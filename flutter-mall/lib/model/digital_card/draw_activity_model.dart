import 'dart:convert';

DrawActivityLandingResponse drawActivityLandingResponseFromJson(String str) =>
    DrawActivityLandingResponse.fromJson(json.decode(str));

PreviewDrawEligibilityResponse previewDrawEligibilityResponseFromJson(
  String str,
) =>
    PreviewDrawEligibilityResponse.fromJson(json.decode(str));

ParticipateDrawResponse participateDrawResponseFromJson(String str) =>
    ParticipateDrawResponse.fromJson(json.decode(str));

QueryMyDrawRecordListResponse queryMyDrawRecordListResponseFromJson(
        String str) =>
    QueryMyDrawRecordListResponse.fromJson(json.decode(str));

class DrawActivityLandingResponse {
  final String code;
  final String message;
  final DrawActivityLandingData data;

  const DrawActivityLandingResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory DrawActivityLandingResponse.fromJson(Map<String, dynamic> json) {
    return DrawActivityLandingResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: DrawActivityLandingData.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class DrawActivityLandingData {
  final int activityId;
  final String activityCode;
  final String name;
  final String ruleSummary;
  final String participantConditionSummary;
  final String consumeRuleSummary;
  final String probabilityRule;
  final String complianceRuleSummary;
  final String circulationLimitSummary;
  final String startTime;
  final String endTime;
  final int realNameRequired;
  final DrawEligibilitySummary eligibility;
  final DrawIdentitySummary identity;
  final List<DrawCardPreview> cardPreviews;
  final List<DrawPoolPreview> pools;
  final List<DrawRecentWin> recentWins;
  final List<DrawMemberRecord> myRecords;

  const DrawActivityLandingData({
    required this.activityId,
    required this.activityCode,
    required this.name,
    required this.ruleSummary,
    required this.participantConditionSummary,
    required this.consumeRuleSummary,
    required this.probabilityRule,
    required this.complianceRuleSummary,
    required this.circulationLimitSummary,
    required this.startTime,
    required this.endTime,
    required this.realNameRequired,
    required this.eligibility,
    required this.identity,
    required this.cardPreviews,
    required this.pools,
    required this.recentWins,
    required this.myRecords,
  });

  factory DrawActivityLandingData.fromJson(Map<String, dynamic> json) {
    return DrawActivityLandingData(
      activityId: _intValue(json['activityId']),
      activityCode: json['activityCode']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      ruleSummary: json['ruleSummary']?.toString() ?? '',
      participantConditionSummary:
          json['participantConditionSummary']?.toString() ?? '',
      consumeRuleSummary: json['consumeRuleSummary']?.toString() ?? '',
      probabilityRule: json['probabilityRule']?.toString() ?? '',
      complianceRuleSummary: json['complianceRuleSummary']?.toString() ?? '',
      circulationLimitSummary:
          json['circulationLimitSummary']?.toString() ?? '',
      startTime: json['startTime']?.toString() ?? '',
      endTime: json['endTime']?.toString() ?? '',
      realNameRequired: _intValue(json['realNameRequired']),
      eligibility: DrawEligibilitySummary.fromJson(
        Map<String, dynamic>.from(
          json['eligibility'] ?? const <String, dynamic>{},
        ),
      ),
      identity: DrawIdentitySummary.fromJson(
        Map<String, dynamic>.from(
          json['identity'] ?? const <String, dynamic>{},
        ),
      ),
      cardPreviews: _listOf<DrawCardPreview>(
        json['cardPreviews'],
        (item) => DrawCardPreview.fromJson(item),
      ),
      pools: _listOf<DrawPoolPreview>(
        json['pools'],
        (item) => DrawPoolPreview.fromJson(item),
      ),
      recentWins: _listOf<DrawRecentWin>(
        json['recentWins'],
        (item) => DrawRecentWin.fromJson(item),
      ),
      myRecords: _listOf<DrawMemberRecord>(
        json['myRecords'],
        (item) => DrawMemberRecord.fromJson(item),
      ),
    );
  }
}

class DrawEligibilitySummary {
  final String eligibilityStatus;
  final String eligibilityCode;
  final String eligibilityMessage;
  final String nextAction;
  final int remainingLotteryTimes;

  const DrawEligibilitySummary({
    required this.eligibilityStatus,
    required this.eligibilityCode,
    required this.eligibilityMessage,
    required this.nextAction,
    required this.remainingLotteryTimes,
  });

  factory DrawEligibilitySummary.fromJson(Map<String, dynamic> json) {
    return DrawEligibilitySummary(
      eligibilityStatus: json['eligibilityStatus']?.toString() ?? '',
      eligibilityCode: json['eligibilityCode']?.toString() ?? '',
      eligibilityMessage: json['eligibilityMessage']?.toString() ?? '',
      nextAction: json['nextAction']?.toString() ?? '',
      remainingLotteryTimes: _intValue(json['remainingLotteryTimes']),
    );
  }
}

class DrawIdentitySummary {
  final String realNameStatus;
  final String realNameStatusText;
  final String realNameMasked;
  final String credentialRef;
  final String verifiedAt;

  const DrawIdentitySummary({
    required this.realNameStatus,
    required this.realNameStatusText,
    required this.realNameMasked,
    required this.credentialRef,
    required this.verifiedAt,
  });

  factory DrawIdentitySummary.fromJson(Map<String, dynamic> json) {
    return DrawIdentitySummary(
      realNameStatus: json['realNameStatus']?.toString() ?? '',
      realNameStatusText: json['realNameStatusText']?.toString() ?? '',
      realNameMasked: json['realNameMasked']?.toString() ?? '',
      credentialRef: json['credentialRef']?.toString() ?? '',
      verifiedAt: json['verifiedAt']?.toString() ?? '',
    );
  }
}

class DrawCardPreview {
  final int templateId;
  final String templateCode;
  final String templateName;
  final String cardFaceImage;
  final String rarity;
  final String displayCopy;

  const DrawCardPreview({
    required this.templateId,
    required this.templateCode,
    required this.templateName,
    required this.cardFaceImage,
    required this.rarity,
    required this.displayCopy,
  });

  factory DrawCardPreview.fromJson(Map<String, dynamic> json) {
    return DrawCardPreview(
      templateId: _intValue(json['templateId']),
      templateCode: json['templateCode']?.toString() ?? '',
      templateName: json['templateName']?.toString() ?? '',
      cardFaceImage: json['cardFaceImage']?.toString() ?? '',
      rarity: json['rarity']?.toString() ?? '',
      displayCopy: json['displayCopy']?.toString() ?? '',
    );
  }
}

class DrawPoolPreview {
  final int poolId;
  final String poolName;
  final String probabilityRule;
  final List<DrawCardPreview> cards;

  const DrawPoolPreview({
    required this.poolId,
    required this.poolName,
    required this.probabilityRule,
    required this.cards,
  });

  factory DrawPoolPreview.fromJson(Map<String, dynamic> json) {
    return DrawPoolPreview(
      poolId: _intValue(json['poolId']),
      poolName: json['poolName']?.toString() ?? '',
      probabilityRule: json['probabilityRule']?.toString() ?? '',
      cards: _listOf<DrawCardPreview>(
        json['cards'],
        (item) => DrawCardPreview.fromJson(item),
      ),
    );
  }
}

class DrawRecentWin {
  final int recordId;
  final int memberId;
  final String memberNameMasked;
  final String resultStatus;
  final String resultStatusText;
  final String templateName;
  final String rarity;
  final String createTime;

  const DrawRecentWin({
    required this.recordId,
    required this.memberId,
    required this.memberNameMasked,
    required this.resultStatus,
    required this.resultStatusText,
    required this.templateName,
    required this.rarity,
    required this.createTime,
  });

  factory DrawRecentWin.fromJson(Map<String, dynamic> json) {
    return DrawRecentWin(
      recordId: _intValue(json['recordId']),
      memberId: _intValue(json['memberId']),
      memberNameMasked: json['memberNameMasked']?.toString() ?? '',
      resultStatus: json['resultStatus']?.toString() ?? '',
      resultStatusText: json['resultStatusText']?.toString() ?? '',
      templateName: json['templateName']?.toString() ?? '',
      rarity: json['rarity']?.toString() ?? '',
      createTime: json['createTime']?.toString() ?? '',
    );
  }
}

class DrawMemberRecord {
  final int id;
  final int activityId;
  final String requestId;
  final String resultType;
  final String resultStatus;
  final String resultStatusText;
  final String failureCode;
  final String failureReason;
  final int poolId;
  final int templateId;
  final String templateName;
  final String rarity;
  final int consumeAmount;
  final int lotteryTimesBefore;
  final int lotteryTimesAfter;
  final int assetInstanceId;
  final String assetNo;
  final String assetStatus;
  final String assetStatusText;
  final String assetCreatedAt;
  final String createTime;

  const DrawMemberRecord({
    required this.id,
    required this.activityId,
    required this.requestId,
    required this.resultType,
    required this.resultStatus,
    required this.resultStatusText,
    required this.failureCode,
    required this.failureReason,
    required this.poolId,
    required this.templateId,
    required this.templateName,
    required this.rarity,
    required this.consumeAmount,
    required this.lotteryTimesBefore,
    required this.lotteryTimesAfter,
    required this.assetInstanceId,
    required this.assetNo,
    required this.assetStatus,
    required this.assetStatusText,
    required this.assetCreatedAt,
    required this.createTime,
  });

  factory DrawMemberRecord.fromJson(Map<String, dynamic> json) {
    return DrawMemberRecord(
      id: _intValue(json['id']),
      activityId: _intValue(json['activityId']),
      requestId: json['requestId']?.toString() ?? '',
      resultType: json['resultType']?.toString() ?? '',
      resultStatus: json['resultStatus']?.toString() ?? '',
      resultStatusText: json['resultStatusText']?.toString() ?? '',
      failureCode: json['failureCode']?.toString() ?? '',
      failureReason: json['failureReason']?.toString() ?? '',
      poolId: _intValue(json['poolId']),
      templateId: _intValue(json['templateId']),
      templateName: json['templateName']?.toString() ?? '',
      rarity: json['rarity']?.toString() ?? '',
      consumeAmount: _intValue(json['consumeAmount']),
      lotteryTimesBefore: _intValue(json['lotteryTimesBefore']),
      lotteryTimesAfter: _intValue(json['lotteryTimesAfter']),
      assetInstanceId: _intValue(json['assetInstanceId']),
      assetNo: json['assetNo']?.toString() ?? '',
      assetStatus: json['assetStatus']?.toString() ?? '',
      assetStatusText: json['assetStatusText']?.toString() ?? '',
      assetCreatedAt: json['assetCreatedAt']?.toString() ?? '',
      createTime: json['createTime']?.toString() ?? '',
    );
  }
}

class PreviewDrawEligibilityResponse {
  final String code;
  final String message;
  final DrawEligibilitySummary data;

  const PreviewDrawEligibilityResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory PreviewDrawEligibilityResponse.fromJson(Map<String, dynamic> json) {
    return PreviewDrawEligibilityResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: DrawEligibilitySummary.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class ParticipateDrawResponse {
  final String code;
  final String message;
  final ParticipateDrawData data;

  const ParticipateDrawResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory ParticipateDrawResponse.fromJson(Map<String, dynamic> json) {
    return ParticipateDrawResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: ParticipateDrawData.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class ParticipateDrawData {
  final DrawMemberRecord record;
  final DrawEligibilitySummary eligibility;

  const ParticipateDrawData({
    required this.record,
    required this.eligibility,
  });

  factory ParticipateDrawData.fromJson(Map<String, dynamic> json) {
    return ParticipateDrawData(
      record: DrawMemberRecord.fromJson(
        Map<String, dynamic>.from(json['record'] ?? const <String, dynamic>{}),
      ),
      eligibility: DrawEligibilitySummary.fromJson(
        Map<String, dynamic>.from(
          json['eligibility'] ?? const <String, dynamic>{},
        ),
      ),
    );
  }
}

class QueryMyDrawRecordListResponse {
  final String code;
  final String message;
  final QueryMyDrawRecordListData data;

  const QueryMyDrawRecordListResponse({
    required this.code,
    required this.message,
    required this.data,
  });

  factory QueryMyDrawRecordListResponse.fromJson(Map<String, dynamic> json) {
    return QueryMyDrawRecordListResponse(
      code: json['code']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      data: QueryMyDrawRecordListData.fromJson(
        Map<String, dynamic>.from(json['data'] ?? const <String, dynamic>{}),
      ),
    );
  }
}

class QueryMyDrawRecordListData {
  final int total;
  final List<DrawMemberRecord> list;

  const QueryMyDrawRecordListData({
    required this.total,
    required this.list,
  });

  factory QueryMyDrawRecordListData.fromJson(Map<String, dynamic> json) {
    return QueryMyDrawRecordListData(
      total: _intValue(json['total']),
      list: _listOf<DrawMemberRecord>(
        json['list'],
        (item) => DrawMemberRecord.fromJson(item),
      ),
    );
  }
}

int _intValue(dynamic value) {
  if (value is int) {
    return value;
  }
  if (value is double) {
    return value.toInt();
  }
  return int.tryParse(value?.toString() ?? '') ?? 0;
}

List<T> _listOf<T>(
  dynamic source,
  T Function(Map<String, dynamic>) builder,
) {
  if (source is! List) {
    return <T>[];
  }
  return source
      .map((item) => builder(Map<String, dynamic>.from(item)))
      .toList();
}
