---
title: '多链 Adapter 三端联动 — chainType 字段透出'
slug: 'multi-chain-frontend-integration'
created: '2026-04-23T16:30:00+0800'
status: 'done'
stepsCompleted: [1, 2, 3, 4, 5, 6]
tech_stack: ['Go 1.25', 'go-zero 1.9.3', 'GORM', 'React (Ant Design Pro)', 'Flutter (Dart)']
depends_on: 'tech-spec-multi-chain-adapter-refactor'
files_to_modify:
  # 后端 — pkg 核心
  - 'pkg/digitalcardmint/service.go (TaskListItem / ActionResult 等结构体增加 ChainType)'
  - 'pkg/digitalcardmint/asset_service.go (DigitalCardAssetAuditItem / MemberDigitalCardAssetItem 增加 ChainType)'
  # 后端 — admin-api types + logic
  - 'api/admin/internal/types/digital_card_chain.go (DigitalCardChainItem / DigitalCardChainDetailData / DigitalCardChainActionResp 增加 chainType)'
  - 'api/admin/internal/types/digital_card_asset.go (DigitalCardAssetItem / DigitalCardAssetMintTaskSummary 增加 chainType)'
  - 'api/admin/internal/logic/sms/digital_card_chain/helper.go (mapTaskItem 透传 chainType)'
  - 'api/admin/internal/logic/sms/digital_card_asset/helper.go (mapAuditItem 透传 chainType)'
  - 'api/admin/doc/api/sms/digital_card_chain.api (DigitalCardChainItem 增加 chainType 字段定义)'
  - 'api/admin/doc/api/sms/digital_card_asset.api (DigitalCardAssetItem / MintTaskSummary 增加 chainType 字段定义)'
  # 后端 — front-api types + logic
  - 'api/front/doc/api/digital_card/digital_card_asset.api (DigitalCardAssetItem 增加 chainType)'
  - 'api/front/internal/types/digital_card_asset.go (DigitalCardAssetItem 增加 chainType)'
  - 'api/front/internal/logic/digital_card/digital_card_asset/helper.go (mapAssetItem 透传 chainType)'
  # Web Admin
  - 'web-admin/src/pages/sms/DigitalCardChainMonitor/data.ts (DigitalCardChainItem 增加 chainType)'
  - 'web-admin/src/pages/sms/DigitalCardChainMonitor/index.tsx (表格列增加"链类型")'
  - 'web-admin/src/pages/sms/DigitalCardChainMonitor/components/ChainDetailDrawer.tsx (详情抽屉增加链类型)'
  - 'web-admin/src/pages/sms/DigitalCardAssetWorkbench/data.ts (DigitalCardAssetItem 增加 chainType)'
  - 'web-admin/src/pages/sms/DigitalCardAssetWorkbench/components/AssetDetailDrawer.tsx (详情抽屉增加链类型)'
  # Flutter (注意：C 端不展示链相关信息，只做 model 字段透传，不在 UI 上显示)
  - 'flutter-mall/lib/model/digital_card/digital_card_asset_model.dart (DigitalCardAssetItem 增加 chainType，仅 model 层)'
  - 'flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart (移除链上状态/链类型/token 展示)'
  - 'flutter-mall/lib/view/digital_card/digital_card_asset_tile.dart (替换 chainStatusText 为 displayStatusText)'
---

# Tech-Spec: 多链 Adapter 三端联动 — chainType 字段透出

**Created:** 2026-04-23T16:30:00+0800  
**Depends on:** tech-spec-multi-chain-adapter-refactor (已完成)

## Overview

### Problem Statement

多链 Adapter 架构重构已完成，`Service.Chain` 现在是 `chainclient.ChainClient` 接口，支持 FISCO BCOS 和 AntChain。但当前 API 响应和前端 UI 均不显示资产/任务所使用的链类型（chainType），运维无法在界面上区分。

### Solution

从核心 Service 层开始，在 `TaskListItem`、`ActionResult`、审计查询结果等结构体中新增 `ChainType string` 字段，通过 admin-api / front-api 的 types 透传到前端；Web Admin 的链路监控和资产工作台展示链类型，Flutter C 端仅在 model 层透传，UI 不展示底层链信息。

### Scope

**In Scope:**

- 后端: `pkg/digitalcardmint` 结构体增加 `ChainType` 字段并从 `Service.Chain.ChainType()` 填充
- 后端: admin-api 和 front-api types + logic 透传 `chainType`
- 后端: `.api` 文件同步更新（仅文档作用，不运行 goctl）
- Web Admin: 链路监控表格列和详情抽屉、资产工作台详情抽屉显示链类型
- Flutter: model 层透传 `chainType`，卡片列表/详情 UI 移除链类型、链上状态和 token 展示

**Out of Scope:**

- 链配置管理页面（Blockchain.Primary 切换等管理后台功能）
- 链类型筛选/搜索（本期只做展示，不做筛选条件）
- 数据库表结构变更（chainType 从运行时 Service.Chain.ChainType() 获取，不持久化）

### 监管约束（强制）

> **C 端（Flutter / front-api 面向普通用户的 UI）严禁展示任何区块链底层信息。**
>
> 以下字段/文案 **禁止在 C 端 UI 上直接展示**：
> - `chainType`（链类型：蚂蚁链 / FISCO BCOS 等）
> - `chainStatus` / `chainStatusText`（链上状态）
> - `chainTxId`（链上交易 ID）
> - `tokenId` / `tokenIdMasked`（token 编号）
> - `lastReceiptJson` / `lastReceiptSummary`（链上回执）
> - 任何包含“蚂蚁链”“AntChain”“FISCO”“区块链”“链上”“上链”“链路”等关键词的文案
>
> C 端只允许展示：卡片编号、模板名、活动名、获取时间、发放状态（mintStatusText）、展示状态、合规状态、合规提示摘要。
> 如果服务端返回的允许字段包含底层链路文案，Flutter UI 必须转换为发放/到账/处理结果口径。
>
> **B 端（Web Admin）不受此约束**，运维人员可以看到完整链信息。
>
> **原因**: 提货卡业务监管要求，C 端不得暴露底层区块链实现细节。

## Implementation Tasks

### Phase 1: 后端核心 — pkg/digitalcardmint 结构体

#### Task 1: `pkg/digitalcardmint/service.go` — TaskListItem 增加 ChainType

在 `TaskListItem` 结构体中增加 `ChainType string` 字段。

在 `QueryTaskList` 方法中，查询完成后对每个 item 填充 `s.Chain.ChainType()`。

在 `ActionResult` 结构体中增加 `ChainType string` 字段，所有返回 `ActionResult` 的方法填充 `s.Chain.ChainType()`。

在 `QueryTaskDetail` 返回的 `TaskDetailData` 中增加 `ChainType string`，填充 `s.Chain.ChainType()`。

#### Task 2: `pkg/digitalcardmint/asset_service.go` — 审计/会员资产结构体增加 ChainType

在 `DigitalCardAssetAuditItem` 和 `MemberDigitalCardAssetItem` 结构体中增加 `ChainType string`。

在 `QueryAssetAuditList`、`QueryAssetAuditDetail`、`QueryMemberAssetList`、`QueryMemberAssetDetail` 方法中填充 `s.Chain.ChainType()`。

### Phase 2: 后端 admin-api types + logic

#### Task 3: `api/admin/internal/types/digital_card_chain.go` — 增加 chainType 字段

- `DigitalCardChainItem` 增加 `ChainType string \`json:"chainType"\``
- `DigitalCardChainDetailData` 增加 `ChainType string \`json:"chainType"\``
- `DigitalCardChainActionResp` 增加 `ChainType string \`json:"chainType"\``

#### Task 4: `api/admin/internal/types/digital_card_asset.go` — 增加 chainType 字段

- `DigitalCardAssetItem` 增加 `ChainType string \`json:"chainType"\``
- `DigitalCardAssetMintTaskSummary` 增加 `ChainType string \`json:"chainType"\``

#### Task 5: `api/admin/internal/logic/sms/digital_card_chain/helper.go` — mapTaskItem 透传

`mapTaskItem` 函数增加 `ChainType: item.ChainType`。

`writeDigitalCardChainOperateLog` 的 payload 增加 `"chainType": result.ChainType`。

所有返回 `DigitalCardChainActionResp` 的 logic（retry/freeze/escalate）增加 `ChainType: result.ChainType`。

#### Task 6: `api/admin/internal/logic/sms/digital_card_asset/` — helper 透传

`mapAuditItem` 增加 `ChainType: item.ChainType`。

`QueryDigitalCardAssetDetailLogic` 的 `MintTaskSummary` 映射增加 `ChainType: detail.MintTaskSummary.ChainType`。

#### Task 7: `.api` 文件更新（仅文档）

- `api/admin/doc/api/sms/digital_card_chain.api` → `DigitalCardChainItem` 增加 `ChainType string \`json:"chainType"\``
- `api/admin/doc/api/sms/digital_card_asset.api` → `DigitalCardAssetItem` 和 `DigitalCardAssetMintTaskSummary` 增加 `ChainType`
- `api/front/doc/api/digital_card/digital_card_asset.api` → `DigitalCardAssetItem` 增加 `ChainType`

### Phase 3: 后端 front-api types + logic

#### Task 8: `api/front/internal/types/digital_card_asset.go` — 增加 chainType 字段

`DigitalCardAssetItem` 增加 `ChainType string \`json:"chainType"\``。

#### Task 9: `api/front/internal/logic/digital_card/digital_card_asset/helper.go` — mapAssetItem 透传

`mapAssetItem` 增加 `ChainType: item.ChainType`。

### Phase 4: Web Admin

#### Task 10: DigitalCardChainMonitor — 数据类型 + 表格列 + 详情抽屉

**data.ts:**
- `DigitalCardChainItem` 增加 `chainType: string`
- `DigitalCardChainDetailData` 增加 `chainType: string`
- `DigitalCardChainActionResp` 增加 `chainType: string`

**index.tsx:**
- 在 columns 的"任务状态"列前插入新列：
```tsx
{
  title: '链类型',
  dataIndex: 'chainType',
  width: 100,
  hideInSearch: true,
  render: (_, row) => (
    <Tag color={row.chainType === 'fisco' ? 'cyan' : row.chainType === 'antchain' ? 'purple' : 'default'}>
      {row.chainType || '-'}
    </Tag>
  ),
},
```

**ChainDetailDrawer.tsx:**
- Descriptions 中"资产编号"后增加：
```tsx
<Descriptions.Item label="链类型">
  <Tag color={item.chainType === 'fisco' ? 'cyan' : 'purple'}>
    {item.chainType || '-'}
  </Tag>
</Descriptions.Item>
```

#### Task 11: DigitalCardAssetWorkbench — 数据类型 + 详情抽屉

**data.ts:**
- `DigitalCardAssetItem` 增加 `chainType: string`
- `DigitalCardAssetMintTaskSummary` 增加 `chainType: string`

**AssetDetailDrawer.tsx:**
- "发放任务摘要" Descriptions 中增加：
```tsx
<Descriptions.Item label="链类型">
  <Tag color={detail?.mintTaskSummary?.chainType === 'fisco' ? 'cyan' : 'purple'}>
    {detail?.mintTaskSummary?.chainType || '-'}
  </Tag>
</Descriptions.Item>
```

### Phase 5: Flutter（C 端监管合规）

> **重要**: C 端不展示任何链相关信息，参见上方「监管约束」。

#### Task 12: Model 字段透传 + UI 链信息移除

**digital_card_asset_model.dart:**
- `DigitalCardAssetItem` 增加 `final String chainType;`（model 层保留，供内部逻辑使用）
- 构造函数和 `fromJson` 增加 `chainType`

**digital_card_asset_detail_page.dart — 移除链信息:**
- 删除「链上状态」、「链类型」、「脱敏 token」三行
- hero 区域 chip 从 `chainStatusText` 替换为 `displayStatusText`

**digital_card_asset_tile.dart — 移除链信息:**
- chip 从 `chainStatusText` 替换为 `displayStatusText`
- `_statusColor` 从 `chainStatus == 'success'` 改为 `mintStatus == 'mint_success'`

**禁止在 Flutter UI 中展示的字段:**
```
chainType, chainStatus, chainStatusText, chainTxId,
tokenId, tokenIdMasked, lastReceiptJson, lastReceiptSummary
```

### Phase 6: 验证

#### Task 13: 编译验证 + 测试

```bash
go build ./pkg/digitalcardmint/... ./api/admin/... ./api/front/...
go test ./pkg/digitalcardmint/... -count=1 -v
go vet ./pkg/digitalcardmint/... ./api/admin/... ./api/front/...
```

Web Admin: `cd web-admin && npx tsc --noEmit`（类型检查）

Flutter: `cd flutter-mall && flutter analyze`

## Verification Checklist

- [x] `go build` 涉及模块全部通过
- [x] `pkg/digitalcardmint` 测试全部 PASS (26 tests)
- [x] `go vet` 零新增错误
- [x] admin-api 返回的 chainItem JSON 包含 `"chainType":"fisco"` 或 `"chainType":"antchain"`
- [x] Web Admin 链路监控表格显示“链类型”列
- [x] Web Admin 链路详情抽屉显示链类型
- [x] Flutter 卡片详情 **不显示** 链类型/链上状态/token（监管合规）
- [x] Flutter 卡片列表/详情 chip 使用 displayStatusText 替代 chainStatusText
- [x] 远程部署全量 smoke 16/16 PASS
