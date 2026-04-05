enum AppRecentTargetType {
  home,
  cart,
  orderList,
  productDetail,
  orderDetail,
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
  }
}

class AppRecentContext {
  static const int currentVersion = 2;

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
  final int? fallbackTabIndex;

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
    required this.fallbackTabIndex,
  });

  factory AppRecentContext.create({
    required AppRecentTargetType targetType,
    int? targetId,
    int? tabIndex,
    required String source,
    required bool requiresAuth,
    int? ownerMemberId,
    AppRecentTargetType fallbackType = AppRecentTargetType.home,
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
      fallbackTabIndex: fallbackTabIndex,
    );
  }

  String get targetTypeValue => appRecentTargetTypeToValue(targetType);

  String get fallbackTypeValue => appRecentTargetTypeToValue(fallbackType);

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
    int? fallbackTabIndex,
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
      fallbackTabIndex: fallbackTabIndex ?? this.fallbackTabIndex,
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
      'fallbackTabIndex': fallbackTabIndex,
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
    final targetType = appRecentTargetTypeFromValue(json['targetType']?.toString());
    final fallbackType = appRecentTargetTypeFromValue(json['fallbackType']?.toString()) ?? AppRecentTargetType.home;
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
      capturedAt: _parseDateTime(json['capturedAt']) ?? DateTime.fromMillisecondsSinceEpoch(0),
      lastValidatedAt: _parseDateTime(json['lastValidatedAt']) ?? DateTime.fromMillisecondsSinceEpoch(0),
      fallbackType: fallbackType,
      fallbackTabIndex: _parseInt(json['fallbackTabIndex']),
    );
  }

  String? validateReason() {
    if (version != currentVersion) {
      return 'unsupported_version';
    }
    if (source.trim().isEmpty) {
      return 'missing_source';
    }
    if (capturedAt.millisecondsSinceEpoch <= 0 || lastValidatedAt.millisecondsSinceEpoch <= 0) {
      return 'missing_timestamp';
    }
    if (!_isValidTabIndex(targetType, tabIndex)) {
      return 'invalid_tab_index';
    }
    if (!_isValidTargetId(targetType, targetId)) {
      return 'invalid_target_id';
    }
    if (!_isValidFallback(fallbackType, fallbackTabIndex)) {
      return 'invalid_fallback';
    }
    return null;
  }

  bool get isRecoverable => validateReason() == null;

  static bool _isValidTargetId(AppRecentTargetType type, int? targetId) {
    switch (type) {
      case AppRecentTargetType.home:
      case AppRecentTargetType.cart:
      case AppRecentTargetType.orderList:
        return targetId == null || targetId > 0;
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
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
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.orderDetail:
        return tabIndex == null;
    }
  }

  static bool _isValidFallback(AppRecentTargetType fallbackType, int? fallbackTabIndex) {
    if (fallbackType == AppRecentTargetType.home) {
      return fallbackTabIndex == null || fallbackTabIndex == 0;
    }
    if (fallbackType == AppRecentTargetType.cart) {
      return fallbackTabIndex == null || fallbackTabIndex == 2;
    }
    if (fallbackType == AppRecentTargetType.orderList) {
      return fallbackTabIndex != null && fallbackTabIndex >= 0 && fallbackTabIndex <= 4;
    }
    return false;
  }

  static int? _parseInt(dynamic value) {
    if (value is int) {
      return value;
    }
    return int.tryParse(value?.toString() ?? '');
  }

  static DateTime? _parseDateTime(dynamic value) {
    final raw = value?.toString();
    if (raw == null || raw.isEmpty) {
      return null;
    }
    return DateTime.tryParse(raw);
  }
}
