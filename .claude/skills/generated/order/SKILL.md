---
name: order
description: "Skill for the Order area of zero-admin. 564 symbols across 223 files."
---

# Order

564 symbols | 223 files | Cohesion: 81%

## When to Use

- Working with code in `api/`
- Understanding how RegisterHandlers, NewRecordHomeAdvertiseClickLogic, NewIndexLogic work
- Modifying order-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `flutter-mall/lib/view/mine/order/order_submit.dart` | buildProductList, build, _buildUpgradeFallbackContext, _ensureUpgradeReady, build (+23) |
| `flutter-mall/lib/view/mine/order/apply_after_sales.dart` | _buildFormState, _ensureUpgradeReady, _resumePendingPermissionFlow, _handlePickImage, _pickImageWithFlow (+19) |
| `flutter-mall/lib/view/category/product/product_detail.dart` | ProductDetail, build, buildPageBody, _cardDecoration, _sectionCard (+18) |
| `flutter-mall/lib/view/mine/order/order_pay.dart` | _buildInitialState, _buildOrderSummaryCard, _buildPayTypeSelector, _buildPayButton, _buildFailedState (+16) |
| `flutter-mall/lib/view/mine/order/order_detail.dart` | _buildBody, _buildSkeletonScreen, _buildContent, _buildOrderStatusBar, _buildAddressCard (+10) |
| `flutter-mall/lib/view/digital_card/draw_activity_page.dart` | _buildBody, _buildHeroCard, _buildStatusPill, _buildGhostPill, _buildHeroCardFace (+9) |
| `api/front/internal/logic/order/order/order_status_service.go` | UpdateOrderStatus, getCurrentOrderStatus, handlePaymentSuccess, handlePaymentFailed, handleConfirmReceive (+6) |
| `flutter-mall/lib/view/mine/order/order_list.dart` | build, _OrderListItem, build, _buildHeader, _buildProductPreview (+5) |
| `flutter-mall/lib/view/cart/cart.dart` | _updateQuantity, _estimateSelectedSavings, _openProductDetail, _groupCartItems, _buildShopSection (+4) |
| `api/front/internal/logic/digital_card/draw_activity/helper.go` | rpcError, apiCodeFromEligibility, mapCardPreviews, mapPoolPreviews, mapDrawRecords (+4) |

## Entry Points

Start here when exploring this area:

- **`RegisterHandlers`** (Function) — `api/front/internal/handler/routes.go:29`
- **`NewRecordHomeAdvertiseClickLogic`** (Function) — `api/front/internal/logic/home/record_home_advertise_click_logic.go:25`
- **`NewIndexLogic`** (Function) — `api/front/internal/logic/home/index_logic.go:32`
- **`RecordHomeAdvertiseClickHandler`** (Function) — `api/front/internal/handler/home/record_home_advertise_click_handler.go:11`
- **`IndexHandler`** (Function) — `api/front/internal/handler/home/index_handler.go:14`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `PriceBreakdownCard` | Class | `flutter-mall/lib/widgets/price_breakdown_card.dart` | 8 |
| `OrderTimelinePanel` | Class | `flutter-mall/lib/widgets/order_timeline_panel.dart` | 9 |
| `EmptyStateWidget` | Class | `flutter-mall/lib/widgets/empty_state_widget.dart` | 15 |
| `ErrorRetryWidget` | Class | `flutter-mall/lib/widgets/empty_state_widget.dart` | 97 |
| `CachedImageWidget` | Class | `flutter-mall/lib/widgets/cached_image_widget.dart` | 12 |
| `PhysicalFulfillmentStatusBadge` | Class | `flutter-mall/lib/view/digital_card/physical_fulfillment_status_badge.dart` | 3 |
| `DrawActivityRulePage` | Class | `flutter-mall/lib/view/digital_card/draw_activity_rule_page.dart` | 4 |
| `OrderLogistics` | Class | `flutter-mall/lib/view/mine/order/order_logistics.dart` | 26 |
| `BrandDetail` | Class | `flutter-mall/lib/view/home/brand/brand_detail.dart` | 17 |
| `ProductList` | Class | `flutter-mall/lib/view/category/product/product_list.dart` | 16 |
| `ProductDetail` | Class | `flutter-mall/lib/view/category/product/product_detail.dart` | 27 |
| `UpgradeGateService` | Class | `flutter-mall/lib/utils/upgrade_gate_service.dart` | 33 |
| `PendingUpgradeContext` | Class | `flutter-mall/lib/model/upgrade_gate_context.dart` | 34 |
| `UpgradeGatePage` | Class | `flutter-mall/lib/layout/upgrade_gate_page.dart` | 14 |
| `AfterSalesDraftSnapshot` | Class | `flutter-mall/lib/model/permission_flow_context.dart` | 491 |
| `OrderPay` | Class | `flutter-mall/lib/view/mine/order/order_pay.dart` | 37 |
| `CouponSelectSheet` | Class | `flutter-mall/lib/view/mine/order/coupon_select_sheet.dart` | 5 |
| `AddressSelectSheet` | Class | `flutter-mall/lib/view/mine/order/address_select_sheet.dart` | 7 |
| `RegisterHandlers` | Function | `api/front/internal/handler/routes.go` | 29 |
| `NewRecordHomeAdvertiseClickLogic` | Function | `api/front/internal/logic/home/record_home_advertise_click_logic.go` | 25 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `RegisterExtraHandlers → CodeError` | cross_community | 7 |
| `Build → Init` | cross_community | 7 |
| `BuildFooter → ProductDetailModel` | cross_community | 7 |
| `OrderPayQueryStatusHandler → CodeError` | cross_community | 6 |
| `UpdateCouponStatusHandler → CodeError` | cross_community | 6 |
| `UpdateMenuTemplateHandler → CodeError` | cross_community | 6 |
| `AddMenuTemplateHandler → CodeError` | cross_community | 6 |
| `UpdateProductFulfillmentRuleStatusHandler → CodeError` | cross_community | 6 |
| `UpdateProductFulfillmentRuleHandler → CodeError` | cross_community | 6 |
| `DeleteProductFulfillmentRuleHandler → CodeError` | cross_community | 6 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Provider | 23 calls |
| Ping_jia | 14 calls |
| Digital_card | 13 calls |
| Test | 13 calls |
| Svc | 10 calls |
| Cart | 9 calls |
| Digitalcardmint | 8 calls |
| Model | 6 calls |

## How to Explore

1. `gitnexus_context({name: "RegisterHandlers"})` — see callers and callees
2. `gitnexus_query({query: "order"})` — find related execution flows
3. Read key files listed above for implementation details
