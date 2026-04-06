import 'dart:convert';
import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:image_picker/image_picker.dart';
import 'package:permission_handler/permission_handler.dart' as permission;

enum BrokerSystemPermissionStatus {
  granted,
  limited,
  denied,
  permanentlyDenied,
  restricted,
}

class PermissionInspectionResult {
  final PermissionFlowContext flow;
  final PermissionFlowStatus status;
  final bool statusCheckOnly;

  const PermissionInspectionResult({
    required this.flow,
    required this.status,
    this.statusCheckOnly = false,
  });

  bool get isUsable =>
      status == PermissionFlowStatus.granted ||
      status == PermissionFlowStatus.limited ||
      status == PermissionFlowStatus.notRequired;

  bool get shouldOpenSettings =>
      status == PermissionFlowStatus.permanentlyDenied;
}

class PermissionMediaResult {
  final PermissionFlowContext flow;
  final PermissionFlowStatus status;
  final List<String> mediaBase64List;
  final String errorMessage;

  const PermissionMediaResult({
    required this.flow,
    required this.status,
    this.mediaBase64List = const <String>[],
    this.errorMessage = '',
  });

  bool get hasMedia => mediaBase64List.isNotEmpty;
}

enum NotificationPreferencePresentation {
  enabled,
  mutedLocally,
  blockedBySystem,
}

class NotificationPreferenceSnapshot {
  final NotificationPreferencePresentation presentation;
  final PermissionFlowStatus systemStatus;
  final bool localPreferenceEnabled;
  final bool statusCheckOnly;

  const NotificationPreferenceSnapshot({
    required this.presentation,
    required this.systemStatus,
    required this.localPreferenceEnabled,
    required this.statusCheckOnly,
  });

  bool get isEnabled =>
      presentation == NotificationPreferencePresentation.enabled;
}

abstract class PermissionBrokerAdapter {
  bool get isIOS;

  bool get isAndroid;

  int? get androidSdkInt;

  Future<BrokerSystemPermissionStatus> getNotificationStatus();

  Future<BrokerSystemPermissionStatus> requestNotificationPermission();

  Future<BrokerSystemPermissionStatus> getPhotoStatus();

  Future<BrokerSystemPermissionStatus> requestPhotoPermission();

  Future<BrokerSystemPermissionStatus> getCameraStatus();

  Future<BrokerSystemPermissionStatus> requestCameraPermission();

  Future<bool> openAppSettings();

  Future<List<String>> retrieveLostMediaBase64();
}

class PermissionBroker {
  final PermissionBrokerAdapter adapter;

  const PermissionBroker({
    this.adapter = const DefaultPermissionBrokerAdapter(),
  });

  Future<PermissionInspectionResult> inspect(PermissionFlowContext flow) async {
    if (_isPhotoPermissionNotRequired(flow)) {
      return PermissionInspectionResult(
        flow: flow,
        status: PermissionFlowStatus.notRequired,
      );
    }

    if (_isNotificationStatusCheckOnly(flow)) {
      final systemStatus = await adapter.getNotificationStatus();
      return PermissionInspectionResult(
        flow: flow,
        status: systemStatus == BrokerSystemPermissionStatus.granted
            ? PermissionFlowStatus.notRequired
            : _mapStatus(systemStatus),
        statusCheckOnly: true,
      );
    }

    final systemStatus = await _resolveCurrentStatus(flow);
    return PermissionInspectionResult(
      flow: flow,
      status: _mapStatus(systemStatus),
    );
  }

  Future<PermissionInspectionResult> request(PermissionFlowContext flow) async {
    if (_isPhotoPermissionNotRequired(flow)) {
      return PermissionInspectionResult(
        flow: flow,
        status: PermissionFlowStatus.notRequired,
      );
    }

    if (_isNotificationStatusCheckOnly(flow)) {
      return inspect(flow);
    }

    final systemStatus = await _requestStatus(flow);
    return PermissionInspectionResult(
      flow: flow,
      status: _mapStatus(systemStatus),
    );
  }

  Future<bool> openSettingsForFlow(PermissionFlowContext flow) async {
    await AppRecoveryStore.savePendingPermissionContext(
      flow.copyWith(status: PermissionFlowStatus.settingsReturnPending),
    );
    final opened = await adapter.openAppSettings();
    if (!opened) {
      await AppRecoveryStore.clearPendingPermissionContext();
    }
    return opened;
  }

  Future<PermissionMediaResult> pickImageAsBase64(
    PermissionFlowContext flow,
  ) async {
    if (flow.source != PermissionFlowSource.gallery &&
        flow.source != PermissionFlowSource.camera) {
      return PermissionMediaResult(
        flow: flow,
        status: flow.status,
        errorMessage: 'unsupported_media_source',
      );
    }

    await AppRecoveryStore.savePendingPermissionContext(flow);
    try {
      final source = flow.source == PermissionFlowSource.camera
          ? ImageSource.camera
          : ImageSource.gallery;
      final picker = ImagePicker();
      final image = await picker.pickImage(
        source: source,
        maxWidth: 1024,
        maxHeight: 1024,
        imageQuality: 80,
      );
      if (image == null) {
        await AppRecoveryStore.clearPendingPermissionContext();
        return PermissionMediaResult(
          flow: flow,
          status: flow.status,
        );
      }
      final bytes = await image.readAsBytes();
      await AppRecoveryStore.clearPendingPermissionContext();
      await AppRecoveryStore.clearPendingLostMedia();
      return PermissionMediaResult(
        flow: flow,
        status: flow.status,
        mediaBase64List: <String>[base64Encode(bytes)],
      );
    } catch (_) {
      await AppRecoveryStore.clearPendingPermissionContext();
      return PermissionMediaResult(
        flow: flow,
        status: PermissionFlowStatus.denied,
        errorMessage: 'image_picker_failed',
      );
    }
  }

  Future<NotificationPreferenceSnapshot> getNotificationPreferenceSnapshot({
    PermissionFlowContext? flow,
  }) async {
    final localEnabled = AppRecoveryStore.getNotificationPreferenceEnabled();
    final inspection = await inspect(
      flow ??
          PermissionFlowContext.create(
            scene: PermissionScene.notificationSubscription,
            permissionType: PermissionType.notification,
            source: PermissionFlowSource.notificationToggle,
            returnTarget: AppRecentTargetType.settings,
            fallbackAction: PermissionFallbackAction.openMessageCenter,
            intentId: 'notification_settings',
            recoveryId: 'notification_settings',
          ),
    );

    if (inspection.isUsable && localEnabled) {
      return NotificationPreferenceSnapshot(
        presentation: NotificationPreferencePresentation.enabled,
        systemStatus: inspection.status,
        localPreferenceEnabled: localEnabled,
        statusCheckOnly: inspection.statusCheckOnly,
      );
    }

    if (inspection.isUsable) {
      return NotificationPreferenceSnapshot(
        presentation: NotificationPreferencePresentation.mutedLocally,
        systemStatus: inspection.status,
        localPreferenceEnabled: localEnabled,
        statusCheckOnly: inspection.statusCheckOnly,
      );
    }

    return NotificationPreferenceSnapshot(
      presentation: NotificationPreferencePresentation.blockedBySystem,
      systemStatus: inspection.status,
      localPreferenceEnabled: localEnabled,
      statusCheckOnly: inspection.statusCheckOnly,
    );
  }

  Future<void> persistNotificationPreference(bool enabled) async {
    await AppRecoveryStore.saveNotificationPreferenceEnabled(enabled);
  }

  Future<void> captureLostMediaOnLaunch() async {
    final pendingFlow = AppRecoveryStore.peekPendingPermissionContext();
    if (pendingFlow == null) {
      return;
    }
    final mediaBase64List = await adapter.retrieveLostMediaBase64();
    if (mediaBase64List.isEmpty) {
      return;
    }
    await AppRecoveryStore.savePendingLostMedia(
      PermissionLostMediaSnapshot(
        recoveryId: pendingFlow.recoveryId,
        scene: pendingFlow.scene,
        source: pendingFlow.source,
        mediaBase64List: mediaBase64List,
        errorMessage: '',
        capturedAt: DateTime.now(),
      ),
    );
  }

  PermissionFlowStatus mapSystemStatusForTests(
    BrokerSystemPermissionStatus status,
  ) {
    return _mapStatus(status);
  }

  bool _isPhotoPermissionNotRequired(PermissionFlowContext flow) {
    return flow.permissionType == PermissionType.photos && adapter.isAndroid;
  }

  bool _isNotificationStatusCheckOnly(PermissionFlowContext flow) {
    if (flow.permissionType != PermissionType.notification) {
      return false;
    }
    if (!adapter.isAndroid) {
      return false;
    }
    final sdkInt = adapter.androidSdkInt;
    return sdkInt != null && sdkInt < 33;
  }

  Future<BrokerSystemPermissionStatus> _resolveCurrentStatus(
    PermissionFlowContext flow,
  ) {
    switch (flow.permissionType) {
      case PermissionType.notification:
        return adapter.getNotificationStatus();
      case PermissionType.photos:
        return adapter.getPhotoStatus();
      case PermissionType.camera:
        return adapter.getCameraStatus();
    }
  }

  Future<BrokerSystemPermissionStatus> _requestStatus(
    PermissionFlowContext flow,
  ) {
    switch (flow.permissionType) {
      case PermissionType.notification:
        return adapter.requestNotificationPermission();
      case PermissionType.photos:
        return adapter.requestPhotoPermission();
      case PermissionType.camera:
        return adapter.requestCameraPermission();
    }
  }

  PermissionFlowStatus _mapStatus(BrokerSystemPermissionStatus status) {
    switch (status) {
      case BrokerSystemPermissionStatus.granted:
        return PermissionFlowStatus.granted;
      case BrokerSystemPermissionStatus.limited:
        return PermissionFlowStatus.limited;
      case BrokerSystemPermissionStatus.denied:
      case BrokerSystemPermissionStatus.restricted:
        return PermissionFlowStatus.denied;
      case BrokerSystemPermissionStatus.permanentlyDenied:
        return PermissionFlowStatus.permanentlyDenied;
    }
  }
}

class DefaultPermissionBrokerAdapter implements PermissionBrokerAdapter {
  const DefaultPermissionBrokerAdapter();

  @override
  bool get isIOS => !kIsWeb && Platform.isIOS;

  @override
  bool get isAndroid => !kIsWeb && Platform.isAndroid;

  @override
  int? get androidSdkInt {
    if (!isAndroid) {
      return null;
    }
    final match = RegExp(r'SDK\s*(\d+)').firstMatch(
      Platform.operatingSystemVersion,
    );
    return int.tryParse(match?.group(1) ?? '');
  }

  @override
  Future<BrokerSystemPermissionStatus> getNotificationStatus() async {
    return _mapPermissionStatus(
        await permission.Permission.notification.status);
  }

  @override
  Future<BrokerSystemPermissionStatus> requestNotificationPermission() async {
    return _mapPermissionStatus(
      await permission.Permission.notification.request(),
    );
  }

  @override
  Future<BrokerSystemPermissionStatus> getPhotoStatus() async {
    return _mapPermissionStatus(await permission.Permission.photos.status);
  }

  @override
  Future<BrokerSystemPermissionStatus> requestPhotoPermission() async {
    return _mapPermissionStatus(await permission.Permission.photos.request());
  }

  @override
  Future<BrokerSystemPermissionStatus> getCameraStatus() async {
    return _mapPermissionStatus(await permission.Permission.camera.status);
  }

  @override
  Future<BrokerSystemPermissionStatus> requestCameraPermission() async {
    return _mapPermissionStatus(await permission.Permission.camera.request());
  }

  @override
  Future<bool> openAppSettings() {
    return permission.openAppSettings();
  }

  @override
  Future<List<String>> retrieveLostMediaBase64() async {
    final picker = ImagePicker();
    final lostData = await picker.retrieveLostData();
    if (lostData.isEmpty) {
      return const <String>[];
    }
    final files = lostData.files ?? <XFile>[];
    final results = <String>[];
    for (final file in files) {
      final bytes = await file.readAsBytes();
      results.add(base64Encode(bytes));
    }
    return results;
  }

  BrokerSystemPermissionStatus _mapPermissionStatus(
    permission.PermissionStatus status,
  ) {
    if (status.isGranted) {
      return BrokerSystemPermissionStatus.granted;
    }
    if (status.isLimited) {
      return BrokerSystemPermissionStatus.limited;
    }
    if (status.isPermanentlyDenied) {
      return BrokerSystemPermissionStatus.permanentlyDenied;
    }
    if (status.isRestricted) {
      return BrokerSystemPermissionStatus.restricted;
    }
    return BrokerSystemPermissionStatus.denied;
  }
}
