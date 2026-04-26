import 'package:flutter/material.dart';
import 'package:flutter_mall/layout/app_bootstrap.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('app bootstrap opens main tab without recovery gate',
      (tester) async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();

    await tester.pumpWidget(
      const MaterialApp(home: AppBootstrap()),
    );

    expect(find.byType(MainTab), findsOneWidget);
  });
}
