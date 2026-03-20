#!/usr/bin/env python3
# Quick Check: 验证环境、获取 Issue 详情、定位 Epic 和任务文件
# Usage: python quick-check.py <issue_number>

import sys
import os
import subprocess
import glob
import re


def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供 issue 编号", file=sys.stderr)
        sys.exit(1)

    issue_number = sys.argv[1]

    # 1. 获取 issue 详情
    print(f"=== 获取 Issue #{issue_number} 详情 ===")
    try:
        result = subprocess.run(
            ["gh", "issue", "view", issue_number, "--json", "state,title,labels,body"],
            capture_output=True, text=True, check=True
        )
        print(result.stdout)
    except (subprocess.CalledProcessError, FileNotFoundError):
        print(f"ERROR: 无法访问 issue #{issue_number}。请检查编号或运行: gh auth login", file=sys.stderr)
        sys.exit(1)

    # 2. 查找本地任务文件并提取 epic
    print("")
    print("=== 查找本地任务文件 ===")

    task_file = ""
    # 按文件名查找
    for match in glob.glob(f".claude/epics/**/{issue_number}.md", recursive=True):
        task_file = match
        break

    # 按内容查找
    if not task_file:
        pattern = f"github:.*issues/{issue_number}"
        for md_file in glob.glob(".claude/epics/**/*.md", recursive=True):
            try:
                with open(md_file, "r", encoding="utf-8") as f:
                    if re.search(pattern, f.read()):
                        task_file = md_file
                        break
            except (IOError, UnicodeDecodeError):
                continue

    if task_file:
        # 从路径提取 epic 名称: .claude/epics/<epic_name>/...
        m = re.match(r"\.claude/epics/([^/]+)/", task_file)
        epic_name = m.group(1) if m else ""
        print(f"TASK_FILE={task_file}")
        print(f"EPIC_NAME={epic_name}")
    else:
        print(f"ERROR: 未找到 issue #{issue_number} 的本地任务文件。该 issue 可能不是通过 PM 系统创建的。", file=sys.stderr)
        sys.exit(1)

    # 3. 检查 analysis 是否存在
    print("")
    print("=== 检查分析文档 ===")
    issue_dir = f".claude/epics/{epic_name}/issues/{issue_number}"
    analysis_file = f"{issue_dir}/analysis.md"

    if os.path.isfile(analysis_file):
        print("ANALYSIS_EXISTS=true")
        print(f"ANALYSIS_FILE={analysis_file}")
    else:
        print("ANALYSIS_EXISTS=false")
        print(f"NOTICE: 未找到 issue #{issue_number} 的分析文档，将在 Step 1 中创建。")

    # 4. 检查 progress 是否存在
    progress_file = f"{issue_dir}/progress.md"
    if os.path.isfile(progress_file):
        print("PROGRESS_EXISTS=true")
        print(f"PROGRESS_FILE={progress_file}")
    else:
        print("PROGRESS_EXISTS=false")

    print("")
    print("=== Quick Check 完成 ===")
    print(f"ISSUE_DIR={issue_dir}")


if __name__ == "__main__":
    main()
