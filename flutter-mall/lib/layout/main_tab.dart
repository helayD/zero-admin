import 'package:flutter/material.dart';

import 'bottom_navigation_bar.dart';

///
/// 底部导航
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
// 参考https://blog.csdn.net/sinat_41144773/article/details/129906589实现
class MainTab extends StatefulWidget {
  final int initialIndex;
  final String? intentSource;

  const MainTab({super.key, this.initialIndex = 0, this.intentSource});

  @override
  State<MainTab> createState() => _MainTabState();
}

class _MainTabState extends State<MainTab> {
  late int _bottomNavigationIndex; //底部导航的索引
  int? _intentSourceActiveIndex;

  @override
  void initState() {
    super.initState();
    _bottomNavigationIndex = widget.initialIndex;
    if ((widget.intentSource ?? '').trim().isNotEmpty) {
      _intentSourceActiveIndex = widget.initialIndex;
    }
  }

  @override
  Widget build(BuildContext context) {
    final pageIntentSource = _bottomNavigationIndex == _intentSourceActiveIndex
        ? widget.intentSource
        : null;
    return Scaffold(
        body: buildPages(
            intentSource: pageIntentSource)[_bottomNavigationIndex], //页面切换
        bottomNavigationBar: _bottomNavigationBar() //底部导航
        );
  }

  //底部导航-样式
  BottomNavigationBar _bottomNavigationBar() {
    return BottomNavigationBar(
      items: items(),
      //底部导航-图标和文字的定义，封装到函数里
      currentIndex: _bottomNavigationIndex,
      onTap: (flag) {
        setState(() {
          _bottomNavigationIndex = flag; //使用底部导航索引
          if (_intentSourceActiveIndex != null &&
              flag != _intentSourceActiveIndex) {
            _intentSourceActiveIndex = null;
          }
        });
      },
      //onTap 点击切换页面
    );
  }
}
