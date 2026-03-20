#!/usr/bin/env python3
# 自动推送到远程仓库
# Usage: python3 auto-push.py

import subprocess
import sys


def run(cmd, check=False):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    if check and result.returncode != 0:
        return ""
    return result.stdout.strip()


def main():
    print("=== 推送到远程 ===")

    current_branch = run("git branch --show-current", check=True)
    if not current_branch:
        print("❌ 无法获取当前分支名")
        sys.exit(1)

    # 检查远程分支是否存在
    remote_check = run(f"git ls-remote --heads origin {current_branch}", check=True)
    remote_exists = len(remote_check.splitlines()) > 0 if remote_check else False

    if not remote_exists:
        print("首次推送，设置上游分支...")
        result = subprocess.run(f"git push -u origin {current_branch}", shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            print(f"✅ 已推送并设置上游: origin/{current_branch}")
        else:
            print("❌ 推送失败")
            print("  认证问题: gh auth login")
            print("  网络问题: 检查网络连接")
            sys.exit(1)
    else:
        result = subprocess.run(f"git push origin {current_branch}", shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            print(f"✅ 已推送到: origin/{current_branch}")
        else:
            print("❌ 推送被拒绝")
            print(f"  远程有新更改: git pull --rebase origin {current_branch}")
            print("  认证问题: gh auth login")
            sys.exit(1)

    commit_hash = run("git rev-parse --short HEAD")
    print()
    print(f"BRANCH={current_branch}")
    print(f"REMOTE=origin/{current_branch}")
    print(f"COMMIT={commit_hash}")


if __name__ == "__main__":
    main()
