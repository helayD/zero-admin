import 'package:flutter/foundation.dart';

import '../model/cart_list.dart';
import '../model/cart_promotion.dart';
import '../model/cart_validate.dart';

///
/// 购物车的状态
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class CartModel with ChangeNotifier, DiagnosticableTreeMixin {
  // 购物车所有商品（key 为购物车记录 id）
  Map<int, CartData> allCartProduct = <int, CartData>{};

  // 购物车选中的商品（key 为购物车记录 id）
  Map<int, CartData> checkCartProduct = <int, CartData>{};

  // 促销信息（按商品ID索引）
  Map<int, CartPromotionData> promotionMap = <int, CartPromotionData>{};

  // 无效购物车项（key: cartItemId, value: errorMessage）
  Map<int, String> invalidCartItems = <int, String>{};

  // 选择商品的价格总和
  int allProductPrice = 0;

  // 初始化购物车商品数据到状态管理中
  void setCartListData(List<CartData> cartListData) {
    for (var cart in cartListData) {
      allCartProduct[cart.id] = cart;
      checkCartProduct[cart.id] = cart;
    }
    notifyListeners();
  }

  // 设置商品选择是否选择的状态
  void setCartItemStatus(int cartItemId) {
    if (checkCartProduct[cartItemId] != null) {
      checkCartProduct.remove(cartItemId);
    } else {
      checkCartProduct[cartItemId] = allCartProduct[cartItemId]!;
    }
    notifyListeners();
  }

  // 点击全选或者反选
  void setAllStatus() {
    if (checkCartProduct.length == allCartProduct.length) {
      checkCartProduct.clear();
    } else {
      checkCartProduct = Map.from(allCartProduct);
    }
    notifyListeners();
  }

  // 获取全选或者反选状态显示
  bool getAllStatus() {
    return checkCartProduct.length == allCartProduct.length;
  }

  // 判断购物车中商品前面的图标是否被选择
  bool getProductIsCheck(int cartItemId) {
    return checkCartProduct[cartItemId] != null;
  }

  // 获取所有商品
  List<CartData> getAllProduct() {
    List<CartData> list = [];
    allCartProduct.forEach((key, value) {
      list.add(value);
    });
    return list;
  }

  // 获取选择中的商品
  List<CartData> getCheckProduct() {
    List<CartData> list = [];
    checkCartProduct.forEach((key, value) {
      list.add(value);
    });
    return list;
  }

  // 设置促销信息
  void setPromotionData(List<CartPromotionData> promotionList) {
    promotionMap.clear();
    for (var item in promotionList) {
      promotionMap[item.productId] = item;
    }
    notifyListeners();
  }

  // 获取商品的促销信息
  CartPromotionData? getPromotion(int productId) {
    return promotionMap[productId];
  }

  // 计算所选中的商品的价格（扣除促销优惠）
  int getProductAllPrice() {
    allProductPrice = 0;
    checkCartProduct.forEach((key, value) {
      // 跳过无效商品
      if (invalidCartItems.containsKey(key)) return;
      final promo = promotionMap[value.productId];
      if (promo != null) {
        allProductPrice += (promo.price - promo.reduceAmount) * promo.quantity;
      } else {
        allProductPrice += value.price * value.quantity;
      }
    });

    return allProductPrice;
  }

  // 获取有效（未被标记为无效）的已选商品
  List<CartData> getValidCheckProduct() {
    return checkCartProduct.values
        .where((item) => !invalidCartItems.containsKey(item.id))
        .toList();
  }

  // 设置校验结果（标记无效商品）
  void setValidationResults(List<CartValidateResult> results) {
    invalidCartItems.clear();
    for (var result in results) {
      if (!result.valid) {
        invalidCartItems[result.id] = result.errorMessage;
      }
    }
    notifyListeners();
  }

  // 清除无效标记
  void clearInvalidItems() {
    invalidCartItems.clear();
    notifyListeners();
  }

  // 获取商品无效原因
  String? getInvalidReason(int cartItemId) {
    return invalidCartItems[cartItemId];
  }

  // 判断商品是否无效
  bool isInvalid(int cartItemId) {
    return invalidCartItems.containsKey(cartItemId);
  }

  // 清空所有购物车数据（清空购物车成功后调用）
  void clearAll() {
    allCartProduct.clear();
    checkCartProduct.clear();
    promotionMap.clear();
    invalidCartItems.clear();
    allProductPrice = 0;
    notifyListeners();
  }

  // 删除单个商品（删除成功后调用）
  void removeItem(int cartItemId) {
    allCartProduct.remove(cartItemId);
    checkCartProduct.remove(cartItemId);
    invalidCartItems.remove(cartItemId);
    notifyListeners();
  }

  // 更新单个商品数量
  void updateQuantity(int cartItemId, int newQuantity) {
    if (allCartProduct[cartItemId] != null) {
      final updated = allCartProduct[cartItemId]!;
      allCartProduct[cartItemId] = CartData(
        id: updated.id,
        memberId: updated.memberId,
        productId: updated.productId,
        productSkuId: updated.productSkuId,
        quantity: newQuantity,
        price: updated.price,
        selected: updated.selected,
        productName: updated.productName,
        productSubTitle: updated.productSubTitle,
        productPic: updated.productPic,
        productSkuCode: updated.productSkuCode,
        productSn: updated.productSn,
        productBrand: updated.productBrand,
        productCategoryId: updated.productCategoryId,
        productAttr: updated.productAttr,
        memberNickname: updated.memberNickname,
        source: updated.source,
        expireTime: updated.expireTime,
        createTime: updated.createTime,
        updateTime: updated.updateTime,
      );
      if (checkCartProduct.containsKey(cartItemId)) {
        checkCartProduct[cartItemId] = allCartProduct[cartItemId]!;
      }
      notifyListeners();
    }
  }

  /// Makes `cartListData` readable inside the devtools by listing all of its properties
  @override
  void debugFillProperties(DiagnosticPropertiesBuilder properties) {
    super.debugFillProperties(properties);
    properties.add(DiagnosticsProperty<Map<int, CartData>>('allCartProduct', allCartProduct));
    properties.add(DiagnosticsProperty<Map<int, CartData>>('checkCartProduct', checkCartProduct));
    properties.add(DiagnosticsProperty<Map<int, CartPromotionData>>('promotionMap', promotionMap));
    properties.add(DiagnosticsProperty<Map<int, String>>('invalidCartItems', invalidCartItems));
    properties.add(IntProperty('allProductPrice', allProductPrice));
  }
}
