import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/cart_promotion.dart';
import 'package:flutter_mall/model/cart_validate.dart';
import 'package:flutter_mall/provider/cart_model.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/spec_formatter.dart';
import 'package:provider/provider.dart';

import '../../layout/main_tab.dart';
import '../../model/cart_list.dart';
import '../category/product/product_detail.dart';
import '../mine/order/order_submit.dart';

///
/// 购物车页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Cart extends StatefulWidget {
  final String? intentSource;

  const Cart({super.key, this.intentSource});

  @override
  State<Cart> createState() => _CartState();
}

class _CartState extends State<Cart> {
  bool _isSubmitting = false;
  bool _cartLoadedSuccessfully = false;

  AppRecentContext _buildCartRecoveryContext([String source = 'manual_open']) {
    return AppRecentContext.create(
      targetType: AppRecentTargetType.cart,
      tabIndex: 2,
      source: widget.intentSource ?? source,
      requiresAuth: true,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );
  }

  @override
  void initState() {
    super.initState();
    AppRecoveryStore.saveActiveIntentCandidate(_buildCartRecoveryContext());
    _queryBrandListData();
  }

  @override
  void dispose() {
    AppRecoveryStore.clearActiveIntentCandidateIfMatches(
      AppRecentTargetType.cart,
      tabIndex: 2,
    );
    super.dispose();
  }

  void _queryBrandListData() async {
    try {
      Response result = await HttpUtil.get(cartDataUrl);
      if (!mounted) return;
      setState(() {
        CartListModel cartListModel = CartListModel.fromJson(result.data);
        context.read<CartModel>().setCartListData(cartListModel.data);
        _cartLoadedSuccessfully = true;
      });
      _queryPromotionData();
    } catch (e) {
      _cartLoadedSuccessfully = false;
    }
  }

  void _queryPromotionData() async {
    try {
      Response result = await HttpUtil.post(cartPromotionUrl, data: {});
      CartPromotionListModel model =
          CartPromotionListModel.fromJson(result.data);
      if (mounted) {
        context.read<CartModel>().setPromotionData(model.data);
        if (_cartLoadedSuccessfully) {
          await AppRecoveryStore.saveRecentContext(
            _buildCartRecoveryContext(),
          );
        }
      }
    } catch (e) {
      debugPrint("获取促销信息失败: $e");
    }
  }

  // 防抖 Timer（Task 11.1 防重复提交）
  Timer? _deleteDebounce;

  // 删除单个商品（Task 6）
  Future<void> _deleteItem(int cartItemId) async {
    if (_deleteDebounce?.isActive ?? false) return;
    _deleteDebounce = Timer(const Duration(milliseconds: 500), () {});

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("确认删除"),
        content: const Text("确认删除该商品？"),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text("取消"),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text("确认"),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() => _isSubmitting = true);
    try {
      Response resp = await HttpUtil.post(deleteCartUrl, data: {
        "ids": [cartItemId]
      });
      final respData = resp.data as Map<String, dynamic>;
      if (respData["code"] == 0) {
        if (mounted) {
          context.read<CartModel>().removeItem(cartItemId);
          _queryPromotionData();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text("删除成功"),
              backgroundColor: Colors.green,
            ),
          );
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(respData["message"] ?? "删除失败"),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        String msg = "删除失败，请稍后重试";
        if (e.response?.data != null && e.response!.data is Map) {
          msg = e.response!.data["message"] ?? msg;
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  // 清空购物车（Task 8）
  Future<void> _clearCart() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("确认清空"),
        content: const Text("确认清空购物车？清空后无法恢复。"),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text("取消"),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text("确认清空"),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() => _isSubmitting = true);
    try {
      Response resp = await HttpUtil.get(clearCartUrl);
      final respData = resp.data as Map<String, dynamic>;
      if (respData["code"] == 0) {
        if (mounted) {
          context.read<CartModel>().clearAll();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text("清空成功"),
              backgroundColor: Colors.green,
            ),
          );
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(respData["message"] ?? "清空失败"),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        String msg = "清空失败，请稍后重试";
        if (e.response?.data != null && e.response!.data is Map) {
          msg = e.response!.data["message"] ?? msg;
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  // 修改商品数量（Task 7）
  Future<void> _updateQuantity(
      int cartItemId, int currentQty, int maxStock) async {
    final controller = TextEditingController(text: currentQty.toString());

    final newQty = await showDialog<int>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("修改数量"),
        content: StatefulBuilder(
          builder: (ctx, setDialogState) => Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              IconButton(
                onPressed: controller.text.isNotEmpty &&
                        int.tryParse(controller.text) != null &&
                        int.parse(controller.text) > 1
                    ? () {
                        final v = int.tryParse(controller.text) ?? 1;
                        if (v > 1) {
                          setDialogState(
                              () => controller.text = (v - 1).toString());
                        }
                      }
                    : null,
                icon: const Icon(Icons.remove_circle_outline),
              ),
              SizedBox(
                width: 60,
                child: TextField(
                  controller: controller,
                  keyboardType: TextInputType.number,
                  textAlign: TextAlign.center,
                  decoration: const InputDecoration(
                    contentPadding:
                        EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    isDense: true,
                    border: OutlineInputBorder(),
                  ),
                  onChanged: (v) {
                    final val = int.tryParse(v);
                    if (val != null && val > maxStock) {
                      controller.text = maxStock.toString();
                      controller.selection = TextSelection.fromPosition(
                        TextPosition(offset: controller.text.length),
                      );
                    }
                    setDialogState(() {});
                  },
                ),
              ),
              IconButton(
                onPressed: int.tryParse(controller.text) != null &&
                        int.parse(controller.text) < maxStock
                    ? () {
                        final v = int.tryParse(controller.text) ?? 1;
                        if (v < maxStock) {
                          setDialogState(
                              () => controller.text = (v + 1).toString());
                        }
                      }
                    : null,
                icon: const Icon(Icons.add_circle_outline),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text("取消"),
          ),
          TextButton(
            onPressed: () {
              final qty = int.tryParse(controller.text) ?? currentQty;
              Navigator.pop(ctx, qty);
            },
            child: const Text("确认"),
          ),
        ],
      ),
    );

    if (newQty == null || newQty == currentQty) return;
    await _submitQuantity(cartItemId, newQty);
  }

  Future<void> _submitQuantity(int cartItemId, int newQty) async {
    setState(() => _isSubmitting = true);
    try {
      Response resp = await HttpUtil.post(
        updateCartQuantityUrl,
        data: {"id": cartItemId, "quantity": newQty},
      );
      final respData = resp.data as Map<String, dynamic>;
      if (respData["code"] == 0) {
        if (mounted) {
          context.read<CartModel>().updateQuantity(cartItemId, newQty);
          _queryPromotionData();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text("数量修改成功"),
              backgroundColor: Colors.green,
            ),
          );
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(respData["message"] ?? "修改失败"),
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        String msg = "修改失败，请稍后重试";
        if (e.response?.data != null && e.response!.data is Map) {
          msg = e.response!.data["message"] ?? msg;
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  Future<void> _quickChangeQuantity(
    CartData item,
    CartPromotionData? promo,
    int delta,
  ) async {
    if (_isSubmitting) return;
    final maxStock = _maxStock(promo);
    final nextQty = (item.quantity + delta).clamp(1, maxStock).toInt();
    if (nextQty == item.quantity) return;
    await _submitQuantity(item.id, nextQty);
  }

  // 批量结算前商品有效性校验（Task 9 & AC#2）
  Future<bool> _validateBeforeCheckout() async {
    final cartModel = context.read<CartModel>();
    final checkItems = cartModel.getCheckProduct();
    if (checkItems.isEmpty) return false;

    try {
      final ids = checkItems.map((e) => e.id).toList();
      Response resp =
          await HttpUtil.post(validateCartItemsUrl, data: {"ids": ids});
      final respData = resp.data as Map<String, dynamic>;
      if (respData["code"] == 0) {
        final validateModel = CartValidateModel.fromJson(respData);
        cartModel.setValidationResults(validateModel.data);
        // 检查是否有无效商品
        final invalidIds =
            validateModel.data.where((r) => !r.valid).map((r) => r.id).toSet();

        if (invalidIds.isNotEmpty) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text("部分商品已下架/库存不足，请移除后再结算"),
                backgroundColor: Colors.orange,
              ),
            );
          }
          return false;
        }
        return true;
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(respData["message"] ?? "校验失败"),
              backgroundColor: Colors.red,
            ),
          );
        }
        return false;
      }
    } on DioException catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text("校验失败，请稍后重试"),
            backgroundColor: Colors.red,
          ),
        );
      }
      return false;
    }
  }

  // 去结算（Task 9）
  void _goToCheckout() async {
    if (_isSubmitting) return;

    // 校验选中的商品
    final valid = await _validateBeforeCheckout();
    if (!valid) return;

    if (!mounted) return;

    // 再次检查有效商品是否为空（可能全被标记无效）
    final validItems = context.read<CartModel>().getValidCheckProduct();
    if (validItems.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text("请先选择有效商品"),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => const OrderSubmit()),
    );
  }

  int _estimateSelectedSavings(CartModel cartModel) {
    int total = 0;
    for (final item in cartModel.getValidCheckProduct()) {
      final promo = cartModel.getPromotion(item.id);
      if (promo != null && promo.reduceAmount > 0) {
        total += promo.reduceAmount * item.quantity;
      }
    }
    return total;
  }

  int _selectedQuantity(CartModel cartModel) {
    return cartModel
        .getValidCheckProduct()
        .fold<int>(0, (sum, item) => sum + item.quantity);
  }

  String _formatPrice(int amountInYuan) {
    return amountInYuan.toString();
  }

  void _openProductDetail(CartData item) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ProductDetail(
          productId: item.productId,
          intentSource: 'cart_item_tap',
        ),
      ),
    );
  }

  int _maxStock(CartPromotionData? promo) {
    if (promo != null && promo.realStock > 0) {
      return promo.realStock;
    }
    return 999;
  }

  String _formatShopName(String raw) {
    final name = raw.trim();
    if (name.isEmpty || name.toLowerCase() == "test") {
      return "九克城自营";
    }
    if (name.contains("旗舰店") || name.contains("自营")) {
      return name;
    }
    return "$name 官方旗舰店";
  }

  String _shopKeyForItem(CartData item) {
    final rawBrand = item.productBrand.trim();
    if (rawBrand.isNotEmpty && rawBrand.toLowerCase() != "test") {
      return rawBrand;
    }

    final name = item.productName.toLowerCase();
    if (name.contains("荣耀")) {
      return "荣耀";
    }
    if (name.contains("三星") || name.contains("samsung")) {
      return "三星";
    }
    if (name.contains("华为")) {
      return "华为";
    }
    if (name.contains("小米") || name.contains("redmi")) {
      return "小米";
    }
    if (name.contains("apple") ||
        name.contains("苹果") ||
        name.contains("iphone")) {
      return "Apple";
    }
    return "九克城自营";
  }

  List<_CartShopGroup> _groupCartItems(List<CartData> cartListData) {
    final groupMap = <String, List<CartData>>{};
    for (final item in cartListData) {
      final key = _shopKeyForItem(item);
      groupMap.putIfAbsent(key, () => <CartData>[]).add(item);
    }
    return groupMap.entries
        .map((entry) => _CartShopGroup(
              name: _formatShopName(entry.key),
              items: entry.value,
            ))
        .toList();
  }

  bool _isGroupSelected(CartModel cartModel, List<CartData> items) {
    return items.isNotEmpty &&
        items.every((item) => cartModel.getProductIsCheck(item.id));
  }

  void _toggleGroupSelection(CartModel cartModel, List<CartData> items) {
    final shouldUnselect = _isGroupSelected(cartModel, items);
    for (final item in items) {
      final isSelected = cartModel.getProductIsCheck(item.id);
      if ((shouldUnselect && isSelected) || (!shouldUnselect && !isSelected)) {
        cartModel.setCartItemStatus(item.id);
      }
    }
  }

  Widget _buildCheckbox({
    required bool selected,
    required VoidCallback onTap,
    double size = 22,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(size),
      child: Image.asset(
        selected
            ? "images/checkbox_round_1.png"
            : "images/checkbox_round_2.png",
        height: size,
        width: size,
      ),
    );
  }

  Widget _buildShopSection(
    BuildContext context,
    CartModel cartModel,
    _CartShopGroup group,
    Map<int, String> invalidItems,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 13, 14, 6),
            child: Row(
              children: [
                _buildCheckbox(
                  selected: _isGroupSelected(cartModel, group.items),
                  onTap: () => _toggleGroupSelection(cartModel, group.items),
                  size: 21,
                ),
                const SizedBox(width: 9),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 5, vertical: 2),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFA436A),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: const Text(
                    "店",
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
                    ),
                  ),
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    group.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                      color: Color(0xFF1F2329),
                    ),
                  ),
                ),
              ],
            ),
          ),
          for (final item in group.items) ...[
            _buildCartItemRow(
              context,
              cartModel,
              item,
              cartModel.getPromotion(item.id),
              invalidItems.containsKey(item.id),
              invalidItems[item.id],
            ),
            if (item != group.items.last)
              const Divider(
                height: 1,
                indent: 48,
                endIndent: 14,
                color: Color(0xFFF0F1F2),
              ),
          ],
        ],
      ),
    );
  }

  Widget _buildCartItemRow(
    BuildContext context,
    CartModel cartModel,
    CartData item,
    CartPromotionData? promo,
    bool isInvalid,
    String? invalidReason,
  ) {
    final currentPrice = promo != null && promo.reduceAmount > 0
        ? promo.price - promo.reduceAmount
        : item.price;
    final originalPrice =
        promo != null && promo.reduceAmount > 0 ? promo.price : item.price;

    return Opacity(
      opacity: isInvalid ? 0.55 : 1,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(14, 10, 14, 14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.only(top: 39),
              child: _buildCheckbox(
                selected: cartModel.getProductIsCheck(item.id),
                onTap: () => cartModel.setCartItemStatus(item.id),
                size: 22,
              ),
            ),
            const SizedBox(width: 10),
            InkWell(
              onTap: () => _openProductDetail(item),
              borderRadius: BorderRadius.circular(12),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: Image.network(
                  kIsWeb ? proxyImageUrl(item.productPic) : item.productPic,
                  width: 92,
                  height: 92,
                  fit: BoxFit.cover,
                  errorBuilder: (ctx, err, stack) => Container(
                    width: 92,
                    height: 92,
                    color: const Color(0xFFF0F1F2),
                    child: const Icon(
                      Icons.image_not_supported_outlined,
                      color: Color(0xFFB0B4BC),
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: GestureDetector(
                          behavior: HitTestBehavior.opaque,
                          onTap: () => _openProductDetail(item),
                          child: Text(
                            item.productName,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              fontSize: 15.5,
                              height: 1.35,
                              fontWeight: FontWeight.w700,
                              color: Color(0xFF1F2329),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      InkWell(
                        onTap:
                            _isSubmitting ? null : () => _deleteItem(item.id),
                        child: const Icon(
                          Icons.close,
                          size: 18,
                          color: Color(0xFF909399),
                        ),
                      ),
                    ],
                  ),
                  if (item.productSubTitle.isNotEmpty) ...[
                    const SizedBox(height: 6),
                    Text(
                      item.productSubTitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF909399),
                      ),
                    ),
                  ],
                  if (item.productAttr.isNotEmpty) ...[
                    const SizedBox(height: 7),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF5F5F5),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        formatSpec(item.productAttr),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF606266),
                        ),
                      ),
                    ),
                  ],
                  const SizedBox(height: 8),
                  _buildPromotionTags(promo),
                  if (promo != null &&
                      promo.realStock > 0 &&
                      promo.realStock <= 10) ...[
                    const SizedBox(height: 7),
                    Text(
                      "库存紧张，仅剩${promo.realStock}件",
                      style: const TextStyle(
                        fontSize: 12,
                        color: Colors.orange,
                      ),
                    ),
                  ],
                  if (isInvalid && invalidReason != null) ...[
                    const SizedBox(height: 8),
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 7,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFFFF4F4),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        invalidReason,
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFFFA436A),
                        ),
                      ),
                    ),
                  ],
                  const SizedBox(height: 12),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Text(
                                  "¥${_formatPrice(currentPrice)}",
                                  style: const TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.w700,
                                    color: Color(0xFFFA436A),
                                  ),
                                ),
                                if (promo != null &&
                                    promo.reduceAmount > 0) ...[
                                  const SizedBox(width: 6),
                                  Text(
                                    "¥${_formatPrice(originalPrice)}",
                                    style: const TextStyle(
                                      fontSize: 12,
                                      color: Color(0xFFB0B4BC),
                                      decoration: TextDecoration.lineThrough,
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ],
                        ),
                      ),
                      _buildQuantityStepper(item, promo, isInvalid),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPromotionTags(CartPromotionData? promo) {
    final hasPromo = promo != null && promo.promotionMessage.trim().isNotEmpty;
    return Wrap(
      spacing: 8,
      runSpacing: 6,
      children: [
        _buildTag(hasPromo ? promo.promotionMessage.trim() : "无优惠"),
        const Text(
          "7天价保",
          style: TextStyle(
            fontSize: 12,
            color: Color(0xFF909399),
          ),
        ),
      ],
    );
  }

  Widget _buildTag(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1F4),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: const TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w600,
          color: Color(0xFFFA436A),
        ),
      ),
    );
  }

  Widget _buildQuantityStepper(
    CartData item,
    CartPromotionData? promo,
    bool isInvalid,
  ) {
    final maxStock = _maxStock(promo);
    final canMinus = !isInvalid && !_isSubmitting && item.quantity > 1;
    final canPlus = !isInvalid && !_isSubmitting && item.quantity < maxStock;

    return Container(
      height: 34,
      decoration: BoxDecoration(
        color: const Color(0xFFF7F7F7),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildQuantityButton(
            icon: Icons.remove_rounded,
            enabled: canMinus,
            onTap: () => _quickChangeQuantity(item, promo, -1),
          ),
          GestureDetector(
            onTap: isInvalid
                ? null
                : () => _updateQuantity(item.id, item.quantity, maxStock),
            child: Container(
              width: 38,
              alignment: Alignment.center,
              child: Text(
                item.quantity.toString(),
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1F2329),
                ),
              ),
            ),
          ),
          _buildQuantityButton(
            icon: Icons.add_rounded,
            enabled: canPlus,
            onTap: () => _quickChangeQuantity(item, promo, 1),
          ),
        ],
      ),
    );
  }

  Widget _buildQuantityButton({
    required IconData icon,
    required bool enabled,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: enabled ? onTap : null,
      borderRadius: BorderRadius.circular(8),
      child: SizedBox(
        width: 30,
        height: 34,
        child: Icon(
          icon,
          size: 18,
          color: enabled ? const Color(0xFF303133) : const Color(0xFFC8CDD4),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 108,
              height: 108,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF1F4),
                borderRadius: BorderRadius.circular(32),
              ),
              child: const Icon(
                Icons.shopping_bag_outlined,
                size: 46,
                color: Color(0xFFFA436A),
              ),
            ),
            const SizedBox(height: 18),
            const Text(
              "购物车还是空的",
              style: TextStyle(
                fontSize: 22,
                fontWeight: FontWeight.w700,
                color: Color(0xFF303133),
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              "去首页挑几件好商品吧，支持加购后统一结算，也支持从详情页直接下单。",
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 14,
                height: 1.6,
                color: Color(0xFF909399),
              ),
            ),
            const SizedBox(height: 22),
            ElevatedButton(
              onPressed: () {
                Navigator.of(context).pushAndRemoveUntil(
                  MaterialPageRoute(builder: (_) => const MainTab()),
                  (route) => false,
                );
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFFFA436A),
                foregroundColor: Colors.white,
                padding:
                    const EdgeInsets.symmetric(horizontal: 28, vertical: 14),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(24),
                ),
              ),
              child: const Text("去首页逛逛"),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCheckoutBar(CartModel cartModel) {
    final selectedKinds = cartModel.getValidCheckProduct().length;
    final selectedQuantity = _selectedQuantity(cartModel);
    final totalPrice = cartModel.getProductAllPrice();
    final estimatedSavings = _estimateSelectedSavings(cartModel);

    return Positioned(
      bottom: 0,
      width: MediaQuery.of(context).size.width,
      child: SafeArea(
        top: false,
        child: Container(
          padding: const EdgeInsets.fromLTRB(16, 10, 12, 10),
          decoration: BoxDecoration(
            color: Colors.white,
            boxShadow: [
              BoxShadow(
                blurRadius: 12,
                offset: const Offset(0, -4),
                color: Colors.black.withAlpha(12),
              ),
            ],
          ),
          child: Row(
            children: [
              InkWell(
                onTap: () => context.read<CartModel>().setAllStatus(),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Image.asset(
                      Provider.of<CartModel>(context, listen: true)
                              .getAllStatus()
                          ? "images/checkbox_round_1.png"
                          : "images/checkbox_round_2.png",
                      height: 22,
                      width: 22,
                    ),
                    const SizedBox(height: 3),
                    const Text(
                      "全选",
                      style: TextStyle(
                        fontSize: 13,
                        color: Color(0xFF303133),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    FittedBox(
                      fit: BoxFit.scaleDown,
                      alignment: Alignment.centerRight,
                      child: Text(
                        "¥${_formatPrice(totalPrice)}",
                        maxLines: 1,
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFFFA436A),
                        ),
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      estimatedSavings > 0
                          ? "已优惠 ¥${_formatPrice(estimatedSavings)}"
                          : "已选$selectedKinds款/$selectedQuantity件",
                      style: const TextStyle(
                        fontSize: 11,
                        color: Color(0xFF909399),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              SizedBox(
                height: 50,
                child: ElevatedButton(
                  onPressed: _isSubmitting ? null : _goToCheckout,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFA436A),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(horizontal: 28),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                    elevation: 0,
                  ),
                  child: const Text(
                    "去结算",
                    style: TextStyle(
                      fontSize: 17,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final cartModel = context.watch<CartModel>();
    List<CartData> cartListData = cartModel.getAllProduct();
    final invalidItems = cartModel.invalidCartItems;
    final shopGroups = _groupCartItems(cartListData);

    return Scaffold(
      appBar: AppBar(
        backgroundColor: const Color(0xFFF5F5F5),
        elevation: 0,
        toolbarHeight: 70,
        titleSpacing: 18,
        title: Row(
          children: [
            Text(
              "购物车 (${cartListData.length})",
              style: const TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.w800,
                color: Colors.black,
              ),
            ),
          ],
        ),
        actions: [
          if (cartListData.isNotEmpty)
            TextButton(
              onPressed: _isSubmitting ? null : _clearCart,
              child: const Text(
                "清空",
                style: TextStyle(
                  color: Color(0xFFFA436A),
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
        ],
      ),
      body: Stack(
        children: [
          Container(
            color: const Color(0xFFF5F5F5),
          ),
          cartListData.isNotEmpty
              ? ListView(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 112),
                  children: [
                    ...shopGroups.map((group) {
                      return _buildShopSection(
                        context,
                        cartModel,
                        group,
                        invalidItems,
                      );
                    }),
                  ],
                )
              : _buildEmptyState(),
          if (cartListData.isNotEmpty) _buildCheckoutBar(cartModel),
          if (_isSubmitting)
            Container(
              color: Colors.black26,
              child: const Center(child: CircularProgressIndicator()),
            ),
        ],
      ),
    );
  }
}

class _CartShopGroup {
  final String name;
  final List<CartData> items;

  const _CartShopGroup({
    required this.name,
    required this.items,
  });
}
