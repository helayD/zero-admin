import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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

  Future<void> _openTransferSheet(DigitalCardAssetItem item) async {
    final TextEditingController controller = TextEditingController();
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
                  builder: (BuildContext dialogContext) {
                    return AlertDialog(
                      title: const Text('确认转赠'),
                      content: Text(
                        '确认转赠给 ${recipient['recipientMobileMasked'] ?? mobile} 吗？',
                      ),
                      actions: <Widget>[
                        TextButton(
                          onPressed: () => Navigator.pop(dialogContext, false),
                          child: const Text('取消'),
                        ),
                        FilledButton(
                          onPressed: () => Navigator.pop(dialogContext, true),
                          child: const Text('确认转赠'),
                        ),
                      ],
                    );
                  },
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
              if (transferCompleted) {
                // 必须先关 sheet 再 show snack，否则：
                // (1) await transfer 期间软键盘隐藏触发 MediaQuery rebuild,
                //     sheet 内 Element 重新注册 dependent；
                // (2) SnackBar 与 sheet 同帧争抢 Overlay 层；
                // (3) pop sheet 时 InheritedElement 仍有未清理 dependent →
                //     framework.dart line 6268 `_dependents.isEmpty` assert fail (红屏)。
                if (mounted) {
                  Navigator.pop(sheetContext);
                }
                // Story 10.11 Follow-up: 转赠成功后退出详情页回列表。
                // 卡片已不属于自己，再调 _loadDetail() 必然 400 record_not_found，
                // 用户会看到「加载提货卡详情失败」，体验断裂。直接 pop 让列表自动刷新。
                if (mounted) {
                  _showSnack('转赠成功');
                  Navigator.of(context).pop();
                }
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
          if (detail != null && _shouldShowDrawSummary(detail))
            _buildDrawSummaryCard(detail.drawSummary),
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
    final bool actionsAvailable = item.mintStatus == 'mint_success';
    final bool canRedeem = actionsAvailable;
    final bool canShare = actionsAvailable && item.transferable == true;
    // Story 10.7 Review Fix HIGH-2: 只把 pending/processing/shipped 视为活跃占用，
    // cancelled/delivered/failed 等终态不应锁定卡片，否则会永久禁用分享/提货入口。
    const Set<String> activeRedemptionStatuses = <String>{
      'pending',
      'processing',
      'shipped',
    };
    final bool hasActiveRedemption = item.redemptionStatus != null &&
        activeRedemptionStatuses.contains(item.redemptionStatus);
    final bool hasActiveShare = item.shareTokenStatus == 'active';

    String actionHint = '';
    if (hasActiveRedemption) {
      actionHint = '该卡片正在提货中，暂不能分享';
    } else if (hasActiveShare) {
      actionHint = '该卡片已分享给朋友，暂不能提货';
    }

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
            actionsAvailable
                ? '可支付邮费、提交提现申请，或转赠给已注册用户。'
                : '待发放完成后，可支付邮费、提交提现申请或转赠给已注册用户。',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          if (actionHint.isNotEmpty) ...<Widget>[
            const SizedBox(height: AppSpacing.sm),
            Text(
              actionHint,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: AppColors.price,
                  ),
            ),
          ],
          const SizedBox(height: AppSpacing.md),
          // 主要操作按钮（全宽）
          if (canRedeem && !hasActiveShare)
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _isActionSubmitting
                    ? null
                    : () => _openRedemptionSheet(item),
                icon: const Icon(Icons.local_mall_outlined),
                label: const Text('我要提货'),
              ),
            ),
          if (canShare && !hasActiveRedemption) ...[
            if (canRedeem && !hasActiveShare) const SizedBox(height: AppSpacing.sm),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: _isActionSubmitting
                    ? null
                    : () => _openShareSheet(item),
                icon: const Icon(Icons.share_outlined),
                label: const Text('分享给朋友'),
              ),
            ),
          ],
          // 次要操作按钮
          if (actionsAvailable) ...[
            const SizedBox(height: AppSpacing.md),
            const Divider(),
            const SizedBox(height: AppSpacing.sm),
            Wrap(
              spacing: AppSpacing.md,
              runSpacing: AppSpacing.sm,
              children: <Widget>[
                TextButton.icon(
                  onPressed: _isActionSubmitting
                      ? null
                      : () => _openPhysicalFulfillment(item),
                  icon: const Icon(Icons.local_shipping_outlined, size: 18),
                  label: const Text('支付邮费'),
                ),
                TextButton.icon(
                  onPressed: _isActionSubmitting
                      ? null
                      : () => _requestWithdraw(item),
                  icon: const Icon(Icons.account_balance_wallet_outlined, size: 18),
                  label: const Text('提现'),
                ),
                TextButton.icon(
                  onPressed: _isActionSubmitting
                      ? null
                      : () => _openTransferSheet(item),
                  icon: const Icon(Icons.ios_share_outlined, size: 18),
                  label: const Text('转赠'),
                ),
              ],
            ),
          ],
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
                          ? '提货卡'
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
          _buildMetaLine('资产来源', _assetSourceName(detail.item)),
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
