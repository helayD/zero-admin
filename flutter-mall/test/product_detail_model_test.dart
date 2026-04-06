import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/product_detail.dart';

void main() {
  SkuStockList buildSku(dynamic specData) {
    return SkuStockList.fromJson({
      "id": 1,
      "spuId": 1,
      "name": "测试规格",
      "skuCode": "sku-1",
      "mainPic": "",
      "albumPics": "",
      "price": 100,
      "promotionPrice": 90,
      "promotionStartTime": "",
      "promotionEndTime": "",
      "stock": 10,
      "lowStock": 1,
      "specData": specData,
      "weight": 0.19,
      "publishStatus": 1,
      "verifyStatus": 1,
      "sort": 1,
      "sales": 0,
      "purchasable": true,
      "purchaseReasonCode": "",
    });
  }

  test('SkuStockList 可解析对象格式规格数据', () {
    final sku = buildSku('{"容量":"128GB","颜色":"金色","网络版本":"全网通版"}');

    expect(sku.parsedSpecData, {
      "容量": "128GB",
      "颜色": "金色",
      "网络版本": "全网通版",
    });
  });

  test('SkuStockList 兼容历史数组格式规格数据', () {
    final sku = buildSku(
      '[{"key":"颜色","value":"黑色"},{"key":"尺码","value":"L"}]',
    );

    expect(sku.parsedSpecData, {
      "颜色": "黑色",
      "尺码": "L",
    });
  });

  test('SkuStockList fromJson 可归一化对象型 specData', () {
    final sku = buildSku({
      "颜色": "银色",
      "容量": "256GB",
    });

    expect(sku.parsedSpecData, {
      "颜色": "银色",
      "容量": "256GB",
    });
  });
}
