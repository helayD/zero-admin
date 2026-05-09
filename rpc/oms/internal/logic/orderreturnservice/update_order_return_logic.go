package orderreturnservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/audit"
	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateOrderReturnLogic 更新退货/售后
/*
Author: LiuFeiHua
Date: 2025/06/30 15:49:28
*/
type UpdateOrderReturnLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderReturnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderReturnLogic {
	return &UpdateOrderReturnLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateOrderReturn 更新退货/售后
func (l *UpdateOrderReturnLogic) UpdateOrderReturn(in *omsclient.OrderReturnReq) (*omsclient.OrderReturnResp, error) {
	q := query.OmsOrderReturn.WithContext(l.ctx)

	// 1.根据退货/售后id查询退货/售后是否已存在
	returnApply, err := q.Where(query.OmsOrderReturn.ID.Eq(in.Id)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "退货/售后不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("退货/售后不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询退货/售后异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询退货/售后异常")
	}

	returnApply.Status = in.Status
	operTime := time.Now()
	// 确认退货
	if in.Status == 1 {
		// returnApply.CompanyAddressID = in.CompanyAddressID
		returnApply.HandleTime = &operTime
		returnApply.HandleNote = in.HandleNote
		returnApply.HandleMan = in.HandleMan
		returnApply.RefundAmount = float64(in.RefundAmount)
	}
	// 完成退货
	if in.Status == 2 {
		// returnApply.CompanyAddressID = in.CompanyAddressId
		returnApply.ReceiveMan = in.ReceiveMan
		returnApply.ReceiveTime = &operTime
		returnApply.ReceiveNote = in.ReceiveNote
	}
	// 拒绝退货
	if in.Status == 3 {
		returnApply.HandleTime = &operTime
		returnApply.HandleNote = in.HandleNote
		returnApply.HandleMan = in.HandleMan
	}

	// 记录旧状态用于判断是否首次进入"已退款"
	previousStatus := int32(0)
	if existing, lookupErr := q.Where(query.OmsOrderReturn.ID.Eq(in.Id)).First(); lookupErr == nil && existing != nil {
		previousStatus = existing.Status
	}

	// 2.退货/售后存在时,则直接更新退货/售后
	_, err = q.Where(query.OmsOrderReturn.ID.Eq(in.Id)).Updates(returnApply)

	if err != nil {
		logc.Errorf(l.ctx, "更新退货/售后失败,参数:%+v,异常:%s", returnApply, err.Error())
		return nil, errors.New("更新退货/售后失败")
	}

	// Story 10.6 Fix #3: 退货状态首次进入"已退款"时（按 proto 注释 status=3=已退款）
	// 发布 order.refund.queue 事件，触发 consumer 对订单购买型提货卡执行 refund_policy 处置
	const orderReturnStatusRefunded = int32(3)
	if in.Status == orderReturnStatusRefunded && previousStatus != orderReturnStatusRefunded {
		l.publishOrderRefundEvent(returnApply.OrderID, returnApply.MemberID)
	}

	return &omsclient.OrderReturnResp{}, nil
}

// publishOrderRefundEvent 发送订单退款成功事件，由 consumer/order_refund_logic 消费
// 事件 payload 字段对齐消费者 EventPayload 结构（entityId/actorId/scope/...）
func (l *UpdateOrderReturnLogic) publishOrderRefundEvent(orderID, memberID int64) {
	if orderID <= 0 || l.svcCtx == nil || l.svcCtx.RabbitMQ == nil {
		if l.svcCtx == nil || l.svcCtx.RabbitMQ == nil {
			logc.Infof(l.ctx, "RabbitMQ 未配置，跳过退款事件发送, orderId=%d", orderID)
		}
		return
	}

	// 读取订单 scope 字段（platform/tenant/merchant/user_id），跨域消费需要这些上下文
	type orderScopeRow struct {
		UserID     int64 `gorm:"column:user_id"`
		PlatformID int64 `gorm:"column:platform_id"`
		TenantID   int64 `gorm:"column:tenant_id"`
		MerchantID int64 `gorm:"column:merchant_id"`
	}
	var scope orderScopeRow
	if l.svcCtx.DB != nil {
		_ = l.svcCtx.DB.WithContext(l.ctx).
			Table("oms_order_main").
			Select("user_id, platform_id, tenant_id, merchant_id").
			Where("id = ? AND is_deleted = 0", orderID).
			Take(&scope).Error
	}
	actorID := scope.UserID
	if actorID == 0 {
		actorID = memberID
	}

	eventID := uuid.New().String()
	traceID := audit.NewTraceID("order.refunded", orderID)
	payload := map[string]interface{}{
		"eventId":    eventID,
		"occurredAt": time.Now().UnixMilli(),
		"traceId":    traceID,
		"platformId": scope.PlatformID,
		"tenantId":   scope.TenantID,
		"merchantId": scope.MerchantID,
		"actorId":    actorID,
		"entityId":   orderID,
		"action":     "refunded",
		"version":    "v1",
	}
	body, err := sonic.Marshal(payload)
	if err != nil {
		logc.Errorf(l.ctx, "序列化退款事件失败, orderId=%d, err=%v", orderID, err)
		return
	}

	if err := l.svcCtx.RabbitMQ.SendMessage("order.event.exchange", "direct",
		"order.refund.queue", "order.refund.key", body); err != nil {
		logc.Errorf(l.ctx, "发送订单退款事件失败, orderId=%d, eventId=%s, err=%v", orderID, eventID, err)
		return
	}
	logc.Infof(l.ctx, "发送订单退款事件成功, orderId=%d, eventId=%s, traceId=%s", orderID, eventID, traceID)
}
