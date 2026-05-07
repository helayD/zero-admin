import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/layout/upgrade_gate_page.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/confirm_order.dart';
import 'package:flutter_mall/model/direct_checkout.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/provider/cart_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:provider/provider.dart';
import 'package:uuid/uuid.dart';

import 'address_select_sheet.dart';
import 'coupon_select_sheet.dart';
import 'order_pay.dart';

///
/// 提交订单页面（Story 5-3 完整实现）
///
/// 作者：LiuFeiHua
/// 日期：2023/11/21 17:17
/// Story 5-3: Task 5-11 实现
///
class OrderSubmit extends StatefulWidget {
  final DirectCheckoutParams? directItem;

  const OrderSubmit({super.key, this.directItem});

  @override
  State<OrderSubmit> createState() => _OrderSubmitState();
}

class _OrderSubmitState extends State<OrderSubmit> {
  ConfirmOrderData? _orderData;
  bool _loading = true;
  String? _errorMessage;

  // Task 5: 地址选择
  int _selectedAddressIndex = 0;

  // Task 6: 优惠券选择
  ConfirmCouponData? _selectedCoupon;

  // Task 7: 积分抵扣
  final TextEditingController _integrationController = TextEditingController();
  bool _integrationConflict = false; // 积分与优惠券互斥
  final TextEditingController _remarkController =
      TextEditingController(); // 订单备注（MEDIUM-3）

  // Task 8: 支付方式（1=支付宝，2=微信，默认微信）
  int _selectedPayType = 2;

  // Task 9: 提交订单防重复
  bool _isSubmitting = false;

  // Price Breakdown Card（Task 11）
  int _previewIntegrationAmount = 0;

  String _formatAmount(num amount) {
    final normalized = amount.toDouble();
    if (normalized == normalized.truncateToDouble()) {
      return normalized.toStringAsFixed(0);
    }
    return normalized.toStringAsFixed(2);
  }

  @override
  void initState() {
    super.initState();
    _loadConfirmOrder();
  }

  bool get _isDirectBuy => widget.directItem != null;

  @override
  void dispose() {
    _integrationController.dispose();
    _remarkController.dispose();
    super.dispose();
  }

  /// 加载确认单数据
  void _loadConfirmOrder() async {
    final ready = await _ensureUpgradeReady();
    if (!ready || !mounted) {
      return;
    }
    try {
      setState(() {
        _loading = true;
        _errorMessage = null;
      });
      final cartModel = context.read<CartModel>();
      final ids = cartModel.getCheckProduct().map((e) => e.id).toList();
      final Map<String, dynamic> requestData = _isDirectBuy
          ? {"directItem": widget.directItem!.toJson()}
          : {"ids": ids};
      Response result =
          await HttpUtil.post(generateConfirmOrderUrl, data: requestData);
      ConfirmOrderModel model = ConfirmOrderModel.fromJson(result.data);
      if (mounted) {
        setState(() {
          _orderData = model.data;
          _loading = false;
          // 自动选择默认地址
          _selectedAddressIndex = 0;
          for (int i = 0; i < model.data.memberReceiveAddressList.length; i++) {
            if (model.data.memberReceiveAddressList[i].isDefault == 1) {
              _selectedAddressIndex = i;
              break;
            }
          }
          // Task 7: 初始化积分抵扣，默认不使用，避免用户未主动选择时隐式抵扣。
          _initIntegrationInput();
        });
      }
    } catch (e) {
      debugPrint("加载确认单失败: $e");
      if (mounted) {
        setState(() {
          _loading = false;
          _errorMessage = "加载确认单失败，请重试";
        });
      }
    }
  }

  AppRecentContext _buildUpgradeFallbackContext() {
    if (_isDirectBuy && widget.directItem != null) {
      return AppRecentContext.create(
        targetType: AppRecentTargetType.productDetail,
        targetId: widget.directItem!.productId,
        source: 'order_confirm',
        requiresAuth: false,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
      );
    }
    return AppRecentContext.create(
      targetType: AppRecentTargetType.cart,
      tabIndex: 2,
      source: 'order_confirm',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );
  }

  Future<bool> _ensureUpgradeReady() async {
    final policy = await UpgradeGateService.queryPolicy(
      scene: 'order_confirm',
      targetType: _isDirectBuy
          ? appRecentTargetTypeToValue(AppRecentTargetType.productDetail)
          : appRecentTargetTypeToValue(AppRecentTargetType.cart),
      targetId: _isDirectBuy ? widget.directItem?.productId : null,
    );
    if (!policy.hasUpgradeGate) {
      return true;
    }
    if (!mounted) {
      return true;
    }
    final navigator = Navigator.of(context);
    final pendingUpgrade = PendingUpgradeContext.forOrderConfirm(
      policy: policy,
      directItem: widget.directItem,
      fallbackContext: _buildUpgradeFallbackContext(),
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

  /// Task 7: 初始化积分输入
  void _initIntegrationInput() {
    final data = _orderData;
    if (data == null) return;
    final integration = data.memberIntegration;
    final setting = data.integrationConsumeSetting;
    _integrationController.text = "0";
    _previewIntegrationAmount = 0;
    if (integration <= 0 || setting.deductionPerAmount <= 0) return;
    _checkIntegrationConflict();
  }

  int _calcIntegrationAmount(int useIntegration) {
    final data = _orderData;
    if (data == null ||
        data.integrationConsumeSetting.deductionPerAmount <= 0) {
      return 0;
    }
    return useIntegration ~/ data.integrationConsumeSetting.deductionPerAmount;
  }

  void _onIntegrationChanged(String val) {
    final data = _orderData;
    if (data == null) return;
    final v = int.tryParse(val) ?? 0;
    final integration = data.memberIntegration;
    final maxByPercent = data.integrationConsumeSetting.maxPercentPerOrder > 0
        ? (data.calcAmount.totalAmount *
                data.integrationConsumeSetting.maxPercentPerOrder ~/
                100) *
            data.integrationConsumeSetting.deductionPerAmount
        : integration;
    final maxIntegration =
        integration < maxByPercent ? integration : maxByPercent;
    if (v > maxIntegration) {
      _integrationController.text = maxIntegration.toString();
      _integrationController.selection = TextSelection.fromPosition(
        TextPosition(offset: _integrationController.text.length),
      );
    }
    setState(() {
      _previewIntegrationAmount = _calcIntegrationAmount(
          int.tryParse(_integrationController.text) ?? 0);
    });
    _checkIntegrationConflict();
  }

  void _onStepperPressed(int delta) {
    final data = _orderData;
    if (data == null) return;
    final cur = int.tryParse(_integrationController.text) ?? 0;
    final step = data.integrationConsumeSetting.useUnit > 0
        ? data.integrationConsumeSetting.useUnit
        : 100;
    final integration = data.memberIntegration;
    final maxByPercent = data.integrationConsumeSetting.maxPercentPerOrder > 0
        ? (data.calcAmount.totalAmount *
                data.integrationConsumeSetting.maxPercentPerOrder ~/
                100) *
            data.integrationConsumeSetting.deductionPerAmount
        : integration;
    final maxIntegration =
        integration < maxByPercent ? integration : maxByPercent;
    var next = (cur + delta * step).clamp(0, maxIntegration);
    next = (next ~/ step) * step;
    _integrationController.text = next.toString();
    _integrationController.selection = TextSelection.fromPosition(
      TextPosition(offset: _integrationController.text.length),
    );
    setState(() {
      _previewIntegrationAmount = _calcIntegrationAmount(next);
    });
    _checkIntegrationConflict();
  }

  void _checkIntegrationConflict() {
    final data = _orderData;
    if (data == null) return;
    final couponStatus = data.integrationConsumeSetting.couponStatus;
    final hasCoupon = _selectedCoupon != null;
    final useIntegration = int.tryParse(_integrationController.text) ?? 0;
    setState(() {
      _integrationConflict =
          hasCoupon && couponStatus == 0 && useIntegration > 0;
    });
  }

  /// Task 5: 打开地址选择弹窗
  void _openAddressSheet() {
    if (_orderData == null) return;
    showModalBottomSheet<int>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => AddressSelectSheet(
        addresses: _orderData!.memberReceiveAddressList,
        selectedAddressId: _orderData!.memberReceiveAddressList.isNotEmpty &&
                _selectedAddressIndex <
                    _orderData!.memberReceiveAddressList.length
            ? _orderData!.memberReceiveAddressList[_selectedAddressIndex].id
            : 0,
        onAddressesChanged: () => _loadConfirmOrder(), // LOW-2: 地址变更后刷新确认单
      ),
    ).then((selectedId) {
      if (selectedId != null && selectedId > 0) {
        setState(() {
          final addrs = _orderData!.memberReceiveAddressList;
          for (int i = 0; i < addrs.length; i++) {
            if (addrs[i].id == selectedId) {
              _selectedAddressIndex = i;
              break;
            }
          }
          // 边界保护：若未找到匹配则保持原值
          if (_selectedAddressIndex >= addrs.length) {
            _selectedAddressIndex = 0;
          }
        });
      }
    });
  }

  /// Task 6: 打开优惠券选择弹窗
  void _openCouponSheet() async {
    if (_orderData == null) return;
    final result = await showModalBottomSheet<ConfirmCouponData?>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => CouponSelectSheet(
        enableList: _orderData!.couponHistoryDetailList.enableList,
        disableList: _orderData!.couponHistoryDetailList.disableList,
        selectedCoupon: _selectedCoupon,
      ),
    );
    if (result != null) {
      setState(() {
        _selectedCoupon = result;
        final data = _orderData;
        if (data != null && data.integrationConsumeSetting.couponStatus == 0) {
          _integrationController.text = "0";
          _previewIntegrationAmount = 0;
        }
      });
      _checkIntegrationConflict();
      _recalcPreview();
    }
  }

  /// Task 11: 重新计算预览金额（本地计算，不调后端）
  void _recalcPreview() {
    // 优惠券金额在 _calcFinalPayAmount 中直接计算，无需独立状态
  }

  /// Task 7: 提交订单（Story 5.4: 幂等键 + postWithHeaders）
  Future<void> _submitOrder() async {
    if (_isSubmitting) return;

    // === Task 9.1: 校验收货地址 ===
    final addresses = _orderData?.memberReceiveAddressList ?? [];
    if (addresses.isEmpty || _selectedAddressIndex >= addresses.length) {
      _showInlineError("请先添加收货地址");
      return;
    }
    final selectedAddr = addresses[_selectedAddressIndex];

    // === Task 9.2: 校验支付方式 ===
    if (_selectedPayType != 1 && _selectedPayType != 2) {
      _showInlineError("请选择支付方式");
      return;
    }

    // === Task 7.1: 积分与优惠券互斥 ===
    if (_integrationConflict) {
      _showInlineError("该优惠券不支持与积分共用");
      return;
    }

    final cartModel = context.read<CartModel>();
    final cartIds = cartModel.getCheckProduct().map((e) => e.id).toList();
    final useIntegration = int.tryParse(_integrationController.text) ?? 0;

    setState(() => _isSubmitting = true);

    try {
      // === Task 7.1: 生成幂等键（使用真实会员 ID，而非地址 ID）===
      // 格式: {userId}:{timestamp}:{cartIds}:{couponId}:{useIntegration}:{uuid}
      final uuid = const Uuid();
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final memberId = selectedAddr.memberId; // 修复 CRITICAL-1：使用会员 ID 而非地址 ID
      final checkoutSource = _isDirectBuy
          ? 'direct:${widget.directItem!.productId}:${widget.directItem!.productSkuId}:${widget.directItem!.quantity}'
          : cartIds.join(',');
      final idempotencyKey =
          '$memberId:$timestamp:$checkoutSource:${_selectedCoupon?.id ?? 0}:$useIntegration:${uuid.v4()}';

      // === Task 7.2: 通过 HTTP header 传递幂等键 ===
      final resp = await HttpUtil.postWithHeaders(
        generateOrderUrl,
        data: _isDirectBuy
            ? {
                "directItem": widget.directItem!.toJson(),
                "memberReceiveAddressId": selectedAddr.id,
                "couponId": _selectedCoupon?.id ?? 0,
                "useIntegration": useIntegration,
                "payType": _selectedPayType,
                "note": _remarkController.text.trim(),
              }
            : {
                "cartIds": cartIds,
                "memberReceiveAddressId": selectedAddr.id,
                "couponId": _selectedCoupon?.id ?? 0,
                "useIntegration": useIntegration,
                "payType": _selectedPayType,
                "note": _remarkController.text.trim(),
              },
        headers: {"X-Idempotency-Key": idempotencyKey},
      );

      if (!mounted) return;

      if (resp.data["code"] == 0) {
        final orderData = resp.data["data"];
        final orderId = orderData?["id"] ?? 0;
        final orderSn = orderData?["orderSn"] ?? '';
        final payAmount = _calcFinalPayAmount();

        // Task 9.3: 清空已结算购物车项
        if (!_isDirectBuy) {
          try {
            await HttpUtil.post(deleteCartUrl, data: {"ids": cartIds});
          } catch (_) {}
        }

        if (!mounted) return;

        // === Task 8: 跳转支付页，传递完整订单信息 ===
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (ctx) => OrderPay(
              orderId: orderId,
              orderSn: orderSn,
              payType: _selectedPayType,
              amount: payAmount,
            ),
          ),
        );
      } else {
        _handleSubmitFailure(resp.data);
        setState(() => _isSubmitting = false);
      }
    } on DioException catch (e) {
      if (mounted) {
        if (e.response != null) {
          _handleSubmitFailure(e.response?.data);
        } else {
          _showInlineError("网络请求失败，请检查网络连接");
        }
        setState(() => _isSubmitting = false);
      }
    } catch (_) {
      if (mounted) {
        _showInlineError("网络请求失败，请检查网络连接");
        setState(() => _isSubmitting = false);
      }
    }
  }

  // Story 5.4 错误码常量（与后端 errorx 保持一致）
  static const String errCodeOrderDuplicatedRequest =
      'OMS_ORDER_DUPLICATED_REQUEST';
  static const String errCodeOrderCompensationFailed =
      'OMS_ORDER_COMPENSATION_FAILED';
  static const String errCodeOrderStockLocked = 'OMS_ORDER_STOCK_LOCKED';
  static const String errCodeOrderStockInsufficient =
      'OMS_ORDER_STOCK_INSUFFICIENT';
  static const String errCodeOrderCouponUnavailable =
      'OMS_ORDER_COUPON_UNAVAILABLE';
  static const String errCodeOrderIntegrationExceed =
      'OMS_ORDER_INTEGRATION_EXCEED';
  static const String errCodeOrderAddressInvalid = 'OMS_ORDER_ADDRESS_INVALID';
  static const String errCodeOrderPayTypeInvalid = 'OMS_ORDER_PAY_TYPE_INVALID';

  void _handleSubmitFailure(dynamic data) {
    final errCode = _extractOrderErrorCode(data);
    final msg = _extractOrderErrorMessage(data);

    if (errCode == errCodeOrderDuplicatedRequest) {
      _showInlineError("订单正在处理中，请稍后查看");
    } else if (errCode == errCodeOrderCompensationFailed) {
      _showSubmitFailureWithActions("订单创建异常，请稍后重试");
    } else if (errCode == errCodeOrderStockLocked) {
      _showSubmitFailureWithActions("库存已被其他订单占用，请返回购物车重新选择");
    } else if (errCode == errCodeOrderStockInsufficient) {
      _showSubmitFailureWithActions("库存不足，请返回购物车重新选择");
    } else if (errCode == errCodeOrderCouponUnavailable) {
      _showSubmitFailureWithActions("优惠券不可用，请返回重新选择");
    } else if (errCode == errCodeOrderIntegrationExceed) {
      _showInlineError("积分不足，请调整使用数量");
    } else if (errCode == errCodeOrderAddressInvalid) {
      _showSubmitFailureWithActions("收货地址无效，请重新选择");
    } else if (errCode == errCodeOrderPayTypeInvalid) {
      _showSubmitFailureWithActions("支付方式无效，请重新选择");
    } else if (data is Map &&
        data["code"] == 0 &&
        data["data"]?["id"] != null) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (ctx) => OrderPay(
            orderId: data["data"]["id"],
            orderSn: data["data"]["orderSn"] ?? '',
            payType: _selectedPayType,
            amount: (data["data"]["payAmount"] ?? 0).toDouble(),
          ),
        ),
      );
    } else {
      _showInlineError(msg);
    }
  }

  String _extractOrderErrorCode(dynamic data) {
    if (data is String) {
      return data;
    }
    if (data is Map) {
      return data["message"]?.toString() ??
          data["data"]?.toString() ??
          data["error"]?.toString() ??
          "";
    }
    return "";
  }

  String _extractOrderErrorMessage(dynamic data) {
    final code = _extractOrderErrorCode(data);
    if (code.isNotEmpty) {
      return code;
    }
    return "提交失败，请稍后重试";
  }

  /// 计算最终实付金额（与底部栏保持一致）
  double _calcFinalPayAmount() {
    final calc = _orderData?.calcAmount;
    if (calc == null) return 0;
    final totalAmount = calc.totalAmount.toDouble();
    final promotionAmount = calc.promotionAmount.toDouble();
    final couponAmount = _selectedCoupon?.amount ?? 0;
    final integrationAmount = _previewIntegrationAmount.toDouble();
    final payAmount =
        totalAmount - promotionAmount - couponAmount - integrationAmount;
    return payAmount < 0 ? 0 : payAmount;
  }

  void _showInlineError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: AppColors.price,
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  // ===== Build 方法 =====

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Scaffold(
        backgroundColor: AppColors.background,
        appBar: _buildOrderAppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (_errorMessage != null) {
      return Scaffold(
        backgroundColor: AppColors.background,
        appBar: _buildOrderAppBar(),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: _checkoutCard(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(
                    Icons.receipt_long_outlined,
                    color: AppColors.textHint,
                    size: 42,
                  ),
                  const SizedBox(height: AppSpacing.md),
                  Text(
                    _errorMessage!,
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      fontSize: 15,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: AppSpacing.lg),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: _loadConfirmOrder,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        foregroundColor: Colors.white,
                        elevation: 0,
                      ),
                      child: const Text("重试"),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: _buildOrderAppBar(),
      body: SafeArea(
        top: false,
        child: Container(
          color: AppColors.background,
          width: MediaQuery.of(context).size.width,
          child: CustomScrollView(
            slivers: [
              buildAddress(),
              buildProductList(),
              buildCoupon(),
              buildIntegration(),
              buildPayType(),
              buildOrderInfo(),
            ],
          ),
        ),
      ),
      bottomNavigationBar: buildSubmit(),
    );
  }

  PreferredSizeWidget _buildOrderAppBar() {
    return AppBar(
      backgroundColor: AppColors.surface,
      surfaceTintColor: AppColors.surface,
      elevation: 0,
      title: const Text("确认订单"),
      titleTextStyle: const TextStyle(
        fontSize: 16,
        color: AppColors.textPrimary,
        fontWeight: FontWeight.w600,
      ),
      centerTitle: true,
    );
  }

  BoxDecoration _checkoutDecoration({Color color = AppColors.surface}) {
    return BoxDecoration(
      color: color,
      borderRadius: BorderRadius.circular(AppRadii.lg),
      border: Border.all(color: AppColors.border.withValues(alpha: 0.78)),
    );
  }

  Widget _checkoutCard({
    required Widget child,
    EdgeInsetsGeometry margin = const EdgeInsets.fromLTRB(
      AppSpacing.md,
      AppSpacing.md,
      AppSpacing.md,
      0,
    ),
    EdgeInsetsGeometry padding = const EdgeInsets.all(AppSpacing.lg),
    Color color = AppColors.surface,
  }) {
    return Container(
      margin: margin,
      padding: padding,
      decoration: _checkoutDecoration(color: color),
      child: child,
    );
  }

  Widget _sectionTitle(String title) {
    return Text(
      title,
      style: const TextStyle(
        fontSize: 15,
        color: AppColors.textPrimary,
        fontWeight: FontWeight.w700,
      ),
    );
  }

  Widget _miniBadge({
    required String label,
    required Color color,
    Color? background,
  }) {
    return Container(
      padding:
          const EdgeInsets.symmetric(horizontal: AppSpacing.sm, vertical: 3),
      decoration: BoxDecoration(
        color: background ?? color.withValues(alpha: 0.10),
        borderRadius: BorderRadius.circular(AppRadii.sm),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 12,
          color: color,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }

  // Task 5: 收货地址行（可点击选择）
  Widget buildAddress() {
    final addresses = _orderData?.memberReceiveAddressList ?? [];
    final addr =
        addresses.isNotEmpty && _selectedAddressIndex < addresses.length
            ? addresses[_selectedAddressIndex]
            : null;

    return SliverToBoxAdapter(
      child: _checkoutCard(
        padding: EdgeInsets.zero,
        child: InkWell(
          onTap: _openAddressSheet,
          borderRadius: BorderRadius.circular(AppRadii.lg),
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.lg),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Container(
                  width: 34,
                  height: 34,
                  decoration: BoxDecoration(
                    color: AppColors.success.withValues(alpha: 0.10),
                    borderRadius: BorderRadius.circular(AppRadii.md),
                  ),
                  child: const Icon(
                    Icons.location_on_outlined,
                    color: AppColors.success,
                    size: 20,
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: addr != null
                      ? Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              "${addr.receiverName} ${addr.receiverPhone}",
                              style: const TextStyle(
                                fontSize: 16,
                                color: AppColors.textPrimary,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: AppSpacing.xs),
                            Text(
                              "${addr.province} ${addr.city} ${addr.district} ${addr.detailAddress}",
                              style: const TextStyle(
                                fontSize: 13,
                                height: 1.35,
                                color: AppColors.textSecondary,
                              ),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ],
                        )
                      : Row(
                          children: [
                            const Text(
                              "去添加收货地址",
                              style: TextStyle(
                                fontSize: 15,
                                color: AppColors.price,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(width: AppSpacing.xs),
                            const Icon(
                              Icons.add,
                              color: AppColors.price,
                              size: 20,
                            ),
                          ],
                        ),
                ),
                const Icon(
                  Icons.chevron_right,
                  color: AppColors.textHint,
                  size: 22,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // 商品列表
  Widget buildProductList() {
    final products = _orderData?.cartPromotionItemList ?? [];
    return SliverToBoxAdapter(
      child: _checkoutCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _sectionTitle(_isDirectBuy ? "立即购买商品" : "商品信息"),
            const SizedBox(height: AppSpacing.md),
            ...products.asMap().entries.map((entry) {
              final index = entry.key;
              final item = entry.value;
              final finalPrice = item.price - item.reduceAmount;
              return Column(
                children: [
                  if (index > 0) const Divider(height: AppSpacing.xxl),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(AppRadii.md),
                        child: CachedImageWidget(
                          76,
                          76,
                          item.productPic,
                          fit: BoxFit.cover,
                        ),
                      ),
                      const SizedBox(width: AppSpacing.md),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              item.productName,
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                fontSize: 14,
                                height: 1.35,
                                color: AppColors.textPrimary,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            if (item.productSubTitle.isNotEmpty) ...[
                              const SizedBox(height: AppSpacing.xs),
                              Text(
                                item.productSubTitle,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  fontSize: 12,
                                  color: AppColors.textHint,
                                ),
                              ),
                            ],
                            if (item.productAttr.isNotEmpty) ...[
                              const SizedBox(height: AppSpacing.xs),
                              Text(
                                item.productAttr,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  fontSize: 12,
                                  color: AppColors.textHint,
                                ),
                              ),
                            ],
                            if (item.promotionMessage.isNotEmpty &&
                                item.reduceAmount > 0) ...[
                              const SizedBox(height: AppSpacing.sm),
                              _miniBadge(
                                label: item.promotionMessage,
                                color: AppColors.price,
                              ),
                            ],
                            // Story 10.6: 数字资产商品履约模式提示
                            if (item.fulfillmentMode == 'digital_asset') ...[
                              const SizedBox(height: AppSpacing.sm),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.blue.shade50,
                                  borderRadius: BorderRadius.circular(4),
                                  border: Border.all(
                                    color: Colors.blue.shade200,
                                    width: 0.5,
                                  ),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(
                                      Icons.card_giftcard,
                                      size: 14,
                                      color: Colors.blue.shade600,
                                    ),
                                    const SizedBox(width: 4),
                                    Flexible(
                                      child: Text(
                                        '该商品为数字资产，支付成功后将发放至您的数字卡包',
                                        style: TextStyle(
                                          fontSize: 11,
                                          color: Colors.blue.shade700,
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                            const SizedBox(height: AppSpacing.sm),
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Text(
                                  "￥${_formatAmount(finalPrice)}",
                                  style: const TextStyle(
                                    fontSize: 17,
                                    color: AppColors.price,
                                    fontWeight: FontWeight.w800,
                                  ),
                                ),
                                const SizedBox(width: AppSpacing.xs),
                                Text(
                                  "x${item.quantity}",
                                  style: const TextStyle(
                                    fontSize: 12,
                                    color: AppColors.textHint,
                                  ),
                                ),
                                if (item.reduceAmount > 0) ...[
                                  const SizedBox(width: AppSpacing.sm),
                                  Text(
                                    "￥${_formatAmount(item.price)}",
                                    style: const TextStyle(
                                      fontSize: 12,
                                      color: AppColors.textHint,
                                      decoration: TextDecoration.lineThrough,
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              );
            }),
          ],
        ),
      ),
    );
  }

  // Task 6: 优惠券选择行
  Widget buildCoupon() {
    final enableCoupons = _orderData?.couponHistoryDetailList.enableList ?? [];

    final valueText = _selectedCoupon != null
        ? "-￥${_formatAmount(_selectedCoupon!.amount)} ${_selectedCoupon!.name}"
        : enableCoupons.isNotEmpty
            ? "${enableCoupons.length}张可用"
            : "暂无可用优惠券";

    return SliverToBoxAdapter(
      child: _checkoutCard(
        padding: EdgeInsets.zero,
        child: InkWell(
          onTap: _openCouponSheet,
          borderRadius: BorderRadius.circular(AppRadii.lg),
          child: Padding(
            padding: const EdgeInsets.symmetric(
              horizontal: AppSpacing.lg,
              vertical: AppSpacing.md,
            ),
            child: Row(
              children: [
                _miniBadge(label: "券", color: AppColors.price),
                const SizedBox(width: AppSpacing.md),
                const Expanded(
                  child: Text(
                    "优惠券",
                    style: TextStyle(
                      fontSize: 14,
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                Flexible(
                  child: Text(
                    valueText,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.right,
                    style: TextStyle(
                      fontSize: 13,
                      color: _selectedCoupon != null || enableCoupons.isNotEmpty
                          ? AppColors.price
                          : AppColors.textHint,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                const SizedBox(width: AppSpacing.xs),
                const Icon(
                  Icons.chevron_right,
                  color: AppColors.textHint,
                  size: 20,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // Task 7: 积分抵扣行
  Widget buildIntegration() {
    final data = _orderData;
    final setting = data?.integrationConsumeSetting;
    final totalAmt = data?.calcAmount.totalAmount ?? 0;
    int maxIntegration = data?.memberIntegration ?? 0;
    if (setting != null && setting.maxPercentPerOrder > 0 && totalAmt > 0) {
      final maxByPercent = (totalAmt * setting.maxPercentPerOrder ~/ 100) *
          setting.deductionPerAmount;
      if (maxByPercent < maxIntegration) maxIntegration = maxByPercent;
    }
    final hasEnoughPoints = (data?.memberIntegration ?? 0) > 0;
    final useUnit = setting?.useUnit ?? 100;

    return SliverToBoxAdapter(
      child: _checkoutCard(
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.lg,
          vertical: AppSpacing.md,
        ),
        margin: const EdgeInsets.fromLTRB(
            AppSpacing.md, AppSpacing.sm, AppSpacing.md, 0),
        child: Row(
          children: [
            _miniBadge(label: "积", color: AppColors.accent),
            const SizedBox(width: AppSpacing.md),
            const Expanded(
              child: Text(
                "积分抵扣",
                style: TextStyle(
                  fontSize: 14,
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
            if (!hasEnoughPoints)
              const Text(
                "暂无可用积分",
                style: TextStyle(fontSize: 13, color: AppColors.textHint),
              )
            else if (_integrationConflict)
              const Flexible(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    Icon(Icons.warning_amber,
                        color: AppColors.accent, size: 16),
                    SizedBox(width: AppSpacing.xs),
                    Flexible(
                      child: Text(
                        "该优惠券不支持与积分共用",
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          fontSize: 12,
                          color: AppColors.accent,
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else ...[
              GestureDetector(
                onTap: () => _onStepperPressed(-1),
                child: Container(
                  width: 30,
                  height: 28,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    border: Border.all(color: AppColors.border),
                    borderRadius: const BorderRadius.horizontal(
                      left: Radius.circular(AppRadii.sm),
                    ),
                  ),
                  child: const Text(
                    "—",
                    style: TextStyle(
                      fontSize: 14,
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              SizedBox(
                width: 62,
                height: 28,
                child: TextField(
                  controller: _integrationController,
                  keyboardType: TextInputType.number,
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.textPrimary,
                  ),
                  decoration: const InputDecoration(
                    contentPadding: EdgeInsets.zero,
                    isDense: true,
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.zero,
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.zero,
                      borderSide: BorderSide(color: AppColors.border),
                    ),
                    filled: true,
                    fillColor: AppColors.surfaceMuted,
                  ),
                  onChanged: _onIntegrationChanged,
                ),
              ),
              GestureDetector(
                onTap: () => _onStepperPressed(1),
                child: Container(
                  width: 30,
                  height: 28,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    border: Border.all(color: AppColors.border),
                    borderRadius: const BorderRadius.horizontal(
                      right: Radius.circular(AppRadii.sm),
                    ),
                  ),
                  child: const Text(
                    "+",
                    style: TextStyle(
                      fontSize: 14,
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: AppSpacing.sm),
              Flexible(
                child: Text(
                  _previewIntegrationAmount > 0
                      ? "-￥${_formatAmount(_previewIntegrationAmount)}"
                      : "可用${data?.memberIntegration ?? 0}积分（$useUnit 积分起）",
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: _previewIntegrationAmount > 0 ? 13 : 12,
                    color: _previewIntegrationAmount > 0
                        ? AppColors.price
                        : AppColors.textHint,
                    fontWeight: _previewIntegrationAmount > 0
                        ? FontWeight.w600
                        : FontWeight.w400,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  // Task 8: 支付方式选择
  Widget buildPayType() {
    return SliverToBoxAdapter(
      child: _checkoutCard(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.lg,
          AppSpacing.lg,
          AppSpacing.sm,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _sectionTitle("支付方式"),
            const SizedBox(height: AppSpacing.sm),
            _PayTypeItem(
              icon: "images/wx_pay.png",
              label: "微信支付",
              desc: "推荐使用",
              value: 2,
              groupValue: _selectedPayType,
              onChanged: (v) => setState(() => _selectedPayType = v),
            ),
            const Divider(height: 1),
            _PayTypeItem(
              icon: "images/ali_pay.png",
              label: "支付宝支付",
              desc: "",
              value: 1,
              groupValue: _selectedPayType,
              onChanged: (v) => setState(() => _selectedPayType = v),
            ),
          ],
        ),
      ),
    );
  }

  // Task 11: Price Breakdown Card
  Widget buildOrderInfo() {
    final calc = _orderData?.calcAmount;
    final totalAmount = calc?.totalAmount ?? 0;
    final freightAmount = calc?.freightAmount ?? 0;
    final promotionAmount = calc?.promotionAmount ?? 0;
    final couponAmount = _selectedCoupon?.amount ?? 0;
    final integrationAmount = _previewIntegrationAmount;

    return SliverToBoxAdapter(
      child: _checkoutCard(
        margin: const EdgeInsets.fromLTRB(
          AppSpacing.md,
          AppSpacing.md,
          AppSpacing.md,
          AppSpacing.lg,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _sectionTitle("金额明细"),
            const SizedBox(height: AppSpacing.sm),
            _AmountRow(label: "商品合计", value: totalAmount),
            const Divider(height: 1),
            _AmountRow(label: "运费", value: freightAmount),
            const Divider(height: 1),
            _AmountRow(
              label: "活动优惠",
              value: -promotionAmount,
              valueColor:
                  promotionAmount > 0 ? AppColors.price : AppColors.textHint,
            ),
            const Divider(height: 1),
            _AmountRow(
              label: "优惠券",
              value: couponAmount > 0 ? -couponAmount : 0,
              valueColor:
                  couponAmount > 0 ? AppColors.price : AppColors.textHint,
            ),
            const Divider(height: 1),
            _AmountRow(
              label: "积分抵扣",
              value: integrationAmount > 0 ? -integrationAmount : 0,
              valueColor:
                  integrationAmount > 0 ? AppColors.price : AppColors.textHint,
            ),
            const Divider(height: 1),
            SizedBox(
              height: 46,
              child: Row(
                children: [
                  const Text(
                    "备注",
                    style: TextStyle(
                      fontSize: 13,
                      color: AppColors.textSecondary,
                    ),
                  ),
                  const SizedBox(width: AppSpacing.md),
                  Expanded(
                    child: TextField(
                      controller: _remarkController,
                      decoration: const InputDecoration(
                        hintText: "选填，可填写备注信息",
                        contentPadding: EdgeInsets.zero,
                        isDense: true,
                        border: InputBorder.none,
                        hintStyle: TextStyle(
                          fontSize: 12,
                          color: AppColors.textHint,
                        ),
                      ),
                      style: const TextStyle(
                        fontSize: 13,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // Task 10: 底部提交栏（含提交前确认对话框）
  Widget buildSubmit() {
    final payAmount = _calcFinalPayAmount();

    return SafeArea(
      top: false,
      child: Container(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.sm,
          AppSpacing.lg,
          AppSpacing.sm,
        ),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          border: Border(
            top: BorderSide(color: AppColors.border, width: 1),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2),
                    child: Text(
                      "实付款 ",
                      style: TextStyle(
                        fontSize: 13,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ),
                  const Padding(
                    padding: EdgeInsets.only(bottom: 2),
                    child: Text(
                      "￥",
                      style: TextStyle(
                        fontSize: 14,
                        color: AppColors.price,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                  TweenAnimationBuilder<double>(
                    tween: Tween<double>(begin: payAmount, end: payAmount),
                    duration: const Duration(milliseconds: 300),
                    builder: (context, val, _) {
                      return Text(
                        _formatAmount(val),
                        style: const TextStyle(
                          fontSize: 22,
                          color: AppColors.price,
                          fontWeight: FontWeight.w800,
                        ),
                      );
                    },
                  ),
                ],
              ),
            ),
            SizedBox(
              width: 136,
              height: 48,
              child: ElevatedButton(
                onPressed:
                    _isSubmitting ? null : () => _confirmAndSubmit(payAmount),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.primary,
                  disabledBackgroundColor: AppColors.textHint,
                  foregroundColor: Colors.white,
                  elevation: 0,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(24),
                  ),
                ),
                child: _isSubmitting
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation(Colors.white),
                        ),
                      )
                    : const Text(
                        '提交订单',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Task 10.1 & HIGH-3: 提交前确认对话框（展示商品摘要 + 提交锁定提示）
  Future<void> _confirmAndSubmit(double payAmount) async {
    final products = _orderData?.cartPromotionItemList ?? [];
    final productSummary = products.length == 1
        ? products.first.productName
        : "${products.first.productName} 等${products.length}件商品";

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("确认提交订单？"),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              "商品：$productSummary",
              style: const TextStyle(fontSize: 14),
            ),
            const SizedBox(height: 4),
            Text(
              "提交后将锁定库存和优惠，实付款 ￥${_formatAmount(payAmount)}，是否继续？",
              style: const TextStyle(fontSize: 14),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text("取消"),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.primary,
            ),
            child: const Text("确认提交"),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _submitOrder();
    }
  }

  // MEDIUM-3: 提交失败双入口（返回购物车 / 重试）
  void _showSubmitFailureWithActions(String msg) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("提交失败"),
        content: Text(msg, style: const TextStyle(fontSize: 14)),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(ctx).pop();
              Navigator.of(context).pop(); // 返回购物车
            },
            child: const Text("返回购物车"),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(ctx).pop();
              setState(() => _isSubmitting = false);
            },
            style: TextButton.styleFrom(
              foregroundColor: AppColors.primary,
            ),
            child: const Text("重试"),
          ),
        ],
      ),
    );
  }
}

// Task 8: 支付方式单选行
class _PayTypeItem extends StatelessWidget {
  final String icon;
  final String label;
  final String desc;
  final int value;
  final int groupValue;
  final ValueChanged<int> onChanged;

  const _PayTypeItem({
    required this.icon,
    required this.label,
    required this.desc,
    required this.value,
    required this.groupValue,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final selected = value == groupValue;

    return InkWell(
      onTap: () => onChanged(value),
      child: Container(
        constraints: const BoxConstraints(minHeight: 56),
        child: Row(
          children: [
            Image.asset(icon, height: 22, width: 22),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    label,
                    style: const TextStyle(
                      fontSize: 14,
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  if (desc.isNotEmpty) const SizedBox(height: 2),
                  if (desc.isNotEmpty)
                    Text(
                      desc,
                      style: const TextStyle(
                        fontSize: 11,
                        color: AppColors.textHint,
                      ),
                    ),
                ],
              ),
            ),
            Icon(
              selected ? Icons.check_circle : Icons.circle_outlined,
              color: selected ? AppColors.primary : AppColors.border,
              size: 22,
            ),
          ],
        ),
      ),
    );
  }
}

// Task 11: 金额行 Widget
class _AmountRow extends StatelessWidget {
  final String label;
  final num value; // 单位：元
  final Color? valueColor;

  const _AmountRow({
    required this.label,
    required this.value,
    this.valueColor,
  });

  @override
  Widget build(BuildContext context) {
    final isNegative = value < 0;
    final absVal = value.abs();
    return SizedBox(
      height: 42,
      child: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
              ),
            ),
          ),
          Text(
            "${isNegative ? "-" : ""}￥${formatCurrencyAmount(absVal)}",
            style: TextStyle(
              fontSize: 13,
              color: valueColor ?? AppColors.textPrimary,
              fontWeight:
                  valueColor != null ? FontWeight.w600 : FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }
}

String formatCurrencyAmount(num amount) {
  final normalized = amount.toDouble();
  if (normalized == normalized.truncateToDouble()) {
    return normalized.toStringAsFixed(0);
  }
  return normalized.toStringAsFixed(2);
}
