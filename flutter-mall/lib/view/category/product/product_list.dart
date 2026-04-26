import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/category/product/product_detail.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';

import '../../../model/product_list.dart';

///
/// 商品列表页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class ProductList extends StatefulWidget {
  final int productCategoryId;
  final String keyword;
  final String title;
  final int? newStatus;
  final int? recommendStatus;

  const ProductList({
    super.key,
    this.productCategoryId = 0,
    this.keyword = '',
    this.title = '商品列表',
    this.newStatus,
    this.recommendStatus,
  });

  @override
  State<ProductList> createState() => _ProductListState();
}

class _ProductListState extends State<ProductList> {
  List<ProductListData> productDataItem = [];
  bool _isLoading = true;
  bool _hasError = false;

  @override
  void initState() {
    super.initState();
    queryProductListData();
  }

  void queryProductListData() async {
    setState(() {
      _isLoading = true;
      _hasError = false;
    });
    try {
      final Map<String, dynamic> queryParameters = <String, dynamic>{};
      if (widget.productCategoryId > 0) {
        queryParameters['productCategoryId'] = widget.productCategoryId;
      }
      final String keyword = widget.keyword.trim();
      if (keyword.isNotEmpty) {
        queryParameters['keyword'] = keyword;
      }
      if (widget.newStatus != null) {
        queryParameters['newStatus'] = widget.newStatus;
      }
      if (widget.recommendStatus != null) {
        queryParameters['recommendStatus'] = widget.recommendStatus;
      }

      Response result = await HttpUtil.get(
        productListQueryUrl,
        queryParameters: queryParameters,
      );
      ProductListModel collectionListModel =
          ProductListModel.fromJson(result.data);
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        productDataItem = collectionListModel.data;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _hasError = true;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          backgroundColor: Colors.white,
          title: Text(widget.title),
          titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
          centerTitle: true,
        ),
        body: _buildBody(context));
  }

  Widget _buildBody(BuildContext context) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_hasError) {
      return ErrorRetryWidget(
        message: '商品列表加载失败',
        onRetry: queryProductListData,
      );
    }
    if (productDataItem.isEmpty) {
      return EmptyStateWidget(
        message: widget.keyword.trim().isEmpty ? '当前分类下暂无商品' : '没有找到相关商品',
        icon: Icons.shopping_bag_outlined,
        actionText: widget.keyword.trim().isEmpty ? '浏览其他分类' : '返回搜索',
        onAction: () => Navigator.of(context).pop(),
      );
    }
    return Container(
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
      width: MediaQuery.of(context).size.width,
      child: Container(
        color: Colors.white,
        margin: const EdgeInsets.only(top: 5),
        padding: const EdgeInsets.all(15),
        child: GridView.builder(
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2, childAspectRatio: 0.58, crossAxisSpacing: 10),
          itemCount: productDataItem.length,
          itemBuilder: (BuildContext context, int index) {
            final item = productDataItem[index];
            return InkWell(
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => ProductDetail(productId: item.id),
                  ),
                );
              },
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Stack(
                    children: [
                      CachedImageWidget(
                        165,
                        165,
                        item.mainPic,
                        fit: BoxFit.fill,
                      ),
                      if (item.stock <= 0)
                        Positioned(
                          top: 0,
                          right: 0,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: Colors.grey.shade600,
                              borderRadius: const BorderRadius.only(
                                bottomLeft: Radius.circular(8),
                              ),
                            ),
                            child: const Text(
                              '缺货',
                              style:
                                  TextStyle(color: Colors.white, fontSize: 10),
                            ),
                          ),
                        )
                      else if (item.stock > 0 && item.stock <= item.lowStock)
                        Positioned(
                          top: 0,
                          right: 0,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: Color(int.parse('fa436a', radix: 16))
                                  .withAlpha(255),
                              borderRadius: const BorderRadius.only(
                                bottomLeft: Radius.circular(8),
                              ),
                            ),
                            child: Text(
                              '仅剩${item.stock}件',
                              style: const TextStyle(
                                  color: Colors.white, fontSize: 10),
                            ),
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(item.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                          fontSize: 16,
                          color: Color(int.parse('303133', radix: 16))
                              .withAlpha(255))),
                  const SizedBox(height: 5),
                  Text(item.subTitle,
                      maxLines: 2,
                      style: TextStyle(
                          fontSize: 12,
                          color: Color(int.parse('707070', radix: 16))
                              .withAlpha(255))),
                  const SizedBox(height: 5),
                  Row(
                    children: [
                      Expanded(
                        child: Text("￥${item.price}",
                            style: TextStyle(
                                fontSize: 16,
                                color: item.stock <= 0
                                    ? const Color(0xFF909399)
                                    : Color(int.parse('fa436a', radix: 16))
                                        .withAlpha(255))),
                      ),
                      Text("已售 ${item.sales}",
                          style: TextStyle(
                              fontSize: 12,
                              color: Color(int.parse('909399', radix: 16))
                                  .withAlpha(255))),
                    ],
                  )
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}
