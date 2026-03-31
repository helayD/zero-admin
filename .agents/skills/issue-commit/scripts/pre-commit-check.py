#!/usr/bin/env python3
# 提交前预检：自动发现 story ID、检查变更、验证分支、查找任务文件
# Usage: python3 pre-commit-check.py [story_id]
#
# Story ID 格式: X-Y  (如 "1-4", "1-3", "2-1")
# 支持的目录:
#   _opcos/implementation-artifacts/  - 已实现 story
#   _opcos/planning-artifacts/        - 规划中 story

import subprocess
import sys
import os
import re
import glob
import shutil

# 搜索路径配置
SEARCH_PATHS = [
    "_opcos/implementation-artifacts/*.md",
    "_opcos/planning-artifacts/**/*.md",
    "_opcos/epics/**/*.md",
]


def run(cmd, check=False):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=get_repo_root())
    if check and result.returncode != 0:
        return ""
    return result.stdout.strip()


def get_repo_root():
    result = subprocess.run("git rev-parse --show-toplevel", shell=True, capture_output=True, text=True)
    if result.returncode == 0:
        return result.stdout.strip()
    return os.getcwd()


def find_story_by_id(story_id):
    """按 story ID 查找 story 文件，匹配 _opcos/{category}/{story_id}-{title}.md"""
    for pattern in SEARCH_PATHS:
        for full_pattern in glob.glob(pattern, recursive=True):
            basename = os.path.basename(full_pattern)
            # 匹配: 1-4-标题.md 或 1-4_标题.md
            if basename.startswith(f"{story_id}-") or basename.startswith(f"{story_id}_"):
                return full_pattern
    return None


def extract_story_title(story_file):
    """从 story 文件提取标题"""
    try:
        with open(story_file, "r", encoding="utf-8") as f:
            content = f.read()
        # 优先从 frontmatter (YAML) 的 title 字段提取
        match = re.search(r'^title:\s*["\']?(.+?)["\']?\s*$', content, re.MULTILINE)
        if match:
            return match.group(1).strip()
        # 备选：从 H1 标题提取 "# Story X-Y: 标题"
        match = re.search(r'^#\s+Story\s+\d+-\d+:\s+(.+)$', content, re.MULTILINE)
        if match:
            return match.group(1).strip()
        return os.path.splitext(os.path.basename(story_file))[0]
    except Exception:
        return os.path.splitext(os.path.basename(story_file))[0]


def main():
    story_id = sys.argv[1] if len(sys.argv) > 1 else ""

    print("=== 提交前预检 ===")

    # 1. 自动发现 story ID（如果未提供）
    if not story_id:
        print("--- 自动发现 Story ID ---")

        # Method 1: 从分支名提取
        current_branch = run("git branch --show-current", check=True)
        if current_branch:
            # 匹配 1-4, 2-1, 10-3 等格式
            match = re.search(r'(\d+-\d+)', current_branch)
            if match:
                story_id = match.group(1)
                print(f"从分支名提取: {current_branch} → Story {story_id}")

        # Method 2: 从最近提交消息提取
        if not story_id:
            log_output = run("git log -5 --pretty=format:'%s'", check=True)
            if log_output:
                match = re.search(r'[Ss]tory\s+(\d+-\d+)', log_output)
                if match:
                    story_id = match.group(1)
                    print(f"从提交消息提取: Story {story_id}")

        if not story_id:
            print("AUTO_DISCOVER=false")
            print("⚠️  无法自动发现 story ID，请手动提供")
            print("   用法: /issue-commit 1-4")
            sys.exit(1)

    print(f"STORY_ID={story_id}")

    # 2. 验证 story ID 格式 (X-Y，纯数字+连字符)
    if not re.match(r'^\d+-\d+$', story_id):
        print(f"❌ Story ID 格式无效: {story_id}（必须为 X-Y 格式，如 1-4）")
        sys.exit(1)

    # 3. 查找本地 story 文件
    story_file = find_story_by_id(story_id)
    if story_file:
        story_title = extract_story_title(story_file)
        print(f"STORY_TITLE={story_title}")
        print(f"STORY_FILE={story_file}")
        # 提取 epic 名称
        parts = story_file.split(os.sep)
        if len(parts) >= 3:
            epic_name = parts[1]
            print(f"EPIC_NAME={epic_name}")
    else:
        print("STORY_TITLE=")
        print("STORY_FILE=")
        print("EPIC_NAME=")
        print("⚠️  未找到本地 story 文件（跳过任务状态更新）")

    # 4. 检查是否有未提交的更改
    changes_output = run("git status --porcelain")
    changes = len([l for l in changes_output.splitlines() if l.strip()]) if changes_output else 0
    if changes == 0:
        print("❌ 没有可提交的更改，工作区干净。")
        sys.exit(1)
    print(f"CHANGES_COUNT={changes}")

    # 5. 验证不在 main/master 分支
    current_branch = run("git branch --show-current", check=True)
    if current_branch in ("main", "master"):
        print(f"❌ 不能直接提交到 {current_branch} 分支")
        print(f"请创建功能分支: git checkout -b story/{story_id}")
        sys.exit(1)
    print(f"BRANCH={current_branch}")

    print()
    print("✅ 预检通过")


if __name__ == "__main__":
    main()
