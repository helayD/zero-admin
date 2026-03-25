package seckillactivityservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateSeckillActivityLogic 更新秒杀活动
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:51
*/
type UpdateSeckillActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSeckillActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSeckillActivityLogic {
	return &UpdateSeckillActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateSeckillActivity 更新秒杀活动
func (l *UpdateSeckillActivityLogic) UpdateSeckillActivity(in *smsclient.UpdateSeckillActivityReq) (*smsclient.UpdateSeckillActivityResp, error) {
	// 2.3 基本字段校验
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("活动名称不能为空")
	}
	if len([]rune(in.Name)) > 100 {
		return nil, fmt.Errorf("活动名称不能超过100个字符")
	}
	if len([]rune(in.Description)) > 500 {
		return nil, fmt.Errorf("活动描述不能超过500个字符")
	}

	// 2.2 时间解析——不再吞掉错误
	startTime, err := time.Parse("2006-01-02 15:04:05", in.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式无效，请使用 yyyy-MM-dd HH:mm:ss 格式")
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", in.EndTime)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式无效，请使用 yyyy-MM-dd HH:mm:ss 格式")
	}
	if !startTime.Before(endTime) {
		return nil, fmt.Errorf("开始时间必须早于结束时间")
	}

	q := query.SmsSeckillActivity.WithContext(l.ctx)

	// 1.根据秒杀活动id查询秒杀活动是否已存在
	detail, err := q.Where(query.SmsSeckillActivity.ID.Eq(in.Id)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "秒杀活动不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("秒杀活动不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询秒杀活动异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询秒杀活动异常")
	}

	// 2.1 状态流转约束：已上线(status=0)的活动不允许修改时间段
	if detail.Status == 0 {
		if startTime != detail.StartTime || endTime != detail.EndTime {
			return nil, fmt.Errorf("已上线的活动不允许修改时间段，仅允许修改描述和启停")
		}
	}

	count, err := q.Where(query.SmsSeckillActivity.Name.Eq(in.Name), query.SmsSeckillActivity.ID.Neq(in.Id)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "更新秒杀活动失败,参数：%+v, 异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("更新秒杀活动失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("活动名称：%s,已存在", in.Name))
	}

	now := time.Now()
	item := &model.SmsSeckillActivity{
		ID:          in.Id,             // 编号
		Name:        in.Name,           // 活动名称
		Description: in.Description,    // 活动描述
		StartTime:   startTime,         // 开始时间
		EndTime:     endTime,           // 结束时间
		Status:      in.Status,         // 状态:0-上线,1-下线
		IsEnabled:   in.IsEnabled,      // 是否启用
		CreateBy:    detail.CreateBy,   // 创建人ID
		CreateTime:  detail.CreateTime, // 创建时间
		UpdateBy:    &in.UpdateBy,      // 更新人ID
		UpdateTime:  &now,              // 更新时间
	}

	// 2.秒杀活动存在时,则直接更新秒杀活动
	err = l.svcCtx.DB.Model(&model.SmsSeckillActivity{}).WithContext(l.ctx).Where(query.SmsSeckillActivity.ID.Eq(in.Id)).Save(item).Error

	if err != nil {
		logc.Errorf(l.ctx, "更新秒杀活动失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("更新秒杀活动失败")
	}

	return &smsclient.UpdateSeckillActivityResp{}, nil
}
