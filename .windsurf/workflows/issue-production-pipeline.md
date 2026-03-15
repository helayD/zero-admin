---
description: Issue 全栈生产就绪流水线。获取下一个待处理 issue，使用 issue-production-ready 完善规格，编译验证，API 测试，前后端字段对接确认。
---

# Issue 全栈生产就绪流水线

> **⚠️ 强制执行项 — 不可跳过**:
> 1. **DDL + 种子数据**: 每个 issue 必须创建 `ddl_and_seed.sql` 并执行导入，确保测试数据存在
> 2. **API 测试脚本**: 每个 issue 必须创建 `test_api.sh` 并实际运行，确保 21+ 测试用例全部通过
> 3. **重编译+重启**: 代码修改后必须 `coze-manage.sh build` + `coze-manage.sh start` 重启服务再测试
> 4. 以上三项必须在流水线中实际执行并输出结果，不可仅创建文件而不运行
>
> **目标**: 端到端完成一个 issue 的「规格完善 → 后端开发 → 编译验证 → DDL+种子数据 → API 测试 → 前后端对接确认」全流程。
> **依赖 Skill**: `@issue-production-ready`、`@issue-start`
> **编译工具**: `coze-studio/scripts/coze-manage.sh`
> **Issue 目录**: `.claude/epics/OPC_OS_HTML_to_Flutter_Migration/`
> **Issue 编号范围**: `345-381`（对应 skill reference 中的 `#001-#037`，映射: GitHub Issue# = 344 + skill序号）
> **Issue 文件命名**: `{issue_number}-{页面名称}.md`（如 `345-首页迁移.md`、`346-AI员工列表页迁移.md`）
> **Issue 工作目录**: `issues/{issue_number}/`（含 `progress.md`、`test_api.sh`、`ddl_and_seed.sql`）
> **Skill 参考文件**: `.claude/skills/issue-production-ready/reference/`（含 `issue-dart-file-index.md`、`api-domain-mapping.md`、`ddl-schema.md`、`shared-services.md`、`code-reuse-patterns.md`）
>
> **重要机制 — 阶段总结**: 每完成一个步骤后，必须在 issue 工作目录下更新 `progress.md` 总结文件，确保上下文连贯。

---

## 前置条件

- coze-studio 后端可编译（Go 环境就绪）
- 数据库服务已启动（MySQL/Redis 等）
- 进度文件存在：`.claude/epics/OPC_OS_HTML_to_Flutter_Migration/production-ready-progress.md`

### 数据库连接方式

使用 `coze-manage.sh` 脚本管理数据库：

```bash
# 查看数据库连接信息
bash coze-studio/scripts/coze-manage.sh db-info

# 数据库调试（连接测试 + 表信息 + OPC 表状态）
bash coze-studio/scripts/coze-manage.sh db-debug

# 执行 DDL 迁移（指定文件）
bash coze-studio/scripts/coze-manage.sh db-migrate path/to/schema.sql

# 执行 DDL 迁移（自动扫描 migrations/ 目录）
bash coze-studio/scripts/coze-manage.sh db-migrate

# 直接用 mysql CLI 连接（默认配置）
mysql -u coze -pcoze123 -h 127.0.0.1 -P 3306 opencoze
```

环境文件加载优先级（未指定 `--env` 时）：
1. `APP_ENV` 环境变量 → `docker/.env.$APP_ENV`
2. 自动扫描: `docker/.env` → `.env.debug` → `.env.local` → `.env.production`

可通过 `--env <file>` 指定环境文件：`bash coze-manage.sh --env docker/.env.debug db-info`

### 服务编译与启动方式

```bash
# 方式 1：使用 coze-manage.sh 一键编译+启动（推荐）
bash coze-studio/scripts/coze-manage.sh start

# 方式 2：仅编译（生成 bin/opencoze）
bash coze-studio/scripts/coze-manage.sh build

# 方式 3：手动编译到 /tmp 并启动（开发调试用）
# 编译
cd coze-studio/backend && go build -o /tmp/coze-studio-server .
# 启动（必须在 bin/ 目录下执行，因为需要 .env 文件）
cd coze-studio/bin && /tmp/coze-studio-server

# 仅编译检查（不生成二进制）
bash coze-studio/scripts/coze-manage.sh check

# OPC 域编译检查
bash coze-studio/scripts/coze-manage.sh check-opc

# 状态总览（Go/Docker/DB）
bash coze-studio/scripts/coze-manage.sh status
```

**注意**: 服务启动时需要 `.env` 文件在工作目录中。`coze-studio/bin/.env` 是本地开发环境配置。
使用 `start` 命令会自动复制 `docker/.env*` 到 `bin/` 目录。

### 🚨 用户信息规范（强制）

所有 OPC 域 Provider 必须通过统一的用户信息 Provider 获取 `spaceId` 和 `userId`，**禁止硬编码**。

**统一入口**: `flutter_client/lib/src/features/shared/state/opc_user_providers.dart`

```dart
import '../../shared/state/opc_user_providers.dart';

// ✅ 正确用法：从全局 Provider 获取
final spaceId = ref.watch(opcSpaceIdProvider);
final userId = ref.watch(opcUserIdProvider);

// ❌ 禁止：硬编码占位值
const spaceId = '1';  // 禁止！
const userId = '1';   // 禁止！
```

**Provider 定义**:
- `opcSpaceIdProvider` — 当前租户 Space ID（MVP 阶段默认 '1'，后续接入多租户）
- `opcUserIdProvider` — 当前用户 ID（从 `currentUserProfileProvider` 获取真实用户 ID）

**已接入的 Provider**:
- `onboarding_providers.dart` / `track_assessment_providers.dart`
- `briefing_providers.dart` / `opc_dashboard_providers.dart`
- `opc_home_providers.dart` / `workers_providers.dart`
- `worker_detail_providers.dart` / `worker_chat_providers.dart`

---

## 步骤 1：获取下一个待处理任务

使用 `@issue-production-ready` skill 获取下一个 `⏳ pending` 状态的 issue：

```
@issue-production-ready next
```

**输出要求**:
- 确认 issue 编号和页面名称
- 确认对应的 dart 文件路径
- 确认对应的 API Domain

如果所有 issue 都已完成，流水线结束。

### 📝 阶段总结 1

创建或更新 issue 目录下的 `progress.md`，记录：

```markdown
# Issue #<NNN> 流水线进度

## 基本信息
- **Issue**: #<NNN> <页面名称>
- **API Domain**: opc-xxx
- **Dart 文件**: features/xxx/state/xxx_providers.dart
- **当前阶段**: 步骤 1 完成 — 任务已获取

## 步骤 1 完成
- 获取任务: ✅
- Issue 编号: #<NNN>
- 页面: <名称>
- API Domain: opc-xxx
```

---

## 步骤 2：检测 issue 规格完成度（前置质量门禁 ≥ 95%）

在执行开发前，先检查当前 issue 文档的完善程度。

### 2.1 完成度检查清单

对 issue `.md` 文件逐项检查以下内容，每项 1 分，满分 20 分：

| # | 检查项 | 权重 | 说明 |
|---|--------|------|------|
| 1 | Frontmatter 完整 | 1 | name/status/created/updated/github/depends_on/parallel/deprecated |
| 2 | 功能描述 | 1 | WHAT + WHY |
| 3 | 用户工作流 | 1 | HOW（触发→交互→导航→反馈→异常） |
| 4 | UI 元素清单 | 1 | 页面+组件列表 |
| 5 | 功能验收标准 | 1 | 前端验收项（已勾选） |
| 6 | 后端验收标准 | 1 | API endpoint + 响应时间 + 错误码 |
| 7 | 前端改造验收 | 1 | fromJson + Repository + Provider 改造 |
| 8 | Task Metadata | 1 | File/Purpose/Leverage/Requirements/Prompt |
| 9 | API 设计 — Endpoint 定义 | 1 | Method + Path + 描述 |
| 10 | API 设计 — Request 字段表 | 1 | 每个字段：名称/类型/必填/说明 |
| 11 | API 设计 — Response 字段表 | 1 | 每个字段精确到类型，子对象独立表，禁止「同上」 |
| 12 | API 设计 — 错误码表 | 1 | code + msg + 触发条件 |
| 13 | DDL 设计 | 1 | 完整 CREATE TABLE 或明确复用说明 |
| 14 | Go Domain 层设计 | 1 | Entity/Repository/Service/Handler |
| 15 | 前端 fromJson 映射 | 1 | 每个 Model 的 fromJson 字段映射表 |
| 16 | 前端 Repository 接口 | 1 | 接口定义 + ApiRepository 实现 |
| 17 | 前端 Provider 改造 | 1 | Mock→API 改造方案 |
| 18 | 实现计划 | 1 | 分步骤实现顺序 |
| 19 | 测试策略 | 1 | 单元测试 + 集成测试 + 边界条件 |
| 20 | 边界情况 | 1 | 空数据/网络错误/权限不足等处理方式 |

### 2.2 计算完成度

```
完成度 = 已满足项数 / 20 × 100%
```

**质量门禁**: 完成度必须 **≥ 95%**（即最多 1 项缺失）。

- **≥ 95%**: 通过，进入步骤 3
- **< 95%**: 不通过，列出缺失项，使用 `@issue-production-ready <issue_number>` 补全后重新检测

### 📝 阶段总结 2

更新 `progress.md`，追加：

```markdown
## 步骤 2 完成 — 规格完成度检测
- 完成度: XX% (XX/20)
- 缺失项: [列表或“无”]
- 门禁结果: ✅ 通过 / ❌ 未通过
- API 数量: X 个
- DDL 表: [opc_xxx, opc_yyy]
- Model 数量: X 个 fromJson
```

---

## 步骤 3：使用 issue-start 执行开发任务

规格通过质量门禁后，使用 `@issue-start` skill 开始实际开发：

```
@issue-start <issue_number>
```

> **🚨 强制要求**: 必须严格按照 `issue-start` skill 的 **9 步工作流**顺序执行，不得跳步、合并步骤或未经验证即继续。

### 3.0 issue-start 9 步流程（必须逐步执行）

| 步骤 | 内容 | 必须产出 |
|------|------|----------|
| Step 1 | quick-check.py + 读取任务 + init-analysis.py + 业务上下文 | `analysis.md` (Business Context) |
| Step 2 | Code Review + 跨模块依赖分析 | `analysis.md` (Technical Approach, Affected Files) |
| Step 3 | detect-stack.py + go vet + dart analyze | 零错误报告 |
| Step 4 | 实现方案决策 + 风险评估 | `analysis.md` (Implementation Plan, Risk Mitigation) |
| Step 5 | init-progress.py + GitHub 分配 | `progress.md` |
| Step 6 | 编码执行 + log-progress.py | 代码文件 + 进度日志 |
| Step 7 | 验证实现结果 | 测试通过报告 |
| Step 8 | Code Review 自审 | 自审清单全通过 |
| Step 9 | verify-docs.py + 文档交叉引用 | frontmatter documentation 字段 |

**必须运行的脚本**（不可跳过）：
```bash
# Step 1
python3 .claude/skills/issue-start/scripts/quick-check.py <issue_number>
python3 .claude/skills/issue-start/scripts/init-analysis.py <issue_number> <epic_name> "任务概述"
# Step 3
python3 .claude/skills/issue-start/scripts/detect-stack.py .
# Step 5
python3 .claude/skills/issue-start/scripts/init-progress.py <issue_number> <epic_name>
# Step 6 (每次重要变更后)
python3 .claude/skills/issue-start/scripts/log-progress.py <issue_number> <epic_name> "完成了什么"
# Step 9
python3 .claude/skills/issue-start/scripts/verify-docs.py <issue_number> <epic_name>
```

### 3.1 后端开发
- 创建/扩展 Go Domain（Entity + Repository + Service）
- 实现 API Handler（注册路由）
- 执行 DDL 迁移（如需新表）

### 3.2 前端改造
- 实现所有 Model 的 `fromJson()` / `toJson()`
- 创建 Repository 接口 + ApiRepository 实现
- 改造 Provider：移除 `_getMock*Data()`，调用 Repository
- **🚨 用户信息**: Provider 中的 `spaceId`/`userId` 必须使用 `opcSpaceIdProvider`/`opcUserIdProvider`，禁止硬编码

### 3.3 开发完成标志
- 后端：所有 API endpoint 代码已写入
- 前端：零 Mock 数据，所有 Provider 调用 Repository
- 路由已注册

### 📝 阶段总结 3

更新 `progress.md`，追加：

```markdown
## 步骤 3 完成 — 开发执行
- 后端文件创建: [entity.go, dao.go, service.go, handler.go, routes 注册]
- 前端文件修改: [models.dart, repository.dart, providers.dart]
- 新增/修改文件清单:
  - `coze-studio/backend/domain/opc_xxx/entity/xxx.go` — Entity 定义
  - `coze-studio/backend/domain/opc_xxx/internal/dal/dao.go` — Repository
  - `coze-studio/backend/api/handler/coze/opc_xxx_handler.go` — Handler
  - `flutter_client/lib/src/features/xxx/data/models/xxx_models.dart` — fromJson
  - `flutter_client/lib/src/features/xxx/data/xxx_repository.dart` — Repository
  - `flutter_client/lib/src/features/xxx/state/xxx_providers.dart` — Provider 改造
- Mock 移除状态: ✅ 已全部移除 / ❌ 剩余 X 处
```

---

## 步骤 4：开发完成度验证（≥ 95%）

开发完成后，逐项验证代码实现的完成度。

### 4.1 后端代码检查清单

| # | 检查项 | 验证方法 |
|---|--------|---------|
| 1 | Entity struct 定义 | 检查 `domain/opc_xxx/entity/` 文件存在且字段与 DDL 一致 |
| 2 | Repository 接口 | 检查 `domain/opc_xxx/internal/dal/dao.go` 接口方法完整 |
| 3 | Service 实现 | 检查 `domain/opc_xxx/service.go` 或 `application/opc_xxx/` |
| 4 | Handler 实现 | 检查 `api/handler/coze/opc_xxx_handler.go` |
| 5 | 路由注册 | 检查 `api/router/coze/opc_routes.go` 中路由已添加 |
| 6 | 错误码定义 | 检查错误码常量已定义 |

### 4.2 前端代码检查清单

| # | 检查项 | 验证方法 |
|---|--------|---------|
| 1 | Model fromJson | grep 检查所有 Model 类包含 `factory Xxx.fromJson` |
| 2 | Repository 接口 | 检查 `data/xxx_repository.dart` 接口定义 |
| 3 | ApiRepository | 检查 HTTP 调用实现 |
| 4 | Provider 改造 | grep 确认无 `_getMock` 方法调用 |
| 5 | 零 Mock 数据 | grep 确认无硬编码 Mock 数据 |
| 6 | 用户信息接入 | grep 确认无 `const spaceId = '1'` 或 `const userId = '1'`，必须使用 `opcSpaceIdProvider`/`opcUserIdProvider` |

### 4.3 计算完成度

```
完成度 = 已通过检查项 / 总检查项 × 100%
```

**质量门禁**: **≥ 95%**

- **≥ 95%**: 通过，进入步骤 5
- **< 95%**: 列出未通过项，修复后重新验证

### 📝 阶段总结 4

更新 `progress.md`，追加：

```markdown
## 步骤 4 完成 — 开发完成度验证
- 后端检查: X/6 通过
- 前端检查: X/5 通过
- 总完成度: XX% (XX/11)
- 门禁结果: ✅ 通过 / ❌ 未通过
- 未通过项: [列表或“无”]
```

---

## 步骤 5：创建 API 测试脚本

在 issue 目录下创建 API 测试脚本：

```
.claude/epics/OPC_OS_HTML_to_Flutter_Migration/issues/<issue_number>/test_api.sh
```

### 5.1 脚本模板

参考已有脚本（`issues/001/test_opc_home_api.sh`、`issues/346/test_api.sh`），新脚本必须包含：

```bash
#!/bin/bash
# Issue #<NNN> — <页面名称> API 测试
# Usage: bash <script_path> [--host <url>] [--space <id>]

BASE_URL="${1:-http://127.0.0.1:8888}"
PASS=0; FAIL=0; TOTAL=0

# 颜色定义...
# run_test() 函数...

# === 测试用例 ===
# 1. 正常请求（200 + code=0）
# 2. 响应字段完整性检查（data 中关键字段存在）
# 3. 空数据场景（返回空数组而非 null）
# 4. 错误参数（400 错误码）
# 5. 不存在资源（404 错误码）

# === 汇总 ===
echo "通过: $PASS / $TOTAL, 失败: $FAIL"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```

### 5.2 测试覆盖要求

每个 API endpoint 至少包含以下测试：
- **正常请求**: 验证 `code=0` + 关键 `data` 字段存在
- **响应结构**: 验证子对象字段（与 Flutter Model 的 fromJson 字段一一对应）
- **边界条件**: 空数据、无效参数、不存在的资源
- **字段类型**: 验证关键字段的类型（string/number/array/object）

### 📝 阶段总结 5

更新 `progress.md`，追加：

```markdown
## 步骤 5 完成 — API 测试脚本
- 脚本路径: issues/<NNN>/test_api.sh
- 测试用例数: X 个
- 覆盖 API: [GET /api/opc/v1/xxx, POST /api/opc/v1/yyy]
- 覆盖场景: 正常请求 / 字段完整性 / 空数据 / 错误参数 / 404
```

---

## 步骤 6：编译验证 + 运行 API 测试

> **⚠️ 强制要求**: 本步骤必须**实际执行** test_api.sh 脚本并确认**全部测试通过 (PASS=TOTAL, FAIL=0)**。
> 仅编译通过不算完成此步骤。必须启动后端服务器，运行 `bash issues/<NNN>/test_api.sh`，看到 `ALL PASSED` 输出后才能进入步骤 7。
> 如果服务器已在运行（`lsof -i :8888` 有输出），可直接运行测试脚本，无需重启。

### 6.1 OPC 域编译检查

// turbo
```bash
bash coze-studio/scripts/coze-manage.sh check-opc
```

如果编译失败，修复编译错误后重新执行。

### 6.2 完整编译（可选，如果 check-opc 通过）

```bash
bash coze-studio/scripts/coze-manage.sh build
```

### 6.3 DDL 执行和种子数据

如果 issue 涉及新表，必须在启动服务前执行 DDL 和插入种子数据。
如果 issue 是聚合查询（如 briefing），虽然不新建表，但**必须确保依赖表中有当日测试数据**，否则 API 返回空数据无法验证字段完整性。
插入测试数据时，`created_at` 使用 `UNIX_TIMESTAMP(NOW())*1000` 确保是今日数据。

**获取数据库连接信息**：
```bash
bash coze-studio/scripts/coze-manage.sh status
```
从输出中获取 MySQL host/port/user/database，密码从 `coze-studio/docker/docker-compose-debug.yml` 中的 `MYSQL_PASSWORD` 获取（默认 `coze123`）。

**执行 DDL**：
```bash
mysql -h 127.0.0.1 -P 3306 -u coze -pcoze123 opencoze -e "<CREATE TABLE ...>"
```

**插入种子数据**：
```bash
mysql -h 127.0.0.1 -P 3306 -u coze -pcoze123 opencoze -e "<INSERT INTO ...>"
```

> **⚠️ MySQL JSON 列注意事项**: JSON 类型列不接受空字符串 `""`，必须传 `NULL` 或有效 JSON。Go Model 中 JSON 列应使用 `*string` 指针类型，DAO 层空字符串转为 `nil`。

### 6.4 启动服务并运行测试

启动后端服务：
```bash
bash coze-studio/scripts/coze-manage.sh start
```

如果端口被占用，先杀掉旧进程：
```bash
lsof -ti:8888 | xargs kill -9 2>/dev/null; sleep 2
bash coze-studio/scripts/coze-manage.sh start
```

等待服务就绪后，运行刚创建的测试脚本：

// turbo
```bash
bash .claude/epics/OPC_OS_HTML_to_Flutter_Migration/issues/<issue_number>/test_api.sh
```

### 6.5 测试结果判定

- **全部通过 (PASS=TOTAL)**: 进入步骤 7
- **存在失败**: 分析失败原因，修复后重新运行测试
  - 编译错误 → 修复 Go 代码 → 重新 `check-opc`
  - API 返回 500 → 查看服务日志定位根因（常见：表不存在、JSON列空值）
  - API 返回错误码 → 检查 Handler/Service 逻辑
  - 字段缺失 → 检查 Entity/Response 映射

### 📝 阶段总结 6

更新 `progress.md`，追加：

```markdown
## 步骤 6 完成 — 编译验证 + API 测试
- check-opc: ✅ 通过 / ❌ 失败
- build: ✅ 通过 / ❌ 失败 / ➖ 跳过
- API 测试结果: PASS X / TOTAL X, FAIL X
- 失败用例: [列表或“无”]
- 修复记录: [修复了什么问题，或“无需修复”]
```

---

## 步骤 7：前后端字段对接确认

最终确认 Flutter 前端 Model 与后端 API Response 的字段完全一致。

### 7.1 字段对照表生成

对每个 API endpoint，生成对照表：

| Flutter Model 字段 | 类型 (Dart) | API Response 字段 | 类型 (Go/JSON) | fromJson 映射 | 状态 |
|-------------------|-------------|-------------------|----------------|--------------|------|
| `id` | `int` | `id` | `int64` | `json['id']` | ✅ |
| `name` | `String` | `name` | `string` | `json['name']` | ✅ |
| `createdAt` | `DateTime` | `created_at` | `int64` (ms) | `DateTime.fromMillisecondsSinceEpoch(json['created_at'])` | ✅ |
| ... | ... | ... | ... | ... | ... |

### 7.2 对接检查项

- [ ] **用户信息传递**: Provider 中的 `spaceId`/`userId` 通过 `opcSpaceIdProvider`/`opcUserIdProvider` 获取，正确传递给 Repository 和 API
- [ ] **字段数量一致**: Flutter Model 字段数 = API Response 字段数（不含内部计算字段）
- [ ] **字段名映射正确**: snake_case (API) → camelCase (Dart) 映射无误
- [ ] **类型兼容**: `int64` → `int`、`string` → `String`、`[]T` → `List<T>`、嵌套对象 → 子 Model
- [ ] **时间字段**: 后端 `int64` 毫秒时间戳 → 前端 `DateTime.fromMillisecondsSinceEpoch()`
- [ ] **可空字段**: 后端 `omitempty` / nullable → 前端 `T?` 可空类型
- [ ] **枚举映射**: 后端 `int/string` → 前端 `enum` 的 fromApi 方法
- [ ] **默认值**: 后端缺失字段 → 前端 fromJson 中有合理默认值
- [ ] **嵌套对象**: 每个子对象都有独立的 `fromJson` 且字段完整

### 7.3 最终判定

- **全部 ✅**: Issue 完成，更新进度文件状态为 `✅ done`
- **存在 ❌**: 列出不一致项，修复后重新验证

### 📝 阶段总结 7

更新 `progress.md`，追加：

```markdown
## 步骤 7 完成 — 前后端字段对接
- 对照表字段数: X 个
- 全部一致: ✅ / ❌
- 不一致项: [列表或“无”]
- 修复记录: [修复了什么，或“无需修复”]
```

---

## 步骤 8：更新进度

更新进度文件 `.claude/epics/OPC_OS_HTML_to_Flutter_Migration/production-ready-progress.md`：

1. 将该 issue 状态改为 `✅ done`
2. 填写完善日期
3. 填写备注（API 数量、DDL 表数量、Model fromJson 数量）
4. 更新统计区的已完善数和完善率

---

## 异常处理

| 异常场景 | 处理方式 |
|---------|---------|
| issue 规格完成度 < 95% | 使用 `@issue-production-ready <N>` 补全，重新检测 |
| 开发完成度 < 95% | 列出缺失项，逐项修复 |
| 编译失败 | 修复 Go 编译错误，重新 `check-opc` |
| API 测试失败 | 分析失败原因，修复代码，重新测试 |
| API 返回 500 (Table not exist) | 先执行 DDL 创建表 + 插入种子数据，再重新测试 |
| API 返回 500 (Invalid JSON) | Go Model 中 JSON 列改为 `*string` 指针，DAO 层空字符串转 nil |
| 前后端字段不一致 | 修复 fromJson 映射或 API Response，重新对照 |
| 数据库服务未启动 | 提示用户启动 MySQL/Redis 等依赖服务 |
| 数据库连接失败 | 运行 `bash coze-studio/scripts/coze-manage.sh status` 获取正确连接信息 |
| 端口被占用 | `lsof -ti:8888 \| xargs kill -9` 后重启 |

---

## 快速参考

| 命令 | 用途 |
|------|------|
| `@issue-production-ready next` | 获取下一个待处理 issue |
| `@issue-production-ready <N>` | 完善指定 issue 规格 |
| `@issue-production-ready status` | 查看整体进度 |
| `@issue-start <N>` | 开始开发指定 issue |
| `bash coze-studio/scripts/coze-manage.sh check-opc` | OPC 域编译检查 |
| `bash coze-studio/scripts/coze-manage.sh build` | 完整编译 |
| `bash coze-studio/scripts/coze-manage.sh start` | 启动服务 |
| `bash coze-studio/scripts/coze-manage.sh status` | 状态总览 |
