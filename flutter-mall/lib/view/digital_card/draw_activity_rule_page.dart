import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';

class DrawActivityRulePage extends StatelessWidget {
  static const Color _pageBackground = Color(0xFFF7F2EA);
  static const Color _surface = Color(0xFFFFFCF7);
  static const Color _ink = Color(0xFF151821);
  static const Color _inkSoft = Color(0xFF2C3240);
  static const Color _gold = Color(0xFFC78A24);
  static const Color _goldLight = Color(0xFFFFE6A6);
  static const Color _line = Color(0xFFE8DDCC);
  static const Color _muted = Color(0xFF746D63);
  static const Color _green = Color(0xFF0F766E);

  final DrawActivityLandingData landing;

  const DrawActivityRulePage({
    super.key,
    required this.landing,
  });

  String _clean(String text) {
    return digitalCardUserFacingText(text).trim();
  }

  List<_RuleSection> get _sections {
    return <_RuleSection>[
      _RuleSection(
        title: '参与说明',
        content: _clean(landing.participantConditionSummary),
        icon: Icons.how_to_reg_outlined,
        color: _gold,
      ),
      _RuleSection(
        title: '活动规则',
        content: _clean(landing.ruleSummary),
        icon: Icons.auto_awesome_outlined,
        color: _gold,
      ),
      _RuleSection(
        title: '合规说明',
        content: _clean(landing.complianceRuleSummary),
        icon: Icons.verified_user_outlined,
        color: _green,
      ),
      _RuleSection(
        title: '消耗规则',
        content: _clean(landing.consumeRuleSummary),
        icon: Icons.confirmation_number_outlined,
        color: _gold,
      ),
      _RuleSection(
        title: '卡池概率',
        content: _clean(landing.probabilityRule).isNotEmpty
            ? _clean(landing.probabilityRule)
            : '本活动每次旋转必然中奖，各格位概率之和为 100%，不存在未中奖格位。',
        icon: Icons.percent_outlined,
        color: _gold,
      ),
      _RuleSection(
        title: '流通限制',
        content: _clean(landing.circulationLimitSummary),
        icon: Icons.lock_outline,
        color: _inkSoft,
      ),
      _RuleSection(
        title: '活动时间',
        content: _buildTimeCopy(),
        icon: Icons.schedule_outlined,
        color: _inkSoft,
      ),
    ].where((section) => section.content.isNotEmpty).toList();
  }

  String _buildTimeCopy() {
    final String start = landing.startTime.trim();
    final String end = landing.endTime.trim();
    if (start.isEmpty && end.isEmpty) {
      return '';
    }
    if (start.isEmpty) {
      return '截止 $end';
    }
    if (end.isEmpty) {
      return '开始 $start';
    }
    return '$start 至 $end';
  }

  @override
  Widget build(BuildContext context) {
    final List<_RuleSection> sections = _sections;
    return Scaffold(
      backgroundColor: _pageBackground,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 0,
        foregroundColor: _ink,
        title: const Text('活动说明'),
        titleTextStyle: const TextStyle(
          color: _ink,
          fontSize: 20,
          fontWeight: FontWeight.w800,
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
        children: <Widget>[
          _buildHeader(),
          const SizedBox(height: 14),
          if (sections.isEmpty)
            _buildEmptyState()
          else
            ...sections.map(_buildSectionCard),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF10131B), Color(0xFF30243C)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(24),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            decoration: BoxDecoration(
              color: _goldLight,
              borderRadius: BorderRadius.circular(999),
            ),
            child: const Text(
              '抽卡说明',
              style: TextStyle(
                color: _ink,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(height: 14),
          Row(
            children: <Widget>[
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
                decoration: BoxDecoration(
                  color: _gold.withValues(alpha: 0.18),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(color: _gold.withValues(alpha: 0.32)),
                ),
                child: const Row(
                  mainAxisSize: MainAxisSize.min,
                  children: <Widget>[
                    Icon(Icons.auto_awesome, color: _goldLight, size: 13),
                    SizedBox(width: 4),
                    Text(
                      '每次旋转必然中奖 · 中奖率 100%',
                      style: TextStyle(
                        color: _goldLight,
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            landing.name.trim().isEmpty ? '数字卡片活动' : landing.name.trim(),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              height: 1.18,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionCard(_RuleSection section) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _line),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: section.color.withValues(alpha: 0.12),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(section.icon, color: section.color, size: 21),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  section.title,
                  style: TextStyle(
                    color: section.color,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  section.content,
                  style: const TextStyle(
                    color: _inkSoft,
                    fontSize: 14,
                    height: 1.55,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _line),
      ),
      child: const Text(
        '暂无更多活动说明',
        style: TextStyle(
          color: _muted,
          fontSize: 14,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

class _RuleSection {
  final String title;
  final String content;
  final IconData icon;
  final Color color;

  const _RuleSection({
    required this.title,
    required this.content,
    required this.icon,
    required this.color,
  });
}
