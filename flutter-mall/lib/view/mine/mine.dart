import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/view/mine/collection/collection.dart';
import 'package:flutter_mall/view/mine/coupon/coupon_list.dart';
import 'package:flutter_mall/view/mine/focus/focus.dart';
import 'package:flutter_mall/view/mine/history/history.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/view/mine/message/message.dart';
import 'package:flutter_mall/view/mine/order/order_list.dart';
import 'package:flutter_mall/view/mine/ping_jia/ping_jia.dart';
import 'package:flutter_mall/view/mine/profile/profile_edit.dart';
import 'package:flutter_mall/view/mine/setting/settings.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../config/constant_param.dart';
import '../../config/service_url.dart';
import '../../model/member_info.dart';
import 'address/address_list.dart';

///
/// 我的页面
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class Mine extends StatefulWidget {
  const Mine({super.key});

  @override
  State<Mine> createState() => _MineState();
}

class _MineState extends State<Mine> {
  bool _isLoggedIn = false;
  MemberInfoData? _memberInfoData;

  @override
  void initState() {
    super.initState();
    _checkLoginAndLoadData();
  }

  // 检查登录状态，已登录才请求会员信息
  Future<void> _checkLoginAndLoadData() async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    String? savedToken = prefs.getString(token);
    bool loggedIn = savedToken != null && savedToken.isNotEmpty;
    if (mounted) {
      setState(() {
        _isLoggedIn = loggedIn;
      });
    }
    if (loggedIn) {
      await _queryMemberInfo();
    }
  }

  // 请求用户个人信息
  Future<void> _queryMemberInfo() async {
    try {
      Response result = await HttpUtil.get(memberInfoDataUrl);
      MemberInfoModel memberInfoModel = MemberInfoModel.fromJson(result.data);
      if (mounted) {
        setState(() {
          _memberInfoData = memberInfoModel.data;
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: const Text('加载会员信息失败，请重试'),
            action: SnackBarAction(
              label: '重试',
              onPressed: _queryMemberInfo,
            ),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
        width: MediaQuery.of(context).size.width,
        child: CustomScrollView(
          shrinkWrap: true,
          slivers: [buildHeader(), buildIntegral(), buildOrderInfo(), buildSetting()],
        ),
      ),
    );
  }

  SliverPadding buildHeader() {
    return SliverPadding(
        padding: const EdgeInsets.all(0),
        sliver: SliverList(
          delegate: SliverChildListDelegate(<Widget>[
            Stack(
              children: [
                SizedBox(
                  width: MediaQuery.of(context).size.width,
                  child: Image.asset(
                    "images/user-bg.jpg",
                    height: 200,
                    // width: MediaQuery.of(context).size.width,
                    fit: BoxFit.cover,
                  ),
                ),
                Container(
                  height: 200,
                  margin: const EdgeInsets.symmetric(horizontal: 15),
                  child: Row(
                    children: [
                      InkWell(
                        onTap: () async {
                          if (_isLoggedIn) {
                            final result = await Navigator.of(context).push<bool>(
                              MaterialPageRoute(
                                builder: (context) => const ProfileEdit(),
                              ),
                            );
                            if (result == true) {
                              _queryMemberInfo();
                            }
                          } else {
                            await Navigator.of(context).push(
                              MaterialPageRoute(
                                builder: (context) => const Login(),
                              ),
                            );
                            await _checkLoginAndLoadData();
                          }
                        },
                        child: ClipOval(
                          child: _isLoggedIn && _memberInfoData != null && _memberInfoData!.avatar.isNotEmpty
                              ? Image.network(
                                  proxyImageUrl(_memberInfoData!.avatar),
                                  width: 70,
                                  height: 70,
                                  fit: BoxFit.cover,
                                  errorBuilder: (context, error, stackTrace) => Container(
                                    width: 70,
                                    height: 70,
                                    color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                                    child: const Icon(Icons.person, size: 45, color: Colors.white),
                                  ),
                                )
                              : Container(
                                  width: 70,
                                  height: 70,
                                  color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                                  child: const Icon(Icons.person, size: 45, color: Colors.white),
                                ),
                        ),
                      ),
                      const SizedBox(
                        width: 10,
                      ),
                      Text(
                          _isLoggedIn && _memberInfoData != null
                              ? _memberInfoData!.nickname
                              : '点击登录',
                          style: TextStyle(fontSize: 25, color: Color(int.parse('303133', radix: 16)).withAlpha(255))),
                    ],
                  ),
                ),
                Container(
                  margin: const EdgeInsets.all(20),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      InkWell(
                        onTap: () {
                          Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (context) => const Message(),
                            ),
                          );
                        },
                        child: Image.asset(
                          "images/message.png",
                          height: 30,
                          width: 30,
                        ),
                      ),
                      const SizedBox(
                        width: 15,
                      ),
                      InkWell(
                        onTap: () async {
                          await Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (context) => const Settings(),
                            ),
                          );
                          _checkLoginAndLoadData();
                        },
                        child: Image.asset(
                          "images/setting_white.png",
                          height: 30,
                          width: 30,
                        ),
                      ),
                    ],
                  ),
                )
              ],
            )
          ]),
        ));
  }

  SliverPadding buildIntegral() {
    var numStyle = TextStyle(fontSize: 16, color: Color(int.parse('303133', radix: 16)).withAlpha(255));
    var txtStyle = TextStyle(fontSize: 12, color: Color(int.parse('707070', radix: 16)).withAlpha(255));
    return SliverPadding(
        padding: const EdgeInsets.all(0),
        sliver: SliverList(
          delegate: SliverChildListDelegate(<Widget>[
            Container(
              decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(5)),
              height: 70,
              margin: const EdgeInsets.only(top: 8, left: 10, right: 10),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_memberInfoData?.points.toString() ?? '0', style: numStyle),
                      Text("积分", style: txtStyle),
                    ],
                  ),
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_memberInfoData?.growthPoint.toString() ?? '0', style: numStyle),
                      Text("成长值", style: txtStyle),
                    ],
                  ),
                  InkWell(
                    onTap: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => const CouponList(),
                        ),
                      );
                    },
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(_memberInfoData?.couponCount.toString() ?? '0', style: numStyle),
                        Text("优惠券", style: txtStyle),
                      ],
                    ),
                  ),
                ],
              ),
            )
          ]),
        ));
  }

  SliverPadding buildOrderInfo() {
    var style = TextStyle(fontSize: 12, color: Color(int.parse('303133', radix: 16)).withAlpha(255));
    return SliverPadding(
        padding: const EdgeInsets.all(0),
        sliver: SliverList(
          delegate: SliverChildListDelegate(<Widget>[
            InkWell(
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => const OrderList(),
                  ),
                );
              },
              child: Container(
                decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(5)),
                margin: const EdgeInsets.only(top: 8, left: 10, right: 10),
                height: 88,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  children: [
                    Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Image.asset(
                          "images/all_order.png",
                          height: 25,
                          width: 24,
                        ),
                        const SizedBox(
                          height: 5,
                        ),
                        Text("全部订单", style: style),
                      ],
                    ),
                    Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Image.asset(
                          "images/daifukuan.png",
                          height: 25,
                          width: 24,
                        ),
                        const SizedBox(
                          height: 5,
                        ),
                        Text("待付款", style: style),
                      ],
                    ),
                    Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Image.asset(
                          "images/daifahuo.png",
                          height: 25,
                          width: 24,
                        ),
                        const SizedBox(
                          height: 5,
                        ),
                        Text("待发货", style: style),
                      ],
                    ),
                    Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Image.asset(
                          "images/yifukuan.png",
                          height: 25,
                          width: 24,
                        ),
                        const SizedBox(
                          height: 5,
                        ),
                        Text("待收货", style: style),
                      ],
                    ),
                    Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Image.asset(
                          "images/tuihuo.png",
                          height: 25,
                          width: 24,
                        ),
                        const SizedBox(
                          height: 5,
                        ),
                        Text("退款/售后", style: style),
                      ],
                    )
                  ],
                ),
              ),
            ),
          ]),
        ));
  }

  SliverPadding buildSetting() {
    var list = ['地址管理', '我的足迹', '我的关注', '我的收藏', '我的评价', '设置'];
    var listImg = [
      'images/address.png',
      'images/history.png',
      'images/guanzhu.png',
      'images/shoucang.png',
      'images/pingjia.png',
      'images/setting.png'
    ];
    var click = [
      const AddressList(),
      const History(),
      const FocusOn(),
      const Collection(),
      const PinJia(),
      const Settings(),
    ];

    var border = BorderSide(width: 1, color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255));
    var boxDecoration = BoxDecoration(
      color: Colors.white,
      border: Border(bottom: border),
    );

    return SliverPadding(
        padding: const EdgeInsets.all(0),
        sliver: SliverList(
          delegate: SliverChildListDelegate(<Widget>[
            Container(
              margin: const EdgeInsets.only(top: 8, left: 10, right: 10),
              padding: const EdgeInsets.symmetric(horizontal: 10),
              decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(5)),
              height: 350,
              child: MediaQuery.removeViewPadding(
                  removeTop: true,
                  context: context,
                  child: ListView.builder(
                      itemCount: list.length,
                      itemBuilder: (BuildContext context, int index) {
                        return InkWell(
                          onTap: () async {
                            await Navigator.of(context).push(
                              MaterialPageRoute(
                                builder: (context) => click[index],
                              ),
                            );
                            _checkLoginAndLoadData();
                          },
                          child: Container(
                            decoration: boxDecoration,
                            padding: const EdgeInsets.symmetric(horizontal: 15),
                            height: 50,
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Row(
                                  children: [
                                    Expanded(
                                        child: Row(
                                      children: [
                                        Image.asset(
                                          listImg[index],
                                          height: 20,
                                          width: 20,
                                        ),
                                        const SizedBox(
                                          width: 10,
                                        ),
                                        Text(list[index],
                                            style: TextStyle(
                                                fontSize: 14,
                                                color: Color(int.parse('303133', radix: 16)).withAlpha(255))),
                                      ],
                                    )),
                                    Image.asset(
                                      "images/right_arrow.png",
                                      height: 15,
                                      width: 15,
                                    )
                                  ],
                                )
                              ],
                            ),
                          ),
                        );
                      })),
            )
          ]),
        ));
  }
}
