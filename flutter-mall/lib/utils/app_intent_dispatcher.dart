import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/version_compare.dart';

enum AppIntentDispatchAction {
  login,
  target,
  fallback,
  upgradeGate,
}

class AppIntentDispatchPlan {
  final AppIntentDispatchAction action;
  final AppRecentContext intent;
  final String? message;
  final String? failureReason;
  final bool shouldMarkMessageRead;
  final AppVersionPolicy? upgradePolicy;

  const AppIntentDispatchPlan({
    required this.action,
    required this.intent,
    this.message,
    this.failureReason,
    this.shouldMarkMessageRead = false,
    this.upgradePolicy,
  });
}

class AppIntentDispatcher {
  static AppIntentDispatchPlan resolve(
    AppRecentContext intent, {
    required bool hasValidToken,
  }) {
    final invalidReason = intent.validateReason();
    if (invalidReason != null) {
      return AppIntentDispatchPlan(
        action: AppIntentDispatchAction.fallback,
        intent: intent,
        message: '消息目标无效，已返回可用页面',
        failureReason: invalidReason,
      );
    }

    if (intent.requiresAuth && !hasValidToken) {
      return AppIntentDispatchPlan(
        action: AppIntentDispatchAction.login,
        intent: intent,
        failureReason: 'login_required',
      );
    }

    final minVersionFailure = _resolveMinVersionFailure(intent);
    if (minVersionFailure != null) {
      return AppIntentDispatchPlan(
        action: AppIntentDispatchAction.upgradeGate,
        intent: intent.copyWith(
          blocked: true,
          failureReason: minVersionFailure,
        ),
        message: intent.recoveryHint.isNotEmpty
            ? intent.recoveryHint
            : '当前版本暂不支持该入口，请先升级后继续',
        failureReason: minVersionFailure,
        shouldMarkMessageRead: true,
      );
    }

    if (intent.blocked || intent.failureReason.trim().isNotEmpty) {
      return AppIntentDispatchPlan(
        action: AppIntentDispatchAction.fallback,
        intent: intent,
        message: intent.recoveryHint.isNotEmpty
            ? intent.recoveryHint
            : '当前入口暂不可直达，已为你返回可用页面',
        failureReason: intent.failureReason,
        shouldMarkMessageRead: intent.failureReason != 'invalid_payload',
      );
    }

    return AppIntentDispatchPlan(
      action: AppIntentDispatchAction.target,
      intent: intent,
      shouldMarkMessageRead: true,
    );
  }

  static String? _resolveMinVersionFailure(AppRecentContext intent) {
    if (intent.minAppVersion.trim().isEmpty) {
      return null;
    }
    if (compareVersion(
          AppVersionService.currentInfo.version,
          intent.minAppVersion,
        ) >=
        0) {
      return null;
    }
    return 'min_version_unmet';
  }
}
