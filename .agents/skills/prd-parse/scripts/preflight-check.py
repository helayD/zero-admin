#!/usr/bin/env python3
"""PRD 预检验证：验证参数、PRD 文件存在性、frontmatter、内容完整性、已有 Epic 检查
Usage: python preflight-check.py <feature_name>
"""

import sys
import os
import re

def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供 feature_name")
        sys.exit(1)

    feature_name = sys.argv[1]
    prd_file = f".claude/prds/{feature_name}.md"
    epic_dir = f".claude/epics/{feature_name}"
    epic_file = f"{epic_dir}/epic.md"

    print(f"=== PRD 预检验证: {feature_name} ===")

    # 1. 验证 PRD 文件存在
    if not os.path.isfile(prd_file):
        print(f"❌ PRD 不存在: {prd_file}")
        print(f"请先创建: /prd-new {feature_name}")
        sys.exit(1)
    print(f"✅ PRD file: {prd_file}")

    with open(prd_file, "r", encoding="utf-8") as f:
        content = f.read()
    lines = content.splitlines()

    # 2. 验证 PRD frontmatter
    print("\n--- Frontmatter 检查 ---")
    required_fields = ["name:", "description:", "status:", "created:"]
    missing_fields = []
    for field in required_fields:
        if not any(line.startswith(field) for line in lines):
            missing_fields.append(field)

    if missing_fields:
        print(f"❌ PRD frontmatter 缺少字段: {' '.join(missing_fields)}")
        print(f"请检查: {prd_file}")
        sys.exit(1)
    print("✅ Frontmatter 完整")

    # 3. 验证 PRD 内容完整性
    print("\n--- 内容完整性检查 ---")
    required_sections = ["User Stor", "Functional Requirement", "Success Criteria"]
    missing_sections = []
    content_lower = content.lower()
    for section in required_sections:
        if section.lower() not in content_lower:
            missing_sections.append(section)

    if missing_sections:
        print("⚠️  PRD 内容可能不完整:")
        for s in missing_sections:
            print(f"  ⚠️  缺少: {s}")
        print("Epic 覆盖率可能受影响。")
    else:
        print("✅ 核心 section 齐全")

    # 4. 统计 PRD 信息
    prd_lines = len(lines)
    api_pattern = re.compile(r'endpoint|api|POST|GET|PUT|DELETE|PATCH', re.IGNORECASE)
    api_count = sum(1 for line in lines if api_pattern.search(line))
    table_pattern = re.compile(r'CREATE TABLE|数据表|表结构|table_name', re.IGNORECASE)
    table_count = sum(1 for line in lines if table_pattern.search(line))
    story_pattern = re.compile(r'^\s*-.*作为|As a|用户可以', re.IGNORECASE)
    user_story_count = sum(1 for line in lines if story_pattern.search(line))

    print("\n--- PRD 统计 ---")
    print(f"PRD_LINES={prd_lines}")
    print(f"API_MENTIONS={api_count}")
    print(f"TABLE_MENTIONS={table_count}")
    print(f"USER_STORY_MENTIONS={user_story_count}")

    # 5. 检查已有 Epic
    print("\n--- Epic 检查 ---")
    if os.path.isfile(epic_file):
        epic_status = "unknown"
        with open(epic_file, "r", encoding="utf-8") as f:
            for line in f:
                if line.startswith("status:"):
                    epic_status = line.split(":", 1)[1].strip()
                    break
        print(f"⚠️  Epic '{feature_name}' 已存在 (status: {epic_status})")
        print("EPIC_EXISTS=true")
        print(f"EPIC_STATUS={epic_status}")
    else:
        print("✅ 无已有 Epic，可以创建")
        print("EPIC_EXISTS=false")

    # 6. 验证目录权限
    if not os.path.isdir(".claude/epics"):
        try:
            os.makedirs(".claude/epics", exist_ok=True)
        except OSError:
            print("❌ 无法创建 epic 目录，请检查权限")
            sys.exit(1)

    print("\n=== 预检完成 ===")
    print(f"FEATURE_NAME={feature_name}")
    print(f"PRD_FILE={prd_file}")
    print(f"EPIC_DIR={epic_dir}")

if __name__ == "__main__":
    main()
