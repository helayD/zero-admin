import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/utils/version_compare.dart';

void main() {
  test('compareVersion covers semantic edge cases', () {
    expect(compareVersion('1.0.0', '1.2.0'), -1);
    expect(compareVersion('1.2', '1.2.0'), 0);
    expect(compareVersion('1.2.0+5', '1.2.0'), 0);
    expect(compareVersion('', '1.0.0'), -1);
  });

  test(
      'policyFromRecallIntent avoids dead-end blocking gate without upgrade url',
      () {
    final intent = AppRecentContext.createRecall(
      intentType: 'coupon_center_recall',
      targetType: AppRecentTargetType.couponCenter,
      fallbackType: AppRecentTargetType.couponList,
      fallbackTabIndex: 0,
      source: 'message_tap',
      requiresAuth: true,
      minAppVersion: '1.2.0',
      intentId: 'member_message:9',
      recoveryHint: '升级后继续领券',
    );
    final policy = AppVersionPolicy.fromRecallIntent(
      intent,
      versionInfo: const AppVersionInfo(
        version: '1.0.0',
        buildNumber: '1',
        platform: 'android',
        installerStore: '',
        channel: 'direct',
        isFallback: false,
      ),
    );

    expect(policy.blocking, isFalse);
    expect(policy.updateMode, UpgradeUpdateMode.recommended);
    expect(policy.requiredDisplayVersion, '1.2.0');
    expect(policy.scene, 'message_recall');
    expect(policy.storeTarget, 'settings_check');
  });

  test('applyRecallRequirement reuses shared upgrade entry for blocking gate',
      () {
    final intent = AppRecentContext.createRecall(
      intentType: 'coupon_center_recall',
      targetType: AppRecentTargetType.couponCenter,
      fallbackType: AppRecentTargetType.couponList,
      fallbackTabIndex: 0,
      source: 'message_tap',
      requiresAuth: true,
      minAppVersion: '1.2.0',
      intentId: 'member_message:10',
      recoveryHint: '升级后继续领券',
    );
    final policy = const AppVersionPolicy(
      currentVersion: '1.0.0',
      minSupportedVersion: '1.0.0',
      recommendedVersion: '1.0.0',
      requiredVersion: '',
      updateMode: UpgradeUpdateMode.none,
      effectiveAt: '',
      deadlineAt: '',
      affectedCapabilities: <String>['message_recall'],
      blocking: false,
      releaseNotesSummary: '',
      upgradeUrl: 'https://example.com/app.apk',
      storeTarget: 'browser_download',
      recoveryHint: '',
      traceId: 'trace-1',
      platform: 'android',
      channel: 'direct',
      installerStore: '',
      scene: 'message_recall',
      targetType: '',
      targetId: null,
    ).applyRecallRequirement(intent);

    expect(policy.blocking, isTrue);
    expect(policy.updateMode, UpgradeUpdateMode.force);
    expect(policy.requiredDisplayVersion, '1.2.0');
    expect(policy.upgradeUrl, 'https://example.com/app.apk');
    expect(policy.targetType, 'coupon_center');
    expect(policy.recoveryHint, '升级后继续领券');
  });
}
