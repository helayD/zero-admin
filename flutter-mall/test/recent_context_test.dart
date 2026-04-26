import 'dart:convert';

import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();
  });

  String buildJwtToken(int memberId) {
    final header = base64Url
        .encode(utf8.encode(jsonEncode({'alg': 'HS256', 'typ': 'JWT'})))
        .replaceAll('=', '');
    final payload = base64Url
        .encode(
            utf8.encode(jsonEncode({'memberId': memberId, 'exp': 9999999999})))
        .replaceAll('=', '');
    return 'Bearer $header.$payload.signature';
  }

  test('recent context persistence is disabled', () async {
    await AppRecoveryStore.saveRecentContext(
      AppRecentContext.create(
        targetType: AppRecentTargetType.orderList,
        tabIndex: 1,
        source: 'manual_open',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.home,
        fallbackTabIndex: 0,
      ),
    );

    expect(AppRecoveryStore.getRecentContext(), isNull);
    expect(AppRecoveryStore.peekActiveIntentCandidate(), isNull);
    expect(AppRecoveryStore.peekPendingIntent(), isNull);
    expect(AppRecoveryStore.peekPendingUpgradeContext(), isNull);
  });

  test('auth token helpers still bind to the current member', () async {
    await SharedPreferencesUtil.saveString(token, buildJwtToken(1001));

    expect(AppRecoveryStore.hasValidToken(), isTrue);
    expect(AppRecoveryStore.getCurrentMemberId(), 1001);
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

  test('structured recall payload tolerates polluted zero optional fields', () {
    final context = AppRecentContext.fromRecallPayload({
      'intentType': 'coupon_center_recall',
      'targetType': 'coupon_center',
      'targetId': 0,
      'targetTab': 0,
      'fallbackType': 'coupon_list',
      'fallbackTargetId': 0,
      'fallbackTab': 0,
      'requiresAuth': true,
      'intentId': 'member_message:22',
      'issuedAt': '2026-04-05T12:10:00',
      'source': 'member_message',
    });

    expect(context.targetType, AppRecentTargetType.couponCenter);
    expect(context.targetId, isNull);
    expect(context.tabIndex, isNull);
    expect(context.fallbackType, AppRecentTargetType.couponList);
    expect(context.fallbackTargetId, isNull);
    expect(context.fallbackTabIndex, 0);
    expect(context.validateReason(), isNull);
  });
}
