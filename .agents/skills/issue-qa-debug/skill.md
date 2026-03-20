---
name: issue-qa-debug
description: 通用 Issue 诊断与 Debug 工作流。针对指定 issue，以测试工程师角色全面诊断功能完整性、代码质量、数据链路正确性，定位并修复问题。适用于后端 Go、前端 Flutter、纯后端集成等各类 issue。
allowed-tools: Bash, Read, Write, LS
argument-hint: "[issue_number]"
disable-model-invocation: true
---

# Issue 诊断与 Debug 工作流

> **角色**: 测试工程师 (QA Engineer)
> **目标**: 针对指定 issue，全面诊断功能完整性，定位问题根因并修复，确保代码质量达到生产标准。

## 测试工程师职责

1. 加载 issue 规格文件，理解功能需求和验收标准
2. 判断 issue 类型，选择对应的诊断策略
3. 编译检查，确保代码可构建
4. 代码质量审查（零 Mock、零硬编码、数据链路真实性）
5. 数据库验证（如涉及）
6. API 实际调用测试（如涉及）
7. 前端对接验证（如涉及）
8. 对照验收标准逐条验证功能完整性
9. 发现问题时定位根因并修复，修复后重新验证
10. 输出诊断报告

## 关键路径

- **编译工具**: `coze-studio/scripts/coze-manage.sh`
- **Epic 目录**: `.Codex/epics/<epic_name>/`
- **Issue 规格文件**: `.Codex/epics/<epic_name>/<issue_id>-<名称>.md`
- **Issue 工作目录（强制）**: `.Codex/epics/<epic_name>/issues/<issue_id>/`（含 `progress.md`、`test_api.sh`、`ddl_and_seed.sql` 等）
- **Skill 参考文件**: `.Codex/skills/issue-production-ready/reference/`

> **🚨 工作目录规范**: 所有 issue 产出物（进度文件、测试脚本、DDL、诊断报告）必须存放在 `.Codex/epics/<epic_name>/issues/<issue_id>/` 目录下。如该目录不存在，必须先创建。

> **🚨 用户信息规范（强制）**: 所有 OPC 域 Provider 必须通过 `opcSpaceIdProvider`/`opcUserIdProvider` 获取 `spaceId`/`userId`，**禁止硬编码**。
> 统一入口: `flutter_client/lib/src/features/shared/state/opc_user_providers.dart`

---

## 🚨 强制规范

### 数据库驱动原则

**所有 CRUD 功能必须由数据库支撑，禁止在代码中硬编码 Mock 数据。**

1. **新建表 → 必须先写 DDL**: 在 `.Codex/epics/<epic_name>/issues/<issue_id>/ddl_and_seed.sql` 中定义
2. **CRUD 测试 → 必须先导入种子数据**: DDL + INSERT 种子数据一起写入 SQL 文件
3. **导入方式 → 只用 `coze-manage.sh`**: `bash coze-studio/scripts/coze-manage.sh db-migrate .Codex/epics/<epic_name>/issues/<issue_id>/ddl_and_seed.sql`
4. **禁止 Docker 操作**: 不要研究 docker-compose.yml、不要用 `docker exec`，本地 MySQL 通过 `coze-manage.sh` 管理
5. **DDL 同步**: 新表需同步到 `coze-studio/docker/atlas/migrations/20250601000000_opc_tables.sql`

### 禁止硬编码清单

| 场景 | 禁止 | 正确做法 |
|------|------|----------|
| 用户信息 | ❌ `spaceId: '1'` | ✅ `ref.watch(opcSpaceIdProvider)` |
| 测试数据 | ❌ Go 代码返回 hardcoded JSON | ✅ 从数据库查询 |
| 数据库配置 | ❌ `mysql -h 127.0.0.1 -u root` | ✅ `coze-manage.sh db-debug` |
| API Base URL | ❌ `http://localhost:8888` 在代码中 | ✅ 从环境配置读取 |

### coze-manage.sh 命令速查

| 命令 | 用途 | 何时用 |
|------|------|--------|
| `check-opc` | OPC 域编译检查 | 每次改后端代码后 |
| `build` | 完整编译 | 生成可执行文件 |
| `start` | 编译+启动服务 | 需要运行时测试时 |
| `db-debug` | 数据库连接+表状态 | 检查数据库是否正常 |
| `db-migrate [file]` | 执行 DDL 迁移 | 导入 DDL + 种子数据 |
| `db-info` | 显示连接信息 | 查看数据库配置 |
| `status` | 全局状态 | 检查环境是否就绪 |

---

## 步骤 1：加载 Issue 上下文

> **🚨 强制要求**: 必须**完整读取 issue 规格文件**，这是诊断的核心上下文来源。

### 1.1 查找并读取 Issue 规格文件（必须，不可跳过）

在所有 epic 目录下搜索对应 issue 的规格文件：

```bash
find .Codex/epics/ -name "$ARGUMENTS-*" -o -name "$ARGUMENTS.*" | head -10
```

**必须从规格文件中提取**：
- 功能描述、验收标准
- 涉及的技术栈（Go 后端 / Flutter 前端 / 数据库 / 第三方集成等）
- API Endpoint 定义（如有）
- 修改文件清单
- 已知问题和调试记录

### 1.2 读取 Issue 进度文件（如存在）

在对应 epic 目录下查找 issue 工作目录：
```bash
find .Codex/epics/ -path "*/issues/$ARGUMENTS" -type d 2>/dev/null
```

进度文件固定路径：`.Codex/epics/<epic_name>/issues/$ARGUMENTS/progress.md`

提取：当前完成阶段、已创建文件清单、API 测试结果、历史 Debug 记录。

> 如果 issue 工作目录不存在，创建它：`mkdir -p .Codex/epics/<epic_name>/issues/$ARGUMENTS/`

### 1.3 判断 Issue 类型（结构化规则）

按以下规则自动判断 issue 类型，**不依赖主观推断**：

#### 判断规则（按优先级从上到下匹配）

| 信号 | 检测方法 | 结论 |
|------|---------|------|
| 规格文件含 `## DDL` 或 `CREATE TABLE` | grep 规格文件 | → 涉及数据库 |
| 修改文件含 `.go` | 检查修改文件清单 | → 涉及 Go 后端 |
| 修改文件含 `.dart` | 检查修改文件清单 | → 涉及 Flutter 前端 |
| 规格文件含 `API` 或 `/api/opc/` | grep 规格文件 | → 涉及 API |
| 规格文件含 `Coze`、`Agent`、`工作流`、`crossdomain` | grep 规格文件 | → 涉及第三方集成 |
| 修改文件含 `.env`、`docker`、`deploy`、`scripts/` | 检查修改文件清单 | → 涉及配置/基础设施 |

```bash
# 自动检测信号
echo "=== Issue 类型检测 ==="
SPEC_FILE="<规格文件路径>"
echo -n "数据库: "; grep -c "DDL\|CREATE TABLE" "$SPEC_FILE" 2>/dev/null || echo "0"
echo -n "API: "; grep -c "/api/opc/\|API Endpoint" "$SPEC_FILE" 2>/dev/null || echo "0"
echo -n "集成: "; grep -c "Coze\|Agent\|工作流\|crossdomain" "$SPEC_FILE" 2>/dev/null || echo "0"
```

#### 类型 → 步骤映射

| Issue 类型 | 组合信号 | 适用步骤 |
|-----------|---------|---------|
| **全栈 CRUD** | 数据库 + Go + Flutter + API | 全部步骤 |
| **纯后端** | Go + API，无 `.dart` | 步骤 2-6, 8（跳过步骤 7） |
| **纯前端** | Flutter，无 `.go` | 步骤 2, 3.3, 7, 8（跳过后端和数据库） |
| **后端集成** | Go + 第三方集成信号 | 步骤 2, 3.2, 6, 8（重点集成链路） |
| **配置/基础设施** | 配置信号 | 步骤 2, 8（重点配置正确性） |

### 1.4 读取关键代码文件

根据规格文件中的**修改文件清单**，逐个读取涉及的代码文件。

如果规格文件未列出修改文件清单，按 issue 类型推断：

**后端 Go**（如涉及）:
- `coze-studio/backend/domain/opc_<domain>/entity/*.go`
- `coze-studio/backend/domain/opc_<domain>/internal/dal/dao.go`
- `coze-studio/backend/domain/opc_<domain>/service.go`
- `coze-studio/backend/api/handler/coze/opc_<domain>_*.go`

**前端 Flutter**（如涉及）:
- `flutter_client/lib/src/features/<feature>/data/models/*_models.dart`
- `flutter_client/lib/src/features/<feature>/data/*_repository.dart`
- `flutter_client/lib/src/features/<feature>/state/*_providers.dart`
- `flutter_client/lib/src/features/<feature>/presentation/widgets/*.dart`

### 1.5 读取 DDL 和测试脚本（如存在）

固定路径：
```
.Codex/epics/<epic_name>/issues/$ARGUMENTS/ddl_and_seed.sql
.Codex/epics/<epic_name>/issues/$ARGUMENTS/test_api.sh
```

### 📝 上下文摘要输出格式

```
Issue #<id> — <名称>
规格文件: <path>
Issue 类型: <全栈 CRUD / 纯后端 / 纯前端 / 后端集成 / 配置>
涉及技术栈: <Go / Flutter / MySQL / Coze Agent / ...>
修改文件: <file list>
验收标准: <N> 条
当前状态: <status>
诊断策略: 执行步骤 <X, Y, Z>（跳过步骤 <A, B>）
```

### 🚦 诊断计划确认（必须）

> 步骤 1 完成后，**必须先输出诊断计划并等待用户确认**，再执行后续步骤。避免跑偏浪费时间。

输出格式：
```
📋 诊断计划 — Issue #<id>

类型判断依据:
  - [信号1]: 检测到 N 处匹配
  - [信号2]: 未检测到

将执行的步骤:
  ✅ 步骤 2: 编译检查（后端 Go）
  ✅ 步骤 3: 代码质量审查（后端数据流）
  ⏭️ 步骤 4: 数据库验证 — 跳过（无 DDL）
  ⏭️ 步骤 5: 字段一致性 — 跳过（非全栈 CRUD）
  ✅ 步骤 6: 生成 test_api.sh 并执行
  ⏭️ 步骤 7: 前端对接 — 跳过（无 .dart 文件）
  ✅ 步骤 8: 验收标准逐条验证

是否按此计划执行？
```

---

## 步骤 2：编译检查

> **🚨 Early Exit**: 编译失败则停在此步骤，先修复编译错误再继续后续诊断。

### 2.1 后端编译（如涉及 Go 代码）

```bash
bash coze-studio/scripts/coze-manage.sh check-opc
```

### 2.2 前端编译（如涉及 Flutter 代码）

```bash
cd flutter_client && dart analyze lib/src/features/<feature>/ 2>&1 | head -50
```

如有 error/warning，先修复再继续。

---

## 步骤 3：代码质量审查

> **⚠️ 仅靠关键词 grep 不够！必须追踪代码实现逻辑，验证数据流链路真实性。**

### 3.1 关键词初筛

对规格文件中列出的所有修改文件进行扫描：

```bash
grep -rn "mock\|Mock\|hardcode\|fake\|stub\|TODO\|FIXME\|_getMock\|sampleData\|dummyData" \
  <涉及的文件/目录>
```

### 3.2 后端数据流链路审查（如涉及 Go 后端）

| 审查层级 | 审查内容 | 判定标准 |
|---------|---------|---------|
| **Handler** | 确认调用 Service 接口方法 | ❌ 直接构造 JSON 返回 = Mock |
| **Service** | 确认调用 Repository/DAO 方法 | ❌ 直接构造 Entity/VO 返回 = Mock |
| **DAO** | 确认使用 GORM/SQL 查询 | ❌ 直接 return 固定数据 = Mock |
| **Application** | 确认注入真实 `*gorm.DB` | ❌ 注入 mock 实例 = Mock |
| **错误降级** | DB 查询失败时的处理 | ❌ 吞掉 error 返回空数组 = 掩盖问题 |
| **调试日志** | 是否有残留的调试日志 | ⚠️ `fmt.Println`、临时 `logs.CtxInfof` 应清理 |
| **硬编码 ID** | 是否有硬编码的 spaceId/userId/agentId | ❌ 硬编码 = 禁止 |

### 3.3 前端数据流链路审查（如涉及 Flutter 前端）

| 审查层级 | 审查内容 | 判定标准 |
|---------|---------|---------|
| **Widget** | 确认 `ref.watch(xxxProvider)` | ❌ 直接构造 Model 实例 = Mock |
| **Provider** | 确认调用 Repository 方法 | ❌ 直接返回构造的 Model = Mock |
| **Repository** | 确认 `dio.post/get` 调用 | ❌ 不调用 API 直接返回 = Mock |
| **spaceId/userId** | 确认来自 `opcSpaceIdProvider`/`opcUserIdProvider` | ❌ 硬编码字符串 = 禁止 |

### 3.4 输出审查表

```
| 层级 | 文件 | 关键行 | 验证结果 |
|------|------|--------|----------|
| <层级> | <文件:行号> | <关键代码> | ✅/❌ |
```

---

## 步骤 4：数据库验证（如涉及数据库）

> 跳过条件：issue 不涉及数据库表操作（如纯后端集成、纯前端 UI）

### 4.1 表存在性检查

```bash
bash coze-studio/scripts/coze-manage.sh db-debug
```

### 4.2 种子数据检查

```bash
mysql -u coze -pcoze123 -h 127.0.0.1 -P 3306 opencoze \
  -e "SELECT id, <key_fields> FROM <table> WHERE deleted_at=0 LIMIT 10;"
```

### 4.3 DDL 与 Atlas 同步检查

确认表结构已同步到 `coze-studio/docker/atlas/migrations/20250601000000_opc_tables.sql`。

### 4.4 数据完整性约束验证

| 检查项 | 方法 | 判定标准 |
|--------|------|---------|
| **唯一约束** | 检查 DDL 中 `UNIQUE KEY` | 业务唯一字段应有唯一约束 |
| **索引覆盖** | 检查 `KEY`/`INDEX` | 高频查询字段应有索引 |
| **软删除索引** | 检查 `deleted_at` 索引 | 使用软删除的表应在联合索引中 |
| **外键逻辑一致性** | 检查关联字段 | 引用的表和数据是否存在 |

---

## 步骤 5：全链路字段一致性验证（如涉及全栈 CRUD）

> 跳过条件：issue 不涉及 DDL↔Go↔Dart 全链路

逐字段比对：
1. **DDL** → **Go Entity** → **DAO Model** → **API Response DTO** → **Dart fromJson**

检查要点：
- 每个 DDL 列在 Go Entity 和 DAO Model 中都有对应字段
- JSON tag 与 DDL 列名一致
- Go DTO 的 JSON tag 与 Dart fromJson 的 key 一致
- 类型匹配：`int64`↔`int`、`string`↔`String`、`bool`↔`bool`

### 字段提取辅助命令

```bash
# 提取 Go Entity JSON tag
grep -oP 'json:"(\w+)"' <entity_file> | grep -oP '"\w+"' | tr -d '"' | sort

# 提取 Dart fromJson key
grep -oP "json\['(\w+)'\]" <dart_model_file> | grep -oP "'(\w+)'" | tr -d "'" | sort
```

---

## 步骤 6：生成测试脚本并执行 API 测试（不可跳过）

> **🚨 强制要求**: 无论 issue 类型如何，只要涉及后端 API，必须生成 `test_api.sh` 并实际执行。测试脚本是诊断的核心产出物之一。

### 6.1 确认服务运行

```bash
lsof -i :8888
# 如未运行：
bash coze-studio/scripts/coze-manage.sh start &
sleep 3
```

### 6.2 生成 test_api.sh（必须）

根据规格文件中的 API Endpoint / 功能点，在 issue 工作目录生成测试脚本：

**输出路径**: `.Codex/epics/<epic_name>/issues/$ARGUMENTS/test_api.sh`

测试脚本必须覆盖：
- **正常场景**: 每个 API Endpoint 至少一个正常调用，验证 `code=0`、Response 字段完整
- **错误场景**: 缺少必填参数、无效资源 ID、关联数据不存在
- **边界场景**: 空字符串参数、极大/极小分页参数
- **集成链路**（如涉及）: 第三方服务调用验证（如 Agent 执行、工作流触发）

脚本必须包含失败计数和汇总逻辑：
```bash
#!/bin/bash
# Issue #<id> — <名称> API 测试脚本
# 生成日期: <date>
BASE_URL="http://127.0.0.1:8888"
PASS=0
FAIL=0
TOTAL=0

run_test() {
  local name="$1"
  local expected_code="$2"
  shift 2
  TOTAL=$((TOTAL + 1))
  echo "--- [$TOTAL] $name ---"
  RESPONSE=$(curl -s -w "\n%{http_code}" "$@")
  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  BODY=$(echo "$RESPONSE" | sed '$d')
  CODE=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code','N/A'))" 2>/dev/null || echo "N/A")
  if [ "$CODE" = "$expected_code" ]; then
    echo "✅ PASS (code=$CODE)"
    PASS=$((PASS + 1))
  else
    echo "❌ FAIL (expected code=$expected_code, got code=$CODE, http=$HTTP_CODE)"
    echo "Response: $BODY" | head -5
    FAIL=$((FAIL + 1))
  fi
  echo ""
}

echo "=== 1. 正常场景 ==="
run_test "<接口描述>" "0" \
  -X POST "$BASE_URL/api/opc/v1/<endpoint>" \
  -H "Content-Type: application/json" \
  -d '<request_body>'

echo "=== 2. 错误场景 ==="
run_test "缺少必填参数" "<expected_error_code>" \
  -X POST "$BASE_URL/api/opc/v1/<endpoint>" \
  -H "Content-Type: application/json" \
  -d '{}'

echo "=== 3. 边界场景 ==="
# ...

echo "==============================="
echo "测试结果: $PASS/$TOTAL 通过, $FAIL 失败"
if [ $FAIL -gt 0 ]; then
  echo "❌ 存在失败用例，需要修复后重新执行"
  exit 1
else
  echo "✅ 全部通过"
  exit 0
fi
```

### 6.3 执行测试脚本（必须）

```bash
bash .Codex/epics/<epic_name>/issues/$ARGUMENTS/test_api.sh
```

逐条检查输出结果，记录每个用例的通过/失败状态。

> **🚨 硬性要求: API 测试通过率必须 100%**。任何失败用例都必须立即定位根因并修复，修复后重新执行测试脚本，直到全部通过。不允许带着失败用例进入后续步骤。

### 6.4 写入后验证（如涉及 CRUD）

每次 POST/PUT/DELETE 后：
1. 查询数据库确认写入成功
2. 调用 GET 接口确认新数据可读

---

## 步骤 7：前端对接验证（如涉及 Flutter）

> 跳过条件：issue 不涉及 Flutter 前端

### 7.1 基础对接检查

| 检查项 | 验证方法 |
|--------|----------|
| Repository 接口定义 | 方法签名与规格文件一致 |
| ApiRepository 实现 | 调用正确的 API Endpoint |
| Data Provider | 调用 Repository 而非 Mock |
| Widget 三态 | data/loading/error 三态完整 |

### 7.2 fromJson 健壮性检查

| 检查项 | 判定标准 |
|--------|---------|
| null 字段处理 | ❌ 直接 `as Type` 无空安全 = 崩溃风险 |
| 类型不匹配兜底 | ❌ 无 `toString()` 转换 = 崩溃 |
| 枚举值兜底 | ❌ 无 `default` 分支 = 崩溃 |
| 嵌套对象 null | ❌ 直接 `fromJson(map['nested'])` 无 null 检查 = 崩溃 |

### 7.3 Provider 生命周期检查

| 检查项 | 判定标准 |
|--------|---------|
| autoDispose 合理性 | 页面级应 autoDispose，跨页面共享不应 |
| invalidate 时机 | 写操作后未 invalidate = 数据不一致 |
| 循环依赖 | Provider 之间循环依赖 = 运行时异常 |

---

## 步骤 8：验收标准逐条验证

> **核心步骤**: 不仅验证"做对了"，还验证"做全了"。

### 8.1 验收标准逐条勾选

从规格文件中提取所有验收标准，逐条检查：

| # | 验收标准 | 代码位置 | 实现状态 | 备注 |
|---|---------|---------|---------|------|
| 1 | 来自规格文件 | 对应文件:行号 | ✅/⚠️/❌ | |

状态说明：
- ✅ **已实现且验证通过**
- ⚠️ **部分实现** — 有遗漏
- ❌ **未实现** — 规格要求但代码中找不到

### 8.2 缺失功能汇总

将所有 ⚠️ 和 ❌ 项汇总，标注：
- 是否为本期必须（规格文件明确要求）
- 是否可延期
- 是否需要升级到 PM Review

---

## 步骤 9：问题修复（如有）

### 9.1 定位根因

- **编译错误**: 从错误信息定位文件和行号
- **API 错误**: 查看后端日志，常见原因：表不存在/JSON 空值/空指针/路由未注册
- **集成错误**: 检查第三方服务配置、鉴权参数、执行链路
- **字段缺失**: 对照规格文件检查映射
- **Mock 残留**: 定位并删除，替换为真实调用

### 9.2 修复后重新验证

```bash
bash coze-studio/scripts/coze-manage.sh check-opc
```

修复后重新执行相关步骤，确保问题解决且无回归。

### 9.3 回归测试记录

```markdown
#### 回归测试 — <修复描述>

**修改文件**:
- `path/to/file` — 修改描述

**重新验证的用例**:
| # | 用例 | 修复前 | 修复后 | 回归验证 |
|---|------|-------|--------|---------|
| 1 | 直接相关的用例 | ❌ | ✅ | — |
| 2 | 可能受影响的关联用例 | ✅ | ✅ | ✅ 无回归 |
```

### 9.4 环境恢复

```bash
# 如果测试导致服务异常，重启服务
lsof -ti:8888 | xargs kill -9 2>/dev/null
bash coze-studio/scripts/coze-manage.sh start &
sleep 3
```

---

## 步骤 10：输出诊断报告

在 `.Codex/epics/<epic_name>/issues/$ARGUMENTS/progress.md` 末尾追加（如文件不存在则创建）：

```markdown
## QA 诊断报告 — <日期>

> **Issue 类型**: <全栈 CRUD / 纯后端 / 纯前端 / 后端集成 / 配置>
> **诊断目标**: <简述>

### 1. 编译检查
| 检查项 | 结果 |
|--------|------|
| 后端编译 | ✅/❌/N/A |
| 前端编译 | ✅/❌/N/A |

### 2. 代码质量审查
| 层级 | 文件 | 关键行 | 验证结果 |
|------|------|--------|----------|

### 3. 运行时测试
| 测试场景 | 预期 | 实际 | 结果 |
|---------|------|------|------|

### 4. 验收标准验证
| # | 验收标准 | 实现状态 | 备注 |
|---|---------|---------|------|

### 5. 诊断结论
**状态**: ✅ 通过 / ⚠️ 有条件通过 / ❌ 不通过

- 发现问题: <列表>
- 已修复: <列表>
- 待跟进: <列表>
```

---

## 异常处理

| 异常场景 | 处理方式 |
|---------|---------|
| 编译错误 — 类型不匹配 | 对照字段类型定义，修正不一致 |
| 编译错误 — 未定义引用 | 检查 import 路径、包名 |
| API 500 — Table not exist | `coze-manage.sh db-debug` 检查表，执行 DDL |
| API 500 — nil pointer | 检查空指针解引用 |
| API 404 — Route not found | 检查路由注册 |
| 端口被占用 | `lsof -ti:8888 \| xargs kill -9` 后重启 |
| 数据库连接失败 | `coze-manage.sh status` 检查连接信息 |
| 找不到规格文件 | 扩大搜索范围 `find .Codex/epics/ -name "*$ARGUMENTS*"` |

---

## 快速参考

### 编译与服务

| 命令 | 用途 |
|------|------|
| `bash coze-studio/scripts/coze-manage.sh check-opc` | OPC 域编译检查 |
| `bash coze-studio/scripts/coze-manage.sh build` | 完整编译 |
| `bash coze-studio/scripts/coze-manage.sh start` | 编译+启动服务 |
| `bash coze-studio/scripts/coze-manage.sh status` | 状态总览 |

### 数据库

| 命令 | 用途 |
|------|------|
| `bash coze-studio/scripts/coze-manage.sh db-info` | 数据库连接信息 |
| `bash coze-studio/scripts/coze-manage.sh db-debug` | 数据库调试 |
| `bash coze-studio/scripts/coze-manage.sh db-migrate [file]` | DDL 迁移 |

### 测试与调试

| 命令 | 用途 |
|------|------|
| `cd flutter_client && dart analyze` | Flutter 静态分析 |
| `lsof -i :8888` | 检查服务是否运行 |
| `lsof -ti:8888 \| xargs kill -9` | 强制停止服务 |

### 参考文件

| 文件 | 用途 |
|------|------|
| `.Codex/skills/issue-production-ready/reference/api-domain-mapping.md` | API-Domain 映射 |
| `.Codex/skills/issue-production-ready/reference/ddl-schema.md` | DDL Schema 参考 |
| `.Codex/skills/issue-production-ready/reference/shared-services.md` | 共享服务层 |
| `.Codex/skills/issue-production-ready/reference/code-reuse-patterns.md` | 代码复用模式 |
| `flutter_client/lib/src/features/shared/state/opc_user_providers.dart` | 用户信息 Provider |
