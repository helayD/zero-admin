import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/view/mine/address/address_edit.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';

import '../../../model/address_list.dart';


///
/// 收货地址列表页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class AddressList extends StatefulWidget {
  const AddressList({super.key});

  @override
  State<AddressList> createState() => _AddressListState();
}

class _AddressListState extends State<AddressList> {
  List<AddressListData> addressListData = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    queryAddressList();
  }

  Future<void> queryAddressList() async {
    try {
      setState(() {
        _isLoading = true;
        _errorMessage = null;
      });
      Response result = await HttpUtil.get(addressListDataUrl);
      AddressListModel addressListModel =
          AddressListModel.fromJson(result.data);
      if (!mounted) return;
      // AC1: 默认地址优先排序
      final sortedData = addressListModel.data
        ..sort((a, b) => b.isDefault.compareTo(a.isDefault));
      setState(() {
        addressListData = sortedData;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _errorMessage = "加载失败，请重试";
      });
    }
  }

  Future<void> _deleteAddress(AddressListData data) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("确认删除"),
        content: Text("确定删除收货地址「${data.receiverName}」吗？"),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text("取消"),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text("删除", style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      Response result = await HttpUtil.get(
        deleteAddressDataUrl,
        queryParameters: {"ids": [data.id]},
      );
      if (!mounted) return;
      if (result.data["code"] == 0) {
        queryAddressList();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(result.data["message"] ?? "删除失败")),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("网络请求失败: $e")),
      );
    }
  }

  Future<void> _setDefault(AddressListData data) async {
    try {
      Response result = await HttpUtil.post(
        updateAddressStatusDataUrl,
        data: {"id": data.id, "isDefault": 1},
      );
      if (!mounted) return;
      if (result.data["code"] == 0) {
        queryAddressList();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(result.data["message"] ?? "设置失败")),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("网络请求失败: $e")),
      );
    }
  }

  void _navigateToEdit({AddressListData? addressData}) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (context) => AddressEdit(addressData: addressData),
      ),
    );
    if (result == true) {
      queryAddressList();
    }
  }

  @override
  Widget build(BuildContext context) {
    final themeColor =
        Color(int.parse('fa436a', radix: 16)).withAlpha(255);

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text(
          "收货地址",
          style: TextStyle(fontSize: 16, color: Colors.black),
        ),
        centerTitle: true,
      ),
      body: Container(
        color: Colors.white,
        child: Column(
          children: [
            Expanded(
              child: _buildBody(themeColor),
            ),
            InkWell(
              onTap: () => _navigateToEdit(),
              child: Container(
                alignment: Alignment.center,
                width: MediaQuery.of(context).size.width,
                height: 50,
                margin: const EdgeInsets.all(15),
                decoration: BoxDecoration(
                  color: themeColor,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: const Text(
                  '新增地址',
                  style: TextStyle(color: Colors.white),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBody(Color themeColor) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(_errorMessage!, style: const TextStyle(fontSize: 15)),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: queryAddressList,
              child: const Text("重试"),
            ),
          ],
        ),
      );
    }

    if (addressListData.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.location_off, size: 60, color: Colors.grey),
            SizedBox(height: 16),
            Text(
              "暂无收货地址，点击下方按钮新增",
              style: TextStyle(fontSize: 15, color: Colors.grey),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      itemCount: addressListData.length,
      itemBuilder: (BuildContext context, int index) {
        AddressListData data = addressListData[index];
        var defaultBorderDecoration = BoxDecoration(
          border: Border.all(width: 1, color: themeColor),
          borderRadius: const BorderRadius.all(Radius.circular(2)),
        );
        return Container(
          padding: const EdgeInsets.all(15),
          decoration: BoxDecoration(
            border: Border(
              top: BorderSide(
                width: 1,
                color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
              ),
            ),
          ),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        if (data.isDefault == 1)
                          Container(
                            margin: const EdgeInsets.only(right: 3),
                            decoration: defaultBorderDecoration,
                            child: Text(
                              "默认",
                              style: TextStyle(
                                fontSize: 12,
                                color: themeColor,
                              ),
                            ),
                          ),
                        if (data.isDefault != 1)
                          InkWell(
                            onTap: () => _setDefault(data),
                            child: Container(
                              margin: const EdgeInsets.only(right: 3),
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 4, vertical: 1),
                              decoration: BoxDecoration(
                                border: Border.all(
                                    width: 1, color: Colors.grey),
                                borderRadius: const BorderRadius.all(
                                    Radius.circular(2)),
                              ),
                              child: const Text(
                                "设为默认",
                                style: TextStyle(
                                    fontSize: 11, color: Colors.grey),
                              ),
                            ),
                          ),
                        Expanded(
                          child: Text(
                            "${data.province} ${data.city} ${data.district} ${data.detailAddress}",
                            style: TextStyle(
                              fontSize: 15,
                              color: Color(int.parse('303133', radix: 16))
                                  .withAlpha(255),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    Text(
                      "${data.receiverName} ${data.receiverPhone}",
                      style: TextStyle(
                        fontSize: 14,
                        color: Color(int.parse('909399', radix: 16))
                            .withAlpha(255),
                      ),
                    ),
                  ],
                ),
              ),
              Row(
                children: [
                  InkWell(
                    onTap: () => _navigateToEdit(addressData: data),
                    child: Image.asset(
                      "images/edit.png",
                      height: 20,
                      width: 22,
                    ),
                  ),
                  const SizedBox(width: 15),
                  InkWell(
                    onTap: () => _deleteAddress(data),
                    child: Image.asset(
                      "images/delete.png",
                      height: 20,
                      width: 22,
                    ),
                  ),
                ],
              ),
            ],
          ),
        );
      },
    );
  }
}
