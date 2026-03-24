package couponservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateCouponStatusLogic 更新优惠券
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type UpdateCouponStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCouponStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponStatusLogic {
	return &UpdateCouponStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCouponStatus 更新优惠券状态
func (l *UpdateCouponStatusLogic) UpdateCouponStatus(in *smsclient.UpdateCouponStatusReq) (*smsclient.UpdateCouponStatusResp, error) {
	q := query.SmsCoupon
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureCouponScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "sms.coupon.status", in.UpdateBy, "", "update coupon status"); err != nil {
		return nil, err
	}

	// 发布校验：当目标状态为1（进行中/已发布）时，对每个优惠券执行发布前置校验
	if in.Status == 1 {
		for _, id := range in.Ids {
			detail, err := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
			if err != nil {
				logc.Errorf(l.ctx, "查询优惠券失败,id:%d,异常:%s", id, err.Error())
				return nil, fmt.Errorf("优惠券(ID:%d)不存在", id)
			}

			// 只有草稿状态(0)可以发布
			if detail.Status != 0 {
				return nil, fmt.Errorf("优惠券「%s」当前状态不允许发布，仅草稿状态可发布", detail.Name)
			}

			// 校验 end_time 晚于当前时间
			if detail.EndTime.Before(time.Now()) {
				return nil, fmt.Errorf("优惠券「%s」已过期，不允许发布", detail.Name)
			}

			// 校验 scope 至少有一条有效记录
			scopeQ := query.SmsCouponScope
			scopeCount, err := scopeQ.WithContext(l.ctx).Where(scopeQ.CouponID.Eq(id)).Count()
			if err != nil {
				logc.Errorf(l.ctx, "查询优惠券scope失败,couponId:%d,异常:%s", id, err.Error())
				return nil, errors.New("查询优惠券适用范围失败")
			}
			if scopeCount == 0 {
				return nil, fmt.Errorf("优惠券「%s」未配置适用范围，请先设置适用范围后再发布", detail.Name)
			}

			// 校验 coupon_type 存在且启用
			typeQ := query.SmsCouponType
			couponType, err := typeQ.WithContext(l.ctx).Where(typeQ.ID.Eq(detail.TypeID)).First()
			if err != nil {
				return nil, fmt.Errorf("优惠券「%s」关联的优惠券类型不存在", detail.Name)
			}
			if couponType.Status != 1 {
				return nil, fmt.Errorf("优惠券「%s」关联的优惠券类型「%s」已停用", detail.Name, couponType.Name)
			}
		}
	}

	// 取消校验：当目标状态为3（已取消）时，只允许从草稿(0)或进行中(1)取消
	if in.Status == 3 {
		for _, id := range in.Ids {
			detail, err := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
			if err != nil {
				return nil, fmt.Errorf("优惠券(ID:%d)不存在", id)
			}
			if detail.Status != 0 && detail.Status != 1 {
				return nil, fmt.Errorf("优惠券「%s」当前状态不允许取消", detail.Name)
			}
		}
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Update(q.Status, in.Status)

	if err != nil {
		logc.Errorf(l.ctx, "更新优惠券状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新优惠券状态失败")
	}

	return &smsclient.UpdateCouponStatusResp{}, nil
}
