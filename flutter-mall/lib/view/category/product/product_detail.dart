import 'package:bottom_sheet/bottom_sheet.dart';
import 'package:card_swiper/card_swiper.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/coupon_model.dart' as coupon_model;
import 'package:flutter_mall/model/direct_checkout.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/cart/cart.dart';
import 'package:flutter_mall/view/mine/order/order_submit.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

import '../../../layout/main_tab.dart';
import '../../../model/product_detail.dart';

/// 规格选择弹窗的来源动作：决定点击"确定"后的行为
enum SkuSheetAction {
  /// 仅选择规格（点击"购买类型"行打开），确定后只关闭弹窗
  select,

  /// 加入购物车（点击"加入购物车"按钮打开），确定后关闭弹窗并加购
  addCart,

  /// 立即购买（点击"立即购买"按钮打开），确定后关闭弹窗并下单
  buyNow,
}

///
/// 商品详情页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class ProductDetail extends StatefulWidget {
  final int productId;
  final String? intentSource;

  const ProductDetail({super.key, required this.productId, this.intentSource});

  @override
  State<ProductDetail> createState() => _ProductDetailState();
}

class _ProductDetailState extends State<ProductDetail> {
  Product? product;
  late Brand brand;
  List<ProductAttributeList> productAttributeList = [];
  List<ProductAttributeValueList> productAttributeValueList = [];
  List<SkuStockList> skuStockList = [];
  List<CouponList> couponList = [];
  List<ProductLadderList> productLadderList = [];
  List<ProductFullReductionList> productFullReductionList = [];
  List<MemberPriceList> memberPriceList = [];
  ProductVisibility visibility = ProductVisibility.fromJson({});
  bool loading = true;
  bool _isAdding = false; // 加购按钮防抖
  bool _isCollecting = false;
  bool _isCollected = false;
  String? _collectionId;
  SkuStockList? selectedSku;
  final Map<String, String> _specSelection = {};

  @override
  void initState() {
    super.initState();
    refreshProductDetail();
  }

  void refreshProductDetail() async {
    try {
      Response result = await HttpUtil.get(
        productDetailDataUrl + widget.productId.toString(),
      );
      ProductDetailModel productDetailModel = ProductDetailModel.fromJson(
        result.data,
      );
      ProductDetailData productDetailData = productDetailModel.data;
      if (!mounted) {
        return;
      }
      setState(() {
        loading = false;
        visibility = productDetailData.visibility;
        if (productDetailData.visibility.visible) {
          product = productDetailData.product;
          brand = productDetailData.brand;
          productAttributeList = productDetailData.productAttributeList;
          productAttributeValueList =
              productDetailData.productAttributeValueList;
          skuStockList = productDetailData.skuStockList;
          couponList = productDetailData.couponList;
          productLadderList = productDetailData.productLadderList;
          productFullReductionList = productDetailData.productFullReductionList;
          memberPriceList = productDetailData.memberPriceList;
          _resetSkuSelection();
        } else {
          product = null;
          productAttributeList = [];
          productAttributeValueList = [];
          skuStockList = [];
          couponList = [];
          productLadderList = [];
          productFullReductionList = [];
          memberPriceList = [];
          _clearSkuSelection();
        }
      });
      if (productDetailData.visibility.visible) {
        await _refreshCouponReceiveStatus();
        await _refreshCollectionStatus();
      }
      if (productDetailData.visibility.visible) {
        await AppRecoveryStore.saveRecentContext(
          AppRecentContext.create(
            targetType: AppRecentTargetType.productDetail,
            targetId: widget.productId,
            source: widget.intentSource ?? 'manual_open',
            requiresAuth: false,
            fallbackType: AppRecentTargetType.home,
            fallbackTabIndex: 0,
          ),
        );
      } else {
        final currentContext = AppRecoveryStore.getRecentContext();
        if (currentContext?.targetType == AppRecentTargetType.productDetail &&
            currentContext?.targetId == widget.productId) {
          await AppRecoveryStore.clearRecentContext();
        }
      }
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        loading = false;
        product = null;
        productAttributeList = [];
        productAttributeValueList = [];
        skuStockList = [];
        couponList = [];
        productLadderList = [];
        productFullReductionList = [];
        memberPriceList = [];
        _clearSkuSelection();
        visibility = ProductVisibility(
          visible: false,
          purchasable: false,
          showPrice: false,
          showStock: false,
          status: "hidden",
          reasonCode: "request_failed",
          reasonMessage: "商品详情加载失败",
          recoveryHint: "请稍后重试，或先返回首页继续浏览",
          fallbackAction: "go_home",
          fallbackTarget: "home",
        );
      });
      final currentContext = AppRecoveryStore.getRecentContext();
      if (currentContext?.targetType == AppRecentTargetType.productDetail &&
          currentContext?.targetId == widget.productId) {
        await AppRecoveryStore.clearRecentContext();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        surfaceTintColor: AppColors.surface,
        elevation: 0,
        title: const Text("商品详情"),
        titleTextStyle: const TextStyle(
          fontSize: 16,
          color: AppColors.textPrimary,
          fontWeight: FontWeight.w600,
        ),
        centerTitle: true,
      ),
      body: Container(
        color: AppColors.background,
        width: MediaQuery.of(context).size.width,
        child: buildPageBody(context),
      ),
      bottomNavigationBar:
          !loading && product != null ? buildFooter(context) : null,
    );
  }

  Widget buildPageBody(BuildContext context) {
    if (loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (!visibility.visible || product == null) {
      return buildUnavailableState(context);
    }
    return ListView(
      padding: const EdgeInsets.only(bottom: AppSpacing.lg),
      children: [
        buildProductPic(),
        buildProductBaseInfo(),
        buildProductShare(),
        buildAttributesInfo(context),
        buildPinJiaInfo(),
        buildBrandInfo(),
        buildImageDetailInfo(),
      ],
    );
  }

  BoxDecoration _cardDecoration({Color color = AppColors.surface}) {
    return BoxDecoration(
      color: color,
      borderRadius: BorderRadius.circular(AppRadii.lg),
      border: Border.all(color: AppColors.border.withValues(alpha: 0.78)),
    );
  }

  Widget _sectionCard({
    required Widget child,
    EdgeInsetsGeometry margin = const EdgeInsets.fromLTRB(
      AppSpacing.md,
      AppSpacing.md,
      AppSpacing.md,
      0,
    ),
    EdgeInsetsGeometry padding = const EdgeInsets.all(AppSpacing.lg),
    Color color = AppColors.surface,
  }) {
    return Container(
      margin: margin,
      padding: padding,
      decoration: _cardDecoration(color: color),
      child: child,
    );
  }

  Widget _statTile(String label, String value) {
    return Expanded(
      child: Text(
        "$label $value",
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        textAlign: TextAlign.center,
        style: const TextStyle(
          fontSize: 12,
          color: AppColors.textHint,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }

  // 商品图片
  Widget buildProductPic() {
    List<String> list = [];
    if (product != null) {
      list = product!.albumPics.split(",");
      //如果画册图片为空,则区主图
      if (list[0] == "") {
        list[0] = product!.mainPic;
      }
    }

    return AspectRatio(
      aspectRatio: 1,
      child: Container(
        clipBehavior: Clip.antiAlias,
        decoration: const BoxDecoration(color: AppColors.surface),
        child: Swiper(
          itemBuilder: (BuildContext context, int index) {
            return CachedImageWidget(
              double.infinity,
              double.infinity,
              list[index],
              fit: BoxFit.cover,
            );
          },
          autoplay: list.length > 1,
          itemCount: list.length,
          pagination: SwiperPagination(
            builder: DotSwiperPaginationBuilder(
              activeColor: AppColors.primary,
              color: AppColors.border,
              activeSize: 7,
              size: 7,
              space: 4,
            ),
          ),
        ),
      ),
    );
  }

  // 商品基本信息
  Widget buildProductBaseInfo() {
    final priceText = selectedSku != null
        ? '${selectedSku!.price}'
        : product!.price.toString();
    final salesText =
        '${selectedSku != null ? selectedSku!.sales : product!.sales}';
    final stockText =
        '${selectedSku != null ? selectedSku!.stock : product!.stock}';

    return _sectionCard(
      margin: const EdgeInsets.only(top: AppSpacing.sm),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              const Text(
                "¥",
                style: TextStyle(
                  fontSize: 16,
                  color: AppColors.price,
                  fontWeight: FontWeight.w800,
                ),
              ),
              Text(
                priceText,
                style: const TextStyle(
                  fontSize: 30,
                  height: 1,
                  color: AppColors.price,
                  fontWeight: FontWeight.w800,
                ),
              ),
              if (selectedSku == null && product!.priceRange.contains('-'))
                Padding(
                  padding: const EdgeInsets.only(
                    left: AppSpacing.sm,
                    bottom: 2,
                  ),
                  child: Text(
                    product!.priceRange,
                    style: const TextStyle(
                      fontSize: 12,
                      color: AppColors.textHint,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          Text(
            product!.name,
            style: const TextStyle(
              fontSize: 16,
              height: 1.35,
              fontWeight: FontWeight.w500,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(
            product!.subTitle,
            style: const TextStyle(
              fontSize: 13,
              height: 1.4,
              color: AppColors.textSecondary,
            ),
          ),
          if (!visibility.purchasable && visibility.reasonMessage.isNotEmpty)
            Container(
              margin: const EdgeInsets.only(top: AppSpacing.md),
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.md,
                vertical: AppSpacing.xs,
              ),
              decoration: BoxDecoration(
                color: AppColors.price.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(AppRadii.sm),
              ),
              child: Text(
                visibility.reasonMessage,
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.price,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          const SizedBox(height: AppSpacing.md),
          const Divider(height: 1),
          const SizedBox(height: AppSpacing.sm),
          Row(
            children: [
              _statTile('销量', salesText),
              _statTile('库存', stockText),
              _statTile('浏览', '768'),
            ],
          ),
        ],
      ),
    );
  }

  // 分享
  Widget buildProductShare() {
    return _sectionCard(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.sm,
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(
              horizontal: AppSpacing.sm,
              vertical: 3,
            ),
            decoration: BoxDecoration(
              color: AppColors.price.withValues(alpha: 0.10),
              borderRadius: BorderRadius.circular(AppRadii.sm),
            ),
            child: const Text(
              "返券",
              style: TextStyle(
                fontSize: 12,
                color: AppColors.price,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          const SizedBox(width: AppSpacing.sm),
          const Expanded(
            child: Text(
              "分享商品可领 49 减 10 红包",
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                fontSize: 14,
                color: AppColors.textPrimary,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          const SizedBox(width: AppSpacing.sm),
          const Text(
            "立即分享",
            style: TextStyle(
              fontSize: 13,
              color: AppColors.accent,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(width: 2),
          const Icon(
            Icons.chevron_right,
            size: 20,
            color: AppColors.accent,
          ),
        ],
      ),
    );
  }

  // 商品属性和规格
  Widget buildAttributesInfo(BuildContext context) {
    return _sectionCard(
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          buildTxt("购买类型", _selectedSkuSummary(), 1, context),
          buildTxt("商品参数", "查看", 2, context),
          buildTxt("优惠券", _couponSummary(), 3, context),
          buildTxt("促销活动", _promotionSummary(), 4, context),
          buildTxt("商家服务", "无忧退货 · 快速退款 · 免费包邮", 5, context),
        ],
      ),
    );
  }

  Widget buildTxt(String title, String value, int flag, BuildContext context) {
    final bool highlight = flag == 3;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: () {
          _openBottomSheetWithInfo(context, title);
        },
        child: Container(
          constraints: const BoxConstraints(minHeight: 56),
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.lg,
            vertical: AppSpacing.md,
          ),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                color: flag == 5 ? Colors.transparent : AppColors.border,
              ),
            ),
          ),
          child: Row(
            children: [
              SizedBox(
                width: 68,
                child: Text(
                  title,
                  style: const TextStyle(
                    fontSize: 14,
                    color: AppColors.textSecondary,
                  ),
                ),
              ),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Text(
                  value,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: 14,
                    color: highlight ? AppColors.price : AppColors.textPrimary,
                    fontWeight: highlight ? FontWeight.w700 : FontWeight.w600,
                  ),
                ),
              ),
              const Icon(
                Icons.chevron_right,
                size: 20,
                color: AppColors.textHint,
              ),
            ],
          ),
        ),
      ),
    );
  }

  // 评价信息
  Widget buildPinJiaInfo() {
    return _sectionCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            height: 40,
            child: Row(
              children: [
                const Text(
                  "评价",
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(width: AppSpacing.xs),
                const Expanded(
                  child: Text(
                    "(86)",
                    style: TextStyle(
                      fontSize: 14,
                      color: AppColors.textHint,
                    ),
                  ),
                ),
                const Text(
                  "好评率 100% ",
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textHint,
                  ),
                ),
                const Icon(
                  Icons.chevron_right,
                  size: 20,
                  color: AppColors.textHint,
                ),
              ],
            ),
          ),
          Text(
            "koobe",
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textPrimary,
              fontWeight: FontWeight.w700,
            ),
          ),
          Container(
            margin: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
            child: const Text(
              "商品收到了，79元两件，质量不错，试了一下有点瘦，但是加个外罩很漂亮，我很喜欢",
              style: TextStyle(
                fontSize: 14,
                height: 1.45,
                color: AppColors.textPrimary,
              ),
            ),
          ),
          const Row(
            children: [
              Expanded(
                child: Text(
                  "购买类型：XL 红色",
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textHint,
                  ),
                ),
              ),
              Text(
                "2019-04-01 19:21",
                style: TextStyle(
                  fontSize: 14,
                  color: AppColors.textHint,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // 品牌信息
  Widget buildBrandInfo() {
    return _sectionCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            "品牌信息",
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: AppSpacing.md),
          Row(
            mainAxisAlignment: MainAxisAlignment.start,
            children: [
              Container(
                width: 112,
                height: 46,
                padding: const EdgeInsets.all(AppSpacing.sm),
                decoration: BoxDecoration(
                  color: AppColors.surfaceMuted,
                  borderRadius: BorderRadius.circular(AppRadii.md),
                ),
                child: CachedImageWidget(
                  96,
                  30,
                  brand.logo,
                  fit: BoxFit.contain,
                ),
              ),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      brand.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.w700,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: AppSpacing.xs),
                    Text(
                      "品牌首字母：${brand.firstLetter}",
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppColors.textHint,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // 图文详情
  Widget buildImageDetailInfo() {
    List<String> imageUrls = [];
    final detailContent =
        product?.detailMobileHtml ?? product?.detailHtml ?? '';
    if (detailContent.isNotEmpty) {
      final urlPattern =
          RegExp(r'(https?://[^\s"<>]+\.(?:jpg|jpeg|png|gif|webp)[^\s"<>]*)');
      imageUrls =
          urlPattern.allMatches(detailContent).map((m) => m.group(0)!).toList();
    }
    if (imageUrls.isEmpty && product != null && product!.albumPics.isNotEmpty) {
      imageUrls =
          product!.albumPics.split(",").where((s) => s.isNotEmpty).toList();
    }

    if (imageUrls.isEmpty) {
      return Column(
        children: [
          _sectionCard(
            child: Column(
              children: const [
                Text(
                  "图文详情",
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
                SizedBox(height: AppSpacing.md),
                Text(
                  '暂无图文详情',
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textHint,
                  ),
                ),
              ],
            ),
          ),
        ],
      );
    }

    return Column(
      children: [
        _sectionCard(
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.lg,
            vertical: AppSpacing.md,
          ),
          child: const Center(
            child: Text(
              "图文详情",
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w700,
                color: AppColors.textPrimary,
              ),
            ),
          ),
        ),
        Container(
          margin: const EdgeInsets.fromLTRB(
            AppSpacing.lg,
            AppSpacing.md,
            AppSpacing.lg,
            0,
          ),
          clipBehavior: Clip.antiAlias,
          decoration: _cardDecoration(),
          child: Column(
            children: imageUrls.map((url) {
              return Image.network(kIsWeb ? proxyImageUrl(url) : url);
            }).toList(),
          ),
        ),
      ],
    );
  }

  // 底部悬浮
  Widget buildFooter(BuildContext context) {
    final bool spuDisabled = !visibility.purchasable;
    final SkuStockList? resolvedSku =
        selectedSku ?? (skuStockList.length == 1 ? skuStockList.first : null);
    final bool skuDisabled = resolvedSku != null && !resolvedSku.purchasable;
    final bool disabled = spuDisabled || skuDisabled;
    final bool isAddingNow = _isAdding;
    final bool needSelectSku =
        !spuDisabled && resolvedSku == null && skuStockList.length > 1;

    final String disabledLabel = spuDisabled
        ? (visibility.reasonMessage.isNotEmpty
            ? visibility.reasonMessage
            : '暂不可购买')
        : (resolvedSku?.purchaseReasonLabel.isNotEmpty == true
            ? resolvedSku!.purchaseReasonLabel
            : '暂不可购买');

    return SafeArea(
      top: false,
      child: Container(
        height: 64,
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.sm,
          vertical: 8,
        ),
        decoration: BoxDecoration(
          color: AppColors.surface,
          border: const Border(
            top: BorderSide(
              color: AppColors.border,
              width: 1,
            ),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              flex: 5,
              child: Row(
                children: [
                  buildImage(Icons.home_outlined, "首页"),
                  const SizedBox(width: 6),
                  buildImage(Icons.shopping_cart_outlined, "购物车"),
                  const SizedBox(width: 6),
                  buildCollectionAction(),
                ],
              ),
            ),
            Expanded(
              flex: 6,
              child: Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: disabled || isAddingNow
                          ? null
                          : needSelectSku
                              ? () => _openBottomSheetWithInfo(context, "购买类型",
                                  skuAction: SkuSheetAction.addCart)
                              : () => _addCart(product!),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 11),
                        foregroundColor: AppColors.primary,
                        side: BorderSide(
                          color: (disabled && !needSelectSku)
                              ? AppColors.border
                              : AppColors.primary,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(24),
                        ),
                      ),
                      child: isAddingNow
                          ? SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor: AlwaysStoppedAnimation<Color>(
                                  AppColors.primary,
                                ),
                              ),
                            )
                          : Text(
                              disabled && !needSelectSku
                                  ? disabledLabel
                                  : '加入购物车',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: disabled
                          ? null
                          : needSelectSku
                              ? () => _openBottomSheetWithInfo(context, "购买类型",
                                  skuAction: SkuSheetAction.buyNow)
                              : _buyNow,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        disabledBackgroundColor: AppColors.textHint,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 11),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(24),
                        ),
                        elevation: 0,
                      ),
                      child: Text(
                        disabled ? disabledLabel : '立即购买',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget buildUnavailableState(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.inventory_2_outlined,
              size: 72,
              color: AppColors.textHint,
            ),
            const SizedBox(height: 16),
            Text(
              visibility.reasonMessage.isNotEmpty
                  ? visibility.reasonMessage
                  : "商品暂不可查看",
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
            ),
            if (visibility.recoveryHint.isNotEmpty) ...[
              const SizedBox(height: 10),
              Text(
                visibility.recoveryHint,
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => _handleFallbackAction(context),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.primary,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                ),
                child: Text(buildFallbackActionLabel()),
              ),
            ),
            if (Navigator.of(context).canPop()) ...[
              const SizedBox(height: 12),
              TextButton(
                onPressed: () => Navigator.of(context).pop(),
                child: const Text("返回上一页"),
              ),
            ],
          ],
        ),
      ),
    );
  }

  String buildFallbackActionLabel() {
    switch (visibility.fallbackAction) {
      case "browse_product_list":
      case "browse_similar":
        return "返回继续逛";
      case "go_home":
        return "回到首页";
      default:
        return "继续浏览";
    }
  }

  void _handleFallbackAction(BuildContext context) {
    switch (visibility.fallbackAction) {
      case "browse_product_list":
      case "browse_similar":
        if (Navigator.of(context).canPop()) {
          Navigator.of(context).pop();
          return;
        }
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (context) => const MainTab(),
          ),
        );
        return;
      case "go_home":
      default:
        Navigator.of(context).pushAndRemoveUntil(
          MaterialPageRoute(
            builder: (context) => const MainTab(),
          ),
          (route) => false,
        );
        return;
    }
  }

  SkuStockList? _resolvePurchaseSku({bool promptSelectorOnMissing = true}) {
    final sku =
        selectedSku ?? (skuStockList.length == 1 ? skuStockList.first : null);
    if (sku == null) {
      if (promptSelectorOnMissing && skuStockList.length > 1) {
        _openBottomSheetWithInfo(context, "购买类型");
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('请先选择规格')),
        );
      }
      return null;
    }
    if (!sku.purchasable) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            sku.purchaseReasonLabel.isNotEmpty
                ? sku.purchaseReasonLabel
                : '该规格暂不可购买',
          ),
        ),
      );
      return null;
    }
    return sku;
  }

  void _buyNow() {
    if (product == null) {
      return;
    }
    final sku = _resolvePurchaseSku();
    if (sku == null) {
      return;
    }
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => OrderSubmit(
          directItem: DirectCheckoutParams(
            productId: product!.id,
            productSkuId: sku.id,
            quantity: 1,
          ),
        ),
      ),
    );
  }

  Widget buildImage(IconData icon, String title) {
    return Semantics(
      button: true,
      label: title,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.md),
          onTap: () {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) {
                  if (title.contains("首页")) {
                    return const MainTab();
                  }
                  return const Cart();
                },
              ),
            );
          },
          child: SizedBox(
            width: 44,
            height: 48,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  icon,
                  size: 22,
                  color: AppColors.textSecondary,
                ),
                const SizedBox(height: 2),
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 11,
                    color: AppColors.textSecondary,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget buildCollectionAction() {
    return Semantics(
      button: true,
      label: _isCollected ? "已收藏，点击取消收藏" : "收藏",
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.md),
          onTap: _isCollecting
              ? null
              : _isCollected
                  ? _deleteCollection
                  : _addCollection,
          child: SizedBox(
            width: 44,
            height: 48,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                _isCollecting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Icon(
                        _isCollected
                            ? Icons.favorite_rounded
                            : Icons.favorite_border_rounded,
                        size: 22,
                        color: _isCollected
                            ? AppColors.primary
                            : AppColors.textSecondary,
                      ),
                const SizedBox(height: 2),
                Text(
                  _isCollected ? "已收藏" : "收藏",
                  style: TextStyle(
                    fontSize: 11,
                    color: _isCollected
                        ? AppColors.primary
                        : AppColors.textSecondary,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _addCollection() async {
    final Product? currentProduct = product;
    if (currentProduct == null) {
      return;
    }
    setState(() {
      _isCollecting = true;
    });
    try {
      await HttpUtil.post(
        addCollectionDataUrl,
        data: {
          "productId": currentProduct.id,
          "productName": currentProduct.name,
          "productPic": currentProduct.mainPic,
          "productSubTitle": currentProduct.subTitle,
          "productPrice": _parseProductPrice(currentProduct.price),
        },
      );
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("已收藏")),
      );
      await _refreshCollectionStatus();
    } catch (_) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("收藏失败，请重试")),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isCollecting = false;
        });
      }
    }
  }

  int _parseProductPrice(String price) {
    return double.tryParse(price.trim())?.round() ?? 0;
  }

  Future<void> _deleteCollection() async {
    final String? currentCollectionId = _collectionId;
    if (currentCollectionId == null || currentCollectionId.isEmpty) {
      await _refreshCollectionStatus();
      if (_collectionId == null || _collectionId!.isEmpty) {
        _showCollectedMessage();
        return;
      }
    }
    setState(() {
      _isCollecting = true;
    });
    try {
      await HttpUtil.get("$deleteCollectionDataUrl?ids=${_collectionId!}");
      if (!mounted) {
        return;
      }
      setState(() {
        _isCollected = false;
        _collectionId = null;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("已取消收藏")),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("取消收藏失败，请重试")),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isCollecting = false;
        });
      }
    }
  }

  Future<void> _refreshCollectionStatus() async {
    try {
      final Response result = await HttpUtil.get(
        collectionListDataUrl,
        redirectOnUnauthorized: false,
      );
      final data =
          result.data is Map<String, dynamic> ? result.data["data"] : null;
      if (data is! List || !mounted) {
        return;
      }
      String? matchedId;
      final bool matched = data.any((item) {
        if (item is! Map) {
          return false;
        }
        final value = item["productId"];
        final bool sameProduct = value is int
            ? value == widget.productId
            : int.tryParse(value?.toString() ?? "") == widget.productId;
        if (sameProduct) {
          matchedId = item["id"]?.toString();
        }
        return sameProduct;
      });
      setState(() {
        _isCollected = matched;
        _collectionId = matched ? matchedId : null;
      });
    } catch (_) {
      // 未登录或接口异常不影响商品详情浏览和后续点击收藏登录流程。
    }
  }

  void _showCollectedMessage() {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text("已在我的收藏中")),
    );
  }

  void _openBottomSheetWithInfo(
    BuildContext context,
    String title, {
    SkuSheetAction skuAction = SkuSheetAction.select,
  }) {
    showFlexibleBottomSheet<void>(
      isExpand: false,
      initHeight: 0.8,
      maxHeight: 0.8,
      context: context,
      builder: (context, controller, offset) {
        if (title.contains("购买类型")) {
          return buildBuyType("title", "value", controller, skuAction);
        }
        if (title.contains("商品参数")) {
          return buildProductParams("title", "value", controller);
        }
        if (title.contains("优惠券")) {
          return buildCouponList(controller);
        }
        if (title.contains("促销活动")) {
          return buildPromotion("title", "value", controller);
        }
        if (title.contains("商家服务")) {
          return buildMerchantServices("title", "value", controller);
        }
        return buildMerchantServices("title", "value", controller);
      },
    );
  }

  String _promotionSummary() {
    final parts = <String>[];
    if (productFullReductionList.isNotEmpty) {
      final fr = productFullReductionList.first;
      parts.add('满${fr.fullPrice}减${fr.reducePrice}');
    }
    if (productLadderList.isNotEmpty) {
      final ld = productLadderList.first;
      parts.add('满${ld.count}件打折');
    }
    if (memberPriceList.isNotEmpty) {
      parts.add('会员价');
    }
    if (parts.isEmpty) return '暂无优惠';
    return parts.join(' · ');
  }

  String _couponSummary() {
    if (couponList.isEmpty) return '暂无优惠券';
    final first = couponList.first;
    final prefix =
        first.amount > 0 ? '满${first.minAmount}减${first.amount}' : first.name;
    if (couponList.length > 1) {
      return '$prefix 等${couponList.length}张券';
    }
    return prefix;
  }

  String _selectedSkuSummary() {
    if (selectedSku != null) {
      final specs = selectedSku!.parsedSpecData;
      if (specs.isNotEmpty) {
        return specs.values.join(' ');
      }
      return selectedSku!.name.isNotEmpty ? selectedSku!.name : '已选规格';
    }
    if (skuStockList.length == 1) {
      return '默认规格';
    }
    return '请选择规格';
  }

  Future<void> _refreshCouponReceiveStatus() async {
    final currentToken = SharedPreferencesUtil.getString(token);
    if (currentToken == null ||
        currentToken.trim().isEmpty ||
        couponList.isEmpty) {
      return;
    }

    try {
      final result = await HttpUtil.get(availableCouponUrl);
      final couponModel = coupon_model.CouponModel.fromJson(result.data);
      final receiveStatusMap = <int, int>{
        for (final item in couponModel.data) item.id: item.receiveStatus,
      };

      bool changed = false;
      for (final coupon in couponList) {
        final latestStatus = receiveStatusMap[coupon.id];
        if (latestStatus != null && latestStatus != coupon.receiveStatus) {
          coupon.receiveStatus = latestStatus;
          changed = true;
        }
      }

      if (changed && mounted) {
        setState(() {});
      }
    } on DioException catch (e) {
      if (e.response?.statusCode == 401 || e.response?.statusCode == 403) {
        return;
      }
    } catch (_) {
      // 忽略优惠券状态同步失败，保留商品详情主链路可用
    }
  }

  void _clearSkuSelection() {
    _specSelection.clear();
    selectedSku = null;
  }

  void _resetSkuSelection() {
    _clearSkuSelection();
    if (skuStockList.isEmpty) {
      return;
    }
    if (skuStockList.length == 1) {
      selectedSku = skuStockList.first;
      _specSelection.addAll(selectedSku!.parsedSpecData);
      return;
    }

    final specOptions = _buildSpecOptions();
    for (final entry in specOptions.entries) {
      if (entry.value.length == 1) {
        _specSelection[entry.key] = entry.value.first;
      }
    }
  }

  Map<String, List<String>> _buildSpecOptions() {
    final map = <String, List<String>>{};
    for (final sku in skuStockList) {
      final specs = sku.parsedSpecData;
      for (final entry in specs.entries) {
        map.putIfAbsent(entry.key, () => []);
        if (!map[entry.key]!.contains(entry.value)) {
          map[entry.key]!.add(entry.value);
        }
      }
    }
    return map;
  }

  SkuStockList? _findExactMatchingSku(Map<String, String> selection) {
    for (final sku in skuStockList) {
      final specs = sku.parsedSpecData;
      if (specs.length != selection.length) continue;
      bool match = true;
      for (final entry in selection.entries) {
        if (specs[entry.key] != entry.value) {
          match = false;
          break;
        }
      }
      if (match) return sku;
    }
    return null;
  }

  bool _isOptionAvailable(Map<String, String> testSelection) {
    for (final sku in skuStockList) {
      final specs = sku.parsedSpecData;
      bool match = true;
      for (final entry in testSelection.entries) {
        if (specs[entry.key] != entry.value) {
          match = false;
          break;
        }
      }
      if (match && sku.purchasable) return true;
    }
    return false;
  }

  // 购买类型（规格选择器）
  Widget buildBuyType(
    String title,
    String value,
    ScrollController controller,
    SkuSheetAction skuAction,
  ) {
    final specOptions = _buildSpecOptions();

    if (specOptions.isEmpty) {
      return ListView(
        controller: controller,
        shrinkWrap: true,
        children: [
          Container(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                if (product != null) _buildSkuPriceHeader(selectedSku),
                const SizedBox(height: 16),
                Text(
                  '该商品无可选规格',
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textHint,
                  ),
                ),
              ],
            ),
          ),
        ],
      );
    }

    return StatefulBuilder(
      builder: (context, setSheetState) {
        return ListView(
          controller: controller,
          shrinkWrap: true,
          children: [
            Container(
              padding: const EdgeInsets.all(15),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSkuPriceHeader(selectedSku),
                  const SizedBox(height: 16),
                  ...specOptions.entries.map((entry) {
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Padding(
                          padding: const EdgeInsets.only(bottom: 8, top: 12),
                          child: Text(
                            entry.key,
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ),
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: entry.value.map((val) {
                            final isSelected = _specSelection[entry.key] == val;

                            final testSelection =
                                Map<String, String>.from(_specSelection);
                            testSelection[entry.key] = val;
                            final isAvailable =
                                _isOptionAvailable(testSelection);

                            return ChoiceChip(
                              label: Text(val),
                              selected: isSelected,
                              selectedColor:
                                  AppColors.price.withValues(alpha: 0.14),
                              backgroundColor:
                                  !isAvailable ? AppColors.surfaceMuted : null,
                              labelStyle: TextStyle(
                                color: isSelected
                                    ? AppColors.price
                                    : !isAvailable
                                        ? AppColors.textHint
                                        : AppColors.textPrimary,
                                fontSize: 13,
                              ),
                              side: isSelected
                                  ? const BorderSide(color: AppColors.price)
                                  : null,
                              onSelected: !isAvailable
                                  ? null
                                  : (selected) {
                                      setSheetState(() {
                                        if (selected) {
                                          _specSelection[entry.key] = val;
                                        } else {
                                          _specSelection.remove(entry.key);
                                        }
                                      });
                                      final matched =
                                          _findExactMatchingSku(_specSelection);
                                      setState(() {
                                        selectedSku = _specSelection.length ==
                                                specOptions.length
                                            ? matched
                                            : null;
                                      });
                                    },
                            );
                          }).toList(),
                        ),
                      ],
                    );
                  }),
                  if (selectedSku != null && !selectedSku!.purchasable)
                    Container(
                      margin: const EdgeInsets.only(top: 16),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 8),
                      decoration: BoxDecoration(
                        color: AppColors.price.withValues(alpha: 0.08),
                        borderRadius:
                            const BorderRadius.all(Radius.circular(8)),
                      ),
                      child: Row(
                        children: [
                          Icon(Icons.info_outline,
                              size: 16, color: AppColors.price),
                          const SizedBox(width: 6),
                          Text(
                            selectedSku!.purchaseReasonLabel,
                            style: const TextStyle(
                              fontSize: 13,
                              color: AppColors.price,
                            ),
                          ),
                        ],
                      ),
                    ),
                  const SizedBox(height: 20),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed:
                          (selectedSku != null && selectedSku!.purchasable)
                              ? () {
                                  Navigator.of(context).pop();
                                  // 根据弹窗来源执行后续动作：
                                  // 加入购物车 → 直接加购；立即购买 → 下单；
                                  // 仅选规格 → 关闭弹窗即可。
                                  switch (skuAction) {
                                    case SkuSheetAction.addCart:
                                      if (product != null) {
                                        _addCart(product!);
                                      }
                                      break;
                                    case SkuSheetAction.buyNow:
                                      _buyNow();
                                      break;
                                    case SkuSheetAction.select:
                                      break;
                                  }
                                }
                              : null,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        disabledBackgroundColor: AppColors.textHint,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(25),
                        ),
                      ),
                      child: Text(
                        selectedSku == null
                            ? '请选择完整规格'
                            : selectedSku!.purchasable
                                ? '确定'
                                : selectedSku!.purchaseReasonLabel,
                        style: const TextStyle(
                            fontSize: 15, fontWeight: FontWeight.bold),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        );
      },
    );
  }

  Widget _buildSkuPriceHeader(SkuStockList? sku) {
    final displayPrice = sku != null ? '${sku.price}' : (product?.price ?? '');
    final displayStock = sku != null ? sku.stock : (product?.stock ?? 0);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Text(
          '¥',
          style: TextStyle(
            fontSize: 16,
            color: AppColors.price,
            fontWeight: FontWeight.bold,
          ),
        ),
        Text(
          displayPrice,
          style: TextStyle(
            fontSize: 24,
            color: AppColors.price,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(width: 12),
        Text(
          '库存: $displayStock',
          style: TextStyle(
            fontSize: 13,
            color: AppColors.textHint,
          ),
        ),
      ],
    );
  }

  // 商品参数
  Widget buildProductParams(
    String title,
    String value,
    ScrollController controller,
  ) {
    final attrMap = <int, String>{};
    for (final attr in productAttributeList) {
      attrMap[attr.id] = attr.name;
    }

    final paramRows = <MapEntry<String, String>>[];
    for (final av in productAttributeValueList) {
      final attrName = attrMap[av.attributeId] ?? '属性${av.attributeId}';
      paramRows.add(MapEntry(attrName, av.value));
    }

    if (paramRows.isEmpty) {
      return ListView(
        controller: controller,
        shrinkWrap: true,
        children: [
          Container(
            padding: const EdgeInsets.all(20),
            alignment: Alignment.center,
            child: Text(
              '暂无商品参数',
              style: TextStyle(
                fontSize: 14,
                color: AppColors.textHint,
              ),
            ),
          ),
        ],
      );
    }

    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: paramRows.length,
      itemBuilder: (context, index) {
        final entry = paramRows[index];
        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 12),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                width: 1,
                color: AppColors.surfaceMuted,
              ),
            ),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SizedBox(
                width: 100,
                child: Text(
                  entry.key,
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textHint,
                  ),
                ),
              ),
              Expanded(
                child: Text(
                  entry.value,
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textPrimary,
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  String _receiveStatusText(int status) {
    switch (status) {
      case 0:
        return '领取';
      case 1:
        return '可使用';
      case 2:
        return '已领完';
      case 3:
        return '未开始';
      case 4:
        return '已过期';
      default:
        return '已领取';
    }
  }

  bool _canTapCouponAction(int status) {
    return status == 0 || status == 1;
  }

  Color _couponActionBackgroundColor(int status) {
    if (status == 0) {
      return AppColors.price;
    }
    if (status == 1) {
      return AppColors.price.withValues(alpha: 0.08);
    }
    return Colors.grey[300]!;
  }

  Color _couponActionForegroundColor(int status) {
    if (status == 1) {
      return AppColors.price;
    }
    return Colors.white;
  }

  void _showCouponUsageHint() {
    if (product == null) {
      return;
    }
    showModalBottomSheet<void>(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (sheetContext) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(20, 18, 20, 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '这张券已经领取成功',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  '下单结算时系统会自动展示可用优惠券。你可以现在立即购买，也可以先加入购物车后再统一结算。',
                  style: TextStyle(
                    fontSize: 14,
                    height: 1.5,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 20),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                        onPressed: () {
                          Navigator.of(sheetContext).pop();
                          _addCart(product!);
                        },
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          foregroundColor: AppColors.primary,
                          side: const BorderSide(color: AppColors.primary),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(24),
                          ),
                        ),
                        child: const Text('加入购物车'),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: ElevatedButton(
                        onPressed: () {
                          Navigator.of(sheetContext).pop();
                          _buyNow();
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.primary,
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(24),
                          ),
                          elevation: 0,
                        ),
                        child: const Text('立即购买'),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  void _handleCouponAction(CouponList coupon, BuildContext actionContext) {
    if (coupon.receiveStatus == 0) {
      _claimCoupon(coupon.id);
      return;
    }
    if (coupon.receiveStatus == 1) {
      if (Navigator.of(actionContext).canPop()) {
        Navigator.of(actionContext).pop();
      }
      _showCouponUsageHint();
    }
  }

  void _claimCoupon(int couponId) async {
    try {
      Response result =
          await HttpUtil.post(addCouponUrl, data: {"couponId": couponId});
      Map<String, dynamic> resp = result.data;
      if (resp["code"] == 0) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
                content: Text(resp["message"] ?? "领取成功"),
                backgroundColor: Colors.green),
          );
          refreshProductDetail();
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
                content: Text(resp["message"] ?? "领取失败"),
                backgroundColor: Colors.red),
          );
        }
      }
    } on DioException catch (e) {
      String msg = "领取失败";
      if (e.response?.data != null && e.response!.data is Map) {
        msg = e.response!.data["message"] ?? msg;
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    }
  }

  // 优惠券列表
  Widget buildCouponList(ScrollController controller) {
    if (couponList.isEmpty) {
      return ListView(
        controller: controller,
        shrinkWrap: true,
        children: [
          Container(
            padding: const EdgeInsets.all(20),
            alignment: Alignment.center,
            child: Text(
              '暂无可用优惠券',
              style: TextStyle(
                fontSize: 14,
                color: AppColors.textHint,
              ),
            ),
          ),
        ],
      );
    }

    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: couponList.length,
      itemBuilder: (context, index) {
        final coupon = couponList[index];
        final endStr =
            '${coupon.endTime.year}-${coupon.endTime.month.toString().padLeft(2, '0')}-${coupon.endTime.day.toString().padLeft(2, '0')}';
        return Container(
          padding: const EdgeInsets.all(15),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                width: 1,
                color: AppColors.surfaceMuted,
              ),
            ),
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
                          coupon.name,
                          style: const TextStyle(
                            fontSize: 16,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        Text(
                          '有效期至$endStr',
                          style: const TextStyle(
                            fontSize: 12,
                            color: AppColors.textHint,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Column(
                    children: [
                      Row(
                        children: [
                          Text(
                            "￥",
                            style: const TextStyle(
                              fontSize: 17,
                              color: AppColors.price,
                            ),
                          ),
                          Text(
                            '${coupon.amount}',
                            style: const TextStyle(
                              fontSize: 22,
                              color: AppColors.price,
                            ),
                          ),
                        ],
                      ),
                      Text(
                        coupon.minAmount > 0 ? '满${coupon.minAmount}可用' : '无门槛',
                        style: const TextStyle(
                          fontSize: 13,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(width: 8),
                  SizedBox(
                    height: 28,
                    child: ElevatedButton(
                      onPressed: _canTapCouponAction(coupon.receiveStatus)
                          ? () => _handleCouponAction(coupon, context)
                          : null,
                      style: ElevatedButton.styleFrom(
                        backgroundColor:
                            _couponActionBackgroundColor(coupon.receiveStatus),
                        foregroundColor:
                            _couponActionForegroundColor(coupon.receiveStatus),
                        padding: const EdgeInsets.symmetric(horizontal: 10),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(14)),
                      ),
                      child: Text(
                        _receiveStatusText(coupon.receiveStatus),
                        style: const TextStyle(fontSize: 12),
                      ),
                    ),
                  ),
                ],
              ),
              if (coupon.description.isNotEmpty) ...[
                Container(height: 5),
                Text(
                  coupon.description,
                  style: TextStyle(
                    fontSize: 12,
                    color: AppColors.textHint,
                  ),
                ),
              ],
            ],
          ),
        );
      },
    );
  }

  // 促销活动
  Widget buildPromotion(
    String title,
    String value,
    ScrollController controller,
  ) {
    final items = <Widget>[];

    if (productFullReductionList.isNotEmpty) {
      items.add(_buildPromotionSection(
          '满减优惠',
          productFullReductionList.map((fr) {
            return '满${fr.fullPrice}元减${fr.reducePrice}元';
          }).toList()));
    }

    if (productLadderList.isNotEmpty) {
      items.add(_buildPromotionSection(
          '阶梯价格',
          productLadderList.map((ld) {
            return '满${ld.count}件，折后¥${ld.price}';
          }).toList()));
    }

    if (memberPriceList.isNotEmpty) {
      items.add(_buildPromotionSection(
          '会员专享',
          memberPriceList.map((mp) {
            return '${mp.memberLevelName}：¥${mp.memberPrice}';
          }).toList()));
    }

    if (items.isEmpty) {
      return ListView(
        controller: controller,
        shrinkWrap: true,
        children: [
          Container(
            padding: const EdgeInsets.all(20),
            alignment: Alignment.center,
            child: Text(
              '暂无促销活动',
              style: TextStyle(
                fontSize: 14,
                color: AppColors.textHint,
              ),
            ),
          ),
        ],
      );
    }

    return ListView(
      controller: controller,
      shrinkWrap: true,
      children: items,
    );
  }

  Widget _buildPromotionSection(String title, List<String> details) {
    return Container(
      padding: const EdgeInsets.all(15),
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(
            width: 1,
            color: AppColors.surfaceMuted,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
            decoration: BoxDecoration(
              color: AppColors.price.withValues(alpha: 0.10),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              title,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: AppColors.price,
              ),
            ),
          ),
          const SizedBox(height: 8),
          ...details.map((d) => Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: Text(
                  d,
                  style: TextStyle(
                    fontSize: 14,
                    color: AppColors.textPrimary,
                  ),
                ),
              )),
        ],
      ),
    );
  }

  // 商家服务
  ListView buildMerchantServices(
    String title,
    String value,
    ScrollController controller,
  ) {
    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: 1,
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(15),
          margin: const EdgeInsets.only(bottom: 50),
          child: Text(
            "无忧退货 · 快速退款 · 免费包邮",
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textSecondary,
            ),
          ),
        );
      },
    );
  }

  //添加商品到购物车（带幂等键 + 按钮防抖 + 错误码映射）
  void _addCart(Product product) async {
    final sku = _resolvePurchaseSku();
    if (sku == null) {
      return;
    }

    // 按钮防抖：防止快速双击
    if (_isAdding) return;
    setState(() => _isAdding = true);

    try {
      final skuId = sku.id;
      // 幂等键：productSkuId + 时间戳
      final idempotencyKey =
          '${skuId}_${DateTime.now().millisecondsSinceEpoch}';
      final Map<String, String> headers = {
        'X-Idempotency-Key': idempotencyKey,
      };

      Map<String, dynamic> addCartParams = <String, dynamic>{};
      addCartParams["productId"] = product.id;
      addCartParams["productSkuId"] = skuId;
      addCartParams["quantity"] = 1;
      addCartParams["price"] = sku.price.toDouble();
      addCartParams["productPic"] =
          sku.mainPic.isNotEmpty ? sku.mainPic : product.mainPic;
      addCartParams["productName"] = product.name;
      addCartParams["productSubTitle"] = product.subTitle;
      addCartParams["productSkuCode"] = sku.skuCode;
      addCartParams["productCategoryId"] = product.categoryId;
      addCartParams["productBrand"] = product.brandName;
      addCartParams["productSn"] = product.productSn;
      addCartParams["memberNickname"] = "test";
      addCartParams["productAttr"] = sku.specData;

      final result = await HttpUtil.postWithHeaders(
        cartAddUrl,
        data: addCartParams,
        headers: headers,
      );

      if (!mounted) return;
      final resp = result.data as Map<String, dynamic>;
      if (resp["code"] == 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: const Text('已加入购物车'),
            duration: const Duration(seconds: 1),
            action: SnackBarAction(
              label: '查看',
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute(builder: (context) => const Cart()),
                );
              },
            ),
          ),
        );
      } else {
        // 后端返回错误码，映射为用户可理解文案
        final errorCode = resp["code"]?.toString() ?? '';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text(_mapCartErrorCode(errorCode, resp)),
              backgroundColor: Colors.red),
        );
      }
    } on DioException catch (e) {
      if (!mounted) return;
      // Dio 网络层错误（如 500/网络不可达）
      String msg = '添加失败，请稍后重试';
      if (e.response?.data != null && e.response!.data is Map) {
        msg = _mapCartErrorCode(
            e.response!.data["code"]?.toString() ?? '', e.response!.data);
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg), backgroundColor: Colors.red),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
            content: Text('添加失败，请稍后重试'), backgroundColor: Colors.red),
      );
    } finally {
      if (mounted) setState(() => _isAdding = false);
    }
  }

  // 后端错误码 → 用户可理解文案
  // 优先使用后端返回的具体 message（如库存不足时后端已含具体数量），
  // code 仅用于兜底未知错误和日志追踪
  String _mapCartErrorCode(String code, Map<String, dynamic> resp) {
    // 优先用后端 message（有具体库存数量等上下文信息）
    final backendMsg = resp["message"] ?? '';
    if (backendMsg.isNotEmpty) {
      return backendMsg;
    }

    // 兜底：根据 code 映射
    switch (code) {
      case 'OMS_CART_PRODUCT_OFFLINE':
        return '该商品已下架';
      case 'OMS_CART_PRODUCT_UNVERIFIED':
        return '商品还在审核中，暂不支持购买';
      case 'OMS_CART_STOCK_INSUFFICIENT':
        return '库存不足，请选择其他规格或减少数量';
      case 'OMS_CART_PRODUCT_NOT_FOUND':
        return '商品不存在或已下架';
      default:
        return '添加失败，请稍后重试';
    }
  }
}
