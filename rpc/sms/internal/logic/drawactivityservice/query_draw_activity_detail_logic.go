package drawactivityservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryDrawActivityDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDrawActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDrawActivityDetailLogic {
	return &QueryDrawActivityDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryDrawActivityDetailLogic) QueryDrawActivityDetail(in *smsclient.QueryDrawActivityDetailReq) (*smsclient.QueryDrawActivityDetailResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	aggregate, err := loadDrawActivityAggregate(l.ctx, l.svcCtx.DB, currentScope, in.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("抽卡活动不存在")
		}
		logc.Errorf(l.ctx, "查询抽卡活动详情失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	return &smsclient.QueryDrawActivityDetailResp{
		Id:                          aggregate.Activity.ID,
		ActivityCode:                aggregate.Activity.ActivityCode,
		Name:                        aggregate.Activity.Name,
		RuleSummary:                 aggregate.Activity.RuleSummary,
		StartTime:                   time_util.TimeToStr(aggregate.Activity.StartTime),
		EndTime:                     time_util.TimeToStr(aggregate.Activity.EndTime),
		RealNameRequired:            aggregate.Activity.RealNameRequired,
		ParticipantConditionSummary: aggregate.Activity.ParticipantConditionSummary,
		ConsumeRuleSummary:          aggregate.Activity.ConsumeRuleSummary,
		ProbabilityRule:             aggregate.Activity.ProbabilityRule,
		ComplianceRuleSummary:       aggregate.Activity.ComplianceRuleSummary,
		CirculationLimitSummary:     aggregate.Activity.CirculationLimitSummary,
		ApprovalRecordRef:           aggregate.Activity.ApprovalRecordRef,
		CopyrightStatus:             aggregate.Activity.CopyrightStatus,
		ContentAuditStatus:          aggregate.Activity.ContentAuditStatus,
		Status:                      aggregate.Activity.Status,
		AuditStatus:                 aggregate.Activity.AuditStatus,
		IsEnabled:                   aggregate.Activity.IsEnabled,
		PublishReadiness:            aggregate.Readiness.PublishReadiness,
		PublishFailureSummary:       aggregate.Readiness.Summary,
		ScopeType:                   activityScopeType(aggregate.Activity.PlatformID, aggregate.Activity.TenantID, aggregate.Activity.MerchantID),
		PlatformId:                  aggregate.Activity.PlatformID,
		TenantId:                    aggregate.Activity.TenantID,
		MerchantId:                  aggregate.Activity.MerchantID,
		CreateTime:                  time_util.TimeToStr(aggregate.Activity.CreateTime),
		UpdateTime:                  time_util.TimeToString(aggregate.Activity.UpdateTime),
		CreateBy:                    aggregate.Activity.CreateBy,
		UpdateBy:                    safeInt64(aggregate.Activity.UpdateBy),
		HomeEntry:                   buildDrawHomeEntry(aggregate.Activity),
		Templates:                   aggregate.Templates,
		Pools:                       aggregate.Pools,
		ReadinessItems:              aggregate.Readiness.Items,
		Audits:                      aggregate.Audits,
	}, nil
}
