#!/usr/bin/env python3
# 创建新 tag 并推送到远程
# Usage: python3 bump-tag.py <new_tag> [message]

import subprocess
import sys


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.returncode, result.stdout.strip()


def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供新 tag 名称")
        sys.exit(1)

    new_tag = sys.argv[1]
    message = sys.argv[2] if len(sys.argv) > 2 else f"Release {new_tag}"

    print("=== 创建并推送 Tag ===")

    # 检查 tag 是否已存在
    code, output = run(f'git tag -l "{new_tag}"')
    if new_tag in output:
        print(f"❌ Tag {new_tag} 已存在")
        sys.exit(1)

    # 创建 annotated tag
    code, _ = run(f'git tag -a "{new_tag}" -m "{message}"')
    if code != 0:
        print(f"❌ 创建 tag 失败")
        sys.exit(1)
    print(f"✅ 已创建 tag: {new_tag}")
    print(f"   消息: {message}")

    # 推送到远程
    code, _ = run(f'git push origin "{new_tag}"')
    if code != 0:
        print("❌ 推送 tag 失败，回滚...")
        run(f'git tag -d "{new_tag}"')
        print(f"已删除本地 tag: {new_tag}")
        sys.exit(1)

    print(f"✅ 已推送到远程: origin/{new_tag}")

    _, commit = run("git rev-parse HEAD")
    print()
    print(f"TAG={new_tag}")
    print(f"MESSAGE={message}")
    print(f"COMMIT={commit}")


if __name__ == "__main__":
    main()
