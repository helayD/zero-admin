import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/layout/intent_recovery_shell.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  Widget buildTestApp(AppLifecycleProvider provider) {
    return ChangeNotifierProvider<AppLifecycleProvider>.value(
      value: provider,
      child: MaterialApp(
        home: IntentRecoveryShell(
          splash: const SizedBox.shrink(),
          targetBuilder: (context) => Scaffold(
            body: Text(
              'target-${context.targetTypeValue}-${context.targetId ?? context.tabIndex ?? 0}',
            ),
          ),
          fallbackBuilder: (_) => const Scaffold(body: Text('fallback-page')),
          loginBuilder: (_) => const Scaffold(body: Text('login-page')),
        ),
      ),
    );
  }

  testWidgets('falls back when there is no recent context', (tester) async {
    final provider = AppLifecycleProvider();

    await tester.pumpWidget(buildTestApp(provider));
    await tester.pumpAndSettle();

    expect(find.text('fallback-page'), findsOneWidget);
    expect(provider.fallbackUsed, isTrue);
    expect(provider.intentReceivedCount, 0);
    expect(provider.latestIntentTelemetry, isNull);
  });

  testWidgets('restores target when a valid recent context exists',
      (tester) async {
    final provider = AppLifecycleProvider();
    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.home,
        tabIndex: 0,
        source: 'resume',
        requiresAuth: false,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
      ),
    );

    await tester.pumpWidget(buildTestApp(provider));
    await tester.pumpAndSettle();

    expect(find.text('target-home-0'), findsOneWidget);
    expect(provider.restoreSucceeded, isTrue);
    expect(provider.intentReceivedCount, 1);
    expect(provider.intentRestoredCount, 1);
    expect(provider.latestIntentTelemetry?.eventName, 'intentRestored');
    expect(provider.latestIntentTelemetry?.targetType, 'home');
    expect(provider.latestIntentTelemetry?.source, 'resume');
  });

  testWidgets('prefers active intent candidate over recent context',
      (tester) async {
    final provider = AppLifecycleProvider();
    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.home,
        tabIndex: 0,
        source: 'resume',
        requiresAuth: false,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
      ),
    );
    await AppRecoveryStore.saveActiveIntentCandidate(
      AppRecentContext.create(
        targetType: AppRecentTargetType.orderDetail,
        targetId: 1001,
        source: 'resume',
        requiresAuth: false,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
      ),
    );

    await tester.pumpWidget(buildTestApp(provider));
    await tester.pumpAndSettle();

    expect(find.text('target-order_detail-1001'), findsOneWidget);
    expect(provider.restoreSucceeded, isTrue);
    expect(provider.intentReceivedCount, 1);
    expect(provider.intentRestoredCount, 1);
    expect(provider.latestIntentTelemetry?.eventName, 'intentRestored');
    expect(provider.latestIntentTelemetry?.targetType, 'order_detail');
  });
}
