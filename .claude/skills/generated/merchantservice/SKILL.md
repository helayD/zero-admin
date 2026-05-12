---
name: merchantservice
description: "Skill for the Merchantservice area of zero-admin. 69 symbols across 19 files."
---

# Merchantservice

69 symbols | 19 files | Cohesion: 79%

## When to Use

- Working with code in `rpc/`
- Understanding how EncodeMerchantPayload, NewCreateMerchantLogic, NewMerchantServiceClient work
- Modifying merchantservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/sys/internal/logic/merchantservice/service_helpers.go` | normalizeCreateMerchantReq, normalizeIDList, createUniqueMerchantCode, ensureTenantExists, ensureTenantActiveForMerchant (+11) |
| `rpc/sys/internal/server/merchantservice/merchantserviceserver.go` | CreateMerchant, QueryMerchantDetail, QueryMerchantList, ApproveMerchant, RejectMerchant (+4) |
| `rpc/sys/client/merchantservice/merchantservice.go` | CreateMerchant, QueryMerchantDetail, QueryMerchantList, ApproveMerchant, RejectMerchant (+4) |
| `rpc/sys/internal/logic/merchantservice/merchant_logic_test.go` | newMerchantTestSvc, seedMerchantTenant, seedMerchantUser, TestCreateApproveAndEnableMerchantFlow, TestApproveMerchantRejectsDisabledTenant (+1) |
| `rpc/sys/sysclient/sys_grpc.pb.go` | NewMerchantServiceClient, CreateMerchant, EnableMerchant, ApproveMerchant, DisableMerchant |
| `rpc/sys/internal/logic/merchantservice/createmerchantlogic.go` | NewCreateMerchantLogic, CreateMerchant |
| `rpc/sys/internal/logic/merchantservice/querymerchantlistlogic.go` | NewQueryMerchantListLogic, QueryMerchantList |
| `rpc/sys/internal/logic/merchantservice/querymerchantdetaillogic.go` | NewQueryMerchantDetailLogic, QueryMerchantDetail |
| `rpc/sys/internal/logic/common/governance.go` | ActivationStatusCode, UpsertUserScopeBinding |
| `rpc/sys/internal/logic/merchantservice/approvemerchantlogic.go` | NewApproveMerchantLogic, ApproveMerchant |

## Entry Points

Start here when exploring this area:

- **`EncodeMerchantPayload`** (Function) — `pkg/audit/merchant.go:32`
- **`NewCreateMerchantLogic`** (Function) — `rpc/sys/internal/logic/merchantservice/createmerchantlogic.go:24`
- **`NewMerchantServiceClient`** (Function) — `rpc/sys/sysclient/sys_grpc.pb.go:1693`
- **`TestCreateApproveAndEnableMerchantFlow`** (Function) — `rpc/sys/internal/logic/merchantservice/merchant_logic_test.go:129`
- **`TestApproveMerchantRejectsDisabledTenant`** (Function) — `rpc/sys/internal/logic/merchantservice/merchant_logic_test.go:201`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `EncodeMerchantPayload` | Function | `pkg/audit/merchant.go` | 32 |
| `NewCreateMerchantLogic` | Function | `rpc/sys/internal/logic/merchantservice/createmerchantlogic.go` | 24 |
| `NewMerchantServiceClient` | Function | `rpc/sys/sysclient/sys_grpc.pb.go` | 1693 |
| `TestCreateApproveAndEnableMerchantFlow` | Function | `rpc/sys/internal/logic/merchantservice/merchant_logic_test.go` | 129 |
| `TestApproveMerchantRejectsDisabledTenant` | Function | `rpc/sys/internal/logic/merchantservice/merchant_logic_test.go` | 201 |
| `TestDisableMerchantUpdatesScopeAndEvictsPermissionCache` | Function | `rpc/sys/internal/logic/merchantservice/merchant_logic_test.go` | 235 |
| `NewQueryMerchantListLogic` | Function | `rpc/sys/internal/logic/merchantservice/querymerchantlistlogic.go` | 18 |
| `NewQueryMerchantDetailLogic` | Function | `rpc/sys/internal/logic/merchantservice/querymerchantdetaillogic.go` | 18 |
| `EncodeGovernanceScopeMetadata` | Function | `pkg/scope/governance.go` | 100 |
| `ActivationStatusCode` | Function | `rpc/sys/internal/logic/common/governance.go` | 185 |
| `UpsertUserScopeBinding` | Function | `rpc/sys/internal/logic/common/governance.go` | 402 |
| `NewApproveMerchantLogic` | Function | `rpc/sys/internal/logic/merchantservice/approvemerchantlogic.go` | 18 |
| `NewRejectMerchantLogic` | Function | `rpc/sys/internal/logic/merchantservice/rejectmerchantlogic.go` | 18 |
| `NewRequestMerchantMaterialLogic` | Function | `rpc/sys/internal/logic/merchantservice/requestmerchantmateriallogic.go` | 18 |
| `NewEnableMerchantLogic` | Function | `rpc/sys/internal/logic/merchantservice/enablemerchantlogic.go` | 18 |
| `NewDisableMerchantLogic` | Function | `rpc/sys/internal/logic/merchantservice/disablemerchantlogic.go` | 18 |
| `NewArchiveMerchantLogic` | Function | `rpc/sys/internal/logic/merchantservice/archivemerchantlogic.go` | 18 |
| `CreateMerchant` | Method | `rpc/sys/internal/server/merchantservice/merchantserviceserver.go` | 25 |
| `CreateMerchant` | Method | `rpc/sys/internal/logic/merchantservice/createmerchantlogic.go` | 32 |
| `CreateMerchant` | Method | `rpc/sys/sysclient/sys_grpc.pb.go` | 1697 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Productcategoryservice | 2 calls |
| Audit | 1 calls |
| Userservice | 1 calls |

## How to Explore

1. `gitnexus_context({name: "EncodeMerchantPayload"})` — see callers and callees
2. `gitnexus_query({query: "merchantservice"})` — find related execution flows
3. Read key files listed above for implementation details
