import 'package:flutter/material.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/view/category/product/product_detail.dart';
import 'package:flutter_mall/view/mine/order/order_detail.dart';
import 'package:flutter_mall/view/mine/order/order_list.dart';

class AppRecoveryRouter {
  static Widget buildTarget(AppRecentContext context) {
    switch (context.targetType) {
      case AppRecentTargetType.home:
        return MainTab(initialIndex: context.tabIndex ?? 0);
      case AppRecentTargetType.cart:
        return MainTab(initialIndex: context.tabIndex ?? 2);
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
        return OrderDetail(orderId: context.targetId!, intentSource: context.source);
    }
  }

  static Widget buildFallback([AppRecentContext? context]) {
    if (context == null) {
      return const MainTab();
    }
    switch (context.fallbackType) {
      case AppRecentTargetType.home:
        return MainTab(initialIndex: context.fallbackTabIndex ?? 0);
      case AppRecentTargetType.cart:
        return MainTab(initialIndex: context.fallbackTabIndex ?? 2);
      case AppRecentTargetType.orderList:
        return OrderList(initialTab: context.fallbackTabIndex ?? 0);
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
        return const MainTab();
    }
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
    final resolvedRecoveryIntent = recoveryIntent != null && recoveryIntent.isRecoverable
        ? recoveryIntent
        : null;
    final canRestoreRecoveryIntent = resolvedRecoveryIntent != null &&
        AppRecoveryStore.isContextAllowedForCurrentMember(resolvedRecoveryIntent);
    if (canRestoreRecoveryIntent) {
      await replaceWithTarget(context, resolvedRecoveryIntent);
      return;
    }
    if (resolvedRecoveryIntent != null) {
      await replaceWithFallback(
        context,
        recentContext: resolvedRecoveryIntent,
        message: '检测到账号已切换，已为你返回安全页面',
      );
      return;
    }
    if (Navigator.of(context).canPop()) {
      Navigator.of(context).pop(true);
      return;
    }
    if (redirectRoute != null && redirectRoute.isNotEmpty) {
      await replaceWithFallback(
        context,
        message: '未找到可恢复的目标，已返回首页',
      );
      return;
    }
    await replaceWithFallback(context);
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
