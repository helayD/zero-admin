import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

import '../../../model/collection_list.dart';
import '../../category/product/product_detail.dart';

///
/// 我的收藏页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Collection extends StatefulWidget {
  const Collection({super.key});

  @override
  State<Collection> createState() => _CollectionState();
}

class _CollectionState extends State<Collection> {
  List<CollectionListData> collectionDataItem = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    queryCollectionList();
  }

  Future<void> queryCollectionList() async {
    try {
      setState(() {
        _isLoading = true;
        _errorMessage = null;
      });
      Response result = await HttpUtil.get(collectionListDataUrl);
      CollectionListModel collectionListModel =
          CollectionListModel.fromJson(result.data);
      if (!mounted) return;
      setState(() {
        collectionDataItem = collectionListModel.data;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        collectionDataItem = [];
        _isLoading = false;
        _errorMessage = "加载失败，请重试";
      });
    }
  }

  Future<void> _clearCollection() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text("清空收藏"),
        content: const Text("确定要清空全部收藏吗？此操作不可恢复。"),
        actions: [
          TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text("取消")),
          TextButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text("确定")),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    try {
      setState(() {
        _isLoading = true;
      });
      await HttpUtil.get(clearCollectionDataUrl);
      if (!mounted) return;
      queryCollectionList();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text("清空收藏失败，请重试")));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("我的收藏"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        actions: [
          if (collectionDataItem.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.delete_sweep, color: Colors.grey),
              onPressed: _clearCollection,
              tooltip: "清空收藏",
            ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: queryCollectionList,
        child: _buildBody(),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          SizedBox(
            height: MediaQuery.of(context).size.height * 0.65,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48, color: Colors.grey),
                const SizedBox(height: 16),
                Text(_errorMessage!,
                    style: const TextStyle(fontSize: 14, color: Colors.grey)),
                const SizedBox(height: 16),
                ElevatedButton(
                    onPressed: queryCollectionList, child: const Text("重试")),
              ],
            ),
          ),
        ],
      );
    }
    if (collectionDataItem.isEmpty) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          SizedBox(
            height: MediaQuery.of(context).size.height * 0.65,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.favorite_border, size: 48, color: Colors.grey),
                const SizedBox(height: 16),
                const Text("暂无收藏",
                    style: TextStyle(fontSize: 14, color: Colors.grey)),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text("去逛逛"),
                ),
              ],
            ),
          ),
        ],
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
          itemCount: collectionDataItem.length,
          itemBuilder: (BuildContext context, int index) {
            CollectionListData item = collectionDataItem[index];
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
                final messenger = ScaffoldMessenger.of(context);
                final confirmed = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text("取消收藏"),
                    content: Text("确定要取消收藏「${item.productName}」吗？"),
                    actions: [
                      TextButton(
                          onPressed: () => Navigator.of(context).pop(false),
                          child: const Text("取消")),
                      TextButton(
                          onPressed: () => Navigator.of(context).pop(true),
                          child: const Text("确定")),
                    ],
                  ),
                );
                if (confirmed != true) return false;
                try {
                  await HttpUtil.get("$deleteCollectionDataUrl?ids=${item.id}");
                  return true;
                } catch (e) {
                  if (!mounted) return false;
                  messenger.showSnackBar(
                      const SnackBar(content: Text("取消收藏失败，请重试")));
                  return false;
                }
              },
              onDismissed: (direction) {
                queryCollectionList();
              },
              child: Container(
                margin: const EdgeInsets.only(bottom: 12),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: Color(int.parse('eeeeee', radix: 16)).withAlpha(255),
                  ),
                ),
                child: InkWell(
                  borderRadius: BorderRadius.circular(12),
                  onTap: () async {
                    await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (context) =>
                            ProductDetail(productId: item.productId),
                      ),
                    );
                    if (!mounted) return;
                    queryCollectionList();
                  },
                  child: Container(
                    padding: const EdgeInsets.all(10),
                    child: Row(
                      children: [
                        CachedImageWidget(103, 125, item.productPic,
                            fit: BoxFit.fill),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(item.productName,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: TextStyle(
                                      fontSize: 16,
                                      color:
                                          Color(int.parse('303133', radix: 16))
                                              .withAlpha(255))),
                              const SizedBox(height: 6),
                              Text(item.productSubTitle,
                                  maxLines: 2,
                                  style: TextStyle(
                                      fontSize: 12,
                                      color:
                                          Color(int.parse('707070', radix: 16))
                                              .withAlpha(255))),
                              const SizedBox(height: 6),
                              Text("￥${item.productPrice}",
                                  style: TextStyle(
                                      fontSize: 16,
                                      color:
                                          Color(int.parse('fa436a', radix: 16))
                                              .withAlpha(255))),
                            ],
                          ),
                        ),
                      ],
                    ),
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
