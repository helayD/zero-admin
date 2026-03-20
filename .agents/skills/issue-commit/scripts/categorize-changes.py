#!/usr/bin/env python3
# 变更文件分类统计
# Usage: python3 categorize-changes.py

import subprocess
import re


def main():
    print("=== 变更文件分类 ===")

    source_count = 0
    test_count = 0
    config_count = 0
    doc_count = 0
    other_count = 0

    result = subprocess.run("git status --porcelain", shell=True, capture_output=True, text=True)
    lines = result.stdout.strip().splitlines() if result.stdout.strip() else []

    for line in lines:
        if not line.strip():
            continue
        file = line[3:]  # 去掉前3个状态字符

        # 测试文件（优先判断）
        if re.search(r'(test|spec)\.', file, re.IGNORECASE) or re.search(r'(/__tests__/|/test/)', file, re.IGNORECASE):
            test_count += 1
        # 源代码文件
        elif re.search(r'\.(ts|tsx|js|jsx|py|go|rs|java|kt|swift|dart|c|cpp|h|rb|php|scala|vue|svelte)$', file, re.IGNORECASE):
            source_count += 1
        # 配置文件
        elif re.search(r'\.(json|yaml|yml|toml|env|ini|cfg|conf|properties|xml)$', file, re.IGNORECASE):
            config_count += 1
        # 文档文件
        elif re.search(r'\.(md|txt|rst|adoc|doc|docx)$', file, re.IGNORECASE):
            doc_count += 1
        else:
            other_count += 1

    total = source_count + test_count + config_count + doc_count + other_count

    print("Changes to commit:")
    print()
    print(f"SOURCE_COUNT={source_count}")
    print(f"TEST_COUNT={test_count}")
    print(f"CONFIG_COUNT={config_count}")
    print(f"DOC_COUNT={doc_count}")
    print(f"OTHER_COUNT={other_count}")
    print(f"TOTAL={total}")
    print()
    print(f"Source files: {source_count}")
    print(f"Test files: {test_count}")
    print(f"Config files: {config_count}")
    print(f"Documentation: {doc_count}")
    print(f"Other files: {other_count}")
    print()
    print(f"Total: {total} files modified")


if __name__ == "__main__":
    main()
