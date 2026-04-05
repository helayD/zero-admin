import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  String buildJwtToken(int memberId) {
    final header = base64Url.encode(utf8.encode(jsonEncode({'alg': 'HS256', 'typ': 'JWT'}))).replaceAll('=', '');
    final payload = base64Url
        .encode(utf8.encode(jsonEncode({'memberId': memberId, 'exp': 9999999999})))
        .replaceAll('=', '');
    return 'Bearer $header.$payload.signature';
  }

  test('recent context roundtrip persists valid order list target', () async {
    final context = AppRecentContext.create(
      targetType: AppRecentTargetType.orderList,
      tabIndex: 1,
      source: 'manual_open',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );

    await AppRecoveryStore.saveRecentContext(context);
    final restored = AppRecoveryStore.getRecentContext();

    expect(restored, isNotNull);
    expect(restored!.targetType, AppRecentTargetType.orderList);
    expect(restored.tabIndex, 1);
    expect(restored.requiresAuth, isTrue);
  });

  test('product detail without target id is rejected', () {
    final parsed = AppRecentContext.tryParse({
      'version': AppRecentContext.currentVersion,
      'targetType': 'product_detail',
      'targetId': null,
      'tabIndex': null,
      'source': 'manual_open',
      'requiresAuth': false,
      'capturedAt': DateTime.now().toIso8601String(),
      'lastValidatedAt': DateTime.now().toIso8601String(),
      'fallbackType': 'home',
      'fallbackTabIndex': 0,
    });

    expect(parsed, isNotNull);
    expect(parsed!.validateReason(), 'invalid_target_id');
    expect(parsed.isRecoverable, isFalse);
  });

  test('unsupported version is rejected', () {
    final parsed = AppRecentContext.tryParse({
      'version': 999,
      'targetType': 'home',
      'targetId': null,
      'tabIndex': 0,
      'source': 'manual_open',
      'requiresAuth': false,
      'capturedAt': DateTime.now().toIso8601String(),
      'lastValidatedAt': DateTime.now().toIso8601String(),
      'fallbackType': 'home',
      'fallbackTabIndex': 0,
    });

    expect(parsed, isNotNull);
    expect(parsed!.validateReason(), 'unsupported_version');
  });

  test('auth required context binds current member id from token snapshot', () async {
    await SharedPreferencesUtil.saveString(token, buildJwtToken(1001));

    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.orderDetail,
        targetId: 88,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderList,
        fallbackTabIndex: 1,
      ),
    );

    final restored = AppRecoveryStore.getRecentContext();
    expect(restored, isNotNull);
    expect(restored!.ownerMemberId, 1001);
  });

  test('auth required recent context is rejected after member switch', () async {
    await SharedPreferencesUtil.saveString(token, buildJwtToken(1001));
    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.orderDetail,
        targetId: 99,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderList,
        fallbackTabIndex: 1,
      ),
    );

    await SharedPreferencesUtil.saveString(token, buildJwtToken(2002));

    expect(AppRecoveryStore.getRecentContext(), isNull);
  });
}
