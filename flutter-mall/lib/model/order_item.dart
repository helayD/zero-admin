import '../utils/spec_formatter.dart';

// OrderItemList 订单商品明细
//
// Story 6-1 Review Fix (HIGH): 合并重复 OrderItemList 类定义
// 此文件作为 OrderItemList 的唯一 canonical 定义，
// 解决 order_list_model.dart 与 order_detail.dart 之间的循环导入问题。
// OrderItemList（canonical，包含 isDeleted 字段）
class OrderItemList {
  int id;
  int orderId;
  String orderNo;
  int orderItemStatus;
  int skuId;
  String skuName;
  String skuPic;
  double skuPrice;
  int skuQuantity;
  String specData;
  double skuTotalAmount;
  double promotionAmount;
  double couponAmount;
  double pointsAmount;
  double discountAmount;
  double realAmount;
  String createTime;
  int isDeleted; // 是否删除：0-正常,1-已删除

  OrderItemList({
    required this.id,
    required this.orderId,
    required this.orderNo,
    required this.orderItemStatus,
    required this.skuId,
    required this.skuName,
    required this.skuPic,
    required this.skuPrice,
    required this.skuQuantity,
    required this.specData,
    required this.skuTotalAmount,
    required this.promotionAmount,
    required this.couponAmount,
    required this.pointsAmount,
    required this.discountAmount,
    required this.realAmount,
    required this.createTime,
    required this.isDeleted,
  });

  factory OrderItemList.fromJson(Map<String, dynamic> json) => OrderItemList(
    id: json["id"] ?? 0,
    orderId: json["orderId"] ?? 0,
    orderNo: json["orderNo"] ?? "",
    orderItemStatus: json["orderItemStatus"] ?? 1,
    skuId: json["skuId"] ?? 0,
    skuName: json["skuName"] ?? "",
    skuPic: json["skuPic"] ?? "",
    skuPrice: (json["skuPrice"] ?? 0).toDouble(),
    skuQuantity: json["skuQuantity"] ?? 1,
    specData: json["specData"] ?? "",
    skuTotalAmount: (json["skuTotalAmount"] ?? 0).toDouble(),
    promotionAmount: (json["promotionAmount"] ?? 0).toDouble(),
    couponAmount: (json["couponAmount"] ?? 0).toDouble(),
    pointsAmount: (json["pointsAmount"] ?? 0).toDouble(),
    discountAmount: (json["discountAmount"] ?? 0).toDouble(),
    realAmount: (json["realAmount"] ?? 0).toDouble(),
    createTime: json["createTime"] ?? "",
    isDeleted: json["isDeleted"] ?? 0,
  );

  Map<String, dynamic> toJson() => {
    "id": id,
    "orderId": orderId,
    "orderNo": orderNo,
    "orderItemStatus": orderItemStatus,
    "skuId": skuId,
    "skuName": skuName,
    "skuPic": skuPic,
    "skuPrice": skuPrice,
    "skuQuantity": skuQuantity,
    "specData": specData,
    "skuTotalAmount": skuTotalAmount,
    "promotionAmount": promotionAmount,
    "couponAmount": couponAmount,
    "pointsAmount": pointsAmount,
    "discountAmount": discountAmount,
    "realAmount": realAmount,
    "createTime": createTime,
    "isDeleted": isDeleted,
  };

  /// Story 10.7：把 specData JSON 字符串解析成「容量: 256GB · 颜色: 金色」的可读格式。
  /// 见 utils/spec_formatter.dart。
  String get formattedSpec => formatSpec(specData);
}
