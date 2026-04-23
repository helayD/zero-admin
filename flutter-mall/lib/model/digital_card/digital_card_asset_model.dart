import 'dart:convert';

QueryMyDigitalCardAssetListResponse queryMyDigitalCardAssetListResponseFromJson(
  String str,
) =>
    QueryMyDigitalCardAssetListResponse.fromJson(json.decode(str));

QueryMyDigitalCardAssetDetailResponse
    queryMyDigitalCardAssetDetailResponseFromJson(String str) =>
        QueryMyDigitalCardAssetDetailResponse.fromJson(json.decode(str));

class QueryMyDigitalCardAssetListResponse {
  final String code;
  final String message;
  final QueryMyDigitalCardAssetListData data;

  const QueryMyDigitalCardAssetListResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory QueryMyDigitalCardAssetListResponse.fromJson(
    Map<String, dynamic> json,
  ) {
    return QueryMyDigitalCardAssetListResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: QueryMyDigitalCardAssetListData.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class QueryMyDigitalCardAssetListData {
  final int total;
  final List<DigitalCardAssetItem> list;

  const QueryMyDigitalCardAssetListData({
    required this.total,
    required this.list,
  });

  factory QueryMyDigitalCardAssetListData.fromJson(Map<String, dynamic> json) {
    return QueryMyDigitalCardAssetListData(
      total: _intValue(json['total']),
      list: _listOf<DigitalCardAssetItem>(
        json['list'],
        (item) => DigitalCardAssetItem.fromJson(item),
      ),
    );
  }
}

class QueryMyDigitalCardAssetDetailResponse {
  final String code;
  final String message;
  final DigitalCardAssetDetailData data;

  const QueryMyDigitalCardAssetDetailResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory QueryMyDigitalCardAssetDetailResponse.fromJson(
    Map<String, dynamic> json,
  ) {
    return QueryMyDigitalCardAssetDetailResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: DigitalCardAssetDetailData.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class DigitalCardAssetItem {
  final int assetInstanceId;
  final String assetNo;
  final int templateId;
  final String templateName;
  final String cardFaceImage;
  final int activityId;
  final String activityName;
  final String rarity;
  final String obtainedAt;
  final String mintStatus;
  final String mintStatusText;
  final String chainStatus;
  final String chainStatusText;
  final String displayStatus;
  final String displayStatusText;
  final String complianceStatus;
  final String complianceStatusText;
  final String tokenStatusText;
  final String complianceRuleSummary;
  final String chainType;

  const DigitalCardAssetItem({
    required this.assetInstanceId,
    required this.assetNo,
    required this.templateId,
    required this.templateName,
    required this.cardFaceImage,
    required this.activityId,
    required this.activityName,
    required this.rarity,
    required this.obtainedAt,
    required this.mintStatus,
    required this.mintStatusText,
    required this.chainStatus,
    required this.chainStatusText,
    required this.displayStatus,
    required this.displayStatusText,
    required this.complianceStatus,
    required this.complianceStatusText,
    required this.tokenStatusText,
    required this.complianceRuleSummary,
    required this.chainType,
  });

  factory DigitalCardAssetItem.fromJson(Map<String, dynamic> json) {
    return DigitalCardAssetItem(
      assetInstanceId: _intValue(json['assetInstanceId']),
      assetNo: json['assetNo']?.toString() ?? '',
      templateId: _intValue(json['templateId']),
      templateName: json['templateName']?.toString() ?? '',
      cardFaceImage: json['cardFaceImage']?.toString() ?? '',
      activityId: _intValue(json['activityId']),
      activityName: json['activityName']?.toString() ?? '',
      rarity: json['rarity']?.toString() ?? '',
      obtainedAt: json['obtainedAt']?.toString() ?? '',
      mintStatus: json['mintStatus']?.toString() ?? '',
      mintStatusText: json['mintStatusText']?.toString() ?? '',
      chainStatus: json['chainStatus']?.toString() ?? '',
      chainStatusText: json['chainStatusText']?.toString() ?? '',
      displayStatus: json['displayStatus']?.toString() ?? '',
      displayStatusText: json['displayStatusText']?.toString() ?? '',
      complianceStatus: json['complianceStatus']?.toString() ?? '',
      complianceStatusText: json['complianceStatusText']?.toString() ?? '',
      tokenStatusText: json['tokenStatusText']?.toString() ?? '',
      complianceRuleSummary: json['complianceRuleSummary']?.toString() ?? '',
      chainType: json['chainType']?.toString() ?? '',
    );
  }

  bool get hasRestriction =>
      displayStatus != 'display_visible' ||
      complianceStatus != 'compliance_clear';
}

class DigitalCardAssetDetailData {
  final DigitalCardAssetItem item;
  final String tokenIdMasked;
  final String latestStatusSummary;
  final String restrictionReason;
  final DigitalCardAssetDrawSummary drawSummary;
  final List<DigitalCardAssetTimelineItem> timeline;

  const DigitalCardAssetDetailData({
    required this.item,
    required this.tokenIdMasked,
    required this.latestStatusSummary,
    required this.restrictionReason,
    required this.drawSummary,
    required this.timeline,
  });

  factory DigitalCardAssetDetailData.fromJson(Map<String, dynamic> json) {
    return DigitalCardAssetDetailData(
      item: DigitalCardAssetItem.fromJson(
        Map<String, dynamic>.from(json['item'] ?? const <String, dynamic>{}),
      ),
      tokenIdMasked: json['tokenIdMasked']?.toString() ?? '',
      latestStatusSummary: json['latestStatusSummary']?.toString() ?? '',
      restrictionReason: json['restrictionReason']?.toString() ?? '',
      drawSummary: DigitalCardAssetDrawSummary.fromJson(
        Map<String, dynamic>.from(
          json['drawSummary'] ?? const <String, dynamic>{},
        ),
      ),
      timeline: _listOf<DigitalCardAssetTimelineItem>(
        json['timeline'],
        (item) => DigitalCardAssetTimelineItem.fromJson(item),
      ),
    );
  }
}

class DigitalCardAssetDrawSummary {
  final int participationRecordId;
  final String resultType;
  final String resultStatus;
  final String resultStatusText;
  final String failureReason;
  final String createTime;

  const DigitalCardAssetDrawSummary({
    required this.participationRecordId,
    required this.resultType,
    required this.resultStatus,
    required this.resultStatusText,
    required this.failureReason,
    required this.createTime,
  });

  factory DigitalCardAssetDrawSummary.fromJson(Map<String, dynamic> json) {
    return DigitalCardAssetDrawSummary(
      participationRecordId: _intValue(json['participationRecordId']),
      resultType: json['resultType']?.toString() ?? '',
      resultStatus: json['resultStatus']?.toString() ?? '',
      resultStatusText: json['resultStatusText']?.toString() ?? '',
      failureReason: json['failureReason']?.toString() ?? '',
      createTime: json['createTime']?.toString() ?? '',
    );
  }
}

class DigitalCardAssetTimelineItem {
  final String operationType;
  final String operationText;
  final String statusText;
  final String reasonText;
  final String createTime;

  const DigitalCardAssetTimelineItem({
    required this.operationType,
    required this.operationText,
    required this.statusText,
    required this.reasonText,
    required this.createTime,
  });

  factory DigitalCardAssetTimelineItem.fromJson(Map<String, dynamic> json) {
    return DigitalCardAssetTimelineItem(
      operationType: json['operationType']?.toString() ?? '',
      operationText: json['operationText']?.toString() ?? '',
      statusText: json['statusText']?.toString() ?? '',
      reasonText: json['reasonText']?.toString() ?? '',
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
