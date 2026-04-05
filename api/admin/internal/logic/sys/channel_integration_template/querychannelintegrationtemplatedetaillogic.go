package channel_integration_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChannelIntegrationTemplateDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryChannelIntegrationTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChannelIntegrationTemplateDetailLogic {
	return &QueryChannelIntegrationTemplateDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryChannelIntegrationTemplateDetailLogic) QueryChannelIntegrationTemplateDetail(req *types.QueryChannelIntegrationTemplateDetailReq) (*types.QueryChannelIntegrationTemplateDetailResp, error) {
	currentScope, err := currentTemplateGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.ChannelIntegrationTemplateService.QueryChannelIntegrationTemplateDetail(l.ctx, &sysclient.QueryChannelIntegrationTemplateDetailReq{Id: req.Id})
	if err != nil {
		logc.Errorf(l.ctx, "查询渠道集成模板详情失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}
	if !allowTemplateByScope(result.Data, currentScope) {
		return nil, errorx.NewDefaultError("当前主体无权查看该模板")
	}

	return &types.QueryChannelIntegrationTemplateDetailResp{
		Code:    "000000",
		Message: "查询渠道集成模板详情成功",
		Data:    mapChannelIntegrationTemplateDetail(result.Data),
	}, nil
}
