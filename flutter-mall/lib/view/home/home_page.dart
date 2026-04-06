import 'package:card_swiper/card_swiper.dart';
import 'package:dio/dio.dart';
import 'package:easy_refresh/easy_refresh.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/home_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/commerce_state_resolver.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/home/brand/brand_detail.dart';
import 'package:flutter_mall/view/home/brand/brand_list.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/widgets/commerce_state_shell.dart';

import '../../config/service_url.dart';
import '../../model/brand_list.dart';
import '../category/product/product_detail.dart';

///
/// 首页
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class HomePage extends StatefulWidget {
  final String? intentSource;

  const HomePage({super.key, this.intentSource});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  static const int _initialGuessLikeCount = 4;
  static const int _guessLikeLoadStep = 2;
  static const int _flashVisibleCount = 4;
  static const int _newVisibleCount = 4;
  static const int _hotVisibleCount = 4;

  List<AdvertiseList> advertiseList = [];
  List<BrandListData> brandList = [];
  HomeFlashPromotion? homeFlashPromotion;
  List<ProductList> flashProductList = [];
  List<ProductList> newProductList = [];
  List<ProductList> hotProductList = [];
  List<PreferredAreaListData> preferredAreaList = [];

  int _count = _initialGuessLikeCount;
  late EasyRefreshController _controller;
  bool _isInitialLoading = true;
  bool _hasLoadedOnce = false;
  Object? _pageError;
  String? _contentBannerText;
  bool _contentBannerIsWeakNetwork = false;

  final List<_HomeShortcut> _shortcutItems = const [
    _HomeShortcut(
      label: '专题',
      message: '专题',
      icon: Icons.explore_outlined,
      accentColor: Color(0xFFF97316),
      backgroundColor: Color(0xFFFFEDD5),
    ),
    _HomeShortcut(
      label: '话题',
      message: '话题',
      icon: Icons.forum_outlined,
      accentColor: Color(0xFFEC4899),
      backgroundColor: Color(0xFFFCE7F3),
    ),
    _HomeShortcut(
      label: '优选',
      message: '优选',
      icon: Icons.auto_awesome_outlined,
      accentColor: Color(0xFF8B5CF6),
      backgroundColor: Color(0xFFF3E8FF),
    ),
    _HomeShortcut(
      label: '特惠',
      message: '特惠',
      icon: Icons.local_offer_outlined,
      accentColor: Color(0xFF65A30D),
      backgroundColor: Color(0xFFECFCCB),
    ),
  ];

  @override
  void initState() {
    super.initState();
    _controller = EasyRefreshController(
      controlFinishRefresh: true,
      controlFinishLoad: true,
    );
    _queryHomeData();
  }

  bool get _hasHomeContent {
    return advertiseList.isNotEmpty ||
        brandList.isNotEmpty ||
        flashProductList.isNotEmpty ||
        newProductList.isNotEmpty ||
        hotProductList.isNotEmpty ||
        preferredAreaList.isNotEmpty;
  }

  List<ProductList> get _featuredFlashProducts {
    if (flashProductList.length <= _flashVisibleCount) {
      return flashProductList;
    }
    return flashProductList.take(_flashVisibleCount).toList();
  }

  List<ProductList> get _featuredNewProducts {
    if (newProductList.length <= _newVisibleCount) {
      return newProductList;
    }
    return newProductList.take(_newVisibleCount).toList();
  }

  List<ProductList> get _featuredHotProducts {
    if (hotProductList.length <= _hotVisibleCount) {
      return hotProductList;
    }
    return hotProductList.take(_hotVisibleCount).toList();
  }

  List<ProductList> get _guessLikePool {
    return _buildGuessLikePool(
      flashProducts: _featuredFlashProducts,
      newProducts: _featuredNewProducts,
      hotProducts: hotProductList,
    );
  }

  int get _guessLikeVisibleCount {
    if (_guessLikePool.isEmpty) {
      return 0;
    }
    return _count > _guessLikePool.length ? _guessLikePool.length : _count;
  }

  List<ProductList> _buildGuessLikePool({
    required List<ProductList> flashProducts,
    required List<ProductList> newProducts,
    required List<ProductList> hotProducts,
  }) {
    final List<ProductList> featuredHotProducts =
        hotProducts.length <= _hotVisibleCount
            ? hotProducts
            : hotProducts.take(_hotVisibleCount).toList();
    final Set<int> excludedIds = <int>{
      ...flashProducts.map((product) => product.id),
      ...newProducts.map((product) => product.id),
      ...featuredHotProducts.map((product) => product.id),
    };
    final List<ProductList> pool = <ProductList>[];
    final Set<int> seenIds = <int>{};

    void appendProducts(List<ProductList> products) {
      for (final ProductList product in products) {
        if (excludedIds.contains(product.id) || seenIds.contains(product.id)) {
          continue;
        }
        seenIds.add(product.id);
        pool.add(product);
      }
    }

    appendProducts(hotProducts);
    appendProducts(newProducts);
    appendProducts(flashProducts);
    return pool;
  }

  Future<void> _queryHomeData({bool isManualRefresh = false}) async {
    final bool hasContentBeforeRefresh = _hasHomeContent;
    if (mounted) {
      setState(() {
        if (!hasContentBeforeRefresh) {
          _isInitialLoading = true;
        }
        _pageError = null;
        _contentBannerText = null;
        _contentBannerIsWeakNetwork = false;
      });
    }

    try {
      final Response result = await HttpUtil.get(homeDataUrl);
      final HomeModel homeModel = HomeModel.fromJson(result.data);
      if (!mounted) {
        return;
      }

      final List<ProductList> loadedHotProducts = homeModel.data.hotProductList;
      final List<ProductList> loadedFlashProducts =
          homeModel.data.homeFlashPromotion.productList;
      final List<ProductList> loadedNewProducts = homeModel.data.newProductList;
      final List<ProductList> guessLikePool = _buildGuessLikePool(
        flashProducts: loadedFlashProducts.length <= _flashVisibleCount
            ? loadedFlashProducts
            : loadedFlashProducts.take(_flashVisibleCount).toList(),
        newProducts: loadedNewProducts.length <= _newVisibleCount
            ? loadedNewProducts
            : loadedNewProducts.take(_newVisibleCount).toList(),
        hotProducts: loadedHotProducts,
      );
      final int initialGuessCount =
          guessLikePool.length < _initialGuessLikeCount
              ? guessLikePool.length
              : _initialGuessLikeCount;

      setState(() {
        advertiseList = homeModel.data.advertiseList;
        brandList = homeModel.data.brandList;
        homeFlashPromotion = homeModel.data.homeFlashPromotion;
        flashProductList = loadedFlashProducts;
        newProductList = loadedNewProducts;
        hotProductList = loadedHotProducts;
        preferredAreaList = homeModel.data.preferredAreaList;
        _count = initialGuessCount;
        _isInitialLoading = false;
        _hasLoadedOnce = true;
        _pageError = null;
        _contentBannerText = null;
        _contentBannerIsWeakNetwork = false;
      });

      await AppRecoveryStore.saveRecentContext(
        AppRecentContext.create(
          targetType: AppRecentTargetType.home,
          tabIndex: 0,
          source: widget.intentSource ?? 'manual_open',
          requiresAuth: false,
          fallbackType: AppRecentTargetType.home,
          fallbackTabIndex: 0,
        ),
      );

      if (isManualRefresh && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('首页已刷新到最新状态')),
        );
      }
    } catch (e) {
      if (!mounted) {
        return;
      }

      final failure = CommerceStateResolver.resolveFailure(
        e,
        errorSummary: '首页加载失败，请重试',
        weakNetworkSummary: '首页加载超时或网络较弱，请检查网络后重试',
      );

      setState(() {
        _isInitialLoading = false;
        if (_hasHomeContent) {
          _contentBannerText = failure?.summary ?? '刷新失败，已保留当前内容';
          _contentBannerIsWeakNetwork = failure?.isWeakNetwork ?? false;
        } else {
          _pageError = e;
        }
      });
    }
  }

  Future<void> _handleLoadMore() async {
    if (_guessLikePool.isEmpty ||
        _guessLikeVisibleCount >= _guessLikePool.length) {
      _controller.finishLoad(IndicatorResult.noMore);
      return;
    }

    await Future.delayed(const Duration(milliseconds: 300));
    if (!mounted) {
      return;
    }

    setState(() {
      final int nextCount = _guessLikeVisibleCount + _guessLikeLoadStep;
      _count =
          nextCount > _guessLikePool.length ? _guessLikePool.length : nextCount;
    });

    _controller.finishLoad(
      _guessLikeVisibleCount >= _guessLikePool.length
          ? IndicatorResult.noMore
          : IndicatorResult.success,
    );
  }

  void _showFeatureInProgress(String label) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('$label 功能建设中，后续会接入完整流程')),
    );
  }

  void _openBrandList() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => const BrandList()),
    );
  }

  void _openBrandDetail(BrandListData brand) {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => BrandDetail(brandId: brand.id)),
    );
  }

  void _openProductDetail(ProductList product) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => ProductDetail(productId: product.id),
      ),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        bottom: false,
        child: _buildPageBody(),
      ),
    );
  }

  Widget _buildPageBody() {
    final CommercePageState pageState = CommerceStateResolver.resolvePageState(
      isLoading: _isInitialLoading,
      hasContent: _hasHomeContent,
      isEmpty: _hasLoadedOnce && !_hasHomeContent,
      error: _pageError,
    );

    if (pageState != CommercePageState.content) {
      final failure = CommerceStateResolver.resolveFailure(
        _pageError,
        errorSummary: '首页加载失败，请重试',
        weakNetworkSummary: '当前网络较弱，首页暂时无法加载',
      );
      return CommerceStateShell(
        state: pageState,
        title: pageState == CommercePageState.empty ? '首页暂时没有内容' : '首页暂不可用',
        summary: pageState == CommercePageState.empty
            ? '当前没有可展示的首页推荐内容，稍后再来看看'
            : failure?.summary ?? '首页加载失败，请稍后重试',
        detail: failure?.detail,
        primaryAction: pageState == CommercePageState.initialLoading
            ? null
            : CommerceStateAction(
                label: '重新加载',
                onPressed: () {
                  _queryHomeData(isManualRefresh: true);
                },
              ),
      );
    }

    return CommerceStateShell(
      state: CommercePageState.content,
      title: '首页',
      showWeakNetworkBanner: _contentBannerText != null,
      weakNetworkBannerText: _contentBannerText,
      weakNetworkBannerAction: CommerceStateAction(
        label: '重试',
        onPressed: () {
          _queryHomeData(isManualRefresh: true);
        },
      ),
      weakNetworkBannerIcon: _contentBannerIsWeakNetwork
          ? Icons.wifi_tethering_error_rounded
          : Icons.info_outline,
      weakNetworkBannerBackgroundColor: _contentBannerIsWeakNetwork
          ? const Color(0xFFFFF7E6)
          : const Color(0xFFF4F4F5),
      weakNetworkBannerForegroundColor: _contentBannerIsWeakNetwork
          ? const Color(0xFFD46B08)
          : AppColors.textSecondary,
      child: EasyRefresh(
        controller: _controller,
        onRefresh: () async {
          await _queryHomeData(isManualRefresh: true);
          _controller.finishRefresh();
          _controller.resetFooter();
        },
        onLoad: _handleLoadMore,
        child: CustomScrollView(
          physics: const BouncingScrollPhysics(
            parent: AlwaysScrollableScrollPhysics(),
          ),
          slivers: [
            _buildTopBar(),
            if (advertiseList.isNotEmpty) _buildBannerSection(),
            _buildShortcutSection(),
            if (preferredAreaList.isNotEmpty) _buildPreferredAreaSection(),
            if (brandList.isNotEmpty) _buildBrandSection(),
            if (flashProductList.isNotEmpty)
              _buildProductGridSection(
                title: '秒杀专区',
                subtitle: _flashPromotionSubtitle(),
                icon: Icons.flash_on_rounded,
                iconColor: AppColors.primaryDark,
                iconBackground: AppColors.primarySoft,
                products: _featuredFlashProducts,
                badgeLabel: '限时价',
              ),
            if (newProductList.isNotEmpty)
              _buildProductGridSection(
                title: '新鲜好物',
                subtitle: '为你挑选高颜值、高口碑的新鲜好物',
                icon: Icons.inventory_2_outlined,
                iconColor: AppColors.primaryDark,
                iconBackground: AppColors.primarySoft,
                products: _featuredNewProducts,
                badgeLabel: '新品',
              ),
            if (hotProductList.isNotEmpty)
              _buildProductGridSection(
                title: '人气推荐',
                subtitle: '口碑热卖商品，浏览和下单都更集中',
                icon: Icons.local_fire_department_outlined,
                iconColor: AppColors.primaryDark,
                iconBackground: AppColors.primarySoft,
                products: _featuredHotProducts,
                badgeLabel: '热卖',
              ),
            if (_guessLikePool.isNotEmpty)
              _buildProductGridSection(
                title: '猜你喜欢',
                subtitle: '根据热卖趋势展示的精选商品',
                icon: Icons.favorite_border_rounded,
                iconColor: AppColors.primaryDark,
                iconBackground: AppColors.primarySoft,
                products: _guessLikePool.take(_guessLikeVisibleCount).toList(),
                badgeLabel: '精选',
              ),
            const SliverToBoxAdapter(child: SizedBox(height: AppSpacing.xxl)),
          ],
        ),
      ),
    );
  }

  String _flashPromotionSubtitle() {
    final String nextStartTime = homeFlashPromotion?.nextStartTime.trim() ?? '';
    if (nextStartTime.isEmpty) {
      return '精选限时好价，先到先得';
    }
    return '下一场 $nextStartTime 开始';
  }

  String _bannerTitle(AdvertiseList advertise) {
    final String title = advertise.name.trim();
    if (title.isEmpty ||
        title.contains('acceptance') ||
        title.contains('_') ||
        title.length > 24) {
      return '首页精选活动';
    }
    return title;
  }

  String _bannerSubtitle(AdvertiseList advertise) {
    final String remark = advertise.remark.trim();
    if (remark.isNotEmpty &&
        !remark.contains('acceptance') &&
        remark.length <= 30) {
      return remark;
    }
    return '品牌好物与限时活动持续更新，逛一逛今天的推荐内容';
  }

  SliverToBoxAdapter _buildTopBar() {
    final ThemeData theme = Theme.of(context);
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.md,
          AppSpacing.lg,
          0,
        ),
        child: Container(
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [AppColors.primary, AppColors.primaryDark],
            ),
            borderRadius: BorderRadius.circular(AppRadii.xxl),
            boxShadow: const [
              BoxShadow(
                color: Color(0x29101828),
                blurRadius: 28,
                offset: Offset(0, 16),
              ),
            ],
          ),
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.lg,
            AppSpacing.lg,
            AppSpacing.lg,
            AppSpacing.xl,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  _TopIconButton(
                    icon: Icons.qr_code_scanner_rounded,
                    semanticLabel: '扫一扫',
                    onTap: () {
                      _showFeatureInProgress('扫一扫');
                    },
                  ),
                  const SizedBox(width: AppSpacing.md),
                  Expanded(
                    child: _SearchTrigger(
                      onTap: () {
                        _showFeatureInProgress('搜索');
                      },
                    ),
                  ),
                  const SizedBox(width: AppSpacing.md),
                  _TopIconButton(
                    icon: Icons.notifications_none_rounded,
                    semanticLabel: '消息',
                    onTap: () {
                      _showFeatureInProgress('消息');
                    },
                  ),
                ],
              ),
              const SizedBox(height: AppSpacing.xl),
              Text(
                '九克城',
                style: theme.textTheme.headlineSmall?.copyWith(
                  color: Colors.white,
                  fontSize: 26,
                ),
              ),
              const SizedBox(height: AppSpacing.sm),
              Text(
                '品牌直供、限时秒杀与精选好物都在这里',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: Colors.white.withValues(alpha: 0.92),
                ),
              ),
              const SizedBox(height: AppSpacing.lg),
              Wrap(
                spacing: AppSpacing.sm,
                runSpacing: AppSpacing.sm,
                children: const [
                  _HeaderTag(label: '品牌直供'),
                  _HeaderTag(label: '限时好价'),
                  _HeaderTag(label: '口碑精选'),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  SliverToBoxAdapter _buildBannerSection() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.xl,
          AppSpacing.lg,
          0,
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(AppRadii.xl),
          child: SizedBox(
            height: 190,
            child: Swiper(
              autoplay: true,
              itemCount: advertiseList.length,
              pagination: SwiperPagination(
                alignment: Alignment.bottomCenter,
                builder: DotSwiperPaginationBuilder(
                  color: Colors.white.withValues(alpha: 0.55),
                  activeColor: Colors.white,
                  size: 7,
                  activeSize: 8,
                ),
              ),
              itemBuilder: (BuildContext context, int index) {
                final AdvertiseList advertise = advertiseList[index];
                return _HomeBannerCard(
                  title: _bannerTitle(advertise),
                  subtitle: _bannerSubtitle(advertise),
                  imageUrl: advertise.pic,
                  onTap: () {
                    final String activityName = advertise.name.trim();
                    _showFeatureInProgress(
                      activityName.isEmpty ? '活动详情' : activityName,
                    );
                  },
                );
              },
              onTap: (int index) {
                final String activityName = advertiseList[index].name.trim();
                _showFeatureInProgress(
                  activityName.isEmpty ? '活动详情' : activityName,
                );
              },
            ),
          ),
        ),
      ),
    );
  }

  SliverToBoxAdapter _buildShortcutSection() {
    return _buildSectionShell(
      child: Row(
        children: _shortcutItems
            .map(
              (item) => Expanded(
                child: Padding(
                  padding:
                      const EdgeInsets.symmetric(horizontal: AppSpacing.xs),
                  child: _ShortcutButton(
                    item: item,
                    onTap: () {
                      _showFeatureInProgress(item.message);
                    },
                  ),
                ),
              ),
            )
            .toList(),
      ),
    );
  }

  SliverToBoxAdapter _buildPreferredAreaSection() {
    return _buildSectionShell(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _SectionHeader(
            icon: Icons.auto_awesome_rounded,
            iconColor: const Color(0xFF7C3AED),
            iconBackground: const Color(0xFFF3E8FF),
            title: '优选专区',
            subtitle: '聚合主题专场和精选活动，帮助用户快速找到想逛的内容',
          ),
          const SizedBox(height: AppSpacing.lg),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: preferredAreaList.length,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              childAspectRatio: 0.74,
              crossAxisSpacing: AppSpacing.md,
              mainAxisSpacing: AppSpacing.md,
            ),
            itemBuilder: (BuildContext context, int index) {
              final PreferredAreaListData item = preferredAreaList[index];
              return _PreferredAreaCard(
                item: item,
                onTap: () {
                  _showFeatureInProgress(
                      item.name.isEmpty ? '优选专区' : item.name);
                },
              );
            },
          ),
        ],
      ),
    );
  }

  SliverToBoxAdapter _buildBrandSection() {
    return _buildSectionShell(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _SectionHeader(
            icon: Icons.storefront_rounded,
            iconColor: const Color(0xFF2563EB),
            iconBackground: const Color(0xFFDBEAFE),
            title: '品牌制造商直供',
            subtitle: '精选品牌馆，突出货源可信与品牌背书',
            actionLabel: '查看全部',
            onActionTap: _openBrandList,
          ),
          const SizedBox(height: AppSpacing.lg),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: brandList.length,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              childAspectRatio: 0.82,
              crossAxisSpacing: AppSpacing.md,
              mainAxisSpacing: AppSpacing.md,
            ),
            itemBuilder: (BuildContext context, int index) {
              final BrandListData brand = brandList[index];
              return _BrandCard(
                brand: brand,
                onTap: () {
                  _openBrandDetail(brand);
                },
              );
            },
          ),
        ],
      ),
    );
  }

  SliverToBoxAdapter _buildProductGridSection({
    required String title,
    required String subtitle,
    required IconData icon,
    required Color iconColor,
    required Color iconBackground,
    required List<ProductList> products,
    String? badgeLabel,
  }) {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.xl,
          AppSpacing.lg,
          0,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _SectionHeader(
              icon: icon,
              iconColor: iconColor,
              iconBackground: iconBackground,
              title: title,
              subtitle: subtitle,
            ),
            const SizedBox(height: AppSpacing.lg),
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: products.length,
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 0.72,
                crossAxisSpacing: AppSpacing.md,
                mainAxisSpacing: AppSpacing.md,
              ),
              itemBuilder: (BuildContext context, int index) {
                final ProductList product = products[index];
                return _ProductGridCard(
                  product: product,
                  badgeLabel: badgeLabel,
                  onTap: () {
                    _openProductDetail(product);
                  },
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  SliverToBoxAdapter _buildSectionShell({
    required Widget child,
    Color backgroundColor = AppColors.surface,
    Color borderColor = AppColors.border,
  }) {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.xl,
          AppSpacing.lg,
          0,
        ),
        child: Container(
          decoration: BoxDecoration(
            color: backgroundColor,
            borderRadius: BorderRadius.circular(AppRadii.xl),
            border: Border.all(color: borderColor),
            boxShadow: const [
              BoxShadow(
                color: Color(0x12101828),
                blurRadius: 24,
                offset: Offset(0, 12),
              ),
            ],
          ),
          padding: const EdgeInsets.all(AppSpacing.lg),
          child: child,
        ),
      ),
    );
  }
}

class _HomeShortcut {
  final String label;
  final String message;
  final IconData icon;
  final Color accentColor;
  final Color backgroundColor;

  const _HomeShortcut({
    required this.label,
    required this.message,
    required this.icon,
    required this.accentColor,
    required this.backgroundColor,
  });
}

class _TopIconButton extends StatelessWidget {
  final IconData icon;
  final String semanticLabel;
  final VoidCallback onTap;

  const _TopIconButton({
    required this.icon,
    required this.semanticLabel,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: semanticLabel,
      child: Material(
        color: Colors.white.withValues(alpha: 0.18),
        borderRadius: BorderRadius.circular(AppRadii.md),
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.md),
          onTap: onTap,
          child: SizedBox(
            width: 44,
            height: 44,
            child: Icon(icon, color: Colors.white),
          ),
        ),
      ),
    );
  }
}

class _SearchTrigger extends StatelessWidget {
  final VoidCallback onTap;

  const _SearchTrigger({required this.onTap});

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Semantics(
      button: true,
      label: '搜索商品',
      child: Material(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppRadii.lg),
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.lg),
          onTap: onTap,
          child: SizedBox(
            height: 48,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
              child: Row(
                children: [
                  Icon(
                    Icons.search_rounded,
                    color: AppColors.textSecondary.withValues(alpha: 0.9),
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  Text(
                    '搜索商品，例如：手机',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _HeaderTag extends StatelessWidget {
  final String label;

  const _HeaderTag({required this.label});

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.sm,
      ),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.16),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: theme.textTheme.labelMedium?.copyWith(
          color: Colors.white,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _HomeBannerCard extends StatelessWidget {
  final String title;
  final String subtitle;
  final String imageUrl;
  final VoidCallback onTap;

  const _HomeBannerCard({
    required this.title,
    required this.subtitle,
    required this.imageUrl,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: Stack(
          fit: StackFit.expand,
          children: [
            CachedImageWidget(
              double.infinity,
              double.infinity,
              imageUrl,
              fallback: const _BannerFallbackVisual(),
            ),
            DecoratedBox(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Colors.black.withValues(alpha: 0.08),
                    Colors.black.withValues(alpha: 0.12),
                    Colors.black.withValues(alpha: 0.54),
                  ],
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: AppSpacing.md,
                      vertical: AppSpacing.sm,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.2),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Text(
                      '今日精选',
                      style: theme.textTheme.labelMedium?.copyWith(
                        color: Colors.white,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                  const Spacer(),
                  Text(
                    title,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.titleLarge?.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: AppSpacing.xs),
                  Text(
                    subtitle,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: Colors.white.withValues(alpha: 0.9),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.md),
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: AppSpacing.md,
                          vertical: AppSpacing.sm,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(999),
                        ),
                        child: Text(
                          '立即查看',
                          style: theme.textTheme.labelMedium?.copyWith(
                            color: AppColors.primaryDark,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ),
                      const SizedBox(width: AppSpacing.sm),
                      Icon(
                        Icons.arrow_forward_rounded,
                        color: Colors.white.withValues(alpha: 0.92),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _BannerFallbackVisual extends StatelessWidget {
  const _BannerFallbackVisual();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            AppColors.primary,
            AppColors.primaryDark,
            Color(0xFF7A1F43),
          ],
        ),
      ),
      child: Stack(
        children: [
          Positioned(
            top: -16,
            right: -12,
            child: Container(
              width: 96,
              height: 96,
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(32),
              ),
            ),
          ),
          Positioned(
            left: 22,
            bottom: 28,
            child: Container(
              width: 72,
              height: 72,
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(24),
              ),
            ),
          ),
          Positioned(
            right: 42,
            bottom: 24,
            child: Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                border: Border.all(
                  color: Colors.white.withValues(alpha: 0.18),
                  width: 1.5,
                ),
                borderRadius: BorderRadius.circular(36),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ShortcutButton extends StatelessWidget {
  final _HomeShortcut item;
  final VoidCallback onTap;

  const _ShortcutButton({
    required this.item,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Semantics(
      button: true,
      label: item.label,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.lg),
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
            child: Column(
              children: [
                Container(
                  width: 56,
                  height: 56,
                  decoration: BoxDecoration(
                    color: item.backgroundColor,
                    borderRadius: BorderRadius.circular(18),
                  ),
                  child: Icon(item.icon, color: item.accentColor, size: 28),
                ),
                const SizedBox(height: AppSpacing.sm),
                Text(
                  item.label,
                  style: theme.textTheme.labelLarge?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final IconData icon;
  final Color iconColor;
  final Color iconBackground;
  final String title;
  final String subtitle;
  final String? actionLabel;
  final VoidCallback? onActionTap;

  const _SectionHeader({
    required this.icon,
    required this.iconColor,
    required this.iconBackground,
    required this.title,
    required this.subtitle,
    this.actionLabel,
    this.onActionTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Row(
      children: [
        Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: iconBackground,
            borderRadius: BorderRadius.circular(AppRadii.md),
          ),
          child: Icon(icon, color: iconColor, size: 24),
        ),
        const SizedBox(width: AppSpacing.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: theme.textTheme.titleMedium),
              const SizedBox(height: AppSpacing.xs),
              Text(
                subtitle,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ),
        if (actionLabel != null && onActionTap != null)
          TextButton(
            onPressed: onActionTap,
            child: Text(actionLabel!),
          ),
      ],
    );
  }
}

class _PreferredAreaCard extends StatelessWidget {
  final PreferredAreaListData item;
  final VoidCallback onTap;

  const _PreferredAreaCard({
    required this.item,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Material(
      color: AppColors.surfaceMuted,
      borderRadius: BorderRadius.circular(AppRadii.lg),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(AppRadii.md),
                  child: CachedImageWidget(
                    double.infinity,
                    double.infinity,
                    item.pic,
                    fit: BoxFit.cover,
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.md),
              Text(
                item.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.titleSmall,
              ),
              const SizedBox(height: AppSpacing.xs),
              Text(
                item.subTitle.isEmpty ? '精选主题专场' : item.subTitle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _BrandCard extends StatelessWidget {
  final BrandListData brand;
  final VoidCallback onTap;

  const _BrandCard({
    required this.brand,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Material(
      color: AppColors.surfaceMuted,
      borderRadius: BorderRadius.circular(AppRadii.lg),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Container(
                  width: double.infinity,
                  decoration: BoxDecoration(
                    color: AppColors.surfaceMuted,
                    borderRadius: BorderRadius.circular(AppRadii.md),
                    border: Border.all(color: AppColors.border),
                  ),
                  clipBehavior: Clip.antiAlias,
                  child: CachedImageWidget(
                    double.infinity,
                    double.infinity,
                    brand.logo,
                    fit: BoxFit.cover,
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.md),
              Text(
                brand.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.titleSmall,
              ),
              const SizedBox(height: AppSpacing.xs),
              Text(
                '商品 ${brand.productCount} · 评价 ${brand.productCommentCount}',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ProductCardContainer extends StatelessWidget {
  final Widget child;
  final VoidCallback onTap;

  const _ProductCardContainer({
    required this.child,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.lg),
        border: Border.all(color: AppColors.border.withValues(alpha: 0.72)),
        boxShadow: const [
          BoxShadow(
            color: Color(0x08101828),
            blurRadius: 12,
            offset: Offset(0, 6),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.lg),
          onTap: onTap,
          child: child,
        ),
      ),
    );
  }
}

class _ProductImagePanel extends StatelessWidget {
  final String imageUrl;

  const _ProductImagePanel({
    required this.imageUrl,
  });

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(AppRadii.md),
      child: DecoratedBox(
        decoration: const BoxDecoration(
          color: AppColors.surfaceMuted,
        ),
        child: CachedImageWidget(
          double.infinity,
          double.infinity,
          imageUrl,
          fit: BoxFit.cover,
        ),
      ),
    );
  }
}

class _ProductPriceRow extends StatelessWidget {
  final String price;
  final String? badgeLabel;

  const _ProductPriceRow({
    required this.price,
    this.badgeLabel,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Row(
      children: [
        if (badgeLabel != null)
          Container(
            padding: const EdgeInsets.symmetric(
              horizontal: AppSpacing.sm,
              vertical: AppSpacing.xs,
            ),
            decoration: BoxDecoration(
              color: AppColors.primarySoft,
              borderRadius: BorderRadius.circular(999),
            ),
            child: Text(
              badgeLabel!,
              style: theme.textTheme.labelSmall?.copyWith(
                color: AppColors.primaryDark,
              ),
            ),
          ),
        if (badgeLabel != null) const SizedBox(width: AppSpacing.sm),
        Expanded(
          child: Text(
            '￥$price',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.titleSmall?.copyWith(
              color: AppColors.primary,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      ],
    );
  }
}

class _ProductGridCard extends StatelessWidget {
  final ProductList product;
  final String? badgeLabel;
  final VoidCallback onTap;

  const _ProductGridCard({
    required this.product,
    required this.onTap,
    this.badgeLabel,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return _ProductCardContainer(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: _ProductImagePanel(imageUrl: product.mainPic),
            ),
            const SizedBox(height: AppSpacing.md),
            Text(
              product.name,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.titleSmall,
            ),
            const SizedBox(height: AppSpacing.xs),
            Text(
              product.subTitle.isEmpty ? '精选热卖商品' : product.subTitle,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            _ProductPriceRow(
              price: product.price,
              badgeLabel: badgeLabel,
            ),
          ],
        ),
      ),
    );
  }
}
