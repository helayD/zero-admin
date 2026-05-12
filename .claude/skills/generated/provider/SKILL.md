---
name: provider
description: "Skill for the Provider area of zero-admin. 53 symbols across 24 files."
---

# Provider

53 symbols | 24 files | Cohesion: 54%

## When to Use

- Working with code in `flutter-mall/`
- Understanding how main, Counter, CommentProvider work
- Modifying provider-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `flutter-mall/lib/provider/app_lifecycle_provider.dart` | recordUpgradeGateShown, recordUpgradeGateBlocked, recordUpgradeActionTapped, _recordUpgradeEvent, recordIntentReceived (+5) |
| `flutter-mall/lib/provider/comment_provider.dart` | queryCommentList, queryCommentDetail, CommentProvider, CommentUploadProvider, submitComment |
| `flutter-mall/lib/view/mine/message/message.dart` | initState, _loadDetail, _handleMessageTap, _restoreMessageTarget |
| `flutter-mall/lib/provider/cart_model.dart` | getCheckProduct, getValidCheckProduct, CartModel, setCartListData |
| `flutter-mall/lib/view/cart/cart.dart` | _clearCart, _validateBeforeCheckout, _goToCheckout |
| `flutter-mall/lib/layout/upgrade_gate_page.dart` | _recordGateShown, _handlePrimaryAction, _handleSecondaryAction |
| `flutter-mall/lib/utils/http_util.dart` | get, post |
| `flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart` | _fetchDetailFromApi, _transferAsset |
| `flutter-mall/lib/view/mine/focus/focus.dart` | queryFocusOnList, _clearAttention |
| `flutter-mall/lib/view/mine/history/history.dart` | queryHistoryList, _clearHistory |

## Entry Points

Start here when exploring this area:

- **`main`** (Function) — `flutter-mall/test/cart_model_test.dart:5`
- **`Counter`** (Class) — `flutter-mall/lib/provider/counter.dart:8`
- **`CommentProvider`** (Class) — `flutter-mall/lib/provider/comment_provider.dart:12`
- **`CommentUploadProvider`** (Class) — `flutter-mall/lib/provider/comment_provider.dart:192`
- **`CartModel`** (Class) — `flutter-mall/lib/provider/cart_model.dart:12`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `Counter` | Class | `flutter-mall/lib/provider/counter.dart` | 8 |
| `CommentProvider` | Class | `flutter-mall/lib/provider/comment_provider.dart` | 12 |
| `CommentUploadProvider` | Class | `flutter-mall/lib/provider/comment_provider.dart` | 192 |
| `CartModel` | Class | `flutter-mall/lib/provider/cart_model.dart` | 12 |
| `main` | Function | `flutter-mall/test/cart_model_test.dart` | 5 |
| `get` | Method | `flutter-mall/lib/utils/http_util.dart` | 114 |
| `queryCommentList` | Method | `flutter-mall/lib/provider/comment_provider.dart` | 81 |
| `queryCommentDetail` | Method | `flutter-mall/lib/provider/comment_provider.dart` | 124 |
| `initState` | Method | `flutter-mall/lib/view/mine/message/message.dart` | 807 |
| `queryFocusOnList` | Method | `flutter-mall/lib/view/mine/focus/focus.dart` | 33 |
| `queryHistoryList` | Method | `flutter-mall/lib/view/mine/history/history.dart` | 33 |
| `queryCollectionList` | Method | `flutter-mall/lib/view/mine/collection/collection.dart` | 33 |
| `queryProductListData` | Method | `flutter-mall/lib/view/category/product/product_list.dart` | 47 |
| `recordUpgradeGateShown` | Method | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 300 |
| `recordUpgradeGateBlocked` | Method | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 325 |
| `recordUpgradeActionTapped` | Method | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 350 |
| `recordIntentReceived` | Method | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 145 |
| `recordIntentLoginRequired` | Method | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 150 |
| `recordIntentFallbackUsed` | Method | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 164 |
| `build` | Method | `flutter-mall/lib/main.dart` | 40 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Build → Init` | cross_community | 7 |
| `Build → Init` | cross_community | 6 |
| `Build → Init` | cross_community | 6 |
| `Build → AppRecoveryStore` | cross_community | 6 |
| `Build → PeekCurrentIntentContext` | cross_community | 6 |
| `Build → AppRecoveryStore` | cross_community | 5 |
| `Build → PeekCurrentIntentContext` | cross_community | 5 |
| `Build → AppRecoveryStore` | cross_community | 5 |
| `Build → PeekCurrentIntentContext` | cross_community | 5 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Message | 6 calls |
| Order | 4 calls |
| Test | 3 calls |
| Cluster_654 | 2 calls |
| Digital_card | 2 calls |
| Cart | 1 calls |
| Static | 1 calls |
| Layout | 1 calls |

## How to Explore

1. `gitnexus_context({name: "main"})` — see callers and callees
2. `gitnexus_query({query: "provider"})` — find related execution flows
3. Read key files listed above for implementation details
