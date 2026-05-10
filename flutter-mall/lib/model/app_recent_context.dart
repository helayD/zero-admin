enum AppRecentTargetType {
  home,
  cart,
  orderList,
  productDetail,
  orderDetail,
  settings,
  commentCompose,
  couponList,
  couponCenter,
  afterSalesApply,
  activity,
  digitalCardAssetList,
  digitalCardAssetDetail,
  digitalCardClaim,
  subject,
  preferredArea,
}

AppRecentTargetType? appRecentTargetTypeFromValue(String? value) {
  switch (value) {
    case 'home':
      return AppRecentTargetType.home;
    case 'cart':
      return AppRecentTargetType.cart;
    case 'order_list':
      return AppRecentTargetType.orderList;
    case 'product_detail':
      return AppRecentTargetType.productDetail;
    case 'order_detail':
      return AppRecentTargetType.orderDetail;
    case 'settings':
      return AppRecentTargetType.settings;
    case 'comment_compose':
      return AppRecentTargetType.commentCompose;
    case 'coupon_list':
      return AppRecentTargetType.couponList;
    case 'coupon_center':
      return AppRecentTargetType.couponCenter;
    case 'after_sales_apply':
      return AppRecentTargetType.afterSalesApply;
    case 'activity':
      return AppRecentTargetType.activity;
    case 'digital_card_asset_list':
      return AppRecentTargetType.digitalCardAssetList;
    case 'digital_card_asset_detail':
      return AppRecentTargetType.digitalCardAssetDetail;
    case 'digital_card_claim':
      return AppRecentTargetType.digitalCardClaim;
    case 'subject':
      return AppRecentTargetType.subject;
    case 'preferred_area':
      return AppRecentTargetType.preferredArea;
    default:
      return null;
  }
}

String appRecentTargetTypeToValue(AppRecentTargetType type) {
  switch (type) {
    case AppRecentTargetType.home:
      return 'home';
    case AppRecentTargetType.cart:
      return 'cart';
    case AppRecentTargetType.orderList:
      return 'order_list';
    case AppRecentTargetType.productDetail:
      return 'product_detail';
    case AppRecentTargetType.orderDetail:
      return 'order_detail';
    case AppRecentTargetType.settings:
      return 'settings';
    case AppRecentTargetType.commentCompose:
      return 'comment_compose';
    case AppRecentTargetType.couponList:
      return 'coupon_list';
    case AppRecentTargetType.couponCenter:
      return 'coupon_center';
    case AppRecentTargetType.afterSalesApply:
      return 'after_sales_apply';
    case AppRecentTargetType.activity:
      return 'activity';
    case AppRecentTargetType.digitalCardAssetList:
      return 'digital_card_asset_list';
    case AppRecentTargetType.digitalCardAssetDetail:
      return 'digital_card_asset_detail';
    case AppRecentTargetType.digitalCardClaim:
      return 'digital_card_claim';
    case AppRecentTargetType.subject:
      return 'subject';
    case AppRecentTargetType.preferredArea:
      return 'preferred_area';
  }
}

class AppRecentContext {
  static const int currentVersion = 3;

  final int version;
  final AppRecentTargetType targetType;
  final int? targetId;
  final int? tabIndex;
  final String source;
  final bool requiresAuth;
  final int? ownerMemberId;
  final DateTime capturedAt;
  final DateTime lastValidatedAt;
  final AppRecentTargetType fallbackType;
  final int? fallbackTargetId;
  final int? fallbackTabIndex;
  final String intentType;
  final String minAppVersion;
  final String intentId;
  final DateTime? issuedAt;
  final String failureReason;
  final String recoveryHint;
  final bool blocked;

  const AppRecentContext({
    required this.version,
    required this.targetType,
    required this.targetId,
    required this.tabIndex,
    required this.source,
    required this.requiresAuth,
    this.ownerMemberId,
    required this.capturedAt,
    required this.lastValidatedAt,
    required this.fallbackType,
    this.fallbackTargetId,
    required this.fallbackTabIndex,
    required this.intentType,
    required this.minAppVersion,
    required this.intentId,
    required this.issuedAt,
    required this.failureReason,
    required this.recoveryHint,
    required this.blocked,
  });

  factory AppRecentContext.create({
    required AppRecentTargetType targetType,
    int? targetId,
    int? tabIndex,
    required String source,
    required bool requiresAuth,
    int? ownerMemberId,
    AppRecentTargetType fallbackType = AppRecentTargetType.home,
    int? fallbackTargetId,
    int? fallbackTabIndex,
    DateTime? capturedAt,
    DateTime? lastValidatedAt,
  }) {
    final now = DateTime.now();
    return AppRecentContext(
      version: currentVersion,
      targetType: targetType,
      targetId: targetId,
      tabIndex: tabIndex,
      source: source,
      requiresAuth: requiresAuth,
      ownerMemberId: ownerMemberId,
      capturedAt: capturedAt ?? now,
      lastValidatedAt: lastValidatedAt ?? now,
      fallbackType: fallbackType,
      fallbackTargetId: fallbackTargetId,
      fallbackTabIndex: fallbackTabIndex,
      intentType: '',
      minAppVersion: '',
      intentId: '',
      issuedAt: null,
      failureReason: '',
      recoveryHint: '',
      blocked: false,
    );
  }

  factory AppRecentContext.createRecall({
    required String intentType,
    required AppRecentTargetType targetType,
    int? targetId,
    int? tabIndex,
    required AppRecentTargetType fallbackType,
    int? fallbackTargetId,
    int? fallbackTabIndex,
    required String source,
    required bool requiresAuth,
    String minAppVersion = '',
    required String intentId,
    DateTime? issuedAt,
    String failureReason = '',
    String recoveryHint = '',
    bool blocked = false,
    int? ownerMemberId,
    DateTime? capturedAt,
    DateTime? lastValidatedAt,
  }) {
    final now = DateTime.now();
    return AppRecentContext(
      version: currentVersion,
      targetType: targetType,
      targetId: targetId,
      tabIndex: tabIndex,
      source: source,
      requiresAuth: requiresAuth,
      ownerMemberId: ownerMemberId,
      capturedAt: capturedAt ?? now,
      lastValidatedAt: lastValidatedAt ?? now,
      fallbackType: fallbackType,
      fallbackTargetId: fallbackTargetId,
      fallbackTabIndex: fallbackTabIndex,
      intentType: intentType,
      minAppVersion: minAppVersion,
      intentId: intentId,
      issuedAt: issuedAt ?? now,
      failureReason: failureReason,
      recoveryHint: recoveryHint,
      blocked: blocked,
    );
  }

  factory AppRecentContext.fromRecallPayload(
    Map<String, dynamic> json, {
    String defaultSource = 'member_message',
  }) {
    final targetType =
        appRecentTargetTypeFromValue(json['targetType']?.toString());
    final fallbackType =
        appRecentTargetTypeFromValue(json['fallbackType']?.toString());
    if (targetType == null || fallbackType == null) {
      throw const FormatException('invalid recall payload');
    }
    final now = DateTime.now();
    return AppRecentContext.createRecall(
      intentType: json['intentType']?.toString() ?? 'message_recall',
      targetType: targetType,
      targetId: _normalizeTargetId(targetType, _parseInt(json['targetId'])),
      tabIndex: _normalizeTabIndex(targetType, _parseInt(json['targetTab'])),
      fallbackType: fallbackType,
      fallbackTargetId:
          _normalizeTargetId(fallbackType, _parseInt(json['fallbackTargetId'])),
      fallbackTabIndex:
          _normalizeTabIndex(fallbackType, _parseInt(json['fallbackTab'])),
      source: json['source']?.toString() ?? defaultSource,
      requiresAuth: json['requiresAuth'] == true,
      minAppVersion: json['minAppVersion']?.toString() ?? '',
      intentId: json['intentId']?.toString() ?? '',
      issuedAt: _parseDateTime(json['issuedAt']),
      failureReason: json['failureReason']?.toString() ?? '',
      recoveryHint: json['recoveryHint']?.toString() ?? '',
      blocked: json['blocked'] == true,
      capturedAt: now,
      lastValidatedAt: now,
    );
  }

  String get targetTypeValue => appRecentTargetTypeToValue(targetType);

  String get fallbackTypeValue => appRecentTargetTypeToValue(fallbackType);

  bool get isRecallIntent =>
      intentType.trim().isNotEmpty ||
      intentId.trim().isNotEmpty ||
      issuedAt != null ||
      minAppVersion.trim().isNotEmpty ||
      failureReason.trim().isNotEmpty ||
      blocked;

  AppRecentContext copyWith({
    int? version,
    AppRecentTargetType? targetType,
    int? targetId,
    int? tabIndex,
    String? source,
    bool? requiresAuth,
    int? ownerMemberId,
    DateTime? capturedAt,
    DateTime? lastValidatedAt,
    AppRecentTargetType? fallbackType,
    int? fallbackTargetId,
    int? fallbackTabIndex,
    String? intentType,
    String? minAppVersion,
    String? intentId,
    DateTime? issuedAt,
    String? failureReason,
    String? recoveryHint,
    bool? blocked,
  }) {
    return AppRecentContext(
      version: version ?? this.version,
      targetType: targetType ?? this.targetType,
      targetId: targetId ?? this.targetId,
      tabIndex: tabIndex ?? this.tabIndex,
      source: source ?? this.source,
      requiresAuth: requiresAuth ?? this.requiresAuth,
      ownerMemberId: ownerMemberId ?? this.ownerMemberId,
      capturedAt: capturedAt ?? this.capturedAt,
      lastValidatedAt: lastValidatedAt ?? this.lastValidatedAt,
      fallbackType: fallbackType ?? this.fallbackType,
      fallbackTargetId: fallbackTargetId ?? this.fallbackTargetId,
      fallbackTabIndex: fallbackTabIndex ?? this.fallbackTabIndex,
      intentType: intentType ?? this.intentType,
      minAppVersion: minAppVersion ?? this.minAppVersion,
      intentId: intentId ?? this.intentId,
      issuedAt: issuedAt ?? this.issuedAt,
      failureReason: failureReason ?? this.failureReason,
      recoveryHint: recoveryHint ?? this.recoveryHint,
      blocked: blocked ?? this.blocked,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'version': version,
      'targetType': targetTypeValue,
      'targetId': targetId,
      'tabIndex': tabIndex,
      'source': source,
      'requiresAuth': requiresAuth,
      'ownerMemberId': ownerMemberId,
      'capturedAt': capturedAt.toIso8601String(),
      'lastValidatedAt': lastValidatedAt.toIso8601String(),
      'fallbackType': fallbackTypeValue,
      'fallbackTargetId': fallbackTargetId,
      'fallbackTabIndex': fallbackTabIndex,
      'intentType': intentType,
      'minAppVersion': minAppVersion,
      'intentId': intentId,
      'issuedAt': issuedAt?.toIso8601String(),
      'failureReason': failureReason,
      'recoveryHint': recoveryHint,
      'blocked': blocked,
    };
  }

  static AppRecentContext? tryParse(dynamic value) {
    if (value is! Map) {
      return null;
    }
    try {
      return AppRecentContext.fromJson(Map<String, dynamic>.from(value));
    } catch (_) {
      return null;
    }
  }

  factory AppRecentContext.fromJson(Map<String, dynamic> json) {
    final targetType =
        appRecentTargetTypeFromValue(json['targetType']?.toString());
    final fallbackType =
        appRecentTargetTypeFromValue(json['fallbackType']?.toString()) ??
            AppRecentTargetType.home;
    if (targetType == null) {
      throw const FormatException('invalid target type');
    }
    return AppRecentContext(
      version: _parseInt(json['version']) ?? 0,
      targetType: targetType,
      targetId: _parseInt(json['targetId']),
      tabIndex: _parseInt(json['tabIndex']),
      source: json['source']?.toString() ?? '',
      requiresAuth: json['requiresAuth'] == true,
      ownerMemberId: _parseInt(json['ownerMemberId']),
      capturedAt: _parseDateTime(json['capturedAt']) ??
          DateTime.fromMillisecondsSinceEpoch(0),
      lastValidatedAt: _parseDateTime(json['lastValidatedAt']) ??
          DateTime.fromMillisecondsSinceEpoch(0),
      fallbackType: fallbackType,
      fallbackTargetId: _parseInt(json['fallbackTargetId']),
      fallbackTabIndex: _parseInt(json['fallbackTabIndex']),
      intentType: json['intentType']?.toString() ?? '',
      minAppVersion: json['minAppVersion']?.toString() ?? '',
      intentId: json['intentId']?.toString() ?? '',
      issuedAt: _parseDateTime(json['issuedAt']),
      failureReason: json['failureReason']?.toString() ?? '',
      recoveryHint: json['recoveryHint']?.toString() ?? '',
      blocked: json['blocked'] == true,
    );
  }

  String? validateReason() {
    if (version != currentVersion) {
      return 'unsupported_version';
    }
    if (source.trim().isEmpty) {
      return 'missing_source';
    }
    if (capturedAt.millisecondsSinceEpoch <= 0 ||
        lastValidatedAt.millisecondsSinceEpoch <= 0) {
      return 'missing_timestamp';
    }
    if (!_isValidTarget(targetType, targetId, tabIndex)) {
      return 'invalid_target_id';
    }
    if (!_isValidFallback(fallbackType, fallbackTargetId, fallbackTabIndex)) {
      return 'invalid_fallback';
    }
    if (isRecallIntent) {
      if (intentType.trim().isEmpty) {
        return 'missing_intent_type';
      }
      if (intentId.trim().isEmpty) {
        return 'missing_intent_id';
      }
      if (issuedAt == null || issuedAt!.millisecondsSinceEpoch <= 0) {
        return 'missing_issued_at';
      }
    }
    return null;
  }

  bool get isRecoverable => validateReason() == null;

  static bool _isValidTarget(
    AppRecentTargetType type,
    int? targetId,
    int? tabIndex,
  ) {
    if (!_isValidTargetId(type, targetId)) {
      return false;
    }
    if (!_isValidTabIndex(type, tabIndex)) {
      return false;
    }
    return true;
  }

  // Story 10.7 Review Fix: 修复 refactor 残留的 pre-existing bugs：
  // 1) _isValidTargetId 旧版漏列 orderDetail/productDetail/commentCompose/afterSalesApply/cart/orderList
  // 2) _isValidTabIndex 旧版重复 case（digitalCardAssetDetail/activity/productDetail/orderDetail/commentCompose/afterSalesApply 均出现两次）
  // 3) _targetSupportsTab 旧版 default 分支塞入了错位的 _isValidFallback body，引用 static 方法里不存在的实例字段
  // 4) _parseInt / _isValidFallback 仅有调用点，缺实现
  static bool _isValidTargetId(AppRecentTargetType type, int? targetId) {
    switch (type) {
      // 这些 target 可缺省（null）或需为正数
      case AppRecentTargetType.home:
      case AppRecentTargetType.cart:
      case AppRecentTargetType.orderList:
      case AppRecentTargetType.settings:
      case AppRecentTargetType.couponList:
      case AppRecentTargetType.couponCenter:
      case AppRecentTargetType.digitalCardAssetList:
      case AppRecentTargetType.digitalCardClaim:
      case AppRecentTargetType.activity:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return targetId == null || targetId > 0;
      // 这些 target 必须携带正数 targetId
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
      case AppRecentTargetType.commentCompose:
      case AppRecentTargetType.afterSalesApply:
      case AppRecentTargetType.digitalCardAssetDetail:
        return targetId != null && targetId > 0;
    }
  }

  static bool _isValidTabIndex(AppRecentTargetType type, int? tabIndex) {
    switch (type) {
      case AppRecentTargetType.home:
        return tabIndex == null || tabIndex == 0;
      case AppRecentTargetType.cart:
        return tabIndex == null || tabIndex == 2;
      case AppRecentTargetType.orderList:
        return tabIndex != null && tabIndex >= 0 && tabIndex <= 4;
      case AppRecentTargetType.couponList:
        return tabIndex == null || (tabIndex >= 0 && tabIndex <= 2);
      // 其余 target 不支持 tab，不允许非空 tabIndex
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
      case AppRecentTargetType.settings:
      case AppRecentTargetType.commentCompose:
      case AppRecentTargetType.couponCenter:
      case AppRecentTargetType.afterSalesApply:
      case AppRecentTargetType.activity:
      case AppRecentTargetType.digitalCardAssetList:
      case AppRecentTargetType.digitalCardAssetDetail:
      case AppRecentTargetType.digitalCardClaim:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return tabIndex == null;
    }
  }

  bool _isValidFallback(
    AppRecentTargetType fallbackType,
    int? fallbackTargetId,
    int? fallbackTabIndex,
  ) {
    return _isValidTarget(fallbackType, fallbackTargetId, fallbackTabIndex);
  }

  static int? _normalizeTargetId(AppRecentTargetType type, int? value) {
    if (value == null) {
      return null;
    }
    // 非正数按未设置处理，防止脏数据导致 _isValidTargetId 判空失败
    if (value <= 0) {
      return null;
    }
    return value;
  }

  static int? _normalizeTabIndex(AppRecentTargetType type, int? value) {
    if (value == null) {
      return null;
    }
    if (_targetSupportsTab(type)) {
      return value;
    }
    // 对于不支持 tab 的 target，忽略任何 tabIndex
    return null;
  }

  static bool _targetSupportsTab(AppRecentTargetType type) {
    switch (type) {
      case AppRecentTargetType.home:
      case AppRecentTargetType.cart:
      case AppRecentTargetType.orderList:
      case AppRecentTargetType.couponList:
        return true;
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
      case AppRecentTargetType.settings:
      case AppRecentTargetType.commentCompose:
      case AppRecentTargetType.couponCenter:
      case AppRecentTargetType.afterSalesApply:
      case AppRecentTargetType.activity:
      case AppRecentTargetType.digitalCardAssetList:
      case AppRecentTargetType.digitalCardAssetDetail:
      case AppRecentTargetType.digitalCardClaim:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return false;
    }
  }

  static int? _parseInt(dynamic value) {
    if (value is int) {
      return value;
    }
    if (value is num) {
      return value.toInt();
    }
    final raw = value?.toString();
    if (raw == null || raw.isEmpty) {
      return null;
    }
    return int.tryParse(raw);
  }

  static DateTime? _parseDateTime(dynamic value) {
    final raw = value?.toString();
    if (raw == null || raw.isEmpty) {
      return null;
    }
    return DateTime.tryParse(raw);
  }
}
