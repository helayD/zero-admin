import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';

void main() {
  test('DrawMemberRecord.fromJson 解析资产字段', () {
    final DrawMemberRecord record = DrawMemberRecord.fromJson(<String, dynamic>{
      'id': 1,
      'activityId': 10,
      'requestId': 'req-model-1',
      'resultType': 'won',
      'resultStatus': 'won_pending_asset',
      'resultStatusText': '已中奖待到账',
      'templateName': 'SSR 卡',
      'rarity': 'SSR',
      'assetInstanceId': 9001,
      'assetNo': 'CARD202604170001',
      'assetStatus': 'asset_created',
      'assetStatusText': '资产已创建，链上处理中',
      'assetCreatedAt': '2026-04-17 10:00:00',
    });

    expect(record.assetInstanceId, 9001);
    expect(record.assetNo, 'CARD202604170001');
    expect(record.assetStatus, 'asset_created');
    expect(record.assetStatusText, '资产已创建，链上处理中');
    expect(record.assetCreatedAt, '2026-04-17 10:00:00');
  });
}
