import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/mine/login/register.dart';

import '../../../config/constant_param.dart';
import '../../../model/login_model.dart';

///
/// 登录页面
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class Login extends StatefulWidget {
  final String? redirectRoute;

  const Login({super.key, this.redirectRoute});

  @override
  State<Login> createState() => _LoginState();
}

class _LoginState extends State<Login> {
  static final RegExp _mobileRegExp = RegExp(r'^1[3-9]\d{9}$');

  //用户名文本控制器
  final TextEditingController _usernameController = TextEditingController();
  //密码文本控制器
  final TextEditingController _passwordController = TextEditingController();

  bool _isLoading = false;
  String? _mobileError;
  String? _passwordError;

  @override
  void initState() {
    super.initState();
    //初始化用户名和密码，便于测试
    _usernameController.text = "13800138001";
    _passwordController.text = "123456";
  }

  // 校验输入
  bool _validateInputs() {
    bool valid = true;
    final mobile = _usernameController.value.text.trim();
    final password = _passwordController.value.text;

    setState(() {
      _mobileError = null;
      _passwordError = null;
    });

    if (!_mobileRegExp.hasMatch(mobile)) {
      setState(() => _mobileError = "请输入正确的手机号");
      valid = false;
    }

    if (password.length < 6) {
      setState(() => _passwordError = "密码长度不能少于6位");
      valid = false;
    }

    return valid;
  }

  //提交登录
  void _submitLoginData() async {
    if (!_validateInputs()) return;
    if (_isLoading) return;

    setState(() => _isLoading = true);

    try {
      //创建一个映射，用于存储用户名和密码
      Map<String, dynamic> loginMap = <String, dynamic>{};
      loginMap["mobile"] = _usernameController.value.text.trim();
      loginMap["password"] = _passwordController.value.text;

      //通过HTTP POST请求提交登录数据
      Response result = await HttpUtil.post(loginDataUrl, data: loginMap);

      //将登录响应数据解析为LoginModel对象
      LoginModel loginModel = LoginModel.fromJson(result.data);

      if (loginModel.code == 0) {
        //保存登录凭证token
        SharedPreferencesUtil.saveString(
            token, "${loginModel.data.tokenHead} ${loginModel.data.token}");

        if (!mounted) return;

        // 登录成功后恢复原始路由或简单 pop
        if (widget.redirectRoute != null) {
          Navigator.of(context).pushReplacementNamed(widget.redirectRoute!);
        } else {
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
      String msg = "登录失败，请稍后重试";
      if (e.response?.data is Map) {
        msg = (e.response?.data as Map)["message"] ?? msg;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg)),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("登录失败，请检查网络连接")),
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
      body: Container(
        decoration: const BoxDecoration(
          color: Colors.white,
        ),
        child: Column(
          children: [
            const SizedBox(height: 150),
            Image.asset(
              "images/icon_main_logo.png",
              width: 80,
              height: 80,
              fit: BoxFit.contain,
            ),
            const SizedBox(height: 12),
            const Text(
              "九克城",
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
                color: Color(0xFF1a1a2e),
                letterSpacing: 2,
              ),
            ),
            const SizedBox(height: 6),
            const Text(
              "欢迎回来",
              style: TextStyle(color: Colors.grey, fontSize: 14),
            ),
            const SizedBox(height: 28),
            Container(
              margin: const EdgeInsets.only(left: 50.0, right: 50),
              child: Column(
                children: [
                  TextField(
                    controller: _usernameController,
                    keyboardType: TextInputType.phone,
                    maxLength: 11,
                    decoration: InputDecoration(
                      hintText: "手机号",
                      counterText: "",
                      errorText: _mobileError,
                    ),
                  ),
                  TextField(
                    controller: _passwordController,
                    obscureText: true,
                    decoration: InputDecoration(
                      hintText: "密码",
                      errorText: _passwordError,
                    ),
                  ),
                  const SizedBox(height: 20),
                  InkWell(
                    onTap: _isLoading ? null : _submitLoginData,
                    child: Container(
                      alignment: Alignment.center,
                      width: MediaQuery.of(context).size.width,
                      height: 50,
                      decoration: BoxDecoration(
                        color: _isLoading
                            ? Colors.grey
                            : Color(int.parse('fa436a', radix: 16)).withAlpha(255),
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
                              '登录',
                              style: TextStyle(color: Colors.white),
                            ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  GestureDetector(
                    onTap: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => const Register(),
                        ),
                      );
                    },
                    child: const Text(
                      "还没有账号？立即注册",
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
    );
  }
}
