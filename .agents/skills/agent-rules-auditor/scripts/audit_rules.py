#!/usr/bin/env python3
import argparse
import json
import os
import glob
import sys

MASTER_DIR = ".agents/rules"
TARGETS = {
    "Cursor": (".cursor/rules", ".mdc"),
    "Claude": (".claude/rules", ".md"),
    "Trae": (".trae/rules", ".md"),
    "Lingma": (".lingma/rules", ".md"),
}


def get_master_rules():
    rules = glob.glob(os.path.join(MASTER_DIR, "*.md"))
    return [os.path.basename(r) for r in rules if os.path.basename(r) != "README.md"]


def strip_fm(content):
    if content.startswith("---"):
        parts = content.split("---", 2)
        if len(parts) >= 3:
            return parts[2].strip()
    return content.strip()


def audit():
    master_files = get_master_rules()
    issues = []

    for tool, (dir_path, ext) in TARGETS.items():
        if not os.path.exists(dir_path):
            continue

        for master_file in master_files:
            target_filename = master_file.replace(".md", ext)
            target_path = os.path.join(dir_path, target_filename)

            if not os.path.exists(target_path):
                issues.append({
                    "tool": tool,
                    "type": "missing_rule",
                    "file": target_filename,
                    "message": f"缺少规则: {target_filename}",
                })
                continue

            master_path = os.path.join(MASTER_DIR, master_file)
            with open(master_path, encoding="utf-8") as f:
                master_content = f.read()
            with open(target_path, encoding="utf-8") as f:
                target_content = f.read()

            master_body = strip_fm(master_content)
            target_body = strip_fm(target_content)

            if master_body != target_body:
                issues.append({
                    "tool": tool,
                    "type": "desync",
                    "file": target_filename,
                    "message": f"内容不同步: {target_filename}",
                })

    return issues


def main():
    parser = argparse.ArgumentParser(description="审计多工具规则同步一致性")
    parser.add_argument("--json", action="store_true", help="以 JSON 格式输出结果")
    args = parser.parse_args()

    issues = audit()
    passed = len(issues) == 0

    if args.json:
        print(json.dumps({
            "status": "pass" if passed else "fail",
            "total_issues": len(issues),
            "issues": issues,
        }, ensure_ascii=False))
    else:
        master_files = get_master_rules()
        print(f"共发现 {len(master_files)} 个主规则文件")
        for issue in issues:
            print(f"[{issue['tool']}] {issue['message']}")
        if passed:
            print("\n所有规则已同步。")
        else:
            print(f"\n共 {len(issues)} 个问题。")

    sys.exit(0 if passed else 1)


if __name__ == "__main__":
    main()
