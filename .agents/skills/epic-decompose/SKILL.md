---
name: epic-decompose
description: Decompose an Epic into executable task files with frontend/backend differentiation. Use when breaking down a feature Epic into implementable tasks, or when the user says "decompose epic".
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[feature_name]"
disable-model-invocation: true
---

# Epic Decompose

将 Epic 拆解为可执行的任务文件，支持前后端项目的差异化处理。

## Scripts

| Script | Purpose |
|--------|---------|
| `preflight-check.py` | Epic 预检：验证存在、统计任务、检测孤立任务 |
| `lock.py` | 并发锁管理：获取/释放/检查 decompose 锁 |
| `compare-tasks.py` | 任务存在性检查：按编号和标题匹配现有任务 |
| `atomic-write.py` | 原子写入：通过临时文件确保写入完整性 |

## Templates

| Template | Purpose |
|----------|---------|
| `github-mapping-template.md` | GitHub Issue 映射文件模板 |

## Additional Resources

- 任务比较算法详见 [comparison-algorithm.md](comparison-algorithm.md)
- 各步骤标准输出格式详见 [output-formats.md](output-formats.md)

---

## Core Concept

> Feature decomposition + UI/UX first, technical details second

- **Epic**: high-level feature description
- **Task**: executable units (1-3 days)
- **Frontend**: components, interactions, UI/UX
- **Backend**: APIs, database, business logic

---

## Role Definition

**You are a senior software engineer with 10 years of development experience (NOT a system architect).** Focus on practical, implementable technical solutions: code structure, API design, data models, and concrete implementation details.

---

## Execution Principles

### ⚠️ CRITICAL: Rules Are MANDATORY, Not Optional

- Every rule in this document and referenced files **MUST** be followed
- No shortcuts, no assumptions, no "good enough" implementations
- Each rule exists to prevent defects or ensure quality

### Mandatory Requirements

1. **Technical Details MUST Include**:
   - API Specifications (endpoint, method, params, response)
   - Data Models (for backend tasks)
   - Component Structure (for frontend tasks)
   - Business Logic (implementation approach)

2. **Each Task MUST Be**:
   - Independently verifiable (1-3 days of work)
   - User-value focused (not just tech implementation)
   - Self-contained (no external dependencies for understanding)

3. **No Partial Implementation**:
   - Don't assume "API details are in Epic"
   - Don't skip "painful" technical details
   - Don't defer complex requirements to later

4. **Existence-First Principle (防重复)**:
   - If task file exists (by number/title/content match) → UPDATE or KEEP
   - Never CREATE a new task if similar file exists
   - Double-check before CREATE to prevent duplicates

---

## Required Rules

Must read in order before execution:

1. **Core Principles**: `.Codex/rules/pm/core-principles.md`
2. **Common Template**: `.Codex/rules/pm/task-template-common.md`
3. **Technical Template**: `.Codex/rules/pm/task-template-tech.md`
4. **Task Structure Template**: `.Codex/rules/pm/task-template-structure.md`
5. **Datetime**: `.Codex/rules/datetime.md`

---

## Workflow: Sequential Execution

```
Step 0: Preflight Check
   ↓
Step 1: Extract Tasks from Epic
   ↓
Step 2: Compare with Existing Tasks
   ↓
Step 3: Detect Conflicts
   ↓
Step 4: Execute Actions (KEEP/UPDATE/CREATE/DEPRECATE)
   ↓
Step 5: Validate & Report
   ↓
Step 6: GitHub Sync Check
```

Each step depends on previous output, strict sequential execution required.

---

## Concurrency Protection

获取锁后再执行：
```bash
python3 .Codex/skills/epic-decompose/scripts/lock.py acquire $ARGUMENTS
```

执行完毕后释放锁：
```bash
python3 .Codex/skills/epic-decompose/scripts/lock.py release $ARGUMENTS
```

Lock file: `.Codex/epics/<epic_name>/.decompose.lock`  
Timeout: 10s | Stale: 5min auto-clean

---

## Execution Steps

### Step 0: Preflight Check

运行预检脚本：
```bash
python3 .Codex/skills/epic-decompose/scripts/preflight-check.py $ARGUMENTS
```

验证：Epic 文件存在、frontmatter 完整、统计现有任务、检测孤立任务。
如果 Epic 状态为 `completed`，发出警告。

---

### Step 1: Extract Tasks from Epic

从 Epic 提取任务和技术规格。

**提取来源**:

- **Source 1: Technical Approach** — 功能列表、组件、API、状态
- **Source 2: Task Breakdown Preview** — 任务列表、状态标记

**Tech Spec 提取（按项目类型）**:

| Spec | Frontend | Backend |
|------|----------|---------|
| `Component:` | ✅ | - |
| `Route:` | ✅ | - |
| `State:` | ✅ | - |
| `UI/UX:` | ✅ | - |
| `API:` | ✅ | ✅ |
| `SQL:` | - | ✅ |
| `Model:` | - | ✅ |

**Task Metadata 提取（per `task-template-structure.md`）**:

| Field | Description |
|-------|-------------|
| **File** | Target file path |
| **Purpose** | One-line purpose statement |
| **Leverage** | Existing code to reuse |
| **Requirements** | Requirement IDs |
| **Prompt** | `Role: ... | Task: ... | Restrictions: ... | Success: ...` |

**Merge Strategy**: Task Breakdown Preview（优先） + Technical Approach（补充详情） → 合并去重

---

### Step 2: Compare with Existing Tasks

对每个提取的任务运行存在性检查：
```bash
python3 .Codex/skills/epic-decompose/scripts/compare-tasks.py $ARGUMENTS <task_number> "<task_title>"
```

**⚠️ CRITICAL: Existence-First Principle**

**Iron Rule**: 如果任务文件存在（编号/标题/内容匹配），**必须** UPDATE 或 KEEP，**绝不** CREATE。

决策流程及详细算法参见 [comparison-algorithm.md](comparison-algorithm.md)。

输出格式详见 [output-formats.md](output-formats.md) Step 2 section。

---

### Step 3: Detect Conflicts

识别相似任务，防止重复。

- 对每个待创建/更新的任务，与所有现有任务比较相似度
- **0.75-0.85** → 潜在冲突，标记 `conflicts_with`
- **≥ 0.85** → 高度相似，建议合并

处理原则：
- **不删除**: 保留所有现有任务
- **标记**: frontmatter 添加 `conflicts_with: ["003", "007"]`
- **日志**: 输出冲突详情供人工审查

---

### Step 4: Execute Actions

应用 KEEP/UPDATE/CREATE/DEPRECATE 操作。

| 操作 | 条件 | 行为 |
|------|------|------|
| **KEEP** | similarity ≥ 0.85 或 completed | 不修改文件，仅日志 |
| **UPDATE** | similarity < 0.85 且 open | 智能合并（保留现有 + 添加新增） |
| **CREATE** | 文件不存在 | 生成新任务文件 |
| **DEPRECATE** | 孤立任务 | 标记 `deprecated: true` |

**Smart Merge（UPDATE）规则**:
1. 保留现有实现细节
2. 添加 Epic 中新增的技术规格
3. 确保 task metadata 完整：File, Purpose, Leverage, Requirements, Prompt
4. 去重
5. 更新 `updated` 时间戳
6. 不覆盖已完成部分

**CREATE 安全检查**:
- ⚠️ 执行前重新扫描目录确认无匹配文件（防重复二次检查）
- 如果发现类似文件，转为 UPDATE

**原子写入**（所有文件操作）:
```bash
echo "content" | python3 .Codex/skills/epic-decompose/scripts/atomic-write.py <target_file>
```

---

### Step 5: Validate & Report

**验证检查**:
- 所有 Epic 功能有对应任务
- 状态标记正确映射（✅→completed, [ ]→open）
- 无循环依赖
- 已完成任务所有 checkbox `[x]`
- Task metadata 完整：File, Purpose, Leverage, Requirements, Prompt

**生成 GitHub Mapping 文件**:
创建 `.Codex/epics/<feature_name>/github-mapping.md`

**输出总结报告** — 详见 [output-formats.md](output-formats.md) Step 5 section。

---

### Step 6: GitHub Sync Check

释放锁并提示下一步：
```bash
python3 .Codex/skills/epic-decompose/scripts/lock.py release $ARGUMENTS
```

```
Epic decompose complete, GitHub sync required separately

Next Steps:
- Run: /epic-sync <feature_name>
- This creates/updates GitHub Issues
- Updates github-mapping.md file
```

---

## Project Type Handling

### Frontend
Priority: Component Structure → State Management → UI/UX → Routing → Performance  
Template: `task-template-tech.md` frontend sections

### Backend
Priority: API Specifications → Database Design → Business Logic → Error Handling → Security  
Template: `task-template-tech.md` backend sections

### Fullstack
Use unified `task-template-tech.md`, select sections by task type.

---

## Summary

1. Read 4 rule files (core-principles, common, tech, structure)
2. Execute 6-step workflow (Preflight → Extract → Compare → Detect → Execute → Report)
3. **Existence-First Principle**: If task exists → UPDATE/KEEP, never CREATE
4. **Required task metadata**: File, Purpose, Leverage, Requirements, Prompt
5. Fully automated (no user interaction)
6. Atomic write protection (data integrity)
7. Concurrency lock (prevent conflicts)
8. **中文输出**: 所有日志、报告、任务内容均使用中文
