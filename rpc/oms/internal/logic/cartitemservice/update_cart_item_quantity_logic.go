package cartitemservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/zeromicro/go-zero/core/logc"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateCartItemQuantityLogic 修改购物车中某个商品的数量
/*
Author: LiuFeiHua
Date: 2024/6/12 10:09
*/
type UpdateCartItemQuantityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCartItemQuantityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartItemQuantityLogic {
	return &UpdateCartItemQuantityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCartItemQuantity 修改购物车中某个商品的数量
// Task 5.1: UpdateSimple 条件已包含 MemberID.Eq(in.MemberId)，防止跨会员越权修改
// Task 5.4: 数量必须 > 0，否则拒绝
func (l *UpdateCartItemQuantityLogic) UpdateCartItemQuantity(in *omsclient.UpdateCartItemQuantityReq) (*omsclient.CartItemResp, error) {
	// Task 5.4: 数量边界保护
	if in.Quantity <= 0 {
		logc.Errorf(l.ctx, "修改购物车数量非法, quantity=%d", in.Quantity)
		return nil, errors.New("数量必须大于0")
	}

	q := query.OmsCartItem
	now := time.Now()
	expireTime := now.AddDate(0, 0, l.svcCtx.C.Cart.Timeout)
	_, err := q.WithContext(l.ctx).Where(q.ID.Eq(in.Id), q.MemberID.Eq(in.MemberId)).UpdateSimple(q.Quantity.Value(in.Quantity), q.ExpireTime.Value(expireTime))

	if err != nil {
		logc.Errorf(l.ctx, "修改购物车中某个商品的数量失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("修改购物车中某个商品的数量失败")
	}

	return &omsclient.CartItemResp{}, nil
}
