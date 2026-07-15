#!/usr/bin/env python3
"""
git-auto-suite 执行器

支持双模式：
  - staged: 仅使用 `git add` 已暂存内容，按已暂存变更直接提交
  - diff  : 暂存区为空时，按工作区 diff 自动分类、按类验证并选择性提交

默认行为（auto 模式）：
  - 暂存区非空 -> 模式 staged
  - 暂存区为空但工作区有变更 -> 模式 diff
  - 暂存区为空且工作区无变更 -> 终止

详见 SKILL.md
"""
from __future__ import annotations

import argparse
import json
import os
import re
import shlex
import subprocess
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# 常量
# ---------------------------------------------------------------------------

REPORT_VERSION = "2.0"
DEFAULT_REPORT_PATH = "./git-auto-suite-report.json"

PROTECTED_BRANCHES = {"main", "master", "develop", "release/*"}

# 自动跳过的路径（生成产物 / 临时文件 / CI 产物）
DEFAULT_EXCLUDE_PATTERNS = [
    re.compile(r"(^|/)node_modules/"),
    re.compile(r"(^|/)dist/"),
    re.compile(r"(^|/)build/"),
    re.compile(r"(^|/)\.next/"),
    re.compile(r"(^|/)coverage/"),
    re.compile(r"(^|/)__pycache__/"),
    re.compile(r"(^|/)\.pytest_cache/"),
    re.compile(r"(^|/)\.mypy_cache/"),
    re.compile(r"(^|/)\.ruff_cache/"),
    re.compile(r"\.(log|tmp|bak|swp)$"),
    re.compile(r"\.lockb$"),
    re.compile(r"(^|/)frontend-v2/test-results/"),
    re.compile(r"(^|/)frontend-v2/playwright-report/"),
    re.compile(r"(^|/)frontend-v2/blob-report/"),
]

# 锁文件：单独聚类
LOCKFILE_BASENAMES = {
    "package-lock.json",
    "pnpm-lock.yaml",
    "yarn.lock",
    "uv.lock",
    "poetry.lock",
    "Pipfile.lock",
    "composer.lock",
    "Gemfile.lock",
    "bun.lockb",
    "bun.lock",
    "Cargo.lock",
    "go.sum",
}


# ---------------------------------------------------------------------------
# 基础工具
# ---------------------------------------------------------------------------

def _log(msg: str) -> None:
    print(msg, file=sys.stderr, flush=True)


def _run(
    cmd: List[str],
    cwd: str = ".",
    timeout: int = 300,
    check: bool = False,
    env: Optional[Dict[str, str]] = None,
) -> Tuple[int, str, str]:
    """同步执行命令并返回 (returncode, stdout, stderr)。"""
    try:
        proc = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=timeout,
            env={**os.environ, **(env or {})},
        )
    except subprocess.TimeoutExpired as e:
        return 124, "", f"timeout after {timeout}s: {' '.join(shlex.quote(c) for c in cmd)}"
    except FileNotFoundError as e:
        return 127, "", f"command not found: {e}"
    if check and proc.returncode != 0:
        raise RuntimeError(
            f"command failed (rc={proc.returncode}): {' '.join(shlex.quote(c) for c in cmd)}\n{proc.stderr}"
        )
    return proc.returncode, proc.stdout, proc.stderr


def _is_excluded(path: str) -> bool:
    for pat in DEFAULT_EXCLUDE_PATTERNS:
        if pat.search(path):
            return True
    return False


def _short_subject(s: str, limit: int = 50) -> str:
    """规范化中文 subject：去除句末标点 + 截断到 limit。"""
    s = s.strip()
    s = s.rstrip("。.!?！？")
    if len(s) > limit:
        s = s[: limit - 1] + "…"
    return s


# ---------------------------------------------------------------------------
# Git 探测
# ---------------------------------------------------------------------------

def _git_current_branch(repo: str) -> str:
    rc, out, _ = _run(["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=repo)
    if rc != 0:
        raise RuntimeError("无法解析当前分支，请确认当前位于 Git 仓库内。")
    return out.strip()


def _git_status_porcelain(repo: str) -> Tuple[List[str], List[str]]:
    """返回 (staged_paths, unstaged_or_untracked_paths)，均为相对仓库根的路径。"""
    rc, out, err = _run(["git", "status", "--porcelain=1", "-z"], cwd=repo)
    if rc != 0:
        raise RuntimeError(f"git status 失败: {err}")
    staged: List[str] = []
    others: List[str] = []
    if not out:
        return staged, others
    for entry in out.split("\x00"):
        if not entry:
            continue
        # 格式: XY<space><path>  其中 X 是 index 状态，Y 是 worktree 状态
        if len(entry) < 4:
            continue
        index_status = entry[0]
        path = entry[3:]
        # 重命名时 path 形如 "old -> new"
        if " -> " in path:
            path = path.split(" -> ", 1)[1]
        if index_status not in (" ", "?"):
            staged.append(path)
        else:
            others.append(path)
    return staged, others


def _git_name_status(repo: str, paths: List[str]) -> List[Tuple[str, str]]:
    """返回 [(status, path), ...]，仅追踪文件。"""
    if not paths:
        return []
    rc, out, err = _run(
        ["git", "diff", "--name-status", "--"] + paths, cwd=repo
    )
    if rc != 0:
        raise RuntimeError(f"git diff --name-status 失败: {err}")
    results: List[Tuple[str, str]] = []
    for line in out.splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        if len(parts) < 2:
            continue
        status = parts[0]
        path = parts[-1]
        if " -> " in path:
            path = path.split(" -> ", 1)[1]
        results.append((status, path))
    return results


def _expand_untracked(repo: str) -> List[str]:
    """展开未跟踪目录，返回所有未跟踪叶子文件路径（相对仓库根）。

    解决空仓库中 `git status` 把未跟踪目录显示为 `?? backend/` 而非具体文件的问题。
    只返回**文件**，目录条目会被自动忽略。
    """
    rc, out, _ = _run(
        [
            "git", "ls-files", "--others", "--exclude-standard", "-z",
        ],
        cwd=repo,
    )
    if rc != 0:
        return []
    return [p for p in out.split("\x00") if p and not p.endswith("/")]


def _filter_files_only(paths: List[str], repo: str) -> List[str]:
    """过滤掉目录条目，只保留文件（已被 git 跟踪或未跟踪的）。"""
    kept: List[str] = []
    for p in paths:
        if not p:
            continue
        if p.endswith("/"):
            continue
        # 二次保险：本地 fs 检查
        if Path(repo, p).is_file():
            kept.append(p)
    return kept


def _git_diff_patch(repo: str, paths: List[str]) -> str:
    if not paths:
        return ""
    rc, out, _ = _run(["git", "diff", "-U0", "--"] + paths, cwd=repo, timeout=120)
    return out if rc == 0 else ""


# ---------------------------------------------------------------------------
# 数据结构
# ---------------------------------------------------------------------------

@dataclass
class Complexity:
    score: int = 0
    level: str = "light"  # light | standard | strict
    signals: Dict[str, Dict] = field(default_factory=dict)

    def to_dict(self) -> Dict:
        return {
            "score": self.score,
            "level": self.level,
            "signals": self.signals,
        }


@dataclass
class Bucket:
    id: str
    type: str
    scope: str
    files: List[str] = field(default_factory=list)
    subject: str = ""
    verify_commands: List[str] = field(default_factory=list)
    verify_level: str = "light"
    complexity: Complexity = field(default_factory=Complexity)
    verify_passed: bool = False
    verify_stderr: str = ""
    commit_sha: str = ""
    status: str = "pending"  # pending | committed | failed
    pre_commit_sha: str = ""
    fix_iterations: int = 0

    def to_dict(self) -> Dict:
        return {
            "id": self.id,
            "type": self.type,
            "scope": self.scope,
            "files": self.files,
            "subject": self.subject,
            "verify": {
                "commands": self.verify_commands,
                "level": self.verify_level,
                "passed": self.verify_passed,
                "stderr": self.verify_stderr[:4000] if self.verify_stderr else "",
            },
            "complexity": self.complexity.to_dict(),
            "commit_sha": self.commit_sha,
            "pre_commit_sha": self.pre_commit_sha,
            "fix_iterations": self.fix_iterations,
            "status": self.status,
        }


# ---------------------------------------------------------------------------
# 分类逻辑
# ---------------------------------------------------------------------------

_DOC_PATTERN = re.compile(r"(^|/)docs/|(^|/)[^/]+\.md$|README", re.IGNORECASE)
_CONFIG_PATTERN = re.compile(
    r"(^|/)(Dockerfile|docker-compose\.ya?ml|\.gitignore|\.editorconfig|\.gitattributes|"
    r"\.eslintrc.*|\.prettierrc.*|\.npmrc|\.nvmrc|pyproject\.toml|setup\.cfg|setup\.py|"
    r"Makefile|requirements.*\.txt|\.github/.*)$"
)
_TEST_PATTERN = re.compile(r"(^|/)(test|tests|__tests__)/|(^|/)[^/]+\.(test|spec)\.[A-Za-z]+$")
_I18N_PATTERN = re.compile(r"(^|/)(locales|i18n|messages)/")
_DB_MIGRATION_PATTERN = re.compile(
    r"(^|/)alembic/versions/|(^|/)sqls/|(^|/)backend/migrations/|(^|/)backend/alembic/"
)
_AGENT_CONFIG_PATTERN = re.compile(
    r"(^|/)\.agents/|(^|/)\.cursor/|(^|/)\.claude/|(^|/)\.lingma/|(^|/)\.trae/|(^|/)\.opencode/|"
    r"^/?AGENTS\.md$|(^|/)CLAUDE\.md$"
)


def _infer_scope_from_path(path: str) -> str:
    parts = path.replace("\\", "/").split("/")
    parts = [p for p in parts if p]
    for candidate in ("backend", "frontend-v2", "admin-frontend", "frontend", "docs"):
        if candidate in parts:
            return candidate
    return parts[0] if parts else "root"


def _looks_like_fix(patch: str) -> bool:
    """通过 diff 关键字粗略判断是否偏 fix。"""
    if not patch:
        return False
    keywords = re.compile(
        r"\b(BUG|FIX|FIXME|TODO|XXX|HACK|fix(?:es|ed)?|bug)\b", re.IGNORECASE
    )
    new_lines = [
        ln[1:].lstrip("+").strip()
        for ln in patch.splitlines()
        if ln.startswith("+") and not ln.startswith("+++")
    ]
    joined = "\n".join(new_lines[:200])
    return bool(keywords.search(joined))


def classify_files(
    paths: List[str],
    name_status: List[Tuple[str, str]],
    repo: str,
) -> List[Bucket]:
    """根据路径与 diff 启发式把变更分桶。"""
    ns_map = {p: s for s, p in name_status}
    # 跳过被排除的路径
    filtered = [p for p in paths if not _is_excluded(p)]
    if not filtered:
        return []

    groups: Dict[str, Bucket] = {}
    for path in filtered:
        basename = os.path.basename(path)
        bucket_id: Optional[str] = None
        ctype = "feat"
        scope = _infer_scope_from_path(path)

        if basename in LOCKFILE_BASENAMES:
            bucket_id = "chore-deps"
            ctype = "chore"
            scope = "deps"
        elif _AGENT_CONFIG_PATTERN.search(path):
            bucket_id = "agent-config"
            ctype = "chore"
            scope = "agents"
        elif _DB_MIGRATION_PATTERN.search(path):
            bucket_id = "db-migration"
            ctype = "feat"
            scope = "db"
        elif _DOC_PATTERN.search(path):
            bucket_id = "docs"
            ctype = "docs"
            scope = _infer_scope_from_path(path) or "global"
        elif _CONFIG_PATTERN.search(path):
            bucket_id = "chore-config"
            ctype = "chore"
            scope = "infra"
        elif _TEST_PATTERN.search(path):
            bucket_id = "test"
            ctype = "test"
            scope = scope
        elif _I18N_PATTERN.search(path):
            bucket_id = "i18n"
            ctype = "feat"
            scope = "i18n"
        else:
            bucket_id = f"{scope}:code"
            ctype = "feat"

        key = bucket_id or f"{scope}:code"
        if key not in groups:
            groups[key] = Bucket(id=key, type=ctype, scope=scope)
        groups[key].files.append(path)

    # 对 code 桶依据 diff 关键字拆分 feat / fix
    final_buckets: Dict[str, Bucket] = {}
    for key, bucket in groups.items():
        if not key.endswith(":code"):
            final_buckets[key] = bucket
            continue
        patch = _git_diff_patch(repo, bucket.files)
        if _looks_like_fix(patch):
            fix_bucket = Bucket(
                id=f"{key}:fix", type="fix", scope=bucket.scope, files=list(bucket.files)
            )
            feat_bucket = Bucket(
                id=f"{key}:feat", type="feat", scope=bucket.scope, files=[]
            )
            # 仅在仅 fix 倾向时整体归为 fix
            if not _has_actual_new_feature(patch):
                final_buckets[fix_bucket.id] = fix_bucket
            else:
                final_buckets[fix_bucket.id] = fix_bucket
                final_buckets[feat_bucket.id] = feat_bucket
        else:
            final_buckets[key] = bucket

    return list(final_buckets.values())


def _has_actual_new_feature(patch: str) -> bool:
    if not patch:
        return False
    added = [
        ln[1:].lstrip("+").strip()
        for ln in patch.splitlines()
        if ln.startswith("+") and not ln.startswith("+++")
    ]
    # 启发：存在多个新增函数定义
    func_defs = sum(1 for ln in added if re.match(r"(async\s+)?def\s+\w+", ln))
    return func_defs >= 2


# ---------------------------------------------------------------------------
# 复杂度评估（按修改内容动态判定验证档位）
# ---------------------------------------------------------------------------

_SENSITIVE_PATH_PATTERNS = [
    re.compile(r"(^|/)migrations/"),
    re.compile(r"(^|/)alembic/versions/"),
    re.compile(r"(^|/)models/"),
    re.compile(r"(^|/)auth/"),
    re.compile(r"(^|/)security/"),
    re.compile(r"(^|/)permissions?[\w-]*\.py$"),
    re.compile(r"(^|/)rbac[\w-]*\.py$"),
    re.compile(r"(^|/)workflow/"),
    re.compile(r"scoped_state\.py$"),
    re.compile(r"conftest\.py$"),
    re.compile(r"__init__\.py$"),
    re.compile(r"\.env($|\.)"),
    re.compile(r"\.pem$"),
    re.compile(r"\.key$"),
]

_BREAKING_PATTERNS = [
    re.compile(r"^-def\s+\w+", re.MULTILINE),
    re.compile(r"^-class\s+\w+", re.MULTILINE),
    re.compile(r"^-export\s+", re.MULTILINE),
    re.compile(r"BREAKING\s+CHANGE", re.IGNORECASE),
    re.compile(r"removed\s+(function|class|export)", re.IGNORECASE),
]


def _count_diff_lines(patch: str) -> int:
    if not patch:
        return 0
    total = 0
    for ln in patch.splitlines():
        if not ln:
            continue
        if ln.startswith(("+++", "---")):
            continue
        if ln.startswith(("+", "-")):
            total += 1
    return total


def _first_level_subdirs(paths: List[str]) -> set:
    """提取每个文件所在的「第一级子目录」（相对仓库根），用于 D5 判定。"""
    subs: set = set()
    for p in paths:
        parts = p.replace("\\", "/").split("/")
        parts = [x for x in parts if x]
        if len(parts) >= 2:
            subs.add(parts[1])
        else:
            subs.add("<root>")
    return subs


def assess_complexity(bucket: Bucket, repo: str) -> Complexity:
    """根据 diff 内容评估该分类的复杂度，决定验证档位。

    评分维度（D1~D5）见 SKILL.md B.3.2。
    """
    patch = _git_diff_patch(repo, bucket.files)

    # D1 变更行数
    diff_lines = _count_diff_lines(patch)
    if diff_lines <= 50:
        d1 = 0
    elif diff_lines <= 200:
        d1 = 1
    elif diff_lines <= 500:
        d1 = 2
    else:
        d1 = 3

    # D2 变更文件数
    n_files = len(bucket.files)
    if n_files <= 1:
        d2 = 0
    elif n_files <= 5:
        d2 = 1
    elif n_files <= 15:
        d2 = 2
    else:
        d2 = 3

    # D3 触及敏感文件
    d3_hits = []
    for p in bucket.files:
        for pat in _SENSITIVE_PATH_PATTERNS:
            if pat.search(p):
                d3_hits.append(p)
                break
    d3 = 3 if d3_hits else 0

    # D4 破坏性改动
    d4_hits = []
    for pat in _BREAKING_PATTERNS:
        m = pat.search(patch or "")
        if m:
            d4_hits.append(m.group(0))
    d4 = 2 if d4_hits else 0

    # D5 跨多模块依赖
    subs = _first_level_subdirs(bucket.files)
    d5 = 1 if len(subs) >= 3 else 0

    total = d1 + d2 + d3 + d4 + d5
    if d3 > 0 or d4 > 0 or total >= 6:
        level = "strict"
    elif total >= 3:
        level = "standard"
    else:
        level = "light"

    return Complexity(
        score=total,
        level=level,
        signals={
            "D1_diff_lines": {"score": d1, "value": diff_lines},
            "D2_file_count": {"score": d2, "value": n_files},
            "D3_sensitive_paths": {
                "score": d3,
                "hits": d3_hits[:5],
                "total_hits": len(d3_hits),
            },
            "D4_breaking_change": {"score": d4, "hits": d4_hits[:5]},
            "D5_multi_module": {
                "score": d5,
                "subdirs": sorted(subs),
            },
        },
    )


# ---------------------------------------------------------------------------
# 验证逻辑
# ---------------------------------------------------------------------------

def build_verify_commands(
    bucket: Bucket,
    repo: str,
    full_verify: bool,
) -> List[str]:
    """根据 bucket 的 scope + complexity.level 决定具体验证命令。"""
    paths = bucket.files or []
    scope = bucket.scope
    level = bucket.complexity.level if bucket.complexity else "light"
    if full_verify:
        level = "strict"
    cmds: List[str] = []

    if scope == "backend":
        py_paths = [p for p in paths if p.endswith(".py")]
        if py_paths:
            cmds.append("ruff check --no-fix " + " ".join(shlex.quote(p) for p in py_paths))
            if level in ("standard", "strict"):
                cmds.append(
                    "mypy --no-incremental --follow-imports=silent "
                    + " ".join(shlex.quote(p) for p in py_paths)
                )
            if level == "strict":
                cmds.append("python -m pytest -q " + " ".join(shlex.quote(p) for p in py_paths))
    elif scope in ("frontend-v2", "admin-frontend", "frontend"):
        frontend_root = (
            "frontend-v2" if scope == "frontend-v2"
            else "admin-frontend" if scope == "admin-frontend"
            else "frontend"
        )
        ts_paths = [p for p in paths if p.endswith((".ts", ".tsx", ".js", ".jsx"))]
        if ts_paths:
            cmds.append(
                f"cd {frontend_root} && npx --no-install eslint --quiet "
                + " ".join(shlex.quote(p) for p in ts_paths)
            )
            if level in ("standard", "strict"):
                cmds.append(
                    f"cd {frontend_root} && npx --no-install tsc --noEmit"
                )
            if level == "strict":
                cmds.append(
                    f"cd {frontend_root} && npm run build --silent"
                )
    elif scope == "docs" or bucket.id == "docs":
        md_paths = [p for p in paths if p.endswith(".md")]
        if level == "light" or not md_paths:
            pass
        elif level == "standard":
            if md_paths:
                cmds.append(
                    f"npx --no-install markdownlint "
                    + " ".join(shlex.quote(p) for p in md_paths)
                )
        elif level == "strict":
            if md_paths:
                cmds.append(
                    f"npx --no-install markdown-link-check "
                    + " ".join(shlex.quote(p) for p in md_paths)
                )
    elif scope == "agents" or bucket.id == "agent-config":
        for p in paths:
            if p.endswith((".yaml", ".yml")):
                cmds.append(
                    f"python3 -c \"import sys, yaml; yaml.safe_load(open(sys.argv[1]))\" {shlex.quote(p)}"
                )
        if level == "standard" and any(p.endswith(".md") for p in paths):
            if Path(repo, ".agents/skills/general-agent-rules-auditor/scripts/audit_rules.py").exists():
                cmds.append(
                    "python3 .agents/skills/general-agent-rules-auditor/scripts/audit_rules.py --quiet"
                )
    elif scope == "db" or bucket.id == "db-migration":
        if Path(repo, "alembic.ini").exists():
            cmds.append("alembic heads")
            if level in ("standard", "strict"):
                cmds.append("alembic check")
            if level == "strict":
                cmds.append("alembic upgrade --sql head:base")
    elif bucket.id == "chore-deps":
        for p in paths:
            if p.endswith(".json"):
                cmds.append(
                    f"python3 -c \"import sys, json; json.load(open(sys.argv[1]))\" {shlex.quote(p)}"
                )
            elif p.endswith((".yaml", ".yml")):
                cmds.append(
                    f"python3 -c \"import sys, yaml; yaml.safe_load(open(sys.argv[1]))\" {shlex.quote(p)}"
                )
        if level == "strict" and any(p.endswith("package-lock.json") for p in paths):
            cmds.append("node -e \"require('./package-lock.json')\"")
    elif bucket.id == "chore-config":
        for p in paths:
            if p.endswith(".json"):
                cmds.append(
                    f"python3 -c \"import sys, json; json.load(open(sys.argv[1]))\" {shlex.quote(p)}"
                )
            elif p.endswith((".yaml", ".yml")):
                cmds.append(
                    f"python3 -c \"import sys, yaml; yaml.safe_load(open(sys.argv[1]))\" {shlex.quote(p)}"
                )
    # test / i18n 默认无专属命令
    return cmds


def run_verify(cmd: str, repo: str) -> Tuple[bool, str]:
    rc, _, stderr = _run(["bash", "-lc", cmd], cwd=repo, timeout=600)
    return rc == 0, stderr


def verify_bucket(bucket: Bucket, repo: str, full_verify: bool) -> None:
    cmds = build_verify_commands(bucket, repo, full_verify)
    bucket.verify_commands = cmds
    bucket.verify_level = bucket.complexity.level if bucket.complexity else "light"
    if not cmds:
        bucket.verify_passed = True
        bucket.verify_stderr = ""
        return
    for cmd in cmds:
        ok, err = run_verify(cmd, repo)
        if not ok:
            bucket.verify_passed = False
            bucket.verify_stderr = err
            return
    bucket.verify_passed = True
    bucket.verify_stderr = ""


# ---------------------------------------------------------------------------
# 提交
# ---------------------------------------------------------------------------

def _build_commit_message(bucket: Bucket) -> Tuple[str, str]:
    title = _short_subject(
        bucket.subject
        or _default_subject_for_bucket(bucket)
    )
    head = f"{bucket.type}({bucket.scope}): {title}"
    body_lines = [
        f"- 涉及文件 ({len(bucket.files)}):",
        *[f"  - {p}" for p in bucket.files[:20]],
    ]
    if len(bucket.files) > 20:
        body_lines.append(f"  - ... 及其他 {len(bucket.files) - 20} 个文件")
    if bucket.verify_commands:
        body_lines.append("- 验证通过: " + " / ".join(bucket.verify_commands))
    body = "\n".join(body_lines)
    return head, body


def _default_subject_for_bucket(bucket: Bucket) -> str:
    label_map = {
        "chore-deps": "同步依赖锁文件",
        "chore-config": "更新项目配置",
        "agent-config": "更新智能体规则配置",
        "db-migration": "新增数据库迁移",
        "test": "更新测试用例",
        "i18n": "更新国际化文案",
        "docs": "更新文档",
    }
    if bucket.id in label_map:
        return label_map[bucket.id]
    # 默认从首个文件提取业务名
    if bucket.files:
        first = bucket.files[0]
        return f"优化 {os.path.basename(first).split('.')[0]} 模块"
    return f"提交 {bucket.scope} 变更"


def _git_add_paths(paths: List[str], repo: str) -> Tuple[int, str, str]:
    if not paths:
        return 0, "", ""
    staged, _ = _git_status_porcelain(repo)
    staged_set = set(staged)
    to_add = []
    for p in paths:
        if Path(repo, p).is_file() or p not in staged_set:
            to_add.append(p)
    if not to_add:
        return 0, "", ""
    rc, out, err = _run(["git", "add", "--"] + to_add, cwd=repo)
    return rc, out, err


def _git_unstage_paths(paths: List[str], repo: str) -> Tuple[int, str, str]:
    if not paths:
        return 0, "", ""
    rc, out, err = _run(["git", "reset", "HEAD", "--"] + paths, cwd=repo)
    return rc, out, err


def _git_commit(title: str, body: str, repo: str) -> Tuple[int, str, str]:
    return _run(
        ["git", "commit", "-m", title, "-m", body], cwd=repo, timeout=300
    )


def _git_head(repo: str) -> str:
    rc, out, _ = _run(["git", "rev-parse", "HEAD"], cwd=repo)
    return out.strip() if rc == 0 else ""


def _git_status_after(paths: List[str], repo: str) -> bool:
    """检查给定路径是否仍有未暂存/未跟踪内容。"""
    rc, out, _ = _run(["git", "status", "--porcelain=1", "-z", "--"] + paths, cwd=repo)
    return rc == 0 and bool(out)


def commit_bucket(bucket: Bucket, repo: str) -> bool:
    """对单个 bucket 执行 add -> verify -> commit，成功返回 True。"""
    if not bucket.verify_passed:
        return False
    bucket.pre_commit_sha = _git_head(repo)
    rc, _, err = _git_add_paths(bucket.files, repo)
    if rc != 0:
        bucket.status = "failed"
        bucket.verify_stderr = f"git add 失败: {err}"
        return False
    title, body = _build_commit_message(bucket)
    rc, _, err = _git_commit(title, body, repo)
    if rc != 0:
        # 撤回暂存
        _git_unstage_paths(bucket.files, repo)
        bucket.status = "failed"
        bucket.verify_stderr = f"git commit 失败: {err}"
        return False
    new_head = _git_head(repo)
    bucket.commit_sha = new_head
    bucket.status = "committed"
    return True


# ---------------------------------------------------------------------------
# 推送
# ---------------------------------------------------------------------------

def push_branch(repo: str, branch: str) -> Tuple[bool, str]:
    if branch in PROTECTED_BRANCHES:
        return False, f"拒绝直接推送受保护分支: {branch}"
    rc, _, err = _run(
        ["git", "push", "origin", branch], cwd=repo, timeout=600
    )
    return rc == 0, err


# ---------------------------------------------------------------------------
# 编排
# ---------------------------------------------------------------------------

def _emit_plan(plan: Dict) -> None:
    _log("[阶段 0] 执行计划预览")
    _log(json.dumps(plan, ensure_ascii=False, indent=2))


def _confirm(prompt: str) -> bool:
    try:
        ans = input(f"{prompt} [y/N]: ").strip().lower()
    except EOFError:
        return False
    return ans in ("y", "yes")


def run_mode_staged(
    repo: str,
    args: argparse.Namespace,
) -> Dict:
    staged, _ = _git_status_porcelain(repo)
    if not staged:
        return {
            "status": "noop",
            "mode": "staged",
            "message": "暂存区为空，未执行任何提交。",
        }
    name_status = _git_name_status(repo, staged)
    bucket = Bucket(
        id="staged:all",
        type="feat",
        scope=_infer_scope_from_path(staged[0]) if staged else "root",
        files=staged,
    )
    bucket.complexity = assess_complexity(bucket, repo)
    if args.verify_level and args.verify_level != "auto":
        bucket.complexity.level = args.verify_level
    _log(
        f"\n[复杂度判定] staged:all score={bucket.complexity.score} "
        f"→ {bucket.complexity.level}"
    )
    verify_bucket(bucket, repo, args.full_verify)
    ok = commit_bucket(bucket, repo) if bucket.verify_passed else False
    branch = _git_current_branch(repo)
    pushed = False
    push_msg = ""
    if ok and args.push:
        pushed, push_msg = push_branch(repo, branch)
    return {
        "status": "committed" if ok else "failed",
        "mode": "staged",
        "branch": branch,
        "buckets": [bucket.to_dict()] if ok else [],
        "failed_buckets": [] if ok else [bucket.to_dict()],
        "push": {"enabled": args.push, "pushed": pushed, "message": push_msg},
    }


def _collect_changes(repo: str) -> Tuple[List[str], List[str]]:
    """收集暂存区为空时的所有变更：未暂存修改 + 未跟踪文件（已展开）。"""
    _, others = _git_status_porcelain(repo)
    untracked = _expand_untracked(repo)
    # 合并并去重 + 过滤掉目录条目
    seen = set()
    combined: List[str] = []
    for p in others + untracked:
        if not p or p.endswith("/") or p in seen:
            continue
        if _is_excluded(p):
            continue
        seen.add(p)
        combined.append(p)
    combined = _filter_files_only(combined, repo)
    return [], combined


def run_mode_diff(
    repo: str,
    args: argparse.Namespace,
) -> Dict:
    staged, _ = _git_status_porcelain(repo)
    if staged:
        return {
            "status": "skipped",
            "mode": "diff",
            "message": "暂存区非空；如需按 diff 模式运行，请先 `git reset HEAD`。",
        }
    _, combined = _collect_changes(repo)
    if not combined:
        return {
            "status": "noop",
            "mode": "diff",
            "message": "暂存区与工作区均无变更。",
        }
    name_status = _git_name_status(repo, combined)
    buckets = classify_files(combined, name_status, repo)
    if not buckets:
        return {
            "status": "noop",
            "mode": "diff",
            "message": "所有变更均被排除规则过滤，无需提交。",
        }

    # 阶段 0.5：按修改内容复杂度评估每个分类，必要时按 --verify-level 覆盖
    for bucket in buckets:
        bucket.complexity = assess_complexity(bucket, repo)
        if args.verify_level and args.verify_level != "auto":
            bucket.complexity.level = args.verify_level

    # 终端打印复杂度判定结果，便于阶段 0 确认
    _log("\n[分类与复杂度判定]")
    for b in buckets:
        c = b.complexity
        _log(
            f"  - {b.id:<30} score={c.score} → {c.level}  "
            f"(D1={c.signals['D1_diff_lines']['score']} "
            f"D2={c.signals['D2_file_count']['score']} "
            f"D3={c.signals['D3_sensitive_paths']['score']} "
            f"D4={c.signals['D4_breaking_change']['score']} "
            f"D5={c.signals['D5_multi_module']['score']})"
        )

    branch = _git_current_branch(repo)
    committed: List[Bucket] = []
    failed: List[Bucket] = []

    for bucket in buckets:
        # 修复循环
        while bucket.fix_iterations < args.max_iterations:
            verify_bucket(bucket, repo, args.full_verify)
            if bucket.verify_passed:
                break
            bucket.fix_iterations += 1
            # 尝试自动修复
            auto_ok = _try_auto_fix(bucket, repo)
            if not auto_ok:
                break
        if not bucket.verify_passed:
            bucket.status = "failed"
            failed.append(bucket)
            continue
        if commit_bucket(bucket, repo):
            committed.append(bucket)
        else:
            bucket.status = "failed"
            failed.append(bucket)

    pushed = False
    push_msg = ""
    if args.push and committed:
        pushed, push_msg = push_branch(repo, branch)

    status = "success" if committed and not failed else (
        "partial" if committed else "failed"
    )
    return {
        "status": status,
        "mode": "diff",
        "branch": branch,
        "buckets": [b.to_dict() for b in committed],
        "failed_buckets": [b.to_dict() for b in failed],
        "push": {"enabled": args.push, "pushed": pushed, "message": push_msg},
    }


_AUTO_FIX_HANDLERS = {
    ("backend", "ruff"): (
        "ruff check --fix {paths}",
        re.compile(r"^backend/.*\.py$"),
    ),
    ("frontend-v2", "eslint"): (
        "cd frontend-v2 && npx --no-install eslint --fix {paths}",
        re.compile(r"^frontend-v2/.*\.(ts|tsx|js|jsx)$"),
    ),
    ("admin-frontend", "eslint"): (
        "cd admin-frontend && npx --no-install eslint --fix {paths}",
        re.compile(r"^admin-frontend/.*\.(ts|tsx|js|jsx)$"),
    ),
}


def _try_auto_fix(bucket: Bucket, repo: str) -> bool:
    """尝试对桶内文件做一次自动修复，返回是否尝试了修复动作。"""
    for (scope, _label), (template, pat) in _AUTO_FIX_HANDLERS.items():
        if scope != bucket.scope:
            continue
        targets = [p for p in bucket.files if pat.match(p)]
        if not targets:
            continue
        cmd = template.format(paths=" ".join(shlex.quote(p) for p in targets))
        rc, _, _ = _run(["bash", "-lc", cmd], cwd=repo, timeout=600)
        if rc == 0:
            return True
    return False


# ---------------------------------------------------------------------------
# 入口
# ---------------------------------------------------------------------------

def _resolve_mode(staged_present: bool, unstaged_present: bool, mode: str) -> str:
    if mode == "staged":
        return "staged"
    if mode == "diff":
        return "diff"
    # auto
    if staged_present:
        return "staged"
    if unstaged_present:
        return "diff"
    return "noop"


def main() -> int:
    parser = argparse.ArgumentParser(
        description="git-auto-suite 执行器：按暂存区或 diff 分类选择性提交"
    )
    parser.add_argument("--input", type=str, required=True, help="JSON 输入载荷")
    parser.add_argument("--repo", type=str, default=".", help="Git 仓库根目录")
    parser.add_argument(
        "--mode",
        choices=["auto", "staged", "diff"],
        default="auto",
    )
    parser.add_argument("--auto-yes", action="store_true", help="跳过阶段 0 确认")
    parser.add_argument("--push", action="store_true", help="提交完成后推送")
    parser.add_argument("--pr", action="store_true", help="推送后创建 draft PR")
    parser.add_argument("--full-verify", action="store_true", help="启用 build / pytest")
    parser.add_argument(
        "--verify-level",
        choices=["auto", "light", "standard", "strict"],
        default="auto",
        help="按修改内容复杂度自动判定，或强制使用某档验证",
    )
    parser.add_argument("--max-iterations", type=int, default=3)
    parser.add_argument("--parallel", action="store_true")
    parser.add_argument("--report", type=str, default=DEFAULT_REPORT_PATH)
    args = parser.parse_args()

    try:
        payload = json.loads(args.input)
    except json.JSONDecodeError as e:
        _print_output({
            "status": "error",
            "message": f"Invalid JSON input: {e}",
            "data": None,
        })
        return 1

    repo = args.repo
    try:
        staged, unstaged = _git_status_porcelain(repo)
    except RuntimeError as e:
        _print_output({"status": "error", "message": str(e), "data": None})
        return 1

    mode = _resolve_mode(bool(staged), bool(unstaged), args.mode)
    plan = {
        "detected_mode": mode,
        "staged_count": len(staged),
        "unstaged_count": len(unstaged),
        "push": args.push,
        "pr": args.pr,
        "full_verify": args.full_verify,
        "verify_level": args.verify_level,
        "max_iterations": args.max_iterations,
    }
    _emit_plan(plan)

    if mode == "noop":
        report = {
            "version": REPORT_VERSION,
            "status": "noop",
            "mode": "noop",
            "message": "无任何待提交变更。",
        }
        _write_report(args.report, report)
        _print_output({
            "status": "success",
            "message": "无变更",
            "data": report,
        })
        return 0

    if not args.auto_yes:
        if not _confirm("确认按以上计划执行？"):
            _print_output({
                "status": "cancelled",
                "message": "用户取消执行。",
                "data": plan,
            })
            return 0

    if mode == "staged":
        report_body = run_mode_staged(repo, args)
    else:
        report_body = run_mode_diff(repo, args)

    report = {
        "version": REPORT_VERSION,
        "generated_at": int(time.time()),
        "plan": plan,
        **report_body,
    }
    _write_report(args.report, report)
    _print_output({
        "status": "success" if report_body.get("status") in ("success", "partial", "committed") else "failed",
        "message": f"git-auto-suite 执行完成，状态: {report_body.get('status')}",
        "data": report,
    })
    return 0


def _write_report(path: str, report: Dict) -> None:
    try:
        with open(path, "w", encoding="utf-8") as f:
            json.dump(report, f, ensure_ascii=False, indent=2)
    except OSError as e:
        _log(f"[警告] 写入报告失败: {e}")


def _print_output(obj: Dict) -> None:
    print(json.dumps(obj, ensure_ascii=False))


if __name__ == "__main__":
    main()
