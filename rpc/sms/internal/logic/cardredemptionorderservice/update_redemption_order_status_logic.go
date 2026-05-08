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

	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table(order.TableName()).
		Where("id = ?", order.ID).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新提货单状态失败: %w", err)
	}

	order.Status = in.Status
	if in.OmsOrderId > 0 {
		order.OmsOrderID = in.OmsOrderId
	}

	return &smsclient.UpdateRedemptionOrderStatusResp{
		Order: buildRedemptionOrderData(&order),
	}, nil
}
