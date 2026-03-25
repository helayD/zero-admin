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
)

// AddSeckillActivityLogic 添加秒杀活动
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:51
*/
type AddSeckillActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddSeckillActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSeckillActivityLogic {
	return &AddSeckillActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddSeckillActivity 添加秒杀活动
func (l *AddSeckillActivityLogic) AddSeckillActivity(in *smsclient.AddSeckillActivityReq) (*smsclient.AddSeckillActivityResp, error) {
	// 1.1 基本字段校验
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("活动名称不能为空")
	}
	if len([]rune(in.Name)) > 100 {
		return nil, fmt.Errorf("活动名称不能超过100个字符")
	}
	if len([]rune(in.Description)) > 500 {
		return nil, fmt.Errorf("活动描述不能超过500个字符")
	}

	// 1.2 时间解析——不再吞掉错误
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

	q := query.SmsSeckillActivity

	count, err := q.WithContext(l.ctx).Where(q.Name.Eq(in.Name)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "添加秒杀活动失败,参数：%+v, 异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("添加秒杀活动失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("活动名称：%s,已存在", in.Name))
	}

	// 1.3 新建活动默认 status=1（下线/草稿），需要显式发布才上线
	item := &model.SmsSeckillActivity{
		Name:        in.Name,        // 活动名称
		Description: in.Description, // 活动描述
		StartTime:   startTime,      // 开始时间
		EndTime:     endTime,        // 结束时间
		Status:      1,              // 默认下线/草稿，需显式发布
		IsEnabled:   in.IsEnabled,   // 是否启用
		CreateBy:    in.CreateBy,    // 创建人ID
	}

	err = q.WithContext(l.ctx).Create(item)
	if err != nil {
		logc.Errorf(l.ctx, "添加秒杀活动失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加秒杀活动失败")
	}

	return &smsclient.AddSeckillActivityResp{}, nil
}
