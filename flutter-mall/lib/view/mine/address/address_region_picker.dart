import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show rootBundle;

class AddressRegion {
  final String code;
  final String name;
  final String parentCode;
  final int level;

  const AddressRegion({
    required this.code,
    required this.name,
    required this.parentCode,
    required this.level,
  });

  factory AddressRegion.fromJson(Map<String, dynamic> json) {
    return AddressRegion(
      code: json['code']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      parentCode: json['parentCode']?.toString() ?? '',
      level: json['level'] is int
          ? json['level'] as int
          : int.tryParse(json['level']?.toString() ?? '') ?? 0,
    );
  }
}

class AddressRegionSelection {
  final AddressRegion? province;
  final AddressRegion? city;
  final AddressRegion? district;
  final String provinceName;
  final String cityName;
  final String districtName;

  const AddressRegionSelection({
    this.province,
    this.city,
    this.district,
    required this.provinceName,
    required this.cityName,
    required this.districtName,
  });

  factory AddressRegionSelection.fromNames({
    required String provinceName,
    required String cityName,
    required String districtName,
  }) {
    return AddressRegionSelection(
      provinceName: provinceName,
      cityName: cityName,
      districtName: districtName,
    );
  }

  bool get isComplete {
    return provinceName.trim().isNotEmpty &&
        cityName.trim().isNotEmpty &&
        districtName.trim().isNotEmpty;
  }

  String get displayText {
    return [provinceName, cityName, districtName]
        .where((item) => item.trim().isNotEmpty)
        .join(' ');
  }
}

class AddressRegionStore {
  static AddressRegionStore? _cache;

  late final Map<String, List<AddressRegion>> _childrenByParent;

  AddressRegionStore._(List<AddressRegion> regions) {
    _childrenByParent = <String, List<AddressRegion>>{};
    for (final region in regions) {
      _childrenByParent
          .putIfAbsent(region.parentCode, () => <AddressRegion>[])
          .add(region);
    }
  }

  static Future<AddressRegionStore> load() async {
    final cached = _cache;
    if (cached != null) {
      return cached;
    }
    final raw = await rootBundle.loadString('assets/data/china_regions.json');
    final decoded = jsonDecode(raw) as List<dynamic>;
    final store = AddressRegionStore._(
      decoded
          .whereType<Map>()
          .map(
              (item) => AddressRegion.fromJson(Map<String, dynamic>.from(item)))
          .toList(),
    );
    _cache = store;
    return store;
  }

  List<AddressRegion> get provinces => childrenOf('0');

  List<AddressRegion> childrenOf(String parentCode) {
    return _childrenByParent[parentCode] ?? const <AddressRegion>[];
  }

  AddressRegionSelection initialSelection(AddressRegionSelection? current) {
    final resolved = resolveByNames(
      provinceName: current?.provinceName ?? '',
      cityName: current?.cityName ?? '',
      districtName: current?.districtName ?? '',
    );
    if (resolved != null) {
      return resolved;
    }

    final province = provinces.isNotEmpty ? provinces.first : null;
    final city =
        province == null ? null : childrenOf(province.code).firstOrNull;
    final district = city == null ? null : childrenOf(city.code).firstOrNull;
    return _buildSelection(province, city, district);
  }

  AddressRegionSelection? resolveByNames({
    required String provinceName,
    required String cityName,
    required String districtName,
  }) {
    final normalizedProvince = provinceName.trim();
    final normalizedCity = cityName.trim();
    final normalizedDistrict = districtName.trim();
    if (normalizedProvince.isEmpty ||
        normalizedCity.isEmpty ||
        normalizedDistrict.isEmpty) {
      return null;
    }

    final province =
        provinces.where((item) => item.name == normalizedProvince).firstOrNull;
    if (province == null) {
      return AddressRegionSelection.fromNames(
        provinceName: normalizedProvince,
        cityName: normalizedCity,
        districtName: normalizedDistrict,
      );
    }

    final cities = childrenOf(province.code);
    AddressRegion? city =
        cities.where((item) => item.name == normalizedCity).firstOrNull;
    if (city == null && normalizedCity == normalizedProvince) {
      city = cities.where((item) => item.name == '市辖区').firstOrNull;
    }
    city ??= cities.firstOrNull;

    final districts =
        city == null ? const <AddressRegion>[] : childrenOf(city.code);
    final district =
        districts.where((item) => item.name == normalizedDistrict).firstOrNull;
    return _buildSelection(province, city, district);
  }

  AddressRegionSelection _buildSelection(
    AddressRegion? province,
    AddressRegion? city,
    AddressRegion? district,
  ) {
    return AddressRegionSelection(
      province: province,
      city: city,
      district: district,
      provinceName: province?.name ?? '',
      cityName: city?.name ?? '',
      districtName: district?.name ?? '',
    );
  }
}

Future<AddressRegionSelection?> showAddressRegionPicker({
  required BuildContext context,
  AddressRegionSelection? initialSelection,
}) async {
  final store = await AddressRegionStore.load();
  if (!context.mounted) return null;

  return showModalBottomSheet<AddressRegionSelection>(
    context: context,
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    builder: (context) {
      return _AddressRegionPickerSheet(
        store: store,
        initialSelection: store.initialSelection(initialSelection),
      );
    },
  );
}

class _AddressRegionPickerSheet extends StatefulWidget {
  final AddressRegionStore store;
  final AddressRegionSelection initialSelection;

  const _AddressRegionPickerSheet({
    required this.store,
    required this.initialSelection,
  });

  @override
  State<_AddressRegionPickerSheet> createState() =>
      _AddressRegionPickerSheetState();
}

class _AddressRegionPickerSheetState extends State<_AddressRegionPickerSheet> {
  static const Color _themeColor = Color(0xFFFA436A);
  static const Color _textPrimary = Color(0xFF303133);
  static const Color _textHint = Color(0xFF909399);

  late AddressRegion? _province;
  late AddressRegion? _city;
  late AddressRegion? _district;
  int _step = 0;

  @override
  void initState() {
    super.initState();
    _province = widget.initialSelection.province;
    _city = widget.initialSelection.city;
    _district = widget.initialSelection.district;
    if (_province != null && _city == null) {
      _city = widget.store.childrenOf(_province!.code).firstOrNull;
    }
    if (_city != null && _district == null) {
      _district = widget.store.childrenOf(_city!.code).firstOrNull;
    }
  }

  @override
  Widget build(BuildContext context) {
    final height = MediaQuery.of(context).size.height * 0.72;
    return Container(
      height: height,
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(22)),
      ),
      child: SafeArea(
        top: false,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(18, 14, 8, 8),
              child: Row(
                children: [
                  const Expanded(
                    child: Text(
                      '选择所在地区',
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.w700,
                        color: _textPrimary,
                      ),
                    ),
                  ),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close_rounded),
                  ),
                ],
              ),
            ),
            _buildTabs(),
            const Divider(height: 1, color: Color(0xFFEDEDED)),
            Expanded(child: _buildOptions()),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
              child: SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: _district == null ? null : _confirm,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: _themeColor,
                    foregroundColor: Colors.white,
                    disabledBackgroundColor: const Color(0xFFFFB3C5),
                    elevation: 0,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(14),
                    ),
                    textStyle: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  child: const Text('确定'),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTabs() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(14, 0, 14, 10),
      child: Row(
        children: [
          _buildTab(0, _province?.name ?? '省份'),
          const SizedBox(width: 8),
          _buildTab(1, _city?.name ?? '城市'),
          const SizedBox(width: 8),
          _buildTab(2, _district?.name ?? '区县'),
        ],
      ),
    );
  }

  Widget _buildTab(int step, String text) {
    final selected = _step == step;
    return Expanded(
      child: InkWell(
        borderRadius: BorderRadius.circular(999),
        onTap: () => setState(() => _step = step),
        child: Container(
          height: 38,
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFEDF2) : const Color(0xFFF5F5F5),
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? const Color(0xFFFFB3C5) : Colors.transparent,
            ),
          ),
          child: Text(
            text,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              fontSize: 14,
              fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
              color: selected ? _themeColor : _textHint,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildOptions() {
    final options = switch (_step) {
      0 => widget.store.provinces,
      1 => _province == null
          ? const <AddressRegion>[]
          : widget.store.childrenOf(_province!.code),
      _ => _city == null
          ? const <AddressRegion>[]
          : widget.store.childrenOf(_city!.code),
    };

    return ListView.separated(
      padding: const EdgeInsets.symmetric(vertical: 8),
      itemCount: options.length,
      separatorBuilder: (_, __) => const Divider(
        height: 1,
        indent: 18,
        endIndent: 18,
        color: Color(0xFFF2F2F2),
      ),
      itemBuilder: (context, index) {
        final option = options[index];
        final selected = switch (_step) {
          0 => option.code == _province?.code,
          1 => option.code == _city?.code,
          _ => option.code == _district?.code,
        };
        return ListTile(
          title: Text(
            option.name,
            style: TextStyle(
              fontSize: 15,
              color: selected ? _themeColor : _textPrimary,
              fontWeight: selected ? FontWeight.w700 : FontWeight.w400,
            ),
          ),
          trailing: selected
              ? const Icon(Icons.check_rounded, color: _themeColor)
              : null,
          onTap: () => _select(option),
        );
      },
    );
  }

  void _select(AddressRegion option) {
    setState(() {
      if (_step == 0) {
        _province = option;
        _city = widget.store.childrenOf(option.code).firstOrNull;
        _district = _city == null
            ? null
            : widget.store.childrenOf(_city!.code).firstOrNull;
        _step = 1;
      } else if (_step == 1) {
        _city = option;
        _district = widget.store.childrenOf(option.code).firstOrNull;
        _step = 2;
      } else {
        _district = option;
      }
    });
  }

  void _confirm() {
    final province = _province;
    final city = _city;
    final district = _district;
    if (province == null || city == null || district == null) {
      return;
    }
    Navigator.pop(
      context,
      AddressRegionSelection(
        province: province,
        city: city,
        district: district,
        provinceName: province.name,
        cityName: city.name,
        districtName: district.name,
      ),
    );
  }
}
