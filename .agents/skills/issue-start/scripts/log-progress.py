#!/usr/bin/env python3
# 向 progress.md 追加一条带时间戳的日志
# Usage: python log-progress.py <issue_number> <epic_name> <message>

import sys
import os
from datetime import datetime, timezone


def main():
    if len(sys.argv) < 4:
        print("ERROR: 必须提供 issue 编号、epic 名称和日志消息", file=sys.stderr)
        print("Usage: python log-progress.py <issue_number> <epic_name> <message>", file=sys.stderr)
        sys.exit(1)

    issue_number = sys.argv[1]
    epic_name = sys.argv[2]
    message = sys.argv[3]

    progress_file = f".claude/epics/{epic_name}/issues/{issue_number}/progress.md"
    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    if not os.path.isfile(progress_file):
        print(f"ERROR: progress.md 不存在: {progress_file}", file=sys.stderr)
        sys.exit(1)

    with open(progress_file, "a", encoding="utf-8") as f:
        f.write(f"- [{timestamp}] {message}\n")

    print(f"✅ 已追加日志到 {progress_file}")


if __name__ == "__main__":
    main()
