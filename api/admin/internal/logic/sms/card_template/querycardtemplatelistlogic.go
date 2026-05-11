package card_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type QueryCardTemplateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCardTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCardTemplateListLogic {
	return &QueryCardTemplateListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryCardTemplateListLogic) QueryCardTemplateList(req *types.QueryCardTemplateListReq) (resp *types.QueryCardTemplateListResp, err error) {
	readScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	// Story 10.10 第二轮 Review 修复 H1: 前端 ProTable 用 current 字段，admin-api 接收后映射到 RPC 的 Page。
	rpcReq := &smsclient.QueryCardTemplateListReq{
		TemplateName:  req.TemplateName,
		TemplateCode:  req.TemplateCode,
		Rarity:        req.Rarity,
		Status:        req.Status,
		DisplayStatus: req.DisplayStatus,
		Page:          req.Current,
		PageSize:      req.PageSize,
		Scope:         common.SMSGovernanceScope(readScope),
	}

	result, err := l.svcCtx.CardTemplateService.QueryCardTemplateList(l.ctx, rpcReq)
	if err != nil {
		logc.Errorf(l.ctx, "查询卡片模板列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	list := make([]types.CardTemplateData, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toCardTemplateData(item))
	}

	return &types.QueryCardTemplateListResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CardTemplateListData{
			List:     list,
			Total:    result.Total,
			Page:     req.Current,
			PageSize: req.PageSize,
		},
	}, nil
}

func toCardTemplateData(item *smsclient.CardTemplateData) types.CardTemplateData {
	return types.CardTemplateData{
		Id:                      item.Id,
		TemplateCode:            item.TemplateCode,
		TemplateName:            item.TemplateName,
		CardFaceImage:           item.CardFaceImage,
		CopyrightOwner:          item.CopyrightOwner,
		CopyrightProofSummary:   item.CopyrightProofSummary,
		Rarity:                  item.Rarity,
		IssueLimit:              item.IssueLimit,
		DisplayCopy:             item.DisplayCopy,
		CirculationLimitSummary: item.CirculationLimitSummary,
		DisplayStatus:           item.DisplayStatus,
		ContentAuditStatus:      item.ContentAuditStatus,
		ProviderCode:            item.ProviderCode,
		CredentialRef:           item.CredentialRef,
		Status:                  item.Status,
		AuditStatus:             item.AuditStatus,
		PlatformId:              item.PlatformId,
		TenantId:                item.TenantId,
		MerchantId:              item.MerchantId,
		CreateBy:                item.CreateBy,
		UpdateBy:                item.UpdateBy,

		CreateTime:   item.CreateTime,
		UpdateTime:   item.UpdateTime,
		RefRuleCount: item.RefRuleCount,
	}
}
