---
name: epic-decompose-validation
description: Validate that epic decomposition produces executable, self-contained tasks with complete technical details. Use when validating task quality after epic decomposition.
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[epic_name]"
disable-model-invocation: true
---

# Epic Decompose Validation

验证 Epic 拆解结果：确保每个任务文件可执行、自包含、技术细节完整。

## Scripts

| Script | Purpose |
|--------|---------|
| `scan-tasks.py` | 扫描 epic 目录下所有任务文件 |
| `validate-task.py` | 单任务文件验证（frontmatter/metadata/技术细节/自包含性） |
| `check-dependencies.py` | 依赖图检查（循环依赖、无效引用） |

## Additional Resources

- 验证报告格式详见 [output-formats.md](output-formats.md)
- 完整验证清单（API 示例、Flutter 示例、Error Handling、Self-Containment patterns）详见 [validation-checklist.md](validation-checklist.md)

---

## Role Definition

**You are a senior software engineer with 10 years of development experience (NOT a system architect).** Focus on practical, implementable technical solutions.

---

## Required Rules

验证前必须读取：
- `.Codex/rules/pm/task-template-structure.md` — 定义任务元数据规范：File, Purpose, Leverage, Requirements, Prompt

---

## Validation Objectives

1. **Completeness** — 每个任务包含所有必需字段
2. **Self-Containment** — 任务不引用外部文件，独立可理解
3. **Executability** — 开发者可直接实现，无需额外说明
4. **Consistency** — 所有任务遵循统一结构

### Success Criteria

- **ALL tasks pass** → Ready for `/epic-sync`
- **ANY task fails** → Must fix before proceeding

---

## 4-Step Workflow

### Step 1: Scan Task Files

```bash
python3 .Codex/skills/epic-decompose-validation/scripts/scan-tasks.py $ARGUMENTS
```

列出所有 `NNN-*.md` 任务文件（排除 `epic.md`、`github-mapping.md`、deprecated 任务）。

---

### Step 2: Validate Each Task

对每个任务文件运行验证脚本：

```bash
python3 .Codex/skills/epic-decompose-validation/scripts/validate-task.py <task_file>
```

**验证检查项**：

#### 2.1 Frontmatter（必需字段）
```yaml
name: Task title
status: open | in-progress | completed
created: 2025-11-27T00:00:00Z
updated: 2025-11-27T00:00:00Z
```

#### 2.2 Task Metadata（per task-template-structure.md）

| Field | Description |
|-------|-------------|
| **File** | Target file path |
| **Purpose** | One-line purpose statement |
| **Leverage** | Existing code to reuse |
| **Requirements** | Requirement IDs |
| **Prompt** | `Role: ... | Task: ... | Restrictions: ... | Success: ...` |

#### 2.3 Technical Details（CRITICAL）

- **API 任务** → 必须包含 Endpoint、Request、Response 规格
- **Backend 任务** → 必须包含 Data Model + Database Schema
- **Frontend 任务** → 必须包含 Component Structure + State Management
- **Flutter 任务** → 必须包含 Widget Structure + State Management + Pubspec Dependencies
- **所有任务** → Implementation Plan + Testing Strategy + Edge Cases

#### 2.4 Self-Containment（零外部引用）

**禁止出现**：
- `"See Epic.md for API details"`
- `"Refer to PRD section 3.2"`
- `TBD`、`TODO`、`to be determined`

#### 2.5 Acceptance Criteria
- SMART 格式（Specific, Measurable, Achievable, Relevant, Time-bound）
- 包含具体指标（%, ms, count）

#### 2.6 Definition of Done
- Code complete、Tests pass、Docs updated、Deployment verified

---

### Step 3: Cross-Task Checks

```bash
python3 .Codex/skills/epic-decompose-validation/scripts/check-dependencies.py $ARGUMENTS
```

- 验证依赖图无循环
- 检查 `depends_on` 引用的任务编号有效
- 验证 Epic 功能覆盖完整

---

### Step 4: Auto-Remediation & Report

**⚠️ 关键原则：发现问题后直接修复，不询问用户确认。**

**自动修复流程**（当任务验证失败时）：

1. 识别问题来源
2. 读取 Epic 源文件 (`.Codex/epics/<epic>/epic.md`) 获取缺失内容
3. **立即修复任务文件（不等待用户确认）**
4. 重新验证确认修复
5. 报告修复结果（仅输出最终报告，不中途询问）

**Auto-Fix 操作**：

| Issue | Fix Action |
|-------|------------|
| 缺少 Frontmatter | 从任务内容 + 当前时间戳生成 |
| 缺少 Task Metadata | 从任务体提取 File/Purpose/Leverage/Requirements，生成 Prompt |
| Prompt 格式无效 | 重组为 `Role|Task|Restrictions|Success` |
| 缺少 API Specs | 从 Epic Technical Approach 提取 |
| 缺少 Technical Details | 根据任务类型从 Epic 提取 |
| 外部引用 | **内联**引用内容到任务中 |
| TBD/占位符 | 从 Epic 提取实际值或标记 `[NEEDS_INPUT]` |
| 缺少 Acceptance Criteria | 从 Core Features 生成 SMART 标准 |
| 缺少 DoD | 添加标准 DoD 模板 |
| 日期格式无效 | 转换为 ISO 8601 |

**无法自动修复的项目**标记为 `[NEEDS_INPUT]`，附 TODO 注释。

**输出报告** — 详见 [output-formats.md](output-formats.md)。

---

## Score Thresholds

| Score | Status | Action |
|-------|--------|--------|
| 90-100 | ✅ Excellent | Ready for implementation |
| 75-89 | ⚠️ Acceptable | Minor fixes recommended |
| 60-74 | ⚠️ Needs Work | Fix before implementation |
| <60 | ❌ Failed | Must fix before proceeding |

**Critical Violations → 自动失败（<60）**

---

## Performance Guidelines

- **Batch Processing**: 最多 5 个任务并行验证
- **Timeout**: 30s/任务，5min/整个 Epic
- **Incremental**: 仅重新验证 `updated` 字段变更的任务

---

## Output Language

- **中文输出**: 所有验证报告、日志、错误消息
