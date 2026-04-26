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

  test('pending login recovery intent is no longer persisted', () async {
    await AppRecoveryStore.savePendingIntent(
      AppRecentContext.create(
        targetType: AppRecentTargetType.orderDetail,
        targetId: 1001,
        source: 'resume',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderList,
        fallbackTabIndex: 1,
      ),
    );

    expect(AppRecoveryStore.peekPendingIntent(), isNull);
    expect(await AppRecoveryStore.consumePendingIntent(), isNull);
  });
}
