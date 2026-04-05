import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

import '../../../model/hisotry_list.dart';
import '../../category/product/product_detail.dart';

///
/// 我的足迹页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class History extends StatefulWidget {
  const History({super.key});

  @override
  State<History> createState() => _HistoryState();
}

class _HistoryState extends State<History> {
  List<HistoryListDataItem> historyListDataItem = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    queryHistoryList();
  }

  Future<void> queryHistoryList() async {
    try {
      setState(() {
        _isLoading = true;
        _errorMessage = null;
      });
      Response result = await HttpUtil.get(historyListDataUrl);
      HistoryListModel historyListModel = HistoryListModel.fromJson(result.data);
      if (!mounted) return;
      setState(() {
        historyListDataItem = historyListModel.data;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        historyListDataItem = [];
        _isLoading = false;
        _errorMessage = "加载失败，请重试";
      });
    }
  }

  String _formatTime(String timeStr) {
    try {
      final dt = DateTime.parse(timeStr);
      return "${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')} ${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}";
    } catch (e) {
      return timeStr;
    }
  }

  Future<void> _clearHistory() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text("清空足迹"),
        content: const Text("确定要清空全部浏览足迹吗？此操作不可恢复。"),
        actions: [
          TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text("取消")),
          TextButton(onPressed: () => Navigator.of(context).pop(true), child: const Text("确定")),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    try {
      setState(() { _isLoading = true; });
      await HttpUtil.get(clearReadHistoryDataUrl);
      if (!mounted) return;
      queryHistoryList();
    } catch (e) {
      if (!mounted) return;
      setState(() { _isLoading = false; });
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("清空足迹失败，请重试")));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("我的足迹"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        actions: [
          if (historyListDataItem.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.delete_sweep, color: Colors.grey),
              onPressed: _clearHistory,
              tooltip: "清空足迹",
            ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 48, color: Colors.grey),
            const SizedBox(height: 16),
            Text(_errorMessage!, style: const TextStyle(fontSize: 14, color: Colors.grey)),
            const SizedBox(height: 16),
            ElevatedButton(onPressed: queryHistoryList, child: const Text("重试")),
          ],
        ),
      );
    }
    if (historyListDataItem.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.history, size: 48, color: Colors.grey),
            const SizedBox(height: 16),
            const Text("暂无足迹", style: TextStyle(fontSize: 14, color: Colors.grey)),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text("去逛逛"),
            ),
          ],
        ),
      );
    }
    return Container(
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
      width: MediaQuery.of(context).size.width,
      child: Container(
        color: Colors.white,
        margin: const EdgeInsets.only(top: 5),
        padding: const EdgeInsets.all(15),
        child: ListView.builder(
          itemCount: historyListDataItem.length,
          itemBuilder: (BuildContext context, int index) {
            HistoryListDataItem item = historyListDataItem[index];
            return Dismissible(
              key: Key(item.id),
              direction: DismissDirection.endToStart,
              background: Container(
                alignment: Alignment.centerRight,
                padding: const EdgeInsets.only(right: 20),
                color: Colors.red,
                child: const Icon(Icons.delete, color: Colors.white),
              ),
              confirmDismiss: (direction) async {
                final confirmed = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text("删除足迹"),
                    content: Text("确定要删除「${item.productName}」的浏览记录吗？"),
                    actions: [
                      TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text("取消")),
                      TextButton(onPressed: () => Navigator.of(context).pop(true), child: const Text("确定")),
                    ],
                  ),
                );
                if (confirmed != true) return false;
                try {
                  await HttpUtil.get("$deleteReadHistoryDataUrl?ids=${item.id}");
                  return true;
                } catch (e) {
                  if (!mounted) return false;
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("删除足迹失败，请重试")));
                  return false;
                }
              },
              onDismissed: (direction) {
                queryHistoryList();
              },
              child: InkWell(
                onTap: () {
                  Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (context) => ProductDetail(productId: item.productId),
                    ),
                  );
                },
                child: Container(
                  margin: const EdgeInsets.only(bottom: 20),
                  child: Row(
                    children: [
                      CachedImageWidget(103, 125, item.productPic, fit: BoxFit.fill),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(item.productName,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: TextStyle(
                                    fontSize: 16, color: Color(int.parse('303133', radix: 16)).withAlpha(255))),
                            const SizedBox(height: 6),
                            Text(item.productSubTitle,
                                maxLines: 2,
                                style: TextStyle(
                                    fontSize: 12, color: Color(int.parse('707070', radix: 16)).withAlpha(255))),
                            const SizedBox(height: 6),
                            Row(
                              children: [
                                Expanded(
                                  child: Text("￥${item.productPrice}",
                                      style: TextStyle(
                                          fontSize: 16,
                                          color: Color(int.parse('fa436a', radix: 16)).withAlpha(255))),
                                ),
                                Text(_formatTime(item.createTime),
                                    style: TextStyle(
                                        fontSize: 12,
                                        color: Color(int.parse('303133', radix: 16)).withAlpha(255))),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }
}
