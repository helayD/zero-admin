package channel_integration_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChannelIntegrationTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChannelIntegrationTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChannelIntegrationTemplateLogic {
	return &CreateChannelIntegrationTemplateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateChannelIntegrationTemplateLogic) CreateChannelIntegrationTemplate(req *types.CreateChannelIntegrationTemplateReq) (*types.CreateChannelIntegrationTemplateResp, error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}

	queryScope, err := resolveTemplateMutationScope(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.ChannelIntegrationTemplateService.CreateChannelIntegrationTemplate(l.ctx, &sysclient.CreateChannelIntegrationTemplateReq{
		TemplateCode:         req.TemplateCode,
		TemplateName:         req.TemplateName,
		TemplateType:         req.TemplateType,
		TargetCode:           req.TargetCode,
		ScopeType:            queryScope.ScopeType,
		PlatformId:           queryScope.PlatformID,
		TenantId:             queryScope.TenantID,
		MerchantId:           queryScope.MerchantID,
		Status:               req.Status,
		MetadataConfig:       normalizeRawJSON(req.MetadataConfig),
		SecretRefConfig:      normalizeRawJSON(req.SecretRefConfig),
		IntentContractConfig: normalizeRawJSON(req.IntentContractConfig),
		ImpactScopeConfig:    normalizeRawJSON(req.ImpactScopeConfig),
		Remark:               req.Remark,
		CreateBy:             userName,
	})
	if err != nil {
		logc.Errorf(l.ctx, "创建渠道集成模板失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.CreateChannelIntegrationTemplateResp{
		Code:    "000000",
		Message: "创建渠道集成模板成功",
		Data: types.CreateChannelIntegrationTemplateData{
			Id: result.Id,
		},
	}, nil
}
