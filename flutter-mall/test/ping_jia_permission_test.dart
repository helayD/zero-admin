import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/model/comment_model.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/mine/ping_jia/ping_jia.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  testWidgets('ping jia keeps draft and falls back from camera to gallery',
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
          home: PinJia(
            orderId: 1001,
            productId: 2001,
            productName: '测试商品',
            permissionBroker: broker,
            commentListLoader: ({
              required int productId,
              required int page,
              required int pageSize,
            }) async {
              return CommentListModel(
                code: 0,
                message: 'ok',
                data: <CommentItem>[],
                total: 0,
              );
            },
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final draft = AppRecoveryStore.getCommentDraft(1001);
    expect(draft, isNotNull);
    expect(draft!.orderId, 1001);
    expect(draft.productId, 2001);

    await tester.enterText(
      find.byType(TextFormField).first,
      '这是保留下来的评价文本内容',
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

    expect(find.text('这是保留下来的评价文本内容'), findsOneWidget);
    expect(find.text('1/3'), findsOneWidget);
    expect(broker.pickedSources, contains(PermissionFlowSource.gallery));
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
