import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';

import '../../../config/constant_param.dart';
import '../../../model/login_model.dart';

///
/// 注册页面
///
/// 日期：2026/03/24
///
class Register extends StatefulWidget {
  const Register({super.key});

  @override
  State<Register> createState() => _RegisterState();
}

class _RegisterState extends State<Register> {
  static final RegExp _mobileRegExp = RegExp(r'^1[3-9]\d{9}$');

  final TextEditingController _mobileController = TextEditingController();
  final TextEditingController _nicknameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _confirmPasswordController =
      TextEditingController();

  bool _isLoading = false;
  String? _mobileError;
  String? _nicknameError;
  String? _passwordError;
  String? _confirmPasswordError;

  // 校验输入
  bool _validateInputs() {
    bool valid = true;
    final mobile = _mobileController.value.text.trim();
    final nickname = _nicknameController.value.text.trim();
    final password = _passwordController.value.text;
    final confirmPassword = _confirmPasswordController.value.text;

    setState(() {
      _mobileError = null;
      _nicknameError = null;
      _passwordError = null;
      _confirmPasswordError = null;
    });

    if (!_mobileRegExp.hasMatch(mobile)) {
      setState(() => _mobileError = "请输入正确的手机号");
      valid = false;
    }

    if (nickname.isEmpty) {
      setState(() => _nicknameError = "请输入昵称");
      valid = false;
    }

    if (password.length < 6) {
      setState(() => _passwordError = "密码长度不能少于6位");
      valid = false;
    }

    if (password != confirmPassword) {
      setState(() => _confirmPasswordError = "两次密码不一致");
      valid = false;
    }

    return valid;
  }

  void _submitRegisterData() async {
    if (!_validateInputs()) return;
    if (_isLoading) return;

    setState(() => _isLoading = true);

    try {
      Map<String, dynamic> registerMap = <String, dynamic>{};
      registerMap["mobile"] = _mobileController.value.text.trim();
      registerMap["nickname"] = _nicknameController.value.text.trim();
      registerMap["password"] = _passwordController.value.text;
      registerMap["confirmPassword"] = _confirmPasswordController.value.text;
      registerMap["source"] = 1;

      Response result =
          await HttpUtil.post(registerDataUrl, data: registerMap);

      LoginModel loginModel = LoginModel.fromJson(result.data);

      if (loginModel.code == 0) {
        // 注册成功，保存 token
        SharedPreferencesUtil.saveString(
            token, "${loginModel.data.tokenHead} ${loginModel.data.token}");

        if (!mounted) return;

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text("注册成功")),
        );

        // 注册成功后安全返回：先 pop 注册页，再尝试 pop 登录页
        Navigator.of(context).pop();
        if (Navigator.of(context).canPop()) {
          Navigator.of(context).pop(true);
        }
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(loginModel.message)),
        );
      }
    } on DioException catch (e) {
      if (!mounted) return;
      String msg = "注册失败，请稍后重试";
      if (e.response?.data is Map) {
        msg = (e.response?.data as Map)["message"] ?? msg;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg)),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("注册失败，请检查网络连接")),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text("注册"),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF1a1a2e),
        elevation: 0,
      ),
      body: Container(
        decoration: const BoxDecoration(
          color: Colors.white,
        ),
        child: SingleChildScrollView(
          child: Column(
            children: [
              const SizedBox(height: 40),
              Image.asset(
                "images/icon_main_logo.png",
                width: 60,
                height: 60,
                fit: BoxFit.contain,
              ),
              const SizedBox(height: 12),
              const Text(
                "注册九克城账号",
                style: TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF1a1a2e),
                ),
              ),
              const SizedBox(height: 28),
              Container(
                margin: const EdgeInsets.only(left: 50.0, right: 50),
                child: Column(
                  children: [
                    TextField(
                      controller: _mobileController,
                      keyboardType: TextInputType.phone,
                      maxLength: 11,
                      decoration: InputDecoration(
                        hintText: "手机号",
                        counterText: "",
                        errorText: _mobileError,
                      ),
                    ),
                    TextField(
                      controller: _nicknameController,
                      decoration: InputDecoration(
                        hintText: "昵称",
                        errorText: _nicknameError,
                      ),
                    ),
                    TextField(
                      controller: _passwordController,
                      obscureText: true,
                      decoration: InputDecoration(
                        hintText: "密码（至少6位）",
                        errorText: _passwordError,
                      ),
                    ),
                    TextField(
                      controller: _confirmPasswordController,
                      obscureText: true,
                      decoration: InputDecoration(
                        hintText: "确认密码",
                        errorText: _confirmPasswordError,
                      ),
                    ),
                    const SizedBox(height: 20),
                    InkWell(
                      onTap: _isLoading ? null : _submitRegisterData,
                      child: Container(
                        alignment: Alignment.center,
                        width: MediaQuery.of(context).size.width,
                        height: 50,
                        decoration: BoxDecoration(
                          color: _isLoading
                              ? Colors.grey
                              : Color(int.parse('fa436a', radix: 16))
                                  .withAlpha(255),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: _isLoading
                            ? const SizedBox(
                                width: 24,
                                height: 24,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: Colors.white,
                                ),
                              )
                            : const Text(
                                '注册',
                                style: TextStyle(color: Colors.white),
                              ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    GestureDetector(
                      onTap: () {
                        Navigator.of(context).pop();
                      },
                      child: const Text(
                        "已有账号？返回登录",
                        style: TextStyle(
                          color: Color(0xFFfa436a),
                          fontSize: 14,
                        ),
                      ),
                    ),
                  ],
                ),
              )
            ],
          ),
        ),
      ),
    );
  }
}
