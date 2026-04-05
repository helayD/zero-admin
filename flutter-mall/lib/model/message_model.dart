///
/// 会员消息数据模型
///
class MessageModel {
  final int code;
  final String? message;
  final List<MessageData> data;

  MessageModel({
    required this.code,
    this.message,
    required this.data,
  });

  factory MessageModel.fromJson(Map<String, dynamic> json) {
    return MessageModel(
      code: json['code'] ?? 0,
      message: json['message'],
      data: (json['data'] as List<dynamic>?)
              ?.map((e) => MessageData.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class MessageData {
  final int id;
  final int memberId;
  final int messageType; // 1-订单 2-支付 3-售后 4-活动 5-会员
  final String title;
  final String content;
  final String? imageUrl;
  final String? linkType; // order/coupon/product
  final String? linkId;
  final int relatedOrderId;
  final int status; // 0-未读 1-已读
  final int isDelete;
  final DateTime? createTime;
  final DateTime? readTime;

  MessageData({
    required this.id,
    required this.memberId,
    required this.messageType,
    required this.title,
    required this.content,
    this.imageUrl,
    this.linkType,
    this.linkId,
    required this.relatedOrderId,
    required this.status,
    required this.isDelete,
    this.createTime,
    this.readTime,
  });

  factory MessageData.fromJson(Map<String, dynamic> json) {
    return MessageData(
      id: json['id'] ?? 0,
      memberId: json['memberId'] ?? 0,
      messageType: json['messageType'] ?? 0,
      title: json['title'] ?? '',
      content: json['content'] ?? '',
      imageUrl: json['imageUrl'],
      linkType: json['linkType'],
      linkId: json['linkId'],
      relatedOrderId: json['relatedOrderId'] ?? 0,
      status: json['status'] ?? 0,
      isDelete: json['isDelete'] ?? 0,
      createTime: json['createTime'] != null
          ? DateTime.tryParse(json['createTime'])
          : null,
      readTime: json['readTime'] != null
          ? DateTime.tryParse(json['readTime'])
          : null,
    );
  }

  /// 消息类型描述
  String get typeName {
    switch (messageType) {
      case 1:
        return '订单';
      case 2:
        return '支付';
      case 3:
        return '售后';
      case 4:
        return '活动';
      case 5:
        return '会员';
      default:
        return '通知';
    }
  }

  /// 消息类型图标颜色
  int get typeColor {
    switch (messageType) {
      case 1:
        return 0xff1AAD19; // 绿色-订单
      case 2:
        return 0xffFA9D3B; // 橙色-支付
      case 3:
        return 0xffFF6B6B; // 红色-售后
      case 4:
        return 0xff5AC8FA; // 蓝色-活动
      case 5:
        return 0xffAF52DE; // 紫色-会员
      default:
        return 0xff888888; // 灰色-默认
    }
  }
}

/// 未读消息数响应
class UnreadCountModel {
  final int code;
  final String? message;
  final int unreadCount;

  UnreadCountModel({
    required this.code,
    this.message,
    required this.unreadCount,
  });

  factory UnreadCountModel.fromJson(Map<String, dynamic> json) {
    return UnreadCountModel(
      code: json['code'] ?? 0,
      message: json['message'],
      unreadCount: json['unreadCount'] ?? 0,
    );
  }
}
