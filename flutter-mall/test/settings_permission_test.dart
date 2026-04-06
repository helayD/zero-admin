import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/mine/setting/settings.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  testWidgets(
      'settings notification keeps switch off and offers message center',
      (tester) async {
    tester.view.physicalSize = const Size(1200, 2000);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final provider = AppLifecycleProvider();
    final broker = PermissionBroker(
      adapter: _SettingsAdapter(
        notificationStatus: BrokerSystemPermissionStatus.denied,
        notificationRequestStatus: BrokerSystemPermissionStatus.denied,
        isAndroid: true,
        androidSdkInt: 33,
      ),
    );

    await tester.pumpWidget(
      ChangeNotifierProvider<AppLifecycleProvider>.value(
        value: provider,
        child: MaterialApp(
          home: Settings(permissionBroker: broker),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byType(Switch));
    await tester.pumpAndSettle();
    expect(find.text('继续授权'), findsOneWidget);

    await tester.tap(find.text('继续授权'));
    await tester.pumpAndSettle();
    expect(find.text('查看站内消息'), findsOneWidget);
    expect(AppRecoveryStore.getNotificationPreferenceEnabled(), isFalse);
  });

  testWidgets(
      'settings notification on android 12 opens system settings instead of looping request',
      (tester) async {
    tester.view.physicalSize = const Size(1200, 2000);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final provider = AppLifecycleProvider();
    final adapter = _SettingsAdapter(
      notificationStatus: BrokerSystemPermissionStatus.denied,
      notificationRequestStatus: BrokerSystemPermissionStatus.denied,
      isAndroid: true,
      androidSdkInt: 32,
    );
    final broker = PermissionBroker(adapter: adapter);

    await tester.pumpWidget(
      ChangeNotifierProvider<AppLifecycleProvider>.value(
        value: provider,
        child: MaterialApp(
          home: Settings(permissionBroker: broker),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byType(Switch));
    await tester.pumpAndSettle();

    expect(find.text('去系统设置'), findsOneWidget);

    await tester.tap(find.text('去系统设置'));
    await tester.pumpAndSettle();

    expect(adapter.openSettingsCalled, isTrue);
    expect(
      AppRecoveryStore.peekPendingPermissionContext()?.status,
      PermissionFlowStatus.settingsReturnPending,
    );
  });
}

class _SettingsAdapter implements PermissionBrokerAdapter {
  @override
  final bool isAndroid;
  @override
  final int? androidSdkInt;
  final BrokerSystemPermissionStatus notificationStatus;
  final BrokerSystemPermissionStatus notificationRequestStatus;

  _SettingsAdapter({
    required this.notificationStatus,
    required this.notificationRequestStatus,
    this.isAndroid = false,
    this.androidSdkInt,
  });

  bool openSettingsCalled = false;

  @override
  bool get isIOS => false;

  @override
  Future<BrokerSystemPermissionStatus> getCameraStatus() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<BrokerSystemPermissionStatus> getNotificationStatus() async =>
      notificationStatus;

  @override
  Future<BrokerSystemPermissionStatus> getPhotoStatus() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<bool> openAppSettings() async {
    openSettingsCalled = true;
    return true;
  }

  @override
  Future<BrokerSystemPermissionStatus> requestCameraPermission() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<BrokerSystemPermissionStatus> requestNotificationPermission() async =>
      notificationRequestStatus;

  @override
  Future<BrokerSystemPermissionStatus> requestPhotoPermission() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<List<String>> retrieveLostMediaBase64() async => const <String>[];
}
