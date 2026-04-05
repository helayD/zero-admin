import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/view/home/brand/brand_detail.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

import '../../../model/attention_list.dart';

///
/// 我的关注页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class FocusOn extends StatefulWidget {
  const FocusOn({super.key});

  @override
  State<FocusOn> createState() => _FocusOnState();
}

class _FocusOnState extends State<FocusOn> {
  List<AttentionListData> attentionDataItem = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    queryFocusOnList();
  }

  Future<void> queryFocusOnList() async {
    try {
      setState(() {
        _isLoading = true;
        _errorMessage = null;
      });
      Response result = await HttpUtil.get(focusOnListDataUrl);
      AttentionListModel focusOnListModel = AttentionListModel.fromJson(result.data);
      if (!mounted) return;
      setState(() {
        attentionDataItem = focusOnListModel.data;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        attentionDataItem = [];
        _isLoading = false;
        _errorMessage = "加载失败，请重试";
      });
    }
  }

  Future<void> _clearAttention() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text("清空关注"),
        content: const Text("确定要清空全部关注吗？此操作不可恢复。"),
        actions: [
          TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text("取消")),
          TextButton(onPressed: () => Navigator.of(context).pop(true), child: const Text("确定")),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    try {
      setState(() { _isLoading = true; });
      await HttpUtil.get(clearAttentionDataUrl);
      if (!mounted) return;
      queryFocusOnList();
    } catch (e) {
      if (!mounted) return;
      setState(() { _isLoading = false; });
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("清空关注失败，请重试")));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("我的关注"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        actions: [
          if (attentionDataItem.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.delete_sweep, color: Colors.grey),
              onPressed: _clearAttention,
              tooltip: "清空关注",
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
            ElevatedButton(onPressed: queryFocusOnList, child: const Text("重试")),
          ],
        ),
      );
    }
    if (attentionDataItem.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.visibility_off, size: 48, color: Colors.grey),
            const SizedBox(height: 16),
            const Text("暂无关注", style: TextStyle(fontSize: 14, color: Colors.grey)),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text("去关注品牌"),
            ),
          ],
        ),
      );
    }
    return Container(
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
      width: MediaQuery.of(context).size.width,
      child: ListView.builder(
        itemCount: attentionDataItem.length,
        itemBuilder: (BuildContext context, int index) {
          AttentionListData item = attentionDataItem[index];
          return Dismissible(
            key: Key(item.brandId.toString()),
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
                  title: const Text("取消关注"),
                  content: Text("确定要取消关注「${item.brandName}」吗？"),
                  actions: [
                    TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text("取消")),
                    TextButton(onPressed: () => Navigator.of(context).pop(true), child: const Text("确定")),
                  ],
                ),
              );
              if (confirmed != true) return false;
              try {
                await HttpUtil.get("$deleteAttentionDataUrl?brandIds=${item.brandId}");
                return true;
              } catch (e) {
                if (!mounted) return false;
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("取消关注失败，请重试")));
                return false;
              }
            },
            onDismissed: (direction) {
              queryFocusOnList();
            },
            child: InkWell(
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => BrandDetail(brandId: item.brandId),
                  ),
                );
              },
              child: Container(
                color: Colors.white,
                margin: const EdgeInsets.only(top: 5),
                padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 10),
                child: Row(
                  children: [
                    CachedImageWidget(103, 85, item.brandLogo, fit: BoxFit.contain),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        item.brandName,
                        style: TextStyle(
                          fontSize: 16,
                          color: Color(int.parse('303133', radix: 16)).withAlpha(255),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}
