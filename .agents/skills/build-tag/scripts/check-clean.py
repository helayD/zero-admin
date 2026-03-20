#!/usr/bin/env python3
# 检查工作区是否干净且所有代码已推送到远程
# Usage: python3 check-clean.py

import subprocess
import sys


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout.strip()


def count_lines(output):
    return len([l for l in output.splitlines() if l.strip()]) if output else 0


def main():
    print("=== 检查代码提交状态 ===")

    # 1. 检查是否有未提交的更改
    unstaged = count_lines(run("git diff --name-only"))
    staged = count_lines(run("git diff --cached --name-only"))
    untracked = count_lines(run("git ls-files --others --exclude-standard"))

    if unstaged > 0 or staged > 0 or untracked > 0:
        print("❌ 工作区不干净")
        if unstaged > 0:
            print(f"  未暂存的更改: {unstaged} 个文件")
        if staged > 0:
            print(f"  已暂存未提交: {staged} 个文件")
        if untracked > 0:
            print(f"  未跟踪的文件: {untracked} 个文件")
        print()
        print("请先提交或暂存所有更改:")
        print("  git add -A && git commit -m 'your message'")
        sys.exit(1)

    print("✅ 工作区干净，无未提交更改")

    # 2. 检查当前分支是否已推送到远程
    current_branch = run("git branch --show-current")
    print()
    print(f"当前分支: {current_branch}")

    # 获取远程最新状态
    fetch_result = subprocess.run("git fetch origin --quiet", shell=True, capture_output=True, text=True)
    if fetch_result.returncode != 0:
        print("❌ 无法连接远程仓库，请检查网络或运行: gh auth login")
        sys.exit(1)

    # 比较本地和远程
    local_hash = run("git rev-parse HEAD")
    remote_hash = run(f"git rev-parse origin/{current_branch}")

    if not remote_hash:
        print(f"❌ 远程分支 origin/{current_branch} 不存在")
        print(f"请先推送: git push -u origin {current_branch}")
        sys.exit(1)

    if local_hash != remote_hash:
        ahead = run(f"git rev-list --count origin/{current_branch}..HEAD") or "0"
        behind = run(f"git rev-list --count HEAD..origin/{current_branch}") or "0"
        print("❌ 本地与远程不同步")
        if int(ahead) > 0:
            print(f"  本地领先远程 {ahead} 个提交")
        if int(behind) > 0:
            print(f"  本地落后远程 {behind} 个提交")
        print()
        print("请先同步:")
        if int(ahead) > 0:
            print(f"  git push origin {current_branch}")
        if int(behind) > 0:
            print(f"  git pull origin {current_branch}")
        sys.exit(1)

    print(f"✅ 所有代码已推送到远程 (origin/{current_branch})")
    print()
    print("CLEAN=true")
    print(f"BRANCH={current_branch}")
    print(f"COMMIT={local_hash}")


if __name__ == "__main__":
    main()
