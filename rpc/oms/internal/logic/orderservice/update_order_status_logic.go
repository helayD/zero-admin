package orderservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderStatusLogic {
	return &UpdateOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新订单状态
func (l *UpdateOrderStatusLogic) UpdateOrderStatus(in *omsclient.UpdateOrderStatusReq) (*omsclient.UpdateOrderStatusResp, error) {
	if len(in.Ids) == 0 {
		return nil, errors.New("缺少有效订单ID")
	}

	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, nil, 0)
	if err != nil {
		return nil, err
	}
	if _, err = logiccommon.EnsureOrderScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "oms.order.update_status", 0, "", "update order status"); err != nil {
		return nil, err
	}

	updates := map[string]any{
		"update_time": time.Now(),
	}
	if in.OrderStatus > 0 {
		updates["order_status"] = in.OrderStatus
	}
	if in.ReceiveStatus == 0 || in.ReceiveStatus == 1 {
		updates["receive_status"] = in.ReceiveStatus
		if in.ReceiveStatus == 1 {
			updates["receive_time"] = time.Now()
		}
	}
	if len(updates) == 1 {
		return nil, errors.New("缺少有效更新字段")
	}

	orderMain := query.OmsOrderMain
	_, err = orderMain.WithContext(l.ctx).Where(orderMain.ID.In(in.Ids...)).Updates(updates)
	if err != nil {
		logc.Errorf(l.ctx, "更新订单状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新订单状态失败")
	}

	return &omsclient.UpdateOrderStatusResp{}, nil
}
