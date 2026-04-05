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

type QueryChannelIntegrationTemplateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryChannelIntegrationTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChannelIntegrationTemplateListLogic {
	return &QueryChannelIntegrationTemplateListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryChannelIntegrationTemplateListLogic) QueryChannelIntegrationTemplateList(req *types.QueryChannelIntegrationTemplateListReq) (*types.QueryChannelIntegrationTemplateListResp, error) {
	queryScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.ChannelIntegrationTemplateService.QueryChannelIntegrationTemplateList(l.ctx, &sysclient.QueryChannelIntegrationTemplateListReq{
		PageNum:      req.Current,
		PageSize:     req.PageSize,
		TemplateName: req.TemplateName,
		TemplateType: req.TemplateType,
		TargetCode:   req.TargetCode,
		Status:       req.Status,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询渠道集成模板列表失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	list := make([]*types.ChannelIntegrationTemplateListData, 0, len(result.List))
	for _, item := range result.List {
		if !allowTemplateByScope(item, queryScope) {
			continue
		}
		if !matchTemplateExtraFilters(item, req.TenantId, req.MerchantId) {
			continue
		}
		list = append(list, mapChannelIntegrationTemplateListItem(item))
	}
	total := result.Total
	if len(list) != len(result.List) || req.TenantId > 0 || req.MerchantId > 0 {
		total = int64(len(list))
	}

	return &types.QueryChannelIntegrationTemplateListResp{
		Code:     "000000",
		Message:  "查询渠道集成模板列表成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    total,
	}, nil
}
