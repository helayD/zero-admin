package seckillactivityservicelogic

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
	"gorm.io/gorm"
)

// DeleteSeckillActivityLogic 删除秒杀活动
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:51
*/
type DeleteSeckillActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSeckillActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSeckillActivityLogic {
	return &DeleteSeckillActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteSeckillActivity 删除秒杀活动
func (l *DeleteSeckillActivityLogic) DeleteSeckillActivity(in *smsclient.DeleteSeckillActivityReq) (*smsclient.DeleteSeckillActivityResp, error) {
	q := query.SmsSeckillActivity

	// 4.1 已上线且在有效期内的活动不允许删除，需先下线
	now := time.Now()
	for _, id := range in.Ids {
		detail, err := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
		if err != nil {
			logc.Errorf(l.ctx, "查询秒杀活动失败,id:%d,异常:%s", id, err.Error())
			return nil, fmt.Errorf("秒杀活动(ID:%d)不存在", id)
		}
		if detail.Status == 0 && detail.EndTime.After(now) {
			return nil, fmt.Errorf("秒杀活动「%s」已上线且在有效期内，请先下线后再删除", detail.Name)
		}
	}

	// 4.2 删除活动时同步删除关联的秒杀商品记录（事务保护）
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if txErr := tx.Where("activity_id IN ? AND is_deleted = 0", in.Ids).Delete(&model.SmsSeckillProduct{}).Error; txErr != nil {
			logc.Errorf(l.ctx, "删除秒杀活动关联商品失败,参数:%+v,异常:%s", in, txErr.Error())
			return txErr
		}
		if txErr := tx.Where("id IN ?", in.Ids).Delete(&model.SmsSeckillActivity{}).Error; txErr != nil {
			logc.Errorf(l.ctx, "删除秒杀活动失败,参数:%+v,异常:%s", in, txErr.Error())
			return txErr
		}
		return nil
	})

	if err != nil {
		return nil, errors.New("删除秒杀活动失败")
	}

	return &smsclient.DeleteSeckillActivityResp{}, nil
}
