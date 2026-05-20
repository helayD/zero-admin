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

  test('daily login points reward is consumed once on the same day', () async {
    final now = DateTime(2026, 5, 20, 9);

    await AppRecoveryStore.markDailyLoginPointsRewardPending(now: now);

    expect(
      await AppRecoveryStore.consumeDailyLoginPointsReward(now: now),
      AppRecoveryStore.dailyLoginPointsRewardAmount,
    );
    expect(
      await AppRecoveryStore.consumeDailyLoginPointsReward(now: now),
      isNull,
    );
  });

  test('daily login points reward does not requeue after it was shown today',
      () async {
    final now = DateTime(2026, 5, 20, 9);

    await AppRecoveryStore.markDailyLoginPointsRewardPending(now: now);
    await AppRecoveryStore.consumeDailyLoginPointsReward(now: now);
    await AppRecoveryStore.markDailyLoginPointsRewardPending(now: now);

    expect(
      await AppRecoveryStore.consumeDailyLoginPointsReward(now: now),
      isNull,
    );
  });

  test('stale daily login points reward is dropped', () async {
    await AppRecoveryStore.markDailyLoginPointsRewardPending(
      now: DateTime(2026, 5, 19, 22),
    );

    expect(
      await AppRecoveryStore.consumeDailyLoginPointsReward(
        now: DateTime(2026, 5, 20, 9),
      ),
      isNull,
    );
  });
}
