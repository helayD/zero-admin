package seckillsessionservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteSeckillSessionLogic 删除秒杀场次
/*
Author: LiuFeiHua
Date: 2025/06/11 10:29:58
*/
type DeleteSeckillSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSeckillSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSeckillSessionLogic {
	return &DeleteSeckillSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteSeckillSession 删除秒杀场次
func (l *DeleteSeckillSessionLogic) DeleteSeckillSession(in *smsclient.DeleteSeckillSessionReq) (*smsclient.DeleteSeckillSessionResp, error) {
	q := query.SmsSeckillSession

	// 6.3 关联保护：有秒杀商品关联的场次不允许删除
	productQ := query.SmsSeckillProduct
	for _, id := range in.Ids {
		productCount, err := productQ.WithContext(l.ctx).Where(productQ.SessionID.Eq(id)).Count()
		if err != nil {
			logc.Errorf(l.ctx, "查询场次关联商品失败,sessionId:%d,异常:%s", id, err.Error())
			return nil, errors.New("查询场次关联商品失败")
		}
		if productCount > 0 {
			sessionDetail, _ := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
			name := fmt.Sprintf("ID:%d", id)
			if sessionDetail != nil {
				name = sessionDetail.Name
			}
			return nil, fmt.Errorf("场次「%s」已关联秒杀商品，请先移除关联商品后再删除", name)
		}
	}

	_, err := q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除秒杀场次失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除秒杀场次失败")
	}

	return &smsclient.DeleteSeckillSessionResp{}, nil
}
