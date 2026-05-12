---
name: test
description: "Skill for the Test area of zero-admin. 52 symbols across 27 files."
---

# Test

52 symbols | 27 files | Cohesion: 63%

## When to Use

- Working with code in `flutter-mall/`
- Understanding how main, main, main work
- Modifying test-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `flutter-mall/lib/utils/permission_broker.dart` | openSettingsForFlow, captureLostMediaOnLaunch, PermissionInspectionResult, inspect, request (+8) |
| `flutter-mall/lib/utils/app_recovery_store.dart` | AppRecoveryStore, savePendingPermissionContext, savePendingLostMedia, saveNotificationPreferenceEnabled, saveCommentDraft (+1) |
| `flutter-mall/lib/model/app_version_policy.dart` | AppVersionPolicy, copyWith, applyRecallRequirement |
| `flutter-mall/test/ping_jia_permission_test.dart` | main, _NoopAdapter |
| `flutter-mall/test/apply_after_sales_permission_test.dart` | main, _NoopAdapter |
| `flutter-mall/lib/utils/upgrade_gate_service.dart` | debugSetPolicyFetcher, debugReset |
| `flutter-mall/lib/utils/shared_preferences_util.dart` | init, saveJsonString |
| `flutter-mall/lib/utils/app_version_service.dart` | debugSetCurrentInfo, debugReset |
| `flutter-mall/lib/utils/app_intent_dispatcher.dart` | AppIntentDispatcher, resolve |
| `flutter-mall/test/version_policy_test.dart` | main |

## Entry Points

Start here when exploring this area:

- **`main`** (Function) — `flutter-mall/test/version_policy_test.dart:6`
- **`main`** (Function) — `flutter-mall/test/settings_upgrade_gate_test.dart:12`
- **`main`** (Function) — `flutter-mall/test/ping_jia_permission_test.dart:12`
- **`main`** (Function) — `flutter-mall/test/intent_recovery_shell_test.dart:10`
- **`main`** (Function) — `flutter-mall/test/apply_after_sales_permission_test.dart:16`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `AppLifecycleProvider` | Class | `flutter-mall/lib/provider/app_lifecycle_provider.dart` | 67 |
| `AppVersionPolicy` | Class | `flutter-mall/lib/model/app_version_policy.dart` | 24 |
| `AppBootstrap` | Class | `flutter-mall/lib/layout/app_bootstrap.dart` | 3 |
| `AppRecoveryStore` | Class | `flutter-mall/lib/utils/app_recovery_store.dart` | 8 |
| `PermissionInspectionResult` | Class | `flutter-mall/lib/utils/permission_broker.dart` | 18 |
| `DrawMemberRecord` | Class | `flutter-mall/lib/model/digital_card/draw_activity_model.dart` | 268 |
| `DigitalCardAssetItem` | Class | `flutter-mall/lib/model/digital_card/digital_card_asset_model.dart` | 79 |
| `AppIntentDispatcher` | Class | `flutter-mall/lib/utils/app_intent_dispatcher.dart` | 30 |
| `PermissionBroker` | Class | `flutter-mall/lib/utils/permission_broker.dart` | 101 |
| `Settings` | Class | `flutter-mall/lib/view/mine/setting/settings.dart` | 21 |
| `PermissionBrokerAdapter` | Class | `flutter-mall/lib/utils/permission_broker.dart` | 77 |
| `main` | Function | `flutter-mall/test/version_policy_test.dart` | 6 |
| `main` | Function | `flutter-mall/test/settings_upgrade_gate_test.dart` | 12 |
| `main` | Function | `flutter-mall/test/ping_jia_permission_test.dart` | 12 |
| `main` | Function | `flutter-mall/test/intent_recovery_shell_test.dart` | 10 |
| `main` | Function | `flutter-mall/test/apply_after_sales_permission_test.dart` | 16 |
| `main` | Function | `flutter-mall/test/app_bootstrap_test.dart` | 7 |
| `main` | Function | `flutter-mall/test/permission_recovery_test.dart` | 8 |
| `main` | Function | `flutter-mall/test/permission_broker_test.dart` | 8 |
| `main` | Function | `flutter-mall/test/draw_activity_model_test.dart` | 3 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Build → AppRecoveryStore` | cross_community | 6 |
| `Build → AppRecoveryStore` | cross_community | 5 |
| `Build → AppRecoveryStore` | cross_community | 5 |
| `Main → PermissionFlowContext` | cross_community | 4 |
| `Main → SaveJsonString` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Model | 7 calls |
| Digital_card | 4 calls |
| Ping_jia | 2 calls |
| Coupon | 2 calls |
| Cluster_647 | 1 calls |
| Setting | 1 calls |

## How to Explore

1. `gitnexus_context({name: "main"})` — see callers and callees
2. `gitnexus_query({query: "test"})` — find related execution flows
3. Read key files listed above for implementation details
