import 'package:flutter_mall/model/cart_list.dart';
import 'package:flutter_mall/model/cart_promotion.dart';
import 'package:flutter_mall/provider/cart_model.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  CartData buildCartItem({
    required int id,
    required int productId,
    required int productSkuId,
    required int quantity,
    required int price,
  }) {
    final now = DateTime.parse('2026-04-06 12:00:00');
    return CartData(
      id: id,
      memberId: 1001,
      productId: productId,
      productSkuId: productSkuId,
      quantity: quantity,
      price: price,
      selected: 1,
      productName: '测试商品$productSkuId',
      productSubTitle: '',
      productPic: 'http://example.com/test.png',
      productSkuCode: 'SKU-$productSkuId',
      productSn: 'SN-$productSkuId',
      productBrand: '测试品牌',
      productCategoryId: 11,
      productAttr: '容量: 256GB',
      memberNickname: 'tester',
      source: 4,
      expireTime: now,
      createTime: now,
      updateTime: now,
    );
  }

  CartPromotionData buildPromotion({
    required int id,
    required int productId,
    required int productSkuId,
    required int quantity,
    required int price,
    required int reduceAmount,
  }) {
    return CartPromotionData(
      id: id,
      memberId: 1001,
      productId: productId,
      productSkuId: productSkuId,
      quantity: quantity,
      price: price,
      selected: 1,
      productName: '测试商品$productSkuId',
      productSubTitle: '',
      productPic: 'http://example.com/test.png',
      productSkuCode: 'SKU-$productSkuId',
      productSn: 'SN-$productSkuId',
      productBrand: '测试品牌',
      productCategoryId: 11,
      productAttr: '容量: 256GB',
      memberNickname: 'tester',
      source: 4,
      deleteStatus: 0,
      expireTime: '',
      createTime: '',
      updateTime: '',
      promotionMessage: reduceAmount > 0 ? '直降' : '',
      reduceAmount: reduceAmount,
      realStock: 99,
      integration: 0,
      growth: 0,
    );
  }

  test('CartModel 按购物车项 id 匹配促销价格，避免同商品不同规格串价', () {
    final model = CartModel();
    model.setCartListData([
      buildCartItem(
          id: 100, productId: 1, productSkuId: 41, quantity: 1, price: 7999),
      buildCartItem(
          id: 101, productId: 1, productSkuId: 42, quantity: 1, price: 8999),
    ]);
    model.setPromotionData([
      buildPromotion(
        id: 100,
        productId: 1,
        productSkuId: 41,
        quantity: 1,
        price: 7999,
        reduceAmount: 100,
      ),
      buildPromotion(
        id: 101,
        productId: 1,
        productSkuId: 42,
        quantity: 1,
        price: 8999,
        reduceAmount: 0,
      ),
    ]);

    expect(model.getPromotion(100)?.price, 7999);
    expect(model.getPromotion(101)?.price, 8999);
    expect(model.getProductAllPrice(), 16898);
  });

  test('CartModel 汇总金额使用当前购物车数量，而不是旧促销快照数量', () {
    final model = CartModel();
    model.setCartListData([
      buildCartItem(
          id: 200, productId: 2, productSkuId: 51, quantity: 2, price: 8999),
    ]);
    model.setPromotionData([
      buildPromotion(
        id: 200,
        productId: 2,
        productSkuId: 51,
        quantity: 1,
        price: 8999,
        reduceAmount: 100,
      ),
    ]);

    expect(model.getProductAllPrice(), 17798);
  });
}
