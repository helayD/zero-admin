package drawparticipationservicelogic

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

type QueryDrawActivityLandingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDrawActivityLandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDrawActivityLandingLogic {
	return &QueryDrawActivityLandingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryDrawActivityLandingLogic) QueryDrawActivityLanding(in *smsclient.QueryDrawActivityLandingReq) (*smsclient.QueryDrawActivityLandingResp, error) {
	if in.ActivityId <= 0 {
		return nil, errors.New("活动ID不能为空")
	}
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	activity, err := loadActivitySnapshot(l.ctx, l.svcCtx.DB, currentScope, in.ActivityId, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("抽卡活动不存在")
		}
		logc.Errorf(l.ctx, "查询抽卡活动落地页失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询抽卡活动落地页失败")
	}

	pools, err := loadPoolSnapshots(l.ctx, l.svcCtx.DB, activity.ID)
	if err != nil {
		return nil, logParticipationFailure(l.ctx, "查询卡池列表", in, errors.New("查询卡池列表失败"))
	}
	poolIDs := make([]int64, 0, len(pools))
	for _, pool := range pools {
		poolIDs = append(poolIDs, pool.ID)
	}
	poolTemplates, err := loadPoolTemplates(l.ctx, l.svcCtx.DB, activity.ID, poolIDs, false)
	if err != nil {
		return nil, logParticipationFailure(l.ctx, "查询卡池模板", in, errors.New("查询卡池模板失败"))
	}
	templateIDs := make([]int64, 0, len(poolTemplates))
	for _, item := range poolTemplates {
		templateIDs = append(templateIDs, item.TemplateID)
	}
	templateMap, err := loadCardTemplates(l.ctx, l.svcCtx.DB, templateIDs)
	if err != nil {
		return nil, logParticipationFailure(l.ctx, "查询卡片模板", in, errors.New("查询卡片模板失败"))
	}

	var (
		member     *drawMemberInfoSnapshot
		identity   *drawMemberIdentitySnapshot
		totalCount int64
		dailyCount int64
		myRecords  []*smsclient.DrawMemberRecordData
	)
	loggedIn := in.MemberId > 0
	if loggedIn {
		member, err = loadMemberInfoSnapshot(l.ctx, l.svcCtx.DB, in.MemberId, false)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, logParticipationFailure(l.ctx, "查询会员信息", in, errors.New("查询会员信息失败"))
		}
		identity, err = loadMemberIdentitySnapshot(l.ctx, l.svcCtx.DB, in.MemberId)
		if err != nil {
			logc.Errorf(l.ctx, "查询实名状态失败,参数:%+v,异常:%s", in, err.Error())
			identity = defaultMemberIdentitySnapshot(in.MemberId)
		}
		totalCount, dailyCount, err = countConsumedRecords(l.ctx, l.svcCtx.DB, activity.ID, in.MemberId)
		if err != nil {
			return nil, logParticipationFailure(l.ctx, "统计参与次数", in, errors.New("统计参与次数失败"))
		}
		_, myRecords, err = queryMemberRecordList(l.ctx, l.svcCtx.DB, activity.ID, in.MemberId, 1, 10)
		if err != nil {
			return nil, logParticipationFailure(l.ctx, "查询会员参与记录", in, errors.New("查询会员参与记录失败"))
		}
	}

	eligibility := buildEligibility(activity, member, identity, totalCount, dailyCount, hasAvailableInventory(poolTemplates), loggedIn)
	recentWins, err := loadRecentWins(l.ctx, l.svcCtx.DB, activity.ID, 10)
	if err != nil {
		return nil, logParticipationFailure(l.ctx, "查询中奖摘要", in, errors.New("查询中奖摘要失败"))
	}

	return &smsclient.QueryDrawActivityLandingResp{
		ActivityId:                  activity.ID,
		ActivityCode:                activity.ActivityCode,
		Name:                        activity.Name,
		RuleSummary:                 activity.RuleSummary,
		ParticipantConditionSummary: activity.ParticipantConditionSummary,
		ConsumeRuleSummary:          activity.ConsumeRuleSummary,
		ProbabilityRule:             activity.ProbabilityRule,
		ComplianceRuleSummary:       activity.ComplianceRuleSummary,
		CirculationLimitSummary:     activity.CirculationLimitSummary,
		StartTime:                   time_util.TimeToStr(activity.StartTime),
		EndTime:                     time_util.TimeToStr(activity.EndTime),
		RealNameRequired:            activity.RealNameRequired,
		EligibilityStatus:           eligibility.Status,
		EligibilityCode:             eligibility.Code,
		EligibilityMessage:          eligibility.Message,
		NextAction:                  eligibility.NextAction,
		RemainingLotteryTimes:       eligibility.RemainingLotteryTime,
		RealNameStatus:              eligibility.RealNameStatus,
		RealNameStatusText:          eligibility.RealNameStatusText,
		RealNameMasked:              eligibility.RealNameMasked,
		CredentialRef:               eligibility.CredentialRef,
		VerifiedAt:                  eligibility.VerifiedAt,
		CardPreviews:                buildCardPreviews(templateMap),
		Pools:                       buildLandingPools(pools, poolTemplates, templateMap),
		RecentWins:                  recentWins,
		MyRecords:                   myRecords,
	}, nil
}
