package couponservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteCouponLogic 删除优惠券
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type DeleteCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCouponLogic {
	return &DeleteCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteCoupon 删除优惠券
func (l *DeleteCouponLogic) DeleteCoupon(in *smsclient.DeleteCouponReq) (*smsclient.DeleteCouponResp, error) {
	q := query.SmsCoupon
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, 0)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureCouponScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "sms.coupon.delete", 0, "", "delete coupon"); err != nil {
		return nil, err
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除优惠券失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除优惠券失败")
	}

	return &smsclient.DeleteCouponResp{}, nil
}
