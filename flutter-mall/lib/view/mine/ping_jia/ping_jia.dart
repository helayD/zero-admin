import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/comment_model.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/provider/comment_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';
import 'package:flutter_mall/widgets/permission_prompt_sheet.dart';
import 'package:flutter_spinkit/flutter_spinkit.dart';
import 'package:provider/provider.dart';

typedef CommentListLoader = Future<CommentListModel> Function({
  required int productId,
  required int page,
  required int pageSize,
});

///
/// 商品评价页面（Story 8-2 重构）
///
/// 支持两种入口：
/// - 入口1：从"我的订单" → 点击"去评价" → 评价提交页
/// - 入口2：从商品详情页 → 查看评价列表（只读）
///
/// 作者：David
/// 日期：2026/04/02
///
class PinJia extends StatefulWidget {
  /// 订单ID（入口1必填）
  final int? orderId;

  /// 商品ID（两个入口都必填）
  /// Design Decision: 此字段实际传入的是 skuId（订单商品维度）。
  /// 评价系统按 SKU 粒度设计——用户对同一商品的每个 SKU 可独立评价一次。
  /// 若未来改为 SPU 粒度，需在 order_detail 传入真实 productId，并修改后端去重校验逻辑。
  final int? productId;

  /// 商品名称（入口1必填）
  final String? productName;

  /// 商品图片（入口1必填）
  final String? productPic;

  /// 购买时的商品属性（Review Fix H-6: 从订单商品 specData 传递）
  final String? productAttribute;

  /// 评价者昵称（Review Fix H-5: 从收货人姓名获取，非硬编码）
  final String? memberNickName;

  final PermissionBroker permissionBroker;

  final CommentListLoader? commentListLoader;

  const PinJia({
    super.key,
    this.orderId,
    this.productId,
    this.productName,
    this.productPic,
    this.productAttribute,
    this.memberNickName,
    this.permissionBroker = const PermissionBroker(),
    this.commentListLoader,
  });

  @override
  State<PinJia> createState() => _PinJiaState();
}

class _PinJiaState extends State<PinJia> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  // 评价提交相关状态
  final _contentController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  int _starRating = 5;
  bool _starValidated = true;
  List<String> _selectedPics = [];
  bool _isSubmitting = false;
  AppLifecycleProvider? _lifecycleProvider;
  int _lastResumeTick = 0;

  // 评价列表相关状态
  bool _isLoadingComments = true;
  List<CommentItem> _commentList = [];
  int _commentTotal = 0;
  bool _hasMore = true;
  int _currentPage = 1;
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _tabController.addListener(_onTabChanged);
    _contentController.addListener(_handleDraftChanged);
    _restoreDraft();
    _loadCommentList();
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      await _saveRecentContext();
      await _persistDraft();
      await _resumePendingPermissionFlow();
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
    _tabController.removeListener(_onTabChanged);
    _tabController.dispose();
    _contentController.removeListener(_handleDraftChanged);
    _contentController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _onTabChanged() {
    if (!_tabController.indexIsChanging) {
      setState(() {});
    }
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

  bool get _hasRecoverableComposeContext {
    final orderId = widget.orderId;
    final productId = widget.productId;
    return orderId != null && orderId > 0 && productId != null && productId > 0;
  }

  Future<void> _saveRecentContext() async {
    if (!_hasRecoverableComposeContext) {
      return;
    }
    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.commentCompose,
        targetId: widget.orderId,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderDetail,
        fallbackTargetId: widget.orderId,
      ),
    );
  }

  void _restoreDraft() {
    final orderId = widget.orderId;
    if (orderId == null || orderId <= 0) {
      return;
    }
    final draft = AppRecoveryStore.getCommentDraft(orderId);
    if (draft == null) {
      return;
    }
    _starRating = draft.starRating;
    _selectedPics = List<String>.from(draft.pics.take(3));
    _contentController.text = draft.content;
  }

  Future<void> _persistDraft() async {
    if (!_hasRecoverableComposeContext) {
      return;
    }
    await AppRecoveryStore.saveCommentDraft(
      CommentDraftSnapshot(
        orderId: widget.orderId!,
        productId: widget.productId!,
        productName: widget.productName ?? '',
        productPic: widget.productPic ?? '',
        productAttribute: widget.productAttribute ?? '',
        memberNickName: widget.memberNickName ?? '',
        starRating: _starRating,
        content: _contentController.text,
        pics: _selectedPics,
        updatedAt: DateTime.now(),
      ),
    );
  }

  Future<void> _resumePendingPermissionFlow() async {
    if (!_hasRecoverableComposeContext) {
      return;
    }
    final pending = AppRecoveryStore.peekPendingPermissionContext();
    final targetId = widget.orderId;
    if (pending == null ||
        targetId == null ||
        !pending.matchesReturnTarget(
          AppRecentTargetType.commentCompose,
          targetId: targetId,
        )) {
      return;
    }

    final lostMedia = AppRecoveryStore.peekPendingLostMedia();
    if (lostMedia != null &&
        lostMedia.recoveryId == pending.recoveryId &&
        lostMedia.scene == PermissionScene.commentImage) {
      await AppRecoveryStore.clearPendingLostMedia();
      await AppRecoveryStore.clearPendingPermissionContext();
      _applyRecoveredPics(lostMedia.mediaBase64List);
      _recordPermissionLostDataRecovered(pending);
      if (mounted) {
        _showToast('已恢复上次未完成的补图内容');
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
      _showToast('权限仍未开启，你仍可继续无图评价');
    }
  }

  // ==================== 评价列表加载 ====================

  Future<void> _loadCommentList({
    bool refresh = false,
    bool loadMore = false,
  }) async {
    if (widget.productId == null) return;

    final page = refresh ? 1 : _currentPage;

    if (!refresh && !loadMore) {
      setState(() => _isLoadingComments = true);
    }

    try {
      final pageSize = 20;
      final model = widget.commentListLoader != null
          ? await widget.commentListLoader!(
              productId: widget.productId!,
              page: page,
              pageSize: pageSize,
            )
          : await _loadCommentListFromApi(
              productId: widget.productId!,
              page: page,
              pageSize: pageSize,
            );

      if (!mounted) return;

      // 维持页面与 provider 缓存一致，避免重复追加同一批评论。
      final commentProvider = context.read<CommentProvider>();
      commentProvider.syncCommentPage(
        widget.productId!,
        data: model.data,
        total: model.total,
        page: page,
        append: loadMore,
      );

      setState(() {
        if (loadMore) {
          _commentList = [..._commentList, ...model.data];
        } else if (refresh) {
          _commentList = model.data;
        } else {
          _commentList = model.data;
        }
        _currentPage = page;
        _hasMore = model.data.length < model.total;
        _commentTotal = model.total;
        _isLoadingComments = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() => _isLoadingComments = false);
    }
  }

  Future<CommentListModel> _loadCommentListFromApi({
    required int productId,
    required int page,
    required int pageSize,
  }) async {
    final result = await HttpUtil.get(
      queryCommentListUrl,
      queryParameters: {
        "productId": productId,
        "pageNum": page,
        "pageSize": pageSize,
      },
    );
    return CommentListModel.fromJson(result.data);
  }

  void _onRefresh() {
    _currentPage = 1;
    _loadCommentList(refresh: true);
  }

  void _onLoadMore() {
    if (!_hasMore || _isLoadingComments) return;
    _currentPage++;
    _loadCommentList(loadMore: true);
  }

  // ==================== 图片上传 ====================

  Future<void> _handlePickImage(PermissionFlowSource source) async {
    if (_selectedPics.length >= 3) {
      _showToast('最多上传 3 张图片');
      return;
    }
    if (!_hasRecoverableComposeContext) {
      _showToast('当前评价上下文不完整，请从订单页重新进入后再补图');
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
      _selectedPics = List.from(_selectedPics)..removeAt(index);
    });
    _persistDraft();
  }

  PermissionFlowContext _buildPermissionFlow(PermissionFlowSource source) {
    return PermissionFlowContext.create(
      scene: PermissionScene.commentImage,
      permissionType: source == PermissionFlowSource.camera
          ? PermissionType.camera
          : PermissionType.photos,
      source: source,
      returnTarget: AppRecentTargetType.commentCompose,
      returnTargetId: widget.orderId,
      orderId: widget.orderId,
      productId: widget.productId,
      fallbackAction: source == PermissionFlowSource.camera
          ? PermissionFallbackAction.switchToGallery
          : PermissionFallbackAction.switchToCamera,
      intentId:
          'comment_${permissionFlowSourceToValue(source)}_${widget.orderId ?? 0}',
      recoveryId:
          'comment_${permissionFlowSourceToValue(source)}_${widget.orderId ?? 0}',
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
      final merged = <String>[..._selectedPics, ...pics];
      _selectedPics = merged.take(3).toList();
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
          _showToast('你仍可继续无图评价');
        }
        return;
      case PermissionPromptAction.dismiss:
        _showToast('已保留当前评价内容，你仍可继续无图评价');
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

  // ==================== 提交评价 ====================

  Future<void> _submitComment() async {
    // Review Fix M-NEW-2: 星评非零校验
    if (_starRating == 0) {
      setState(() => _starValidated = false);
      _showToast('请选择商品评分');
      return;
    }
    if (!_formKey.currentState!.validate()) return;
    if (_isSubmitting) return;

    setState(() => _isSubmitting = true);

    try {
      final req = AddCommentReqData(
        productId: widget.productId ?? 0,
        orderId: widget.orderId ?? 0,
        star: _starRating,
        content: _contentController.text.trim(),
        pics: _selectedPics.join(','),
        // Review Fix H-5: 从 widget.memberNickName 获取（由 order_detail 传入收货人姓名）
        memberNickName: widget.memberNickName ?? '匿名用户',
        productName: widget.productName ?? '',
        // Review Fix H-6: 从 widget.productAttribute 获取（由 order_detail 传入 specData）
        productAttribute: widget.productAttribute ?? '',
      );

      final result = await HttpUtil.post(addCommentUrl, data: req.toJson());
      final resp = AddCommentRespData.fromJson(result.data);

      if (!mounted) return;

      if (resp.isSuccess) {
        // Review Fix R-5: 提交成功后切换到评价列表页，让用户确认提交内容
        if ((widget.orderId ?? 0) > 0) {
          await AppRecoveryStore.clearCommentDraft(widget.orderId!);
        }
        await AppRecoveryStore.clearPendingPermissionContext();
        await AppRecoveryStore.clearPendingLostMedia();
        _showToast('评价提交成功，审核通过后可查看');
        setState(() => _isSubmitting = false);
        // 切换到评价列表 Tab，让用户看到提交后的状态
        if (_tabController.index != 1) {
          _tabController.animateTo(1);
        }
        // 2秒后自动返回订单详情页
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) {
            Navigator.pop(context);
          }
        });
      } else {
        _showToast(resp.message.isNotEmpty ? resp.message : '提交失败');
        setState(() => _isSubmitting = false);
      }
    } catch (e) {
      if (!mounted) return;
      _showToast('网络异常，请检查网络后重试');
      setState(() => _isSubmitting = false);
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
        title: const Text("商品评价"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: const Color(0xFFFA436A),
          labelColor: const Color(0xFFFA436A),
          unselectedLabelColor: const Color(0xFF606266),
          tabs: const [
            Tab(text: '评价提交'),
            Tab(text: '评价列表'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [_buildSubmitTab(), _buildCommentListTab()],
      ),
    );
  }

  // ==================== 评价提交 Tab ====================

  Widget _buildSubmitTab() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Form(
        key: _formKey,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 商品信息卡片
            _buildProductCard(),
            const SizedBox(height: 20),

            // 评分
            _buildSectionTitle('商品评分 *'),
            const SizedBox(height: 8),
            _buildStarRating(),
            const SizedBox(height: 20),

            // 评价内容
            _buildSectionTitle('评价内容 *'),
            const SizedBox(height: 8),
            _buildContentInput(),
            const SizedBox(height: 20),

            // 图片上传
            _buildSectionTitle('上传图片（可选）'),
            const SizedBox(height: 4),
            Text(
              '最多 3 张，支持相册和拍照',
              style: TextStyle(fontSize: 12, color: Colors.grey[500]),
            ),
            const SizedBox(height: 12),
            _buildPicGrid(),
            const SizedBox(height: 32),

            // 提交按钮
            _buildSubmitButton(),
          ],
        ),
      ),
    );
  }

  Widget _buildProductCard() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: Row(
        children: [
          if (widget.productPic != null && widget.productPic!.isNotEmpty)
            CachedImageWidget(60, 60, widget.productPic!)
          else
            Container(
              width: 60,
              height: 60,
              color: Colors.grey[200],
              child: const Icon(Icons.image, color: Colors.grey),
            ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              widget.productName ?? '商品评价',
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 14, color: Color(0xFF303133)),
            ),
          ),
        ],
      ),
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

  Widget _buildStarRating() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.grey[200]!),
          ),
          child: Row(
            children: List.generate(5, (index) {
              final star = index + 1;
              return GestureDetector(
                onTap: () {
                  setState(() {
                    _starRating = star;
                    _starValidated = true;
                  });
                  _persistDraft();
                },
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Icon(
                    star <= _starRating ? Icons.star : Icons.star_border,
                    color: const Color(0xFFFF9900),
                    size: 32,
                  ),
                ),
              );
            }),
          ),
        ),
        // Review Fix M-NEW-2: 未选评分时显示提示
        if (!_starValidated)
          Padding(
            padding: const EdgeInsets.only(top: 4, left: 4),
            child: Text(
              '请选择评分',
              style: TextStyle(fontSize: 12, color: Colors.red[400]),
            ),
          ),
      ],
    );
  }

  Widget _buildContentInput() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: TextFormField(
        controller: _contentController,
        maxLines: 5,
        maxLength: 500,
        decoration: InputDecoration(
          hintText: '请输入您的评价（至少10个字）',
          hintStyle: TextStyle(fontSize: 14, color: Colors.grey[400]),
          contentPadding: const EdgeInsets.all(12),
          border: InputBorder.none,
          counterStyle: TextStyle(color: Colors.grey[400], fontSize: 12),
        ),
        validator: (value) {
          if (value == null || value.trim().isEmpty) {
            return '请输入评价内容';
          }
          if (value.trim().length < 10) {
            return '评价内容至少10个字';
          }
          return null;
        },
      ),
    );
  }

  Widget _buildPicGrid() {
    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: [
        ..._selectedPics.asMap().entries.map((entry) {
          return Stack(
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Image.memory(
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
              ),
              Positioned(
                top: 2,
                right: 2,
                child: GestureDetector(
                  onTap: () => _removePic(entry.key),
                  child: Container(
                    padding: const EdgeInsets.all(2),
                    decoration: const BoxDecoration(
                      color: Colors.black54,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.close,
                      size: 14,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ],
          );
        }),
        if (_selectedPics.length < 3) _buildAddPicButton(),
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
              '${_selectedPics.length}/3',
              style: TextStyle(fontSize: 11, color: Colors.grey[500]),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSubmitButton() {
    return SizedBox(
      width: double.infinity,
      height: 48,
      child: ElevatedButton(
        style: ElevatedButton.styleFrom(
          backgroundColor: const Color(0xFFFA436A),
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          elevation: 0,
        ),
        onPressed: _isSubmitting ? null : _submitComment,
        child: _isSubmitting
            ? const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                ),
              )
            : const Text(
                '提交评价',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
              ),
      ),
    );
  }

  // ==================== 评价列表 Tab ====================

  Widget _buildCommentListTab() {
    if (_isLoadingComments) {
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

    if (_commentList.isEmpty) {
      return EmptyStateWidget(
        message: "暂无评价",
        icon: Icons.rate_review_outlined,
        actionText: "去评价",
        onAction: () {
          _tabController.animateTo(0);
        },
      );
    }

    return Column(
      children: [
        // 评价统计
        Container(
          padding: const EdgeInsets.all(12),
          color: Colors.white,
          child: Row(
            children: [
              const Text(
                '全部评价',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: Color(0xFF303133),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                '$_commentTotal 条',
                style: const TextStyle(fontSize: 14, color: Color(0xFFFA436A)),
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        // 评价列表
        Expanded(
          child: RefreshIndicator(
            onRefresh: () async => _onRefresh(),
            color: const Color(0xFFFA436A),
            child: NotificationListener<ScrollNotification>(
              onNotification: (notification) {
                if (notification is ScrollEndNotification) {
                  final metrics = notification.metrics;
                  if (metrics.pixels >= metrics.maxScrollExtent - 100) {
                    _onLoadMore();
                  }
                }
                return false;
              },
              child: ListView.separated(
                controller: _scrollController,
                padding: const EdgeInsets.all(12),
                itemCount: _commentList.length + (_hasMore ? 1 : 0),
                separatorBuilder: (_, __) => const SizedBox(height: 12),
                itemBuilder: (context, index) {
                  if (index >= _commentList.length) {
                    return const Center(
                      child: Padding(
                        padding: EdgeInsets.all(16),
                        child: Text(
                          "加载中...",
                          style: TextStyle(color: Colors.grey, fontSize: 13),
                        ),
                      ),
                    );
                  }
                  return _CommentCard(comment: _commentList[index]);
                },
              ),
            ),
          ),
        ),
      ],
    );
  }
}

/// 评价卡片组件
class _CommentCard extends StatelessWidget {
  final CommentItem comment;

  const _CommentCard({required this.comment});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 用户信息 + 评分
          Row(
            children: [
              // 头像
              CircleAvatar(
                radius: 16,
                backgroundColor: Colors.grey[200],
                backgroundImage: comment.memberIcon.isNotEmpty
                    ? NetworkImage(comment.memberIcon) as ImageProvider
                    : null,
                child: comment.memberIcon.isEmpty
                    ? const Icon(Icons.person, size: 16, color: Colors.grey)
                    : null,
              ),
              const SizedBox(width: 8),
              // 昵称
              Expanded(
                child: Text(
                  comment.memberNickName.isNotEmpty
                      ? comment.memberNickName
                      : '匿名用户',
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    color: Color(0xFF303133),
                  ),
                ),
              ),
              // 星级
              Row(
                children: List.generate(5, (index) {
                  return Icon(
                    index < comment.star ? Icons.star : Icons.star_border,
                    color: const Color(0xFFFF9900),
                    size: 16,
                  );
                }),
              ),
            ],
          ),
          const SizedBox(height: 8),

          // 评价内容
          Text(
            comment.content,
            style: const TextStyle(
              fontSize: 14,
              color: Color(0xFF606266),
              height: 1.5,
            ),
          ),
          const SizedBox(height: 8),

          // 评价图片
          if (comment.picList.isNotEmpty) ...[
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: comment.picList.map((picUrl) {
                return ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: CachedImageWidget(60, 60, picUrl),
                );
              }).toList(),
            ),
            const SizedBox(height: 8),
          ],

          // 商品属性
          if (comment.productAttribute.isNotEmpty) ...[
            Text(
              comment.productAttribute,
              style: TextStyle(fontSize: 12, color: Colors.grey[500]),
            ),
            const SizedBox(height: 4),
          ],

          // 评价时间
          Text(
            _formatTime(comment.createTime),
            style: TextStyle(fontSize: 12, color: Colors.grey[400]),
          ),

          // 回复数
          if (comment.replayCount > 0) ...[
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(
                  Icons.chat_bubble_outline,
                  size: 14,
                  color: Colors.grey[500],
                ),
                const SizedBox(width: 4),
                Text(
                  '${comment.replayCount} 条回复',
                  style: TextStyle(fontSize: 12, color: Colors.grey[500]),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  String _formatTime(String isoTime) {
    if (isoTime.isEmpty) return "";
    try {
      final parts = isoTime.split("T");
      if (parts.length >= 2) {
        final datePart = parts[0];
        final timePart = parts[1].substring(0, 5);
        return "$datePart $timePart";
      }
      return isoTime;
    } catch (_) {
      return isoTime;
    }
  }
}
