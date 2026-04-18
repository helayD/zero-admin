import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/digital_card/compliance_rule_banner.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_detail_page.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_tile.dart';
import 'package:flutter_mall/view/digital_card/mint_status_timeline.dart';
import 'package:flutter_mall/view/digital_card/my_digital_card_page.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await SharedPreferencesUtil.init();
  });

  testWidgets('DigitalCardAssetTile 展示编号与状态摘要', (WidgetTester tester) async {
    const DigitalCardAssetItem item = DigitalCardAssetItem(
      assetInstanceId: 1001,
      assetNo: 'CARD202604180001',
      templateId: 21,
      templateName: 'SSR 卡',
      cardFaceImage: '',
      activityId: 2001,
      activityName: '春季抽卡',
      rarity: 'SSR',
      obtainedAt: '2026-04-18 10:00:00',
      mintStatus: 'mint_processing',
      mintStatusText: '链上处理中',
      chainStatus: 'processing',
      chainStatusText: '处理中',
      displayStatus: 'display_hidden',
      displayStatusText: '受限展示',
      complianceStatus: 'compliance_review',
      complianceStatusText: '人工复核中',
      tokenStatusText: '处理中',
      complianceRuleSummary: '合规复核中',
    );

    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: DigitalCardAssetTile(item: item),
        ),
      ),
    );

    expect(find.text('SSR 卡'), findsOneWidget);
    expect(find.textContaining('编号 CARD202604180001'), findsOneWidget);
    expect(find.text('链上处理中'), findsOneWidget);
    expect(find.text('合规复核中'), findsOneWidget);
  });

  testWidgets('ComplianceRuleBanner 与 MintStatusTimeline 展示核心说明',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Column(
            children: const <Widget>[
              ComplianceRuleBanner(
                title: '合规说明',
                summary: '该资产当前处于人工复核中，请稍后查看最新结果。',
                statusText: '人工复核中',
              ),
              MintStatusTimeline(
                timeline: <DigitalCardAssetTimelineItem>[
                  DigitalCardAssetTimelineItem(
                    operationType: 'mint_dispatching',
                    operationText: '已派发链路任务',
                    statusText: '链上处理中',
                    reasonText: '等待链上回执',
                    createTime: '2026-04-18 10:00:00',
                  ),
                  DigitalCardAssetTimelineItem(
                    operationType: 'asset_review',
                    operationText: '进入人工复核',
                    statusText: '人工复核中',
                    reasonText: '命中合规规则',
                    createTime: '2026-04-18 10:05:00',
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );

    expect(find.text('合规说明'), findsOneWidget);
    expect(find.text('人工复核中'), findsWidgets);
    expect(find.text('已派发链路任务'), findsOneWidget);
    expect(find.text('进入人工复核'), findsOneWidget);
    expect(find.text('等待链上回执'), findsOneWidget);
    final Offset dispatchOffset = tester.getTopLeft(find.text('已派发链路任务'));
    final Offset reviewOffset = tester.getTopLeft(find.text('进入人工复核'));
    expect(dispatchOffset.dy, lessThan(reviewOffset.dy));
  });

  testWidgets('MyDigitalCardPage 覆盖 loading 与 empty 状态',
      (WidgetTester tester) async {
    final Completer<QueryMyDigitalCardAssetListResponse> completer =
        Completer<QueryMyDigitalCardAssetListResponse>();

    await tester.pumpWidget(
      MaterialApp(
        home: MyDigitalCardPage(
          fetchList: () => completer.future,
        ),
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    completer.complete(
      const QueryMyDigitalCardAssetListResponse(
        code: '200',
        message: 'success',
        data: QueryMyDigitalCardAssetListData(
          total: 0,
          list: <DigitalCardAssetItem>[],
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('还没有数字卡片'), findsOneWidget);
    expect(find.text('下拉刷新'), findsOneWidget);
  });

  testWidgets('MyDigitalCardPage 覆盖 error 状态', (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: MyDigitalCardPage(
          fetchList: () async => throw Exception('boom'),
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('加载失败'), findsOneWidget);
    expect(find.text('重新加载'), findsOneWidget);
  });

  testWidgets('DigitalCardAssetDetailPage 覆盖 loading 与 error 状态',
      (WidgetTester tester) async {
    final Completer<QueryMyDigitalCardAssetDetailResponse> completer =
        Completer<QueryMyDigitalCardAssetDetailResponse>();

    await tester.pumpWidget(
      MaterialApp(
        home: DigitalCardAssetDetailPage(
          assetInstanceId: 1001,
          fetchDetail: (_) => completer.future,
        ),
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    completer.completeError(Exception('detail failed'));
    await tester.pumpAndSettle();

    expect(find.text('加载数字卡片详情失败，请稍后重试'), findsOneWidget);
    expect(find.text('重新加载'), findsOneWidget);
  });
}
