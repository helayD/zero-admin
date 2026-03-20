# Code Reuse Patterns

37 个 issue 的代码复用规范。避免冗余，提取公共模块。

---

## 原则

1. **共享模型上提** — 被 2+ issue 使用的 Model 移至 `shared/models/`
2. **共享 Repository 上提** — 被 2+ issue 使用的 Repository 移至 `shared/repositories/`
3. **共享 Provider 上提** — 被 2+ issue 使用的 Provider 移至 `shared/providers/`
4. **Feature 内聚** — 仅被 1 个 issue 使用的代码留在 `features/{module}/`
5. **不重复设计 API** — 同一共享服务组内，API 只在「首次定义」issue 中完整设计

---

## 1. 已有共享代码（现状）

### OPC 共享 Widget
- **位置**: `features/home/presentation/widgets/opc_shared_widgets.dart`
- **组件**: `OPCTag`, `OPCAvatar`, `OPCIconBox`, `OPCSectionHeader`, `OPCStatValue`, `OPCSkeletonBox`
- **使用者**: 全部 37 个 issue
- **改造**: 移至 `shared/widgets/opc/` (未来)

### OPC 共享枚举
- **位置**: `features/home/data/models/opc_home_models.dart`
- **枚举**: `OpcTagColor` (green/blue/purple/amber/rose/cyan/orange/gray), `WorkerStatus`
- **使用者**: 全部 37 个 issue
- **改造**: 移至 `shared/models/opc_enums.dart`

---

## 2. 新增共享层（目标结构）

```
lib/src/
├── core/
│   └── network/
│       ├── api_client.dart              # HTTP 基础客户端（token 注入、错误拦截）
│       ├── api_client_provider.dart     # ApiClient Riverpod Provider
│       ├── api_response.dart            # 统一响应 { code, msg, data }
│       └── api_exception.dart           # 业务异常（含错误码）
├── shared/
│   ├── models/
│   │   ├── opc_enums.dart              # OpcTagColor, WorkerStatus 等（从 home 迁出）
│   │   └── pagination.dart             # PaginatedList<T> { items, total, page, pageSize }
│   ├── repositories/
│   │   ├── worker_repository.dart      # WorkerService 共享 Repository
│   │   ├── task_repository.dart        # TaskService 共享 Repository
│   │   ├── credits_repository.dart     # CreditsService 共享 Repository
│   │   └── phone_repository.dart       # PhoneService 共享 Repository
│   ├── providers/
│   │   ├── user_providers.dart         # currentUserProvider, currentSpaceIdProvider
│   │   ├── worker_providers.dart       # workerListProvider (多页面复用)
│   │   ├── credits_providers.dart      # creditsBalanceProvider (多页面复用)
│   │   └── phone_providers.dart        # phoneStatusProvider (多页面复用)
│   └── widgets/
│       └── opc/                        # OPCTag 等共享组件（从 home 迁出）
└── features/
    └── {module}/
        ├── data/
        │   ├── models/                 # Feature 专属 Model（含 fromJson/toJson）
        │   └── repositories/           # Feature 专属 Repository（如果不共享）
        ├── state/                      # Feature 专属 Provider
        └── presentation/
```

---

## 3. 复用模式

### Pattern A: 共享 Repository + Feature Provider

当多个 issue 需要同一数据源，但展示逻辑不同时：

```dart
// shared/repositories/worker_repository.dart
abstract class WorkerRepository {
  Future<List<WorkerBrief>> listWorkers({required String spaceId, String? keyword});
  Future<WorkerDetail> getWorkerDetail({required String spaceId, required String workerId});
}

// features/home/ — 只取前 5 个员工
final homeWorkersProvider = FutureProvider.autoDispose<List<WorkerBrief>>((ref) async {
  final repo = ref.watch(workerRepositoryProvider);
  final spaceId = ref.watch(currentSpaceIdProvider);
  final all = await repo.listWorkers(spaceId: spaceId);
  return all.take(5).toList();
});

// features/ai_workers/ — 完整列表 + 搜索
final workersPageProvider = FutureProvider.autoDispose<List<WorkerBrief>>((ref) async {
  final repo = ref.watch(workerRepositoryProvider);
  final spaceId = ref.watch(currentSpaceIdProvider);
  final keyword = ref.watch(workerSearchProvider);
  return repo.listWorkers(spaceId: spaceId, keyword: keyword);
});
```

**适用 issue**: #001↔#002 (员工列表), #001↔#032 (Phone 状态), #001↔#016 (积分)

---

### Pattern B: 聚合 API + 子服务拆分

当一个页面需要多个域的数据时（如首页）：

```dart
// 后端: 聚合 API 一次返回
// POST /api/opc/v1/home/page-data → { workers, credits, phone, tasks, ... }

// Flutter 端: 一个 Provider 获取聚合数据
final homePageDataProvider = FutureProvider.autoDispose<HomePageData>((ref) async {
  final repo = ref.watch(homeRepositoryProvider);
  return repo.getPageData(spaceId: ..., userId: ...);
});

// 但子数据也可以通过共享 Provider 单独刷新
// 例如积分变化后，只刷新积分部分
final creditsBalanceProvider = FutureProvider.autoDispose<CreditsBalance>((ref) async {
  final repo = ref.watch(creditsRepositoryProvider);
  return repo.getBalance(spaceId: ..., userId: ...);
});
```

**适用 issue**: #001 (首页聚合), #005 (简报聚合), #006 (看板聚合)

---

### Pattern C: 同模块内复用

同一 feature 模块内的多个 issue 共享 Model 和 Repository：

```dart
// features/finance/ 下 #016(积分) + #021(财务) + #022(托管) 共享:
// - data/models/opc_finance_common.dart  — 共享枚举和基础模型
// - data/repositories/finance_repository.dart — 共享 Repository

// #016 专属
// - data/models/opc_credits_models.dart
// - state/opc_credits_providers.dart

// #021 专属
// - data/models/opc_finance_models.dart
// - state/opc_finance_providers.dart
```

**适用 issue**: #010↔#011 (内容), #018↔#019↔#020 (市场), #032↔#033↔#034 (手机)

---

### Pattern D: coze-studio 服务代理

当 OPC 功能直接映射到 coze-studio 已有服务时：

```dart
// 后端 Go: OPC handler 代理到 coze-studio domain
func GetCozeWorkflows(ctx context.Context, c *app.RequestContext) {
    // 调用 coze-studio 已有的 workflow domain
    workflows, err := workflowSVC.ListWorkflows(ctx, spaceID)
    // 转换为 OPC 响应格式
    resp := convertToOPCResponse(workflows)
    c.JSON(200, resp)
}

// Flutter 端: 正常调用 OPC API（不直接调 coze-studio API）
final cozeWorkflowsProvider = FutureProvider.autoDispose<CozeWorkflowsPageData>((ref) async {
  final repo = ref.watch(automationRepositoryProvider);
  return repo.getCozeWorkflows(spaceId: ...);
});
```

**适用 issue**: #026 (Coze 工作流), #028 (MCP 工具), #029 (知识库), #004 (聊天), #024 (记忆)

---

## 4. fromJson/toJson 规范

所有需要与 API 交互的 Model 必须实现序列化：

```dart
class WorkerBrief {
  final String id;
  final String name;
  final String title;
  final String avatarChar;
  final OpcTagColor avatarColor;
  final WorkerStatus status;
  final String currentTask;

  const WorkerBrief({
    required this.id,
    required this.name,
    required this.title,
    required this.avatarChar,
    required this.avatarColor,
    required this.status,
    required this.currentTask,
  });

  factory WorkerBrief.fromJson(Map<String, dynamic> json) => WorkerBrief(
    id: json['id'] as String,
    name: json['name'] as String,
    title: json['title'] as String,
    avatarChar: json['avatar_char'] as String,
    avatarColor: OpcTagColor.values.byName(json['avatar_color'] as String),
    status: WorkerStatus.values.byName(json['status'] as String),
    currentTask: json['current_task'] as String? ?? '',
  );

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'title': title,
    'avatar_char': avatarChar,
    'avatar_color': avatarColor.name,
    'status': status.name,
    'current_task': currentTask,
  };
}
```

**命名映射规则**: Dart `camelCase` ↔ API `snake_case`

---

## 5. Error Handling 规范

```dart
// core/network/api_exception.dart
class ApiException implements Exception {
  final int code;
  final String message;
  ApiException(this.code, this.message);
}

// core/network/api_client.dart
class ApiClient {
  Future<ApiResponse<T>> post<T>(String path, {Map<String, dynamic>? body, T Function(dynamic)? fromJson}) async {
    final response = await _dio.post(path, data: body);
    final apiResp = ApiResponse.fromJson(response.data);
    if (apiResp.code != 0) {
      throw ApiException(apiResp.code, apiResp.msg);
    }
    return apiResp;
  }
}

// Feature Provider 中的错误处理由 Riverpod AsyncValue 自动管理
// Page 中的三态展示:
asyncData.when(
  data: (data) => _Content(data: data),
  loading: () => _LoadingSkeleton(),
  error: (error, _) => _ErrorView(
    message: error is ApiException ? error.message : '网络错误',
    onRetry: () => ref.invalidate(provider),
  ),
);
```

---

## 6. 冗余检测清单

在完善每个 issue 时，检查以下冗余：

| 检查项 | 说明 | 处理方式 |
|--------|------|---------|
| 重复 Model | 同一实体在多个 feature 中定义 | 提取到 shared/models/ |
| 重复 Provider | 同一数据在多个 feature 中获取 | 提取到 shared/providers/ |
| 重复 API 设计 | 同一 endpoint 在多个 issue 中描述 | 只在首次定义 issue 中完整描述 |
| 重复 Widget | 同一 UI 组件在多个 feature 中实现 | 提取到 shared/widgets/ |
| 重复枚举 | 同一枚举在多个 Model 文件中定义 | 统一到 shared/models/opc_enums.dart |
