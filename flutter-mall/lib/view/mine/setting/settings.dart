import 'package:flutter/material.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/view/mine/profile/profile_edit.dart';

///
/// 设置页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Settings extends StatefulWidget {
  const Settings({super.key});

  @override
  State<Settings> createState() => _SettingsState();
}

class _SettingsState extends State<Settings> {
  @override
  Widget build(BuildContext context) {
    var border = BorderSide(width: 1, color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255));
    var boxDecoration = BoxDecoration(color: Colors.white, border: Border(bottom: border));
    var list = ['个人资料', '收货地址', '实名认证', '消息推送', '清除缓存', '关于九克城', '检查更新'];
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("设置"),
        centerTitle: true,
      ),
      body: Container(
        color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
        // padding: const EdgeInsets.symmetric(horizontal: 15),
        width: MediaQuery.of(context).size.width,
        child: Column(
          children: [
            Container(
              color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
              height: 400,
              child: ListView.builder(
                  itemCount: list.length,
                  itemBuilder: (BuildContext context, int index) {
                    return InkWell(
                      onTap: () {
                        if (index == 0) {
                          Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (context) => const ProfileEdit(),
                            ),
                          );
                        }
                        if (index == 4) {
                          AppRecoveryStore.clearRecoveryState();
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('缓存已清理')),
                          );
                        }
                      },
                      child: Container(
                      height: 51,
                      padding: const EdgeInsets.all(15),
                      decoration: boxDecoration,
                      margin: index == 3 ? const EdgeInsets.symmetric(vertical: 5) : const EdgeInsets.all(0),
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(list[index],
                                style: TextStyle(
                                    fontSize: 15, color: Color(int.parse('303133', radix: 16)).withAlpha(255))),
                          ),
                          Visibility(
                            visible: index == 6,
                            child: Text("当前版本 1.0.0",
                                style: TextStyle(
                                    fontSize: 12, color: Color(int.parse('707070', radix: 16)).withAlpha(255))),
                          ),
                          Visibility(
                              visible: index != 3,
                              child: Image.asset(
                                "images/right_arrow.png",
                                height: 16,
                                width: 17,
                              )),
                          Visibility(
                            visible: index == 3,
                            child: Switch(
                              value: true,
                              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                              onChanged: (value) {},
                              activeThumbColor: Colors.white,
                              activeTrackColor: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                            ),
                          ),
                        ],
                      ),
                    ),
                    );
                  }),
            ),
            InkWell(
              onTap: () async {
                await AppRecoveryStore.clearAuthToken();
                await AppRecoveryStore.clearRecoveryState();
                if (context.mounted) {
                  Navigator.of(context).pop();
                }
              },
              child: Container(
                alignment: Alignment.center,
                width: MediaQuery.of(context).size.width,
                height: 50,
                margin: const EdgeInsets.only(top: 15),
                decoration: const BoxDecoration(
                  color: Colors.white,
                ),
                child: Text(
                  '退出登录',
                  style: TextStyle(color: Color(int.parse('fa436a', radix: 16)).withAlpha(255), fontSize: 15),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
