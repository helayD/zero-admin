// To parse this JSON data, do
//
//     final orderDetailModel = orderDetailModelFromJson(jsonString);

import 'dart:convert';
import 'order_item.dart'; // OrderItemList canonical definition

OrderDetailModel orderDetailModelFromJson(String str) => OrderDetailModel.fromJson(json.decode(str));

String orderDetailModelToJson(OrderDetailModel data) => json.encode(data.toJson());

// OrderDetailModel 订单详情响应（Story 6-1 Task 9.2）
class OrderDetailModel {
  int code;
  String message;
  OrderDetailData data;

  OrderDetailModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory OrderDetailModel.fromJson(Map<String, dynamic> json) => OrderDetailModel(
    code: json["code"] ?? 0,
    message: json["message"] ?? "",
    data: json["data"] != null
        ? OrderDetailData.fromJson(json["data"])
        : OrderDetailData.empty(),
  );

  Map<String, dynamic> toJson() => {
    "code": code,
    "message": message,
    "data": data.toJson(),
  };
}

// OrderDetailData 订单详情数据（Story 6-1 Task 9.2 增强）
class OrderDetailData {
  int id;
  String orderNo;
  int userId;
  int orderStatus;      // OMS订单状态：0=待支付,1=已支付/待发货,2=已取消,3=已完成,4=售后中
  int payStatus;        // 支付状态：0=未支付,1=已支付
  int deliveryStatus;   // 发货状态：0=未发货,1=已发货,2=已收货
  int aftersaleStatus; // 售后状态：0=无售后,1=售后申请中,2=售后完成
  double totalAmount;
  double promotionAmount;
  double couponAmount;
  double pointsAmount;
  double discountAmount;
  double freightAmount;
  double payAmount;
  int payType;
  String payTime;
  String deliveryTime;
  String receiveTime;
  String confirmTime;
  String finishTime;
  String commentTime;
  int sourceType;
  String expressOrderNumber;
  String logisticsCompany;
  int usePoints;
  int receiveStatus;
  String remark;
  String createTime;
  String updateTime;
  List<OrderItemList> orderItemData;
  MemberReceiveAddress memberReceiveAddress;
  // Story 6-1 新增字段
  List<TimelineNode> timeline;
  PriceBreakdown priceBreakdown;

  OrderDetailData({
    required this.id,
    required this.orderNo,
    required this.userId,
    required this.orderStatus,
    required this.payStatus,
    required this.deliveryStatus,
    required this.aftersaleStatus,
    required this.totalAmount,
    required this.promotionAmount,
    required this.couponAmount,
    required this.pointsAmount,
    required this.discountAmount,
    required this.freightAmount,
    required this.payAmount,
    required this.payType,
    required this.payTime,
    required this.deliveryTime,
    required this.receiveTime,
    required this.confirmTime,
    required this.finishTime,
    required this.commentTime,
    required this.sourceType,
    required this.expressOrderNumber,
    required this.logisticsCompany,
    required this.usePoints,
    required this.receiveStatus,
    required this.remark,
    required this.createTime,
    required this.updateTime,
    required this.orderItemData,
    required this.memberReceiveAddress,
    required this.timeline,
    required this.priceBreakdown,
  });

  /// 空数据兜底
  factory OrderDetailData.empty() => OrderDetailData(
    id: 0,
    orderNo: "",
    userId: 0,
    orderStatus: 0,
    payStatus: 0,
    deliveryStatus: 0,
    aftersaleStatus: 0,
    totalAmount: 0,
    promotionAmount: 0,
    couponAmount: 0,
    pointsAmount: 0,
    discountAmount: 0,
    freightAmount: 0,
    payAmount: 0,
    payType: 0,
    payTime: "",
    deliveryTime: "",
    receiveTime: "",
    confirmTime: "",
    finishTime: "",
    commentTime: "",
    sourceType: 0,
    expressOrderNumber: "",
    logisticsCompany: "",
    usePoints: 0,
    receiveStatus: 0,
    remark: "",
    createTime: "",
    updateTime: "",
    orderItemData: [],
    memberReceiveAddress: MemberReceiveAddress.empty(),
    timeline: [],
    priceBreakdown: PriceBreakdown.empty(),
  );

  factory OrderDetailData.fromJson(Map<String, dynamic> json) => OrderDetailData(
    id: json["id"] ?? 0,
    orderNo: json["orderNo"] ?? "",
    userId: json["userId"] ?? 0,
    orderStatus: json["orderStatus"] ?? 0,
    payStatus: json["payStatus"] ?? 0,
    deliveryStatus: json["deliveryStatus"] ?? 0,
    aftersaleStatus: json["aftersaleStatus"] ?? 0,
    totalAmount: (json["totalAmount"] ?? 0).toDouble(),
    promotionAmount: (json["promotionAmount"] ?? 0).toDouble(),
    couponAmount: (json["couponAmount"] ?? 0).toDouble(),
    pointsAmount: (json["pointsAmount"] ?? 0).toDouble(),
    discountAmount: (json["discountAmount"] ?? 0).toDouble(),
    freightAmount: (json["freightAmount"] ?? 0).toDouble(),
    payAmount: (json["payAmount"] ?? 0).toDouble(),
    payType: json["payType"] ?? 0,
    payTime: json["payTime"] ?? "",
    deliveryTime: json["deliveryTime"] ?? "",
    receiveTime: json["receiveTime"] ?? "",
    confirmTime: json["confirmTime"] ?? "",
    finishTime: json["finishTime"] ?? "",
    commentTime: json["commentTime"] ?? "",
    sourceType: json["sourceType"] ?? 0,
    expressOrderNumber: json["expressOrderNumber"] ?? "",
    logisticsCompany: json["logisticsCompany"] ?? "",
    usePoints: json["usePoints"] ?? 0,
    receiveStatus: json["receiveStatus"] ?? 0,
    remark: json["remark"] ?? "",
    createTime: json["createTime"] ?? "",
    updateTime: json["updateTime"] ?? "",
    orderItemData: json["orderItemData"] != null
        ? List<OrderItemList>.from(json["orderItemData"].map((x) => OrderItemList.fromJson(x)))
        : [],
    memberReceiveAddress: json["memberReceiveAddress"] != null
        ? MemberReceiveAddress.fromJson(json["memberReceiveAddress"])
        : MemberReceiveAddress.empty(),
    timeline: json["timeline"] != null
        ? List<TimelineNode>.from(json["timeline"].map((x) => TimelineNode.fromJson(x)))
        : [],
    priceBreakdown: json["priceBreakdown"] != null
        ? PriceBreakdown.fromJson(json["priceBreakdown"])
        : PriceBreakdown.empty(),
  );

  Map<String, dynamic> toJson() => {
    "id": id,
    "orderNo": orderNo,
    "userId": userId,
    "orderStatus": orderStatus,
    "payStatus": payStatus,
    "deliveryStatus": deliveryStatus,
    "aftersaleStatus": aftersaleStatus,
    "totalAmount": totalAmount,
    "promotionAmount": promotionAmount,
    "couponAmount": couponAmount,
    "pointsAmount": pointsAmount,
    "discountAmount": discountAmount,
    "freightAmount": freightAmount,
    "payAmount": payAmount,
    "payType": payType,
    "payTime": payTime,
    "deliveryTime": deliveryTime,
    "receiveTime": receiveTime,
    "confirmTime": confirmTime,
    "finishTime": finishTime,
    "commentTime": commentTime,
    "sourceType": sourceType,
    "expressOrderNumber": expressOrderNumber,
    "logisticsCompany": logisticsCompany,
    "usePoints": usePoints,
    "receiveStatus": receiveStatus,
    "remark": remark,
    "createTime": createTime,
    "updateTime": updateTime,
    "orderItemData": List<dynamic>.from(orderItemData.map((x) => x.toJson())),
    "memberReceiveAddress": memberReceiveAddress.toJson(),
    "timeline": List<dynamic>.from(timeline.map((x) => x.toJson())),
    "priceBreakdown": priceBreakdown.toJson(),
  };

  /// 兼容旧字段名
  String get orderNoStr => orderNo;
}

// TimelineNode 时间线节点（Story 6-1 Task 9.2）
// status: completed/current/pending/interrupted
class TimelineNode {
  String status;
  String title;
  String time;
  String detail;

  TimelineNode({
    required this.status,
    required this.title,
    required this.time,
    required this.detail,
  });

  factory TimelineNode.fromJson(Map<String, dynamic> json) => TimelineNode(
    status: json["status"] ?? "pending",
    title: json["title"] ?? "",
    time: json["time"] ?? "",
    detail: json["detail"] ?? "",
  );

  Map<String, dynamic> toJson() => {
    "status": status,
    "title": title,
    "time": time,
    "detail": detail,
  };

  bool get isCompleted => status == "completed";
  bool get isCurrent => status == "current";
  bool get isPending => status == "pending";
  bool get isInterrupted => status == "interrupted";
}

// PriceBreakdown 金额拆分（Story 6-1 Task 9.2）
class PriceBreakdown {
  double orderAmount;
  double freightAmount;
  double promotionAmount;
  double couponAmount;
  double pointsAmount;
  double discountAmount;
  double payAmount;

  PriceBreakdown({
    required this.orderAmount,
    required this.freightAmount,
    required this.promotionAmount,
    required this.couponAmount,
    required this.pointsAmount,
    required this.discountAmount,
    required this.payAmount,
  });

  factory PriceBreakdown.empty() => PriceBreakdown(
    orderAmount: 0,
    freightAmount: 0,
    promotionAmount: 0,
    couponAmount: 0,
    pointsAmount: 0,
    discountAmount: 0,
    payAmount: 0,
  );

  factory PriceBreakdown.fromJson(Map<String, dynamic> json) => PriceBreakdown(
    orderAmount: (json["orderAmount"] ?? 0).toDouble(),
    freightAmount: (json["freightAmount"] ?? 0).toDouble(),
    promotionAmount: (json["promotionAmount"] ?? 0).toDouble(),
    couponAmount: (json["couponAmount"] ?? 0).toDouble(),
    pointsAmount: (json["pointsAmount"] ?? 0).toDouble(),
    discountAmount: (json["discountAmount"] ?? 0).toDouble(),
    payAmount: (json["payAmount"] ?? 0).toDouble(),
  );

  Map<String, dynamic> toJson() => {
    "orderAmount": orderAmount,
    "freightAmount": freightAmount,
    "promotionAmount": promotionAmount,
    "couponAmount": couponAmount,
    "pointsAmount": pointsAmount,
    "discountAmount": discountAmount,
    "payAmount": payAmount,
  };
}

// MemberReceiveAddress 收货地址（Story 6-1 Task 9.2）
class MemberReceiveAddress {
  int id;
  int memberId;
  String receiverName;
  String receiverPhone;
  String province;
  String city;
  String district;
  String detailAddress;
  String postalCode;
  String tag;
  int isDefault;

  MemberReceiveAddress({
    required this.id,
    required this.memberId,
    required this.receiverName,
    required this.receiverPhone,
    required this.province,
    required this.city,
    required this.district,
    required this.detailAddress,
    required this.postalCode,
    required this.tag,
    required this.isDefault,
  });

  factory MemberReceiveAddress.empty() => MemberReceiveAddress(
    id: 0,
    memberId: 0,
    receiverName: "",
    receiverPhone: "",
    province: "",
    city: "",
    district: "",
    detailAddress: "",
    postalCode: "",
    tag: "",
    isDefault: 0,
  );

  factory MemberReceiveAddress.fromJson(Map<String, dynamic> json) => MemberReceiveAddress(
    id: json["id"] ?? 0,
    memberId: json["memberId"] ?? 0,
    receiverName: json["receiverName"] ?? "",
    receiverPhone: json["receiverPhone"] ?? "",
    province: json["province"] ?? "",
    city: json["city"] ?? "",
    district: json["district"] ?? "",
    detailAddress: json["detailAddress"] ?? "",
    postalCode: json["postalCode"] ?? "",
    tag: json["tag"] ?? "",
    isDefault: json["isDefault"] ?? 0,
  );

  Map<String, dynamic> toJson() => {
    "id": id,
    "memberId": memberId,
    "receiverName": receiverName,
    "receiverPhone": receiverPhone,
    "province": province,
    "city": city,
    "district": district,
    "detailAddress": detailAddress,
    "postalCode": postalCode,
    "tag": tag,
    "isDefault": isDefault,
  };

  String get fullAddress => "$province $city $district $detailAddress";

  /// 手机号脱敏（138****5678）
  /// 后端已脱敏，若字段已包含 **** 直接返回；否则按原始手机号脱敏
  String get maskedPhone {
    if (receiverPhone.contains('****')) return receiverPhone;
    if (receiverPhone.length < 7) return receiverPhone;
    return '${receiverPhone.substring(0, 3)}****${receiverPhone.substring(receiverPhone.length - 4)}';
  }
}
