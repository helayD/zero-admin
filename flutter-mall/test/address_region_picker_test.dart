import 'package:flutter_mall/view/mine/address/address_region_picker.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('region asset loads province city and district hierarchy', () async {
    final store = await AddressRegionStore.load();

    final guangdong = store.provinces.firstWhere(
      (region) => region.name == '广东省',
    );
    final shenzhen = store.childrenOf(guangdong.code).firstWhere(
          (region) => region.name == '深圳市',
        );
    final nanshan = store.childrenOf(shenzhen.code).firstWhere(
          (region) => region.name == '南山区',
        );

    expect(guangdong.level, 1);
    expect(shenzhen.level, 2);
    expect(nanshan.level, 3);

    final selection = store.resolveByNames(
      provinceName: '广东省',
      cityName: '深圳市',
      districtName: '南山区',
    );
    expect(selection?.displayText, '广东省 深圳市 南山区');
  });
}
