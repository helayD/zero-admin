---
name: commentservice
description: "Skill for the Commentservice area of zero-admin. 51 symbols across 14 files."
---

# Commentservice

51 symbols | 14 files | Cohesion: 82%

## When to Use

- Working with code in `rpc/`
- Understanding how NewUpdateCommentLogic, NewSubmitCommentAppealLogic, NewRestoreCommentLogic work
- Modifying commentservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/pms/internal/server/commentservice/commentserviceserver.go` | UpdateComment, RestoreComment, SubmitCommentAppeal, HandleCommentAppeal, QueryCommentList (+4) |
| `rpc/pms/client/commentservice/commentservice.go` | AddComment, DeleteComment, RestoreComment, QueryCommentDetail, QueryCommentList (+4) |
| `rpc/pms/internal/logic/commentservice/comment_audit_helper.go` | loadScopedComment, buildCommentAuditSnapshot, buildCommentAppealSnapshot, recordCommentAuditLogWithRollback, updateCommentStatsAsync (+1) |
| `rpc/pms/internal/logic/commentservice/audit_consistency_test.go` | newAuditFailureDB, TestRestoreCommentRollsBackWhenAuditLogInsertFails, TestHandleCommentAppealRollsBackWhenAuditLogInsertFails, newAuditSuccessDB, TestSubmitCommentAppealClearsHandledAtAfterResubmission (+1) |
| `rpc/pms/pmsclient/pms_grpc.pb.go` | NewCommentServiceClient, HandleCommentAppeal, SubmitCommentAppeal, UpdateComment |
| `rpc/pms/internal/logic/commentservice/updatecommentlogic.go` | NewUpdateCommentLogic, UpdateComment |
| `rpc/pms/internal/logic/commentservice/submitcommentappeallogic.go` | NewSubmitCommentAppealLogic, SubmitCommentAppeal |
| `rpc/pms/internal/logic/commentservice/restorecommentlogic.go` | NewRestoreCommentLogic, RestoreComment |
| `rpc/pms/internal/logic/commentservice/handlecommentappeallogic.go` | NewHandleCommentAppealLogic, HandleCommentAppeal |
| `rpc/pms/internal/logic/commentservice/querycommentlistlogic.go` | NewQueryCommentListLogic, QueryCommentList |

## Entry Points

Start here when exploring this area:

- **`NewUpdateCommentLogic`** (Function) — `rpc/pms/internal/logic/commentservice/updatecommentlogic.go:25`
- **`NewSubmitCommentAppealLogic`** (Function) — `rpc/pms/internal/logic/commentservice/submitcommentappeallogic.go:22`
- **`NewRestoreCommentLogic`** (Function) — `rpc/pms/internal/logic/commentservice/restorecommentlogic.go:18`
- **`NewHandleCommentAppealLogic`** (Function) — `rpc/pms/internal/logic/commentservice/handlecommentappeallogic.go:21`
- **`NewCommentServiceClient`** (Function) — `rpc/pms/pmsclient/pms_grpc.pb.go:2001`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewUpdateCommentLogic` | Function | `rpc/pms/internal/logic/commentservice/updatecommentlogic.go` | 25 |
| `NewSubmitCommentAppealLogic` | Function | `rpc/pms/internal/logic/commentservice/submitcommentappeallogic.go` | 22 |
| `NewRestoreCommentLogic` | Function | `rpc/pms/internal/logic/commentservice/restorecommentlogic.go` | 18 |
| `NewHandleCommentAppealLogic` | Function | `rpc/pms/internal/logic/commentservice/handlecommentappeallogic.go` | 21 |
| `NewCommentServiceClient` | Function | `rpc/pms/pmsclient/pms_grpc.pb.go` | 2001 |
| `TestRestoreCommentRollsBackWhenAuditLogInsertFails` | Function | `rpc/pms/internal/logic/commentservice/audit_consistency_test.go` | 234 |
| `TestHandleCommentAppealRollsBackWhenAuditLogInsertFails` | Function | `rpc/pms/internal/logic/commentservice/audit_consistency_test.go` | 319 |
| `TestSubmitCommentAppealClearsHandledAtAfterResubmission` | Function | `rpc/pms/internal/logic/commentservice/audit_consistency_test.go` | 271 |
| `NewQueryCommentListLogic` | Function | `rpc/pms/internal/logic/commentservice/querycommentlistlogic.go` | 24 |
| `TestUpdateCommentRollsBackWhenAuditLogInsertFails` | Function | `rpc/pms/internal/logic/commentservice/audit_consistency_test.go` | 195 |
| `NewQueryCommentDetailLogic` | Function | `rpc/pms/internal/logic/commentservice/querycommentdetaillogic.go` | 23 |
| `NewQueryCommentAuditLogLogic` | Function | `rpc/pms/internal/logic/commentservice/querycommentauditloglogic.go` | 20 |
| `NewAddCommentLogic` | Function | `rpc/pms/internal/logic/commentservice/addcommentlogic.go` | 24 |
| `NewDeleteCommentLogic` | Function | `rpc/pms/internal/logic/commentservice/deletecommentlogic.go` | 24 |
| `UpdateComment` | Method | `rpc/pms/internal/logic/commentservice/updatecommentlogic.go` | 34 |
| `SubmitCommentAppeal` | Method | `rpc/pms/internal/logic/commentservice/submitcommentappeallogic.go` | 31 |
| `RestoreComment` | Method | `rpc/pms/internal/logic/commentservice/restorecommentlogic.go` | 27 |
| `HandleCommentAppeal` | Method | `rpc/pms/internal/logic/commentservice/handlecommentappeallogic.go` | 30 |
| `UpdateComment` | Method | `rpc/pms/internal/server/commentservice/commentserviceserver.go` | 38 |
| `RestoreComment` | Method | `rpc/pms/internal/server/commentservice/commentserviceserver.go` | 44 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Userservice | 2 calls |

## How to Explore

1. `gitnexus_context({name: "NewUpdateCommentLogic"})` — see callers and callees
2. `gitnexus_query({query: "commentservice"})` — find related execution flows
3. Read key files listed above for implementation details
