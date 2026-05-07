package orderservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderLogic {
	return &UpdateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateOrder 更新订单
func (l *UpdateOrderLogic) UpdateOrder(in *omsclient.UpdateOrderReq) (*omsclient.UpdateOrderResp, error) {
	q := query.OmsOrderMain.WithContext(l.ctx)
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, 0)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureOrderScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "oms.order.update", 0, "", "update order"); err != nil {
		return nil, err
	}

	// 1.根据订单id查询订单是否已存在
	_, err = q.Where(query.OmsOrderMain.ID.Eq(in.Id)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "订单不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("订单不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询订单异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询订单异常")
	}

	// 2.订单存在时,则直接更新订单
	updateMap := make(map[string]interface{})
	if in.OrderStatus != 0 {
		updateMap["order_status"] = in.OrderStatus
	}
	if in.Remark != "" {
		updateMap["remark"] = in.Remark
	}
	if in.ExpressOrderNumber != "" {
		updateMap["express_order_number"] = in.ExpressOrderNumber
		now := time.Now()
		updateMap["delivery_time"] = &now
	}
	if in.FreightAmount != 0 {
		updateMap["freight_amount"] = float64(in.FreightAmount)
	}
	if in.DiscountAmount != 0 {
		updateMap["discount_amount"] = float64(in.DiscountAmount)
	}

	if len(updateMap) > 0 {
		_, err = q.Where(query.OmsOrderMain.ID.Eq(in.Id)).Updates(updateMap)
	} else {
		return &omsclient.UpdateOrderResp{}, nil
	}

	if err != nil {
		logc.Errorf(l.ctx, "更新订单失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("更新订单失败")
	}

	return &omsclient.UpdateOrderResp{}, nil
}
