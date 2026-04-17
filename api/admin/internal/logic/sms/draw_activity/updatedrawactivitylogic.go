package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDrawActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDrawActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDrawActivityLogic {
	return &UpdateDrawActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDrawActivityLogic) UpdateDrawActivity(req *types.UpdateDrawActivityReq) (*types.BaseResp, error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	userName, err := common.GetUserName(l.ctx)
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
	_, err = l.svcCtx.DrawActivityService.UpdateDrawActivity(l.ctx, &smsclient.UpdateDrawActivityReq{
		Id:                          req.Id,
		Scope:                       common.SMSGovernanceScope(writeScope),
		ActivityCode:                req.ActivityCode,
		Name:                        req.Name,
		RuleSummary:                 req.RuleSummary,
		StartTime:                   req.StartTime,
		EndTime:                     req.EndTime,
		RealNameRequired:            req.RealNameRequired,
		ParticipantConditionSummary: req.ParticipantConditionSummary,
		ConsumeRuleSummary:          req.ConsumeRuleSummary,
		ProbabilityRule:             req.ProbabilityRule,
		ComplianceRuleSummary:       req.ComplianceRuleSummary,
		CirculationLimitSummary:     req.CirculationLimitSummary,
		ApprovalRecordRef:           req.ApprovalRecordRef,
		CopyrightStatus:             req.CopyrightStatus,
		ContentAuditStatus:          req.ContentAuditStatus,
		AuditStatus:                 req.AuditStatus,
		Status:                      req.Status,
		IsEnabled:                   req.IsEnabled,
		OperatorName:                userName,
		UpdateBy:                    userId,
		HomeEntry:                   toSMSHomeEntry(req.HomeEntry),
		Templates:                   toSMSTemplates(req.Templates),
		Pools:                       toSMSPools(req.Pools),
	})
	if err != nil {
		return nil, grpcDrawError(err)
	}
	return &types.BaseResp{Code: "000000", Message: "更新抽卡活动成功"}, nil
}
