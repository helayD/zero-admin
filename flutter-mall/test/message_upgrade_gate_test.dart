import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/utils/app_intent_dispatcher.dart';
import 'package:flutter_mall/utils/app_version_service.dart';

void main() {
  setUp(() {
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
  });

  tearDown(AppVersionService.debugReset);

  test('message recall enters upgrade gate when min app version is unmet', () {
    final intent = AppRecentContext.createRecall(
      intentType: 'coupon_center_recall',
      targetType: AppRecentTargetType.couponCenter,
      fallbackType: AppRecentTargetType.couponList,
      fallbackTabIndex: 0,
      source: 'message_tap',
      requiresAuth: true,
      minAppVersion: '1.2.0',
      intentId: 'member_message:2',
      recoveryHint: '升级后继续领券中心',
    );

    final plan = AppIntentDispatcher.resolve(intent, hasValidToken: true);

    expect(plan.action, AppIntentDispatchAction.upgradeGate);
    expect(plan.failureReason, 'min_version_unmet');
    expect(plan.upgradePolicy, isNull);
  });
}
