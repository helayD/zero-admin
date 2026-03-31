---
name: zero-admin-qa-debug
description: Zero-Admin 后端 Issue 通用诊断与 Debug 工作流。适用于 zero-admin Go go-zero 后端 issue 的 QA 测试，涵盖编译检查、代码质量审查、API 测试、数据链路验证。
allowed-tools: Bash, Read, Write, LS, Grep, Glob, AstGrep
argument-hint: "[issue_id]"
disable-model-invocation: true
---

# Zero-Admin 后端 QA 诊断工作流（通用版）

> **角色**: 测试工程师 (QA Engineer)
> **目标**: 针对 zero-admin 后端 issue，全面诊断功能完整性、代码质量、数据链路正确性
> **适用范围**: Zero-Admin Go go-zero 后端项目（RPC + API 双层架构，GORM + MySQL + RabbitMQ）

## 测试工程师职责

1. 理解 Issue 规格和验收标准
2. 编译检查，确保代码可构建
3. 代码质量审查（零 Mock、零硬编码）
4. Bug 修复验证（如有）
5. API 实际调用测试
6. 对照验收标准逐条验证
7. 输出诊断报告

---

## 🚨 Zero-Admin 项目强制规范

### ⚠️ 数据库安全约束（最高优先级）

**数据库操作只能新增数据，禁止执行任何清理或删除操作。**

| 操作 | 允许 | 禁止 |
|------|------|------|
| SELECT | ✅ | — |
| INSERT | ✅ | — |
| UPDATE | ✅（仅测试数据修正） | — |
| DELETE | ❌ | ✅ |
| TRUNCATE | ❌ | ✅ |
| DROP | ❌ | ✅ |

**唯一索引冲突处理**：使用时间戳生成唯一值重试，禁止 DELETE 清理

### 编码规范

| 场景 | 禁止 | 正确 |
|------|------|------|
| 用户信息 | ❌ 硬编码 userId/tenantId | ✅ 从 JWT 解析或参数传入 |
| 测试数据 | ❌ hardcoded JSON | ✅ 从数据库查询 |
| 调试日志 | ❌ fmt.Println 残留 | ✅ 已清理或日志级别控制 |

### Zero-Admin 项目结构

```
zero-admin/
├── rpc/                    # gRPC 服务 (sys/ums/pms/oms/sms/cms/search)
│   └── {service}/
│       ├── internal/
│       │   ├── logic/      # 业务逻辑 (*_logic.go)
│       │   ├── svc/        # 服务上下文
│       │   └── config/    # 配置
│       └── gen/           # GORM 模型
├── api/                    # HTTP API 服务
│   └── {api}/
│       ├── internal/
│       │   ├── handler/    # HTTP Handler
│       │   ├── logic/     # 业务逻辑
│       │   └── types/     # 请求/响应类型
├── consumer/               # RabbitMQ 消费者
│   └── internal/
│       └── mq/           # 队列消费者
├── job/                    # 定时任务
├── pkg/                    # 共享包
├── makefile               # 构建命令
└── _opcos/               # 需求与进度追踪
    └── implementation-artifacts/  # Issue 规格文件
```

### 服务端口配置

| 服务 | 端口 | 配置文件 |
|------|------|---------|
| admin-api | 8888 | `api/admin/etc/admin-api.yaml` |
| front-api | 8001 | `api/front/etc/front-api.yaml` |
| sys-rpc | 8070 | `rpc/sys/etc/sys.yaml` |
| ums-rpc | 8081 | `rpc/ums/etc/ums.yaml` |
| pms-rpc | 8082 | `rpc/pms/etc/pms.yaml` |
| oms-rpc | 8083 | `rpc/oms/etc/oms.yaml` |
| sms-rpc | 8084 | `rpc/sms/etc/sms.yaml` |
| cms-rpc | 8086 | `rpc/cms/etc/cms.yaml` |
| search-rpc | 8087 | `rpc/search/etc/search.yaml` |

---

## 步骤 1：理解 Issue 上下文

### 1.1 定位 Issue 规格文件

Issue 规格文件位于 `_opcos/implementation-artifacts/` 目录：

```bash
# 方式1：按 issue ID 搜索（推荐）
find _opcos/implementation-artifacts/ -name "*$ARGUMENTS*" -type f 2>/dev/null

# 方式2：列出所有 implementation-artifacts
ls -la _opcos/implementation-artifacts/

# 方式3：在 rpc/ 或 api/ 下查找相关实现
find rpc/ -name "*.go" | head -10
find api/ -name "*.go" | head -10
```

### 1.2 读取 Issue 规格

读取找到的 Issue 规格文件，提取：
- **功能描述**：要实现/修复什么
- **验收标准**：如何判定完成
- **涉及文件**：哪些代码需要检查
- **API Endpoint**：如有的话

### 1.3 确定涉及技术栈

| 信号 | 结论 |
|------|------|
| `.go` 文件 + `rpc/` | → 涉及 RPC 服务 |
| `.go` 文件 + `api/` | → 涉及 HTTP API |
| `.go` 文件 + `consumer/` | → 涉及 MQ 消费者 |
| `gen/model/` | → 涉及 GORM 模型 |
| `sysclient/` | → 涉及 gRPC 客户端调用 |
| `types/` | → 涉及 API 类型定义 |

### 1.4 输出上下文摘要

```
Issue: <名称/路径>
涉及技术栈: Go / RPC / API / GORM / MySQL / RabbitMQ
关键文件: <列表>
验收标准: <N> 条
诊断策略: 执行步骤 <X, Y, Z>
```

---

## 步骤 2：编译检查

> **🚨 Early Exit**: 编译失败则停在此步骤

```bash
# 方式1：使用 Makefile（推荐）
make build

# 方式2：直接编译所有服务
go build ./...

# 方式3：针对特定服务
go build -o target/sys-rpc/sys-rpc ./rpc/sys/sys.go
go build -o target/admin-api/admin-api ./api/admin/admin.go
go build -o target/oms-rpc/oms-rpc ./rpc/oms/oms.go
```

**常见错误**：
- `undefined` → 检查 import 路径
- `未导入` → 检查包名
- `类型不匹配` → 对照定义修正

---

## 步骤 3：代码质量审查

### 3.1 关键词初筛

```bash
grep -rn "mock\|Mock\|hardcode\|TODO\|FIXME\|sampleData\|dummyData" \
  rpc/ api/ consumer/ job/ 2>/dev/null
```

### 3.2 Mock 数据识别

| 层级 | 判定标准 |
|------|---------|
| Handler/Logic | ❌ 直接构造 JSON 返回 |
| Service | ❌ 直接构造 Entity 返回 |
| DAO/ORM | ❌ 直接 return 固定数据 |

### 3.3 硬编码识别

```bash
# 查找硬编码的 userId/tenantId/menuId
grep -rn "userId\s*[:=]\s*[0-9]\|tenantId\s*[:=]\s*[0-9]" \
  rpc/ api/ consumer/ 2>/dev/null
```

### 3.4 调试日志检查

```bash
# 查找残留的调试日志
grep -rn "fmt\.Print\|log\.\|logrus\." \
  rpc/ api/ consumer/ 2>/dev/null
```

### 3.5 RabbitMQ 消费模式检查

```bash
# 检查 auto-ack 模式（问题模式）
grep -rn "auto-ack\|ConsumeSimple\|context.Background()" \
  consumer/ pkg/mq/ 2>/dev/null

# 检查是否使用手动 ACK
grep -rn "ConsumeSimpleWithAck\|channel\.Ack\|nack" \
  consumer/ pkg/mq/ 2>/dev/null
```

---

## 步骤 4：gRPC/RPC 路由验证

### 4.1 查找 RPC 服务定义

```bash
# 查看 RPC 服务 proto 文件
ls -la rpc/*/proto/

# 查看 RPC 服务入口
grep -rn "main" rpc/*/*.go | head -10
```

### 4.2 确认 RPC 服务端口配置

```bash
# 查看 RPC 服务配置
cat rpc/*/etc/*.yaml | grep -A5 "ListenOn:"
```

### 4.3 确认 Logic 层方法

```bash
# 查找 Logic 方法定义
grep -n "func.*Logic" rpc/*/internal/logic/*/*.go
```

---

## 步骤 5：Bug 修复验证

### 5.1 验证方法

根据 Issue 规格中记录的 Bug，使用 grep 验证修复：

```bash
# 验证 Bug 已修复：grep 搜索问题代码，应返回空
grep -n "<问题代码>" rpc/*/internal/logic/*.go

# 示例：验证某个错误的调用已修复
grep -n "EquipmentVendorBatchDelete" rpc/*/internal/logic/*/*.go
```

### 5.2 输出验证结果

```
| Bug 描述 | 验证方法 | 结果 |
|---------|---------|------|
| XXX 调用错误模型 | grep XXX | ✅ 已修复（未找到） |
| XXX 残留调试日志 | grep XXX | ❌ 未清理（找到 N 处） |
```

---

## 步骤 6：API 测试

### 6.1 确认服务运行

```bash
# 检查 admin-api 端口（默认 8888）
lsof -i :8888

# 检查 front-api 端口（默认 8001）
lsof -i :8001

# 或健康检查
curl -sf http://127.0.0.1:8888/health
```

### 6.2 JWT Token 获取

**登录接口**：
```
POST /api/sys/user/login
Content-Type: application/json

Request:
{
  "account": "<账号>",
  "password": "<密码>"
}

Response (成功):
{
  "code": "0",
  "message": "",
  "data": {
    "token": "<token>"
  }
}
```

**获取 Token 脚本**：
```bash
BASE_URL="http://127.0.0.1:8888"

LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/sys/user/login" \
  -H 'Content-Type: application/json' \
  -d "{\"account\":\"<账号>\",\"password\":\"<密码>\"}")

TOKEN=$(echo "$LOGIN_RESP" | python3 -c \
  "import sys,json; print(json.load(sys.stdin).get('data',{}).get('token',''))" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "❌ Token 获取失败"; exit 1
fi
echo "✅ Token 获取成功: $TOKEN"
```

**测试账号来源**：
- 查看 `rpc/*/etc/*.yaml` 配置文件中的数据库初始化数据
- 默认开发账号：`admin / 123456`

### 6.3 生成 test_api.sh

**基本模板**：
```bash
#!/bin/bash
# Zero-Admin API 测试脚本
# Issue: <名称>
BASE_URL="http://127.0.0.1:8888"
PASS=0; FAIL=0; TOTAL=0

RED='\033[0;31m'; GREEN='\033[0;32m'; NC='\033[0m'

run_test() {
  local name="$1"; local expected="$2"; shift 2
  TOTAL=$((TOTAL+1))
  echo "--- [$TOTAL] $name ---"
  RESP=$(curl -s -w "\n%{http_code}" "$@")
  HTTP=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')
  CODE=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code','N/A'))" 2>/dev/null || echo "N/A")
  if [ "$CODE" = "$expected" ]; then
    echo -e "${GREEN}✅ PASS${NC} (code=$CODE)"
    PASS=$((PASS+1))
  else
    echo -e "${RED}❌ FAIL${NC} (expected=$expected, got=$CODE)"
    echo "Response: $(echo "$BODY" | head -3)"
    FAIL=$((FAIL+1))
  fi
}

# 1. 获取 Token
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/sys/user/login" \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"123456"}')
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('token',''))" 2>/dev/null)
AUTH="Authorization: Bearer $TOKEN"

# 2. 测试用例
# List 操作
run_test "List <资源>" "0" \
  -X GET "$BASE_URL/<api_path>?current=1&pageSize=20" -H "$AUTH"

# Create 操作（使用唯一值避免冲突）
TS=$(date +%s)
UNIQUE_NO="<前缀>${TS}"
run_test "Create <资源>" "0" \
  -X POST "$BASE_URL/<api_path>" -H "$AUTH" \
  -H 'Content-Type: application/json' \
  --data-raw "{\"name\":\"${UNIQUE_NO}\",\"desc\":\"测试${TS}\"}"

# ... 更多用例

echo "==============================="
echo -e "结果: ${GREEN}${PASS}/${TOTAL}${NC} 通过, ${RED}${FAIL}${NC} 失败"
[ $FAIL -gt 0 ] && exit 1 || exit 0
```

### 6.4 执行测试

```bash
bash <issue_dir>/test_api.sh
```

### 6.5 测试覆盖标准

| 类型 | 要求 |
|------|------|
| List | ✅ 必测，验证分页和返回格式 |
| Create | ✅ 必测，验证唯一约束和必填校验 |
| Get | ✅ 必测，验证单个资源获取 |
| Update | ✅ 必测，验证更新生效 |
| Delete | ⚠️ 按需，禁止删除测试数据 |
| 错误场景 | ✅ 必测，无效 ID、缺少参数 |

### 6.6 🚨 测试通过率要求：100%

> **🚨 硬性要求：API 测试必须 100% 通过。任何失败用例都必须立即修复，修复后重新执行测试，直到全部通过。不允许带着失败用例进入下一步骤。**

**失败处理流程**：
1. 分析失败原因
2. 如果是测试数据问题（如唯一索引冲突）→ 使用更唯一的值重试
3. 如果是 API 实现问题 → 修复代码，重新编译，再测试
4. 重复直到 100% 通过

### 6.7 唯一值策略

> ⚠️ 禁止清理数据，唯一索引冲突使用时间戳生成唯一值

```bash
TS=$(date +%s)  # Unix 时间戳

# 唯一编号
NO="PREFIX${TS}"

# 如遇冲突，等待下一秒重试
sleep 1
TS=$(date +%s)
NO="PREFIX${TS}"
```

---

## 步骤 7：验收标准验证

从 Issue 规格提取验收标准，逐条检查：

| # | 验收标准 | 代码位置 | 状态 |
|---|---------|---------|------|
| 1 | ... | file:line | ✅/⚠️/❌ |

**状态说明**：
- ✅ 已实现且验证通过
- ⚠️ 部分实现，有遗漏
- ❌ 未实现

---

## 步骤 8：输出诊断报告

**输出位置**：`<issue_dir>/progress.md`

```markdown
## QA 诊断报告 — <日期>

**Issue**: <名称>
**Issue 类型**: <纯后端 RPC / API + RPC / 后端集成>
**诊断目标**: <简述>

### 1. 编译检查
| 检查项 | 结果 |
|--------|------|
| Go 编译 | ✅/❌ |

### 2. 代码质量
| 检查项 | 结果 |
|--------|------|
| Mock 数据 | ✅ 无 / ❌ 发现 N 处 |
| 硬编码 | ✅ 无 / ⚠️ 发现 N 处 |
| 调试日志 | ✅ 已清理 / ⚠️ 发现 N 处 |

### 3. Bug 修复验证
| Bug | 验证方法 | 结果 |
|-----|---------|-----|
| ... | ... | ✅/❌ |

### 4. API 测试
| 用例 | 结果 |
|------|------|
| List XXX | ✅ PASS |
| Create XXX | ✅ PASS |
| ... | ... |

**通过率**: X/Y (Z%)

### 5. 验收标准
| # | 标准 | 状态 |
|---|------|------|
| 1 | ... | ✅ |

### 6. 诊断结论
**状态**: ✅ 通过 / ⚠️ 有条件通过 / ❌ 不通过

- 已验证: <列表>
- 待跟进: <列表>
```

---

## 快速参考

### 服务管理
```bash
make build                    # 构建所有服务（产物在 target/）
make start                    # 启动所有服务
make stop                     # 停止所有服务
make restart                  # 重启所有服务
make test                     # 快速测试（build + restart）
```

### 编译构建
```bash
go build ./...                # 编译所有
go vet ./...                  # 静态检查
go fmt ./...                  # 格式化
make lint                     # golangci-lint 检查
make lint-fix                 # golangci-lint 自动修复
```

### API 测试
```bash
# 健康检查
curl http://127.0.0.1:8888/health

# JWT 登录（路径: /api/sys/user/login）
curl -X POST http://127.0.0.1:8888/api/sys/user/login \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"123456"}'

# 带 Token 请求
curl -X GET "http://127.0.0.1:8888/<api_path>" \
  -H "Authorization: Bearer <token>"
```

### 日志查看
```bash
# 查看服务日志
tail -f target/logs/admin/admin-api.log   # admin-api 日志
tail -f target/logs/oms/oms-service.log   # oms-rpc 日志
```

---

## 异常处理

| 异常 | 处理方式 |
|------|---------|
| 编译错误 undefined | 检查 import 路径和包名 |
| 编译错误类型不匹配 | 对照字段定义修正 |
| API 500 Table not exist | 检查 GORM 模型和数据库 |
| API 404 Route not found | 检查 API 路由注册 |
| API 401 Unauthorized | 重新获取 JWT token |
| 端口被占用 | `lsof -ti:<port> | xargs kill -9` |
| 唯一索引冲突 (1062) | 使用时间戳生成唯一值重试 |
| 数据库连接失败 | 检查 etc/*.yaml 配置 |

---

## 关键文件参考

| 文件 | 用途 |
|------|------|
| `rpc/*/internal/logic/` | RPC 业务逻辑层 |
| `api/*/internal/logic/` | API 业务逻辑层 |
| `consumer/internal/mq/` | RabbitMQ 消费者 |
| `rpc/*/gen/model/` | GORM 数据模型 |
| `rpc/*/etc/*.yaml` | RPC 服务配置 |
| `api/*/etc/*.yaml` | API 服务配置 |
| `makefile` | 构建脚本 |
| `target/` | 构建产物目录 |
| `_opcos/implementation-artifacts/` | Issue 规格文件目录 |
