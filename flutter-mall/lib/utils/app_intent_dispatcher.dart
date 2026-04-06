import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_recent_context.dart';

enum AppIntentDispatchAction {
  login,
  target,
  fallback,
}

class AppIntentDispatchPlan {
  final AppIntentDispatchAction action;
  final AppRecentContext intent;
  final String? message;
  final String? failureReason;
  final bool shouldMarkMessageRead;

  const AppIntentDispatchPlan({
    required this.action,
    required this.intent,
    this.message,
    this.failureReason,
    this.shouldMarkMessageRead = false,
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
        action: AppIntentDispatchAction.fallback,
        intent: intent.copyWith(
          blocked: true,
          failureReason: minVersionFailure,
        ),
        message: intent.recoveryHint.isNotEmpty
            ? intent.recoveryHint
            : '当前版本暂不支持直达该入口，已为你返回可用页面',
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
    if (_compareVersion(appVersion, intent.minAppVersion) >= 0) {
      return null;
    }
    return 'min_version_unmet';
  }

  static int _compareVersion(String current, String minimum) {
    final currentParts = _splitVersion(current);
    final minimumParts = _splitVersion(minimum);
    final maxLength = currentParts.length > minimumParts.length
        ? currentParts.length
        : minimumParts.length;
    for (var i = 0; i < maxLength; i++) {
      final currentValue = i < currentParts.length ? currentParts[i] : 0;
      final minimumValue = i < minimumParts.length ? minimumParts[i] : 0;
      if (currentValue > minimumValue) {
        return 1;
      }
      if (currentValue < minimumValue) {
        return -1;
      }
    }
    return 0;
  }

  static List<int> _splitVersion(String value) {
    return value.trim().split('.').where((part) => part.isNotEmpty).map((part) {
      final match = RegExp(r'^(\d+)').firstMatch(part.trim());
      if (match == null) {
        return 0;
      }
      return int.tryParse(match.group(1) ?? '0') ?? 0;
    }).toList();
  }
}
