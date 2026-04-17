import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/utils/app_intent_dispatcher.dart';
import 'package:flutter_mall/utils/app_version_service.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

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

  test('requires login when auth intent has no valid token', () {
    final intent = AppRecentContext.createRecall(
      intentType: 'order_recall',
      targetType: AppRecentTargetType.orderDetail,
      targetId: 9001,
      fallbackType: AppRecentTargetType.orderList,
      fallbackTabIndex: 1,
      source: 'message_tap',
      requiresAuth: true,
      intentId: 'member_message:1',
    );

    final plan = AppIntentDispatcher.resolve(intent, hasValidToken: false);

    expect(plan.action, AppIntentDispatchAction.login);
    expect(plan.failureReason, 'login_required');
  });

  test('enters upgrade gate when min app version is not satisfied', () {
    final intent = AppRecentContext.createRecall(
      intentType: 'coupon_center_recall',
      targetType: AppRecentTargetType.couponCenter,
      fallbackType: AppRecentTargetType.couponList,
      fallbackTabIndex: 0,
      source: 'message_tap',
      requiresAuth: true,
      minAppVersion: '1.2.0',
      intentId: 'member_message:2',
      recoveryHint: '当前版本暂不支持直达领券中心，已为你切回优惠券列表',
    );

    final plan = AppIntentDispatcher.resolve(intent, hasValidToken: true);

    expect(plan.action, AppIntentDispatchAction.upgradeGate);
    expect(plan.failureReason, 'min_version_unmet');
    expect(plan.shouldMarkMessageRead, isTrue);
    expect(plan.upgradePolicy, isNull);
  });

  test('keeps valid product recall on target path', () {
    final intent = AppRecentContext.createRecall(
      intentType: 'product_recall',
      targetType: AppRecentTargetType.productDetail,
      targetId: 9527,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
      source: 'message_tap',
      requiresAuth: false,
      intentId: 'member_message:3',
    );

    final plan = AppIntentDispatcher.resolve(intent, hasValidToken: true);

    expect(plan.action, AppIntentDispatchAction.target);
    expect(plan.shouldMarkMessageRead, isTrue);
  });
}
