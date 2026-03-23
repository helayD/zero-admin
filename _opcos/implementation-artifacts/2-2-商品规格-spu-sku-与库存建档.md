# Story 2.2: 商品规格、SPU/SKU 与库存建档

Status: in-progress

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 商户管理员,
I want 完成商品规格、SPU、SKU、价格和库存的建档闭环,
so that 我可以把自有商品准备成具备进入后续审核流程条件的可售商品草稿。

## Acceptance Criteria

1. **Given** 商户管理员已完成分类、品牌和属性配置  
   **When** 创建或编辑商品 SPU、SKU、规格、价格和库存信息  
   **Then** 系统保存完整的商品建档数据并绑定当前商户作用域  
   **And** 生成的商品草稿至少包含后续审核所需的最小关键信息：有效 SPU 基础信息、至少一个有效 SKU、合法规格组合、合法价格、合法库存与正确主体归属。
2. **Given** 商品规格、价格、库存、目录引用或商户归属信息不合法  
   **When** 商户管理员尝试保存  
   **Then** 系统阻止保存并明确指出冲突原因  
   **And** 不允许生成缺少库存、缺少有效 SKU、目录引用失效或归属错误的可售商品记录。

## Tasks / Subtasks

- [ ] 1. 收口商品建档契约，统一 SPU/SKU/库存/规格的 API、RPC 与生成链源定义（AC: 1, 2）
  - [ ] 盘点 `api/admin/doc/api/pms/` 下商品建档相关 `.api`，重点覆盖 `product.api`、`sku_stock.api`、`product_operate_log.api`、图片/属性/阶梯价/满减等与建档直接相关的契约，确认列表、详情、保存、更新、上下架前校验、草稿保存都落在同一套接口语义里
  - [x] 盘点 `rpc/pms/proto/` 中 SPU、SKU、库存、商品属性值、规格值、图片与价格相关 proto，避免继续沿用旧单商户字段命名或页面私有 DTO
  - [ ] 对请求参数补齐分页默认值、作用域字段、状态字段和必要校验字段，保证 admin-api 与 rpc/pms 之间的映射在 logic 中显式完成，而不是依赖隐式默认或手改生成物
  - [x] 严守生成链：只改 `.api`、`.proto`、SQL 源和手工 logic；不直接修改 `types.go`、`routes.go`、`*_pb.go`、`client`、`gen/query`、`gen/model` 等生成产物

- [ ] 2. 以商户作用域为真相源建立 SPU/SKU/库存数据模型与约束（AC: 1, 2）
  - [x] 核查 `script/sql/pms/` 中 `pms_product`、`pms_sku_stock` 及相关商品属性/图片/会员价/阶梯价/满减等表结构，确认是否已具备 `platform_id/tenant_id/merchant_id`、审计字段、状态字段、版本/更新时间字段以及高频查询索引
  - [x] 对缺失的商户归属字段、联合唯一约束和库存/编码检索索引补 migration，保证同一商户可维护自己的 SPU/SKU，不污染其他商户数据，也不因共享编码造成冲突
  - [x] 设计最小历史数据回填策略，避免作用域字段上线后把既有商品建档数据全部打成“无主体”或默认平台全局数据
  - [ ] 明确商品草稿、待审核、上架、下架等状态边界；2.2 只负责建档与可进入审核的草稿准备，不把上架审核流提前塞进本 story

- [ ] 3. 复用 1.6A / 1.6B 的治理底座，打通 admin-api -> rpc/pms 的作用域感知读写闭环（AC: 1, 2）
  - [ ] 复用 Story 1.6A 已建立的 query scope helper，让商品列表、详情、SKU 列表、库存查询自动按当前平台 / 租户 / 商户上下文过滤，而不是页面手工拼接主体 ID
  - [ ] 复用 Story 1.6B 已建立的 write scope guard，对创建、编辑、删除、复制、状态切换、批量导入等写路径执行资源归属校验与越权审计
  - [ ] 在 `api/admin/internal/logic/pms/` 与 `rpc/pms/internal/logic/` 中显式校验商户管理员不能通过手填 tenantId/merchantId 越界写入他方商品、库存或 SKU
  - [ ] 对批量保存 SKU、批量更新库存、删除规格值、复制商品等“隐蔽写操作”应用同一套治理逻辑，避免只修主保存接口而留下旁路

- [ ] 4. 建立可复用的商品建档聚合模型，覆盖分类/品牌/属性与规格/库存的完整关系（AC: 1）
  - [ ] 基于 Story 2.1 已完成的分类、品牌、属性、属性分组、规格基础数据，明确 SPU/SKU 建档时的依赖关系与最小必填集合，避免再次创建平行目录语义
  - [ ] 统一 SPU 层与 SKU 层字段：商品基础信息、主图/图集、类目、品牌、销售属性、规格组合、售价、库存、安全库存、上下文归属、展示状态等，保证前后台和后续 2.3 审核链路使用同一套数据口径
  - [ ] 明确哪些字段由 SPU 继承到 SKU、哪些字段允许 SKU 覆写，避免后续详情页、库存页和审核页对同一商品出现多套解释
  - [ ] 对被 2.1 基础目录引用的禁用/删除数据建立前置校验：分类、品牌、属性已失效时，商品建档页必须给出可理解阻断原因，而不是保存后才失败

- [ ] 5. 完成库存、价格与规格组合校验，防止生成缺少库存或归属错误的可售记录（AC: 2）
  - [x] 在保存前校验 SKU 规格组合唯一性、SKU 编码唯一性、价格合法性、库存非负、安全库存边界、必填销售属性完整性与商户归属一致性
  - [x] 对“有 SPU 无 SKU”“有价格无库存”“有库存无规格主键”“商户主体与分类/品牌归属不一致”等场景返回结构化错误，并在表单层定位到对应字段或行
  - [x] 对草稿商品进入 2.3 上架审核前定义最小可发布条件，例如：至少一个有效 SKU、主图/基础信息完整、价格库存合法、目录引用有效、主体归属合法
  - [ ] 如涉及库存批量维护或规格矩阵编辑，优先复用现有 PMS 模型与页面交互，不新增平行库存中心或手工导入依赖

- [ ] 6. 交付商户可用的商品建档后台体验，并保持与现有 Web Admin 结构一致（AC: 1, 2）
  - [ ] 在 `web-admin/src/pages/pms/` 下核查并修正商品/SPU/SKU/库存相关页面，保持 `index.tsx + service.ts + data.d.ts + components/*` 与 `PageContainer + ProTable + Drawer/Modal` 模式
  - [ ] 页面持续显示当前主体上下文，符合 `Scope Context Bar` 语义，让平台、租户、商户在建档时一眼知道自己正在操作谁的数据
  - [x] 建档表单要优先服务“核对流程”而不是“填空流程”：默认带出 2.1 的目录基础数据，规格矩阵、价格和库存采用分段编辑与即时校验，降低商户出错成本
  - [ ] 对新增、编辑、复制、删除、批量库存调整等高风险动作提供 `Consequence Preview` 与 `Recovery-first Feedback`；被引用、越权、库存非法、价格冲突等问题都要给出可修复提示

- [ ] 7. 为后续 2.3 审核与 2.5 商品详情提供可复用的基础字段（AC: 1）
  - [ ] 确保商品主数据中包含后续审核和前台展示所需的基础字段：主图、图集、卖点、品牌、类目、规格摘要、价格区间、库存摘要、商品状态、商户归属与作用域信息
  - [ ] 明确“草稿商品”“可送审商品”“审核失败商品”“已下架商品”等状态在建档层的最小表达，但不在本 story 内完成完整审核编排或前台详情展示逻辑
  - [ ] 对商品详情页未来需要复用的规格组合、库存可售标记、价格摘要与优惠前基础价格保持稳定输出口径，但不额外扩展新的商品详情聚合模型
  - [ ] 沿用现有日志 / 审计机制记录关键建档动作，满足后续审核与治理追踪所需的最小可追溯性，不在本 story 内新增独立审计产品能力

- [ ] 8. 补齐测试、回归与跨 story 兼容验证（AC: 1, 2）
  - [x] 为 `rpc/pms` 增加 SPU/SKU/库存读写测试，覆盖同主体成功、跨主体拒绝、详情越权拒绝、规格组合冲突、价格/库存非法、草稿最小发布条件、批量库存更新一致性
  - [x] 为 admin-api 侧增加 scope 透传与错误映射测试，确认平台 / 租户 / 商户上下文下商品建档行为一致，且错误能被页面正确消费
  - [ ] 回归 Story 1.4 的角色 / 菜单模板、1.5 的商户主体与启停、1.6A 的查询隔离、1.6B 的写路径越权校验，以及 Story 2.1 的目录基础数据可复用性，确保 2.2 不破坏已有治理底座
  - [x] 对 2.3 商品上架审核与作用域可见性做最小联调验证，确认本 story 产出的草稿商品确实可进入后续审核流，而不是形成新的接口缺口

## Dev Notes

### Scope Guardrails

- 本 story 只负责商品规格、SPU/SKU、价格和库存建档闭环，以及“商品草稿具备进入发布审核流程最小条件”这一层，不直接实现商品审核、上架/下架可见性控制、首页导购或商品详情前台消费闭环；这些分别属于 Story 2.3、2.4、2.5、2.6 的范围。[Source: `_opcos/planning-artifacts/1-new-feature/epic.md#Epic-2-商品目录发布与前台导购可见性`; `_opcos/planning-artifacts/1-new-feature/epic.md#Story-22-商品规格SPUSKU-与库存建档`; `_opcos/planning-artifacts/1-new-feature/epic.md#Story-23-商品上架审核与作用域可见性控制`; `_opcos/planning-artifacts/1-new-feature/epic.md#Story-25-商品详情与规格库存优惠摘要`]
- 本 story 必须建立在 Epic 1 已完成的治理底座之上，特别是 Story 1.5 的商户主体启停、1.6A 的查询范围过滤与主体透传、1.6B 的关键写操作校验与越权审计；不能回退到旧单商户模式，也不能另造一套商品建档专用 scope 逻辑。[Source: `_opcos/implementation-artifacts/1-5-商户入驻审核与主体启停.md`; `_opcos/implementation-artifacts/1-6a-核心业务对象查询范围过滤与主体透传.md`; `_opcos/implementation-artifacts/1-6b-核心业务对象关键写操作校验与越权审计.md`]
- 本 story 必须直接消费 Story 2.1 已建立的分类、品牌、属性、属性分组、规格等目录基础语义，不允许重新发明平行商品目录体系。[Source: `/Users/helay/Documents/GitHub/zero-admin/_opcos/implementation-artifacts/2-1-商品分类-品牌与属性基础配置.md`; `_opcos/planning-artifacts/1-new-feature/epic.md#Story-21-商品分类品牌与属性基础配置`]

### Story Foundation / Business Context

- Epic 2 的目标是让商户和运营在正确作用域内发布商品与导购内容，让消费者通过首页、分类、品牌、专题与详情页发现可信、可售的商品。2.2 是把“目录基础语义”推进成“可销售商品草稿”的关键一步，它把 2.1 的分类/品牌/属性配置转化为 SPU/SKU/库存真相源。[Source: `_opcos/planning-artifacts/1-new-feature/epic.md#Epic-2-商品目录发布与前台导购可见性`]
- Story 2.2 的 AC 明确要求：在商户管理员已经完成分类、品牌和属性配置后，创建或编辑 SPU、SKU、规格、价格和库存时，系统要保存完整商品建档数据并绑定当前商户作用域；当规格、价格、库存、目录引用或商户归属不合法时，必须阻止保存并明确指出冲突原因。同时，本 story 仅要求商品草稿满足进入后续审核流程的最小条件，而不是在本 story 内完成完整审核规则编排。 [Source: `_opcos/planning-artifacts/1-new-feature/epic.md#Story-22-商品规格SPUSKU-与库存建档`]
- PRD 中 FR3 要求消费者可以查看商品详情，包括价格、规格、库存、卖点、品牌和图文详情；FR49 要求商户管理员只能在授权范围内管理本商户的商品、库存、订单、营销、内容和售后。这两个需求在 2.2 上汇合成同一件事：商品建档既要准备前台详情所需的完整商品语义，也要把商户归属和作用域隔离落实到数据真相源。[Source: `_opcos/planning-artifacts/1-new-feature/prd.md#商品发现与内容导购`; `_opcos/planning-artifacts/1-new-feature/prd.md#平台租户与多商户治理`]
- UX 明确把“商品详情、规格切换、金额确认、库存表达与支付前确认”视为消费者建立信任的关键节点，因此 2.2 不只是后台录入页，而是后续 2.5 前台详情可信体验的源头。[Source: `_opcos/planning-artifacts/1-new-feature/ux-design.md#Critical-Success-Moments`; `_opcos/planning-artifacts/1-new-feature/ux-design.md#Trusted-Confirmation-Flow`]

### Existing Repository Reality

- 当前仓库的 PMS 能力并非空白：`api/admin/doc/api/pms/` 下已有 `product.api`、`sku_stock.api`、`product_attribute_value.api`、`product_full_reduction.api`、`product_ladder.api`、`product_operate_log.api` 等契约；`web-admin/src/pages/pms/` 下也已有商品、SKU、库存与属性相关页面骨架。这意味着 2.2 的重点是把这些壳与治理底座、建档规则和后续可复用输出真正接通，而不是从零新开模块。[Source: `api/admin/doc/api/pms/`; `web-admin/src/pages/pms/`; `_opcos/project-context.md`]
- 既有菜单种子和前端路由已经为商品管理提供入口，因此 2.2 应复用现有导航结构与页面骨架，在既有页面上补 scope 感知、规格矩阵校验、价格/库存验证和错误反馈，而不是再造新导航树。[Source: `script/sql/sys/sys_menu.sql`; `web-admin/config/routes.ts`]
- 当前 PMS 商品建档相关表与 SQL 资产需要重点核查：
  - 商品主表：`script/sql/pms/pms_product.sql`
  - SKU/库存：`script/sql/pms/pms_sku_stock.sql`
  - 商品属性值：`script/sql/pms/pms_product_attribute_value.sql`
  - 商品阶梯价 / 满减 / 会员价 / 操作日志等相关表
  这些表是否已具备 `platform_id/tenant_id/merchant_id`、审计字段、状态字段、索引和历史数据可回填策略，是 2.2 的现实基础。[Source: `script/sql/pms/`; `_opcos/implementation-artifacts/1-6a-核心业务对象查询范围过滤与主体透传.md`]
- Story 2.1 已经把分类、品牌、属性、属性分组、规格这一层整理成“商品目录基础语义层”；2.2 现在应直接复用 2.1 产出的字段、页面和 scope 约束，不再回头重做目录基础配置。[Source: `/Users/helay/Documents/GitHub/zero-admin/_opcos/implementation-artifacts/2-1-商品分类-品牌与属性基础配置.md`]

### Technical Requirements

- 商品建档必须与现有生成链兼容：新增请求字段、商品状态、规格矩阵结构、库存字段或返回结构时，必须修改 `.api`、`.proto`、SQL 源并重新生成；不手改 `types.go`、`routes.go`、`pb.go`、client 或 `gen/query` 生成物。[Source: `_opcos/project-context.md`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Migration--Schema-Evolution`]
- SPU/SKU/库存数据必须以 MySQL 为真相源；Redis 仅承担缓存、热点读和协调，不能让库存可售状态只存在于缓存层。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Data-Architecture`]
- 商品建档必须显式记录 `status`、`version`、`updated_at`、`updated_by` 等关键状态流转字段，并对高风险写操作（如批量库存更新、商品复制、重复提交）预留幂等保护与审计入口。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Data-Architecture`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Authentication--Security`]
- 规格组合、SKU 编码、价格、库存、安全库存、商户归属和目录引用都要做保存前校验；不能先生成“看起来可售”的商品记录，再把异常丢给 2.3 审核或前台详情页兜底。[Inference from AC and trusted confirmation flow; Source: `_opcos/planning-artifacts/1-new-feature/epic.md#Story-22-商品规格SPUSKU-与库存建档`; `_opcos/planning-artifacts/1-new-feature/ux-design.md#Trusted-Confirmation-Flow`]
- 后台服务层必须继续走 Umi `request`，Go 侧必须继续遵循 `handler -> logic -> svc`；商品建档页和逻辑层不能引入裸请求或把业务规则塞进 handler。[Source: `_opcos/project-context.md`]

### Architecture Compliance

- 继续遵循 `web-admin -> admin-api -> rpc/pms -> MySQL` 的边界，不新增商品建档专用微服务、前端直连 DB/ES 的捷径，或把商品详情聚合逻辑散到页面层。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Project-Structure--Boundaries`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Service-Boundaries`]
- 商品建档虽属 PMS 域，但必须继承统一平台 / 租户 / 商户作用域模型、JWT claim、scope middleware、audit context 和幂等约束；商户管理员不能通过手填主体 ID 越权维护他方商品。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Authentication--Security`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Decision-Priority-Analysis`]
- Web 端继续保持 `index.tsx + service.ts + data.d.ts + components/*` 组织模式与 `PageContainer + ProTable + Drawer/Modal` 结构；商城与后台共用同一套商品/库存/状态语义，不新造第二套命名体系。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Frontend-Architecture`; `_opcos/project-context.md`]
- 2.2 产出的商品聚合结构应服务于后续 2.3 审核、2.4 导购浏览、2.5 商品详情消费，不要只满足当前后台表单保存而忽略 downstream 复用。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Requirements-to-Structure-Mapping`; `_opcos/planning-artifacts/1-new-feature/epic.md#Story-23-商品上架审核与作用域可见性控制`; `_opcos/planning-artifacts/1-new-feature/epic.md#Story-25-商品详情与规格库存优惠摘要`]

### Library / Framework Requirements

- 继续沿用项目当前基线：Go `1.25`、go-zero `1.9.3`、gRPC `1.77.0`、GORM `1.31.1`、React `17.0.0`、Umi `3.5.0`、Ant Design `4.17.0`、TypeScript `4.5.0`；本 story 不引入框架升级或新的页面组织范式。[Source: `_opcos/project-context.md`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Selected-Starter-Existing-Repository-Baseline`]
- 架构文档已完成 2026-03-20 版本核验：上游 Go、go-zero、React、Ant Design、Flutter 均已有更高版本，但当前项目决策是“业务闭环优先、升级后置”，因此 2.2 仅做实现，不借机发起大版本迁移。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Version-Verification-Note-2026-03-20`]

### File Structure Requirements

- 预计主要触达的手工入口包括：
  - `api/admin/doc/api/pms/`
  - `api/admin/internal/logic/pms/`
  - `rpc/pms/proto/`
  - `rpc/pms/internal/logic/`
  - `script/sql/pms/`
  - `web-admin/src/pages/pms/`
- 重点可能涉及的既有页面 / 模块包括：
  - 商品管理页（SPU）
  - SKU/库存维护页或商品编辑内的规格矩阵区域
  - 商品属性值、图片、价格阶梯/满减相关编辑模块
  - 后续 2.3 审核、2.5 详情页会消费的商品详情输出逻辑
- 不要手改生成物或构建产物：
  - `api/admin/internal/types/types.go`
  - `api/admin/internal/handler/routes.go`
  - `rpc/pms/*pb.go`
  - `rpc/pms/client/*`
  - `rpc/pms/gen/query/*`
  - `rpc/pms/gen/model/*`
  - `web-admin/dist/**/*`
  - `target/**/*`
- 若需要新增共享 helper，优先放在 `pkg/scope/`、`pkg/audit/`、`api/admin/internal/common/` 或 `rpc/pms/internal/logic/common/`，不要在商品、库存、SKU 多个目录里复制一份校验逻辑。[Source: `_opcos/project-context.md`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Project-Structure--Boundaries`]

### Web Admin / UX Requirements

- 商品建档页要优先服务“核对流程”而不是“填空流程”：默认带出 Story 2.1 的目录基础数据，按“基础信息 → 规格矩阵 → 价格库存 → 图文与说明 → 状态摘要”分段组织，减少商户来回跳页确认。[Source: `_opcos/planning-artifacts/1-new-feature/ux-design.md#Flow-Optimization-Principles`; `_opcos/planning-artifacts/1-new-feature/ux-design.md#Effortless-Interactions`]
- 页面需要持续显示当前主体上下文，符合 `Scope Context Bar` 语义，让平台 / 租户 / 商户视角切换清晰可见，避免商户误以为自己在改全局商品数据。[Source: `_opcos/planning-artifacts/1-new-feature/ux-design.md#Scope-Context-Bar`]
- 规格矩阵、库存、价格和错误提示要遵循 `Recovery-first Feedback`：例如某个 SKU 规格重复、库存非法、价格为空、目录引用失效时，应就地高亮并给出修复入口，而不是只弹模糊 toast。[Source: `_opcos/planning-artifacts/1-new-feature/ux-design.md#Feedback-Patterns`; `_opcos/planning-artifacts/1-new-feature/ux-design.md#Recovery-first-Feedback`]
- 商品建档结果、草稿状态、可送审条件和后续下一步（送审、继续编辑、查看详情）要在保存后给出明确摘要，帮助商户建立“我已准备好进入 2.3 审核”的认知。[Source: `_opcos/planning-artifacts/1-new-feature/ux-design.md#Success-Criteria`; `_opcos/planning-artifacts/1-new-feature/ux-design.md#Completion`]

### Previous Story Intelligence

- Story 2.1 已经把分类、品牌、属性、属性分组、规格基础配置整理成可复用的目录语义层，并明确要求这些基础数据为 2.2 的 SPU/SKU 建档直接服务。因此 2.2 不应重复修目录，而是直接消费 2.1 的成果，并验证其可用性。[Source: `/Users/helay/Documents/GitHub/zero-admin/_opcos/implementation-artifacts/2-1-商品分类-品牌与属性基础配置.md`]
- Story 1.5 让商户主体与启停状态成为可信真相源，2.2 现在可以直接依赖 merchant scope 绑定商品、库存和价格数据，而不需要再猜测商户上下文。[Source: `_opcos/implementation-artifacts/1-5-商户入驻审核与主体启停.md`]
- Story 1.6A 已经把查询范围过滤与主体透传打通到核心业务对象，2.2 应直接复用其 query scope helper 来做商品、SKU、库存列表/详情过滤。[Source: `_opcos/implementation-artifacts/1-6a-核心业务对象查询范围过滤与主体透传.md`]
- Story 1.6B 已经把写路径越权校验与审计闭环打通，因此 2.2 的商品创建、编辑、复制、库存变更、状态切换都要沿用同一套 write guard，而不是另写简化逻辑。[Source: `_opcos/implementation-artifacts/1-6b-核心业务对象关键写操作校验与越权审计.md`]

### Git Intelligence Summary

- 最近提交集中在首页聚合 fallback、重复 job scaffold 清理等修复，说明仓库当前已从 Epic 1 治理建设切换到收口稳定性和新 Epic 开发准备阶段；2.2 需要在不引入额外脚手架噪音的前提下推进商品建档主线。[Source: `git log --oneline -5`]
- `sprint-status.yaml` 中 Epic 2 当前为 `in-progress`，2-1 已先落为实施文档，2-2 是当前第一条 backlog story，因此按 sprint 顺序自动选中 2.2 符合 create-story 工作流要求。[Source: `_opcos/implementation-artifacts/sprint-status.yaml`; `/Users/helay/Documents/GitHub/zero-admin/_opcos/implementation-artifacts/2-1-商品分类-品牌与属性基础配置.md`]

### Latest Tech Information

- 架构文档已经完成关键技术栈的最新稳定版核验：Go 官方稳定版 `1.26.1`、go-zero `v1.10.0`、React `19.2/19.2.1`、Ant Design `6.3.3`、Flutter 文档稳定通道 `3.41.2`。当前项目明确决策是不在 Epic 2 中做框架升级，避免把商品建档需求与跨大版本迁移耦合。[Source: `_opcos/planning-artifacts/1-new-feature/architecture.md#Version-Verification-Note-2026-03-20`]
- 因此本 story 的“最新技术信息”结论很明确：实现时要遵守当前仓库锁定版本和现有生成链，不引入依赖升级、UI 框架升级或生成器升级作为“顺手优化”。[Source: `_opcos/project-context.md`; `_opcos/planning-artifacts/1-new-feature/architecture.md#Selected-Starter-Existing-Repository-Baseline`]

### Testing Requirements

- `rpc/pms` 至少覆盖：
  - SPU/SKU/库存的同主体成功、跨主体拒绝、详情越权拒绝
  - 规格组合重复、SKU 编码冲突、价格非法、库存非法、安全库存边界错误
  - 草稿最小发布条件校验
  - 批量库存更新 / 批量 SKU 保存一致性
- `admin-api` 至少覆盖：
  - scope 透传
  - 错误信息映射
  - 商品建档相关列表 / 详情 / 写操作在平台 / 租户 / 商户上下文下的一致行为
- `web-admin` 至少验证：
  - 商品建档表单必填与分段编辑行为
  - 规格矩阵校验与就地错误反馈
  - 价格库存非法时的阻断提示
  - 保存草稿 / 继续编辑 / 删除 / 复制等高风险动作反馈
- 回归重点：
  - Story 1.4 的角色 / 菜单模板 / 数据范围
  - Story 1.5 的商户主体与启停
  - Story 1.6A 的查询范围过滤
  - Story 1.6B 的写前归属校验与越权审计
  - Story 2.1 的目录基础数据可消费性
  - Story 2.3 审核流对 2.2 草稿商品的可接入性

### Project Context Reference

- 项目当前硬约束：用户沟通与文档默认中文；Go 保持 `handler -> logic -> svc`；Web Admin 请求统一走 Umi `request`；Flutter 侧网络统一走 `HttpUtil`；生成文件不得手改。[Source: `_opcos/project-context.md`]
- 对本 story 最关键的项目规则是：优先复用现有 PMS 契约、页面骨架与共享 helper，不用“临时捷径”绕过正式生成链、scope helper 或共享请求层。[Source: `_opcos/project-context.md`]

### References

- `_opcos/planning-artifacts/1-new-feature/epic.md`
- `_opcos/planning-artifacts/1-new-feature/prd.md`
- `_opcos/planning-artifacts/1-new-feature/architecture.md`
- `_opcos/planning-artifacts/1-new-feature/ux-design.md`
- `_opcos/project-context.md`
- `_opcos/implementation-artifacts/1-5-商户入驻审核与主体启停.md`
- `_opcos/implementation-artifacts/1-6a-核心业务对象查询范围过滤与主体透传.md`
- `_opcos/implementation-artifacts/1-6b-核心业务对象关键写操作校验与越权审计.md`
- `_opcos/implementation-artifacts/2-1-商品分类-品牌与属性基础配置.md`
- `api/admin/doc/api/pms/`
- `rpc/pms/proto/`
- `script/sql/pms/`
- `web-admin/src/pages/pms/`

## Dev Agent Record

### Agent Model Used

GPT-5 via BMAD `create-story` workflow

### Debug Log References

- `_bmad/core/tasks/workflow.xml`
- `_bmad/bmm/workflows/4-implementation/create-story/workflow.yaml`
- `_opcos/implementation-artifacts/sprint-status.yaml`
- `_opcos/planning-artifacts/1-new-feature/epic.md`
- `_opcos/planning-artifacts/1-new-feature/architecture.md`
- `_opcos/planning-artifacts/1-new-feature/prd.md`
- `_opcos/planning-artifacts/1-new-feature/ux-design.md`
- `_opcos/project-context.md`
- `_opcos/implementation-artifacts/2-1-商品分类-品牌与属性基础配置.md`
- `git log --oneline -5`

### Completion Notes List

- 已按 sprint-status 自动选择当前第一条 backlog story：`2-2-商品规格-spu-sku-与库存建档`
- 已基于 Epic / PRD / Architecture / UX / project-context / 前置故事 1.5、1.6A、1.6B、2.1 生成完整实施文档，并将状态设为 `ready-for-dev`
- 已将 2-2 明确收口为“商品建档与库存真相源层”，避免与 2.3 的审核可见性、2.4 的导购浏览、2.5 的详情消费范围重叠
- 已把对后续故事的输出契约（审核、详情、导购）显式写入 Dev Notes，减少后续 story 再回头补结构的风险
- 已在 `rpc/pms/internal/logic/productspuservice/` 将 SPU 保存链路改为复用 `EnsureSkuCode` / `RefreshSpuDraftSummary`，统一空 SKU 编码生成策略，避免继续使用时间戳随机码并让 SPU 汇总字段以真实落库 SKU 为准
- 已为 `productspuservice` 增加 `productspu_draft_test.go`，覆盖 SPU 新增 / 更新时的 SKU 编码生成、库存汇总与作用域回写；同时为 ES 同步发送增加空 `RabbitMQ` 保护，避免测试与本地最小环境发生空指针
- 已补 `rpc/pms/proto/product_spu.proto`、`rpc/pms/proto/product_sku.proto` 的 2.2 契约源：为 SPU 嵌套的会员价 / 阶梯价 / 满减 / 属性值 / SKU 明细补可回传 `id` 字段，并为库存锁定请求补治理范围字段，给后续生成链收口留出正式契约入口
- 已新增 `script/sql/pms/migration_20260323_product_draft_scope_constraints.sql`，为 `pms_product_spu` / `pms_product_sku` 补联合唯一约束与作用域索引，并为 `pms_product_attribute_value`、`pms_member_price`、`pms_product_ladder`、`pms_product_full_reduction` 补 `platform_id/tenant_id/merchant_id` 和最小历史回填 SQL
- 已将 `ApplyProductScope` 扩展到 SPU 关联的属性值、会员价、阶梯价、满减表，避免商品草稿主记录有 scope、关联明细仍是“无主体”数据
- 已将 2.2 的正式生成链源同步到 `api/admin/doc/api/pms/*.api` 与 `rpc/pms/pms.proto`：补齐独立 `AddProductSku` 的 `spuId`、SPU 嵌套会员价/阶梯价/满减/属性值/SKU 明细的可回传 `id`，并把库存锁定请求的 `scope` 正式写回可生成源，而不是只停留在拆分 proto 草稿
- 已新增 admin-api 定向测试，覆盖 `AddProductSku` 的 `spuId + scope` 透传，以及 `build*List` / `buildUpdate*List` 对嵌套明细 `id` 的显式映射，确认 2.2 契约收口不再依赖手工补字段
- 已补 `QueryProductSpuList` 对 `productSn` 的透传，以及 `UpdateVerifyStatus` 对 `updateBy/reviewMan/detail/scope` 的显式映射，避免 2.2 商品草稿查询和送审备注链路继续丢字段
- 已修正 `api/admin/doc/api/pms/product_spu.api` 中 `UpdateProductSpuStatusReq.detail` 的正式契约标签，从 `form` 改为 `json`，并新增 handler 级解析回归，确保 Web Admin 以 JSON body 提交审核备注时不会被静默吞掉
- 已补 `product_sku` admin-api 测试覆盖：新增 `AddProductSku` 错误映射、`QueryProductSkuList` scope+筛选透传、`UpdateProductSku` updateBy+scope 透传、`DeleteProductSku` scope+错误映射回归，继续向 story 8 的 admin-api 验证面收口
- 已修正 `web-admin/src/pages/pms/ProductSpu/service.ts` 的提交构造：不再把 `ladderList/fullList/memberPriceList/skuList/attributeValueList` 强制清空或误塞进 `productData`，至少保证后续表单一旦提供嵌套数据，请求层不会先把 payload 丢掉
- 已继续收口 `web-admin/src/pages/pms/ProductSpu/`：编辑前先调用详情接口装载当前 SPU 的 `sku/memberPrice/ladder/full/attributeValue/subjectIds/prefrenceAreaIds`，更新时用详情快照保留这些嵌套明细，避免只改主表字段时把现有建档内容整体覆盖掉
- 已把 `ProductSpu` 页面主提交流与正式契约再拉近一段：补 `productSn` 搜索/编辑、补 `subTitle` 表单字段、过滤 `brief/description/priceRange/createBy/updateBy/isDeleted` 等历史页面噪音字段，并修正 `promotionType` 的前端取值与展示口径
- 已通过 `npx eslint src/pages/pms/ProductSpu/data.d.ts src/pages/pms/ProductSpu/service.ts src/pages/pms/ProductSpu/index.tsx src/pages/pms/ProductSpu/components/AddModal.tsx src/pages/pms/ProductSpu/components/UpdateModal.tsx --format unix`，确认这轮 `ProductSpu` 前端定向改动无新增 lint 告警
- 已为 `web-admin/src/pages/pms/ProductSpu/components/` 新增 `NestedDraftSections.tsx`，并把 `AddModal` / `UpdateModal` 接到真实的 `sku/memberPrice/ladder/full/attributeValue` 分段表单，2.2 的 SPU 建档页已不再只能“保留已有嵌套数据”，而是可以直接编辑并提交这些明细
- 已将 `ProductSpu` 新增 / 编辑弹窗扩为可滚动大弹窗，补默认状态值与首条 SKU 草稿初始化，同时把主表单的状态单选文案调整到更贴近正式商品语义的“下架/上架、否/是、未审核/审核通过”
- 已通过 `npx prettier --write src/pages/pms/ProductSpu/data.d.ts src/pages/pms/ProductSpu/service.ts src/pages/pms/ProductSpu/index.tsx src/pages/pms/ProductSpu/components/AddModal.tsx src/pages/pms/ProductSpu/components/UpdateModal.tsx src/pages/pms/ProductSpu/components/NestedDraftSections.tsx` 与对应 `eslint` 定向校验，确认嵌套建档组件接线后无新增前端 lint 问题
- 已新增 `useCatalogOptions.ts` 并接入 `ProductSpu` 新增 / 编辑弹窗，直接复用 Story 2.1 的 `ProductCategory / ProductBrand / ProductAttribute` 查询服务，避免 SPU 建档继续依赖手工输入分类、品牌和属性 ID
- 已将 `ProductSpu` 的基础建档字段从自由输入收口为目录选择器：商品分类改为树形选择并自动回填 `categoryName/categoryIds`，商品品牌改为候选下拉并自动回填 `brandName`，属性值明细里的 `attributeId` 也改为属性候选项下拉
- 已再次通过 `ProductSpu` 定向 prettier + eslint，确认目录联动接入后前端仍保持可编译、无新增 lint 告警
- 已新增 `web-admin/src/pages/pms/ProductSpu/draftFeedback.ts` 与对应 Jest 用例，把后端返回的草稿错误重新映射到 `skuList` / `attributeValueList` 的具体字段，并为 `NestedDraftSections` 补齐 SKU 规格 JSON、SKU 编码、价格、库存、预警库存、属性绑定的即时校验，避免前端只能提示“建档失败”却无法定位具体问题行
- 已补 `api/admin/internal/logic/pms/product_spu/productspu_logic_test.go` 的 `UpdateVerifyStatus` 错误映射断言，以及 `rpc/pms/internal/logic/productspuservice/productspu_draft_test.go` 的送审记录断言，给 2.2 -> 2.3 的最小草稿送审闭环补上自动化证据

### Stage Acceptance Notes

- `2026-03-23`：`git log --oneline -10` 已出现 `feat: advance story 2.2 product draft validation and feedback`，且提交实际落到了 `rpc/pms/internal/logic/common/product_draft_validation.go`、`rpc/pms/internal/logic/productskuservice/*`、`rpc/pms/internal/logic/productspuservice/*`、`web-admin/src/pages/pms/ProductSpu/index.tsx`、`web-admin/src/pages/pms/ProductSku/index.tsx` 等 2.2 实现文件，说明开发已开始，`ready-for-dev` 与真实代码状态不一致。
- 本次分析已将 story 与 sprint 状态回写为 `in-progress`，避免后续调度继续把一个已启动开发的 story 误判为未开始。
- 当前实现与最近提交部分一致：已覆盖商品草稿校验、SKU 维护、错误反馈与针对性测试；但与 story 完整范围仍不完全一致，暂无证据表明可直接进入 `code-review`。
- `2026-03-23`：已补 `rpc/pms/internal/logic/productspuservice/productspu_draft_test.go`，并通过 `cd rpc/pms && go test ./internal/logic/common ./internal/logic/productspuservice ./internal/logic/productskuservice`，说明 2.2 在 SPU/SKU 保存与校验链路上又向前推进一段。
- `2026-03-23`：已通过 `goctl api go -api ./api/admin/doc/api/admin.api -dir ./api/admin/` 与 `goctl rpc protoc rpc/pms/pms.proto --go_out=./rpc/pms/ --go-grpc_out=./rpc/pms/ --zrpc_out=./rpc/pms/ -m` 回刷生成链，并通过 `go test ./api/admin/internal/logic/pms/product_spu ./api/admin/internal/logic/pms/product_sku ./rpc/pms/internal/logic/productspuservice ./rpc/pms/internal/logic/productskuservice`，说明 admin-api 与 pms-rpc 的 2.2 契约源已开始真正同步收口。
- `2026-03-23`：已通过 `go test ./api/admin/internal/handler/pms/product_spu ./api/admin/internal/logic/pms/product_spu ./api/admin/internal/logic/pms/product_sku`，确认 `UpdateProductSpuStatusReq.detail` 的 JSON 解析、`product_spu` 的审核/列表透传，以及 `product_sku` 的 scope/错误映射回归全部为绿；同时 `web-admin` 提交层已不再默认清空 2.2 所需嵌套 payload。
- `2026-03-23`：已通过 `web-admin` 的 `ProductSpu` 定向 prettier + eslint，且编辑态现在会先取详情再提交更新，说明 2.2 的前端建档链已从“会丢嵌套数据的生成页”推进到“主字段可编辑且不会覆盖掉既有嵌套明细”的状态。
- `2026-03-23`：已把 `NestedDraftSections` 接入 `ProductSpu` 新增 / 编辑弹窗，并通过定向 prettier + eslint，说明 2.2 的 Web Admin 已具备直接编辑 `SKU/会员价/阶梯价/满减/属性值` 嵌套 payload 的能力，不再只停留在“保留旧明细、无法新增或修改”的阶段。
- `2026-03-23`：已将 2.1 目录基础数据接入 `ProductSpu` 建档表单，分类/品牌/属性不再完全依赖手输，说明 2.2 的 Web Admin 已开始符合“按目录语义建档而非纯字段录入”的目标。
- `2026-03-23`：已通过 `./node_modules/.bin/eslint src/pages/pms/ProductSpu/draftFeedback.ts src/pages/pms/ProductSpu/draftFeedback.test.ts src/pages/pms/ProductSpu/components/NestedDraftSections.tsx src/pages/pms/ProductSpu/components/useCatalogOptions.ts src/pages/pms/ProductSpu/components/AddModal.tsx src/pages/pms/ProductSpu/components/UpdateModal.tsx src/pages/pms/ProductSpu/data.d.ts src/pages/pms/ProductSpu/index.tsx --format unix`，确认这轮前端新增的表单反馈与候选项接线没有引入新的 lint 问题。
- `2026-03-23`：已通过 `./node_modules/.bin/umi test --runInBand src/pages/pms/ProductSpu/draftFeedback.test.ts`，确认 SKU 规格归一化、重复校验与后端错误映射 helper 可用；执行时需避免 login shell，否则会被本机 `SecItemCopyMatching failed -50` 环境噪音打断。
- `2026-03-23`：已通过 `GOCACHE=/tmp/zero-admin-go-cache go test ./api/admin/internal/logic/pms/product_spu ./api/admin/internal/logic/pms/product_sku ./api/admin/internal/handler/pms/product_spu ./rpc/pms/internal/logic/productspuservice`，确认 admin-api 的 scope/错误映射与 rpc/pms 的草稿送审记录链路均为绿，2.2 已具备“草稿可送审”的最小自动化证据。
- 当前缺口集中在：Web Admin 虽已接入目录候选项、即时校验和字段级错误回填，但目录禁用/失效态的前置阻断、1.4/1.5/1.6A/1.6B/2.1 的系统级回归，以及高风险动作的 `Consequence Preview` 仍未补齐，说明本 story 仍更接近“开发中途”而非“开发完成”。
- 建议下一 BMAD 节点：继续 `dev-story:2-2-商品规格-spu-sku-与库存建档`，优先补齐契约 / 数据模型 / 剩余测试与任务回写；完成后再进入 `code-review`。

### File List

- `_opcos/implementation-artifacts/2-2-商品规格-spu-sku-与库存建档.md`
- `_opcos/implementation-artifacts/sprint-status.yaml`
- `rpc/pms/internal/logic/productspuservice/addproductspulogic.go`
- `rpc/pms/internal/logic/productspuservice/updateproductspulogic.go`
- `rpc/pms/internal/logic/productspuservice/es_sync.go`
- `rpc/pms/internal/logic/productspuservice/productspu_draft_test.go`
- `rpc/pms/internal/logic/common/write_scope.go`
- `rpc/pms/internal/logic/common/write_scope_test.go`
- `rpc/pms/internal/logic/productskuservice/maintainproductsku_test.go`
- `api/admin/doc/api/pms/product_attribute_value.api`
- `api/admin/doc/api/pms/product_full_reduction.api`
- `api/admin/doc/api/pms/product_ladder.api`
- `api/admin/doc/api/pms/product_member_price.api`
- `api/admin/doc/api/pms/product_spu.api`
- `api/admin/doc/api/pms/product_sku.api`
- `api/admin/internal/handler/routes.go`
- `api/admin/internal/handler/pms/product_spu/updateverifystatushandler_test.go`
- `api/admin/internal/types/types.go`
- `api/admin/internal/logic/pms/product_sku/addproductskulogic.go`
- `api/admin/internal/logic/pms/product_sku/addproductskulogic_test.go`
- `api/admin/internal/logic/pms/product_spu/addproductspulogic.go`
- `api/admin/internal/logic/pms/product_spu/queryproductspulistlogic.go`
- `api/admin/internal/logic/pms/product_spu/updateverifystatuslogic.go`
- `api/admin/internal/logic/pms/product_spu/productspu_logic_test.go`
- `api/admin/internal/logic/pms/product_spu/updateproductspulogic.go`
- `api/admin/internal/logic/pms/product_spu/productspu_mapping_test.go`
- `web-admin/src/pages/pms/ProductSpu/data.d.ts`
- `web-admin/src/pages/pms/ProductSpu/draftFeedback.ts`
- `web-admin/src/pages/pms/ProductSpu/draftFeedback.test.ts`
- `web-admin/src/pages/pms/ProductSpu/components/AddModal.tsx`
- `web-admin/src/pages/pms/ProductSpu/components/NestedDraftSections.tsx`
- `web-admin/src/pages/pms/ProductSpu/components/UpdateModal.tsx`
- `web-admin/src/pages/pms/ProductSpu/components/useCatalogOptions.ts`
- `web-admin/src/pages/pms/ProductSpu/index.tsx`
- `web-admin/src/pages/pms/ProductSpu/service.ts`
- `rpc/pms/pms.proto`
- `rpc/pms/pmsclient/pms.pb.go`
- `rpc/pms/proto/product_spu.proto`
- `rpc/pms/proto/product_sku.proto`
- `script/sql/pms/migration_20260323_product_draft_scope_constraints.sql`
