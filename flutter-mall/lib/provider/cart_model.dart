import 'package:flutter/foundation.dart';

import '../model/cart_list.dart';
import '../model/cart_promotion.dart';

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
      final promo = promotionMap[value.productId];
      if (promo != null) {
        allProductPrice += (promo.price - promo.reduceAmount) * promo.quantity;
      } else {
        allProductPrice += value.price * value.quantity;
      }
    });

    return allProductPrice;
  }

  /// Makes `cartListData` readable inside the devtools by listing all of its properties
  @override
  void debugFillProperties(DiagnosticPropertiesBuilder properties) {
    super.debugFillProperties(properties);
    properties.add(DiagnosticsProperty<Map<int, CartData>>('allCartProduct', allCartProduct));
    properties.add(DiagnosticsProperty<Map<int, CartData>>('checkCartProduct', checkCartProduct));
    properties.add(IntProperty('allProductPrice', allProductPrice));
  }
}
