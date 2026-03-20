#!/usr/bin/env python3
# 测试报告目录初始化脚本
# 创建版本化的测试报告结构
# 使用方法: python3 setup-test-reports.py <issue_number> <epic_name>

import sys
import os
from datetime import datetime


def main():
    issue_number = sys.argv[1] if len(sys.argv) > 1 else ""
    epic_name = sys.argv[2] if len(sys.argv) > 2 else "default"

    if not issue_number:
        print("错误: 请提供 issue 编号")
        print("使用方法: python3 setup-test-reports.py <issue_number> <epic_name>")
        sys.exit(1)

    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    dt = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    base_dir = f".claude/epics/{epic_name}/issues/{issue_number}"
    reports_dir = f"{base_dir}/reports"
    logs_dir = f"{reports_dir}/test_logs"

    # 创建目录结构
    print("📁 创建测试报告目录结构...")
    os.makedirs(logs_dir, exist_ok=True)

    # 创建测试报告文件
    report_file = f"{reports_dir}/test_report_{timestamp}.md"
    with open(report_file, "w", encoding="utf-8") as f:
        f.write(f"""---
issue: {issue_number}
report_id: {timestamp}
created: {dt}
status: in_progress
---

# Test Report - Issue #{issue_number}

## Environment
- Base URL: TBD
- Database: TBD
- Version: TBD

## Test Execution

### E2E Tests
- Status: Pending
- Passed: 0
- Failed: 0

### API Tests
- Status: Pending
- Total: 0
- Success: 0
- Failed: 0
- Avg Response Time: 0ms

### Database Consistency
- Status: Pending
- Pre-test snapshot: TBD
- Post-test snapshot: TBD

## Results
TBD

## Issues Found
None yet

## Performance Metrics
TBD

---
*Report generated: {dt}*
""")

    # 创建 API 请求日志
    api_log = f"{logs_dir}/api_request_log_{timestamp}.json"
    with open(api_log, "w", encoding="utf-8") as f:
        f.write(f"""{{"issue": "{issue_number}", "report_id": "{timestamp}", "start_time": "{dt}", "api_requests": [], "summary": {{"total": 0, "success": 0, "failed": 0, "avg_response_time_ms": 0}}}}
""")

    # 创建数据库一致性日志
    db_log = f"{logs_dir}/db_consistency_{timestamp}.json"
    with open(db_log, "w", encoding="utf-8") as f:
        f.write(f"""{{"issue": "{issue_number}", "report_id": "{timestamp}", "start_time": "{dt}", "pre_test": {{}}, "post_test": {{}}, "checks": [], "summary": {{"total_checks": 0, "passed": 0, "failed": 0, "status": "pending"}}}}
""")

    # 创建 E2E 日志
    e2e_log = f"{logs_dir}/e2e_{timestamp}.log"
    with open(e2e_log, "w", encoding="utf-8") as f:
        f.write(f"# E2E Test Log - Issue #{issue_number}\n")
        f.write(f"# Started: {dt}\n\n")

    print()
    print("✅ 测试报告结构创建完成")
    print()
    print("📄 文件列表:")
    print(f"  - {report_file}")
    print(f"  - {api_log}")
    print(f"  - {db_log}")
    print(f"  - {e2e_log}")
    print()
    print(f"📋 时间戳: {timestamp}")


if __name__ == "__main__":
    main()
