---
name: user
description: "Skill for the User area of zero-admin. 994 symbols across 646 files."
---

# User

994 symbols | 646 files | Cohesion: 94%

## When to Use

- Working with code in `api/`
- Understanding how RegisterHandlers, NewUpdateMemberTagStatusLogic, NewUpdateMemberTagLogic work
- Modifying user-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `web-admin/src/pages/system/user/index.tsx` | handleAdd, handleUpdate, UserList, showDeleteConfirm, showReSetPasswordConfirm (+4) |
| `api/admin/internal/common/query_scope.go` | CurrentGovernanceScope, ResolveQueryGovernanceScope, PMSGovernanceScope, OMSGovernanceScope, SMSGovernanceScope (+3) |
| `api/admin/internal/logic/sms/digital_card_physical_fulfillment/physical_fulfillment_logic.go` | QueryDigitalCardPhysicalFulfillmentDetail, EnsureDigitalCardPhysicalFulfillment, UpdateDigitalCardPhysicalProductionStatus, ShipDigitalCardPhysicalFulfillment, MarkDigitalCardPhysicalFulfillmentException (+1) |
| `web-admin/src/pages/system/user/service.ts` | addUser, updateUser, queryUserList, updateUserRoleList, removeUser (+1) |
| `api/admin/internal/logic/sms/draw_activity/helper.go` | grpcDrawError, toSMSHomeEntry, toSMSTemplates, toSMSPools, toAPIReadinessItems |
| `api/admin/internal/logic/sms/digital_card_physical_fulfillment/helper.go` | resolvePhysicalFulfillmentWriteScope, resolvePhysicalActionContext, mapPhysicalActionResp, writePhysicalFulfillmentOperateLog |
| `api/admin/internal/logic/oms/chain_monitor/retrychainlogic.go` | NewRetryChainLogic, writeChainInterventionLog, RetryChain |
| `api/admin/internal/common/errorx/baseerror.go` | NewCodeError, NewDefaultError, Error |
| `api/admin/internal/logic/sms/digital_card_chain/helper.go` | resolveDigitalCardChainWriteScope, validateActionReason, writeDigitalCardChainOperateLog |
| `api/admin/internal/logic/ums/member_tag/updatemembertagstatuslogic.go` | NewUpdateMemberTagStatusLogic, UpdateMemberTagStatus |

## Entry Points

Start here when exploring this area:

- **`RegisterHandlers`** (Function) — `api/admin/internal/handler/routes.go:72`
- **`NewUpdateMemberTagStatusLogic`** (Function) — `api/admin/internal/logic/ums/member_tag/updatemembertagstatuslogic.go:26`
- **`NewUpdateMemberTagLogic`** (Function) — `api/admin/internal/logic/ums/member_tag/updatemembertaglogic.go:26`
- **`NewQueryMemberTagListLogic`** (Function) — `api/admin/internal/logic/ums/member_tag/querymembertaglistlogic.go:25`
- **`NewQueryMemberTagDetailLogic`** (Function) — `api/admin/internal/logic/ums/member_tag/querymembertagdetaillogic.go:25`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `RegisterHandlers` | Function | `api/admin/internal/handler/routes.go` | 72 |
| `NewUpdateMemberTagStatusLogic` | Function | `api/admin/internal/logic/ums/member_tag/updatemembertagstatuslogic.go` | 26 |
| `NewUpdateMemberTagLogic` | Function | `api/admin/internal/logic/ums/member_tag/updatemembertaglogic.go` | 26 |
| `NewQueryMemberTagListLogic` | Function | `api/admin/internal/logic/ums/member_tag/querymembertaglistlogic.go` | 25 |
| `NewQueryMemberTagDetailLogic` | Function | `api/admin/internal/logic/ums/member_tag/querymembertagdetaillogic.go` | 25 |
| `NewDeleteMemberTagLogic` | Function | `api/admin/internal/logic/ums/member_tag/deletemembertaglogic.go` | 25 |
| `NewAddMemberTagLogic` | Function | `api/admin/internal/logic/ums/member_tag/addmembertaglogic.go` | 26 |
| `NewQueryMemberStatisticsInfoListLogic` | Function | `api/admin/internal/logic/ums/member_statistics/querymemberstatisticsinfolistlogic.go` | 25 |
| `NewQueryMemberStatisticsInfoDetailLogic` | Function | `api/admin/internal/logic/ums/member_statistics/querymemberstatisticsinfodetaillogic.go` | 20 |
| `NewQueryMemberSignLogListLogic` | Function | `api/admin/internal/logic/ums/member_sign/querymembersignloglistlogic.go` | 25 |
| `NewQueryMemberSignLogDetailLogic` | Function | `api/admin/internal/logic/ums/member_sign/querymembersignlogdetaillogic.go` | 25 |
| `NewUpdateMemberTaskStatusLogic` | Function | `api/admin/internal/logic/ums/member_task/updatemembertaskstatuslogic.go` | 26 |
| `NewUpdateMemberTaskLogic` | Function | `api/admin/internal/logic/ums/member_task/updatemembertasklogic.go` | 26 |
| `NewQueryMemberTaskListLogic` | Function | `api/admin/internal/logic/ums/member_task/querymembertasklistlogic.go` | 25 |
| `NewQueryMemberTaskDetailLogic` | Function | `api/admin/internal/logic/ums/member_task/querymembertaskdetaillogic.go` | 25 |
| `NewDeleteMemberTaskLogic` | Function | `api/admin/internal/logic/ums/member_task/deletemembertasklogic.go` | 25 |
| `NewAddMemberTaskLogic` | Function | `api/admin/internal/logic/ums/member_task/addmembertasklogic.go` | 26 |
| `NewUpdateMemberLevelStatusLogic` | Function | `api/admin/internal/logic/ums/member_level/updatememberlevelstatuslogic.go` | 27 |
| `NewUpdateMemberLevelLogic` | Function | `api/admin/internal/logic/ums/member_level/updatememberlevellogic.go` | 28 |
| `NewQueryMemberLevelListLogic` | Function | `api/admin/internal/logic/ums/member_level/querymemberlevellistlogic.go` | 27 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `ExportRepeatPurchaseAnalysisHandler → GovernanceScope` | cross_community | 6 |
| `ExportRepeatPurchaseAnalysisHandler → InferScopeType` | cross_community | 6 |
| `RegisterExtraHandlers → ReadContextInt64` | cross_community | 6 |
| `RegisterExtraHandlers → CodeError` | cross_community | 6 |
| `UpdateCouponStatusHandler → CodeError` | cross_community | 6 |
| `UpdateMenuTemplateHandler → CodeError` | cross_community | 6 |
| `AddMenuTemplateHandler → CodeError` | cross_community | 6 |
| `UpdateProductFulfillmentRuleStatusHandler → CodeError` | cross_community | 6 |
| `UpdateProductFulfillmentRuleStatusHandler → GovernanceScope` | cross_community | 6 |
| `UpdateProductFulfillmentRuleStatusHandler → InferScopeType` | cross_community | 6 |

## Connected Areas

| Area | Connections |
|------|-------------|
| MenuTemplate | 6 calls |
| Audit | 5 calls |
| Productcategoryservice | 4 calls |
| Scope | 3 calls |
| Order | 3 calls |
| Channel_integration_template | 3 calls |
| Chain_monitor | 2 calls |
| Product_spu | 2 calls |

## How to Explore

1. `gitnexus_context({name: "RegisterHandlers"})` — see callers and callees
2. `gitnexus_query({query: "user"})` — find related execution flows
3. Read key files listed above for implementation details
