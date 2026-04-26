package couponrecordservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
)

// UpdateCouponRecordLogic 更新优惠券领取记录
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type UpdateCouponRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCouponRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponRecordLogic {
	return &UpdateCouponRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCouponRecord 更新优惠券领取记录
func (l *UpdateCouponRecordLogic) UpdateCouponRecord(in *smsclient.UpdateCouponRecordReq) (*smsclient.UpdateCouponRecordResp, error) {
	q := query.SmsCouponRecord

	for _, couponId := range in.CouponIds {
		// 1.查询优惠券领取记录是否已存在
		coupon, err := q.WithContext(l.ctx).Where(q.MemberID.Eq(in.MemberId), q.CouponID.Eq(couponId)).First()

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			logc.Errorf(l.ctx, "优惠券领取记录不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
			return nil, errors.New("优惠券领取记录不存在")
		case err != nil:
			logc.Errorf(l.ctx, "查询优惠券领取记录异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
			return nil, errors.New("查询优惠券领取记录异常")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status": in.Status,
		}

		if in.Status == 1 {
			updates["use_time"] = &now
			updates["order_id"] = in.OrderId
			updates["order_amount"] = float64(in.OrderAmount)
			updates["discount_amount"] = float64(in.DiscountAmount)
		}

		if in.Status == 0 {
			updates["use_time"] = nil
			updates["order_id"] = 0
			updates["order_amount"] = 0
			updates["discount_amount"] = 0
			updates["invalid_time"] = nil
			updates["invalid_reason"] = ""
		}

		if in.Status == 3 {
			invalidTime, _ := time.Parse("2006-01-02 15:04:05", in.InvalidTime)
			updates["invalid_time"] = &invalidTime
			updates["invalid_reason"] = in.InvalidReason
		}

		// 2.优惠券领取记录存在时,则按领取记录主键更新，避免 GORM 拦截无条件更新。
		err = l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_coupon_record").
			Where("id = ?", coupon.ID).
			Updates(updates).Error

		if err != nil {
			logc.Errorf(l.ctx, "更新优惠券领取记录失败,recordId:%d,参数:%+v,异常:%s", coupon.ID, in, err.Error())
			return nil, errors.New("更新优惠券领取记录失败")
		}

		if in.Status == 2 || in.Status == 3 {
			return &smsclient.UpdateCouponRecordResp{}, nil
		}

		c := query.SmsCoupon
		smsCoupon, err := c.WithContext(l.ctx).Where(c.ID.Eq(couponId)).First()
		if err != nil {
			logc.Errorf(l.ctx, "更新优惠券使用失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("更新优惠券使用失败")
		}

		count := smsCoupon.UsedCount
		if in.Status == 0 {
			count = count - 1
		} else {
			count = count + 1
		}
		_, err = c.WithContext(l.ctx).Where(c.ID.Eq(couponId)).Update(c.UsedCount, count)

		if err != nil {
			logc.Errorf(l.ctx, "更新优惠券使用失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("更新优惠券使用失败")
		}

	}
	return &smsclient.UpdateCouponRecordResp{}, nil
}
