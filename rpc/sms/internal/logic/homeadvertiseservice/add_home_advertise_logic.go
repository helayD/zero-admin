package homeadvertiseservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// AddHomeAdvertiseLogic 添加首页轮播广告
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type AddHomeAdvertiseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddHomeAdvertiseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddHomeAdvertiseLogic {
	return &AddHomeAdvertiseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddHomeAdvertise 添加首页轮播广告
func (l *AddHomeAdvertiseLogic) AddHomeAdvertise(in *smsclient.AddHomeAdvertiseReq) (*smsclient.AddHomeAdvertiseResp, error) {
	// 字段校验
	if len(in.Name) == 0 {
		return nil, errors.New("广告名称不能为空")
	}
	if len(in.Name) > 100 {
		return nil, errors.New("广告名称长度不能超过100个字符")
	}
	if len(in.Pic) == 0 {
		return nil, errors.New("请先上传图片")
	}
	if len(in.Url) == 0 {
		return nil, errors.New("请填写链接地址")
	}

	// 时间解析与校验
	startTime, err := time.Parse("2006-01-02 15:04:05", in.StartTime)
	if err != nil {
		logc.Errorf(l.ctx, "添加首页轮播广告开始时间解析失败,参数：%+v, 异常:%s", in, err.Error())
		return nil, errors.New("开始时间格式不正确")
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", in.EndTime)
	if err != nil {
		logc.Errorf(l.ctx, "添加首页轮播广告结束时间解析失败,参数：%+v, 异常:%s", in, err.Error())
		return nil, errors.New("结束时间格式不正确")
	}
	if !startTime.Before(endTime) {
		return nil, errors.New("开始时间必须早于结束时间")
	}

	q := query.SmsHomeAdvertise
	count, err := q.WithContext(l.ctx).Where(q.Name.Eq(in.Name)).Count()

	if err != nil {
		logc.Errorf(l.ctx, "添加首页轮播广告失败,参数：%+v, 异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("添加首页轮播广告失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("首页轮播广告：%s,已存在", in.Name))
	}

	item := &model.SmsHomeAdvertise{
		Name:      in.Name,   // 名称
		Type:      in.Type,   // 轮播位置：0->PC首页轮播；1->app首页轮播
		Pic:       in.Pic,    // 图片地址
		StartTime: startTime, // 开始时间
		EndTime:   endTime,   // 结束时间
		Status:    0,         // 新建广告默认下线，需显式上线才投放
		URL:       in.Url,    // 链接地址
		Remark:    in.Remark, // 备注
		Sort:      in.Sort,   // 排序
	}

	err = q.WithContext(l.ctx).Create(item)
	if err != nil {
		logc.Errorf(l.ctx, "添加首页轮播广告失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加首页轮播广告失败")
	}

	return &smsclient.AddHomeAdvertiseResp{}, nil
}
