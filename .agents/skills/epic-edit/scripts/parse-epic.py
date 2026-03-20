#!/usr/bin/env python3
# 解析 Epic frontmatter 和内容段落，输出结构化信息
# Usage: python3 parse-epic.py <epic_name>

import sys
import os
import re


def parse_frontmatter(content):
    """解析 YAML frontmatter"""
    fm = {}
    match = re.match(r'^---\s*\n(.*?)\n---\s*\n', content, re.DOTALL)
    if not match:
        return fm, content
    fm_text = match.group(1)
    body = content[match.end():]
    for line in fm_text.splitlines():
        if ':' in line:
            key, _, value = line.partition(':')
            fm[key.strip()] = value.strip()
    return fm, body


def extract_sections(body):
    """提取 Markdown 二级标题段落"""
    sections = []
    current_title = None
    current_lines = []
    for line in body.splitlines():
        match = re.match(r'^##\s+(.+)', line)
        if match:
            if current_title is not None:
                sections.append((current_title, '\n'.join(current_lines).strip()))
            current_title = match.group(1).strip()
            current_lines = []
        else:
            current_lines.append(line)
    if current_title is not None:
        sections.append((current_title, '\n'.join(current_lines).strip()))
    return sections


def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供 epic_name")
        print("Usage: python3 parse-epic.py <epic_name>")
        sys.exit(1)

    epic_name = sys.argv[1]
    epic_file = f".claude/epics/{epic_name}/epic.md"

    if not os.path.isfile(epic_file):
        print(f"ERROR: Epic 文件不存在: {epic_file}")
        sys.exit(1)

    with open(epic_file, "r", encoding="utf-8") as f:
        content = f.read()

    fm, body = parse_frontmatter(content)
    sections = extract_sections(body)

    print(f"=== Epic 解析: {epic_name} ===")
    print(f"EPIC_FILE={epic_file}")
    print()

    # Frontmatter
    print("--- Frontmatter ---")
    for k, v in fm.items():
        print(f"  {k}: {v}")
    print()

    # Sections
    print(f"--- 内容段落 ({len(sections)} 个) ---")
    for i, (title, content_text) in enumerate(sections, 1):
        line_count = len(content_text.splitlines()) if content_text else 0
        print(f"  [{i}] {title} ({line_count} 行)")
    print()

    # 可编辑项提示
    print("--- 可编辑项 ---")
    editable = [
        "Name/Title（名称/标题）",
        "Description/Overview（描述/概述）",
        "Architecture decisions（架构决策）",
        "Technical approach（技术方案）",
        "Dependencies（依赖关系）",
        "Success criteria（验收标准）",
    ]
    for item in editable:
        print(f"  - {item}")

    # GitHub 状态
    print()
    if "github" in fm:
        print(f"GITHUB_LINKED=true")
        print(f"GITHUB_URL={fm['github']}")
    else:
        print("GITHUB_LINKED=false")

    print()
    print("✅ 解析完成")


if __name__ == "__main__":
    main()
