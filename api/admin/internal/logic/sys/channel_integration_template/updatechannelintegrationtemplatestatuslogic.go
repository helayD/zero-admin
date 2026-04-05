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

type UpdateChannelIntegrationTemplateStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChannelIntegrationTemplateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChannelIntegrationTemplateStatusLogic {
	return &UpdateChannelIntegrationTemplateStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateChannelIntegrationTemplateStatusLogic) UpdateChannelIntegrationTemplateStatus(req *types.UpdateChannelIntegrationTemplateStatusReq) (*types.BaseResp, error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	currentScope, err := currentTemplateGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	for _, id := range req.Ids {
		detail, detailErr := l.svcCtx.ChannelIntegrationTemplateService.QueryChannelIntegrationTemplateDetail(l.ctx, &sysclient.QueryChannelIntegrationTemplateDetailReq{Id: id})
		if detailErr != nil {
			logc.Errorf(l.ctx, "查询待更新模板状态详情失败, 模板ID:%d, 异常:%s", id, detailErr.Error())
			return nil, grpcError(detailErr)
		}
		if !allowTemplateMutationByScope(detail.Data, currentScope) {
			return nil, errorx.NewDefaultError("当前主体无权修改所选模板状态")
		}
	}

	_, err = l.svcCtx.ChannelIntegrationTemplateService.UpdateChannelIntegrationTemplateStatus(l.ctx, &sysclient.UpdateChannelIntegrationTemplateStatusReq{
		Ids:      req.Ids,
		Status:   req.Status,
		UpdateBy: userName,
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新渠道集成模板状态失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.BaseResp{Code: "000000", Message: "更新渠道集成模板状态成功"}, nil
}
