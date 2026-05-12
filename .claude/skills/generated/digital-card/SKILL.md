---
name: digital-card
description: "Skill for the Digital_card area of zero-admin. 88 symbols across 34 files."
---

# Digital_card

88 symbols | 34 files | Cohesion: 60%

## When to Use

- Working with code in `flutter-mall/`
- Understanding how main, NewCardRedemptionOrderServiceClient, RedemptionRequested work
- Modifying digital_card-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart` | DigitalCardAssetDetailPage, _openPhysicalFulfillment, initState, _loadDetail, _requestWithdraw (+10) |
| `flutter-mall/lib/view/digital_card/digital_card_physical_fulfillment_page.dart` | DigitalCardPhysicalFulfillmentPage, _loadDetail, _confirmReceipt, _confirmShippingFee, _openAddressSheet (+3) |
| `flutter-mall/lib/view/digital_card/my_digital_card_page.dart` | MyDigitalCardPage, _fetchListFromApi, initState, _loadAssets, _openDetail (+2) |
| `consumer/internal/mq/digital_card/redemption_requested.go` | recordRedemptionLog, RedemptionRequested, createOmsOrder, isRetryableError, containsIgnoreCase (+1) |
| `flutter-mall/lib/view/digital_card/draw_activity_page.dart` | _buildRecoveryContext, _handlePrimaryAction, _safeLandingMessage, initState, _loadLanding (+1) |
| `flutter-mall/lib/view/digital_card/digital_card_display_text.dart` | digitalCardUserFacingText, digitalCardMintStatusText, digitalCardAssetPrimaryCopy, DigitalCardStatusCopy, digitalCardAssetStatusText (+1) |
| `flutter-mall/lib/view/digital_card/digital_card_claim_page.dart` | DigitalCardClaimPage, _validateToken, _claimCard, _checkLoginStatus, _responseData |
| `flutter-mall/lib/model/digital_card/digital_card_asset_model.dart` | QueryMyDigitalCardAssetListData, DigitalCardAssetDetailData, DigitalCardAssetDrawSummary, DigitalCardAssetTimelineItem, queryMyDigitalCardAssetListResponseFromJson |
| `flutter-mall/lib/utils/app_recovery_router.dart` | buildTarget, buildFallback |
| `flutter-mall/lib/view/digital_card/draw_result_sheet.dart` | _openAssetCenter, DrawResultSheet |

## Entry Points

Start here when exploring this area:

- **`main`** (Function) — `flutter-mall/test/app_recovery_router_test.dart:16`
- **`NewCardRedemptionOrderServiceClient`** (Function) — `rpc/sms/smsclient/card_redemption_order.go:119`
- **`RedemptionRequested`** (Function) — `consumer/internal/mq/digital_card/redemption_requested.go:138`
- **`main`** (Function) — `flutter-mall/test/draw_result_sheet_test.dart:5`
- **`main`** (Function) — `flutter-mall/test/digital_card_physical_fulfillment_test.dart:8`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `MyDigitalCardPage` | Class | `flutter-mall/lib/view/digital_card/my_digital_card_page.dart` | 16 |
| `DigitalCardClaimPage` | Class | `flutter-mall/lib/view/digital_card/digital_card_claim_page.dart` | 9 |
| `DigitalCardAssetDetailPage` | Class | `flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart` | 22 |
| `PinJia` | Class | `flutter-mall/lib/view/mine/ping_jia/ping_jia.dart` | 34 |
| `OrderList` | Class | `flutter-mall/lib/view/mine/order/order_list.dart` | 21 |
| `OrderDetail` | Class | `flutter-mall/lib/view/mine/order/order_detail.dart` | 26 |
| `ApplyAfterSales` | Class | `flutter-mall/lib/view/mine/order/apply_after_sales.dart` | 31 |
| `CouponList` | Class | `flutter-mall/lib/view/mine/coupon/coupon_list.dart` | 16 |
| `AvailableCouponList` | Class | `flutter-mall/lib/view/mine/coupon/available_coupon_list.dart` | 12 |
| `DrawResultSheet` | Class | `flutter-mall/lib/view/digital_card/draw_result_sheet.dart` | 6 |
| `DrawEligibilitySummary` | Class | `flutter-mall/lib/model/digital_card/draw_activity_model.dart` | 126 |
| `DigitalCardPhysicalFulfillmentPage` | Class | `flutter-mall/lib/view/digital_card/digital_card_physical_fulfillment_page.dart` | 17 |
| `PhysicalFulfillmentDetailData` | Class | `flutter-mall/lib/model/digital_card/physical_fulfillment_model.dart` | 30 |
| `PhysicalFulfillmentTimelineItem` | Class | `flutter-mall/lib/model/digital_card/physical_fulfillment_model.dart` | 122 |
| `QueryMyDigitalCardAssetListData` | Class | `flutter-mall/lib/model/digital_card/digital_card_asset_model.dart` | 35 |
| `DigitalCardAssetDetailData` | Class | `flutter-mall/lib/model/digital_card/digital_card_asset_model.dart` | 160 |
| `DigitalCardAssetDrawSummary` | Class | `flutter-mall/lib/model/digital_card/digital_card_asset_model.dart` | 195 |
| `DigitalCardAssetTimelineItem` | Class | `flutter-mall/lib/model/digital_card/digital_card_asset_model.dart` | 224 |
| `PhysicalFulfillmentAddressSheet` | Class | `flutter-mall/lib/view/digital_card/physical_fulfillment_address_sheet.dart` | 4 |
| `Login` | Class | `flutter-mall/lib/view/mine/login/login.dart` | 19 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Build → DigitalCardUserFacingText` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Provider | 12 calls |
| Order | 11 calls |
| Test | 6 calls |
| Coupon | 3 calls |
| Login | 2 calls |
| Home | 2 calls |
| Orderservice | 2 calls |
| Model | 2 calls |

## How to Explore

1. `gitnexus_context({name: "main"})` — see callers and callees
2. `gitnexus_query({query: "digital_card"})` — find related execution flows
3. Read key files listed above for implementation details
