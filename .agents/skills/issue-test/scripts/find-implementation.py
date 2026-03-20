#!/usr/bin/env python3
# 查找实现文件脚本
# 用于分析 issue 相关的实现代码
# 使用方法: python3 find-implementation.py [search_path]

import subprocess
import sys


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout.strip()


def main():
    search_path = sys.argv[1] if len(sys.argv) > 1 else "."

    print("=========================================")
    print("  查找实现文件")
    print("=========================================")
    print()

    # 查找 Controller/Service/Repository 文件
    print("📁 Controller/Service/Repository 文件:")
    output = run(
        f'find "{search_path}" -type f \\( -name "*.dart" -o -name "*.java" -o -name "*.kt" '
        f'-o -name "*.ts" -o -name "*.py" -o -name "*.go" \\) 2>/dev/null | '
        f'grep -E "(controller|Controller|service|Service|repository|Repository|provider|Provider)" | '
        f'head -20'
    )
    if output:
        print(output)

    print()
    print("📁 路由定义文件:")
    output = run(
        f'find "{search_path}" -type f \\( -name "*route*" -o -name "*router*" '
        f'-o -name "*Route*" -o -name "*Router*" \\) 2>/dev/null | head -10'
    )
    if output:
        print(output)

    print()
    print("📁 API 端点文件:")
    output = run(
        f'find "{search_path}" -type f \\( -name "*api*" -o -name "*Api*" '
        f'-o -name "*API*" -o -name "*endpoint*" \\) 2>/dev/null | '
        f'grep -v -E "(node_modules|build|dist|\\.git)" | head -10'
    )
    if output:
        print(output)

    print()
    print("=========================================")
    print("  分析完成")
    print("=========================================")


if __name__ == "__main__":
    main()
