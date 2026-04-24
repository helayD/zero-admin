import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';
import 'package:flutter_mall/view/digital_card/draw_result_sheet.dart';
import 'package:flutter_mall/view/digital_card/draw_rule_banner.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
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

  DrawActivityLandingData? _landing;
  bool _isLoading = true;
  bool _isParticipating = false;
  bool _useAnonymousLandingOnly = false;
  String? _errorMessage;

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
    return '加载抽卡活动失败，请稍后重试';
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

    setState(() {
      _isParticipating = true;
    });

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
        ),
      );
      await _loadLanding();
    } catch (_) {
      if (!mounted) {
        return;
      }
      _showSnackBar('参与抽卡失败，请稍后重试');
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
        return _isParticipating ? '抽卡中...' : '立即抽卡';
      default:
        return '暂不可参与';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF4F1EA),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(widget.activityTitle?.trim().isNotEmpty == true
            ? widget.activityTitle!.trim()
            : '数字卡片活动'),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline,
                  size: 52, color: Color(0xFFB45309)),
              const SizedBox(height: 12),
              Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 15, height: 1.5),
              ),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _loadLanding,
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
      onRefresh: _loadLanding,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 28),
        children: <Widget>[
          _buildHeroCard(landing),
          const SizedBox(height: 14),
          DrawRuleBanner(
            title: '合规说明',
            content: landing.complianceRuleSummary,
            icon: Icons.verified_user_outlined,
            color: const Color(0xFF3366CC),
          ),
          const SizedBox(height: 12),
          DrawRuleBanner(
            title: '活动规则',
            content: landing.ruleSummary,
            icon: Icons.auto_awesome_outlined,
            color: const Color(0xFFC97C00),
          ),
          const SizedBox(height: 18),
          _buildSectionTitle('卡片预览'),
          const SizedBox(height: 10),
          _buildCardPreviewList(landing),
          const SizedBox(height: 18),
          _buildSectionTitle('卡池说明'),
          const SizedBox(height: 10),
          ...landing.pools.map(_buildPoolCard),
          const SizedBox(height: 18),
          _buildSectionTitle('最近中奖'),
          const SizedBox(height: 10),
          ...landing.recentWins.map(_buildRecentWinTile),
          const SizedBox(height: 18),
          _buildSectionTitle('我的参与记录'),
          const SizedBox(height: 10),
          if (landing.myRecords.isEmpty)
            _buildEmptyCard(
              landing.eligibility.eligibilityCode == 'need_login'
                  ? '登录后可查看你的抽卡记录'
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
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF1F2937), Color(0xFF7C2D12)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(28),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.16),
              borderRadius: BorderRadius.circular(999),
            ),
            child: Text(
              _statusLabel(eligibility),
              style: const TextStyle(
                color: Colors.white,
                fontSize: 12,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            landing.name.trim().isEmpty
                ? widget.activityTitle?.trim() ?? '数字卡片活动'
                : landing.name,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 26,
              height: 1.2,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            landing.participantConditionSummary.trim().isEmpty
                ? '查看资格、规则与卡池后参与抽卡'
                : landing.participantConditionSummary,
            style: const TextStyle(
                color: Colors.white70, fontSize: 14, height: 1.45),
          ),
          const SizedBox(height: 18),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: <Widget>[
              _buildHeroMetric('参与状态', eligibility.eligibilityMessage),
              _buildHeroMetric('剩余次数', '${eligibility.remainingLotteryTimes}'),
              _buildHeroMetric(
                '兑卡实名',
                landing.realNameRequired == 1
                    ? landing.identity.realNameStatusText
                    : '无需实名兑卡',
              ),
            ],
          ),
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            child: FilledButton(
              onPressed: _isParticipating ? null : _handlePrimaryAction,
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFFF3C969),
                foregroundColor: const Color(0xFF1F2937),
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(18),
                ),
              ),
              child: Text(
                _primaryActionText(eligibility),
                style:
                    const TextStyle(fontSize: 15, fontWeight: FontWeight.w700),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeroMetric(String label, String value) {
    return Container(
      constraints: const BoxConstraints(minWidth: 92),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(18),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(color: Colors.white70, fontSize: 12),
          ),
          const SizedBox(height: 4),
          Text(
            value.trim().isEmpty ? '--' : value,
            style: const TextStyle(
                color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700),
    );
  }

  Widget _buildCardPreviewList(DrawActivityLandingData landing) {
    if (landing.cardPreviews.isEmpty) {
      return _buildEmptyCard('当前暂无卡片预览');
    }
    return SizedBox(
      height: 196,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemBuilder: (_, int index) {
          final item = landing.cardPreviews[index];
          return Container(
            width: 150,
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(22),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Expanded(
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: CachedImageWidget(
                      double.infinity,
                      double.infinity,
                      item.cardFaceImage,
                      fit: BoxFit.cover,
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                Text(
                  item.templateName,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 4),
                Text(
                  item.rarity.trim().isEmpty ? item.displayCopy : item.rarity,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                ),
              ],
            ),
          );
        },
        separatorBuilder: (_, __) => const SizedBox(width: 12),
        itemCount: landing.cardPreviews.length,
      ),
    );
  }

  Widget _buildPoolCard(DrawPoolPreview pool) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(pool.poolName,
              style:
                  const TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
          if (pool.probabilityRule.trim().isNotEmpty) ...<Widget>[
            const SizedBox(height: 6),
            Text(
              pool.probabilityRule,
              style: TextStyle(
                  fontSize: 13, color: Colors.grey[700], height: 1.45),
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
                    color: const Color(0xFFF5E7C5),
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: Text(
                    '${card.templateName} ${card.rarity}'.trim(),
                    style: const TextStyle(
                        fontSize: 12, fontWeight: FontWeight.w600),
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
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: const Color(0xFFF5E7C5),
              borderRadius: BorderRadius.circular(14),
            ),
            child: const Icon(Icons.workspace_premium_outlined,
                color: Color(0xFFC97C00)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  '${item.memberNameMasked} 抽中了 ${item.templateName}'.trim(),
                  style: const TextStyle(fontWeight: FontWeight.w600),
                ),
                const SizedBox(height: 4),
                Text(
                  '${item.resultStatusText} · ${item.createTime}'.trim(),
                  style: TextStyle(fontSize: 12, color: Colors.grey[700]),
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
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  item.resultStatusText.trim().isEmpty
                      ? '抽卡记录'
                      : item.resultStatusText,
                  style: const TextStyle(
                      fontSize: 15, fontWeight: FontWeight.w700),
                ),
              ),
              Text(
                item.createTime,
                style: TextStyle(fontSize: 12, color: Colors.grey[600]),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            item.templateName.trim().isEmpty
                ? (item.failureReason.trim().isEmpty
                    ? '本次未命中卡片'
                    : item.failureReason)
                : '卡片: ${item.templateName} · ${item.rarity}',
            style:
                TextStyle(fontSize: 13, color: Colors.grey[800], height: 1.45),
          ),
          if (item.assetStatusText.trim().isNotEmpty) ...<Widget>[
            const SizedBox(height: 6),
            Text(
              item.assetNo.trim().isNotEmpty
                  ? '${item.assetStatusText} · 编号 ${item.assetNo}'
                  : item.assetStatusText,
              style: TextStyle(fontSize: 12, color: Colors.grey[700]),
            ),
          ],
          const SizedBox(height: 6),
          Text(
            '抽奖次数 ${item.lotteryTimesBefore} -> ${item.lotteryTimesAfter}',
            style: TextStyle(fontSize: 12, color: Colors.grey[600]),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyCard(String message) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
      ),
      child: Text(
        message,
        style: TextStyle(fontSize: 13, color: Colors.grey[700]),
      ),
    );
  }
}
