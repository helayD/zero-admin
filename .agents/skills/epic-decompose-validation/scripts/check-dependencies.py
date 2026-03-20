#!/usr/bin/env python3
"""依赖图检查：检测循环依赖、无效引用
Usage: python check-dependencies.py <epic_name>
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

    print(f"=== 依赖图检查: {epic_name} ===")

    # 收集所有任务编号和文件映射
    pattern = os.path.join(epic_dir, "[0-9][0-9][0-9]-*.md")
    task_files = sorted(glob.glob(pattern))

    task_nums = []
    num_to_file = {}
    for f in task_files:
        basename = os.path.basename(f)
        m = re.match(r'^(\d+)', basename)
        if m:
            num = m.group(1)
            task_nums.append(num)
            num_to_file[num] = f

    if not task_nums:
        print("⚠️  未找到任务文件")
        sys.exit(0)

    print(f"任务总数: {len(task_nums)}")

    errors = 0

    def extract_depends_on(filepath):
        """从文件中提取 depends_on 列表"""
        deps = []
        with open(filepath, "r", encoding="utf-8") as fh:
            for line in fh:
                if line.startswith("depends_on:"):
                    raw = line.split(":", 1)[1].strip()
                    # 移除 [] 和引号
                    raw = raw.strip("[]")
                    raw = raw.replace('"', '').replace("'", "")
                    for part in raw.split(","):
                        part = part.strip()
                        if part:
                            deps.append(part)
                    break
        return deps

    # 检查每个任务的 depends_on
    for task_num, task_file in num_to_file.items():
        deps = extract_depends_on(task_file)

        for dep in deps:
            # 检查依赖目标是否存在
            if dep not in task_nums:
                print(f"❌ Task {task_num}: depends_on 引用了不存在的任务 {dep}")
                errors += 1

            # 简单循环检测（A→B, B→A）
            if dep in num_to_file:
                reverse_deps = extract_depends_on(num_to_file[dep])
                if task_num in reverse_deps:
                    print(f"❌ 循环依赖: Task {task_num} ↔ Task {dep}")
                    errors += 1

    print()
    if errors > 0:
        print(f"DEPENDENCY_ERRORS={errors}")
        print(f"❌ 发现 {errors} 个依赖问题")
        sys.exit(1)
    else:
        print("✅ 依赖图检查通过，无循环依赖")
        sys.exit(0)

if __name__ == "__main__":
    main()
