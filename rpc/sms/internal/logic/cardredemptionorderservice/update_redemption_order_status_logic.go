package cardredemptionorderservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateRedemptionOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRedemptionOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRedemptionOrderStatusLogic {
	return &UpdateRedemptionOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRedemptionOrderStatusLogic) UpdateRedemptionOrderStatus(in *smsclient.UpdateRedemptionOrderStatusReq) (*smsclient.UpdateRedemptionOrderStatusResp, error) {
	if in.OrderId <= 0 {
		return nil, errors.New("提货单ID无效")
	}
	if in.Status == "" {
		return nil, errors.New("状态不能为空")
	}

	var order redemptionOrderRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table(order.TableName()).
		Where("id = ? AND is_deleted = 0", in.OrderId).
		Take(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("提货单不存在")
		}
		return nil, err
	}

	validTransitions := map[string][]string{
		redemptionOrderStatusPending:    {redemptionOrderStatusProcessing, redemptionOrderStatusCancelled},
		redemptionOrderStatusProcessing: {redemptionOrderStatusShipped, redemptionOrderStatusCancelled},
		redemptionOrderStatusShipped:    {redemptionOrderStatusDelivered},
	}

	allowedNextStatuses, ok := validTransitions[order.Status]
	if !ok {
		return nil, fmt.Errorf("当前状态%s不允许变更", order.Status)
	}

	allowed := false
	for _, s := range allowedNextStatuses {
		if s == in.Status {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("状态从%s变更为%s不允许", order.Status, in.Status)
	}

	updates := map[string]interface{}{
		"status":     in.Status,
		"updated_at": time.Now(),
	}
	if in.OmsOrderId > 0 {
		updates["oms_order_id"] = in.OmsOrderId
	}
	if in.Status == redemptionOrderStatusShipped {
		now := time.Now()
		updates["shipped_at"] = now
	}
	if in.Status == redemptionOrderStatusDelivered {
		now := time.Now()
		updates["delivered_at"] = now
	}

	// Review #5 H4: 绑定 OMS 订单时使用 CAS，确保单条提货单只能绑定一次。
	// 同一 MQ 事件被并发消费时，第二次绑定会得到 RowsAffected==0 并被识别为冲突。
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table(order.TableName()).
		Where("id = ? AND is_deleted = 0", order.ID)
	if in.OmsOrderId > 0 {
		query = query.Where("oms_order_id = 0")
	}
	updRes := query.Updates(updates)
	if updRes.Error != nil {
		return nil, fmt.Errorf("更新提货单状态失败: %w", updRes.Error)
	}
	if updRes.RowsAffected == 0 {
		if in.OmsOrderId > 0 {
			return nil, errors.New("提货单 OMS 订单已绑定，无法重复绑定")
		}
		return nil, errors.New("提货单状态已变化，更新失败")
	}

	order.Status = in.Status
	if in.OmsOrderId > 0 {
		order.OmsOrderID = in.OmsOrderId
	}

	return &smsclient.UpdateRedemptionOrderStatusResp{
		Order: buildRedemptionOrderData(&order),
	}, nil
}
