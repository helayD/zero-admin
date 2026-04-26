import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/compliance_rule_banner.dart';
import 'package:flutter_mall/view/digital_card/digital_card_physical_fulfillment_page.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';
import 'package:flutter_mall/view/digital_card/mint_status_timeline.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:url_launcher/url_launcher.dart';

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
  bool _isActionSubmitting = false;
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

  Future<void> _requestWithdraw(DigitalCardAssetItem item) async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: const Text('提现'),
        content: const Text('确认提交这张卡片的提现申请吗？'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('提交'),
          ),
        ],
      ),
    );
    if (confirmed != true || _isActionSubmitting) return;
    setState(() => _isActionSubmitting = true);
    try {
      final Response response = await HttpUtil.post(
        requestDigitalCardWithdrawUrl,
        data: <String, dynamic>{
          'assetInstanceId': item.assetInstanceId,
          'reason': '会员提交提现申请',
          'requestId': 'withdraw-${DateTime.now().millisecondsSinceEpoch}',
        },
      );
      final Map<String, dynamic> data = _responseData(response.data);
      if (!mounted) return;
      _showSnack(data['withdrawText']?.toString() ?? '提现申请已提交');
      await _loadDetail();
    } finally {
      if (mounted) setState(() => _isActionSubmitting = false);
    }
  }

  Future<void> _openTransferSheet(DigitalCardAssetItem item) async {
    final TextEditingController controller = TextEditingController();
    final BuildContext pageContext = context;
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (BuildContext sheetContext) {
        bool isChecking = false;
        String? message;
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setSheetState) {
            Future<void> submit() async {
              final String mobile = controller.text.trim();
              if (mobile.isEmpty || isChecking) return;
              bool transferCompleted = false;
              setSheetState(() {
                isChecking = true;
                message = null;
              });
              try {
                final Map<String, dynamic> recipient =
                    await _resolveTransferRecipient(item, mobile);
                final bool canTransfer = recipient['canTransfer'] == true;
                if (!canTransfer) {
                  setSheetState(() {
                    message = recipient['actionHint']?.toString() ??
                        recipient['recipientStatusText']?.toString() ??
                        '接收人需先完成注册';
                  });
                  await _showRegisterPrompt(recipient);
                  return;
                }
                final bool? confirmed = await showDialog<bool>(
                  context: sheetContext,
                  builder: (BuildContext context) => AlertDialog(
                    title: const Text('确认转赠'),
                    content: Text(
                      '确认转赠给 ${recipient['recipientMobileMasked'] ?? mobile} 吗？',
                    ),
                    actions: <Widget>[
                      TextButton(
                        onPressed: () => Navigator.pop(context, false),
                        child: const Text('取消'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.pop(context, true),
                        child: const Text('确认转赠'),
                      ),
                    ],
                  ),
                );
                if (confirmed != true) return;
                await _transferAsset(item, mobile);
                transferCompleted = true;
              } catch (_) {
                setSheetState(() => message = '转赠处理失败，请稍后重试');
              } finally {
                if (!transferCompleted) {
                  setSheetState(() => isChecking = false);
                }
              }
              if (transferCompleted && mounted) {
                Navigator.pop(sheetContext);
                _showSnack('转赠成功');
                Navigator.of(pageContext).pop();
              }
            }

            return SafeArea(
              top: false,
              child: Padding(
                padding: EdgeInsets.fromLTRB(
                  20,
                  8,
                  20,
                  20 + MediaQuery.of(context).viewInsets.bottom,
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text('转赠卡片', style: Theme.of(context).textTheme.titleLarge),
                    const SizedBox(height: AppSpacing.sm),
                    Text(
                      '输入接收人的注册手机号。已注册可直接接收，未注册会生成 H5 注册提示页。',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextField(
                      controller: controller,
                      keyboardType: TextInputType.phone,
                      maxLength: 11,
                      decoration: const InputDecoration(
                        labelText: '接收人手机号',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    if (message != null) ...<Widget>[
                      const SizedBox(height: AppSpacing.sm),
                      Text(
                        message!,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: AppColors.price,
                            ),
                      ),
                    ],
                    const SizedBox(height: AppSpacing.md),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        onPressed: isChecking ? null : submit,
                        icon: isChecking
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child:
                                    CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Icon(Icons.ios_share_outlined),
                        label: const Text('确认接收人'),
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
    controller.dispose();
  }

  Future<Map<String, dynamic>> _resolveTransferRecipient(
    DigitalCardAssetItem item,
    String mobile,
  ) async {
    final Response response = await HttpUtil.post(
      resolveDigitalCardTransferRecipientUrl,
      data: <String, dynamic>{
        'assetInstanceId': item.assetInstanceId,
        'recipientMobile': mobile,
      },
    );
    return _responseData(response.data);
  }

  Future<void> _transferAsset(DigitalCardAssetItem item, String mobile) async {
    await HttpUtil.post(
      transferDigitalCardAssetUrl,
      data: <String, dynamic>{
        'assetInstanceId': item.assetInstanceId,
        'recipientMobile': mobile,
        'requestId': 'transfer-${DateTime.now().millisecondsSinceEpoch}',
      },
    );
  }

  Future<void> _showRegisterPrompt(Map<String, dynamic> recipient) async {
    final String registerUrl = recipient['registerUrl']?.toString() ?? '';
    if (registerUrl.trim().isEmpty || !mounted) return;
    await showDialog<void>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: const Text('接收人需注册'),
        content: Text('请将注册页面发给接收人：$registerUrl'),
        actions: <Widget>[
          TextButton(
            onPressed: () async {
              await Clipboard.setData(ClipboardData(text: registerUrl));
              if (context.mounted) Navigator.pop(context);
              _showSnack('注册链接已复制');
            },
            child: const Text('复制链接'),
          ),
          FilledButton(
            onPressed: () async {
              Navigator.pop(context);
              await _openRegisterUrl(registerUrl);
            },
            child: const Text('打开页面'),
          ),
        ],
      ),
    );
  }

  Future<void> _openRegisterUrl(String registerUrl) async {
    final String normalized =
        registerUrl.startsWith('http') ? registerUrl : '$baseUrl$registerUrl';
    final Uri? uri = Uri.tryParse(normalized);
    if (uri == null) return;
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  }

  Map<String, dynamic> _responseData(dynamic raw) {
    if (raw is Map && raw['data'] is Map) {
      return Map<String, dynamic>.from(raw['data'] as Map);
    }
    return <String, dynamic>{};
  }

  void _showSnack(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
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
    final DigitalCardStatusCopy statusCopy = digitalCardAssetPrimaryCopy(
      item,
      latestStatusSummary: detail?.latestStatusSummary ?? '',
    );

    return RefreshIndicator(
      onRefresh: _loadDetail,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        children: <Widget>[
          _buildHeroCard(item, statusCopy),
          const SizedBox(height: AppSpacing.lg),
          ComplianceRuleBanner(
            title: '合规说明',
            summary: detail?.restrictionReason.trim().isNotEmpty == true
                ? digitalCardUserFacingText(detail!.restrictionReason)
                : statusCopy.description,
            statusText: statusCopy.label,
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            '到账进度',
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
            _buildStatusCard(detail, statusCopy),
            const SizedBox(height: AppSpacing.lg),
            _buildPhysicalFulfillmentEntry(detail.item),
            const SizedBox(height: AppSpacing.lg),
            _buildAssetActionCard(detail.item),
          ],
        ],
      ),
    );
  }

  Widget _buildAssetActionCard(DigitalCardAssetItem item) {
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
          Text('卡片操作', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.sm),
          Text(
            '可支付邮费、提交提现申请，或转赠给已注册用户。',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: AppSpacing.md),
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.sm,
            children: <Widget>[
              FilledButton.icon(
                onPressed: _isActionSubmitting
                    ? null
                    : () => _openPhysicalFulfillment(item),
                icon: const Icon(Icons.local_shipping_outlined),
                label: const Text('支付邮费'),
              ),
              OutlinedButton.icon(
                onPressed:
                    _isActionSubmitting ? null : () => _requestWithdraw(item),
                icon: const Icon(Icons.account_balance_wallet_outlined),
                label: const Text('提现'),
              ),
              OutlinedButton.icon(
                onPressed:
                    _isActionSubmitting ? null : () => _openTransferSheet(item),
                icon: const Icon(Icons.ios_share_outlined),
                label: const Text('转赠'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPhysicalFulfillmentEntry(DigitalCardAssetItem item) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Row(
        children: <Widget>[
          const Icon(Icons.local_shipping_outlined, color: Color(0xFF2563EB)),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  '实体卡进度',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        color: AppColors.textPrimary,
                      ),
                ),
                const SizedBox(height: AppSpacing.xs),
                Text(
                  _physicalEntryHint(item),
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
          ),
          IconButton(
            tooltip: '查看实体卡进度',
            onPressed: () => _openPhysicalFulfillment(item),
            icon: const Icon(Icons.chevron_right_rounded),
          ),
        ],
      ),
    );
  }

  Future<void> _openPhysicalFulfillment(DigitalCardAssetItem item) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => DigitalCardPhysicalFulfillmentPage(
          assetInstanceId: item.assetInstanceId,
        ),
      ),
    );
  }

  String _physicalEntryHint(DigitalCardAssetItem item) {
    if (item.mintStatus == 'mint_success') {
      return '查看地址确认、制作、配送和签收进度。';
    }
    if (item.mintStatus == 'mint_pending' ||
        item.mintStatus == 'mint_processing') {
      return '待到账后可继续确认地址和查看制作配送进度。';
    }
    if (item.mintStatus == 'mint_failed') {
      return '当前暂不可发货，请等待处理结果更新。';
    }
    return '查看实体卡制作与配送进度。';
  }

  Widget _buildHeroCard(
    DigitalCardAssetItem item,
    DigitalCardStatusCopy statusCopy,
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
                          statusCopy.label,
                        ),
                        _buildHeroChip(
                          digitalCardDisplayStatusText(
                            item.displayStatus,
                            item.displayStatusText,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            statusCopy.description,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.white.withValues(alpha: 0.8),
                ),
          ),
          if (statusCopy.actionHint.trim().isNotEmpty) ...<Widget>[
            const SizedBox(height: AppSpacing.xs),
            Text(
              statusCopy.actionHint,
              style: Theme.of(context).textTheme.labelMedium?.copyWith(
                    color: Colors.white,
                    fontWeight: FontWeight.w700,
                  ),
            ),
          ],
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
            digitalCardUserFacingText(summary.resultStatusText.trim().isEmpty
                ? summary.resultStatus
                : summary.resultStatusText),
          ),
          _buildMetaLine(
              '参与时间', summary.createTime.isEmpty ? '待同步' : summary.createTime),
          if (summary.failureReason.trim().isNotEmpty)
            _buildMetaLine(
              '失败原因',
              digitalCardUserFacingText(summary.failureReason),
            ),
        ],
      ),
    );
  }

  Widget _buildStatusCard(
    DigitalCardAssetDetailData detail,
    DigitalCardStatusCopy statusCopy,
  ) {
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
          _buildMetaLine('当前进展', statusCopy.label),
          _buildMetaLine('状态说明', statusCopy.description),
          if (statusCopy.actionHint.trim().isNotEmpty)
            _buildMetaLine('下一步', statusCopy.actionHint),
          _buildMetaLine('所属活动', detail.item.activityName),
          _buildMetaLine(
            '获取时间',
            detail.item.obtainedAt.isEmpty ? '待同步' : detail.item.obtainedAt,
          ),
          _buildMetaLine(
            '展示状态',
            digitalCardDisplayStatusText(
              detail.item.displayStatus,
              detail.item.displayStatusText,
            ),
          ),
          _buildMetaLine(
            '合规状态',
            digitalCardUserFacingText(detail.item.complianceStatusText.isEmpty
                ? detail.item.complianceStatus
                : detail.item.complianceStatusText),
          ),
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
