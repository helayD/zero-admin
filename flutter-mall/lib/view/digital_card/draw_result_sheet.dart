import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_detail_page.dart';
import 'package:flutter_mall/view/digital_card/my_digital_card_page.dart';

class DrawResultSheet extends StatelessWidget {
  static const Color _ink = Color(0xFF151821);
  static const Color _inkSoft = Color(0xFF2C3240);
  static const Color _gold = Color(0xFFC78A24);
  static const Color _surface = Color(0xFFFFFCF7);
  static const Color _line = Color(0xFFE8DDCC);
  static const Color _muted = Color(0xFF746D63);

  final DrawMemberRecord record;
  final DrawEligibilitySummary eligibility;
  final bool requiresRealNameForRedemption;

  const DrawResultSheet({
    super.key,
    required this.record,
    required this.eligibility,
    this.requiresRealNameForRedemption = false,
  });

  DigitalCardStatusCopy get _assetStatusCopy =>
      digitalCardDrawResultCopy(record);

  String get _title {
    switch (record.resultStatus) {
      case 'won_pending_asset':
        if (requiresRealNameForRedemption) {
          return '待实名兑卡';
        }
        if (_assetStatusCopy.label.trim().isNotEmpty) {
          return _assetStatusCopy.label;
        }
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
      return digitalCardUserFacingText(record.failureReason);
    }
    if (record.resultStatus == 'won_pending_asset' &&
        requiresRealNameForRedemption) {
      final String cardName = record.templateName.trim();
      final String prefix = cardName.isEmpty ? '你已中奖' : '你抽中了 $cardName';
      final String suffix =
          record.assetNo.trim().isEmpty ? '' : '唯一编号 ${record.assetNo}。';
      return '$prefix，完成实名认证后可继续兑卡并发放到我的数字卡片。$suffix';
    }
    if (_assetStatusCopy.description.trim().isNotEmpty) {
      if (record.assetNo.trim().isNotEmpty) {
        return '你抽中了 ${record.templateName}，${_assetStatusCopy.description}唯一编号 ${record.assetNo}。';
      }
      return '你抽中了 ${record.templateName}，${_assetStatusCopy.description}';
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
        return _gold;
      case 'rejected_need_real_name':
        return const Color(0xFF2563EB);
      case 'rejected_quota_exhausted':
      case 'rejected_inventory_exhausted':
        return const Color(0xFFD97706);
      default:
        return _inkSoft;
    }
  }

  bool get _isWon => record.resultStatus == 'won_pending_asset';

  String get _resultBadge {
    if (_isWon) {
      return '中奖结果';
    }
    return '抽卡结果';
  }

  bool get _canOpenAssetCenter => record.resultStatus == 'won_pending_asset';

  String get _assetActionLabel {
    if (record.assetInstanceId > 0) {
      return '支付邮费/提现/转赠';
    }
    return '查看我的卡片';
  }

  Future<void> _openAssetCenter(BuildContext context) async {
    final NavigatorState rootNavigator =
        Navigator.of(context, rootNavigator: true);
    Navigator.of(context).pop();
    await Future<void>.delayed(Duration.zero);
    if (record.assetInstanceId > 0) {
      await rootNavigator.push(
        MaterialPageRoute(
          builder: (_) => DigitalCardAssetDetailPage(
            assetInstanceId: record.assetInstanceId,
            intentSource: 'draw_result_sheet',
          ),
        ),
      );
      return;
    }
    await rootNavigator.push(
      MaterialPageRoute(
        builder: (_) => const MyDigitalCardPage(
          intentSource: 'draw_result_sheet',
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: SingleChildScrollView(
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
                  color: const Color(0xFFD7D1C8),
                  borderRadius: BorderRadius.circular(999),
                ),
              ),
            ),
            const SizedBox(height: 18),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                color: _surface,
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: _line),
                boxShadow: const <BoxShadow>[
                  BoxShadow(
                    color: Color(0x12101828),
                    blurRadius: 22,
                    offset: Offset(0, 12),
                  ),
                ],
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Container(
                        width: 48,
                        height: 48,
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            colors: <Color>[
                              _accentColor,
                              _accentColor.withValues(alpha: 0.72),
                            ],
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                          ),
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Icon(
                          _isWon
                              ? Icons.workspace_premium_outlined
                              : Icons.info_outline,
                          color: Colors.white,
                          size: 25,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            _buildResultBadge(),
                            const SizedBox(height: 8),
                            Text(
                              _title,
                              style: const TextStyle(
                                color: _ink,
                                fontSize: 24,
                                height: 1.15,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 14),
                  Text(
                    _description,
                    style: const TextStyle(
                      color: _inkSoft,
                      fontSize: 14,
                      height: 1.55,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  if (_isWon ||
                      record.templateName.trim().isNotEmpty) ...<Widget>[
                    const SizedBox(height: 16),
                    _buildPrizePanel(),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 14),
            _buildMetaPanel(),
            const SizedBox(height: 18),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: () => Navigator.of(context).pop(),
                style: FilledButton.styleFrom(
                  backgroundColor: _ink,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 15),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
                child: const Text(
                  '我知道了',
                  style: TextStyle(fontWeight: FontWeight.w800),
                ),
              ),
            ),
            if (_canOpenAssetCenter) ...<Widget>[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  onPressed: () {
                    _openAssetCenter(context);
                  },
                  style: OutlinedButton.styleFrom(
                    foregroundColor: _ink,
                    side: const BorderSide(color: _ink),
                    padding: const EdgeInsets.symmetric(vertical: 15),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                  child: Text(
                    _assetActionLabel,
                    style: const TextStyle(fontWeight: FontWeight.w800),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildResultBadge() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
      decoration: BoxDecoration(
        color: _accentColor.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: _accentColor.withValues(alpha: 0.18)),
      ),
      child: Text(
        _resultBadge,
        style: TextStyle(
          color: _accentColor,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }

  Widget _buildPrizePanel() {
    return Container(
      width: double.infinity,
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF111827), Color(0xFF30243C)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Stack(
        children: <Widget>[
          Positioned(
            right: -44,
            top: -52,
            child: Container(
              width: 132,
              height: 132,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _gold.withValues(alpha: 0.18),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  record.templateName.trim().isEmpty
                      ? '本次抽卡结果'
                      : record.templateName.trim(),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 20,
                    height: 1.2,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 10),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: <Widget>[
                    if (record.rarity.trim().isNotEmpty)
                      _buildDarkChip('稀有度 ${record.rarity.trim()}'),
                    if (record.assetNo.trim().isNotEmpty)
                      _buildDarkChip('编号 ${record.assetNo.trim()}'),
                    if (_assetStatusCopy.actionHint.trim().isNotEmpty &&
                        record.assetNo.trim().isEmpty)
                      _buildDarkChip(_assetStatusCopy.actionHint.trim()),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDarkChip(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: Colors.white.withValues(alpha: 0.14)),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: Colors.white,
          fontSize: 12,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }

  Widget _buildMetaPanel() {
    final List<Widget> rows = <Widget>[];
    if (record.createTime.trim().isNotEmpty) {
      rows.add(_buildMetaRow('抽卡时间', record.createTime.trim()));
    }
    if (record.assetCreatedAt.trim().isNotEmpty &&
        !requiresRealNameForRedemption) {
      rows.add(_buildMetaRow('获取时间', record.assetCreatedAt.trim()));
    }
    if (record.assetNo.trim().isNotEmpty) {
      rows.add(_buildMetaRow('卡片编号', record.assetNo.trim()));
    }
    if (rows.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _line),
      ),
      child: Column(
        children: rows,
      ),
    );
  }

  Widget _buildMetaRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 5),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            width: 72,
            child: Text(
              label,
              style: const TextStyle(
                color: _muted,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              textAlign: TextAlign.right,
              style: const TextStyle(
                color: _ink,
                fontSize: 13,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
