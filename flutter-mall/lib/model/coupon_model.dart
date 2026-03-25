// To parse this JSON data, do
//
//     final couponModel = couponModelFromJson(jsonString);

import 'dart:convert';

CouponModel couponModelFromJson(String str) => CouponModel.fromJson(json.decode(str));

String couponModelToJson(CouponModel data) => json.encode(data.toJson());

class CouponModel {
  int code;
  String message;
  List<CouponData> data;

  CouponModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory CouponModel.fromJson(Map<String, dynamic> json) => CouponModel(
    code: json["code"],
    message: json["message"],
    data: List<CouponData>.from(json["data"].map((x) => CouponData.fromJson(x))),
  );

  Map<String, dynamic> toJson() => {
    "code": code,
    "message": message,
    "data": List<dynamic>.from(data.map((x) => x.toJson())),
  };
}

class CouponData {
  int id;
  int typeId;
  String name;
  String code;
  num amount;
  num minAmount;
  String startTime;
  String endTime;
  int perLimit;
  int status;
  String description;
  int scopeType;
  int receiveStatus;
  int totalCount;
  int receivedCount;

  CouponData({
    required this.id,
    required this.typeId,
    required this.name,
    required this.code,
    required this.amount,
    required this.minAmount,
    required this.startTime,
    required this.endTime,
    required this.perLimit,
    required this.status,
    required this.description,
    required this.scopeType,
    this.receiveStatus = 0,
    this.totalCount = 0,
    this.receivedCount = 0,
  });

  factory CouponData.fromJson(Map<String, dynamic> json) => CouponData(
    id: json["id"],
    typeId: json["typeId"],
    name: json["name"],
    code: json["code"],
    amount: json["amount"] ?? 0,
    minAmount: json["minAmount"] ?? 0,
    startTime: json["startTime"],
    endTime: json["endTime"],
    perLimit: json["perLimit"],
    status: json["status"],
    description: json["description"],
    scopeType: json["scopeType"],
    receiveStatus: json["receiveStatus"] ?? 0,
    totalCount: json["totalCount"] ?? 0,
    receivedCount: json["receivedCount"] ?? 0,
  );

  Map<String, dynamic> toJson() => {
    "id": id,
    "typeId": typeId,
    "name": name,
    "code": code,
    "amount": amount,
    "minAmount": minAmount,
    "startTime": startTime,
    "endTime": endTime,
    "perLimit": perLimit,
    "status": status,
    "description": description,
    "scopeType": scopeType,
    "receiveStatus": receiveStatus,
    "totalCount": totalCount,
    "receivedCount": receivedCount,
  };
}
