package couponrecordservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAvailableCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAvailableCouponsLogic {
	return &QueryAvailableCouponsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryAvailableCoupons 查询可领取的优惠券列表
// 条件：status=1（进行中） + is_enabled=1 + 在有效期内 + received_count < total_count
func (l *QueryAvailableCouponsLogic) QueryAvailableCoupons(in *smsclient.QueryAvailableCouponsReq) (*smsclient.QueryAvailableCouponsResp, error) {
	// 1.查询可领取的优惠券（status=1 + is_enabled=1 + 在有效期内 + 未领完）
	var result []model.SmsCoupon
	pageNum := in.PageNum
	pageSize := in.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize

	sql := `SELECT * FROM sms_coupon 
			WHERE status = 1 
			  AND is_enabled = 1 
			  AND NOW() BETWEEN start_time AND end_time 
			  AND received_count < total_count 
			  AND is_deleted = 0
			ORDER BY create_time DESC
			LIMIT ? OFFSET ?`
	db := l.svcCtx.DB
	err := db.WithContext(l.ctx).Raw(sql, pageSize, offset).Find(&result).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询可领取优惠券列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	// 2.批量查询该会员已领取的优惠券记录（用于判断 receiveStatus）
	receivedMap := make(map[int64]int64)
	if in.MemberId > 0 {
		record := query.SmsCouponRecord
		memberRecords, memberErr := record.WithContext(l.ctx).Where(record.MemberID.Eq(in.MemberId)).Find()
		if memberErr != nil {
			logc.Errorf(l.ctx, "批量查询会员优惠券领取记录失败,memberId:%d,异常:%s", in.MemberId, memberErr.Error())
		} else {
			for _, r := range memberRecords {
				receivedMap[r.CouponID]++
			}
		}
	}

	// 3.批量查询 scope 信息
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

	// 4.组装返回数据
	var list []*smsclient.QueryCouponData
	for _, item := range result {
		// 计算领取状态
		var receiveStatus int32 = 0 // 默认可领取
		memberCount := receivedMap[item.ID]
		if memberCount >= int64(item.PerLimit) {
			receiveStatus = 1 // 已领取（达到限领上限）
		}

		list = append(list, &smsclient.QueryCouponData{
			Id:            item.ID,
			TypeId:        item.TypeID,
			Name:          item.Name,
			Code:          item.Code,
			Amount:        float32(item.Amount),
			MinAmount:     float32(item.MinAmount),
			StartTime:     time_util.TimeToStr(item.StartTime),
			EndTime:       time_util.TimeToStr(item.EndTime),
			PerLimit:      item.PerLimit,
			Status:        item.Status,
			Description:   item.Description,
			ScopeType:     scopeMap[item.ID],
			ReceiveStatus: receiveStatus,
			TotalCount:    item.TotalCount,
			ReceivedCount: item.ReceivedCount,
		})
	}

	return &smsclient.QueryAvailableCouponsResp{
		List: list,
	}, nil
}
