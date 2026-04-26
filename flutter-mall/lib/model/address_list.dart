// To parse this JSON data, do
//
//     final addressListModel = addressListModelFromJson(jsonString);

import 'dart:convert';

AddressListModel addressListModelFromJson(String str) =>
    AddressListModel.fromJson(json.decode(str));

String addressListModelToJson(AddressListModel data) =>
    json.encode(data.toJson());

class AddressListModel {
  int code;
  String message;
  int current;
  List<AddressListData> data;
  int pageSize;
  bool success;
  int total;

  AddressListModel({
    required this.code,
    required this.message,
    required this.current,
    required this.data,
    required this.pageSize,
    required this.success,
    required this.total,
  });

  factory AddressListModel.fromJson(Map<String, dynamic> json) {
    final rawData = json["data"];
    final List<AddressListData> items = rawData is List
        ? rawData
            .whereType<Map>()
            .map((x) => AddressListData.fromJson(Map<String, dynamic>.from(x)))
            .toList()
        : <AddressListData>[];

    return AddressListModel(
      code: _asInt(json["code"]),
      message: _asString(json["message"]),
      current: _asInt(json["current"], fallback: 1),
      data: items,
      pageSize: _asInt(json["pageSize"], fallback: items.length),
      success: json["success"] == true || _asInt(json["code"]) == 0,
      total: _asInt(json["total"], fallback: items.length),
    );
  }

  Map<String, dynamic> toJson() => {
        "code": code,
        "message": message,
        "current": current,
        "data": List<dynamic>.from(data.map((x) => x.toJson())),
        "pageSize": pageSize,
        "success": success,
        "total": total,
      };
}

class AddressListData {
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
  DateTime createTime;
  String updateTime;

  AddressListData({
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
    required this.createTime,
    required this.updateTime,
  });

  factory AddressListData.fromJson(Map<String, dynamic> json) =>
      AddressListData(
        id: _asInt(json["id"]),
        memberId: _asInt(json["memberId"]),
        receiverName: _asString(json["receiverName"]),
        receiverPhone: _asString(json["receiverPhone"]),
        province: _asString(json["province"]),
        city: _asString(json["city"]),
        district: _asString(json["district"]),
        detailAddress: _asString(json["detailAddress"]),
        postalCode: _asString(json["postalCode"]),
        tag: _asString(json["tag"]),
        isDefault: _asInt(json["isDefault"]),
        createTime: _asDateTime(json["createTime"]),
        updateTime: _asString(json["updateTime"]),
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
        "createTime": createTime.toIso8601String(),
        "updateTime": updateTime,
      };

  String get fullAddress {
    return [province, city, district, detailAddress]
        .where((item) => item.trim().isNotEmpty)
        .join(' ');
  }

  String get contactText {
    return [receiverName, receiverPhone]
        .where((item) => item.trim().isNotEmpty)
        .join('  ');
  }
}

int _asInt(dynamic value, {int fallback = 0}) {
  if (value is int) return value;
  if (value is num) return value.toInt();
  return int.tryParse(value?.toString() ?? '') ?? fallback;
}

String _asString(dynamic value) => value?.toString() ?? '';

DateTime _asDateTime(dynamic value) {
  final raw = value?.toString() ?? '';
  return DateTime.tryParse(raw) ?? DateTime.fromMillisecondsSinceEpoch(0);
}
