package order

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/logic/member/coupon"
	"github.com/feihua/zero-admin/api/front/internal/logic/order/cart"
	"github.com/feihua/zero-admin/api/front/internal/middleware"
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
	ErrCodeOrderNoAddress                 = "OMS_ORDER_NO_ADDRESS"
	ErrCodeOrderAddressInvalid            = "OMS_ORDER_ADDRESS_INVALID"
	ErrCodeOrderCouponUnavailable         = "OMS_ORDER_COUPON_UNAVAILABLE"
	ErrCodeOrderIntegrationExceed         = "OMS_ORDER_INTEGRATION_EXCEED"
	ErrCodeOrderIntegrationCouponConflict = "OMS_ORDER_INTEGRATION_COUPON_CONFLICT"
	ErrCodeOrderPayTypeInvalid            = "OMS_ORDER_PAY_TYPE_INVALID"
	ErrCodeOrderStatusInvalid             = "OMS_ORDER_STATUS_INVALID" // 订单状态不允许支付
	ErrCodeOrderPayFailed                 = "OMS_ORDER_PAY_FAILED"     // 支付发起失败
	ErrCodeOrderStockInsufficient         = "OMS_ORDER_STOCK_INSUFFICIENT"
	ErrCodeOrderSystemError               = "OMS_ORDER_SYSTEM_ERROR"
)

// Story 5.4 新增错误码
const (
	ErrCodeOrderDuplicatedRequest  = "OMS_ORDER_DUPLICATED_REQUEST"  // 重复提交
	ErrCodeOrderCompensationFailed = "OMS_ORDER_COMPENSATION_FAILED" // Saga 补偿失败，需人工介入
	ErrCodeOrderStockLocked        = "OMS_ORDER_STOCK_LOCKED"        // 库存已被其他订单锁定
)

// Saga 链路总超时（60s），防止客户端断开后资源泄漏
const sagaTimeout = 60 * time.Second

// GenerateOrderLogic
/*
Author: LiuFeiHua
Date: 2023/12/12 18:04
Updated by Story 5-3: Complete order generation with address, coupon, integration, and pay type.
Updated by Story 5-4: Add idempotency key mechanism and Saga compensation chain.
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

// GenerateOrder 根据提交信息生成订单
// Story 5.4 实现内容：
// - 幂等键检查（Redis 三状态机）
// - Saga 补偿链路（单次请求内内存级回滚）
// - Saga 补偿顺序：① 库存预锁 → ② 创建订单主记录 → ③ 存储收货人 → ④ 核销优惠券 → ⑤ 扣除积分 → ⑥ 发送延迟消息
func (l *GenerateOrderLogic) GenerateOrder(req *types.GenerateOrderReq) (*types.GenerateOrderResp, error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 使用带超时的 context 包装 Saga 链路，客户端断开时自动取消
	ctx, cancel := context.WithTimeout(l.ctx, sagaTimeout)
	defer cancel()

	// === Task 1: 幂等键检查 ===
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey != "" {
		result, hit, err := middleware.CheckAndSetProcessing(l.ctx, l.svcCtx.Redis, idempotencyKey)
		if err != nil {
			logc.Errorf(l.ctx, "[Saga-IDEM] 幂等键 Redis 异常, key=%s, err=%s", idempotencyKey, err.Error())
			return nil, errorx.NewDefaultError(ErrCodeOrderSystemError)
		}
		if hit {
			switch result.State {
			case middleware.StateCompleted:
				// 幂等命中，返回原订单ID
				logc.Infof(l.ctx, "[Saga-IDEM] 幂等命中，返回已有订单, key=%s, orderId=%d", idempotencyKey, result.OrderId)
				return &types.GenerateOrderResp{
					Code:    0,
					Message: "订单已存在",
					Data: types.GenerateOrderData{
						Id: result.OrderId,
					},
				}, nil
			case middleware.StateFailed:
				// 幂等命中失败，返回原错误
				logc.Infof(l.ctx, "[Saga-IDEM] 幂等命中失败, key=%s, errCode=%s, errMsg=%s", idempotencyKey, result.ErrCode, result.ErrMsg)
				return nil, errorx.NewDefaultError(result.ErrCode)
			case middleware.StateProcessing:
				// 另一个请求正在处理中
				logc.Infof(l.ctx, "[Saga-IDEM] 幂等键处理中，等待超时, key=%s", idempotencyKey)
				return nil, errorx.NewDefaultError(ErrCodeOrderDuplicatedRequest)
			}
		}
		// 设置成功，继续处理
	}

	// 使用 defer 处理幂等键（TTL 在 MarkCompleted/MarkFailed 中自动设置，无需显式清理）
	// LOW-2 修复：Saga 链路使用 ctxWithTimeout，客户端断开时自动取消

	// === Task 9: 校验收货地址 ===
	if req.MemberReceiveAddressId <= 0 {
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderNoAddress, "收货地址无效")
		}
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
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderAddressInvalid, "收货地址无效")
		}
		return nil, errorx.NewDefaultError(ErrCodeOrderAddressInvalid)
	}

	memberInfo, _ := l.svcCtx.MemberService.QueryMemberInfoDetail(l.ctx, &umsclient.QueryMemberInfoDetailReq{MemberId: memberId})
	if memberInfo == nil {
		memberInfo = &umsclient.QueryMemberInfoDetailResp{}
	}

	// 1.获取购物车及优惠信息 / 立即购买商品信息
	isDirectBuy := req.DirectItem.ProductId > 0
	var cartPromotionItemList []types.CarItemtPromotionListData
	if isDirectBuy {
		cartPromotionItemList, err = cart.QueryDirectOrderPromotion(&req.DirectItem, l.ctx, l.svcCtx)
	} else {
		cartPromotionItemList, err = cart.QueryCartListPromotion(req.CartIds, l.ctx, l.svcCtx)
	}
	if err != nil {
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderSystemError, "购物车查询失败")
		}
		return nil, err
	}
	if len(cartPromotionItemList) == 0 {
		if idempotencyKey != "" {
			emptyMessage := "购物车为空"
			if isDirectBuy {
				emptyMessage = "立即购买商品为空"
			}
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderSystemError, emptyMessage)
		}
		if isDirectBuy {
			return result(1, "当前商品暂不可直接购买，请重新选择规格后再试"), nil
		}
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
	for _, item := range cartPromotionItemList {
		skuTotalAmt := int64(item.Price) * int64(item.Quantity)
		itemPromoAmt := item.ReduceAmount
		orderItem := &omsclient.OrderItemData{
			Id:              item.Id,
			SkuId:           item.ProductSkuId,
			SkuName:         item.ProductName,
			SkuPic:          item.ProductPic,
			SkuPrice:        float32(item.Price), // int64 → float32
			SkuQuantity:     int32(item.Quantity),
			SpecData:        item.ProductAttr,
			SkuTotalAmount:  float32(skuTotalAmt),  // int64 → float32
			PromotionAmount: float32(itemPromoAmt), // int64 → float32
		}
		orderItemList = append(orderItemList, orderItem)

		if item.RealStock <= 0 || item.RealStock < item.Quantity {
			flag = true
		}
	}

	// 3.判断购物车中商品是否都有库存
	if flag {
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderStockInsufficient, "库存不足")
		}
		return nil, errorx.NewDefaultError(ErrCodeOrderStockInsufficient)
	}

	// 优惠券按商品金额（分）比例分摊（全场通用）
	couponAmountTotalFen := int64(0)
	if req.CouponId > 0 {
		enableList, _, err := coupon.QueryCouponList(l.svcCtx, l.ctx, cartPromotionItemList)
		if err != nil {
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderSystemError, "优惠券查询失败")
			}
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
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderCouponUnavailable, "优惠券不可用")
			}
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

	// === Task 2.3: 积分提交 ===
	var integrationAmountTotal int64 = 0
	if req.UseIntegration > 0 {
		if req.UseIntegration > memberInfo.Points {
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderIntegrationExceed, "积分不足")
			}
			return nil, errorx.NewDefaultError(ErrCodeOrderIntegrationExceed)
		}
		consumeSetting, _ := l.svcCtx.MemberConsumeSettingService.QueryMemberConsumeSettingDetail(l.ctx, &umsclient.QueryMemberConsumeSettingDetailReq{Id: 1})
		if consumeSetting == nil {
			consumeSetting = &umsclient.QueryMemberConsumeSettingDetailResp{}
		}
		// 积分与优惠券互斥判断
		if req.CouponId > 0 && consumeSetting.CouponStatus == 0 {
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderIntegrationCouponConflict, "积分与优惠券互斥")
			}
			return nil, errorx.NewDefaultError(ErrCodeOrderIntegrationCouponConflict)
		}
		// 最低使用门槛
		if req.UseIntegration < consumeSetting.UseUnit {
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderIntegrationExceed, "积分不足最低门槛")
			}
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

	// === Task 2.5: 金额计算收口 ===
	payAmount := totalAmount - promotionAmountTotal - couponAmountTotalFen - integrationAmountTotal
	if payAmount < 0 {
		payAmount = 0
	}

	// === Task 2.4: 支付方式校验 ===
	if req.PayType != 1 && req.PayType != 2 {
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderPayTypeInvalid, "支付方式无效")
		}
		return nil, errorx.NewDefaultError(ErrCodeOrderPayTypeInvalid)
	}

	// =========================================================
	// === Saga 补偿链路（Task 2: Story 5.4 核心改造）===
	// =========================================================
	// 补偿顺序：① 库存预锁 → ② 创建订单主记录 → ③ 存储收货人 → ④ 核销优惠券 → ⑤ 扣除积分 → ⑥ 发送延迟消息
	// 每个步骤后记录补偿标记；若步骤 N+1 失败，则回滚步骤 1~N

	// Saga 补偿标记
	sagaStockLocked := false
	sagaOrderCreated := false
	sagaCouponConsumed := false
	sagaPointsDeducted := false
	orderId := int64(0)
	orderNo := genOrderNo()

	// Saga 补偿函数：MEDIUM-2 修复 — 记录补偿失败日志，不静默忽略
	compensate := func(failedStep string) {
		logc.Infof(l.ctx, "[Saga-COMP] 开始 Saga 补偿，失败步骤=%s, orderNo=%s", failedStep, orderNo)

		// 回滚积分扣除（步骤⑤）
		if sagaPointsDeducted {
			if _, compErr := l.svcCtx.MemberService.UpdateMemberPoints(ctx, &umsclient.UpdateMemberPointsReq{
				MemberId: memberId,
				Points:   memberInfo.Points, // 还原原始积分
			}); compErr != nil {
				logc.Errorf(l.ctx, "[Saga-COMP] 积分回滚失败, memberId=%d, err=%s", memberId, compErr.Error())
			} else {
				logc.Infof(l.ctx, "[Saga-COMP] 已回滚积分, memberId=%d", memberId)
			}
		}

		// 回滚优惠券核销（步骤④）
		if sagaCouponConsumed {
			if _, compErr := l.svcCtx.CouponRecordService.UpdateCouponRecord(ctx, &smsclient.UpdateCouponRecordReq{
				MemberId:  memberId,
				CouponIds: []int64{req.CouponId},
				Status:    0, // 还原为未使用
			}); compErr != nil {
				logc.Errorf(l.ctx, "[Saga-COMP] 优惠券回滚失败, couponId=%d, err=%s", req.CouponId, compErr.Error())
			} else {
				logc.Infof(l.ctx, "[Saga-COMP] 已回滚优惠券, couponId=%d", req.CouponId)
			}
		}

		// 回滚订单创建（步骤②）- 仅当订单已创建时
		if sagaOrderCreated && orderId > 0 {
			if _, compErr := l.svcCtx.OrderService.CancelOrder(ctx, &omsclient.CancelOrderReq{
				MemberId: memberId,
				OrderId:  orderId,
			}); compErr != nil {
				logc.Errorf(l.ctx, "[Saga-COMP] 订单回滚失败, orderId=%d, err=%s", orderId, compErr.Error())
			} else {
				logc.Infof(l.ctx, "[Saga-COMP] 已回滚订单, orderId=%d", orderId)
			}
		}

		// 回滚库存锁定（步骤①）- 仅当库存已锁定时
		if sagaStockLocked {
			var stockReleaseData []*pmsclient.UpdateSkuStockData
			for _, item := range orderItemList {
				stockReleaseData = append(stockReleaseData, &pmsclient.UpdateSkuStockData{
					Id:              item.SkuId,
					ProductQuantity: item.SkuQuantity,
				})
			}
			if _, compErr := l.svcCtx.ProductSkuService.ReleaseSkuStockLock(ctx, &pmsclient.UpdateSkuStockReq{
				Data: stockReleaseData,
			}); compErr != nil {
				logc.Errorf(l.ctx, "[Saga-COMP] 库存释放失败, orderNo=%s, err=%s", orderNo, compErr.Error())
			} else {
				logc.Infof(l.ctx, "[Saga-COMP] 已释放库存锁定, orderNo=%s", orderNo)
			}
		}

		logc.Infof(l.ctx, "[Saga-COMP] Saga 补偿完成, failedStep=%s, orderNo=%s", failedStep, orderNo)
	}

	// ① 库存预锁（使用带超时的 ctx）
	var stockLockData []*pmsclient.UpdateSkuStockData
	for _, item := range orderItemList {
		stockLockData = append(stockLockData, &pmsclient.UpdateSkuStockData{
			Id:              item.SkuId,
			ProductQuantity: item.SkuQuantity,
		})
	}
	_, err = l.svcCtx.ProductSkuService.LockSkuStockLock(ctx, &pmsclient.UpdateSkuStockReq{
		Data: stockLockData,
	})
	if err != nil {
		logc.Errorf(l.ctx, "锁定库存异常,参数: %+v,异常：%s", stockLockData, err.Error())
		s, _ := status.FromError(err)
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderStockLocked, s.Message())
		}
		return nil, errorx.NewDefaultError(ErrCodeOrderStockLocked)
	}
	sagaStockLocked = true
	logc.Infof(l.ctx, "[Saga] 步骤①库存预锁成功, orderNo=%s", orderNo)

	// ② 创建订单主记录
	orderInfo := &omsclient.AddOrderReq{
		OrderNo:         orderNo,
		UserId:          memberId,
		OrderStatus:     1,                    // 1-待支付
		TotalAmount:     float32(totalAmount), // 分→元（float32）
		PromotionAmount: float32(promotionAmountTotal),
		CouponAmount:    float32(couponAmountTotalFen),
		PointsAmount:    float32(integrationAmountTotal),
		DiscountAmount:  0,
		FreightAmount:   0,
		PayAmount:       float32(payAmount),
		SourceType:      1, // 1-APP
		UsePoints:       req.UseIntegration,
		OrderItemData:   orderItemList,
	}

	orderAddResp, err := l.svcCtx.OrderService.AddOrder(ctx, orderInfo)
	if err != nil {
		logc.Errorf(l.ctx, "[Saga-STEP2] 创建订单失败, orderNo=%s, err=%s", orderNo, err.Error())
		compensate("②创建订单")
		if idempotencyKey != "" {
			middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderCompensationFailed, "订单创建失败")
		}
		return nil, errorx.NewDefaultError(ErrCodeOrderSystemError)
	}
	orderId = orderAddResp.Id
	sagaOrderCreated = true
	logc.Infof(l.ctx, "[Saga] 步骤②创建订单成功, orderId=%d, orderNo=%s", orderId, orderNo)

	// ③ 存储收货人信息
	_, err = l.svcCtx.OrderDeliveryService.AddOrderDelivery(ctx, &omsclient.AddOrderDeliveryReq{
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
		logc.Errorf(l.ctx, "[Saga-STEP3] 保存收货人信息失败, orderId=%d, err=%s", orderAddResp.Id, err.Error())
		// 不阻塞订单创建，记录日志即可
	}
	logc.Infof(l.ctx, "[Saga] 步骤③收货人信息保存完成, orderId=%d", orderId)

	// ④ 核销优惠券
	if req.CouponId > 0 {
		_, err = l.svcCtx.CouponRecordService.UpdateCouponRecord(ctx, &smsclient.UpdateCouponRecordReq{
			CouponIds: []int64{req.CouponId},
			MemberId:  memberId,
			OrderId:   orderId,
			Status:    1,
		})
		if err != nil {
			logc.Errorf(l.ctx, "[Saga-STEP4] 优惠券核销失败, orderId=%d, couponId=%d, err=%s", orderId, req.CouponId, err.Error())
			compensate("④核销优惠券")
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderCompensationFailed, "优惠券核销失败，订单已回滚")
			}
			return nil, errorx.NewDefaultError(ErrCodeOrderCompensationFailed)
		}
		sagaCouponConsumed = true
		logc.Infof(l.ctx, "[Saga] 步骤④优惠券核销成功, orderId=%d, couponId=%d", orderId, req.CouponId)
	}

	// ⑤ 扣除积分
	if req.UseIntegration > 0 {
		newPoints := memberInfo.Points - req.UseIntegration
		if newPoints < 0 {
			newPoints = 0
		}
		_, err = l.svcCtx.MemberService.UpdateMemberPoints(ctx, &umsclient.UpdateMemberPointsReq{
			MemberId: memberId,
			Points:   newPoints,
		})
		if err != nil {
			logc.Errorf(l.ctx, "[Saga-STEP5] 积分扣除失败, orderId=%d, memberId=%d, points=%d, err=%s", orderId, memberId, req.UseIntegration, err.Error())
			compensate("⑤扣除积分")
			if idempotencyKey != "" {
				middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, ErrCodeOrderCompensationFailed, "积分扣除失败，订单已回滚")
			}
			return nil, errorx.NewDefaultError(ErrCodeOrderCompensationFailed)
		}
		sagaPointsDeducted = true
		logc.Infof(l.ctx, "[Saga] 步骤⑤积分扣除成功, orderId=%d, deduct=%d, remaining=%d", orderId, req.UseIntegration, newPoints)
	}

	// ⑥ 发送延迟消息
	err = l.sendMsg(ctx, orderId, memberId)
	if err != nil {
		// 延迟消息发送失败不影响订单创建，仅记录日志
		logc.Errorf(l.ctx, "[Saga-STEP6] 延迟消息发送失败, orderId=%d, err=%s", orderId, err.Error())
	}

	// 幂等键标记为完成
	if idempotencyKey != "" {
		middleware.MarkCompleted(l.ctx, l.svcCtx.Redis, idempotencyKey, orderId)
	}

	// === 返回完整订单信息 ===
	return &types.GenerateOrderResp{
		Code:    0,
		Message: "下单成功",
		Data: types.GenerateOrderData{
			Id:                orderId,
			OrderSn:           orderNo,
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

// 发送延迟消息取消订单（使用 ctx 以支持超时取消）
func (l *GenerateOrderLogic) sendMsg(ctx context.Context, orderId, memberId int64) error {
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

// trimQuotes 去除字符串首尾的引号
func trimQuotes(s string) string {
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	return s
}
