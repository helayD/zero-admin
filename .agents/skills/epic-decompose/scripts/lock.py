#!/usr/bin/env python3
# 并发锁管理：获取/释放 decompose 锁，防止并发修改
# Usage: python3 lock.py <action> <feature_name>
#   action: acquire | release | check
#   feature_name: Epic 目录名

import sys
import os
import time


def main():
    if len(sys.argv) < 3:
        print("ERROR: 必须提供 action (acquire|release|check) 和 feature_name")
        sys.exit(1)

    action = sys.argv[1]
    feature_name = sys.argv[2]

    lock_file = f".claude/epics/{feature_name}/.decompose.lock"
    stale_seconds = 300  # 5 分钟视为过期

    if action == "acquire":
        # 检查并清理过期锁
        if os.path.isfile(lock_file):
            try:
                lock_time = int(os.path.getmtime(lock_file))
            except Exception:
                lock_time = 0
            now = int(time.time())
            age = now - lock_time
            if age > stale_seconds:
                print(f"⚠️  清理过期锁 (age: {age}s)")
                os.remove(lock_file)
            else:
                try:
                    with open(lock_file, "r") as f:
                        lock_pid = f.read().strip()
                except Exception:
                    lock_pid = "unknown"
                print(f"ERROR: 检测到正在进行的 decompose 操作 (PID: {lock_pid})")
                print(f"Lock file: {lock_file}")
                print("请等待 10 秒或稍后重试")
                sys.exit(1)

        # 获取锁
        os.makedirs(os.path.dirname(lock_file), exist_ok=True)
        with open(lock_file, "w") as f:
            f.write(str(os.getpid()))
        print(f"✅ 已获取锁: {lock_file} (PID: {os.getpid()})")

    elif action == "release":
        if os.path.isfile(lock_file):
            os.remove(lock_file)
            print(f"✅ 已释放锁: {lock_file}")
        else:
            print(f"⚠️  锁文件不存在: {lock_file}")

    elif action == "check":
        if os.path.isfile(lock_file):
            try:
                with open(lock_file, "r") as f:
                    lock_pid = f.read().strip()
            except Exception:
                lock_pid = "unknown"
            try:
                lock_time = int(os.path.getmtime(lock_file))
            except Exception:
                lock_time = 0
            now = int(time.time())
            age = now - lock_time
            print("LOCKED=true")
            print(f"LOCK_PID={lock_pid}")
            print(f"LOCK_AGE={age}s")
            print(f"LOCK_STALE={'true' if age > stale_seconds else 'false'}")
        else:
            print("LOCKED=false")

    else:
        print(f"ERROR: 未知操作: {action} (支持: acquire|release|check)")
        sys.exit(1)


if __name__ == "__main__":
    main()
