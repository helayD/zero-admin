import 'package:flutter_mall/model/cart_list.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  Map<String, dynamic> buildCartItem(dynamic productAttr) {
    return {
      "id": 121,
      "memberId": 1001,
      "productId": 4,
      "productSkuId": 40,
      "quantity": 1,
      "price": 8999,
      "selected": 1,
      "productName": "测试商品",
      "productSubTitle": "测试副标题",
      "productPic": "http://example.com/test.png",
      "productSkuCode": "SKU-001",
      "productSn": "SN-001",
      "productBrand": "测试品牌",
      "productCategoryId": 11,
      "productAttr": productAttr,
      "memberNickname": "test",
      "source": 4,
      "expireTime": "2026-06-05 11:11:58",
      "createTime": "2026-04-06 11:11:58",
      "updateTime": "2026-04-06 11:11:58",
    };
  }

  test('CartData 兼容对象格式 productAttr', () {
    final cart = CartData.fromJson(
      buildCartItem({
        "容量": "256GB",
        "颜色": "暗紫色",
        "网络版本": "全网通版",
      }),
    );

    expect(cart.productAttr, '容量: 256GB · 颜色: 暗紫色 · 网络版本: 全网通版');
  });

  test('CartData 兼容数组格式 productAttr', () {
    final cart = CartData.fromJson(
      buildCartItem([
        {"颜色": "金色"},
        {"容量": "128GB"},
      ]),
    );

    expect(cart.productAttr, '颜色: 金色 · 容量: 128GB');
  });

  test('CartData 兼容字符串 JSON 格式 productAttr', () {
    final cart = CartData.fromJson(
      buildCartItem('{"颜色":"银色","容量":"512GB"}'),
    );

    expect(cart.productAttr, '颜色: 银色 · 容量: 512GB');
  });

  test('CartData 兼容空数组 productAttr', () {
    final cart = CartData.fromJson(buildCartItem([]));

    expect(cart.productAttr, isEmpty);
  });

  test('CartData 兼容空字符串 updateTime', () {
    final cart = CartData.fromJson(
      buildCartItem('[]')
        ..['createTime'] = '2026-04-06 11:11:58'
        ..['updateTime'] = '',
    );

    expect(cart.updateTime, DateTime.parse('2026-04-06 11:11:58'));
  });
}
