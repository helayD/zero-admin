package couponservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/pointerprocess"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCouponByScopeIdLogic 根据商品Id和分类id查询可用的优惠券
/*
Author: LiuFeiHua
Date: 2025/6/18 11:30
*/
type QueryCouponByScopeIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCouponByScopeIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCouponByScopeIdLogic {
	return &QueryCouponByScopeIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCouponByScopeId 根据商品Id和分类id查询可用的优惠券
func (l *QueryCouponByScopeIdLogic) QueryCouponByScopeId(in *smsclient.QueryCouponByScopeIdReq) (*smsclient.QueryCouponByScopeIdResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		logc.Errorf(l.ctx, "根据商品Id和分类id查询可用的优惠券scope非法,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("根据商品Id和分类id查询可用的优惠券(app)失败")
	}

	sql := `select distinct t1.*
	from sms_coupon t1
		 join sms_coupon_scope t2 on t1.id = t2.coupon_id
	where (t2.scope_type = 0 or t2.scope_id in (?))
	  and t1.platform_id = ?
	  and t1.tenant_id = ?
	  and t1.merchant_id = ?
	  and t1.status = 1
	  and t1.is_enabled = 1
	  and t1.received_count < t1.total_count
	  and now() between t1.start_time and t1.end_time`

	var result []model.SmsCoupon
	db := l.svcCtx.DB
	err = db.WithContext(l.ctx).Raw(sql, in.ScopeIds, current.PlatformID, current.TenantID, current.MerchantID).Scan(&result).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据商品Id和分类id查询可用的优惠券(app)失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("根据商品Id和分类id查询可用的优惠券(app)失败")
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

		// 过期记录状态更新已移至定时Job处理，查询接口不再执行写操作
	}

	return &smsclient.QueryCouponByScopeIdResp{
		List: list,
	}, nil
}
