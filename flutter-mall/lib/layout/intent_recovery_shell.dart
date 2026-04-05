import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:provider/provider.dart';

typedef RecoveryTargetBuilder = Widget Function(AppRecentContext context);
typedef RecoveryFallbackBuilder = Widget Function(AppRecentContext? context);
typedef RecoveryLoginBuilder = Widget Function(AppRecentContext intent);

class IntentRecoveryShell extends StatefulWidget {
  final Widget splash;
  final Duration decisionTimeout;
  final RecoveryTargetBuilder? targetBuilder;
  final RecoveryFallbackBuilder? fallbackBuilder;
  final RecoveryLoginBuilder? loginBuilder;

  const IntentRecoveryShell({
    super.key,
    required this.splash,
    this.decisionTimeout = const Duration(seconds: 3),
    this.targetBuilder,
    this.fallbackBuilder,
    this.loginBuilder,
  });

  @override
  State<IntentRecoveryShell> createState() => _IntentRecoveryShellState();
}

class _IntentRecoveryShellState extends State<IntentRecoveryShell> {
  bool _handled = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _resolveInitialIntent();
    });
  }

  Future<void> _resolveInitialIntent() async {
    if (!mounted || _handled) {
      return;
    }
    _handled = true;
    final lifecycleProvider = context.read<AppLifecycleProvider>();
    lifecycleProvider.markRestoreStarted();

    final result = await _buildDecision().timeout(
      widget.decisionTimeout,
      onTimeout: () => const _RecoveryDecision.fallback(
        message: '恢复超时，已为你返回首页',
      ),
    );

    if (!mounted) {
      return;
    }

    switch (result.action) {
      case _RecoveryAction.login:
        lifecycleProvider.markRestoreFailed('login_required');
        await AppRecoveryStore.savePendingIntent(result.context!);
        if (!mounted) {
          return;
        }
        await Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) => widget.loginBuilder?.call(result.context!) ?? Login(recoveryIntent: result.context!),
          ),
        );
        return;
      case _RecoveryAction.target:
        lifecycleProvider.markRestoreSucceeded();
        if (widget.targetBuilder != null) {
          if (!mounted) {
            return;
          }
          await Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => widget.targetBuilder!(result.context!)),
          );
          return;
        }
        await AppRecoveryRouter.replaceWithTarget(
          context,
          result.context!,
          message: result.message,
        );
        return;
      case _RecoveryAction.fallback:
        lifecycleProvider.markFallbackUsed(result.reason);
        if (widget.fallbackBuilder != null) {
          await Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => widget.fallbackBuilder!(result.context)),
          );
          return;
        }
        await AppRecoveryRouter.replaceWithFallback(
          context,
          recentContext: result.context,
          message: result.message,
        );
        return;
    }
  }

  Future<_RecoveryDecision> _buildDecision() async {
    final pendingIntent = AppRecoveryStore.peekPendingIntent();
    final activeIntentCandidate = AppRecoveryStore.peekActiveIntentCandidate();
    final recentContext = AppRecoveryStore.getRecentContext();
    final candidate = pendingIntent ?? activeIntentCandidate ?? recentContext;

    if (candidate == null) {
      return const _RecoveryDecision.fallback();
    }

    final invalidReason = candidate.validateReason();
    if (invalidReason != null) {
      if (pendingIntent != null) {
        await AppRecoveryStore.clearPendingIntent();
      }
      if (activeIntentCandidate != null) {
        await AppRecoveryStore.clearActiveIntentCandidate();
      }
      if (pendingIntent == null && activeIntentCandidate == null) {
        await AppRecoveryStore.clearRecentContext();
      }
      return _RecoveryDecision.fallback(
        context: candidate,
        reason: invalidReason,
        message: '最近上下文已失效，已返回可用页面',
      );
    }

    if (candidate.requiresAuth && !AppRecoveryStore.hasValidToken()) {
      return _RecoveryDecision.login(candidate);
    }

    if (pendingIntent != null) {
      await AppRecoveryStore.clearPendingIntent();
    } else if (activeIntentCandidate != null) {
      await AppRecoveryStore.clearActiveIntentCandidate();
    }

    return _RecoveryDecision.target(candidate);
  }

  @override
  Widget build(BuildContext context) {
    return widget.splash;
  }
}

enum _RecoveryAction { login, target, fallback }

class _RecoveryDecision {
  final _RecoveryAction action;
  final AppRecentContext? context;
  final String? reason;
  final String? message;

  const _RecoveryDecision._({
    required this.action,
    this.context,
    this.reason,
    this.message,
  });

  const _RecoveryDecision.login(AppRecentContext context)
      : this._(action: _RecoveryAction.login, context: context);

  const _RecoveryDecision.target(AppRecentContext context)
      : this._(action: _RecoveryAction.target, context: context);

  const _RecoveryDecision.fallback({
    AppRecentContext? context,
    String? reason,
    String? message,
  }) : this._(
          action: _RecoveryAction.fallback,
          context: context,
          reason: reason,
          message: message,
        );
}
