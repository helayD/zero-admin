import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/after_sales.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/mine/order/apply_after_sales.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  testWidgets('after sales keeps form content and falls back to gallery',
      (tester) async {
    tester.view.physicalSize = const Size(1200, 2000);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    final broker = _FakePagePermissionBroker(
      inspectStatuses: <PermissionFlowSource, PermissionFlowStatus>{
        PermissionFlowSource.camera: PermissionFlowStatus.denied,
        PermissionFlowSource.gallery: PermissionFlowStatus.notRequired,
      },
      requestStatuses: <PermissionFlowSource, PermissionFlowStatus>{
        PermissionFlowSource.camera: PermissionFlowStatus.denied,
      },
      mediaResults: <PermissionFlowSource, List<String>>{
        PermissionFlowSource.gallery: <String>['aGVsbG8='],
      },
    );

    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AppLifecycleProvider(),
        child: MaterialApp(
          home: ApplyAfterSales(
            orderId: 3001,
            permissionBroker: broker,
            reasonLoader: () async => ReturnReasonListData(
              code: 0,
              message: 'ok',
              reasonList: <ReturnReasonItemData>[
                ReturnReasonItemData(id: 1, name: '质量问题'),
              ],
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextFormField).first,
      '这是保留下来的售后描述内容',
    );

    await tester.ensureVisible(find.byIcon(Icons.add_a_photo).first);
    await tester.tap(find.byIcon(Icons.add_a_photo).first);
    await tester.pumpAndSettle();
    await tester.tap(find.text('拍照'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('继续授权'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('改用相册'));
    await tester.pumpAndSettle();

    expect(find.text('这是保留下来的售后描述内容'), findsOneWidget);
    expect(find.text('1/3'), findsOneWidget);
    expect(broker.pickedSources, contains(PermissionFlowSource.gallery));
  });

  testWidgets('after sales success clears recoverable context', (tester) async {
    tester.view.physicalSize = const Size(1200, 2000);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    await AppRecoveryStore.saveAfterSalesDraft(
      AfterSalesDraftSnapshot(
        orderId: 3001,
        typeValue: AfterSalesType.returnRefund.value,
        reasonId: 1,
        reasonName: '质量问题',
        description: '待提交的售后说明',
        proofPics: const <String>[],
        updatedAt: DateTime.now(),
      ),
    );

    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AppLifecycleProvider(),
        child: MaterialApp(
          home: ApplyAfterSales(
            orderId: 3001,
            permissionBroker: _FakePagePermissionBroker(),
            reasonLoader: () async => ReturnReasonListData(
              code: 0,
              message: 'ok',
              reasonList: <ReturnReasonItemData>[
                ReturnReasonItemData(id: 1, name: '质量问题'),
              ],
            ),
            submitter: (_) async => ApplyAfterSalesRespData(
              code: 0,
              message: 'ok',
              returnId: 1,
              returnNo: 'TH20260406001',
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      AppRecoveryStore.getRecentContext()?.targetType,
      AppRecentTargetType.afterSalesApply,
    );
    expect(AppRecoveryStore.getAfterSalesDraft(3001), isNotNull);

    await tester.tap(find.text('提交申请'));
    await tester.pumpAndSettle();

    expect(AppRecoveryStore.getAfterSalesDraft(3001), isNull);
    expect(AppRecoveryStore.getRecentContext(), isNull);
  });
}

class _FakePagePermissionBroker extends PermissionBroker {
  final Map<PermissionFlowSource, PermissionFlowStatus> inspectStatuses;
  final Map<PermissionFlowSource, PermissionFlowStatus> requestStatuses;
  final Map<PermissionFlowSource, List<String>> mediaResults;
  final List<PermissionFlowSource> pickedSources = <PermissionFlowSource>[];

  _FakePagePermissionBroker({
    this.inspectStatuses = const <PermissionFlowSource, PermissionFlowStatus>{},
    this.requestStatuses = const <PermissionFlowSource, PermissionFlowStatus>{},
    this.mediaResults = const <PermissionFlowSource, List<String>>{},
  }) : super(adapter: _NoopAdapter());

  @override
  Future<PermissionInspectionResult> inspect(PermissionFlowContext flow) async {
    return PermissionInspectionResult(
      flow: flow,
      status: inspectStatuses[flow.source] ?? PermissionFlowStatus.denied,
    );
  }

  @override
  Future<PermissionInspectionResult> request(PermissionFlowContext flow) async {
    return PermissionInspectionResult(
      flow: flow,
      status: requestStatuses[flow.source] ?? PermissionFlowStatus.denied,
    );
  }

  @override
  Future<PermissionMediaResult> pickImageAsBase64(
      PermissionFlowContext flow) async {
    pickedSources.add(flow.source);
    return PermissionMediaResult(
      flow: flow,
      status: flow.status,
      mediaBase64List: mediaResults[flow.source] ?? const <String>[],
    );
  }
}

class _NoopAdapter implements PermissionBrokerAdapter {
  const _NoopAdapter();

  @override
  bool get isAndroid => false;

  @override
  int? get androidSdkInt => null;

  @override
  bool get isIOS => false;

  @override
  Future<BrokerSystemPermissionStatus> getCameraStatus() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<BrokerSystemPermissionStatus> getNotificationStatus() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<BrokerSystemPermissionStatus> getPhotoStatus() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<bool> openAppSettings() async => true;

  @override
  Future<BrokerSystemPermissionStatus> requestCameraPermission() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<BrokerSystemPermissionStatus> requestNotificationPermission() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<BrokerSystemPermissionStatus> requestPhotoPermission() async =>
      BrokerSystemPermissionStatus.denied;

  @override
  Future<List<String>> retrieveLostMediaBase64() async => const <String>[];
}
