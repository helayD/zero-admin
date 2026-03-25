package seckillactivityservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateSeckillActivityStatusLogic 更新秒杀活动
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:51
*/
type UpdateSeckillActivityStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSeckillActivityStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSeckillActivityStatusLogic {
	return &UpdateSeckillActivityStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateSeckillActivityStatus 更新秒杀活动状态
func (l *UpdateSeckillActivityStatusLogic) UpdateSeckillActivityStatus(in *smsclient.UpdateSeckillActivityStatusReq) (*smsclient.UpdateSeckillActivityStatusResp, error) {
	q := query.SmsSeckillActivity

	// 3.1 上线（status=0）前校验
	if in.Status == 0 {
		for _, id := range in.Ids {
			detail, err := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
			if err != nil {
				logc.Errorf(l.ctx, "查询秒杀活动失败,id:%d,异常:%s", id, err.Error())
				return nil, fmt.Errorf("秒杀活动(ID:%d)不存在", id)
			}

			// 校验 end_time 晚于当前时间
			if detail.EndTime.Before(time.Now()) {
				return nil, fmt.Errorf("秒杀活动「%s」已过期，不允许上线", detail.Name)
			}

			// 校验 is_enabled=1
			if detail.IsEnabled != 1 {
				return nil, fmt.Errorf("秒杀活动「%s」未启用，请先启用后再上线", detail.Name)
			}

			// 校验至少关联了一个已上架的秒杀商品
			productQ := query.SmsSeckillProduct
			productCount, err := productQ.WithContext(l.ctx).Where(
				productQ.ActivityID.Eq(id),
				productQ.Status.Eq(1),
			).Count()
			if err != nil {
				logc.Errorf(l.ctx, "查询秒杀商品失败,activityId:%d,异常:%s", id, err.Error())
				return nil, errors.New("查询秒杀商品失败")
			}
			if productCount == 0 {
				return nil, fmt.Errorf("秒杀活动「%s」未添加秒杀商品，请先添加秒杀商品后再上线", detail.Name)
			}
		}
	}

	_, err := q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Update(q.Status, in.Status)

	if err != nil {
		logc.Errorf(l.ctx, "更新秒杀活动状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新秒杀活动状态失败")
	}

	return &smsclient.UpdateSeckillActivityStatusResp{}, nil
}
