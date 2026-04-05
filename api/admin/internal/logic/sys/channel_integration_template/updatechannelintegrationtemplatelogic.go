package channel_integration_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChannelIntegrationTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChannelIntegrationTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChannelIntegrationTemplateLogic {
	return &UpdateChannelIntegrationTemplateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateChannelIntegrationTemplateLogic) UpdateChannelIntegrationTemplate(req *types.UpdateChannelIntegrationTemplateReq) (*types.BaseResp, error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	currentScope, err := currentTemplateGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	nextScope, err := resolveTemplateMutationScope(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}
	existing, err := l.svcCtx.ChannelIntegrationTemplateService.QueryChannelIntegrationTemplateDetail(l.ctx, &sysclient.QueryChannelIntegrationTemplateDetailReq{Id: req.Id})
	if err != nil {
		logc.Errorf(l.ctx, "查询待更新模板详情失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}
	if !allowTemplateMutationByScope(existing.Data, currentScope) {
		return nil, errorx.NewDefaultError("当前主体无权修改该模板")
	}

	_, err = l.svcCtx.ChannelIntegrationTemplateService.UpdateChannelIntegrationTemplate(l.ctx, &sysclient.UpdateChannelIntegrationTemplateReq{
		Id:                   req.Id,
		TemplateCode:         req.TemplateCode,
		TemplateName:         req.TemplateName,
		TemplateType:         req.TemplateType,
		TargetCode:           req.TargetCode,
		ScopeType:            nextScope.ScopeType,
		PlatformId:           nextScope.PlatformID,
		TenantId:             nextScope.TenantID,
		MerchantId:           nextScope.MerchantID,
		Status:               req.Status,
		MetadataConfig:       normalizeRawJSON(req.MetadataConfig),
		SecretRefConfig:      normalizeRawJSON(req.SecretRefConfig),
		IntentContractConfig: normalizeRawJSON(req.IntentContractConfig),
		ImpactScopeConfig:    normalizeRawJSON(req.ImpactScopeConfig),
		Remark:               req.Remark,
		UpdateBy:             userName,
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新渠道集成模板失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.BaseResp{Code: "000000", Message: "更新渠道集成模板成功"}, nil
}
