---
name: issue-production-ready
description: 将 OPC_OS_HTML_to_Flutter_Migration 的 issue 从硬编码 Mock 升级为生产就绪，补充 API 设计、DDL、共享服务层、代码复用规范。Use when upgrading migration issues to production-ready specs, or when the user says "production ready" or "升级 issue".
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[issue_number: NNN | next | status]"
disable-model-invocation: true
---

# Issue Production Ready Skill

逐个完善 OPC-OS Epic 的 issue 文档：前端已实现（使用硬编码 Mock 数据），需要补充后端 API/DDL/共享服务规格，使 issue 达到「全栈生产就绪」标准。

**前端现状**: 37 个页面的 Flutter 代码已全部完成，但 Provider 中使用 `_getMock*Data()` 硬编码数据。
**本 skill 目标**: 为每个 issue 补充后端规格 + 前端改造计划，使开发者可直接实现正式功能。

**执行模式**: 每次调用处理 **1 个 issue**，更新进度文件，等待用户确认后再处理下一个。支持后续动态新增 issue。

**⚠️ 核心原则**:
- **不允许有任何 mock 数据、TBD、TODO、占位符**
- 每个 API Response 字段、每个 DDL 列、每个 fromJson 映射都必须精确定义
- 完成后的 issue 必须通过 `epic-decompose-validation` 规范校验
- **issue 是规格说明书，不是代码文件** — 禁止出现大段 Go/Dart/Python 代码块，用结构化文字和表格描述

---

## 调用方式

| 命令 | 说明 |
|------|------|
| `/issue-production-ready 001` | 完善指定 issue（支持任意编号，不限 001-037） |
| `/issue-production-ready next` | 自动选择下一个未完善的 issue |
| `/issue-production-ready status` | 查看整体进度 |
| `/issue-production-ready add 038` | 将新 issue 加入进度跟踪 |

---

## Role Definition

**You are a senior full-stack engineer with 10 years experience in Go backend (DDD/Hertz) and Flutter frontend (Riverpod).** Focus on practical, production-grade API design and code reuse.

---

## 必须遵循的规范

完善后的每个 issue 必须同时满足：
1. **本 skill 的生产就绪标准**（API/DDL/Repository/fromJson 完整）
2. **`epic-decompose-validation` 规范**（参见 `.Codex/skills/epic-decompose-validation/SKILL.md`）

### epic-decompose-validation 关键要求

以下每一项都必须在完善后的 issue 中存在，缺失则补充：

| 检查项 | 要求 | 示例 |
|--------|------|------|
| **Frontmatter** | name, status, created, updated, github, depends_on, parallel, deprecated | YAML 格式 |
| **Task Metadata** | File, Purpose, Leverage, Requirements | 见下方模板 |
| **Prompt** | `Role: ... \| Task: ... \| Restrictions: ... \| Success: ...` | 单行格式 |
| **验收标准** | SMART 格式，含具体指标 (%, ms, count) | `API 响应 < 200ms` |
| **完成定义** | Code complete + Tests pass + Docs updated + Deployment verified | 4 项都要 |
| **工作量评估** | Size (XS/S/M/L/XL) + Hours + Risk | `Size: M, 8-16h, Risk: 中` |
| **技术详情** | API: Endpoint+Request+Response；Flutter: Widget+State | 表格格式 |
| **实现计划** | 分步骤的实现顺序 | 后端→前端→测试 |
| **测试策略** | 单元测试 + 集成测试 + 边界条件 | 具体测试点 |
| **边界情况** | 空数据、网络错误、权限不足等 | 列出处理方式 |
| **自包含** | 零外部引用 | 禁止 "See Epic.md" |

### Task Metadata 模板

在 issue 的 `## 技术详情` 前补充（如果缺失）：

```markdown
## Task Metadata

- **File**: `features/{module}/state/{module}_providers.dart` (前端改造) + `domain/opc_{domain}/` (后端新增)
- **Purpose**: 为 {page_name} 实现后端 API 并将前端 Mock 数据替换为真实 API 调用
- **Leverage**: `coze-studio/backend/domain/` 已有服务、`shared/` 共享模块
- **Requirements**: 零 Mock 数据、API 字段完整、复用共享服务
- **Prompt**: Role: 全栈工程师（Go+Flutter） | Task: 实现 {page_name} 后端 API 和前端 Mock→API 改造 | Restrictions: 零硬编码、复用共享服务、coze-studio DDD 风格 | Success: API 可调用、前端无 Mock、测试通过
```

---

## 动态 Issue 管理

当用户新增 issue 时（如 `add 038`）：
1. 扫描 `.Codex/epics/OPC_OS_HTML_to_Flutter_Migration/` 目录找到新 issue 文件
2. 将其加入进度文件
3. 如果有对应的 dart 文件，更新 `reference/issue-dart-file-index.md`
4. 如果涉及新的 API 域，更新 `reference/api-domain-mapping.md`

---

## 进度跟踪

进度文件: `.Codex/epics/OPC_OS_HTML_to_Flutter_Migration/production-ready-progress.md`

每次执行完一个 issue 后，必须更新此文件。格式见下方「进度文件格式」。

### 进度文件格式

```markdown
# Production Ready Progress

| # | 页面 | 状态 | 完善日期 | 备注 |
|---|------|------|---------|------|
| 001 | 首页 | ✅ done | 2026-02-14 | API 3个, DDL 复用 opc_user_profile |
| 002 | AI员工列表 | ⏳ pending | | |
| ... | ... | ... | | |
```

**状态值**: `✅ done` / `⏳ pending` / `🔄 in-progress` / `⚠️ needs-review`

---

## 单个 Issue 执行流程（5 步）

### Step 1: 读取上下文（必须认真阅读 dart 文件）

1. 读取进度文件，确认当前 issue 状态
2. 读取 issue 文件: `.Codex/epics/OPC_OS_HTML_to_Flutter_Migration/NNN-*.md`
3. **认真阅读**对应 Flutter dart 文件（这是最关键的一步）:
   - `data/models/*_models.dart` — 逐字段分析每个 class 的属性
   - `state/*_providers.dart` — 逐行分析 `_getMock*Data()` 中的硬编码值
   - `presentation/pages/*_page.dart` — 理解 UI 如何消费数据
4. 读取参考资料:
   - `reference/issue-dart-file-index.md` — **首先查此文件**，获取该 issue 的精确 dart 文件路径和核心 Model 列表
   - `reference/api-domain-mapping.md` — 查找该 issue 对应的 API 域和 endpoint 设计
   - `reference/ddl-schema.md` — 查找该 issue 需要的数据库表
   - `reference/shared-services.md` — 查找该 issue 复用的共享服务
   - `reference/code-reuse-patterns.md` — 查找代码复用模式

**输出**:
- 逐个列出 dart Model 的每个字段（字段名、类型、用途）
- 逐个列出 Mock 数据中的硬编码值（哪些需要从 API 获取、哪些需要从 DB 查询、哪些是计算值）
- 对应的 API 域和需要的数据库表

---

### Step 2: 设计后端规格

基于 Step 1 的分析，为该 issue 设计：

#### 2.1 API 设计
- 每个 API 必须包含: Endpoint / Method / Request（每个字段） / Response（每个字段，精确到类型和含义） / Error Codes
- **Response 字段必须与 dart Model 字段一一对应** — 不能遗漏任何 Mock 中出现的字段
- **Response 中的每个子对象必须有独立的字段表** — 禁止“类似映射”“同上”等偷懒写法
- 聚合查询优先（一个页面尽量一个聚合 API）
- 写操作独立 endpoint
- HTTP Method: 查询用 GET，创建/修改用 POST，根据实际场景选择
- 优先复用 coze-studio 已有 API（agent/workflow/knowledge/plugin/conversation/memory）
- 统一响应格式: `{ "code": 0, "msg": "success", "data": { ... } }`

#### 2.2 DDL 设计
- 如果 `reference/ddl-schema.md` 已有该 issue 需要的表，写明「复用 ddl-schema.md 中的 opc_xxx 表」
- **如果本 issue 是某张表的「首次定义」，必须给出完整可执行的 CREATE TABLE 语句**
- 如果复用 coze-studio 已有表（agent/workflow 等），说明复用方式
- **DDL 风格必须与 coze-studio 一致**: BIGINT 时间戳（非 DATETIME）、软删除 `deleted_at BIGINT DEFAULT 0`、JSON 扩展字段

#### 2.3 Go Domain 层设计
- Go 包路径、Entity struct、Repository 接口、Service 接口、Handler 实现
- **Go Entity 风格必须与 coze-studio 一致**:
  - 时间字段用 `int64`（Unix 毫秒时间戳），不用 `time.Time`
  - 主键用 `int64`，不用 `uint64`
  - JSON tag 用 `json:"snake_case"`
  - 不用 gorm tag（coze-studio 不用 gorm）
  - 软删除用 `DeletedAt int64` 默认 0

---

### Step 3: 设计前端改造规格

基于 Step 1 的 Mock 数据分析：

#### 3.1 数据模型改造
- 列出需要增加 `fromJson` / `toJson` 的**每一个** Model class
- **每个 Model 都必须有完整字段映射表**，禁止"类似映射""同上""其他 Model 类似实现"等偷懒写法
- **用映射表描述转换逻辑，不要写 fromJson 代码块**
- 字段映射表格式:
```
| Dart 字段 | Dart 类型 | JSON 字段 | JSON 类型 | 转换说明 |
|-----------|----------|-----------|----------|----------|
| avatarColor | OpcTagColor | avatar_color | string | OpcTagColor.values.byName() |
```

#### 3.2 Repository 层
- 接口定义（方法签名）
- API 实现类

#### 3.3 Provider 改造
- 原 Mock Provider → 新 API Provider
- 共享 Provider 复用（引用 `shared/providers/` 下的已有 Provider）

#### 3.4 代码复用
- 该 issue 与其他 issue 共享的 Model / Provider / Repository
- 需要提取到 `shared/` 的公共代码

---

### Step 4: 更新 Issue 文件

**⚠️ 直接修改 issue 文件，不询问用户确认。**

在 issue 的 `## 技术详情` 部分，**保留原有前端内容**，追加后端和改造内容。

**写入标准**: 写入的每一行都必须是可直接用于实现的精确规格。开发者读完 issue 后，不需要再查阅任何其他文档就能开始编码。

**禁止出现的写法**（发现则必须展开写完整）:
- ❌ "类似映射"、"同上"、"参考 xxx"、"详见 xxx"
- ❌ "其他 Model 类似实现 fromJson"
- ❌ "字段略"、"// ..."、"etc."
- ❌ 一个映射表覆盖多个 Model（每个 Model 必须有独立的映射表）
- ❌ 大段 Go/Dart/Python 代码块（issue 是规格说明书，不是代码文件）

**代码 vs 文字的原则**:
- Go Entity → 用表格列出字段名、类型、说明，不要写 struct 代码
- Go Repository → 用文字列出方法签名（一行一个），不要写完整 interface 代码块
- Go Handler → 用文字描述路由绑定和处理流程，不要写函数实现
- Dart fromJson → 用映射表描述，不要写 factory 代码
- Dart Repository → 用文字描述接口方法和文件路径，不要写 abstract class 代码
- Dart Provider → 用文字描述改造方案（删除什么、替换为什么），不要写 Provider 代码
- DDL → CREATE TABLE 语句是唯一允许的代码块（因为需要直接执行）

**验收标准/完成定义处理规则**:
- 如果原 issue 已有验收标准/完成定义，**在原有基础上追加**后端相关项，不要另建第二份
- 已完成的前端项保持 `[x]`，新增的后端项用 `[ ]`

#### 写入内容模板

```markdown
### 后端 (coze-studio)

#### API 设计

##### API 1: GET /api/opc/v1/{domain}/{action}

**Description**: [一句话]

**Request**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| space_id | string | Y | 工作空间ID |

**Response Data**:
| 字段 | 类型 | 说明 |
|------|------|------|
| user_name | string | 用户显示名 |
| workers | []WorkerBrief | 员工列表 |
| workers[].id | string | 员工ID |
| workers[].name | string | 员工名 |

**Error Codes**:
| code | msg | 触发条件 |
|------|-----|----------|
| 400001 | invalid_param | 缺少必填参数 |
| 404001 | not_found | 资源不存在 |
| 500001 | internal_error | 服务端异常 |

#### DDL

[首次定义的表给完整 CREATE TABLE 语句；复用的表写明「复用 ddl-schema.md 中的 opc_xxx 表」]

#### Go Domain 层

**Entity** (`domain/opc_{domain}/entity.go`):
| 字段 | 类型 | JSON tag | 说明 |
|------|------|----------|------|
| ID | int64 | id | 主键 |
| SpaceID | int64 | space_id | 工作空间 |
| ... | ... | ... | ... |

**Repository** (`domain/opc_{domain}/repository.go`):
- `GetByUserID(ctx, spaceID, userID int64, limit int) ([]Entity, error)`
- `Create(ctx, entity *Entity) error`

**Service** (`application/opc_{domain}/service.go`):
- `GetPageData(ctx, spaceID, userID int64) (*PageData, error)` — 聚合多个 Repository 数据

**Handler** (`api/handler/coze/opc_{domain}_service.go`):
- 路由: `GET /api/opc/v1/{domain}/{action}`
- 流程: 参数校验 → 调用 Service → 返回 JSON

### 前端改造计划

#### 数据模型 fromJson/toJson

**{ModelName}**:
| Dart 字段 | Dart 类型 | JSON 字段 | JSON 类型 | 转换 |
|-----------|----------|-----------|----------|------|
| id | String | id | string | 直接 |
| avatarColor | OpcTagColor | avatar_color | string | enum byName |

[每个子 Model 都必须有独立的映射表，不能省略]

#### Repository 层（新增文件）

- **接口**: `data/repositories/{feature}_repository.dart`
  - `Future<{PageData}> getPageData({required String spaceId, required String userId})`
- **实现**: `data/repositories/{feature}_api_repository.dart`
  - 注入 ApiClient，调用 `GET /api/opc/v1/{domain}/{action}`
  - 解析 `{code, msg, data}` 响应，调用 `{PageData}.fromJson(data)`

#### Provider 改造

- **删除**: `_getMock*Data()` 函数及所有硬编码数据
- **新增**: `{feature}RepositoryProvider` — 提供 Repository 实例
- **改造**: `{feature}DataProvider` — 从 Mock 改为调用 Repository

#### 共享服务复用
- [列出该 issue 复用的共享服务，引用首次定义的 issue 编号]

#### 代码复用分析
- [与其他 issue 共享的 Model/Provider]
- [需提取到 shared/ 的公共模块]
```

同时更新 frontmatter:
```yaml
updated: {当前时间 ISO 8601}
```

同时确保 issue 包含以下 `epic-decompose-validation` 必需内容。如果缺失则补充，已有则追加后端相关项：

**1. Task Metadata**（如缺失，在 `## 技术详情` 前补充）:
```markdown
## Task Metadata

- **File**: `features/{module}/state/{module}_providers.dart` + `domain/opc_{domain}/`
- **Purpose**: 为 {page_name} 实现后端 API 并将 Mock 替换为 API 调用
- **Leverage**: coze-studio 已有服务、shared/ 共享模块
- **Requirements**: 零 Mock、API 字段完整、复用共享服务
- **Prompt**: Role: 全栈工程师(Go+Flutter) | Task: 实现 {page_name} 后端 API+前端改造 | Restrictions: 零硬编码/复用共享服务/DDD风格 | Success: API可调用/前端无Mock/测试通过
```

**2. 验收标准**（追加到已有部分，不另建第二份）:
```markdown
### 后端验收
- [ ] 所有 API endpoint 实现并可调用
- [ ] API 响应时间 < 200ms
- [ ] 错误码返回正确

### 前端改造验收
- [ ] 所有 Model.fromJson() 实现
- [ ] Repository 层实现
- [ ] Provider 改造完成（移除 _getMock*Data）
- [ ] 零硬编码 Mock 数据
```

**3. 完成定义**（追加到已有部分，不另建第二份）:
```markdown
### 后端实现
- [ ] API 实现并提交
- [ ] 后端单元测试
- [ ] API 集成测试

### 前端改造
- [ ] fromJson/toJson 完成
- [ ] Provider Mock → API 完成
- [ ] 无遗留 TODO/FIXME

### 部署验证
- [ ] API 部署到测试环境
- [ ] 前后端联调通过
```

**4. 实现计划**（如缺失则补充）:
```markdown
## 实现计划

1. 后端: 创建 DDL → Entity → Repository → Service → Handler → 注册路由
2. 前端: Model 添加 fromJson → 创建 Repository 接口+实现 → 改造 Provider → 删除 Mock
3. 测试: 后端单元测试 → API 集成测试 → 前后端联调
```

**5. 测试策略**（如缺失则补充）:
```markdown
## 测试策略

### 后端测试
- Service 层单元测试: 每个方法的正常/异常路径
- API 集成测试: 每个 endpoint 的 200/400/404/500 响应

### 前端测试
- fromJson 单元测试: 每个 Model 的序列化/反序列化
- Provider 测试: 数据加载、错误处理、刷新
```

**6. 边界情况**（如缺失则补充）:
```markdown
## 边界情况

| 场景 | 处理方式 |
|------|----------|
| 网络错误 | 显示错误页+重试按钮 |
| 空数据 | 显示引导页 |
| Token 过期 | 跳转登录页 |
| 参数无效 | 返回 400 + 错误描述 |
```

> **注意**: 不修改 `status: completed`，因为前端已完成。后端实现是新的工作。

---

### Step 5: 更新进度

1. 更新进度文件中该 issue 的状态为 `✅ done`
2. 输出完成摘要:
   - API 数量
   - 新建/复用的表
   - 复用的共享服务
   - 下一个待处理的 issue 编号

---

## Architecture Context

### Backend (coze-studio)
- **Framework**: Go + Hertz HTTP + Thrift IDL
- **Architecture**: DDD (domain/ → application/ → api/)
- **Database**: MySQL 8.4.5
- **Cache**: Redis 8.0
- **API Pattern**: POST JSON, `BindAndValidate`, `c.JSON(200, resp)`
- **IDL Location**: `coze-studio/idl/`
- **Handler Location**: `coze-studio/backend/api/handler/coze/`
- **Domain Location**: `coze-studio/backend/domain/`

### Frontend (Flutter)
- **State**: Riverpod (ConsumerWidget / FutureProvider / StateProvider)
- **Current Pattern**: Provider → `_getMock*Data()` → 硬编码数据
- **Target Pattern**: Provider → Repository → ApiClient → HTTP → coze-studio API
- **Mock Location**: 各 `state/*_providers.dart` 内 `_getMock*Data()` 函数
- **Shared Widgets**: `features/home/presentation/widgets/opc_shared_widgets.dart`

### coze-studio 可复用服务

| OPC 功能 | 复用 coze-studio 服务 |
|---------|---------------------|
| 员工聊天 | `conversation` domain |
| 员工记忆 | `memory` domain |
| Coze 工作流 | `workflow` domain |
| MCP 工具 | `plugin` domain |
| 知识库 | `knowledge` domain |
| Agent 配置 | `agent` domain |

---

## 共享服务速查表

执行每个 issue 时，检查该 issue 是否属于以下共享服务组：

| 共享服务 | 涉及 Issue | 首次定义于 |
|----------|-----------|-----------|
| WorkerService | #001,#002,#003,#004,#005,#023,#024 | #002 |
| TaskService | #001,#003,#005,#013,#034 | #001 |
| CreditsService | #001,#005,#016,#021 | #016 |
| PhoneService | #001,#032,#033,#034 | #032 |
| ContentService | #010,#011 | #010 |
| CrmService | #012 | #012 |
| FinanceService | #021,#022 | #021 |
| MarketplaceService | #018,#019,#020 | #018 |
| AutomationService | #025,#026,#027,#028 | #025 |
| NotificationService | #009,#030 | #009 |
| ProfileService | #014,#015,#031 | #014 |
| QualityService | #035,#036,#037 | #035 |

**规则**: 如果当前 issue 不是「首次定义于」的 issue，则在文档中写「复用 #{首次定义} 定义的 {Service}」，不重复设计。

---

## Validation Checklist（严格执行）

每个完善后的 issue 必须通过以下**全部**检查，任何一项不通过则不能标记为 done：

### 生产就绪检查
- [ ] **零 Mock** — 新增内容中不出现 "Mock"、"模拟数据" 等字样（仅在「改造前」描述中可提及）
- [ ] **零 TBD** — 无 TBD、TODO、占位符、"待定"、"后续补充"
- [ ] **API 字段完整** — dart Model 的每个字段都在 API Response 表中有对应行
- [ ] **DDL 字段完整** — API Response 的每个持久化字段都在 DDL 中有对应列（计算字段需注明计算逻辑）
- [ ] **fromJson 映射完整** — 每个需要序列化的 Model 都有完整的字段映射表
- [ ] **共享服务不重复** — 不重复定义已有共享服务的 API/DDL
- [ ] **Repository 接口明确** — 有完整的方法签名
- [ ] **Error Codes 完整** — 每个 API 都有错误码定义
- [ ] **自包含** — 开发者只读该 issue 就能实现，不需要查阅其他文档

### epic-decompose-validation 合规检查
- [ ] **Frontmatter 完整** — name, status, created, updated, depends_on, github, parallel, deprecated
- [ ] **Task Metadata** — File, Purpose, Leverage, Requirements, Prompt 五项齐全
- [ ] **Prompt 格式** — `Role: ... | Task: ... | Restrictions: ... | Success: ...` 单行格式
- [ ] **验收标准** — 含具体指标（%, ms, count），后端+前端改造项齐全
- [ ] **完成定义** — Code complete + Tests pass + Docs updated + Deployment verified
- [ ] **工作量评估** — Size + Hours + Risk
- [ ] **实现计划** — 分步骤的实现顺序（后端→前端→测试）
- [ ] **测试策略** — 后端单元测试 + API 集成测试 + 前端测试
- [ ] **边界情况** — 网络错误、空数据、Token 过期等处理方式

---

## Output Language

- **中文输出**: 所有 issue 文档内容、进度报告
- **英文**: 代码、API path、字段名、DDL、Go 代码
