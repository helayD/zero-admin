///
/// 接口地址
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
///

// const baseUrl = "http://10.168.96.23:9999";
// const baseUrl = "http://127.0.0.1:9999";
const baseUrl = "http://47.107.224.56:9999";
// 图片代理地址（解决 Web 端 CanvasKit 图片跨域问题）
const imageProxyBase = "http://47.107.224.56:8001/image-proxy";

/// 将外部图片 URL 转为通过代理加载
/// 例: http://129.204.203.29/hua_s.jpg -> http://47.107.224.56:8001/image-proxy/http/129.204.203.29/hua_s.jpg
String proxyImageUrl(String url) {
  if (url.isEmpty) return url;
  final uri = Uri.tryParse(url);
  if (uri == null || !uri.hasScheme || !uri.hasAuthority) return url;
  return "$imageProxyBase/${uri.scheme}/${uri.host}${uri.path}${uri.hasQuery ? '?${uri.query}' : ''}";
}
// 首页
const homeDataUrl = "$baseUrl/api/home/index";
// 品牌列表
const brandListDataUrl = "$baseUrl/api/product/queryBrandList";
// 品牌详情
const brandDetailDataUrl = "$baseUrl/api/product/queryBrandDetail?brandId=";
// 分类
const categoriesDataUrl = "$baseUrl/api/product/queryProductCateList";
// 购物车
const cartDataUrl = "$baseUrl/api/order/queryCarItemList";
// 添加商品进购物车
const cartAddUrl = "$baseUrl/api/order/addCart";
// 商品列表
const productListDataUrl = "$baseUrl/api/product/queryProductList?productCategoryId=";
// 商品详情
const productDetailDataUrl = "$baseUrl/api/product/queryProductDetail?productId=";
// 通知消息
const messageListDataUrl = "$baseUrl/api/member/message/list/";
// 优惠券
const couponDataUrl = "$baseUrl/api/member/coupon/queryCouponList?useStatus=";
// 订单列表
const orderListDataUrl = "$baseUrl/api/order/queryOrderList?status=";
// 订单详情
const orderDetailDataUrl = "$baseUrl/api/order/queryOrderDetail?orderId=";
// 收货地址列表
const addressListDataUrl = "$baseUrl/api/member/queryAddressList";
// 添加会员地址
const addAddressDataUrl = "$baseUrl/api/member/addAddress";
// 删除会员地址
const deleteAddressDataUrl = "$baseUrl/api/member/deleteAddress";
// 更新会员地址
const updateAddressDataUrl = "$baseUrl/api/member/updateAddress";
// 查询地址详情
const queryAddressDetailDataUrl = "$baseUrl/api/member/querAddressDetail";
// 更新地址默认状态
const updateAddressStatusDataUrl = "$baseUrl/api/member/updateAddressStatus";
// 我的足迹
const historyListDataUrl = "$baseUrl/api/member/queryReadHistoryList";
// 删除足迹
const deleteReadHistoryDataUrl = "$baseUrl/api/member/deleteReadHistory";
// 清空足迹
const clearReadHistoryDataUrl = "$baseUrl/api/member/clearReadHistory";
// 我的收藏
const collectionListDataUrl = "$baseUrl/api/member/queryCollectionList";
// 删除收藏
const deleteCollectionDataUrl = "$baseUrl/api/member/deleteCollection";
// 清空收藏
const clearCollectionDataUrl = "$baseUrl/api/member/clearCollection";
// 我的关注
const focusOnListDataUrl = "$baseUrl/api/member/queryAttentionList";
// 删除关注
const deleteAttentionDataUrl = "$baseUrl/api/member/deleteAttention";
// 清空关注
const clearAttentionDataUrl = "$baseUrl/api/member/clearAttention";
// 登录
const loginDataUrl = "$baseUrl/api/member/login";
// 注册
const registerDataUrl = "$baseUrl/api/member/register";
// 获取用户信息
const memberInfoDataUrl = "$baseUrl/api/member/info";
// 更新会员信息
const updateMemberDataUrl = "$baseUrl/api/member/updateMember";
