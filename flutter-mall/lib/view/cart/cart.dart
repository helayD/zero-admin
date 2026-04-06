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
import 'package:provider/provider.dart';

import '../../model/cart_list.dart';
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

  var boxDecoration = BoxDecoration(
    color: Colors.white,
    border: Border(
      bottom: BorderSide(
        width: 1,
        color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
      ),
    ),
  );

  @override
  Widget build(BuildContext context) {
    final cartModel = context.watch<CartModel>();
    List<CartData> cartListData = cartModel.getAllProduct();
    final invalidItems = cartModel.invalidCartItems;

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("购物车"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        actions: [
          if (cartListData.isNotEmpty)
            TextButton(
              onPressed: _isSubmitting ? null : _clearCart,
              child: const Text("清空", style: TextStyle(color: Colors.red)),
            ),
        ],
      ),
      body: Stack(
        children: [
          if (_isSubmitting)
            Container(
              color: Colors.black26,
              child: const Center(child: CircularProgressIndicator()),
            ),
          cartListData.isNotEmpty
              ? ListView.builder(
                  itemCount: cartListData.length,
                  itemBuilder: (BuildContext context, int index) {
                    final item = cartListData[index];
                    final promo = cartModel.getPromotion(item.productId);
                    final isInvalid = invalidItems.containsKey(item.id);
                    final invalidReason = invalidItems[item.id];

                    return Opacity(
                      opacity: isInvalid ? 0.5 : 1.0,
                      child: Container(
                        decoration: boxDecoration,
                        padding: const EdgeInsets.all(15),
                        child: Stack(
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                // 勾选框
                                InkWell(
                                  onTap: () {
                                    cartModel.setCartItemStatus(item.id);
                                  },
                                  child: Image.asset(
                                    cartModel.getProductIsCheck(item.id)
                                        ? "images/checkbox_round_1.png"
                                        : "images/checkbox_round_2.png",
                                    height: 23,
                                    width: 22,
                                  ),
                                ),
                                const SizedBox(width: 10),
                                // 商品图片
                                ClipRRect(
                                  borderRadius: const BorderRadius.all(
                                      Radius.circular(20)),
                                  child: Image.network(
                                    kIsWeb
                                        ? proxyImageUrl(item.productPic)
                                        : item.productPic,
                                    width: 80,
                                    height: 80,
                                    fit: BoxFit.cover,
                                    errorBuilder: (ctx, err, stack) =>
                                        Container(
                                      width: 80,
                                      height: 80,
                                      color: Colors.grey[200],
                                      child:
                                          const Icon(Icons.image_not_supported),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 10),
                                // 商品信息
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        item.productName,
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                        style: TextStyle(
                                          fontSize: 14,
                                          color: Color(
                                            int.parse('303133', radix: 16),
                                          ).withAlpha(255),
                                          decoration: isInvalid
                                              ? TextDecoration.lineThrough
                                              : null,
                                        ),
                                      ),
                                      Padding(
                                        padding: const EdgeInsets.symmetric(
                                            vertical: 2),
                                        child: Text(
                                          item.productAttr,
                                          style: TextStyle(
                                            fontSize: 12,
                                            color: Color(
                                              int.parse('909399', radix: 16),
                                            ).withAlpha(255),
                                          ),
                                        ),
                                      ),
                                      // 促销标签
                                      if (promo != null &&
                                          promo.reduceAmount > 0) ...[
                                        Container(
                                          padding: const EdgeInsets.symmetric(
                                              horizontal: 6, vertical: 2),
                                          margin:
                                              const EdgeInsets.only(bottom: 2),
                                          decoration: BoxDecoration(
                                            color: const Color(0xFFFFF0F0),
                                            borderRadius:
                                                BorderRadius.circular(4),
                                          ),
                                          child: Text(
                                            promo.promotionMessage,
                                            style: const TextStyle(
                                              fontSize: 11,
                                              color: Color(0xFFFA436A),
                                            ),
                                          ),
                                        ),
                                        Row(
                                          children: [
                                            Text(
                                              "¥${(promo.price - promo.reduceAmount) ~/ 100}",
                                              style: TextStyle(
                                                fontSize: 14,
                                                color: Color(
                                                  int.parse('303133',
                                                      radix: 16),
                                                ).withAlpha(255),
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                            const SizedBox(width: 4),
                                            Text(
                                              "¥${promo.price ~/ 100}",
                                              style: const TextStyle(
                                                fontSize: 11,
                                                color: Colors.grey,
                                                decoration:
                                                    TextDecoration.lineThrough,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ] else
                                        Text(
                                          "¥${item.price ~/ 100}",
                                          style: TextStyle(
                                            fontSize: 14,
                                            color: Color(
                                              int.parse('303133', radix: 16),
                                            ).withAlpha(255),
                                            fontWeight: FontWeight.bold,
                                          ),
                                        ),
                                      // 库存提示
                                      if (promo != null &&
                                          promo.realStock > 0 &&
                                          promo.realStock <= 10)
                                        Padding(
                                          padding:
                                              const EdgeInsets.only(top: 2),
                                          child: Text(
                                            "仅剩${promo.realStock}件",
                                            style: const TextStyle(
                                              fontSize: 11,
                                              color: Colors.orange,
                                            ),
                                          ),
                                        ),
                                      // 数量显示（可点击修改，失效商品不可改，Task 10.3）
                                      GestureDetector(
                                        onTap: isInvalid
                                            ? null
                                            : () => _updateQuantity(
                                                  item.id,
                                                  item.quantity,
                                                  promo?.realStock ?? 999,
                                                ),
                                        child: Container(
                                          margin: const EdgeInsets.only(top: 4),
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 8,
                                            vertical: 2,
                                          ),
                                          decoration: BoxDecoration(
                                            border: Border.all(
                                              color: Colors.grey[300]!,
                                            ),
                                            borderRadius:
                                                BorderRadius.circular(4),
                                          ),
                                          child: Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Text(
                                                "×${item.quantity}",
                                                style: const TextStyle(
                                                    fontSize: 13),
                                              ),
                                              const SizedBox(width: 4),
                                              const Icon(Icons.edit, size: 12),
                                            ],
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                // 删除按钮（Task 6.1）
                                GestureDetector(
                                  onTap: _isSubmitting
                                      ? null
                                      : () => _deleteItem(item.id),
                                  child: Image.asset(
                                    "images/close.png",
                                    height: 16,
                                    width: 17,
                                    color: Color(
                                      int.parse('909399', radix: 16),
                                    ).withAlpha(255),
                                  ),
                                ),
                              ],
                            ),
                            // 无效商品角标（Task 9.2 & Task 10.2）
                            if (isInvalid && invalidReason != null)
                              Positioned(
                                left: 0,
                                top: 0,
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 6,
                                    vertical: 2,
                                  ),
                                  decoration: const BoxDecoration(
                                    color: Colors.red,
                                    borderRadius: BorderRadius.only(
                                      topLeft: Radius.circular(4),
                                      bottomRight: Radius.circular(4),
                                    ),
                                  ),
                                  child: Text(
                                    invalidReason,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 10,
                                    ),
                                  ),
                                ),
                              ),
                          ],
                        ),
                      ),
                    );
                  },
                )
              : Container(
                  alignment: Alignment.topCenter,
                  margin: const EdgeInsets.only(top: 200),
                  child: Text(
                    "囧~ 购物车还是空的",
                    style: TextStyle(
                      fontSize: 20,
                      color:
                          Color(int.parse('909399', radix: 16)).withAlpha(255),
                    ),
                  ),
                ),
          // 底部结算栏
          Visibility(
            visible: cartListData.isNotEmpty,
            child: Positioned(
              bottom: 0,
              width: MediaQuery.of(context).size.width,
              child: Container(
                height: 60,
                padding: const EdgeInsets.symmetric(horizontal: 15),
                margin: const EdgeInsets.all(15),
                decoration: BoxDecoration(
                  color: Colors.white,
                  border: Border.all(color: Colors.grey, width: 1),
                  boxShadow: const [
                    BoxShadow(
                      blurRadius: 2,
                      spreadRadius: 1,
                      color: Colors.grey,
                    ),
                  ],
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    InkWell(
                      onTap: () {
                        context.read<CartModel>().setAllStatus();
                      },
                      child: Image.asset(
                        Provider.of<CartModel>(context, listen: true)
                                .getAllStatus()
                            ? "images/checkbox_round_1.png"
                            : "images/checkbox_round_2.png",
                        height: 30,
                        width: 30,
                      ),
                    ),
                    Row(
                      children: [
                        Text(
                          "￥${Provider.of<CartModel>(context, listen: true).getProductAllPrice() ~/ 100}",
                          style: const TextStyle(fontSize: 16),
                        ),
                        const SizedBox(width: 25),
                        Container(
                          height: 40,
                          width: 90,
                          decoration: BoxDecoration(
                            color: Color(int.parse('fa436a', radix: 16))
                                .withAlpha(255),
                            borderRadius:
                                const BorderRadius.all(Radius.circular(50)),
                            boxShadow: const [
                              BoxShadow(
                                blurRadius: 2,
                                spreadRadius: 1,
                                color: Colors.grey,
                              ),
                            ],
                          ),
                          child: TextButton(
                            onPressed: _isSubmitting ? null : _goToCheckout,
                            child: const Text(
                              '去结算',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 15,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
