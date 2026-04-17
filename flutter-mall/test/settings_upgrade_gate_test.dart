import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/view/mine/setting/settings.dart';
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
        minSupportedVersion: '1.0.0',
        recommendedVersion: '1.1.0',
        requiredVersion: '1.1.0',
        updateMode: UpgradeUpdateMode.recommended,
        effectiveAt: '2026-04-30T00:00:00+08:00',
        deadlineAt: '2026-05-31T23:59:59+08:00',
        affectedCapabilities: const <String>['settings_check'],
        blocking: false,
        releaseNotesSummary: '升级后可获得更稳定的恢复链路。',
        upgradeUrl: 'https://example.com/download/app.apk',
        storeTarget: 'browser_download',
        recoveryHint: '升级后可继续刚才的任务。',
        traceId: 'settings-test',
        platform: 'android',
        channel: 'direct',
        installerStore: '',
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

  testWidgets('settings check update uses shared upgrade gate flow',
      (tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider<AppLifecycleProvider>.value(
        value: AppLifecycleProvider(),
        child: const MaterialApp(
          home: Settings(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('当前版本 1.0.0'), findsOneWidget);

    await tester.tap(find.text('检查更新'));
    await tester.pumpAndSettle();

    expect(find.text('发现更合适的新版本'), findsOneWidget);
    expect(find.textContaining('1.1.0'), findsOneWidget);
  });
}
