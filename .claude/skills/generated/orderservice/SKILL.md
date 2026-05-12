---
name: orderservice
description: "Skill for the Orderservice area of zero-admin. 123 symbols across 33 files."
---

# Orderservice

123 symbols | 33 files | Cohesion: 62%

## When to Use

- Working with code in `rpc/`
- Understanding how NewOrderServiceClient, NewConfirmOrderLogic, NewCancelOrderLogic work
- Modifying orderservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/oms/client/orderservice/order_service.go` | DeleteOrder, UpdateOrder, UpdateOrderStatus, QueryOperateOrderFunnel, QueryRepeatPurchaseDetailList (+15) |
| `rpc/oms/internal/server/orderservice/order_service_server.go` | CancelOrder, ConfirmOrder, AddOrder, DeleteOrder, QueryRepeatPurchaseDetailList (+11) |
| `rpc/oms/internal/logic/orderservice/repeat_purchase_snapshot.go` | queryRepeatPurchaseTrackingState, normalizeRepeatPurchaseTrackingTime, parseRepeatPurchaseTimeText, BuildRepeatPurchaseSnapshot, buildRepeatPurchaseDetailRowsAndOverview (+6) |
| `rpc/oms/internal/logic/orderservice/query_operate_order_funnel_logic.go` | NewQueryOperateOrderFunnelLogic, buildOperateOrderBucketQuery, operateOrderChannelCaseSQL, queryOrderCreatedBuckets, queryPaySuccessBuckets (+1) |
| `rpc/oms/omsclient/oms_grpc.pb.go` | NewOrderServiceClient, AddOrder, QueryRepeatPurchaseAnalysis, QueryOrderDetail, QueryOrderList |
| `rpc/oms/internal/logic/orderservice/add_order_logic_test.go` | newAddOrderTestSvc, newAddOrderReq, TestAddOrderMarksCartItemDeletedByCartItemID, TestAddOrderDirectBuyDoesNotDeleteCartItems, TestAddOrderFallsBackToPlatformScopeWhenMemberIsNotSysUser |
| `rpc/oms/internal/logic/orderservice/repeat_purchase_query_builder_test.go` | TestBuildRepeatPurchasePriorOrderExistsSQLHonorsLookbackAndScope, newRepeatPurchaseDryRunDB, TestBuildRepeatPurchaseQualifiedOrderQueryAppliesCurrentOrderFilters, TestBuildRepeatPurchaseQualifiedOrderQuerySupportsNoneActivityAndPlatformScope, TestNewRepeatPurchaseQueryBuilderUsesSvcDB |
| `rpc/oms/internal/logic/orderservice/repeat_purchase_query_builder.go` | buildRepeatPurchaseQualifiedOrderQuery, buildRepeatPurchasePriorOrderExistsSQL, applyRepeatPurchaseCurrentOrderFilters, repeatPurchaseOrderChannelCaseSQL, newRepeatPurchaseQueryBuilder |
| `rpc/oms/internal/logic/orderservice/queryrepeatpurchaseanalysislogic_test.go` | newRepeatPurchaseTestSvc, seedRepeatPurchaseOrder, TestQueryRepeatPurchaseAnalysisHonorsEffectiveOrderScopeAndCurrentFilters, TestQueryRepeatPurchaseDetailListUsesStableSortAndAggregatesRepeatOrders |
| `rpc/oms/internal/logic/orderservice/query_order_scope_test.go` | seedOrderMain, TestQueryOrderDetailRejectsCrossScopeLookup, TestQueryOrderListFiltersByGovernanceScope |

## Entry Points

Start here when exploring this area:

- **`NewOrderServiceClient`** (Function) — `rpc/oms/omsclient/oms_grpc.pb.go:999`
- **`NewConfirmOrderLogic`** (Function) — `rpc/oms/internal/logic/orderservice/confirm_order_logic.go:23`
- **`NewCancelOrderLogic`** (Function) — `rpc/oms/internal/logic/orderservice/cancel_order_logic.go:22`
- **`ResolveActorScope`** (Function) — `rpc/oms/internal/logic/common/write_scope.go:19`
- **`TestAddOrderMarksCartItemDeletedByCartItemID`** (Function) — `rpc/oms/internal/logic/orderservice/add_order_logic_test.go:132`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewOrderServiceClient` | Function | `rpc/oms/omsclient/oms_grpc.pb.go` | 999 |
| `NewConfirmOrderLogic` | Function | `rpc/oms/internal/logic/orderservice/confirm_order_logic.go` | 23 |
| `NewCancelOrderLogic` | Function | `rpc/oms/internal/logic/orderservice/cancel_order_logic.go` | 22 |
| `ResolveActorScope` | Function | `rpc/oms/internal/logic/common/write_scope.go` | 19 |
| `TestAddOrderMarksCartItemDeletedByCartItemID` | Function | `rpc/oms/internal/logic/orderservice/add_order_logic_test.go` | 132 |
| `TestAddOrderDirectBuyDoesNotDeleteCartItems` | Function | `rpc/oms/internal/logic/orderservice/add_order_logic_test.go` | 171 |
| `TestAddOrderFallsBackToPlatformScopeWhenMemberIsNotSysUser` | Function | `rpc/oms/internal/logic/orderservice/add_order_logic_test.go` | 190 |
| `NewAddOrderLogic` | Function | `rpc/oms/internal/logic/orderservice/add_order_logic.go` | 27 |
| `TestRepeatPurchasePartialMetricsOnlyWhenActivityTrackingIsIncomplete` | Function | `pkg/operatefunnel/repeat_purchase_test.go` | 159 |
| `RepeatPurchasePartialMetrics` | Function | `pkg/operatefunnel/repeat_purchase.go` | 121 |
| `TestQueryRepeatPurchaseAnalysisHonorsEffectiveOrderScopeAndCurrentFilters` | Function | `rpc/oms/internal/logic/orderservice/queryrepeatpurchaseanalysislogic_test.go` | 52 |
| `RepeatPurchaseScopeOrderColumns` | Function | `pkg/operatefunnel/repeat_purchase.go` | 66 |
| `TestBuildRepeatPurchasePriorOrderExistsSQLHonorsLookbackAndScope` | Function | `rpc/oms/internal/logic/orderservice/repeat_purchase_query_builder_test.go` | 73 |
| `TestQueryOrderDetailRejectsCrossScopeLookup` | Function | `rpc/oms/internal/logic/orderservice/query_order_scope_test.go` | 278 |
| `CancelTimeOutOrder` | Function | `job/internal/jobs/cancel_timeout_order_logic.go` | 42 |
| `NewDeleteOrderLogic` | Function | `rpc/oms/internal/logic/orderservice/delete_order_logic.go` | 20 |
| `ResolveWriteScope` | Function | `rpc/oms/internal/logic/common/write_scope.go` | 36 |
| `NewQueryRepeatPurchaseDetailListLogic` | Function | `rpc/oms/internal/logic/orderservice/queryrepeatpurchasedetaillistlogic.go` | 18 |
| `TestQueryRepeatPurchaseDetailListUsesStableSortAndAggregatesRepeatOrders` | Function | `rpc/oms/internal/logic/orderservice/queryrepeatpurchaseanalysislogic_test.go` | 105 |
| `NewQueryChainActionsLogic` | Function | `rpc/oms/internal/logic/orderservice/query_chain_actions_logic.go` | 19 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `JobHandler → OrderSettingServiceClient` | cross_community | 7 |
| `JobHandler → OrderServiceClient` | cross_community | 7 |
| `JobHandler → QueryDefaultSetting` | cross_community | 6 |
| `JobHandler → QueryTimeOutOrderList` | cross_community | 6 |
| `JobHandler → UpdateOrderConsistency` | cross_community | 6 |
| `JobHandler → TimeToStr` | cross_community | 5 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Operatefunnel | 13 calls |
| Productcategoryservice | 10 calls |
| Drawactivityservice | 8 calls |
| Scope | 5 calls |
| Operatedashboardservice | 5 calls |
| Userservice | 4 calls |
| Cartitemservice | 2 calls |
| Svc | 2 calls |

## How to Explore

1. `gitnexus_context({name: "NewOrderServiceClient"})` — see callers and callees
2. `gitnexus_query({query: "orderservice"})` — find related execution flows
3. Read key files listed above for implementation details
