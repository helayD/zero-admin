import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/view/mine/coupon/available_coupon_list.dart';
import 'package:flutter_mall/view/mine/coupon/coupon_list.dart';
import 'package:flutter_mall/view/mine/order/order_list.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('fallback widgets keep original intent source', () {
    const source = 'message_recall';

    final homeFallback = AppRecoveryRouter.buildFallback(
      AppRecentContext.createRecall(
        intentType: 'activity_recall',
        targetType: AppRecentTargetType.activity,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
        source: source,
        requiresAuth: false,
        intentId: 'member_message:31',
      ),
    );
    expect(homeFallback, isA<MainTab>());
    expect((homeFallback as MainTab).intentSource, source);

    final orderListFallback = AppRecoveryRouter.buildFallback(
      AppRecentContext.createRecall(
        intentType: 'message_recall',
        targetType: AppRecentTargetType.home,
        fallbackType: AppRecentTargetType.orderList,
        fallbackTabIndex: 1,
        source: source,
        requiresAuth: true,
        intentId: 'member_message:32',
      ),
    );
    expect(orderListFallback, isA<OrderList>());
    expect((orderListFallback as OrderList).intentSource, source);

    final couponListFallback = AppRecoveryRouter.buildFallback(
      AppRecentContext.createRecall(
        intentType: 'message_recall',
        targetType: AppRecentTargetType.home,
        fallbackType: AppRecentTargetType.couponList,
        fallbackTabIndex: 0,
        source: source,
        requiresAuth: true,
        intentId: 'member_message:33',
      ),
    );
    expect(couponListFallback, isA<CouponList>());
    expect((couponListFallback as CouponList).intentSource, source);

    final couponCenterFallback = AppRecoveryRouter.buildFallback(
      AppRecentContext.createRecall(
        intentType: 'message_recall',
        targetType: AppRecentTargetType.home,
        fallbackType: AppRecentTargetType.couponCenter,
        source: source,
        requiresAuth: true,
        intentId: 'member_message:34',
      ),
    );
    expect(couponCenterFallback, isA<AvailableCouponList>());
    expect((couponCenterFallback as AvailableCouponList).intentSource, source);
  });
}
