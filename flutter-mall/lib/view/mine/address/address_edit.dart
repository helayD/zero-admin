import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/address_list.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/mine/address/address_region_picker.dart';

///
/// 地址编辑页面（新增/编辑双模式）
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class AddressEdit extends StatefulWidget {
  final AddressListData? addressData;
  final bool initialDefault;

  const AddressEdit({
    super.key,
    this.addressData,
    this.initialDefault = false,
  });

  @override
  State<AddressEdit> createState() => _AddressEditState();
}

class _AddressEditState extends State<AddressEdit> {
  static final RegExp _phoneRegExp = RegExp(r'^1[3-9]\d{9}$');
  static final RegExp _postalCodeRegExp = RegExp(r'^\d{6}$');

  static const Color _themeColor = Color(0xFFFA436A);
  static const Color _pageBg = Color(0xFFF7F7F7);
  static const Color _textPrimary = Color(0xFF303133);
  static const Color _textHint = Color(0xFF909399);

  final _formKey = GlobalKey<FormState>();

  late final TextEditingController _nameController;
  late final TextEditingController _phoneController;
  late final TextEditingController _detailAddressController;
  late final TextEditingController _postalCodeController;

  final List<String> _tagOptions = const ['家', '公司', '学校', '父母家', '其他'];
  late String _selectedTag;
  late bool _isDefault;
  AddressRegionSelection? _regionSelection;
  String? _regionError;
  bool _isSubmitting = false;
  bool _submittedOnce = false;

  bool get _isEditMode => widget.addressData != null;
  bool get _hasCompleteRegion => _regionSelection?.isComplete ?? false;

  @override
  void initState() {
    super.initState();
    final data = widget.addressData;
    _nameController = TextEditingController(text: data?.receiverName ?? '');
    _phoneController = TextEditingController(text: data?.receiverPhone ?? '');
    final initialRegion = AddressRegionSelection.fromNames(
      provinceName: data?.province ?? '',
      cityName: data?.city ?? '',
      districtName: data?.district ?? '',
    );
    _regionSelection = initialRegion.displayText.isEmpty ? null : initialRegion;
    _detailAddressController =
        TextEditingController(text: data?.detailAddress ?? '');
    _postalCodeController = TextEditingController(text: data?.postalCode ?? '');
    _selectedTag = _tagOptions.contains(data?.tag) ? data!.tag : '家';
    _isDefault =
        data?.isDefault == 1 || (!_isEditMode && widget.initialDefault);
  }

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _detailAddressController.dispose();
    _postalCodeController.dispose();
    super.dispose();
  }

  Future<void> _submitForm() async {
    FocusScope.of(context).unfocus();
    setState(() {
      _submittedOnce = true;
      _regionError = _hasCompleteRegion ? null : '请选择省、市、区县';
    });

    if (!_formKey.currentState!.validate() ||
        !_hasCompleteRegion ||
        _isSubmitting) {
      return;
    }

    setState(() => _isSubmitting = true);
    final region = _regionSelection!;

    final requestMap = <String, dynamic>{
      'receiverName': _nameController.text.trim(),
      'receiverPhone': _phoneController.text.trim(),
      'province': region.provinceName.trim(),
      'city': region.cityName.trim(),
      'district': region.districtName.trim(),
      'detailAddress': _detailAddressController.text.trim(),
      'postalCode': _postalCodeController.text.trim(),
      'tag': _selectedTag,
      'isDefault': _isDefault ? 1 : 0,
    };

    try {
      final result = _isEditMode
          ? await HttpUtil.post(
              updateAddressDataUrl,
              data: <String, dynamic>{
                ...requestMap,
                'id': widget.addressData!.id,
              },
            )
          : await HttpUtil.post(addAddressDataUrl, data: requestMap);

      if (!mounted) return;
      if (result.data is Map && result.data['code'] == 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_isEditMode ? '地址已保存' : '地址已新增')),
        );
        Navigator.pop(context, true);
      } else {
        _showError(result.data?['message']?.toString() ?? '保存失败');
      }
    } catch (_) {
      if (!mounted) return;
      _showError('网络请求失败，请稍后重试');
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _pageBg,
      appBar: AppBar(
        backgroundColor: Colors.white,
        foregroundColor: Colors.black,
        elevation: 0,
        title: Text(
          _isEditMode ? '编辑收货地址' : '新增收货地址',
          style: const TextStyle(
            fontSize: 18,
            color: Colors.black,
            fontWeight: FontWeight.w700,
          ),
        ),
        centerTitle: true,
      ),
      body: SafeArea(
        child: Form(
          key: _formKey,
          autovalidateMode: _submittedOnce
              ? AutovalidateMode.onUserInteraction
              : AutovalidateMode.disabled,
          child: Column(
            children: [
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.fromLTRB(14, 12, 14, 18),
                  children: [
                    _buildSection(
                      children: [
                        _buildTextField(
                          controller: _nameController,
                          label: '收件人',
                          hintText: '请输入收件人姓名',
                          icon: Icons.person_outline_rounded,
                          maxLength: 20,
                          textInputAction: TextInputAction.next,
                          validator: (value) => _required(value, '请输入收件人姓名'),
                        ),
                        _buildTextField(
                          controller: _phoneController,
                          label: '手机号码',
                          hintText: '请输入 11 位手机号',
                          icon: Icons.phone_iphone_rounded,
                          keyboardType: TextInputType.phone,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                            LengthLimitingTextInputFormatter(11),
                          ],
                          textInputAction: TextInputAction.next,
                          validator: (value) {
                            final phone = value?.trim() ?? '';
                            if (phone.isEmpty) return '请输入手机号码';
                            if (!_phoneRegExp.hasMatch(phone)) {
                              return '请输入正确的 11 位手机号码';
                            }
                            return null;
                          },
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    _buildSection(
                      children: [
                        _buildRegionSelector(),
                        _buildTextField(
                          controller: _detailAddressController,
                          label: '详细地址',
                          hintText: '街道、门牌号等',
                          icon: Icons.home_work_outlined,
                          maxLength: 100,
                          minLines: 1,
                          maxLines: 3,
                          textInputAction: TextInputAction.newline,
                          validator: (value) {
                            final text = value?.trim() ?? '';
                            if (text.isEmpty) return '请输入详细地址';
                            if (text.length < 4) return '详细地址不能少于 4 个字';
                            return null;
                          },
                        ),
                        _buildTextField(
                          controller: _postalCodeController,
                          label: '邮政编码',
                          hintText: '选填',
                          icon: Icons.markunread_mailbox_outlined,
                          keyboardType: TextInputType.number,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                            LengthLimitingTextInputFormatter(6),
                          ],
                          validator: (value) {
                            final code = value?.trim() ?? '';
                            if (code.isNotEmpty &&
                                !_postalCodeRegExp.hasMatch(code)) {
                              return '邮政编码需为 6 位数字';
                            }
                            return null;
                          },
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    _buildTagSection(),
                    const SizedBox(height: 10),
                    _buildDefaultSwitch(),
                  ],
                ),
              ),
              _buildSubmitButton(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSection({required List<Widget> children}) {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 4, 14, 4),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Column(children: children),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hintText,
    required IconData icon,
    int? maxLength,
    int minLines = 1,
    int maxLines = 1,
    TextInputType? keyboardType,
    TextInputAction? textInputAction,
    List<TextInputFormatter>? inputFormatters,
    String? Function(String?)? validator,
  }) {
    return TextFormField(
      controller: controller,
      maxLength: maxLength,
      minLines: minLines,
      maxLines: maxLines,
      keyboardType: keyboardType,
      textInputAction: textInputAction,
      inputFormatters: inputFormatters,
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        prefixIcon: Icon(icon, color: _textHint),
        border: InputBorder.none,
        counterText: '',
      ),
      validator: validator,
    );
  }

  Widget _buildRegionSelector() {
    final regionText = _regionSelection?.displayText ?? '';
    final hasValue = regionText.isNotEmpty;
    final showError = _regionError != null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        InkWell(
          borderRadius: BorderRadius.circular(10),
          onTap: _openRegionPicker,
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Row(
              children: [
                const SizedBox(
                  width: 48,
                  child: Icon(Icons.place_outlined, color: _textHint),
                ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        '所在地区',
                        style: TextStyle(fontSize: 12, color: _textHint),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        hasValue ? regionText : '请选择省、市、区县',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          fontSize: 16,
                          color: hasValue ? _textPrimary : _textHint,
                        ),
                      ),
                    ],
                  ),
                ),
                const Icon(
                  Icons.chevron_right_rounded,
                  color: Color(0xFFC0C4CC),
                ),
              ],
            ),
          ),
        ),
        if (showError)
          Padding(
            padding: const EdgeInsets.only(left: 48, bottom: 8),
            child: Text(
              _regionError!,
              style: const TextStyle(fontSize: 12, color: Colors.redAccent),
            ),
          ),
      ],
    );
  }

  Future<void> _openRegionPicker() async {
    final selection = await showAddressRegionPicker(
      context: context,
      initialSelection: _regionSelection,
    );
    if (!mounted || selection == null) return;
    setState(() {
      _regionSelection = selection;
      _regionError = null;
    });
  }

  Widget _buildTagSection() {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 14, 14, 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '地址标签',
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: _textPrimary,
            ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _tagOptions.map((tag) {
              final isSelected = _selectedTag == tag;
              return ChoiceChip(
                label: Text(tag),
                selected: isSelected,
                selectedColor: const Color(0xFFFFEDF2),
                checkmarkColor: _themeColor,
                labelStyle: TextStyle(
                  color: isSelected ? _themeColor : _textPrimary,
                  fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                ),
                side: BorderSide(
                  color: isSelected
                      ? const Color(0xFFFFB3C5)
                      : const Color(0xFFE4E7ED),
                ),
                onSelected: (_) => setState(() => _selectedTag = tag),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildDefaultSwitch() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
      ),
      child: SwitchListTile(
        value: _isDefault,
        activeThumbColor: _themeColor,
        contentPadding: EdgeInsets.zero,
        title: const Text(
          '设为默认地址',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w700,
            color: _textPrimary,
          ),
        ),
        subtitle: const Text(
          '下单时优先使用该地址',
          style: TextStyle(fontSize: 13, color: _textHint),
        ),
        onChanged: (value) => setState(() => _isDefault = value),
      ),
    );
  }

  Widget _buildSubmitButton() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
      child: SizedBox(
        width: double.infinity,
        height: 52,
        child: ElevatedButton(
          onPressed: _isSubmitting ? null : _submitForm,
          style: ElevatedButton.styleFrom(
            backgroundColor: _themeColor,
            foregroundColor: Colors.white,
            disabledBackgroundColor: _themeColor.withAlpha(120),
            elevation: 0,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
            textStyle: const TextStyle(
              fontSize: 17,
              fontWeight: FontWeight.w700,
            ),
          ),
          child: _isSubmitting
              ? const SizedBox(
                  width: 22,
                  height: 22,
                  child: CircularProgressIndicator(
                    strokeWidth: 2.2,
                    color: Colors.white,
                  ),
                )
              : Text(_isEditMode ? '保存地址' : '新增地址'),
        ),
      ),
    );
  }

  String? _required(String? value, String message) {
    return (value?.trim() ?? '').isEmpty ? message : null;
  }
}
