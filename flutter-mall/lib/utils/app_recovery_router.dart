import 'package:flutter/material.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/app_intent_dispatcher.dart';
import 'package:flutter_mall/view/mine/ping_jia/ping_jia.dart';
import 'package:flutter_mall/view/category/product/product_detail.dart';
import 'package:flutter_mall/view/mine/coupon/available_coupon_list.dart';
import 'package:flutter_mall/view/mine/coupon/coupon_list.dart';
import 'package:flutter_mall/view/mine/order/apply_after_sales.dart';
import 'package:flutter_mall/view/mine/order/order_detail.dart';
import 'package:flutter_mall/view/mine/order/order_list.dart';
import 'package:flutter_mall/view/mine/setting/settings.dart';

class AppRecoveryRouter {
  static Widget buildTarget(AppRecentContext context) {
    switch (context.targetType) {
      case AppRecentTargetType.home:
        return MainTab(
          initialIndex: context.tabIndex ?? 0,
          intentSource: context.source,
        );
      case AppRecentTargetType.cart:
        return MainTab(
          initialIndex: context.tabIndex ?? 2,
          intentSource: context.source,
        );
      case AppRecentTargetType.orderList:
        return OrderList(
          initialTab: context.tabIndex ?? 0,
          intentSource: context.source,
        );
      case AppRecentTargetType.productDetail:
        return ProductDetail(
          productId: context.targetId!,
          intentSource: context.source,
        );
      case AppRecentTargetType.orderDetail:
        return OrderDetail(
          orderId: context.targetId!,
          intentSource: context.source,
        );
      case AppRecentTargetType.settings:
        return const Settings();
      case AppRecentTargetType.commentCompose:
        final targetId = context.targetId;
        if (targetId == null || targetId <= 0) {
          return buildFallback(context);
        }
        final draft = AppRecoveryStore.getCommentDraft(targetId);
        if (draft == null || draft.orderId <= 0 || draft.productId <= 0) {
          return buildFallback(context);
        }
        return PinJia(
          orderId: draft.orderId,
          productId: draft.productId,
          productName: draft.productName,
          productPic: draft.productPic,
          productAttribute: draft.productAttribute,
          memberNickName: draft.memberNickName,
        );
      case AppRecentTargetType.couponList:
        return CouponList(
          initialTab: context.tabIndex ?? 0,
          intentSource: context.source,
        );
      case AppRecentTargetType.couponCenter:
        return AvailableCouponList(intentSource: context.source);
      case AppRecentTargetType.afterSalesApply:
        return ApplyAfterSales(
          orderId: context.targetId!,
          intentSource: context.source,
        );
      case AppRecentTargetType.activity:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return buildFallback(context);
    }
  }

  static Widget buildFallback([AppRecentContext? context]) {
    if (context == null) {
      return const MainTab();
    }
    switch (context.fallbackType) {
      case AppRecentTargetType.home:
        return MainTab(
          initialIndex: context.fallbackTabIndex ?? 0,
          intentSource: context.source,
        );
      case AppRecentTargetType.cart:
        return MainTab(
          initialIndex: context.fallbackTabIndex ?? 2,
          intentSource: context.source,
        );
      case AppRecentTargetType.orderList:
        return OrderList(
          initialTab: context.fallbackTabIndex ?? 0,
          intentSource: context.source,
        );
      case AppRecentTargetType.orderDetail:
        if (context.fallbackTargetId != null && context.fallbackTargetId! > 0) {
          return OrderDetail(
            orderId: context.fallbackTargetId!,
            intentSource: context.source,
          );
        }
        return MainTab(intentSource: context.source);
      case AppRecentTargetType.couponList:
        return CouponList(
          initialTab: context.fallbackTabIndex ?? 0,
          intentSource: context.source,
        );
      case AppRecentTargetType.couponCenter:
        return AvailableCouponList(intentSource: context.source);
      case AppRecentTargetType.settings:
      case AppRecentTargetType.commentCompose:
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.afterSalesApply:
      case AppRecentTargetType.activity:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return MainTab(intentSource: context.source);
    }
  }

  static Future<void> pushTarget(
    BuildContext context,
    AppRecentContext recentContext, {
    String? message,
  }) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (routeContext) {
          _showMessage(routeContext, message);
          return buildTarget(recentContext);
        },
      ),
    );
  }

  static Future<void> pushFallback(
    BuildContext context, {
    AppRecentContext? recentContext,
    String? message,
  }) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (routeContext) {
          _showMessage(routeContext, message);
          return buildFallback(recentContext);
        },
      ),
    );
  }

  static Future<void> replaceWithTarget(
    BuildContext context,
    AppRecentContext recentContext, {
    String? message,
  }) async {
    await Navigator.of(context).pushReplacement(
      MaterialPageRoute(
        builder: (routeContext) {
          _showMessage(routeContext, message);
          return buildTarget(recentContext);
        },
      ),
    );
  }

  static Future<void> replaceWithFallback(
    BuildContext context, {
    AppRecentContext? recentContext,
    String? message,
  }) async {
    await Navigator.of(context).pushReplacement(
      MaterialPageRoute(
        builder: (routeContext) {
          _showMessage(routeContext, message);
          return buildFallback(recentContext);
        },
      ),
    );
  }

  static Future<void> restoreAfterLogin(
    BuildContext context, {
    required AppRecentContext? recoveryIntent,
    String? redirectRoute,
  }) async {
    final navigator = Navigator.of(context);
    final resolvedRecoveryIntent =
        recoveryIntent != null && recoveryIntent.isRecoverable
            ? recoveryIntent
            : null;
    final canRestoreRecoveryIntent = resolvedRecoveryIntent != null &&
        AppRecoveryStore.isContextAllowedForCurrentMember(
          resolvedRecoveryIntent,
        );
    if (canRestoreRecoveryIntent) {
      final allowedRecoveryIntent = resolvedRecoveryIntent;
      final plan = AppIntentDispatcher.resolve(
        allowedRecoveryIntent,
        hasValidToken: AppRecoveryStore.hasValidToken(),
      );
      switch (plan.action) {
        case AppIntentDispatchAction.target:
          await navigator.pushReplacement(
            MaterialPageRoute(
              builder: (routeContext) {
                _showMessage(routeContext, plan.message);
                return buildTarget(plan.intent);
              },
            ),
          );
          return;
        case AppIntentDispatchAction.fallback:
          await navigator.pushReplacement(
            MaterialPageRoute(
              builder: (routeContext) {
                _showMessage(routeContext, plan.message);
                return buildFallback(plan.intent);
              },
            ),
          );
          return;
        case AppIntentDispatchAction.login:
          break;
      }
    }
    if (!navigator.mounted) {
      return;
    }
    if (resolvedRecoveryIntent != null) {
      await navigator.pushReplacement(
        MaterialPageRoute(
          builder: (routeContext) {
            _showMessage(routeContext, '检测到账号已切换，已为你返回安全页面');
            return buildFallback(resolvedRecoveryIntent);
          },
        ),
      );
      return;
    }
    if (navigator.canPop()) {
      navigator.pop(true);
      return;
    }
    if (redirectRoute != null && redirectRoute.isNotEmpty) {
      await navigator.pushReplacement(
        MaterialPageRoute(
          builder: (routeContext) {
            _showMessage(routeContext, '未找到可恢复的目标，已返回首页');
            return buildFallback();
          },
        ),
      );
      return;
    }
    await navigator.pushReplacement(
      MaterialPageRoute(
        builder: (_) => buildFallback(),
      ),
    );
  }

  static void _showMessage(BuildContext context, String? message) {
    if (message == null || message.trim().isEmpty) {
      return;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final messenger = ScaffoldMessenger.maybeOf(context);
      if (messenger == null) {
        return;
      }
      messenger.showSnackBar(
        SnackBar(content: Text(message)),
      );
    });
  }
}
