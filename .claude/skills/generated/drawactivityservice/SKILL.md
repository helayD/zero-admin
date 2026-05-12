---
name: drawactivityservice
description: "Skill for the Drawactivityservice area of zero-admin. 52 symbols across 16 files."
---

# Drawactivityservice

52 symbols | 16 files | Cohesion: 61%

## When to Use

- Working with code in `rpc/`
- Understanding how CleanupExpiredClaimTokens, CheckRedemptionOrderTimeout, ApplyDrawActivityScope work
- Modifying drawactivityservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/sms/internal/logic/drawactivityservice/draw_activity_helper.go` | upsertDrawTemplates, replaceDrawPools, ensureUniqueActivityCode, saveDrawAggregate, safeInt64 (+6) |
| `rpc/sms/internal/logic/drawactivityservice/draw_activity_logic_test.go` | newDrawActivityLogicSvc, merchantScope, validDrawAddReq, TestAddDrawActivityRejectsInvalidProbabilitySum, TestAddDrawActivityRejectsCrossScopeWrite (+3) |
| `rpc/sms/client/drawactivityservice/draw_activity_service.go` | AddDrawActivity, UpdateDrawActivity, DeleteDrawActivity, UpdateDrawActivityStatus, QueryDrawActivityDetail (+2) |
| `rpc/sms/smsclient/sms_grpc.pb.go` | NewDrawActivityServiceClient, UpdateDrawActivity, UpdateDrawActivityStatus, PreviewDrawActivityPublishReadiness, AddDrawActivity |
| `rpc/sms/internal/server/drawactivityservice/draw_activity_service_server.go` | DeleteDrawActivity, PreviewDrawActivityPublishReadiness, AddDrawActivity, UpdateDrawActivity, UpdateDrawActivityStatus |
| `rpc/sms/internal/logic/drawactivityservice/delete_draw_activity_logic.go` | NewDeleteDrawActivityLogic, DeleteDrawActivity |
| `rpc/sms/internal/logic/drawactivityservice/preview_draw_activity_publish_readiness_logic.go` | NewPreviewDrawActivityPublishReadinessLogic, PreviewDrawActivityPublishReadiness |
| `rpc/sms/internal/logic/drawactivityservice/add_draw_activity_logic.go` | NewAddDrawActivityLogic, AddDrawActivity |
| `rpc/sms/internal/logic/drawactivityservice/update_draw_activity_logic.go` | NewUpdateDrawActivityLogic, UpdateDrawActivity |
| `rpc/sms/internal/logic/drawactivityservice/update_draw_activity_status_logic.go` | NewUpdateDrawActivityStatusLogic, UpdateDrawActivityStatus |

## Entry Points

Start here when exploring this area:

- **`CleanupExpiredClaimTokens`** (Function) — `job/internal/jobs/cleanup_expired_claim_tokens.go:26`
- **`CheckRedemptionOrderTimeout`** (Function) — `job/internal/jobs/check_redemption_order_timeout.go:51`
- **`ApplyDrawActivityScope`** (Function) — `rpc/sms/internal/logic/common/write_scope.go:82`
- **`ApplySkuScopeBatch`** (Function) — `rpc/pms/internal/logic/common/write_scope.go:78`
- **`NewDrawActivityServiceClient`** (Function) — `rpc/sms/smsclient/sms_grpc.pb.go:1476`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `CleanupExpiredClaimTokens` | Function | `job/internal/jobs/cleanup_expired_claim_tokens.go` | 26 |
| `CheckRedemptionOrderTimeout` | Function | `job/internal/jobs/check_redemption_order_timeout.go` | 51 |
| `ApplyDrawActivityScope` | Function | `rpc/sms/internal/logic/common/write_scope.go` | 82 |
| `ApplySkuScopeBatch` | Function | `rpc/pms/internal/logic/common/write_scope.go` | 78 |
| `NewDrawActivityServiceClient` | Function | `rpc/sms/smsclient/sms_grpc.pb.go` | 1476 |
| `TestAddDrawActivityRejectsInvalidProbabilitySum` | Function | `rpc/sms/internal/logic/drawactivityservice/draw_activity_logic_test.go` | 272 |
| `TestAddDrawActivityRejectsCrossScopeWrite` | Function | `rpc/sms/internal/logic/drawactivityservice/draw_activity_logic_test.go` | 289 |
| `TestPreviewDrawActivityPersistsStructuredFailures` | Function | `rpc/sms/internal/logic/drawactivityservice/draw_activity_logic_test.go` | 300 |
| `TestUpdateDrawActivityStatusBlocksPublishWhenPreviewFails` | Function | `rpc/sms/internal/logic/drawactivityservice/draw_activity_logic_test.go` | 343 |
| `TestAddAndUpdateAppendAuditRecords` | Function | `rpc/sms/internal/logic/drawactivityservice/draw_activity_logic_test.go` | 367 |
| `NewDeleteDrawActivityLogic` | Function | `rpc/sms/internal/logic/drawactivityservice/delete_draw_activity_logic.go` | 21 |
| `NewPreviewDrawActivityPublishReadinessLogic` | Function | `rpc/sms/internal/logic/drawactivityservice/preview_draw_activity_publish_readiness_logic.go` | 17 |
| `NewAddDrawActivityLogic` | Function | `rpc/sms/internal/logic/drawactivityservice/add_draw_activity_logic.go` | 19 |
| `NewUpdateDrawActivityLogic` | Function | `rpc/sms/internal/logic/drawactivityservice/update_draw_activity_logic.go` | 21 |
| `NewUpdateDrawActivityStatusLogic` | Function | `rpc/sms/internal/logic/drawactivityservice/update_draw_activity_status_logic.go` | 22 |
| `Updates` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 456 |
| `UpdateDrawActivity` | Method | `rpc/sms/smsclient/sms_grpc.pb.go` | 1489 |
| `UpdateDrawActivityStatus` | Method | `rpc/sms/smsclient/sms_grpc.pb.go` | 1507 |
| `PreviewDrawActivityPublishReadiness` | Method | `rpc/sms/smsclient/sms_grpc.pb.go` | 1534 |
| `AddDrawActivity` | Method | `rpc/sms/client/drawactivityservice/draw_activity_service.go` | 222 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Userservice | 5 calls |
| Cluster_1173 | 4 calls |
| Scope | 4 calls |
| Digitalcardmint | 1 calls |
| Audit | 1 calls |

## How to Explore

1. `gitnexus_context({name: "CleanupExpiredClaimTokens"})` — see callers and callees
2. `gitnexus_query({query: "drawactivityservice"})` — find related execution flows
3. Read key files listed above for implementation details
