---
name: prd-parse
description: Convert PRD to technical implementation epic. Use when parsing a PRD into an epic, or when the user says "parse PRD" or "create epic from PRD".
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[feature_name]"
disable-model-invocation: true
---

# PRD Parse

将 PRD 转换为技术实现 Epic，确保零遗漏覆盖。

## Scripts

| Script | Purpose |
|--------|---------|
| `preflight-check.py` | PRD 预检验证（文件存在、frontmatter、内容完整性、已有 Epic 检查） |

## Templates

| Template | Purpose |
|----------|---------|
| `epic-template.md` | Epic 文件结构模板（frontmatter + 8 sections） |

## Additional Resources

- API/DDL 提取规范详见 [api-ddl-extraction.md](api-ddl-extraction.md)
- 质量验证清单详见 [quality-validation.md](quality-validation.md)

---

## Role Definition

**You are a senior software engineer with 10 years of development experience (NOT a system architect).** Focus on practical, implementable technical solutions: code structure, API design, data models, and concrete implementation details.

**⚠️ Epic 的核心是技术实现细节，不是架构设计。** 避免抽象的系统架构讨论，聚焦于：
- **具体的 API 端点设计**（endpoint、method、request/response schema）
- **具体的数据模型**（表结构、字段、索引、约束）
- **具体的代码结构**（文件路径、类/函数、接口定义）
- **具体的业务逻辑实现方案**（算法、流程、规则）

---

## Required Rules

执行前必须读取：
1. `.Codex/rules/datetime.md` — ISO 8601 时间戳
2. `.Codex/commands/pm/prd-parse-functional-checklist.md` — 功能挖掘清单

---

## ⚠️ CRITICAL REQUIREMENTS

- **Response Language**: 始终用中文回复
- **Zero-Omission Principle**: PRD 中每个功能必须覆盖
- **Mandatory Review**: 逐行对比 PRD 进行审查
- **Completeness First**: 宁多勿漏

---

## 5-Step Workflow

### Step 0: Preflight Check

```bash
python3 .Codex/skills/prd-parse/scripts/preflight-check.py $ARGUMENTS
```

脚本验证后，继续执行增强功能挖掘：

**Enhanced Feature Mining**（参考 `prd-parse-functional-checklist.md`）：
- 提取 **所有** 显式功能需求
- 挖掘功能变体、扩展、边界情况
- 提取 **所有** API 接口设计 — 详见 [api-ddl-extraction.md](api-ddl-extraction.md)
- 提取 **所有** DDL 数据模型设计 — 详见 [api-ddl-extraction.md](api-ddl-extraction.md)
- 识别功能依赖和关系

**Feature Value Assessment**（每个功能评估）：
- Business value / User value / Strategic value
- 分配优先级：P0 Core / P1 Important / P2 Value-Add
- **阻断规则**: 如果任何 P0/高价值功能缺失，STOP

**Feature Completeness Validation**：
- 交叉验证 PRD 需求与挖掘功能
- 确认无有价值功能遗漏

如果已有 Epic 且用户未确认覆盖，STOP。

---

### Step 1: Read PRD

- 加载 PRD: `.Codex/prds/$ARGUMENTS.md`
- 分析所有需求和约束
- 理解 User Stories 和 Success Criteria
- 提取 PRD description from frontmatter
- **创建 Feature Inventory**: 列出每个功能

---

### Step 2: Technical Detail Analysis

- 确定技术栈和实现方案（**不是架构选型**）
- **将每个 PRD 功能映射到具体的代码实现**：API 端点、数据表、业务类/函数
- 列出每个功能需要的 API 接口完整规格
- 列出每个功能涉及的数据表和字段
- 识别集成点和依赖
- **验证映射完整性**: 确认所有功能已映射到具体技术实现

---

### Step 3: Create Epic File

**Location**: `.Codex/epics/$ARGUMENTS/epic.md`

使用模板创建：参考 `templates/epic-template.md`

**Epic 必须包含 8 个 section**:

1. **Overview** — 技术实现方案简述（聚焦实现，非架构）
2. **Technical Decisions** — 技术选型决策 + 具体理由（如选 MyBatis 而非 JPA，选 Redis 而非本地缓存）
3. **System Overview** — Mermaid 图仅作辅助，**重点是具体的模块、API、数据流**
4. **Technical Details** — **核心 section**，包含：
   - 完整的 API 端点列表（endpoint、method、params、response）
   - 完整的数据模型（CREATE TABLE、字段、索引）
   - 具体的业务逻辑实现方案（代码级描述）
   - 每项引用 PRD 需求
5. **Implementation Strategy** — 阶段划分、风险缓解、测试方案
6. **Task Breakdown Preview** — 5-10 个具体可执行任务
7. **Dependencies** — 外部/内部依赖、前置工作、阻碍
8. **Success Criteria (Technical)** — 性能基准、质量门、验收标准

**Task Breakdown 要求**:
- 每个任务追溯到具体 PRD 需求
- 按层/功能/阶段分组
- 1-3 天工作量（8-24h）
- 包含 PRD Coverage 引用
- 标注并行化机会和关键路径

---

### Step 4: Quality Validation

**⚠️ MANDATORY: PRD 覆盖率审查**

完整验证清单详见 [quality-validation.md](quality-validation.md)。

核心检查流程：
1. **Feature Inventory Comparison** — 每个 PRD 功能检查 ✅/❌
2. **Section-by-Section PRD Review** — User Stories / Requirements / Criteria 逐项对比
3. **Cross-Reference Validation** — 每个 PRD 条目有对应 Epic 条目
4. **Technical Detail Validation** — API/Table/Field 数量 diff，gap > 0 则 STOP

**如果检测到功能遗漏**：
- 立即停止 Epic 创建
- 列出所有遗漏功能及 PRD 引用
- 修订 Epic 包含所有缺失功能
- 重新运行覆盖率审查确认 100%

---

### Step 5: Post-Creation

1. **输出覆盖率报告** — 格式详见 [quality-validation.md](quality-validation.md) Coverage Report section
2. **确认创建**: `✅ Epic created: .Codex/epics/$ARGUMENTS/epic.md`
3. **显示摘要**: 任务分类数、关键架构决策
4. **建议下一步**: `Ready to decompose? Run: /epic-decompose $ARGUMENTS`

---

## Error Recovery

| 错误 | 处理 |
|------|------|
| PRD 不完整 | 列出缺失 section |
| 技术方案不明确 | 标识需要澄清的点 |
| 功能遗漏 | STOP，列出遗漏，修订后重新验证 |
| 目录权限问题 | 提示检查权限 |

**绝不在信息不完整时创建 Epic**

---

## Important Notes

- **ZERO-OMISSION is TOP PRIORITY** — 永远不为简洁牺牲覆盖率
- 目标 ≤10 任务，但 PRD 功能需要独立任务时必须创建
- Epic 侧重 **技术实现细节**（API、数据模型、业务逻辑），不是抽象架构
- 包含必要的代码示例（接口定义、DDL、关键算法），但避免完整实现代码
- **中文输出**: 所有报告、日志使用中文
