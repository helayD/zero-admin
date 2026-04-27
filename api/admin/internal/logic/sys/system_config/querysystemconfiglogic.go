package system_config

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QuerySystemConfigLogic 查询系统配置
type QuerySystemConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuerySystemConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuerySystemConfigLogic {
	return &QuerySystemConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QuerySystemConfig 查询系统配置
func (l *QuerySystemConfigLogic) QuerySystemConfig() (*types.QuerySystemConfigResp, error) {
	data, err := QuerySystemConfig(l.ctx, l.svcCtx)
	if err != nil {
		logc.Errorf(l.ctx, "查询系统配置失败,异常:%s", err.Error())
		return nil, err
	}

	return &types.QuerySystemConfigResp{
		Code:    "000000",
		Message: "查询系统配置成功",
		Data:    data,
	}, nil
}
