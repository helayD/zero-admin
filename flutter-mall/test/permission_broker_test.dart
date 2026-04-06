import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  test('android gallery inspect maps to not required', () async {
    final broker = PermissionBroker(
      adapter: _FakeAdapter(
        isAndroid: true,
        androidSdkInt: 34,
        photoStatus: BrokerSystemPermissionStatus.denied,
      ),
    );

    final result = await broker.inspect(
      PermissionFlowContext.create(
        scene: PermissionScene.commentImage,
        permissionType: PermissionType.photos,
        source: PermissionFlowSource.gallery,
        returnTarget: AppRecentTargetType.commentCompose,
        returnTargetId: 1001,
        fallbackAction: PermissionFallbackAction.switchToCamera,
        intentId: 'comment_gallery_1001',
        recoveryId: 'comment_gallery_1001',
      ),
    );

    expect(result.status, PermissionFlowStatus.notRequired);
    expect(result.isUsable, isTrue);
  });

  test('ios limited photo access is preserved as limited', () async {
    final broker = PermissionBroker(
      adapter: _FakeAdapter(
        isIOS: true,
        photoStatus: BrokerSystemPermissionStatus.limited,
      ),
    );

    final result = await broker.inspect(
      PermissionFlowContext.create(
        scene: PermissionScene.commentImage,
        permissionType: PermissionType.photos,
        source: PermissionFlowSource.gallery,
        returnTarget: AppRecentTargetType.commentCompose,
        returnTargetId: 1001,
        fallbackAction: PermissionFallbackAction.switchToCamera,
        intentId: 'comment_gallery_1001',
        recoveryId: 'comment_gallery_1001',
      ),
    );

    expect(result.status, PermissionFlowStatus.limited);
  });

  test('android below 13 notification uses status check only state', () async {
    final broker = PermissionBroker(
      adapter: _FakeAdapter(
        isAndroid: true,
        androidSdkInt: 32,
        notificationStatus: BrokerSystemPermissionStatus.granted,
      ),
    );

    final flow = PermissionFlowContext.create(
      scene: PermissionScene.notificationSubscription,
      permissionType: PermissionType.notification,
      source: PermissionFlowSource.notificationToggle,
      returnTarget: AppRecentTargetType.settings,
      fallbackAction: PermissionFallbackAction.openMessageCenter,
      intentId: 'settings_notification_toggle',
      recoveryId: 'settings_notification_toggle',
    );

    final inspection = await broker.inspect(flow);
    expect(inspection.status, PermissionFlowStatus.notRequired);
    expect(inspection.statusCheckOnly, isTrue);

    var snapshot = await broker.getNotificationPreferenceSnapshot(flow: flow);
    expect(
        snapshot.presentation, NotificationPreferencePresentation.mutedLocally);

    await broker.persistNotificationPreference(true);
    snapshot = await broker.getNotificationPreferenceSnapshot(flow: flow);
    expect(snapshot.presentation, NotificationPreferencePresentation.enabled);
  });

  test('openSettingsForFlow stores pending permission context', () async {
    final adapter = _FakeAdapter();
    final broker = PermissionBroker(adapter: adapter);
    final flow = PermissionFlowContext.create(
      scene: PermissionScene.notificationSubscription,
      permissionType: PermissionType.notification,
      source: PermissionFlowSource.notificationToggle,
      returnTarget: AppRecentTargetType.settings,
      fallbackAction: PermissionFallbackAction.openMessageCenter,
      intentId: 'settings_notification_toggle',
      recoveryId: 'settings_notification_toggle',
    );

    final opened = await broker.openSettingsForFlow(flow);

    expect(opened, isTrue);
    expect(adapter.openSettingsCalled, isTrue);
    expect(
      AppRecoveryStore.peekPendingPermissionContext()?.status,
      PermissionFlowStatus.settingsReturnPending,
    );
  });
}

class _FakeAdapter implements PermissionBrokerAdapter {
  @override
  final bool isIOS;
  @override
  final bool isAndroid;
  @override
  final int? androidSdkInt;
  final BrokerSystemPermissionStatus notificationStatus;
  final BrokerSystemPermissionStatus notificationRequestStatus;
  final BrokerSystemPermissionStatus photoStatus;
  final BrokerSystemPermissionStatus photoRequestStatus;
  final BrokerSystemPermissionStatus cameraStatus;
  final BrokerSystemPermissionStatus cameraRequestStatus;
  final List<String> lostMediaBase64;
  bool openSettingsCalled = false;

  _FakeAdapter({
    this.isIOS = false,
    this.isAndroid = false,
    this.androidSdkInt,
    this.notificationStatus = BrokerSystemPermissionStatus.denied,
    this.notificationRequestStatus = BrokerSystemPermissionStatus.denied,
    this.photoStatus = BrokerSystemPermissionStatus.denied,
    this.photoRequestStatus = BrokerSystemPermissionStatus.denied,
    this.cameraStatus = BrokerSystemPermissionStatus.denied,
    this.cameraRequestStatus = BrokerSystemPermissionStatus.denied,
    this.lostMediaBase64 = const <String>[],
  });

  @override
  Future<BrokerSystemPermissionStatus> getCameraStatus() async => cameraStatus;

  @override
  Future<BrokerSystemPermissionStatus> getNotificationStatus() async =>
      notificationStatus;

  @override
  Future<BrokerSystemPermissionStatus> getPhotoStatus() async => photoStatus;

  @override
  Future<bool> openAppSettings() async {
    openSettingsCalled = true;
    return true;
  }

  @override
  Future<BrokerSystemPermissionStatus> requestCameraPermission() async =>
      cameraRequestStatus;

  @override
  Future<BrokerSystemPermissionStatus> requestNotificationPermission() async =>
      notificationRequestStatus;

  @override
  Future<BrokerSystemPermissionStatus> requestPhotoPermission() async =>
      photoRequestStatus;

  @override
  Future<List<String>> retrieveLostMediaBase64() async => lostMediaBase64;
}
