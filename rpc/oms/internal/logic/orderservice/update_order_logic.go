package orderservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

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
		// Story 10.7 Fix: 状态变更为「已支付」时若 pay_time 仍为空则自动回填，
		// 让订单详情时间线能正确显示「支付成功」节点（buildTimeline 依赖 pay_time != ""）。
		// 只有当 pay_time IS NULL 时才回填，避免幂等重试覆盖原始支付时间。
		if in.OrderStatus == 2 {
			now := time.Now()
			var existing struct {
				PayTime *time.Time
			}
			if err = l.svcCtx.DB.WithContext(l.ctx).
				Table("oms_order_main").
				Select("pay_time").
				Where("id = ? AND is_deleted = 0", in.Id).
				Scan(&existing).Error; err == nil && existing.PayTime == nil {
				updateMap["pay_time"] = &now
			}
		}
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
		logc.Errorf(l.ctx, "更新订单失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新订单失败")
	}

	return &omsclient.UpdateOrderResp{}, nil
}
