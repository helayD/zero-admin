import 'package:flutter/material.dart';
import 'layout/app_bootstrap.dart';

///
/// 欢迎页面(启动页)
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Welcome extends StatefulWidget {
  const Welcome({super.key});

  @override
  State<Welcome> createState() => _WelcomeState();
}

class _WelcomeState extends State<Welcome> {
  @override
  Widget build(BuildContext context) {
    return const AppBootstrap();
  }
}
