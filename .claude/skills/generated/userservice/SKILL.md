---
name: userservice
description: "Skill for the Userservice area of zero-admin. 456 symbols across 199 files."
---

# Userservice

456 symbols | 199 files | Cohesion: 86%

## When to Use

- Working with code in `rpc/`
- Understanding how TimeToString, TimeToString1, TimeToStr work
- Modifying userservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/sys/internal/logic/common/governance.go` | NormalizeProtoScope, ProtoScope, DefaultScope, ActivationStatusName, ValidateTenantWritable (+8) |
| `rpc/sys/internal/server/userservice/userserviceserver.go` | AddUser, UpdateUser, QueryUserDetail, QueryUserList, QueryUserRoleList (+7) |
| `rpc/sys/client/userservice/userservice.go` | AddUser, DeleteUser, UpdateUser, UpdateUserStatus, QueryUserDetail (+7) |
| `rpc/pms/internal/logic/productspuservice/queryproductspudetaillogic.go` | NewQueryProductSpuDetailLogic, QueryProductSpuDetail, buildBrandListData, buildProductAttributeListData, buildProductAttributeValueListData (+1) |
| `rpc/oms/internal/server/orderservice/order_service_server.go` | QueryOrderDetail, QueryOrderList, QueryCompensationChainList, QueryTimeOutOrderList, QueryManualRequiredOrders |
| `rpc/sys/internal/server/postservice/postserviceserver.go` | AddPost, UpdatePost, QueryPostDetail, QueryPostList |
| `rpc/sys/internal/server/noticeservice/noticeserviceserver.go` | AddNotice, UpdateNotice, QueryNoticeDetail, QueryNoticeList |
| `rpc/sys/internal/server/dictitemservice/dictitemserviceserver.go` | AddDictItem, UpdateDictItem, QueryDictItemDetail, QueryDictItemList |
| `rpc/sys/internal/server/dicttypeservice/dicttypeserviceserver.go` | AddDictType, UpdateDictType, QueryDictTypeDetail, QueryDictTypeList |
| `rpc/sys/internal/server/deptservice/deptserviceserver.go` | AddDept, UpdateDept, QueryDeptDetail, QueryDeptList |

## Entry Points

Start here when exploring this area:

- **`TimeToString`** (Function) — `pkg/time_util/time_string.go:5`
- **`TimeToString1`** (Function) — `pkg/time_util/time_string.go:12`
- **`TimeToStr`** (Function) — `pkg/time_util/time_string.go:20`
- **`ValidateUserRoleAssignment`** (Function) — `pkg/scope/role_scope.go:43`
- **`ScopeFilterSQL`** (Function) — `pkg/scope/query_filter.go:9`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TimeToString` | Function | `pkg/time_util/time_string.go` | 5 |
| `TimeToString1` | Function | `pkg/time_util/time_string.go` | 12 |
| `TimeToStr` | Function | `pkg/time_util/time_string.go` | 20 |
| `ValidateUserRoleAssignment` | Function | `pkg/scope/role_scope.go` | 43 |
| `ScopeFilterSQL` | Function | `pkg/scope/query_filter.go` | 9 |
| `ApplyGovernanceScope` | Function | `pkg/scope/query_filter.go` | 34 |
| `DefaltData` | Function | `pkg/pointerprocess/processnil.go` | 9 |
| `NewUpdateUserRoleListLogic` | Function | `rpc/sys/internal/logic/userservice/updateuserrolelistlogic.go` | 27 |
| `NewUpdateUserLogic` | Function | `rpc/sys/internal/logic/userservice/updateuserlogic.go` | 30 |
| `NewQueryUserRoleListLogic` | Function | `rpc/sys/internal/logic/userservice/queryuserrolelistlogic.go` | 27 |
| `NewQueryUserListLogic` | Function | `rpc/sys/internal/logic/userservice/queryuserlistlogic.go` | 26 |
| `NewQueryUserDetailLogic` | Function | `rpc/sys/internal/logic/userservice/queryuserdetaillogic.go` | 31 |
| `NewQueryDeptAndPostListLogic` | Function | `rpc/sys/internal/logic/userservice/querydeptandpostlistlogic.go` | 22 |
| `NewAddUserLogic` | Function | `rpc/sys/internal/logic/userservice/adduserlogic.go` | 28 |
| `NewQueryRoleUserListLogic` | Function | `rpc/sys/internal/logic/roleservice/queryroleuserlistlogic.go` | 26 |
| `NewQueryRoleListLogic` | Function | `rpc/sys/internal/logic/roleservice/queryrolelistlogic.go` | 25 |
| `NewQueryRoleDetailLogic` | Function | `rpc/sys/internal/logic/roleservice/queryroledetaillogic.go` | 30 |
| `NewUpdateNoticeLogic` | Function | `rpc/sys/internal/logic/noticeservice/updatenoticelogic.go` | 27 |
| `NewQueryNoticeListLogic` | Function | `rpc/sys/internal/logic/noticeservice/querynoticelistlogic.go` | 25 |
| `NewQueryNoticeDetailLogic` | Function | `rpc/sys/internal/logic/noticeservice/querynoticedetaillogic.go` | 27 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `JobHandler → TimeToStr` | cross_community | 5 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Drawactivityservice | 8 calls |
| Productcategoryservice | 8 calls |
| Scope | 7 calls |
| Roleservice | 5 calls |
| Order | 4 calls |
| Operatefunnel | 4 calls |
| Query | 2 calls |
| Merchantservice | 2 calls |

## How to Explore

1. `gitnexus_context({name: "TimeToString"})` — see callers and callees
2. `gitnexus_query({query: "userservice"})` — find related execution flows
3. Read key files listed above for implementation details
