---
name: epic-sync
description: Synchronize epic tasks with GitHub Issues in sequential order. Use when syncing local task files to GitHub, or when the user says "sync epic".
allowed-tools: Bash, Read, Write, LS, Edit, Grep
argument-hint: "[feature_name] [--check-only | --sync-all]"
disable-model-invocation: true
---

# Epic Sync

将 Epic 任务文件与 GitHub Issues 同步，保持本地与远程 1:1 一致。

## Scripts

| Script | Purpose |
|--------|---------|
| `validate-env.py` | Step 1: 验证 GitHub CLI、认证、仓库权限、安全检查 |
| `load-epic-tasks.py` | Step 2: 加载 Epic 元数据和所有任务文件信息 |
| `sync-task.py` | Step 5: 执行单个任务的同步操作（CREATE/UPDATE/SYNCED/SKIP） |

---

## Synchronization Objective

**维护本地任务文件与 GitHub Issues 之间的 1:1 一致性。**

- 每个本地任务文件 → 恰好一个 GitHub Issue
- 内容和状态必须在本地和 GitHub 之间保持匹配
- 任务按严格顺序处理（001 → 002 → 003...）
- 不并行或批量操作

### File Naming Convention

**任务文件必须使用格式: `{issue_id}-{title}.md`**

- **首次同步前**: 文件使用序号: `001-task_name.md`
- **同步后**: 文件重命名为: `{issue_id}-{title}.md`
- **强制重命名**: CREATE、UPDATE、SYNCED 操作中都必须检查并确保文件名格式正确

---

## Usage

**Auto-discovery mode** (recommended):
```
Use this skill when you need to sync epic tasks to GitHub
```

**Arguments:**
- `<feature_name>`: Epic 目录名
- `--check-only`: Dry-run 模式，只显示计划不执行
- `--sync-all`: (默认) 全自动同步模式

---

## Required Rules

Before executing this skill, read and follow:
1. `.Codex/rules/github-operations.md` - GitHub CLI 操作和仓库保护
2. `.Codex/rules/datetime.md` - 时间戳格式（最高优先级）
3. `.Codex/rules/strip-frontmatter.md` - 同步前移除 frontmatter
4. `.Codex/rules/frontmatter-operations.md` - YAML frontmatter 原子更新

---

## Core Principles

### Separation of Concerns
- `epic-decompose`: 只管理本地任务文件
- `epic-sync`: 只管理 GitHub Issue 同步

### Sequential Execution Guarantee
- 按任务编号严格顺序处理（001, 002, 003...）
- 不批量或并行操作
- 每个任务同步完成后才处理下一个

### Automation Guarantee
全自动执行，不需要用户交互。所有决策基于 Epic 内容和任务状态自动做出。

---

## 8-Step Workflow

```
Step 1: Validate Environment
   ↓
Step 2: Load Epic and Task Data
   ↓
Step 3: Analyze Sync Status (Automated Decision)
   ↓
Step 4: Sort and Determine Sync Strategy
   ↓
Step 5: Execute Sync Operations (Sequential)
   ↓
Step 6: Update Epic Sync Timestamp
   ↓
Step 7: Display Summary
   ↓
Step 8: Update GitHub Mapping File
```

---

## Execution Steps

### Step 1: Validate Environment

运行环境验证脚本：
```bash
python3 .Codex/skills/epic-sync/scripts/validate-env.py
```

验证项：
1. GitHub CLI 已安装（`command -v gh`）
2. GitHub 已认证（`gh auth status`）
3. 仓库可访问（`gh repo view`）
4. 安全检查：不允许同步到 CCPM/automazeio 仓库

---

### Step 2: Load Epic and Task Data

加载 Epic 和任务数据：
```bash
python3 .Codex/skills/epic-sync/scripts/load-epic-tasks.py $ARGUMENTS
```

**操作：**
1. 验证 Epic 文件存在
2. 提取 Epic 元数据（name, github, last_sync）
3. 加载所有任务文件并提取 frontmatter
4. 输出结构化数据

---

### Step 3: Analyze Sync Status (Automated Decision)

**决策逻辑：**
- 统计需要 CREATE 的任务（github 为空且 status=open）
- 统计需要 UPDATE 的任务（github 不为空且 task_updated > last_sync）
- 统计已 SYNCED 的任务（github 不为空且 task_updated <= last_sync）

**自动决策：**
- 如果 last_sync 为空或有需要同步的任务 → 继续 Step 4
- 如果所有任务已同步 → 显示 "✅ 全部已同步" 并退出

---

### Step 4: Sort and Determine Sync Strategy

**按编号排序后，为每个任务确定操作：**

| Action | 条件 | 含义 |
|--------|------|------|
| **CREATE** | github 为空且 status=open | 需要创建新 GitHub Issue |
| **UPDATE** | github 不为空且 updated > last_sync | 需要更新内容 |
| **SYNCED** | github 不为空且 updated <= last_sync | 已同步 |
| **SKIP** | github 为空且 status≠open | 已完成但未同步，跳过 |

---

### Step 5: Execute Sync Operations (Sequential)

**对每个任务执行同步：**
```bash
python3 .Codex/skills/epic-sync/scripts/sync-task.py <feature_name> <task_file> <action>
```

#### Action: CREATE
1. 读取任务内容，剥离 frontmatter
2. **校验 body 不为空** — 如果 frontmatter 之后无内容，报错退出
3. `gh issue create --title "[{task_num}] {task_name}" --body "{body}" --label "epic:{feature_name}"`
4. 提取 Issue ID
5. 更新任务文件 frontmatter（`github: {issue_url}`）
6. **重命名文件**: `{task_num}-name.md` → `{issue_id}-name.md`

#### Action: UPDATE
1. 提取 Issue 编号
2. 读取并剥离 frontmatter
3. **校验 body 不为空** — 如果 frontmatter 之后无内容，报错退出
4. `gh issue edit {issue_num} --body "{body}"`
5. **强制检查并重命名文件**

#### Action: SYNCED
1. **强制检查并重命名文件**
2. 显示已同步状态

#### Action: SKIP
- 显示跳过状态

#### Rate Limiting
每次 CREATE/UPDATE 后：`sleep 0.5`

---

### Step 6: Update Epic Sync Timestamp

```bash
date -u +"%Y-%m-%dT%H:%M:%SZ"
```

更新 `epic.md` frontmatter 中的 `last_sync` 字段。

---

### Step 7: Display Summary

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 GitHub Sync Complete: {feature_name}

Operations:
- 🆕 Created: {create_count} new issues
- 🔄 Updated: {update_count} existing issues
- ✅ Synced: {synced_count} already in sync
- ⏭️  Skipped: {skipped_count} skipped
- 📝 Renamed: {rename_count} files

Results:
- ✅ Successful: {success_count}
- ❌ Failed: {failure_count}

Next steps:
- View issues: gh issue list --label 'epic:{feature_name}'
- Continue working: use issue-start skill
━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### Step 8: Update GitHub Mapping File

如果 `.Codex/epics/{feature_name}/github-mapping.md` 存在：
1. 更新所有同步的任务条目（URL、文件名、状态）
2. 更新统计信息和同步元数据
3. 记录同步操作详情

---

## Check-Only Mode

当提供 `--check-only` 参数时：
1. 执行 Step 1-4
2. Step 5 只显示计划，不执行
3. 跳过 Step 6（不更新时间戳）
4. 跳过 Step 8（不更新 mapping）

---

## Error Handling

| 错误类型 | 处理 |
|----------|------|
| Rate Limiting | 显示警告并退出 |
| Permission Error | 显示权限错误并退出 |
| Network Error | 显示网络错误并退出 |
| 单任务失败 | 记录错误，继续处理下一个任务 |
| 文件重命名失败 | 标记部分失败，继续 |

---

## Data Integrity Guarantees

- **顺序执行**: 严格按编号顺序处理
- **1:1 映射**: 每个本地文件对应一个 GitHub Issue
- **原子操作**: 所有文件写入使用 temp + move 模式
- **一致性**: 同步后内容/状态/文件名/元数据全部匹配
- **Frontmatter 处理**: 同步到 GitHub 时必须剥离 frontmatter

---

## Workflow Constraints Summary

**MUST:**
- 按顺序处理任务（001 → 002 → 003...）
- **所有任务文件必须使用 `{issue_id}-{title}.md` 格式**
- 同步到 GitHub 前剥离 frontmatter
- 使用原子文件写入
- GitHub API 调用间隔 0.5s
- 单任务失败后继续处理
- 同步完成后更新 `last_sync`
- 更新 `github-mapping.md`

**MUST NOT:**
- 并行或批量操作
- 同步到 CCPM/automazeio 仓库
- GitHub Issue body 中包含 frontmatter
- 提示用户做决策
- 单任务失败时停止执行
