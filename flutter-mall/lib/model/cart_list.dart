// To parse this JSON data, do
//
//     final cartListModel = cartListModelFromJson(jsonString);

import 'dart:convert';

CartListModel cartListModelFromJson(String str) =>
    CartListModel.fromJson(json.decode(str));

String cartListModelToJson(CartListModel data) => json.encode(data.toJson());

class CartListModel {
  int code;
  String message;
  List<CartData> data;

  CartListModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory CartListModel.fromJson(Map<String, dynamic> json) => CartListModel(
        code: json["code"],
        message: json["message"],
        data:
            List<CartData>.from(json["data"].map((x) => CartData.fromJson(x))),
      );

  Map<String, dynamic> toJson() => {
        "code": code,
        "message": message,
        "data": List<dynamic>.from(data.map((x) => x.toJson())),
      };
}

class CartData {
  int id;
  int memberId;
  int productId;
  int productSkuId;
  int quantity;
  int price;
  int selected;
  String productName;
  String productSubTitle;
  String productPic;
  String productSkuCode;
  String productSn;
  String productBrand;
  int productCategoryId;
  String productAttr;
  String memberNickname;
  int source;
  DateTime expireTime;
  DateTime createTime;
  DateTime updateTime;

  CartData({
    required this.id,
    required this.memberId,
    required this.productId,
    required this.productSkuId,
    required this.quantity,
    required this.price,
    required this.selected,
    required this.productName,
    required this.productSubTitle,
    required this.productPic,
    required this.productSkuCode,
    required this.productSn,
    required this.productBrand,
    required this.productCategoryId,
    required this.productAttr,
    required this.memberNickname,
    required this.source,
    required this.expireTime,
    required this.createTime,
    required this.updateTime,
  });

  factory CartData.fromJson(Map<String, dynamic> json) => CartData(
        id: json["id"],
        memberId: json["memberId"],
        productId: json["productId"],
        productSkuId: json["productSkuId"],
        quantity: json["quantity"],
        price: json["price"],
        selected: json["selected"],
        productName: json["productName"],
        productSubTitle: json["productSubTitle"],
        productPic: json["productPic"],
        productSkuCode: json["productSkuCode"],
        productSn: json["productSn"],
        productBrand: json["productBrand"],
        productCategoryId: json["productCategoryId"],
        productAttr: _normalizeProductAttr(json["productAttr"]),
        memberNickname: json["memberNickname"],
        source: json["source"],
        expireTime: _parseDateTime(json["expireTime"]),
        createTime: _parseDateTime(json["createTime"]),
        updateTime:
            _parseDateTime(json["updateTime"], fallback: json["createTime"]),
      );

  Map<String, dynamic> toJson() => {
        "id": id,
        "memberId": memberId,
        "productId": productId,
        "productSkuId": productSkuId,
        "quantity": quantity,
        "price": price,
        "selected": selected,
        "productName": productName,
        "productSubTitle": productSubTitle,
        "productPic": productPic,
        "productSkuCode": productSkuCode,
        "productSn": productSn,
        "productBrand": productBrand,
        "productCategoryId": productCategoryId,
        "productAttr": productAttr,
        "memberNickname": memberNickname,
        "source": source,
        "expireTime": expireTime.toIso8601String(),
        "createTime": createTime.toIso8601String(),
        "updateTime": updateTime.toIso8601String(),
      };

  static String _normalizeProductAttr(dynamic raw) {
    if (raw == null) {
      return '';
    }
    if (raw is String) {
      final trimmed = raw.trim();
      if (trimmed.isEmpty) {
        return '';
      }
      if ((trimmed.startsWith('{') && trimmed.endsWith('}')) ||
          (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
        try {
          return _normalizeProductAttr(jsonDecode(trimmed));
        } catch (_) {
          return trimmed;
        }
      }
      return trimmed;
    }
    if (raw is Map) {
      final parts = <String>[];
      for (final entry in raw.entries) {
        final key = entry.key.toString().trim();
        final value = entry.value?.toString().trim() ?? '';
        if (key.isEmpty && value.isEmpty) {
          continue;
        }
        if (key.isEmpty) {
          parts.add(value);
          continue;
        }
        if (value.isEmpty) {
          parts.add(key);
          continue;
        }
        parts.add('$key: $value');
      }
      return parts.join(' · ');
    }
    if (raw is List) {
      final parts = <String>[];
      for (final item in raw) {
        final normalized = _normalizeProductAttr(item).trim();
        if (normalized.isNotEmpty) {
          parts.add(normalized);
        }
      }
      return parts.join(' · ');
    }
    return raw.toString();
  }

  static DateTime _parseDateTime(dynamic raw, {dynamic fallback}) {
    final candidates = [raw, fallback];
    for (final candidate in candidates) {
      if (candidate == null) {
        continue;
      }
      final text = candidate.toString().trim();
      if (text.isEmpty) {
        continue;
      }
      try {
        return DateTime.parse(text);
      } catch (_) {
        continue;
      }
    }
    return DateTime.fromMillisecondsSinceEpoch(0);
  }
}
