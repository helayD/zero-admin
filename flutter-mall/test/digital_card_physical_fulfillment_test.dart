import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/digital_card/physical_fulfillment_model.dart';
import 'package:flutter_mall/view/digital_card/digital_card_physical_fulfillment_page.dart';
import 'package:flutter_mall/view/digital_card/physical_fulfillment_timeline.dart';

void main() {
  test('PhysicalFulfillmentDetailData 不解析底层链路字段', () {
    final PhysicalFulfillmentDetailData detail =
        PhysicalFulfillmentDetailData.fromJson(<String, dynamic>{
      'fulfillmentId': 1,
      'fulfillmentNo': 'PF2026042600000001',
      'assetInstanceId': 970001,
      'assetNo': 'CARD-001',
      'templateName': 'SSR 卡',
      'activityName': '春季抽卡',
      'obtainedAt': '2026-04-26 10:00:00',
      'mintStatusText': '已到账',
      'fulfillmentStatus': 'shipped',
      'fulfillmentStatusText': '已发货',
      'productionStatusText': '制作完成',
      'shippingStatusText': '已发货',
      'receiverNameMasked': '张*',
      'receiverPhoneMasked': '138****8001',
      'addressSummary': '广东省深圳市南山区科技园',
      'carrierName': '顺丰速运',
      'trackingNo': 'SF123456789',
      'complianceTipSummary': '实体卡履约仅展示制作、配送和签收进度。',
      'chainType': 'antchain',
      'chainStatus': 'success',
      'tokenId': 'token-1',
      'timeline': <Map<String, dynamic>>[
        <String, dynamic>{
          'action': 'physical_shipped',
          'actionText': '实体卡发货',
          'statusText': '已发货',
          'reason': '发货',
          'createTime': '2026-04-26 11:00:00',
        },
      ],
    });

    expect(detail.fulfillmentNo, 'PF2026042600000001');
    expect(detail.canConfirmReceipt, isTrue);
    final String encoded = jsonEncode(<String, dynamic>{
      'fulfillmentNo': detail.fulfillmentNo,
      'status': detail.fulfillmentStatusText,
      'timeline': detail.timeline.map((item) => item.actionText).toList(),
    });
    expect(encoded.contains('chainType'), isFalse);
    expect(encoded.contains('tokenId'), isFalse);
  });

  testWidgets('实体卡履约页不展示底层链路文案', (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: DigitalCardPhysicalFulfillmentPage(
          assetInstanceId: 970001,
          fetchDetail: (_) async =>
              const QueryMyPhysicalFulfillmentDetailResponse(
            code: 'SUCCESS',
            message: 'success',
            data: PhysicalFulfillmentDetailData(
              fulfillmentId: 1,
              fulfillmentNo: 'PF2026042600000001',
              assetInstanceId: 970001,
              assetNo: 'CARD-001',
              templateName: 'SSR 卡',
              activityName: '春季抽卡',
              obtainedAt: '2026-04-26 10:00:00',
              mintStatusText: '已到账',
              fulfillmentStatus: 'shipped',
              fulfillmentStatusText: '已发货',
              productionStatusText: '制作完成',
              shippingStatusText: '已发货',
              receiverNameMasked: '张*',
              receiverPhoneMasked: '138****8001',
              addressSummary: '广东省深圳市南山区科技园',
              carrierName: '顺丰速运',
              trackingNo: 'SF123456789',
              complianceTipSummary: '实体卡履约仅展示制作、配送和签收进度。',
              blockedReason: '',
              blockedReasonText: '',
              timeline: <PhysicalFulfillmentTimelineItem>[
                PhysicalFulfillmentTimelineItem(
                  action: 'physical_shipped',
                  actionText: '实体卡发货',
                  statusText: '已发货',
                  reason: '发货',
                  createTime: '2026-04-26 11:00:00',
                ),
              ],
            ),
          ),
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('实体卡进度'), findsWidgets);
    expect(find.text('SSR 卡'), findsOneWidget);
    expect(find.text('确认已收到'), findsOneWidget);
    expect(find.textContaining('蚂蚁链'), findsNothing);
    expect(find.textContaining('AntChain'), findsNothing);
    expect(find.textContaining('FISCO'), findsNothing);
    expect(find.textContaining('区块链'), findsNothing);
    expect(find.textContaining('链上'), findsNothing);
    expect(find.textContaining('token'), findsNothing);
  });

  testWidgets('PhysicalFulfillmentTimeline 展示空态与时间线',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: PhysicalFulfillmentTimeline(
            timeline: <PhysicalFulfillmentTimelineItem>[],
          ),
        ),
      ),
    );
    expect(find.text('暂无实体卡进度，稍后下拉刷新可获取最新状态。'), findsOneWidget);
  });
}
