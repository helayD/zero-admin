import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/compliance_rule_banner.dart';
import 'package:flutter_mall/view/digital_card/mint_status_timeline.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

typedef DigitalCardAssetDetailFetcher
    = Future<QueryMyDigitalCardAssetDetailResponse> Function(
        int assetInstanceId);

class DigitalCardAssetDetailPage extends StatefulWidget {
  final int assetInstanceId;
  final DigitalCardAssetItem? initialItem;
  final String? intentSource;
  final DigitalCardAssetDetailFetcher? fetchDetail;

  const DigitalCardAssetDetailPage({
    super.key,
    required this.assetInstanceId,
    this.initialItem,
    this.intentSource,
    this.fetchDetail,
  });

  @override
  State<DigitalCardAssetDetailPage> createState() =>
      _DigitalCardAssetDetailPageState();
}

class _DigitalCardAssetDetailPageState
    extends State<DigitalCardAssetDetailPage> {
  DigitalCardAssetDetailData? _detail;
  bool _isLoading = true;
  String? _errorMessage;

  AppRecentContext _buildRecoveryContext([String? source]) {
    return AppRecentContext.create(
      targetType: AppRecentTargetType.digitalCardAssetDetail,
      targetId: widget.assetInstanceId,
      source: source ?? widget.intentSource ?? 'digital_card_asset_detail',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.digitalCardAssetList,
    );
  }

  @override
  void initState() {
    super.initState();
    AppRecoveryStore.saveActiveIntentCandidate(_buildRecoveryContext());
    _loadDetail();
  }

  @override
  void dispose() {
    AppRecoveryStore.clearActiveIntentCandidateIfMatches(
      AppRecentTargetType.digitalCardAssetDetail,
      targetId: widget.assetInstanceId,
    );
    super.dispose();
  }

  Future<void> _loadDetail() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final QueryMyDigitalCardAssetDetailResponse parsed =
          await (widget.fetchDetail?.call(widget.assetInstanceId) ??
              _fetchDetailFromApi());
      if (!mounted) {
        return;
      }

      setState(() {
        _detail = parsed.data;
        _isLoading = false;
      });
      await AppRecoveryStore.saveRecentContext(
        _buildRecoveryContext('digital_card_asset_detail_view'),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = '加载数字卡片详情失败，请稍后重试';
        _detail = null;
        _isLoading = false;
      });
    }
  }

  Future<QueryMyDigitalCardAssetDetailResponse> _fetchDetailFromApi() async {
    final Response response = await HttpUtil.get(
      queryMyDigitalCardAssetDetailUrl,
      queryParameters: <String, dynamic>{
        'assetInstanceId': widget.assetInstanceId,
      },
    );
    return queryMyDigitalCardAssetDetailResponseFromJson(
      jsonEncode(response.data),
    );
  }

  @override
  Widget build(BuildContext context) {
    final String title = _detail?.item.templateName.trim().isNotEmpty == true
        ? _detail!.item.templateName
        : widget.initialItem?.templateName.trim().isNotEmpty == true
            ? widget.initialItem!.templateName
            : '数字卡片详情';

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Text(title),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading && _detail == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null && _detail == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline,
                  size: 52, color: Color(0xFFB42318)),
              const SizedBox(height: AppSpacing.md),
              Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                onPressed: _loadDetail,
                child: const Text('重新加载'),
              ),
            ],
          ),
        ),
      );
    }

    final DigitalCardAssetDetailData? detail = _detail;
    final DigitalCardAssetItem? item = detail?.item ?? widget.initialItem;
    if (item == null) {
      return const SizedBox.shrink();
    }

    return RefreshIndicator(
      onRefresh: _loadDetail,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        children: <Widget>[
          _buildHeroCard(item, detail),
          const SizedBox(height: AppSpacing.lg),
          ComplianceRuleBanner(
            title: '合规说明',
            summary: detail?.restrictionReason.trim().isNotEmpty == true
                ? detail!.restrictionReason
                : item.complianceRuleSummary,
            statusText: item.complianceStatusText.trim().isNotEmpty
                ? item.complianceStatusText
                : item.displayStatusText,
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            '到账进度时间线',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: AppSpacing.md),
          MintStatusTimeline(
              timeline:
                  detail?.timeline ?? const <DigitalCardAssetTimelineItem>[]),
          const SizedBox(height: AppSpacing.lg),
          if (detail != null) _buildDrawSummaryCard(detail.drawSummary),
          if (detail != null) ...<Widget>[
            const SizedBox(height: AppSpacing.lg),
            _buildStatusCard(detail),
          ],
        ],
      ),
    );
  }

  Widget _buildHeroCard(
    DigitalCardAssetItem item,
    DigitalCardAssetDetailData? detail,
  ) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.xl),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF111827), Color(0xFF8B1E3F)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(AppRadii.xxl),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              ClipRRect(
                borderRadius: BorderRadius.circular(AppRadii.lg),
                child: CachedImageWidget(
                  96,
                  128,
                  item.cardFaceImage,
                ),
              ),
              const SizedBox(width: AppSpacing.lg),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      item.templateName.trim().isEmpty
                          ? '数字卡片'
                          : item.templateName,
                      style:
                          Theme.of(context).textTheme.headlineSmall?.copyWith(
                                color: Colors.white,
                                fontSize: 24,
                              ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Text(
                      '编号 ${item.assetNo.isEmpty ? '待分配' : item.assetNo}',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Colors.white.withValues(alpha: 0.86),
                          ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Wrap(
                      spacing: AppSpacing.sm,
                      runSpacing: AppSpacing.sm,
                      children: <Widget>[
                        _buildHeroChip(
                          item.mintStatusText.trim().isEmpty
                              ? item.mintStatus
                              : item.mintStatusText,
                        ),
                        _buildHeroChip(
                          item.tokenStatusText.trim().isEmpty
                              ? item.chainStatusText
                              : item.tokenStatusText,
                        ),
                        if (item.displayStatusText.trim().isNotEmpty &&
                            item.displayStatus != 'display_visible')
                          _buildHeroChip(item.displayStatusText),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            detail?.latestStatusSummary.trim().isNotEmpty == true
                ? detail!.latestStatusSummary
                : item.complianceRuleSummary.trim().isNotEmpty
                    ? item.complianceRuleSummary
                    : '当前状态以服务端确认为准。',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.white.withValues(alpha: 0.8),
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildDrawSummaryCard(DigitalCardAssetDrawSummary summary) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            '抽卡来源摘要',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: AppSpacing.md),
          _buildMetaLine('抽卡记录', '${summary.participationRecordId}'),
          _buildMetaLine(
            '结果状态',
            summary.resultStatusText.trim().isEmpty
                ? summary.resultStatus
                : summary.resultStatusText,
          ),
          _buildMetaLine(
              '参与时间', summary.createTime.isEmpty ? '待同步' : summary.createTime),
          if (summary.failureReason.trim().isNotEmpty)
            _buildMetaLine('失败原因', summary.failureReason),
        ],
      ),
    );
  }

  Widget _buildStatusCard(DigitalCardAssetDetailData detail) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            '状态与标识',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: AppSpacing.md),
          _buildMetaLine('所属活动', detail.item.activityName),
          _buildMetaLine(
            '获取时间',
            detail.item.obtainedAt.isEmpty ? '待同步' : detail.item.obtainedAt,
          ),
          _buildMetaLine(
            '链上状态',
            detail.item.chainStatusText.isEmpty
                ? detail.item.chainStatus
                : detail.item.chainStatusText,
          ),
          _buildMetaLine(
            '展示状态',
            detail.item.displayStatusText.isEmpty
                ? detail.item.displayStatus
                : detail.item.displayStatusText,
          ),
          _buildMetaLine(
            '合规状态',
            detail.item.complianceStatusText.isEmpty
                ? detail.item.complianceStatus
                : detail.item.complianceStatusText,
          ),
          if (detail.tokenIdMasked.trim().isNotEmpty)
            _buildMetaLine('脱敏 token', detail.tokenIdMasked),
        ],
      ),
    );
  }

  Widget _buildHeroChip(String text) {
    final String trimmed = text.trim();
    if (trimmed.isEmpty) {
      return const SizedBox.shrink();
    }
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.sm,
        vertical: AppSpacing.xs,
      ),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.14),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        trimmed,
        style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: Colors.white,
            ),
      ),
    );
  }

  Widget _buildMetaLine(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            width: 84,
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
          Expanded(
            child: Text(
              value.trim().isEmpty ? '-' : value,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: AppColors.textPrimary,
                  ),
            ),
          ),
        ],
      ),
    );
  }
}
