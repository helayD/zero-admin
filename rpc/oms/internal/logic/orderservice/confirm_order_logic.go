package orderservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmOrderLogic {
	return &ConfirmOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ConfirmOrder 确认收货(app)
func (l *ConfirmOrderLogic) ConfirmOrder(in *omsclient.ConfirmOrderReq) (*omsclient.ConfirmOrderResp, error) {
	order := query.OmsOrderMain
	item, err := order.WithContext(l.ctx).Where(order.ID.Eq(in.OrderId), order.UserID.Eq(in.MemberId), order.IsDeleted.Eq(0), order.OrderStatus.Eq(3)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "订单不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("订单不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询订单异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询订单异常")
	}

	item.OrderStatus = 4
	item.ReceiveStatus = 1

	_, err = order.WithContext(l.ctx).Where(order.ID.Eq(in.OrderId), order.UserID.Eq(in.MemberId)).Updates(item)

	if err != nil {
		logc.Errorf(l.ctx, "更新订单失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("更新订单失败")
	}

	currentScope, scopeErr := common.ResolveActorScope(l.ctx, l.svcCtx.DB, in.MemberId)
	if scopeErr != nil {
		logc.Errorf(l.ctx, "解析用户作用域失败,memberId:%d,异常:%s", in.MemberId, scopeErr.Error())
	}

	sendOrderEvent(l.ctx, l.svcCtx, "order.confirm.queue", "order.confirmed.key", "order.confirmed", in.OrderId, currentScope, in.MemberId, map[string]interface{}{
		"orderNo": item.OrderNo,
	})

	return &omsclient.ConfirmOrderResp{}, nil
}
