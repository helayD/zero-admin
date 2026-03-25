package couponrecordservicelogic

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
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

// couponRecordRow 用于接收 record + coupon JOIN 查询结果
type couponRecordRow struct {
	RecordID      int64     `gorm:"column:record_id"`
	RecordStatus  int32     `gorm:"column:record_status"`
	CouponID      int64     `gorm:"column:coupon_id"`
	TypeID        int64     `gorm:"column:type_id"`
	Name          string    `gorm:"column:name"`
	Code          string    `gorm:"column:code"`
	Amount        float64   `gorm:"column:amount"`
	MinAmount     float64   `gorm:"column:min_amount"`
	StartTime     time.Time `gorm:"column:start_time"`
	EndTime       time.Time `gorm:"column:end_time"`
	PerLimit      int32     `gorm:"column:per_limit"`
	CouponStatus  int32     `gorm:"column:coupon_status"`
	Description   string    `gorm:"column:description"`
	TotalCount    int32     `gorm:"column:total_count"`
	ReceivedCount int32     `gorm:"column:received_count"`
}

// QueryMemberCouponList 获取会员优惠券
func (l *QueryMemberCouponListLogic) QueryMemberCouponList(in *smsclient.QueryMemberCouponListReq) (*smsclient.QueryMemberCouponListResp, error) {
	var rows []couponRecordRow
	db := l.svcCtx.DB
	var err error

	// 查询 record 行并 JOIN coupon 详情，每条领取记录独立返回（解决 DISTINCT 合并问题）
	baseSQL := `SELECT t1.id AS record_id, t1.status AS record_status,
					   t2.id AS coupon_id, t2.type_id, t2.name, t2.code,
					   t2.amount, t2.min_amount, t2.start_time, t2.end_time,
					   t2.per_limit, t2.status AS coupon_status, t2.description,
					   t2.total_count, t2.received_count
				FROM sms_coupon_record t1
				JOIN sms_coupon t2 ON t1.coupon_id = t2.id
				WHERE t1.member_id = ?`

	if in.Status < 0 {
		err = db.Raw(baseSQL, in.MemberId).Find(&rows).Error
	} else {
		err = db.Raw(baseSQL+" AND t1.status = ?", in.MemberId, in.Status).Find(&rows).Error
	}

	if err != nil {
		logc.Errorf(l.ctx, "查询优惠券领取记录列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优惠券领取记录列表失败")
	}

	// 批量查询 scope 信息，避免 N+1 问题
	couponIDSet := make(map[int64]struct{})
	for _, row := range rows {
		couponIDSet[row.CouponID] = struct{}{}
	}
	couponIDs := make([]int64, 0, len(couponIDSet))
	for id := range couponIDSet {
		couponIDs = append(couponIDs, id)
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

	for _, row := range rows {
		// 过期记录状态更新已移至定时Job处理，查询接口不再执行写操作
		// 此处仅跳过过期的未使用优惠券（不在列表中展示）
		if in.Status == 0 && row.EndTime.Before(time.Now()) {
			continue
		}

		list = append(list, &smsclient.QueryCouponData{
			Id:            row.CouponID,                       // 优惠券ID
			TypeId:        row.TypeID,                         // 优惠券类型ID
			Name:          row.Name,                           // 优惠券名称
			Code:          row.Code,                           // 优惠券码
			Amount:        float32(row.Amount),                // 优惠金额/折扣率
			MinAmount:     float32(row.MinAmount),             // 最低使用金额
			StartTime:     time_util.TimeToStr(row.StartTime), // 生效时间
			EndTime:       time_util.TimeToStr(row.EndTime),   // 失效时间
			PerLimit:      row.PerLimit,                       // 每人限领数量
			Status:        row.CouponStatus,                   // 状态：0-未开始，1-进行中，2-已结束，3-已取消
			Description:   row.Description,                    // 使用说明
			ScopeType:     scopeMap[row.CouponID],             // 从预批量查询 scopeMap 获取，避免 N+1
			ReceiveStatus: 1,                                  // 已领取（此接口返回的是已领优惠券）
			TotalCount:    row.TotalCount,
			ReceivedCount: row.ReceivedCount,
		})
	}
	return &smsclient.QueryMemberCouponListResp{
		List: list,
	}, nil
}
