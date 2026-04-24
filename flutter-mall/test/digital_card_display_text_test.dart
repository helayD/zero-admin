import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';

void main() {
  test('digitalCardUserFacingText 隐藏底层链路文案', () {
    expect(digitalCardUserFacingText('链上处理中'), '到账中');
    expect(digitalCardUserFacingText('等待链上回执'), '等待处理结果');
    expect(digitalCardUserFacingText('AntChain token 已同步'), '服务 编号 已同步');
  });

  test('digitalCardAssetPrimaryCopy 优先展示用户可理解状态', () {
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
      tokenStatusText: '底层凭证处理中',
      complianceRuleSummary: '合规复核中',
      chainType: 'antchain',
    );

    final DigitalCardStatusCopy copy = digitalCardAssetPrimaryCopy(item);

    expect(copy.label, '合规复核中');
    expect(copy.description, '卡片正在复核，结果更新后会自动同步。');
    expect(copy.actionHint, '无需重复操作，稍后刷新查看结果。');
  });

  test('digitalCardDrawResultCopy 将资产创建状态表达为到账中', () {
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

    final DigitalCardStatusCopy copy = digitalCardDrawResultCopy(record);

    expect(copy.label, '到账中');
    expect(copy.description, '卡片正在到账，稍后会同步到我的数字卡片。');
  });
}
