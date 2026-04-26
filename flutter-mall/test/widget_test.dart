import 'package:flutter_mall/layout/app_bootstrap.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('app bootstrap is available', () {
    expect(const AppBootstrap(), isA<AppBootstrap>());
  });
}
