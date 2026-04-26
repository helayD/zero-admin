import 'package:flutter_mall/model/app_recent_context.dart';

///
/// 会员消息数据模型
///
class MessageModel {
  final int code;
  final String? message;
  final List<MessageData> data;
  final int total;
  final int pageNum;
  final int pageSize;

  MessageModel({
    required this.code,
    this.message,
    required this.data,
    this.total = 0,
    this.pageNum = 1,
    this.pageSize = 20,
  });

  factory MessageModel.fromJson(Map<String, dynamic> json) {
    return MessageModel(
      code: json['code'] ?? 0,
      message: json['message'],
      data: (json['data'] as List<dynamic>?)
              ?.map((e) => MessageData.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      total: _parseInt(json['total']),
      pageNum: _parseInt(json['pageNum'], fallback: 1),
      pageSize: _parseInt(json['pageSize'], fallback: 20),
    );
  }
}

class MessageDetailModel {
  final int code;
  final String? message;
  final MessageData? data;

  MessageDetailModel({
    required this.code,
    this.message,
    this.data,
  });

  factory MessageDetailModel.fromJson(Map<String, dynamic> json) {
    final rawData = json['data'];
    return MessageDetailModel(
      code: json['code'] ?? 0,
      message: json['message'],
      data: rawData is Map<String, dynamic>
          ? MessageData.fromJson(rawData)
          : rawData is Map
              ? MessageData.fromJson(Map<String, dynamic>.from(rawData))
              : null,
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
  final AppRecentContext? intent;

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
    this.intent,
  });

  factory MessageData.fromJson(Map<String, dynamic> json) {
    final createTime = json['createTime'] != null
        ? DateTime.tryParse(json['createTime'])
        : null;
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
      createTime: createTime,
      readTime:
          json['readTime'] != null ? DateTime.tryParse(json['readTime']) : null,
      intent: _buildRecallIntent(json, createTime),
    );
  }

  MessageData copyWith({
    int? status,
    AppRecentContext? intent,
  }) {
    return MessageData(
      id: id,
      memberId: memberId,
      messageType: messageType,
      title: title,
      content: content,
      imageUrl: imageUrl,
      linkType: linkType,
      linkId: linkId,
      relatedOrderId: relatedOrderId,
      status: status ?? this.status,
      isDelete: isDelete,
      createTime: createTime,
      readTime: readTime,
      intent: intent ?? this.intent,
    );
  }

  static AppRecentContext? _buildRecallIntent(
    Map<String, dynamic> json,
    DateTime? createTime,
  ) {
    final rawIntent = json['intent'];
    if (rawIntent is Map) {
      try {
        return AppRecentContext.fromRecallPayload(
          Map<String, dynamic>.from(rawIntent),
        );
      } catch (_) {}
    }

    return _buildLegacyIntent(json, createTime);
  }

  static AppRecentContext _buildLegacyIntent(
    Map<String, dynamic> json,
    DateTime? createTime,
  ) {
    final linkType = (json['linkType']?.toString() ?? '').trim().toLowerCase();
    final linkId = (json['linkId']?.toString() ?? '').trim();
    final relatedOrderId = json['relatedOrderId'] ?? 0;
    final title = (json['title']?.toString() ?? '').trim();
    final content = (json['content']?.toString() ?? '').trim();
    final issuedAt = createTime ?? DateTime.now();
    final intentId = 'member_message:${json['id'] ?? 0}';

    switch (linkType) {
      case 'order':
      case 'order_detail':
        final orderId =
            _parsePositiveInt(linkId) ?? _positiveOrNull(relatedOrderId);
        if (orderId != null) {
          return AppRecentContext.createRecall(
            intentType: 'order_recall',
            targetType: AppRecentTargetType.orderDetail,
            targetId: orderId,
            fallbackType: AppRecentTargetType.orderList,
            fallbackTabIndex: 1,
            source: 'member_message',
            requiresAuth: true,
            intentId: intentId,
            issuedAt: issuedAt,
          );
        }
        return _invalidIntent(
          targetType: AppRecentTargetType.orderDetail,
          intentType: 'order_recall',
          intentId: intentId,
          issuedAt: issuedAt,
          requiresAuth: true,
          recoveryHint: '订单入口缺少必要信息，已返回订单列表',
          fallbackType: AppRecentTargetType.orderList,
          fallbackTabIndex: 1,
        );
      case 'coupon':
      case 'coupon_list':
        return AppRecentContext.createRecall(
          intentType: _isCouponCenterTitle(title, content)
              ? 'coupon_center_recall'
              : 'coupon_list_recall',
          targetType: _isCouponCenterTitle(title, content)
              ? AppRecentTargetType.couponCenter
              : AppRecentTargetType.couponList,
          fallbackType: _isCouponCenterTitle(title, content)
              ? AppRecentTargetType.couponList
              : AppRecentTargetType.home,
          fallbackTabIndex: _isCouponCenterTitle(title, content) ? 0 : 0,
          source: 'member_message',
          requiresAuth: true,
          intentId: intentId,
          issuedAt: issuedAt,
        );
      case 'coupon_center':
      case 'available_coupon':
      case 'coupon_receive_center':
        return AppRecentContext.createRecall(
          intentType: 'coupon_center_recall',
          targetType: AppRecentTargetType.couponCenter,
          fallbackType: AppRecentTargetType.couponList,
          fallbackTabIndex: 0,
          source: 'member_message',
          requiresAuth: true,
          intentId: intentId,
          issuedAt: issuedAt,
        );
      case 'product':
      case 'product_detail':
        final productId = _parsePositiveInt(linkId);
        if (productId != null) {
          return AppRecentContext.createRecall(
            intentType: 'product_recall',
            targetType: AppRecentTargetType.productDetail,
            targetId: productId,
            fallbackType: AppRecentTargetType.home,
            fallbackTabIndex: 0,
            source: 'member_message',
            requiresAuth: false,
            intentId: intentId,
            issuedAt: issuedAt,
          );
        }
        return _invalidIntent(
          targetType: AppRecentTargetType.productDetail,
          intentType: 'product_recall',
          intentId: intentId,
          issuedAt: issuedAt,
          requiresAuth: false,
          recoveryHint: '商品入口缺少必要信息，已返回首页继续浏览',
        );
      case 'after_sales':
      case 'after_sale':
      case 'after_sales_apply':
        final orderId =
            _parsePositiveInt(linkId) ?? _positiveOrNull(relatedOrderId);
        if (orderId != null) {
          return AppRecentContext.createRecall(
            intentType: 'after_sales_recall',
            targetType: AppRecentTargetType.afterSalesApply,
            targetId: orderId,
            fallbackType: AppRecentTargetType.orderDetail,
            fallbackTargetId: orderId,
            source: 'member_message',
            requiresAuth: true,
            intentId: intentId,
            issuedAt: issuedAt,
          );
        }
        return _invalidIntent(
          targetType: AppRecentTargetType.afterSalesApply,
          intentType: 'after_sales_recall',
          intentId: intentId,
          issuedAt: issuedAt,
          requiresAuth: true,
          recoveryHint: '售后入口缺少必要信息，已返回订单列表',
          fallbackType: AppRecentTargetType.orderList,
          fallbackTabIndex: 1,
        );
      case 'activity':
        return AppRecentContext.createRecall(
          intentType: 'activity_recall',
          targetType: AppRecentTargetType.activity,
          targetId: _parsePositiveInt(linkId),
          fallbackType: AppRecentTargetType.home,
          fallbackTabIndex: 0,
          source: 'member_message',
          requiresAuth: false,
          intentId: intentId,
          issuedAt: issuedAt,
          recoveryHint: '已为你打开活动入口，可继续查看抽卡规则与参与记录',
          blocked: false,
        );
      case 'subject':
        return AppRecentContext.createRecall(
          intentType: 'subject_recall',
          targetType: AppRecentTargetType.subject,
          targetId: _parsePositiveInt(linkId),
          fallbackType: AppRecentTargetType.home,
          fallbackTabIndex: 0,
          source: 'member_message',
          requiresAuth: false,
          intentId: intentId,
          issuedAt: issuedAt,
          failureReason: 'unsupported_target',
          recoveryHint: '当前专题入口暂不可直达，已返回首页继续浏览',
          blocked: true,
        );
      case 'preferred_area':
        return AppRecentContext.createRecall(
          intentType: 'preferred_area_recall',
          targetType: AppRecentTargetType.preferredArea,
          targetId: _parsePositiveInt(linkId),
          fallbackType: AppRecentTargetType.home,
          fallbackTabIndex: 0,
          source: 'member_message',
          requiresAuth: false,
          intentId: intentId,
          issuedAt: issuedAt,
          failureReason: 'unsupported_target',
          recoveryHint: '当前优选专区入口暂不可直达，已返回首页继续浏览',
          blocked: true,
        );
      default:
        if (_positiveOrNull(relatedOrderId) != null) {
          return AppRecentContext.createRecall(
            intentType: 'order_recall',
            targetType: AppRecentTargetType.orderDetail,
            targetId: relatedOrderId,
            fallbackType: AppRecentTargetType.orderList,
            fallbackTabIndex: 1,
            source: 'member_message',
            requiresAuth: true,
            intentId: intentId,
            issuedAt: issuedAt,
          );
        }
        return _invalidIntent(
          targetType: AppRecentTargetType.home,
          intentType: 'message_recall',
          intentId: intentId,
          issuedAt: issuedAt,
          requiresAuth: false,
          recoveryHint: '消息目标无效，已返回首页继续浏览',
        );
    }
  }

  static AppRecentContext _invalidIntent({
    required AppRecentTargetType targetType,
    required String intentType,
    required String intentId,
    required DateTime issuedAt,
    required bool requiresAuth,
    String recoveryHint = '消息目标无效，已返回首页继续浏览',
    AppRecentTargetType fallbackType = AppRecentTargetType.home,
    int? fallbackTargetId,
    int? fallbackTabIndex = 0,
  }) {
    return AppRecentContext.createRecall(
      intentType: intentType,
      targetType: targetType,
      fallbackType: fallbackType,
      fallbackTargetId: fallbackTargetId,
      fallbackTabIndex: fallbackTabIndex,
      source: 'member_message',
      requiresAuth: requiresAuth,
      intentId: intentId,
      issuedAt: issuedAt,
      failureReason: 'invalid_payload',
      recoveryHint: recoveryHint,
      blocked: true,
    );
  }

  static int? _parsePositiveInt(String value) {
    final parsed = int.tryParse(value.trim());
    if (parsed == null || parsed <= 0) {
      return null;
    }
    return parsed;
  }

  static int? _positiveOrNull(dynamic value) {
    if (value is int && value > 0) {
      return value;
    }
    return null;
  }

  static bool _isCouponCenterTitle(String title, String content) {
    const keywords = ['领券', '可领', '领取', '发券', '优惠券到账'];
    final joined = '$title $content';
    for (final keyword in keywords) {
      if (joined.contains(keyword)) {
        return true;
      }
    }
    return false;
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
      unreadCount: _parseInt(json['unreadCount']),
    );
  }
}

int _parseInt(dynamic value, {int fallback = 0}) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse(value?.toString() ?? '') ?? fallback;
}
