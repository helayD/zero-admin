import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/commerce_state_resolver.dart';
import 'package:flutter_mall/view/category/product/product_list.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/widgets/commerce_state_shell.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';

import '../../model/categories_model.dart';

///
/// 商品分类页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Categories extends StatefulWidget {
  const Categories({super.key});

  @override
  State<Categories> createState() => _CategoriesState();
}

// 商品分类页面
class _CategoriesState extends State<Categories> {
  // 商品一级分类
  List<CategoriesData>? firstCategoriesData = [];

  // 商品二级分类
  List<CategoriesData>? secondCategoriesData = [];

  // 默认选中的一级菜单项
  String selectedCategory = '';

  // 选择数据的索引
  int selectedIndex = 0;

  bool _isLoading = true;
  bool _hasError = false;
  Object? _pageError;

  @override
  void initState() {
    super.initState();
    _queryCategoriesData();
  }

  // 请求分类列表数据
  void _queryCategoriesData() async {
    setState(() {
      _isLoading = true;
      _hasError = false;
      _pageError = null;
    });
    try {
      Response result = await HttpUtil.get(categoriesDataUrl);
      CategoriesModel categoriesModel = CategoriesModel.fromJson(result.data);
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        firstCategoriesData = categoriesModel.data;
        if (categoriesModel.data.isNotEmpty) {
          secondCategoriesData = categoriesModel.data[0].children;
          selectedCategory = categoriesModel.data[0].name;
        }
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _hasError = true;
        _pageError = e;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("分类"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    final bool hasContent = firstCategoriesData != null && firstCategoriesData!.isNotEmpty;
    final CommercePageState pageState = CommerceStateResolver.resolvePageState(
      isLoading: _isLoading,
      hasContent: hasContent,
      isEmpty: !_isLoading && !_hasError && !hasContent,
      error: _pageError,
    );

    if (pageState != CommercePageState.content) {
      final failure = CommerceStateResolver.resolveFailure(
        _pageError,
        errorSummary: '分类数据加载失败',
        weakNetworkSummary: '当前网络较弱，分类数据暂时无法加载',
      );
      return CommerceStateShell(
        state: pageState,
        title: pageState == CommercePageState.empty ? '暂无商品分类' : '分类页暂不可用',
        summary: pageState == CommercePageState.empty
            ? '当前还没有可浏览的分类，稍后再来看看'
            : failure?.summary ?? '分类数据加载失败',
        detail: failure?.detail,
        icon: pageState == CommercePageState.empty ? Icons.category_outlined : null,
        primaryAction: pageState == CommercePageState.initialLoading
            ? null
            : CommerceStateAction(label: '重试', onPressed: _queryCategoriesData),
      );
    }

    if (firstCategoriesData == null || firstCategoriesData!.isEmpty) {
      return const EmptyStateWidget(
        message: '暂无商品分类',
        icon: Icons.category_outlined,
      );
    }
    return Container(
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
      child: Row(
        children: <Widget>[
          buildFirstCategories(),
          Expanded(
            child: Container(
              alignment: Alignment.topCenter,
              color: Colors.white,
              margin: const EdgeInsets.only(left: 6, top: 6),
              padding: const EdgeInsets.all(16.0),
              child: (secondCategoriesData != null && secondCategoriesData!.isNotEmpty)
                  ? buildSecondCategories()
                  : const EmptyStateWidget(
                      message: '该分类下暂无子分类',
                      icon: Icons.folder_open_outlined,
                    ),
            ),
          ),
        ],
      ),
    );
  }

  // 构建一级商品分类
  Container buildFirstCategories() {
    return Container(
      color: Colors.white,
      width: 100.0,
      child: ListView.builder(
          itemCount: firstCategoriesData!.length,
          itemBuilder: (BuildContext context, int index) {
            String name = firstCategoriesData![index].name;
            return InkWell(
              onTap: () {
                setState(() {
                  selectedCategory = name;
                  secondCategoriesData = firstCategoriesData![index].children;
                });
              },
              child: Stack(
                alignment: Alignment.centerLeft,
                children: [
                  Container(
                    alignment: Alignment.center,
                    padding: const EdgeInsets.all(15),
                    height: 50,
                    color: name == selectedCategory ? Color(int.parse('f5f5f5', radix: 16)).withAlpha(255) : null,
                    child: Text(
                      name,
                      style: TextStyle(
                          fontSize: 14,
                          color: name == selectedCategory
                              ? Color(int.parse('fa436a', radix: 16)).withAlpha(255)
                              : Color(int.parse('606266', radix: 16)).withAlpha(255)),
                    ),
                  ),
                  Visibility(
                    visible: name == selectedCategory,
                    child: Text(
                      "|",
                      style: TextStyle(
                        color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ],
              ),
            );
          }),
    );
  }

  // 根据选中的一级分类项构建对应的二级分类
  Widget buildSecondCategories() {
    return Wrap(
      spacing: 8.0,
      runSpacing: 8.0,
      children: secondCategoriesData!.map((item) {
        return GestureDetector(
          onTap: () {
            // 处理二级菜单项点击事件，跳转到商品详情页
            Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => ProductList(
                  productCategoryId: item.id,
                ),
              ),
            );
          },
          child: Column(
            children: [
              CachedImageWidget(
                70,
                70,
                item.imageUrl,
                fit: BoxFit.contain,
              ),
              Text(
                item.name,
                style: TextStyle(fontSize: 13, color: Color(int.parse('666666', radix: 16)).withAlpha(255)),
              )
            ],
          ),
        );
      }).toList(),
    );
  }
}
