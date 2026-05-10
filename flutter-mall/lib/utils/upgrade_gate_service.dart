import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:url_launcher/url_launcher.dart';

class UpgradePolicyQuery {
  final String scene;
  final String targetType;
  final int? targetId;
  final String channel;
  final String installerStore;
  final AppVersionInfo versionInfo;

  const UpgradePolicyQuery({
    required this.scene,
    required this.targetType,
    required this.targetId,
    required this.channel,
    required this.installerStore,
    required this.versionInfo,
  });
}

typedef UpgradePolicyFetcher = Future<AppVersionPolicy> Function(
  UpgradePolicyQuery query,
);

typedef UpgradeUrlLauncher = Future<bool> Function(Uri uri);

class UpgradeGateService {
  static UpgradePolicyFetcher? _debugPolicyFetcher;
  static UpgradeUrlLauncher? _debugUrlLauncher;

  static String recoverySceneFor(AppRecentContext context) {
    if (context.isRecallIntent) {
      return 'message_recall';
    }
    switch (context.targetType) {
      case AppRecentTargetType.orderDetail:
      case AppRecentTargetType.afterSalesApply:
      case AppRecentTargetType.commentCompose:
        return 'order_recovery';
      case AppRecentTargetType.home:
      case AppRecentTargetType.cart:
      case AppRecentTargetType.orderList:
      case AppRecentTargetType.productDetail:
      case AppRecentTargetType.settings:
      case AppRecentTargetType.couponList:
      case AppRecentTargetType.couponCenter:
      case AppRecentTargetType.activity:
      case AppRecentTargetType.digitalCardAssetList:
      case AppRecentTargetType.digitalCardAssetDetail:
      case AppRecentTargetType.digitalCardClaim:
      case AppRecentTargetType.subject:
      case AppRecentTargetType.preferredArea:
        return 'app_bootstrap';
    }
  }

  static Future<AppVersionPolicy> queryPolicy({
    required String scene,
    AppRecentContext? recoveryContext,
    String targetType = '',
    int? targetId,
  }) async {
    final versionInfo = await AppVersionService.getCurrentInfo();
    final resolvedTargetType = recoveryContext?.targetTypeValue ?? targetType;
    final resolvedTargetId = recoveryContext?.targetId ?? targetId;
    final query = UpgradePolicyQuery(
      scene: scene,
      targetType: resolvedTargetType,
      targetId: resolvedTargetId,
      channel: versionInfo.channel,
      installerStore: versionInfo.installerStore,
      versionInfo: versionInfo,
    );

    if (_debugPolicyFetcher != null) {
      return _debugPolicyFetcher!(query);
    }

    final response = await HttpUtil.get(
      appVersionPolicyUrl,
      queryParameters: <String, dynamic>{
        'scene': query.scene,
        'targetType': query.targetType,
        if (query.targetId != null) 'targetId': query.targetId,
        'channel': query.channel,
        'installerStore': query.installerStore,
      },
    );

    return AppVersionPolicy.fromResponse(
      Map<String, dynamic>.from(response.data as Map),
      versionInfo: versionInfo,
      scene: scene,
      targetType: resolvedTargetType,
      targetId: resolvedTargetId,
    );
  }

  static Future<AppVersionPolicy> queryPendingPolicy(
    PendingUpgradeContext pending,
  ) {
    final recoveryContext = pending.recoveryContext;
    if (recoveryContext != null && recoveryContext.isRecallIntent) {
      return queryRecallPolicy(
        recoveryContext,
        scene: recoverySceneFor(recoveryContext),
      );
    }
    return queryPolicy(
      scene: pending.scene,
      recoveryContext: recoveryContext,
      targetType: recoveryContext?.targetTypeValue ?? '',
      targetId: recoveryContext?.targetId ?? pending.orderId,
    );
  }

  static Future<AppVersionPolicy> queryRecallPolicy(
    AppRecentContext intent, {
    String? scene,
  }) async {
    final sharedPolicy = await queryPolicy(
      scene: scene ?? recoverySceneFor(intent),
      recoveryContext: intent,
      targetType: intent.targetTypeValue,
      targetId: intent.targetId,
    );
    return sharedPolicy.applyRecallRequirement(intent);
  }

  static AppVersionPolicy policyFromRecallIntent(AppRecentContext intent) {
    return AppVersionPolicy.fromRecallIntent(
      intent,
      versionInfo: AppVersionService.currentInfo,
    );
  }

  static Future<bool> launchUpgrade(AppVersionPolicy policy) async {
    final url = policy.upgradeUrl.trim();
    if (url.isEmpty) {
      return false;
    }
    final uri = Uri.tryParse(url);
    if (uri == null) {
      return false;
    }
    if (_debugUrlLauncher != null) {
      return _debugUrlLauncher!(uri);
    }
    return launchUrl(uri, mode: LaunchMode.externalApplication);
  }

  static void debugSetPolicyFetcher(UpgradePolicyFetcher? fetcher) {
    _debugPolicyFetcher = fetcher;
  }

  static void debugSetUrlLauncher(UpgradeUrlLauncher? launcher) {
    _debugUrlLauncher = launcher;
  }

  static void debugReset() {
    _debugPolicyFetcher = null;
    _debugUrlLauncher = null;
  }
}
