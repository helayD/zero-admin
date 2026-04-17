import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/layout/upgrade_gate_page.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/after_sales.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/widgets/permission_prompt_sheet.dart';
import 'package:flutter_spinkit/flutter_spinkit.dart';
import 'package:provider/provider.dart';

///
/// 售后申请页面（Story 6-4 Task 9）
///
/// 功能：
/// - 售后类型选择（退货退款 / 仅退款 / 换货）
/// - 退货原因下拉选择（从接口加载）
/// - 问题描述（可选）
/// - 凭证图片上传（最多 3 张，base64 方案）
/// - 提交后展示售后单号 + 成功提示
///
/// Intent Recovery 预留（Story 9-3）：
/// - intentSource?: string 参数为 Epic 9 预留
///
class ApplyAfterSales extends StatefulWidget {
  /// 订单ID（必填）
  final int orderId;

  /// Intent 来源（Story 9-3 预留）
  final String? intentSource;

  final PermissionBroker permissionBroker;

  final Future<ReturnReasonListData> Function()? reasonLoader;
  final AfterSalesSubmitter? submitter;

  const ApplyAfterSales({
    super.key,
    required this.orderId,
    this.intentSource,
    this.permissionBroker = const PermissionBroker(),
    this.reasonLoader,
    this.submitter,
  });

  @override
  State<ApplyAfterSales> createState() => ApplyAfterSalesState();
}

// ==================== 页面状态枚举（外置避免与字段名冲突）====================
enum SalesPageState {
  loading,
  ready,
  submitting,
  success,
  error,
}

typedef AfterSalesSubmitter = Future<ApplyAfterSalesRespData> Function(
  ApplyAfterSalesReqData request,
);

class ApplyAfterSalesState extends State<ApplyAfterSales> {
  // ==================== 表单控制器 ====================
  final _formKey = GlobalKey<FormState>();
  final _descController = TextEditingController();

  // ==================== 状态变量 ====================
  SalesPageState _pageState = SalesPageState.loading;
  AfterSalesType _selectedType = AfterSalesType.returnRefund;
  ReturnReasonItemData? _selectedReason;
  List<String> _proofPics = [];
  String? _errorMessage;
  String? _successReturnNo;
  List<ReturnReasonItemData> _reasonList = [];
  bool _isSubmitting = false;
  AppLifecycleProvider? _lifecycleProvider;
  int _lastResumeTick = 0;

  // ==================== 生命周期 ====================
  @override
  void initState() {
    super.initState();
    _descController.addListener(_handleDraftChanged);
    _restoreDraft();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadReasonList();
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final provider = context.read<AppLifecycleProvider>();
    if (_lifecycleProvider == provider) {
      return;
    }
    _lifecycleProvider?.removeListener(_handleLifecycleChanged);
    _lifecycleProvider = provider;
    _lastResumeTick = provider.resumeTick;
    provider.addListener(_handleLifecycleChanged);
  }

  @override
  void dispose() {
    _lifecycleProvider?.removeListener(_handleLifecycleChanged);
    _descController.removeListener(_handleDraftChanged);
    _descController.dispose();
    super.dispose();
  }

  void _handleLifecycleChanged() {
    final provider = _lifecycleProvider;
    if (provider == null) {
      return;
    }
    if (_lastResumeTick == provider.resumeTick) {
      return;
    }
    _lastResumeTick = provider.resumeTick;
    _resumePendingPermissionFlow();
  }

  void _handleDraftChanged() {
    _persistDraft();
  }

  void _restoreDraft() {
    final draft = AppRecoveryStore.getAfterSalesDraft(widget.orderId);
    if (draft == null) {
      return;
    }
    _selectedType = AfterSalesType.fromValue(draft.typeValue);
    _proofPics = List<String>.from(draft.proofPics.take(3));
    _descController.text = draft.description;
    if (draft.reasonId != null && draft.reasonId! > 0) {
      _selectedReason = ReturnReasonItemData(
        id: draft.reasonId!,
        name: draft.reasonName,
      );
    }
  }

  Future<void> _persistDraft() async {
    await AppRecoveryStore.saveAfterSalesDraft(
      AfterSalesDraftSnapshot(
        orderId: widget.orderId,
        typeValue: _selectedType.value,
        reasonId: _selectedReason?.id,
        reasonName: _selectedReason?.name ?? '',
        description: _descController.text,
        proofPics: _proofPics,
        updatedAt: DateTime.now(),
      ),
    );
  }

  // ==================== 加载原因列表 ====================
  Future<void> _loadReasonList() async {
    final ready = await _ensureUpgradeReady();
    if (!ready || !mounted) {
      return;
    }
    try {
      final data = widget.reasonLoader != null
          ? await widget.reasonLoader!.call()
          : await _loadReasonListFromApi();
      if (!mounted) return;
      setState(() {
        _reasonList = data.reasonList;
        if (_selectedReason != null) {
          final matched =
              _reasonList.where((item) => item.id == _selectedReason!.id);
          if (matched.isNotEmpty) {
            _selectedReason = matched.first;
          }
        }
        _pageState = SalesPageState.ready;
      });
      await AppRecoveryStore.saveRecentContext(
        AppRecentContext.create(
          targetType: AppRecentTargetType.afterSalesApply,
          targetId: widget.orderId,
          source: widget.intentSource ?? 'manual_open',
          requiresAuth: true,
          fallbackType: AppRecentTargetType.orderDetail,
          fallbackTargetId: widget.orderId,
        ),
      );
      await _resumePendingPermissionFlow();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _pageState = SalesPageState.error;
        _errorMessage = '加载失败，请重试';
      });
    }
  }

  Future<bool> _ensureUpgradeReady() async {
    final policy = await UpgradeGateService.queryPolicy(
      scene: 'after_sales_apply',
      targetType: appRecentTargetTypeToValue(
        AppRecentTargetType.afterSalesApply,
      ),
      targetId: widget.orderId,
    );
    if (!policy.hasUpgradeGate) {
      return true;
    }
    if (!mounted) {
      return true;
    }
    final navigator = Navigator.of(context);
    final pendingUpgrade = PendingUpgradeContext.forRecentContext(
      scene: 'after_sales_apply',
      recoveryContext: AppRecentContext.create(
        targetType: AppRecentTargetType.afterSalesApply,
        targetId: widget.orderId,
        source: widget.intentSource ?? 'after_sales_apply',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderDetail,
        fallbackTargetId: widget.orderId,
      ),
      policy: policy,
    );
    await AppRecoveryStore.savePendingUpgradeContext(pendingUpgrade);
    if (!mounted) {
      return true;
    }
    final result = await navigator.push<bool>(
      MaterialPageRoute(
        builder: (_) => UpgradeGatePage(
          pendingContext: pendingUpgrade,
          initialPolicy: policy,
          onResolved: (gateContext, _) async {
            Navigator.of(gateContext).pop(true);
          },
          onContinueLater: policy.canContinueLater
              ? (gateContext, _) async {
                  Navigator.of(gateContext).pop(true);
                }
              : null,
        ),
      ),
    );
    return result == true;
  }

  Future<ReturnReasonListData> _loadReasonListFromApi() async {
    final resp = await HttpUtil.post(queryReturnReasonListUrl);
    return ReturnReasonListData.fromJson(resp.data);
  }

  Future<ApplyAfterSalesRespData> _submitAfterSales(
    ApplyAfterSalesReqData request,
  ) async {
    if (widget.submitter != null) {
      return widget.submitter!(request);
    }
    final resp = await HttpUtil.post(
      applyAfterSalesUrl,
      data: request.toJson(),
    );
    return ApplyAfterSalesRespData.fromJson(resp.data);
  }

  Future<void> _resumePendingPermissionFlow() async {
    final pending = AppRecoveryStore.peekPendingPermissionContext();
    if (pending == null ||
        !pending.matchesReturnTarget(
          AppRecentTargetType.afterSalesApply,
          targetId: widget.orderId,
        )) {
      return;
    }

    final lostMedia = AppRecoveryStore.peekPendingLostMedia();
    if (lostMedia != null &&
        lostMedia.recoveryId == pending.recoveryId &&
        lostMedia.scene == PermissionScene.afterSalesProof) {
      await AppRecoveryStore.clearPendingLostMedia();
      await AppRecoveryStore.clearPendingPermissionContext();
      _applyRecoveredPics(lostMedia.mediaBase64List);
      _recordPermissionLostDataRecovered(pending);
      if (mounted) {
        _showToast('已恢复上次未完成的凭证图片');
      }
      return;
    }

    if (pending.status != PermissionFlowStatus.settingsReturnPending) {
      return;
    }

    await AppRecoveryStore.clearPendingPermissionContext();
    final inspection = await widget.permissionBroker.inspect(pending);
    if (inspection.isUsable && pending.retryOnResume) {
      _recordPermissionGranted(pending);
      await _pickImageWithFlow(
        pending.copyWith(status: inspection.status),
      );
      return;
    }

    _recordPermissionDenied(pending);
    if (mounted) {
      _showToast('权限仍未开启，你仍可继续无图提交售后申请');
    }
  }

  // ==================== 图片上传 ====================
  /// 阶段一（当前 Story）：转为 base64
  /// 阶段二（未来 Story）：上传 OSS 后获取 URL
  Future<void> _handlePickImage(PermissionFlowSource source) async {
    if (_proofPics.length >= 3) {
      _showToast('最多上传 3 张凭证图片');
      return;
    }

    final flow = _buildPermissionFlow(source);
    final inspection = await widget.permissionBroker.inspect(flow);
    if (inspection.isUsable) {
      await _pickImageWithFlow(flow.copyWith(status: inspection.status));
      return;
    }

    final action = await _showPermissionPrompt(
      flow,
      inspection.status,
      alternativeLabel: _alternativeLabel(source),
    );
    if (!mounted || action == null) {
      return;
    }

    await _handlePermissionPromptAction(
      flow,
      action,
      currentStatus: inspection.status,
    );
  }

  void _showPickImageSheet() {
    showModalBottomSheet(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_library),
              title: const Text('相册'),
              onTap: () {
                Navigator.pop(ctx);
                _handlePickImage(PermissionFlowSource.gallery);
              },
            ),
            ListTile(
              leading: const Icon(Icons.camera_alt),
              title: const Text('拍照'),
              onTap: () {
                Navigator.pop(ctx);
                _handlePickImage(PermissionFlowSource.camera);
              },
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  void _removePic(int index) {
    setState(() {
      _proofPics = List.from(_proofPics)..removeAt(index);
    });
    _persistDraft();
  }

  PermissionFlowContext _buildPermissionFlow(PermissionFlowSource source) {
    return PermissionFlowContext.create(
      scene: PermissionScene.afterSalesProof,
      permissionType: source == PermissionFlowSource.camera
          ? PermissionType.camera
          : PermissionType.photos,
      source: source,
      returnTarget: AppRecentTargetType.afterSalesApply,
      returnTargetId: widget.orderId,
      orderId: widget.orderId,
      fallbackAction: source == PermissionFlowSource.camera
          ? PermissionFallbackAction.switchToGallery
          : PermissionFallbackAction.switchToCamera,
      intentId:
          'after_sales_${permissionFlowSourceToValue(source)}_${widget.orderId}',
      recoveryId:
          'after_sales_${permissionFlowSourceToValue(source)}_${widget.orderId}',
      requiresAuth: true,
      retryOnResume: true,
    );
  }

  Future<void> _pickImageWithFlow(PermissionFlowContext flow) async {
    final result = await widget.permissionBroker.pickImageAsBase64(flow);
    if (!mounted) {
      return;
    }
    if (result.hasMedia) {
      _applyRecoveredPics(result.mediaBase64List);
      _recordPermissionGranted(flow);
      return;
    }
    if (result.errorMessage.isNotEmpty) {
      _showToast('图片选择失败');
    }
  }

  void _applyRecoveredPics(List<String> pics) {
    if (pics.isEmpty) {
      return;
    }
    setState(() {
      final merged = <String>[..._proofPics, ...pics];
      _proofPics = merged.take(3).toList();
    });
    _persistDraft();
  }

  Future<void> _handlePermissionPromptAction(
    PermissionFlowContext flow,
    PermissionPromptAction action, {
    required PermissionFlowStatus currentStatus,
  }) async {
    switch (action) {
      case PermissionPromptAction.request:
        final result = await widget.permissionBroker.request(flow);
        if (result.isUsable) {
          _recordPermissionGranted(flow);
          await _pickImageWithFlow(
            flow.copyWith(status: result.status),
          );
          return;
        }
        _recordPermissionDenied(flow);
        final followUpAction = await _showPermissionPrompt(
          flow,
          result.status,
          alternativeLabel: _alternativeLabel(flow.source),
        );
        if (!mounted || followUpAction == null) {
          return;
        }
        await _handlePermissionPromptAction(
          flow,
          followUpAction,
          currentStatus: result.status,
        );
        return;
      case PermissionPromptAction.openSettings:
        _recordPermissionSettingsRedirected(flow);
        await widget.permissionBroker.openSettingsForFlow(
          flow.copyWith(status: currentStatus),
        );
        return;
      case PermissionPromptAction.alternative:
        _recordPermissionFallbackUsed(flow);
        final alternative = flow.source == PermissionFlowSource.camera
            ? PermissionFlowSource.gallery
            : PermissionFlowSource.camera;
        await _handlePickImage(alternative);
        if (mounted) {
          _showToast('已保留当前售后内容，你仍可继续无图提交');
        }
        return;
      case PermissionPromptAction.dismiss:
        _showToast('已保留当前售后内容，你仍可继续无图提交');
        return;
    }
  }

  Future<PermissionPromptAction?> _showPermissionPrompt(
    PermissionFlowContext flow,
    PermissionFlowStatus status, {
    required String alternativeLabel,
  }) {
    _recordPermissionPromptShown(flow);
    return PermissionPromptSheet.show(
      context,
      flow: flow,
      status: status,
      alternativeLabel: alternativeLabel,
    );
  }

  String _alternativeLabel(PermissionFlowSource source) {
    if (source == PermissionFlowSource.camera) {
      return '改用相册';
    }
    return '改用拍照';
  }

  void _recordPermissionPromptShown(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionPromptShown(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionGranted(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionGranted(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionDenied(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionDenied(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionSettingsRedirected(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionSettingsRedirected(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionFallbackUsed(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionFallbackUsed(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionLostDataRecovered(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionLostDataRecovered(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  // ==================== 提交售后申请 ====================
  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedReason == null) {
      _showToast('请选择退货原因');
      return;
    }
    if (_isSubmitting) return;

    setState(() {
      _isSubmitting = true;
      _pageState = SalesPageState.submitting;
    });

    try {
      final req = ApplyAfterSalesReqData(
        orderId: widget.orderId,
        type: _selectedType.value,
        reasonId: _selectedReason!.id,
        description: _descController.text.trim(),
        proofPics: _proofPics.join(','),
      );
      final result = await _submitAfterSales(req);

      if (!mounted) return;

      if (result.isSuccess) {
        // 幂等返回（已有待审核售后单）：提示用户，不走完整成功流程
        if (result.message == "已存在待审核的售后单") {
          _showToast('已有待审核售后单：${result.returnNo}');
          setState(() {
            _isSubmitting = false;
          });
          return;
        }
        // 新提交成功
        await AppRecoveryStore.clearAfterSalesDraft(widget.orderId);
        await AppRecoveryStore.clearPendingPermissionContext();
        await AppRecoveryStore.clearPendingLostMedia();
        await AppRecoveryStore.clearActiveIntentCandidateIfMatches(
          AppRecentTargetType.afterSalesApply,
          targetId: widget.orderId,
        );
        await AppRecoveryStore.clearRecentContextIfMatches(
          AppRecentTargetType.afterSalesApply,
          targetId: widget.orderId,
        );
        setState(() {
          _pageState = SalesPageState.success;
          _successReturnNo = result.returnNo;
          _isSubmitting = false;
        });
      } else {
        setState(() {
          _pageState = SalesPageState.error;
          _errorMessage =
              result.message.isNotEmpty ? result.message : '提交失败，请稍后重试';
          _isSubmitting = false;
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _pageState = SalesPageState.error;
        _errorMessage = '网络异常，请检查网络后重试';
        _isSubmitting = false;
      });
    }
  }

  void _showToast(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), duration: const Duration(seconds: 2)),
    );
  }

  // ==================== 构建 ====================
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('申请售后'),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    switch (_pageState) {
      case SalesPageState.loading:
        return _buildLoadingState();
      case SalesPageState.ready:
        return _buildFormState();
      case SalesPageState.submitting:
        return _buildSubmittingState();
      case SalesPageState.success:
        return _buildSuccessState();
      case SalesPageState.error:
        return _buildErrorState();
    }
  }

  // ==================== 各状态 UI ====================

  Widget _buildLoadingState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SpinKitCircle(color: Color(0xFFFA436A), size: 40),
          const SizedBox(height: 16),
          Text(
            '加载中...',
            style: TextStyle(color: Colors.grey[600], fontSize: 14),
          ),
        ],
      ),
    );
  }

  Widget _buildFormState() {
    final border = BorderSide(width: 1, color: const Color(0xFFF0F0F0));
    final labelStyle = const TextStyle(fontSize: 14, color: Color(0xFF303133));
    final hintStyle = TextStyle(fontSize: 14, color: Colors.grey[400]);

    return Stack(
      children: [
        Form(
          key: _formKey,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // ---- 售后类型选择 ----
              _buildSectionTitle('售后类型'),
              const SizedBox(height: 8),
              Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: border.color),
                ),
                child: RadioGroup<AfterSalesType>(
                  groupValue: _selectedType,
                  onChanged: (value) {
                    if (value != null) {
                      setState(() => _selectedType = value);
                      _persistDraft();
                    }
                  },
                  child: Column(
                    children: AfterSalesType.values.map((type) {
                      return ListTile(
                        dense: true,
                        contentPadding:
                            const EdgeInsets.symmetric(horizontal: 8),
                        title: Text(type.label, style: labelStyle),
                        leading: Radio<AfterSalesType>(
                          value: type,
                          activeColor: const Color(0xFFFA436A),
                        ),
                        onTap: () {
                          setState(() => _selectedType = type);
                          _persistDraft();
                        },
                      );
                    }).toList(),
                  ),
                ),
              ),

              const SizedBox(height: 20),

              // ---- 退货原因选择 ----
              _buildSectionTitle('退货原因 *'),
              const SizedBox(height: 8),
              Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: border.color),
                ),
                child: DropdownButtonFormField<ReturnReasonItemData>(
                  initialValue: _selectedReason,
                  hint: Text('请选择退货原因', style: hintStyle),
                  decoration: const InputDecoration(
                    contentPadding:
                        EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                    border: InputBorder.none,
                  ),
                  items: _reasonList.map((reason) {
                    return DropdownMenuItem<ReturnReasonItemData>(
                      value: reason,
                      child: Text(reason.name,
                          style: labelStyle, overflow: TextOverflow.ellipsis),
                    );
                  }).toList(),
                  validator: (v) => v == null ? '请选择退货原因' : null,
                  onChanged: (v) {
                    setState(() => _selectedReason = v);
                    _persistDraft();
                  },
                ),
              ),

              const SizedBox(height: 20),

              // ---- 问题描述 ----
              _buildSectionTitle('问题描述'),
              const SizedBox(height: 8),
              Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: border.color),
                ),
                child: TextFormField(
                  controller: _descController,
                  maxLines: 4,
                  maxLength: 500,
                  decoration: InputDecoration(
                    hintText: '请描述您遇到的问题（可选）',
                    hintStyle: hintStyle,
                    contentPadding: const EdgeInsets.all(12),
                    border: InputBorder.none,
                    counterStyle:
                        TextStyle(color: Colors.grey[400], fontSize: 12),
                  ),
                ),
              ),

              const SizedBox(height: 20),

              // ---- 凭证图片 ----
              _buildSectionTitle('上传凭证（可选）'),
              const SizedBox(height: 4),
              Text(
                '最多 3 张，支持相册和拍照',
                style: TextStyle(fontSize: 12, color: Colors.grey[500]),
              ),
              const SizedBox(height: 12),
              _buildProofPicsGrid(),

              const SizedBox(height: 32),

              // ---- 提交按钮 ----
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFA436A),
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    elevation: 0,
                  ),
                  onPressed: _submit,
                  child: const Text(
                    '提交申请',
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
                  ),
                ),
              ),

              const SizedBox(height: 16),

              // ---- 提示 ----
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.orange[50],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(Icons.info_outline,
                        size: 16, color: Colors.orange[700]),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        '提交后预计 1-3 个工作日处理，请耐心等待',
                        style:
                            TextStyle(fontSize: 12, color: Colors.orange[800]),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: const TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        color: Color(0xFF303133),
      ),
    );
  }

  Widget _buildProofPicsGrid() {
    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: [
        ..._proofPics.asMap().entries.map((entry) {
          return _buildPicItem(
            Image.memory(
              base64Decode(entry.value),
              fit: BoxFit.cover,
              width: 80,
              height: 80,
              errorBuilder: (_, __, ___) => Container(
                width: 80,
                height: 80,
                color: Colors.grey[200],
                child: const Icon(Icons.broken_image, color: Colors.grey),
              ),
            ),
            onRemove: () => _removePic(entry.key),
          );
        }),
        if (_proofPics.length < 3) _buildAddPicButton(),
      ],
    );
  }

  Widget _buildPicItem(Widget image, {required VoidCallback onRemove}) {
    return Stack(
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: SizedBox(width: 80, height: 80, child: image),
        ),
        Positioned(
          top: 2,
          right: 2,
          child: GestureDetector(
            onTap: onRemove,
            child: Container(
              padding: const EdgeInsets.all(2),
              decoration: const BoxDecoration(
                color: Colors.black54,
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.close, size: 14, color: Colors.white),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildAddPicButton() {
    return GestureDetector(
      onTap: _showPickImageSheet,
      child: Container(
        width: 80,
        height: 80,
        decoration: BoxDecoration(
          color: Colors.grey[100],
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey[300]!),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.add_a_photo, size: 24, color: Colors.grey[500]),
            const SizedBox(height: 4),
            Text(
              '${_proofPics.length}/3',
              style: TextStyle(fontSize: 11, color: Colors.grey[500]),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSubmittingState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SpinKitCircle(color: Color(0xFFFA436A), size: 40),
          const SizedBox(height: 16),
          const Text(
            '提交中...',
            style: TextStyle(fontSize: 16, color: Color(0xFF606266)),
          ),
          const SizedBox(height: 8),
          Text(
            '请勿关闭页面',
            style: TextStyle(fontSize: 13, color: Colors.grey[500]),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.error_outline, size: 60, color: Colors.red[300]),
            const SizedBox(height: 16),
            Text(
              _errorMessage ?? '提交失败',
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 15, color: Color(0xFF606266)),
            ),
            const SizedBox(height: 24),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                OutlinedButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('返回订单'),
                ),
                const SizedBox(width: 16),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFA436A),
                    foregroundColor: Colors.white,
                  ),
                  onPressed: () {
                    setState(() {
                      _pageState = SalesPageState.ready;
                      _errorMessage = null;
                    });
                  },
                  child: const Text('重新提交'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSuccessState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.green[50],
                shape: BoxShape.circle,
              ),
              child:
                  Icon(Icons.check_circle, size: 60, color: Colors.green[400]),
            ),
            const SizedBox(height: 24),
            const Text(
              '售后申请已提交',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.w600,
                color: Color(0xFF303133),
              ),
            ),
            const SizedBox(height: 12),
            if (_successReturnNo != null) ...[
              Text(
                '售后单号：$_successReturnNo',
                style: TextStyle(fontSize: 14, color: Colors.grey[600]),
              ),
              const SizedBox(height: 8),
            ],
            Text(
              '预计 1-3 个工作日处理，请耐心等待',
              style: TextStyle(fontSize: 13, color: Colors.grey[500]),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            SizedBox(
              width: double.infinity,
              height: 44,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFFA436A),
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  elevation: 0,
                ),
                onPressed: () => Navigator.pop(context),
                child: const Text('返回订单'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
