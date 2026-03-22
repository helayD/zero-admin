package couponservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/pkg/pointerprocess"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// QueryCouponListLogic 查询优惠券列表
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type QueryCouponListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCouponListLogic {
	return &QueryCouponListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCouponList 查询优惠券列表
func (l *QueryCouponListLogic) QueryCouponList(in *smsclient.QueryCouponListReq) (*smsclient.QueryCouponListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		logc.Errorf(l.ctx, "查询优惠券列表scope非法,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优惠券列表失败")
	}

	q := pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.SmsCoupon{}),
		current,
		"",
	)
	if in.TypeId != 0 {
		q = q.Where("type_id = ?", in.TypeId)
	}
	if len(in.Name) > 0 {
		q = q.Where("name LIKE ?", "%"+in.Name+"%")
	}
	if len(in.Code) > 0 {
		q = q.Where("code LIKE ?", "%"+in.Code+"%")
	}

	if len(in.StartTime) > 0 {
		startTime, _ := time.Parse("2006-01-02 15:04:05", in.StartTime)
		q = q.Where("start_time >= ?", startTime)
	}
	if len(in.EndTime) > 0 {
		endTime, _ := time.Parse("2006-01-02 15:04:05", in.EndTime)
		q = q.Where("end_time <= ?", endTime)
	}

	if in.Status != 4 {
		q = q.Where("status = ?", in.Status)
	}
	if in.IsEnabled != 2 {
		q = q.Where("is_enabled = ?", in.IsEnabled)
	}

	var (
		result []model.SmsCoupon
		count  int64
	)
	err = q.Session(&gorm.Session{}).Count(&count).Error
	if err == nil {
		err = q.Offset(int((in.PageNum - 1) * in.PageSize)).Limit(int(in.PageSize)).Find(&result).Error
	}

	if err != nil {
		logc.Errorf(l.ctx, "查询优惠券列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优惠券列表失败")
	}

	var list []*smsclient.CouponListData

	for _, item := range result {
		list = append(list, &smsclient.CouponListData{
			Id:            item.ID,                                          // 优惠券ID
			TypeId:        item.TypeID,                                      // 优惠券类型ID
			Name:          item.Name,                                        // 优惠券名称
			Code:          item.Code,                                        // 优惠券码
			Amount:        float32(item.Amount),                             // 优惠金额/折扣率
			MinAmount:     float32(item.MinAmount),                          // 最低使用金额
			StartTime:     time_util.TimeToStr(item.StartTime),              // 开始时间
			EndTime:       time_util.TimeToStr(item.EndTime),                // 结束时间
			TotalCount:    item.TotalCount,                                  // 发放总量
			ReceivedCount: item.ReceivedCount,                               // 已领取数量
			UsedCount:     item.UsedCount,                                   // 已使用数量
			PerLimit:      item.PerLimit,                                    // 每人限领数量
			Status:        item.Status,                                      // 状态：0-未开始，1-进行中，2-已结束，3-已取消
			IsEnabled:     item.IsEnabled,                                   // 是否启用
			Description:   item.Description,                                 // 使用说明
			CreateBy:      item.CreateBy,                                    // 创建人ID
			CreateTime:    time_util.TimeToStr(item.CreateTime),             // 创建时间
			UpdateBy:      pointerprocess.DefaltData(item.UpdateBy).(int64), // 更新人ID
			UpdateTime:    time_util.TimeToString(item.UpdateTime),          // 更新时间

		})

		isExpired := item.EndTime.Before(time.Now())
		if isExpired {
			record := query.SmsCouponRecord
			_, _ = record.WithContext(l.ctx).Where(record.CouponID.Eq(item.ID), record.Status.Eq(0)).Update(record.Status, 2)
		}
	}

	return &smsclient.QueryCouponListResp{
		Total: count,
		List:  list,
	}, nil
}
