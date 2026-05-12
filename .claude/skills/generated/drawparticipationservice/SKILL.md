---
name: drawparticipationservice
description: "Skill for the Drawparticipationservice area of zero-admin. 52 symbols across 9 files."
---

# Drawparticipationservice

52 symbols | 9 files | Cohesion: 72%

## When to Use

- Working with code in `rpc/`
- Understanding how NewQueryDrawActivityLandingLogic, NewPreviewDrawEligibilityLogic, NewParticipateDrawLogic work
- Modifying drawparticipationservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/sms/internal/logic/drawparticipationservice/draw_participation_helper.go` | loadActivitySnapshot, loadMemberInfoSnapshot, loadMemberIdentitySnapshot, defaultMemberIdentitySnapshot, loadPoolSnapshots (+21) |
| `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go` | TestBuildEligibilityDoesNotRequireRealNameForDraw, newDrawParticipationSvc, TestParticipateDrawPassesScopeAndReturnsRejectedWhenRuleFails, TestParticipateDrawCreatesAssetSnapshotForWinningRecordAndKeepsRequestIdempotent, newDrawParticipationTestDB (+3) |
| `rpc/sms/internal/server/drawparticipationservice/draw_participation_service_server.go` | QueryDrawActivityLanding, PreviewDrawEligibility, ParticipateDraw, QueryMemberDrawRecordList |
| `rpc/sms/client/drawparticipationservice/draw_participation_service.go` | QueryDrawActivityLanding, PreviewDrawEligibility, ParticipateDraw, QueryMemberDrawRecordList |
| `rpc/sms/internal/logic/drawparticipationservice/query_draw_activity_landing_logic.go` | NewQueryDrawActivityLandingLogic, QueryDrawActivityLanding |
| `rpc/sms/internal/logic/drawparticipationservice/preview_draw_eligibility_logic.go` | NewPreviewDrawEligibilityLogic, PreviewDrawEligibility |
| `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go` | NewParticipateDrawLogic, ParticipateDraw |
| `rpc/sms/smsclient/sms_grpc.pb.go` | NewDrawParticipationServiceClient, ParticipateDraw |
| `rpc/sms/internal/logic/drawparticipationservice/query_member_draw_record_list_logic.go` | NewQueryMemberDrawRecordListLogic, QueryMemberDrawRecordList |

## Entry Points

Start here when exploring this area:

- **`NewQueryDrawActivityLandingLogic`** (Function) — `rpc/sms/internal/logic/drawparticipationservice/query_draw_activity_landing_logic.go:22`
- **`NewPreviewDrawEligibilityLogic`** (Function) — `rpc/sms/internal/logic/drawparticipationservice/preview_draw_eligibility_logic.go:21`
- **`NewParticipateDrawLogic`** (Function) — `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go:25`
- **`TestBuildEligibilityDoesNotRequireRealNameForDraw`** (Function) — `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go:220`
- **`NewDrawParticipationServiceClient`** (Function) — `rpc/sms/smsclient/sms_grpc.pb.go:1782`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewQueryDrawActivityLandingLogic` | Function | `rpc/sms/internal/logic/drawparticipationservice/query_draw_activity_landing_logic.go` | 22 |
| `NewPreviewDrawEligibilityLogic` | Function | `rpc/sms/internal/logic/drawparticipationservice/preview_draw_eligibility_logic.go` | 21 |
| `NewParticipateDrawLogic` | Function | `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go` | 25 |
| `TestBuildEligibilityDoesNotRequireRealNameForDraw` | Function | `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go` | 220 |
| `NewDrawParticipationServiceClient` | Function | `rpc/sms/smsclient/sms_grpc.pb.go` | 1782 |
| `TestParticipateDrawPassesScopeAndReturnsRejectedWhenRuleFails` | Function | `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go` | 252 |
| `TestParticipateDrawCreatesAssetSnapshotForWinningRecordAndKeepsRequestIdempotent` | Function | `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go` | 321 |
| `NewQueryMemberDrawRecordListLogic` | Function | `rpc/sms/internal/logic/drawparticipationservice/query_member_draw_record_list_logic.go` | 19 |
| `TestLoadActivitySnapshotHonorsScope` | Function | `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go` | 194 |
| `TestCreateParticipationRecordSetsCreateTime` | Function | `rpc/sms/internal/logic/drawparticipationservice/draw_participation_logic_test.go` | 300 |
| `QueryDrawActivityLanding` | Method | `rpc/sms/internal/server/drawparticipationservice/draw_participation_service_server.go` | 25 |
| `PreviewDrawEligibility` | Method | `rpc/sms/internal/server/drawparticipationservice/draw_participation_service_server.go` | 30 |
| `ParticipateDraw` | Method | `rpc/sms/internal/server/drawparticipationservice/draw_participation_service_server.go` | 35 |
| `QueryDrawActivityLanding` | Method | `rpc/sms/internal/logic/drawparticipationservice/query_draw_activity_landing_logic.go` | 30 |
| `PreviewDrawEligibility` | Method | `rpc/sms/internal/logic/drawparticipationservice/preview_draw_eligibility_logic.go` | 29 |
| `ParticipateDraw` | Method | `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go` | 33 |
| `QueryDrawActivityLanding` | Method | `rpc/sms/client/drawparticipationservice/draw_participation_service.go` | 219 |
| `PreviewDrawEligibility` | Method | `rpc/sms/client/drawparticipationservice/draw_participation_service.go` | 224 |
| `ParticipateDraw` | Method | `rpc/sms/client/drawparticipationservice/draw_participation_service.go` | 229 |
| `QueryMemberDrawRecordList` | Method | `rpc/sms/client/drawparticipationservice/draw_participation_service.go` | 234 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Userservice | 10 calls |
| Cardassetservice | 5 calls |
| Productcategoryservice | 1 calls |
| Operatefunnel | 1 calls |
| Digitalcardmint | 1 calls |

## How to Explore

1. `gitnexus_context({name: "NewQueryDrawActivityLandingLogic"})` — see callers and callees
2. `gitnexus_query({query: "drawparticipationservice"})` — find related execution flows
3. Read key files listed above for implementation details
