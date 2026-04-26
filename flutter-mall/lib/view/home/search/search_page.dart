import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/categories_model.dart';
import 'package:flutter_mall/model/home_model.dart' as home;
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/view/category/product/product_list.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

class SearchPage extends StatefulWidget {
  const SearchPage({super.key});

  @override
  State<SearchPage> createState() => _SearchPageState();
}

class _SearchPageState extends State<SearchPage> {
  static const String _recentSearchKey = 'home_recent_search_keywords';
  static const int _recentSearchLimit = 8;

  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocusNode = FocusNode();

  List<String> _recentKeywords = [];
  List<String> _hotKeywords = [];
  List<CategoriesData> _hotCategories = [];
  bool _isHotKeywordLoading = true;
  bool _hasHotKeywordError = false;
  bool _isCategoryLoading = true;
  bool _hasCategoryError = false;

  @override
  void initState() {
    super.initState();
    _loadRecentKeywords();
    _queryHotKeywords();
    _queryHotCategories();
  }

  @override
  void dispose() {
    _searchController.dispose();
    _searchFocusNode.dispose();
    super.dispose();
  }

  void _loadRecentKeywords() {
    final String raw = SharedPreferencesUtil.getString(_recentSearchKey) ?? '';
    if (raw.isEmpty) {
      return;
    }
    try {
      final dynamic decoded = jsonDecode(raw);
      if (decoded is! List) {
        return;
      }
      setState(() {
        _recentKeywords = decoded
            .map((item) => item.toString().trim())
            .where((item) => item.isNotEmpty)
            .take(_recentSearchLimit)
            .toList();
      });
    } catch (_) {
      // 搜索历史损坏时忽略，避免影响进入搜索页。
    }
  }

  Future<void> _saveRecentKeyword(String keyword) async {
    final String normalized = keyword.trim();
    if (normalized.isEmpty) {
      return;
    }

    final List<String> nextKeywords = <String>[
      normalized,
      ..._recentKeywords.where((item) => item != normalized),
    ].take(_recentSearchLimit).toList();

    await SharedPreferencesUtil.saveString(
      _recentSearchKey,
      jsonEncode(nextKeywords),
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _recentKeywords = nextKeywords;
    });
  }

  Future<void> _clearRecentKeywords() async {
    await SharedPreferencesUtil.remove(_recentSearchKey);
    if (!mounted) {
      return;
    }
    setState(() {
      _recentKeywords = [];
    });
  }

  Future<void> _queryHotKeywords() async {
    setState(() {
      _isHotKeywordLoading = true;
      _hasHotKeywordError = false;
    });

    try {
      final Response result = await HttpUtil.get(homeDataUrl);
      final home.HomeModel homeModel = home.HomeModel.fromJson(result.data);
      final Set<String> seenKeywords = <String>{};
      final List<String> nextKeywords = <String>[];

      void appendKeyword(String? value) {
        final String normalized = value?.trim() ?? '';
        if (normalized.isEmpty || seenKeywords.contains(normalized)) {
          return;
        }
        seenKeywords.add(normalized);
        nextKeywords.add(normalized);
      }

      void appendProduct(home.ProductList product) {
        appendKeyword(product.brandName);
        appendKeyword(product.categoryName);
      }

      for (final home.ProductList product
          in homeModel.data.homeFlashPromotion.productList) {
        appendProduct(product);
      }
      for (final home.ProductList product in homeModel.data.hotProductList) {
        appendProduct(product);
      }
      for (final home.ProductList product in homeModel.data.newProductList) {
        appendProduct(product);
      }
      for (final brand in homeModel.data.brandList) {
        appendKeyword(brand.name);
      }
      for (final subject in homeModel.data.subjectList) {
        appendKeyword(subject.categoryName);
      }

      if (!mounted) {
        return;
      }
      setState(() {
        _hotKeywords = nextKeywords.take(10).toList();
        _isHotKeywordLoading = false;
        _hasHotKeywordError = false;
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isHotKeywordLoading = false;
        _hasHotKeywordError = true;
      });
    }
  }

  Future<void> _queryHotCategories() async {
    setState(() {
      _isCategoryLoading = true;
      _hasCategoryError = false;
    });

    try {
      final Response result = await HttpUtil.get(categoriesDataUrl);
      final CategoriesModel categoriesModel =
          CategoriesModel.fromJson(result.data);
      final List<CategoriesData> flattened = <CategoriesData>[];
      for (final CategoriesData category in categoriesModel.data) {
        final List<CategoriesData> children = category.children ?? [];
        if (children.isEmpty) {
          flattened.add(category);
          continue;
        }
        flattened.addAll(children);
      }
      if (!mounted) {
        return;
      }
      setState(() {
        _hotCategories = flattened.take(8).toList();
        _isCategoryLoading = false;
        _hasCategoryError = false;
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isCategoryLoading = false;
        _hasCategoryError = true;
      });
    }
  }

  Future<void> _openSearchResult(String keyword) async {
    final String normalized = keyword.trim();
    if (normalized.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请输入要搜索的商品关键词')),
      );
      return;
    }

    await _saveRecentKeyword(normalized);
    if (!mounted) {
      return;
    }
    FocusScope.of(context).unfocus();
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => ProductList(
          keyword: normalized,
          title: '“$normalized”搜索结果',
        ),
      ),
    );
  }

  void _openCategory(CategoriesData category) {
    FocusScope.of(context).unfocus();
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => ProductList(
          productCategoryId: category.id,
          title: category.name,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final MediaQueryData mediaQuery = MediaQuery.of(context);
    return MediaQuery(
      data: mediaQuery.copyWith(
        textScaler: mediaQuery.textScaler.clamp(maxScaleFactor: 1.18),
      ),
      child: Scaffold(
        backgroundColor: AppColors.background,
        body: SafeArea(
          child: Column(
            children: [
              _buildSearchBar(),
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.fromLTRB(
                    AppSpacing.lg,
                    AppSpacing.lg,
                    AppSpacing.lg,
                    AppSpacing.xxl,
                  ),
                  children: [
                    if (_recentKeywords.isNotEmpty) ...[
                      _buildRecentSearchSection(),
                      const SizedBox(height: AppSpacing.xl),
                    ],
                    _buildHotKeywordSection(),
                    const SizedBox(height: AppSpacing.xl),
                    _buildHotCategorySection(),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSearchBar() {
    final ThemeData theme = Theme.of(context);
    return Container(
      color: AppColors.surface,
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.sm,
        AppSpacing.sm,
        AppSpacing.lg,
        AppSpacing.md,
      ),
      child: Row(
        children: [
          IconButton(
            onPressed: () => Navigator.of(context).pop(),
            icon: const Icon(Icons.arrow_back_rounded),
            tooltip: '返回',
          ),
          Expanded(
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: AppColors.background,
                borderRadius: BorderRadius.circular(AppRadii.md),
                border: Border.all(color: AppColors.border),
              ),
              child: TextField(
                controller: _searchController,
                focusNode: _searchFocusNode,
                textInputAction: TextInputAction.search,
                onSubmitted: _openSearchResult,
                decoration: InputDecoration(
                  hintText: '搜索商品、品牌或分类',
                  border: InputBorder.none,
                  prefixIcon: const Icon(Icons.search_rounded),
                  suffixIcon: _searchController.text.trim().isEmpty
                      ? null
                      : IconButton(
                          onPressed: () {
                            setState(() {
                              _searchController.clear();
                            });
                          },
                          icon: const Icon(Icons.close_rounded),
                          tooltip: '清空',
                        ),
                  contentPadding: const EdgeInsets.symmetric(
                    vertical: AppSpacing.md,
                  ),
                ),
                onChanged: (_) => setState(() {}),
              ),
            ),
          ),
          const SizedBox(width: AppSpacing.sm),
          TextButton(
            onPressed: () => _openSearchResult(_searchController.text),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.accent,
              textStyle: theme.textTheme.labelLarge,
              minimumSize: const Size(52, 44),
            ),
            child: const Text('搜索'),
          ),
        ],
      ),
    );
  }

  Widget _buildRecentSearchSection() {
    return _SearchSection(
      title: '最近搜索',
      action: TextButton(
        onPressed: _clearRecentKeywords,
        style: TextButton.styleFrom(
          foregroundColor: AppColors.textHint,
          minimumSize: const Size(44, 36),
        ),
        child: const Text('清空'),
      ),
      child: _KeywordWrap(
        keywords: _recentKeywords,
        onTap: _openSearchResult,
      ),
    );
  }

  Widget _buildHotKeywordSection() {
    if (_isHotKeywordLoading) {
      return const _SearchSection(
        title: '热门搜索',
        child: Padding(
          padding: EdgeInsets.symmetric(vertical: AppSpacing.xl),
          child: Center(child: CircularProgressIndicator()),
        ),
      );
    }

    if (_hasHotKeywordError) {
      return _SearchSection(
        title: '热门搜索',
        child: _SearchRetryPanel(
          message: '热门搜索加载失败，请稍后重试',
          onRetry: _queryHotKeywords,
        ),
      );
    }

    if (_hotKeywords.isEmpty) {
      return const _SearchSection(
        title: '热门搜索',
        child: _SearchEmptyPanel(message: '暂无热门搜索'),
      );
    }

    return _SearchSection(
      title: '热门搜索',
      child: _KeywordWrap(
        keywords: _hotKeywords,
        onTap: _openSearchResult,
      ),
    );
  }

  Widget _buildHotCategorySection() {
    if (_isCategoryLoading) {
      return const _SearchSection(
        title: '热门分类',
        child: Padding(
          padding: EdgeInsets.symmetric(vertical: AppSpacing.xl),
          child: Center(child: CircularProgressIndicator()),
        ),
      );
    }

    if (_hasCategoryError) {
      return _SearchSection(
        title: '热门分类',
        child: _SearchRetryPanel(
          message: '热门分类加载失败，请稍后重试',
          onRetry: _queryHotCategories,
        ),
      );
    }

    if (_hotCategories.isEmpty) {
      return const _SearchSection(
        title: '热门分类',
        child: _SearchEmptyPanel(message: '暂无热门分类'),
      );
    }

    return _SearchSection(
      title: '热门分类',
      child: GridView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemCount: _hotCategories.length,
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          childAspectRatio: 2.85,
          crossAxisSpacing: AppSpacing.md,
          mainAxisSpacing: AppSpacing.md,
        ),
        itemBuilder: (context, index) {
          final CategoriesData category = _hotCategories[index];
          return _CategoryTile(
            category: category,
            onTap: () => _openCategory(category),
          );
        },
      ),
    );
  }
}

class _SearchSection extends StatelessWidget {
  final String title;
  final Widget child;
  final Widget? action;

  const _SearchSection({
    required this.title,
    required this.child,
    this.action,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              title,
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
            const Spacer(),
            if (action != null) action!,
          ],
        ),
        const SizedBox(height: AppSpacing.md),
        child,
      ],
    );
  }
}

class _KeywordWrap extends StatelessWidget {
  final List<String> keywords;
  final ValueChanged<String> onTap;

  const _KeywordWrap({
    required this.keywords,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Wrap(
      spacing: AppSpacing.sm,
      runSpacing: AppSpacing.sm,
      children: keywords.map((keyword) {
        return Material(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(999),
          child: InkWell(
            borderRadius: BorderRadius.circular(999),
            onTap: () => onTap(keyword),
            child: Container(
              constraints: const BoxConstraints(minHeight: 40),
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.lg,
                vertical: AppSpacing.sm,
              ),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(999),
                border: Border.all(
                  color: AppColors.border,
                ),
              ),
              child: Text(
                keyword,
                style: theme.textTheme.labelLarge?.copyWith(
                  color: AppColors.textPrimary,
                ),
              ),
            ),
          ),
        );
      }).toList(),
    );
  }
}

class _CategoryTile extends StatelessWidget {
  final CategoriesData category;
  final VoidCallback onTap;

  const _CategoryTile({
    required this.category,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Material(
      color: AppColors.surface,
      borderRadius: BorderRadius.circular(AppRadii.lg),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.all(AppSpacing.sm),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(AppRadii.lg),
            border: Border.all(color: AppColors.border),
          ),
          child: Row(
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(AppRadii.md),
                child: CachedImageWidget(
                  44,
                  44,
                  category.imageUrl,
                  fit: BoxFit.cover,
                ),
              ),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Text(
                  category.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.labelLarge?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              const Icon(
                Icons.chevron_right_rounded,
                color: AppColors.textHint,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SearchRetryPanel extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _SearchRetryPanel({
    required this.message,
    required this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.lg),
        border: Border.all(color: AppColors.border),
      ),
      child: Row(
        children: [
          const Icon(Icons.info_outline_rounded, color: AppColors.textHint),
          const SizedBox(width: AppSpacing.sm),
          Expanded(child: Text(message)),
          TextButton(onPressed: onRetry, child: const Text('重试')),
        ],
      ),
    );
  }
}

class _SearchEmptyPanel extends StatelessWidget {
  final String message;

  const _SearchEmptyPanel({required this.message});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.lg),
        border: Border.all(color: AppColors.border),
      ),
      child: Text(
        message,
        style: Theme.of(context).textTheme.bodyMedium,
      ),
    );
  }
}
