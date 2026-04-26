import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_test/flutter_test.dart';
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

  test('bootstrap recovery contexts are ignored', () async {
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

    expect(AppRecoveryStore.getRecentContext(), isNull);
    expect(AppRecoveryStore.peekPendingUpgradeContext(), isNull);
  });
}
