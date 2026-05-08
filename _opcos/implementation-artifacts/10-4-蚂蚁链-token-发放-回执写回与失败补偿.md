# Story 10.4: 蚂蚁链 token 发放、回执写回与失败补偿

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 技术运营人员,
I want 让中奖卡片异步发放蚂蚁链 token 并可补偿追踪,
so that 链上确权失败时我可以及时识别、重试和人工介入。

## Acceptance Criteria

1. **Given** 本地卡片实例已创建且满足链上发放条件 **When** 异步发放任务执行 **Then** 系统调用蚂蚁链开放联盟链完成 token 发放 **And** 回写 token 标识、链上状态、失败原因和最近执行结果。

2. **Given** 链上发放超时、失败或回写异常 **When** 技术运营查看链路状态 **Then** 系统展示失败原因、重试次数和补偿记录 **And** 支持对失败链路执行重试、冻结或人工复核。

## AC 覆盖矩阵

- **AC1（异步发放链路 + 回执写回 + token 映射）**：Task 1.1-1.8、2.1-2.7、3.1-3.12、4.1-4.7
- **AC2（失败补偿 + 重试/冻结/人工复核 + 运维可见）**：Task 1.3-1.8、2.4-2.8、3.4-3.12、5.1-5.8

## 背景与关键风险（必须优先阅读）

> **⚠️ CRITICAL - 10.4 只能建立在 10.3 的链下资产真相源之上，绝不能绕过 `sms_card_instance` 重新发卡。** 当前仓库已经通过 `rpc/sms/internal/logic/cardassetservice/` 保证“中奖记录 -> 本地资产实例”幂等成立，且 `sms_card_instance` 已固定 `mint_status = mint_pending` 作为链上发放前状态。10.4 必须直接消费这个资产实例，而不是再次根据抽卡记录重新推断模板、重建资产，或把链上回执当作第一次落账时机。
>
> **⚠️ CRITICAL - 当前仓库还没有 `pkg/antchain`，也没有 `sms_card_mint_task`。** 这意味着 10.4 不是“把一个现成 SDK 接上去”这么简单，而是第一次把“统一链上集成层 + 发放任务表 + consumer/job 补偿治理 + 运维干预接口”真正落到现有仓库里。任何把链上调用散落到 `ParticipateDraw`、front 聚合、Flutter 或临时脚本的做法，都会直接破坏后续可维护性。
>
> **⚠️ CRITICAL - 链上发放必须异步执行，不能卡在抽卡请求事务里。** PRD、架构和 UX 都把“中奖即有资产真相源、链上处理中有明确状态”写成了产品要求。10.4 必须遵循“事务内确保发放任务存在，事务外异步触发链上调用”的模式；如果把外部链上调用放进 `ParticipateDraw` 的事务内，一旦蚂蚁链接口超时或抖动，就会把前台中奖结果路径拖进高延迟甚至回滚。
>
> **⚠️ CRITICAL - MQ 投递成功不等于链路可恢复，链下任务表才是 10.4 的主真相源。** 现有 `pkg/mq` 已支持持久化消息和手动 ACK，但没有 publisher confirm 封装，也没有“消息一定入队”的数据库联动保证。根据 RabbitMQ 官方数据安全文档，publisher confirm 和 consumer ack 解决的是不同方向的数据安全问题。对当前仓库来说，10.4 不能把 `SendMessage` 返回成功当成唯一排队成功信号，必须先把 `sms_card_mint_task` 持久化，再用 `consumer` 触发执行、`job` 扫描兜底。
>
> **⚠️ CRITICAL - 当前仓库的 `consumer`、`job`、`rpc/sms` 都不是自动发现 wiring，10.4 必须把接线点补完整。** 已有实现表明：`consumer/internal/svc/service_context.go` 通过显式 goroutine 启动订阅，`job/internal/logic/job_logic.go` 通过 `switch req.Name` 手工分发任务，`rpc/sms/sms.go` 手工注册 gRPC service。也就是说，10.4 不能只新增 `consumer/internal/mq/digital_card/*.go`、`job/internal/jobs/*.go` 或 `cardminttaskservice/*.go` 就算完成；如果这些入口没接上，代码会“写完但永远不会执行”。
>
> **⚠️ CRITICAL - 10.4 需要三层幂等，不是只做一层唯一键。** 至少要同时保证：
> 1. 同一 `sms_card_instance` 只创建一条有效发放任务；
> 2. 同一发放任务对蚂蚁链只持有一个稳定幂等键；
> 3. 同一 token / 回执回写不会把多个实例绑定到同一链上结果。
> 否则在“消息重复消费 / 网络超时后重试 / 人工回放 / job 补偿”任一场景下，都可能出现重复发 token。
>
> **⚠️ CRITICAL - `asset_status`、`mint_status` 和链上状态不是一回事，绝不能混写。** 10.3 已经明确 `asset_status` 描述资产生命周期，`mint_status` 描述发放阶段。10.4 还要引入链上回执摘要和 token 标识。若把这些状态混成一个字段，后续 10.5 的资产详情、审计检索和运营冻结会很快失真。
>
> **⚠️ CRITICAL - 失败补偿必须复用同一执行器，不能形成“实时执行一套、重试一套、job 一套”。** 10.4 至少会出现三种入口：consumer 收到发放请求、后台人工重试、job 扫描超时或漏投任务。三者必须共用同一个“锁任务 -> 调链 -> 回写 -> 记日志”的核心逻辑，否则线上线下行为很容易分叉。
>
> **⚠️ CRITICAL - 运维干预不是前台展示问题，而是受作用域保护的后台能力。** Epic 10.4 明确“技术运营查看链路状态并执行重试、冻结或人工复核”。这类接口和页面只能在 admin / web-admin 中出现，并沿用现有 `oms/chain-monitor` 的作用域校验和动作语义；不允许把重试入口暴露给前台会员接口，也不允许让普通商户越权查看其他主体的 token / 回执。
>
> **⚠️ CRITICAL - 10.4 必须给 10.5 留下可消费的状态时间线。** UX 规范已经定义了 `实例已创建 -> 发放请求中 -> 链上确认成功 -> 补偿中 -> 人工复核 -> 已冻结` 的时间线语义。10.4 需要把这些状态稳定落进数据库与日志，而不是只写一条“成功/失败”的粗粒度结果。

## 明确非目标（本 Story 不做）

- 不重做抽卡活动配置、实名认证校验、抽卡参与和本地资产建账；这些分别属于 10.1、10.2、10.3。
- 不建设完整“我的提货卡资产列表 / 详情页”与客服审计详情页；完整资产展示与审计检索属于 10.5。
- 不在 Flutter、front-api handler 或页面组件里直连蚂蚁链接口，也不通过临时脚本绕过 `pkg/antchain`。
- 不把“人工重试”实现成直接修改数据库字段后假装成功；所有重试、冻结、人工复核都必须经过受控服务逻辑与日志留痕。
- 不在本 Story 顺手升级 go-zero、GORM、RabbitMQ 客户端或引入新的消息中间件。
- 不开放提货卡二级交易、连续挂牌、收益承诺、虚拟币计价或其他违反去金融化约束的能力。

## 实施顺序（必须按序推进）

1. 先补 `sms_card_mint_task` 及 `sms_card_instance` 的链上发放摘要字段，明确任务真相源、唯一约束与状态机。
2. 再补 `pkg/antchain` 统一集成层和 `rpc/sms/internal/config/config.go` / `svc/service_context.go` 注入，保证链上调用边界收口。
3. 然后补 `sms` 域内部的“确保发放任务存在 / 执行发放 / 回写回执 / 记录失败补偿”的共享逻辑。
4. 再把任务触发接到 MQ + consumer，并补 job 扫描兜底，确保“消息丢失 / 超时 / 漏执行”有恢复路径。
5. 再补 admin / web-admin 的链路查询与干预动作，沿用现有 `chain-monitor` 的模式实现重试、冻结和人工复核。
6. 最后补最小前台/会员状态透出、专项测试和可观测指标，确保 10.5 可以在不返工底层链路的前提下继续向上构建。

## 完成判定（Definition of Done）

- 任一 `mint_pending` 的本地资产实例都可以被幂等地生成一条发放任务，并拥有稳定的任务状态与幂等键。
- consumer、人工重试、job 补偿三种入口重复执行同一资产时，不会创建第二条有效发放任务，也不会重复发 token。
- 链上发放成功后，可以在本地稳定回写 `token 标识 / 链上状态 / 最近执行结果 / 最近回执时间`，并能追溯到对应资产实例与中奖记录。
- 链上超时、失败、回执写回异常时，系统会把任务推进到明确的失败/补偿/人工复核状态，而不是停留在“处理中”无结论状态。
- 技术运营可以在授权范围内查看链路状态、失败原因、重试次数、最近执行时间和可用干预动作。
- 冻结、重试、人工复核等干预动作都会留下审计记录，并遵守平台/租户/商户作用域校验。
- 现有前台抽卡结果和个人记录不会被 10.4 破坏；如补充最小状态透出，必须仍然遵循“服务端确认优先”。
- 10.5 可以直接复用 10.4 产出的 token 状态、时间线日志和任务摘要，而不需要重新定义链上发放基础模型。

## Tasks / Subtasks

### 一、发放任务 DDL 与状态机基线

#### Task 1 (AC: 1, 2): 为链上发放建立本地任务真相源

- [x] 1.1 在 `script/sql/sms/` 新增迁移，例如 `migration_20260418_card_mint_task.sql`，至少创建 `sms_card_mint_task`。
- [x] 1.2 `sms_card_mint_task` 至少包含：`platform_id`、`tenant_id`、`merchant_id`、`asset_instance_id`、`participation_record_id`、`activity_id`、`member_id`、`request_id`、`trace_id`、`idempotency_key`、`task_status`、`mint_status`、`chain_status`、`token_id`、`chain_tx_id`、`retry_count`、`max_retry_count`、`last_error_code`、`last_error_reason`、`last_receipt_summary`、`last_receipt_json`、`last_execute_at`、`next_retry_at`、`manual_required`、`frozen`、`freeze_reason`、`create_by/update_by`、`create_time/update_time`、`is_deleted`。
- [x] 1.3 `sms_card_mint_task` 必须具备至少两类唯一性约束：
  - `asset_instance_id + is_deleted` 或等价“每个资产仅一条有效任务”约束
  - `idempotency_key + is_deleted` 或等价“对链请求幂等键唯一”约束
- [x] 1.4 扩展 `sms_card_instance`，至少补链上摘要字段：`token_id`、`chain_status`、`last_receipt_at`、`mint_task_id` 或等价最小快照；`sms_card_instance` 继续作为用户资产真相源，`sms_card_mint_task` 作为发放过程真相源。
- [x] 1.5 `sms_card_asset_log` 继续承担状态时间线，10.4 至少新增这些操作语义：
  - `mint_requested`
  - `mint_dispatching`
  - `mint_succeeded`
  - `mint_failed`
  - `mint_retry_requested`
  - `mint_frozen`
  - `mint_manual_review`
- [x] 1.6 状态机必须显式区分：
  - `task_status`：任务调度/补偿层状态（pending/dispatched/running/succeeded/failed/manual_review/frozen）
  - `mint_status`：资产发放阶段状态（mint_pending/mint_processing/mint_success/mint_failed/mint_compensating/mint_manual_review/mint_frozen）
  - `chain_status`：链上结果摘要（processing/success/failed/unknown/frozen）
- [x] 1.7 所有与 token / 回执 / 失败原因有关的字段都要带查询索引或组合索引，至少覆盖：
  - `task_status + next_retry_at + id`
  - `chain_status + update_time + id`
  - `activity_id + member_id + id`
  - `token_id + is_deleted`
- [x] 1.8 如需生成 GORM model/query，必须遵守项目生成链约束；未获用户授权前，不把“修改 proto/api 后再批量生成”当作默认修复路线。

### 二、统一链上集成层与内部协议

#### Task 2 (AC: 1, 2): 让蚂蚁链接入、任务执行和后台干预拥有稳定接口边界

- [x] 2.1 在 `pkg/antchain/` 新增统一集成层，至少拆出：
  - `client.go`：对外统一调用入口
  - `types.go`：请求/响应/回执结构
  - `mock.go` 或等价测试桩
  不要把第三方 SDK 细节散落到 `logic`、`consumer` 或 `job`。
- [x] 2.2 在 `rpc/sms/internal/config/config.go` 为蚂蚁链增加独立配置段，例如 `AntChain.Endpoint`、`AppId`、`AccessKey`、`Secret`、`TimeoutSeconds`、`Enabled`；真实凭据只允许来自配置文件或环境变量，不得写死在代码和 story 文档中。
- [x] 2.3 在 `rpc/sms/internal/svc/service_context.go` 注入 antchain client，保持与现有 `DB`、`RabbitMQ` 一样由 `svcCtx` 统一传递。
- [ ] 2.4 内部 RPC / proto 建议新增 `rpc/sms/proto/card_mint_task.proto` 或等价服务定义，至少包含：
  - `EnsureCardMintTask`
  - `ExecuteCardMintTask`
  - `QueryCardMintTaskList`
  - `QueryCardMintTaskDetail`
  - `RetryCardMintTask`
  - `FreezeCardMintTask`
  - `EscalateCardMintTask`
  也可以按仓库习惯扩展 `card_asset.proto`，但不要把任务、资产、活动全部揉进一个超大 message。
- [x] 2.5 协议语义必须明确：
  - `asset_status` 描述资产生命周期
  - `mint_status` 描述发放阶段
  - `task_status` 描述调度/补偿任务状态
  - `chain_status` 描述链上结果摘要
  任何一个字段都不能代替另外三个。
- [x] 2.6 admin 侧为提货卡补偿链路新增独立 API 文档文件，例如 `api/admin/doc/api/sms/digital_card_chain.api`，并在 `api/admin/doc/api/sms/sms.api` 中显式引入。
- [x] 2.7 如需向 front/Flutter 透出最小链上状态，只允许增加服务端聚合字段，例如 `mintStatus`、`mintStatusText`、`chainStatus`；不得让客户端自行根据本地时间或文案猜测“已到账”。
- [x] 2.8 所有后台动作接口必须携带并校验 `scopeType/platformId/tenantId/merchantId`，复用现有治理范围模式。
- [ ] 2.9 如新增 `CardMintTaskService` 或等价 gRPC service，必须同步在 `rpc/sms/sms.go` 完成 server 注册；只创建 logic / proto 文件但不注册，不算完成。

### 三、SMS 域发放编排、回执写回与幂等执行器

#### Task 3 (AC: 1, 2): 让“任务创建 -> 异步发放 -> 回写 -> 补偿”成为同一套可重入逻辑

- [x] 3.1 在 `rpc/sms/internal/logic/` 下新增 `cardminttaskservice/` 或等价目录，集中实现发放任务编排；不要把 10.4 的全部逻辑继续塞进 `cardassetservice` 或 `ParticipateDraw`。
- [x] 3.2 提供一个共享 helper / service 方法，例如 `EnsureCardMintTaskByAssetInstance`，输入至少包含 `asset_instance_id`；该方法必须在以下场景复用：
  - 实时中奖后的链上发放排队
  - 历史 `mint_pending` 资产回填任务
  - consumer 重复消费兜底
  - 后台人工重试
  - job 补偿重扫
- [ ] 3.3 创建发放任务时，必须校验资产实例前置条件：
  - 资产实例存在且未删除
  - `asset_status` 允许继续发放
  - `mint_status` 为 `mint_pending` 或等价可重试状态
  - 所属活动 / 模板 / 实名 / 合规快照未处于明确禁止发放状态
- [x] 3.4 创建任务与更新 `sms_card_instance.mint_status` 必须在显式事务内完成；由于 `rpc/sms` 已开启 `SkipDefaultTransaction: true`，不能依赖默认事务。
- [x] 3.5 外部链上调用不得放入创建任务事务内；正确顺序应为：
  1. 事务内确保任务存在并把实例状态推进到“待发放/待调度”
  2. 事务提交
  3. 事务外通过 MQ 或受控执行器触发真正链上请求
- [x] 3.6 发放执行器必须使用任务级锁或等价保护（例如 `SELECT ... FOR UPDATE` + 状态机判断），确保同一任务在 consumer 重复消息、人工重试和 job 同时扫描下只会有一个执行者真正出链。
- [x] 3.7 对蚂蚁链的请求必须携带稳定幂等键；推荐以 `asset_instance_id` 或 `mint_task.id` 为核心，组合活动/模板/环境信息生成，不允许用随机字符串临时拼接后丢失。
- [x] 3.8 发放成功后，至少回写这些结果：
  - `sms_card_mint_task.token_id`
  - `sms_card_mint_task.chain_status`
  - `sms_card_mint_task.last_receipt_summary`
  - `sms_card_mint_task.last_receipt_json`
  - `sms_card_mint_task.last_execute_at`
  - `sms_card_instance.token_id`
  - `sms_card_instance.chain_status`
  - `sms_card_instance.mint_status = mint_success`
  - `sms_card_asset_log` 追加成功日志
- [x] 3.9 发放失败、超时、回执写回异常时，必须把任务推进到 `mint_failed` / `mint_compensating` / `mint_manual_review` 的明确状态之一，并记录最近失败原因；不允许维持“processing”假象。
- [x] 3.10 任何一次链上执行都要保留 `trace_id`、`request_id`、`scope`、操作者、失败原因、最近回执摘要，以便 10.5 和后台审计直接复用。
- [ ] 3.11 如果一次执行里“链上已成功，但本地回写失败”，必须定义可恢复策略：
  - 优先依据任务幂等键重新查询或补写回执
  - 严禁简单重发第二次 token 请求
- [x] 3.12 所有“重试 / 冻结 / 人工复核”动作都必须复用同一发放执行器与状态机，不得单独复制一套逻辑。

### 四、MQ 触发、Consumer 执行与 Job 补偿兜底

#### Task 4 (AC: 1, 2): 让异步链路既能跑通，也能在消息抖动和外部超时下恢复

- [x] 4.1 事件命名遵循架构文档的 `{domain}.{entity}.{action}.v{n}` 约定，建议新增：
  - `sms.digital_card.mint_requested.v1`
  - 如需回写查询补偿，可新增 `sms.digital_card.mint_reconcile.v1`
- [x] 4.2 在 `consumer/internal/mq/digital_card/` 新增对应的 consumer 逻辑文件，并优先使用 `pkg/mq` 的 `ConsumeSimpleWithAck` 手动 ACK 模式；只有在任务状态和日志已经稳定写库后才 ACK。
- [x] 4.3 consumer 处理失败时应 NACK 并重新入队或把任务推进到补偿状态，由任务状态机决定是否继续自动重试；不要吞错后直接 ACK。
- [x] 4.4 新增 consumer 逻辑后，必须同步把订阅 wiring 接到 `consumer/internal/svc/service_context.go`；当前仓库不是自动扫描 `consumer/internal/mq/` 目录，未接线即不会消费。
- [x] 4.5 RabbitMQ 的 exchange / queue / routing key 命名应对齐现有仓库风格，至少保持：
  - exchange: `*.event.exchange`
  - queue: `*.queue`
  - key: `*.key`
- [x] 4.6 在 `job/internal/jobs/` 新增兜底扫描任务，例如 `handle_card_mint_timeout_logic.go`，至少负责：
  - 扫描长时间停留在 `pending/dispatched/processing` 的任务
  - 识别“消息可能漏投”与“执行可能超时”的任务
  - 按重试上限与 `next_retry_at` 触发重放或升级人工复核
- [x] 4.7 新增 job 后，必须同步把任务名接到 `job/internal/logic/job_logic.go` 的 `switch req.Name` 分发；只放一个新 job 文件而不接分发，不算完成。
- [x] 4.8 job 补偿必须复用 `ExecuteCardMintTask` 或同一共享执行器，不允许在 job 里手写另一套链上调用和回写逻辑。
- [x] 4.9 若 MQ publish 失败，但本地任务已成功创建，必须保留任务为 `pending_dispatch` 或等价状态，依赖 job 扫描补发；不得把“消息没发出去”伪装成“任务不存在”。
- [ ] 4.10 需要最小观测指标与日志：
  - 发放任务创建量
  - 成功率 / 超时量 / 失败量
  - 自动重试次数
  - 人工复核数量
  - token 回执对账差异
  - traceId/requestId/scope 级日志串联

### 五、后台运维工作台与干预动作

#### Task 5 (AC: 2): 让技术运营能在授权范围内看见问题并立即处理

- [x] 5.1 参考 `oms/chain-monitor`，在 admin 侧新增提货卡链路查询逻辑，例如：
  - `api/admin/internal/logic/sms/digital_card_chain/query...`
  - `retry...`
  - `freeze...`
  - `escalate...`
- [x] 5.2 查询列表至少支持按这些维度过滤：
  - 活动 ID / 活动名
  - 会员 ID
  - 卡片模板 ID
  - `asset_no`
  - `token_id`
  - `task_status`
  - `mint_status`
  - `chain_status`
  - `manual_required`
  - 时间范围
- [x] 5.3 列表项至少展示：`traceId`、`assetNo`、`memberId`、`activityId`、`templateId`、`tokenId`、`taskStatus`、`mintStatus`、`chainStatus`、`retryCount`、`lastError`、`lastExecuteAt`、`manualRequired`、`frozen`。
- [x] 5.4 动作语义建议直接复用现有补偿链路心智：
  - `retry`：重新执行同一任务
  - `freeze`：冻结资产发放或后续流转
  - `escalate`：升级为人工复核
  如确有需要可补 `replay`，但不要先设计一套和 OMS 完全不同的动作词典。
- [x] 5.5 在 `web-admin/src/pages/sms/` 下新增提货卡链路工作台页面，例如 `DigitalCardChainMonitor/`，优先复用 `oms/chain-monitor` 的“筛选区 + 列表 + 详情抽屉 + 动作按钮”结构。
- [x] 5.6 详情抽屉至少展示：
  - 资产实例与中奖来源快照
  - 最近一次回执摘要
  - 完整状态时间线
  - 失败原因与重试历史
  - 可执行动作
- [x] 5.7 所有后台查询与动作都必须做平台/租户/商户作用域校验；默认不向普通商户暴露平台级或跨租户发卡数据。
- [x] 5.8 冻结或人工复核动作必须同步写入 `sms_card_asset_log`，并保留操作人、原因、traceId 和时间。
- [x] 5.9 admin 接线必须补齐完整链路：
  - `api/admin/internal/types/` 的请求/响应结构
  - `api/admin/internal/logic/sms/digital_card_chain/` 的查询与动作逻辑
  - `api/admin/internal/handler/sms/digital_card_chain/` 的 handler 入口
  - `web-admin` 对应 `service.ts`、`data.d.ts` 和页面调用
  不能只写 RPC logic 或只补一个前端页面。

### 六、最小前台状态承接与测试保护网

#### Task 6 (AC: 1, 2): 保证前台不说错话，后台和异步链路不悄悄失真

- [ ] 6.1 如需让现有抽卡结果页或个人记录反映 10.4 状态，只做最小字段扩展：
  - `mintStatus`
  - `mintStatusText`
  - `chainStatus`
  - 必要时补 `tokenStatusText`
  完整资产列表和详情仍交给 10.5。
- [ ] 6.2 `api/front/internal/logic/digital_card/draw_activity/helper.go` 与 Flutter 模型只能透传服务端确认状态，不允许客户端根据倒计时、轮询失败或旧文案猜测“已到账/失败”。
- [ ] 6.3 `rpc/sms` 单测至少覆盖：
  - 首次创建发放任务成功
  - 同一资产重复创建任务仍返回同一任务
  - consumer 重复执行不重复发 token
  - 链上成功但回写失败的恢复路径
  - 超时 / 失败 / 人工复核状态推进
  - scope 不匹配时禁止干预
- [ ] 6.4 `consumer` 测试至少覆盖：
  - 手动 ACK 成功路径
  - handler 返回错误时 NACK 重入队
  - 重复消息不导致重复发 token
- [ ] 6.5 `job` 测试至少覆盖：
  - pending_dispatch 任务补发
  - processing 超时任务推进补偿
  - 达到重试上限后升级人工复核
- [x] 6.6 `api/admin` / `web-admin` 测试至少覆盖：
  - 列表筛选与字段映射
  - retry/freeze/escalate 可用动作边界
  - 跨主体越权访问被拒绝
- [ ] 6.7 如补充 front/Flutter 状态字段，测试至少覆盖：
  - `链上处理中`
  - `补偿中`
  - `已到账`
  - `人工复核中`
  不允许把补偿中误报成失败或成功。
- [ ] 6.8 最低验收命令建议为：

```bash
go test ./rpc/sms/internal/logic/cardminttaskservice/... -v
go test ./consumer/internal/mq/... -v
go test ./job/internal/jobs/... -v
go test ./api/admin/internal/logic/sms/... -v
go test ./api/front/internal/logic/digital_card/... -v
cd web-admin && npm run test -- DigitalCardChainMonitor
cd flutter-mall && flutter test
```

## Dev Notes

### Previous Story Intelligence

1. Story 10.1 已经把活动、卡池、模板、版权与合规规则固定在 `sms` 域，10.4 不能绕开这些配置直接“裸发 token”。
2. Story 10.2 已经把实名认证、资格校验、请求幂等和中奖记录落到了真实仓库；链上发放不应重新校验抽卡随机结果，只能复用既有中奖记录与实名快照。
3. Story 10.3 已经把本地资产台账、`cardassetservice` 和 `sms_card_instance` 建好了，并明确：
  - `asset_status = asset_created`
  - `mint_status = mint_pending`
  - `sms_card_asset_log` 已是资产时间线真相源之一
4. 因此 10.4 的核心不是“再证明用户中奖了”，而是把“已有本地资产实例”可靠推进到“已发链 / 补偿中 / 人工复核 / 冻结”这些可运营状态。
5. Story 10.5 将直接消费 10.4 产出的 token 状态、时间线和干预结果，所以 10.4 需要把状态语义一次性定义稳。

### Git Intelligence Summary

1. 最近与 Epic 10 直接相关的提交是 `bed45518 feat: complete draw participation asset flow`，它已经把改动落到了：
  - `rpc/sms/internal/logic/cardassetservice/`
  - `script/sql/sms/migration_20260417_card_asset_ledger.sql`
  - `api/front/internal/logic/digital_card/draw_activity/`
  - `flutter-mall/lib/view/digital_card/`
2. 这说明当前仓库对提货卡链路的真实演进方式是“围绕既有模块增量补齐”，而不是新起服务或新起前台入口。
3. 现有后台已经有 `oms/chain-monitor` 成熟模式；10.4 应优先借鉴它的查询、动作、抽屉和作用域治理方式，而不是从零造一个完全不同的运维工作台。
4. `oms/chain-monitor` 可复用的不是单个页面，而是 `types + handler + logic + web-admin service + operate log/干预动作语义` 的整套接法；10.4 应整体借鉴，而不是只抄一个列表 UI。

### 架构约束（必须遵守）

1. **`sms` 域继续拥有提货卡资产台账与链上发放编排。**
2. **链下主台账优先于链上结果。** `sms_card_instance` / `sms_card_mint_task` 才是本地真相源，链上 token 只是确权结果。
3. **MQ 负责触发，job 负责兜底，二者都不是真相源。**
4. **外部链上能力必须通过统一集成层接入。**
5. **前端永远只消费聚合后的服务端状态，不直接操作链上接口。**
6. **所有后台干预动作都必须受治理范围约束。**

### 技术要求与 Guardrails

1. **不要在 `ParticipateDraw` 事务里直接调蚂蚁链。**
   - 实时请求只负责确保任务存在并异步排队
   - 外部 IO 必须在事务外执行
2. **`pkg/antchain` 是唯一允许出现链上 SDK 细节的地方。**
   - `logic` / `consumer` / `job` 只能依赖抽象 client
3. **任务创建与执行都必须幂等。**
   - 创建任务：按资产实例唯一
   - 发链请求：按稳定 idempotency key 唯一
   - 回执写回：按 token/task 唯一
4. **consumer 使用手动 ACK。**
   - 只有数据库状态和日志稳定写入后才 ACK
   - handler 失败时允许 NACK 重回队列
5. **现有 `pkg/mq` 发送端没有 publisher confirm 封装。**
   - 因此“publish 成功”只能视为“已尝试投递”
   - 不能视为“任务一定被 broker 接管且一定会消费”
   - 必须用 `sms_card_mint_task` + job 扫描兜底
6. **`asset_status` 与 `mint_status` 不得混写。**
   - 成功发链后，资产仍然是资产，只是发放阶段变化
7. **回执摘要要写快照，原始回执要保留但受控暴露。**
   - front 只看聚合文案
   - admin 才能看详细失败原因或回执摘要
8. **冻结 / 人工复核必须改变后续自动执行行为。**
   - 冻结后 consumer/job 不再自动出链
   - 人工复核状态需要显性停机位
9. **日志必须能串起“中奖记录 -> 资产实例 -> 发放任务 -> 链上回执 -> 人工动作”。**
10. **不要让 10.4 变成 10.5。**
   - 10.4 只补最小前台状态透出
   - 完整资产详情和搜索式审计留给 10.5
11. **新增 consumer 文件但没在 `consumer/internal/svc/service_context.go` 接线，不算完成。**
12. **新增 job 文件但没在 `job/internal/logic/job_logic.go` 分发，不算完成。**
13. **新增 gRPC server 但没在 `rpc/sms/sms.go` 注册，不算完成。**

### File Structure Requirements

- `script/sql/sms/migration_20260418_card_mint_task.sql`
  - 新增发放任务表，并按需扩展 `sms_card_instance`。
- `pkg/antchain/`
  - 统一的蚂蚁链开放联盟链接入层。
- `rpc/sms/proto/card_mint_task.proto`
  - 发放任务查询、执行和干预协议定义。
- `rpc/sms/internal/logic/cardminttaskservice/`
  - 任务创建、执行、查询、重试、冻结、升级逻辑。
- `rpc/sms/internal/logic/cardassetservice/`
  - 仅做资产真相源与最小快照配合，不承载全部 10.4 编排。
- `rpc/sms/sms.go`
  - 如新增 `CardMintTaskService`，必须在此注册 gRPC server。
- `consumer/internal/mq/digital_card/`
  - `mint_requested` 事件消费与执行入口。
- `consumer/internal/svc/service_context.go`
  - 显式启动提货卡发放相关订阅，不依赖目录自动扫描。
- `job/internal/jobs/`
  - pending / timeout / retry 补偿扫描任务。
- `job/internal/logic/job_logic.go`
  - 把新的补偿任务名接入 `switch req.Name` 分发。
- `api/admin/doc/api/sms/digital_card_chain.api`
  - 后台工作台接口。
- `api/admin/internal/types/`
  - 提货卡链路查询、详情和干预动作结构定义。
- `api/admin/internal/logic/sms/digital_card_chain/`
  - 列表查询、详情、retry/freeze/escalate 等逻辑。
- `api/admin/internal/handler/sms/digital_card_chain/`
  - 提货卡链路工作台 HTTP handler 入口。
- `web-admin/src/pages/sms/DigitalCardChainMonitor/`
  - 提货卡链路工作台页面。
- `api/front/doc/api/digital_card/draw_activity.api`
  - 如需最小状态透出，在此增补 `mintStatus` 等字段。
- `flutter-mall/lib/model/digital_card/draw_activity_model.dart`
  - 同步最小链上状态字段；不扩展完整资产中心页面。

### 明确反模式（命中任一项都不算完成）

1. **在 `ParticipateDraw`、front-api 或 Flutter 页面里直接调用蚂蚁链 SDK / HTTP。**
2. **同一 `sms_card_instance` 重复创建多条有效发放任务。**
3. **MQ 发布成功后就把任务视为绝对完成排队，没有 DB 任务和 job 兜底。**
4. **consumer 在本地未落库前就 ACK 消息。**
5. **把 `asset_status`、`mint_status`、`chain_status` 合并成一个字段图省事。**
6. **链上请求超时后直接再发第二次，而不依赖稳定幂等键或查询回执。**
7. **成功发链但回写失败时，简单重发 token 而不是先做幂等恢复。**
8. **把 token 标识、失败回执或人工干预入口暴露给未授权主体。**
9. **为了做后台工作台而绕开现有 `oms/chain-monitor` 的治理模式。**
10. **顺手把完整资产中心、审计检索、活动编排都塞进本 Story。**
11. **新增 `consumer/internal/mq/digital_card/*.go` 但没有在 `consumer/internal/svc/service_context.go` 启动订阅。**
12. **新增 `job/internal/jobs/*.go` 但没有在 `job/internal/logic/job_logic.go` 接上任务名分发。**
13. **新增 `cardminttaskservice` / gRPC server 文件但没有在 `rpc/sms/sms.go` 注册服务。**

### Testing Requirements

1. 必须证明同一资产实例不会创建第二条有效发放任务。
2. 必须证明同一任务被重复消费或重复重试时，不会重复发 token。
3. 必须证明超时、失败、回执写回异常都会进入明确的补偿或人工状态。
4. 必须证明 consumer 手动 ACK 只发生在本地状态稳定写入之后。
5. 必须证明 job 扫描能补到漏投和卡住的任务，而不会制造重复任务。
6. 必须证明后台干预接口遵守平台/租户/商户作用域隔离。
7. 如补充前台状态字段，必须证明客户端不会把“补偿中/人工复核中”误渲染为“失败/已到账”。

### Latest Technical Information（2026-04-17 调研）

1. **go-zero 官方 GitHub Releases 当前最新版本为 `v1.10.1`，发布时间是 2026-03-28；仓库当前仍固定在 `go-zero 1.9.3`。**  
   10.4 应继续基于现有 go-zero 链路做增量实现，不把发放任务和链上接入与框架升级绑定。  
   参考：[go-zero Releases](https://github.com/zeromicro/go-zero/releases)
2. **GORM 官方事务文档仍明确要求：在 `db.Transaction(func(tx *gorm.DB) error { ... })` 中返回错误会整体回滚，并支持嵌套事务。**  
   这直接支撑 10.4 的“任务创建 / 资产快照更新 / 状态日志”必须显式放进事务。  
   参考：[GORM Transactions](https://gorm.io/docs/transactions.html)
3. **GORM 官方索引文档在 2026-03-30 仍强调：唯一复合索引需要显式定义，且复合索引列顺序会影响性能。**  
   10.4 的任务幂等唯一键、token 唯一键和扫描索引必须在 DDL 层明确设计，不能只靠应用逻辑约定。  
   参考：[GORM Database Indexes](https://gorm.io/docs/indexes.html)
4. **RabbitMQ 官方数据安全文档明确指出：publisher confirms 与 consumer acknowledgements 分别覆盖“生产者到 broker”与“broker 到消费者”两段数据安全，二者彼此独立。**  
   对当前仓库意味着：即便 consumer 用了手动 ACK，如果生产端没有 confirm 保证，也仍然需要本地任务表作为真正的可恢复真相源。  
   参考：[RabbitMQ Consumer Acknowledgements and Publisher Confirms](https://www.rabbitmq.com/docs/3.13/confirms)
5. **RabbitMQ 官方文档同时说明：手动 ACK 模式下，未 ACK 的消息会在连接或 channel 关闭时自动重新入队。**  
   这正适合 10.4 的 consumer 在链上调用失败或本地回写失败时走 NACK / 重入队，但执行器必须本身幂等。  
   参考：[RabbitMQ Consumer Acknowledgements and Publisher Confirms](https://www.rabbitmq.com/docs/4.0/confirms)
6. **蚂蚁数字科技开放联盟链官方产品页仍把接入过程定义为“三步上链”：开通体验或购买、业务和合约开发、接入业务端上链，并明确支持集成 SDK、合约模板、分布式数字身份、实人认证和内容安全能力。**  
   这说明 10.4 应优先走“统一集成层 + 既有实名/合规体系协同”的路线，而不是为单个页面写临时链上脚本。  
   参考：[蚂蚁链开放联盟链](https://www.antdigital.com/products/openchain)

### Project Structure Notes

- 当前真实仓库已经有独立的 `consumer/`、`job/` 和 `pkg/mq/`，因此 10.4 不需要假设“异步链路只能写在 sms RPC 内部”。
- `pkg/mq/rabbitmq_util.go` 已经提供：
  - `SendMessage`
  - `ConsumeSimple`
  - `ConsumeSimpleWithAck`
  但没有 publisher confirm 封装，所以 10.4 的消息可靠性策略必须建立在 DB 任务表 + job 扫描之上。
- `consumer/internal/svc/service_context.go` 当前通过硬编码 goroutine 启动各类 MQ 订阅，不会自动发现新 consumer；10.4 的提货卡发放订阅必须显式接到这里。
- `job` 当前通过 `job-api /from/:name` + `job/internal/logic/job_logic.go` 的 `switch req.Name` 分发执行，不会自动执行新增 job 文件；10.4 的补偿任务名必须手工接入。
- `rpc/sms/sms.go` 当前手工注册各个 gRPC service；若新增 `CardMintTaskService`，必须同步注册才能对外提供能力。
- 当前 `web-admin` 已经有 `oms/chain-monitor` 页面与 admin 动作接口，这为 10.4 的提货卡发放工作台提供了现成模式。
- 当前 `web-admin/src/pages/sms/DigitalCardActivity/` 只负责活动配置，不负责发卡链路监控；10.4 应新增链路工作台，而不是把补偿动作塞进活动配置页。
- 当前 `flutter-mall` 和 front 聚合已经能展示“资产已创建，链上处理中”，但还没有 token 成功、补偿中、人工复核、冻结这些明确状态。

### References

- `/_opcos/planning-artifacts/1-new-feature/epic.md#Story 10.4`
- `/_opcos/planning-artifacts/1-new-feature/prd.md#Journey 10`
- `/_opcos/planning-artifacts/1-new-feature/prd.md#提货卡抽赏与链上资产`
- `/_opcos/planning-artifacts/1-new-feature/architecture.md#Digital Card Asset Modeling`
- `/_opcos/planning-artifacts/1-new-feature/architecture.md#API & Communication Patterns`
- `/_opcos/planning-artifacts/1-new-feature/architecture.md#Infrastructure & Deployment`
- `/_opcos/planning-artifacts/1-new-feature/architecture.md#Data Flow`
- `/_opcos/planning-artifacts/1-new-feature/ux-design.md#提货卡抽赏与链上资产到账流`
- `/_opcos/planning-artifacts/1-new-feature/ux-design.md#Digital Asset Status`
- `/_opcos/planning-artifacts/1-new-feature/ux-design.md#Ops Audit Flow`
- `/_opcos/project-context.md`
- `/_opcos/implementation-artifacts/10-3-卡片实例生成-唯一编号与本地资产台账.md`
- `/script/sql/sms/migration_20260417_card_asset_ledger.sql`
- `/rpc/sms/internal/logic/cardassetservice/card_asset_helper.go`
- `/rpc/sms/internal/logic/cardassetservice/ensure_card_instance_by_participation_record_logic.go`
- `/rpc/sms/internal/svc/service_context.go`
- `/pkg/mq/rabbitmq_util.go`
- `/consumer/internal/svc/service_context.go`
- `/job/internal/logic/job_logic.go`
- `/rpc/sms/sms.go`
- `/web-admin/src/pages/oms/chain-monitor/service.ts`
- `/api/admin/internal/types/chain_monitor.go`
- `/api/admin/internal/types/chain_intervention.go`
- [go-zero Releases](https://github.com/zeromicro/go-zero/releases)
- [GORM Transactions](https://gorm.io/docs/transactions.html)
- [GORM Database Indexes](https://gorm.io/docs/indexes.html)
- [RabbitMQ Consumer Acknowledgements and Publisher Confirms](https://www.rabbitmq.com/docs/3.13/confirms)
- [RabbitMQ Consumer Acknowledgements and Publisher Confirms](https://www.rabbitmq.com/docs/4.0/confirms)
- [蚂蚁链开放联盟链](https://www.antdigital.com/products/openchain)

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- create-story workflow analysis
- sprint-status selection: `10-4-蚂蚁链-token-发放-回执写回与失败补偿`
- validation note: `_bmad/core/tasks/validate-workflow.xml` 在仓库中缺失，已按 `create-story/checklist.md` 进行手工交叉校验
- context sources: Epic 10 / PRD / 架构 / UX / project-context / Story 10.3 / 现有 `consumer`、`job`、`pkg/mq`、`oms/chain-monitor`
- web research: go-zero releases、GORM transactions/indexes、RabbitMQ confirms/acks、蚂蚁链开放联盟链官方产品页
- implementation verification:
  - `go test ./pkg/digitalcardmint ./rpc/sms/internal/logic/... ./consumer/internal/... ./job/internal/... ./api/admin/internal/handler ./api/admin/internal/handler/sms/digital_card_chain ./api/admin/internal/logic/sms/digital_card_chain ./api/admin/internal/svc`
  - `go test ./consumer/internal/mq/digital_card ./job/internal/jobs ./api/admin/internal/logic/sms/digital_card_chain`
  - `cd web-admin && npm test -- DigitalCardChainMonitor`
  - `cd web-admin && npm run tsc -- --pretty false`（失败，均为仓库既有页面类型问题，当前新增页面相关报错已清零）

### Completion Notes List

- 已新增 `pkg/antchain` 与 `pkg/digitalcardmint`，落地发放任务表模型、幂等键、状态机、回执写回、失败补偿与可用动作判定。
- 已把链路接入 `rpc/sms -> consumer -> job -> admin/web-admin`：中奖/回填会确保任务存在并事务外派发，consumer 执行发链，job 在无 MQ 场景下可直接复用执行器做兜底补偿。
- 已新增提货卡链路后台工作台：`extra_routes`、API 文档、admin handler/logic/types、`web-admin/src/pages/sms/DigitalCardChainMonitor/` 全链路打通。
- 已补充自动化测试：`pkg/digitalcardmint`、`consumer/internal/mq/digital_card`、`job/internal/jobs`、`api/admin/internal/logic/sms/digital_card_chain`、`web-admin` helper 测试均通过。
- 已更正 6.1 状态：当前 `smsclient.DrawMemberRecordData` 仍未承载 `mintStatus` / `mintStatusText` / `chainStatus`，front/Flutter 只能继续透传服务端确认后的 `assetStatusText`，该项保持未完成。
- 受仓库 AGENTS 约束，本次未新增 `rpc/sms/proto/card_mint_task.proto` 或生成 gRPC service，而是以手工共享服务 `pkg/digitalcardmint` 收口执行边界。
- `web-admin` 全量 `tsc` 仍被仓库既有页面（`oms/customerServiceWorkstation`、`oms/merchantWorkstation`、`system/*` 等）阻塞；当前新增 `DigitalCardChainMonitor` 相关类型错误已修复。

### File List

- `_opcos/implementation-artifacts/10-4-蚂蚁链-token-发放-回执写回与失败补偿.md`
- `_opcos/implementation-artifacts/sprint-status.yaml`
- `script/sql/sms/migration_20260418_card_mint_task.sql`
- `pkg/antchain/client.go`
- `pkg/antchain/mock.go`
- `pkg/antchain/types.go`
- `pkg/digitalcardmint/constants.go`
- `pkg/digitalcardmint/model.go`
- `pkg/digitalcardmint/service.go`
- `pkg/digitalcardmint/service_test.go`
- `rpc/sms/etc/sms.yaml`
- `rpc/sms/internal/config/config.go`
- `rpc/sms/internal/svc/service_context.go`
- `rpc/sms/internal/logic/cardminttaskservice/card_mint_task_helper.go`
- `rpc/sms/internal/logic/cardassetservice/ensure_card_instance_by_participation_record_logic.go`
- `rpc/sms/internal/logic/cardassetservice/backfill_winning_card_instances_logic.go`
- `rpc/sms/internal/logic/cardassetservice/card_asset_helper.go`
- `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go`
- `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go`
- `rpc/sms/internal/logic/drawparticipationservice/draw_participation_helper.go`
- `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go`
- `consumer/etc/consumer-api.yaml`
- `consumer/internal/config/config.go`
- `consumer/internal/svc/service_context.go`
- `consumer/internal/mq/digital_card/mint_requested.go`
- `consumer/internal/mq/digital_card/mint_requested_test.go`
- `job/etc/job-api.yaml`
- `job/internal/config/config.go`
- `job/internal/svc/service_context.go`
- `job/internal/logic/job_logic.go`
- `job/internal/jobs/handle_card_mint_timeout_logic.go`
- `job/internal/jobs/handle_card_mint_timeout_logic_test.go`
- `api/admin/etc/admin-api.yaml`
- `api/admin/internal/config/config.go`
- `api/admin/internal/svc/servicecontext.go`
- `api/admin/internal/types/digital_card_chain.go`
- `api/admin/internal/handler/extra_routes.go`
- `api/admin/internal/handler/sms/digital_card_chain/*`
- `api/admin/internal/logic/sms/digital_card_chain/*`
- `api/admin/doc/api/sms/digital_card_chain.api`
- `api/admin/doc/api/sms/sms.api`
- `web-admin/config/routes.ts`
- `web-admin/src/pages/sms/DigitalCardChainMonitor/*`

## Change Log

- 2026-04-17: 创建 Story 10.4，补全统一链上集成层、发放任务表、MQ+Job 补偿、后台干预工作台与幂等回执写回约束。
- 2026-04-18: 完成 10.4 主体实现，新增发放任务 DDL、统一链上集成层、共享发放执行器、consumer/job 兜底、admin/API 文档与 `web-admin` 提货卡链路工作台，并补充核心自动化测试；同时更正 6.1 未完成状态与 File List 漏项。
