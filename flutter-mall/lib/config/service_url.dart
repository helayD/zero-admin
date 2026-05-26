// 接口地址配置
// 作者：David
// 日期：2023/11/21 17:17

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
// 首页广告点击
const recordHomeAdvertiseClickUrl =
    "$baseUrl/api/home/recordHomeAdvertiseClick";
// 品牌列表
const brandListDataUrl = "$baseUrl/api/product/queryBrandList";
// 品牌详情
const brandDetailDataUrl = "$baseUrl/api/product/queryBrandDetail?brandId=";
// 分类
const categoriesDataUrl = "$baseUrl/api/product/queryProductCateList";
// 购物车
const cartDataUrl = "$baseUrl/api/order/queryCarItemList";
// 购物车促销试算
const cartPromotionUrl = "$baseUrl/api/order/queryPromotionList";
// 清空购物车
const clearCartUrl = "$baseUrl/api/order/clear";
// 删除购物车商品
const deleteCartUrl = "$baseUrl/api/order/deleteCartItem";
// 修改购物车商品数量
const updateCartQuantityUrl = "$baseUrl/api/order/updateCartItemQuantity";
// 批量结算前商品有效性校验
const validateCartItemsUrl = "$baseUrl/api/order/validateCartItems";
// 生成确认单
const generateConfirmOrderUrl = "$baseUrl/api/order/generateConfirmOrder";
// 提交订单（Story 5-3 HIGH-4）
const generateOrderUrl = "$baseUrl/api/order/generateOrder";
// 添加商品进购物车
const cartAddUrl = "$baseUrl/api/order/addCart";
// 商品列表
const productListQueryUrl = "$baseUrl/api/product/queryProductList";
const productListDataUrl =
    "$baseUrl/api/product/queryProductList?productCategoryId=";
// 商品详情
const productDetailDataUrl =
    "$baseUrl/api/product/queryProductDetail?productId=";
// 站内消息
const memberMessageBaseUrl = "$baseUrl/api/member/message";
// 通知消息列表（全部）
const messageListDataUrl = "$memberMessageBaseUrl/list";
// 消息详情
String messageDetailDataUrl(int messageId) =>
    "$memberMessageBaseUrl/$messageId";
// 标记单条消息已读
String messageReadDataUrl(int messageId) =>
    "$memberMessageBaseUrl/$messageId/read";
// 兼容旧版标记单条消息已读入口
const messageReadUrl = "$memberMessageBaseUrl/read";
// 标记全部已读
const markAllReadUrl = "$memberMessageBaseUrl/readAll";
// 查询未读消息数
const unreadCountUrl = "$memberMessageBaseUrl/unreadCount";
// 删除消息
String messageDeleteDataUrl(int messageId) =>
    "$memberMessageBaseUrl/$messageId";
// 兼容旧版删除消息入口
const messageDeleteUrl = "$memberMessageBaseUrl/delete";
// 积分记录
const pointsLogListUrl = "$baseUrl/api/member/points/list";
// 优惠券
const couponDataUrl = "$baseUrl/api/member/coupon/queryCouponList?useStatus=";
// 可领取优惠券列表
const availableCouponUrl = "$baseUrl/api/member/coupon/queryAvailableCoupons";
// 领取优惠券
const addCouponUrl = "$baseUrl/api/member/coupon/addCoupon";
// 订单列表
const orderListDataUrl = "$baseUrl/api/order/queryOrderList?status=";
// 订单详情
const orderDetailDataUrl = "$baseUrl/api/order/queryOrderDetail?orderId=";
// 取消订单（Story 6.2）
const cancelOrderUrl = "$baseUrl/api/order/cancelUserOrder?orderId=";
// 确认收货（Story 6.2）
const confirmReceiveUrl = "$baseUrl/api/order/confirmReceiveOrder?orderId=";
// 物流查询（Story 6.3）
const logisticsDataUrl = "$baseUrl/api/order/queryLogistics?orderId=";
// 售后原因列表（Story 6.4）
const queryReturnReasonListUrl = "$baseUrl/api/order/queryReturnReasonList";
// 提交售后申请（Story 6.4）
const applyAfterSalesUrl = "$baseUrl/api/order/applyAfterSales";
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
// 添加收藏
const addCollectionDataUrl = "$baseUrl/api/member/addtCollection";
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
// Story 3.1.1: 手机号 + 短信验证码合并登录注册
//   - sendSmsCodeUrl：发送短信验证码（mock 模式固定 123456，前期调试用）
//   - smsLoginUrl：验证码合并登录注册接口（已注册→直接登录；未注册→自动建号并登录）
// 旧的 /api/member/login（手机号+密码登录）与 /api/member/register（手机号+密码注册）已下线。
const sendSmsCodeUrl = "$baseUrl/api/member/auth/sms/send";
const smsLoginUrl = "$baseUrl/api/member/auth/login";
// 获取用户信息
const memberInfoDataUrl = "$baseUrl/api/member/info";
// 更新会员信息
const updateMemberDataUrl = "$baseUrl/api/member/updateMember";
// 提货卡活动落地页（匿名）
const queryDrawActivityLandingUrl =
    "$baseUrl/api/digitalCard/queryDrawActivityLanding";
// 提货卡活动落地页（登录态）
const queryMyDrawActivityLandingUrl =
    "$baseUrl/api/digitalCard/queryMyDrawActivityLanding";
// 预检抽卡资格
const previewDrawEligibilityUrl =
    "$baseUrl/api/digitalCard/previewDrawEligibility";
// 参与抽卡
const participateDrawUrl = "$baseUrl/api/digitalCard/participateDraw";
// 查询我的抽卡记录
const queryMyDrawRecordListUrl =
    "$baseUrl/api/digitalCard/queryMyDrawRecordList";
// 查询我的提货卡资产列表
const queryMyDigitalCardAssetListUrl =
    "$baseUrl/api/digitalCard/asset/queryMyDigitalCardAssetList";
// 查询我的提货卡资产详情
const queryMyDigitalCardAssetDetailUrl =
    "$baseUrl/api/digitalCard/asset/queryMyDigitalCardAssetDetail";
// 识别提货卡转赠接收人
const resolveDigitalCardTransferRecipientUrl =
    "$baseUrl/api/digitalCard/asset/resolveTransferRecipient";
// 转赠提货卡
const transferDigitalCardAssetUrl =
    "$baseUrl/api/digitalCard/asset/transferDigitalCardAsset";
// 提交提货卡提现申请
const requestDigitalCardWithdrawUrl =
    "$baseUrl/api/digitalCard/asset/requestDigitalCardWithdraw";
// 查询我的实体卡履约详情
const queryMyPhysicalFulfillmentDetailUrl =
    "$baseUrl/api/digitalCard/physicalFulfillment/queryMyPhysicalFulfillmentDetail";
// 确认实体卡收货地址
const confirmPhysicalFulfillmentAddressUrl =
    "$baseUrl/api/digitalCard/physicalFulfillment/confirmPhysicalFulfillmentAddress";
// 确认实体卡邮费支付
const confirmPhysicalFulfillmentShippingFeeUrl =
    "$baseUrl/api/digitalCard/physicalFulfillment/confirmPhysicalFulfillmentShippingFee";
// 确认实体卡签收
const confirmPhysicalCardReceiptUrl =
    "$baseUrl/api/digitalCard/physicalFulfillment/confirmPhysicalCardReceipt";

// ==================== 提货卡提货与转赠（Story 10.7）====================
// 创建提货单
const createRedemptionOrderUrl =
    "$baseUrl/api/digitalCard/asset/createRedemptionOrder";
// 查询提货单详情
const queryRedemptionOrderUrl =
    "$baseUrl/api/digitalCard/asset/queryRedemptionOrder";
// 生成分享链接
const generateShareLinkUrl = "$baseUrl/api/digitalCard/asset/generateShareLink";
// 校验分享凭证（匿名）
const validateClaimTokenUrl = "$baseUrl/api/digitalCard/validateClaimToken";
// 领取分享卡片
const claimDigitalCardUrl = "$baseUrl/api/digitalCard/claim";

// ==================== 支付相关（Story 5.5 Task 11）====================
// 发起支付
const orderPayUrl = "$baseUrl/api/order/orderPay";
// 支付状态查询（Flutter 使用 query 参数：?orderId=xxx）
const orderPayQueryUrl = "$baseUrl/api/order/orderPayQueryStatus";
// 统一版本策略
const appVersionPolicyUrl = "$baseUrl/api/app/version/policy";

// ==================== 商品评价相关（Story 8-2 Review Fix R-3）====================
// 提交商品评价（注意：路由注册为 /comment/add，而非 /comment/addComment）
const addCommentUrl = "$baseUrl/api/product/comment/add";
// 查询商品评价列表
const queryCommentListUrl = "$baseUrl/api/product/comment/queryCommentList";
// 查询商品评价详情
const queryCommentDetailUrl = "$baseUrl/api/product/comment/queryCommentDetail";
