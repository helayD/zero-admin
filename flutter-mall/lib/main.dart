import 'package:flutter/material.dart';
import 'package:flutter_mall/provider/cart_model.dart';
import 'package:flutter_mall/provider/counter.dart';
import 'package:flutter_mall/provider/comment_provider.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/welcome.dart';
import 'package:provider/provider.dart';

import 'config/nav_key.dart';

///
/// 应用入口页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
// 程序的入口点
void main() async {
  // 确保Flutter的绑定被初始化，以便在主函数中使用Flutter的特性
  WidgetsFlutterBinding.ensureInitialized();
  // 初始化SharedPreferences
  await SharedPreferencesUtil.init();
  await AppVersionService.init();
  await const PermissionBroker().captureLostMediaOnLaunch();
  // 启动应用程序
  runApp(const MyApp());
}

// MyApp类是应用的根部件
class MyApp extends StatelessWidget {
  // 构造函数，使用super.key来初始化StatelessWidget的key属性
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    // 使用MultiProvider来管理应用的状态
    return MultiProvider(
      providers: [
        // 提供一个Counter实例
        ChangeNotifierProvider(create: (_) => Counter()),
        // 提供一个CartModel实例
        ChangeNotifierProvider(create: (_) => CartModel()),
        ChangeNotifierProvider(
          create: (_) => AppLifecycleProvider()..startObserving(),
        ),
        // 提供一个CommentProvider实例（Story 8-2 Review Fix R-4）
        ChangeNotifierProvider(create: (_) => CommentProvider()),
        ChangeNotifierProvider(create: (_) => CommentUploadProvider()),
      ],
      child: MaterialApp(
        // 设置全局的navigatorKey，以便在应用的任何地方进行导航
        navigatorKey: NavKey.navKey,
        title: '九克城',
        theme: AppTheme.lightTheme(),
        // 设置应用的首页为Welcome部件
        home: const Welcome(),
      ),
    );
  }
}
