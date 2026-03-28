import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/confirm_order.dart';
import 'package:flutter_mall/provider/cart_model.dart';
import 'package:flutter_mall/utils/http_util.dart';
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
  const OrderSubmit({super.key});

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
  final TextEditingController _remarkController = TextEditingController(); // 订单备注（MEDIUM-3）

  // Task 8: 支付方式（1=支付宝，2=微信，默认微信）
  int _selectedPayType = 2;

  // Task 9: 提交订单防重复
  bool _isSubmitting = false;

  // Price Breakdown Card（Task 11）
  int _previewCouponAmount = 0;
  int _previewIntegrationAmount = 0;

  @override
  void initState() {
    super.initState();
    _loadConfirmOrder();
  }

  @override
  void dispose() {
    _integrationController.dispose();
    _remarkController.dispose();
    super.dispose();
  }

  /// 加载确认单数据
  void _loadConfirmOrder() async {
    try {
      setState(() {
        _loading = true;
        _errorMessage = null;
      });
      final cartModel = context.read<CartModel>();
      final ids = cartModel.getCheckProduct().map((e) => e.id).toList();
      Response result = await HttpUtil.post(
          generateConfirmOrderUrl, data: {"ids": ids});
      ConfirmOrderModel model = ConfirmOrderModel.fromJson(result.data);
      if (mounted) {
        setState(() {
          _orderData = model.data;
          _loading = false;
          // 自动选择默认地址
          _selectedAddressIndex = 0;
          for (int i = 0;
              i < model.data.memberReceiveAddressList.length;
              i++) {
            if (model.data.memberReceiveAddressList[i].isDefault == 1) {
              _selectedAddressIndex = i;
              break;
            }
          }
          // Task 7: 计算最大可用积分并默认填入
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

  /// Task 7: 初始化积分输入
  void _initIntegrationInput() {
    final data = _orderData;
    if (data == null) return;
    final integration = data.memberIntegration;
    final setting = data.integrationConsumeSetting;
    if (integration <= 0 || setting.deductionPerAmount <= 0) return;
    final totalAmt = data.calcAmount.totalAmount;
    int maxByPercent = 0;
    if (setting.maxPercentPerOrder > 0 && totalAmt > 0) {
      maxByPercent =
          (totalAmt * setting.maxPercentPerOrder ~/ 100) *
              setting.deductionPerAmount;
    }
    final maxIntegration =
        integration < maxByPercent ? integration : maxByPercent;
    _integrationController.text = maxIntegration.toString();
    _previewIntegrationAmount = _calcIntegrationAmount(maxIntegration);
    _checkIntegrationConflict();
  }

  int _calcIntegrationAmount(int useIntegration) {
    final data = _orderData;
    if (data == null || data.integrationConsumeSetting.deductionPerAmount <= 0) {
      return 0;
    }
    return useIntegration ~/
        data.integrationConsumeSetting.deductionPerAmount;
  }

  void _onIntegrationChanged(String val) {
    final data = _orderData;
    if (data == null) return;
    final v = int.tryParse(val) ?? 0;
    final integration = data.memberIntegration;
    final maxByPercent =
        data.integrationConsumeSetting.maxPercentPerOrder > 0
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
    final maxByPercent =
        data.integrationConsumeSetting.maxPercentPerOrder > 0
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
    setState(() {
      _integrationConflict = hasCoupon && couponStatus == 0;
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
        selectedAddressId:
            _orderData!.memberReceiveAddressList.isNotEmpty &&
                    _selectedAddressIndex <
                        _orderData!.memberReceiveAddressList.length
                ? _orderData!.memberReceiveAddressList[_selectedAddressIndex].id
                : 0,
        onAddressesChanged: () => _loadConfirmOrder(),  // LOW-2: 地址变更后刷新确认单
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
        if (data != null &&
            data.integrationConsumeSetting.couponStatus == 0) {
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
    final data = _orderData;
    if (data == null) return;
    setState(() {
      _previewCouponAmount = _selectedCoupon != null
          ? (_selectedCoupon!.amount * 100).toInt()
          : 0;
    });
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
    final useIntegration =
        int.tryParse(_integrationController.text) ?? 0;

    setState(() => _isSubmitting = true);

    try {
      // === Task 7.1: 生成幂等键 ===
      // 格式: {userId}:{timestamp}:{cartIds}:{couponId}:{useIntegration}:{uuid}
      final uuid = const Uuid();
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final memberId = selectedAddr.id;
      final cartIdsStr = cartIds.join(',');
      final idempotencyKey = '$memberId:$timestamp:${cartIdsStr}:${_selectedCoupon?.id ?? 0}:$useIntegration:${uuid.v4()}';

      // === Task 7.2: 通过 HTTP header 传递幂等键 ===
      final resp = await HttpUtil.postWithHeaders(
        generateOrderUrl,
        data: {
          "cartIds": cartIds,
          "memberReceiveAddressId": selectedAddr.id,
          "couponId": _selectedCoupon?.id ?? 0,
          "useIntegration": useIntegration,
          "payType": _selectedPayType,
          "note": _remarkController.text.trim(),  // MEDIUM-3：订单备注
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
        try {
          await HttpUtil.post(deleteCartUrl, data: {"ids": cartIds});
        } catch (_) {}

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
        // === Task 7.3 & Task 8: 幂等命中或其他错误的差异化提示 ===
        final msg = resp.data["message"] ?? "提交失败，请稍后重试";
        final code = resp.data["code"] ?? -1;

        if (code == 'OMS_ORDER_DUPLICATED_REQUEST') {
          _showInlineError("订单正在处理中，请稍后查看");
        } else if (code == 'OMS_ORDER_COMPENSATION_FAILED') {
          _showInlineError("订单创建异常，请稍后重试");
        } else if (code == 'OMS_ORDER_STOCK_LOCKED') {
          _showInlineError("库存已被其他订单占用，请返回购物车重新选择");
        } else {
          _showInlineError(msg);
        }
        setState(() => _isSubmitting = false);
      }
    } catch (e) {
      if (mounted) {
        _showInlineError("网络请求失败，请检查网络连接");
        setState(() => _isSubmitting = false);
      }
    }
  }

  /// 计算最终实付金额（与底部栏保持一致）
  int _calcFinalPayAmount() {
    final calc = _orderData?.calcAmount;
    if (calc == null) return 0;
    final totalAmount = calc.totalAmount;
    final promotionAmount = calc.promotionAmount;
    final couponAmount = _selectedCoupon != null
        ? (_selectedCoupon!.amount * 100).toInt()
        : 0;
    final integrationAmount = _previewIntegrationAmount;
    return totalAmount - promotionAmount - couponAmount - integrationAmount;
  }

  void _showInlineError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: Colors.red,
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
        appBar: AppBar(
          backgroundColor: Colors.white,
          title: const Text("创建订单"),
          titleTextStyle:
              const TextStyle(fontSize: 16, color: Colors.black),
          centerTitle: true,
        ),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (_errorMessage != null) {
      return Scaffold(
        appBar: AppBar(
          backgroundColor: Colors.white,
          title: const Text("创建订单"),
          titleTextStyle:
              const TextStyle(fontSize: 16, color: Colors.black),
          centerTitle: true,
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(_errorMessage!, style: const TextStyle(fontSize: 15)),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: _loadConfirmOrder,
                child: const Text("重试"),
              ),
            ],
          ),
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("创建订单"),
        titleTextStyle:
            const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
      ),
      body: SafeArea(
        child: Container(
          color: Colors.white,
          width: MediaQuery.of(context).size.width,
          child: Stack(
            alignment: Alignment.topCenter,
            children: [
              CustomScrollView(
                slivers: [
                  buildAddress(),
                  buildProductTitle(),
                  buildProductList(),
                  buildCoupon(),
                  buildIntegration(),
                  buildPayType(),
                  buildOrderInfo(),
                ],
              ),
              Positioned(bottom: 0, child: buildSubmit()),
            ],
          ),
        ),
      ),
    );
  }

  // Task 5: 收货地址行（可点击选择）
  SliverPadding buildAddress() {
    final border = BorderSide(
        width: 5, color: const Color(0xFFF5F5F5));
    final addresses = _orderData?.memberReceiveAddressList ?? [];
    final addr = addresses.isNotEmpty &&
            _selectedAddressIndex < addresses.length
        ? addresses[_selectedAddressIndex]
        : null;

    return SliverPadding(
      padding: EdgeInsets.zero,
      sliver: SliverList(
        delegate: SliverChildListDelegate(<Widget>[
          InkWell(
            onTap: _openAddressSheet,
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                border: Border(bottom: border),
              ),
              height: 78,
              padding: const EdgeInsets.symmetric(horizontal: 15),
              child: Row(
                children: [
                  Image.asset(
                    "images/address.png",
                    height: 24,
                    width: 22,
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: addr != null
                        ? Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text(
                                "${addr.receiverName} ${addr.receiverPhone}",
                                style: const TextStyle(
                                    fontSize: 17,
                                    fontWeight: FontWeight.w500),
                              ),
                              const SizedBox(height: 5),
                              Text(
                                "${addr.province} ${addr.city} ${addr.district} ${addr.detailAddress}",
                                style: const TextStyle(
                                    fontSize: 14, color: Colors.grey),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ],
                          )
                        : Row(
                            children: [
                              Text(
                                "去添加收货地址",
                                style: TextStyle(
                                  fontSize: 15,
                                  color: Color(0xFFFA436A),
                                ),
                              ),
                              const SizedBox(width: 8),
                              Icon(
                                Icons.add,
                                color: Color(0xFFFA436A),
                                size: 20,
                              ),
                            ],
                          ),
                  ),
                  Icon(Icons.chevron_right, color: Colors.grey),
                ],
              ),
            ),
          ),
        ])),
    );
  }

  // 商品标题
  SliverPadding buildProductTitle() {
    final border = BorderSide(
        width: 1, color: const Color(0xFFF5F5F5));
    return SliverPadding(
      padding: EdgeInsets.zero,
      sliver: SliverList(
        delegate: SliverChildListDelegate(<Widget>[
          Container(
            alignment: AlignmentDirectional.centerStart,
            padding: const EdgeInsets.only(left: 15),
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border(bottom: border),
            ),
            height: 42,
            child: const Text("商品信息",
                style: TextStyle(fontSize: 15, color: Color(0xFF606266))),
          ),
        ])),
    );
  }

  // 商品列表
  SliverList buildProductList() {
    final products = _orderData?.cartPromotionItemList ?? [];
    return SliverList.builder(
      itemCount: products.length,
      itemBuilder: (context, index) {
        final item = products[index];
        final finalPrice = item.price - item.reduceAmount;
        return Container(
          color: Colors.white,
          padding: const EdgeInsets.all(15),
          child: Row(
            children: [
              CachedImageWidget(70, 70, item.productPic,
                  fit: BoxFit.cover),
              const SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(item.productName,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontSize: 15)),
                    const SizedBox(height: 6),
                    Text(item.productSubTitle,
                        maxLines: 1,
                        style:
                            const TextStyle(fontSize: 13, color: Colors.grey)),
                    const SizedBox(height: 4),
                    if (item.promotionMessage.isNotEmpty &&
                        item.reduceAmount > 0)
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 6, vertical: 2),
                        margin: const EdgeInsets.only(bottom: 4),
                        decoration: BoxDecoration(
                          color: const Color(0xFFFFF0F0),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(item.promotionMessage,
                            style: const TextStyle(
                                fontSize: 11, color: Color(0xFFFA436A))),
                      ),
                    Row(
                      children: [
                        Text("￥$finalPrice",
                            style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold)),
                        Text(" x${item.quantity}",
                            style: const TextStyle(
                                fontSize: 13, color: Colors.grey)),
                        if (item.reduceAmount > 0) ...[
                          const SizedBox(width: 6),
                          Text("￥${item.price}",
                              style: const TextStyle(
                                  fontSize: 12,
                                  color: Colors.grey,
                                  decoration: TextDecoration.lineThrough)),
                        ],
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  // Task 6: 优惠券选择行
  SliverPadding buildCoupon() {
    final border = BorderSide(
        width: 5, color: const Color(0xFFF5F5F5));
    final enableCoupons =
        _orderData?.couponHistoryDetailList.enableList ?? [];

    return SliverPadding(
      padding: EdgeInsets.zero,
      sliver: SliverList(
        delegate: SliverChildListDelegate(<Widget>[
          Container(
            margin: const EdgeInsets.only(top: 5),
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border(top: border),
            ),
            child: InkWell(
              onTap: _openCouponSheet,
              child: Container(
                height: 45,
                padding: const EdgeInsets.symmetric(horizontal: 15),
                child: Row(
                  children: [
                    Container(
                      width: 20,
                      height: 20,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        color: const Color(0xFFF85E52),
                        borderRadius: BorderRadius.circular(3),
                      ),
                      child: const Text("券",
                          style:
                              TextStyle(fontSize: 12, color: Colors.white)),
                    ),
                    const SizedBox(width: 5),
                    const Expanded(
                        child:
                            Text("优惠券", style: TextStyle(fontSize: 13))),
                    if (_selectedCoupon != null)
                      Text(
                        "-￥${_selectedCoupon!.amount.toStringAsFixed(2)} ${_selectedCoupon!.name}",
                        style: const TextStyle(
                            fontSize: 13, color: Color(0xFFFA436A)),
                      )
                    else
                      Text(
                        enableCoupons.isNotEmpty
                            ? "${enableCoupons.length}张可用"
                            : "暂无可用优惠券",
                        style: const TextStyle(
                            fontSize: 13, color: Color(0xFFFA436A)),
                      ),
                    const SizedBox(width: 4),
                    const Icon(Icons.chevron_right,
                        color: Colors.grey, size: 20),
                  ],
                ),
              ),
            ),
          ),
        ])),
    );
  }

  // Task 7: 积分抵扣行
  SliverPadding buildIntegration() {
    final data = _orderData;
    final setting = data?.integrationConsumeSetting;
    final totalAmt = data?.calcAmount.totalAmount ?? 0;
    int maxIntegration = data?.memberIntegration ?? 0;
    if (setting != null && setting.maxPercentPerOrder > 0 && totalAmt > 0) {
      final maxByPercent = (totalAmt *
              setting.maxPercentPerOrder ~/
              100) *
          setting.deductionPerAmount;
      if (maxByPercent < maxIntegration) maxIntegration = maxByPercent;
    }
    final hasEnoughPoints = (data?.memberIntegration ?? 0) > 0;
    final useUnit = setting?.useUnit ?? 100;

    return SliverPadding(
      padding: EdgeInsets.zero,
      sliver: SliverList(
        delegate: SliverChildListDelegate(<Widget>[
          Container(
            decoration: const BoxDecoration(
              color: Colors.white,
              border: Border(
                  top: BorderSide(width: 1, color: Color(0xFFF5F5F5))),
            ),
            padding: const EdgeInsets.symmetric(horizontal: 15),
            height: 52,
            child: Row(
              children: [
                Container(
                  width: 20,
                  height: 20,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFAA0E),
                    borderRadius: BorderRadius.circular(3),
                  ),
                  child: const Text("积",
                      style: TextStyle(fontSize: 12, color: Colors.white)),
                ),
                const SizedBox(width: 5),
                const Expanded(
                    child: Text("积分抵扣", style: TextStyle(fontSize: 13))),
                if (!hasEnoughPoints)
                  const Text("暂无可用积分",
                      style: TextStyle(fontSize: 13, color: Colors.grey))
                else if (_integrationConflict)
                  const Row(
                    children: [
                      Icon(Icons.warning_amber,
                          color: Colors.orange, size: 16),
                      SizedBox(width: 4),
                      Text("该优惠券不支持与积分共用",
                          style: TextStyle(fontSize: 12, color: Colors.orange)),
                    ],
                  )
                else ...[
                  // Task 7.2: 步进按钮 + 输入框
                  GestureDetector(
                    onTap: () => _onStepperPressed(-1),
                    child: Container(
                      width: 28,
                      height: 24,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey.shade300),
                        borderRadius:
                            const BorderRadius.horizontal(
                                left: Radius.circular(4)),
                      ),
                      child: const Text("—",
                          style: TextStyle(
                              fontSize: 14, fontWeight: FontWeight.bold)),
                    ),
                  ),
                  SizedBox(
                    width: 60,
                    height: 24,
                    child: TextField(
                      controller: _integrationController,
                      keyboardType: TextInputType.number,
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontSize: 13),
                      decoration: InputDecoration(
                        contentPadding: EdgeInsets.zero,
                        isDense: true,
                        border: const OutlineInputBorder(
                          borderRadius: BorderRadius.zero,
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.zero,
                          borderSide:
                              BorderSide(color: Colors.grey.shade300),
                        ),
                        disabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.zero,
                          borderSide:
                              BorderSide(color: Colors.grey.shade200),
                        ),
                        filled: true,
                        fillColor: Colors.grey.shade100,
                      ),
                      onChanged: _onIntegrationChanged,
                    ),
                  ),
                  GestureDetector(
                    onTap: () => _onStepperPressed(1),
                    child: Container(
                      width: 28,
                      height: 24,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey.shade300),
                        borderRadius:
                            const BorderRadius.horizontal(
                                right: Radius.circular(4)),
                      ),
                      child: const Text("+",
                          style: TextStyle(
                              fontSize: 14, fontWeight: FontWeight.bold)),
                    ),
                  ),
                  const SizedBox(width: 4),
                  if (_previewIntegrationAmount > 0)
                    Text(
                      "-￥${(_previewIntegrationAmount / 100).toStringAsFixed(2)}",
                      style: const TextStyle(
                          fontSize: 13, color: Color(0xFFFA436A)),
                    )
                  else
                    Text(
                      "可用${data?.memberIntegration ?? 0}积分（$useUnit 积分起）",
                      style: const TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                ],
              ],
            ),
          ),
        ])),
    );
  }

  // Task 8: 支付方式选择
  SliverPadding buildPayType() {
    final border = BorderSide(
        width: 5, color: const Color(0xFFF5F5F5));
    return SliverPadding(
      padding: EdgeInsets.zero,
      sliver: SliverList(
        delegate: SliverChildListDelegate(<Widget>[
          Container(
            margin: const EdgeInsets.only(top: 5),
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border(top: border),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.only(left: 15, top: 10, bottom: 8),
                  child: const Text("支付方式",
                      style:
                          TextStyle(fontSize: 13, color: Color(0xFF606266))),
                ),
                // Task 8.1: 微信支付
                _PayTypeItem(
                  icon: "images/wx_pay.png",
                  label: "微信支付",
                  desc: "推荐使用",
                  value: 2,
                  groupValue: _selectedPayType,
                  onChanged: (v) => setState(() => _selectedPayType = v),
                ),
                const Divider(height: 1, indent: 15),
                // Task 8.1: 支付宝
                _PayTypeItem(
                  icon: "images/ali_pay.png",
                  label: "支付宝支付",
                  desc: "",
                  value: 1,
                  groupValue: _selectedPayType,
                  onChanged: (v) => setState(() => _selectedPayType = v),
                ),
                const SizedBox(height: 8),
              ],
            ),
          ),
        ])),
    );
  }

  // Task 11: Price Breakdown Card
  SliverPadding buildOrderInfo() {
    final calc = _orderData?.calcAmount;
    final totalAmount = calc?.totalAmount ?? 0;
    final freightAmount = calc?.freightAmount ?? 0;
    final promotionAmount = calc?.promotionAmount ?? 0;
    final couponAmount = _selectedCoupon != null
        ? (_selectedCoupon!.amount * 100).toInt()
        : 0;
    final integrationAmount = _previewIntegrationAmount;

    return SliverPadding(
      padding: const EdgeInsets.only(bottom: 80),
      sliver: SliverList(
        delegate: SliverChildListDelegate(<Widget>[
          Container(
            decoration: const BoxDecoration(
              color: Colors.white,
              border: Border(
                  top: BorderSide(width: 5, color: Color(0xFFF5F5F5))),
            ),
            padding: const EdgeInsets.symmetric(horizontal: 15),
            child: Column(
              children: [
                _AmountRow(
                    label: "商品合计", value: totalAmount, prefix: "￥"),
                const Divider(height: 1),
                _AmountRow(label: "运费", value: freightAmount, prefix: "￥"),
                const Divider(height: 1),
                _AmountRow(
                    label: "活动优惠",
                    value: -promotionAmount,
                    prefix: "-￥",
                    valueColor: const Color(0xFFFA436A)),
                const Divider(height: 1),
                _AmountRow(
                    label: "优惠券",
                    value: couponAmount > 0 ? -couponAmount : 0,
                    prefix: "-￥",
                    valueColor: const Color(0xFFFA436A)),
                const Divider(height: 1),
                _AmountRow(
                    label: "积分抵扣",
                    value: integrationAmount > 0 ? -integrationAmount : 0,
                    prefix: "-￥",
                    valueColor: const Color(0xFFFA436A)),
                const Divider(height: 1),
                SizedBox(
                  height: 45,
                  child: Row(
                    textBaseline: TextBaseline.alphabetic,
                    children: [
                      const Text("备注",
                          style: TextStyle(fontSize: 13, color: Colors.grey)),
                      const SizedBox(width: 10),
                      Expanded(
                        child: TextField(
                          controller: _remarkController,
                          decoration: InputDecoration(
                            hintText: "选填，可填写备注信息",
                            contentPadding: EdgeInsets.zero,
                            isDense: true,
                            border: InputBorder.none,
                            hintStyle: TextStyle(fontSize: 12),
                          ),
                          style: TextStyle(fontSize: 13),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 5),
              ],
            ),
          ),
        ])),
    );
  }

  // Task 10: 底部提交栏（含提交前确认对话框）
  Container buildSubmit() {
    final calc = _orderData?.calcAmount;
    final payAmount = _calcFinalPayAmount();

    return Container(
      padding: const EdgeInsets.only(left: 15),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(
          top: BorderSide(color: Colors.grey, width: 0.5),
        ),
        boxShadow: [
          BoxShadow(
            blurRadius: 4,
            spreadRadius: 1,
            color: Colors.black12,
          ),
        ],
      ),
      height: 60,
      width: MediaQuery.of(context).size.width,
      child: Row(
        children: [
          Expanded(
            flex: 2,
            child: Row(
              children: [
                const Text("实付款 ",
                    style: TextStyle(
                        fontSize: 15, color: Color(0xFF606266))),
                Text("￥",
                    style: TextStyle(
                        fontSize: 15, color: Colors.red.shade400)),
                TweenAnimationBuilder<int>(
                  tween: IntTween(begin: payAmount, end: payAmount),
                  duration: const Duration(milliseconds: 300),
                  builder: (context, val, _) {
                    return Text((val / 100).toStringAsFixed(2),
                        style: TextStyle(
                            fontSize: 18,
                            color: Colors.red.shade400,
                            fontWeight: FontWeight.bold));
                  },
                ),
              ],
            ),
          ),
          Expanded(
            flex: 1,
            child: GestureDetector(
              onTap: _isSubmitting ? null : () => _confirmAndSubmit(payAmount),
              child: Container(
                alignment: Alignment.center,
                height: 60,
                decoration: BoxDecoration(
                  color:
                      _isSubmitting ? Colors.grey : const Color(0xFFFA436A),
                ),
                child: _isSubmitting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor:
                              AlwaysStoppedAnimation(Colors.white),
                        ),
                      )
                    : const Text(
                        '提交订单',
                        style: TextStyle(color: Colors.white, fontSize: 16),
                      ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Task 10.1: 提交前确认对话框
  Future<void> _confirmAndSubmit(int payAmount) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("确认提交订单？"),
        content: Text(
          "提交后将锁定库存和优惠，实付款 ￥${(payAmount / 100).toStringAsFixed(2)}，是否继续？",
          style: const TextStyle(fontSize: 14),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text("取消"),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFFFA436A),
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
    const themeColor = Color(0xFFFA436A);

    return InkWell(
      onTap: () => onChanged(value),
      child: Container(
        height: 48,
        padding: const EdgeInsets.symmetric(horizontal: 15),
        child: Row(
          children: [
            Image.asset(icon, height: 22, width: 22),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(label, style: const TextStyle(fontSize: 14)),
                  if (desc.isNotEmpty)
                    Text(desc,
                        style:
                            const TextStyle(fontSize: 11, color: Colors.grey)),
                ],
              ),
            ),
            Icon(
              selected ? Icons.check_circle : Icons.circle_outlined,
              color: selected ? themeColor : Colors.grey.shade300,
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
  final int value; // 单位：分
  final String prefix;
  final Color? valueColor;

  const _AmountRow({
    required this.label,
    required this.value,
    this.prefix = "￥",
    this.valueColor,
  });

  @override
  Widget build(BuildContext context) {
    final isNegative = value < 0;
    final absVal = value.abs();
    return SizedBox(
      height: 45,
      child: Row(
        children: [
          Expanded(
            child: Text(label,
                style: const TextStyle(fontSize: 13, color: Colors.grey)),
          ),
          Text(
            "$prefix${isNegative ? "-" : ""}${(absVal / 100).toStringAsFixed(2)}",
            style: TextStyle(
                fontSize: 13,
                color: valueColor ?? const Color(0xFF303133)),
          ),
        ],
      ),
    );
  }
}
