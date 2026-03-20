#!/usr/bin/env python3
# Step 1: 验证 GitHub CLI、认证、仓库权限、安全检查
# Usage: python3 validate-env.py

import subprocess
import sys


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.returncode, result.stdout.strip(), result.stderr.strip()


def main():
    print("=== 验证 GitHub 环境 ===")

    # 1. Check GitHub CLI installation
    code, out, err = run("command -v gh")
    if code != 0:
        print("❌ GitHub CLI (gh) 未安装")
        print("安装地址: https://cli.github.com/")
        sys.exit(1)
    print("✅ GitHub CLI 已安装")

    # 2. Check GitHub authentication
    code, out, err = run("gh auth status")
    if code != 0:
        print("❌ GitHub 未认证")
        print("请运行: gh auth login")
        sys.exit(1)
    print("✅ GitHub 已认证")

    # 3. Verify repository access
    code, out, err = run("gh repo view --json name --jq '.name'")
    if code != 0:
        print("❌ 无法访问仓库或不在 git 仓库中")
        sys.exit(1)
    print(f"✅ 仓库可访问: {out}")

    # 4. Safety check: Prevent CCPM template sync
    code, remote_url, err = run("git remote get-url origin")
    if code == 0 and ("ccpm" in remote_url.lower() or "automazeio" in remote_url.lower()):
        print("❌ 不允许同步到 CCPM/automazeio 模板仓库")
        sys.exit(1)
    print("✅ 安全检查通过")

    print()
    print("✅ 环境验证完成，可以继续同步")


if __name__ == "__main__":
    main()
