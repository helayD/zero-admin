import 'package:flutter_mall/config/order_status.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('order list tabs map to backend order statuses', () {
    expect(flutterTabToBackendStatus(0), 0);
    expect(flutterTabToBackendStatus(1), 1);
    expect(flutterTabToBackendStatus(2), 2);
    expect(flutterTabToBackendStatus(3), 4);
    expect(flutterTabToBackendStatus(4), 5);
    expect(flutterTabToBackendStatus(5), 7);
  });

  test('after-sales status text is user facing', () {
    expect(getOmsOrderStatusTxt(7), '售后中');
  });
}
