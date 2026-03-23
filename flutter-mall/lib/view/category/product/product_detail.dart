import 'package:bottom_sheet/bottom_sheet.dart';
import 'package:card_swiper/card_swiper.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/cart/cart.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

import '../../../layout/main_tab.dart';
import '../../../model/product_detail.dart';

///
/// 商品详情页面
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class ProductDetail extends StatefulWidget {
  final int productId;

  const ProductDetail({super.key, required this.productId});

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
  ProductVisibility visibility = ProductVisibility.fromJson({});
  bool loading = true;

  @override
  void initState() {
    super.initState();
    queryCollectionList();
  }

  void queryCollectionList() async {
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
        } else {
          product = null;
          productAttributeList = [];
          productAttributeValueList = [];
          skuStockList = [];
          couponList = [];
        }
      });
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
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("详情展示"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
      ),
      body: Stack(
        children: [
          Container(
            color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
            width: MediaQuery.of(context).size.width,
            child: buildPageBody(context),
          ),
          if (!loading && product != null) buildFooter(context),
        ],
      ),
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

  // 商品图片
  SizedBox buildProductPic() {
    List<String> list = [];
    if (product != null) {
      list = product!.albumPics.split(",");
      //如果画册图片为空,则区主图
      if (list[0] == "") {
        list[0] = product!.mainPic;
      }
    }

    return SizedBox(
      height: 360,
      child: Swiper(
        itemBuilder: (BuildContext context, int index) {
          return CachedImageWidget(
            double.infinity,
            double.infinity,
            list[index],
            fit: BoxFit.cover,
          );
        },
        autoplay: true,
        itemCount: list.length,
        pagination: const SwiperPagination(),
        viewportFraction: 0.8,
        scale: 0.8,
      ),
    );
  }

  // 商品基本信息
  Container buildProductBaseInfo() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.all(15),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            product!.name,
            style: TextStyle(
              fontSize: 16,
              color: Color(int.parse('303133', radix: 16)).withAlpha(255),
            ),
          ),
          Text(
            product!.subTitle,
            style: TextStyle(
              fontSize: 14,
              color: Color(int.parse('909399', radix: 16)).withAlpha(255),
            ),
          ),
          if (!visibility.purchasable && visibility.reasonMessage.isNotEmpty)
            Container(
              margin: const EdgeInsets.only(top: 8),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
              decoration: BoxDecoration(
                color: Color(int.parse('fff5f5', radix: 16)).withAlpha(255),
                borderRadius: const BorderRadius.all(Radius.circular(20)),
              ),
              child: Text(
                visibility.reasonMessage,
                style: TextStyle(
                  fontSize: 12,
                  color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                ),
              ),
            ),
          SizedBox(
            height: 32,
            child: Row(
              children: [
                Text(
                  "¥",
                  style: TextStyle(
                    fontSize: 15,
                    color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                  ),
                ),
                Text(
                  product!.price.toString(),
                  style: TextStyle(
                    fontSize: 17,
                    color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                  ),
                ),
              ],
            ),
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                "销量: ${product!.sales}",
                style: TextStyle(
                  fontSize: 14,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
              Text(
                "库存: ${product!.stock}",
                style: TextStyle(
                  fontSize: 14,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
              Text(
                "浏览量: 768",
                style: TextStyle(
                  fontSize: 14,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // 分享
  Container buildProductShare() {
    // var border = BorderSide(width: 1, color: Color(int.parse('fa436a', radix: 16)).withAlpha(255));
    var boxDecoration = BoxDecoration(
      border: Border.all(
        width: 1,
        color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
      ),
      borderRadius: const BorderRadius.all(Radius.circular(5)),
    );
    return Container(
      height: 40,
      margin: const EdgeInsets.symmetric(horizontal: 15),
      child: Row(
        children: [
          Container(
            decoration: boxDecoration,
            child: Row(
              children: [
                Image.asset("images/five.png", height: 16, width: 17),
                Text(
                  "返",
                  style: TextStyle(
                    fontSize: 14,
                    color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Text(
              " 该商品分享可领49减10红包",
              style: TextStyle(
                fontSize: 14,
                color: Color(int.parse('606266', radix: 16)).withAlpha(255),
              ),
            ),
          ),
          Text(
            "立即分享",
            style: TextStyle(
              fontSize: 14,
              color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
            ),
          ),
          Image.asset(
            "images/right_arrow1.png",
            height: 13,
            width: 13,
            color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
          ),
        ],
      ),
    );
  }

  // 商品属性和规格
  Container buildAttributesInfo(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(15),
      color: Colors.white,
      child: Column(
        children: [
          buildTxt("购买类型", "银色 64G", 1, context),
          buildTxt("商品参数", "查看", 2, context),
          buildTxt("优惠券   ", "领取优惠券", 3, context),
          buildTxt("促销活动", "暂无优惠", 4, context),
          buildTxt("商家服务", "无忧退货 · 快速退款 · 免费包邮 ·", 5, context),
        ],
      ),
    );
  }

  InkWell buildTxt(String title, String value, int flag, BuildContext context) {
    var border = BorderSide(
      width: 1,
      color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
    );
    var boxDecoration = BoxDecoration(border: Border(bottom: border));
    return InkWell(
      onTap: () {
        _openBottomSheetWithInfo(context, title);
      },
      child: Container(
        decoration: boxDecoration,
        height: 47,
        child: Row(
          children: [
            Text(
              title,
              style: TextStyle(
                fontSize: 14,
                color: Color(int.parse('606266', radix: 16)).withAlpha(255),
              ),
            ),
            const SizedBox(width: 20),
            Expanded(
              child: Text(
                value,
                style: TextStyle(
                  fontSize: 14,
                  color: Color(
                    int.parse(flag == 3 ? 'fa436a' : "303133", radix: 16),
                  ).withAlpha(255),
                ),
              ),
            ),
            Image.asset("images/right_arrow1.png", height: 13, width: 13),
          ],
        ),
      ),
    );
  }

  // 评价信息
  Container buildPinJiaInfo() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.all(15),
      margin: const EdgeInsets.only(top: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            height: 40,
            child: Row(
              children: [
                Text(
                  "评价",
                  style: TextStyle(
                    fontSize: 15,
                    color: Color(int.parse('303133', radix: 16)).withAlpha(255),
                  ),
                ),
                Expanded(
                  child: Text(
                    "(86)",
                    style: TextStyle(
                      fontSize: 14,
                      color: Color(
                        int.parse('909399', radix: 16),
                      ).withAlpha(255),
                    ),
                  ),
                ),
                Text(
                  "好评率 100% ",
                  style: TextStyle(
                    fontSize: 14,
                    color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                  ),
                ),
                Image.asset("images/right_arrow1.png", height: 15, width: 15),
              ],
            ),
          ),
          Text(
            "koobe",
            style: TextStyle(
              fontSize: 14,
              color: Color(int.parse('606266', radix: 16)).withAlpha(255),
              fontWeight: FontWeight.bold,
            ),
          ),
          Container(
            margin: const EdgeInsets.symmetric(vertical: 10),
            child: Text(
              "商品收到了，79元两件，质量不错，试了一下有点瘦，但是加个外罩很漂亮，我很喜欢",
              style: TextStyle(
                fontSize: 14,
                color: Color(int.parse('303133', radix: 16)).withAlpha(255),
              ),
            ),
          ),
          Row(
            children: [
              Expanded(
                child: Text(
                  "购买类型：XL 红色",
                  style: TextStyle(
                    fontSize: 14,
                    color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                  ),
                ),
              ),
              Text(
                "2019-04-01 19:21",
                style: TextStyle(
                  fontSize: 14,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // 品牌信息
  Container buildBrandInfo() {
    return Container(
      width: MediaQuery.of(context).size.width,
      color: Colors.white,
      margin: const EdgeInsets.symmetric(vertical: 8),
      padding: const EdgeInsets.symmetric(vertical: 25, horizontal: 25),
      child: Column(
        children: [
          SizedBox(
            height: 40,
            child: Text(
              "品牌信息",
              style: TextStyle(
                fontSize: 15,
                color: Color(int.parse('303133', radix: 16)).withAlpha(255),
              ),
            ),
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.start,
            children: [
              CachedImageWidget(105, 35, brand.logo, fit: BoxFit.contain),
              const SizedBox(width: 20),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    brand.name,
                    style: TextStyle(
                      fontSize: 18,
                      color: Color(
                        int.parse('303133', radix: 16),
                      ).withAlpha(255),
                    ),
                  ),
                  Text(
                    "品牌首字母：${brand.firstLetter}",
                    style: TextStyle(
                      fontSize: 12,
                      color: Color(
                        int.parse('909399', radix: 16),
                      ).withAlpha(255),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  // 图文详情
  Column buildImageDetailInfo() {
    List<String> imageUrls = [
      'https://img12.360buyimg.com/cms/jfs/t1/75040/28/21026/295081/634ed154E981e4d10/2ceef91d6f2b2273.jpg!q70.dpg.webp',
      'https://img13.360buyimg.com/cms/jfs/t1/191028/1/28802/401291/634ed15eEb234dc40/5ab89f83531e1023.jpg!q70.dpg.webp',
      'https://img14.360buyimg.com/cms/jfs/t1/32758/24/18599/330590/634ed15eEc3db173c/c52953dc8008a576.jpg!q70.dpg.webp',
    ];
    return Column(
      children: [
        Container(
          width: MediaQuery.of(context).size.width,
          alignment: Alignment.center,
          height: 40,
          color: Colors.white,
          child: Text(
            "图文详情",
            style: TextStyle(
              fontSize: 15,
              color: Color(int.parse('303133', radix: 16)).withAlpha(255),
            ),
          ),
        ),
        Column(
          children: imageUrls.map((url) {
            return Image.network(kIsWeb ? proxyImageUrl(url) : url);
          }).toList(),
        ),
      ],
    );
  }

  // 底部悬浮
  Positioned buildFooter(BuildContext context) {
    bool disabled = !visibility.purchasable;
    return Positioned(
      bottom: 0,
      width: MediaQuery.of(context).size.width,
      child: Container(
        height: 60,
        padding: const EdgeInsets.symmetric(horizontal: 15),
        margin: const EdgeInsets.all(15),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(
            color: Colors.grey, //边框颜色
            width: 1, //边框宽度
          ),
          boxShadow: [
            BoxShadow(
              blurRadius: 3, //阴影范围
              spreadRadius: 1, //阴影浓度
              color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
            ),
            //阴影颜色
          ],
          borderRadius: BorderRadius.circular(10), // 圆角也可控件一边圆角大小
        ),
        child: Row(
          // mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              child: Row(
                children: [
                  buildImage("images/home.png", "首页"),
                  buildImage("images/cart.png", "购物车"),
                  buildImage("images/love1.png", "收藏"),
                  // buildImage("images/love1.png", context.watch<Counter>().count.toString()),
                ],
              ),
            ),
            Container(
              height: 40,
              width: 100,
              decoration: BoxDecoration(
                color: disabled
                    ? Color(int.parse('c0c4cc', radix: 16)).withAlpha(255)
                    : Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                borderRadius: const BorderRadius.all(Radius.circular(50)),
                boxShadow: const [
                  BoxShadow(
                    blurRadius: 2, //阴影范围
                    spreadRadius: 1, //阴影浓度
                    color: Colors.grey, //阴影颜色
                  ),
                ],
              ),
              child: TextButton(
                onPressed: disabled
                    ? null
                    : () {
                        _addCart(product!);
                      },
                child: Text(
                  disabled
                      ? (visibility.reasonMessage.isNotEmpty
                          ? visibility.reasonMessage
                          : '暂不可购买')
                      : '加入购物车',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                  ),
                ),
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
              color: Color(int.parse('c0c4cc', radix: 16)).withAlpha(255),
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
                color: Color(int.parse('303133', radix: 16)).withAlpha(255),
              ),
            ),
            if (visibility.recoveryHint.isNotEmpty) ...[
              const SizedBox(height: 10),
              Text(
                visibility.recoveryHint,
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => _handleFallbackAction(context),
                style: ElevatedButton.styleFrom(
                  backgroundColor:
                      Color(int.parse('fa436a', radix: 16)).withAlpha(255),
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

  InkWell buildImage(String url, String title) {
    return InkWell(
      onTap: () {
        if (title.contains("收藏")) {
          return;
        }
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) {
              if (title.contains("首页")) {
                return const MainTab();
              } else {
                return const Cart();
              }
            },
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(right: 25),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Image.asset(
              url,
              color: Color(int.parse('909399', radix: 16)).withAlpha(255),
              height: 21,
              width: 21,
            ),
            Text(
              title,
              style: TextStyle(
                fontSize: 12,
                color: Color(int.parse('707070', radix: 16)).withAlpha(255),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _openBottomSheetWithInfo(BuildContext context, String title) {
    showFlexibleBottomSheet<void>(
      isExpand: false,
      initHeight: 0.8,
      maxHeight: 0.8,
      context: context,
      builder: (context, controller, offset) {
        if (title.contains("购买类型")) {
          return buildBuyType("title", "value", controller);
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

  // 购买类型
  ListView buildBuyType(
    String title,
    String value,
    ScrollController controller,
  ) {
    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: 5,
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(15),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                width: 5,
                color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
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
                          "全品类通用券",
                          style: TextStyle(
                            fontSize: 16,
                            color: Color(
                              int.parse('303133', radix: 16),
                            ).withAlpha(255),
                          ),
                        ),
                        Text(
                          "有效期至2023-11-30 00:00:00",
                          style: TextStyle(
                            fontSize: 12,
                            color: Color(
                              int.parse('909399', radix: 16),
                            ).withAlpha(255),
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
                            style: TextStyle(
                              fontSize: 17,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                          Text(
                            "10",
                            style: TextStyle(
                              fontSize: 22,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                        ],
                      ),
                      Text(
                        "满100可用",
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(
                            int.parse('707070', radix: 16),
                          ).withAlpha(255),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Container(height: 5),
              Text(
                "全场通用",
                style: TextStyle(
                  fontSize: 12,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  // 商品参数
  ListView buildProductParams(
    String title,
    String value,
    ScrollController controller,
  ) {
    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: 10,
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(15),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                width: 5,
                color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
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
                          "全品类通用券",
                          style: TextStyle(
                            fontSize: 16,
                            color: Color(
                              int.parse('303133', radix: 16),
                            ).withAlpha(255),
                          ),
                        ),
                        Text(
                          "有效期至2023-11-30 00:00:00",
                          style: TextStyle(
                            fontSize: 12,
                            color: Color(
                              int.parse('909399', radix: 16),
                            ).withAlpha(255),
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
                            style: TextStyle(
                              fontSize: 17,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                          Text(
                            "10",
                            style: TextStyle(
                              fontSize: 22,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                        ],
                      ),
                      Text(
                        "满100可用",
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(
                            int.parse('707070', radix: 16),
                          ).withAlpha(255),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Container(height: 5),
              Text(
                "全场通用",
                style: TextStyle(
                  fontSize: 12,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  // 优惠券列表
  ListView buildCouponList(ScrollController controller) {
    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: 10,
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(15),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                width: 5,
                color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
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
                          "全品类通用券",
                          style: TextStyle(
                            fontSize: 16,
                            color: Color(
                              int.parse('303133', radix: 16),
                            ).withAlpha(255),
                          ),
                        ),
                        Text(
                          "有效期至2023-11-30 00:00:00",
                          style: TextStyle(
                            fontSize: 12,
                            color: Color(
                              int.parse('909399', radix: 16),
                            ).withAlpha(255),
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
                            style: TextStyle(
                              fontSize: 17,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                          Text(
                            "10",
                            style: TextStyle(
                              fontSize: 22,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                        ],
                      ),
                      Text(
                        "满100可用",
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(
                            int.parse('707070', radix: 16),
                          ).withAlpha(255),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Container(height: 5),
              Text(
                "全场通用",
                style: TextStyle(
                  fontSize: 12,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  // 促销活动
  ListView buildPromotion(
    String title,
    String value,
    ScrollController controller,
  ) {
    return ListView.builder(
      controller: controller,
      shrinkWrap: true,
      itemCount: 10,
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(15),
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(
                width: 5,
                color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
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
                          "全品类通用券",
                          style: TextStyle(
                            fontSize: 16,
                            color: Color(
                              int.parse('303133', radix: 16),
                            ).withAlpha(255),
                          ),
                        ),
                        Text(
                          "有效期至2023-11-30 00:00:00",
                          style: TextStyle(
                            fontSize: 12,
                            color: Color(
                              int.parse('909399', radix: 16),
                            ).withAlpha(255),
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
                            style: TextStyle(
                              fontSize: 17,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                          Text(
                            "10",
                            style: TextStyle(
                              fontSize: 22,
                              color: Color(
                                int.parse('fa436a', radix: 16),
                              ).withAlpha(255),
                            ),
                          ),
                        ],
                      ),
                      Text(
                        "满100可用",
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(
                            int.parse('707070', radix: 16),
                          ).withAlpha(255),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Container(height: 5),
              Text(
                "全场通用",
                style: TextStyle(
                  fontSize: 12,
                  color: Color(int.parse('909399', radix: 16)).withAlpha(255),
                ),
              ),
            ],
          ),
        );
      },
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
            "无忧退货 · 快速退款 · 免费包邮 ·",
            style: TextStyle(
              fontSize: 14,
              color: Color(int.parse('707070', radix: 16)).withAlpha(255),
            ),
          ),
        );
      },
    );
  }

  //添加商品到购物车
  void _addCart(Product product) async {
    Map<String, dynamic> addCartParams = <String, dynamic>{};
    addCartParams["productId"] = product.id;
    addCartParams["productSkuId"] = 221;
    addCartParams["quantity"] = 1;
    addCartParams["price"] = double.parse(product.price);
    addCartParams["productPic"] = product.mainPic;
    addCartParams["productName"] = product.name;
    addCartParams["productSubTitle"] = product.subTitle;
    addCartParams["productSkuCode"] = "202211040040001";
    addCartParams["productCategoryId"] = product.categoryId;
    addCartParams["productBrand"] = product.brandName;
    addCartParams["productSn"] = product.productSn;
    addCartParams["memberNickname"] = "test";
    addCartParams["productAttr"] =
        "[{\"key\":\"颜色\",\"value\":\"黑色\"},{\"key\":\"容量\",\"value\":\"128G\"}]";
    await HttpUtil.post(cartAddUrl, data: addCartParams);

    // LoginModel loginModel = LoginModel.fromJson(result.data);
    //
    // //保存登录凭证token
    // SharedPreferencesUtil.saveString(token, "${loginModel.data.tokenHead} ${loginModel.data.token}");
  }
}
