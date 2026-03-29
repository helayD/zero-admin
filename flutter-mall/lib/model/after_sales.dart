// 售后数据模型（Story 6-4 Task 8）
// 包含售后原因列表、售后申请请求/响应、售后类型枚举

// ==================== 售后原因列表 ====================

// ReturnReasonItem 售后原因项（Story 6-4 Task 8）
class ReturnReasonItemData {
  final int id;
  final String name;

  ReturnReasonItemData({
    required this.id,
    required this.name,
  });

  factory ReturnReasonItemData.fromJson(Map<String, dynamic> json) =>
      ReturnReasonItemData(
        id: json["id"] ?? 0,
        name: json["name"] ?? "",
      );

  Map<String, dynamic> toJson() => {
        "id": id,
        "name": name,
      };
}

// ReturnReasonListData 售后原因列表响应（Story 6-4 Task 8）
class ReturnReasonListData {
  final int code;
  final String message;
  final List<ReturnReasonItemData> reasonList;

  ReturnReasonListData({
    required this.code,
    required this.message,
    required this.reasonList,
  });

  factory ReturnReasonListData.fromJson(Map<String, dynamic> json) =>
      ReturnReasonListData(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        reasonList: json["reasonList"] != null
            ? (json["reasonList"] as List)
                .map((e) => ReturnReasonItemData.fromJson(e))
                .toList()
            : [],
      );
}

// ==================== 售后申请提交 ====================

// ApplyAfterSalesReqData 售后申请请求数据（Story 6-4 Task 8）
class ApplyAfterSalesReqData {
  final int orderId;
  final int type; // 0=退货退款, 1=仅退款, 2=换货
  final int reasonId;
  final String description;
  final String proofPics; // base64，逗号分隔

  ApplyAfterSalesReqData({
    required this.orderId,
    required this.type,
    required this.reasonId,
    this.description = "",
    this.proofPics = "",
  });

  Map<String, dynamic> toJson() => {
        "orderId": orderId,
        "type": type,
        "reasonId": reasonId,
        "description": description,
        "proofPics": proofPics,
      };
}

// ApplyAfterSalesRespData 售后申请响应数据（Story 6-4 Task 8）
class ApplyAfterSalesRespData {
  final int code;
  final String message;
  final int returnId;
  final String returnNo;

  ApplyAfterSalesRespData({
    required this.code,
    required this.message,
    required this.returnId,
    required this.returnNo,
  });

  factory ApplyAfterSalesRespData.fromJson(Map<String, dynamic> json) =>
      ApplyAfterSalesRespData(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        returnId: json["returnId"] ?? 0,
        returnNo: json["returnNo"] ?? "",
      );

  bool get isSuccess => code == 0;
}

// 售后类型枚举（Story 6-4 Task 8）
enum AfterSalesType {
  returnRefund(0, "退货退款"),
  refundOnly(1, "仅退款"),
  exchange(2, "换货");

  final int value;
  final String label;

  const AfterSalesType(this.value, this.label);

  static AfterSalesType fromValue(int value) {
    return AfterSalesType.values.firstWhere(
      (e) => e.value == value,
      orElse: () => AfterSalesType.returnRefund,
    );
  }
}
