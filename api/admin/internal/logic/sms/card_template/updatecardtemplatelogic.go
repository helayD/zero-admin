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

type UpdateCardTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCardTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCardTemplateLogic {
	return &UpdateCardTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCardTemplateLogic) UpdateCardTemplate(req *types.UpdateCardTemplateReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	logc.Infof(l.ctx, "更新卡片模板，操作人：%d，模板ID：%d，治理范围：%+v", userId, req.Id, writeScope)

	rpcReq := &smsclient.UpdateCardTemplateReq{
		Id:                      req.Id,
		TemplateName:            req.TemplateName,
		CardFaceImage:           req.CardFaceImage,
		CopyrightOwner:          req.CopyrightOwner,
		CopyrightProofSummary:   req.CopyrightProofSummary,
		Rarity:                  req.Rarity,
		IssueLimit:              req.IssueLimit,
		DisplayCopy:             req.DisplayCopy,
		CirculationLimitSummary: req.CirculationLimitSummary,
		DisplayStatus:           req.DisplayStatus,
		ContentAuditStatus:      req.ContentAuditStatus,
		ProviderCode:            req.ProviderCode,
		CredentialRef:           req.CredentialRef,
		Scope:                   common.SMSGovernanceScope(writeScope),
		OperatorType:            "admin",
	}

	if _, err = l.svcCtx.CardTemplateService.UpdateCardTemplate(l.ctx, rpcReq); err != nil {
		logc.Errorf(l.ctx, "更新卡片模板失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.BaseResp{Code: "000000", Message: "更新卡片模板成功"}, nil
}
