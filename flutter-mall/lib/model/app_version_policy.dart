import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:flutter_mall/utils/version_compare.dart';

enum UpgradeUpdateMode {
  none,
  recommended,
  deadline,
  force,
}

UpgradeUpdateMode upgradeUpdateModeFromValue(String? value) {
  switch ((value ?? '').trim().toLowerCase()) {
    case 'recommended':
      return UpgradeUpdateMode.recommended;
    case 'deadline':
      return UpgradeUpdateMode.deadline;
    case 'force':
      return UpgradeUpdateMode.force;
    default:
      return UpgradeUpdateMode.none;
  }
}

class AppVersionPolicy {
  final String currentVersion;
  final String minSupportedVersion;
  final String recommendedVersion;
  final String requiredVersion;
  final UpgradeUpdateMode updateMode;
  final String effectiveAt;
  final String deadlineAt;
  final List<String> affectedCapabilities;
  final bool blocking;
  final String releaseNotesSummary;
  final String upgradeUrl;
  final String storeTarget;
  final String recoveryHint;
  final String traceId;
  final String platform;
  final String channel;
  final String installerStore;
  final String scene;
  final String targetType;
  final int? targetId;

  const AppVersionPolicy({
    required this.currentVersion,
    required this.minSupportedVersion,
    required this.recommendedVersion,
    required this.requiredVersion,
    required this.updateMode,
    required this.effectiveAt,
    required this.deadlineAt,
    required this.affectedCapabilities,
    required this.blocking,
    required this.releaseNotesSummary,
    required this.upgradeUrl,
    required this.storeTarget,
    required this.recoveryHint,
    required this.traceId,
    required this.platform,
    required this.channel,
    required this.installerStore,
    required this.scene,
    required this.targetType,
    required this.targetId,
  });

  factory AppVersionPolicy.none({
    required AppVersionInfo versionInfo,
    required String scene,
    String targetType = '',
    int? targetId,
  }) {
    return AppVersionPolicy(
      currentVersion: versionInfo.version,
      minSupportedVersion: versionInfo.version,
      recommendedVersion: versionInfo.version,
      requiredVersion: '',
      updateMode: UpgradeUpdateMode.none,
      effectiveAt: '',
      deadlineAt: '',
      affectedCapabilities: const <String>[],
      blocking: false,
      releaseNotesSummary: '当前已是最新版本，可继续使用当前能力。',
      upgradeUrl: '',
      storeTarget: '',
      recoveryHint: '',
      traceId: '',
      platform: versionInfo.platform,
      channel: versionInfo.channel,
      installerStore: versionInfo.installerStore,
      scene: scene,
      targetType: targetType,
      targetId: targetId,
    );
  }

  factory AppVersionPolicy.fromResponse(
    Map<String, dynamic> json, {
    required AppVersionInfo versionInfo,
    required String scene,
    String targetType = '',
    int? targetId,
  }) {
    final payload = json['data'] is Map<String, dynamic>
        ? json['data'] as Map<String, dynamic>
        : json;
    return AppVersionPolicy(
      currentVersion:
          (payload['currentVersion']?.toString() ?? '').trim().isEmpty
              ? versionInfo.version
              : payload['currentVersion'].toString().trim(),
      minSupportedVersion:
          payload['minSupportedVersion']?.toString().trim() ?? '',
      recommendedVersion:
          payload['recommendedVersion']?.toString().trim() ?? '',
      requiredVersion: payload['requiredVersion']?.toString().trim() ?? '',
      updateMode: upgradeUpdateModeFromValue(payload['updateMode']),
      effectiveAt: payload['effectiveAt']?.toString().trim() ?? '',
      deadlineAt: payload['deadlineAt']?.toString().trim() ?? '',
      affectedCapabilities: (payload['affectedCapabilities'] as List<dynamic>?)
              ?.map((item) => item.toString())
              .where((item) => item.trim().isNotEmpty)
              .toList() ??
          const <String>[],
      blocking: payload['blocking'] == true,
      releaseNotesSummary:
          payload['releaseNotesSummary']?.toString().trim() ?? '',
      upgradeUrl: payload['upgradeUrl']?.toString().trim() ?? '',
      storeTarget: payload['storeTarget']?.toString().trim() ?? '',
      recoveryHint: payload['recoveryHint']?.toString().trim() ?? '',
      traceId: payload['traceId']?.toString().trim() ?? '',
      platform: payload['platform']?.toString().trim().isNotEmpty == true
          ? payload['platform'].toString().trim()
          : versionInfo.platform,
      channel: payload['channel']?.toString().trim().isNotEmpty == true
          ? payload['channel'].toString().trim()
          : versionInfo.channel,
      installerStore:
          payload['installerStore']?.toString().trim().isNotEmpty == true
              ? payload['installerStore'].toString().trim()
              : versionInfo.installerStore,
      scene: payload['scene']?.toString().trim().isNotEmpty == true
          ? payload['scene'].toString().trim()
          : scene,
      targetType: payload['targetType']?.toString().trim().isNotEmpty == true
          ? payload['targetType'].toString().trim()
          : targetType,
      targetId: _parseInt(payload['targetId']) ?? targetId,
    );
  }

  factory AppVersionPolicy.fromRecallIntent(
    AppRecentContext intent, {
    required AppVersionInfo versionInfo,
  }) {
    return AppVersionPolicy.none(
      versionInfo: versionInfo,
      scene: 'message_recall',
      targetType: intent.targetTypeValue,
      targetId: intent.targetId,
    ).applyRecallRequirement(intent);
  }

  AppVersionPolicy copyWith({
    String? currentVersion,
    String? minSupportedVersion,
    String? recommendedVersion,
    String? requiredVersion,
    UpgradeUpdateMode? updateMode,
    String? effectiveAt,
    String? deadlineAt,
    List<String>? affectedCapabilities,
    bool? blocking,
    String? releaseNotesSummary,
    String? upgradeUrl,
    String? storeTarget,
    String? recoveryHint,
    String? traceId,
    String? platform,
    String? channel,
    String? installerStore,
    String? scene,
    String? targetType,
    int? targetId,
  }) {
    return AppVersionPolicy(
      currentVersion: currentVersion ?? this.currentVersion,
      minSupportedVersion: minSupportedVersion ?? this.minSupportedVersion,
      recommendedVersion: recommendedVersion ?? this.recommendedVersion,
      requiredVersion: requiredVersion ?? this.requiredVersion,
      updateMode: updateMode ?? this.updateMode,
      effectiveAt: effectiveAt ?? this.effectiveAt,
      deadlineAt: deadlineAt ?? this.deadlineAt,
      affectedCapabilities:
          affectedCapabilities ?? List<String>.from(this.affectedCapabilities),
      blocking: blocking ?? this.blocking,
      releaseNotesSummary: releaseNotesSummary ?? this.releaseNotesSummary,
      upgradeUrl: upgradeUrl ?? this.upgradeUrl,
      storeTarget: storeTarget ?? this.storeTarget,
      recoveryHint: recoveryHint ?? this.recoveryHint,
      traceId: traceId ?? this.traceId,
      platform: platform ?? this.platform,
      channel: channel ?? this.channel,
      installerStore: installerStore ?? this.installerStore,
      scene: scene ?? this.scene,
      targetType: targetType ?? this.targetType,
      targetId: targetId ?? this.targetId,
    );
  }

  AppVersionPolicy applyRecallRequirement(AppRecentContext intent) {
    final localRequiredVersion = intent.minAppVersion.trim();
    final mergedRecoveryHint = intent.recoveryHint.trim().isNotEmpty
        ? intent.recoveryHint.trim()
        : recoveryHint;
    final mergedCapabilities = _mergeCapabilities(<String>[
      'message_recall',
      intent.targetTypeValue,
      ...affectedCapabilities,
    ]);

    final basePolicy = copyWith(
      recoveryHint: mergedRecoveryHint,
      affectedCapabilities: mergedCapabilities,
      scene: scene.trim().isNotEmpty ? scene : 'message_recall',
      targetType:
          targetType.trim().isNotEmpty ? targetType : intent.targetTypeValue,
      targetId: targetId ?? intent.targetId,
      traceId: traceId.trim().isNotEmpty ? traceId : intent.intentId.trim(),
    );

    if (localRequiredVersion.isEmpty ||
        compareVersion(currentVersion, localRequiredVersion) >= 0) {
      return basePolicy;
    }

    final remoteRequiredVersion = basePolicy.requiredDisplayVersion.trim();
    if (remoteRequiredVersion.isNotEmpty &&
        compareVersion(remoteRequiredVersion, localRequiredVersion) >= 0) {
      return basePolicy;
    }

    final hasLaunchableUpgradePath = basePolicy.upgradeUrl.trim().isNotEmpty;
    return basePolicy.copyWith(
      minSupportedVersion: _pickHigherVersion(
          basePolicy.minSupportedVersion, localRequiredVersion),
      recommendedVersion: _pickHigherVersion(
          basePolicy.recommendedVersion, localRequiredVersion),
      requiredVersion: localRequiredVersion,
      updateMode: hasLaunchableUpgradePath
          ? UpgradeUpdateMode.force
          : UpgradeUpdateMode.recommended,
      blocking: hasLaunchableUpgradePath,
      storeTarget:
          hasLaunchableUpgradePath ? basePolicy.storeTarget : 'settings_check',
      releaseNotesSummary: basePolicy.hasUpgradeGate &&
              basePolicy.releaseNotesSummary.trim().isNotEmpty
          ? basePolicy.releaseNotesSummary
          : '升级后可继续当前消息唤回目标，并保持恢复链路不丢失。',
      recoveryHint: mergedRecoveryHint.isNotEmpty
          ? mergedRecoveryHint
          : hasLaunchableUpgradePath
              ? '升级后可继续当前消息唤回目标，并保持恢复链路不丢失。'
              : '当前版本暂不支持该消息目标，请稍后在设置页检查更新后重试。',
    );
  }

  bool get hasLaunchableUpgradePath => upgradeUrl.trim().isNotEmpty;

  static List<String> _mergeCapabilities(List<String> values) {
    final merged = <String>[];
    for (final value in values) {
      final normalized = value.trim();
      if (normalized.isEmpty || merged.contains(normalized)) {
        continue;
      }
      merged.add(normalized);
    }
    return merged;
  }

  static String _pickHigherVersion(String current, String candidate) {
    final currentTrimmed = current.trim();
    final candidateTrimmed = candidate.trim();
    if (currentTrimmed.isEmpty) {
      return candidateTrimmed;
    }
    if (candidateTrimmed.isEmpty) {
      return currentTrimmed;
    }
    return compareVersion(currentTrimmed, candidateTrimmed) >= 0
        ? currentTrimmed
        : candidateTrimmed;
  }

  bool get hasUpgradeGate => updateMode != UpgradeUpdateMode.none;

  bool get canContinueLater => !blocking;

  bool get isUpToDate => updateMode == UpgradeUpdateMode.none;

  String get requiredDisplayVersion {
    if (requiredVersion.trim().isNotEmpty) {
      return requiredVersion.trim();
    }
    if (blocking && minSupportedVersion.trim().isNotEmpty) {
      return minSupportedVersion.trim();
    }
    return recommendedVersion.trim();
  }

  static int? _parseInt(dynamic value) {
    if (value is int) {
      return value;
    }
    return int.tryParse(value?.toString() ?? '');
  }
}
