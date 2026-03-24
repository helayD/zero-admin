import 'package:flutter/material.dart';
import 'package:flutter_mall/view/mine/login/login.dart';

import 'layout/main_tab.dart';


///
/// 欢迎页面(启动页)
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class Welcome extends StatefulWidget {
  const Welcome({super.key});

  @override
  State<Welcome> createState() => _WelcomeState();
}

class _WelcomeState extends State<Welcome> {
  @override
  void initState() {
    super.initState();
    //延迟1秒执行
    Future.delayed(const Duration(seconds: 1), () {
      if (!mounted) return;
      //跳转至应用首页
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (context) => const MainTab(),
        ),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    return Container(
        color: Colors.white,
        alignment: Alignment.center,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Image.asset(
              "images/icon_main_logo.png",
              width: 100,
              height: 100,
              fit: BoxFit.contain,
            ),
            const SizedBox(height: 20),
            const Text(
              "九克城",
              style: TextStyle(
                fontSize: 32,
                fontWeight: FontWeight.bold,
                color: Color(0xFF1a1a2e),
                letterSpacing: 4,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              "品质生活，从这里开始",
              style: TextStyle(
                fontSize: 14,
                color: Color(0xFF909399),
              ),
            ),
          ],
        ));
  }
}
