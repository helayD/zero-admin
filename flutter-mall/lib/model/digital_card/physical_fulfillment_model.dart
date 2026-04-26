import 'dart:convert';

QueryMyPhysicalFulfillmentDetailResponse
    queryMyPhysicalFulfillmentDetailResponseFromJson(String str) =>
        QueryMyPhysicalFulfillmentDetailResponse.fromJson(json.decode(str));

class QueryMyPhysicalFulfillmentDetailResponse {
  final String code;
  final String message;
  final PhysicalFulfillmentDetailData data;

  const QueryMyPhysicalFulfillmentDetailResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory QueryMyPhysicalFulfillmentDetailResponse.fromJson(
    Map<String, dynamic> json,
  ) {
    return QueryMyPhysicalFulfillmentDetailResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: PhysicalFulfillmentDetailData.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class PhysicalFulfillmentDetailData {
  final int fulfillmentId;
  final String fulfillmentNo;
  final int assetInstanceId;
  final String assetNo;
  final String templateName;
  final String activityName;
  final String obtainedAt;
  final String mintStatusText;
  final String fulfillmentStatus;
  final String fulfillmentStatusText;
  final String productionStatusText;
  final String shippingStatusText;
  final String shippingFeeStatus;
  final String shippingFeeStatusText;
  final int shippingFeeAmount;
  final String receiverNameMasked;
  final String receiverPhoneMasked;
  final String addressSummary;
  final String carrierName;
  final String trackingNo;
  final String complianceTipSummary;
  final String blockedReason;
  final String blockedReasonText;
  final List<PhysicalFulfillmentTimelineItem> timeline;

  const PhysicalFulfillmentDetailData({
    required this.fulfillmentId,
    required this.fulfillmentNo,
    required this.assetInstanceId,
    required this.assetNo,
    required this.templateName,
    required this.activityName,
    required this.obtainedAt,
    required this.mintStatusText,
    required this.fulfillmentStatus,
    required this.fulfillmentStatusText,
    required this.productionStatusText,
    required this.shippingStatusText,
    required this.shippingFeeStatus,
    required this.shippingFeeStatusText,
    required this.shippingFeeAmount,
    required this.receiverNameMasked,
    required this.receiverPhoneMasked,
    required this.addressSummary,
    required this.carrierName,
    required this.trackingNo,
    required this.complianceTipSummary,
    required this.blockedReason,
    required this.blockedReasonText,
    required this.timeline,
  });

  factory PhysicalFulfillmentDetailData.fromJson(Map<String, dynamic> json) {
    return PhysicalFulfillmentDetailData(
      fulfillmentId: _intValue(json['fulfillmentId']),
      fulfillmentNo: json['fulfillmentNo']?.toString() ?? '',
      assetInstanceId: _intValue(json['assetInstanceId']),
      assetNo: json['assetNo']?.toString() ?? '',
      templateName: json['templateName']?.toString() ?? '',
      activityName: json['activityName']?.toString() ?? '',
      obtainedAt: json['obtainedAt']?.toString() ?? '',
      mintStatusText: json['mintStatusText']?.toString() ?? '',
      fulfillmentStatus: json['fulfillmentStatus']?.toString() ?? '',
      fulfillmentStatusText: json['fulfillmentStatusText']?.toString() ?? '',
      productionStatusText: json['productionStatusText']?.toString() ?? '',
      shippingStatusText: json['shippingStatusText']?.toString() ?? '',
      shippingFeeStatus: json['shippingFeeStatus']?.toString() ?? '',
      shippingFeeStatusText: json['shippingFeeStatusText']?.toString() ?? '',
      shippingFeeAmount: _intValue(json['shippingFeeAmount']),
      receiverNameMasked: json['receiverNameMasked']?.toString() ?? '',
      receiverPhoneMasked: json['receiverPhoneMasked']?.toString() ?? '',
      addressSummary: json['addressSummary']?.toString() ?? '',
      carrierName: json['carrierName']?.toString() ?? '',
      trackingNo: json['trackingNo']?.toString() ?? '',
      complianceTipSummary: json['complianceTipSummary']?.toString() ?? '',
      blockedReason: json['blockedReason']?.toString() ?? '',
      blockedReasonText: json['blockedReasonText']?.toString() ?? '',
      timeline: _listOf<PhysicalFulfillmentTimelineItem>(
        json['timeline'],
        (item) => PhysicalFulfillmentTimelineItem.fromJson(item),
      ),
    );
  }

  bool get needsAddress => fulfillmentStatus == 'pending_address';
  bool get needsShippingFee => !isBlocked && shippingFeeStatus != 'paid';
  bool get canConfirmReceipt =>
      fulfillmentStatus == 'shipped' || fulfillmentStatus == 'in_transit';
  bool get isBlocked => blockedReason.trim().isNotEmpty;
}

class PhysicalFulfillmentTimelineItem {
  final String action;
  final String actionText;
  final String statusText;
  final String reason;
  final String createTime;

  const PhysicalFulfillmentTimelineItem({
    required this.action,
    required this.actionText,
    required this.statusText,
    required this.reason,
    required this.createTime,
  });

  factory PhysicalFulfillmentTimelineItem.fromJson(Map<String, dynamic> json) {
    return PhysicalFulfillmentTimelineItem(
      action: json['action']?.toString() ?? '',
      actionText: json['actionText']?.toString() ?? '',
      statusText: json['statusText']?.toString() ?? '',
      reason: json['reason']?.toString() ?? '',
      createTime: json['createTime']?.toString() ?? '',
    );
  }
}

int _intValue(dynamic value) {
  if (value is int) {
    return value;
  }
  return int.tryParse(value?.toString() ?? '') ?? 0;
}

List<T> _listOf<T>(
  dynamic raw,
  T Function(Map<String, dynamic>) parser,
) {
  if (raw is! List) {
    return <T>[];
  }
  return raw
      .whereType<Map>()
      .map((item) => parser(Map<String, dynamic>.from(item)))
      .toList();
}
