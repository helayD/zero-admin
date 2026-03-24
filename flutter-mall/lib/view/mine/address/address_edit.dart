import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';

import '../../../model/address_list.dart';

///
/// 地址编辑页面（新增/编辑双模式）
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class AddressEdit extends StatefulWidget {
  final AddressListData? addressData;

  const AddressEdit({super.key, this.addressData});

  @override
  State<AddressEdit> createState() => _AddressEditState();
}

class _AddressEditState extends State<AddressEdit> {
  final _formKey = GlobalKey<FormState>();

  late TextEditingController _nameController;
  late TextEditingController _phoneController;
  late TextEditingController _provinceController;
  late TextEditingController _cityController;
  late TextEditingController _districtController;
  late TextEditingController _detailAddressController;
  late TextEditingController _postalCodeController;

  final List<String> _tagOptions = ['家', '公司', '学校', '其他'];
  late String _selectedTag;
  late bool _isDefault;
  bool _isSubmitting = false;

  bool get _isEditMode => widget.addressData != null;

  @override
  void initState() {
    super.initState();
    final data = widget.addressData;
    _nameController = TextEditingController(text: data?.receiverName ?? '');
    _phoneController = TextEditingController(text: data?.receiverPhone ?? '');
    _provinceController = TextEditingController(text: data?.province ?? '');
    _cityController = TextEditingController(text: data?.city ?? '');
    _districtController = TextEditingController(text: data?.district ?? '');
    _detailAddressController =
        TextEditingController(text: data?.detailAddress ?? '');
    _postalCodeController =
        TextEditingController(text: data?.postalCode ?? '');
    _selectedTag = (data?.tag != null && data!.tag.isNotEmpty) ? data.tag : '家';
    _isDefault = data?.isDefault == 1;
  }

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _provinceController.dispose();
    _cityController.dispose();
    _districtController.dispose();
    _detailAddressController.dispose();
    _postalCodeController.dispose();
    super.dispose();
  }

  Future<void> _submitForm() async {
    if (!_formKey.currentState!.validate()) return;
    if (_isSubmitting) return;

    setState(() {
      _isSubmitting = true;
    });

    try {
      Map<String, dynamic> requestMap = {
        "receiverName": _nameController.text.trim(),
        "receiverPhone": _phoneController.text.trim(),
        "province": _provinceController.text.trim(),
        "city": _cityController.text.trim(),
        "district": _districtController.text.trim(),
        "detailAddress": _detailAddressController.text.trim(),
        "postalCode": _postalCodeController.text.trim(),
        "tag": _selectedTag,
        "isDefault": _isDefault ? 1 : 0,
      };

      Response result;
      if (_isEditMode) {
        requestMap["id"] = widget.addressData!.id;
        result = await HttpUtil.post(updateAddressDataUrl, data: requestMap);
      } else {
        result = await HttpUtil.post(addAddressDataUrl, data: requestMap);
      }

      if (!mounted) return;

      if (result.data["code"] == 0) {
        Navigator.pop(context, true);
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(result.data["message"] ?? "保存失败")),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("网络请求失败: $e")),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final themeColor =
        Color(int.parse('fa436a', radix: 16)).withAlpha(255);
    final bgColor =
        Color(int.parse('f5f5f5', radix: 16)).withAlpha(255);

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: Text(
          _isEditMode ? "编辑收货地址" : "新增收货地址",
          style: const TextStyle(fontSize: 16, color: Colors.black),
        ),
        centerTitle: true,
      ),
      body: Container(
        color: bgColor,
        child: Form(
          key: _formKey,
          child: ListView(
            children: [
              _buildTextField(
                controller: _nameController,
                label: "收件人姓名",
                maxLength: 20,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return "请输入收件人姓名";
                  }
                  return null;
                },
              ),
              _buildTextField(
                controller: _phoneController,
                label: "手机号码",
                keyboardType: TextInputType.phone,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return "请输入手机号码";
                  }
                  if (!RegExp(r'^1[3-9]\d{9}$').hasMatch(value.trim())) {
                    return "请输入正确的11位手机号码";
                  }
                  return null;
                },
              ),
              _buildTextField(
                controller: _provinceController,
                label: "省份",
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return "请输入省份";
                  }
                  return null;
                },
              ),
              _buildTextField(
                controller: _cityController,
                label: "城市",
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return "请输入城市";
                  }
                  return null;
                },
              ),
              _buildTextField(
                controller: _districtController,
                label: "区县",
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return "请输入区县";
                  }
                  return null;
                },
              ),
              _buildTextField(
                controller: _detailAddressController,
                label: "详细地址",
                maxLength: 100,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return "请输入详细地址";
                  }
                  return null;
                },
              ),
              _buildTextField(
                controller: _postalCodeController,
                label: "邮政编码（选填）",
                keyboardType: TextInputType.number,
                validator: (value) {
                  if (value != null &&
                      value.trim().isNotEmpty &&
                      !RegExp(r'^\d{6}$').hasMatch(value.trim())) {
                    return "邮政编码需为6位数字";
                  }
                  return null;
                },
              ),
              // 地址标签
              Container(
                color: Colors.white,
                padding:
                    const EdgeInsets.symmetric(horizontal: 15, vertical: 10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "地址标签",
                      style: TextStyle(
                        fontSize: 14,
                        color: Color(int.parse('909399', radix: 16))
                            .withAlpha(255),
                      ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      children: _tagOptions.map((tag) {
                        final isSelected = _selectedTag == tag;
                        return ChoiceChip(
                          label: Text(tag),
                          selected: isSelected,
                          selectedColor: themeColor.withAlpha(50),
                          onSelected: (selected) {
                            if (selected) {
                              setState(() {
                                _selectedTag = tag;
                              });
                            }
                          },
                        );
                      }).toList(),
                    ),
                  ],
                ),
              ),
              // 设为默认
              Container(
                color: Colors.white,
                margin: const EdgeInsets.only(top: 8, bottom: 20),
                padding:
                    const EdgeInsets.symmetric(vertical: 7, horizontal: 15),
                child: Row(
                  children: [
                    Expanded(
                      flex: 1,
                      child: Text(
                        "设为默认",
                        style: TextStyle(
                          fontSize: 15,
                          color: Color(int.parse('303133', radix: 16))
                              .withAlpha(255),
                        ),
                      ),
                    ),
                    Switch(
                      value: _isDefault,
                      materialTapTargetSize:
                          MaterialTapTargetSize.shrinkWrap,
                      onChanged: (value) {
                        setState(() {
                          _isDefault = value;
                        });
                      },
                      activeColor: Colors.white,
                      activeTrackColor: themeColor,
                    ),
                  ],
                ),
              ),
              // 提交按钮
              InkWell(
                onTap: _isSubmitting ? null : _submitForm,
                child: Container(
                  alignment: Alignment.center,
                  width: MediaQuery.of(context).size.width,
                  height: 40,
                  margin: const EdgeInsets.all(15),
                  decoration: BoxDecoration(
                    color: _isSubmitting
                        ? themeColor.withAlpha(128)
                        : themeColor,
                    borderRadius: BorderRadius.circular(5),
                  ),
                  child: _isSubmitting
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor:
                                AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        )
                      : const Text(
                          '保存',
                          style:
                              TextStyle(color: Colors.white, fontSize: 16),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    int? maxLength,
    TextInputType? keyboardType,
    String? Function(String?)? validator,
  }) {
    final border = BorderSide(
      width: 1,
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
    );
    return Container(
      decoration:
          BoxDecoration(color: Colors.white, border: Border(bottom: border)),
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 5),
      child: TextFormField(
        controller: controller,
        maxLength: maxLength,
        keyboardType: keyboardType,
        decoration: InputDecoration(
          labelText: label,
          border: InputBorder.none,
          counterText: '',
        ),
        validator: validator,
      ),
    );
  }
}
