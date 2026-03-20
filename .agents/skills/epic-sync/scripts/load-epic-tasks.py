#!/usr/bin/env python3
# Step 2: 加载 Epic 元数据和所有任务文件信息
# Usage: python3 load-epic-tasks.py <feature_name>

import sys
import os
import re
import glob


def parse_frontmatter(filepath):
    """解析文件的 YAML frontmatter"""
    fm = {}
    try:
        with open(filepath, "r", encoding="utf-8") as f:
            content = f.read()
        match = re.match(r'^---\s*\n(.*?)\n---\s*\n', content, re.DOTALL)
        if match:
            for line in match.group(1).splitlines():
                if ':' in line:
                    key, _, value = line.partition(':')
                    fm[key.strip()] = value.strip()
    except Exception:
        pass
    return fm


def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供 feature_name")
        print("Usage: python3 load-epic-tasks.py <feature_name>")
        sys.exit(1)

    feature_name = sys.argv[1]
    epic_dir = f".claude/epics/{feature_name}"
    epic_file = f"{epic_dir}/epic.md"

    print(f"=== 加载 Epic 数据: {feature_name} ===")

    # 1. 验证 Epic 文件存在
    if not os.path.isfile(epic_file):
        print(f"ERROR: Epic 文件不存在: {epic_file}")
        sys.exit(1)

    # 2. 提取 Epic 元数据
    fm = parse_frontmatter(epic_file)
    epic_name = fm.get("name", feature_name)
    epic_github = fm.get("github", "")
    epic_last_sync = fm.get("last_sync", "")

    print(f"📋 Epic: {epic_name}")
    print(f"🔗 GitHub: {epic_github or 'Not yet synced'}")
    print(f"🕐 Last Sync: {epic_last_sync or 'Never'}")
    print()

    # 3. 加载所有任务文件
    task_files = sorted(glob.glob(f"{epic_dir}/[0-9]*.md"))
    # 排除 epic.md 和 github-mapping.md
    task_files = [f for f in task_files if os.path.basename(f) not in ("epic.md", "github-mapping.md")]

    print(f"--- 任务文件 ({len(task_files)} 个) ---")

    create_count = 0
    update_count = 0
    synced_count = 0
    skip_count = 0

    for tf in task_files:
        basename = os.path.basename(tf)
        tfm = parse_frontmatter(tf)

        task_name = tfm.get("name", tfm.get("title", ""))
        if not task_name:
            # 从文件名提取
            task_name = re.sub(r'^\d+-', '', basename).replace('.md', '').replace('-', ' ').replace('_', ' ')

        task_status = tfm.get("status", "unknown")
        task_github = tfm.get("github", "")
        task_updated = tfm.get("updated", "")

        # 提取任务编号
        num_match = re.match(r'^(\d+)', basename)
        task_num = num_match.group(1) if num_match else "000"

        # 决定操作
        if not task_github and task_status == "open":
            action = "CREATE"
            create_count += 1
        elif task_github and epic_last_sync and task_updated > epic_last_sync:
            action = "UPDATE"
            update_count += 1
        elif task_github and (not epic_last_sync or task_updated <= epic_last_sync):
            action = "SYNCED"
            synced_count += 1
        elif not task_github and task_status != "open":
            action = "SKIP"
            skip_count += 1
        elif task_github:
            action = "SYNCED"
            synced_count += 1
        else:
            action = "CREATE"
            create_count += 1

        print(f"  [{task_num}] {basename}")
        print(f"       name={task_name} | status={task_status} | github={task_github or 'N/A'} | action={action}")

    print()
    print(f"--- 同步计划 ---")
    print(f"EPIC_NAME={epic_name}")
    print(f"EPIC_GITHUB={epic_github}")
    print(f"EPIC_LAST_SYNC={epic_last_sync}")
    print(f"TOTAL_TASKS={len(task_files)}")
    print(f"CREATE_COUNT={create_count}")
    print(f"UPDATE_COUNT={update_count}")
    print(f"SYNCED_COUNT={synced_count}")
    print(f"SKIP_COUNT={skip_count}")
    print(f"SYNC_NEEDED={create_count + update_count}")
    print()

    if create_count + update_count > 0:
        print(f"ℹ️  发现 {create_count + update_count} 个任务需要同步（{create_count} 个新建, {update_count} 个更新）")
        print("Auto-proceeding with sync...")
    else:
        print("✅ 所有任务已同步，无需操作")

    print()
    print("✅ 数据加载完成")


if __name__ == "__main__":
    main()
