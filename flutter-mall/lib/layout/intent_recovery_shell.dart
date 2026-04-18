import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_mall/layout/upgrade_gate_page.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_intent_dispatcher.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/view/mine/order/order_pay.dart';
import 'package:flutter_mall/view/mine/order/order_submit.dart';
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
        lifecycleProvider.recordIntentLoginRequired(result.context!);
        await AppRecoveryStore.savePendingIntent(result.context!);
        if (!mounted) {
          return;
        }
        await Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) =>
                widget.loginBuilder?.call(result.context!) ??
                Login(recoveryIntent: result.context!),
          ),
        );
        return;
      case _RecoveryAction.target:
        lifecycleProvider.markRestoreSucceeded();
        lifecycleProvider.recordIntentRestored(result.context!);
        if (widget.targetBuilder != null) {
          if (!mounted) {
            return;
          }
          await Navigator.of(context).pushReplacement(
            MaterialPageRoute(
                builder: (_) => widget.targetBuilder!(result.context!)),
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
        if (result.context != null) {
          lifecycleProvider.recordIntentFallbackUsed(
            result.context!,
            failureReason: result.reason,
          );
        }
        if (widget.fallbackBuilder != null) {
          await Navigator.of(context).pushReplacement(
            MaterialPageRoute(
                builder: (_) => widget.fallbackBuilder!(result.context)),
          );
          return;
        }
        await AppRecoveryRouter.replaceWithFallback(
          context,
          recentContext: result.context,
          message: result.message,
        );
        return;
      case _RecoveryAction.upgradeGate:
        final pendingUpgradeContext = result.pendingUpgradeContext;
        final upgradePolicy = result.upgradePolicy;
        if (pendingUpgradeContext == null || upgradePolicy == null) {
          await AppRecoveryRouter.replaceWithFallback(
            context,
            recentContext: result.context,
            message: '升级门闸上下文缺失，已返回可用页面',
          );
          return;
        }
        await AppRecoveryStore.savePendingUpgradeContext(pendingUpgradeContext);
        if (!mounted) {
          return;
        }
        await Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) => UpgradeGatePage(
              pendingContext: pendingUpgradeContext,
              initialPolicy: upgradePolicy,
              onResolved: _resumePendingUpgrade,
              onContinueLater: upgradePolicy.canContinueLater
                  ? _continuePendingUpgrade
                  : null,
            ),
          ),
        );
        return;
    }
  }

  Future<_RecoveryDecision> _buildDecision() async {
    final pendingUpgradeContext = AppRecoveryStore.peekPendingUpgradeContext();
    if (pendingUpgradeContext != null) {
      try {
        final upgradePolicy = await UpgradeGateService.queryPendingPolicy(
          pendingUpgradeContext,
        );
        return _RecoveryDecision.upgradeGate(
          pendingUpgradeContext,
          upgradePolicy,
        );
      } catch (_) {
        await AppRecoveryStore.clearPendingUpgradeContext();
        return _RecoveryDecision.fallback(
          context: pendingUpgradeContext.recoveryContext ??
              pendingUpgradeContext.fallbackContext,
          reason: 'upgrade_policy_unavailable',
          message: '暂时无法确认升级策略，已返回可用页面',
        );
      }
    }

    final pendingIntent = AppRecoveryStore.peekPendingIntent();
    final activeIntentCandidate = AppRecoveryStore.peekActiveIntentCandidate();
    final recentContext = AppRecoveryStore.getRecentContext();
    final candidate = pendingIntent ?? activeIntentCandidate ?? recentContext;
    final isPassiveRecentRestore = pendingIntent == null &&
        activeIntentCandidate == null &&
        recentContext != null;

    if (candidate == null) {
      return const _RecoveryDecision.fallback();
    }

    if (isPassiveRecentRestore && _shouldSkipPassiveRecentRestore(candidate)) {
      await AppRecoveryStore.clearRecentContext();
      return const _RecoveryDecision.fallback();
    }

    context.read<AppLifecycleProvider>().recordIntentReceived(candidate);

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

    final plan = AppIntentDispatcher.resolve(
      candidate,
      hasValidToken: AppRecoveryStore.hasValidToken(),
    );
    try {
      switch (plan.action) {
        case AppIntentDispatchAction.login:
          return _RecoveryDecision.login(plan.intent);
        case AppIntentDispatchAction.target:
          final scene = UpgradeGateService.recoverySceneFor(plan.intent);
          final upgradePolicy = await _queryUpgradePolicyForRecovery(
            plan.intent,
            scene: scene,
          );
          if (upgradePolicy.hasUpgradeGate) {
            return _RecoveryDecision.upgradeGate(
              PendingUpgradeContext.forRecentContext(
                scene: scene,
                recoveryContext: plan.intent,
                policy: upgradePolicy,
              ),
              upgradePolicy,
            );
          }
          return _RecoveryDecision.target(plan.intent, message: plan.message);
        case AppIntentDispatchAction.fallback:
          return _RecoveryDecision.fallback(
            context: plan.intent,
            reason: plan.failureReason,
            message: plan.message,
          );
        case AppIntentDispatchAction.upgradeGate:
          final scene = UpgradeGateService.recoverySceneFor(plan.intent);
          final upgradePolicy = await _queryUpgradePolicyForRecovery(
            plan.intent,
            scene: scene,
          );
          if (!upgradePolicy.hasUpgradeGate) {
            return _RecoveryDecision.target(plan.intent, message: plan.message);
          }
          return _RecoveryDecision.upgradeGate(
            PendingUpgradeContext.forRecentContext(
              scene: scene,
              recoveryContext: plan.intent,
              policy: upgradePolicy,
            ),
            upgradePolicy,
          );
      }
    } on _RecoveryUpgradePolicyUnavailable {
      return _RecoveryDecision.fallback(
        context: plan.intent,
        reason: 'upgrade_policy_unavailable',
        message: '暂时无法确认升级策略，已返回可用页面',
      );
    }
  }

  Future<AppVersionPolicy> _queryUpgradePolicyForRecovery(
    AppRecentContext context, {
    required String scene,
  }) async {
    try {
      if (context.isRecallIntent) {
        return await UpgradeGateService.queryRecallPolicy(
          context,
          scene: scene,
        );
      }
      return await UpgradeGateService.queryPolicy(
        scene: scene,
        recoveryContext: context,
      );
    } catch (_) {
      throw const _RecoveryUpgradePolicyUnavailable();
    }
  }

  Future<void> _resumePendingUpgrade(
    BuildContext pageContext,
    PendingUpgradeContext pendingUpgradeContext,
  ) async {
    switch (pendingUpgradeContext.targetType) {
      case UpgradeRecoveryTargetType.recentContext:
        if (pendingUpgradeContext.recoveryContext == null) {
          await AppRecoveryRouter.replaceWithFallback(
            pageContext,
            recentContext: pendingUpgradeContext.fallbackContext,
            message: pendingUpgradeContext.recoveryHint,
          );
          return;
        }
        await AppRecoveryRouter.replaceWithTarget(
          pageContext,
          pendingUpgradeContext.recoveryContext!,
          message: pendingUpgradeContext.recoveryHint,
        );
        return;
      case UpgradeRecoveryTargetType.orderConfirm:
        await Navigator.of(pageContext).pushReplacement(
          MaterialPageRoute(
            builder: (_) =>
                OrderSubmit(directItem: pendingUpgradeContext.directItem),
          ),
        );
        return;
      case UpgradeRecoveryTargetType.orderPay:
        await Navigator.of(pageContext).pushReplacement(
          MaterialPageRoute(
            builder: (_) => OrderPay(
              orderId: pendingUpgradeContext.orderId,
              orderSn: pendingUpgradeContext.orderSn,
              payType: pendingUpgradeContext.payType,
              amount: pendingUpgradeContext.payAmount ?? 0,
            ),
          ),
        );
        return;
    }
  }

  Future<void> _continuePendingUpgrade(
    BuildContext pageContext,
    PendingUpgradeContext pendingUpgradeContext,
  ) async {
    await AppRecoveryStore.clearPendingUpgradeContext();
    await AppRecoveryRouter.replaceWithFallback(
      pageContext,
      recentContext: pendingUpgradeContext.recoveryContext ??
          pendingUpgradeContext.fallbackContext,
      message: pendingUpgradeContext.recoveryHint,
    );
  }

  bool _shouldSkipPassiveRecentRestore(AppRecentContext context) {
    switch (context.targetType) {
      case AppRecentTargetType.couponList:
      case AppRecentTargetType.couponCenter:
        return true;
      case AppRecentTargetType.home:
      case AppRecentTargetType.cart:
      case AppRecentTargetType.orderList:
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
      case AppRecentTargetType.settings:
      case AppRecentTargetType.commentCompose:
      case AppRecentTargetType.afterSalesApply:
      case AppRecentTargetType.activity:
      case AppRecentTargetType.digitalCardAssetList:
      case AppRecentTargetType.digitalCardAssetDetail:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return false;
    }
  }

  @override
  Widget build(BuildContext context) {
    return widget.splash;
  }
}

enum _RecoveryAction { login, target, fallback, upgradeGate }

class _RecoveryDecision {
  final _RecoveryAction action;
  final AppRecentContext? context;
  final String? reason;
  final String? message;
  final PendingUpgradeContext? pendingUpgradeContext;
  final AppVersionPolicy? upgradePolicy;

  const _RecoveryDecision._({
    required this.action,
    this.context,
    this.reason,
    this.message,
    this.pendingUpgradeContext,
    this.upgradePolicy,
  });

  const _RecoveryDecision.login(AppRecentContext context)
      : this._(action: _RecoveryAction.login, context: context);

  const _RecoveryDecision.target(
    AppRecentContext context, {
    String? message,
  }) : this._(
          action: _RecoveryAction.target,
          context: context,
          message: message,
        );

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

  _RecoveryDecision.upgradeGate(
    PendingUpgradeContext pendingUpgradeContext,
    AppVersionPolicy upgradePolicy,
  ) : this._(
          action: _RecoveryAction.upgradeGate,
          context: pendingUpgradeContext.recoveryContext,
          pendingUpgradeContext: pendingUpgradeContext,
          upgradePolicy: upgradePolicy,
        );
}

class _RecoveryUpgradePolicyUnavailable implements Exception {
  const _RecoveryUpgradePolicyUnavailable();
}
