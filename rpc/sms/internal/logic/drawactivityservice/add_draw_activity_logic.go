package drawactivityservicelogic

import (
	"context"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AddDrawActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddDrawActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddDrawActivityLogic {
	return &AddDrawActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddDrawActivityLogic) AddDrawActivity(in *smsclient.AddDrawActivityReq) (*smsclient.AddDrawActivityResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}
	row, err := buildActivityRowFromAdd(in)
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		return saveDrawAggregate(l.ctx, tx, row, currentScope, in.Templates, in.Pools, in.CreateBy, in.OperatorName, "created", true)
	})
	if err != nil {
		logc.Errorf(l.ctx, "新增抽卡活动失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	return &smsclient.AddDrawActivityResp{Id: row.ID}, nil
}
