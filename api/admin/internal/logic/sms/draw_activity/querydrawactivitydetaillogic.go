package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDrawActivityDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDrawActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDrawActivityDetailLogic {
	return &QueryDrawActivityDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDrawActivityDetailLogic) QueryDrawActivityDetail(req *types.QueryDrawActivityDetailReq) (*types.QueryDrawActivityDetailResp, error) {
	queryScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.DrawActivityService.QueryDrawActivityDetail(l.ctx, &smsclient.QueryDrawActivityDetailReq{
		Id:    req.Id,
		Scope: common.SMSGovernanceScope(queryScope),
	})
	if err != nil {
		return nil, grpcDrawError(err)
	}
	return &types.QueryDrawActivityDetailResp{
		Code:    "000000",
		Message: "查询抽卡活动详情成功",
		Data: types.QueryDrawActivityDetailData{
			Id:                          result.Id,
			ActivityCode:                result.ActivityCode,
			Name:                        result.Name,
			RuleSummary:                 result.RuleSummary,
			StartTime:                   result.StartTime,
			EndTime:                     result.EndTime,
			RealNameRequired:            result.RealNameRequired,
			ParticipantConditionSummary: result.ParticipantConditionSummary,
			ConsumeRuleSummary:          result.ConsumeRuleSummary,
			ProbabilityRule:             result.ProbabilityRule,
			ComplianceRuleSummary:       result.ComplianceRuleSummary,
			CirculationLimitSummary:     result.CirculationLimitSummary,
			ApprovalRecordRef:           result.ApprovalRecordRef,
			CopyrightStatus:             result.CopyrightStatus,
			ContentAuditStatus:          result.ContentAuditStatus,
			Status:                      result.Status,
			AuditStatus:                 result.AuditStatus,
			IsEnabled:                   result.IsEnabled,
			PublishReadiness:            result.PublishReadiness,
			PublishFailureSummary:       result.PublishFailureSummary,
			ScopeType:                   result.ScopeType,
			PlatformId:                  result.PlatformId,
			TenantId:                    result.TenantId,
			MerchantId:                  result.MerchantId,
			CreateTime:                  result.CreateTime,
			UpdateTime:                  result.UpdateTime,
			CreateBy:                    result.CreateBy,
			UpdateBy:                    result.UpdateBy,
			HomeEntry: types.DrawHomeEntryConfig{
				ShowOnHome:         result.HomeEntry.ShowOnHome,
				HomeEntryTitle:     result.HomeEntry.HomeEntryTitle,
				HomeEntrySubtitle:  result.HomeEntry.HomeEntrySubtitle,
				HomeEntryImage:     result.HomeEntry.HomeEntryImage,
				HomeEntrySort:      result.HomeEntry.HomeEntrySort,
				LandingTargetType:  result.HomeEntry.LandingTargetType,
				LandingTargetValue: result.HomeEntry.LandingTargetValue,
				IsEnabled:          result.HomeEntry.IsEnabled,
			},
			Templates:      toAPITemplates(result.Templates),
			Pools:          toAPIPools(result.Pools),
			ReadinessItems: toAPIReadinessItems(result.ReadinessItems),
			Audits:         toAPIAudits(result.Audits),
		},
	}, nil
}
