import 'dart:convert';

PointsLogListResponse pointsLogListResponseFromJson(String str) =>
    PointsLogListResponse.fromJson(json.decode(str));

class PointsLogListResponse {
  final int total;
  final List<PointsLogItem> list;

  const PointsLogListResponse({required this.total, required this.list});

  factory PointsLogListResponse.fromJson(Map<String, dynamic> json) {
    final rawList = json['list'];
    final items = rawList is List
        ? rawList
            .map((e) =>
                PointsLogItem.fromJson(Map<String, dynamic>.from(e ?? {})))
            .toList()
        : <PointsLogItem>[];
    return PointsLogListResponse(
      total: (json['total'] as num?)?.toInt() ?? 0,
      list: items,
    );
  }
}

class PointsLogItem {
  final int id;
  final int changeType;   // 1=增加, 2=减少
  final int changePoints; // 变更积分绝对值
  final int sourceType;   // 0-其他,1-订单,2-活动,3-签到,4-管理员
  final String description;
  final String createTime;

  const PointsLogItem({
    required this.id,
    required this.changeType,
    required this.changePoints,
    required this.sourceType,
    required this.description,
    required this.createTime,
  });

  bool get isAdd => changeType == 1;

  String get sourceLabel {
    switch (sourceType) {
      case 1:
        return '订单';
      case 2:
        return '活动';
      case 3:
        return '签到';
      case 4:
        return '管理员';
      default:
        return '其他';
    }
  }

  factory PointsLogItem.fromJson(Map<String, dynamic> json) {
    return PointsLogItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      changeType: (json['changeType'] as num?)?.toInt() ?? 0,
      changePoints: (json['changePoints'] as num?)?.toInt() ?? 0,
      sourceType: (json['sourceType'] as num?)?.toInt() ?? 0,
      description: json['description']?.toString() ?? '',
      createTime: json['createTime']?.toString() ?? '',
    );
  }
}
