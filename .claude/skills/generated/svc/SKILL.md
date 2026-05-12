---
name: svc
description: "Skill for the Svc area of zero-admin. 81 symbols across 77 files."
---

# Svc

81 symbols | 77 files | Cohesion: 85%

## When to Use

- Working with code in `rpc/`
- Understanding how NewServiceContext, NewUserService, NewRoleService work
- Modifying svc-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `job/internal/svc/service_context.go` | NewServiceContext, buildChainClient |
| `rpc/sys/internal/svc/servicecontext.go` | NewServiceContext, settingLogConfig |
| `rpc/oms/internal/svc/service_context.go` | NewServiceContext, settingLogConfig |
| `rpc/cms/internal/svc/servicecontext.go` | NewServiceContext, settingLogConfig |
| `rpc/sys/client/userservice/userservice.go` | NewUserService |
| `rpc/sys/client/roleservice/roleservice.go` | NewRoleService |
| `rpc/sys/client/loginlogservice/loginlogservice.go` | NewLoginLogService |
| `rpc/sys/client/menuservice/menuservice.go` | NewMenuService |
| `rpc/sys/client/deptservice/deptservice.go` | NewDeptService |
| `rpc/sms/client/seckillsessionservice/seckill_session_service.go` | NewSeckillSessionService |

## Entry Points

Start here when exploring this area:

- **`NewServiceContext`** (Function) — `job/internal/svc/service_context.go:40`
- **`NewUserService`** (Function) — `rpc/sys/client/userservice/userservice.go:235`
- **`NewRoleService`** (Function) — `rpc/sys/client/roleservice/roleservice.go:231`
- **`NewLoginLogService`** (Function) — `rpc/sys/client/loginlogservice/loginlogservice.go:217`
- **`NewMenuService`** (Function) — `rpc/sys/client/menuservice/menuservice.go:224`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewServiceContext` | Function | `job/internal/svc/service_context.go` | 40 |
| `NewUserService` | Function | `rpc/sys/client/userservice/userservice.go` | 235 |
| `NewRoleService` | Function | `rpc/sys/client/roleservice/roleservice.go` | 231 |
| `NewLoginLogService` | Function | `rpc/sys/client/loginlogservice/loginlogservice.go` | 217 |
| `NewMenuService` | Function | `rpc/sys/client/menuservice/menuservice.go` | 224 |
| `NewDeptService` | Function | `rpc/sys/client/deptservice/deptservice.go` | 223 |
| `NewSeckillSessionService` | Function | `rpc/sms/client/seckillsessionservice/seckill_session_service.go` | 223 |
| `NewSeckillReservationService` | Function | `rpc/sms/client/seckillreservationservice/seckill_reservation_service.go` | 221 |
| `NewSeckillProductService` | Function | `rpc/sms/client/seckillproductservice/seckill_product_service.go` | 223 |
| `NewSeckillActivityService` | Function | `rpc/sms/client/seckillactivityservice/seckill_activity_service.go` | 223 |
| `NewOperateDashboardService` | Function | `rpc/sms/client/operatedashboardservice/operate_dashboard_service.go` | 213 |
| `NewHomeAdvertiseService` | Function | `rpc/sms/client/homeadvertiseservice/home_advertise_service.go` | 221 |
| `NewCouponService` | Function | `rpc/sms/client/couponservice/coupon_service.go` | 227 |
| `NewCouponScopeService` | Function | `rpc/sms/client/couponscopeservice/coupon_scope_service.go` | 219 |
| `NewCouponRecordService` | Function | `rpc/sms/client/couponrecordservice/coupon_record_service.go` | 223 |
| `NewCouponTypeService` | Function | `rpc/sms/client/coupontypeservice/coupon_type_service.go` | 221 |
| `NewCardMintAdminService` | Function | `rpc/sms/client/cardmintadminservice/card_mint_admin_service.go` | 38 |
| `NewMemberTaskService` | Function | `rpc/ums/client/membertaskservice/membertaskservice.go` | 235 |
| `NewMemberTagService` | Function | `rpc/ums/client/membertagservice/membertagservice.go` | 235 |
| `NewMemberTaskRelationService` | Function | `rpc/ums/client/membertaskrelationservice/membertaskrelationservice.go` | 229 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Main → SysDept` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |
| `Main → SysDictItem` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |
| `Main → SysDictType` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → OmsCartItem` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Query | 3 calls |
| Order | 3 calls |
| Digitalcardmint | 2 calls |
| Mq | 2 calls |

## How to Explore

1. `gitnexus_context({name: "NewServiceContext"})` — see callers and callees
2. `gitnexus_query({query: "svc"})` — find related execution flows
3. Read key files listed above for implementation details
