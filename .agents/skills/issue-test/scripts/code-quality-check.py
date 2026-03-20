#!/usr/bin/env python3
# 代码质量检查脚本
# 检查 issue 相关修改中的代码质量问题：
# 1. 代码冗余 (重复代码)
# 2. 硬编码 (魔法数字、硬编码字符串)
# 3. TODO/FIXME 未完成标记
#
# 使用方法: python3 code-quality-check.py <issue_number> [base_branch]

import subprocess
import sys
import os
import re
from datetime import datetime


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout.strip()


# 统计变量
TOTAL_ISSUES = 0
REDUNDANCY_ISSUES = 0
HARDCODE_ISSUES = 0
TODO_ISSUES = 0
CHANGED_FILES = []


def get_changed_files(issue_number, base_branch, report_file):
    global CHANGED_FILES
    print(f"正在获取 Issue #{issue_number} 相关的修改文件...\n")

    # 方法1: 从当前分支与基准分支的差异获取
    files = run(f"git diff --name-only {base_branch}...HEAD")
    if not files:
        files = run("git diff --name-only --cached")
    if not files:
        files = run("git diff --name-only")

    # 过滤只保留代码文件
    code_ext = re.compile(r'\.(java|kt|dart|ts|tsx|js|jsx|py|go|rs|swift|vue|sh)$')
    CHANGED_FILES = [f for f in files.splitlines() if f.strip() and code_ext.search(f)]

    if not CHANGED_FILES:
        print("⚠️  未找到相关的代码修改文件")
        with open(report_file, "a", encoding="utf-8") as rf:
            rf.write("## 检查结果\n\n未找到相关的代码修改文件\n")
        return False

    print(f"找到 {len(CHANGED_FILES)} 个修改的代码文件\n")
    with open(report_file, "a", encoding="utf-8") as rf:
        rf.write(f"## 检查范围\n\n共检查 {len(CHANGED_FILES)} 个文件:\n\n")
        for f in CHANGED_FILES:
            rf.write(f"- `{f}`\n")
        rf.write("\n")
    return True


def check_redundancy(report_file):
    global REDUNDANCY_ISSUES, TOTAL_ISSUES
    print("[1/3] 检查代码冗余...")
    issues_found = 0

    with open(report_file, "a", encoding="utf-8") as rf:
        rf.write("## 1. 代码冗余检查\n\n")

        for filepath in CHANGED_FILES:
            if not os.path.isfile(filepath):
                continue
            try:
                with open(filepath, "r", encoding="utf-8", errors="ignore") as f:
                    lines = f.readlines()
            except Exception:
                continue

            # 检查重复的代码块 (连续3行以上相同)
            seen_blocks = {}
            duplicates = []
            for i in range(2, len(lines)):
                block = lines[i-2] + lines[i-1] + lines[i]
                block_stripped = block.strip()
                if not block_stripped or len(block_stripped) < 20:
                    continue
                if block_stripped in seen_blocks:
                    duplicates.append(f"重复代码块 (行 {i-1}-{i+1}): {lines[i].strip()}")
                seen_blocks[block_stripped] = i

            if duplicates:
                rf.write(f"### `{filepath}`\n\n```\n")
                for d in duplicates[:10]:
                    rf.write(d + "\n")
                rf.write("```\n\n")
                issues_found += 1

        REDUNDANCY_ISSUES = issues_found
        TOTAL_ISSUES += issues_found

        if issues_found == 0:
            print("  ✅ 未发现代码冗余问题")
            rf.write("✅ 未发现代码冗余问题\n")
        else:
            print(f"  ❌ 发现 {issues_found} 处潜在冗余")
        rf.write("\n")


def check_hardcode(report_file):
    global HARDCODE_ISSUES, TOTAL_ISSUES
    print("[2/3] 检查硬编码...")
    issues_found = 0

    with open(report_file, "a", encoding="utf-8") as rf:
        rf.write("## 2. 硬编码检查\n\n")

        for filepath in CHANGED_FILES:
            if not os.path.isfile(filepath):
                continue
            try:
                with open(filepath, "r", encoding="utf-8", errors="ignore") as f:
                    content = f.read()
                    f.seek(0)
                    numbered_lines = list(enumerate(f.readlines(), 1))
            except Exception:
                continue

            file_issues = []

            for lineno, line in numbered_lines:
                stripped = line.strip()
                # 跳过注释行
                if stripped.startswith(("//", "#", "*", "/*")):
                    continue

                # 检查硬编码 URL
                if re.search(r'https?://[a-zA-Z0-9]', line):
                    if not re.search(r'(localhost|127\.0\.0\.1|example\.com|test\.|mock\.)', line):
                        file_issues.append(f"  {lineno}: [URL] {stripped[:100]}")

                # 检查硬编码密钥/密码
                if re.search(r'(password|secret|key|token|api_key|apikey)\s*[:=]\s*["\'][^"\']+["\']', line, re.IGNORECASE):
                    if not re.search(r'(env\.|process\.env|getenv|System\.getenv|@Value)', line):
                        file_issues.append(f"  {lineno}: [SECRET] {stripped[:100]}")

                # 检查硬编码数据库连接
                if re.search(r'(jdbc:|mongodb://|mysql://|postgres://|redis://)', line):
                    if not re.search(r'(localhost|127\.0\.0\.1|\$\{|env\.)', line):
                        file_issues.append(f"  {lineno}: [DB] {stripped[:100]}")

            if file_issues:
                rf.write(f"### `{filepath}`\n\n```\n")
                for issue in file_issues[:10]:
                    rf.write(issue + "\n")
                rf.write("```\n\n")
                issues_found += 1

        HARDCODE_ISSUES = issues_found
        TOTAL_ISSUES += issues_found

        if issues_found == 0:
            print("  ✅ 未发现硬编码问题")
            rf.write("✅ 未发现硬编码问题\n")
        else:
            print(f"  ❌ 发现 {issues_found} 处硬编码")
        rf.write("\n")


def check_todo(report_file):
    global TODO_ISSUES, TOTAL_ISSUES
    print("[3/3] 检查 TODO/FIXME 标记...")
    issues_found = 0

    with open(report_file, "a", encoding="utf-8") as rf:
        rf.write("## 3. TODO/FIXME 检查\n\n")

        for filepath in CHANGED_FILES:
            if not os.path.isfile(filepath):
                continue
            try:
                with open(filepath, "r", encoding="utf-8", errors="ignore") as f:
                    numbered_lines = list(enumerate(f.readlines(), 1))
            except Exception:
                continue

            todos = []
            for lineno, line in numbered_lines:
                if re.search(r'(TODO|FIXME|XXX|HACK|BUG|UNDONE|OPTIMIZE):?', line):
                    todos.append(f"  {lineno}: {line.strip()}")

            if todos:
                rf.write(f"### `{filepath}` ({len(todos)} 处)\n\n```\n")
                for t in todos:
                    rf.write(t + "\n")
                rf.write("```\n\n")
                issues_found += len(todos)

        TODO_ISSUES = issues_found
        TOTAL_ISSUES += issues_found

        if issues_found == 0:
            print("  ✅ 未发现 TODO/FIXME 标记")
            rf.write("✅ 未发现 TODO/FIXME 标记\n")
        else:
            print(f"  ❌ 发现 {issues_found} 处 TODO/FIXME")
        rf.write("\n")


def generate_summary(report_file):
    print("\n=========================================")
    print("   检查完成 - 总结报告")
    print("=========================================\n")

    with open(report_file, "a", encoding="utf-8") as rf:
        rf.write("---\n\n## 总结\n\n")
        rf.write("| 检查项 | 问题数量 | 状态 |\n")
        rf.write("|--------|----------|------|\n")

        status_r = "✅ 通过" if REDUNDANCY_ISSUES == 0 else "❌ 需修复"
        status_h = "✅ 通过" if HARDCODE_ISSUES == 0 else "❌ 需修复"
        status_t = "✅ 通过" if TODO_ISSUES == 0 else "❌ 需修复"

        rf.write(f"| 代码冗余 | {REDUNDANCY_ISSUES} | {status_r} |\n")
        rf.write(f"| 硬编码 | {HARDCODE_ISSUES} | {status_h} |\n")
        rf.write(f"| TODO/FIXME | {TODO_ISSUES} | {status_t} |\n")
        rf.write("\n")

        print(f"  代码冗余:    {'✅ 0 处' if REDUNDANCY_ISSUES == 0 else f'❌ {REDUNDANCY_ISSUES} 处'}")
        print(f"  硬编码:      {'✅ 0 处' if HARDCODE_ISSUES == 0 else f'❌ {HARDCODE_ISSUES} 处'}")
        print(f"  TODO/FIXME:  {'✅ 0 处' if TODO_ISSUES == 0 else f'❌ {TODO_ISSUES} 处'}")

        if TOTAL_ISSUES == 0:
            rf.write("### ✅ 所有检查通过！代码质量良好\n")
            print("\n✅ 所有检查通过！代码质量良好")
        else:
            rf.write(f"### ❌ 发现 {TOTAL_ISSUES} 处问题需要修复\n\n请在提交前修复以上问题。\n")
            print(f"\n❌ 发现 {TOTAL_ISSUES} 处问题需要修复")

        rf.write(f"\n---\n*报告生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}*\n")

    print(f"\n报告已保存到: {report_file}")


def main():
    issue_number = sys.argv[1] if len(sys.argv) > 1 else ""
    base_branch = sys.argv[2] if len(sys.argv) > 2 else "main"

    if not issue_number:
        print("错误: 请提供 issue 编号")
        print("使用方法: python3 code-quality-check.py <issue_number> [base_branch]")
        sys.exit(1)

    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    log_dir = "./quality_reports"
    os.makedirs(log_dir, exist_ok=True)
    report_file = f"{log_dir}/quality_report_{issue_number}_{timestamp}.md"

    print("=========================================")
    print(f"   代码质量检查 - Issue #{issue_number}")
    print("=========================================\n")

    # 初始化报告
    with open(report_file, "w", encoding="utf-8") as rf:
        rf.write(f"# 代码质量检查报告\n\n")
        rf.write(f"- **Issue**: #{issue_number}\n")
        rf.write(f"- **检查时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        rf.write(f"- **基准分支**: {base_branch}\n\n---\n\n")

    if get_changed_files(issue_number, base_branch, report_file):
        check_redundancy(report_file)
        check_hardcode(report_file)
        check_todo(report_file)
        generate_summary(report_file)

    sys.exit(0 if TOTAL_ISSUES == 0 else 1)


if __name__ == "__main__":
    main()
