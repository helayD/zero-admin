import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/address_list.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/compliance_rule_banner.dart';
import 'package:flutter_mall/view/digital_card/digital_card_physical_fulfillment_page.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';
import 'package:flutter_mall/view/digital_card/mint_status_timeline.dart';
import 'package:flutter_mall/view/digital_card/physical_fulfillment_address_sheet.dart';
import 'package:flutter_mall/view/digital_card/redemption_order_detail_page.dart';
import 'package:flutter_mall/view/digital_card/share_digital_card_page.dart';
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
        _errorMessage = '加载提货卡详情失败，请稍后重试';
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

  Future<void> _openRedemptionSheet(DigitalCardAssetItem item) async {
    final TextEditingController nameController = TextEditingController();
    final TextEditingController phoneController = TextEditingController();
    final TextEditingController addressController = TextEditingController();
    // Story 10.7 Review Fix: 进入提货表单时异步加载地址簿，用户可一键从默认地址填充
    List<AddressListData> addressBook = <AddressListData>[];
    try {
      final Response addrResp = await HttpUtil.get(addressListDataUrl);
      final AddressListModel model = AddressListModel.fromJson(addrResp.data);
      if (model.code == 0) {
        addressBook = model.data;
        // 默认回填：优先默认地址，其次列表首项
        final AddressListData? defaultAddr = addressBook.isEmpty
            ? null
            : addressBook.firstWhere(
                (AddressListData a) => a.isDefault == 1,
                orElse: () => addressBook.first,
              );
        if (defaultAddr != null) {
          nameController.text = defaultAddr.receiverName;
          phoneController.text = defaultAddr.receiverPhone;
          addressController.text = defaultAddr.fullAddress;
        }
      }
    } catch (_) {
      // 地址簿加载失败不阻断流程，用户仍可手填
    }
    if (!mounted) {
      nameController.dispose();
      phoneController.dispose();
      addressController.dispose();
      return;
    }

    int? createdOrderId;
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (BuildContext sheetContext) {
        bool isSubmitting = false;
        String? message;
        // 实时验证状态
        String? nameError;
        String? phoneError;
        String? addressError;

        // 手机号验证正则
        final RegExp phoneRegExp = RegExp(r'^1[3-9]\d{9}$');

        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setSheetState) {
            // 实时验证函数
            void validateFields() {
              final String name = nameController.text.trim();
              final String phone = phoneController.text.trim();
              final String address = addressController.text.trim();

              setSheetState(() {
                nameError = name.isEmpty ? '请输入收货人姓名' : null;
                phoneError = phone.isEmpty
                    ? '请输入手机号'
                    : (!phoneRegExp.hasMatch(phone) ? '请输入有效的11位手机号' : null);
                addressError = address.isEmpty ? '请输入详细收货地址' : null;
              });
            }

            Future<void> pickFromAddressBook() async {
              if (addressBook.isEmpty) {
                _showSnack('暂无已保存的收货地址');
                return;
              }
              await showModalBottomSheet<void>(
                context: sheetContext,
                isScrollControlled: true,
                showDragHandle: true,
                builder: (_) => PhysicalFulfillmentAddressSheet(
                  addresses: addressBook,
                  selectedAddressId: null,
                  onSelected: (AddressListData picked) {
                    Navigator.of(sheetContext).pop();
                    setSheetState(() {
                      nameController.text = picked.receiverName;
                      phoneController.text = picked.receiverPhone;
                      addressController.text = picked.fullAddress;
                      nameError = null;
                      phoneError = null;
                      addressError = null;
                    });
                  },
                ),
              );
            }

            Future<void> submit() async {
              validateFields();
              if (nameError != null || phoneError != null || addressError != null) {
                return;
              }
              if (isSubmitting) return;

              // 显示确认对话框
              final bool? confirmed = await showDialog<bool>(
                context: sheetContext,
                builder: (BuildContext dialogContext) {
                  return AlertDialog(
                    title: const Text('确认提货'),
                    content: const Text('提交后将创建提货单，确认要提货吗？'),
                    actions: <Widget>[
                      TextButton(
                        onPressed: () => Navigator.pop(dialogContext, false),
                        child: const Text('取消'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.pop(dialogContext, true),
                        child: const Text('确认提货'),
                      ),
                    ],
                  );
                },
              );
              if (confirmed != true) return;

              setSheetState(() {
                isSubmitting = true;
                message = null;
              });
              try {
                final Response resp = await HttpUtil.post(
                  createRedemptionOrderUrl,
                  data: <String, dynamic>{
                    'cardInstanceId': item.assetInstanceId,
                    // Review Fix MEDIUM-1: 不再传 holderId（后端从 JWT 取 memberID）
                    'receiverName': nameController.text.trim(),
                    'receiverPhone': phoneController.text.trim(),
                    'receiverAddress': addressController.text.trim(),
                    'requestId': 'redemption-${DateTime.now().millisecondsSinceEpoch}',
                  },
                );
                final Map<String, dynamic> data = _responseData(resp.data);
                createdOrderId = (data['id'] as num?)?.toInt();
                if (!mounted) return;
                _showSnack('提货单创建成功');
                await _loadDetail();
                if (!mounted) return;
                if (mounted) {
                  Navigator.pop(sheetContext);
                }
              } catch (e) {
                setSheetState(() {
                  isSubmitting = false;
                  message = _getErrorMessage(e);
                });
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
                    Row(
                      children: <Widget>[
                        Expanded(
                          child: Text('我要提货',
                              style: Theme.of(context).textTheme.titleLarge),
                        ),
                        TextButton.icon(
                          onPressed: pickFromAddressBook,
                          icon:
                              const Icon(Icons.location_on_outlined, size: 18),
                          label: const Text('从地址簿选择'),
                        ),
                      ],
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Text(
                      '填写收货信息，我们将为您配送实体卡片。',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextField(
                      controller: nameController,
                      decoration: InputDecoration(
                        labelText: '收货人姓名',
                        border: const OutlineInputBorder(),
                        errorText: nameError,
                      ),
                      onChanged: (_) => validateFields(),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    TextField(
                      controller: phoneController,
                      keyboardType: TextInputType.phone,
                      maxLength: 11,
                      decoration: InputDecoration(
                        labelText: '收货人手机号',
                        border: const OutlineInputBorder(),
                        errorText: phoneError,
                        counterText: '',
                      ),
                      onChanged: (_) => validateFields(),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    TextField(
                      controller: addressController,
                      maxLines: 3,
                      decoration: InputDecoration(
                        labelText: '详细收货地址',
                        border: const OutlineInputBorder(),
                        errorText: addressError,
                      ),
                      onChanged: (_) => validateFields(),
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
                        onPressed: isSubmitting ? null : submit,
                        icon: isSubmitting
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Icon(Icons.local_mall_outlined),
                        label: const Text('提交提货申请'),
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
    nameController.dispose();
    phoneController.dispose();
    addressController.dispose();

    // Story 10.7 Review Fix MEDIUM-3: 提货单创建成功后跳转到提货单详情页，承载物流进度与取消入口
    if (!mounted) return;
    if (createdOrderId != null && createdOrderId! > 0) {
      await Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => RedemptionOrderDetailPage(
            orderId: createdOrderId!,
            cardInstanceId: item.assetInstanceId,
          ),
        ),
      );
      if (mounted) {
        await _loadDetail();
      }
    }
  }

  String _getErrorMessage(dynamic error) {
    if (error is DioException) {
      if (error.response?.statusCode == 400) {
        return '请求参数错误，请检查输入信息';
      } else if (error.response?.statusCode == 401) {
        return '登录已过期，请重新登录';
      } else if (error.response?.statusCode == 403) {
        return '您没有权限执行此操作';
      } else if (error.response?.statusCode == 409) {
        return '该卡片状态已变更，请刷新后重试';
      } else if (error.type == DioExceptionType.connectionTimeout ||
          error.type == DioExceptionType.receiveTimeout) {
        return '网络连接超时，请检查网络后重试';
      } else if (error.type == DioExceptionType.connectionError) {
        return '网络连接失败，请检查网络设置';
      }
    }
    return '操作失败，请稍后重试';
  }

  // Story 10.7 Review Fix CRITICAL-2 / HIGH-4:
  // 分享入口改为跳转到独立的分享页（带二维码 + 有效期 + 可领取次数）。
  // 旧的 bottom sheet 只能复制链接，无法覆盖"扫码=H5 页面"的闭环要求。
  Future<void> _openShareSheet(DigitalCardAssetItem item) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ShareDigitalCardPage(
          assetInstanceId: item.assetInstanceId,
          templateName: item.templateName,
          cardFaceImage: item.cardFaceImage,
        ),
      ),
    );
    if (mounted) {
      await _loadDetail();
    }
  }

  // Story 10.7 Review #5 H1: _openTransferSheet / _resolveTransferRecipient /
  // _transferAsset / _showRegisterPrompt / _openRegisterUrl 已下线。
  // 转赠路径必须使用「分享给朋友」按钮触发的 ShareDigitalCardPage（claim_token 流程）。

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
            : '提货卡详情';

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
      return const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            CircularProgressIndicator(),
            SizedBox(height: AppSpacing.lg),
            Text(
              '正在加载...',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 14,
              ),
            ),
          ],
        ),
      );
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
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.inbox_outlined,
                  size: 52, color: AppColors.textSecondary),
              const SizedBox(height: AppSpacing.md),
              Text(
                '暂无数据',
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
    final DigitalCardStatusCopy statusCopy = digitalCardAssetPrimaryCopy(
      item,
      latestStatusSummary: detail?.latestStatusSummary ?? '',
    );

    // 只在异常状态（受限/回收/下线）时才显示合规说明 banner
    final bool showComplianceBanner =
        detail?.restrictionReason.trim().isNotEmpty == true ||
            item.complianceStatus == 'compliance_restricted' ||
            item.complianceStatus == 'compliance_recycled' ||
            item.displayStatus == 'display_offlined' ||
            item.displayStatus == 'display_recycled';

    return RefreshIndicator(
      onRefresh: _loadDetail,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
        children: <Widget>[
          _buildHeroCard(item, statusCopy),
          if (showComplianceBanner) ...<Widget>[
            const SizedBox(height: AppSpacing.md),
            ComplianceRuleBanner(
              title: '卡片说明',
              summary: detail?.restrictionReason.trim().isNotEmpty == true
                  ? digitalCardUserFacingText(detail!.restrictionReason)
                  : statusCopy.description,
              statusText: statusCopy.label,
            ),
          ],
          const SizedBox(height: AppSpacing.md),
          Text(
            '卡片进度',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: AppSpacing.sm),
          MintStatusTimeline(
              timeline:
                  detail?.timeline ?? const <DigitalCardAssetTimelineItem>[]),
          if (detail != null && _shouldShowDrawSummary(detail)) ...<Widget>[
            const SizedBox(height: AppSpacing.md),
            _buildDrawSummaryCard(detail.drawSummary),
          ],
          if (detail != null) ...<Widget>[
            const SizedBox(height: AppSpacing.md),
            _buildActionAndShippingCard(detail.item),
          ],
        ],
      ),
    );
  }

  /// 合并「可用操作」+「实体卡进度」为一个紧凑卡片
  Widget _buildActionAndShippingCard(DigitalCardAssetItem item) {
    final bool actionsAvailable = item.mintStatus == 'mint_success';
    final bool canRedeem = actionsAvailable;
    final bool canShare = actionsAvailable && item.transferable == true;
    const Set<String> activeRedemptionStatuses = <String>{
      'pending', 'processing', 'shipped',
    };
    final bool hasActiveRedemption = item.redemptionStatus != null &&
        activeRedemptionStatuses.contains(item.redemptionStatus);
    final bool hasActiveShare = item.shareTokenStatus == 'active';

    String actionHint = '';
    if (hasActiveRedemption) {
      actionHint = '配送中，暂不能分享';
    } else if (hasActiveShare) {
      actionHint = '已分享给朋友，暂不能申请配送';
    }

    return Container(
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          // ── 实体卡进度入口（行式，紧凑）──
          InkWell(
            onTap: () => _openPhysicalFulfillment(item),
            borderRadius: BorderRadius.vertical(
              top: Radius.circular(AppRadii.xl),
            ),
            child: Padding(
              padding: const EdgeInsets.symmetric(
                  horizontal: AppSpacing.lg, vertical: AppSpacing.md),
              child: Row(
                children: <Widget>[
                  const Icon(Icons.local_shipping_outlined,
                      size: 18, color: Color(0xFF2563EB)),
                  const SizedBox(width: AppSpacing.sm),
                  Expanded(
                    child: Text(
                      '实体卡配送',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppColors.textPrimary,
                            fontWeight: FontWeight.w500,
                          ),
                    ),
                  ),
                  Text(
                    _physicalEntryStatus(item),
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                  const SizedBox(width: 2),
                  const Icon(Icons.chevron_right_rounded,
                      size: 18, color: AppColors.textSecondary),
                ],
              ),
            ),
          ),
          Divider(height: 1, color: AppColors.border),
          // ── 操作按钮区 ──
          Padding(
            padding: const EdgeInsets.fromLTRB(
                AppSpacing.lg, AppSpacing.md, AppSpacing.lg, AppSpacing.md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                if (actionHint.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(bottom: AppSpacing.sm),
                    child: Row(
                      children: <Widget>[
                        const Icon(Icons.info_outline,
                            size: 13, color: AppColors.price),
                        const SizedBox(width: 4),
                        Expanded(
                          child: Text(
                            actionHint,
                            style:
                                Theme.of(context).textTheme.bodySmall?.copyWith(
                                      color: AppColors.price,
                                    ),
                          ),
                        ),
                      ],
                    ),
                  ),
                Row(
                  children: <Widget>[
                    if (canRedeem && !hasActiveShare)
                      Expanded(
                        child: FilledButton.icon(
                          onPressed: _isActionSubmitting
                              ? null
                              : () => _openRedemptionSheet(item),
                          icon: const Icon(Icons.local_shipping_outlined,
                              size: 16),
                          label: const Text('申请配送'),
                          style: FilledButton.styleFrom(
                            visualDensity: VisualDensity.compact,
                          ),
                        ),
                      ),
                    if (canRedeem && !hasActiveShare && canShare && !hasActiveRedemption)
                      const SizedBox(width: AppSpacing.sm),
                    if (canShare && !hasActiveRedemption)
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: _isActionSubmitting
                              ? null
                              : () => _openShareSheet(item),
                          icon: const Icon(Icons.share_outlined, size: 16),
                          label: const Text('分享给朋友'),
                          style: OutlinedButton.styleFrom(
                            visualDensity: VisualDensity.compact,
                          ),
                        ),
                      ),
                    if (!actionsAvailable)
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: null,
                          icon: const Icon(Icons.hourglass_empty_outlined,
                              size: 16),
                          label: const Text('等待卡片到账'),
                          style: OutlinedButton.styleFrom(
                            visualDensity: VisualDensity.compact,
                          ),
                        ),
                      ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _physicalEntryStatus(DigitalCardAssetItem item) {
    if (item.mintStatus == 'mint_success') return '可申请';
    if (item.mintStatus == 'mint_pending' ||
        item.mintStatus == 'mint_processing') return '待到账';
    if (item.mintStatus == 'mint_failed') return '暂不可用';
    return '查看进度';
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

  Widget _buildHeroCard(
    DigitalCardAssetItem item,
    DigitalCardStatusCopy statusCopy,
  ) {
    // 根据稀有度选择渐变色
    final List<Color> gradientColors = _rarityGradient(item.rarity);

    return Container(
      padding: const EdgeInsets.all(AppSpacing.xl),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: gradientColors,
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(AppRadii.xxl),
        boxShadow: <BoxShadow>[
          BoxShadow(
            color: gradientColors.last.withValues(alpha: 0.35),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              // 卡面图片 / 无图占位
              _buildCardFaceImage(item),
              const SizedBox(width: AppSpacing.lg),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    // 稀有度标签
                    if (item.rarity.trim().isNotEmpty)
                      _buildRarityBadge(item.rarity),
                    if (item.rarity.trim().isNotEmpty)
                      const SizedBox(height: AppSpacing.xs),
                    Text(
                      item.templateName.trim().isEmpty
                          ? '提货卡'
                          : item.templateName,
                      style:
                          Theme.of(context).textTheme.headlineSmall?.copyWith(
                                color: Colors.white,
                                fontSize: 22,
                                fontWeight: FontWeight.w700,
                                height: 1.2,
                              ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Text(
                      '编号 ${item.assetNo.isEmpty ? '待分配' : item.assetNo}',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.white.withValues(alpha: 0.7),
                            letterSpacing: 0.4,
                          ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Wrap(
                      spacing: AppSpacing.sm,
                      runSpacing: AppSpacing.sm,
                      children: <Widget>[
                        _buildHeroChip(statusCopy.label),
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
          const SizedBox(height: AppSpacing.md),
          Divider(color: Colors.white.withValues(alpha: 0.2), height: 1),
          const SizedBox(height: AppSpacing.sm),
          Row(
            children: <Widget>[
              Expanded(
                child: _buildHeroMeta('获取方式', _assetSourceName(item)),
              ),
              Expanded(
                child: _buildHeroMeta(
                  '获取时间',
                  item.obtainedAt.isEmpty ? '待同步' : item.obtainedAt,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// 卡面图片，无图时显示稀有度占位
  Widget _buildCardFaceImage(DigitalCardAssetItem item) {
    const double w = 88;
    const double h = 120;
    final bool hasImage = item.cardFaceImage.trim().isNotEmpty &&
        !item.cardFaceImage.contains('example.com');

    if (hasImage) {
      return ClipRRect(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        child: CachedImageWidget(w, h, item.cardFaceImage),
      );
    }

    // 无图占位：渐变背景 + 稀有度文字
    final String rarityLabel =
        item.rarity.trim().isEmpty ? 'N' : item.rarity.trim().toUpperCase();
    return Container(
      width: w,
      height: h,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        gradient: LinearGradient(
          colors: <Color>[
            Colors.white.withValues(alpha: 0.18),
            Colors.white.withValues(alpha: 0.06),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.25),
          width: 1.5,
        ),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          Icon(
            Icons.card_giftcard_rounded,
            color: Colors.white.withValues(alpha: 0.6),
            size: 32,
          ),
          const SizedBox(height: 6),
          Text(
            rarityLabel,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.85),
              fontSize: 18,
              fontWeight: FontWeight.w900,
              letterSpacing: 1,
            ),
          ),
        ],
      ),
    );
  }

  /// 根据稀有度返回渐变色
  List<Color> _rarityGradient(String rarity) {
    switch (rarity.trim().toUpperCase()) {
      case 'SSR':
        return const <Color>[Color(0xFF1A0533), Color(0xFF7C3AED)];
      case 'SR':
        return const <Color>[Color(0xFF0F2044), Color(0xFF1D4ED8)];
      case 'R':
        return const <Color>[Color(0xFF0D2B1F), Color(0xFF059669)];
      default:
        return const <Color>[Color(0xFF111827), Color(0xFF374151)];
    }
  }

  /// 稀有度角标
  Widget _buildRarityBadge(String rarity) {
    final Color badgeColor;
    switch (rarity.trim().toUpperCase()) {
      case 'SSR':
        badgeColor = const Color(0xFFF59E0B);
      case 'SR':
        badgeColor = const Color(0xFF60A5FA);
      case 'R':
        badgeColor = const Color(0xFF34D399);
      default:
        badgeColor = Colors.white54;
    }
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: badgeColor.withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: badgeColor.withValues(alpha: 0.6)),
      ),
      child: Text(
        rarity.trim().toUpperCase(),
        style: TextStyle(
          color: badgeColor,
          fontSize: 11,
          fontWeight: FontWeight.w800,
          letterSpacing: 1.5,
        ),
      ),
    );
  }

  Widget _buildDrawSummaryCard(DigitalCardAssetDrawSummary summary) {
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          tilePadding:
              const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
          childrenPadding: const EdgeInsets.fromLTRB(
            AppSpacing.lg, 0, AppSpacing.lg, AppSpacing.lg,
          ),
          title: Text(
            '抽卡来源',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          children: <Widget>[
            _buildMetaLine(
              '结果状态',
              digitalCardUserFacingText(
                  summary.resultStatusText.trim().isEmpty
                      ? summary.resultStatus
                      : summary.resultStatusText),
            ),
            _buildMetaLine(
              '参与时间',
              summary.createTime.isEmpty ? '待同步' : summary.createTime,
            ),
            if (summary.failureReason.trim().isNotEmpty)
              _buildMetaLine(
                '失败原因',
                digitalCardUserFacingText(summary.failureReason),
              ),
          ],
        ),
      ),
    );
  }

  bool _shouldShowDrawSummary(DigitalCardAssetDetailData detail) {
    return detail.item.sourceType == 'draw' &&
        detail.drawSummary.participationRecordId > 0;
  }

  String _assetSourceName(DigitalCardAssetItem item) {
    if (item.sourceDisplayName.trim().isNotEmpty) {
      return item.sourceDisplayName.trim();
    }
    if (item.activityName.trim().isNotEmpty) {
      return item.activityName.trim();
    }
    return '待同步';
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

  Widget _buildHeroMeta(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: Theme.of(context).textTheme.labelSmall?.copyWith(
                color: Colors.white.withValues(alpha: 0.6),
              ),
        ),
        const SizedBox(height: 2),
        Text(
          value.trim().isEmpty ? '-' : value,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Colors.white.withValues(alpha: 0.9),
              ),
        ),
      ],
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
