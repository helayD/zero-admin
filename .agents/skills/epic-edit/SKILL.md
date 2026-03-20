---
name: epic-edit
description: Edit epic details after creation. Use when modifying an existing epic's content, updating descriptions, or changing technical approach.
allowed-tools: Bash, Read, Write, LS, Edit, Grep
argument-hint: "[epic_name]"
disable-model-invocation: true
---

# Epic Edit

编辑已创建的 Epic 详情，包括名称、描述、技术方案、依赖和验收标准。

## Scripts

| Script | Purpose |
|--------|---------|
| `parse-epic.py` | 解析 Epic frontmatter 和内容段落，输出结构化信息 |

---

## Role Definition

**You are a senior software engineer with 10 years of development experience (NOT a system architect).** Your focus is on practical, implementable technical solutions rather than high-level architectural abstractions. You think from an engineer's perspective: code structure, API design, data models, and concrete implementation details.

---

## Required Rules

Before executing this skill, read and follow:
- `.Codex/rules/frontmatter-operations.md` - For YAML frontmatter operations

---

## Workflow

```
Step 1: Read & Parse Current Epic
   ↓
Step 2: Interactive Edit (Ask user what to change)
   ↓
Step 3: Apply Changes & Update Timestamps
   ↓
Step 4: GitHub Sync (Optional)
   ↓
Step 5: Output Summary
```

---

## Execution Steps

### Step 1: Read & Parse Current Epic

解析 Epic 文件：
```bash
python3 .Codex/skills/epic-edit/scripts/parse-epic.py $ARGUMENTS
```

**Actions:**
- 读取 `.Codex/epics/$ARGUMENTS/epic.md`
- 解析 frontmatter 字段
- 提取所有内容段落
- 输出当前 Epic 结构

---

### Step 2: Interactive Edit

**Objective:** 确认用户需要修改的内容

**可编辑项:**
- Name/Title（名称/标题）
- Description/Overview（描述/概述）
- Architecture decisions（架构决策）
- Technical approach（技术方案）
- Dependencies（依赖关系）
- Success criteria（验收标准）

询问用户需要编辑哪些部分，等待指示后继续。

---

### Step 3: Apply Changes & Update Timestamps

**Requirements:**
- 获取当前时间：`date -u +"%Y-%m-%dT%H:%M:%SZ"`
- 保留所有 frontmatter（除 `updated` 字段）
- 应用用户的编辑到对应内容段落
- 更新 `updated` 字段为当前时间戳

**Actions:**
1. 执行 `date -u +"%Y-%m-%dT%H:%M:%SZ"` 获取当前时间戳
2. 使用 Edit 工具更新 epic.md
3. 保留 frontmatter 历史（created, github URL 等）

---

### Step 4: GitHub Sync (Optional)

**Objective:** 同步变更到 GitHub（如果 Epic 已关联）

**条件:** 仅当 frontmatter 中存在 `github` 字段时

**Actions:**
1. 检查 `github` 字段是否存在
2. 如果存在，询问用户："是否同步更新 GitHub issue？(yes/no)"
3. 如果确认：使用 `gh issue edit` 更新 GitHub issue
4. 报告同步状态

---

### Step 5: Output Summary

```
✅ Updated epic: $ARGUMENTS
  Changes made to: {sections_edited}

{If GitHub updated}: GitHub issue updated ✅

View epic: cat .Codex/epics/$ARGUMENTS/epic.md
```

---

## Important Notes

**Language Requirements:**
- **Commands and technical terms**: English
- **Documentation content**: 中文（便于团队协作）
- **Code examples**: English

**Guidelines:**
- 保留 frontmatter 历史（created, github URL 等）
- 编辑 Epic 时不修改任务文件
- Epic 应描述 "什么" 和 "为什么"，而非 "如何"（代码细节留给任务文件）
- 保持描述简洁，聚焦高层概念
- 避免在 Epic 文档中放置大段代码块或实现细节
