---
name: productspuservice
description: "Skill for the Productspuservice area of zero-admin. 82 symbols across 23 files."
---

# Productspuservice

82 symbols | 23 files | Cohesion: 71%

## When to Use

- Working with code in `rpc/`
- Understanding how Use, NewUpdateVerifyStatusLogic, NewUpdateRecommendStatusLogic work
- Modifying productspuservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/pms/client/productspuservice/productspuservice.go` | DeleteProductSpu, QueryProductSpuList, UpdateNewStatus, UpdateDeleteStatus, UpdateNewStatusSort (+7) |
| `rpc/pms/internal/server/productspuservice/productspuserviceserver.go` | AddProductSpu, DeleteProductSpu, UpdateProductSpu, UpdateVerifyStatus, UpdatePublishStatus (+5) |
| `rpc/pms/internal/logic/productspuservice/productspu_draft_test.go` | newProductSpuDraftTestSvc, TestUpdateRecommendStatusRejectsInvisibleProduct, TestUpdateRecommendStatusPersistsOperationMetadata, TestAddProductSpuGeneratesDeterministicSkuCodesAndScopeSummary, TestUpdateProductSpuRebuildsSkuSummaryWithDeterministicCodes (+2) |
| `rpc/pms/pmsclient/pms_grpc.pb.go` | NewProductSpuServiceClient, UpdateRecommendStatus, AddProductSpu, UpdateProductSpu, UpdateVerifyStatus (+1) |
| `rpc/pms/internal/logic/productspuservice/es_sync.go` | buildProductEventMeta, sendProductESSync, sendProductESSyncBatch, sendProductESDelete, syncProductIndexVisibility |
| `rpc/pms/internal/logic/common/write_scope.go` | ResolveWriteScope, ApplyProductScope, ApplySkuScope, EnsureProductScope, EnsureSpuScopeForSku |
| `rpc/pms/internal/logic/productspuservice/review_metadata.go` | loadProductReviewMetadata, latestReviewRecord, formatReviewTime, reviewRecordTime |
| `rpc/pms/gen/query/gen.go` | Use, WithContext, SetDefault |
| `rpc/pms/internal/logic/common/product_sku_maintenance.go` | ComposeSkuValidationSet, RefreshSpuDraftSummary, EnsureSkuCode |
| `rpc/pms/internal/logic/productspuservice/updateverifystatuslogic.go` | NewUpdateVerifyStatusLogic, UpdateVerifyStatus |

## Entry Points

Start here when exploring this area:

- **`Use`** (Function) — `rpc/pms/gen/query/gen.go:53`
- **`NewUpdateVerifyStatusLogic`** (Function) — `rpc/pms/internal/logic/productspuservice/updateverifystatuslogic.go:24`
- **`NewUpdateRecommendStatusLogic`** (Function) — `rpc/pms/internal/logic/productspuservice/updaterecommendstatuslogic.go:23`
- **`NewUpdatePublishStatusLogic`** (Function) — `rpc/pms/internal/logic/productspuservice/updatepublishstatuslogic.go:24`
- **`NewUpdateProductSpuLogic`** (Function) — `rpc/pms/internal/logic/productspuservice/updateproductspulogic.go:29`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `Use` | Function | `rpc/pms/gen/query/gen.go` | 53 |
| `NewUpdateVerifyStatusLogic` | Function | `rpc/pms/internal/logic/productspuservice/updateverifystatuslogic.go` | 24 |
| `NewUpdateRecommendStatusLogic` | Function | `rpc/pms/internal/logic/productspuservice/updaterecommendstatuslogic.go` | 23 |
| `NewUpdatePublishStatusLogic` | Function | `rpc/pms/internal/logic/productspuservice/updatepublishstatuslogic.go` | 24 |
| `NewUpdateProductSpuLogic` | Function | `rpc/pms/internal/logic/productspuservice/updateproductspulogic.go` | 29 |
| `NewUpdateNewStatusLogic` | Function | `rpc/pms/internal/logic/productspuservice/updatenewstatuslogic.go` | 22 |
| `NewUpdateDeleteStatusLogic` | Function | `rpc/pms/internal/logic/productspuservice/updatedeletestatuslogic.go` | 22 |
| `NewDeleteProductSpuLogic` | Function | `rpc/pms/internal/logic/productspuservice/deleteproductspulogic.go` | 24 |
| `NewAddProductSpuLogic` | Function | `rpc/pms/internal/logic/productspuservice/addproductspulogic.go` | 27 |
| `NewUpdateProductSkuLogic` | Function | `rpc/pms/internal/logic/productskuservice/updateproductskulogic.go` | 28 |
| `NewAddProductSkuLogic` | Function | `rpc/pms/internal/logic/productskuservice/addproductskulogic.go` | 28 |
| `ResolveWriteScope` | Function | `rpc/pms/internal/logic/common/write_scope.go` | 36 |
| `ApplyProductScope` | Function | `rpc/pms/internal/logic/common/write_scope.go` | 44 |
| `ApplySkuScope` | Function | `rpc/pms/internal/logic/common/write_scope.go` | 74 |
| `EnsureProductScope` | Function | `rpc/pms/internal/logic/common/write_scope.go` | 94 |
| `EnsureSpuScopeForSku` | Function | `rpc/pms/internal/logic/common/write_scope.go` | 102 |
| `ComposeSkuValidationSet` | Function | `rpc/pms/internal/logic/common/product_sku_maintenance.go` | 25 |
| `RefreshSpuDraftSummary` | Function | `rpc/pms/internal/logic/common/product_sku_maintenance.go` | 66 |
| `EnsureSkuCode` | Function | `rpc/pms/internal/logic/common/product_sku_maintenance.go` | 101 |
| `NewProductSpuServiceClient` | Function | `rpc/pms/pmsclient/pms_grpc.pb.go` | 4443 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Main → PmsFeightTemplate` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |
| `Main → PmsMemberPrice` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |
| `Main → PmsProductAttribute` | cross_community | 6 |
| `Main → Query` | cross_community | 5 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Query | 14 calls |
| Scope | 10 calls |
| Drawactivityservice | 4 calls |
| Cluster_1324 | 3 calls |
| Subjectservice | 2 calls |
| Productcategoryservice | 2 calls |
| Audit | 2 calls |
| Product | 1 calls |

## How to Explore

1. `gitnexus_context({name: "Use"})` — see callers and callees
2. `gitnexus_query({query: "productspuservice"})` — find related execution flows
3. Read key files listed above for implementation details
