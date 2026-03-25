package seckillsessionservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// AddSeckillSessionLogic 添加秒杀场次
/*
Author: LiuFeiHua
Date: 2025/06/11 10:29:58
*/
type AddSeckillSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddSeckillSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSeckillSessionLogic {
	return &AddSeckillSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddSeckillSession 添加秒杀场次
func (l *AddSeckillSessionLogic) AddSeckillSession(in *smsclient.AddSeckillSessionReq) (*smsclient.AddSeckillSessionResp, error) {
	q := query.SmsSeckillSession

	// 基本校验
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("场次名称不能为空")
	}
	if strings.TrimSpace(in.StartTime) == "" || strings.TrimSpace(in.EndTime) == "" {
		return nil, fmt.Errorf("场次开始时间和结束时间不能为空")
	}
	if in.StartTime >= in.EndTime {
		return nil, fmt.Errorf("场次开始时间必须早于结束时间")
	}

	count, err := q.WithContext(l.ctx).Where(q.Name.Eq(in.Name)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "添加秒杀场次失败,参数：%+v, 异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("添加秒杀场次失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("秒杀场次：%s,已存在", in.Name))
	}

	// 6.1 时间段冲突校验：新建场次的 start_time ~ end_time 不可与已有启用场次重叠
	// HH:mm:ss 格式的字典序等于时间序，可直接用字符串比较
	existingSessions, err := q.WithContext(l.ctx).Where(q.Status.Eq(1)).Find()
	if err != nil {
		logc.Errorf(l.ctx, "查询已有场次失败,异常:%s", err.Error())
		return nil, errors.New("查询已有场次失败")
	}
	for _, s := range existingSessions {
		// 两个时间段 [A_start, A_end) 和 [B_start, B_end) 重叠的条件：A_start < B_end && B_start < A_end
		if in.StartTime < s.EndTime && s.StartTime < in.EndTime {
			return nil, fmt.Errorf("场次时间段与已有启用场次「%s」(%s~%s)冲突", s.Name, s.StartTime, s.EndTime)
		}
	}

	item := &model.SmsSeckillSession{
		Name:      in.Name,      // 场次名称
		StartTime: in.StartTime, // 开始时间
		EndTime:   in.EndTime,   // 结束时间
		Status:    in.Status,    // 状态：0-禁用，1-启用
		Sort:      in.Sort,      // 排序
		CreateBy:  in.CreateBy,  // 创建人ID
	}

	err = q.WithContext(l.ctx).Create(item)
	if err != nil {
		logc.Errorf(l.ctx, "添加秒杀场次失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加秒杀场次失败")
	}

	return &smsclient.AddSeckillSessionResp{}, nil
}
