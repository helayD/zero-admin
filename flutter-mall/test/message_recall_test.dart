import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/message_model.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('legacy order message falls back to relatedOrderId when linkId is empty',
      () {
    final message = MessageData.fromJson({
      'id': 11,
      'messageType': 1,
      'title': '订单提醒',
      'content': '点击查看订单详情',
      'linkType': 'order',
      'linkId': '',
      'relatedOrderId': 9001,
      'status': 0,
      'createTime': '2026-04-05T12:00:00',
    });

    expect(message.intent, isNotNull);
    expect(message.intent!.targetType, AppRecentTargetType.orderDetail);
    expect(message.intent!.targetId, 9001);
    expect(message.intent!.fallbackType, AppRecentTargetType.orderList);
    expect(message.intent!.fallbackTabIndex, 1);
    expect(message.intent!.blocked, isFalse);
  });

  test('legacy activity message becomes direct activity intent', () {
    final message = MessageData.fromJson({
      'id': 12,
      'messageType': 4,
      'title': '活动提醒',
      'content': '春季大促活动火热进行中',
      'linkType': 'activity',
      'linkId': '88',
      'status': 0,
      'createTime': '2026-04-05T12:05:00',
    });

    expect(message.intent, isNotNull);
    expect(message.intent!.targetType, AppRecentTargetType.activity);
    expect(message.intent!.blocked, isFalse);
    expect(message.intent!.targetId, 88);
    expect(message.intent!.fallbackType, AppRecentTargetType.home);
  });

  test('structured recall payload is parsed directly from api response', () {
    final message = MessageData.fromJson({
      'id': 13,
      'messageType': 5,
      'title': '优惠券到账',
      'content': '登录领取新券',
      'status': 0,
      'createTime': '2026-04-05T12:10:00',
      'intent': {
        'intentType': 'coupon_center_recall',
        'targetType': 'coupon_center',
        'fallbackType': 'coupon_list',
        'fallbackTab': 0,
        'requiresAuth': true,
        'intentId': 'member_message:13',
        'issuedAt': '2026-04-05T12:10:00',
        'source': 'member_message',
      },
    });

    expect(message.intent, isNotNull);
    expect(message.intent!.targetType, AppRecentTargetType.couponCenter);
    expect(message.intent!.intentId, 'member_message:13');
    expect(message.intent!.isRecallIntent, isTrue);
  });
}
