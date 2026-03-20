#!/usr/bin/env python3
# 验证 issue 所需的文档文件是否完整
# Usage: python verify-docs.py <issue_number> <epic_name>

import sys
import os
import glob
import re


def main():
    if len(sys.argv) < 3:
        print("ERROR: 必须提供 issue 编号和 epic 名称", file=sys.stderr)
        print("Usage: python verify-docs.py <issue_number> <epic_name>", file=sys.stderr)
        sys.exit(1)

    issue_number = sys.argv[1]
    epic_name = sys.argv[2]

    issue_dir = f".claude/epics/{epic_name}/issues/{issue_number}"
    task_file = f".claude/epics/{epic_name}/{issue_number}.md"
    errors = 0

    print(f"=== 文档完整性验证 - Issue #{issue_number} ===")
    print("")

    # 检查任务文件
    if os.path.isfile(task_file):
        print(f"✅ 任务文件: {task_file}")
    else:
        # 尝试旧命名：在 epic 目录下搜索包含 github issue 引用的 md 文件
        alt_task = ""
        epic_dir = f".claude/epics/{epic_name}"
        if os.path.isdir(epic_dir):
            for md_file in glob.glob(f"{epic_dir}/*.md"):
                try:
                    with open(md_file, "r", encoding="utf-8") as f:
                        if re.search(f"github:.*issues/{issue_number}", f.read()):
                            alt_task = md_file
                            break
                except (IOError, UnicodeDecodeError):
                    continue

        if alt_task:
            print(f"✅ 任务文件: {alt_task}")
        else:
            print("❌ 任务文件: 未找到")
            errors += 1

    # 检查 analysis.md
    analysis_path = f"{issue_dir}/analysis.md"
    if os.path.isfile(analysis_path):
        print(f"✅ analysis.md: {analysis_path}")
        # 检查必要 section
        try:
            with open(analysis_path, "r", encoding="utf-8") as f:
                analysis_content = f.read()
            for section in ["Task Overview", "Business Context", "Technical Approach", "Affected Files", "Implementation Plan"]:
                if f"## {section}" in analysis_content:
                    print(f"   ✅ Section: {section}")
                else:
                    print(f"   ⚠️  Section 缺失: {section}")
        except (IOError, UnicodeDecodeError):
            print("   ⚠️  无法读取 analysis.md")
    else:
        print("❌ analysis.md: 未找到 (必需)")
        errors += 1

    # 检查 progress.md
    progress_path = f"{issue_dir}/progress.md"
    if os.path.isfile(progress_path):
        print(f"✅ progress.md: {progress_path}")
        try:
            with open(progress_path, "r", encoding="utf-8") as f:
                progress_content = f.read()
            for section in ["Code Review Results", "Implementation Log"]:
                if f"## {section}" in progress_content:
                    print(f"   ✅ Section: {section}")
                else:
                    print(f"   ⚠️  Section 缺失: {section}")
        except (IOError, UnicodeDecodeError):
            print("   ⚠️  无法读取 progress.md")
    else:
        print("❌ progress.md: 未找到 (必需)")
        errors += 1

    # 检查 tests 目录
    tests_dir = f"{issue_dir}/tests"
    if os.path.isdir(tests_dir):
        test_count = sum(1 for _ in glob.glob(f"{tests_dir}/**/*", recursive=True) if os.path.isfile(_))
        print(f"✅ tests/: {test_count} 个文件")
    else:
        print("⚠️  tests/: 目录不存在 (建议)")

    print("")
    if errors > 0:
        print(f"❌ 验证失败: {errors} 个必需文档缺失")
        sys.exit(1)
    else:
        print("✅ 所有必需文档已就绪")
        sys.exit(0)


if __name__ == "__main__":
    main()
