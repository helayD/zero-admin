import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';
import 'package:flutter_mall/view/digital_card/draw_activity_rule_page.dart';
import 'package:flutter_mall/view/digital_card/draw_result_sheet.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/view/mine/points/points_log_page.dart';
import 'package:flutter_mall/widgets/lucky_wheel_widget.dart';
import 'package:uuid/uuid.dart';

class DrawActivityPage extends StatefulWidget {
  final int activityId;
  final String? activityTitle;
  final String? intentSource;

  const DrawActivityPage({
    super.key,
    required this.activityId,
    this.activityTitle,
    this.intentSource,
  });

  @override
  State<DrawActivityPage> createState() => _DrawActivityPageState();
}

class _DrawActivityPageState extends State<DrawActivityPage> {
  static const Uuid _uuid = Uuid();
  static const Color _pageBackground = Color(0xFFF7F2EA);
  static const Color _surface = Color(0xFFFFFCF7);
  static const Color _ink = Color(0xFF151821);
  static const Color _inkSoft = Color(0xFF2C3240);
  static const Color _gold = Color(0xFFC78A24);
  static const Color _goldLight = Color(0xFFFFE6A6);
  static const Color _line = Color(0xFFE8DDCC);
  static const Color _muted = Color(0xFF746D63);
  static const Color _green = Color(0xFF0F766E);

  DrawActivityLandingData? _landing;
  bool _isLoading = true;
  bool _isParticipating = false;
  bool _useAnonymousLandingOnly = false;
  String? _errorMessage;
  final GlobalKey<LuckyWheelWidgetState> _wheelKey =
      GlobalKey<LuckyWheelWidgetState>();

  AppRecentContext _buildRecoveryContext([String? source]) {
    return AppRecentContext.create(
      targetType: AppRecentTargetType.activity,
      targetId: widget.activityId,
      source: source ?? widget.intentSource ?? 'digital_card',
      requiresAuth: false,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );
  }

  @override
  void initState() {
    super.initState();
    AppRecoveryStore.saveActiveIntentCandidate(_buildRecoveryContext());
    _loadLanding();
  }

  @override
  void dispose() {
    AppRecoveryStore.clearActiveIntentCandidateIfMatches(
      AppRecentTargetType.activity,
      targetId: widget.activityId,
    );
    super.dispose();
  }

  Future<void> _loadLanding() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final bool hasToken =
          !_useAnonymousLandingOnly && AppRecoveryStore.hasValidToken();
      Response response;
      try {
        response = await _requestLanding(withMemberContext: hasToken);
      } on DioException catch (err) {
        if (hasToken && _shouldUseAnonymousLandingFallback(err)) {
          if (err.response?.statusCode == 401) {
            await AppRecoveryStore.clearAuthToken();
          }
          _useAnonymousLandingOnly = true;
          response = await _requestLanding(withMemberContext: false);
        } else {
          rethrow;
        }
      }
      final model = drawActivityLandingResponseFromJson(
        jsonEncode(response.data),
      );
      if (!mounted) {
        return;
      }

      if (_shouldTreatAsError(model.code, model.data.activityId)) {
        setState(() {
          _errorMessage = _safeLandingMessage(
            model.message,
            fallback: '活动暂时不可用',
          );
          _isLoading = false;
          _landing = null;
        });
        return;
      }

      setState(() {
        _landing = model.data;
        _isLoading = false;
      });
      await AppRecoveryStore.saveRecentContext(
          _buildRecoveryContext('digital_card_view'));
    } catch (err) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = _landingErrorMessage(err);
        _landing = null;
        _isLoading = false;
      });
    }
  }

  Future<Response> _requestLanding({required bool withMemberContext}) {
    final String path = withMemberContext
        ? queryMyDrawActivityLandingUrl
        : queryDrawActivityLandingUrl;
    return HttpUtil.get(
      path,
      queryParameters: <String, dynamic>{'activityId': widget.activityId},
      redirectOnUnauthorized: !withMemberContext,
    );
  }

  String _landingErrorMessage(Object err) {
    if (err is DioException) {
      final data = err.response?.data;
      if (data is Map) {
        final message = _safeLandingMessage(data['message']?.toString() ?? '');
        if (message.isNotEmpty) {
          return message;
        }
      }
      if (data is String) {
        final message = _safeLandingMessage(data);
        if (message.isNotEmpty) {
          return message;
        }
      }
    }
    return '加载大转盘活动失败，请稍后重试';
  }

  String _safeLandingMessage(
    String message, {
    String fallback = '',
  }) {
    final text = digitalCardUserFacingText(message);
    return text.isEmpty ? fallback : text;
  }

  bool _shouldUseAnonymousLandingFallback(DioException err) {
    final statusCode = err.response?.statusCode ?? 0;
    return statusCode == 400 ||
        statusCode == 401 ||
        statusCode == 500 ||
        statusCode == 502 ||
        statusCode == 503;
  }

  bool _shouldTreatAsError(String code, int activityId) {
    return activityId <= 0 || code == 'DRAW_ACTIVITY_OFFLINE';
  }

  Future<void> _handlePrimaryAction() async {
    final landing = _landing;
    if (landing == null || _isParticipating) {
      return;
    }

    final eligibility = landing.eligibility;
    switch (eligibility.eligibilityCode) {
      case 'need_login':
        await Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) =>
                Login(recoveryIntent: _buildRecoveryContext('draw_login')),
          ),
        );
        if (!mounted) {
          return;
        }
        _useAnonymousLandingOnly = false;
        await _loadLanding();
        return;
      case 'need_real_name':
        _showSnackBar(
          landing.identity.credentialRef.trim().isEmpty
              ? '请先完成实名认证后再兑卡'
              : '请先根据实名提示完成认证后再兑卡',
        );
        return;
      case 'eligible':
        break;
      default:
        _showSnackBar(
          eligibility.eligibilityMessage.trim().isEmpty
              ? '当前暂不可参与，请稍后再试'
              : eligibility.eligibilityMessage,
        );
        return;
    }

    if (!_isWheelReady(landing)) {
      _showSnackBar('转盘需配置 5 个格位且概率总和等于 1，请联系管理员');
      return;
    }

    setState(() {
      _isParticipating = true;
    });

    final LuckyWheelWidgetState? wheelState = _wheelKey.currentState;
    wheelState?.startSpin();
    final DateTime spinStarted = DateTime.now();

    try {
      final String requestId = _uuid.v4();
      final Response response = await HttpUtil.post(
        participateDrawUrl,
        data: <String, dynamic>{
          'activityId': widget.activityId,
          'requestId': requestId,
          'channel': 'app',
          'entrySource': widget.intentSource ?? 'digital_card_page',
        },
      );
      final parsed = participateDrawResponseFromJson(jsonEncode(response.data));
      if (!mounted) {
        return;
      }

      final Duration elapsed = DateTime.now().difference(spinStarted);
      final Duration remaining = const Duration(seconds: 2) - elapsed;
      if (remaining > Duration.zero) {
        await Future.delayed(remaining);
      }
      if (!mounted) {
        return;
      }

      if (wheelState != null) {
        final int winSlot = _findWinningSlot(parsed.data.record, landing);
        await wheelState.stopAt(winSlot);
      }
      if (!mounted) {
        return;
      }

      final bool requiresRealNameForRedemption =
          parsed.data.record.resultStatus == 'won_pending_asset' &&
              landing.identity.realNameStatus.trim() != 'verified';

      await showModalBottomSheet<void>(
        context: context,
        isScrollControlled: true,
        backgroundColor: Colors.white,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
        ),
        builder: (_) => DrawResultSheet(
          record: parsed.data.record,
          eligibility: parsed.data.eligibility,
          requiresRealNameForRedemption: requiresRealNameForRedemption,
        ),
      );
      await _loadLanding();
    } catch (_) {
      if (!mounted) {
        return;
      }
      _showSnackBar('参与旋转失败，请稍后重试');
    } finally {
      if (mounted) {
        setState(() {
          _isParticipating = false;
        });
      }
    }
  }

  void _showSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  String _statusLabel(DrawEligibilitySummary eligibility) {
    switch (eligibility.eligibilityCode) {
      case 'need_login':
        return '待登录';
      case 'need_real_name':
        return '待实名';
      case 'eligible':
        return '可参与';
      case 'quota_exhausted':
        return '资格不足';
      case 'inventory_exhausted':
        return '库存不足';
      case 'activity_offline':
        return '已结束';
      default:
        return '资格校验中';
    }
  }

  String _primaryActionText(DrawEligibilitySummary eligibility) {
    switch (eligibility.eligibilityCode) {
      case 'need_login':
        return '登录后参与';
      case 'need_real_name':
        return '前往实名';
      case 'eligible':
        if (_isParticipating) return '旋转中...';
        final landing = _landing;
        if (landing != null && landing.consumeType == 'points' && landing.consumeAmount > 0) {
          return '旋转（消耗 ${landing.consumeAmount} 积分）';
        }
        return '立即旋转';
      default:
        return '暂不可参与';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _pageBackground,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 0,
        foregroundColor: _ink,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: _gold),
      );
    }
    if (_errorMessage != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(28),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Container(
                width: 68,
                height: 68,
                decoration: BoxDecoration(
                  color: _goldLight.withValues(alpha: 0.55),
                  borderRadius: BorderRadius.circular(24),
                ),
                child: const Icon(
                  Icons.error_outline,
                  size: 34,
                  color: _gold,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: _ink,
                  fontSize: 15,
                  height: 1.5,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 18),
              FilledButton(
                onPressed: _loadLanding,
                style: FilledButton.styleFrom(
                  backgroundColor: _ink,
                  foregroundColor: Colors.white,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 28, vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14),
                  ),
                ),
                child: const Text('重新加载'),
              ),
            ],
          ),
        ),
      );
    }

    final landing = _landing;
    if (landing == null) {
      return const SizedBox.shrink();
    }

    return RefreshIndicator(
      color: _gold,
      onRefresh: _loadLanding,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
        children: <Widget>[
          _buildHeroCard(landing),
          const SizedBox(height: 12),
          _buildRuleEntry(landing),
          const SizedBox(height: 18),
          _buildSectionTitle('格位详情'),
          const SizedBox(height: 10),
          _buildSlotDetailList(landing),
          const SizedBox(height: 18),
          _buildSectionTitle('卡池说明'),
          const SizedBox(height: 10),
          if (landing.pools.isEmpty)
            _buildEmptyCard('当前暂无卡池说明')
          else
            ...landing.pools.map(_buildPoolCard),
          const SizedBox(height: 18),
          _buildSectionTitle('最近中奖'),
          const SizedBox(height: 10),
          if (landing.recentWins.isEmpty)
            _buildEmptyCard('暂无公开中奖动态')
          else
            ...landing.recentWins.map(_buildRecentWinTile),
          const SizedBox(height: 18),
          _buildSectionTitle('我的参与记录'),
          const SizedBox(height: 10),
          if (landing.myRecords.isEmpty)
            _buildEmptyCard(
              landing.eligibility.eligibilityCode == 'need_login'
                  ? '登录后可查看你的旋转记录'
                  : '还没有参与记录，准备好就试试手气吧',
            )
          else
            ...landing.myRecords.map(_buildRecordTile),
        ],
      ),
    );
  }

  Widget _buildHeroCard(DrawActivityLandingData landing) {
    final eligibility = landing.eligibility;
    final String title = landing.name.trim().isEmpty
        ? widget.activityTitle?.trim() ?? '大转盘活动'
        : landing.name.trim();
    final List<DrawCardPreview> slots = landing.pools.isNotEmpty
        ? landing.pools.first.cards
        : const <DrawCardPreview>[];
    final bool wheelReady = _isWheelReady(landing);

    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF10131B), Color(0xFF2A2130)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(30),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x24101828),
            blurRadius: 26,
            offset: Offset(0, 16),
          ),
        ],
      ),
      child: Stack(
        children: <Widget>[
          Positioned(
            right: -48,
            top: -72,
            child: Container(
              width: 176,
              height: 176,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _gold.withValues(alpha: 0.2),
              ),
            ),
          ),
          Positioned(
            left: -36,
            bottom: -52,
            child: Container(
              width: 142,
              height: 142,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _green.withValues(alpha: 0.12),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                Row(
                  children: <Widget>[
                    _buildStatusPill(_statusLabel(eligibility)),
                    const SizedBox(width: 8),
                    _buildGhostPill('大转盘'),
                  ],
                ),
                const SizedBox(height: 12),
                Text(
                  title,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 22,
                    height: 1.15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 20),
                if (wheelReady)
                  Center(
                    child: LuckyWheelWidget(
                      key: _wheelKey,
                      slots: slots,
                      size: 280,
                    ),
                  )
                else
                  _buildWheelPlaceholder(),
                const SizedBox(height: 14),
                _buildCurrencyBadge(landing),
                const SizedBox(height: 14),
                Semantics(
                  button: true,
                  label: wheelReady
                      ? _primaryActionText(eligibility)
                      : '转盘格位配置不完整',
                  child: SizedBox(
                    width: double.infinity,
                    child: FilledButton(
                      onPressed:
                          (_isParticipating || !wheelReady)
                              ? null
                              : _handlePrimaryAction,
                      style: FilledButton.styleFrom(
                        backgroundColor: _goldLight,
                        disabledBackgroundColor:
                            _goldLight.withValues(alpha: 0.58),
                        foregroundColor: _ink,
                        disabledForegroundColor: _ink.withValues(alpha: 0.56),
                        padding: const EdgeInsets.symmetric(vertical: 15),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                      ),
                      child: Text(
                        wheelReady
                            ? _primaryActionText(eligibility)
                            : '转盘格位配置不完整',
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCurrencyBadge(DrawActivityLandingData landing) {
    final int cost = landing.consumeAmount;
    final bool isPoints = landing.consumeType == 'points';
    final String eligCode = landing.eligibility.eligibilityCode;
    if (cost <= 0 || eligCode == 'need_login') {
      return const SizedBox.shrink();
    }
    final int balance = landing.eligibility.remainingLotteryTimes;
    final String unit = isPoints ? '积分' : '次';
    final bool canAfford = balance >= cost;
    final Color textColor = canAfford ? _goldLight : Colors.redAccent;
    final IconData icon =
        isPoints ? Icons.stars_rounded : Icons.confirmation_number_outlined;

    final Widget badge = Row(
      mainAxisAlignment: MainAxisAlignment.center,
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Icon(icon, color: textColor, size: 15),
        const SizedBox(width: 5),
        Text(
          '剩余 $balance $unit · 本次消耗 $cost $unit',
          style: TextStyle(
            color: textColor,
            fontSize: 13,
            fontWeight: FontWeight.w600,
          ),
        ),
        if (isPoints) ...<Widget>[
          const SizedBox(width: 4),
          Icon(
            Icons.chevron_right,
            color: textColor.withValues(alpha: 0.7),
            size: 15,
          ),
        ],
      ],
    );

    if (!isPoints) return badge;

    return GestureDetector(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute<void>(
            builder: (_) => PointsLogPage(currentPoints: balance),
          ),
        );
      },
      child: badge,
    );
  }

  Widget _buildStatusPill(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: _goldLight,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: _ink,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }

  Widget _buildGhostPill(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: Colors.white.withValues(alpha: 0.14)),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: Colors.white.withValues(alpha: 0.78),
          fontSize: 12,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }

  Widget _buildRuleEntry(DrawActivityLandingData landing) {
    return Semantics(
      button: true,
      label: '查看活动说明',
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(20),
          onTap: () {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => DrawActivityRulePage(landing: landing),
              ),
            );
          },
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
            decoration: BoxDecoration(
              color: _surface,
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: _line),
            ),
            child: Row(
              children: <Widget>[
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: _goldLight.withValues(alpha: 0.55),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.description_outlined,
                    color: _gold,
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                const Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        '活动说明',
                        style: TextStyle(
                          color: _ink,
                          fontSize: 15,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      SizedBox(height: 2),
                      Text(
                        '规则、合规与卡池概率',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: _muted,
                          fontSize: 12,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ),
                ),
                const Icon(
                  Icons.chevron_right,
                  color: _muted,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Row(
      children: <Widget>[
        Container(
          width: 4,
          height: 18,
          decoration: BoxDecoration(
            color: _gold,
            borderRadius: BorderRadius.circular(999),
          ),
        ),
        const SizedBox(width: 8),
        Text(
          title,
          style: const TextStyle(
            color: _ink,
            fontSize: 18,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }

  Widget _buildPoolCard(DrawPoolPreview pool) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: _line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 34,
                height: 34,
                decoration: BoxDecoration(
                  color: _goldLight.withValues(alpha: 0.7),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(
                  Icons.view_carousel_outlined,
                  color: _gold,
                  size: 19,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  pool.poolName,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
          if (pool.probabilityRule.trim().isNotEmpty) ...<Widget>[
            const SizedBox(height: 10),
            Text(
              pool.probabilityRule,
              style: const TextStyle(
                fontSize: 13,
                color: _muted,
                height: 1.45,
              ),
            ),
          ],
          if (pool.cards.isNotEmpty) ...<Widget>[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: pool.cards.map((card) {
                return Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  decoration: BoxDecoration(
                    color: _goldLight.withValues(alpha: 0.62),
                    borderRadius: BorderRadius.circular(999),
                    border: Border.all(color: _gold.withValues(alpha: 0.2)),
                  ),
                  child: Text(
                    '${card.templateName} ${card.rarity}'.trim(),
                    style: const TextStyle(
                      color: _ink,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                );
              }).toList(),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildRecentWinTile(DrawRecentWin item) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _line),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: _goldLight.withValues(alpha: 0.72),
              borderRadius: BorderRadius.circular(14),
            ),
            child: const Icon(
              Icons.workspace_premium_outlined,
              color: _gold,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  '${item.memberNameMasked} 旋转赢得 ${item.templateName}'.trim(),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ink,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  '${digitalCardUserFacingText(item.resultStatusText)} · ${item.createTime}'
                      .trim(),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 12, color: _muted),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRecordTile(DrawMemberRecord item) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  digitalCardUserFacingText(item.resultStatusText).isEmpty
                      ? '旋转记录'
                      : digitalCardUserFacingText(item.resultStatusText),
                  style: const TextStyle(
                    color: _ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              Text(
                item.createTime,
                style: const TextStyle(fontSize: 12, color: _muted),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            item.templateName.trim().isEmpty
                ? (digitalCardUserFacingText(item.failureReason).trim().isEmpty
                    ? '旋转结果处理中'
                    : digitalCardUserFacingText(item.failureReason))
                : '卡片: ${item.templateName} · ${item.rarity}',
            style: const TextStyle(
              fontSize: 13,
              color: _inkSoft,
              height: 1.45,
            ),
          ),
          if (digitalCardUserFacingText(item.assetStatusText)
              .trim()
              .isNotEmpty) ...<Widget>[
            const SizedBox(height: 6),
            Text(
              item.assetNo.trim().isNotEmpty
                  ? '${digitalCardUserFacingText(item.assetStatusText)} · 编号 ${item.assetNo}'
                  : digitalCardUserFacingText(item.assetStatusText),
              style: const TextStyle(fontSize: 12, color: _muted),
            ),
          ],
          const SizedBox(height: 6),
          Text(
            '消耗 ${item.currencyBefore} -> ${item.currencyAfter}',
            style: const TextStyle(fontSize: 12, color: _muted),
          ),
        ],
      ),
    );
  }

  bool _isWheelReady(DrawActivityLandingData landing) {
    if (landing.pools.isEmpty) return false;
    final List<DrawCardPreview> cards = landing.pools.first.cards;
    if (cards.length != 5) return false;
    final Set<int> indices =
        cards.map((DrawCardPreview c) => c.slotIndex).toSet();
    if (indices.length != 5 ||
        !indices.containsAll(<int>[1, 2, 3, 4, 5])) {
      return false;
    }
    final double sum =
        cards.fold(0.0, (double s, DrawCardPreview c) => s + c.probability);
    return (sum - 1.0).abs() <= 0.001;
  }

  Widget _buildWheelPlaceholder() {
    return SizedBox.square(
      dimension: 280,
      child: DecoratedBox(
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: Colors.white.withValues(alpha: 0.04),
          border: Border.all(
            color: Colors.white.withValues(alpha: 0.12),
            width: 2,
          ),
        ),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(
                Icons.lock_outline_rounded,
                color: _goldLight.withValues(alpha: 0.45),
                size: 40,
              ),
              const SizedBox(height: 10),
              Text(
                '转盘配置中\n需要 5 个格位且概率总和 = 1',
                textAlign: TextAlign.center,
                style: TextStyle(
                  color: Colors.white.withValues(alpha: 0.45),
                  fontSize: 13,
                  height: 1.5,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  int _findWinningSlot(
    DrawMemberRecord record,
    DrawActivityLandingData landing,
  ) {
    if (record.slotIndex > 0) return record.slotIndex;
    if (landing.pools.isNotEmpty) {
      for (final DrawCardPreview card in landing.pools.first.cards) {
        if (card.templateId == record.templateId && card.slotIndex > 0) {
          return card.slotIndex;
        }
      }
    }
    return 1;
  }

  Widget _buildSlotDetailList(DrawActivityLandingData landing) {
    if (landing.pools.isEmpty || landing.pools.first.cards.isEmpty) {
      return _buildEmptyCard('当前暂无格位信息');
    }
    final List<DrawCardPreview> slots =
        List<DrawCardPreview>.from(landing.pools.first.cards);
    slots.sort(
      (DrawCardPreview a, DrawCardPreview b) =>
          a.slotIndex.compareTo(b.slotIndex),
    );
    return Column(
      children: slots
          .map((DrawCardPreview s) => _buildSlotTile(s))
          .toList(),
    );
  }

  Widget _buildSlotTile(DrawCardPreview slot) {
    final int idx = (slot.slotIndex - 1).clamp(0, kWheelColors.length - 1);
    final Color color = kWheelColors[idx];
    final double prob = slot.probability * 100;

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withValues(alpha: 0.22)),
      ),
      clipBehavior: Clip.antiAlias,
      child: IntrinsicHeight(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            Container(width: 4, color: color),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 11,
                ),
                child: Row(
                  children: <Widget>[
                    Container(
                      width: 36,
                      height: 36,
                      decoration: BoxDecoration(
                        color: color.withValues(alpha: 0.13),
                        shape: BoxShape.circle,
                      ),
                      child: Center(
                        child: Text(
                          '${slot.slotIndex}',
                          style: TextStyle(
                            color: color,
                            fontWeight: FontWeight.w900,
                            fontSize: 16,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: <Widget>[
                          Text(
                            slot.templateName.trim().isEmpty
                                ? '格位 ${slot.slotIndex}'
                                : slot.templateName,
                            style: const TextStyle(
                              color: _ink,
                              fontSize: 14,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          if (slot.rarity.trim().isNotEmpty)
                            Text(
                              slot.rarity,
                              style:
                                  const TextStyle(color: _muted, fontSize: 12),
                            ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: color.withValues(alpha: 0.12),
                        borderRadius: BorderRadius.circular(20),
                        border:
                            Border.all(color: color.withValues(alpha: 0.32)),
                      ),
                      child: Text(
                        prob > 0 ? '${prob.toStringAsFixed(1)}%' : '-',
                        style: TextStyle(
                          color: color,
                          fontSize: 13,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyCard(String message) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _line),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 30,
            height: 30,
            decoration: BoxDecoration(
              color: _goldLight.withValues(alpha: 0.5),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(
              Icons.inbox_outlined,
              size: 18,
              color: _gold,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(fontSize: 13, color: _muted),
            ),
          ),
        ],
      ),
    );
  }
}
