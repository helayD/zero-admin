import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';

class DrawResultSheet extends StatelessWidget {
  final DrawMemberRecord record;
  final DrawEligibilitySummary eligibility;

  const DrawResultSheet({
    super.key,
    required this.record,
    required this.eligibility,
  });

  String get _title {
    switch (record.resultStatus) {
      case 'won_pending_asset':
        return '已中奖待到账';
      case 'rejected_need_real_name':
        return '待实名';
      case 'rejected_quota_exhausted':
        return '资格不足';
      case 'rejected_inventory_exhausted':
        return '库存不足';
      case 'rejected_activity_offline':
      case 'rejected_member_disabled':
        return '暂不可参与';
      default:
        return '未中奖';
    }
  }

  String get _description {
    if (record.failureReason.trim().isNotEmpty) {
      return record.failureReason.trim();
    }
    if (eligibility.eligibilityMessage.trim().isNotEmpty) {
      return eligibility.eligibilityMessage.trim();
    }
    if (record.templateName.trim().isNotEmpty) {
      return '你抽中了 ${record.templateName}，当前状态为已中奖待到账。';
    }
    return '本次抽卡结果已生成，你可以稍后回到活动页查看记录。';
  }

  Color get _accentColor {
    switch (record.resultStatus) {
      case 'won_pending_asset':
        return const Color(0xFFC97C00);
      case 'rejected_need_real_name':
        return const Color(0xFF3366CC);
      case 'rejected_quota_exhausted':
      case 'rejected_inventory_exhausted':
        return const Color(0xFFD35400);
      default:
        return const Color(0xFF3E4A59);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Center(
              child: Container(
                width: 44,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[300],
                  borderRadius: BorderRadius.circular(999),
                ),
              ),
            ),
            const SizedBox(height: 18),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: <Color>[
                    _accentColor.withValues(alpha: 0.92),
                    _accentColor.withValues(alpha: 0.72),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(24),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    _title,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 24,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Text(
                    _description,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      height: 1.5,
                    ),
                  ),
                  if (record.templateName.trim().isNotEmpty) ...<Widget>[
                    const SizedBox(height: 14),
                    Text(
                      '卡片: ${record.templateName}',
                      style: const TextStyle(color: Colors.white, fontSize: 13),
                    ),
                  ],
                  if (record.rarity.trim().isNotEmpty)
                    Text(
                      '稀有度: ${record.rarity}',
                      style: const TextStyle(color: Colors.white, fontSize: 13),
                    ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            Text(
              '请求ID: ${record.requestId}',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
            if (record.createTime.trim().isNotEmpty) ...<Widget>[
              const SizedBox(height: 4),
              Text(
                '时间: ${record.createTime}',
                style: TextStyle(fontSize: 12, color: Colors.grey[600]),
              ),
            ],
            const SizedBox(height: 18),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: () => Navigator.of(context).pop(),
                style: FilledButton.styleFrom(
                  backgroundColor: const Color(0xFF1F2937),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
                child: const Text('我知道了'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
