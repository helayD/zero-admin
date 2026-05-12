---
name: digital-card-asset
description: "Skill for the Digital_card_asset area of zero-admin. 59 symbols across 25 files."
---

# Digital_card_asset

59 symbols | 25 files | Cohesion: 59%

## When to Use

- Working with code in `api/`
- Understanding how NewReviewDigitalCardAssetComplianceLogic, NewRecycleDigitalCardAssetLogic, TestReviewDigitalCardAssetComplianceRejectsCrossScopeWrite work
- Modifying digital_card_asset-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `api/front/internal/logic/digital_card/digital_card_asset/redemption_logic.go` | NewClaimDigitalCardLogic, NewCreateRedemptionOrderLogic, CreateRedemptionOrder, convertRedemptionOrderData, NewQueryRedemptionOrderLogic (+5) |
| `api/front/internal/logic/digital_card/digital_card_asset/asset_action_logic.go` | NewRequestDigitalCardWithdrawLogic, RequestDigitalCardWithdraw, NewResolveDigitalCardTransferRecipientLogic, ResolveDigitalCardTransferRecipient, NewTransferDigitalCardAssetLogic (+1) |
| `api/admin/internal/logic/sms/digital_card_asset/digital_card_asset_logic_test.go` | newAdminAssetCtx, newAdminAssetServiceContext, TestReviewDigitalCardAssetComplianceRejectsCrossScopeWrite, TestReviewThenRecycleDigitalCardAssetSuccess, TestQueryDigitalCardAssetListRespectsMerchantScope |
| `api/front/internal/handler/digital_card/digital_card_asset/redemption_handler.go` | ClaimDigitalCardHandler, CreateRedemptionOrderHandler, QueryRedemptionOrderHandler, GenerateShareLinkHandler, ValidateClaimTokenHandler |
| `api/front/internal/logic/digital_card/digital_card_asset/digital_card_asset_logic_test.go` | newFrontAssetCtx, newFrontAssetServiceContext, TestQueryMyDigitalCardAssetListOnlyReturnsCurrentMemberAssets, TestQueryMyDigitalCardAssetDetailRejectsOtherMemberAsset |
| `api/front/internal/handler/digital_card/digital_card_asset/asset_action_handler.go` | RequestDigitalCardWithdrawHandler, ResolveDigitalCardTransferRecipientHandler, TransferDigitalCardAssetHandler |
| `api/admin/internal/logic/sms/digital_card_asset/reviewdigitalcardassetcompliancelogic.go` | NewReviewDigitalCardAssetComplianceLogic, ReviewDigitalCardAssetCompliance |
| `api/admin/internal/logic/sms/digital_card_asset/querydigitalcardassetdetaillogic.go` | NewQueryDigitalCardAssetDetailLogic, QueryDigitalCardAssetDetail |
| `api/admin/internal/logic/sms/digital_card_asset/helper.go` | mapAuditItem, mapAssetLogs |
| `api/front/internal/logic/digital_card/digital_card_asset/query_my_digital_card_asset_list_logic.go` | NewQueryMyDigitalCardAssetListLogic, QueryMyDigitalCardAssetList |

## Entry Points

Start here when exploring this area:

- **`NewReviewDigitalCardAssetComplianceLogic`** (Function) — `api/admin/internal/logic/sms/digital_card_asset/reviewdigitalcardassetcompliancelogic.go:18`
- **`NewRecycleDigitalCardAssetLogic`** (Function) — `api/admin/internal/logic/sms/digital_card_asset/recycledigitalcardassetlogic.go:18`
- **`TestReviewDigitalCardAssetComplianceRejectsCrossScopeWrite`** (Function) — `api/admin/internal/logic/sms/digital_card_asset/digital_card_asset_logic_test.go:176`
- **`TestReviewThenRecycleDigitalCardAssetSuccess`** (Function) — `api/admin/internal/logic/sms/digital_card_asset/digital_card_asset_logic_test.go:194`
- **`ReviewDigitalCardAssetComplianceHandler`** (Function) — `api/admin/internal/handler/sms/digital_card_asset/reviewdigitalcardassetcompliancehandler.go:11`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewReviewDigitalCardAssetComplianceLogic` | Function | `api/admin/internal/logic/sms/digital_card_asset/reviewdigitalcardassetcompliancelogic.go` | 18 |
| `NewRecycleDigitalCardAssetLogic` | Function | `api/admin/internal/logic/sms/digital_card_asset/recycledigitalcardassetlogic.go` | 18 |
| `TestReviewDigitalCardAssetComplianceRejectsCrossScopeWrite` | Function | `api/admin/internal/logic/sms/digital_card_asset/digital_card_asset_logic_test.go` | 176 |
| `TestReviewThenRecycleDigitalCardAssetSuccess` | Function | `api/admin/internal/logic/sms/digital_card_asset/digital_card_asset_logic_test.go` | 194 |
| `ReviewDigitalCardAssetComplianceHandler` | Function | `api/admin/internal/handler/sms/digital_card_asset/reviewdigitalcardassetcompliancehandler.go` | 11 |
| `RecycleDigitalCardAssetHandler` | Function | `api/admin/internal/handler/sms/digital_card_asset/recycledigitalcardassethandler.go` | 11 |
| `RegisterExtraHandlers` | Function | `api/front/internal/handler/extra_routes.go` | 13 |
| `NewConfirmPhysicalFulfillmentAddressLogic` | Function | `api/front/internal/logic/digital_card/physical_fulfillment/confirm_physical_fulfillment_address_logic.go` | 17 |
| `NewClaimDigitalCardLogic` | Function | `api/front/internal/logic/digital_card/digital_card_asset/redemption_logic.go` | 197 |
| `ConfirmPhysicalFulfillmentAddressHandler` | Function | `api/front/internal/handler/digital_card/physical_fulfillment/confirm_physical_fulfillment_address_handler.go` | 11 |
| `DigitalCardRegisterHintHandler` | Function | `api/front/internal/handler/digital_card/digital_card_asset/register_hint_handler.go` | 80 |
| `ClaimDigitalCardHandler` | Function | `api/front/internal/handler/digital_card/digital_card_asset/redemption_handler.go` | 65 |
| `NewQueryDigitalCardAssetDetailLogic` | Function | `api/admin/internal/logic/sms/digital_card_asset/querydigitalcardassetdetaillogic.go` | 18 |
| `QueryDigitalCardAssetDetailHandler` | Function | `api/admin/internal/handler/sms/digital_card_asset/querydigitalcardassetdetailhandler.go` | 11 |
| `NewCreateRedemptionOrderLogic` | Function | `api/front/internal/logic/digital_card/digital_card_asset/redemption_logic.go` | 23 |
| `CreateRedemptionOrderHandler` | Function | `api/front/internal/handler/digital_card/digital_card_asset/redemption_handler.go` | 11 |
| `NewQueryMyDigitalCardAssetListLogic` | Function | `api/front/internal/logic/digital_card/digital_card_asset/query_my_digital_card_asset_list_logic.go` | 17 |
| `QueryMyDigitalCardAssetListHandler` | Function | `api/front/internal/handler/digital_card/digital_card_asset/query_my_digital_card_asset_list_handler.go` | 11 |
| `NewRequestDigitalCardWithdrawLogic` | Function | `api/front/internal/logic/digital_card/digital_card_asset/asset_action_logic.go` | 104 |
| `RequestDigitalCardWithdrawHandler` | Function | `api/front/internal/handler/digital_card/digital_card_asset/asset_action_handler.go` | 47 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `RegisterExtraHandlers → CodeError` | cross_community | 7 |
| `RegisterExtraHandlers → QueryMyDigitalCardAssetListLogic` | cross_community | 4 |
| `RegisterExtraHandlers → MapAssetItem` | cross_community | 4 |
| `RegisterExtraHandlers → QueryMyDigitalCardAssetDetailLogic` | cross_community | 4 |
| `RegisterExtraHandlers → MapDrawSummary` | cross_community | 4 |
| `RegisterExtraHandlers → MapTimeline` | cross_community | 4 |
| `RegisterExtraHandlers → ResolveDigitalCardTransferRecipientLogic` | cross_community | 4 |
| `RegisterExtraHandlers → TransferDigitalCardAssetLogic` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| User | 11 calls |
| Message | 8 calls |
| Physical_fulfillment | 4 calls |
| Digitalcardmint | 3 calls |
| Order | 2 calls |
| Middleware | 2 calls |
| Svc | 1 calls |

## How to Explore

1. `gitnexus_context({name: "NewReviewDigitalCardAssetComplianceLogic"})` — see callers and callees
2. `gitnexus_query({query: "digital_card_asset"})` — find related execution flows
3. Read key files listed above for implementation details
