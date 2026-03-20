#!/usr/bin/env python3
# Step 5: 执行单个任务的同步操作（CREATE/UPDATE/SYNCED/SKIP）
# Usage: python3 sync-task.py <feature_name> <task_file> <action> [--check-only]
#
# Actions: CREATE, UPDATE, SYNCED, SKIP

import subprocess
import sys
import os
import re
import time


def run(cmd):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.returncode, result.stdout.strip(), result.stderr.strip()


def parse_frontmatter(filepath):
    """解析文件的 YAML frontmatter，返回 (fm_dict, body_without_fm)"""
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    fm = {}
    body = content
    match = re.match(r'^---\s*\n(.*?)\n---\s*\n', content, re.DOTALL)
    if match:
        for line in match.group(1).splitlines():
            if ':' in line:
                key, _, value = line.partition(':')
                fm[key.strip()] = value.strip()
        body = content[match.end():]

    return fm, body


def update_frontmatter_field(filepath, key, value):
    """更新文件 frontmatter 中的指定字段"""
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    match = re.match(r'^(---\s*\n)(.*?)(\n---\s*\n)', content, re.DOTALL)
    if not match:
        return False

    fm_text = match.group(2)
    rest = content[match.end():]

    # 检查字段是否存在
    pattern = re.compile(rf'^{re.escape(key)}:.*$', re.MULTILINE)
    if pattern.search(fm_text):
        fm_text = pattern.sub(f'{key}: {value}', fm_text)
    else:
        fm_text += f'\n{key}: {value}'

    new_content = match.group(1) + fm_text + match.group(3) + rest

    # 原子写入
    tmp_file = filepath + ".tmp"
    with open(tmp_file, "w", encoding="utf-8") as f:
        f.write(new_content)
    os.replace(tmp_file, filepath)
    return True


def ensure_filename(feature_name, task_file, issue_id, task_name):
    """确保文件名符合 {issue_id}-{title}.md 格式，返回新路径或原路径"""
    epic_dir = f".claude/epics/{feature_name}"
    current_basename = os.path.basename(task_file)

    # 生成期望的文件名
    # 从当前文件名提取 title 部分
    title_part = re.sub(r'^\d+-', '', current_basename).replace('.md', '')
    expected_basename = f"{issue_id}-{title_part}.md"

    if current_basename == expected_basename:
        return task_file, False

    new_path = os.path.join(epic_dir, expected_basename)
    if os.path.exists(task_file):
        os.rename(task_file, new_path)
        return new_path, True

    return task_file, False


def main():
    if len(sys.argv) < 4:
        print("ERROR: 必须提供 feature_name, task_file, action")
        print("Usage: python3 sync-task.py <feature_name> <task_file> <action> [--check-only]")
        sys.exit(1)

    feature_name = sys.argv[1]
    task_file = sys.argv[2]
    action = sys.argv[3].upper()
    check_only = "--check-only" in sys.argv

    basename = os.path.basename(task_file)
    num_match = re.match(r'^(\d+)', basename)
    task_num = num_match.group(1) if num_match else "000"

    if not os.path.isfile(task_file):
        print(f"ERROR: 任务文件不存在: {task_file}")
        sys.exit(1)

    fm, body = parse_frontmatter(task_file)
    task_name = fm.get("name", fm.get("title", ""))
    if not task_name:
        task_name = re.sub(r'^\d+-', '', basename).replace('.md', '').replace('-', ' ').replace('_', ' ')

    # === SKIP ===
    if action == "SKIP":
        print(f"⏭️  Task #{task_num}: Skipped")
        return

    # === CHECK-ONLY MODE ===
    if check_only:
        if action == "CREATE":
            print(f"Would create Issue for Task #{task_num}: {task_name}")
        elif action == "UPDATE":
            print(f"Would update Issue {fm.get('github', '?')} for Task #{task_num}: {task_name}")
        elif action == "SYNCED":
            print(f"Task #{task_num} already synced")
        return

    # === CREATE ===
    if action == "CREATE":
        print(f"🆕 Creating Issue for Task #{task_num}: {task_name}")

        # 校验 body 不为空
        if not body.strip():
            print(f"  ❌ 任务文件 body 内容为空，无法创建 Issue")
            print(f"  请先在任务文件中添加内容（frontmatter 之后的部分）")
            sys.exit(1)

        # 创建 GitHub Issue
        title = f"[{task_num}] {task_name}"
        # 将 body 写入临时文件避免 shell 转义问题
        tmp_body = f"/tmp/epic-sync-body-{task_num}.md"
        with open(tmp_body, "w", encoding="utf-8") as f:
            f.write(body)

        code, out, err = run(
            f'gh issue create --title "{title}" --body-file "{tmp_body}" --label "epic:{feature_name}" 2>&1 || '
            f'gh issue create --title "{title}" --body-file "{tmp_body}" 2>&1'
        )

        # 清理临时文件
        if os.path.exists(tmp_body):
            os.remove(tmp_body)

        if code != 0 and not out:
            print(f"  ❌ 创建失败: {err}")
            sys.exit(1)

        # 提取 Issue URL
        issue_url = ""
        for line in out.splitlines():
            if "github.com" in line and "/issues/" in line:
                issue_url = line.strip()
                break

        if not issue_url:
            # 尝试从输出中提取
            url_match = re.search(r'https://github\.com/[^/]+/[^/]+/issues/\d+', out)
            if url_match:
                issue_url = url_match.group()

        if not issue_url:
            print(f"  ⚠️  Issue 可能已创建但无法提取 URL: {out}")
            sys.exit(1)

        # 提取 Issue ID
        issue_id = issue_url.rstrip('/').split('/')[-1]
        print(f"  ✅ Created: {issue_url}")

        # 更新 frontmatter
        update_frontmatter_field(task_file, "github", issue_url)

        # 重命名文件
        new_path, renamed = ensure_filename(feature_name, task_file, issue_id, task_name)
        if renamed:
            print(f"  📝 Renamed: {basename} → {os.path.basename(new_path)}")

        # Rate limiting
        time.sleep(0.5)

    # === UPDATE ===
    elif action == "UPDATE":
        task_github = fm.get("github", "")
        issue_num_match = re.search(r'/issues/(\d+)', task_github)
        if not issue_num_match:
            print(f"  ❌ 无法从 github URL 提取 Issue 编号: {task_github}")
            sys.exit(1)

        issue_num = issue_num_match.group(1)
        print(f"🔄 Updating Issue #{issue_num} for Task #{task_num}: {task_name}")

        # 校验 body 不为空
        if not body.strip():
            print(f"  ❌ 任务文件 body 内容为空，无法更新 Issue")
            print(f"  请先在任务文件中添加内容（frontmatter 之后的部分）")
            sys.exit(1)

        # 写入临时文件
        tmp_body = f"/tmp/epic-sync-body-{task_num}.md"
        with open(tmp_body, "w", encoding="utf-8") as f:
            f.write(body)

        code, out, err = run(f'gh issue edit {issue_num} --body-file "{tmp_body}"')

        if os.path.exists(tmp_body):
            os.remove(tmp_body)

        if code != 0:
            print(f"  ❌ 更新失败: {err}")
            sys.exit(1)

        print(f"  ✅ Updated: {task_github}")

        # 强制检查文件名
        new_path, renamed = ensure_filename(feature_name, task_file, issue_num, task_name)
        if renamed:
            print(f"  📝 Renamed: {basename} → {os.path.basename(new_path)}")

        # Rate limiting
        time.sleep(0.5)

    # === SYNCED ===
    elif action == "SYNCED":
        task_github = fm.get("github", "")
        issue_num_match = re.search(r'/issues/(\d+)', task_github)
        issue_num = issue_num_match.group(1) if issue_num_match else task_num

        # 强制检查文件名
        new_path, renamed = ensure_filename(feature_name, task_file, issue_num, task_name)

        print(f"✓ Task #{task_num}: Already synced")
        if renamed:
            print(f"  📝 Renamed: {basename} → {os.path.basename(new_path)}")

    else:
        print(f"ERROR: 未知操作: {action}")
        sys.exit(1)


if __name__ == "__main__":
    main()
