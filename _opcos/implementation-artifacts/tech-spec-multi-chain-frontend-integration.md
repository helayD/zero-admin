---
title: '多链 Adapter 三端联动 — chainType 字段透出'
slug: 'multi-chain-frontend-integration'
created: '2026-04-23T16:30:00+0800'
status: 'ready-for-dev'
stepsCompleted: [1, 2, 3, 4]
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
  # Flutter
  - 'flutter-mall/lib/model/digital_card/digital_card_asset_model.dart (DigitalCardAssetItem 增加 chainType)'
  - 'flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart (_buildStatusCard 增加链类型行)'
---

# Tech-Spec: 多链 Adapter 三端联动 — chainType 字段透出

**Created:** 2026-04-23T16:30:00+0800  
**Depends on:** tech-spec-multi-chain-adapter-refactor (已完成)

## Overview

### Problem Statement

多链 Adapter 架构重构已完成，`Service.Chain` 现在是 `chainclient.ChainClient` 接口，支持 FISCO BCOS 和 AntChain。但当前 API 响应和前端 UI 均不显示资产/任务所使用的链类型（chainType），运维和用户无法在界面上区分。

### Solution

从核心 Service 层开始，在 `TaskListItem`、`ActionResult`、审计查询结果等结构体中新增 `ChainType string` 字段，通过 admin-api / front-api 的 types 透传到前端，最终在 Web Admin 的链路监控和资产工作台、Flutter 的卡片详情中展示。

### Scope

**In Scope:**

- 后端: `pkg/digitalcardmint` 结构体增加 `ChainType` 字段并从 `Service.Chain.ChainType()` 填充
- 后端: admin-api 和 front-api types + logic 透传 `chainType`
- 后端: `.api` 文件同步更新（仅文档作用，不运行 goctl）
- Web Admin: 链路监控表格列和详情抽屉、资产工作台详情抽屉显示链类型
- Flutter: 卡片详情页"状态与标识"区域显示链类型

**Out of Scope:**

- 链配置管理页面（Blockchain.Primary 切换等管理后台功能）
- 链类型筛选/搜索（本期只做展示，不做筛选条件）
- 数据库表结构变更（chainType 从运行时 Service.Chain.ChainType() 获取，不持久化）

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

### Phase 5: Flutter

#### Task 12: Model + 详情页

**digital_card_asset_model.dart:**
- `DigitalCardAssetItem` 增加 `final String chainType;`
- 构造函数和 `fromJson` 增加 `chainType`

**digital_card_asset_detail_page.dart:**
- `_buildStatusCard` 在"链上状态"行后增加：
```dart
_buildMetaLine('链类型', detail.item.chainType.isEmpty ? '-' : detail.item.chainType),
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

- [ ] `go build` 涉及模块全部通过
- [ ] `pkg/digitalcardmint` 测试全部 PASS
- [ ] `go vet` 零错误
- [ ] Web Admin TypeScript 类型检查通过
- [ ] Flutter analyze 无新增错误
- [ ] admin-api 返回的 chainItem JSON 包含 `"chainType":"fisco"` 或 `"chainType":"antchain"`
- [ ] Web Admin 链路监控表格显示"链类型"列
- [ ] Web Admin 链路详情抽屉显示链类型
- [ ] Flutter 卡片详情"状态与标识"显示链类型
