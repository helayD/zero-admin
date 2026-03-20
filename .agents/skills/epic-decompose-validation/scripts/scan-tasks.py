#!/usr/bin/env python3
"""扫描 epic 目录下所有任务文件
Usage: python scan-tasks.py <epic_name>
"""

import sys
import os
import re
import glob

def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供 epic_name")
        sys.exit(1)

    epic_name = sys.argv[1]
    epic_dir = f".claude/epics/{epic_name}"

    if not os.path.isdir(epic_dir):
        print(f"ERROR: Epic 目录不存在: {epic_dir}")
        sys.exit(1)

    print(f"=== 扫描任务文件: {epic_name} ===")

    task_count = 0
    pattern = os.path.join(epic_dir, "[0-9][0-9][0-9]-*.md")
    task_files = sorted(glob.glob(pattern))

    for f in task_files:
        task_count += 1
        basename_f = os.path.basename(f)
        task_num = re.match(r'^(\d+)', basename_f)
        task_num = task_num.group(1) if task_num else ""

        title = ""
        status = "unknown"
        deprecated = "false"

        with open(f, "r", encoding="utf-8") as fh:
            in_frontmatter = False
            frontmatter_count = 0
            for line in fh:
                stripped = line.strip()
                if stripped == "---":
                    frontmatter_count += 1
                    if frontmatter_count == 1:
                        in_frontmatter = True
                        continue
                    elif frontmatter_count == 2:
                        break  # frontmatter 结束
                if in_frontmatter:
                    if stripped.startswith("name:"):
                        title = stripped.split(":", 1)[1].strip()
                    elif stripped.startswith("status:"):
                        status = stripped.split(":", 1)[1].strip()
                    elif stripped.startswith("deprecated:"):
                        deprecated = stripped.split(":", 1)[1].strip()

        if deprecated == "true":
            print(f"SKIP: {basename_f} (deprecated)")
            continue

        print(f"TASK: {basename_f} | num={task_num} | status={status} | title={title}")

    print()
    print(f"TASK_COUNT={task_count}")
    print(f"EPIC_DIR={epic_dir}")

if __name__ == "__main__":
    main()
