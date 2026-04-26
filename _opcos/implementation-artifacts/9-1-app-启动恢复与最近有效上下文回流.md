# Story 9.1: App 启动直达首页与页面异常恢复退役

Status: review

## Scope Change / Retirement Notice

- 2026-04-26 决策：页面异常恢复 / 最近有效上下文回流 / pending recovery intent / active recovery candidate 功能退役。
- 原因：实际使用频率低、对调试干扰大，旧恢复状态黏连会让问题定位变复杂。
- 新基线：App 冷启动、热启动均直接进入 MainTab；不再恢复商品、订单、购物车等最近页面。
- 保留：权限 lost media、评价草稿、售后草稿、通知偏好等局部业务保存。
- 保留：消息页内显式点击、当前会话内手动导航仍可跳转目标页，不属于启动异常恢复。
- 禁止：恢复 store 再落盘 recent context / pending intent / active candidate / pending upgrade context。
- 禁止：恢复壳层重新接入 App 启动流程。

## Story

As a 移动端开发与测试人员,
I want App 启动时不再自动恢复上一次页面,
so that 调试、回归和问题定位不会被旧页面状态干扰。

## Acceptance Criteria

1. Given App 冷启动或热启动 When 启动完成 Then 直接进入 MainTab 默认首页，不读取 recent context / pending intent / active candidate。
2. Given 页面调用 AppRecoveryStore.saveRecentContext / savePendingIntent / saveActiveIntentCandidate / savePendingUpgradeContext When 这些方法被触发 Then 不落盘、不参与下次启动恢复，并清理对应旧键。
3. Given App 进入后台再恢复 When 生命周期回调发生 Then 不生成 active recovery candidate，不自动 push 上一次业务页。
4. Given 用户正在填写评价、售后或经历权限图片选择中断 When 启动清理页面恢复状态 Then 不清除 comment draft / after sales draft / pending permission lost media 等业务草稿。
5. Given 旧 BMAD story 仍引用 9.1 恢复底座 When 后续开发读取本文 Then 以本退役记录为准，不再把 IntentRecoveryShell / recent context 当强制基线。

## Tasks / Subtasks

- [x] 1. 启动入口改为直达 MainTab
  - [x] `flutter-mall/lib/layout/app_bootstrap.dart` 返回 `MainTab`
  - [x] 删除 `flutter-mall/lib/layout/intent_recovery_shell.dart`
- [x] 2. 页面恢复存储改为 no-op
  - [x] `AppRecoveryStore.saveRecentContext/getRecentContext` 不再持久化/读取
  - [x] pending intent / active candidate / pending upgrade context 不再持久化/读取
  - [x] 新增 `clearPageRecoveryState()`，仅清页面恢复键
- [x] 3. 生命周期不再写恢复候选
  - [x] `AppLifecycleProvider.didChangeAppLifecycleState` 只维护 resumeTick，不写 active candidate
- [x] 4. 保留局部业务草稿能力
  - [x] `clearRecoveryState()` 仍可清权限上下文、lost media、通知偏好和草稿
  - [x] 启动只调用 `clearPageRecoveryState()`，不误删业务草稿
- [x] 5. 更新测试语义
  - [x] recovery tests 改为断言恢复状态不持久化
  - [x] 售后/权限草稿测试继续覆盖
- [x] 6. 验证
  - [x] `flutter analyze ...`
  - [x] `flutter test`
  - [x] `flutter build apk --debug`

## Dev Notes

- 页面异常恢复功能已退役；不要重新引入 `IntentRecoveryShell` 作为启动壳层。
- `AppRecentContext` / `AppRecoveryRouter` 可暂留给消息页内显式意图、升级门闸等现有直接跳转兼容使用，但不得作为 App 启动恢复真相源。
- `AppRecoveryStore` 中页面恢复相关 API 保留为兼容壳，行为为 no-op / null，避免一次性重构所有调用点。
- 后续 story 若提到“9.1 恢复底座”“recent context 成功态写入”“pending recovery intent”，均视为历史设计，除非重新开新 story 明确恢复该能力。

## Dev Agent Record

### Agent Model Used

Codex

### Debug Log References

- `flutter analyze ...`
- `flutter test`
- `flutter build apk --debug`

### Completion Notes List

- 已按产品决策退役页面异常恢复，启动不再受旧页面状态影响。
- 已保留权限 lost media 与评价/售后草稿等实际有用的局部恢复能力。
- 已更新测试为“页面恢复不持久化”的新语义。

### File List

- `_opcos/implementation-artifacts/9-1-app-启动恢复与最近有效上下文回流.md`
- `flutter-mall/lib/layout/app_bootstrap.dart`
- `flutter-mall/lib/layout/intent_recovery_shell.dart`
- `flutter-mall/lib/main.dart`
- `flutter-mall/lib/provider/app_lifecycle_provider.dart`
- `flutter-mall/lib/utils/app_recovery_store.dart`
- `flutter-mall/test/app_bootstrap_test.dart`
- `flutter-mall/test/recent_context_test.dart`
- `flutter-mall/test/login_restore_test.dart`
- `flutter-mall/test/intent_recovery_shell_test.dart`
- `flutter-mall/test/apply_after_sales_permission_test.dart`
- `flutter-mall/test/widget_test.dart`
