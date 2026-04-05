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

  testWidgets('opens login builder and stores pending intent when auth is required', (tester) async {
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

    final pendingIntent = AppRecoveryStore.peekPendingIntent();
    expect(pendingIntent, isNotNull);
    expect(pendingIntent!.targetType, AppRecentTargetType.orderDetail);
    expect(pendingIntent.targetId, 1001);
  });
}
