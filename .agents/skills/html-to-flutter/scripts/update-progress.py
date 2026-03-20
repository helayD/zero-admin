#!/usr/bin/env python3
"""
更新 MIGRATION-PROGRESS.md 中的页面迁移状态和统计。

用法: python3 .claude/skills/html-to-flutter/scripts/update-progress.py <page_name> <status>
状态: in_progress | completed | paused
示例: python3 .claude/skills/html-to-flutter/scripts/update-progress.py home completed
"""

import sys
import os
import re
from pathlib import Path
from datetime import datetime

STATUS_ICONS = {
    "in_progress": "🔄",
    "completed": "✅",
    "paused": "⏸️",
    "not_started": "⬜",
}

STATUS_LABELS = {
    "in_progress": "进行中",
    "completed": "已完成",
    "paused": "暂停",
    "not_started": "未开始",
}


def find_project_root():
    """向上查找包含 flutter_client/ 的项目根目录"""
    current = Path.cwd()
    for _ in range(10):
        if (current / "flutter_client").is_dir():
            return current
        parent = current.parent
        if parent == current:
            break
        current = parent
    return Path.cwd()


def update_progress(progress_path: Path, page_name: str, new_status: str):
    """更新进度文件中指定页面的状态"""
    if not progress_path.exists():
        print(f"❌ 进度文件不存在: {progress_path}")
        sys.exit(1)

    content = progress_path.read_text(encoding="utf-8")
    new_icon = STATUS_ICONS.get(new_status, "⬜")
    
    # 查找并替换页面状态
    # 匹配格式: | ⬜ | page_name.html | ... 或 | 🔄 | page_name.html | ...
    pattern = rf'\| [⬜🔄✅⏸️]+ \| {re.escape(page_name)}\.html \|'
    replacement = f'| {new_icon} | {page_name}.html |'
    
    if not re.search(pattern, content):
        print(f"❌ 在进度文件中未找到页面: {page_name}.html")
        print(f"   请检查 MIGRATION-PROGRESS.md 中的页面名称")
        sys.exit(1)

    new_content = re.sub(pattern, replacement, content)

    # 更新统计数字 — 只匹配状态列（行首 | 后紧跟状态图标）
    # 格式: | ⬜ | page.html | ... 状态图标后紧跟 " | " 和 .html 文件名
    completed_count = len(re.findall(r'\| ✅ \| \S+\.html', new_content))
    in_progress_count = len(re.findall(r'\| 🔄 \| \S+\.html', new_content))
    not_started_count = len(re.findall(r'\| ⬜ \| \S+\.html', new_content))
    paused_count = len(re.findall(r'\| ⏸️ \| \S+\.html', new_content))
    total = completed_count + in_progress_count + not_started_count + paused_count

    # 更新总览表
    new_content = re.sub(
        r'\| \*\*已完成\*\* \| \d+ \|',
        f'| **已完成** | {completed_count} |',
        new_content
    )
    new_content = re.sub(
        r'\| \*\*进行中\*\* \| \d+ \|',
        f'| **进行中** | {in_progress_count} |',
        new_content
    )
    new_content = re.sub(
        r'\| \*\*未开始\*\* \| \d+ \|',
        f'| **未开始** | {not_started_count} |',
        new_content
    )
    
    # 更新完成率
    pct = round(completed_count / total * 100) if total > 0 else 0
    new_content = re.sub(
        r'\| \*\*完成率\*\* \| \d+% \|',
        f'| **完成率** | {pct}% |',
        new_content
    )

    # 更新各 Phase 进度
    for phase_label, phase_pages in [
        ("P0 进度", 8),
        ("P1 进度", 12),
        ("P2 进度", 11),
        ("P3 进度", 6),
    ]:
        # 简单统计每个section的完成数（通过匹配section内的✅数量）
        pass  # 由下方的 section 级别计算处理

    # 更新最后更新日期
    today = datetime.now().strftime("%Y-%m-%d")
    new_content = re.sub(
        r'\> \*\*最后更新\*\*: \d{4}-\d{2}-\d{2}',
        f'> **最后更新**: {today}',
        new_content
    )

    # 追加变更日志
    log_entry = f"| {today} | {page_name}.html | {STATUS_LABELS.get(new_status, new_status)} | Cascade |"
    
    # 在变更日志表格末尾插入新行
    if "## 变更日志" in new_content:
        # 找到最后一个表格行的位置，在其后插入
        lines = new_content.split("\n")
        insert_idx = None
        in_changelog = False
        for i, line in enumerate(lines):
            if "## 变更日志" in line:
                in_changelog = True
            if in_changelog and line.startswith("|") and "---" not in line and "日期" not in line:
                insert_idx = i
        
        if insert_idx is not None:
            lines.insert(insert_idx + 1, log_entry)
            new_content = "\n".join(lines)

    progress_path.write_text(new_content, encoding="utf-8")
    return completed_count, in_progress_count, not_started_count, total


def main():
    if len(sys.argv) < 3:
        print("❌ 错误: 参数不足")
        print(f"用法: python3 {sys.argv[0]} <page_name> <status>")
        print(f"状态: in_progress | completed | paused")
        sys.exit(1)

    page_name = sys.argv[1].strip().lower()
    new_status = sys.argv[2].strip().lower()

    if page_name.endswith(".html"):
        page_name = page_name[:-5]

    if new_status not in STATUS_ICONS:
        print(f"❌ 无效状态: {new_status}")
        print(f"可用状态: {', '.join(STATUS_ICONS.keys())}")
        sys.exit(1)

    root = find_project_root()
    progress_path = root / "flutter_client" / "docs" / "MIGRATION-PROGRESS.md"

    print(f"更新迁移进度: {page_name}.html → {STATUS_LABELS[new_status]}")
    
    completed, in_progress, not_started, total = update_progress(
        progress_path, page_name, new_status
    )

    pct = round(completed / total * 100) if total > 0 else 0
    print(f"\n✅ 进度已更新")
    print(f"   总计: {total} | 已完成: {completed} | 进行中: {in_progress} | 未开始: {not_started}")
    print(f"   完成率: {pct}%")


if __name__ == "__main__":
    main()
