---
name: query
description: "Skill for the Query area of zero-admin. 735 symbols across 86 files."
---

# Query

735 symbols | 86 files | Cohesion: 91%

## When to Use

- Working with code in `rpc/`
- Understanding how NewAddCouponRecordLogic, Use, Use work
- Modifying query-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/ums/gen/query/ums_member_message.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+13) |
| `rpc/sys/gen/query/sys_user_role.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_user_post.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_user.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_role_menu.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_role.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_post.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_operate_log.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_notice.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |
| `rpc/sys/gen/query/sys_menu.gen.go` | Clauses, Limit, Offset, FindByPage, ScanByPage (+4) |

## Entry Points

Start here when exploring this area:

- **`NewAddCouponRecordLogic`** (Function) — `rpc/sms/internal/logic/couponrecordservice/add_coupon_record_logic.go:28`
- **`Use`** (Function) — `rpc/sys/gen/query/gen.go:51`
- **`Use`** (Function) — `rpc/sms/gen/query/gen.go:43`
- **`Use`** (Function) — `rpc/ums/gen/query/gen.go:53`
- **`Use`** (Function) — `rpc/oms/gen/query/gen.go:49`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewAddCouponRecordLogic` | Function | `rpc/sms/internal/logic/couponrecordservice/add_coupon_record_logic.go` | 28 |
| `Use` | Function | `rpc/sys/gen/query/gen.go` | 51 |
| `Use` | Function | `rpc/sms/gen/query/gen.go` | 43 |
| `Use` | Function | `rpc/ums/gen/query/gen.go` | 53 |
| `Use` | Function | `rpc/oms/gen/query/gen.go` | 49 |
| `Use` | Function | `rpc/cms/gen/query/gen.go` | 49 |
| `Clauses` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 241 |
| `Limit` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 297 |
| `Offset` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 301 |
| `FindByPage` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 415 |
| `ScanByPage` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 430 |
| `Scan` | Method | `rpc/ums/gen/query/ums_member_message.gen.go` | 440 |
| `Clauses` | Method | `rpc/sys/gen/query/sys_user_role.gen.go` | 193 |
| `Limit` | Method | `rpc/sys/gen/query/sys_user_role.gen.go` | 249 |
| `Offset` | Method | `rpc/sys/gen/query/sys_user_role.gen.go` | 253 |
| `FindByPage` | Method | `rpc/sys/gen/query/sys_user_role.gen.go` | 365 |
| `ScanByPage` | Method | `rpc/sys/gen/query/sys_user_role.gen.go` | 380 |
| `Clauses` | Method | `rpc/sys/gen/query/sys_user_post.gen.go` | 193 |
| `Limit` | Method | `rpc/sys/gen/query/sys_user_post.gen.go` | 249 |
| `Offset` | Method | `rpc/sys/gen/query/sys_user_post.gen.go` | 253 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Main → UmsMemberAddress` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |
| `Main → UmsMemberConsumeSetting` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |
| `Main → UmsMemberGrowthLog` | cross_community | 6 |
| `Main → SmsCoupon` | cross_community | 6 |
| `Main → TableName` | cross_community | 6 |
| `Main → FillFieldMap` | cross_community | 6 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Roleservice | 3 calls |
| Drawactivityservice | 2 calls |

## How to Explore

1. `gitnexus_context({name: "NewAddCouponRecordLogic"})` — see callers and callees
2. `gitnexus_query({query: "query"})` — find related execution flows
3. Read key files listed above for implementation details
