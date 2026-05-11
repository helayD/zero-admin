# 区块链技术选型与演进策略

> **项目**：zero-admin（九克城电商平台）
> **文档类型**：架构决策记录（ADR）/ 技术选型说明
> **版本**：v2.0
> **最后更新**：2026-04-23
> **状态**：实施中（Active）

---

## 目录

- [一、背景与动机](#一背景与动机)
- [二、市场调研](#二市场调研)
- [三、候选方案对比](#三候选方案对比)
- [四、选型决策](#四选型决策)
- [五、三阶段演进路线](#五三阶段演进路线)
- [六、业务场景映射](#六业务场景映射)
- [七、当前实现概述](#七当前实现概述)
- [八、系统架构设计](#八系统架构设计)
- [九、数据模型](#九数据模型)
- [十、代码结构](#十代码结构)
- [十一、多链并存与切换策略](#十一多链并存与切换策略)
- [十二、风险与注意事项](#十二风险与注意事项)
- [十三、成本评估](#十三成本评估)
- [十四、TODO / 下一步](#十四todo--下一步)
- [附录 A：术语表](#附录-a术语表)
- [附录 B：相关 Story 索引](#附录-b相关-story-索引)
- [附录 C：FISCO BCOS 3.x 私钥与证书轮换 SOP](#附录-cfisco-bcos-3x-私钥与证书轮换-sop)
- [附录 D：合约源码仓库（git submodule）](#附录-d合约源码仓库git-submodule)

---

## 一、背景与动机

### 1.1 业务驱动

zero-admin（九克城）作为电商平台，预期在以下场景中引入区块链能力：

- **商品溯源 / 防伪**：关键品类（例如生鲜、奢侈品、礼品）的流转链路上链
- **订单 / 合同存证**：重要订单、退换货协议、企业客户合同等留痕
- **会员权益凭证**：VIP 权益、限定联名款、积分等数字化凭证
- **营销活动公平性**：抽奖、秒杀等关键事件的不可篡改证据
- **运营审计**：重要后台操作（价格变更、优惠券发放等）存证

### 1.2 原方案的问题

初始调研时倾向直接接入 **蚂蚁链 BaaS**，理由是大厂背书、司法采信强。进入阿里云 / 蚂蚁数科控制台后发现：

- **最低套餐约 9900 元/月**（2026 年 4 月实测报价）
- 年化成本 ≥ 12 万元
- 该费用仅为"开通费 / 节点费"，不含存证调用量
- 对当前拟发布的商业化首发阶段和中小规模业务而言，成本严重过高

### 1.3 决策目标

在不牺牲未来可扩展性的前提下，选择一套：

1. **商业化首发成本可控**（支持小规模生产部署，而不是以单机试验为前提）
2. **技术栈主流**（避免后期无路可走）
3. **能平滑演进**到司法级存证方案和大厂 BaaS
4. **不绑死在任一厂商生态**

的区块链集成方案。

---

## 二、市场调研

### 2.1 头部数藏 / 区块链应用平台选型

| 平台 | 背景 | 底层链 | 现状 |
|------|------|--------|------|
| 鲸探 | 蚂蚁集团 | **蚂蚁链** | 运营中，行业标杆 |
| 幻核 | 腾讯 | **至信链**（FISCO BCOS 改造） | 2022.8 停运 |
| 灵稀 | 京东科技 | **京东智臻链** | 运营中 |
| 元视觉 | 视觉中国 | **蚂蚁链** | 图片版权方向 |
| 小红书 R-SPACE | 小红书 | **Conflux 树图链** | 国内少数合规公链 |
| 网易星球 / 藏品 | 网易 | **伏羲通宝**（自研） | 游戏资产 |
| iBox 链盒 | 独立 | **BSN-DDC / 文昌链** | 多链切换史 |
| 唯一艺术 | 独立 | 以太坊 → Polygon → **BSN** | 经历监管转向 |
| 数藏中国 | 文创国家队 | **国版链 / 文创链** | 国家版权方向 |

### 2.2 链选型五大阵营

1. **大厂自家链**：蚂蚁链、至信链、智臻链、百度 XuperChain、网易伏羲 —— 头部平台首选
2. **国家队 / 准国家队**：BSN（文昌链、武汉链、泰安链）、长安链、国版链 —— 二三线平台主力
3. **开源自建派**：FISCO BCOS、Hyperledger Fabric —— 技术驱动团队的选择
4. **合规公链**：Conflux 树图 —— 公链里少有的国内合规项
5. **灰色地带**：以太坊 / Polygon / BSC —— 2022 年后国内项目基本退出

### 2.3 2022 年后的监管大变化

- **严禁二级市场 / 严禁炒作**：数藏从"交易属性"转向"权益凭证"
- **禁止公链承载**：国内项目全部迁往联盟链
- **平台大规模出清**：2023 年估计 80% 中小数藏平台关停
- **新方向**：IP 授权、文博联名、会员权益、消费凭证

### 2.4 FISCO BCOS 市占情况（截至 2024 官方披露）

- **4000+ 应用**
- **300+ 产业落地项目**
- **100+ 深度参与机构**

典型用户覆盖：微众银行、平安银行、招商银行、建设银行、上海票交所、海关总署、国家电网、粤澳健康码、深圳税务发票、广州仲裁委、至信链、BSN-DDC 多条链等。

> **关键事实**：**至信链的底层就是 FISCO BCOS 的深度定制版本**；BSN-DDC 的文昌链、武汉链等主流链也基于 FISCO BCOS。这意味着 FISCO 是国内联盟链最大的"技术动脉"，从 FISCO 出发，往上游迁移的成本最低。

---

## 三、候选方案对比

| 维度 | 自建 FISCO BCOS | 至信链 | 蚂蚁链 | BSN-DDC（文昌链等） | 长安链 | Hyperledger Fabric |
|------|-----------------|--------|--------|----------------------|--------|--------------------|
| 入场费 | **0** | 0 | **9900/月起** | 几百~几千/月 | 几千/月 | 0 |
| 按量费用 | 几乎 0 | 0.1~1 元/条 | 套餐内 | 低 | 低 | 几乎 0 |
| 司法采信 | ❌ 需组合方案 | ✅ | ✅ | 部分 ✅ | 部分 ✅ | ❌ |
| 国产合规 | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ 国际 |
| 开源可控 | ✅ | ❌ | ❌ | 部分 | 部分 | ✅ |
| 学习曲线 | 中 | 低（REST API） | 低（REST API） | 中 | 中 | 高 |
| 中文生态 | ✅ 极好 | ✅ | ✅ | ✅ | ✅ | ⚠️ 一般 |
| 适用阶段 | 商业化首发 ~ 中型 | 司法增强期 | 大型 / 资产化业务 | 中小型商业化 | 中大型商业化 | 国际化项目 |

---

## 四、选型决策

### 4.1 结论

采用 **"FISCO 做基础存证底座 + 特殊场景按需接入至信链 / 蚂蚁链 + 多链 Adapter 解耦"** 的稳态架构：

- **基础存证层**：默认由自建 FISCO BCOS 承担商品溯源、订单存证、运营审计等通用存证需求
- **特殊场景层**：司法采信类场景按需接入至信链，数字权益 / 提货卡类场景按需接入蚂蚁链或长安链
- **代码层面**：通过 Adapter 和链路由配置解耦业务逻辑与具体链实现
- **长期形态**：不是"只选一条链"，而是让 FISCO 长期保留为基础底座，特殊业务按能力接入专用链并可多链并存

### 4.2 选型理由

1. **商业化首发成本可控**：相较托管 BaaS，自建 FISCO 能以更低成本完成首发所需的基础集群、监控和备份建设
2. **通用存证与特殊业务天然分层**：普通审计 / 溯源 / 存证不必一开始就绑定高成本 BaaS，司法和资产场景再单独升级更稳
3. **中文生态成熟**：微众银行开源，WeBASE、WeIdentity、WeEvent 等配套完善
4. **Go SDK 成熟**：与 zero-admin 的 Go / go-zero 技术栈天然契合
5. **不绑定厂商**：基础层留在自建链，特殊能力按需接入至信链、蚂蚁链、长安链，避免单厂商锁定
6. **业务数据资产在自己手里**：链只存哈希或凭证指纹，原始数据长期保留在 MySQL
7. **演进风险更低**：未来是"新增专用链能力"，而不是"整体替换基础链"，迁移面更可控

### 4.3 反对意见与回应

| 反对意见 | 回应 |
|----------|------|
| "自建链没有司法效力" | 关键业务可以**双写**：自建 FISCO + 按量调用至信链存证 API |
| "运维成本高" | 商业化首发可采用 4 节点 + 监控告警 + 备份的生产部署方案，运维复杂度仍明显低于高成本专有 BaaS 体系 |
| "以后换蚂蚁链要重构" | 不会。基础存证继续留在 FISCO，数字权益 / 资产类场景通过独立 Adapter 接入蚂蚁链即可 |
| "找不到区块链工程师" | FISCO BCOS 有详细中文文档 + 活跃社区，Go / Java 工程师上手 1~2 周 |

---

## 五、三阶段演进路线

### 阶段一：商业化首发（0~6 个月）

- **目标**：以正式发布标准把"基础存证底座"接入 zero-admin，并上线首批真实业务场景
- **链**：生产部署 FISCO BCOS（至少 4 节点 + 1 个 WeBASE + 监控告警 + 备份恢复）
- **范围**：商品溯源、订单 / 合同存证、关键运营日志等通用存证场景
- **成本**：相较托管 BaaS 明显可控，满足商业化首发预算
- **产出**：
  - `blockchain_evidence` 表落库
  - `Evidencer` 接口 + `fisco` Adapter
  - `PutEvidence` / `VerifyEvidence` RPC
  - 监控告警、备份恢复、密钥管理、链路由配置等生产基线能力

### 阶段二：规模化运营期（6~18 个月）

- **目标**：根据真实业务规模和合规要求，明确哪些场景继续留在 FISCO，哪些场景需要专用链增强
- **链**：FISCO 继续作为基础存证底座
- **可选扩展**：
  - 增加节点数量、引入 WeBASE / WeEvent
  - 将关键节点托管到阿里云 / 腾讯云 BCS（混合部署）
  - 为司法采信和数字权益场景预留路由规则与 Adapter 配置
- **判断指标**：
  - 月存证条数是否超过 **10 万**
  - 是否出现需司法采信的诉讼 / 仲裁
  - 是否要发行数字化会员权益凭证

### 阶段三 A：存证型升级（按需触发）

- **触发条件**：出现电子合同、法律文书、重要订单需要司法采信
- **方案**：在保留 FISCO 基础存证的前提下，新增 **至信链 Adapter**，对关键业务双写（FISCO + 至信链）
- **计费**：按条数，0.1~1 元/条
- **改动**：
  - 新增 `zxchain` Adapter 文件
  - 配置里增加 `Blockchain.ZXChain.{AppID, AppKey}`
  - 仅调整链路由配置，业务代码**零改动**

### 阶段三 B：资产型升级（按需触发）

- **触发条件**：发行会员数字权益凭证、做合规提货卡、IP 授权
- **方案**：在保留 FISCO 基础存证的前提下，接入 **长安链 BaaS** 或 **蚂蚁链 BaaS**
- **计费**：几千~几万/月
- **改动**：
  - 新增 `antchain` 或 `chainmaker` Adapter
  - 提货卡相关业务路由到新链
  - 通用存证类业务继续保留在 FISCO，司法增强类业务保留 FISCO / 至信链 双写

### 演进路线示意

```text
 ┌────────────┐
 │ Stage 1    │  FISCO BCOS 自建
 │ 首发版     │  先承接基础存证
 └─────┬──────┘
       │
       ├─── 通用存证 / 审计 / 溯源 ──► 持续留在 FISCO（长期底座）
       │
       ├─── 需司法采信 ───────────► 新增 至信链 Adapter（FISCO + 至信链）
       │
       └─── 数字权益 / 资产场景 ───► 新增 蚂蚁链 / 长安链 Adapter
```

---

## 六、业务场景映射

### 6.1 zero-admin 场景下的上链方案

| 业务场景 | 是否上链 | 链类型（当前） | 未来演进 | 说明 |
|----------|----------|----------------|----------|------|
| 商品溯源（生鲜 / 奢品） | ✅ | FISCO | 保留自建 | 批次号 + 流转节点哈希上链 |
| 订单存证（普通） | ⚠️ 选择性 | FISCO | 保留自建 | 仅大额 / VIP 订单 |
| 电子合同 / B 端合同 | ✅ | FISCO | → 至信链 双写 | 法律效力强需求 |
| 退换货协议 | ⚠️ 选择性 | FISCO | → 至信链 | 涉及争议的存证 |
| 秒杀 / 抽奖结果 | ✅ | FISCO | 保留自建 | 公平性证据 |
| 优惠券发放记录 | ✅ | FISCO | 保留自建 | 审计用 |
| 后台关键操作日志 | ✅ | FISCO | 保留自建 | 管理员操作审计 |
| 会员数字权益凭证 | ✅ | FISCO（初期） | → 蚂蚁链（有能力后） | 提货卡发放已上线 |
| 提货卡 | ❌ | - | → 蚂蚁链 / 文昌链 | 监管严格，需合规 BaaS |

### 6.2 哪些业务**不应该上链**

- 频繁更新的数据（链是 append-only）
- 涉及个人隐私的原始数据（只能上哈希）
- 可以用数据库 + 签名替代的场景
- 低价值、高频率的日志（性价比过低）

---

## 七、当前实现概述

> **核心策略**：蚂蚁链 BaaS 年费 12 万+，不适合初期项目搭建。因此项目初期选择 **FISCO BCOS**（开源免费、自主可控）作为提货卡发放的链底座。架构设计上，业务代码只依赖 `ChainClient` 接口，不绑定任何具体链——**有能力后随时可以切换到蚂蚁链**，只需改一行配置，无需改代码。
>
> 蚂蚁链 Adapter 已在代码中完整实现（`pkg/antchain`），作为升级路径预留。当项目营收支撑蚂蚁链费用时，切换即可获得蚂蚁链的合规背书和更高的商业可信度。

### 7.1 已落地能力总览

| 能力 | 状态 | 技术实现 |
|------|------|---------|
| 提货卡链上发放（Mint Token） | 已上线 | `pkg/digitalcardmint` + `ChainClient` 接口 |
| FISCO BCOS 链底座（初期默认） | 已上线 | `pkg/fisco`（开源免费，初期首选） |
| 蚂蚁链 Adapter（升级路径预留） | 已实现 | `pkg/antchain`（代码就绪，改配置即可启用） |
| 发放任务状态机（pending/dispatched/running/succeeded/failed/frozen） | 已上线 | `digitalcardmint.Service` |
| 异步 MQ 派发 + 消费者执行 | 已上线 | RabbitMQ + `consumer/` |
| 超时补偿扫描（ScanDueTasks） | 已上线 | `job/` 定时任务 |
| 链路管理后台（查询/重试/冻结/升级人工复核） | 已上线 | admin-api `digital_card_chain` |
| 资产合规管理（审核/下线/回收） | 已上线 | admin-api `digital_card_asset` |
| 会员提货卡查看（我的卡片/详情/时间线） | 已上线 | front-api `digital_card_asset` |
| 抽卡活动系统（资格校验/概率抽奖/库存扣减） | 已上线 | `drawparticipationservice` |
| 幂等链上回执查询（QueryMintToken） | 已上线 | `ChainClient.QueryMintToken` |
| 司法增强存证（至信链） | 未启动 | 长期规划 |

### 7.2 与原始策略的对应关系

原始策略的阶段一（FISCO）是当前落地场景，阶段三 B（蚂蚁链）作为升级路径预留：

```text
原始规划路径                          实际执行路径
──────────────────────                ──────────────────────
Stage 1: FISCO 基础底座       ──►    首先落地: FISCO (免费、初期成本低)
Stage 3B: 蚂蚁链资产型升级    ──►    升级路径: 蚂蚁链 Adapter 已实现，改配置即可启用
Stage 3A: 至信链司法增强               (长期规划)
```

多链 Adapter 思想已在代码层面完整体现：`ChainClient` 接口同时有 FISCO 和 AntChain 两个实现，切换只需改配置。

---

## 八、系统架构设计

### 8.1 核心原则（实际执行）

1. **业务逻辑与链交互解耦**：`digitalcardmint.Service` 持有 `ChainClient` 接口，初期使用 FISCO 实现，有能力后可切换为蚂蚁链，业务代码无需改动
2. **异步执行 + 补偿机制**：链上操作通过 MQ 异步派发，失败后有自动重试和手工介入两条路径
3. **状态机驱动**：每个发放任务有明确的状态流转（pending_dispatch → dispatched → running → succeeded/failed/frozen）
4. **幂等性保证**：通过 `idempotency_key` 防止重复发放，通过 `QueryMintToken` 回执查询实现断点续传
5. **多租户 Scope 治理**：所有链路操作受 `GovernanceScope`（platform/tenant/merchant）约束

### 8.2 实际架构图

```text
┌─────────────────────────────────────────────────────────────────┐
│ API Layer                                                       │
│ ┌───────────────────┐  ┌───────────────────┐                    │
│ │ admin-api         │  │ front-api         │                    │
│ │ digital_card_chain│  │ digital_card_asset│                    │
│ │ digital_card_asset│  │ draw_activity     │                    │
│ └────────┬──────────┘  └────────┬──────────┘                    │
└──────────┼──────────────────────┼───────────────────────────────┘
           │ gRPC                 │ gRPC
           ▼                      ▼
┌──────────────────────────────────────────────────────────────────┐
│ sms-rpc (营销服务)                                                │
│ ┌──────────────────────────────────────────────┐                 │
│ │ CardMintAdminService (gRPC)                  │                 │
│ │  QueryTaskList / QueryTaskDetail             │                 │
│ │  RetryTask / FreezeTask / EscalateTask       │                 │
│ │  QueryAssetAuditList / ReviewCompliance / ... │                 │
│ └───────────────────┬──────────────────────────┘                 │
│                     │                                            │
│ ┌───────────────────▼──────────────────────────┐                 │
│ │ digitalcardmint.Service (pkg/digitalcardmint)│                 │
│ │  EnsureTaskTx() - 创建发放任务                │                 │
│ │  DispatchTask() - MQ 异步派发                 │                 │
│ │  ExecuteTask()  - 调用 ChainClient 铸造       │                 │
│ │  ScanDueTasks() - 超时补偿扫描                │                 │
│ │  RetryTask() / FreezeTask() / EscalateTask() │                 │
│ │  QueryMember/AuditAsset*()                   │                 │
│ │  Review/Offline/Recycle*()                    │                 │
│ └──────────┬───────────────────┬───────────────┘                 │
│            │                   │                                  │
│  ┌─────────▼─────────┐  ┌─────▼──────────────┐                  │
│  │ ChainClient        │  │ MQPublisher        │                  │
│  │ (FISCO / AntChain) │  │ (pkg/mq.RabbitMQ)  │                  │
│  │ MintToken()        │  │ SendMessage()      │                  │
│  │ QueryMintToken()   │  └─────┬──────────────┘                  │
│  └──────────┬─────────┘        │                                  │
└─────────────┼──────────────────┼──────────────────────────────────┘
              │ SDK/HTTP          │ AMQP
              ▼                   ▼
   [FISCO / 蚂蚁链 BaaS]   ┌─────────────┐     ┌──────────────┐
                           │  RabbitMQ   │────►│  consumer    │
                           └─────────────┘     │  MintRequested│
                                               │  → ExecuteTask│
                                               └──────────────┘
                                                      │
                                               ┌──────▼──────┐
                                               │    job      │
                                               │ ScanDueTasks│
                                               │ 超时补偿     │
                                               └─────────────┘
```

### 8.3 提货卡发放完整流程

```text
1. 抽卡活动参与
   drawparticipationservice.ParticipateDraw()
   → 校验资格 → 概率抽奖 → 扣减库存 → 创建 sms_card_instance
   → EnsureTaskTx() 创建 sms_card_mint_task
   → DispatchTask() 发送 MQ 消息

2. 异步链上铸造
   consumer/MintRequested() 收到 MQ 消息
   → ExecuteTask()
   → ChainClient.MintToken() 调用链底座 API (初期 FISCO)
   → 成功: transitionSuccess() 更新 token_id/chain_tx_id
   → 失败: transitionFailure() 更新错误码, 安排重试

3. 超时补偿
   job/HandleCardMintTimeout()
   → ScanDueTasks() 扫描滞留任务
   → 状态为 pending_dispatch → 重新 DispatchTask()
   → 状态为 dispatched/running 超时 → 直接 ExecuteTask()
   → 超过最大重试次数 → 升级为 manual_review

4. 人工干预（admin-api）
   → RetryTask(): 手动触发重试
   → FreezeTask(): 冻结链路，停止自动补偿
   → EscalateTask(): 升级为人工复核

5. 资产合规管理（admin-api）
   → ReviewAssetCompliance(): 标记人工复核
   → OfflineAssetDisplay(): 下线展示
   → RecycleAsset(): 回收资产
```

### 8.4 任务状态机

```text
                  ┌──────────────────┐
                  │ pending_dispatch │ ◄── EnsureTaskTx() / RetryTask()
                  └────────┬─────────┘
                           │ DispatchTask()
                  ┌────────▼─────────┐
                  │   dispatched     │
                  └────────┬─────────┘
                           │ ExecuteTask()
                  ┌────────▼─────────┐
                  │    running       │
                  └──┬─────────┬─────┘
          成功       │         │ 失败
     ┌───────────────▼┐  ┌────▼──────────┐
     │   succeeded    │  │    failed     │
     └────────────────┘  └──┬────┬───────┘
                            │    │ 超过重试上限
                    RetryTask│  ┌▼───────────────┐
                            │  │ manual_review   │
                            │  └─────────────────┘
                            │
                  ┌─────────▼────────┐
                  │     frozen       │ ◄── FreezeTask()
                  └──────────────────┘
```

---

## 九、数据模型

### 9.1 核心表一览

| 表名 | 用途 | 关键字段 |
|------|------|---------|
| `sms_card_mint_task` | 链上发放任务 | task_status, mint_status, chain_status, token_id, chain_tx_id |
| `sms_card_instance` | 提货卡资产实例 | asset_status, mint_status, chain_status, display_status, compliance_status |
| `sms_card_asset_log` | 资产操作日志（审计轨迹） | operation_type, from_status, to_status |
| `sms_draw_activity` | 抽卡活动配置 | status, real_name_required, probability_rule |
| `sms_draw_pool` | 奖池配置 | probability_rule |
| `sms_draw_pool_template` | 奖池-卡片模板关联 | probability, remaining_limit |
| `sms_card_template` | 卡片模板 | template_name, card_face_image, display_status |
| `sms_draw_participation_record` | 参与记录 | result_type, result_status, asset_instance_id |

### 9.2 `sms_card_mint_task` 表（链上发放任务）

```sql
CREATE TABLE sms_card_mint_task (
  id                      BIGINT PRIMARY KEY AUTO_INCREMENT,
  platform_id             BIGINT        NOT NULL COMMENT '平台ID',
  tenant_id               BIGINT        NOT NULL COMMENT '租户ID',
  merchant_id             BIGINT        NOT NULL COMMENT '商户ID',
  asset_instance_id       BIGINT        NOT NULL COMMENT '关联资产实例ID',
  participation_record_id BIGINT        NOT NULL COMMENT '参与记录ID',
  activity_id             BIGINT        NOT NULL COMMENT '活动ID',
  member_id               BIGINT        NOT NULL COMMENT '会员ID',
  request_id              VARCHAR(64)            COMMENT '请求ID',
  trace_id                VARCHAR(64)            COMMENT '链路追踪ID',
  idempotency_key         VARCHAR(128)  NOT NULL COMMENT '幂等键 (防重复铸造)',
  task_status             VARCHAR(32)   NOT NULL COMMENT 'pending_dispatch/dispatched/running/succeeded/failed/manual_review/frozen',
  mint_status             VARCHAR(32)   NOT NULL COMMENT 'mint_pending/mint_processing/mint_success/mint_failed/mint_compensating/mint_manual_review/mint_frozen',
  chain_status            VARCHAR(32)   NOT NULL COMMENT 'unknown/processing/success/failed/frozen',
  token_id                VARCHAR(128)           COMMENT '链上 Token ID',
  chain_tx_id             VARCHAR(128)           COMMENT '链上交易哈希',
  retry_count             INT           NOT NULL DEFAULT 0,
  max_retry_count         INT           NOT NULL DEFAULT 3,
  last_error_code         VARCHAR(64)            COMMENT '最近错误码',
  last_error_reason       VARCHAR(512)           COMMENT '最近错误原因',
  last_receipt_summary    VARCHAR(512)           COMMENT '最近链上回执摘要',
  last_receipt_json       TEXT                   COMMENT '最近链上回执原文 (JSON)',
  last_execute_at         DATETIME               COMMENT '最近执行时间',
  next_retry_at           DATETIME               COMMENT '下次重试时间',
  manual_required         TINYINT       NOT NULL DEFAULT 0 COMMENT '是否需要人工介入',
  frozen                  TINYINT       NOT NULL DEFAULT 0 COMMENT '是否已冻结',
  freeze_reason           VARCHAR(512)           COMMENT '冻结原因',
  create_by               BIGINT                 COMMENT '创建人',
  update_by               BIGINT                 COMMENT '更新人',
  create_time             DATETIME      NOT NULL,
  update_time             DATETIME,
  is_deleted              TINYINT       NOT NULL DEFAULT 0
) COMMENT '提货卡链上发放任务';
```

### 9.3 `sms_card_instance` 表（提货卡资产实例）

```sql
CREATE TABLE sms_card_instance (
  id                      BIGINT PRIMARY KEY AUTO_INCREMENT,
  platform_id             BIGINT        NOT NULL,
  tenant_id               BIGINT        NOT NULL,
  merchant_id             BIGINT        NOT NULL,
  activity_id             BIGINT        NOT NULL,
  member_id               BIGINT        NOT NULL,
  participation_record_id BIGINT        NOT NULL,
  request_id              VARCHAR(64),
  trace_id                VARCHAR(64),
  scope                   VARCHAR(32),
  pool_id                 BIGINT        NOT NULL,
  template_id             BIGINT        NOT NULL,
  rarity                  VARCHAR(32),
  asset_no                VARCHAR(64)   COMMENT '资产编号',
  asset_status            VARCHAR(32)   COMMENT 'asset_created',
  mint_status             VARCHAR(32)   COMMENT '铸造状态 (同步自 mint_task)',
  token_id                VARCHAR(128)  COMMENT '链上 Token ID',
  chain_status            VARCHAR(32),
  display_status          VARCHAR(32)   COMMENT 'display_visible/display_hidden/display_offlined/display_recycled',
  compliance_status       VARCHAR(32)   COMMENT 'compliance_clear/compliance_review/compliance_restricted/compliance_recycled',
  display_reason          VARCHAR(512),
  compliance_reason       VARCHAR(512),
  rule_snapshot_json      TEXT,
  last_receipt_at         DATETIME,
  mint_task_id            BIGINT        COMMENT '关联铸造任务ID',
  issued_at               DATETIME,
  disposed_at             DATETIME,
  disposed_by             BIGINT,
  create_by               BIGINT,
  create_time             DATETIME,
  update_by               BIGINT,
  update_time             DATETIME,
  is_deleted              TINYINT       NOT NULL DEFAULT 0
) COMMENT '提货卡资产实例';
```

### 9.4 `sms_card_asset_log` 表（资产操作日志）

```sql
CREATE TABLE sms_card_asset_log (
  id                      BIGINT PRIMARY KEY AUTO_INCREMENT,
  asset_instance_id       BIGINT        NOT NULL,
  participation_record_id BIGINT        NOT NULL,
  from_status             VARCHAR(32)   COMMENT '操作前状态',
  to_status               VARCHAR(32)   COMMENT '操作后状态',
  operation_type          VARCHAR(64)   NOT NULL COMMENT 'mint_requested/mint_dispatching/mint_succeeded/mint_failed/...',
  operator_type           VARCHAR(32)   NOT NULL COMMENT 'system/job/manual',
  trace_id                VARCHAR(64),
  reason_code             VARCHAR(64),
  reason_text             VARCHAR(512),
  payload_json            TEXT          COMMENT '操作上下文快照 (JSON)',
  create_time             DATETIME      NOT NULL
) COMMENT '提货卡资产操作日志';
```

### 9.5 设计要点

- **三层状态**：`task_status`（任务调度层）、`mint_status`（业务语义层）、`chain_status`（链层）分离，互不耦合
- **幂等键**：`idempotency_key` 由 `buildIdempotencyKey(assetInstanceID)` 生成，保证同一资产不会重复铸造
- **回执持久化**：`last_receipt_json` 保留链上原始回执，断电重启后可从回执恢复状态（receipt replay）
- **合规双维度**：`display_status` 控制展示可见性，`compliance_status` 控制合规审核状态，二者独立管理
- **Scope 治理**：所有表都有 `platform_id` / `tenant_id` / `merchant_id`，查询时通过 `GovernanceScope` 过滤

---

## 十、代码结构

### 10.1 实际目录布局

```text
pkg/
├── antchain/                          # 蚂蚁链客户端（可替换的链适配层）
│   ├── client.go                      # Client 接口 + httpClient 实现 + disabledClient
│   ├── types.go                       # Config, MintTokenRequest/Response, QueryMintTokenRequest
│   └── mock.go                        # MockClient (用于单元测试)
│
├── digitalcardmint/                   # 提货卡发放核心服务
│   ├── constants.go                   # 状态常量 (TaskStatus*, MintStatus*, ChainStatus*, Operation*, Event*)
│   ├── asset_constants.go             # 资产合规常量 (DisplayStatus*, ComplianceStatus*)
│   ├── model.go                       # GORM 模型 + RPC DTO (CardMintTaskRow, CardInstanceRow, ...)
│   ├── service.go                     # 核心状态机 (EnsureTaskTx, DispatchTask, ExecuteTask, ScanDueTasks, ...)
│   ├── asset_service.go               # 资产查询与合规管理 (QueryMemberAsset*, QueryAudit*, Review/Offline/Recycle*)
│   ├── scan_time.go                   # nullableTime 辅助类型
│   ├── service_test.go                # 核心状态机单元测试
│   └── asset_service_test.go          # 资产服务单元测试
│
rpc/sms/                               # 营销 RPC 服务 (承载提货卡链路)
├── internal/
│   ├── config/config.go               # Blockchain (FISCO/AntChain) + Rabbitmq 配置
│   ├── svc/service_context.go         # 初始化 ChainClient + digitalcardmint.Service
│   └── logic/
│       ├── cardassetservice/          # 资产查询 RPC 逻辑
│       └── drawparticipationservice/  # 抽卡参与 + 发放触发 RPC 逻辑
├── client/
│   └── cardmintadminservice/          # CardMintAdminService RPC 客户端
│       └── card_mint_admin_service.go # QueryTaskList/RetryTask/FreezeTask/...
└── cardmintadminrpc/
    └── rpc.go                         # gRPC ServiceDesc + structpb 编解码

api/admin/internal/logic/sms/
├── digital_card_chain/                # 链路管理 (查询/重试/冻结/升级)
│   ├── querydigitalcardchainlistlogic.go
│   ├── querydigitalcardchaindetaillogic.go
│   ├── querydigitalcardchainactionslogic.go
│   ├── retrydigitalcardchainlogic.go
│   ├── freezedigitalcardchainlogic.go
│   └── escalatedigitalcardchainlogic.go
└── digital_card_asset/                # 资产合规管理
    ├── querydigitalcardassetlistlogic.go
    ├── querydigitalcardassetdetaillogic.go
    ├── reviewdigitalcardassetcompliancelogic.go
    ├── offlinedigitalcardassetdisplaylogic.go
    └── recycledigitalcardassetlogic.go

api/front/internal/logic/digital_card/
├── digital_card_asset/                # 会员提货卡 (我的卡片)
│   ├── query_my_digital_card_asset_list_logic.go
│   └── query_my_digital_card_asset_detail_logic.go
└── draw_activity/                     # 抽卡活动 (落地页/资格/参与/记录)
    ├── query_draw_activity_landing_logic.go
    ├── preview_draw_eligibility_logic.go
    ├── participate_draw_logic.go
    └── query_my_draw_record_list_logic.go

consumer/internal/mq/digital_card/
└── mint_requested.go                  # MQ 消费者: 接收 MintRequested 事件 → ExecuteTask()

job/internal/jobs/
└── handle_card_mint_timeout_logic.go  # 定时任务: ScanDueTasks() 超时补偿
```

### 10.2 `ChainClient` 接口

```go
// 通用链客户端接口（FISCO 和 AntChain 均实现此接口）
type ChainClient interface {
    MintToken(ctx context.Context, req *MintTokenRequest) (*MintTokenResponse, error)
    QueryMintToken(ctx context.Context, req *QueryMintTokenRequest) (*MintTokenResponse, error)
    ChainType() string  // "fisco" 或 "antchain"
}
```

**FISCO 实现**（`pkg/fisco`，初期默认）：
- **`fiscoClient`**：通过 Go SDK 调用 FISCO BCOS 节点

**蚂蚁链实现**（`pkg/antchain`，升级路径预留）：
- **`httpClient`**：通过 HTTP POST 调用蚂蚁链 BaaS API
- **`disabledClient`**：未启用时返回 “蚂蚁链能力未启用” 错误
- **`MockClient`**：测试用，支持 `SetError()` / `SetQueryError()` 注入故障

### 10.3 配置示例（实际）

```yaml
# rpc/sms/etc/sms-rpc.yaml
Blockchain:
  Primary: fisco                    # 初期默认 FISCO

Fisco:
  Endpoint: "127.0.0.1:20200"
  GroupID: 1
  ContractAddress: "0x..."
  PrivateKeyPath: "./conf/fisco_sdk.key"
  Enabled: true

AntChain:
  Endpoint: "https://antchain-api.example.com/mint"
  ReceiptEndpoint: "https://antchain-api.example.com/query"
  AppId: "your-app-id"
  AccessKey: "your-access-key"
  Secret: "your-secret"
  TimeoutSeconds: 10
  Enabled: false                    # 升级时改为 true

Rabbitmq:
  Host: "127.0.0.1"
  Port: 5672
  UserName: "guest"
  Password: "guest"
```

### 10.4 ServiceContext 初始化（实际）

```go
// rpc/sms/internal/svc/service_context.go
func NewServiceContext(c config.Config) *ServiceContext {
    // ... DB, RabbitMQ 初始化 ...

    // 根据配置选择链客户端（初期 FISCO，有能力后切蚂蚁链）
    var chainClient digitalcardmint.ChainClient
    switch c.Blockchain.Primary {
    case "antchain":
        chainClient = antchain.NewClient(antchain.Config{...})
    default: // "fisco"
        chainClient = fisco.NewClient(fisco.Config{...})
    }

    cardMintService := digitalcardmint.NewService(DB, rabbitmq, chainClient)
    return &ServiceContext{
        Config:          c,
        DB:              DB,
        RabbitMQ:        rabbitmq,
        ChainClient:     chainClient,
        CardMintService: cardMintService,
    }
}
```

### 10.5 多链注册表（当前实现）

`ChainClient` 接口已有 FISCO 和 AntChain 两个实现，通过配置 `Blockchain.Primary` 选择当前活动链。未来引入至信链等新链时，只需新增一个实现并注册即可：

```go
// 当前已实现
type Registry map[string]ChainClient

var defaultRegistry = Registry{
    "fisco":    fisco.NewClient(fiscoConfig),      // 初期默认
    "antchain": antchain.NewClient(antchainConfig), // 升级路径预留
}

// 未来扩展: 新增至信链
// "zxchain": zxchain.NewClient(zxConfig),
```

---

## 十一、多链并存与切换策略

### 11.1 核心原则：FISCO 与蚂蚁链并存，后台随时可切

FISCO 与蚂蚁链都服务于同一业务领域——**提货卡发放、卡片铸造、权益凭证**。初期使用 FISCO（开源免费）控制成本，有能力后随时切换到蚂蚁链（商业 BaaS，合规背书更强），**无需修改代码、无需重新部署**。

```text
┌──────────────────────────────────────────────────────────┐
│ 业务层 (Logic)                                            │
│ 提货卡发放 / 卡片铸造 / 权益凭证                          │
│ 不关心底层用的是哪条链，只调用 ChainClient 接口               │
└───────────────────────┬──────────────────────────────────┘
                        │
            ┌───────────▼───────────────┐
            │ Chain Router (配置驱动)    │
            │ Primary: fisco (初期)    │
            │ 有能力后切为 antchain    │
            └─────┬──────────────┬──────┘
                  │              │
        ┌─────────▼────┐  ┌─────▼──────────┐
        │ FISCO BCOS   │  │ 蚂蚁链 AntChain │
        │ ★ 初期首选    │  │ ☆ 升级路径    │
        │ 开源免费     │  │ 商业 BaaS      │
        │ 自主可控     │  │ 合规背书更强   │
        └──────────────┘  └────────────────┘
```

**设计要点**：

1. **初期用 FISCO，控制成本**：FISCO 开源免费、自主部署，年成本约 7,000 元（服务器 + 带宽）
2. **有能力后切蚂蚁链**：蚂蚁链 BaaS 年费 12 万+，但合规背书更强、商业可信度更高，营收支撑后即可切换
3. **配置驱动切换**：`Blockchain.Primary` 从 `fisco` 改为 `antchain`，重启服务即生效，无需改代码
4. **历史数据不迁移**：切换前的数据按 `chain_type` 回查原链，链上数据不可搬家
5. **可选双写**：过渡期可配置同时写入两条链，确保平滑切换

### 11.2 配置切换示例

```yaml
# rpc/sms/etc/sms-rpc.yaml
Blockchain:
  Primary: fisco                  # 初期使用 FISCO (有能力后改为 antchain)
  Enabled: [fisco, antchain]      # 两条链同时就绪，随时可切
  DualWrite: false                # 过渡期可设为 true，同时写入两条链

  Fisco:
    Endpoint: "127.0.0.1:20200"
    GroupID: 1
    ContractAddress: "0x..."
    PrivateKeyPath: "./conf/fisco_sdk.key"
    Enabled: true                 # 初期默认启用

  AntChain:
    Endpoint: "https://antchain-api.example.com/mint"
    ReceiptEndpoint: "https://antchain-api.example.com/query"
    AppId: "your-app-id"
    AccessKey: "your-access-key"
    Secret: "your-secret"
    Enabled: false                # 初期不启用，升级时改为 true
```

**升级操作**：当项目营收能支撑蚂蚁链费用时，将 `Primary` 从 `fisco` 改为 `antchain`，同时将 `AntChain.Enabled` 设为 `true`，重启服务即生效。切换后新的卡片铸造走蚂蚁链，已铸造的历史数据仍在 FISCO 上回查。未来可扩展为热更新（通过 etcd/nacos 推送配置，无需重启）。

### 11.3 切换时的影响面

| 项目 | 是否受影响 | 说明 |
|------|-----------|------|
| 业务 Logic 代码 | ❌ 无变化 | 只依赖 `ChainClient` 接口 |
| 数据库表结构 | ❌ 无变化 | 每条记录已有 `chain_type` 字段 |
| 配置文件 | ✅ 改 Primary | 改一行 YAML 即可切换 |
| 历史数据验证 | ❌ 无变化 | 按记录上的 `chain_type` 回查原链 |
| 新链 Adapter | ✅ 需实现 | 单文件单 Package，实现 `ChainClient` 接口 |
| 已有链 | ❌ 不下线 | 并存运行，互不干扰 |

### 11.4 不可能做的事

- ❌ **跨链迁移账本数据**：不同链的账本无法互通，历史存证不可搬家
- ❌ **一键切换所有历史数据**：切换只影响新数据，旧数据永远在原链上

---

## 十二、风险与注意事项

### 12.1 技术风险

| 风险 | 对策 |
|------|------|
| **哈希算法不统一** | 全项目统一 SHA256，禁用国密 SM3（避免跨链兼容问题） |
| **JSON 序列化差异** | 上链前做 canonical 形式（字段排序、空格规范），避免验证时哈希对不上 |
| **tx_hash 丢失** | `tx_hash` 必须 `NOT NULL`（上链成功后）；链数据只能通过它回查 |
| **本地时间不可信** | 使用链上 `block_time`，不用 `time.Now()` |
| **FISCO 节点故障** | 至少部署 4 个共识节点，跨 AZ 部署；关键业务开双写 |
| **证书 / 私钥管理** | 使用 Vault 或 KMS，不写入 git |

### 12.2 合规风险

| 风险 | 对策 |
|------|------|
| 发行提货卡被认定为代币 | **禁止二级市场**、禁止任何形式的流通 / 交易 |
| 数藏业务监管收紧 | 阶段三 B 才真正接入，到时用合规 BaaS |
| 跨境业务 | 不使用海外公链，不在境外节点留存数据 |
| 个人信息上链 | **只上哈希，不上原始数据**；原始数据留在 MySQL 加密存储 |

### 12.3 运维风险

| 风险 | 对策 |
|------|------|
| 链节点宕机 | 部署监控告警（Prometheus + Grafana） |
| 账本膨胀 | 定期归档；非核心存证可配置 TTL（FISCO 支持） |
| 升级兼容 | 锁定 FISCO 大版本，跨版本升级走灰度 |

---

## 十三、成本评估

### 13.1 阶段一（商业化首发版 FISCO）一年成本估算

| 项目 | 配置 | 年成本（元） |
|------|------|--------------|
| 云服务器 | 4C8G × 2 台（链节点 + 应用节点） | ~4,000 |
| 存储 | 200 GB SSD | ~600 |
| 带宽 | 按量 5 Mbps | ~2,000 |
| 监控 / 日志 | 轻量自建 | ~500 |
| **合计** | | **≈ 7,000 元** |

> 对比：蚂蚁链同等业务量 ≈ **12 万元/年**，差距 **17 倍**。

### 13.2 阶段三 A（叠加至信链存证）增量成本

- 假设月存证 1 万条（只对法律敏感业务）
- 单价 ~0.5 元/条（具体看合同）
- 月成本 ≈ 5,000 元，年 ≈ 6 万元

### 13.3 阶段三 B（蚂蚁链 / 长安链 BaaS）

> **备注**：当前提货卡场景通过 FISCO 运行。蚂蚁链 Adapter 已在代码中实现，当营收支撑时可切换，届时费用取决于蚂蚁链合同条款。

- 蚂蚁链：9,900 元/月起，年 ≥ 12 万
- 长安链 BaaS（腾讯云 / 华为云）：3,000~8,000 元/月，年 ≈ 4~10 万

---

## 十四、TODO / 下一步

### 已完成

- [x] 确定首个区块链场景：提货卡链上发放（初期 FISCO，蚂蚁链作为升级路径）
- [x] 实现 `pkg/antchain` 蚂蚁链客户端（Client 接口 + httpClient + disabledClient + MockClient）——升级路径预留
- [x] 实现 `pkg/digitalcardmint` 核心服务（任务状态机 / MQ 派发 / 链上铸造 / 超时补偿）
- [x] 实现资产管理服务（会员资产查询 / 审计列表 / 合规处置）
- [x] 在 `rpc/sms` 中集成 CardMintAdminService gRPC 服务
- [x] 实现 admin-api 链路管理（查询/重试/冻结/升级）
- [x] 实现 admin-api 资产合规管理（审核/下线/回收）
- [x] 实现 front-api 会员提货卡查看（列表/详情/时间线）
- [x] 实现抽卡活动系统（资格校验/概率抽奖/库存扣减/发放触发）
- [x] 实现 consumer MQ 消费者（MintRequested → ExecuteTask）
- [x] 实现 job 定时任务（HandleCardMintTimeout → ScanDueTasks）
- [x] 编写单元测试（service_test.go / asset_service_test.go / digital_card_*_logic_test.go）
- [x] 创建数据表（sms_card_mint_task / sms_card_instance / sms_card_asset_log / sms_draw_* 系列）
- [x] sms-rpc / consumer / job 三个服务均配置链接口参数（FISCO 默认 + AntChain 预留）

### 近期优化

- [ ] 補充链上回执对账机制（定时批量 QueryMintToken 确认链上结果）
- [ ] 为 `sms_card_mint_task` 补充索引优化（`idx_status_retry`, `idx_idempotency`）
- [ ] 补充链路监控告警（失败率 / 滞留任务数 / manual_review 积压）
- [ ] 补充端到端测试：抽卡参与 → MQ 消费 → 链上铸造 → 会员查看
- [ ] 添加操作日志中间件记录链路管理操作

### 中期扩展（升级蚂蚁链）

- [ ] 当营收支撑蚂蚁链费用时，将 `Primary` 切换为 `antchain`
- [ ] 经过双写过渡期验证蚂蚁链链路稳定性
- [ ] 下线 FISCO 链路（保留只读回查，用于历史数据验证）

### 长期

- [ ] 引入至信链司法增强存证（电子合同、争议订单）
- [ ] 形成 zero-admin 内部的“区块链服务 SDK”供其他业务模块接入
- [ ] 评估是否将区块链服务独立为 `bcs-rpc` 微服务
- [ ] 探索更多链提供商（长安链 / BSN-DDC）以进一步降低成本或增强合规性

---

## 附录 A：术语表

| 术语 | 说明 |
|------|------|
| **BaaS** | Blockchain as a Service，云厂商提供的区块链托管服务 |
| **联盟链** | 参与节点受限的区块链，适合企业场景 |
| **公链** | 任意节点可加入的区块链（如以太坊、比特币） |
| **存证** | 将数据哈希写入链，作为不可篡改证据 |
| **采信** | 法院 / 仲裁机构接受链上证据作为合法证据 |
| **DDC** | Distributed Digital Certificate，分布式数字凭证，BSN 的 NFT 规范 |
| **Adapter** | 适配器模式，本文指每条链对应的实现类 |
| **切换点（Cut-over）** | 从旧方案切到新方案的时间点 |
| **MintToken** | 链上铸造 / 发放提货卡，初期通过 FISCO，升级后可切换为蚂蚁链 BaaS |
| **GovernanceScope** | 多租户治理范围（platform/tenant/merchant） |
| **状态机** | 发放任务的生命周期转换（pending_dispatch → ... → succeeded/frozen） |
| **ChainClient** | 抽象的链客户端接口 |
| **Registry** | 链注册表 |

---

## 附录 B：相关 Story 索引

| Story | 主题 | 状态 |
|-------|------|------|
| Story 10.4 | 蚂蚁链 token 发放与补偿（任务表 + consumer + job + 三层幂等） | 已上线 |
| Story 10.5 | 资产展示与审计检索 | 已上线 |
| Story 10.10 | 提货卡后台配置中心（商业化版） | 已上线 |
| Story 10.11 | FISCO BCOS 3.x 真实上链商业化接入（替换免费链 mock 底座） | 已上线 |

---

## 附录 C：FISCO BCOS 3.x 私钥与证书轮换 SOP

> **适用范围**：Story 10.11 落地后 sms-rpc / consumer / job 三服务共用的 FISCO BCOS 3.x 签名私钥（`/opt/fisco/keys/sms-rpc.key`）和 SDK mTLS 证书（`/opt/fisco/sdk/{ca.crt,sdk.crt,sdk.key}`）的运维操作规范。

### C.1 触发场景

| 场景 | 触发条件 | 处理优先级 |
|------|----------|-----------|
| 例行轮换 | 每 12 个月一次（与 SDK CA 默认有效期 825 天保持冗余）| 中 |
| 怀疑泄露 | 私钥文件权限被改 / 服务器被入侵 / 私钥误传 git | **紧急** |
| 证书过期 | sdk.crt 还剩 30 天到期，运维巡检告警 | 高 |
| 合约迁移 | grantMinter 给新合约后旧私钥不再使用 | 低（直接废弃即可） |

### C.2 私钥轮换标准流程（零停机）

**前提**：Story 10.11 的 ChainClient 已支持配置热加载；如未支持，需要先重启服务（< 1 分钟窗口）。

1. **生成新私钥**（在堡垒机或离线机器上，**不要在生产服务器**）：
   ```bash
   openssl ecparam -name secp256k1 -genkey -noout -out sms-rpc.key.new
   chmod 0400 sms-rpc.key.new
   ```
2. **派生新地址**并记录：
   ```bash
   # 用 console2 或 ethers.js 都可以
   ./console2.sh getAddressFromKey sms-rpc.key.new
   # 输出: 0xNEW_ADDR
   ```
3. **on-chain 授权**：用现役私钥（旧私钥仍是 minter）调合约 `grantMinter(0xNEW_ADDR)`：
   ```bash
   ./console2.sh call CardToken 0xCONTRACT grantMinter 0xNEW_ADDR
   ```
4. **分发新私钥**到生产服务器（**通过受控渠道**，禁止 plain SCP，建议加密 zip + 单独信道传密码）：
   ```bash
   # 服务器上替换
   mv /opt/fisco/keys/sms-rpc.key /opt/fisco/keys/sms-rpc.key.bak.$(date +%s)
   install -m 0400 -o root sms-rpc.key.new /opt/fisco/keys/sms-rpc.key
   ```
5. **重启 sms-rpc / consumer / job**（按部署顺序）。
6. **验证**：观察日志首行 `buildChainClient[*]: chainType="fisco_bcos_3x" ...`，等 30 秒看自循环扫描日志，确认有新交易上链且 `from = 0xNEW_ADDR`。
7. **on-chain 撤销旧地址**：
   ```bash
   ./console2.sh call CardToken 0xCONTRACT revokeMinter 0xOLD_ADDR
   ```
8. **物理销毁旧私钥**：`shred -uvz /opt/fisco/keys/sms-rpc.key.bak.*`（保险柜留存的纸质副本除外）。

### C.3 证书（CA / SDK）轮换流程

> 证书轮换只影响 mTLS 通道，不影响链上数据。

1. 在 FISCO 节点宿主机上重新生成证书：
   ```bash
   cd /opt/fisco/build_chain
   ./build_chain.sh -l "127.0.0.1:4" -p 30300,20200,8545 -T -c
   # 注意 -c 表示重新生成证书但保留原 nodeID/group
   ```
2. 把新的 `nodes/127.0.0.1/sdk/{ca.crt,sdk.crt,sdk.key}` 拷贝到 `/opt/fisco/sdk/`（保留旧版本到 `/opt/fisco/sdk.bak/`）：
   ```bash
   install -m 0400 -o root nodes/127.0.0.1/sdk/ca.crt /opt/fisco/sdk/ca.crt
   install -m 0400 -o root nodes/127.0.0.1/sdk/sdk.crt /opt/fisco/sdk/sdk.crt
   install -m 0400 -o root nodes/127.0.0.1/sdk/sdk.key /opt/fisco/sdk/sdk.key
   ```
3. 重启 sms-rpc / consumer / job。
4. 验证：日志没有 `tls_handshake_failed` 错误，新 mint 交易能正常上链。

### C.4 紧急吊销流程（怀疑泄露）

> **窗口目标**：从发现泄露到完成吊销 ≤ 30 分钟。

```
T+0    钉钉告警 / 安全事件单进
T+5    SRE on-call 用 grantMinter 给临时新私钥（步骤 C.2.1-3）
T+15   把临时新私钥分发到三服务、重启
T+20   on-chain revokeMinter 旧地址
T+25   验证旧地址不再能 mint（用旧私钥手动签名调 mint，应得 "AccessControl: missing role"）
T+30   写故障报告，提 PR 把私钥从可能泄露的位置擦除
```

### C.5 审计要求

每次轮换后必须留下：
- on-chain `grantMinter` / `revokeMinter` 两笔 tx 的 hash（区块浏览器永久可查）
- `/opt/fisco/keys/` 下 `ls -la` 的截图（证明新文件权限 0400 + 旧文件已删除）
- 三服务重启后的首屏 buildChainClient 日志截图

把以上证据归档到 `_opcos/operations-evidence/fisco-key-rotation-YYYY-MM-DD.md`。

### C.6 当前不在范围

下列内容由 Story 10.13+ 负责，**本 SOP 不覆盖**：
- 私钥托管到 KMS / Vault
- 多签名（multi-sig）minter 合约
- HSM 硬件模块
- 跨可用区 / 跨机房的私钥分发

---

## 附录 D：合约源码仓库（git submodule）

Story 10.11 / Review H2 处置：合约源码仓库 `helayD/zero-admin-contracts` 作为 git submodule 挂载在主仓 `contracts/` 路径下，**与主仓双向版本控制对齐**。

### D.1 Clone 主仓时同步拉取合约源码

```bash
# 推荐：clone 时一次性拉全
git clone --recursive https://github.com/helayD/zero-admin.git

# 已有 clone：补拉 submodule
git submodule update --init --recursive
```

### D.2 升级 contracts 到最新 main

```bash
git submodule update --remote contracts
git add contracts && git commit -m "chore: bump contracts submodule to <new-hash>"
git push
```

升级前必须先把 contracts 仓的新 commit `git push origin main` 到独立 GitHub 仓，主仓只记录 commit hash 指针。

### D.3 在 contracts 内部开发

```bash
cd contracts
# 在 main 分支正常开发，提交并 push 到独立仓
git checkout main && git pull
# ...修改 src/CardToken.sol...
git add . && git commit -m "feat: ..." && git push
# 回到主仓，更新 submodule 指针
cd .. && git add contracts && git commit -m "chore: bump contracts to <hash>"
```

### D.4 ABI 同步约束（Story 10.11 / Review L4）

每次合约变更都必须同步刷新主仓 `pkg/fisco/contracts/CardToken.abi`，防止运行时 method selector 与合约实际不符。建议把以下校验加入 CI（见 Story 10.11 Review Follow-up L4）：

```bash
forge inspect --root contracts CardToken abi > /tmp/abi.json
diff <(jq -S . /tmp/abi.json) <(jq -S . pkg/fisco/contracts/CardToken.abi) || \
  { echo "CardToken.abi 与合约源码不一致"; exit 1; }
```

## 变更记录

| 日期 | 版本 | 变更 | 作者 |
|------|------|------|------|
| 2026-04-18 | v1.0 | 初稿，基于蚂蚁链励退后的调研 | 九克城技术团队 |
| 2026-04-23 | v2.0 | 对齐实际代码实现：新增「当前实现概述」章节，重写架构/数据模型/代码结构，更新 TODO | 九克城技术团队 |
| 2026-05-10 | v2.1 | Story 10.11 落地：新增附录 B 相关 Story 索引，新增附录 C FISCO BCOS 3.x 私钥与证书轮换 SOP | 九克城技术团队 |
| 2026-05-11 | v2.2 | Story 10.11 Review H2 处置：新增附录 D 合约源码 submodule 工作流（contracts/ → helayD/zero-admin-contracts） | 九克城技术团队 |
