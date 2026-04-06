import 'package:flutter_mall/model/app_recent_context.dart';

enum PermissionScene {
  notificationSubscription,
  commentImage,
  afterSalesProof,
}

enum PermissionType {
  notification,
  photos,
  camera,
}

enum PermissionFlowSource {
  notificationToggle,
  gallery,
  camera,
  settingsReturn,
  lostData,
}

enum PermissionFlowStatus {
  initial,
  granted,
  limited,
  denied,
  permanentlyDenied,
  notRequired,
  settingsReturnPending,
  pickerLostData,
}

enum PermissionFallbackAction {
  stayOnCurrentTask,
  openMessageCenter,
  continueWithoutImage,
  switchToGallery,
  switchToCamera,
}

PermissionScene? permissionSceneFromValue(String? value) {
  switch (value) {
    case 'notification_subscription':
      return PermissionScene.notificationSubscription;
    case 'comment_image':
      return PermissionScene.commentImage;
    case 'after_sales_proof':
      return PermissionScene.afterSalesProof;
    default:
      return null;
  }
}

String permissionSceneToValue(PermissionScene scene) {
  switch (scene) {
    case PermissionScene.notificationSubscription:
      return 'notification_subscription';
    case PermissionScene.commentImage:
      return 'comment_image';
    case PermissionScene.afterSalesProof:
      return 'after_sales_proof';
  }
}

PermissionType? permissionTypeFromValue(String? value) {
  switch (value) {
    case 'notification':
      return PermissionType.notification;
    case 'photos':
      return PermissionType.photos;
    case 'camera':
      return PermissionType.camera;
    default:
      return null;
  }
}

String permissionTypeToValue(PermissionType type) {
  switch (type) {
    case PermissionType.notification:
      return 'notification';
    case PermissionType.photos:
      return 'photos';
    case PermissionType.camera:
      return 'camera';
  }
}

PermissionFlowSource? permissionFlowSourceFromValue(String? value) {
  switch (value) {
    case 'notification_toggle':
      return PermissionFlowSource.notificationToggle;
    case 'gallery':
      return PermissionFlowSource.gallery;
    case 'camera':
      return PermissionFlowSource.camera;
    case 'settings_return':
      return PermissionFlowSource.settingsReturn;
    case 'lost_data':
      return PermissionFlowSource.lostData;
    default:
      return null;
  }
}

String permissionFlowSourceToValue(PermissionFlowSource source) {
  switch (source) {
    case PermissionFlowSource.notificationToggle:
      return 'notification_toggle';
    case PermissionFlowSource.gallery:
      return 'gallery';
    case PermissionFlowSource.camera:
      return 'camera';
    case PermissionFlowSource.settingsReturn:
      return 'settings_return';
    case PermissionFlowSource.lostData:
      return 'lost_data';
  }
}

PermissionFlowStatus? permissionFlowStatusFromValue(String? value) {
  switch (value) {
    case 'initial':
      return PermissionFlowStatus.initial;
    case 'granted':
      return PermissionFlowStatus.granted;
    case 'limited':
      return PermissionFlowStatus.limited;
    case 'denied':
      return PermissionFlowStatus.denied;
    case 'permanently_denied':
      return PermissionFlowStatus.permanentlyDenied;
    case 'not_required':
      return PermissionFlowStatus.notRequired;
    case 'settings_return_pending':
      return PermissionFlowStatus.settingsReturnPending;
    case 'picker_lost_data':
      return PermissionFlowStatus.pickerLostData;
    default:
      return null;
  }
}

String permissionFlowStatusToValue(PermissionFlowStatus status) {
  switch (status) {
    case PermissionFlowStatus.initial:
      return 'initial';
    case PermissionFlowStatus.granted:
      return 'granted';
    case PermissionFlowStatus.limited:
      return 'limited';
    case PermissionFlowStatus.denied:
      return 'denied';
    case PermissionFlowStatus.permanentlyDenied:
      return 'permanently_denied';
    case PermissionFlowStatus.notRequired:
      return 'not_required';
    case PermissionFlowStatus.settingsReturnPending:
      return 'settings_return_pending';
    case PermissionFlowStatus.pickerLostData:
      return 'picker_lost_data';
  }
}

PermissionFallbackAction? permissionFallbackActionFromValue(String? value) {
  switch (value) {
    case 'stay_on_current_task':
      return PermissionFallbackAction.stayOnCurrentTask;
    case 'open_message_center':
      return PermissionFallbackAction.openMessageCenter;
    case 'continue_without_image':
      return PermissionFallbackAction.continueWithoutImage;
    case 'switch_to_gallery':
      return PermissionFallbackAction.switchToGallery;
    case 'switch_to_camera':
      return PermissionFallbackAction.switchToCamera;
    default:
      return null;
  }
}

String permissionFallbackActionToValue(PermissionFallbackAction action) {
  switch (action) {
    case PermissionFallbackAction.stayOnCurrentTask:
      return 'stay_on_current_task';
    case PermissionFallbackAction.openMessageCenter:
      return 'open_message_center';
    case PermissionFallbackAction.continueWithoutImage:
      return 'continue_without_image';
    case PermissionFallbackAction.switchToGallery:
      return 'switch_to_gallery';
    case PermissionFallbackAction.switchToCamera:
      return 'switch_to_camera';
  }
}

class PermissionFlowContext {
  final PermissionScene scene;
  final PermissionType permissionType;
  final PermissionFlowStatus status;
  final PermissionFlowSource source;
  final AppRecentTargetType returnTarget;
  final int? returnTargetId;
  final int? orderId;
  final int? productId;
  final PermissionFallbackAction fallbackAction;
  final String intentId;
  final String recoveryId;
  final bool requiresAuth;
  final bool retryOnResume;
  final bool requestedToggleEnabled;
  final DateTime createdAt;

  const PermissionFlowContext({
    required this.scene,
    required this.permissionType,
    required this.status,
    required this.source,
    required this.returnTarget,
    this.returnTargetId,
    this.orderId,
    this.productId,
    required this.fallbackAction,
    required this.intentId,
    required this.recoveryId,
    required this.requiresAuth,
    required this.retryOnResume,
    required this.requestedToggleEnabled,
    required this.createdAt,
  });

  factory PermissionFlowContext.create({
    required PermissionScene scene,
    required PermissionType permissionType,
    required PermissionFlowSource source,
    required AppRecentTargetType returnTarget,
    int? returnTargetId,
    int? orderId,
    int? productId,
    required PermissionFallbackAction fallbackAction,
    required String intentId,
    required String recoveryId,
    bool requiresAuth = false,
    bool retryOnResume = false,
    bool requestedToggleEnabled = false,
    PermissionFlowStatus status = PermissionFlowStatus.initial,
    DateTime? createdAt,
  }) {
    return PermissionFlowContext(
      scene: scene,
      permissionType: permissionType,
      status: status,
      source: source,
      returnTarget: returnTarget,
      returnTargetId: returnTargetId,
      orderId: orderId,
      productId: productId,
      fallbackAction: fallbackAction,
      intentId: intentId,
      recoveryId: recoveryId,
      requiresAuth: requiresAuth,
      retryOnResume: retryOnResume,
      requestedToggleEnabled: requestedToggleEnabled,
      createdAt: createdAt ?? DateTime.now(),
    );
  }

  factory PermissionFlowContext.fromJson(Map<String, dynamic> json) {
    final scene = permissionSceneFromValue(json['scene']?.toString());
    final permissionType =
        permissionTypeFromValue(json['permissionType']?.toString());
    final status = permissionFlowStatusFromValue(json['status']?.toString());
    final source = permissionFlowSourceFromValue(json['source']?.toString());
    final returnTarget =
        appRecentTargetTypeFromValue(json['returnTarget']?.toString());
    final fallbackAction =
        permissionFallbackActionFromValue(json['fallbackAction']?.toString());
    if (scene == null ||
        permissionType == null ||
        status == null ||
        source == null ||
        returnTarget == null ||
        fallbackAction == null) {
      throw const FormatException('invalid permission flow context');
    }
    return PermissionFlowContext(
      scene: scene,
      permissionType: permissionType,
      status: status,
      source: source,
      returnTarget: returnTarget,
      returnTargetId: _parseInt(json['returnTargetId']),
      orderId: _parseInt(json['orderId']),
      productId: _parseInt(json['productId']),
      fallbackAction: fallbackAction,
      intentId: json['intentId']?.toString() ?? '',
      recoveryId: json['recoveryId']?.toString() ?? '',
      requiresAuth: json['requiresAuth'] == true,
      retryOnResume: json['retryOnResume'] == true,
      requestedToggleEnabled: json['requestedToggleEnabled'] == true,
      createdAt: DateTime.tryParse(json['createdAt']?.toString() ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'scene': permissionSceneToValue(scene),
      'permissionType': permissionTypeToValue(permissionType),
      'status': permissionFlowStatusToValue(status),
      'source': permissionFlowSourceToValue(source),
      'returnTarget': appRecentTargetTypeToValue(returnTarget),
      'returnTargetId': returnTargetId,
      'orderId': orderId,
      'productId': productId,
      'fallbackAction': permissionFallbackActionToValue(fallbackAction),
      'intentId': intentId,
      'recoveryId': recoveryId,
      'requiresAuth': requiresAuth,
      'retryOnResume': retryOnResume,
      'requestedToggleEnabled': requestedToggleEnabled,
      'createdAt': createdAt.toIso8601String(),
    };
  }

  PermissionFlowContext copyWith({
    PermissionScene? scene,
    PermissionType? permissionType,
    PermissionFlowStatus? status,
    PermissionFlowSource? source,
    AppRecentTargetType? returnTarget,
    int? returnTargetId,
    int? orderId,
    int? productId,
    PermissionFallbackAction? fallbackAction,
    String? intentId,
    String? recoveryId,
    bool? requiresAuth,
    bool? retryOnResume,
    bool? requestedToggleEnabled,
    DateTime? createdAt,
  }) {
    return PermissionFlowContext(
      scene: scene ?? this.scene,
      permissionType: permissionType ?? this.permissionType,
      status: status ?? this.status,
      source: source ?? this.source,
      returnTarget: returnTarget ?? this.returnTarget,
      returnTargetId: returnTargetId ?? this.returnTargetId,
      orderId: orderId ?? this.orderId,
      productId: productId ?? this.productId,
      fallbackAction: fallbackAction ?? this.fallbackAction,
      intentId: intentId ?? this.intentId,
      recoveryId: recoveryId ?? this.recoveryId,
      requiresAuth: requiresAuth ?? this.requiresAuth,
      retryOnResume: retryOnResume ?? this.retryOnResume,
      requestedToggleEnabled:
          requestedToggleEnabled ?? this.requestedToggleEnabled,
      createdAt: createdAt ?? this.createdAt,
    );
  }

  static PermissionFlowContext? tryParse(dynamic value) {
    if (value is! Map) {
      return null;
    }
    try {
      return PermissionFlowContext.fromJson(Map<String, dynamic>.from(value));
    } catch (_) {
      return null;
    }
  }

  bool matchesReturnTarget(
    AppRecentTargetType targetType, {
    int? targetId,
  }) {
    if (returnTarget != targetType) {
      return false;
    }
    if (targetId == null) {
      return true;
    }
    return returnTargetId == targetId;
  }

  static int? _parseInt(dynamic value) {
    if (value is int) {
      return value;
    }
    return int.tryParse(value?.toString() ?? '');
  }
}

class CommentDraftSnapshot {
  final int orderId;
  final int productId;
  final String productName;
  final String productPic;
  final String productAttribute;
  final String memberNickName;
  final int starRating;
  final String content;
  final List<String> pics;
  final DateTime updatedAt;

  const CommentDraftSnapshot({
    required this.orderId,
    required this.productId,
    required this.productName,
    required this.productPic,
    required this.productAttribute,
    required this.memberNickName,
    required this.starRating,
    required this.content,
    required this.pics,
    required this.updatedAt,
  });

  factory CommentDraftSnapshot.fromJson(Map<String, dynamic> json) {
    return CommentDraftSnapshot(
      orderId: PermissionFlowContext._parseInt(json['orderId']) ?? 0,
      productId: PermissionFlowContext._parseInt(json['productId']) ?? 0,
      productName: json['productName']?.toString() ?? '',
      productPic: json['productPic']?.toString() ?? '',
      productAttribute: json['productAttribute']?.toString() ?? '',
      memberNickName: json['memberNickName']?.toString() ?? '',
      starRating: PermissionFlowContext._parseInt(json['starRating']) ?? 5,
      content: json['content']?.toString() ?? '',
      pics: (json['pics'] as List<dynamic>? ?? <dynamic>[])
          .map((item) => item.toString())
          .toList(),
      updatedAt: DateTime.tryParse(json['updatedAt']?.toString() ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'orderId': orderId,
      'productId': productId,
      'productName': productName,
      'productPic': productPic,
      'productAttribute': productAttribute,
      'memberNickName': memberNickName,
      'starRating': starRating,
      'content': content,
      'pics': pics,
      'updatedAt': updatedAt.toIso8601String(),
    };
  }

  CommentDraftSnapshot copyWith({
    int? orderId,
    int? productId,
    String? productName,
    String? productPic,
    String? productAttribute,
    String? memberNickName,
    int? starRating,
    String? content,
    List<String>? pics,
    DateTime? updatedAt,
  }) {
    return CommentDraftSnapshot(
      orderId: orderId ?? this.orderId,
      productId: productId ?? this.productId,
      productName: productName ?? this.productName,
      productPic: productPic ?? this.productPic,
      productAttribute: productAttribute ?? this.productAttribute,
      memberNickName: memberNickName ?? this.memberNickName,
      starRating: starRating ?? this.starRating,
      content: content ?? this.content,
      pics: pics ?? this.pics,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  static CommentDraftSnapshot? tryParse(dynamic value) {
    if (value is! Map) {
      return null;
    }
    try {
      return CommentDraftSnapshot.fromJson(Map<String, dynamic>.from(value));
    } catch (_) {
      return null;
    }
  }
}

class AfterSalesDraftSnapshot {
  final int orderId;
  final int typeValue;
  final int? reasonId;
  final String reasonName;
  final String description;
  final List<String> proofPics;
  final DateTime updatedAt;

  const AfterSalesDraftSnapshot({
    required this.orderId,
    required this.typeValue,
    required this.reasonId,
    required this.reasonName,
    required this.description,
    required this.proofPics,
    required this.updatedAt,
  });

  factory AfterSalesDraftSnapshot.fromJson(Map<String, dynamic> json) {
    return AfterSalesDraftSnapshot(
      orderId: PermissionFlowContext._parseInt(json['orderId']) ?? 0,
      typeValue: PermissionFlowContext._parseInt(json['typeValue']) ?? 0,
      reasonId: PermissionFlowContext._parseInt(json['reasonId']),
      reasonName: json['reasonName']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      proofPics: (json['proofPics'] as List<dynamic>? ?? <dynamic>[])
          .map((item) => item.toString())
          .toList(),
      updatedAt: DateTime.tryParse(json['updatedAt']?.toString() ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'orderId': orderId,
      'typeValue': typeValue,
      'reasonId': reasonId,
      'reasonName': reasonName,
      'description': description,
      'proofPics': proofPics,
      'updatedAt': updatedAt.toIso8601String(),
    };
  }

  AfterSalesDraftSnapshot copyWith({
    int? orderId,
    int? typeValue,
    int? reasonId,
    String? reasonName,
    String? description,
    List<String>? proofPics,
    DateTime? updatedAt,
  }) {
    return AfterSalesDraftSnapshot(
      orderId: orderId ?? this.orderId,
      typeValue: typeValue ?? this.typeValue,
      reasonId: reasonId ?? this.reasonId,
      reasonName: reasonName ?? this.reasonName,
      description: description ?? this.description,
      proofPics: proofPics ?? this.proofPics,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  static AfterSalesDraftSnapshot? tryParse(dynamic value) {
    if (value is! Map) {
      return null;
    }
    try {
      return AfterSalesDraftSnapshot.fromJson(Map<String, dynamic>.from(value));
    } catch (_) {
      return null;
    }
  }
}

class PermissionLostMediaSnapshot {
  final String recoveryId;
  final PermissionScene scene;
  final PermissionFlowSource source;
  final List<String> mediaBase64List;
  final String errorMessage;
  final DateTime capturedAt;

  const PermissionLostMediaSnapshot({
    required this.recoveryId,
    required this.scene,
    required this.source,
    required this.mediaBase64List,
    required this.errorMessage,
    required this.capturedAt,
  });

  factory PermissionLostMediaSnapshot.fromJson(Map<String, dynamic> json) {
    final scene = permissionSceneFromValue(json['scene']?.toString());
    final source = permissionFlowSourceFromValue(json['source']?.toString());
    if (scene == null || source == null) {
      throw const FormatException('invalid lost media snapshot');
    }
    return PermissionLostMediaSnapshot(
      recoveryId: json['recoveryId']?.toString() ?? '',
      scene: scene,
      source: source,
      mediaBase64List:
          (json['mediaBase64List'] as List<dynamic>? ?? <dynamic>[])
              .map((item) => item.toString())
              .toList(),
      errorMessage: json['errorMessage']?.toString() ?? '',
      capturedAt: DateTime.tryParse(json['capturedAt']?.toString() ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'recoveryId': recoveryId,
      'scene': permissionSceneToValue(scene),
      'source': permissionFlowSourceToValue(source),
      'mediaBase64List': mediaBase64List,
      'errorMessage': errorMessage,
      'capturedAt': capturedAt.toIso8601String(),
    };
  }

  static PermissionLostMediaSnapshot? tryParse(dynamic value) {
    if (value is! Map) {
      return null;
    }
    try {
      return PermissionLostMediaSnapshot.fromJson(
        Map<String, dynamic>.from(value),
      );
    } catch (_) {
      return null;
    }
  }
}
