---
name: digitalcardmint
description: "Skill for the Digitalcardmint area of zero-admin. 154 symbols across 17 files."
---

# Digitalcardmint

154 symbols | 17 files | Cohesion: 70%

## When to Use

- Working with code in `pkg/`
- Understanding how TestEnsureTaskTxIsIdempotent, TestEnsureTaskTxRejectsForbiddenPrerequisites, TestEnsureTaskTxUsesCurrentRealNameStatus work
- Modifying digitalcardmint-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `pkg/digitalcardmint/service.go` | now, DispatchTask, ExecuteTask, QueryTaskDetail, QueryAvailableActions (+51) |
| `pkg/digitalcardmint/physical_fulfillment.go` | TableName, EnsurePhysicalFulfillmentByAsset, mutatePhysicalFulfillment, appendPhysicalFulfillmentLogTx, applyPhysicalAddressSnapshot (+16) |
| `pkg/digitalcardmint/service_test.go` | newDigitalCardMintTestDB, TestEnsureTaskTxIsIdempotent, TestEnsureTaskTxRejectsForbiddenPrerequisites, TestEnsureTaskTxUsesCurrentRealNameStatus, TestDispatchTaskSkipsTerminalTask (+11) |
| `pkg/digitalcardmint/asset_service.go` | HandleRefundCardDispose, QueryMemberDigitalCardAssetList, QueryMemberDigitalCardAssetDetail, memberAssetSelectColumns, buildMemberAssetItem (+8) |
| `pkg/digitalcardmint/asset_service_test.go` | newAssetServiceTestDB, TestQueryMemberDigitalCardAssetDetailReturnsTimelineWithoutToken, TestEnsureOrderPurchaseAssetCreatesIdempotentAsset, TestHandleRefundCardDisposeFreezeCard, merchantScope (+4) |
| `pkg/digitalcardmint/asset_transfer.go` | TransferDigitalCardAsset, RequestDigitalCardAssetWithdraw, validateOwnedAssetAction, validateTransferableAsset, validateAssetWithdrawable (+3) |
| `pkg/digitalcardmint/physical_fulfillment_test.go` | preparePhysicalFulfillmentTestDB, physicalTestScope, TestEnsurePhysicalFulfillmentByAssetCreatesPendingAddressOnce, TestEnsurePhysicalFulfillmentBlocksDigitalPendingAndRealName, TestConfirmAddressSnapshotsAndMasksMemberDetail (+2) |
| `pkg/digitalcardmint/asset_constants.go` | displayStatusText, complianceStatusText, tokenStatusText, complianceRuleSummary |
| `pkg/digitalcardmint/asset_transfer_test.go` | prepareAssetTransferTestDB, TestResolveAssetTransferRecipientReturnsRegisterH5ForUnknownMobile, TestTransferDigitalCardAssetMovesOwnershipToRegisteredMember, TestRequestDigitalCardAssetWithdrawAppendsAuditLog |
| `pkg/digitalcardmint/order_purchase_asset.go` | EnsureOrderPurchaseAsset, ensureOrderPurchaseAssetTx, ensureTaskTxForAsset |

## Entry Points

Start here when exploring this area:

- **`TestEnsureTaskTxIsIdempotent`** (Function) — `pkg/digitalcardmint/service_test.go:272`
- **`TestEnsureTaskTxRejectsForbiddenPrerequisites`** (Function) — `pkg/digitalcardmint/service_test.go:318`
- **`TestEnsureTaskTxUsesCurrentRealNameStatus`** (Function) — `pkg/digitalcardmint/service_test.go:379`
- **`TestDispatchTaskSkipsTerminalTask`** (Function) — `pkg/digitalcardmint/service_test.go:398`
- **`TestScanDueTasksReplaysStaleDispatchedTaskWithoutExecuteLease`** (Function) — `pkg/digitalcardmint/service_test.go:451`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestEnsureTaskTxIsIdempotent` | Function | `pkg/digitalcardmint/service_test.go` | 272 |
| `TestEnsureTaskTxRejectsForbiddenPrerequisites` | Function | `pkg/digitalcardmint/service_test.go` | 318 |
| `TestEnsureTaskTxUsesCurrentRealNameStatus` | Function | `pkg/digitalcardmint/service_test.go` | 379 |
| `TestDispatchTaskSkipsTerminalTask` | Function | `pkg/digitalcardmint/service_test.go` | 398 |
| `TestScanDueTasksReplaysStaleDispatchedTaskWithoutExecuteLease` | Function | `pkg/digitalcardmint/service_test.go` | 451 |
| `TestExecuteTaskSkipsLeasedRunningTask` | Function | `pkg/digitalcardmint/service_test.go` | 523 |
| `TestExecuteTaskMovesToManualReviewWhenTokenAlreadyBound` | Function | `pkg/digitalcardmint/service_test.go` | 583 |
| `TestExecuteTaskSuccessWritesBackTokenAndLogs` | Function | `pkg/digitalcardmint/service_test.go` | 656 |
| `TestExecuteTaskReplaysStoredReceiptAfterWritebackFailure` | Function | `pkg/digitalcardmint/service_test.go` | 711 |
| `TestExecuteTaskEscalatesToManualReviewAtRetryLimit` | Function | `pkg/digitalcardmint/service_test.go` | 784 |
| `TestExecuteTaskRevalidatesPrerequisitesBeforeMint` | Function | `pkg/digitalcardmint/service_test.go` | 839 |
| `TestExecuteTaskReconcilesReceiptBeforeRetryMint` | Function | `pkg/digitalcardmint/service_test.go` | 889 |
| `TestExecuteTaskEscalatesWhenReceiptCannotBeReconciled` | Function | `pkg/digitalcardmint/service_test.go` | 957 |
| `TestRetryTaskReturnsDispatchError` | Function | `pkg/digitalcardmint/service_test.go` | 1022 |
| `TestFreezeTaskRequiresReason` | Function | `pkg/digitalcardmint/service_test.go` | 1050 |
| `NewService` | Function | `pkg/digitalcardmint/service.go` | 56 |
| `NewMockClient` | Function | `pkg/antchain/mock.go` | 23 |
| `TestHandleCardMintTimeoutExecutesPendingDispatchWithoutMQ` | Function | `job/internal/jobs/handle_card_mint_timeout_logic_test.go` | 207 |
| `TestHandleCardMintTimeoutEscalatesFailedTaskAtRetryLimit` | Function | `job/internal/jobs/handle_card_mint_timeout_logic_test.go` | 221 |
| `HandleCardMintTimeout` | Function | `job/internal/jobs/handle_card_mint_timeout_logic.go` | 9 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Drawactivityservice | 14 calls |
| Userservice | 3 calls |
| Cardassetservice | 3 calls |

## How to Explore

1. `gitnexus_context({name: "TestEnsureTaskTxIsIdempotent"})` — see callers and callees
2. `gitnexus_query({query: "digitalcardmint"})` — find related execution flows
3. Read key files listed above for implementation details
