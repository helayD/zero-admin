#!/usr/bin/env python3
# 任务存在性检查：按编号和标题匹配现有任务文件
# Usage: python3 compare-tasks.py <feature_name> <task_number> [task_title]
# 输出匹配结果和建议操作

import sys
import os
import re
import glob


def get_status(filepath):
    try:
        with open(filepath, "r", encoding="utf-8") as f:
            content = f.read()
        m = re.search(r'^status:\s*(.+)', content, re.MULTILINE)
        return m.group(1).strip() if m else "unknown"
    except Exception:
        return "unknown"


def get_title(filepath):
    try:
        with open(filepath, "r", encoding="utf-8") as f:
            content = f.read()
        m = re.search(r'^title:\s*(.+)', content, re.MULTILINE)
        if m:
            return m.group(1).strip()
    except Exception:
        pass
    # 从文件名提取
    basename = os.path.basename(filepath).replace(".md", "")
    return re.sub(r'^\d+-', '', basename)


def tokenize(text):
    """将文本分词为小写关键词列表"""
    return [w for w in re.findall(r'[a-zA-Z\u4e00-\u9fff]+', text.lower()) if len(w) >= 3]


def main():
    if len(sys.argv) < 3:
        print("ERROR: 必须提供 feature_name 和 task_number")
        print("Usage: python3 compare-tasks.py <feature_name> <task_number> [task_title]")
        sys.exit(1)

    feature_name = sys.argv[1]
    task_number = sys.argv[2]
    task_title = sys.argv[3] if len(sys.argv) > 3 else ""

    epic_dir = f".claude/epics/{feature_name}"

    print(f"=== 任务存在性检查: {task_number} ===")

    # 1. 编号匹配
    number_matches = glob.glob(f"{epic_dir}/{task_number}*.md") + glob.glob(f"{epic_dir}/{task_number}-*.md")
    # 去重
    number_matches = list(set(number_matches))

    if number_matches:
        number_match = sorted(number_matches)[0]
        print("MATCH_TYPE=number")
        print(f"MATCHED_FILE={number_match}")

        status = get_status(number_match)
        print(f"STATUS={status}")

        if status in ("completed", "in_progress", "in-progress"):
            print("ACTION=KEEP")
            print(f"REASON=status is {status}")
        else:
            print("ACTION=UPDATE_CANDIDATE")
            print("REASON=number match found, needs deep comparison")
        return

    # 2. 标题匹配（简单关键词匹配）
    if task_title:
        title_words = tokenize(task_title)
        best_match = ""
        best_match_count = 0

        task_files = glob.glob(f"{epic_dir}/[0-9]*.md")
        for f in task_files:
            existing_title = get_title(f)
            existing_words = tokenize(existing_title)

            match_count = sum(1 for w in title_words if len(w) >= 3 and any(w in ew for ew in existing_words))

            if match_count > best_match_count:
                best_match_count = match_count
                best_match = f

        total_words = sum(1 for w in title_words if len(w) >= 3)

        if total_words > 0 and best_match_count > 0:
            similarity_pct = best_match_count * 100 // total_words

            if similarity_pct >= 70:
                print("MATCH_TYPE=title")
                print(f"MATCHED_FILE={best_match}")
                print(f"SIMILARITY={similarity_pct}%")

                status = get_status(best_match)
                print(f"STATUS={status}")

                if status in ("completed", "in_progress", "in-progress"):
                    print("ACTION=KEEP")
                elif similarity_pct >= 85:
                    print("ACTION=KEEP")
                else:
                    print("ACTION=UPDATE_CANDIDATE")
                return

    # 3. 内容匹配（基于功能关键词的 Jaccard 近似）
    if task_title:
        task_keywords = set(tokenize(task_title))

        best_content_match = ""
        best_content_score = 0

        task_files = glob.glob(f"{epic_dir}/[0-9]*.md")
        for f in task_files:
            try:
                with open(f, "r", encoding="utf-8") as fh:
                    # 读取前 50 行
                    lines = []
                    for i, line in enumerate(fh):
                        if i >= 50:
                            break
                        lines.append(line)
                    file_text = "".join(lines)
            except Exception:
                continue

            file_keywords = set(tokenize(file_text))
            if not file_keywords:
                continue

            matched = len(task_keywords & file_keywords)
            max_total = max(len(task_keywords), len(file_keywords))

            if max_total > 0:
                content_score = matched * 100 // max_total
                if content_score > best_content_score:
                    best_content_score = content_score
                    best_content_match = f

        if best_content_score >= 75 and best_content_match:
            print("MATCH_TYPE=content")
            print(f"MATCHED_FILE={best_content_match}")
            print(f"SIMILARITY={best_content_score}%")

            status = get_status(best_content_match)
            print(f"STATUS={status}")

            if status in ("completed", "in_progress", "in-progress"):
                print("ACTION=KEEP")
            elif best_content_score >= 85:
                print("ACTION=KEEP")
            else:
                print("ACTION=UPDATE_CANDIDATE")
            return

    # 4. 无匹配
    print("MATCH_TYPE=none")
    print("ACTION=CREATE_CANDIDATE")
    print("REASON=no matching file found")


if __name__ == "__main__":
    main()
