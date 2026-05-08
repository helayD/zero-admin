---
stepsCompleted:
  - 1
  - 2
  - 3
  - 4
  - 5
  - 6
  - 7
  - 8
inputDocuments:
  - _opcos/planning-artifacts/1-new-feature/prd.md
  - _opcos/planning-artifacts/1-new-feature/ux-design.md
  - _opcos/planning-artifacts/1-new-feature/epic.md
  - _opcos/planning-artifacts/implementation-readiness-report-2026-03-20.md
  - _opcos/project-context.md
  - docs/index.md
  - docs/project-overview.md
  - docs/backend/README.md
  - docs/frontend/README.md
  - docs/mobile/README.md
  - docs/api/README.md
workflowType: 'architecture'
project_name: 'ai_flutter_client'
user_name: 'David'
date: '2026-03-20T17:38:06+0800'
prd_dir: '1-new-feature'
lastStep: 8
status: 'complete'
completedAt: '2026-03-20T17:38:06+0800'
---

# Architecture Decision Document

_本架构文档基于 PRD、UX 设计、项目上下文与现有代码结构整理，目标是为多 AI 代理协作实现提供单一技术事实来源，降低 brownfield 多端电商项目在实现阶段的分歧与返工。_

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

当前 PRD 共定义 71 条功能需求，可归纳为 10 个架构能力簇：

- **商品发现与内容导购**：首页、分类、品牌、推荐、专题、搜索入口需要共同消费商品、内容、营销与搜索读模型，不能继续依赖前台静态占位或后台孤立配置。
- **会员账户与用户资产**：会员资料、地址、积分、优惠券、收藏与关注既是用户中心能力，也是订单、营销与留存能力的输入，因此必须由统一会员域提供可复用接口与作用域校验。
- **购物车与结算**：购物车、确认订单、金额拆分、库存校验与优惠计算属于同一交易前链路，必须由 OMS 统一聚合，避免前台页面自行计算导致口径漂移。
- **订单交易与售后服务**：订单创建、支付发起、订单恢复、取消、确认收货、退货申请、售后处理需要稳定状态机、幂等、补偿与后台统一视图。
- **营销活动与内容运营**：优惠券、秒杀、广告、推荐位、专题与优选专区既服务前台转化，也服务商户和平台运营，因此 SMS/CMS 需要通过清晰的发布与生效边界接入前台。
- **搜索、同步与平台协同**：商品 ES 同步、推荐内容生效、跨端状态一致与消息/补偿链路需要异步事件与可观测性，不应由人工刷库或隐式脚本兜底。
- **后台管理与权限治理**：平台后台、租户后台、商户后台都依赖统一 RBAC 与数据范围控制，且必须支持审计留痕、批量操作与高密度工作流。
- **平台租户与多商户治理 / 增长扩展**：平台-租户-商户三级作用域、商户治理、复购运营、评价沉淀与留存分析要求系统从“单商家商城项目”升级为“平台化经营底座”。
- **移动端召回、上下文恢复与渠道合规**：冷启动、热启动、登录态恢复、消息/活动唤回、弱网重试、场景化权限请求与升级闸门已经进入正式 FR 范围，不能再视为 Flutter 端体验优化项。
- **提货卡抽赏与链上资产**：抽卡活动、卡池、卡片模板、唯一编号、用户资产台账、蚂蚁链 token 发放、链上回执、合规审核与失败补偿必须被视为同一条可信资产链路，而不是页面营销玩法的附属功能。

**Non-Functional Requirements:**

- 核心交易接口目标可用性 `99.9%`，商品详情、购物车、确认订单、创建订单、支付、订单查询链路 `P95` 需控制在 `500ms-1s`。
- 商品变更同步到 Elasticsearch 的延迟应控制在 `5 分钟` 以内，且需要可监控、可重试、可人工介入。
- 多租户、多商户隔离必须成为一等约束，平台级、租户级、商户级菜单与数据查询不得串读或误写。
- 支付、库存、优惠券、积分、订单状态流转必须具备幂等与补偿机制，异常恢复不是实现细节，而是产品要求。
- 移动端 App 需要满足冷启动进入首页或最近有效上下文 `P90 <= 3s`、上下文恢复成功率 `>= 99%`、消息/活动唤回成功链路 `>= 95%` 的正式约束。
- 设备权限请求、弱网反馈与版本升级提示必须服从“不中断当前任务、完成后返回原目标”的体验契约，不能由页面各自实现。
- 提货卡中奖后需要先落本地资产记录，再异步发放蚂蚁链 token；同一抽卡请求、同一卡片实例和同一链上发放任务必须 100% 幂等。
- 提货卡活动需要满足实名、版权、内容安全与反金融化约束，默认不支持连续挂牌、集中竞价、收益承诺和虚拟币计价。
- UX 明确要求统一状态语义、统一作用域表达、统一异常恢复模式，因此架构必须支持“可信确认流”和“作用域优先”体验。

**Scale & Complexity:**

- Primary domain: 平台化多端电商系统（Web 管理后台 + Flutter 商城端 + API/RPC 微服务）
- Complexity level: 高
- Estimated architectural components: 14 个核心能力单元（admin-api、front-api、sys、ums、pms、oms、sms、cms、search、job、consumer、web-admin、flutter-mall、shared pkg）

### Technical Constraints & Dependencies

- 当前仓库是 **brownfield** 项目，后端已采用 go-zero 微服务分层，前端已落在 React 17 + Umi 3 + Ant Design Pro 5.2 / Ant Design 4，移动端为 Flutter + Provider + Dio。
- 生成式工件是第一约束：`.api`、`.proto`、`gorm/gen` 产物必须通过源定义和生成流程维护，不能直接手改生成文件。
- Go 网关遵循 `handler -> logic -> svc`；网关负责协议转换与错误封装，业务逻辑必须留在 `logic` 与下游服务。
- 管理端与前台网关的错误包装与中间件链不同，不能假定二者对称。
- 当前项目已有 Dockerfile、`make build/start/stop`、K8s manifest 和 `service_manager.sh`，说明部署结构已具备多服务独立交付能力，但 CI/CD 还未在仓库中标准化。
- 多租户、多商户模型已在 PRD 中被正式确立，但代码层仍缺少统一的作用域上下文包、作用域中间件和跨域查询/写入的一致约束，这是 MVP 实现阶段的第一优先缺口。
- 此前旧版 `architecture.md` 与 `epic.md` 曾停留在 `59 FR` 基线；本轮已把 `FR60-FR64` 作为正式架构输入并同步收口到最新 Epic/Story 链。

### Cross-Cutting Concerns Identified

- 认证、RBAC、数据范围与平台/租户/商户三级作用域
- 交易一致性、库存锁定、支付回调、优惠券回退、积分补偿
- 商品与营销内容的异步发布、生效与搜索同步
- 提货卡实例、链上 token、用户资产展示与合规审核的一致性
- 移动端 App 生命周期、目标意图恢复、弱网重试、场景化权限与升级闸门
- 多端状态语义一致性（前台订单状态、后台工单状态、补偿状态、审核状态）
- 可观测性、审计日志、异常定位与人工干预入口
- 生成代码约束、共享基础设施复用与 AI 代理实现一致性

## Starter Template Evaluation

### Primary Technology Domain

本项目不是 greenfield scaffold 选型题，而是 **brownfield 全栈电商平台基线选型题**。真实技术域由以下三部分共同组成：

- Go 1.25 + go-zero 1.9.3 的 HTTP/RPC 微服务底座
- React 17 + Umi 3.5 + Ant Design Pro 5.2 + Ant Design 4.17 的管理后台
- Flutter（Dart `^3.7.0`）+ Provider + Dio 的移动商城端

### Starter Options Considered

1. **现有仓库基线（Selected）**
   - 优点：保留现有服务边界、生成流程、部署脚本、前后台页面资产与 Flutter 页面资产。
   - 风险：需要在旧版本栈上做增量演进，而不是借一次全量重脚手架完成现代化升级。

2. **Umi / Ant Design Pro 新版脚手架（Not Selected）**
   - 官方生态已经走到 Umi 4.x、Ant Design 6.x、React 18/19 生态。
   - Ant Design 官方文档当前展示版本为 `6.3.3`，且明确说明 `antd 6.0` 后不再支持 React 16/17。
   - 对当前 React 17 / Ant Design 4 的管理端来说，直接重脚手架会把“业务改造”与“跨大版本迁移”强耦合，不适合作为本轮 MVP 基础。

3. **Flutter 官方新项目脚手架（Not Selected）**
   - Flutter 官方当前文档反映的是 `3.41.2` 时代的稳定通道文档。
   - 对已存在的 `flutter-mall` 而言，重新 `flutter create` 只能带来目录重置与工程升级噪音，不能替代业务闭环改造。

4. **全新 go-zero 服务脚手架（Not Selected）**
   - Go 官方当前稳定版已到 `1.26.1`，go-zero 官方发布说明已到 `v1.10.0`。
   - 但当前仓库已以 go-zero `1.9.3` 与既有 `.api/.proto` 生成链为基础运行，直接用新脚手架重建会破坏现有服务边界和生成关系。

### Selected Starter: Existing Repository Baseline

**Rationale for Selection:**

本轮架构选择明确采用 **“保留现有 brownfield 基线，不进行重脚手架初始化”** 的策略。原因如下：

- PRD 的核心目标是把现有能力收敛成可运营、可扩展、可验证的产品基线，而不是换技术演示。
- 当前仓库已经具备服务拆分、网关分层、Docker / K8s 交付与多端 UI 资产；问题在于作用域治理、状态一致性、交易补偿和体验闭环，而不是缺脚手架。
- 上游生态版本已经明显领先于当前仓库，若把 React/Umi/Antd/Flutter/Go 升级与业务需求改造合并，会显著增加实施风险。

**Initialization Command:**

```bash
# 本轮不执行新脚手架初始化，保留既有仓库作为实现基线
make build
cd web-admin && npm install
cd flutter-mall && flutter pub get
```

**Architectural Decisions Provided by Baseline:**

**Language & Runtime:**

- Backend: Go `1.25`, go-zero `1.9.3`, gRPC `1.77.0`, GORM `1.31.1`
- Web Admin: React `17.0.0`, Umi `3.5.0`, Ant Design Pro `5.2.0`, Ant Design `4.17.0`, TypeScript `4.5.0`
- Mobile: Dart SDK `^3.7.0`, Provider, Dio, cached_network_image, easy_refresh

**Styling Solution:**

- Web 采用 Ant Design / Less 主题体系与 Ant Design Pro 页面模板
- Mobile 采用 Flutter 组件与现有 `widgets/` 自定义组件扩展

**Build Tooling:**

- 后端通过 `make build`, `make gen`, `make model` 与 `service_manager.sh` 驱动
- Web 通过 `umi build`, `npm run lint`, `npm run tsc`, `npm run test`
- Mobile 通过 `flutter analyze`, `flutter test`, Flutter 平台构建链驱动

**Testing Framework:**

- Go：按服务定向 build / test
- Web：Jest + Playwright 现有测试基础
- Mobile：Flutter test / analyze

**Code Organization:**

- 网关分 admin-api / front-api
- 业务服务按 `sys/ums/pms/oms/sms/cms/search`
- 异步处理分 `consumer`、定时任务分 `job`
- 共享基础设施下沉到 `pkg/`

**Development Experience:**

- `.api` / `.proto` / `gorm/gen` 已形成标准化生成流程
- Web 端已具备 proxy + `request` 中心化调用模式
- Flutter 已有统一 `service_url.dart` 与 `HttpUtil`

**Version Verification Note (2026-03-20):**

- Go 官方稳定版：`1.26.1`
- React 官方博客当前最新主版本文章：`19.2`（并已发布 `19.2.1` 修复）
- Ant Design 官方文档当前版本：`6.3.3`
- Flutter 官方文档当前反映稳定通道版本：`3.41.2`
- go-zero 官方当前最新 release：`v1.10.0`

结论：**本轮架构实施不做全栈大版本升级，升级工作单列为后续 Epic。**

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**

- 统一平台 / 租户 / 商户作用域模型与上下文传递
- 明确 MySQL 为交易与治理主数据源，Redis / ES / MQ 为辅
- 固化 “REST 网关 + gRPC 服务 + MQ 异步事件 + Job 补偿” 通信拓扑
- 统一订单、库存、支付、优惠券、积分的幂等与补偿框架
- 固化 AI 代理必须遵守的目录边界、命名、格式和生成代码约束

**Important Decisions (Shape Architecture):**

- 统一 JWT claim 与鉴权 / 授权 / 数据范围校验模式
- 统一前后台状态语义与共享业务组件实现边界
- 统一缓存命名空间、事件命名规范、监控字段与日志上下文
- 明确 CI/CD、Docker / K8s、监控告警与灰度发布演进方向

**Deferred Decisions (Post-MVP):**

- React / Umi / Ant Design 大版本升级
- Flutter 目录重构或完整 design token 共享包
- GraphQL、微前端、事件溯源、开放平台插件化、消息中心独立服务

### Data Architecture

**Primary Transaction Store: MySQL**

- `sys/ums/pms/oms/sms/cms` 的核心业务数据以 MySQL 为唯一写入真相源。
- 所有交易、治理、会员、营销与内容主状态必须先落 MySQL，再通过事件同步到其他系统。
- 新增平台化能力时，实体必须带清晰作用域字段：`platform_id`、`tenant_id`、`merchant_id`（按场景取其一或多者组合），并保留操作人、审计与状态机字段。
- 提货卡抽赏采用“链下主台账 + 链上确权凭证”模型：抽卡结果、卡片实例、唯一编号、用户归属、发放任务与回执先落 MySQL，蚂蚁链 token 作为链上确权结果回写，不把链上状态当作唯一真相源。

**Cache & Coordination: Redis**

- Redis 用于登录态、验证码、热点缓存、幂等键、分布式锁、短期聚合数据。
- Redis 不是订单、库存、优惠券状态的真相源；关键状态必须由 MySQL 决定。
- 缓存 key 统一使用带作用域前缀的命名方式，例如：
  - `za:{env}:{scope}:{domain}:{entity}:{id}`
  - 示例：`za:prod:tenant-12:pms:product:20019`

**Search Read Model: Elasticsearch**

- ES 只承担搜索和推荐读模型，不承担交易与配置真相。
- 商品上下架、详情修改、品牌/分类/推荐关系变化通过事件异步同步到 ES。
- 同步链路必须具备 lag 指标、失败重试与人工补偿入口。

**Document / Flexible Data: MongoDB**

- MongoDB 仅在现有文档型、报表型、内容扩展型场景下继续使用。
- 不得将订单、支付、库存、退款等交易核心状态迁入 MongoDB。

**Data Modeling Strategy**

- 保持现有域前缀表命名：`sys_*`, `ums_*`, `pms_*`, `oms_*`, `sms_*`, `cms_*`
- 新增平台化实体延续现有命名风格，不引入第二套命名体系。
- 对涉及状态流转的数据，显式记录 `status`, `version`, `updated_at`, `updated_by`，为幂等与并发更新提供基础。
- 金额以最小货币单位存储，前端负责格式化显示，避免浮点误差。

**Digital Card Asset Modeling**

- 继续沿用 `sms_*` 域前缀，避免为单一新增能力临时拆出第二套服务命名。
- 推荐新增的核心实体包括：
  - `sms_draw_activity`：抽卡活动主表，承接活动时间窗、规则、资格、实名要求与合规状态
  - `sms_draw_pool`：卡池定义，承接稀有度、发售数量、概率披露与展示配置
  - `sms_card_template`：卡片模板，承接卡面、版权来源、描述、上下架与内容审核状态
  - `sms_card_instance`：中奖后生成的真实卡片资产实例，承接唯一编号、用户归属、实例状态
  - `sms_card_mint_task`：链上发放任务，承接幂等键、任务状态、失败原因、重试次数与回执摘要
  - `sms_card_asset_log`：面向审计的资产状态日志，承接实例状态变化、合规动作、冻结/回收记录
- `sms_card_instance` 与链上 token 必须保持一对一映射；允许“实例已创建、token 待发放”的处理中状态，但不允许多个 token 对应同一实例或同一 token 绑定多个实例。

**Migration & Schema Evolution**

- 数据变更通过 SQL 脚本与 `gorm/gen` 源头同步演进。
- 禁止手改 `gen/model`、`query` 或 `pb` 生成文件；必须修改 `.api`、`.proto` 或生成脚本后再生成。
- 每个跨域数据变更要附带兼容策略，避免一次发布同时破坏 web-admin、front-api 与 flutter-mall。

### Authentication & Security

**Authentication Model**

- 管理端与前台统一采用 JWT，但 claim 必须扩展为平台化版本。
- 最低要求的 claim 集合：
  - `userId`
  - `roleIds`
  - `subjectType`（platform / tenant / merchant / member）
  - `tenantId`
  - `merchantId`
  - `dataScope`
  - `traceId`

**Authorization Patterns**

- `sys` 域继续负责菜单、角色、数据范围元数据。
- 网关负责解析身份与上下文，`logic` 层负责资源级授权与数据范围校验。
- 禁止页面仅靠前端按钮隐藏来完成授权；所有写操作与关键查询都需要服务端二次校验。

**Security Middleware**

- `admin-api`、`front-api` 增加统一的：
  - 鉴权中间件
  - 作用域解析中间件
  - 审计上下文中间件
  - 幂等键中间件（订单、支付、租户开通、商户审核等关键写操作）
  - 限流中间件（会员侧交易入口与后台高风险写接口）

**Sensitive Data Handling**

- 手机号、地址、支付相关标识、商户入驻敏感材料必须在日志与 UI 中按角色做脱敏。
- 第三方支付密钥、短信、OSS、ES、MQ 凭据只允许放在配置中心 / 环境配置，不进入 AI 规则文档或源码常量。
- 管理端默认最小可见原则：普通运营角色不能看见不必要的敏感字段。

### API & Communication Patterns

**North-South API Pattern**

- 对 Web Admin 和 Flutter Mall 暴露的接口统一保持 REST 风格，由 `admin-api` 与 `front-api` 提供。
- OpenAPI 静态文档从 `.api` 生成，不允许维护脱离源定义的手工接口文档。

**East-West Service Pattern**

- 网关与领域服务之间通过 gRPC 通信。
- 领域服务拥有自己的数据模型和写边界，禁止网关直接访问跨服务数据库。
- 跨领域查询优先通过网关编排；跨领域状态传播优先通过事件。

**Async Event Pattern**

- RabbitMQ 承担异步发布、搜索同步、营销生效、补偿触发等非同步主链路行为。
- 事件命名统一采用：`{domain}.{entity}.{action}.v{n}`
  - 示例：`pms.product.published.v1`
  - 示例：`oms.order.closed.v1`
  - 示例：`sms.coupon.reverted.v1`
  - 示例：`sms.digital_card.draw_resulted.v1`
  - 示例：`sms.digital_card.mint_requested.v1`
  - 示例：`sms.digital_card.minted.v1`
- 事件 payload 统一包含：
  - `eventId`
  - `occurredAt`
  - `traceId`
  - `platformId/tenantId/merchantId`
  - `actorId`
  - `entityId`
  - `action`
  - `version`
  - `data`

**Client Context & Recovery Contract**

- `front-api` 需要识别并透传最小客户端上下文字段，用于支持 `FR60-FR64`：
  - `appVersion`
  - `platform`
  - `deviceId`
  - `intentSource`
  - `intentId`
  - `networkState`
- 消息、活动、优惠券、待支付订单等唤回入口统一使用意图契约，不允许每个页面私定义跳转参数。推荐最小字段：
  - `intentType`
  - `targetType`
  - `targetId`
  - `fallbackType`
  - `requiresAuth`
  - `minAppVersion`
  - `issuedAt`
- 当前端目标已失效、版本不兼容或权限未完成时，`front-api` 需要返回结构化恢复提示，而不是让 App 自行猜测降级路径。

**Error Handling Standards**

- Admin API 保持现有自定义错误包装与上下文日志策略。
- Front API 保持面向用户的轻量错误输出，但交易类错误必须返回可恢复语义，而不是笼统失败。
- UI 层禁止直接展示 RPC / SQL 原始错误文本。

### Frontend Architecture

**Web Admin**

- 保持 Umi 3 + React 17 + Ant Design 4 / Ant Design Pro 5.2 基线，不在本轮引入 Umi 4 / Antd 6 混合迁移。
- 继续采用 `index.tsx + service.ts + data.d.ts + components/*` 模块结构。
- 所有网络请求必须走 Umi `request` 与现有拦截器，不允许页面手工拼接 token 或硬编码后端 host。
- 列表 / 详情 / 批量操作页面继续以 `PageContainer + ProTable + Drawer/Modal` 为主模式。

**Flutter Mall**

- 保持 Provider + Dio + `HttpUtil` + `service_url.dart` 的集中式网络模式。
- Widgets 不直接创建裸 `Dio` 或私有 token 流；网络与鉴权继续集中在共享层处理。
- 高风险交易节点采用“状态确认优先”的交互：确认订单、支付发起、订单恢复、售后申请必须以服务端状态为准。
- 新增统一 **App Lifecycle Shell**，负责冷启动、热启动、登录态恢复、前后台切换与目标意图恢复，避免在 `home/cart/mine/order` 页面各自维护会话恢复逻辑。
- 新增统一 **Commerce State Shell**，覆盖首页、分类、搜索、商品详情、购物车、确认订单、订单列表/详情和售后页面的加载态、空态、错误态与弱网态。
- 新增统一 **Permission Request Sheet** 与 **Upgrade Gate**；具体权限编排由内部 **Permission Broker** 承接，只在真实业务触发时请求通知 / 相册 / 相机权限，并在强制或限时升级后恢复原始目标上下文。
- 提货卡资产页、抽卡活动页和“我的卡片”页统一走服务端确认优先模式；中奖结果、唯一编号、token 状态与合规限制文案由服务端聚合返回，客户端不自行拼装“已到账”语义。

**Cross-End UX Architecture**

- 统一以下共享业务组件与语义，而不是在每个端各自命名：
  - `Scope Context Bar`
  - `Order Timeline Panel`
  - `Price Breakdown Card`
  - `Promotion Stack / Coupon Sheet`
  - `Sync Status Badge`
  - `Commerce State Shell`
  - `Intent Recovery Loop`
  - `Permission Request Sheet`
  - `Upgrade Gate`
- 同一业务对象在前后台必须使用统一状态命名、金额口径与时间语义。

### Infrastructure & Deployment

**Build & Packaging**

- Go 服务继续通过 `make build` 输出到 `target/<service>/`。
- `target` 是发布工件目录，不是源码目录；所有 AI 代理应把变更落在源目录，再由构建流程产出目标文件。
- Web 管理端与 Flutter 客户端分别独立构建，不纳入后端二进制打包。

**Deployment Topology**

- 每个网关与 RPC 服务保持独立 Docker 镜像与 K8s 清单。
- `consumer` 与 `job` 保持脱离请求主链路的独立运行单元，防止异步补偿拖慢用户路径。
- 后续如引入灰度或按租户隔离部署，应在网关层与配置层实现，而不是复制业务逻辑。

**Observability**

- 最小可观测字段：
  - `traceId`
  - `requestId`
  - `platformId`
  - `tenantId`
  - `merchantId`
  - `userId`
  - `service`
  - `handler`
  - `action`
  - `result`
- 至少建立以下看板 / 告警：
  - 搜索同步 lag
  - 订单超时关闭任务执行情况
  - 支付回调失败 / 重试
  - 优惠券回滚失败
  - 租户 / 商户审核与启停审计事件
  - App 冷启动 / 会话恢复成功率
  - 消息 / 活动 / 优惠券唤回到达率
- 弱网重试成功率与失败摘要
- 强制升级拦截次数与版本分布
- 权限拒绝后的替代路径使用率
- 提货卡抽卡成功率、链上发放成功率、处理中超时数量、补偿重试次数与链上回执对账差异

**Scaling Strategy**

- 水平扩展优先级：`front-api` / `admin-api` / `oms` / `pms` / `search` / `consumer`
- Job 保持单实例或分布式锁控制的调度模式，避免重复补偿
- 高读场景优先通过 Redis 与 ES 扩展，关键写链路优先通过幂等和分区扩展保护

### Decision Impact Analysis

**Implementation Sequence:**

1. 建立统一作用域模型（platform / tenant / merchant / member）
2. 扩展 JWT claim、网关中间件与审计上下文
3. 为 `oms/pms/sms/cms/ums` 补充作用域过滤与资源校验
4. 统一事件命名、payload 结构与消费幂等
5. 在 `sms` 域补充抽卡活动、卡池、卡片模板、资产实例与链上发放任务的数据模型
6. 建立蚂蚁链开放联盟链统一集成层、幂等发放键与回执对账逻辑
7. 为 Flutter Mall 建立意图恢复壳层、状态壳层、权限闸门与升级闸门
8. 完成商品同步、订单补偿、支付回调、优惠回滚和提货卡发放补偿等主链路改造
9. 在 web-admin / flutter-mall 中落地统一状态与作用域组件
10. 建立 CI / 监控 / 验证基线

**Cross-Component Dependencies:**

- 作用域模型会影响网关、RBAC、查询过滤、缓存 key、事件 payload 与 UI 上下文条
- 订单状态机会影响 OMS、SMS、库存、支付、Job 与 Consumer
- 搜索同步规则会影响 PMS、CMS、SMS、Search 与前台导购体验
- 统一状态语义会影响 PRD 验收、UX 验收、自动化测试与客服后台定位流程

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:**

至少有 10 类实现冲突点会导致多 AI 代理产出不兼容代码：

- 作用域字段命名与传递方式
- JSON 字段风格
- 事件命名与 payload 结构
- Web 页面模块组织方式
- Flutter 网络调用位置
- Go 业务逻辑落点
- 生成代码维护方式
- 错误封装与日志字段
- 缓存 key 命名
- 搜索 / 补偿链路实现方式

### Naming Patterns

**Database Naming Conventions:**

- 延续现有域前缀命名：`sys_*`, `ums_*`, `pms_*`, `oms_*`, `sms_*`, `cms_*`
- 表名与字段名统一 `snake_case`
- 新增作用域字段统一使用：
  - `platform_id`
  - `tenant_id`
  - `merchant_id`
- 审计字段统一使用：
  - `created_by`
  - `updated_by`
  - `created_at`
  - `updated_at`
- 索引命名统一：`idx_<table>_<field>` 或 `uk_<table>_<field>`

**API Naming Conventions:**

- 保持现有网关路由命名空间，不在已有模块上引入第二套 REST 风格。
- 对外 JSON 字段统一 `camelCase`，与现有 `pageNum/pageSize`、前端消费习惯保持一致。
- 网关层允许把下游 proto 的 `snake_case` 映射成前端使用的 `camelCase`，但禁止把 proto 命名直接暴露给前端。
- 分页参数统一：`pageNum`, `pageSize`

**Code Naming Conventions:**

- Go package：小写
- Go 逻辑目录：按业务域分组（如 `order_main`, `coupon_scope`），不要再新建平行命名体系
- Web 页面目录：沿用现有目录大小写与页面模式，不混用 kebab / Pascal / snake 三套目录名
- React 组件：组件名 `PascalCase`，页面入口文件统一 `index.tsx`
- Flutter 文件：`snake_case`，类名 `PascalCase`

### Structure Patterns

**Project Organization:**

- Go 网关与服务统一遵循：
  - `handler`：参数解析、调用逻辑、返回响应
  - `logic`：业务编排、RPC 聚合、错误转换
  - `svc`：服务上下文、依赖注入
- 共享基础能力落 `pkg/`，不要散落到每个服务各自实现。
- `consumer` 只做异步消费与副作用处理；`job` 只做定时扫描、补偿与维护任务。
- Web 页面必须把请求代码放在 `service.ts` 或共享 `src/services/` 中，不能把 HTTP 请求写进组件 JSX。

**File Structure Patterns:**

- 生成源优先：
  - API：`api/*/doc/api/**/*.api`
  - RPC：`rpc/*/proto/*.proto`
  - GORM：`rpc/*/gen/generator.go`
- 生成结果作为构建产物和编译输入，不作为手工编辑入口。
- 配置文件统一放在各服务 `etc/` 下，构建后复制到 `target/<service>/`

### Format Patterns

**API Response Formats:**

- 对外 REST 统一采用现有响应包裹：`{ code, msg, data }`
- 业务成功与失败必须使用结构化错误码 / 语义错误，不返回裸字符串
- 交易和治理类失败响应必须可映射为：
  - 用户可恢复
  - 权限拒绝
  - 资源状态冲突
  - 系统错误

**Data Exchange Formats:**

- 时间字段：
  - 新增对外接口优先使用 RFC3339 / ISO 8601 字符串
  - 现有接口若使用旧格式，保持兼容，不在同一接口中混用两套时间语义
- 金额字段：
  - 后端存储与计算使用整数最小货币单位
  - UI 只负责格式化展示
- 布尔与枚举：
  - 对外保持语义明确，不要用 `0/1/2/3` 直接裸露给前端而不给注释或转换

### Communication Patterns

**Event System Patterns:**

- 命名：`{domain}.{entity}.{action}.v{n}`
- 主题 / 队列不以页面名、操作按钮名命名，而以业务事实命名
- 事件消费者必须做到：
  - 幂等消费
  - 重试可控
  - 死信可追踪
  - 日志可关联 `traceId`
- 推送 / 站内消息 / 活动唤回使用统一意图对象，不使用“裸 URL + 页面局部 query”传递业务语义。

**State Management Patterns:**

- Web Admin 以页面局部状态和服务端状态为主，不引入新的全局 Redux 状态树
- Flutter 继续使用 Provider，状态由业务 Provider 和共享工具层维护
- 对交易类动作默认 **后端确认优先**，只在低风险场景允许乐观更新（如简单开关 / 排序）
- App 生命周期状态与页面业务状态分离：`intent / lifecycle / network / upgrade` 不与商品、购物车、订单 Provider 混写。

### Process Patterns

**Error Handling Patterns:**

- Handler 不写业务判断；业务错误在 `logic` 中返回并做网关层转换
- 日志记录与用户提示分离：
  - 日志保留上下文与内部原因
  - UI 只展示可理解、可恢复的文案
- 所有高风险失败必须至少返回一个明确恢复动作：
  - 重新支付
  - 重新查询状态
  - 联系客服 / 打开工单
  - 重新同步 / 人工补偿

**Loading State Patterns:**

- Web Admin：
  - 列表页使用表格 loading / skeleton
  - 写操作期间按钮禁用并保留上下文
- Flutter：
  - 交易关键流使用页面级 loading + 结果态
  - 不能用短暂 toast 替代关键状态更新
  - 首页、分类、搜索、商品详情、购物车、确认订单、订单列表/详情、售后页面统一走 `Commerce State Shell`
- 异步生效类操作必须展示同步状态，而不是“保存成功”后无后续反馈

**Recovery / Permission / Upgrade Patterns:**

- 登录恢复、消息唤回、版本升级、权限授权与弱网重试必须串成一条 `Intent Recovery Loop`，完成系统门槛后优先回到原目标页。
- 权限请求只允许发生在真实触发点，例如消息订阅、评价上传、售后凭证上传；拒绝后必须给替代路径，不阻断当前主任务。
- 升级提示必须明确：
  - 受影响功能
  - 最迟生效时间
  - 是否阻断使用
  - 升级入口
- 目标资源失效时必须提供替代落点或明确失败反馈，不允许静默回首页。

### Enforcement Guidelines

**All AI Agents MUST:**

- 遵守生成源优先原则，不手改 `pb.go`、`gen/query`、`types.go` 等生成文件
- 遵守作用域字段、事件命名、REST JSON 命名和共享日志字段规范
- 在各自子系统内部遵守本地约定，不为了“统一感”强行改写别的子系统命名风格

**Pattern Enforcement:**

- 代码审查时优先检查：
  - 作用域是否穿透到查询 / 写入
  - 是否绕过共享 HTTP / request / HttpUtil
  - 是否把逻辑塞进 handler / widget / page component
  - 是否直接修改生成代码
- 文档中新增模式前，先确认现有 `project-context.md` 是否已有约束，避免双重标准

### Pattern Examples

**Good Examples:**

- `oms.order.closed.v1` 事件携带 `tenantId`, `merchantId`, `orderId`, `traceId`
- Web 新页面采用 `index.tsx + service.ts + data.d.ts`
- 新增 Go 依赖统一放进 `internal/svc/service_context.go`
- 消息唤回对象统一携带 `intentType`, `targetType`, `targetId`, `requiresAuth`, `minAppVersion`

**Anti-Patterns:**

- 在 React 组件中直接写 `fetch('http://127.0.0.1:8000/...')`
- 在 Flutter widget 中 new 一个裸 `Dio`
- 在网关 handler 内直接写跨服务业务判断
- 为图省事直接手改 `.pb.go` 或 `.gen.go`
- 用页面按钮名命名 MQ 事件，如 `clickSaveCoupon`
- 消息唤回后丢失目标意图，直接把用户打回首页

## Project Structure & Boundaries

### Complete Project Directory Structure

```text
zero-admin/
├── README.md
├── DEPLOY.md
├── go.mod
├── makefile
├── service_manager.sh
├── api/
│   ├── admin/
│   │   ├── doc/api/
│   │   ├── etc/
│   │   ├── internal/
│   │   │   ├── common/
│   │   │   │   ├── errorx/
│   │   │   │   └── res/
│   │   │   ├── config/
│   │   │   ├── handler/
│   │   │   │   ├── cms/
│   │   │   │   ├── oms/
│   │   │   │   ├── pms/
│   │   │   │   ├── sms/
│   │   │   │   ├── sys/
│   │   │   │   └── ums/
│   │   │   ├── logic/
│   │   │   │   ├── cms/
│   │   │   │   ├── oms/
│   │   │   │   ├── pms/
│   │   │   │   ├── sms/
│   │   │   │   ├── sys/
│   │   │   │   └── ums/
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go                # 现有/延续
│   │   │   │   ├── scope.go               # 规划新增：平台/租户/商户作用域解析
│   │   │   │   ├── audit.go               # 规划新增：审计上下文
│   │   │   │   └── idempotency.go         # 规划新增：关键写接口幂等
│   │   │   ├── svc/
│   │   │   └── types/
│   │   ├── static/
│   │   └── Dockerfile
│   └── front/
│       ├── doc/api/
│       ├── etc/
│       ├── internal/
│       │   ├── config/
│       │   ├── handler/
│       │   │   ├── home/
│       │   │   ├── member/
│       │   │   ├── order/
│       │   │   └── product/
│       │   ├── logic/
│       │   │   ├── common/
│       │   │   ├── home/
│       │   │   ├── member/
│       │   │   ├── order/
│       │   │   └── product/
│       │   ├── middleware/
│       │   │   ├── auth.go                # 规划新增：会员 / 商户上下文校验
│       │   │   ├── scope.go               # 规划新增：租户/商户入口透传
│       │   │   ├── client_meta.go         # 规划新增：App 版本/渠道/意图头解析
│       │   │   └── idempotency.go         # 规划新增：下单/支付防重
│       │   ├── svc/
│       │   └── types/
│       ├── static/
│       └── Dockerfile
├── rpc/
│   ├── sys/
│   │   ├── client/
│   │   ├── etc/
│   │   ├── gen/
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── logic/
│   │   │   ├── server/
│   │   │   └── svc/
│   │   ├── proto/
│   │   └── Dockerfile
│   ├── ums/
│   ├── pms/
│   ├── oms/
│   ├── sms/
│   ├── cms/
│   └── search/
├── consumer/
│   ├── etc/
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   ├── logic/
│   │   ├── mq/
│   │   └── svc/
│   └── Dockerfile
├── job/
│   ├── etc/
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   ├── jobs/
│   │   ├── logic/
│   │   └── svc/
│   └── Dockerfile
├── pkg/
│   ├── errorx/
│   ├── es/
│   ├── mq/
│   ├── pointerprocess/
│   ├── time_util/
│   ├── scope/                           # 规划新增：统一作用域模型与 helper
│   ├── audit/                           # 规划新增：审计字段与上下文
│   ├── observability/                   # 规划新增：日志/指标/tracing helper
│   ├── idempotency/                     # 规划新增：幂等 key helper
│   └── antchain/                        # 规划新增：蚂蚁链开放联盟链统一集成层
├── web-admin/
│   ├── config/
│   ├── public/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ScopeContextBar/         # 规划新增
│   │   │   ├── OrderTimelinePanel/      # 规划新增
│   │   │   ├── PriceBreakdownCard/      # 规划新增
│   │   │   ├── BatchActionDock/         # 规划新增
│   │   │   └── SyncStatusBadge/         # 规划新增
│   │   ├── pages/
│   │   │   ├── cms/
│   │   │   ├── oms/
│   │   │   ├── pms/
│   │   │   ├── sms/
│   │   │   │   └── digital_card/        # 规划新增：抽卡与提货卡后台管理
│   │   │   ├── system/
│   │   │   ├── ums/
│   │   │   └── user/
│   │   ├── services/
│   │   ├── utils/
│   │   ├── app.tsx
│   │   └── global.tsx
│   └── tests/
├── flutter-mall/
│   ├── lib/
│   │   ├── config/
│   │   ├── layout/
│   │   │   ├── app_bootstrap.dart         # 规划新增：启动、登录恢复、升级闸门
│   │   │   └── intent_recovery_shell.dart # 规划新增：唤回与上下文恢复壳层
│   │   ├── model/
│   │   ├── provider/
│   │   │   ├── app_lifecycle_provider.dart
│   │   │   ├── intent_recovery_provider.dart
│   │   │   └── network_state_provider.dart
│   │   ├── utils/
│   │   │   ├── http_util.dart
│   │   │   └── shared_preferences_util.dart
│   │   ├── view/
│   │   │   ├── home/
│   │   │   ├── digital_card/            # 规划新增：抽卡页 / 我的卡片资产
│   │   │   ├── category/
│   │   │   ├── cart/
│   │   │   └── mine/
│   │   ├── widgets/
│   │   │   ├── cached_image_widget.dart
│   │   │   ├── commerce_state_shell.dart  # 规划新增：加载/空态/错误/弱网统一壳层
│   │   │   ├── permission_prompt_sheet.dart
│   │   │   ├── upgrade_gate_dialog.dart
│   │   │   ├── scope_context_bar.dart   # 规划新增
│   │   │   ├── order_timeline_panel.dart# 规划新增
│   │   │   └── price_breakdown_card.dart# 规划新增
│   │   └── main.dart
│   └── test/
├── docs/
├── script/
│   ├── account/
│   ├── configmap/
│   ├── *.yaml
│   └── sql/                             # 用于手工/阶段性 schema 辅助脚本
└── _opcos/
    ├── planning-artifacts/
    ├── implementation-artifacts/
    └── test-artifacts/
```

### Architectural Boundaries

**API Boundaries:**

- `admin-api`：平台治理、租户/商户运营、客服、履约、配置后台入口
- `front-api`：移动商城端与用户前台入口
- 前端只能访问网关，绝不能直接打 RPC 服务
- 支付、搜索、运营配置等第三方或异步场景由网关 / 服务端统一编排，不由客户端直连

**Component Boundaries:**

- Web 页面组件只负责呈现、交互与调度 `service.ts`
- Flutter widget 只负责视图和交互，不直接持有基础网络栈
- Handler 不写业务逻辑，Logic 不做视图层拼装细节，Svc 只负责依赖组织

**Service Boundaries:**

- `sys`：账号、角色、菜单、部门、数据范围、治理元数据
- `ums`：会员、地址、积分、成长值、收藏/关注等用户资产
- `pms`：商品、分类、品牌、属性、规格、SKU / SPU
- `oms`：购物车、订单、履约、退货、订单设置、支付记录
- `sms`：优惠券、秒杀、广告、推荐、活动编排，以及抽卡活动、卡池、卡片模板、提货卡资产台账与链上发放编排
- `cms`：专题、优选专区、内容运营
- `search`：商品搜索与搜索读模型
- `consumer`：异步事件副作用
- `job`：定时补偿、同步修复、巡检任务

**Data Boundaries:**

- 每个 RPC 服务拥有自己的 MySQL 聚合写边界
- 搜索索引由 `search` 服务与异步消费链共同维护
- `pkg/` 只承载基础设施与共享工具，不拥有业务数据真相
- 跨域联查通过 gRPC / MQ / 聚合 API 实现，不通过直接跨库读写实现

### Requirements to Structure Mapping

**Feature / FR Mapping:**

- **商品发现与内容导购**
  - Front: `api/front/internal/handler/{home,product}`, `api/front/internal/logic/{home,product}`
  - Service: `rpc/pms`, `rpc/cms`, `rpc/sms`, `rpc/search`
  - Web: `web-admin/src/pages/pms`, `web-admin/src/pages/sms`, `web-admin/src/pages/cms`
  - Mobile: `flutter-mall/lib/view/home`, `flutter-mall/lib/view/category`

- **会员账户与用户资产**
  - Front: `api/front/internal/handler/member`, `api/front/internal/logic/member`
  - Service: `rpc/ums`
  - Web: `web-admin/src/pages/ums`
  - Mobile: `flutter-mall/lib/view/mine`, `flutter-mall/lib/provider`

- **购物车与结算**
  - Front: `api/front/internal/handler/order`, `api/front/internal/logic/order`
  - Service: `rpc/oms`
  - Mobile: `flutter-mall/lib/view/cart`

- **订单交易与售后服务**
  - Admin: `api/admin/internal/logic/oms`
  - Front: `api/front/internal/logic/order`
  - Service: `rpc/oms`
  - Async / Compensation: `job/internal/jobs`, `consumer/internal/mq`

- **营销活动与内容运营**
  - Service: `rpc/sms`, `rpc/cms`
  - Admin: `web-admin/src/pages/sms`, `web-admin/src/pages/cms`
  - Front / Mobile: 首页导购、优惠券与活动入口

- **提货卡抽赏与链上资产**
  - Service: `rpc/sms`, `pkg/antchain`
  - Admin: `web-admin/src/pages/sms/digital_card`
  - Mobile: `flutter-mall/lib/view/digital_card`
  - Async: `consumer/internal/mq`, `job/internal/jobs`

- **搜索、同步与平台协同**
  - Service: `rpc/search`
  - Shared: `pkg/es`, `pkg/mq`
  - Async: `consumer`, `job`

- **后台管理与权限治理**
  - Admin: `api/admin/internal/logic/sys`, `web-admin/src/pages/system`
  - Service: `rpc/sys`
  - Shared: `pkg/scope`, `pkg/audit`

- **平台租户与多商户治理**
  - Shared: `pkg/scope`, `api/*/internal/middleware/scope.go`
  - Admin Governance UI: 以 `system` 域和后续平台治理页面承载
  - Domain Services: `sys/pms/oms/sms/cms/ums` 全部接入作用域过滤与写入校验

- **移动端召回、上下文恢复与渠道合规**
  - Front: `api/front/internal/middleware/client_meta.go`, `api/front/internal/logic/common`
  - Mobile Shell: `flutter-mall/lib/layout`, `flutter-mall/lib/provider`, `flutter-mall/lib/widgets`
  - Mobile Pages: `flutter-mall/lib/view/home`, `flutter-mall/lib/view/cart`, `flutter-mall/lib/view/mine/message`, `flutter-mall/lib/view/mine/order`
  - Shared Observability: `pkg/observability`

### Integration Points

**Internal Communication:**

- `web-admin -> admin-api -> RPC services`
- `flutter-mall -> front-api -> RPC services`
- `RPC services -> RabbitMQ -> consumer/search/job`
- `job -> RPC services` 做补偿与定时治理
- `message/push/activity entry -> Flutter intent recovery shell -> login/upgrade gate -> target page -> front-api revalidation`

**External Integrations:**

- 支付渠道（支付宝 / 微信等）
- Redis
- Elasticsearch
- RabbitMQ
- MongoDB
- 推送 / 站内消息分发渠道
- App 版本分发与升级元数据
- OSS / 图片资源服务（如项目后续启用）
- 蚂蚁链开放联盟链、实人认证与内容安全能力

**Data Flow:**

- 商品发布流：PMS/CMS/SMS 写库 -> 事件 -> Search 更新索引 -> Front/Home/Search 消费
- 订单流：Front 下单 -> OMS 写库 -> 支付发起 -> 回调更新 -> Job/Consumer 补偿库存/优惠券/积分
- 提货卡流：Front 发起抽卡 -> SMS 落抽卡记录与卡片实例 -> 事件触发链上发放任务 -> `pkg/antchain` 调用蚂蚁链开放联盟链 -> 回写 token / 回执 / 失败原因 -> Job / Consumer 补偿或重试 -> Front 资产页查询最新状态
- 平台治理流：Admin 审核租户/商户 -> Sys / 业务域更新作用域元数据 -> Front/Admin 查询即时生效
- 召回恢复流：消息/活动入口 -> App 恢复目标意图 -> 登录 / 升级 / 权限闸门 -> 目标页重新拉取服务端状态 -> 失败时给替代落点

### File Organization Patterns

**Configuration Files:**

- 服务运行配置：各服务 `etc/*.yaml`
- 发布配置：`script/*.yaml`
- 工件配置：`target/<service>/*.yaml`

**Source Organization:**

- 业务代码按子系统归属，不跨域乱放
- 共享基础能力集中在 `pkg/`
- 规划性新增目录必须优先挂在现有结构之下，不得另起平行根目录

**Test Organization:**

- Go：按服务 / 包内定向测试与 smoke build
- Web：`web-admin/tests` + 组件 / 页面测试
- Mobile：`flutter-mall/test`

**Asset Organization:**

- Web 静态资源：`web-admin/public` / `images`
- Flutter 图片资源：`flutter-mall/images`
- API 静态文档：`api/*/static`

### Development Workflow Integration

**Development Server Structure:**

- Backend 本地由 `make build/start` 或 `service_manager.sh` 驱动
- Web 本地由 `npm run start:dev` 驱动，依赖 `/api` 代理
- Flutter 本地依赖 `flutter run` 与共享 `HttpUtil`

**Build Process Structure:**

- `.api/.proto` 变更后使用 `make gen`
- GORM 生成源变更后使用 `make model`
- 后端构建产物统一落 `target/`

**Deployment Structure:**

- Dockerfile 已按服务拆分
- K8s manifest 已按服务拆分
- 后续 CI/CD 直接围绕现有目录结构增量建设，不重构部署资产根布局

## Architecture Validation Results

### Coherence Validation

**Decision Compatibility:**

- 所有核心决策都围绕“保留 brownfield 基线、先修产品闭环与平台化治理、后做框架升级”展开，没有互相冲突的技术方向。
- `REST + gRPC + MQ + Job` 的通信拓扑与当前仓库现实完全一致，因此不会出现“架构文档要求一套、代码库实际是另一套”的落差。
- 将 MySQL 作为真相源、Redis/ES/MQ 作为辅助系统的决策，与交易一致性和搜索异步同步的 NFR 兼容。

**Pattern Consistency:**

- 实现模式与 `project-context.md` 中的关键约束对齐：Go 逻辑分层、Web `request`、Flutter `HttpUtil`、生成代码不可手改。
- 本文档没有试图强行统一不同子系统的历史命名习惯，而是规定“在各自子系统内保持一致”，能最大化降低代理冲突。

**Structure Alignment:**

- 项目结构章节既尊重现有仓库，也明确标出了 `pkg/scope`、`pkg/audit`、作用域中间件、共享 UI 组件这些首批应新增的位置。
- 结构边界与 PRD 的 9 大能力簇能一一对应，不存在“需求无法落到目录”的情况。

### Requirements Coverage Validation

**Epic / Feature Coverage:**

- 当前 `epic.md` 已完成覆盖 `FR1-FR64` 的 Epic/Story 刷新，移动端新增需求已进入正式 Story 链。
- 当前覆盖性验证的重点不再是补齐缺失范围，而是确保 Story 尺寸可执行、Flutter 相关组件命名一致，并与 readiness 结论收口。
- 在现有前提下，全部 9 个功能能力簇都已有明确的服务归属、网关边界和 UI 承载位置。

**Functional Requirements Coverage:**

- 64 条 FR 已通过“发现导购、会员资产、购物车结算、订单售后、营销内容、搜索同步、后台治理、平台化经营、移动端召回与恢复”九类能力进行架构归位。
- 新增 `FR60-FR64` 已映射到前台网关客户端上下文契约、Flutter 生命周期壳层、统一状态壳层、权限闸门和升级闸门。
- 交易闭环、后台闭环、租户 / 商户治理、复购闭环、搜索同步闭环与移动端恢复闭环均被映射到了现有服务与新增共享基础设施上。

**Non-Functional Requirements Coverage:**

- Performance：通过缓存、读模型、独立网关 / 服务扩展、前后端状态管理模式承接
- Security：通过 JWT claim 扩展、RBAC、数据范围、日志脱敏、最小权限承接
- Reliability：通过幂等、Job 补偿、MQ 重试、审计与告警承接
- Scalability：通过独立服务扩展、MQ 解耦、ES 读模型承接
- Accessibility / UX：通过统一状态语义、作用域条、错误恢复、`Commerce State Shell` 与一致的共享组件承接
- Tenant Isolation：通过统一作用域模型、服务端资源校验、缓存 key 与事件 payload 作用域字段承接
- Mobile Lifecycle & Compliance：通过 `Intent Recovery Loop`、场景化权限闸门、升级闸门与恢复指标承接

### Implementation Readiness Validation

**Decision Completeness:**

- 已明确记录当前基线版本、上游参考版本、是否升级、为何不升级。
- 已明确关键决策优先级、数据与通信拓扑、鉴权模式、部署与观测方向。
- 已把 `FR60-FR64` 对应的生命周期、意图恢复、权限与升级策略提升为正式架构决策，而不是实现备注。

**Structure Completeness:**

- 已给出完整根目录树与关键新增目录建议。
- 已把主要 FR 类别映射到实际目录和服务。
- 已为移动端恢复壳层、状态壳层、权限闸门和升级闸门给出落位目录。

**Pattern Completeness:**

- 已对命名、结构、格式、通信、流程五类冲突点给出可执行规则。
- 已明确 AI 代理在生成代码、作用域、请求层和共享依赖方面的硬约束。
- 已补齐多代理最容易分叉的移动端恢复、消息唤回、权限请求和升级中断模式。

### Gap Analysis Results

**Critical Gaps:**

- 当前未发现新的架构级阻塞缺口；下一步重点是通过 readiness 复核确认 PRD、UX、Architecture 与 Epic/Story 已完全对齐。

**Important Gaps:**

- 仓库中尚未形成标准 CI 工作流文件，需在实现阶段补齐。
- 平台 / 租户 / 商户统一作用域包、网关作用域中间件与审计上下文尚未在代码中落地。
- App 版本闸门、消息唤回意图对象和统一状态壳层还未在代码中形成共享实现。

**Nice-to-Have Gaps:**

- Web / Flutter 共享 design token 仍未抽出统一实现
- 消息通知、物流轨迹、评价等扩展能力还缺独立服务化规划
- 事件注册表和补偿策略可以在后续文档中进一步模板化

### Validation Issues Addressed

- 已明确拒绝“为追新而全量升级技术栈”的高风险路线，改为“业务闭环优先、升级后置”的低风险路线。
- 已将 `FR60-FR64`、`Commerce State Shell`、`Intent Recovery Loop`、`Permission Request Sheet` 和 `Upgrade Gate` 正式纳入架构主文档。
- 已将 Epic/Story 规划纳入整体实施验证链路，因此后续 readiness 校验可以直接对照 Story 清单检查跨文档一致性。

### Architecture Completeness Checklist

**✅ Requirements Analysis**

- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed
- [x] Technical constraints identified
- [x] Cross-cutting concerns mapped

**✅ Architectural Decisions**

- [x] Critical decisions documented with versions
- [x] Technology stack fully specified
- [x] Integration patterns defined
- [x] Performance considerations addressed

**✅ Implementation Patterns**

- [x] Naming conventions established
- [x] Structure patterns defined
- [x] Communication patterns specified
- [x] Process patterns documented

**✅ Project Structure**

- [x] Complete directory structure defined
- [x] Component boundaries established
- [x] Integration points mapped
- [x] Requirements to structure mapping complete

### Architecture Readiness Assessment

**Overall Status:** ARCHITECTURE READY / DOWNSTREAM EPICS REQUIRE REFRESH

**Confidence Level:** 高

**Key Strengths:**

- 明确保留了既有仓库的现实边界，避免脱离代码库的理想化设计
- 把交易一致性、搜索同步、平台化作用域与多端状态一致性提升为一等架构约束
- 为多 AI 代理协作补齐了命名、结构、格式、通信与流程五类一致性规则
- 已把移动端恢复、消息唤回、弱网、权限与升级治理从 UX 约束上升为可落地的架构边界

**Areas for Future Enhancement:**

- 单列技术升级 Epic（Go 1.26 / go-zero 1.10 / React 19 / Umi 4 / Antd 6）
- 统一前后端与移动端的类型契约生成链
- 将通知 / 物流 / 留存分析进一步独立成清晰服务边界

### Implementation Handoff

**AI Agent Guidelines:**

- 严格按本文档定义的作用域、边界、事件和目录规则实现
- 先从共享基础设施与关键链路改造入手，不从页面微调或零散功能开始
- 任何生成文件改动都必须回到源定义和生成流程
- 实现时优先保持与当前仓库局部约定一致，而不是追求跨子系统表面统一

**First Implementation Priority:**

第一批实现故事应围绕以下内容展开：

1. 先基于已收口的 `epic.md` 重新确认实施顺序，不再沿用旧的 sprint / story 状态直接推进
2. 建立 `pkg/scope`、网关 `scope` / `audit` / `idempotency` 中间件
3. 把 `oms/pms/sms/cms/ums` 的关键查询与写操作接入平台 / 租户 / 商户作用域校验
4. 建立移动端 `Intent Recovery Loop`、`Commerce State Shell`、权限闸门与升级闸门
5. 建立订单补偿、搜索同步与关键审计日志的统一事件与观测基线

## Completion Summary & Next Workflow Recommendations

架构工作流已经完成，当前文档已经吸收最新 PRD / UX / readiness 对 `FR60-FR64` 的修订，并与最新 Epic/Story 链完成收口，可以继续作为实现阶段的技术单一事实源使用。按照 BMAD 流程，建议的后续步骤是：

1. **Check Implementation Readiness**
   - Command: `/bmad-bmm-check-implementation-readiness`
   - Agent: 🏗️ Architect
   - 目的：重新校验 PRD、UX、Architecture 与 Epics/Stories 是否已经完全对齐

2. **Sprint Planning**
   - Command: `/bmad-bmm-sprint-planning`
   - Agent: 🏃 Scrum Master
   - 目的：在 readiness 通过后生成实现阶段的 sprint plan

3. **Create Story / Code Review**
   - Command: `/bmad-bmm-create-story` 或 `/bmad-bmm-code-review`
   - Agent: 🏃 Scrum Master / 💻 Developer Agent
   - 目的：根据新的 sprint 结果决定是继续推进下一个故事，还是先收口当前已进入 review 的故事

如需继续，我建议下一步直接进入 **Check Implementation Readiness**，先确认规划层已经完全对齐，再刷新实施队列。
