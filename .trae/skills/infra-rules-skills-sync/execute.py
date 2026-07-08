#!/usr/bin/env python3
import argparse
import json
import sys
import re
import shutil
import subprocess
from pathlib import Path

# 定义源目录和目标目录
RULES_SRC_DIR = ".agents/rules"
SKILLS_SRC_DIR = ".agents/skills"
AGENTS_SRC_DIR = ".agents/agents"

GEMINI_AGENTS_DIR = ".gemini/agents"

# MCP 同步配置
# SSOT: 用户级 Gemini MCP 配置（全局唯一事实源）
MCP_CONFIG_SRC = Path.home() / ".gemini" / "config" / "mcp_config.json"

# MCP 目标文件列表：各 AI 助手的 MCP 配置文件
# 每项: (相对路径, 格式类型)
#   格式类型说明：
#     "mcpServers"   — 顶层 mcpServers 段，serverUrl/command+args 字段（Gemini/Claude/Cursor 原生格式）
#     "mcp"          — 顶层 mcp 段，type/url/command[]/enabled 字段（OpenCode 格式）
MCP_TARGET_FILES = [
    (".claude/settings.json",    "mcpServers"),
    (".gemini/settings.json",    "mcpServers"),
    (".cursor/mcp.json",         "mcpServers"),
    (".mcp.json",                "mcpServers_stdio_only"),
    ("opencode.json",            "mcp"),
]


CURSOR_RULES_DIR = ".cursor/rules"
LINGMA_RULES_DIR = ".lingma/rules"
CLAUDE_RULES_DIR = ".claude/rules"
TRAE_RULES_DIR = ".trae/rules"
OPENCODE_AGENTS_DIR = ".opencode/agents"
OPENCODE_SKILLS_DIR = ".opencode/skills"
OPENCODE_COMMANDS_DIR = ".opencode/commands"
MIMOCODE_AGENTS_DIR = ".mimocode/agents"
MIMOCODE_SKILLS_DIR = ".mimocode/skills"
MIMOCODE_COMMANDS_DIR = ".mimocode/commands"
CLAUDE_SUBAGENTS_DIR = ".claude/subagents"
CLAUDE_SKILLS_DIR = ".claude/skills"
TRAE_SKILLS_DIR = ".trae/skills"

# 同步目标根目录列表（仅对已落盘的目标目录进行 git 暂存，避免污染源文件）
SYNC_TARGET_ROOTS = [
    CURSOR_RULES_DIR,
    LINGMA_RULES_DIR,
    CLAUDE_RULES_DIR,
    TRAE_RULES_DIR,
    OPENCODE_SKILLS_DIR,
    OPENCODE_COMMANDS_DIR,
    MIMOCODE_SKILLS_DIR,
    MIMOCODE_COMMANDS_DIR,
    CLAUDE_SKILLS_DIR,
    CLAUDE_SUBAGENTS_DIR,
    TRAE_SKILLS_DIR,
    GEMINI_AGENTS_DIR,
    # MCP 目标文件（settings.json / mcp.json / opencode.json，只更新 MCP 段）
    ".claude/settings.json",
    ".gemini/settings.json",
    ".cursor/mcp.json",
    ".mcp.json",
    "opencode.json",
]

# 加载配置文件
CONFIG_PATH = Path(__file__).parent / "sync_config.json"
if CONFIG_PATH.exists():
    try:
        with open(CONFIG_PATH, "r", encoding="utf-8") as f:
            _config = json.load(f)
            COMMAND_SKILLS = _config.get("command_skills", [])
            SUBAGENT_GROUPS = _config.get("subagent_groups", [])
            RULE_GLOBS = _config.get("rule_globs", {})
    except Exception as e:
        print(f"Warning: Failed to load sync_config.json: {e}")
        COMMAND_SKILLS = []
        SUBAGENT_GROUPS = []
        RULE_GLOBS = {}
else:
    print(f"Warning: Configuration file {CONFIG_PATH} not found.")
    COMMAND_SKILLS = []
    SUBAGENT_GROUPS = []
    RULE_GLOBS = {}

def parse_frontmatter(content: str):
    """提取 YAML frontmatter 和 markdown 正文"""
    match = re.match(r'^---\s*\n(.*?)\n---\s*\n', content, re.DOTALL)
    if not match:
        return {}, content
    
    yaml_text = match.group(1)
    body = content[match.end():]
    
    metadata = {}
    for line in yaml_text.splitlines():
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        if ':' in line:
            key, val = line.split(':', 1)
            key = key.strip()
            val = val.strip().strip('"').strip("'")
            metadata[key] = val
    return metadata, body

def get_globs_for_rule(rule_name: str):
    """根据 sync_config.json 或规则前缀智能匹配适配的文件 globs"""
    # 优先读取 sync_config.json 中的精确配置
    if rule_name in RULE_GLOBS:
        return RULE_GLOBS[rule_name]
    # 降级：按前缀推导（比 config 宽松，仅作兜底）
    if rule_name.startswith("backend-"):
        return ["mcp-proxy/**/*", "douyin-mcp/**/*.py", "jules-mcp-server/**/*.py"]
    elif rule_name.startswith("frontend-"):
        return ["web/**/*"]
    else:
        return ["**/*"]



def sync_rules(sync_log, changed_files):
    """同步规则（.agents/rules -> .cursor/rules, .lingma/rules, .claude/rules）"""
    src_dir = Path(RULES_SRC_DIR)
    if not src_dir.exists() or not src_dir.is_dir():
        sync_log.append("源规则目录不存在，跳过规则同步。")
        return

    # 创建目标目录
    Path(CURSOR_RULES_DIR).mkdir(parents=True, exist_ok=True)
    Path(LINGMA_RULES_DIR).mkdir(parents=True, exist_ok=True)
    Path(CLAUDE_RULES_DIR).mkdir(parents=True, exist_ok=True)
    Path(TRAE_RULES_DIR).mkdir(parents=True, exist_ok=True)

    synced_count = 0
    source_rule_names = set()

    for file_path in src_dir.iterdir():
        if file_path.is_file() and file_path.suffix == ".md" and file_path.name != "README.md":
            content = file_path.read_text(encoding="utf-8")
            metadata, body = parse_frontmatter(content)
            rule_name = file_path.stem
            source_rule_names.add(rule_name)

            # 1. 同步到通义灵码 (.lingma/rules/<rule_name>.md) - 直接拷贝源文件
            lingma_rule_path = Path(LINGMA_RULES_DIR) / file_path.name
            lingma_rule_path.write_text(content, encoding="utf-8")
            changed_files.append(lingma_rule_path)

            # 2. 同步到 Claude (.claude/rules/<rule_name>.md) - 直接拷贝源文件
            claude_rule_path = Path(CLAUDE_RULES_DIR) / file_path.name
            claude_rule_path.write_text(content, encoding="utf-8")
            changed_files.append(claude_rule_path)

            # 3. 同步到 Trae (.trae/rules/<rule_name>.md) - 转为 Trae 前端格式
            trae_rule_path = Path(TRAE_RULES_DIR) / file_path.name
            # 读取源规则的 trigger 字段：always_on → alwaysApply: true
            source_always_apply = metadata.get("trigger", "").strip().lower() == "always_on"
            trae_lines = ["---", f"alwaysApply: {'true' if source_always_apply else 'false'}"]
            if not source_always_apply:
                trae_description = metadata.get("description", f"项目规则: {rule_name}")
                trae_lines.append(f"description: {trae_description}")
                trae_globs = get_globs_for_rule(rule_name)
                trae_lines.append(f"globs: {', '.join(trae_globs)}")
            trae_lines.append("---")
            trae_lines.append("")
            trae_lines.append(body.strip())
            trae_content = "\n".join(trae_lines)
            trae_rule_path.write_text(trae_content, encoding="utf-8")
            changed_files.append(trae_rule_path)

            # 4. 同步到 Cursor (.cursor/rules/<rule_name>.mdc) - 转为 MDC 格式
            cursor_rule_path = Path(CURSOR_RULES_DIR) / f"{rule_name}.mdc"

            # 解析 Cursor MDC 的 frontmatter
            description = metadata.get("description", f"项目规则: {rule_name}")
            globs = get_globs_for_rule(rule_name)

            # 判断 alwaysApply（读取源 frontmatter 的 trigger 字段：always_on → true）
            always_apply = "true" if metadata.get("trigger", "").strip().lower() == "always_on" else "false"

            mdc_content = f"""---
description: {description}
globs: {json.dumps(globs)}
alwaysApply: {always_apply}
---

{body.strip()}
"""
            cursor_rule_path.write_text(mdc_content, encoding="utf-8")
            changed_files.append(cursor_rule_path)
            synced_count += 1

    sync_log.append(f"成功同步规则，共计 {synced_count} 个规则文件。")

    # 4. 清理目标规则目录中已失效的内容
    deleted_rules_count = 0

    # 清理 Cursor 规则 (.mdc)
    cursor_dir = Path(CURSOR_RULES_DIR)
    if cursor_dir.exists() and cursor_dir.is_dir():
        for file_path in cursor_dir.iterdir():
            if file_path.is_file() and file_path.suffix == ".mdc":
                if file_path.stem not in source_rule_names:
                    file_path.unlink()
                    deleted_rules_count += 1
                    sync_log.append(f"清理已失效的 Cursor 规则: {file_path.name}")
                    changed_files.append(file_path)  # 标记为已删除

    # 清理通义灵码规则 (.md)
    lingma_dir = Path(LINGMA_RULES_DIR)
    if lingma_dir.exists() and lingma_dir.is_dir():
        for file_path in lingma_dir.iterdir():
            if file_path.is_file() and file_path.suffix == ".md" and file_path.name != "README.md":
                if file_path.stem not in source_rule_names:
                    file_path.unlink()
                    deleted_rules_count += 1
                    sync_log.append(f"清理已失效的通义灵码规则: {file_path.name}")
                    changed_files.append(file_path)

    # 清理 Claude 规则 (.md)
    claude_dir = Path(CLAUDE_RULES_DIR)
    if claude_dir.exists() and claude_dir.is_dir():
        for file_path in claude_dir.iterdir():
            if file_path.is_file() and file_path.suffix == ".md" and file_path.name != "README.md":
                if file_path.stem not in source_rule_names:
                    file_path.unlink()
                    deleted_rules_count += 1
                    sync_log.append(f"清理已失效的 Claude 规则: {file_path.name}")
                    changed_files.append(file_path)

    # 清理 Trae 规则 (.md)
    trae_dir = Path(TRAE_RULES_DIR)
    if trae_dir.exists() and trae_dir.is_dir():
        for file_path in trae_dir.iterdir():
            if file_path.is_file() and file_path.suffix == ".md" and file_path.name != "README.md":
                if file_path.stem not in source_rule_names:
                    file_path.unlink()
                    deleted_rules_count += 1
                    sync_log.append(f"清理已失效的 Trae 规则: {file_path.name}")
                    changed_files.append(file_path)

    if deleted_rules_count > 0:
        sync_log.append(f"清理规则完成，共移除了 {deleted_rules_count} 个已失效的规则文件。")

def validate_skill_name_consistency(sync_log):
    """验证每个技能的 frontmatter name 与文件夹名一致"""
    src_dir = Path(SKILLS_SRC_DIR)
    if not src_dir.exists() or not src_dir.is_dir():
        sync_log.append("源技能目录不存在，跳过技能名一致性验证。")
        return True

    all_ok = True
    for skill_path in sorted(src_dir.iterdir()):
        if not skill_path.is_dir():
            continue
        folder_name = skill_path.name
        skill_md = skill_path / "SKILL.md"
        if not skill_md.exists():
            sync_log.append(f"⚠️ 目录 {folder_name}/ 缺少 SKILL.md")
            all_ok = False
            continue

        content = skill_md.read_text(encoding="utf-8")
        metadata, _ = parse_frontmatter(content)
        skill_name = metadata.get("name", "")
        if skill_name != folder_name:
            sync_log.append(f"❌ 名称不一致: 文件夹={folder_name!r}, frontmatter name={skill_name!r}")
            all_ok = False

    if all_ok:
        sync_log.append("✅ 所有技能的 frontmatter name 与文件夹名一致。")
    else:
        sync_log.append("⚠️ 存在名称不一致的技能，请优先修复后再同步。")
    return all_ok


def sync_skills(sync_log, changed_files):
    """同步技能（.agents/skills -> .opencode/skills, .claude/subagents）"""
    src_dir = Path(SKILLS_SRC_DIR)
    if not src_dir.exists() or not src_dir.is_dir():
        sync_log.append("源技能目录不存在，跳过技能同步。")
        return

    # 先执行技能名一致性验证（不一致则停止，避免同步脏数据）
    if not validate_skill_name_consistency(sync_log):
        sync_log.append("❌ 因技能名不一致，终止同步。请修复后再运行。")
        return

    # 创建目标目录
    Path(OPENCODE_SKILLS_DIR).mkdir(parents=True, exist_ok=True)
    Path(OPENCODE_COMMANDS_DIR).mkdir(parents=True, exist_ok=True)
    Path(MIMOCODE_SKILLS_DIR).mkdir(parents=True, exist_ok=True)
    Path(MIMOCODE_COMMANDS_DIR).mkdir(parents=True, exist_ok=True)
    Path(CLAUDE_SKILLS_DIR).mkdir(parents=True, exist_ok=True)
    Path(CLAUDE_SUBAGENTS_DIR).mkdir(parents=True, exist_ok=True)
    Path(TRAE_SKILLS_DIR).mkdir(parents=True, exist_ok=True)

    synced_count = 0
    for skill_path in src_dir.iterdir():
        if skill_path.is_dir():
            skill_name = skill_path.name
            skill_md_path = skill_path / "SKILL.md"
            if not skill_md_path.exists():
                continue

            content = skill_md_path.read_text(encoding="utf-8")
            metadata, body = parse_frontmatter(content)
            description = metadata.get("description", f"项目技能: {skill_name}")

            # 1. 复制整个技能文件夹到 .opencode/skills/<skill_name>/
            dest_skill_dir = Path(OPENCODE_SKILLS_DIR) / skill_name
            if dest_skill_dir.exists():
                shutil.rmtree(dest_skill_dir)
            shutil.copytree(skill_path, dest_skill_dir)
            for f in dest_skill_dir.rglob("*"):
                if f.is_file():
                    changed_files.append(f)

            # 1b. 复制整个技能文件夹到 .mimocode/skills/<skill_name>/
            dest_mimo_skill_dir = Path(MIMOCODE_SKILLS_DIR) / skill_name
            if dest_mimo_skill_dir.exists():
                shutil.rmtree(dest_mimo_skill_dir)
            shutil.copytree(skill_path, dest_mimo_skill_dir)
            for f in dest_mimo_skill_dir.rglob("*"):
                if f.is_file():
                    changed_files.append(f)

            # 2. 复制整个技能文件夹到 .claude/skills/<skill_name>/
            claude_skill_dir = Path(CLAUDE_SKILLS_DIR) / skill_name
            if claude_skill_dir.exists():
                shutil.rmtree(claude_skill_dir)
            shutil.copytree(skill_path, claude_skill_dir)
            for f in claude_skill_dir.rglob("*"):
                if f.is_file():
                    changed_files.append(f)

            # 2b. 复制整个技能文件夹到 .trae/skills/<skill_name>/
            trae_skill_dir = Path(TRAE_SKILLS_DIR) / skill_name
            if trae_skill_dir.exists():
                shutil.rmtree(trae_skill_dir)
            shutil.copytree(skill_path, trae_skill_dir)
            for f in trae_skill_dir.rglob("*"):
                if f.is_file():
                    changed_files.append(f)

            # 3. 同步到命令目录 (如果配置了)
            if skill_name in COMMAND_SKILLS:
                command_md = f"---\ndescription: {description}\n---\n\n{body.strip()}\n"

                # 同步到 .opencode
                command_file = Path(OPENCODE_COMMANDS_DIR) / f"{skill_name}.md"
                command_file.write_text(command_md, encoding="utf-8")
                changed_files.append(command_file)

                # 同步到 .mimocode
                mimo_command_file = Path(MIMOCODE_COMMANDS_DIR) / f"{skill_name}.md"
                mimo_command_file.write_text(command_md, encoding="utf-8")
                changed_files.append(mimo_command_file)

            synced_count += 1

    # 4. 清理目标技能目录中已失效的内容
    deleted_skills_count = 0
    source_skill_names = {d.name for d in src_dir.iterdir() if d.is_dir()}

    # 清理 .opencode/skills 中的过期目录
    opencode_skills_path = Path(OPENCODE_SKILLS_DIR)
    if opencode_skills_path.exists() and opencode_skills_path.is_dir():
        for d in opencode_skills_path.iterdir():
            if d.is_dir() and d.name not in source_skill_names:
                for f in d.rglob("*"):
                    if f.is_file():
                        changed_files.append(f)
                shutil.rmtree(d)
                deleted_skills_count += 1
                sync_log.append(f"清理已失效的 OpenCode 技能目录: {d.name}")

    # 清理 .mimocode/skills 中的过期目录
    mimocode_skills_path = Path(MIMOCODE_SKILLS_DIR)
    if mimocode_skills_path.exists() and mimocode_skills_path.is_dir():
        for d in mimocode_skills_path.iterdir():
            if d.is_dir() and d.name not in source_skill_names:
                for f in d.rglob("*"):
                    if f.is_file():
                        changed_files.append(f)
                shutil.rmtree(d)
                deleted_skills_count += 1
                sync_log.append(f"清理已失效的 Mimocode 技能目录: {d.name}")

    # 清理 .claude/skills 中的过期目录
    claude_skills_path = Path(CLAUDE_SKILLS_DIR)
    if claude_skills_path.exists() and claude_skills_path.is_dir():
        for d in claude_skills_path.iterdir():
            if d.is_dir() and d.name not in source_skill_names:
                for f in d.rglob("*"):
                    if f.is_file():
                        changed_files.append(f)
                shutil.rmtree(d)
                deleted_skills_count += 1
                sync_log.append(f"清理已失效的 Claude 技能目录: {d.name}")

    # 清理 .trae/skills 中的过期目录
    trae_skills_path = Path(TRAE_SKILLS_DIR)
    if trae_skills_path.exists() and trae_skills_path.is_dir():
        for d in trae_skills_path.iterdir():
            if d.is_dir() and d.name not in source_skill_names:
                for f in d.rglob("*"):
                    if f.is_file():
                        changed_files.append(f)
                shutil.rmtree(d)
                deleted_skills_count += 1
                sync_log.append(f"清理已失效的 Trae 技能目录: {d.name}")

    # 清理 .opencode/commands 中的过期命令文件
    opencode_cmds_path = Path(OPENCODE_COMMANDS_DIR)
    if opencode_cmds_path.exists() and opencode_cmds_path.is_dir():
        for f in opencode_cmds_path.iterdir():
            if f.is_file() and f.suffix == ".md" and f.name != "README.md":
                if f.stem not in source_skill_names:
                    f.unlink()
                    deleted_skills_count += 1
                    sync_log.append(f"清理已失效的 OpenCode 命令: {f.name}")
                    changed_files.append(f)

    # 清理 .mimocode/commands 中的过期命令文件
    mimocode_cmds_path = Path(MIMOCODE_COMMANDS_DIR)
    if mimocode_cmds_path.exists() and mimocode_cmds_path.is_dir():
        for f in mimocode_cmds_path.iterdir():
            if f.is_file() and f.suffix == ".md" and f.name != "README.md":
                if f.stem not in source_skill_names:
                    f.unlink()
                    deleted_skills_count += 1
                    sync_log.append(f"清理已失效的 Mimocode 命令: {f.name}")
                    changed_files.append(f)

    if deleted_skills_count > 0:
        sync_log.append(f"清理技能完成，共移除了 {deleted_skills_count} 个已失效的技能或命令文件。")

    # 3. 生成合并后的 Claude Code 子代理 (.claude/subagents/<group>.md + .json)
    #    先清理旧文件：删除名称匹配单个技能名的 .md/.json（旧自动生成产物）
    skill_dir_names = {d.name for d in src_dir.iterdir() if d.is_dir()}
    for old_file in Path(CLAUDE_SUBAGENTS_DIR).iterdir():
        stem = old_file.stem
        if stem in skill_dir_names and old_file.suffix in (".md", ".json"):
            old_file.unlink()
            changed_files.append(old_file)

    for group in SUBAGENT_GROUPS:
        group_name = group["name"]
        group_desc = group["description"]
        skill_names = group["skills"]

        # 收集该组所有技能的 body 内容
        combined_body_parts = []
        for sn in skill_names:
            skill_path = src_dir / sn / "SKILL.md"
            if skill_path.exists():
                _, body = parse_frontmatter(skill_path.read_text(encoding="utf-8"))
                combined_body_parts.append(f"## {sn}\n\n{body.strip()}")
        combined_body = "\n\n---\n\n".join(combined_body_parts)

        group_md = Path(CLAUDE_SUBAGENTS_DIR) / f"{group_name}.md"
        group_md.write_text(f"""---
name: {group_name}
description: {group_desc}
model: inherit
---

# {group_name}

来源：`.agents/skills/` 合并组: {", ".join(skill_names)}

{combined_body}
""", encoding="utf-8")
        changed_files.append(group_md)

        group_json = Path(CLAUDE_SUBAGENTS_DIR) / f"{group_name}.json"
        group_json.write_text(json.dumps({
            "name": group_name,
            "tools": {
                "bash": "ask",
                "read_file": "allow",
                "write_file": "ask",
            },
        }, indent=2, ensure_ascii=False), encoding="utf-8")
        changed_files.append(group_json)

    sync_log.append(f"成功同步技能，共计 {synced_count} 个技能文件夹。")
    sync_log.append(f"生成合并子代理，共 {len(SUBAGENT_GROUPS)} 个（{', '.join(g['name'] for g in SUBAGENT_GROUPS)}）。")

def sync_agents(sync_log, changed_files):
    """同步 Gemini 智能体 (.agents/agents -> .gemini/agents)"""
    src_dir = Path(AGENTS_SRC_DIR)
    if not src_dir.exists() or not src_dir.is_dir():
        sync_log.append("源智能体目录不存在，跳过智能体同步。")
        return

    # 创建目标目录
    dest_dir = Path(GEMINI_AGENTS_DIR)
    dest_dir.mkdir(parents=True, exist_ok=True)

    synced_count = 0
    source_agent_names = set()

    for file_path in src_dir.iterdir():
        if file_path.is_file() and file_path.suffix == ".json" and file_path.name != "README.md":
            content = file_path.read_text(encoding="utf-8")
            source_agent_names.add(file_path.name)

            dest_path = dest_dir / file_path.name
            dest_path.write_text(content, encoding="utf-8")
            changed_files.append(dest_path)
            synced_count += 1

    sync_log.append(f"成功同步 Gemini 智能体，共计 {synced_count} 个智能体文件。")

    # 清理已失效的智能体 (清理失效的 .json 或残留的 .md 文件)
    deleted_count = 0
    if dest_dir.exists() and dest_dir.is_dir():
        for file_path in dest_dir.iterdir():
            if file_path.is_file() and file_path.name != "README.md":
                if file_path.suffix == ".md" or (file_path.suffix == ".json" and file_path.name not in source_agent_names):
                    file_path.unlink()
                    deleted_count += 1
                    sync_log.append(f"清理已失效的 Gemini 智能体: {file_path.name}")
                    changed_files.append(file_path)

    if deleted_count > 0:
        sync_log.append(f"清理智能体完成，共移除了 {deleted_count} 个已失效的智能体文件。")


def _convert_mcp_for_target(ssot_mcp: dict, target_format: str) -> dict:
    """将 Gemini SSOT 的 MCP 配置转换为目标格式

    SSOT 格式 (mcpServers):
      {"serverUrl": "...", "headers": {...}, "disabled": true}
      {"command": "...", "args": [...], "disabled": true}  (stdio)

    target_format="mcpServers" (Claude/Cursor/.mcp.json):
      serverUrl → url, 其余字段保留

    target_format="mcp" (OpenCode):
      远程: → {"type": "remote", "url": "...", "headers": {...}, "enabled": bool}
      stdio: → {"type": "local", "command": [...]}
    """
    result = {}
    for name, cfg in ssot_mcp.items():
        if not isinstance(cfg, dict):
            result[name] = cfg
            continue

        entry = json.loads(json.dumps(cfg))  # 深拷贝
        disabled = entry.pop("disabled", False)

        if target_format == "mcpServers":
            # 转换 serverUrl -> url
            if "serverUrl" in entry:
                entry["url"] = entry.pop("serverUrl")
            # 保留 stdio 字段 (command/args/type) 不变
            # 保留 disabled 字段（部分工具支持）
            if disabled:
                entry["disabled"] = True
            result[name] = entry

        elif target_format == "mcpServers_stdio_only":
            # .mcp.json 只支持 stdio 型服务器（必须有 command 字段）
            if "command" in entry:
                # 转换 serverUrl -> url（兼容 stdio+url 混合场景）
                entry.pop("serverUrl", None)
                if disabled:
                    entry["disabled"] = True
                result[name] = entry
            # 远程 HTTP MCP（只有 url/serverUrl 无 command）跳过
            # 不写入 result，等价于过滤

        elif target_format == "mcp":
            # OpenCode 格式
            if "command" in entry:
                # stdio -> local
                cmd = entry["command"]
                args = entry.get("args", [])
                if isinstance(cmd, str):
                    cmd_list = [cmd] + list(args)
                else:
                    cmd_list = list(cmd) + list(args)
                opencode_entry = {
                    "type": "local",
                    "command": cmd_list,
                }
                # OpenCode 的 stdio server 不需要 enabled 字段（有即启用）
                # 但如果 SSOT 标记了 disabled，保持 disabled 语义
                if disabled:
                    opencode_entry["enabled"] = False
            elif "serverUrl" in entry:
                # 远程
                opencode_entry = {
                    "type": "remote",
                    "url": entry["serverUrl"],
                }
                if entry.get("headers"):
                    opencode_entry["headers"] = entry["headers"]
                if disabled:
                    opencode_entry["enabled"] = False
            else:
                # 兜底：原样保留
                opencode_entry = entry

            result[name] = opencode_entry

        else:
            result[name] = entry

    return result


def sync_mcp_servers(sync_log, changed_files):
    """同步 MCP 服务器配置

    SSOT: ~/.gemini/config/mcp_config.json（全局唯一事实源）
    目标文件: .claude/settings.json, .gemini/settings.json, .cursor/mcp.json,
              .mcp.json, opencode.json

    同步策略：将 SSOT 的 mcpServers 段转换后写入目标文件的对应段，
    保留目标文件其他所有字段不变。
    """
    src_path = MCP_CONFIG_SRC
    if not src_path.exists():
        sync_log.append(f"MCP 配置源文件不存在: {src_path}，跳过 MCP 同步。")
        return

    try:
        with open(src_path, "r", encoding="utf-8") as f:
            ssot_data = json.load(f)
    except (json.JSONDecodeError, OSError) as e:
        sync_log.append(f"读取 MCP 配置源文件失败: {e}，跳过 MCP 同步。")
        return

    ssot_mcp = ssot_data.get("mcpServers")
    if ssot_mcp is None:
        sync_log.append("MCP 配置源文件中缺少 mcpServers 字段，跳过 MCP 同步。")
        return

    synced_count = 0
    for rel_path, target_format in MCP_TARGET_FILES:
        target_path = Path(rel_path)
        if not target_path.exists():
            sync_log.append(f"MCP 目标文件不存在，跳过: {rel_path}")
            continue

        try:
            with open(target_path, "r", encoding="utf-8") as f:
                target_data = json.load(f)
        except (json.JSONDecodeError, OSError) as e:
            sync_log.append(f"读取 MCP 目标文件失败: {rel_path} - {e}")
            continue

        # 转换并写入对应段
        converted = _convert_mcp_for_target(ssot_mcp, target_format)
        if target_format == "mcp":
            target_data["mcp"] = converted
        else:
            target_data["mcpServers"] = converted

        try:
            with open(target_path, "w", encoding="utf-8") as f:
                json.dump(target_data, f, indent=2, ensure_ascii=False)
                f.write("\n")
            changed_files.append(target_path)
            synced_count += 1
            sync_log.append(f"MCP 配置已同步: {rel_path} ({target_format})")
        except OSError as e:
            sync_log.append(f"写入 MCP 目标文件失败: {rel_path} - {e}")

    sync_log.append(f"MCP 服务器配置同步完成，共更新 {synced_count} 个目标文件。")


def run_audit(sync_log):
    """同步后调用审计脚本验证结果

    审计脚本位于 `.agents/skills/agent-rules-auditor/scripts/audit_rules.py`
    （路径在 2026-06 技能重命名时修正）。
    """
    audit_script = ".agents/skills/agent-rules-auditor/scripts/audit_rules.py"
    if not Path(audit_script).exists():
        sync_log.append("审计脚本不存在，跳过同步后验证。")
        return

    try:
        result = subprocess.run(
            [sys.executable, audit_script, "--json"],
            capture_output=True, text=True, timeout=30,
        )
        audit_result = json.loads(result.stdout)
        if audit_result["status"] == "pass":
            sync_log.append("审计验证：所有规则已同步。")
        else:
            for issue in audit_result["issues"]:
                sync_log.append(f"审计提醒 [{issue['tool']}] {issue['message']}")
    except Exception as e:
        sync_log.append(f"审计验证失败: {e}")


def _is_in_git_repo():
    """检测当前目录是否位于 git 仓库中"""
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--is-inside-work-tree"],
            capture_output=True, text=True, timeout=5,
        )
        return result.returncode == 0 and result.stdout.strip() == "true"
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False


def _is_git_ignored(rel_path: str) -> bool:
    """判断给定相对路径是否被 .gitignore 忽略（通过 git check-ignore 检测）"""
    try:
        result = subprocess.run(
            ["git", "check-ignore", "--", rel_path],
            capture_output=True, text=True, timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False


def auto_stage_files(sync_log, changed_files, target_roots):
    """将同步过程中变更的目标文件加入 git 暂存区

    仅对位于 `target_roots` 之下的路径执行暂存操作，避免误把源文件
    （`.agents/rules/`、`.agents/skills/`、`.agents/agents/`）一并 stage。
    - 已存在的文件 -> `git add`（自动跳过被 .gitignore 忽略的文件）
    - 已删除的文件 -> `git rm`
    """
    if not changed_files:
        sync_log.append("无文件变更，跳过 git auto-stage。")
        return

    if not _is_in_git_repo():
        sync_log.append("当前目录不在 git 仓库中，跳过 git auto-stage。")
        return

    cwd = Path.cwd().resolve()
    target_paths = [Path(p).resolve() for p in target_roots]

    def _under_targets(p: Path) -> bool:
        try:
            p_resolved = p.resolve()
        except (OSError, RuntimeError):
            return False
        for root in target_paths:
            try:
                p_resolved.relative_to(root)
                return True
            except ValueError:
                continue
        return False

    added, removed, ignored, failed = 0, 0, 0, 0
    for p in changed_files:
        if not _under_targets(p):
            continue
        try:
            rel_path = str(p.resolve().relative_to(cwd))
        except ValueError:
            continue

        if not p.exists():
            # 已删除 -> git rm
            proc = subprocess.run(
                ["git", "rm", "--", rel_path],
                capture_output=True, text=True, timeout=10,
            )
            if proc.returncode == 0:
                removed += 1
            else:
                failed += 1
                sync_log.append(f"git rm 失败: {rel_path} - {proc.stderr.strip()}")
            continue

        proc = subprocess.run(
            ["git", "add", "--", rel_path],
            capture_output=True, text=True, timeout=10,
        )
        if proc.returncode == 0:
            added += 1
        elif "ignored by one of your .gitignore files" in proc.stderr:
            # 文件在 .gitignore 中，不报 failure，仅计入 ignored
            ignored += 1
        else:
            failed += 1
            sync_log.append(f"git add 失败: {rel_path} - {proc.stderr.strip()}")

    sync_log.append(
        f"git auto-stage 完成：新增/更新 {added} 个、删除 {removed} 个、忽略 {ignored} 个、失败 {failed} 个。"
    )


def main():
    parser = argparse.ArgumentParser(description="Synchronize all agent rules, skills and workflows across multiple platforms")
    parser.add_argument("--input", type=str, required=True, help="Input JSON payload")
    args = parser.parse_args()

    # 简单解析输入载荷
    try:
        payload = json.loads(args.input)
    except json.JSONDecodeError:
        payload = {}

    # 输入控制项：
    # - force_refresh: 兼容旧参数，保留但暂未使用
    # - auto_stage: 是否在同步后自动 git add（默认 True）
    # - run_audit_after: 是否在同步后调用审计脚本（默认 True）
    auto_stage = payload.get("auto_stage", True)
    run_audit_after = payload.get("run_audit_after", True)

    changed_files = []
    sync_log = []
    try:
        sync_rules(sync_log, changed_files)
        sync_skills(sync_log, changed_files)
        sync_agents(sync_log, changed_files)
        sync_mcp_servers(sync_log, changed_files)

        if run_audit_after:
            run_audit(sync_log)

        if auto_stage:
            auto_stage_files(sync_log, changed_files, SYNC_TARGET_ROOTS)

        output = {
            "status": "success",
            "message": "同步执行成功",
            "data": {
                "sync_details": sync_log,
                "payload_received": payload,
                "auto_stage_enabled": auto_stage,
                "audit_enabled": run_audit_after,
                "changed_files_count": len(changed_files),
            }
        }
        print(json.dumps(output, ensure_ascii=False))
    except Exception as e:
        output = {
            "status": "error",
            "message": f"同步执行失败: {str(e)}",
            "data": None
        }
        print(json.dumps(output, ensure_ascii=False))
        sys.exit(1)

if __name__ == "__main__":
    main()
