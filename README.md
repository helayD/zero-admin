# 将会在flutter星球更新getx版本的flutter_mall  🎉🎉🎉
# 将会在element plus星球更新element plus版本的的zero-vue-admin  🎉🎉🎉

# Zero-Admin 电商系统

> 注：ORM持久层已经整体切换成gorm
> 后端接口改动较大,前端正在重新适配中

> 基础代码生成工具

## goctl-helper is out! 🎉🎉🎉
**goland可视化插件，基于database生成api和protobuf文件(用于goctl官方插件使用)**

[goctl-helper](https://plugins.jetbrains.com/plugin/25693-goctl-helper)

## 安装
```shell
go install github.com/feihua/generate-code@latest

generate-code golang zero --dsn "root:123456@tcp(127.0.0.1:3306)/demo" --tableNames sys_ --prefix sys_  --rpcClient sysclient --author liufeihua
```
Zero-Admin 是一套基于 go-zero 框架实现的电商系统，采用 Docker 容器化部署，包含前台商城系统和后台管理系统。


## 前台商城系统

### 模块介绍

1. **首页门户**: 提供用户访问网站的入口，展示热门商品和推荐信息。

2. **商品推荐**: 根据用户的历史行为和个人喜好，推荐个性化商品。

3. **商品搜索**: 强大的商品搜索功能，支持关键字搜索、筛选等。

4. **商品展示**: 以优雅的方式展示商品信息，包括详细描述、价格、评价等。

5. **购物车**: 用户可以将喜欢的商品添加到购物车，方便批量购买。

6. **订单流程**: 提供完整的订单流程，包括下单、支付、发货、收货等环节。

7. **会员中心**: 用户可以管理个人信息、查看订单状态、积分等。

8. **帮助中心**: 提供用户常见问题解答、售后政策等信息。

### 技术栈

- go-zero 框架实现，高性能、易扩展。
- 前端采用现代化的前端框架，例如 React 或 Vue。
- Docker 容器化部署，方便快捷。

## 后台管理系统

### 模块介绍

1. **商品管理**: 管理商品信息，包括添加、编辑、删除商品。

2. **订单管理**: 实时监控订单状态，支持订单发货、取消等操作。

3. **会员管理**: 管理用户信息，包括注册用户、会员等级等。

4. **促销管理**: 管理营销活动，例如满减、打折等。

5. **运营管理**: 管理广告、推广等运营活动。

6. **内容管理**: 管理网站内容，包括公告、资讯等。

7. **权限管理**: 管理系统用户权限，确保安全性。

8. **设置**: 系统配置，包括支付方式、物流信息等。

### 技术栈

- go-zero 框架提供后台接口支持。
- 使用现代化的前端框架进行界面开发。
- 数据库采用 mysql。
- Docker 容器化部署，方便管理和维护。

## 提货卡产品本质（业务复盘文档）

> 记录于 2026-05-10，Story 10.10 完成后整理。本节用于团队后续复盘"提货卡"业务形态的产品定位与代码映射。

### 一句话定义

**提货卡 = 链上数字资产 + 链下实物提货权**。本质等价于"链上 NFT + 兑换券"的组合：链上保证所有权可信流转，链下按持卡人意愿兑换实物。

### 三层能力对应代码

| 业务能力 | 实现机制 | 表 / 服务 |
|---|---|---|
| **数字资产（=上链）** | 每发一张卡都上链存证 token，所有权变更可审计 | `sms_card_instance.chain_tx_id / token_id` |
| **可转让（=点对点）** | 短链/二维码 token，扫一下卡就转 | `sms_card_claim_token`（一次性 claim token） |
| **线下交易 / 提货** | 持卡人发起提货 → 仓库出实物 | `sms_card_redemption_order` → `sms_digital_card_physical_fulfillment` |

### 链上 / 链下分层架构

```text
┌─────────────────────────────────────────┐
│ 链上层（蚂蚁链 / FISCO BCOS）             │
│   - tokenId、chainTxId、上链回执          │
│   - 不可篡改的所有权变更轨迹              │
└─────────────────────────────────────────┘
              ↑↓ 双写同步
┌─────────────────────────────────────────┐
│ 链下层（MySQL / 业务系统）                │
│   - sms_card_instance.member_id 当前持卡人│
│   - sms_card_redemption_order 兑换单      │
│   - sms_digital_card_physical_fulfillment │
└─────────────────────────────────────────┘
              ↑↓ 业务读写
┌─────────────────────────────────────────┐
│ 应用层                                    │
│   B 端运维: 看链 + 业务（全字段）          │
│   C 端用户: 只看业务（卡号/状态/有效期）   │ ← 监管约束
└─────────────────────────────────────────┘
```

> **C 端监管约束**：Flutter（C 端）严禁展示 `chainType` / `chainTxId` / `tokenId` / `lastReceiptJson` 等链上字段及"区块链/上链/链路/AntChain/FISCO"等关键词文案。详见 `AGENTS.md` 的"C 端监管约束（提货卡）"章节。B 端 Web Admin 不受此约束，运维需要看完整链信息进行审计。

### "可转让"的合规边界

代码当前支持 **点对点转赠**，**不支持** 公开二级市场。原因：

- 代币化二级市场在国内是监管红线（NFT 炒作 / 违规交易）
- 提货卡作为"提货凭证"性质，转赠属于民法典中的**赠与合同**或**债权转让**，合规
- 一旦做成自由竞价交易，性质就变了

`sms_product_fulfillment_rule` 通过 `transferable` 与 `transfer_limit` 字段精细控制：

| 配置 | 业务含义 | 适用场景 |
|---|---|---|
| `transferable=0` | 完全不可转 | 实名提货卡，类似车票 |
| `transferable=1, transfer_limit=1` | 可送一次 | 礼品场景（最常见） |
| `transferable=1, transfer_limit=N` | 多次转赠 | 流通性最高，需更严合规审查 |

### "线下交易"的两种形态

| 形态 | 系统支持 | 说明 |
|---|---|---|
| **持卡人提货** | ✅ `redemption_order → physical_fulfillment` | 标准链路 |
| **二手买卖→受让方提货** | ✅ `claim_token` 转赠 → 受让方再发起提货 | 系统不感知交易对价（用户私下结算） |
| **在售提货卡的二级市场报价** | ❌ 不做 | 合规风险，不在代码范围内 |

### 与传统电商对比

| 维度 | 传统电商（physical_delivery） | 提货卡电商（digital_asset） |
|---|---|---|
| 下单后 | 仓库直接发货 | 用户拿到一张卡 |
| 用户能做什么 | 等收货 | 可送人 / 可延期提 / 可凭卡线下提 |
| 适用场景 | 日用品、生鲜 | 礼品（生日/年节）、预售（期货）、限定纪念品、商务礼包 |
| 库存压力 | 立即出库 | 卡可发出但货可分批入库 |
| 防黄牛 | 弱 | 强（卡可设转赠次数 + 实名提货） |

### 关键数据模型

| 表 | 角色 |
|---|---|
| `pms_product_spu.fulfillment_mode` | 商品履约模式：`physical_delivery` / `digital_asset` |
| `pms_product_spu.fulfillment_rule_id` | 关联发卡规则 ID（仅 digital_asset 模式有效） |
| `sms_card_template` | 卡片模板（卡面、稀有度、版权、发行上限等元数据） |
| `sms_product_fulfillment_rule` | 发卡规则（关联模板 + 兑换/转赠/退款条件） |
| `sms_card_instance` | 用户名下的卡片实例（asset_no、mint_status、asset_status） |
| `sms_card_claim_token` | 转赠/领取一次性 token |
| `sms_card_redemption_order` | 兑换提货生成的子单 |
| `sms_digital_card_physical_fulfillment` | 实物履约单（仓库 / 物流） |

### 完整业务流（购买 → 提货）

```text
1. 运营预先配置:
   卡片模板 (sms_card_template) + 发卡规则 (sms_product_fulfillment_rule)
   ↓ 关联到商品
   pms_product_spu.fulfillment_mode='digital_asset' + fulfillment_rule_id

2. 用户下单 + 支付完成:
   oms 触发发卡 → sms_card_instance 创建一条记录
   member_id = 用户 / asset_status='owned' / mint_status='success'
   同步上链 → 写回 chain_tx_id / token_id

3. 用户在 C 端"我的卡包"看到卡:
   只展示: 卡号、模板名、获取时间、状态、有效期 (符合监管约束)

4. 用户操作（任选）:
   ├── 转赠 → sms_card_claim_token 生成 token → 受让方扫码领取
   └── 提货 → sms_card_redemption_order (填地址) → sms_digital_card_physical_fulfillment (仓库出库)

5. 完成:
   sms_card_instance.asset_status='redeemed'
   卡变"已兑换"，C 端归档展示
```

### Story 10.10 在产品形态升级中的位置

把卡片模板 / 发卡规则从硬编码 mock 数据 **变成可后台维护的元数据**，意味着运营可以：

- 给手机配 SSR 稀有度卡（限量纪念）
- 给家电配 N 普通卡（实用提货凭证）
- 给礼包配可转赠 3 次的卡（社交流转）
- 给奢侈品配实名不可转赠卡（合规防伪）

**配置即业务**，不再写代码。这是从"卖货"到"卖凭证"的产品形态升级关键里程碑。

### 后续演进方向

- Story 10.11 — SKU 级履约模式覆盖（同一 SPU 下不同 SKU 用不同卡）
- Story 10.12 — 发卡规则下钻已绑定商品列表（运营回查"这条规则关联了哪些商品"）
- Story 10.13（建议）— pms-rpc 协议加 `fulfillment_mode` 字段，履约模式过滤走 SQL 不走内存
- 卡片模板字段清空（M2 follow-up）— PB 协议加 `clear_<field> bool` 标志或改 `*string`

# 文档地址
[https://feihua.github.io/](https://feihua.github.io/) 正在完善
#
[zero-admin-ui是后台的pc管理端](https://github.com/feihua/zero-admin-ui)是一个基于react实现的管理后台

[flutter_mall是zero-admin的app端](https://github.com/feihua/flutter_mall)是一个Flutter的电商实战项目，包括首页、列表页、详细页、购物车页、会员中心和支付(支付对接的是支付宝)

[zero-pc-web 是 zero-admin 的网页端](https://github.com/feihua/zero-pc-web)zero-pc-web 是一个基于 React 框架实现的 web
端电商系统(预览地址[http://129.204.203.29/pc/](http://129.204.203.29/pc/))



# android版本
android版本体验地址 [flutter-mall-app](https://www.pgyer.com/OoW2Zy)


# 项目模板
[zero-admin-template](https://github.com/feihua/zero-admin-template)(只包含基础的rbac权限)

# 1.项目预览

**预览地址**http://129.204.203.29/mall <span  style="color: red;"> 账号：admin 密码: 123456</span>

**vue版本预览地址**http://129.204.203.29/vue/login <span  style="color: red;"> 账号：admin 密码: 123456</span>

**web端**预览地址[http://129.204.203.29/pc/](http://129.204.203.29/pc/)
> 注：演示账号部分功能修改删除权限未开放。



# 2.感谢
[go-zero](https://github.com/zeromicro/go-zero)
<p></p>

[mall](https://github.com/macrozheng/mall)

## 许可证

本项目采用 Apache License 2.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

如果您觉得这个项目对您有帮助，请给我们一个 ⭐️，这将鼓励我们持续改进！
