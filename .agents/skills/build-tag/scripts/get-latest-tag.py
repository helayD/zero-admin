#!/usr/bin/env python3
# 获取远程最新的 tag 版本号
# Usage: python3 get-latest-tag.py [prefix]
#   prefix: 可选的 tag 前缀，如 "v"（默认为 "v"）

import subprocess
import sys
import re


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout.strip()


def main():
    prefix = sys.argv[1] if len(sys.argv) > 1 else "v"

    print("=== 获取远程最新 Tag ===")

    # 从远程获取所有 tags
    fetch_result = subprocess.run("git fetch origin --tags --quiet", shell=True, capture_output=True, text=True)
    if fetch_result.returncode != 0:
        print("❌ 无法获取远程 tags，请检查网络")
        sys.exit(1)

    # 获取所有匹配前缀的 tags，按版本号排序
    latest_tag = run(f'git tag -l "{prefix}*" --sort=-v:refname')
    if latest_tag:
        latest_tag = latest_tag.splitlines()[0]
    else:
        # 尝试不带前缀
        latest_tag = run("git tag -l --sort=-v:refname")
        if latest_tag:
            latest_tag = latest_tag.splitlines()[0]

    if not latest_tag:
        print("⚠️  未找到任何 tag")
        print("LATEST_TAG=")
        print("MAJOR=0")
        print("MINOR=0")
        print("PATCH=0")
        print(f"NEXT_TAG={prefix}0.0.1")
        return

    print(f"最新 Tag: {latest_tag}")

    # 解析版本号（支持 v1.2.3 和 1.2.3 格式）
    version = latest_tag.lstrip(prefix).lstrip("v")

    # 解析 major.minor.patch
    parts = version.split(".")
    major = re.match(r'^\d+', parts[0]).group() if len(parts) > 0 and re.match(r'^\d+', parts[0]) else "0"
    minor = re.match(r'^\d+', parts[1]).group() if len(parts) > 1 and re.match(r'^\d+', parts[1]) else "0"
    patch = re.match(r'^\d+', parts[2]).group() if len(parts) > 2 and re.match(r'^\d+', parts[2]) else "0"

    print()
    print(f"LATEST_TAG={latest_tag}")
    print(f"PREFIX={prefix}")
    print(f"MAJOR={major}")
    print(f"MINOR={minor}")
    print(f"PATCH={patch}")

    # 计算下一个版本（patch +1）
    next_patch = int(patch) + 1
    next_tag = f"{prefix}{major}.{minor}.{next_patch}"
    print(f"NEXT_TAG={next_tag}")


if __name__ == "__main__":
    main()
