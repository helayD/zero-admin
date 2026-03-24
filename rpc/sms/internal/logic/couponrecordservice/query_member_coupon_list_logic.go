package couponrecordservicelogic

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryMemberCouponListLogic 获取会员优惠券
/*
Author: LiuFeiHua
Date: 2025/6/11 14:48
*/
type QueryMemberCouponListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberCouponListLogic {
	return &QueryMemberCouponListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryMemberCouponList 获取会员优惠券
func (l *QueryMemberCouponListLogic) QueryMemberCouponList(in *smsclient.QueryMemberCouponListReq) (*smsclient.QueryMemberCouponListResp, error) {
	var result []model.SmsCoupon
	db := l.svcCtx.DB
	var err error
	if in.Status < 0 {
		sql := `select distinct t2.*
					from sms_coupon_record t1
							 join sms_coupon t2 on t1.coupon_id = t2.id
					where t1.member_id = ?;`
		err = db.Raw(sql, in.MemberId).Find(&result).Error
	} else {
		sql := `select distinct t2.*
					from sms_coupon_record t1
							 join sms_coupon t2 on t1.coupon_id = t2.id
					where t1.member_id = ?
					  and t1.status = ?;`
		err = db.Raw(sql, in.MemberId, in.Status).Find(&result).Error
	}

	if err != nil {
		logc.Errorf(l.ctx, "查询优惠券领取记录列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优惠券领取记录列表失败")
	}

	// 批量查询 scope 信息，避免 N+1 问题
	couponIDs := make([]int64, 0, len(result))
	for _, item := range result {
		couponIDs = append(couponIDs, item.ID)
	}
	scopeMap := make(map[int64]int32)
	if len(couponIDs) > 0 {
		scope := query.SmsCouponScope
		scopes, scopeErr := scope.WithContext(l.ctx).Where(scope.CouponID.In(couponIDs...)).Find()
		if scopeErr != nil {
			logc.Errorf(l.ctx, "批量查询优惠券scope失败,异常:%s", scopeErr.Error())
		} else {
			for _, s := range scopes {
				scopeMap[s.CouponID] = s.ScopeType
			}
		}
	}

	var list []*smsclient.QueryCouponData

	for _, item := range result {
		isExpired := item.EndTime.Before(time.Now())
		if in.Status == 0 && isExpired {
			record := query.SmsCouponRecord
			_, updateErr := record.WithContext(l.ctx).Where(record.CouponID.Eq(item.ID), record.Status.Eq(0)).Update(record.Status, 2)
			if updateErr != nil {
				logc.Errorf(l.ctx, "自动更新过期优惠券记录状态失败,couponId:%d,异常:%s", item.ID, updateErr.Error())
			}
			continue
		}

		list = append(list, &smsclient.QueryCouponData{
			Id:            item.ID,                             // 优惠券ID
			TypeId:        item.TypeID,                         // 优惠券类型ID
			Name:          item.Name,                           // 优惠券名称
			Code:          item.Code,                           // 优惠券码
			Amount:        float32(item.Amount),                // 优惠金额/折扣率
			MinAmount:     float32(item.MinAmount),             // 最低使用金额
			StartTime:     time_util.TimeToStr(item.StartTime), // 生效时间
			EndTime:       time_util.TimeToStr(item.EndTime),   // 失效时间
			PerLimit:      item.PerLimit,                       // 每人限领数量
			Status:        item.Status,                         // 状态：0-未开始，1-进行中，2-已结束，3-已取消
			Description:   item.Description,                    // 使用说明
			ScopeType:     scopeMap[item.ID],                   // 从预批量查询 scopeMap 获取，避免 N+1
			ReceiveStatus: 1,                                   // 已领取（此接口返回的是已领优惠券）
			TotalCount:    item.TotalCount,
			ReceivedCount: item.ReceivedCount,
		})
	}
	return &smsclient.QueryMemberCouponListResp{
		List: list,
	}, nil
}
