package couponservicelogic

import (
	"context"
	"errors"
	"fmt"

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

	// 已发布且有领取记录的优惠券不允许物理删除，仅可取消(status=3)
	for _, id := range in.Ids {
		detail, err := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
		if err != nil {
			return nil, fmt.Errorf("优惠券(ID:%d)不存在或查询失败", id)
		}
		if detail.Status == 1 {
			recordQ := query.SmsCouponRecord
			recordCount, err := recordQ.WithContext(l.ctx).Where(recordQ.CouponID.Eq(id)).Count()
			if err != nil {
				logc.Errorf(l.ctx, "查询优惠券领取记录失败,couponId:%d,异常:%s", id, err.Error())
				return nil, errors.New("查询优惠券领取记录失败")
			}
			if recordCount > 0 {
				return nil, fmt.Errorf("优惠券「%s」已发布且有领取记录，不允许删除，请使用取消操作", detail.Name)
			}
		}
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除优惠券失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除优惠券失败")
	}

	return &smsclient.DeleteCouponResp{}, nil
}
