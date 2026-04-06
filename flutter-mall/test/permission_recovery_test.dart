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

  test('captureLostMediaOnLaunch stores pending media for matching recovery id',
      () async {
    final flow = PermissionFlowContext.create(
      scene: PermissionScene.afterSalesProof,
      permissionType: PermissionType.camera,
      source: PermissionFlowSource.camera,
      returnTarget: AppRecentTargetType.afterSalesApply,
      returnTargetId: 3001,
      orderId: 3001,
      fallbackAction: PermissionFallbackAction.switchToGallery,
      intentId: 'after_sales_camera_3001',
      recoveryId: 'after_sales_camera_3001',
      retryOnResume: true,
    );
    await AppRecoveryStore.savePendingPermissionContext(flow);

    final broker = PermissionBroker(
      adapter: _LostMediaAdapter(
        lostMediaBase64: const <String>['aGVsbG8='],
      ),
    );

    await broker.captureLostMediaOnLaunch();

    final lostMedia = AppRecoveryStore.peekPendingLostMedia();
    expect(lostMedia, isNotNull);
    expect(lostMedia!.recoveryId, flow.recoveryId);
    expect(lostMedia.mediaBase64List, const <String>['aGVsbG8=']);
  });

  test('clearRecoveryState removes notification preference and all drafts',
      () async {
    await AppRecoveryStore.saveNotificationPreferenceEnabled(true);
    await AppRecoveryStore.saveCommentDraft(
      CommentDraftSnapshot(
        orderId: 1001,
        productId: 2001,
        productName: '测试商品A',
        productPic: 'a.png',
        productAttribute: '红色',
        memberNickName: '张三',
        starRating: 5,
        content: 'draft-a',
        pics: const <String>['aGVsbG8='],
        updatedAt: DateTime.now(),
      ),
    );
    await AppRecoveryStore.saveCommentDraft(
      CommentDraftSnapshot(
        orderId: 1002,
        productId: 2002,
        productName: '测试商品B',
        productPic: 'b.png',
        productAttribute: '蓝色',
        memberNickName: '李四',
        starRating: 4,
        content: 'draft-b',
        pics: const <String>[],
        updatedAt: DateTime.now(),
      ),
    );
    await AppRecoveryStore.saveAfterSalesDraft(
      AfterSalesDraftSnapshot(
        orderId: 3001,
        typeValue: 0,
        reasonId: 1,
        reasonName: '质量问题',
        description: 'after-sales-a',
        proofPics: const <String>['aGVsbG8='],
        updatedAt: DateTime.now(),
      ),
    );
    await AppRecoveryStore.saveAfterSalesDraft(
      AfterSalesDraftSnapshot(
        orderId: 3002,
        typeValue: 1,
        reasonId: 2,
        reasonName: '尺寸不合适',
        description: 'after-sales-b',
        proofPics: const <String>[],
        updatedAt: DateTime.now(),
      ),
    );
    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.settings,
        source: 'manual_open',
        requiresAuth: false,
        fallbackType: AppRecentTargetType.home,
      ),
    );

    await AppRecoveryStore.clearRecoveryState();

    expect(AppRecoveryStore.getNotificationPreferenceEnabled(), isFalse);
    expect(AppRecoveryStore.getCommentDraft(1001), isNull);
    expect(AppRecoveryStore.getCommentDraft(1002), isNull);
    expect(AppRecoveryStore.getAfterSalesDraft(3001), isNull);
    expect(AppRecoveryStore.getAfterSalesDraft(3002), isNull);
    expect(AppRecoveryStore.getRecentContext(), isNull);
  });
}

class _LostMediaAdapter implements PermissionBrokerAdapter {
  final List<String> lostMediaBase64;

  _LostMediaAdapter({required this.lostMediaBase64});

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
  Future<List<String>> retrieveLostMediaBase64() async => lostMediaBase64;
}
