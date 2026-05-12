---
name: cardassetservice
description: "Skill for the Cardassetservice area of zero-admin. 57 symbols across 12 files."
---

# Cardassetservice

57 symbols | 12 files | Cohesion: 67%

## When to Use

- Working with code in `rpc/`
- Understanding how TestEnsureCardInstanceByParticipationRecordCreatesLedgerAndSnapshot, TestEnsureCardInstanceByParticipationRecordIsIdempotent, TestEnsureCardInstanceByParticipationRecordRetriesAssetNoCollision work
- Modifying cardassetservice-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `rpc/sms/internal/logic/cardassetservice/card_asset_helper.go` | buildCardInstanceData, TableName, loadCardInstanceByParticipationRecord, ensureInitialAssetLog, createCardInstanceWithRetry (+13) |
| `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | newCardAssetTestDB, seedWinningParticipationRecord, newMerchantScope, TestEnsureCardInstanceByParticipationRecordCreatesLedgerAndSnapshot, TestEnsureCardInstanceByParticipationRecordIsIdempotent (+8) |
| `rpc/sms/internal/logic/cardassetservice/ensure_order_purchase_card_instance_logic.go` | EnsureOrderPurchaseCardInstance, TableName, loadFulfillmentRule, ensureOrderPurchaseAssetLog, ensureOrderPurchaseCardInstance (+2) |
| `rpc/sms/smsclient/sms_grpc.pb.go` | EnsureCardInstanceByParticipationRecord, BackfillWinningCardInstances, NewCardAssetServiceClient |
| `rpc/sms/internal/server/cardassetservice/card_asset_service_server.go` | EnsureCardInstanceByParticipationRecord, BackfillWinningCardInstances, QueryCardInstanceByParticipationRecord |
| `rpc/sms/client/cardassetservice/card_asset_service.go` | EnsureCardInstanceByParticipationRecord, QueryCardInstanceByParticipationRecord, BackfillWinningCardInstances |
| `rpc/sms/internal/logic/cardminttaskservice/card_mint_task_helper.go` | EnsureCardMintTaskByAssetInstance, DispatchCardMintTask |
| `rpc/sms/internal/logic/cardassetservice/ensure_card_instance_by_participation_record_logic.go` | NewEnsureCardInstanceByParticipationRecordLogic, EnsureCardInstanceByParticipationRecord |
| `rpc/sms/internal/logic/cardassetservice/backfill_winning_card_instances_logic.go` | NewBackfillWinningCardInstancesLogic, BackfillWinningCardInstances |
| `rpc/sms/internal/logic/cardassetservice/query_card_instance_by_participation_record_logic.go` | NewQueryCardInstanceByParticipationRecordLogic, QueryCardInstanceByParticipationRecord |

## Entry Points

Start here when exploring this area:

- **`TestEnsureCardInstanceByParticipationRecordCreatesLedgerAndSnapshot`** (Function) — `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go:147`
- **`TestEnsureCardInstanceByParticipationRecordIsIdempotent`** (Function) — `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go:187`
- **`TestEnsureCardInstanceByParticipationRecordRetriesAssetNoCollision`** (Function) — `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go:212`
- **`TestEnsureCardInstanceByParticipationRecordRejectsNonWinningRecord`** (Function) — `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go:246`
- **`TestEnsureCardInstanceByParticipationRecordRollsBackWithTransaction`** (Function) — `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go:267`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestEnsureCardInstanceByParticipationRecordCreatesLedgerAndSnapshot` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 147 |
| `TestEnsureCardInstanceByParticipationRecordIsIdempotent` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 187 |
| `TestEnsureCardInstanceByParticipationRecordRetriesAssetNoCollision` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 212 |
| `TestEnsureCardInstanceByParticipationRecordRejectsNonWinningRecord` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 246 |
| `TestEnsureCardInstanceByParticipationRecordRollsBackWithTransaction` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 267 |
| `TestBackfillWinningCardInstancesIsReentrant` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 298 |
| `TestBackfillWinningCardInstancesDoesNotProcessNonWinningRecord` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 327 |
| `TestEnsureCardInstanceByParticipationRecordLogicRejectsScopeMismatch` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 352 |
| `TestQueryCardInstanceByParticipationRecordLogicRejectsScopeMismatch` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 368 |
| `TestBackfillWinningCardInstancesLogicFiltersByScope` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_logic_test.go` | 386 |
| `EnsureCardMintTaskByAssetInstance` | Function | `rpc/sms/internal/logic/cardminttaskservice/card_mint_task_helper.go` | 10 |
| `DispatchCardMintTask` | Function | `rpc/sms/internal/logic/cardminttaskservice/card_mint_task_helper.go` | 24 |
| `NewEnsureCardInstanceByParticipationRecordLogic` | Function | `rpc/sms/internal/logic/cardassetservice/ensure_card_instance_by_participation_record_logic.go` | 19 |
| `NewBackfillWinningCardInstancesLogic` | Function | `rpc/sms/internal/logic/cardassetservice/backfill_winning_card_instances_logic.go` | 19 |
| `ResolveAssetStatusText` | Function | `pkg/digitalcardmint/constants.go` | 60 |
| `QueryCardInstanceByParticipationRecord` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_helper.go` | 184 |
| `NewCardAssetServiceClient` | Function | `rpc/sms/smsclient/sms_grpc.pb.go` | 39 |
| `EnsureCardInstanceByParticipationRecord` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_helper.go` | 142 |
| `BackfillWinningCardInstances` | Function | `rpc/sms/internal/logic/cardassetservice/card_asset_helper.go` | 195 |
| `NewQueryCardInstanceByParticipationRecordLogic` | Function | `rpc/sms/internal/logic/cardassetservice/query_card_instance_by_participation_record_logic.go` | 17 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Userservice | 5 calls |
| Scope | 3 calls |
| Digitalcardmint | 1 calls |

## How to Explore

1. `gitnexus_context({name: "TestEnsureCardInstanceByParticipationRecordCreatesLedgerAndSnapshot"})` — see callers and callees
2. `gitnexus_query({query: "cardassetservice"})` — find related execution flows
3. Read key files listed above for implementation details
