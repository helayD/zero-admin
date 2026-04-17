import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/layout/intent_recovery_shell.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
    AppVersionService.debugSetCurrentInfo(
      const AppVersionInfo(
        version: '1.0.0',
        buildNumber: '1',
        platform: 'android',
        installerStore: '',
        channel: 'direct',
        isFallback: false,
      ),
    );
    UpgradeGateService.debugSetPolicyFetcher((query) async {
      return AppVersionPolicy.none(
        versionInfo: query.versionInfo,
        scene: query.scene,
        targetType: query.targetType,
        targetId: query.targetId,
      );
    });
  });

  tearDown(() {
    AppVersionService.debugReset();
    UpgradeGateService.debugReset();
  });

  testWidgets(
      'opens login builder and stores pending intent when auth is required',
      (tester) async {
    final provider = AppLifecycleProvider();
    final context = AppRecentContext.create(
      targetType: AppRecentTargetType.orderDetail,
      targetId: 1001,
      source: 'resume',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.orderList,
      fallbackTabIndex: 1,
    );
    await AppRecoveryStore.saveRecentContext(context);

    await tester.pumpWidget(
      ChangeNotifierProvider<AppLifecycleProvider>.value(
        value: provider,
        child: MaterialApp(
          home: IntentRecoveryShell(
            splash: const SizedBox.shrink(),
            targetBuilder: (_) => const Scaffold(body: Text('target-page')),
            fallbackBuilder: (_) => const Scaffold(body: Text('fallback-page')),
            loginBuilder: (_) => const Scaffold(body: Text('login-page')),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('login-page'), findsOneWidget);
    expect(provider.restoreFailedReason, 'login_required');
    expect(provider.intentReceivedCount, 1);
    expect(provider.intentLoginRequiredCount, 1);
    expect(provider.latestIntentTelemetry?.eventName, 'intentLoginRequired');
    expect(provider.latestIntentTelemetry?.targetType, 'order_detail');
    expect(provider.latestIntentTelemetry?.failureReason, 'login_required');

    final pendingIntent = AppRecoveryStore.peekPendingIntent();
    expect(pendingIntent, isNotNull);
    expect(pendingIntent!.targetType, AppRecentTargetType.orderDetail);
    expect(pendingIntent.targetId, 1001);
  });
}
