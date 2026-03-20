#!/usr/bin/env python3
"""单任务文件验证：检查 frontmatter、metadata、技术细节、自包含性
Usage: python validate-task.py <task_file>
输出验证结果和分数
"""

import sys
import os
import re

def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供任务文件路径")
        sys.exit(1)

    task_file = sys.argv[1]

    if not os.path.isfile(task_file):
        print(f"ERROR: 文件不存在: {task_file}")
        sys.exit(1)

    basename = os.path.basename(task_file)
    score = 100
    issues = []
    critical = 0

    def add_issue(severity, msg, deduct):
        nonlocal score, critical
        issues.append(f"  {severity} {msg}")
        score -= deduct
        if severity == "❌":
            critical += 1

    print(f"=== 验证: {basename} ===")

    with open(task_file, "r", encoding="utf-8") as f:
        content = f.read()
    lines = content.splitlines()
    content_lower = content.lower()

    # 1. Frontmatter 检查
    print("--- Frontmatter ---")
    if not lines or not lines[0].strip().startswith("---"):
        add_issue("❌", "CRITICAL: 缺少 YAML frontmatter", 20)
    else:
        # 提取 frontmatter 行（两个 --- 之间）
        fm_lines = []
        fm_started = False
        for line in lines:
            stripped = line.strip()
            if stripped == "---":
                if not fm_started:
                    fm_started = True
                    continue
                else:
                    break
            if fm_started:
                fm_lines.append(stripped)

        for field in ["name:", "status:", "created:", "updated:"]:
            if not any(fl.startswith(field) for fl in fm_lines):
                add_issue("⚠️", f"Frontmatter 缺少字段: {field}", 5)

        # 日期格式检查
        for date_field in ["created", "updated"]:
            for fl in fm_lines:
                if fl.startswith(f"{date_field}:"):
                    date_val = fl.split(":", 1)[1].strip()
                    if date_val and not re.match(r'^\d{4}-\d{2}-\d{2}', date_val):
                        add_issue("⚠️", f"{date_field} 不是有效的 ISO 8601 格式", 3)
                    break

    # 2. Task Metadata 检查
    print("--- Task Metadata ---")
    meta_fields = {"File:": 5, "Purpose:": 5, "Leverage:": 5, "Requirements:": 5, "Prompt:": 10}
    for meta_field, deduct in meta_fields.items():
        field_lower = meta_field.lower()
        # File: 也匹配 Files: 和 **File:** / **Files:**
        if field_lower == "file:":
            found = any(p in content_lower for p in ["file:", "files:"])
        else:
            found = field_lower in content_lower
        if not found:
            severity = "❌" if meta_field == "Prompt:" else "⚠️"
            label = "CRITICAL: 缺少" if meta_field == "Prompt:" else "缺少 metadata 字段:"
            add_issue(severity, f"{label} {meta_field}", deduct)

    # Prompt 格式检查
    if "prompt:" in content_lower:
        for prompt_part in ["Role:", "Task:", "Restrictions:", "Success:"]:
            if prompt_part.lower() not in content_lower:
                add_issue("⚠️", f"Prompt 缺少部分: {prompt_part}", 3)

    # 3. Technical Details 检查
    print("--- Technical Details ---")
    tech_patterns = ["technical details", "技术细节", "api specification",
                     "component structure", "data model", "database schema"]
    has_tech = any(p in content_lower for p in tech_patterns)
    if not has_tech:
        add_issue("❌", "CRITICAL: 缺少 Technical Details section", 15)

    # API 任务检查
    api_keywords = ["api", "endpoint", "post", "get", "put", "delete"]
    if any(kw in content_lower for kw in api_keywords):
        spec_keywords = ["request", "response", "endpoint"]
        if not any(kw in content_lower for kw in spec_keywords):
            add_issue("❌", "CRITICAL: API 任务缺少 API 规格说明", 15)

    # 4. Self-Containment 检查
    print("--- Self-Containment ---")
    ext_ref_pattern = re.compile(r'(see|refer|check|参见|详见).*(epic\.md|PRD)', re.IGNORECASE)
    if ext_ref_pattern.search(content):
        add_issue("❌", "CRITICAL: 包含外部引用", 15)

    tbd_pattern = re.compile(r'\bTBD\b|TODO|to be determined|待定|待补充', re.IGNORECASE)
    # 排除 DoD/checklist 中的描述性文本（如 "无 TODO/FIXME 注释残留"）
    tbd_count = 0
    for line in lines:
        if tbd_pattern.search(line):
            line_stripped = line.strip()
            # 跳过 checklist 描述（"- [ ] 无 TODO" 等）
            if re.match(r'^-\s*\[[ x]\].*无\s*(TODO|FIXME|TBD)', line_stripped, re.IGNORECASE):
                continue
            # 跳过注释中的 "no TODO" 描述
            if re.search(r'(no|无|不含|不留|禁止).{0,5}(TODO|FIXME|TBD)', line_stripped, re.IGNORECASE):
                continue
            tbd_count += 1
    if tbd_count > 0:
        add_issue("❌", f"CRITICAL: 包含 {tbd_count} 处 TBD/TODO 占位符", 10)

    # 5. Acceptance Criteria 检查
    print("--- Acceptance Criteria ---")
    ac_patterns = ["acceptance criteria", "验收标准", "验收条件"]
    if not any(p in content_lower for p in ac_patterns):
        add_issue("⚠️", "缺少 Acceptance Criteria section", 5)

    # 6. Definition of Done 检查
    dod_patterns = ["definition of done", "完成定义", "dod"]
    if not any(p in content_lower for p in dod_patterns):
        add_issue("⚠️", "缺少 Definition of Done", 3)

    # 7. Effort Estimate 检查
    print("--- Effort Estimate ---")
    effort_pattern = re.compile(r'Size:|Hours:|Risk:|工作量|估算', re.IGNORECASE)
    if not effort_pattern.search(content):
        add_issue("⚠️", "缺少 Effort Estimate (Size/Hours/Risk)", 3)

    # 确保分数不低于 0
    score = max(score, 0)

    # 强制失败：有 critical 问题且分数 >=60 时降到 59
    if critical > 0 and score >= 60:
        score = 59

    # 输出结果
    print()
    print(f"TASK={basename}")
    print(f"SCORE={score}")
    print(f"CRITICAL_COUNT={critical}")

    if score >= 90:
        print("STATUS=✅ Excellent")
    elif score >= 75:
        print("STATUS=⚠️ Acceptable")
    elif score >= 60:
        print("STATUS=⚠️ Needs Work")
    else:
        print("STATUS=❌ Failed")

    if issues:
        print()
        print("Issues:")
        for issue in issues:
            print(issue)

if __name__ == "__main__":
    main()
