// To parse this JSON data, do
//
//     final loginModel = loginModelFromJson(jsonString);

import 'dart:convert';

LoginModel loginModelFromJson(String str) => LoginModel.fromJson(json.decode(str));

String loginModelToJson(LoginModel data) => json.encode(data.toJson());

class LoginModel {
  int code;
  String message;
  LoginData data;

  LoginModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory LoginModel.fromJson(Map<String, dynamic> json) => LoginModel(
        code: json["code"],
        message: json["message"],
        data: LoginData.fromJson(json["data"]),
      );

  Map<String, dynamic> toJson() => {
        "code": code,
        "message": message,
        "data": data.toJson(),
      };
}

class LoginData {
  String tokenHead;
  String token;

  /// Story 3.1.1: 是否为本次调用触发的自动建号
  /// 客户端可据此向新用户展示「欢迎，注册成功」等差异化文案。
  bool isNewUser;

  LoginData({
    required this.tokenHead,
    required this.token,
    this.isNewUser = false,
  });

  factory LoginData.fromJson(Map<String, dynamic> json) => LoginData(
        tokenHead: json["tokenHead"] ?? "Bearer",
        token: json["token"] ?? "",
        isNewUser: json["isNewUser"] == true,
      );

  Map<String, dynamic> toJson() => {
        "tokenHead": tokenHead,
        "token": token,
        "isNewUser": isNewUser,
      };
}
