import 'dart:convert';

ConfirmOrderModel confirmOrderModelFromJson(String str) =>
    ConfirmOrderModel.fromJson(json.decode(str));

class ConfirmOrderModel {
  int code;
  String message;
  ConfirmOrderData data;

  ConfirmOrderModel({
    required this.code,
    required this.message,
    required this.data,
  });

  factory ConfirmOrderModel.fromJson(Map<String, dynamic> json) =>
      ConfirmOrderModel(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        data: ConfirmOrderData.fromJson(json["data"] ?? {}),
      );
}

class ConfirmOrderData {
  List<ConfirmCartItem> cartPromotionItemList;
  List<ConfirmAddress> memberReceiveAddressList;
  ConfirmCouponList couponHistoryDetailList;
  IntegrationConsumeSetting integrationConsumeSetting;
  int memberIntegration;
  CalcAmountData calcAmount;

  ConfirmOrderData({
    required this.cartPromotionItemList,
    required this.memberReceiveAddressList,
    required this.couponHistoryDetailList,
    required this.integrationConsumeSetting,
    required this.memberIntegration,
    required this.calcAmount,
  });

  factory ConfirmOrderData.fromJson(Map<String, dynamic> json) =>
      ConfirmOrderData(
        cartPromotionItemList: json["cartPromotionItemList"] == null
            ? []
            : List<ConfirmCartItem>.from(
                json["cartPromotionItemList"]
                    .map((x) => ConfirmCartItem.fromJson(x))),
        memberReceiveAddressList: json["memberReceiveAddressList"] == null
            ? []
            : List<ConfirmAddress>.from(
                json["memberReceiveAddressList"]
                    .map((x) => ConfirmAddress.fromJson(x))),
        couponHistoryDetailList: ConfirmCouponList.fromJson(
            json["couponHistoryDetailList"] ?? {}),
        integrationConsumeSetting: IntegrationConsumeSetting.fromJson(
            json["integrationConsumeSetting"] ?? {}),
        memberIntegration: json["memberIntegration"] ?? 0,
        calcAmount: CalcAmountData.fromJson(json["calcAmount"] ?? {}),
      );
}

class ConfirmCartItem {
  int id;
  int productId;
  int productSkuId;
  int memberId;
  int quantity;
  int price;
  String productPic;
  String productName;
  String productSubTitle;
  String productSkuCode;
  String memberNickname;
  String createDate;
  String modifyDate;
  int deleteStatus;
  int productCategoryId;
  String productBrand;
  String productSn;
  String productAttr;
  String promotionMessage;
  int reduceAmount;
  int realStock;
  int integration;
  int growth;

  ConfirmCartItem({
    required this.id,
    required this.productId,
    required this.productSkuId,
    required this.memberId,
    required this.quantity,
    required this.price,
    required this.productPic,
    required this.productName,
    required this.productSubTitle,
    required this.productSkuCode,
    required this.memberNickname,
    required this.createDate,
    required this.modifyDate,
    required this.deleteStatus,
    required this.productCategoryId,
    required this.productBrand,
    required this.productSn,
    required this.productAttr,
    required this.promotionMessage,
    required this.reduceAmount,
    required this.realStock,
    required this.integration,
    required this.growth,
  });

  factory ConfirmCartItem.fromJson(Map<String, dynamic> json) =>
      ConfirmCartItem(
        id: json["id"] ?? 0,
        productId: json["productId"] ?? 0,
        productSkuId: json["productSkuId"] ?? 0,
        memberId: json["memberId"] ?? 0,
        quantity: json["quantity"] ?? 0,
        price: json["price"] ?? 0,
        productPic: json["productPic"] ?? "",
        productName: json["productName"] ?? "",
        productSubTitle: json["productSubTitle"] ?? "",
        productSkuCode: json["productSkuCode"] ?? "",
        memberNickname: json["memberNickname"] ?? "",
        createDate: json["createDate"] ?? "",
        modifyDate: json["modifyDate"] ?? "",
        deleteStatus: json["deleteStatus"] ?? 0,
        productCategoryId: json["productCategoryId"] ?? 0,
        productBrand: json["productBrand"] ?? "",
        productSn: json["productSn"] ?? "",
        productAttr: json["productAttr"] ?? "",
        promotionMessage: json["promotionMessage"] ?? "",
        reduceAmount: json["reduceAmount"] ?? 0,
        realStock: json["realStock"] ?? 0,
        integration: json["integration"] ?? 0,
        growth: json["growth"] ?? 0,
      );
}

class ConfirmAddress {
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

  ConfirmAddress({
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

  factory ConfirmAddress.fromJson(Map<String, dynamic> json) =>
      ConfirmAddress(
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
}

class ConfirmCouponList {
  List<ConfirmCouponData> enableList;
  List<ConfirmCouponData> disableList;

  ConfirmCouponList({
    required this.enableList,
    required this.disableList,
  });

  factory ConfirmCouponList.fromJson(Map<String, dynamic> json) =>
      ConfirmCouponList(
        enableList: json["enableList"] == null
            ? []
            : List<ConfirmCouponData>.from(
                json["enableList"].map((x) => ConfirmCouponData.fromJson(x))),
        disableList: json["disableList"] == null
            ? []
            : List<ConfirmCouponData>.from(
                json["disableList"]
                    .map((x) => ConfirmCouponData.fromJson(x))),
      );
}

class ConfirmCouponData {
  int id;
  String name;
  double amount;
  double minAmount;
  String startTime;
  String endTime;
  String description;
  String disableReason;

  ConfirmCouponData({
    required this.id,
    required this.name,
    required this.amount,
    required this.minAmount,
    required this.startTime,
    required this.endTime,
    required this.description,
    required this.disableReason,
  });

  factory ConfirmCouponData.fromJson(Map<String, dynamic> json) =>
      ConfirmCouponData(
        id: json["id"] ?? 0,
        name: json["name"] ?? "",
        amount: (json["amount"] ?? 0).toDouble(),
        minAmount: (json["minAmount"] ?? 0).toDouble(),
        startTime: json["startTime"] ?? "",
        endTime: json["endTime"] ?? "",
        description: json["description"] ?? "",
        disableReason: json["disableReason"] ?? "",
      );
}

class IntegrationConsumeSetting {
  int id;
  int deductionPerAmount;
  int maxPercentPerOrder;
  int useUnit;
  int couponStatus;

  IntegrationConsumeSetting({
    required this.id,
    required this.deductionPerAmount,
    required this.maxPercentPerOrder,
    required this.useUnit,
    required this.couponStatus,
  });

  factory IntegrationConsumeSetting.fromJson(Map<String, dynamic> json) =>
      IntegrationConsumeSetting(
        id: json["id"] ?? 0,
        deductionPerAmount: json["deductionPerAmount"] ?? 0,
        maxPercentPerOrder: json["maxPercentPerOrder"] ?? 0,
        useUnit: json["useUnit"] ?? 0,
        couponStatus: json["couponStatus"] ?? 0,
      );
}

class CalcAmountData {
  int totalAmount;
  int freightAmount;
  int promotionAmount;
  int payAmount;

  CalcAmountData({
    required this.totalAmount,
    required this.freightAmount,
    required this.promotionAmount,
    required this.payAmount,
  });

  factory CalcAmountData.fromJson(Map<String, dynamic> json) =>
      CalcAmountData(
        totalAmount: json["totalAmount"] ?? 0,
        freightAmount: json["freightAmount"] ?? 0,
        promotionAmount: json["promotionAmount"] ?? 0,
        payAmount: json["payAmount"] ?? 0,
      );
}
