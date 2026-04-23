---
title: '多链 Adapter 架构重构'
slug: 'multi-chain-adapter-refactor'
created: '2026-04-23T15:55:00+0800'
status: 'ready-for-dev'
stepsCompleted: [1, 2, 3, 4]
tech_stack: ['Go 1.25', 'go-zero 1.9.3', 'GORM', 'gRPC', 'FISCO BCOS Go SDK', 'RabbitMQ']
files_to_modify:
  - 'pkg/chainclient/chainclient.go (新建)'
  - 'pkg/chainclient/types.go (新建)'
  - 'pkg/fisco/client.go (新建)'
  - 'pkg/fisco/types.go (新建)'
  - 'pkg/fisco/client_test.go (新建)'
  - 'pkg/antchain/client.go (适配)'
  - 'pkg/antchain/types.go (适配)'
  - 'pkg/antchain/mock.go (适配)'
  - 'pkg/digitalcardmint/service.go (重构核心)'
  - 'pkg/digitalcardmint/service_test.go (更新 stub)'
  - 'rpc/sms/internal/config/config.go (增加 Blockchain/Fisco 配置)'
  - 'rpc/sms/internal/svc/service_context.go (链路由)'
  - 'rpc/sms/etc/sms.yaml (新增配置段)'
  - 'consumer/internal/config/config.go (增加 Blockchain/Fisco 配置)'
  - 'consumer/internal/svc/service_context.go (链路由)'
  - 'job/internal/config/config.go (增加 Blockchain/Fisco 配置)'
  - 'job/internal/svc/service_context.go (链路由)'
code_patterns:
  - 'ServiceContext 注入 (handler→logic→svc)'
  - 'MQPublisher 接口模式 (digitalcardmint 已用)'
  - 'antchain.Client 接口 + disabledClient + httpClient + MockClient'
  - 'Config struct 嵌套 (AntChain/Rabbitmq/Mysql 块)'
  - 'stubAntChainClient 测试模式'
test_patterns:
  - 'stubAntChainClient 实现 antchain.Client (service_test.go)'
  - 'countingPublisher 实现 MQPublisher (service_test.go)'
  - 'newDigitalCardMintTestDB() 创建 SQLite 内存测试库'
  - 'antchain.MockClient 带 SetError/SetQueryError 注入'
---

# Tech-Spec: 多链 Adapter 架构重构

**Created:** 2026-04-23T15:55:00+0800

## Overview

### Problem Statement

`docs/blockchain-strategy.md` 设计了多链 Adapter 架构（通用 `ChainClient` 接口 + FISCO 默认底座 + AntChain 升级路径 + `Blockchain.Primary` 链路由配置），但当前代码直接硬依赖 `antchain.Client`，存在以下差距：

1. **无抽象层**：`digitalcardmint.Service` 直接持有 `antchain.Client`，没有通用 `ChainClient` 接口
2. **无 FISCO 实现**：`pkg/fisco/` 完全不存在，阶段一的免费链底座缺失
3. **无链路由**：配置中没有 `Blockchain.Primary` 字段，无法在 FISCO 和 AntChain 之间切换
4. **三个服务均硬绑定**：sms-rpc、consumer、job 的 ServiceContext 都直接构造 `antchain.Client`

### Solution

提取通用 `ChainClient` 接口到独立包（`pkg/chainclient`），实现 FISCO Adapter（`pkg/fisco`），重构 `pkg/antchain` 适配新接口，重构 `digitalcardmint.Service` 依赖接口而非具体实现，增加 `Blockchain.Primary` 链路由配置，让 sms-rpc / consumer / job 三个服务根据配置选择活动链。

### Scope

**In Scope:**

- 通用 `ChainClient` 接口定义（`MintToken` / `QueryMintToken` / `ChainType()`）
- `pkg/fisco` FISCO BCOS 客户端实现（通过 Go SDK 调用 FISCO BCOS 节点）
- 重构 `pkg/antchain` 适配新 `ChainClient` 接口
- 重构 `pkg/digitalcardmint/Service` 解耦 `antchain.Client` → `ChainClient`
- `rpc/sms`、`consumer`、`job` 的 Config + ServiceContext 适配链路由
- 配置驱动的链选择（`Blockchain.Primary: fisco | antchain`）
- 单元测试覆盖新接口和 FISCO 客户端

**Out of Scope:**

- 至信链 Adapter（阶段三 A）
- 双写机制
- 回执对账、索引优化、监控告警（近期 TODO 项）
- 数据库表结构变更
- 前端 / Flutter 改动
- FISCO BCOS 节点部署和运维

## Context for Development

### `antchain.Client` 耦合点完整映射

#### A. `pkg/digitalcardmint/service.go`（1881 行，重构核心）

| 行号 | 耦合点 | 类型 |
|------|--------|------|
| 11 | `import "github.com/feihua/zero-admin/pkg/antchain"` | 导入 |
| 24 | `AntChain antchain.Client` (Service 字段) | 字段声明 |
| 57 | `NewService(db, mq, client antchain.Client)` | 构造函数参数 |
| 413 | `var receipt *antchain.MintTokenResponse` (ExecuteTask) | 变量类型 |
| 444 | `s.AntChain == nil` 检查 | nil 检查 |
| 448 | `s.AntChain.MintToken(ctx, &antchain.MintTokenRequest{...})` | 调用 MintToken |
| 967 | `transitionSuccess(..., receipt *antchain.MintTokenResponse, ...)` | 参数类型 |
| 1080 | `reconcilePreviousReceipt(...) (*antchain.MintTokenResponse, ...)` | 返回类型 |
| 1089 | `errors.Is(err, antchain.ErrReceiptNotFound)` | 错误判断 |
| 1095 | `errors.Is(err, antchain.ErrReceiptNotFound)` | 错误判断 |
| 1107 | `s.AntChain == nil` 检查 | nil 检查 |
| 1113 | `s.AntChain.QueryMintToken(ctx, &antchain.QueryMintTokenRequest{...})` | 调用 QueryMintToken |
| 1190 | `bestEffortMarkWritebackFailure(..., receipt *antchain.MintTokenResponse, ...)` | 参数类型 |
| 1232 | `handleSuccessPersistenceFailure(..., receipt *antchain.MintTokenResponse, ...)` | 参数类型 |
| 1239 | `markTokenBindingConflict(..., receipt *antchain.MintTokenResponse, ...)` | 参数类型 |
| 1276 | `buildReceiptFromTask() → *antchain.MintTokenResponse` | 返回类型 |
| 1288 | `&antchain.MintTokenResponse{...}` 构造 | 类型构造 |

#### B. `rpc/sms/internal/svc/service_context.go`（82 行）

| 行号 | 耦合点 |
|------|--------|
| 7 | `import "github.com/feihua/zero-admin/pkg/antchain"` |
| 22 | `AntChain antchain.Client` (ServiceContext 字段) |
| 44 | `antchain.NewClient(antchain.Config{...})` |
| 53 | `digitalcardmint.NewService(DB, rabbitmq, antChainClient)` |
| 59 | `AntChain: antChainClient` |

#### C. `consumer/internal/svc/service_context.go`（243 行）

| 行号 | 耦合点 |
|------|--------|
| 13 | `import "github.com/feihua/zero-admin/pkg/antchain"` |
| 41 | `AntChain antchain.Client` (ServiceContext 字段) |
| 87 | `antchain.NewClient(antchain.Config{...})` |
| 96 | `digitalcardmint.NewService(db, rabbitmq, antChainClient)` |
| 119 | `AntChain: antChainClient` |

#### D. `job/internal/svc/service_context.go`（99 行）

| 行号 | 耦合点 |
|------|--------|
| 5 | `import "github.com/feihua/zero-admin/pkg/antchain"` |
| 24 | `AntChain antchain.Client` (ServiceContext 字段) |
| 56 | `antchain.NewClient(antchain.Config{...})` |
| 65 | `digitalcardmint.NewService(db, nil, antChainClient)` |
| 86 | `AntChain: antChainClient` |

#### E. Config 文件（三个服务均有 `AntChain` 块，需增加 `Blockchain` + `Fisco`）

- `rpc/sms/internal/config/config.go` (29 行)
- `consumer/internal/config/config.go` (54 行)
- `job/internal/config/config.go` (34 行)

#### F. 测试文件

- `pkg/digitalcardmint/service_test.go`: `stubAntChainClient` 实现 `antchain.Client` 接口
- `pkg/antchain/mock.go`: `MockClient` 实现 `antchain.Client`

### 现有 `antchain.Client` 接口签名

```go
// pkg/antchain/client.go
type Client interface {
    MintToken(ctx context.Context, req *MintTokenRequest) (*MintTokenResponse, error)
    QueryMintToken(ctx context.Context, req *QueryMintTokenRequest) (*MintTokenResponse, error)
}
```

### 目标 `ChainClient` 接口签名（来自策略文档）

```go
// pkg/chainclient/chainclient.go
type ChainClient interface {
    MintToken(ctx context.Context, req *MintTokenRequest) (*MintTokenResponse, error)
    QueryMintToken(ctx context.Context, req *QueryMintTokenRequest) (*MintTokenResponse, error)
    ChainType() string  // "fisco" | "antchain"
}
```

### Codebase Patterns

- **ServiceContext 注入**：`handler → logic → svc`，所有链客户端通过 `svcCtx` 传递
- **接口抽象模式**：`MQPublisher` 接口已是现有范式（定义在 `digitalcardmint` 包内），`ChainClient` 应提升到独立包以避免循环依赖
- **disabled 模式**：`antchain` 已有 `disabledClient`（返回 "蚂蚁链能力未启用"），FISCO 应遵循同样模式
- **Config 嵌套结构体**：每个服务的 `Config` 使用内嵌匿名结构体（`AntChain struct{...}`），新增 `Blockchain` 和 `Fisco` 块保持一致
- **测试 stub 模式**：`stubAntChainClient` 在 `service_test.go` 内部定义，直接嵌入字段替代行为

### Files to Reference

| File | Purpose | 行数 |
| ---- | ------- | ---- |
| `pkg/antchain/client.go` | 现有蚂蚁链客户端（Client 接口 + httpClient + disabledClient） | 272 |
| `pkg/antchain/types.go` | 蚂蚁链请求/响应/Config 类型 | 52 |
| `pkg/antchain/mock.go` | MockClient（SetError/SetQueryError 注入故障） | 110 |
| `pkg/digitalcardmint/service.go` | 核心状态机（EnsureTaskTx/DispatchTask/ExecuteTask/ScanDueTasks 等） | 1881 |
| `pkg/digitalcardmint/model.go` | CardMintTaskRow/CardInstanceRow 等 GORM 模型 | 246 |
| `pkg/digitalcardmint/constants.go` | TaskStatus*/MintStatus*/ChainStatus* 常量 | 88 |
| `pkg/digitalcardmint/service_test.go` | stubAntChainClient + 单元测试 | 1036 |
| `rpc/sms/internal/config/config.go` | sms-rpc Config（含 AntChain 块） | 29 |
| `rpc/sms/internal/svc/service_context.go` | sms-rpc ServiceContext 构造 | 82 |
| `rpc/sms/etc/sms.yaml` | sms-rpc YAML 配置（含 AntChain 空配置） | 35 |
| `consumer/internal/config/config.go` | consumer Config（含 AntChain 块） | 54 |
| `consumer/internal/svc/service_context.go` | consumer ServiceContext 构造 | 243 |
| `job/internal/config/config.go` | job Config（含 AntChain 块） | 34 |
| `job/internal/svc/service_context.go` | job ServiceContext 构造 | 99 |
| `consumer/internal/mq/digital_card/mint_requested.go` | MQ 消费者（不直接引用 antchain） | 28 |
| `job/internal/jobs/handle_card_mint_timeout_logic.go` | 超时补偿（不直接引用 antchain） | 24 |

### Technical Decisions

1. **接口放置位置**：`ChainClient` 接口 + 通用类型定义在 `pkg/chainclient/` 独立包中。原因：`digitalcardmint` 依赖链客户端接口，如果接口在 `antchain` 内则 `fisco` 也得导入 `antchain`，产生语义耦合。独立包避免循环依赖和语义污染
2. **通用类型提升**：`MintTokenRequest`、`MintTokenResponse`、`QueryMintTokenRequest`、`ErrReceiptNotFound` 从 `antchain` 提升到 `chainclient`。`antchain` 包内部可保留 HTTP payload/response 私有类型用于序列化
3. **FISCO 实现**：`pkg/fisco` 实现 `chainclient.ChainClient`。初期通过 FISCO BCOS Go SDK 调用链节点，`ChainType()` 返回 `"fisco"`
4. **AntChain 适配**：`antchain.httpClient` 和 `disabledClient` 直接实现 `chainclient.ChainClient`（方法签名只需改参数/返回类型为 `chainclient.*`）。`antchain.MockClient` 同理适配
5. **配置层级**：新增顶层 `Blockchain struct { Primary string }` 配置。`Fisco` 和 `AntChain` 配置块并列存在。ServiceContext 根据 `Blockchain.Primary` 选择构造哪个 client
6. **向后兼容**：`AntChain` 配置块保留不删除，只是 ServiceContext 不再直接使用它构造，而是通过链路由逻辑间接使用
7. **不改数据库**：数据表无需 DDL 变更。未来若需记录每条记录使用的链类型，可在业务逻辑中添加

## Implementation Plan

### Tasks

#### 阶段一：创建通用接口层（无依赖）

- [ ] Task 1: 创建 `pkg/chainclient/chainclient.go` — ChainClient 接口定义
  - File: `pkg/chainclient/chainclient.go` (新建)
  - Action: 定义 `ChainClient` 接口，包含 `MintToken(ctx, *MintTokenRequest) (*MintTokenResponse, error)`、`QueryMintToken(ctx, *QueryMintTokenRequest) (*MintTokenResponse, error)`、`ChainType() string` 三个方法
  - Notes: 包名 `chainclient`，接口命名 `ChainClient`

- [ ] Task 2: 创建 `pkg/chainclient/types.go` — 通用请求/响应类型
  - File: `pkg/chainclient/types.go` (新建)
  - Action: 从 `pkg/antchain/types.go` 提升以下类型到此文件：`MintTokenRequest`、`QueryMintTokenRequest`、`MintTokenResponse`、`ErrReceiptNotFound`。字段完全保持不变
  - Notes: `ErrReceiptNotFound = errors.New("未找到链上回执")`，错误文案保持一致

#### 阶段二：适配现有 AntChain 包（依赖 Task 1-2）

- [ ] Task 3: 重构 `pkg/antchain/types.go` — 移除已提升的类型
  - File: `pkg/antchain/types.go`
  - Action: 删除 `MintTokenRequest`、`QueryMintTokenRequest`、`MintTokenResponse`、`ErrReceiptNotFound` 的定义。保留 `Config` 结构体。保留包内私有 HTTP payload/response 类型不变
  - Notes: 为保证渐进迁移兼容，在文件中添加类型别名：`type MintTokenRequest = chainclient.MintTokenRequest` 等，后续可移除

- [ ] Task 4: 重构 `pkg/antchain/client.go` — 适配 ChainClient 接口
  - File: `pkg/antchain/client.go`
  - Action:
    1. 将 `import` 中添加 `"github.com/feihua/zero-admin/pkg/chainclient"`
    2. 将 `Client` 接口改为嵌入 `chainclient.ChainClient`（或直接删除 `Client` 接口，让 `httpClient` 和 `disabledClient` 直接实现 `chainclient.ChainClient`）
    3. `httpClient` 添加 `ChainType() string` 方法返回 `"antchain"`
    4. `disabledClient` 添加 `ChainType() string` 方法返回 `"antchain"`
    5. 方法签名中 `*MintTokenRequest` → `*chainclient.MintTokenRequest`，`*MintTokenResponse` → `*chainclient.MintTokenResponse` 等
    6. `NewClient()` 返回类型改为 `chainclient.ChainClient`
  - Notes: `buildMintTokenResponse` 内部辅助函数返回类型同步改为 `*chainclient.MintTokenResponse`

- [ ] Task 5: 重构 `pkg/antchain/mock.go` — MockClient 适配 ChainClient
  - File: `pkg/antchain/mock.go`
  - Action:
    1. `MockClient` 方法签名适配 `chainclient.*` 类型
    2. 添加 `ChainType() string` 返回 `"antchain-mock"`
    3. `cloneMintTokenResponse` 参数/返回类型改为 `*chainclient.MintTokenResponse`
  - Notes: `MockClient` 仍保留在 `antchain` 包，但实现 `chainclient.ChainClient`

#### 阶段三：创建 FISCO 客户端（依赖 Task 1-2）

- [ ] Task 6: 创建 `pkg/fisco/types.go` — FISCO 配置类型
  - File: `pkg/fisco/types.go` (新建)
  - Action: 定义 `Config struct`，包含 `NodeAddr string`、`GroupID int`、`ChainID int64`、`ContractAddr string`、`PrivateKey string`、`TimeoutSeconds int64`、`Enabled bool`
  - Notes: 配置项参考 FISCO BCOS Go SDK 连接参数

- [ ] Task 7: 创建 `pkg/fisco/client.go` — FISCO ChainClient 实现
  - File: `pkg/fisco/client.go` (新建)
  - Action:
    1. 定义 `fiscoClient struct`（内部持有 FISCO SDK 连接和 Config）
    2. 实现 `MintToken(ctx, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error)` — 调用 FISCO 合约的 mint 方法
    3. 实现 `QueryMintToken(ctx, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error)` — 查询链上交易回执
    4. 实现 `ChainType() string` 返回 `"fisco"`
    5. 定义 `disabledClient struct` + 三个方法均返回 `errors.New("FISCO BCOS 能力未启用")`
    6. `NewClient(cfg Config) chainclient.ChainClient` — 如 `cfg.Enabled` 为 false 返回 `disabledClient`
  - Notes: 初期 FISCO SDK 调用可使用占位实现（TODO 注释标记），核心是接口对齐和架构正确

- [ ] Task 8: 创建 `pkg/fisco/client_test.go` — FISCO 客户端测试
  - File: `pkg/fisco/client_test.go` (新建)
  - Action:
    1. `TestNewClientDisabled` — Enabled=false 时 MintToken/QueryMintToken 返回 "FISCO BCOS 能力未启用" 错误
    2. `TestNewClientEnabled` — Enabled=true 时返回 `fiscoClient` 实例（验证 ChainType()="fisco"）
    3. `TestFiscoClientImplementsChainClient` — 编译期接口断言 `var _ chainclient.ChainClient = (*fiscoClient)(nil)`
  - Notes: 不需要真正连接 FISCO 节点，测试 disabled 模式和接口兼容性即可

#### 阶段四：重构核心服务（依赖 Task 1-5）

- [ ] Task 9: 重构 `pkg/digitalcardmint/service.go` — 解耦 antchain 直接依赖
  - File: `pkg/digitalcardmint/service.go`
  - Action:
    1. 将 `import "github.com/feihua/zero-admin/pkg/antchain"` 替换为 `"github.com/feihua/zero-admin/pkg/chainclient"`
    2. `Service.AntChain antchain.Client` → `Service.Chain chainclient.ChainClient`
    3. `NewService(db, mq, client antchain.Client)` → `NewService(db, mq, client chainclient.ChainClient)`，内部 `s.AntChain = client` → `s.Chain = client`
    4. 所有 `s.AntChain` 引用改为 `s.Chain`（约 4 处）
    5. 所有 `antchain.MintTokenRequest` → `chainclient.MintTokenRequest`（1 处）
    6. 所有 `antchain.MintTokenResponse` → `chainclient.MintTokenResponse`（约 10 处）
    7. 所有 `antchain.QueryMintTokenRequest` → `chainclient.QueryMintTokenRequest`（1 处）
    8. 所有 `antchain.ErrReceiptNotFound` → `chainclient.ErrReceiptNotFound`（2 处）
    9. 错误消息 "蚂蚁链客户端未初始化" → "链客户端未初始化"
  - Notes: 纯机械替换，不改变任何业务逻辑。替换后 `antchain` 导入应完全消失

- [ ] Task 10: 重构 `pkg/digitalcardmint/service_test.go` — 更新测试 stub
  - File: `pkg/digitalcardmint/service_test.go`
  - Action:
    1. 将 `import "github.com/feihua/zero-admin/pkg/antchain"` 替换为 `"github.com/feihua/zero-admin/pkg/chainclient"`
    2. `stubAntChainClient` → `stubChainClient`，所有字段类型中的 `antchain.*` → `chainclient.*`
    3. 添加 `func (c *stubChainClient) ChainType() string { return "test" }`
    4. `antchain.ErrReceiptNotFound` → `chainclient.ErrReceiptNotFound`
    5. 所有 `stubAntChainClient` 实例化处更新为 `stubChainClient`
  - Notes: 运行 `go test ./pkg/digitalcardmint/...` 验证测试仍全部通过

#### 阶段五：配置适配（依赖 Task 6）

- [ ] Task 11: 更新 `rpc/sms/internal/config/config.go` — 新增 Blockchain + Fisco 配置
  - File: `rpc/sms/internal/config/config.go`
  - Action: 在 `Config` 中新增：
    ```go
    Blockchain struct {
        Primary string // "fisco" | "antchain"，默认 "fisco"
    }
    Fisco struct {
        NodeAddr       string
        GroupID        int
        ChainID        int64
        ContractAddr   string
        PrivateKey     string
        TimeoutSeconds int64
        Enabled        bool
    }
    ```
  - Notes: `AntChain` 块保留不删除

- [ ] Task 12: 更新 `consumer/internal/config/config.go` — 同 Task 11
  - File: `consumer/internal/config/config.go`
  - Action: 新增同样的 `Blockchain` 和 `Fisco` 结构体
  - Notes: 与 Task 11 结构完全一致

- [ ] Task 13: 更新 `job/internal/config/config.go` — 同 Task 11
  - File: `job/internal/config/config.go`
  - Action: 新增同样的 `Blockchain` 和 `Fisco` 结构体
  - Notes: 与 Task 11 结构完全一致

#### 阶段六：ServiceContext 链路由（依赖 Task 3-4, 7, 9, 11-13）

- [ ] Task 14: 重构 `rpc/sms/internal/svc/service_context.go` — 链路由逻辑
  - File: `rpc/sms/internal/svc/service_context.go`
  - Action:
    1. 导入 `"github.com/feihua/zero-admin/pkg/chainclient"` 和 `"github.com/feihua/zero-admin/pkg/fisco"`
    2. `ServiceContext.AntChain antchain.Client` → `ServiceContext.ChainClient chainclient.ChainClient`
    3. 新增 `buildChainClient(c config.Config) chainclient.ChainClient` 函数：
       - `c.Blockchain.Primary == "antchain"` → `antchain.NewClient(...)`
       - 默认（含 `"fisco"` 和空值）→ `fisco.NewClient(...)`
    4. 构造处改为 `chainClient := buildChainClient(c)` → `digitalcardmint.NewService(DB, rabbitmq, chainClient)`
    5. `AntChain: antChainClient` → `ChainClient: chainClient`
  - Notes: `antchain` 导入保留（用于 antchain 分支），新增 `fisco` 导入

- [ ] Task 15: 重构 `consumer/internal/svc/service_context.go` — 同 Task 14
  - File: `consumer/internal/svc/service_context.go`
  - Action: 同 Task 14 的逻辑，`ServiceContext.AntChain` → `ServiceContext.ChainClient`，新增 `buildChainClient` 或内联链路由逻辑
  - Notes: consumer 还有 `AntChain: antChainClient` 赋值需更新

- [ ] Task 16: 重构 `job/internal/svc/service_context.go` — 同 Task 14
  - File: `job/internal/svc/service_context.go`
  - Action: 同 Task 14，job 的 `NewService(db, nil, antChainClient)` → `NewService(db, nil, chainClient)`
  - Notes: job 不使用 MQ（第二参数为 nil），只传链客户端

#### 阶段七：YAML 配置更新

- [ ] Task 17: 更新 YAML 配置文件 — 新增 Blockchain + Fisco 配置段
  - Files:
    - `rpc/sms/etc/sms.yaml`
    - `consumer/etc/consumer-api.yaml`
    - `job/etc/job-api.yaml`
  - Action: 在每个 yaml 文件中新增：
    ```yaml
    Blockchain:
      Primary: fisco

    Fisco:
      NodeAddr: ""
      GroupID: 1
      ChainID: 1
      ContractAddr: ""
      PrivateKey: ""
      TimeoutSeconds: 10
      Enabled: false
    ```
  - Notes: 初始 `Enabled: false`，`Primary: fisco` 表明默认使用 FISCO。现有 `AntChain` 段保留不动

#### 阶段八：验证

- [ ] Task 18: 编译验证 + 全量测试
  - Action:
    1. `go build ./...` — 确保所有包编译通过
    2. `go test ./pkg/chainclient/... ./pkg/antchain/... ./pkg/fisco/... ./pkg/digitalcardmint/... -v` — 核心包测试
    3. `go test ./rpc/sms/... ./consumer/... ./job/... -v -short` — 服务层测试
    4. `go vet ./...` — 静态检查
  - Notes: 重点关注 `digitalcardmint/service_test.go`，它是覆盖率最高的测试文件

### Acceptance Criteria

#### 接口层

- [ ] AC 1: Given `pkg/chainclient/` 包已创建, when 编译 `go build ./pkg/chainclient/...`, then 编译成功且 `ChainClient` 接口包含 `MintToken`、`QueryMintToken`、`ChainType` 三个方法
- [ ] AC 2: Given `chainclient.MintTokenRequest` 类型已定义, when 比较其字段与原 `antchain.MintTokenRequest`, then 所有字段名、类型、tag 完全一致

#### AntChain 适配

- [ ] AC 3: Given AntChain 配置 `Enabled: true`, when 调用 `antchain.NewClient(cfg)`, then 返回的对象实现 `chainclient.ChainClient` 接口且 `ChainType()` 返回 `"antchain"`
- [ ] AC 4: Given AntChain 配置 `Enabled: false`, when 调用 `antchain.NewClient(cfg)`, then 返回 `disabledClient` 且 `MintToken()` 返回 `errors.New("蚂蚁链能力未启用")`
- [ ] AC 5: Given `antchain.MockClient` 实例, when 调用 `MintToken()` 后调用 `QueryMintToken()`, then 能查到之前 mint 的回执（现有测试行为不变）

#### FISCO 实现

- [ ] AC 6: Given FISCO 配置 `Enabled: false`, when 调用 `fisco.NewClient(cfg)`, then `MintToken()` 和 `QueryMintToken()` 均返回 `errors.New("FISCO BCOS 能力未启用")`
- [ ] AC 7: Given FISCO 配置 `Enabled: true`, when 调用 `fisco.NewClient(cfg)`, then 返回的对象 `ChainType()` 返回 `"fisco"` 且实现 `chainclient.ChainClient`

#### 核心服务解耦

- [ ] AC 8: Given `digitalcardmint.Service` 已重构, when 用 `grep -r "antchain" pkg/digitalcardmint/` 搜索, then 无任何匹配结果（`antchain` 导入完全消除）
- [ ] AC 9: Given `stubChainClient` 注入到 Service, when 执行 `TestEnsureTaskTxIsIdempotent`, then 测试通过（幂等性不变）
- [ ] AC 10: Given 一个 stub ChainClient (MintToken 返回成功), when 调用 `ExecuteTask()`, then task 状态转为 `succeeded` 且 `token_id` 被写入

#### 链路由配置

- [ ] AC 11: Given Config 中 `Blockchain.Primary: "antchain"`, when 构造 ServiceContext, then `svcCtx.ChainClient.ChainType()` 返回 `"antchain"`
- [ ] AC 12: Given Config 中 `Blockchain.Primary: "fisco"`, when 构造 ServiceContext, then `svcCtx.ChainClient.ChainType()` 返回 `"fisco"`
- [ ] AC 13: Given Config 中 `Blockchain.Primary` 为空, when 构造 ServiceContext, then 默认使用 FISCO 客户端

#### 编译与测试

- [ ] AC 14: Given 所有代码变更完成, when 运行 `go build ./...`, then 编译成功无错误
- [ ] AC 15: Given 所有代码变更完成, when 运行 `go test ./pkg/... -count=1`, then 所有测试通过

## Additional Context

### Dependencies

- **新增依赖**: `github.com/FISCO-BCOS/go-sdk` — FISCO BCOS Go SDK（用于 `pkg/fisco` 实现）
- **现有依赖不变**: `gorm.io/gorm`、`github.com/zeromicro/go-zero`、`github.com/rabbitmq/amqp091-go`
- **无外部服务依赖**: FISCO 初始实现为 disabled 模式，不需要真实 FISCO 节点
- **内部依赖链**: `chainclient` ← `antchain`/`fisco` ← `digitalcardmint` ← `svc`

### Testing Strategy

#### 单元测试

- `pkg/fisco/client_test.go` (新建): disabled 模式 + 接口断言 + ChainType 返回值
- `pkg/digitalcardmint/service_test.go` (修改): `stubAntChainClient` → `stubChainClient`，全部现有测试无功能性改动
- `pkg/antchain/mock.go` (修改): MockClient 适配 chainclient 接口

#### 编译期验证

- 在 `pkg/fisco/client.go` 和 `pkg/antchain/client.go` 中添加接口断言：
  ```go
  var _ chainclient.ChainClient = (*fiscoClient)(nil)
  var _ chainclient.ChainClient = (*httpClient)(nil)
  var _ chainclient.ChainClient = (*disabledClient)(nil)
  ```

#### 手动验证

1. `go build ./...` — 全量编译
2. `go test ./pkg/chainclient/... ./pkg/antchain/... ./pkg/fisco/... ./pkg/digitalcardmint/... -v -count=1` — 核心包测试
3. `go vet ./...` — 静态分析
4. 检查 `grep -r "antchain" pkg/digitalcardmint/` 输出为空

### Notes

#### 高风险项

- **`service.go` 1881 行大文件重构**: 虽然是机械替换，但涉及 17 处 `antchain.*` 引用，需逐一确认无遗漏。建议使用 `replace_all` 模式和编译验证
- **类型别名过渡期**: `antchain/types.go` 中保留类型别名可能导致两份类型定义共存，需在后续清理

#### 已知限制

- FISCO 客户端初期为**占位实现**（MintToken/QueryMintToken 内部 TODO），真正的 SDK 集成需要 FISCO 节点环境
- 不含双写机制：切换链后历史数据仍指向原链，无自动迁移

#### 未来考虑（Out of Scope）

- 至信链 ZXChain Adapter（阶段三 A）
- 双写并行运行模式
- 链上回执自动对账
- `sms_card_mint_task` 表增加 `chain_type` 字段记录每笔任务使用的链
- 链路由支持按租户/商户维度配置不同链
