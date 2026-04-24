import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';
import 'package:flutter_mall/view/digital_card/draw_result_sheet.dart';

void main() {
  testWidgets('DrawResultSheet 展示资产状态与唯一编号', (WidgetTester tester) async {
    const DrawMemberRecord record = DrawMemberRecord(
      id: 1,
      activityId: 10,
      requestId: 'req-widget-1',
      resultType: 'won',
      resultStatus: 'won_pending_asset',
      resultStatusText: '已中奖待到账',
      failureCode: '',
      failureReason: '',
      poolId: 1,
      templateId: 2,
      templateName: 'SSR 卡',
      rarity: 'SSR',
      consumeAmount: 1,
      lotteryTimesBefore: 3,
      lotteryTimesAfter: 2,
      assetInstanceId: 9001,
      assetNo: 'CARD202604170001',
      assetStatus: 'asset_created',
      assetStatusText: '资产已创建，链上处理中',
      assetCreatedAt: '2026-04-17 10:00:00',
      createTime: '2026-04-17 10:00:01',
    );
    const DrawEligibilitySummary eligibility = DrawEligibilitySummary(
      eligibilityStatus: 'eligible',
      eligibilityCode: 'eligible',
      eligibilityMessage: '',
      nextAction: 'none',
      remainingLotteryTimes: 2,
    );

    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: DrawResultSheet(record: record, eligibility: eligibility),
        ),
      ),
    );

    expect(find.text('到账中'), findsOneWidget);
    expect(find.textContaining('卡片正在到账，稍后会同步到我的数字卡片。唯一编号 CARD202604170001'),
        findsOneWidget);
    expect(find.text('资产已创建，处理中'), findsNothing);
    expect(find.text('资产已创建，链上处理中'), findsNothing);
    expect(find.textContaining('唯一编号: CARD202604170001'), findsOneWidget);
    expect(find.textContaining('建账时间: 2026-04-17 10:00:00'), findsOneWidget);
  });

  testWidgets('DrawResultSheet 在无编号时仍展示服务端资产状态', (WidgetTester tester) async {
    const DrawMemberRecord record = DrawMemberRecord(
      id: 1,
      activityId: 10,
      requestId: 'req-widget-2',
      resultType: 'won',
      resultStatus: 'won_pending_asset',
      resultStatusText: '已中奖待到账',
      failureCode: '',
      failureReason: '',
      poolId: 1,
      templateId: 2,
      templateName: 'SSR 卡',
      rarity: 'SSR',
      consumeAmount: 1,
      lotteryTimesBefore: 3,
      lotteryTimesAfter: 2,
      assetInstanceId: 9002,
      assetNo: '',
      assetStatus: 'asset_created',
      assetStatusText: '资产已创建，链上处理中',
      assetCreatedAt: '',
      createTime: '2026-04-17 10:00:01',
    );
    const DrawEligibilitySummary eligibility = DrawEligibilitySummary(
      eligibilityStatus: 'eligible',
      eligibilityCode: 'eligible',
      eligibilityMessage: '',
      nextAction: 'none',
      remainingLotteryTimes: 2,
    );

    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: DrawResultSheet(record: record, eligibility: eligibility),
        ),
      ),
    );

    expect(find.text('到账中'), findsOneWidget);
    expect(find.textContaining('卡片正在到账，稍后会同步到我的数字卡片。'), findsOneWidget);
    expect(find.text('可以先关闭弹窗，稍后在卡包查看。'), findsOneWidget);
    expect(find.text('资产已创建，处理中'), findsNothing);
    expect(find.text('资产已创建，链上处理中'), findsNothing);
    expect(find.textContaining('唯一编号:'), findsNothing);
  });
}
