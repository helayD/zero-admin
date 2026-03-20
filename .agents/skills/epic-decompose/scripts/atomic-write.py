#!/usr/bin/env python3
# 原子写入：通过临时文件确保写入完整性
# Usage: python3 atomic-write.py <target_file>
# 从 stdin 读取内容，原子写入到目标文件

import sys
import os


def main():
    if len(sys.argv) < 2:
        print("ERROR: 必须提供目标文件路径")
        sys.exit(1)

    target_file = sys.argv[1]
    tmp_file = target_file + ".tmp"

    # 确保目录存在
    os.makedirs(os.path.dirname(target_file), exist_ok=True)

    # 从 stdin 读取并写入临时文件
    content = sys.stdin.read()
    with open(tmp_file, "w", encoding="utf-8") as f:
        f.write(content)

    # 验证临时文件非空
    if os.path.getsize(tmp_file) == 0:
        print("ERROR: 临时文件为空，中止写入")
        os.remove(tmp_file)
        sys.exit(1)

    # 原子移动
    os.replace(tmp_file, target_file)
    print(f"✅ 原子写入完成: {target_file}")


if __name__ == "__main__":
    main()
