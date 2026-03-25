import 'dart:convert';

CartPromotionListModel cartPromotionListModelFromJson(String str) =>
    CartPromotionListModel.fromJson(json.decode(str));

class CartPromotionListModel {
  int code;
  String message;
  List<CartPromotionData> data;

  CartPromotionListModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory CartPromotionListModel.fromJson(Map<String, dynamic> json) =>
      CartPromotionListModel(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        data: json["data"] == null
            ? []
            : List<CartPromotionData>.from(
                json["data"].map((x) => CartPromotionData.fromJson(x))),
      );
}

class CartPromotionData {
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
  int deleteStatus;
  String expireTime;
  String createTime;
  String updateTime;
  String promotionMessage;
  int reduceAmount;
  int realStock;
  int integration;
  int growth;

  CartPromotionData({
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
    required this.deleteStatus,
    required this.expireTime,
    required this.createTime,
    required this.updateTime,
    required this.promotionMessage,
    required this.reduceAmount,
    required this.realStock,
    required this.integration,
    required this.growth,
  });

  factory CartPromotionData.fromJson(Map<String, dynamic> json) =>
      CartPromotionData(
        id: json["id"] ?? 0,
        memberId: json["memberId"] ?? 0,
        productId: json["productId"] ?? 0,
        productSkuId: json["productSkuId"] ?? 0,
        quantity: json["quantity"] ?? 0,
        price: json["price"] ?? 0,
        selected: json["selected"] ?? 0,
        productName: json["productName"] ?? "",
        productSubTitle: json["productSubTitle"] ?? "",
        productPic: json["productPic"] ?? "",
        productSkuCode: json["productSkuCode"] ?? "",
        productSn: json["productSn"] ?? "",
        productBrand: json["productBrand"] ?? "",
        productCategoryId: json["productCategoryId"] ?? 0,
        productAttr: json["productAttr"] ?? "",
        memberNickname: json["memberNickname"] ?? "",
        source: json["source"] ?? 0,
        deleteStatus: json["deleteStatus"] ?? 0,
        expireTime: json["expireTime"] ?? "",
        createTime: json["createTime"] ?? "",
        updateTime: json["updateTime"] ?? "",
        promotionMessage: json["promotionMessage"] ?? "",
        reduceAmount: json["reduceAmount"] ?? 0,
        realStock: json["realStock"] ?? 0,
        integration: json["integration"] ?? 0,
        growth: json["growth"] ?? 0,
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
        "deleteStatus": deleteStatus,
        "expireTime": expireTime,
        "createTime": createTime,
        "updateTime": updateTime,
        "promotionMessage": promotionMessage,
        "reduceAmount": reduceAmount,
        "realStock": realStock,
        "integration": integration,
        "growth": growth,
      };
}
