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
      return AppVersionPolicy(
        currentVersion: '1.0.0',
        minSupportedVersion: '1.2.0',
        recommendedVersion: '1.3.0',
        requiredVersion: '1.2.0',
        updateMode: UpgradeUpdateMode.force,
        effectiveAt: '',
        deadlineAt: '',
        affectedCapabilities: const <String>['app_bootstrap'],
        blocking: true,
        releaseNotesSummary: '升级后继续恢复启动上下文。',
        upgradeUrl: 'https://example.com/download/app.apk',
        storeTarget: 'browser_download',
        recoveryHint: '升级完成后将恢复原始目标。',
        traceId: 'bootstrap-test',
        platform: 'android',
        channel: 'direct',
        installerStore: '',
        scene: query.scene,
        targetType: query.targetType,
        targetId: query.targetId,
      );
    });
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
  });

  tearDown(() {
    AppVersionService.debugReset();
    UpgradeGateService.debugReset();
  });

  testWidgets(
      'intent recovery shell shows upgrade gate when bootstrap policy blocks',
      (tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider<AppLifecycleProvider>.value(
        value: AppLifecycleProvider(),
        child: MaterialApp(
          home: IntentRecoveryShell(
            splash: const SizedBox.shrink(),
            fallbackBuilder: (_) => const Scaffold(body: Text('fallback')),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('升级后继续当前任务'), findsOneWidget);
    expect(find.textContaining('1.2.0'), findsOneWidget);
  });
}
