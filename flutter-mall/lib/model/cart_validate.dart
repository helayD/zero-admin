/// 批量结算前商品有效性校验响应模型
/// Story 5-2 Task 3 & AC#2
class CartValidateModel {
  int code;
  String message;
  List<CartValidateResult> data;

  CartValidateModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory CartValidateModel.fromJson(Map<String, dynamic> json) =>
      CartValidateModel(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        data: json["data"] == null
            ? []
            : List<CartValidateResult>.from(
                json["data"].map((x) => CartValidateResult.fromJson(x)),
              ),
      );
}

/// 单项商品校验结果
class CartValidateResult {
  int id; // 购物车项ID
  bool valid; // 是否有效
  String errorCode; // 错误码
  String errorMessage; // 错误信息

  CartValidateResult({
    required this.id,
    required this.valid,
    required this.errorCode,
    required this.errorMessage,
  });

  factory CartValidateResult.fromJson(Map<String, dynamic> json) =>
      CartValidateResult(
        id: json["id"] ?? 0,
        valid: json["valid"] ?? false,
        errorCode: json["errorCode"] ?? "",
        errorMessage: json["errorMessage"] ?? "",
      );
}
