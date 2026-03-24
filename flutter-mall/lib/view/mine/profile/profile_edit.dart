import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:intl/intl.dart';

import '../../../config/service_url.dart';
import '../../../model/member_info.dart';

///
/// 个人资料编辑页
///
class ProfileEdit extends StatefulWidget {
  const ProfileEdit({super.key});

  @override
  State<ProfileEdit> createState() => _ProfileEditState();
}

class _ProfileEditState extends State<ProfileEdit> {
  final _formKey = GlobalKey<FormState>();
  final _nicknameController = TextEditingController();
  final _signatureController = TextEditingController();

  int _gender = 0; // 0-未知，1-男，2-女
  DateTime? _birthday;
  String _avatarUrl = '';
  String _mobile = '';
  int? _memberId; // info 接口返回的主键 id，用于 update

  bool _isLoading = true;
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    _loadMemberInfo();
  }

  @override
  void dispose() {
    _nicknameController.dispose();
    _signatureController.dispose();
    super.dispose();
  }

  Future<void> _loadMemberInfo() async {
    try {
      Response result = await HttpUtil.get(memberInfoDataUrl);
      MemberInfoModel model = MemberInfoModel.fromJson(result.data);
      MemberInfoData data = model.data;
      if (!mounted) return;
      setState(() {
        _memberId = data.id;
        _nicknameController.text = data.nickname;
        _signatureController.text = data.signature;
        _gender = data.gender;
        _avatarUrl = data.avatar;
        _mobile = data.mobile;
        if (data.birthday.isNotEmpty) {
          try {
            _birthday = DateFormat('yyyy-MM-dd').parse(data.birthday);
          } catch (_) {}
        }
        _isLoading = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: const Text('加载个人资料失败'),
            action: SnackBarAction(
              label: '重试',
              onPressed: () {
                setState(() {
                  _isLoading = true;
                });
                _loadMemberInfo();
              },
            ),
          ),
        );
      }
    }
  }

  Future<void> _saveMemberInfo() async {
    if (!_formKey.currentState!.validate()) return;
    if (_memberId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('资料未加载完成，请稍后重试')),
      );
      return;
    }

    setState(() {
      _isSaving = true;
    });

    try {
      Map<String, dynamic> data = {
        'id': _memberId,
        'nickname': _nicknameController.text.trim(),
        'mobile': _mobile,
        'avatar': _avatarUrl,
        'signature': _signatureController.text.trim(),
        'gender': _gender,
      };
      if (_birthday != null) {
        data['birthday'] = DateFormat('yyyy-MM-dd').format(_birthday!);
      }

      Response result = await HttpUtil.post(updateMemberDataUrl, data: data);

      if (mounted) {
        // updateMember 返回 code 为 string 类型 "000000" 表示成功
        final respData = result.data;
        final code = respData is Map ? respData['code'] : null;
        if (code != null && code.toString() != '000000') {
          final msg = respData['message'] ?? '保存失败';
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(msg.toString())),
          );
          return;
        }
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('保存成功')),
        );
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('保存失败，请重试')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isSaving = false;
        });
      }
    }
  }

  Future<void> _selectBirthday() async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: _birthday ?? DateTime(1990, 1, 1),
      firstDate: DateTime(1900),
      lastDate: now,
    );
    if (picked != null) {
      setState(() {
        _birthday = picked;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('个人资料'),
        centerTitle: true,
        actions: [
          TextButton(
            onPressed: _isSaving ? null : _saveMemberInfo,
            child: _isSaving
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Text(
                    '保存',
                    style: TextStyle(
                      color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                      fontSize: 16,
                    ),
                  ),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _buildForm(),
    );
  }

  Widget _buildForm() {
    var border = BorderSide(
      width: 1,
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
    );
    var boxDecoration = BoxDecoration(
      color: Colors.white,
      border: Border(bottom: border),
    );
    var labelStyle = TextStyle(
      fontSize: 15,
      color: Color(int.parse('303133', radix: 16)).withAlpha(255),
    );
    var valueStyle = TextStyle(
      fontSize: 15,
      color: Color(int.parse('606266', radix: 16)).withAlpha(255),
    );

    return Form(
      key: _formKey,
      child: ListView(
        children: [
          // 头像
          Container(
            decoration: boxDecoration,
            padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 12),
            child: Row(
              children: [
                Text('头像', style: labelStyle),
                const Spacer(),
                ClipOval(
                  child: _avatarUrl.isNotEmpty
                      ? Image.network(
                          proxyImageUrl(_avatarUrl),
                          width: 50,
                          height: 50,
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) =>
                              Container(
                            width: 50,
                            height: 50,
                            color: Color(int.parse('fa436a', radix: 16))
                                .withAlpha(255),
                            child: const Icon(Icons.person,
                                size: 30, color: Colors.white),
                          ),
                        )
                      : Container(
                          width: 50,
                          height: 50,
                          color: Color(int.parse('fa436a', radix: 16))
                              .withAlpha(255),
                          child: const Icon(Icons.person,
                              size: 30, color: Colors.white),
                        ),
                ),
                const SizedBox(width: 8),
                Image.asset("images/right_arrow.png", height: 15, width: 15),
              ],
            ),
          ),

          // 昵称
          Container(
            decoration: boxDecoration,
            padding: const EdgeInsets.symmetric(horizontal: 15),
            child: Row(
              children: [
                Text('昵称', style: labelStyle),
                const SizedBox(width: 20),
                Expanded(
                  child: TextFormField(
                    controller: _nicknameController,
                    maxLength: 20,
                    decoration: const InputDecoration(
                      border: InputBorder.none,
                      counterText: '',
                      hintText: '请输入昵称',
                    ),
                    textAlign: TextAlign.right,
                    style: valueStyle,
                    validator: (value) {
                      if (value == null || value.trim().isEmpty) {
                        return '昵称不能为空';
                      }
                      return null;
                    },
                  ),
                ),
              ],
            ),
          ),

          // 性别
          Container(
            decoration: boxDecoration,
            padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 12),
            child: Row(
              children: [
                Text('性别', style: labelStyle),
                const Spacer(),
                DropdownButton<int>(
                  value: _gender,
                  underline: const SizedBox(),
                  style: valueStyle,
                  items: const [
                    DropdownMenuItem(value: 0, child: Text('未知')),
                    DropdownMenuItem(value: 1, child: Text('男')),
                    DropdownMenuItem(value: 2, child: Text('女')),
                  ],
                  onChanged: (value) {
                    if (value != null) {
                      setState(() {
                        _gender = value;
                      });
                    }
                  },
                ),
              ],
            ),
          ),

          // 生日
          InkWell(
            onTap: _selectBirthday,
            child: Container(
              decoration: boxDecoration,
              padding:
                  const EdgeInsets.symmetric(horizontal: 15, vertical: 16),
              child: Row(
                children: [
                  Text('生日', style: labelStyle),
                  const Spacer(),
                  Text(
                    _birthday != null
                        ? DateFormat('yyyy-MM-dd').format(_birthday!)
                        : '请选择',
                    style: valueStyle,
                  ),
                  const SizedBox(width: 8),
                  Image.asset("images/right_arrow.png",
                      height: 15, width: 15),
                ],
              ),
            ),
          ),

          // 个性签名
          Container(
            decoration: boxDecoration,
            padding: const EdgeInsets.symmetric(horizontal: 15),
            child: Row(
              children: [
                Text('签名', style: labelStyle),
                const SizedBox(width: 20),
                Expanded(
                  child: TextFormField(
                    controller: _signatureController,
                    maxLength: 100,
                    decoration: const InputDecoration(
                      border: InputBorder.none,
                      counterText: '',
                      hintText: '请输入个性签名',
                    ),
                    textAlign: TextAlign.right,
                    style: valueStyle,
                  ),
                ),
              ],
            ),
          ),

          // 手机号（只读）
          Container(
            decoration: boxDecoration,
            padding:
                const EdgeInsets.symmetric(horizontal: 15, vertical: 16),
            child: Row(
              children: [
                Text('手机号', style: labelStyle),
                const Spacer(),
                Text(_mobile, style: valueStyle),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
