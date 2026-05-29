import 'package:card_swiper/card_swiper.dart';
import 'package:dio/dio.dart';
import 'package:easy_refresh/easy_refresh.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/home_model.dart';
import 'package:flutter_mall/model/message_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/commerce_state_resolver.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/digital_card_scan_page.dart';
import 'package:flutter_mall/view/digital_card/draw_activity_page.dart';
import 'package:flutter_mall/view/home/brand/brand_detail.dart';
import 'package:flutter_mall/view/home/brand/brand_list.dart';
import 'package:flutter_mall/view/home/search/search_page.dart';
import 'package:flutter_mall/view/mine/coupon/available_coupon_list.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/view/mine/message/message.dart';
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

  final GlobalKey _flashSectionKey = GlobalKey();
  final GlobalKey _newProductSectionKey = GlobalKey();

  int _count = _initialGuessLikeCount;
  late EasyRefreshController _controller;
  bool _isInitialLoading = true;
  bool _hasLoadedOnce = false;
  Object? _pageError;
  String? _contentBannerText;
  bool _contentBannerIsWeakNetwork = false;
  int _unreadMessageCount = 0;
  bool _isOpeningMessageCenter = false;

  final List<_HomeShortcut> _shortcutItems = const [
    _HomeShortcut(
      target: _HomeShortcutTarget.brand,
      label: '品牌馆',
      icon: Icons.storefront_rounded,
      accentColor: Color(0xFF9A5C00),
      backgroundColor: Color(0xFFFFF4DB),
    ),
    _HomeShortcut(
      target: _HomeShortcutTarget.flash,
      label: '限时购',
      icon: Icons.bolt_rounded,
      accentColor: Color(0xFFC2410C),
      backgroundColor: Color(0xFFFFE7D6),
    ),
    _HomeShortcut(
      target: _HomeShortcutTarget.newProduct,
      label: '新品',
      icon: Icons.inventory_2_rounded,
      accentColor: Color(0xFF0F766E),
      backgroundColor: Color(0xFFE0F2F1),
    ),
    _HomeShortcut(
      target: _HomeShortcutTarget.memberBenefits,
      label: '会员权益',
      icon: Icons.workspace_premium_rounded,
      accentColor: Color(0xFF475569),
      backgroundColor: Color(0xFFEFF6FF),
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
    _queryUnreadMessageCount();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _showPendingDailyLoginPointsReward();
    });
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

  Future<void> _showPendingDailyLoginPointsReward() async {
    if (!AppRecoveryStore.hasValidToken()) {
      return;
    }
    final points = await AppRecoveryStore.consumeDailyLoginPointsReward();
    if (!mounted || points == null || points <= 0) {
      return;
    }
    HapticFeedback.lightImpact();
    await showDialog<void>(
      context: context,
      barrierDismissible: true,
      builder: (context) => _DailyLoginPointsRewardDialog(points: points),
    );
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

  /// 打开扫码页（Story 10.7 Task 10，AC7）：
  /// 识别提货卡分享二维码后跳转 [DigitalCardClaimPage] 走领取闭环；
  /// 非提货卡链接弹窗确认是否打开浏览器；非 URL 文本提示不可识别。
  void _openScanPage() {
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => const DigitalCardScanPage(),
      ),
    );
  }

  Future<void> _onBannerTap(AdvertiseList advertise) async {
    try {
      await HttpUtil.post(
        recordHomeAdvertiseClickUrl,
        data: <String, dynamic>{'advertiseId': advertise.id},
      );
    } catch (_) {
      // 点击埋点失败不阻塞真实跳转
    }

    if (!mounted) {
      return;
    }

    if (advertise.activityType == 'digital_card_draw') {
      final int activityId = advertise.activityId;
      if (activityId <= 0) {
        _showFeatureInProgress('抽卡活动');
        return;
      }
      await Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => DrawActivityPage(
            activityId: activityId,
            activityTitle: advertise.name.trim(),
            intentSource: 'home_banner',
          ),
        ),
      );
      return;
    }

    final String activityName = advertise.name.trim();
    _showFeatureInProgress(activityName.isEmpty ? '活动详情' : activityName);
  }

  void _openBrandList() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => const BrandList()),
    );
  }

  void _openSearchPage() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => const SearchPage()),
    );
  }

  Future<void> _queryUnreadMessageCount() async {
    if (!AppRecoveryStore.hasValidToken()) {
      if (mounted && _unreadMessageCount != 0) {
        setState(() {
          _unreadMessageCount = 0;
        });
      }
      return;
    }

    try {
      final Response result = await HttpUtil.get(
        unreadCountUrl,
        redirectOnUnauthorized: false,
      );
      final UnreadCountModel model = UnreadCountModel.fromJson(result.data);
      if (!mounted) {
        return;
      }
      setState(() {
        _unreadMessageCount = model.unreadCount;
      });
    } catch (_) {
      // 首页未读数失败时不阻塞首页内容展示。
    }
  }

  Future<void> _openMessageCenter() async {
    if (_isOpeningMessageCenter) {
      return;
    }

    setState(() {
      _isOpeningMessageCenter = true;
    });

    try {
      if (!AppRecoveryStore.hasValidToken()) {
        await Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => const Login()),
        );
        if (!mounted || !AppRecoveryStore.hasValidToken()) {
          return;
        }
      }

      await Navigator.of(context).push(
        MaterialPageRoute(builder: (context) => const Message()),
      );
      await _queryUnreadMessageCount();
    } finally {
      if (mounted) {
        setState(() {
          _isOpeningMessageCenter = false;
        });
      }
    }
  }

  Future<void> _openMemberBenefits() async {
    final AppRecentContext recoveryContext = AppRecentContext.create(
      targetType: AppRecentTargetType.couponCenter,
      source: 'home_member_benefits',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );
    await AppRecoveryStore.saveActiveIntentCandidate(recoveryContext);
    if (!mounted) {
      return;
    }

    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => const AvailableCouponList(
          intentSource: 'home_member_benefits',
        ),
      ),
    );
    await AppRecoveryStore.clearActiveIntentCandidateIfMatches(
      AppRecentTargetType.couponCenter,
    );
  }

  Future<void> _scrollToHomeSection({
    required GlobalKey key,
    required String sectionName,
  }) async {
    final BuildContext? sectionContext = key.currentContext;
    if (sectionContext == null) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('当前暂无$sectionName内容，请下拉刷新后再试')),
      );
      return;
    }

    await Scrollable.ensureVisible(
      sectionContext,
      duration: const Duration(milliseconds: 360),
      curve: Curves.easeOutCubic,
      alignment: 0.04,
    );
  }

  Future<void> _handleShortcutTap(_HomeShortcut item) async {
    switch (item.target) {
      case _HomeShortcutTarget.brand:
        _openBrandList();
        return;
      case _HomeShortcutTarget.flash:
        await _scrollToHomeSection(
          key: _flashSectionKey,
          sectionName: '限时好价',
        );
        return;
      case _HomeShortcutTarget.newProduct:
        await _scrollToHomeSection(
          key: _newProductSectionKey,
          sectionName: '新品首发',
        );
        return;
      case _HomeShortcutTarget.memberBenefits:
        await _openMemberBenefits();
        return;
    }
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
    final MediaQueryData mediaQuery = MediaQuery.of(context);
    return MediaQuery(
      data: mediaQuery.copyWith(
        textScaler: mediaQuery.textScaler.clamp(maxScaleFactor: 1.18),
      ),
      child: Scaffold(
        backgroundColor: AppColors.background,
        body: SafeArea(
          bottom: false,
          child: _buildPageBody(),
        ),
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
          await _queryUnreadMessageCount();
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
            if (flashProductList.isNotEmpty)
              _buildProductGridSection(
                sectionKey: _flashSectionKey,
                title: '限时好价',
                subtitle: _flashPromotionSubtitle(),
                icon: Icons.flash_on_rounded,
                iconColor: AppColors.price,
                iconBackground: const Color(0xFFFFE7D6),
                products: _featuredFlashProducts,
                badgeLabel: '限时价',
              ),
            if (newProductList.isNotEmpty)
              _buildProductGridSection(
                sectionKey: _newProductSectionKey,
                title: '新品首发',
                subtitle: '本周上新，精选更适合日常使用的好物',
                icon: Icons.inventory_2_outlined,
                iconColor: const Color(0xFF0F766E),
                iconBackground: const Color(0xFFE0F2F1),
                products: _featuredNewProducts,
                badgeLabel: '新品',
              ),
            if (hotProductList.isNotEmpty)
              _buildProductGridSection(
                title: '热卖榜单',
                subtitle: '近期浏览和下单更集中的口碑商品',
                icon: Icons.local_fire_department_outlined,
                iconColor: const Color(0xFF9A5C00),
                iconBackground: AppColors.accentSoft,
                products: _featuredHotProducts,
                badgeLabel: '热卖',
              ),
            if (brandList.isNotEmpty) _buildBrandSection(),
            if (preferredAreaList.isNotEmpty) _buildPreferredAreaSection(),
            if (_guessLikePool.isNotEmpty)
              _buildProductGridSection(
                title: '猜你喜欢',
                subtitle: '根据热卖趋势展示的精选商品',
                icon: Icons.favorite_border_rounded,
                iconColor: const Color(0xFF334155),
                iconBackground: const Color(0xFFEFF6FF),
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
      child: Container(
        decoration: BoxDecoration(
          color: AppColors.surface,
          border: Border(
            bottom: BorderSide(
              color: AppColors.border.withValues(alpha: 0.72),
            ),
          ),
        ),
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.sm,
          AppSpacing.lg,
          AppSpacing.md,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '九克城',
                        style: theme.textTheme.titleLarge?.copyWith(
                          color: AppColors.textPrimary,
                          fontSize: 22,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: AppSpacing.xs),
                      Text(
                        '品牌直供 · 限时好价 · 精选好物',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ),
                ),
                _TopIconButton(
                  icon: Icons.qr_code_scanner_rounded,
                  semanticLabel: '扫一扫',
                  onTap: _openScanPage,
                ),
                const SizedBox(width: AppSpacing.sm),
                _TopIconButton(
                  icon: Icons.notifications_none_rounded,
                  semanticLabel: '消息',
                  badgeCount: _unreadMessageCount,
                  onTap: _openMessageCenter,
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.md),
            _SearchTrigger(
              onTap: _openSearchPage,
            ),
            const SizedBox(height: AppSpacing.md),
            Wrap(
              spacing: AppSpacing.sm,
              runSpacing: AppSpacing.sm,
              children: const [
                _HeaderTag(label: '官方甄选'),
                _HeaderTag(label: '正品保障'),
                _HeaderTag(label: '好价上新'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  SliverToBoxAdapter _buildBannerSection() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.md,
          AppSpacing.lg,
          0,
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(AppRadii.xl),
          child: SizedBox(
            height: 168,
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
                  onTap: () => _onBannerTap(advertise),
                );
              },
              onTap: (int index) {
                _onBannerTap(advertiseList[index]);
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
                    onTap: () async {
                      await _handleShortcutTap(item);
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
            iconColor: const Color(0xFF334155),
            iconBackground: const Color(0xFFEFF6FF),
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
            iconColor: const Color(0xFF9A5C00),
            iconBackground: AppColors.accentSoft,
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
    Key? sectionKey,
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
        key: sectionKey,
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.lg,
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
          AppSpacing.lg,
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
                color: Color(0x0D101828),
                blurRadius: 14,
                offset: Offset(0, 6),
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

class _DailyLoginPointsRewardDialog extends StatelessWidget {
  final int points;

  const _DailyLoginPointsRewardDialog({
    required this.points,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    final Size size = MediaQuery.sizeOf(context);
    final double dialogWidth = size.width < 420 ? size.width - 40 : 380;

    return Dialog(
      insetPadding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.xl,
        vertical: AppSpacing.xxl,
      ),
      backgroundColor: Colors.transparent,
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: dialogWidth),
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: BorderRadius.circular(AppRadii.xxl),
            border: Border.all(color: AppColors.border),
            boxShadow: const [
              BoxShadow(
                color: Color(0x26101828),
                blurRadius: 28,
                offset: Offset(0, 16),
              ),
            ],
          ),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.xl,
              AppSpacing.xxl,
              AppSpacing.xl,
              AppSpacing.xl,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 58,
                  height: 58,
                  decoration: BoxDecoration(
                    color: AppColors.accentSoft,
                    borderRadius: BorderRadius.circular(AppRadii.xl),
                  ),
                  child: const Icon(
                    Icons.workspace_premium_rounded,
                    color: AppColors.accent,
                    size: 30,
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                Text(
                  '今日登录积分到账',
                  textAlign: TextAlign.center,
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
                Text(
                  '+$points',
                  textAlign: TextAlign.center,
                  style: theme.textTheme.headlineSmall?.copyWith(
                    color: AppColors.price,
                    fontSize: 44,
                    fontWeight: FontWeight.w900,
                    height: 1.0,
                  ),
                ),
                const SizedBox(height: AppSpacing.xs),
                Text(
                  '积分',
                  textAlign: TextAlign.center,
                  style: theme.textTheme.labelLarge?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                Text(
                  '已加入你的账户，可在下单和会员权益中使用。',
                  textAlign: TextAlign.center,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: AppColors.textSecondary,
                    height: 1.5,
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    onPressed: () => Navigator.of(context).pop(),
                    style: FilledButton.styleFrom(
                      minimumSize: const Size(double.infinity, 48),
                      backgroundColor: AppColors.primary,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(AppRadii.md),
                      ),
                    ),
                    child: const Text('知道了'),
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

enum _HomeShortcutTarget {
  brand,
  flash,
  newProduct,
  memberBenefits,
}

class _HomeShortcut {
  final _HomeShortcutTarget target;
  final String label;
  final IconData icon;
  final Color accentColor;
  final Color backgroundColor;

  const _HomeShortcut({
    required this.target,
    required this.label,
    required this.icon,
    required this.accentColor,
    required this.backgroundColor,
  });
}

class _TopIconButton extends StatelessWidget {
  final IconData icon;
  final String semanticLabel;
  final VoidCallback onTap;
  final int badgeCount;

  const _TopIconButton({
    required this.icon,
    required this.semanticLabel,
    required this.onTap,
    this.badgeCount = 0,
  });

  @override
  Widget build(BuildContext context) {
    final String badgeText = badgeCount > 99 ? '99+' : badgeCount.toString();
    return Semantics(
      button: true,
      label: badgeCount > 0 ? '$semanticLabel，$badgeCount 条未读' : semanticLabel,
      child: Material(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(AppRadii.md),
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.md),
          onTap: onTap,
          child: SizedBox(
            width: 40,
            height: 40,
            child: Stack(
              clipBehavior: Clip.none,
              children: [
                Center(
                  child: Icon(icon, color: AppColors.textPrimary, size: 22),
                ),
                if (badgeCount > 0)
                  Positioned(
                    top: 5,
                    right: 4,
                    child: Container(
                      constraints: const BoxConstraints(
                        minWidth: 16,
                        minHeight: 16,
                      ),
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      decoration: BoxDecoration(
                        color: AppColors.price,
                        borderRadius: BorderRadius.circular(999),
                        border: Border.all(color: AppColors.surface, width: 1),
                      ),
                      alignment: Alignment.center,
                      child: Text(
                        badgeText,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 9,
                          fontWeight: FontWeight.w700,
                          height: 1,
                        ),
                      ),
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

class _SearchTrigger extends StatelessWidget {
  final VoidCallback onTap;

  const _SearchTrigger({required this.onTap});

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    return Semantics(
      button: true,
      label: '搜索商品',
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: AppColors.background,
          borderRadius: BorderRadius.circular(AppRadii.md),
          border: Border.all(color: AppColors.border),
        ),
        child: Material(
          color: Colors.transparent,
          borderRadius: BorderRadius.circular(AppRadii.md),
          child: InkWell(
            borderRadius: BorderRadius.circular(AppRadii.md),
            onTap: onTap,
            child: SizedBox(
              height: 46,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
                child: Row(
                  children: [
                    Icon(
                      Icons.search_rounded,
                      color: AppColors.textSecondary.withValues(alpha: 0.9),
                      size: 22,
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(
                      child: Text(
                        '搜索商品，例如：手机、家电',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ),
                  ],
                ),
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
        vertical: AppSpacing.xs,
      ),
      decoration: BoxDecoration(
        color: AppColors.primarySoft,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: theme.textTheme.labelMedium?.copyWith(
          color: AppColors.textSecondary,
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
    final String normalizedImageUrl = imageUrl.trim();
    final bool shouldShowCopy = normalizedImageUrl.isEmpty ||
        normalizedImageUrl.contains('example.com');
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: LayoutBuilder(
          builder: (context, constraints) {
            final double textWidth = constraints.maxWidth * 0.66;
            return Stack(
              fit: StackFit.expand,
              children: [
                CachedImageWidget(
                  double.infinity,
                  double.infinity,
                  imageUrl,
                  fallback: const _BannerFallbackVisual(),
                ),
                if (shouldShowCopy)
                  DecoratedBox(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.centerLeft,
                        end: Alignment.centerRight,
                        colors: [
                          Colors.black.withValues(alpha: 0.72),
                          Colors.black.withValues(alpha: 0.38),
                          Colors.black.withValues(alpha: 0.06),
                        ],
                      ),
                    ),
                  ),
                if (shouldShowCopy)
                  Padding(
                    padding: const EdgeInsets.all(AppSpacing.lg),
                    child: SizedBox(
                      width: textWidth,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: AppSpacing.sm,
                              vertical: AppSpacing.xs,
                            ),
                            decoration: BoxDecoration(
                              color: AppColors.accent.withValues(alpha: 0.96),
                              borderRadius: BorderRadius.circular(999),
                            ),
                            child: Text(
                              '今日精选',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: Colors.white,
                                fontWeight: FontWeight.w800,
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
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                          const SizedBox(height: AppSpacing.xs),
                          Text(
                            subtitle,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: Colors.white.withValues(alpha: 0.9),
                              height: 1.35,
                            ),
                          ),
                          const SizedBox(height: AppSpacing.sm),
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Flexible(
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: AppSpacing.md,
                                    vertical: AppSpacing.xs,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.white,
                                    borderRadius: BorderRadius.circular(999),
                                  ),
                                  child: Text(
                                    '立即查看',
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style:
                                        theme.textTheme.labelMedium?.copyWith(
                                      color: AppColors.primaryDark,
                                      fontWeight: FontWeight.w800,
                                    ),
                                  ),
                                ),
                              ),
                              const SizedBox(width: AppSpacing.sm),
                              Icon(
                                Icons.arrow_forward_rounded,
                                color: Colors.white.withValues(alpha: 0.92),
                                size: 20,
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
              ],
            );
          },
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
            Color(0xFF92400E),
          ],
        ),
      ),
      child: Stack(
        children: [
          Positioned(
            top: 22,
            right: 18,
            child: Container(
              width: 116,
              height: 64,
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(AppRadii.xl),
              ),
            ),
          ),
          Positioned(
            left: 24,
            bottom: 22,
            child: Container(
              width: 132,
              height: 44,
              decoration: BoxDecoration(
                color: AppColors.accent.withValues(alpha: 0.22),
                borderRadius: BorderRadius.circular(AppRadii.lg),
              ),
            ),
          ),
          Positioned(
            right: 34,
            bottom: 22,
            child: Container(
              width: 108,
              height: 108,
              decoration: BoxDecoration(
                border: Border.all(
                  color: Colors.white.withValues(alpha: 0.18),
                  width: 1.5,
                ),
                borderRadius: BorderRadius.circular(AppRadii.xxl),
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
                  width: 52,
                  height: 52,
                  decoration: BoxDecoration(
                    color: item.backgroundColor,
                    borderRadius: BorderRadius.circular(AppRadii.xl),
                  ),
                  child: Icon(item.icon, color: item.accentColor, size: 26),
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
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
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
            style: TextButton.styleFrom(
              foregroundColor: AppColors.primary,
              textStyle: theme.textTheme.labelMedium,
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm),
            ),
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
              color: AppColors.accentSoft,
              borderRadius: BorderRadius.circular(999),
            ),
            child: Text(
              badgeLabel!,
              style: theme.textTheme.labelSmall?.copyWith(
                color: AppColors.price,
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
              color: AppColors.price,
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
