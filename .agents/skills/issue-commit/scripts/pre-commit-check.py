#!/usr/bin/env python3
# 提交前预检：自动发现 issue 号、检查变更、验证分支、查找任务文件
# Usage: python3 pre-commit-check.py [issue_number]

import subprocess
import sys
import os
import re
import glob


def run(cmd, check=False):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    if check and result.returncode != 0:
        return ""
    return result.stdout.strip()


def main():
    issue_number = sys.argv[1] if len(sys.argv) > 1 else ""

    print("=== 提交前预检 ===")

    # 1. 自动发现 issue 号（如果未提供）
    if not issue_number:
        print("--- 自动发现 Issue 号 ---")

        # Method 1: 从分支名提取
        current_branch = run("git branch --show-current", check=True)
        if current_branch:
            match = re.search(r'\d+', current_branch)
            if match:
                issue_number = match.group()
                print(f"从分支名提取: {current_branch} → Issue #{issue_number}")

        # Method 2: 从最近提交消息提取
        if not issue_number:
            log_output = run("git log -5 --pretty=format:'%s'", check=True)
            if log_output:
                match = re.search(r'#(\d+)', log_output)
                if match:
                    issue_number = match.group(1)
                    print(f"从提交消息提取: Issue #{issue_number}")

        if not issue_number:
            print("AUTO_DISCOVER=false")
            print("⚠️  无法自动发现 issue 号，请手动提供")
            sys.exit(1)

    print(f"ISSUE_NUMBER={issue_number}")

    # 2. 验证 issue 号格式
    if not re.match(r'^\d+$', issue_number):
        print(f"❌ Issue 号格式无效: {issue_number}（必须为纯数字）")
        sys.exit(1)

    # 3. 验证 issue 在 GitHub 上存在
    issue_title = run(f"gh issue view {issue_number} --json title --jq '.title'", check=True)
    if not issue_title:
        print(f"❌ 无法访问 issue #{issue_number}。请检查编号或运行: gh auth login")
        sys.exit(1)
    print(f"ISSUE_TITLE={issue_title}")

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
        print(f"请创建功能分支: git checkout -b feature/{issue_number}")
        sys.exit(1)
    print(f"BRANCH={current_branch}")

    # 6. 查找本地任务文件（可选）
    task_file = ""
    epic_name = ""

    # 搜索 .claude/epics 目录
    pattern = f".claude/epics/**/{issue_number}*.md"
    matches = glob.glob(pattern, recursive=True)
    if matches:
        task_file = matches[0]

    if not task_file:
        # 尝试 grep 搜索
        grep_result = run(f'grep -r "github:.*issues/{issue_number}" .claude/epics --include="*.md" -l', check=True)
        if grep_result:
            task_file = grep_result.splitlines()[0]

    if task_file:
        # 提取 epic 名称
        match = re.search(r'\.claude/epics/([^/]+)/', task_file)
        if match:
            epic_name = match.group(1)
        print(f"TASK_FILE={task_file}")
        print(f"EPIC_NAME={epic_name}")
    else:
        print("TASK_FILE=")
        print("EPIC_NAME=")
        print("⚠️  未找到本地任务文件（将跳过任务状态更新）")

    print()
    print("✅ 预检通过")


if __name__ == "__main__":
    main()
