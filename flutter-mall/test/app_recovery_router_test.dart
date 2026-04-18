import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/digital_card/draw_activity_page.dart';
import 'package:flutter_mall/view/digital_card/my_digital_card_page.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_detail_page.dart';
import 'package:flutter_mall/view/mine/coupon/available_coupon_list.dart';
import 'package:flutter_mall/view/mine/coupon/coupon_list.dart';
import 'package:flutter_mall/view/mine/order/order_list.dart';
import 'package:flutter_mall/view/mine/ping_jia/ping_jia.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

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

  test('comment compose recovery uses saved draft snapshot', () async {
    await AppRecoveryStore.saveCommentDraft(
      CommentDraftSnapshot(
        orderId: 1001,
        productId: 2001,
        productName: '测试商品',
        productPic: 'https://example.com/pic.png',
        productAttribute: '红色',
        memberNickName: '张三',
        starRating: 5,
        content: '',
        pics: <String>[],
        updatedAt: DateTime.now(),
      ),
    );

    final target = AppRecoveryRouter.buildTarget(
      AppRecentContext.create(
        targetType: AppRecentTargetType.commentCompose,
        targetId: 1001,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderDetail,
        fallbackTargetId: 1001,
      ),
    );

    expect(target, isA<PinJia>());
    expect((target as PinJia).orderId, 1001);
    expect(target.productId, 2001);
    expect(target.productName, '测试商品');
  });

  test('comment compose recovery falls back when draft snapshot is missing',
      () {
    final target = AppRecoveryRouter.buildTarget(
      AppRecentContext.create(
        targetType: AppRecentTargetType.commentCompose,
        targetId: 1001,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.home,
      ),
    );

    expect(target, isA<MainTab>());
  });

  test('activity recovery opens draw activity page when target id is present',
      () {
    final target = AppRecoveryRouter.buildTarget(
      AppRecentContext.createRecall(
        intentType: 'activity_recall',
        targetType: AppRecentTargetType.activity,
        targetId: 88,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
        source: 'member_message',
        requiresAuth: false,
        intentId: 'member_message:88',
      ),
    );

    expect(target, isA<DrawActivityPage>());
    expect((target as DrawActivityPage).activityId, 88);
  });

  test('digital card asset recovery opens list and detail pages', () {
    final listTarget = AppRecoveryRouter.buildTarget(
      AppRecentContext.create(
        targetType: AppRecentTargetType.digitalCardAssetList,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.home,
      ),
    );
    expect(listTarget, isA<MyDigitalCardPage>());

    final detailTarget = AppRecoveryRouter.buildTarget(
      AppRecentContext.create(
        targetType: AppRecentTargetType.digitalCardAssetDetail,
        targetId: 901,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.digitalCardAssetList,
      ),
    );
    expect(detailTarget, isA<DigitalCardAssetDetailPage>());
    expect((detailTarget as DigitalCardAssetDetailPage).assetInstanceId, 901);
  });
}
