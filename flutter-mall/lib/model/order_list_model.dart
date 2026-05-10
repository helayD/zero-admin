// To parse this JSON data, do
//
//     final orderListModel = orderListModelFromJson(jsonString);

import 'dart:convert';
import 'order_item.dart'; // OrderItemList canonical definition

OrderListModel orderListModelFromJson(String str) => OrderListModel.fromJson(json.decode(str));

String orderListModelToJson(OrderListModel data) => json.encode(data.toJson());

// OrderListModel 订单列表响应（含分页，Story 6-1 Task 9.1）
class OrderListModel {
  int code;
  String message;
  int pageNum;
  int pageSize;
  int total;
  List<OrderListData> data;

  OrderListModel({
    required this.code,
    required this.message,
    required this.pageNum,
    required this.pageSize,
    required this.total,
    required this.data,
  });

  factory OrderListModel.fromJson(Map<String, dynamic> json) => OrderListModel(
    code: json["code"] ?? 0,
    message: json["message"] ?? "",
    pageNum: json["pageNum"] ?? 1,
    pageSize: json["pageSize"] ?? 10,
    total: json["total"] ?? 0,
    data: json["data"] != null
        ? List<OrderListData>.from(json["data"].map((x) => OrderListData.fromJson(x)))
        : [],
  );

  Map<String, dynamic> toJson() => {
    "code": code,
    "message": message,
    "pageNum": pageNum,
    "pageSize": pageSize,
    "total": total,
    "data": List<dynamic>.from(data.map((x) => x.toJson())),
  };

  bool get hasMore => data.length < total;
}

// OrderListData 订单列表项（Story 6-1 Task 9.1 增强）
class OrderListData {
  int id;
  String orderNo;        // 订单编号
  String merchantName;    // 商户名称（Story 6-1 新增）
  int status;            // OMS订单状态：0=待支付,1=已支付/待发货,2=已取消,3=已完成,4=售后中
  int payStatus;         // 支付状态：0=未支付,1=已支付
  double totalAmount;    // 订单总金额
  double payAmount;       // 实付金额
  String thumbnail;      // 首商品缩略图
  String createTime;     // 下单时间
  List<OrderItemList> orderItemData;
  // Story 10.7：订单是否含提货卡，用于列表卡片显示「含提货卡」徽标
  bool hasDigitalCards;

  OrderListData({
    required this.id,
    required this.orderNo,
    required this.merchantName,
    required this.status,
    required this.payStatus,
    required this.totalAmount,
    required this.payAmount,
    required this.thumbnail,
    required this.createTime,
    required this.orderItemData,
    this.hasDigitalCards = false,
  });

  factory OrderListData.fromJson(Map<String, dynamic> json) => OrderListData(
    id: json["id"] ?? 0,
    orderNo: json["orderNo"] ?? "",
    merchantName: json["merchantName"] ?? "",
    status: json["orderStatus"] ?? 0,
    payStatus: json["payStatus"] ?? 0,
    totalAmount: (json["totalAmount"] ?? 0).toDouble(),
    payAmount: (json["payAmount"] ?? 0).toDouble(),
    thumbnail: json["thumbnail"] ?? "",
    createTime: json["createTime"] ?? "",
    orderItemData: json["orderItemData"] != null
        ? List<OrderItemList>.from(json["orderItemData"].map((x) => OrderItemList.fromJson(x)))
        : [],
    hasDigitalCards: json["hasDigitalCards"] == true,
  );

  Map<String, dynamic> toJson() => {
    "id": id,
    "orderNo": orderNo,
    "merchantName": merchantName,
    "orderStatus": status,
    "payStatus": payStatus,
    "totalAmount": totalAmount,
    "payAmount": payAmount,
    "thumbnail": thumbnail,
    "createTime": createTime,
    "orderItemData": List<dynamic>.from(orderItemData.map((x) => x.toJson())),
    "hasDigitalCards": hasDigitalCards,
  };

  /// 兼容旧字段名
  int get orderStatus => status;
  double get orderAmount => totalAmount;
  String get orderSn => orderNo;
}
