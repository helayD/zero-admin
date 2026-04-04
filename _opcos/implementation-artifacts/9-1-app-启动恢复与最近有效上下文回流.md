# Story 9.1: App 启动恢复与最近有效上下文回流

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 消费者,
I want 在冷启动、热启动、登录态恢复和前后台切换后回到最近一次有效上下文,
so that 我不用重新寻找刚才正在处理的商品、专题、购物车或订单。

## Acceptance Criteria

1. **Given** App 在首页、专题、商品详情、购物车、待支付订单或订单详情页被系统回收或切到后台 **When** 用户重新打开 App **Then** 系统优先恢复最近一次有效目标上下文 **And** 若原目标已失效，3 秒内提供替代落点或明确失败反馈。

2. **Given** 登录态需要恢复 **When** 用户完成登录恢复 **Then** 系统返回原始目标上下文 **And** 不把用户静默送回首页。

## AC 覆盖矩阵

- **AC1（最近有效上下文恢复 + 失效目标回退 + 3 秒反馈）**：Task 1-4、5.1-5.4
- **AC2（登录恢复后返回原目标 + 不静默回首页）**：Task 2.4、3.4-3.6、4.1-4.6、5.2-5.4

## 背景与关键风险（必须优先阅读）

> **⚠️ CRITICAL - 当前 Flutter 端没有统一启动恢复壳层：** `flutter-mall/lib/welcome.dart` 目前仅等待 1 秒后直接进入 `MainTab`，没有根据最近有效上下文、登录态、页面失效状态或恢复目标做决策。9.1 必须先建立统一恢复入口，而不是在首页、购物车、订单页各自加一段启动判断。
>
> **⚠️ CRITICAL - 当前导航体系以 `MaterialPageRoute` 为主，不能把恢复方案建立在字符串路由名上：** 仓库里大量页面通过 `Navigator.of(context).push(MaterialPageRoute(...))` 进入，例如 `HomePage`、`Cart`、`ProductDetail`、`OrderDetail`；但 `MaterialApp` 里没有统一 `routes` / `onGenerateRoute`，`HttpUtil` 与 `Login` 的 `redirectRoute` 仅能覆盖非常有限的命名路由场景。9.1 必须使用“语义化上下文对象”恢复目标，而不是继续放大字符串路由依赖。
>
> **⚠️ CRITICAL - 当前 401 登录恢复链存在断层：** `flutter-mall/lib/utils/http_util.dart` 在 401 时会取 `ModalRoute.of(navContext)?.settings.name` 作为 `redirectRoute`，而当前多数页面没有命名路由；`flutter-mall/lib/view/mine/login/login.dart` 登录成功后又调用 `pushReplacementNamed(widget.redirectRoute!)`。如果直接沿用该模式，很多页面会在登录恢复后丢失目标上下文或恢复失败。
>
> **⚠️ CRITICAL - `NavKey` 取到的不是当前业务页 Route context：** `flutter-mall/lib/utils/http_util.dart` 通过 `NavKey.navKey.currentState!.context` 再调用 `ModalRoute.of(...).settings.name` 推断当前路由，但这个 context 属于全局 Navigator，不是具体业务页面的 route context，恢复目标很可能为空或不可靠。9.1 必须把 pending recovery intent / recent context 作为单一真相源，而不是继续试图修补这条字符串路由链。
>
> **⚠️ CRITICAL - 只能恢复“最近一次有效上下文”，不能恢复无效页或失败态：** `ProductDetail`、`OrderDetail` 这类页面都可能因为商品下架、订单状态变化、接口失败而失效。9.1 只能在页面已拿到可用服务端数据后记录上下文；接口失败、商品不可见、订单无效、占位页或临时页面都不能覆盖掉最近一次有效上下文。
>
> **⚠️ CRITICAL - `SharedPreferencesUtil` 目前只有基础类型读写能力：** `flutter-mall/lib/utils/shared_preferences_util.dart` 只支持字符串、数字和布尔值，尚无结构化 recent context 的版本控制、序列化、校验或清理逻辑。9.1 应在此基础上扩展最小可用的结构化存储，而不是绕开它再私建一套本地存储设施。
>
> **⚠️ CRITICAL - 本 Story 只恢复已有页面或已有业务语义：** 当前仓库可确认的稳定目标至少包括 `HomePage`、`Cart`、`OrderList`、`ProductDetail`、`OrderDetail`，以及已具备 `resumed` 状态重查能力的 `OrderPay`。Epic/PRD 虽然提到“专题”“活动”等上下文，但如果目标页面或统一入口尚未在当前 Flutter 仓库落地，9.1 必须明确降级到已存在页面或给出失败反馈，不能虚构页面路由。
>
> **⚠️ CRITICAL - 退出登录当前只移除了 token，没有清理恢复状态：** `flutter-mall/lib/view/mine/setting/settings.dart` 目前只执行 `prefs.remove(token)` 后返回上一页，未清理 recent context / pending recovery intent。9.1 必须把“账号切换 / 退出登录 -> 清理恢复目标”设为硬约束，避免新账号被带回旧账号的订单或商品上下文。
>
> **⚠️ CRITICAL - `main()` 初始化顺序不能被破坏：** `flutter-mall/lib/main.dart` 已先执行 `WidgetsFlutterBinding.ensureInitialized()`，再执行 `SharedPreferencesUtil.init()`，最后 `runApp()`。根据项目上下文，这个顺序必须保留，恢复壳层不能为了图省事把 SharedPreferences 初始化挪到 Widget 生命周期里。

## 明确非目标（本 Story 不做）

- 不在本 Story 中实现 FR61 的统一加载 / 空态 / 错误态壳层；该能力由 9.2 落地。
- 不在本 Story 中实现 FR62 的消息、活动、优惠券统一意图唤回全链路；9.1 只为该链路提供可复用的最近有效上下文恢复底座。
- 不在本 Story 中实现 FR63 权限请求或 FR64 升级闸门；但恢复设计必须为后续权限 / 升级恢复预留统一入口。
- 不重写整个 Flutter 路由体系，不强行把现有所有页面一次性改造成命名路由。
- 不恢复未提交表单、评价输入、售后草稿等高风险临时编辑态；本 Story 只恢复“最近一次有效页面上下文”。
- 不创建新的网络层、鉴权层或本地存储框架，继续复用 `HttpUtil`、`SharedPreferencesUtil`、现有 Provider 与页面结构。

## 实施顺序（必须按序推进）

1. 先定义“最近有效上下文”的最小语义对象、允许恢复的目标类型、持久化格式和失效校验规则。
2. 再把启动入口从 `Welcome -> MainTab` 的固定跳转改为统一恢复壳层决策，接入冷启动、热启动和前后台切换。
3. 再为首页、购物车、商品详情、订单详情等真实目标页面补“有效上下文写入点”，确保只在服务端状态可用时落库。
4. 最后收口登录恢复、401 场景和兜底回退，让恢复结果统一回到语义目标，而不是回到零散字符串路由。

## Tasks / Subtasks

### 一、定义最近有效上下文契约与本地持久化边界

#### Task 1 (AC: 1): 设计统一的 recent context 语义对象、允许恢复目标和存储规则

**目标：** 先把“恢复什么、何时写入、何时失效、如何回退”定义清楚，再改启动与页面。

- [ ] 1.1 在 `flutter-mall/lib/` 中新增最小上下文模型与辅助层，建议路径为：
  - `lib/model/app_recent_context.dart`
  - `lib/layout/app_bootstrap.dart`
  - `lib/layout/intent_recovery_shell.dart`
  - `lib/provider/app_lifecycle_provider.dart`
- [ ] 1.2 recent context 至少包含：`version`、`targetType`、`targetId`、`tabIndex`、`source`、`requiresAuth`、`capturedAt`、`lastValidatedAt`。只存语义化目标，不存 Widget 实例、BuildContext、完整路由栈或大块页面 JSON。
- [ ] 1.3 `targetType` 首批至少允许使用当前仓库中已存在且能稳定恢复的目标：`home`、`cart`、`order_list`、`product_detail`、`order_detail`。其中 `order_list` 必须支持 `tabIndex=1` 表示“待支付”恢复入口。若需要纳入 `order_pay`，必须同时持有 `orderId/orderSn`，并在恢复后立即重查服务端状态。`subject`、`activity` 等若暂无稳定入口，只能通过 fallback 语义描述，不得先写死不存在的页面实现。
- [ ] 1.4 为 recent context 建立版本化与最小校验规则：版本不兼容、目标类型未知、目标 ID 非法、关键字段缺失时，必须直接判定为不可恢复并走 fallback，而不是带着脏数据继续跳转。
- [ ] 1.5 recent context 只允许在“页面已拿到有效服务端数据或目标已确认可展示”时写入；请求失败页、商品不可见页、订单无效页、空白占位页、临时弹层页不得覆盖已有有效上下文。
- [ ] 1.6 在 `SharedPreferencesUtil` 基础上补结构化读写 helper，例如 `saveJsonString/getJsonString` 或专用 recent-context helper；不要引入新存储依赖。
- [ ] 1.7 recent context 清理规则必须明确：注销登录、账号切换、显式清空缓存、目标已确认失效时清理或替换；`flutter-mall/lib/view/mine/setting/settings.dart` 的退出登录逻辑不能只 `remove(token)`，必须一并清理 pending recovery intent / recent context，必要时重置账号绑定的本地恢复状态。不能让旧账号的订单 / 商品上下文泄露到新账号恢复链路。

### 二、建立统一启动恢复壳层与生命周期监听

#### Task 2 (AC: 1, 2): 把启动、前后台切换和登录恢复统一收口到 App Lifecycle Shell

**目标：** 不再由 `Welcome` 或零散页面各自决定恢复逻辑，而是通过单一壳层完成恢复决策。

- [ ] 2.1 保持 `main.dart` 中 `WidgetsFlutterBinding.ensureInitialized()` -> `SharedPreferencesUtil.init()` -> `runApp()` 的顺序不变；恢复壳层只能建立在该初始化完成之后。
- [ ] 2.2 改造 `flutter-mall/lib/welcome.dart` 的固定 1 秒跳转逻辑，使其进入 `AppBootstrap` / `IntentRecoveryShell`，由壳层在 3 秒预算内决定：恢复最近有效上下文、回到 `MainTab`、或展示明确失败反馈。
- [ ] 2.3 新增统一生命周期监听（如 `WidgetsBindingObserver` + `AppLifecycleProvider`），覆盖至少 `paused` / `resumed` 场景；前后台切换时记录候选上下文，恢复时重新验证，而不是页面各自 ad-hoc 处理。
- [ ] 2.4 恢复链必须兼容“需要先恢复登录态”的场景：壳层先判断 token / 会员状态，再决定是否进入登录流程；登录完成后回到原始目标，不允许静默落回首页。
- [ ] 2.5 若 recent context 已失效，壳层必须在 3 秒内给出明确结果：
  - 可替代时进入 `home` 或对应 tab 目标；
  - 不可替代时展示可理解反馈并保留“返回首页 / 重试”的动作。
- [ ] 2.6 不要把恢复逻辑分散写进 `HomePage`、`Cart`、`ProductDetail`、`OrderDetail` 的 `initState()` 里相互竞争；统一由壳层串联，页面只负责报告自己可否被记录为“有效上下文”。

### 三、为真实目标页面补充“有效上下文”写入点与失效校验

#### Task 3 (AC: 1): 只对已确认有效的真实目标页记录上下文

- [ ] 3.1 `HomePage` 在首页内容成功加载或首页根视图确认可展示后，记录 `targetType=home` 与当前 tab 语义；不要在接口尚未返回时立即写入。
- [ ] 3.2 `Cart` 在购物车列表和关键促销数据成功返回后记录 `targetType=cart`；请求失败、空白页或未登录异常状态不能覆盖最近有效上下文。
- [ ] 3.3 `ProductDetail` 仅在商品详情接口成功且 `visibility.visible == true` 时记录 `targetType=product_detail` + `productId`；`reasonCode=request_failed`、商品不可见或 fallback 展示时不得覆盖 recent context。
- [ ] 3.4 `OrderDetail` 仅在订单详情成功拉取且目标订单可访问时记录 `targetType=order_detail` + `orderId`；若订单无效、越权、接口失败或状态已失效，则必须触发 fallback 语义而不是强留旧目标。
- [ ] 3.5 `OrderList` 已是稳定入口，必须支持记录 `targetType=order_list` + `tabIndex`；其中 `tabIndex=1` 对应“待支付”恢复入口，满足 FR60 的待支付订单场景，不应再把该能力误判为仓库缺失。
- [ ] 3.6 `OrderPay` 当前已经通过 `didChangeAppLifecycleState(AppLifecycleState.resumed)` 重新查询支付状态；9.1 若把支付页纳入恢复目标，必须复用这条“恢复后先重查服务端状态”的模式，并在订单已支付 / 已取消 / 已过期时回退到 `OrderDetail` 或 `OrderList(tabIndex=1)`，而不是停留在陈旧支付页。
- [ ] 3.7 记录上下文时统一透传 `source`（如 `manual_open`、`login_restore`、`message`、`resume`），为后续 9.3 意图唤回埋点和恢复分析做准备。

### 四、收口登录恢复、401 场景与现有导航断层

#### Task 4 (AC: 2): 用语义化恢复对象替换脆弱的字符串路由恢复

- [ ] 4.1 改造 `flutter-mall/lib/utils/http_util.dart` 的 401 恢复流程，不再依赖 `NavKey.navKey.currentState!.context` + `ModalRoute.of(...).settings.name` 作为唯一恢复依据；401 时应保存 pending recovery intent / recent context，并跳转登录页。
- [ ] 4.2 改造 `flutter-mall/lib/view/mine/login/login.dart` 登录成功后的恢复逻辑：优先消费 pending recovery intent / recent context，若无明确目标再走 `pop(true)` 或首页回退；`redirectRoute` 只能作为兼容兜底，不能继续对不存在的命名路由执行 `pushReplacementNamed`。
- [ ] 4.3 `flutter-mall/lib/view/mine/message/message.dart` 当前同时存在 `MaterialPageRoute` 与 `pushNamed('/product')` / `pushNamed('/activity')` 混用，但 `MaterialApp` 中未定义全局 `routes` / `onGenerateRoute`。9.1 不要求把消息链全部做完，但必须明确恢复底座不依赖该未收口的命名路由体系。
- [ ] 4.4 恢复对象统一使用语义化字段（`targetType/targetId/fallbackType/requiresAuth`），不要继续在登录恢复链或启动链里传裸 route name、页面标题字符串或局部 query 片段。
- [ ] 4.5 防止重复登录 / 重复恢复：连续 401、前后台切换和冷启动恢复不能同时弹多个登录页或重复 push 同一目标页。
- [ ] 4.6 `flutter-mall/lib/view/mine/setting/settings.dart` 退出登录时，除了移除 `token` 外，还必须清理 pending recovery intent / recent context，并确保退出后不会自动恢复到旧账号的订单、购物车或商品目标。

### 五、测试、验收与恢复成功率观测

#### Task 5 (AC: 1, 2): 为最近有效上下文恢复补足单测、Widget 测试与关键 UAT

- [ ] 5.1 `flutter test` 至少覆盖 recent context helper：
  - 非法 JSON / 缺字段 / 版本不兼容时返回不可恢复
  - `targetType` 非白名单时拒绝恢复
  - 账号切换或显式清理时 recent context 被清空
- [ ] 5.2 Widget / 页面级测试至少覆盖：
  - `Welcome/AppBootstrap` 能在 recent context 可用时优先进入对应目标
  - 登录恢复完成后回到 `order_list(tabIndex=1)`、`product_detail` 或 `order_detail`，而不是静默回首页
  - 401 触发后 pending recovery intent 能被登录成功链消费
- [ ] 5.3 页面行为测试至少覆盖：
  - `ProductDetail` 仅在商品可见时写入 recent context
  - `OrderDetail` 在接口失败或无效订单场景不覆盖已有有效上下文
  - `Cart` / `HomePage` / `OrderList(tabIndex=1)` 成功加载后可写入 recent context
  - 退出登录后 recent context / pending recovery intent 被清空
- [ ] 5.4 UAT / 验收至少覆盖：
  - 冷启动恢复商品详情
  - 冷启动恢复待支付订单列表（`tabIndex=1`）
  - 前后台切换恢复订单详情
  - 登录恢复后返回购物车、待支付订单列表或订单详情目标
  - 目标失效时 3 秒内进入 fallback 并给出明确反馈
- [ ] 5.5 为后续监控预留最小统计字段：`restoreAttempted`、`restoreSucceeded`、`fallbackUsed`、`restoreFailedReason`、`source`。本 Story 可先输出日志或本地调试埋点，不要求一次性接入完整监控平台。

#### 最低验收命令

```bash
flutter test test/**/recent_context* test/**/app_bootstrap* test/**/login_restore* 
flutter analyze
```

## Dev Notes

### 架构约束（必须遵守）

1. **保留启动初始化顺序：** `main()` 里 `SharedPreferencesUtil.init()` 必须先于任何 UI / 网络恢复逻辑执行。
2. **恢复对象只存语义，不存路由实现细节：** recent context / pending intent 必须描述 `targetType + targetId + fallback`，不能把 `MaterialPageRoute`、Widget 实例、页面标题或裸 route name 当成真相源。
3. **不能继续放大命名路由假设：** 当前 Flutter Mall 没有统一 `routes/onGenerateRoute`，而页面跳转主要靠 `MaterialPageRoute`；9.1 的恢复底座必须兼容这一现实。
4. **交易类目标恢复必须先服务端确认：** 商品详情、订单详情、支付恢复等高风险页面在恢复时都应先 revalidate，再决定是否展示原目标。
5. **只记录有效页，不记录失败页：** 请求失败、空白页、不可见商品、无效订单和临时弹层都不能覆盖最近一次有效上下文。
6. **保持共享层复用：** 网络仍走 `HttpUtil`，本地存储仍走 `SharedPreferencesUtil`，状态管理优先围绕现有 Provider 扩展，不得私建平行网络栈或存储栈。
7. **退出登录必须清理恢复状态：** `token`、pending recovery intent、recent context 不能分开清理，更不能在账号切换后继续恢复到旧账号目标。

### 关键数据源与复用点

| 主题 | 现有真相源 / 复用点 | 9.1 约束 |
| --- | --- | --- |
| 启动入口 | `flutter-mall/lib/main.dart` | 保持 `SharedPreferencesUtil.init()` 顺序，不破坏应用入口 |
| 启动页 | `flutter-mall/lib/welcome.dart` | 从固定延时跳转改为统一恢复壳层 |
| 主导航 | `flutter-mall/lib/layout/main_tab.dart`、`flutter-mall/lib/layout/bottom_navigation_bar.dart` | 首页 / 分类 / 购物车 / 我的 tab 是 fallback 基座 |
| 全局导航句柄 | `flutter-mall/lib/config/nav_key.dart` | 只能用于全局跳转，不应再被当成当前业务页 route 真相源 |
| Token 常量 | `flutter-mall/lib/config/constant_param.dart` | 恢复链、登录链与退出链必须围绕统一 token key 协同 |
| 本地缓存 | `flutter-mall/lib/utils/shared_preferences_util.dart` | 在现有工具上扩结构化 recent context helper |
| 网络与 401 | `flutter-mall/lib/utils/http_util.dart` | 401 恢复不能再依赖缺失的命名路由体系 |
| 登录恢复 | `flutter-mall/lib/view/mine/login/login.dart` | 登录成功优先恢复 pending intent / recent context |
| 退出登录 | `flutter-mall/lib/view/mine/setting/settings.dart` | 必须与恢复状态清理联动，避免跨账号恢复污染 |
| 首页目标 | `flutter-mall/lib/view/home/home_page.dart` | 成功加载后才能记录 `home` 上下文 |
| 购物车目标 | `flutter-mall/lib/view/cart/cart.dart` | 数据成功后才能记录 `cart` 上下文 |
| 订单列表目标 | `flutter-mall/lib/view/mine/order/order_list.dart` | `tabIndex=1` 是真实待支付恢复入口 |
| 商品详情目标 | `flutter-mall/lib/view/category/product/product_detail.dart` | 仅在 `visibility.visible=true` 时记录 |
| 订单详情目标 | `flutter-mall/lib/view/mine/order/order_detail.dart` | 仅在订单详情可用时记录 |
| 支付恢复模式参考 | `flutter-mall/lib/view/mine/order/order_pay.dart` | `resumed` 后重查状态的模式应推广到恢复链 |
| 当前消息跳转断层 | `flutter-mall/lib/view/mine/message/message.dart` | 已存在 `pushNamed('/product')` / `pushNamed('/activity')`，但缺全局路由注册 |

### 当前仓库信号与实现情报

1. `Welcome` 目前只做品牌启动页展示并在 1 秒后进入 `MainTab`，说明启动恢复壳层尚未落地，9.1 必须从这里切入，而不是在目标页补丁式恢复。
2. `HttpUtil` 的 401 拦截已具备“跳登录页”的最小能力，但其恢复目标依赖 `ModalRoute.settings.name`，与当前大量 `MaterialPageRoute` 实现不匹配；这是 9.1 必须优先修复的断层。
3. `OrderPay` 已在 `didChangeAppLifecycleState(AppLifecycleState.resumed)` 时重新查询支付状态，说明仓库已经接受“恢复后先 revalidate 服务端状态”的模式；9.1 应复用这一思路，而不是纯本地乐观恢复。
4. `ProductDetail` 已具备商品不可见 / 请求失败后的 fallback 语义字段（如 `reasonCode`、`fallbackAction`、`fallbackTarget`），说明恢复链可以复用现有目标失效表达，而不必另造第三套错误模型。
5. `Message` 页面当前对 `order` / `coupon` 使用直接页面跳转，对 `product` / `activity` 使用 `pushNamed`，进一步证明仓库导航体系尚未统一；9.1 必须先收口恢复底座，再让 9.3 在其上扩展消息意图唤回。
6. `OrderList` 页面已经具备 `全部 / 待支付 / 待发货 / 已完成 / 已取消` tab 和真实 API，说明“待支付订单”在当前仓库并非缺失入口；9.1 应把它纳入首批可恢复目标。
7. `Settings` 页面当前只 `remove(token)` 就返回上一页，说明退出登录与恢复状态清理尚未收口；9.1 必须补齐账号切换后的恢复状态隔离。

### 最新技术信息（2026-04-04 调研）

1. Flutter 官方文档当前仍将 deep linking 视为统一导航入口能力，适合把外部入口、恢复入口和 App 内导航入口收敛到同一套语义目标上；9.1 应先沉淀统一 recovery intent，而不是继续让启动恢复、登录恢复、消息跳转各自维护一套参数。
2. Flutter 官方关于 key-value 持久化的最佳实践仍适合保存轻量、可序列化的小对象；9.1 的 recent context 应保持“最小语义对象”而不是持久化大块页面状态快照。

### Project Structure Notes

| 层级 | 路径 | 责任 |
| --- | --- | --- |
| App Entry | `flutter-mall/lib/main.dart` | 初始化 SharedPreferences 与 Provider，挂载恢复壳层 |
| Bootstrap / Shell | `flutter-mall/lib/layout/app_bootstrap.dart`、`flutter-mall/lib/layout/intent_recovery_shell.dart` | 冷启动、热启动、登录恢复、fallback 决策 |
| Lifecycle State | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 生命周期监听与恢复状态流转 |
| Recent Context Model | `flutter-mall/lib/model/app_recent_context.dart` | 最近有效上下文模型、版本与校验 |
| Shared Storage | `flutter-mall/lib/utils/shared_preferences_util.dart` | recent context 序列化读写 |
| 401 / Auth Bridge | `flutter-mall/lib/utils/http_util.dart`、`flutter-mall/lib/view/mine/login/login.dart` | pending recovery intent 保存与恢复 |
| Home Target | `flutter-mall/lib/view/home/home_page.dart` | 首页成功加载后的有效上下文写入 |
| Cart Target | `flutter-mall/lib/view/cart/cart.dart` | 购物车成功加载后的有效上下文写入 |
| Product Target | `flutter-mall/lib/view/category/product/product_detail.dart` | 商品详情恢复、失效校验与 fallback |
| Order Target | `flutter-mall/lib/view/mine/order/order_detail.dart` | 订单详情恢复、状态重查与 fallback |

### References

- [Source: _opcos/planning-artifacts/1-new-feature/epic.md — Epic 9 / Story 9.1]
- [Source: _opcos/planning-artifacts/1-new-feature/prd.md — FR60 / Mobile App Experience / Performance & Stability Targets]
- [Source: _opcos/planning-artifacts/1-new-feature/architecture.md — Client Context & Recovery Contract / Flutter Mall / Recovery Patterns]
- [Source: _opcos/planning-artifacts/1-new-feature/ux-design.md — Intent Recovery Loop / 消息召回、弱网与上下文恢复流 / Commerce State Shell]
- [Source: _opcos/project-context.md — 移动端技术栈、SharedPreferences 初始化顺序、共享层复用规则]
- [Source: flutter-mall/lib/main.dart]
- [Source: flutter-mall/lib/welcome.dart]
- [Source: flutter-mall/lib/layout/main_tab.dart]
- [Source: flutter-mall/lib/layout/bottom_navigation_bar.dart]
- [Source: flutter-mall/lib/config/nav_key.dart]
- [Source: flutter-mall/lib/config/constant_param.dart]
- [Source: flutter-mall/lib/utils/shared_preferences_util.dart]
- [Source: flutter-mall/lib/utils/http_util.dart]
- [Source: flutter-mall/lib/view/mine/login/login.dart]
- [Source: flutter-mall/lib/view/mine/setting/settings.dart]
- [Source: flutter-mall/lib/view/home/home_page.dart]
- [Source: flutter-mall/lib/view/cart/cart.dart]
- [Source: flutter-mall/lib/view/mine/order/order_list.dart]
- [Source: flutter-mall/lib/view/category/product/product_detail.dart]
- [Source: flutter-mall/lib/view/mine/order/order_detail.dart]
- [Source: flutter-mall/lib/view/mine/order/order_pay.dart]
- [Source: flutter-mall/lib/view/mine/message/message.dart]
- [Source: https://docs.flutter.dev/ui/navigation/deep-linking]
- [Source: https://docs.flutter.dev/cookbook/persistence/key-value]

## Dev Agent Record

### Agent Model Used

Cascade

### Debug Log References

- BMAD create-story workflow
- Sprint status auto-discovery (`9-1-app-启动恢复与最近有效上下文回流`)
- Flutter startup / lifecycle / navigation / shared preferences code scan

### Completion Notes List

- 已按当前 backlog 顺序生成 Story 9.1
- 已将恢复目标约束为当前 Flutter 仓库中已存在且可验证的页面，避免对不存在路由做空想实现
- 已明确 401 / 登录恢复与 `MaterialPageRoute` / `pushNamed` 混用之间的断层，要求先落统一语义恢复对象
- 已把冷启动、热启动、前后台切换与登录恢复统一收口到恢复壳层，而不是分散到各页面

### File List

- _opcos/implementation-artifacts/9-1-app-启动恢复与最近有效上下文回流.md
