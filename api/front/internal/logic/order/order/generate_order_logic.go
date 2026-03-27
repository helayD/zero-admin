package order

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/logic/member/coupon"
	"github.com/feihua/zero-admin/api/front/internal/logic/order/cart"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// OrderErrorCode 订单错误码（Story 5-3 Task 3）
const (
	ErrCodeOrderNoAddress               = "OMS_ORDER_NO_ADDRESS"
	ErrCodeOrderAddressInvalid          = "OMS_ORDER_ADDRESS_INVALID"
	ErrCodeOrderCouponUnavailable       = "OMS_ORDER_COUPON_UNAVAILABLE"
	ErrCodeOrderIntegrationExceed       = "OMS_ORDER_INTEGRATION_EXCEED"
	ErrCodeOrderIntegrationCouponConflict = "OMS_ORDER_INTEGRATION_COUPON_CONFLICT"
	ErrCodeOrderPayTypeInvalid          = "OMS_ORDER_PAY_TYPE_INVALID"
	ErrCodeOrderStockInsufficient       = "OMS_ORDER_STOCK_INSUFFICIENT"
	ErrCodeOrderSystemError             = "OMS_ORDER_SYSTEM_ERROR"
)

// GenerateOrderLogic
/*
Author: LiuFeiHua
Date: 2023/12/12 18:04
Updated by Story 5-3: Complete order generation with address, coupon, integration, and pay type.
*/
type GenerateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateOrderLogic {
	return &GenerateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// genOrderNo 生成订单编号（格式：OR + yyyyMMddHHmmss + 4位随机数）
func genOrderNo() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("OR%s%04d",
		time.Now().Format("20060102150405"),
		r.Intn(10000))
}

// GenerateOrder 根据提交信息生成订单（Story 5-3 重启注释代码，完整闭环）
// 1.获取购物车及优惠信息
// 2.生成下单商品信息（启用所有字段）
// 3.判断购物车中商品是否都有库存
// 4.判断是否使用了优惠券（重新启用）
// 5.判断是否使用积分（重新启用）
// 6.计算order_item的实付金额（重新启用）
// 7.进行库存锁定
// 8.计算应付金额
// 9.校验收货地址
// 10.转化订单信息并插入数据库
// 11.保存收货地址信息
// 12.如果使用优惠券,更新优惠券使用状态（重新启用）
// 13.如果使用积分,需要扣除积分
// 14.发送延迟消息取消订单
func (l *GenerateOrderLogic) GenerateOrder(req *types.GenerateOrderReq) (*types.GenerateOrderResp, error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// === Task 9: 校验收货地址 ===
	if req.MemberReceiveAddressId <= 0 {
		return nil, errorx.NewDefaultError(ErrCodeOrderNoAddress)
	}
	addressDetail, err := l.svcCtx.MemberAddressService.QueryMemberAddressDetail(l.ctx, &umsclient.QueryMemberAddressDetailReq{
		MemberId: memberId,
		Id:       req.MemberReceiveAddressId,
	})
	if err != nil || addressDetail == nil {
		s, _ := status.FromError(err)
		if err != nil && s.Code() != 0 {
			logc.Errorf(l.ctx, "查询收货地址异常, addressId=%d, err=%s", req.MemberReceiveAddressId, err.Error())
		}
		return nil, errorx.NewDefaultError(ErrCodeOrderAddressInvalid)
	}

	memberInfo, _ := l.svcCtx.MemberService.QueryMemberInfoDetail(l.ctx, &umsclient.QueryMemberInfoDetailReq{MemberId: memberId})
	if memberInfo == nil {
		memberInfo = &umsclient.QueryMemberInfoDetailResp{}
	}

	// 1.获取购物车及优惠信息
	cartPromotionItemList, err := cart.QueryCartListPromotion(req.CartIds, l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if len(cartPromotionItemList) == 0 {
		return result(1, "购物车还没有商品,请先添加商品到购物车!"), nil
	}

	// === Task 2.6: 计算总金额（单位：分 int64）===
	var totalAmount int64 = 0
	var promotionAmountTotal int64 = 0
	for _, item := range cartPromotionItemList {
		totalAmount += int64(item.Price) * int64(item.Quantity)
		promotionAmountTotal += item.ReduceAmount * int64(item.Quantity)
	}

	// 2.生成下单商品信息（启用所有字段）
	var flag = false
	orderItemList := make([]*omsclient.OrderItemData, 0)
	cartItemIds := make([]int64, 0)
	for _, item := range cartPromotionItemList {
		skuTotalAmt := int64(item.Price) * int64(item.Quantity)
		itemPromoAmt := item.ReduceAmount
		orderItem := &omsclient.OrderItemData{
			SkuId:           item.ProductSkuId,
			SkuName:         item.ProductName,
			SkuPic:          item.ProductPic,
			SkuPrice:        float32(item.Price), // int64 → float32
			SkuQuantity:     int32(item.Quantity),
			SpecData:        item.ProductAttr,
			SkuTotalAmount:  float32(skuTotalAmt), // int64 → float32
			PromotionAmount: float32(itemPromoAmt), // int64 → float32
		}
		orderItemList = append(orderItemList, orderItem)

		if item.RealStock <= 0 || item.RealStock < item.Quantity {
			flag = true
		}
		cartItemIds = append(cartItemIds, item.Id)
	}

	// 3.判断购物车中商品是否都有库存
	if flag {
		return nil, errorx.NewDefaultError(ErrCodeOrderStockInsufficient)
	}

	// 优惠券按商品金额（分）比例分摊（全场通用）
		couponAmountTotalFen := int64(0)
		if req.CouponId > 0 {
			enableList, _, err := coupon.QueryCouponList(l.svcCtx, l.ctx, cartPromotionItemList)
			if err != nil {
				return nil, errorx.NewDefaultError(ErrCodeOrderSystemError)
			}
			var selectedCoupon *types.CouponData
			for i := range enableList {
				if enableList[i].Id == req.CouponId {
					selectedCoupon = &enableList[i]
					break
				}
			}
			if selectedCoupon == nil {
				return nil, errorx.NewDefaultError(ErrCodeOrderCouponUnavailable)
			}
			// 金额口径：CouponData.Amount 单位是元（float32），转为分（int64）再参与分摊
			couponAmountFen := int64(selectedCoupon.Amount * 100)
			var applicableTotalFen int64 = 0
			for _, item := range cartPromotionItemList {
				applicableTotalFen += int64(item.Price) * int64(item.Quantity)
			}
			if applicableTotalFen > 0 {
				for _, item := range orderItemList {
					itemCouponFen := couponAmountFen * int64(item.SkuTotalAmount) / applicableTotalFen
					item.CouponAmount = float32(itemCouponFen) // OMS proto 约束，仍写 float32
					couponAmountTotalFen += itemCouponFen
				}
			}
		}

	// === Task 2.3: 积分提交（重新启用，已在原代码中部分实现）===
	var integrationAmountTotal int64 = 0
	if req.UseIntegration > 0 {
		if req.UseIntegration > memberInfo.Points {
			return nil, errorx.NewDefaultError(ErrCodeOrderIntegrationExceed)
		}
		consumeSetting, _ := l.svcCtx.MemberConsumeSettingService.QueryMemberConsumeSettingDetail(l.ctx, &umsclient.QueryMemberConsumeSettingDetailReq{Id: 1})
		if consumeSetting == nil {
			consumeSetting = &umsclient.QueryMemberConsumeSettingDetailResp{}
		}
		// 积分与优惠券互斥判断
		if req.CouponId > 0 && consumeSetting.CouponStatus == 0 {
			return nil, errorx.NewDefaultError(ErrCodeOrderIntegrationCouponConflict)
		}
		// 最低使用门槛
		if req.UseIntegration < consumeSetting.UseUnit {
			return nil, errorx.NewDefaultError(ErrCodeOrderIntegrationExceed)
		}
		// 每积分抵扣金额
		integrationAmount := int64(req.UseIntegration / int32(consumeSetting.DeductionPerAmount))
		// 每单最高抵扣比例
		var maxIntegrationAmount int64 = 0
		if consumeSetting.MaxPercentPerOrder > 0 && totalAmount > 0 {
			maxIntegrationAmount = totalAmount * int64(consumeSetting.MaxPercentPerOrder) / 100
		}
		if integrationAmount > maxIntegrationAmount && maxIntegrationAmount > 0 {
			integrationAmount = maxIntegrationAmount
		}
		// 积分按商品金额比例分摊
		if totalAmount > 0 {
			for _, item := range orderItemList {
				ratio := float64(item.SkuTotalAmount) / float64(totalAmount)
				item.PointsAmount = float32(int64(float64(integrationAmount) * ratio))
				integrationAmountTotal += int64(item.PointsAmount)
			}
		}
	}

	// 6.计算order_item的实付金额（重新启用）
	// 实付金额 = 原价 - 促销优惠 - 优惠券抵扣 - 积分抵扣
	for _, item := range orderItemList {
		realAmt := item.SkuPrice*float32(item.SkuQuantity) - item.PromotionAmount*float32(item.SkuQuantity) - item.CouponAmount - item.PointsAmount
		if realAmt < 0 {
			realAmt = 0
		}
		item.RealAmount = realAmt
	}

	// 7.进行库存锁定
	var stockLockData []*pmsclient.UpdateSkuStockData
	for _, item := range orderItemList {
		stockLockData = append(stockLockData, &pmsclient.UpdateSkuStockData{
			Id:              item.SkuId,
			ProductQuantity: item.SkuQuantity,
		})
	}
	_, err = l.svcCtx.ProductSkuService.LockSkuStockLock(l.ctx, &pmsclient.UpdateSkuStockReq{
		Data: stockLockData,
	})
	if err != nil {
		logc.Errorf(l.ctx, "锁定库存异常,参数: %+v,异常：%s", stockLockData, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// === Task 2.5: 金额计算收口 ===
	// payAmount = totalAmount - promotionAmount - couponAmount - integrationAmount（分）
	payAmount := totalAmount - promotionAmountTotal - couponAmountTotalFen - integrationAmountTotal
	if payAmount < 0 {
		payAmount = 0
	}

	// === Task 2.4: 支付方式校验 ===
	if req.PayType != 1 && req.PayType != 2 {
		return nil, errorx.NewDefaultError(ErrCodeOrderPayTypeInvalid)
	}
	// OMS proto AddOrderReq 无 PayType 字段；实际支付方式由支付回调写入 OMS
	// 前端提交时已做 PayType 校验（1=支付宝，2=微信）

	// === Task 2.1: 收货人信息传递（通过 AddOrderDelivery）===
	orderNo := genOrderNo()

	// 10.转化为订单信息并插入数据库
	orderInfo := &omsclient.AddOrderReq{
		OrderNo:         orderNo,
		UserId:          memberId,
		OrderStatus:     1, // 1-待支付
		TotalAmount:     float32(totalAmount),       // 分→元（float32）
		PromotionAmount: float32(promotionAmountTotal),
		CouponAmount:    float32(couponAmountTotalFen),
		PointsAmount:    float32(integrationAmountTotal),
		DiscountAmount:  0,
		FreightAmount:   0, // MVP 阶段运费写死为 0，Story 5.4 后续实现
		PayAmount:       float32(payAmount),
		SourceType:      1, // 1-APP
		UsePoints:       req.UseIntegration,
		OrderItemData:   orderItemList,
	}

	orderAddResp, err := l.svcCtx.OrderService.AddOrder(l.ctx, orderInfo)
	if err != nil {
		return nil, errorx.NewDefaultError(ErrCodeOrderSystemError)
	}

	// === Task 2.7 & MEDIUM-6: 保存收货人信息到 OrderDelivery ===
	_, err = l.svcCtx.OrderDeliveryService.AddOrderDelivery(l.ctx, &omsclient.AddOrderDeliveryReq{
		OrderId:          orderAddResp.Id,
		OrderNo:          orderNo,
		ReceiverName:     addressDetail.ReceiverName,
		ReceiverPhone:    addressDetail.ReceiverPhone,
		ReceiverProvince: addressDetail.Province,
		ReceiverCity:     addressDetail.City,
		ReceiverDistrict: addressDetail.District,
		ReceiverAddress:  addressDetail.DetailAddress,
	})
	if err != nil {
		logc.Errorf(l.ctx, "保存收货人信息失败, orderId=%d, err=%s", orderAddResp.Id, err.Error())
		// 不阻塞订单创建，记录日志即可
	}

	// === Task 2.2: 优惠券核销 ===
	// 【重要】优惠券核销和积分扣除均在订单创建之后执行，存在分布式事务不一致风险：
	// - 最优方案：使用 Saga 模式或 TCC 事务框架统一编排（Story 5.7 异步链路监控视图处理）
	// - 当前方案（MVP）：订单创建成功后执行后置操作，失败时仅记录日志并返回成功
	//   （理由：订单已不可逆，优惠券/积分的补偿可通过后台对账人工处理）
	//   若后续需要强一致性，需在 Story 5.7 中引入 Saga 编排器。
	if req.CouponId > 0 {
		_, err = l.svcCtx.CouponRecordService.UpdateCouponRecord(l.ctx, &smsclient.UpdateCouponRecordReq{
			CouponIds: []int64{req.CouponId},
			MemberId:  memberId,
			OrderId:   orderAddResp.Id,
		})
		if err != nil {
			// ⚠️ 警告：核销失败，订单已创建。Story 5.7 引入 Saga 补偿事务后此处需触发补偿。
			logc.Errorf(l.ctx, "[Saga-WARN] 优惠券核销失败，orderId=%d, couponId=%d, err=%s",
				orderAddResp.Id, req.CouponId, err.Error())
		}
	}

	// 13.如果使用积分,需要扣除积分（同样存在 Saga 补偿风险，见上方说明）
	if req.UseIntegration > 0 {
		newPoints := memberInfo.Points - req.UseIntegration
		if newPoints < 0 {
			newPoints = 0
		}
		_, err = l.svcCtx.MemberService.UpdateMemberPoints(l.ctx, &umsclient.UpdateMemberPointsReq{MemberId: memberId, Points: newPoints})
		if err != nil {
			// ⚠️ 警告：积分扣除失败，订单已创建。Story 5.7 引入 Saga 补偿事务后此处需触发补偿。
			logc.Errorf(l.ctx, "[Saga-WARN] 积分扣除失败，orderId=%d, memberId=%d, points=%d, err=%s",
				orderAddResp.Id, memberId, req.UseIntegration, err.Error())
		}
	}

	orderId := orderAddResp.Id
	// 14.发送延迟消息取消订单
	err = l.sendMsg(orderId, memberId)
	if err != nil {
		return nil, err
	}

	// === Task 9.4 & MEDIUM-5: 返回完整订单信息 ===
	return &types.GenerateOrderResp{
		Code:    0,
		Message: "下单成功",
		Data: types.GenerateOrderData{
			Id:                orderId,
			MemberId:          memberId,
			MemberUsername:    memberInfo.Nickname,
			TotalAmount:       totalAmount,
			PayAmount:         payAmount,
			FreightAmount:     0,
			PromotionAmount:   promotionAmountTotal,
			IntegrationAmount: integrationAmountTotal,
			CouponAmount:      couponAmountTotalFen,
			DiscountAmount:    0,
			PayType:           req.PayType,
			SourceType:        1,
			OrderType:         1,
			Integration:       int32(req.UseIntegration),
			UseIntegration:    req.UseIntegration,
		},
	}, nil
}

// 发送延迟消息取消订单
func (l *GenerateOrderLogic) sendMsg(orderId, memberId int64) error {
	delayMinutes := 30 // 延迟时间(分钟)
	message := map[string]any{"orderId": orderId, "memberId": memberId}
	body, err := sonic.Marshal(message)
	if err != nil {
		logc.Errorf(l.ctx, "序列化 JSON 失败: %v", err)
		return errorx.NewDefaultError("序列化 JSON 错误")
	}
	err = l.svcCtx.RabbitMQ.SendDelayMessage("order.delay.exchange", "order.delay.cancel.queue", "order.delay.cancel", body, delayMinutes)

	if err != nil {
		logc.Errorf(l.ctx, "订单 %d 延时取消,消息发送失败,异常:%s", orderId, err.Error())
		return errorx.NewDefaultError(fmt.Sprintf("订单 %d 延时取消,消息发送失败,异常:%s", orderId, err.Error()))
	}

	logc.Infof(l.ctx, "订单 %d 延时取消,消息已发送，将在 %d 分钟后处理", orderId, delayMinutes)
	return nil
}

func result(code int64, message string) *types.GenerateOrderResp {
	return &types.GenerateOrderResp{
		Code:    code,
		Message: message,
	}
}
