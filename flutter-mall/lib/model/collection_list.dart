// To parse this JSON data, do
//
//     final collectionListModel = collectionListModelFromJson(jsonString);

import 'dart:convert';

CollectionListModel collectionListModelFromJson(String str) =>
    CollectionListModel.fromJson(json.decode(str));

String collectionListModelToJson(CollectionListModel data) =>
    json.encode(data.toJson());

class CollectionListModel {
  int code;
  String message;
  List<CollectionListData> data;

  CollectionListModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory CollectionListModel.fromJson(Map<String, dynamic> json) =>
      CollectionListModel(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        data: List<CollectionListData>.from(
          (json["data"] ?? []).map((x) => CollectionListData.fromJson(x)),
        ),
      );

  Map<String, dynamic> toJson() => {
        "code": code,
        "message": message,
        "data": List<dynamic>.from(data.map((x) => x.toJson())),
      };
}

class CollectionListData {
  String id;
  int memberId;
  String memberNickName;
  String memberIcon;
  int productId;
  String productName;
  String productPic;
  String productSubTitle;
  int productPrice;
  DateTime createTime;

  CollectionListData({
    required this.id,
    required this.memberId,
    required this.memberNickName,
    required this.memberIcon,
    required this.productId,
    required this.productName,
    required this.productPic,
    required this.productSubTitle,
    required this.productPrice,
    required this.createTime,
  });

  factory CollectionListData.fromJson(Map<String, dynamic> json) =>
      CollectionListData(
        id: json["id"] ?? "",
        memberId: json["memberId"] ?? 0,
        memberNickName: json["memberNickName"] ?? "",
        memberIcon: json["memberIcon"] ?? "",
        productId: json["productId"] ?? 0,
        productName: json["productName"] ?? "",
        productPic: json["productPic"] ?? "",
        productSubTitle: json["productSubTitle"] ?? "",
        productPrice: _parseInt(json["productPrice"]),
        createTime: DateTime.tryParse(json["createTime"] ?? "") ??
            DateTime.fromMillisecondsSinceEpoch(0),
      );

  Map<String, dynamic> toJson() => {
        "id": id,
        "memberId": memberId,
        "memberNickName": memberNickName,
        "memberIcon": memberIcon,
        "productId": productId,
        "productName": productName,
        "productPic": productPic,
        "productSubTitle": productSubTitle,
        "productPrice": productPrice,
        "createTime": createTime.toIso8601String(),
      };
}

int _parseInt(dynamic value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse(value?.toString() ?? "") ?? 0;
}
