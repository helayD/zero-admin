import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';

void main() {
  test('DigitalCardAssetItem.fromJson 解析资产状态字段', () {
    final DigitalCardAssetItem item = DigitalCardAssetItem.fromJson(
      <String, dynamic>{
        'assetInstanceId': 1001,
        'assetNo': 'CARD202604180001',
        'templateId': 21,
        'templateName': 'SSR 卡',
        'cardFaceImage': 'https://img.example.com/card.png',
        'activityId': 2001,
        'activityName': '春季抽卡',
        'rarity': 'SSR',
        'obtainedAt': '2026-04-18 10:00:00',
        'mintStatus': 'mint_processing',
        'mintStatusText': '链上处理中',
        'chainStatus': 'processing',
        'chainStatusText': '处理中',
        'displayStatus': 'display_hidden',
        'displayStatusText': '受限展示',
        'complianceStatus': 'compliance_review',
        'complianceStatusText': '人工复核中',
        'tokenStatusText': '处理中',
        'complianceRuleSummary': '合规复核中',
      },
    );

    expect(item.assetInstanceId, 1001);
    expect(item.templateName, 'SSR 卡');
    expect(item.displayStatus, 'display_hidden');
    expect(item.complianceStatusText, '人工复核中');
    expect(item.hasRestriction, isTrue);
  });
}
