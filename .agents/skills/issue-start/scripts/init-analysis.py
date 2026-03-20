#!/usr/bin/env python3
# 从模板创建 analysis.md
# Usage: python init-analysis.py <issue_number> <epic_name> [task_summary]

import sys
import os
from datetime import datetime, timezone


def main():
    if len(sys.argv) < 3:
        print("ERROR: 必须提供 issue 编号和 epic 名称", file=sys.stderr)
        print("Usage: python init-analysis.py <issue_number> <epic_name> [task_summary]", file=sys.stderr)
        sys.exit(1)

    issue_number = sys.argv[1]
    epic_name = sys.argv[2]
    task_summary = sys.argv[3] if len(sys.argv) > 3 else "TBD - 从任务文件读取"

    issue_dir = f".claude/epics/{epic_name}/issues/{issue_number}"
    analysis_file = f"{issue_dir}/analysis.md"
    skill_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    template = os.path.join(skill_dir, "templates", "analysis-template.md")
    dt = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    # 如果已存在则跳过
    if os.path.isfile(analysis_file):
        print(f"analysis.md 已存在: {analysis_file}")
        sys.exit(0)

    # 创建目录
    os.makedirs(issue_dir, exist_ok=True)

    # 从模板生成
    if os.path.isfile(template):
        with open(template, "r", encoding="utf-8") as f:
            content = f.read()
        content = content.replace("{{ISSUE_NUMBER}}", issue_number)
        content = content.replace("{{DATETIME}}", dt)
        content = content.replace("{{TASK_SUMMARY}}", task_summary)
        with open(analysis_file, "w", encoding="utf-8") as f:
            f.write(content)
    else:
        # 内联模板兜底
        content = f"""---
issue: {issue_number}
created: {dt}
---

# Issue #{issue_number} Analysis

## Task Overview
{task_summary}

## Business Context
{{TBD in step 1.5}}

## Technical Approach
{{TBD in step 2}}

## Affected Files
{{TBD in step 2}}

## Dependencies & Integration
{{TBD in step 2.5}}

## Implementation Plan
{{TBD in step 4}}

## Risk Mitigation
{{TBD in step 4}}
"""
        with open(analysis_file, "w", encoding="utf-8") as f:
            f.write(content)

    print(f"✅ 已创建 analysis.md: {analysis_file}")


if __name__ == "__main__":
    main()
