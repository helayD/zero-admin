---
name: message
description: "Skill for the Message area of zero-admin. 52 symbols across 21 files."
---

# Message

52 symbols | 21 files | Cohesion: 63%

## When to Use

- Working with code in `api/`
- Understanding how ReadClientRequestMetadata, WithRecallRequestMetadata, NewMessageListLogic work
- Modifying message-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `flutter-mall/lib/view/mine/message/message.dart` | _loadMessages, _loadMore, _markAllAsRead, build, _buildTabBar (+8) |
| `api/front/internal/logic/member/message/recall_intent_helper.go` | WithRecallRequestMetadata, messageIntentResponseFromRPC, recallTabPointer |
| `api/front/internal/logic/member/message/message_recall_logic_test.go` | newFrontMessageTestContext, TestMessageListMapsStructuredRecallIntent, TestMessageListBlocksRecallWhenAppVersionIsBelowMinimum |
| `flutter-mall/lib/utils/http_util.dart` | delete, postForm, _buildHeaders |
| `flutter-mall/lib/utils/upgrade_gate_service.dart` | recoverySceneFor, queryPendingPolicy, queryRecallPolicy |
| `api/front/internal/handler/member/message/path_handler_test.go` | TestMarkMessageReadByPathAcceptsJsonNullBody, TestDeleteMessageByPathAcceptsJsonNullBody, newPathRequest |
| `api/front/internal/logic/common/client_metadata.go` | ReadClientRequestMetadata, CompareAppVersion |
| `api/front/internal/logic/member/message/messagelistlogic.go` | NewMessageListLogic, MessageList |
| `api/front/internal/logic/member/message/markmessagereadlogic.go` | NewMarkMessageReadLogic, MarkMessageRead |
| `api/front/internal/handler/member/message/markmessagereadhandler.go` | MarkMessageReadHandler, MarkMessageReadByPathHandler |

## Entry Points

Start here when exploring this area:

- **`ReadClientRequestMetadata`** (Function) — `api/front/internal/logic/common/client_metadata.go:20`
- **`WithRecallRequestMetadata`** (Function) — `api/front/internal/logic/member/message/recall_intent_helper.go:34`
- **`NewMessageListLogic`** (Function) — `api/front/internal/logic/member/message/messagelistlogic.go:25`
- **`TestMessageListMapsStructuredRecallIntent`** (Function) — `api/front/internal/logic/member/message/message_recall_logic_test.go:33`
- **`TestMessageListBlocksRecallWhenAppVersionIsBelowMinimum`** (Function) — `api/front/internal/logic/member/message/message_recall_logic_test.go:90`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `ReadClientRequestMetadata` | Function | `api/front/internal/logic/common/client_metadata.go` | 20 |
| `WithRecallRequestMetadata` | Function | `api/front/internal/logic/member/message/recall_intent_helper.go` | 34 |
| `NewMessageListLogic` | Function | `api/front/internal/logic/member/message/messagelistlogic.go` | 25 |
| `TestMessageListMapsStructuredRecallIntent` | Function | `api/front/internal/logic/member/message/message_recall_logic_test.go` | 33 |
| `TestMessageListBlocksRecallWhenAppVersionIsBelowMinimum` | Function | `api/front/internal/logic/member/message/message_recall_logic_test.go` | 90 |
| `MessageListHandler` | Function | `api/front/internal/handler/member/message/messagelisthandler.go` | 11 |
| `NewMarkMessageReadLogic` | Function | `api/front/internal/logic/member/message/markmessagereadlogic.go` | 25 |
| `MarkMessageReadHandler` | Function | `api/front/internal/handler/member/message/markmessagereadhandler.go` | 11 |
| `MarkMessageReadByPathHandler` | Function | `api/front/internal/handler/member/message/markmessagereadhandler.go` | 29 |
| `NewDeleteMessageLogic` | Function | `api/front/internal/logic/member/message/deletemessagelogic.go` | 25 |
| `DeleteMessageHandler` | Function | `api/front/internal/handler/member/message/deletemessagehandler.go` | 11 |
| `DeleteMessageByPathHandler` | Function | `api/front/internal/handler/member/message/deletemessagehandler.go` | 29 |
| `CompareAppVersion` | Function | `api/front/internal/logic/common/client_metadata.go` | 57 |
| `NewQueryUnreadCountLogic` | Function | `api/front/internal/logic/member/message/queryunreadcountlogic.go` | 25 |
| `QueryUnreadCountHandler` | Function | `api/front/internal/handler/member/message/queryunreadcounthandler.go` | 11 |
| `NewMessageDetailLogic` | Function | `api/front/internal/logic/member/message/messagedetaillogic.go` | 25 |
| `MessageDetailHandler` | Function | `api/front/internal/handler/member/message/messagedetailhandler.go` | 11 |
| `NewMarkAllMessagesReadLogic` | Function | `api/front/internal/logic/member/message/markallmessagesreadlogic.go` | 25 |
| `MarkAllMessagesReadHandler` | Function | `api/front/internal/handler/member/message/markallmessagesreadhandler.go` | 11 |
| `TestMarkMessageReadByPathAcceptsJsonNullBody` | Function | `api/front/internal/handler/member/message/path_handler_test.go` | 30 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Build → Init` | cross_community | 7 |
| `Build → Init` | cross_community | 6 |
| `Build → Init` | cross_community | 6 |
| `Build → AppRecoveryStore` | cross_community | 6 |
| `Build → PeekCurrentIntentContext` | cross_community | 6 |
| `MessageListHandler → CodeError` | cross_community | 6 |
| `MessageDetailHandler → CodeError` | cross_community | 6 |
| `RegisterHandlers → SplitAppVersion` | cross_community | 5 |
| `Build → AppRecoveryStore` | cross_community | 5 |
| `Build → PeekCurrentIntentContext` | cross_community | 5 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Order | 13 calls |
| Provider | 6 calls |
| Home | 2 calls |
| Setting | 1 calls |
| Test | 1 calls |
| Category | 1 calls |
| Widgets | 1 calls |
| Version | 1 calls |

## How to Explore

1. `gitnexus_context({name: "ReadClientRequestMetadata"})` — see callers and callees
2. `gitnexus_query({query: "message"})` — find related execution flows
3. Read key files listed above for implementation details
