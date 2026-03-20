#!/usr/bin/env python3
# Epic Decompose 预检：验证 Epic 存在、统计现有任务、检测孤立任务
# Usage: python3 preflight-check.py <feature_name>

import sys
import os
import re
import glob


def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供 feature_name")
        sys.exit(1)

    feature_name = sys.argv[1]
    epic_dir = f".claude/epics/{feature_name}"
    epic_file = f"{epic_dir}/epic.md"

    print("=== Epic Decompose 预检 ===")

    # 1. 检查 Epic 文件
    if not os.path.isfile(epic_file):
        print(f"ERROR: Epic 文件不存在: {epic_file}")
        sys.exit(1)
    print(f"✅ Epic file: {epic_file}")

    # 2. 检查 Epic frontmatter 完整性
    epic_status = "unknown"
    with open(epic_file, "r", encoding="utf-8") as f:
        epic_content = f.read()

    if not re.search(r'^status:', epic_content, re.MULTILINE):
        print("⚠️  Epic frontmatter 缺少 status 字段")

    match = re.search(r'^status:\s*(.+)', epic_content, re.MULTILINE)
    if match:
        epic_status = match.group(1).strip()

    if epic_status == "completed":
        print("⚠️  WARNING: Epic 状态为 completed，确认是否需要重新拆解")

    # 3. 统计现有任务文件
    task_files = sorted(glob.glob(f"{epic_dir}/[0-9]*.md"))
    task_count = len(task_files)
    print(f"EXISTING_TASKS={task_count}")

    # 4. 列出现有任务文件
    if task_count > 0:
        print()
        print("--- 现有任务文件 ---")
        for f in task_files:
            print(os.path.basename(f))
            title = ""
            status = "unknown"
            try:
                with open(f, "r", encoding="utf-8") as fh:
                    content = fh.read()
                m = re.search(r'^title:\s*(.+)', content, re.MULTILINE)
                if m:
                    title = m.group(1).strip()
                m = re.search(r'^status:\s*(.+)', content, re.MULTILINE)
                if m:
                    status = m.group(1).strip()
            except Exception:
                pass
            print(f"  title: {title or 'N/A'} | status: {status}")

    # 5. 检测孤立任务（文件存在但 Epic 中未引用）
    orphan_count = 0
    print()
    print("--- 孤立任务检测 ---")
    for f in task_files:
        basename = os.path.basename(f)
        match = re.match(r'^(\d+)', basename)
        if match:
            task_num = match.group(1)
            if task_num not in epic_content:
                print(f"ORPHAN: {basename} (未在 Epic 中引用)")
                orphan_count += 1

    if orphan_count == 0:
        print("无孤立任务")
    print(f"ORPHAN_COUNT={orphan_count}")

    print()
    print("=== 预检完成 ===")
    print(f"FEATURE_NAME={feature_name}")
    print(f"EPIC_FILE={epic_file}")
    print(f"EPIC_STATUS={epic_status}")
    print("Continue auto-execution...")


if __name__ == "__main__":
    main()
