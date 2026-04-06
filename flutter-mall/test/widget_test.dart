import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/layout/app_bootstrap.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('app bootstrap smoke test', (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferencesUtil.init();

    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AppLifecycleProvider(),
        child: MaterialApp(
          home: AppBootstrap(
            fallbackBuilder: (_) => const Scaffold(
              body: Text('fallback-page'),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('fallback-page'), findsOneWidget);
  });
}
