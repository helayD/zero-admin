import 'package:flutter_mall/model/address_list.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('address list model treats null data as an empty list', () {
    final model = AddressListModel.fromJson({
      'code': 0,
      'message': 'success',
      'current': 1,
      'data': null,
      'pageSize': 20,
      'success': true,
      'total': 0,
    });

    expect(model.code, 0);
    expect(model.data, isEmpty);
    expect(model.success, isTrue);
  });

  test('address list data tolerates missing optional response fields', () {
    final data = AddressListData.fromJson({
      'id': 12,
      'memberId': 4,
      'receiverName': '张三',
      'receiverPhone': '16698129676',
      'province': '广东省',
      'city': '深圳市',
      'district': '南山区',
      'detailAddress': '科技园测试路1号',
      'isDefault': 1,
    });

    expect(data.postalCode, '');
    expect(data.tag, '');
    expect(data.fullAddress, '广东省 深圳市 南山区 科技园测试路1号');
    expect(data.contactText, '张三  16698129676');
  });
}
