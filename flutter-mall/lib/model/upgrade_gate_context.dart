import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/model/direct_checkout.dart';

enum UpgradeRecoveryTargetType {
  recentContext,
  orderConfirm,
  orderPay,
}

UpgradeRecoveryTargetType? upgradeRecoveryTargetTypeFromValue(String? value) {
  switch ((value ?? '').trim()) {
    case 'recent_context':
      return UpgradeRecoveryTargetType.recentContext;
    case 'order_confirm':
      return UpgradeRecoveryTargetType.orderConfirm;
    case 'order_pay':
      return UpgradeRecoveryTargetType.orderPay;
    default:
      return null;
  }
}

String upgradeRecoveryTargetTypeToValue(UpgradeRecoveryTargetType value) {
  switch (value) {
    case UpgradeRecoveryTargetType.recentContext:
      return 'recent_context';
    case UpgradeRecoveryTargetType.orderConfirm:
      return 'order_confirm';
    case UpgradeRecoveryTargetType.orderPay:
      return 'order_pay';
  }
}

class PendingUpgradeContext {
  final UpgradeRecoveryTargetType targetType;
  final String scene;
  final AppRecentContext? recoveryContext;
  final AppRecentContext? fallbackContext;
  final DirectCheckoutParams? directItem;
  final int? orderId;
  final String orderSn;
  final int? payType;
  final double? payAmount;
  final String blockingReason;
  final String currentVersion;
  final String requiredVersion;
  final String platform;
  final String channel;
  final String installerStore;
  final String updateMode;
  final bool blocking;
  final String upgradeUrl;
  final String storeTarget;
  final String recoveryHint;
  final String traceId;
  final DateTime enteredAt;
  final String actionTaken;
  final List<String> affectedCapabilities;

  const PendingUpgradeContext({
    required this.targetType,
    required this.scene,
    required this.recoveryContext,
    required this.fallbackContext,
    required this.directItem,
    required this.orderId,
    required this.orderSn,
    required this.payType,
    required this.payAmount,
    required this.blockingReason,
    required this.currentVersion,
    required this.requiredVersion,
    required this.platform,
    required this.channel,
    required this.installerStore,
    required this.updateMode,
    required this.blocking,
    required this.upgradeUrl,
    required this.storeTarget,
    required this.recoveryHint,
    required this.traceId,
    required this.enteredAt,
    required this.actionTaken,
    required this.affectedCapabilities,
  });

  factory PendingUpgradeContext.forRecentContext({
    required String scene,
    required AppRecentContext recoveryContext,
    required AppVersionPolicy policy,
  }) {
    return PendingUpgradeContext(
      targetType: UpgradeRecoveryTargetType.recentContext,
      scene: scene,
      recoveryContext: recoveryContext,
      fallbackContext: null,
      directItem: null,
      orderId: recoveryContext.targetId,
      orderSn: '',
      payType: null,
      payAmount: null,
      blockingReason: recoveryContext.failureReason,
      currentVersion: policy.currentVersion,
      requiredVersion: policy.requiredDisplayVersion,
      platform: policy.platform,
      channel: policy.channel,
      installerStore: policy.installerStore,
      updateMode: policy.updateMode.name,
      blocking: policy.blocking,
      upgradeUrl: policy.upgradeUrl,
      storeTarget: policy.storeTarget,
      recoveryHint: policy.recoveryHint,
      traceId: policy.traceId,
      enteredAt: DateTime.now(),
      actionTaken: 'gate_shown',
      affectedCapabilities: List<String>.from(policy.affectedCapabilities),
    );
  }

  factory PendingUpgradeContext.forOrderConfirm({
    required AppVersionPolicy policy,
    DirectCheckoutParams? directItem,
    AppRecentContext? fallbackContext,
  }) {
    return PendingUpgradeContext(
      targetType: UpgradeRecoveryTargetType.orderConfirm,
      scene: 'order_confirm',
      recoveryContext: null,
      fallbackContext: fallbackContext,
      directItem: directItem,
      orderId: null,
      orderSn: '',
      payType: null,
      payAmount: null,
      blockingReason: 'min_version_unmet',
      currentVersion: policy.currentVersion,
      requiredVersion: policy.requiredDisplayVersion,
      platform: policy.platform,
      channel: policy.channel,
      installerStore: policy.installerStore,
      updateMode: policy.updateMode.name,
      blocking: policy.blocking,
      upgradeUrl: policy.upgradeUrl,
      storeTarget: policy.storeTarget,
      recoveryHint: policy.recoveryHint,
      traceId: policy.traceId,
      enteredAt: DateTime.now(),
      actionTaken: 'gate_shown',
      affectedCapabilities: List<String>.from(policy.affectedCapabilities),
    );
  }

  factory PendingUpgradeContext.forOrderPay({
    required AppVersionPolicy policy,
    required int orderId,
    required double payAmount,
    int? payType,
    String orderSn = '',
    AppRecentContext? fallbackContext,
  }) {
    return PendingUpgradeContext(
      targetType: UpgradeRecoveryTargetType.orderPay,
      scene: 'order_pay',
      recoveryContext: null,
      fallbackContext: fallbackContext,
      directItem: null,
      orderId: orderId,
      orderSn: orderSn,
      payType: payType,
      payAmount: payAmount,
      blockingReason: 'min_version_unmet',
      currentVersion: policy.currentVersion,
      requiredVersion: policy.requiredDisplayVersion,
      platform: policy.platform,
      channel: policy.channel,
      installerStore: policy.installerStore,
      updateMode: policy.updateMode.name,
      blocking: policy.blocking,
      upgradeUrl: policy.upgradeUrl,
      storeTarget: policy.storeTarget,
      recoveryHint: policy.recoveryHint,
      traceId: policy.traceId,
      enteredAt: DateTime.now(),
      actionTaken: 'gate_shown',
      affectedCapabilities: List<String>.from(policy.affectedCapabilities),
    );
  }

  PendingUpgradeContext copyWith({
    UpgradeRecoveryTargetType? targetType,
    String? scene,
    AppRecentContext? recoveryContext,
    AppRecentContext? fallbackContext,
    DirectCheckoutParams? directItem,
    int? orderId,
    String? orderSn,
    int? payType,
    double? payAmount,
    String? blockingReason,
    String? currentVersion,
    String? requiredVersion,
    String? platform,
    String? channel,
    String? installerStore,
    String? updateMode,
    bool? blocking,
    String? upgradeUrl,
    String? storeTarget,
    String? recoveryHint,
    String? traceId,
    DateTime? enteredAt,
    String? actionTaken,
    List<String>? affectedCapabilities,
  }) {
    return PendingUpgradeContext(
      targetType: targetType ?? this.targetType,
      scene: scene ?? this.scene,
      recoveryContext: recoveryContext ?? this.recoveryContext,
      fallbackContext: fallbackContext ?? this.fallbackContext,
      directItem: directItem ?? this.directItem,
      orderId: orderId ?? this.orderId,
      orderSn: orderSn ?? this.orderSn,
      payType: payType ?? this.payType,
      payAmount: payAmount ?? this.payAmount,
      blockingReason: blockingReason ?? this.blockingReason,
      currentVersion: currentVersion ?? this.currentVersion,
      requiredVersion: requiredVersion ?? this.requiredVersion,
      platform: platform ?? this.platform,
      channel: channel ?? this.channel,
      installerStore: installerStore ?? this.installerStore,
      updateMode: updateMode ?? this.updateMode,
      blocking: blocking ?? this.blocking,
      upgradeUrl: upgradeUrl ?? this.upgradeUrl,
      storeTarget: storeTarget ?? this.storeTarget,
      recoveryHint: recoveryHint ?? this.recoveryHint,
      traceId: traceId ?? this.traceId,
      enteredAt: enteredAt ?? this.enteredAt,
      actionTaken: actionTaken ?? this.actionTaken,
      affectedCapabilities:
          affectedCapabilities ?? List<String>.from(this.affectedCapabilities),
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'targetType': upgradeRecoveryTargetTypeToValue(targetType),
      'scene': scene,
      'recoveryContext': recoveryContext?.toJson(),
      'fallbackContext': fallbackContext?.toJson(),
      'directItem': directItem?.toJson(),
      'orderId': orderId,
      'orderSn': orderSn,
      'payType': payType,
      'payAmount': payAmount,
      'blockingReason': blockingReason,
      'currentVersion': currentVersion,
      'requiredVersion': requiredVersion,
      'platform': platform,
      'channel': channel,
      'installerStore': installerStore,
      'updateMode': updateMode,
      'blocking': blocking,
      'upgradeUrl': upgradeUrl,
      'storeTarget': storeTarget,
      'recoveryHint': recoveryHint,
      'traceId': traceId,
      'enteredAt': enteredAt.toIso8601String(),
      'actionTaken': actionTaken,
      'affectedCapabilities': affectedCapabilities,
    };
  }

  factory PendingUpgradeContext.fromJson(Map<String, dynamic> json) {
    final targetType =
        upgradeRecoveryTargetTypeFromValue(json['targetType']?.toString()) ??
            UpgradeRecoveryTargetType.recentContext;
    final directItemJson = json['directItem'];
    return PendingUpgradeContext(
      targetType: targetType,
      scene: json['scene']?.toString() ?? '',
      recoveryContext: AppRecentContext.tryParse(json['recoveryContext']),
      fallbackContext: AppRecentContext.tryParse(json['fallbackContext']),
      directItem: directItemJson is Map<String, dynamic>
          ? DirectCheckoutParams.fromJson(directItemJson)
          : null,
      orderId: _parseInt(json['orderId']),
      orderSn: json['orderSn']?.toString() ?? '',
      payType: _parseInt(json['payType']),
      payAmount: _parseDouble(json['payAmount']),
      blockingReason: json['blockingReason']?.toString() ?? '',
      currentVersion: json['currentVersion']?.toString() ?? '',
      requiredVersion: json['requiredVersion']?.toString() ?? '',
      platform: json['platform']?.toString() ?? '',
      channel: json['channel']?.toString() ?? '',
      installerStore: json['installerStore']?.toString() ?? '',
      updateMode: json['updateMode']?.toString() ?? '',
      blocking: json['blocking'] == true,
      upgradeUrl: json['upgradeUrl']?.toString() ?? '',
      storeTarget: json['storeTarget']?.toString() ?? '',
      recoveryHint: json['recoveryHint']?.toString() ?? '',
      traceId: json['traceId']?.toString() ?? '',
      enteredAt: DateTime.tryParse(json['enteredAt']?.toString() ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
      actionTaken: json['actionTaken']?.toString() ?? '',
      affectedCapabilities: (json['affectedCapabilities'] as List<dynamic>?)
              ?.map((item) => item.toString())
              .toList() ??
          const <String>[],
    );
  }

  static PendingUpgradeContext? tryParse(dynamic value) {
    if (value is! Map) {
      return null;
    }
    try {
      return PendingUpgradeContext.fromJson(Map<String, dynamic>.from(value));
    } catch (_) {
      return null;
    }
  }

  static int? _parseInt(dynamic value) {
    if (value is int) {
      return value;
    }
    return int.tryParse(value?.toString() ?? '');
  }

  static double? _parseDouble(dynamic value) {
    if (value is double) {
      return value;
    }
    if (value is int) {
      return value.toDouble();
    }
    return double.tryParse(value?.toString() ?? '');
  }
}
