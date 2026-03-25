package homeadvertiseservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateHomeAdvertiseStatusLogic 更新首页轮播广告
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type UpdateHomeAdvertiseStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateHomeAdvertiseStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateHomeAdvertiseStatusLogic {
	return &UpdateHomeAdvertiseStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateHomeAdvertiseStatus 更新首页轮播广告状态
func (l *UpdateHomeAdvertiseStatusLogic) UpdateHomeAdvertiseStatus(in *smsclient.UpdateHomeAdvertiseStatusReq) (*smsclient.UpdateHomeAdvertiseStatusResp, error) {
	q := query.SmsHomeAdvertise

	// 上线（status=1）前校验
	if in.Status == 1 {
		ads, err := q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Find()
		if err != nil {
			logc.Errorf(l.ctx, "查询首页轮播广告失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("查询首页轮播广告失败")
		}
		now := time.Now()
		for _, ad := range ads {
			if ad.EndTime.Before(now) {
				return nil, errors.New("广告已过期，不允许上线")
			}
			if len(ad.Pic) == 0 {
				return nil, errors.New("请先上传图片")
			}
			if len(ad.URL) == 0 {
				return nil, errors.New("请填写链接地址")
			}
		}
	}

	_, err := q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Update(q.Status, in.Status)

	if err != nil {
		logc.Errorf(l.ctx, "更新首页轮播广告状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新首页轮播广告状态失败")
	}

	return &smsclient.UpdateHomeAdvertiseStatusResp{}, nil
}
