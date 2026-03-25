package seckillactivityservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/pkg/pointerprocess"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QuerySeckillActivityListLogic 查询秒杀活动列表
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:51
*/
type QuerySeckillActivityListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQuerySeckillActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuerySeckillActivityListLogic {
	return &QuerySeckillActivityListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QuerySeckillActivityList 查询秒杀活动列表
func (l *QuerySeckillActivityListLogic) QuerySeckillActivityList(in *smsclient.QuerySeckillActivityListReq) (*smsclient.QuerySeckillActivityListResp, error) {
	seckillActivity := query.SmsSeckillActivity
	q := seckillActivity.WithContext(l.ctx)
	if len(in.Name) > 0 {
		q = q.Where(seckillActivity.Name.Like("%" + in.Name + "%"))
	}
	if len(in.StartTime) > 0 {
		startTime, parseErr := time.Parse("2006-01-02 15:04:05", in.StartTime)
		if parseErr != nil {
			logc.Errorf(l.ctx, "查询秒杀活动列表-开始时间解析失败,参数:%s,异常:%s", in.StartTime, parseErr.Error())
		} else {
			q = q.Where(seckillActivity.StartTime.Gte(startTime))
		}
	}
	if len(in.EndTime) > 0 {
		endTime, parseErr := time.Parse("2006-01-02 15:04:05", in.EndTime)
		if parseErr != nil {
			logc.Errorf(l.ctx, "查询秒杀活动列表-结束时间解析失败,参数:%s,异常:%s", in.EndTime, parseErr.Error())
		} else {
			q = q.Where(seckillActivity.EndTime.Lte(endTime))
		}
	}
	if in.Status != 2 {
		q = q.Where(seckillActivity.Status.Eq(in.Status))
	}
	if in.IsEnabled != 2 {
		q = q.Where(seckillActivity.IsEnabled.Eq(in.IsEnabled))
	}

	result, count, err := q.FindByPage(int((in.PageNum-1)*in.PageSize), int(in.PageSize))

	if err != nil {
		logc.Errorf(l.ctx, "查询秒杀活动列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询秒杀活动列表失败")
	}

	// 5.1 批量收集活动ID，预查询 productCount 和 sessionCount
	var activityIds []int64
	for _, item := range result {
		activityIds = append(activityIds, item.ID)
	}

	productCountMap := make(map[int64]int64)
	sessionCountMap := make(map[int64]int64)
	if len(activityIds) > 0 {
		// 查询每个活动关联的已上架秒杀商品数量
		type countResult struct {
			ActivityID int64 `gorm:"column:activity_id"`
			Cnt        int64 `gorm:"column:cnt"`
		}
		var productCounts []countResult
		err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.SmsSeckillProduct{}).
			Select("activity_id, count(*) as cnt").
			Where("activity_id IN ? AND status = 1 AND is_deleted = 0", activityIds).
			Group("activity_id").
			Find(&productCounts).Error
		if err != nil {
			logc.Errorf(l.ctx, "查询秒杀商品数量失败,异常:%s", err.Error())
		} else {
			for _, pc := range productCounts {
				productCountMap[pc.ActivityID] = pc.Cnt
			}
		}

		// 查询每个活动关联的场次数量（通过秒杀商品表的 session_id 去重）
		var sessionCounts []countResult
		err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.SmsSeckillProduct{}).
			Select("activity_id, count(distinct session_id) as cnt").
			Where("activity_id IN ? AND is_deleted = 0", activityIds).
			Group("activity_id").
			Find(&sessionCounts).Error
		if err != nil {
			logc.Errorf(l.ctx, "查询秒杀场次数量失败,异常:%s", err.Error())
		} else {
			for _, sc := range sessionCounts {
				sessionCountMap[sc.ActivityID] = sc.Cnt
			}
		}

	}

	var list []*smsclient.SeckillActivityListData

	for _, item := range result {
		list = append(list, &smsclient.SeckillActivityListData{
			Id:           item.ID,                                          // 编号
			Name:         item.Name,                                        // 活动名称
			Description:  item.Description,                                 // 活动描述
			StartTime:    time_util.TimeToStr(item.StartTime),              // 开始时间
			EndTime:      time_util.TimeToStr(item.EndTime),                // 结束时间
			Status:       item.Status,                                      // 状态:0-上线,1-下线
			IsEnabled:    item.IsEnabled,                                   // 是否启用
			CreateBy:     item.CreateBy,                                    // 创建人ID
			CreateTime:   time_util.TimeToStr(item.CreateTime),             // 创建时间
			UpdateBy:     pointerprocess.DefaltData(item.UpdateBy).(int64), // 更新人ID
			UpdateTime:   time_util.TimeToString(item.UpdateTime),          // 更新时间
			ProductCount: productCountMap[item.ID],                         // 关联已上架秒杀商品数量
			SessionCount: sessionCountMap[item.ID],                         // 关联场次数量

		})
	}

	return &smsclient.QuerySeckillActivityListResp{
		Total: count,
		List:  list,
	}, nil
}
