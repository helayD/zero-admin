package couponrecordservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddCouponRecordLogic 添加优惠券领取记录
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type AddCouponRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCouponRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCouponRecordLogic {
	return &AddCouponRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddCouponRecord 添加优惠券领取记录
// 1.查询优惠券是否存在
// 2.校验优惠券状态、启用、有效期
// 3.查询是否已经领取过优惠券了
// 4.原子更新优惠券领取数量（防并发超领）
// 5.添加领取优惠券记录
func (l *AddCouponRecordLogic) AddCouponRecord(in *smsclient.AddCouponRecordReq) (*smsclient.AddCouponRecordResp, error) {
	// 1.查询优惠券是否存在
	coupon := query.SmsCoupon
	smsCoupon, err := coupon.WithContext(l.ctx).Where(coupon.ID.Eq(in.CouponId)).First()

	if err != nil {
		return nil, errors.New("优惠券不存在")
	}

	// 2.校验优惠券状态：仅 status=1（进行中）允许领取
	switch smsCoupon.Status {
	case 0:
		return nil, errors.New("该优惠券暂不可领取")
	case 2:
		return nil, errors.New("该优惠券活动已结束")
	case 3:
		return nil, errors.New("该优惠券已取消")
	case 1:
		// 进行中，继续校验
	default:
		return nil, errors.New("该优惠券状态异常")
	}

	// 3.校验是否启用
	if smsCoupon.IsEnabled != 1 {
		return nil, errors.New("该优惠券暂不可领取")
	}

	// 4.校验有效期：当前时间必须在 start_time ~ end_time 范围内
	now := time.Now()
	if now.Before(smsCoupon.StartTime) {
		return nil, errors.New("该优惠券活动尚未开始")
	}
	if now.After(smsCoupon.EndTime) {
		return nil, errors.New("该优惠券已过期")
	}

	// 5.校验库存
	if smsCoupon.TotalCount-smsCoupon.ReceivedCount <= 0 {
		return nil, errors.New("优惠券已被领完")
	}

	// 6.事务：限领检查 + 原子更新优惠券领取数量 + 添加领取记录
	db := l.svcCtx.DB
	txErr := db.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 6a.事务内查询已领取次数（SELECT FOR UPDATE 锁行防并发超领）
		var count int64
		if err := tx.Raw(
			"SELECT COUNT(*) FROM sms_coupon_record WHERE coupon_id = ? AND member_id = ? FOR UPDATE",
			in.CouponId, in.MemberId,
		).Scan(&count).Error; err != nil {
			logc.Errorf(l.ctx, "查询优惠券领取记录失败,参数:%+v,异常:%s", in, err.Error())
			return errors.New("查询领取记录失败")
		}
		if count >= int64(smsCoupon.PerLimit) {
			return errors.New("您已达到该优惠券的领取上限")
		}

		// 6b.原子更新优惠券领取数量（CAS: received_count < total_count）
		result := tx.Exec(
			"UPDATE sms_coupon SET received_count = received_count + 1 WHERE id = ? AND received_count < total_count",
			in.CouponId,
		)
		if result.Error != nil {
			logc.Errorf(l.ctx, "原子更新优惠券领取数量失败,couponId:%d,异常:%s", in.CouponId, result.Error.Error())
			return errors.New("领取优惠券失败，请稍后重试")
		}
		if result.RowsAffected == 0 {
			return errors.New("优惠券已被领完")
		}

		// 6c.添加领取优惠券记录
		item := &model.SmsCouponRecord{
			CouponID: in.CouponId, // 优惠券ID
			MemberID: in.MemberId, // 用户ID
			GetTime:  time.Now(),  // 领取时间
			GetType:  in.GetType,  // 获取类型：0->后台赠送；1->主动获取
			Status:   0,           // 状态：0-未使用，1-已使用，2-已过期，3-已失效
		}
		if err := tx.Create(item).Error; err != nil {
			logc.Errorf(l.ctx, "添加优惠券领取记录失败,参数:%+v,异常:%s", item, err.Error())
			return errors.New("添加优惠券领取记录失败")
		}
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return &smsclient.AddCouponRecordResp{}, nil
}
