# Shared Services Layer

跨 Issue 共享的服务层设计。避免 37 个 issue 各自重复实现相同功能。

---

## 设计原则

1. **单一职责** — 每个共享服务只负责一个业务域
2. **接口隔离** — Flutter 端通过 Repository 接口访问，不直接依赖实现
3. **后端聚合** — 页面级聚合 API 在 application 层组合多个 domain service
4. **缓存策略** — 高频读取数据（用户信息/员工列表）本地缓存，低频数据实时请求

---

## 共享服务清单

### 1. AuthService（认证服务）

**复用范围**: 所有 37 个 issue

**后端**:
- 复用 coze-studio `passport` domain
- 提供 `space_id` + `user_id` 上下文

**Flutter 端**:
```
lib/src/core/services/
├── auth_service.dart          # 接口
├── auth_service_impl.dart     # 实现
└── api_client.dart            # HTTP 基础客户端（含 token 注入）
```

**Provider**:
```dart
final apiClientProvider = Provider<ApiClient>((ref) => ApiClient(baseUrl: env.apiBaseUrl));
final currentUserProvider = FutureProvider<UserInfo>((ref) async { ... });
final currentSpaceIdProvider = Provider<String>((ref) => ...);
```

---

### 2. WorkerService（AI 员工服务）

**复用范围**: #001, #002, #003, #004, #005, #023, #024

**后端 Go 包**: `domain/opc_worker/`

**核心接口**:
```go
type WorkerService interface {
    ListWorkers(ctx context.Context, spaceID int64, filter *WorkerFilter) ([]*Worker, int64, error)
    GetWorkerDetail(ctx context.Context, spaceID int64, workerID string) (*WorkerDetail, error)
    UpdateWorkerConfig(ctx context.Context, spaceID int64, workerID string, config *WorkerConfig) error
    GetWorkerStats(ctx context.Context, spaceID int64, workerID string, period string) (*WorkerStats, error)
}
```

**Flutter 端**:
```
lib/src/shared/services/worker/
├── worker_repository.dart           # 接口
├── worker_api_repository.dart       # API 实现
└── worker_cache_repository.dart     # 缓存装饰器
```

**Provider**:
```dart
final workerRepositoryProvider = Provider<WorkerRepository>((ref) {
  final client = ref.watch(apiClientProvider);
  return WorkerApiRepository(client);
});

// 员工列表（多页面复用）
final workerListProvider = FutureProvider.autoDispose<List<WorkerBrief>>((ref) async {
  final repo = ref.watch(workerRepositoryProvider);
  final spaceId = ref.watch(currentSpaceIdProvider);
  return repo.listWorkers(spaceId: spaceId);
});
```

**复用场景**:
| Issue | 使用的方法 |
|-------|-----------|
| #001 首页 | `listWorkers` (简要列表) |
| #002 员工列表 | `listWorkers` (完整列表+搜索) |
| #003 员工详情 | `getWorkerDetail` |
| #005 简报 | `getWorkerStats` (当日统计) |
| #023 协作 | `listWorkers` (参与者信息) |

---

### 3. TaskService（任务服务）

**复用范围**: #001, #003, #005, #013, #034

**后端 Go 包**: `domain/opc_worker/` (子模块 task)

**核心接口**:
```go
type TaskService interface {
    ListTasks(ctx context.Context, spaceID int64, filter *TaskFilter) ([]*Task, int64, error)
    GetTaskDetail(ctx context.Context, taskID int64) (*TaskDetail, error)
    CreateTask(ctx context.Context, task *Task) (int64, error)
    UpdateTaskStatus(ctx context.Context, taskID int64, status TaskStatus) error
    GetDailyTaskStats(ctx context.Context, spaceID int64, date string) (*DailyTaskStats, error)
}
```

**Flutter 端**:
```
lib/src/shared/services/task/
├── task_repository.dart
└── task_api_repository.dart
```

**复用场景**:
| Issue | 使用的方法 |
|-------|-----------|
| #001 首页 | `ListTasks` (最近任务) |
| #003 员工详情 | `ListTasks` (员工任务历史) |
| #005 简报 | `GetDailyTaskStats` |
| #013 项目 | `ListTasks` (项目关联任务) |
| #034 手机任务 | `ListTasks` (手机任务) |

---

### 4. CreditsService（积分服务）

**复用范围**: #001, #005, #016, #021

**后端 Go 包**: `domain/opc_finance/` (子模块 credits)

**核心接口**:
```go
type CreditsService interface {
    GetBalance(ctx context.Context, spaceID, userID int64) (*CreditsBalance, error)
    ListTransactions(ctx context.Context, spaceID, userID int64, filter *TxFilter) ([]*Transaction, int64, error)
    GetUsageByWorker(ctx context.Context, spaceID, userID int64) ([]*WorkerUsage, error)
}
```

**Flutter 端**:
```
lib/src/shared/services/credits/
├── credits_repository.dart
└── credits_api_repository.dart
```

**复用场景**:
| Issue | 使用的方法 |
|-------|-----------|
| #001 首页 | `GetBalance` (积分概览) |
| #005 简报 | `GetUsageByWorker` (当日消耗) |
| #016 积分管理 | 全部方法 |
| #021 财务概览 | `GetBalance` |

---

### 5. PhoneService（手机服务）

**复用范围**: #001, #032, #033, #034

**后端 Go 包**: `domain/opc_phone/`

**核心接口**:
```go
type PhoneService interface {
    GetDeviceStatus(ctx context.Context, spaceID, userID int64) (*DeviceStatus, error)
    GetCurrentTask(ctx context.Context, deviceID string) (*PhoneTask, error)
    ListRecentTasks(ctx context.Context, deviceID string, limit int) ([]*PhoneTask, error)
    UpdateDeviceSettings(ctx context.Context, deviceID string, settings *DeviceSettings) error
    PauseTask(ctx context.Context, taskID int64) error
    AbortTask(ctx context.Context, taskID int64) error
}
```

**Flutter 端**:
```
lib/src/shared/services/phone/
├── phone_repository.dart
└── phone_api_repository.dart
```

**复用场景**:
| Issue | 使用的方法 |
|-------|-----------|
| #001 首页 | `GetDeviceStatus` (Phone 卡片) |
| #032 手机OS | `GetDeviceStatus` + `GetCurrentTask` + `ListRecentTasks` |
| #033 手机编辑 | `UpdateDeviceSettings` |
| #034 手机任务 | `GetCurrentTask` + `ListRecentTasks` + `PauseTask` + `AbortTask` |

---

### 6. ContentService（内容服务）

**复用范围**: #010, #011

**后端 Go 包**: `domain/opc_content/`

**核心接口**:
```go
type ContentService interface {
    GetOverview(ctx context.Context, spaceID, userID int64) (*ContentOverview, error)
    ListContent(ctx context.Context, spaceID, userID int64, filter *ContentFilter) ([]*Content, int64, error)
    ListPendingReview(ctx context.Context, spaceID, userID int64) ([]*Content, error)
    ApproveContent(ctx context.Context, contentID int64, note string) error
    RejectContent(ctx context.Context, contentID int64, note string) error
    ScheduleContent(ctx context.Context, contentID int64, scheduledAt int64) error
}
```

---

### 7. CrmService（客户服务）

**复用范围**: #012

**后端 Go 包**: `domain/opc_crm/`

**核心接口**:
```go
type CrmService interface {
    GetOverview(ctx context.Context, spaceID, userID int64) (*CrmOverview, error)
    ListCustomers(ctx context.Context, spaceID, userID int64, filter *CustomerFilter) ([]*Customer, int64, error)
    GetPipeline(ctx context.Context, spaceID, userID int64) ([]*PipelineStage, error)
    ListFollowUps(ctx context.Context, spaceID, userID int64) ([]*FollowUp, error)
    CompleteFollowUp(ctx context.Context, followUpID int64) error
}
```

---

### 8. FinanceService（财务服务）

**复用范围**: #021, #022

**后端 Go 包**: `domain/opc_finance/`

**核心接口**:
```go
type FinanceService interface {
    GetBalanceOverview(ctx context.Context, spaceID, userID int64) (*BalanceOverview, error)
    GetRevenueChart(ctx context.Context, spaceID, userID int64, period string) ([]*ChartBar, error)
    ListIncomeSources(ctx context.Context, spaceID, userID int64) ([]*IncomeSource, error)
    ListExpenses(ctx context.Context, spaceID, userID int64, filter *ExpenseFilter) ([]*Expense, int64, error)
    ListEscrows(ctx context.Context, spaceID, userID int64) ([]*Escrow, error)
    RequestWithdrawal(ctx context.Context, escrowID int64, amount int64) error
}
```

---

### 9. MarketplaceService（市场服务）

**复用范围**: #018, #019, #020

**后端 Go 包**: `domain/opc_marketplace/`

**核心接口**:
```go
type MarketplaceService interface {
    ListDemands(ctx context.Context, filter *DemandFilter) ([]*Demand, int64, error)
    GetDemandDetail(ctx context.Context, demandID int64) (*DemandDetail, error)
    ListMyApplications(ctx context.Context, userID int64) ([]*Application, error)
    ApplyToDemand(ctx context.Context, app *Application) (int64, error)
    GetEnterpriseOverview(ctx context.Context, userID int64) (*EnterpriseOverview, error)
}
```

---

### 10. AutomationService（自动化服务）

**复用范围**: #025, #026, #027, #028

**后端 Go 包**: `domain/opc_automation/` + 复用 coze-studio `workflow`/`plugin` domain

**核心接口**:
```go
type AutomationService interface {
    GetOverview(ctx context.Context, spaceID int64) (*AutomationOverview, error)
    ListWorkflows(ctx context.Context, spaceID int64) ([]*Workflow, error)
    ToggleWorkflow(ctx context.Context, workflowID int64, enabled bool) error
    // Coze 工作流 — 代理到 coze-studio workflow domain
    ListCozeAgentWorkflows(ctx context.Context, spaceID int64) ([]*CozeAgentWorkflow, error)
    // ComfyUI — 独立实现
    ListComfyUITemplates(ctx context.Context, spaceID int64) ([]*ComfyUITemplate, error)
    GenerateComfyUI(ctx context.Context, req *ComfyUIGenerateReq) (*ComfyUIGeneration, error)
    // MCP — 代理到 coze-studio plugin domain
    ListConnectedMCPTools(ctx context.Context, spaceID int64) ([]*MCPTool, error)
    ListRecommendedMCPTools(ctx context.Context, spaceID int64) ([]*MCPTool, error)
    ConnectMCPTool(ctx context.Context, toolID int64) error
    DisconnectMCPTool(ctx context.Context, toolID int64) error
}
```

---

### 11. NotificationService（通知服务）

**复用范围**: #009, #030

**后端 Go 包**: `domain/opc_notification/`

**核心接口**:
```go
type NotificationService interface {
    ListNotifications(ctx context.Context, spaceID, userID int64, filter *NotifFilter) ([]*Notification, int64, error)
    MarkRead(ctx context.Context, notifIDs []int64) error
    MarkAllRead(ctx context.Context, spaceID, userID int64) error
    GetUnreadCount(ctx context.Context, spaceID, userID int64) (int64, error)
}
```

---

### 12. QualityService（质量服务）

**复用范围**: #035, #036, #037

**后端 Go 包**: `domain/opc_quality/`

**核心接口**:
```go
type QualityService interface {
    GetCreditScore(ctx context.Context, spaceID, userID int64) (*CreditScore, error)
    GetAIQualityOverview(ctx context.Context, spaceID, userID int64) (*AIQualityOverview, error)
    GetComplianceStatus(ctx context.Context, spaceID, userID int64) (*ComplianceStatus, error)
    ListContracts(ctx context.Context, spaceID, userID int64) ([]*Contract, error)
    ListIPItems(ctx context.Context, spaceID, userID int64) ([]*IPItem, error)
}
```

---

## Flutter 端共享层目录结构

```
lib/src/
├── core/
│   ├── services/
│   │   ├── api_client.dart              # HTTP 基础客户端
│   │   ├── api_client_provider.dart     # ApiClient Provider
│   │   └── auth_service.dart            # 认证服务
│   └── config/
│       └── env_config.dart              # 环境配置（API base URL）
├── shared/
│   ├── models/
│   │   ├── api_response.dart            # 统一响应模型 {code, msg, data}
│   │   ├── pagination.dart              # 分页模型 {page, page_size, total}
│   │   └── opc_enums.dart               # 共享枚举（OpcTagColor, WorkerStatus 等）
│   ├── repositories/
│   │   ├── worker_repository.dart       # AI 员工
│   │   ├── task_repository.dart         # 任务
│   │   ├── credits_repository.dart      # 积分
│   │   ├── phone_repository.dart        # 手机
│   │   ├── content_repository.dart      # 内容
│   │   ├── crm_repository.dart          # CRM
│   │   ├── finance_repository.dart      # 财务
│   │   ├── marketplace_repository.dart  # 市场
│   │   ├── automation_repository.dart   # 自动化
│   │   ├── notification_repository.dart # 通知
│   │   └── quality_repository.dart      # 质量
│   └── providers/
│       ├── worker_providers.dart        # 跨页面复用的 Worker Provider
│       ├── credits_providers.dart       # 跨页面复用的 Credits Provider
│       ├── phone_providers.dart         # 跨页面复用的 Phone Provider
│       └── user_providers.dart          # 用户信息 Provider
└── features/
    └── ... (各 feature 模块引用 shared/ 下的 repository 和 provider)
```

---

## 后端共享层目录结构

```
coze-studio/backend/
├── domain/
│   ├── opc_worker/
│   │   ├── entity.go           # Worker, WorkerConfig, WorkerStats
│   │   ├── repository.go       # WorkerRepository 接口
│   │   ├── service.go          # WorkerService 实现
│   │   └── task/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       └── service.go
│   ├── opc_finance/
│   │   ├── credits/
│   │   ├── subscription/
│   │   ├── escrow/
│   │   └── record/
│   ├── opc_content/
│   ├── opc_crm/
│   ├── opc_marketplace/
│   ├── opc_phone/
│   ├── opc_automation/
│   ├── opc_notification/
│   ├── opc_quality/
│   └── opc_profile/
├── application/
│   ├── opc_home/               # 首页聚合服务
│   ├── opc_briefing/           # 简报聚合服务
│   ├── opc_dashboard/          # 看板聚合服务
│   └── opc_onboarding/         # 引导服务
├── api/
│   ├── handler/coze/
│   │   ├── opc_home_service.go
│   │   ├── opc_worker_service.go
│   │   ├── opc_finance_service.go
│   │   └── ...
│   └── model/opc/              # OPC API 请求/响应模型
├── infra/
│   └── opc/                    # OPC 数据库实现
│       ├── worker_repo_impl.go
│       ├── customer_repo_impl.go
│       └── ...
```

---

## 缓存策略

| 数据 | 缓存位置 | TTL | 失效策略 |
|------|---------|-----|---------|
| 用户信息 | Flutter 本地 + Redis | 5min | 修改时失效 |
| 员工列表 | Flutter 本地 + Redis | 1min | 状态变更时失效 |
| 积分余额 | Redis | 30s | 交易时失效 |
| 通知未读数 | Redis | 10s | 新通知时失效 |
| 设备状态 | Redis | 5s | 心跳更新 |
| 信用评分 | Redis | 1h | 每日重算 |
| 看板数据 | Redis | 30s | 周期切换时重新请求 |
