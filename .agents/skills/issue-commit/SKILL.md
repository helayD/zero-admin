---
name: issue-commit
description: Commit changes with proper categorization and link to local story file. Use when committing code changes related to a specific story ID (e.g. 1-4, 2-1). Replaces old GitHub issue-based workflow with story-ID-based workflow.
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[story_id]"
disable-model-invocation: true
---

# Issue Commit (Story-ID Based)

Commit changes with proper categorization and link to local story file.

## 核心变更

**旧版**：基于 GitHub Issue 号（`gh issue view` 验证）
**新版**：基于 Story ID（如 `1-4`），从本地 `_opcos/` 目录查找 story 文件

## Story ID 格式

- 格式：`X-Y`（如 `1-4`、`2-1`、`10-3`）
- 查找路径：
  - `_opcos/implementation-artifacts/` — 已实现 story
  - `_opcos/planning-artifacts/` — 规划中 story
  - `_opcos/epics/` — Epic 文件

## Scripts

| Script | Purpose |
|--------|---------|
| `pre-commit-check.py` | 预检：自动发现 story ID、检查变更、验证分支、查找 story 文件 |
| `categorize-changes.py` | 变更文件分类统计（Source/Test/Config/Doc/Other） |
| `auto-push.py` | 自动推送到远程（首次推送自动设置上游） |

## Usage

**Auto-discovery mode** (recommended):
```
Use this skill when you need to commit changes
```
- 自动从分支名或提交消息提取 story ID（格式 `X-Y`）
- 自动查找对应 story 文件并提取标题
- Prompts for manual input if not found

**Manual mode**:
- 直接指定 story ID：`/issue-commit 1-4`

## Quick Check

运行预检脚本：
```bash
python3 .claude/skills/issue-commit/scripts/pre-commit-check.py $ARGUMENTS
```

此脚本会自动完成：
1. 自动发现 story ID（分支名 → 提交消息 → 手动）
2. 验证 story ID 格式（`X-Y`）
3. 查找本地 story 文件，提取标题
4. 检查是否有未提交的更改
5. 验证不在 main/master 分支

输出 `STORY_ID`、`STORY_TITLE`、`STORY_FILE`、`EPIC_NAME`、`BRANCH` 等变量。

---

## Instructions

### 1. Analyze Changed Files

运行分类统计脚本：
```bash
python3 .claude/skills/issue-commit/scripts/categorize-changes.py
```

### 2. Review Changes

```bash
git status -s
git diff --stat
```

### 3. Stage Changes

```bash
git add -A
```

### 4. Create Commit Message

```bash
git commit -m "Story $ARGUMENTS: {story_title}

Changes:
- {Summarize main changes}

See _opcos/..."
```

### 5. Verify Commit

```bash
git log -1 --oneline
git show --stat HEAD
```

### 6. Update Task Status

如果本地 story 文件存在且可写：
- 如果 status 不是 `in_progress`：更新 frontmatter 为 `status: in_progress`
- 修改前先备份，失败时从备份恢复
- 在文件末尾追加 commit 引用

### 7. Push to Remote

```bash
python3 .claude/skills/issue-commit/scripts/auto-push.py
```

### 8. Output Summary

```
✓ Commit completed and pushed to remote

Story: $ARGUMENTS - {story_title}
Branch: {current_branch}
Commit: {commit_hash}
Remote: origin/{current_branch}

Files committed:
  Source: {count}
  Tests: {count}
  Config: {count}
  Docs: {count}

Next steps:
- Continue work: Make more changes and run this skill again
- Complete story: Mark story as done in sprint-status.yaml
- View changes: git log --oneline -5
```

## Additional Resources

- Best Practices、Common Issues、Auto-Discovery Examples、Important Notes 详见 [best-practices.md](best-practices.md)
