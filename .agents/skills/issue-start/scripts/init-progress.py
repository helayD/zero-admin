#!/usr/bin/env python3
# 从模板创建 progress.md
# Usage: python init-progress.py <issue_number> <epic_name> [code_review_summary] [quality_summary] [decision_summary]

import sys
import os
from datetime import datetime, timezone


def main():
    if len(sys.argv) < 3:
        print("ERROR: 必须提供 issue 编号和 epic 名称", file=sys.stderr)
        print("Usage: python init-progress.py <issue_number> <epic_name> [code_review_summary] [quality_summary] [decision_summary]", file=sys.stderr)
        sys.exit(1)

    issue_number = sys.argv[1]
    epic_name = sys.argv[2]
    code_review_summary = sys.argv[3] if len(sys.argv) > 3 else "TBD - 从 Step 2 填充"
    quality_summary = sys.argv[4] if len(sys.argv) > 4 else "TBD - 从 Step 3 填充"
    decision_summary = sys.argv[5] if len(sys.argv) > 5 else "TBD - 从 Step 4 填充"

    issue_dir = f".claude/epics/{epic_name}/issues/{issue_number}"
    progress_file = f"{issue_dir}/progress.md"
    skill_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    template = os.path.join(skill_dir, "templates", "progress-template.md")
    dt = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    # 如果已存在则跳过
    if os.path.isfile(progress_file):
        print(f"progress.md 已存在: {progress_file}")
        sys.exit(0)

    # 创建目录
    os.makedirs(issue_dir, exist_ok=True)

    # 从模板生成
    if os.path.isfile(template):
        with open(template, "r", encoding="utf-8") as f:
            content = f.read()
        content = content.replace("{{ISSUE_NUMBER}}", issue_number)
        content = content.replace("{{DATETIME}}", dt)
        content = content.replace("{{CODE_REVIEW_SUMMARY}}", code_review_summary)
        content = content.replace("{{QUALITY_CHECK_SUMMARY}}", quality_summary)
        content = content.replace("{{IMPLEMENTATION_DECISION}}", decision_summary)
        with open(progress_file, "w", encoding="utf-8") as f:
            f.write(content)
    else:
        content = f"""---
issue: {issue_number}
started: {dt}
status: in_progress
---

# Issue #{issue_number} Progress

## Code Review Results
{code_review_summary}

## Code Quality Check
{quality_summary}

## Implementation Decision
{decision_summary}

## Implementation Log
- [{dt}] Started implementation
"""
        with open(progress_file, "w", encoding="utf-8") as f:
            f.write(content)

    print(f"✅ 已创建 progress.md: {progress_file}")


if __name__ == "__main__":
    main()
