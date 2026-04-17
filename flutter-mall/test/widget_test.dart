import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/layout/app_bootstrap.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('app bootstrap smoke test', (WidgetTester tester) async {
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

    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AppLifecycleProvider(),
        child: MaterialApp(
          home: AppBootstrap(
            fallbackBuilder: (_) => const Scaffold(
              body: Text('fallback-page'),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('fallback-page'), findsOneWidget);

    AppVersionService.debugReset();
    UpgradeGateService.debugReset();
  });
}
