---
name: issue-commit
description: Commit changes with proper categorization and link to GitHub issue. Use when committing code changes related to a specific issue number.
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[issue_number]"
disable-model-invocation: true
---

# Issue Commit

Commit changes with proper categorization and link to GitHub issue.

## Scripts

| Script | Purpose |
|--------|---------|
| `pre-commit-check.py` | 预检：自动发现 issue 号、检查变更、验证分支、查找任务文件 |
| `categorize-changes.py` | 变更文件分类统计（Source/Test/Config/Doc/Other） |
| `auto-push.py` | 自动推送到远程（首次推送自动设置上游） |

## Usage

**Auto-discovery mode** (recommended):
```
Use this skill when you need to commit changes
```
- Auto-extracts issue number from branch name or recent commits
- Prompts for manual input if not found

**Manual mode**:
- Directly specify issue number via $ARGUMENTS

## Required Rules

Before executing this skill, read and follow:
- `.Codex/rules/github-operations.md` - For GitHub CLI operations
- `.Codex/rules/standard-patterns.md` - For validation patterns

---

## Quick Check

运行预检脚本：
```bash
python3 .Codex/skills/issue-commit/scripts/pre-commit-check.py $ARGUMENTS
```

此脚本会自动完成：
1. 自动发现 issue 号（分支名 → 提交消息 → 手动）
2. 验证 issue 号格式和 GitHub 可访问性
3. 检查是否有未提交的更改
4. 验证不在 main/master 分支
5. 查找本地任务文件（可选）

输出 `ISSUE_NUMBER`、`ISSUE_TITLE`、`BRANCH`、`TASK_FILE`、`EPIC_NAME` 等变量。

---

## Instructions

### 1. Analyze Changed Files

运行分类统计脚本：
```bash
python3 .Codex/skills/issue-commit/scripts/categorize-changes.py
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
git commit -m "Issue #$ARGUMENTS: {issue_title}

Changes:
- {Summarize main changes}

Related to: #$ARGUMENTS"
```

### 5. Verify Commit

```bash
git log -1 --oneline
git show --stat HEAD
```

### 6. Update Task Status

If local task file exists and is writable:
- If status is not `in_progress`: Update frontmatter to `status: in_progress`
- Create backup before modification, restore from backup if update fails
- Append commit reference to task file

### 7. Push to Remote

```bash
python3 .Codex/skills/issue-commit/scripts/auto-push.py
```

### 8. Output Summary

```
✓ Commit completed and pushed to remote

Issue: #$ARGUMENTS - {issue_title}
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
- Complete task: When ready, run issue-close skill
- View changes: git log --oneline -5
```

## Additional Resources

- Best Practices、Common Issues、Auto-Discovery Examples、Important Notes 详见 [best-practices.md](best-practices.md)
