package orderservicelogic

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrderConsistencyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderConsistencyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderConsistencyLogic {
	return &UpdateOrderConsistencyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateOrderConsistency 更新订单一致性阶段（用于补偿链路）
func (l *UpdateOrderConsistencyLogic) UpdateOrderConsistency(in *omsclient.UpdateOrderConsistencyReq) (*omsclient.UpdateOrderConsistencyResp, error) {
	updates := map[string]interface{}{
		"consistency_stage":     in.ConsistencyStage,
		"consistency_result":   in.ConsistencyResult,
		"last_error":          in.LastError,
		"retry_count":          in.RetryCount,
		"last_compensation_at": time.Now(),
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.OmsOrderMain{}).
		Where("id = ? AND is_deleted = 0", in.OrderId).
		Updates(updates)

	if result.Error != nil {
		logc.Errorf(l.ctx, "更新订单一致性阶段失败,orderId=%d,err=%s", in.OrderId, result.Error.Error())
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		logc.Infof(l.ctx, "更新订单一致性阶段:订单不存在或已删除,orderId=%d", in.OrderId)
	}

	logc.Infof(l.ctx, "更新订单一致性阶段成功,orderId=%d,stage=%d,result=%d,retryCount=%d",
		in.OrderId, in.ConsistencyStage, in.ConsistencyResult, in.RetryCount)

	return &omsclient.UpdateOrderConsistencyResp{Pong: "pong"}, nil
}
