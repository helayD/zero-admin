package draw_activity

import (
	"context"
	"strings"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"
)

const (
	drawCodeSuccess            = "SUCCESS"
	drawCodeNeedLogin          = "DRAW_NEED_LOGIN"
	drawCodeNeedRealName       = "DRAW_NEED_REAL_NAME"
	drawCodeQuotaExhausted     = "DRAW_QUOTA_EXHAUSTED"
	drawCodeActivityOffline    = "DRAW_ACTIVITY_OFFLINE"
	drawCodeInventoryExhausted = "DRAW_INVENTORY_EXHAUSTED"
	drawCodeMemberDisabled     = "DRAW_MEMBER_DISABLED"
)

func rpcError(ctx context.Context, action string, payload interface{}, err error) error {
	logc.Errorf(ctx, "%s失败,参数:%+v,异常:%s", action, payload, err.Error())
	s, _ := status.FromError(err)
	message := strings.TrimSpace(s.Message())
	if message == "" {
		message = err.Error()
	}
	return errorx.NewDefaultError(message)
}

func apiCodeFromEligibility(code string) string {
	switch strings.TrimSpace(code) {
	case "need_login":
		return drawCodeNeedLogin
	case "need_real_name":
		return drawCodeNeedRealName
	case "quota_exhausted":
		return drawCodeQuotaExhausted
	case "activity_offline":
		return drawCodeActivityOffline
	case "inventory_exhausted":
		return drawCodeInventoryExhausted
	case "member_disabled":
		return drawCodeMemberDisabled
	default:
		return drawCodeSuccess
	}
}

func mapEligibilitySummary(resp *smsclient.PreviewDrawEligibilityResp) types.DrawEligibilitySummary {
	if resp == nil {
		return types.DrawEligibilitySummary{}
	}
	return types.DrawEligibilitySummary{
		EligibilityStatus:     resp.EligibilityStatus,
		EligibilityCode:       resp.EligibilityCode,
		EligibilityMessage:    resp.EligibilityMessage,
		NextAction:            resp.NextAction,
		RemainingLotteryTimes: resp.RemainingLotteryTimes,
	}
}

func mapIdentitySummary(statusText, realNameStatus, realNameMasked, credentialRef, verifiedAt string) types.DrawIdentitySummary {
	return types.DrawIdentitySummary{
		RealNameStatus:     realNameStatus,
		RealNameStatusText: statusText,
		RealNameMasked:     realNameMasked,
		CredentialRef:      credentialRef,
		VerifiedAt:         verifiedAt,
	}
}

func mapCardPreviews(items []*smsclient.DrawLandingCardPreview) []types.DrawCardPreview {
	result := make([]types.DrawCardPreview, 0, len(items))
	for _, item := range items {
		result = append(result, types.DrawCardPreview{
			TemplateId:    item.TemplateId,
			TemplateCode:  item.TemplateCode,
			TemplateName:  item.TemplateName,
			CardFaceImage: item.CardFaceImage,
			Rarity:        item.Rarity,
			DisplayCopy:   item.DisplayCopy,
			SlotIndex:     item.SlotIndex,
			Probability:   item.Probability,
		})
	}
	return result
}

func mapPoolPreviews(items []*smsclient.DrawLandingPoolPreview) []types.DrawPoolPreview {
	result := make([]types.DrawPoolPreview, 0, len(items))
	for _, item := range items {
		result = append(result, types.DrawPoolPreview{
			PoolId:          item.PoolId,
			PoolName:        item.PoolName,
			ProbabilityRule: item.ProbabilityRule,
			Cards:           mapCardPreviews(item.Cards),
		})
	}
	return result
}

func mapRecentWins(items []*smsclient.DrawLandingRecentWin) []types.DrawRecentWin {
	result := make([]types.DrawRecentWin, 0, len(items))
	for _, item := range items {
		result = append(result, types.DrawRecentWin{
			RecordId:         item.RecordId,
			MemberId:         item.MemberId,
			MemberNameMasked: item.MemberNameMasked,
			ResultStatus:     item.ResultStatus,
			ResultStatusText: item.ResultStatusText,
			TemplateName:     item.TemplateName,
			Rarity:           item.Rarity,
			CreateTime:       item.CreateTime,
		})
	}
	return result
}

func mapDrawRecords(items []*smsclient.DrawMemberRecordData) []types.DrawMemberRecord {
	result := make([]types.DrawMemberRecord, 0, len(items))
	for _, item := range items {
		result = append(result, mapDrawRecord(item))
	}
	return result
}

func mapDrawRecord(item *smsclient.DrawMemberRecordData) types.DrawMemberRecord {
	if item == nil {
		return types.DrawMemberRecord{}
	}
	return types.DrawMemberRecord{
		Id:                 item.Id,
		ActivityId:         item.ActivityId,
		RequestId:          item.RequestId,
		ResultType:         item.ResultType,
		ResultStatus:       item.ResultStatus,
		ResultStatusText:   item.ResultStatusText,
		FailureCode:        item.FailureCode,
		FailureReason:      item.FailureReason,
		PoolId:             item.PoolId,
		TemplateId:         item.TemplateId,
		TemplateName:       item.TemplateName,
		Rarity:             item.Rarity,
		ConsumeAmount:      item.ConsumeAmount,
		LotteryTimesBefore: item.LotteryTimesBefore,
		LotteryTimesAfter:  item.LotteryTimesAfter,
		AssetInstanceId:    item.AssetInstanceId,
		AssetNo:            item.AssetNo,
		AssetStatus:        item.AssetStatus,
		AssetStatusText:    item.AssetStatusText,
		AssetCreatedAt:     item.AssetCreatedAt,
		CreateTime:         item.CreateTime,
	}
}

func mapLandingResponse(resp *smsclient.QueryDrawActivityLandingResp) *types.DrawActivityLandingResp {
	if resp == nil {
		return &types.DrawActivityLandingResp{
			Code:    drawCodeActivityOffline,
			Message: "活动不存在或不可访问",
		}
	}

	return &types.DrawActivityLandingResp{
		Code:    apiCodeFromEligibility(resp.EligibilityCode),
		Message: firstNonEmpty(strings.TrimSpace(resp.EligibilityMessage), "查询抽卡活动成功"),
		Data: types.DrawActivityLandingData{
			ActivityId:                  resp.ActivityId,
			ActivityCode:                resp.ActivityCode,
			Name:                        resp.Name,
			RuleSummary:                 resp.RuleSummary,
			ParticipantConditionSummary: resp.ParticipantConditionSummary,
			ConsumeRuleSummary:          resp.ConsumeRuleSummary,
			ConsumeType:                 resp.ConsumeType,
			ConsumeAmount:               resp.ConsumeAmount,
			ProbabilityRule:             resp.ProbabilityRule,
			ComplianceRuleSummary:       resp.ComplianceRuleSummary,
			CirculationLimitSummary:     resp.CirculationLimitSummary,
			StartTime:                   resp.StartTime,
			EndTime:                     resp.EndTime,
			RealNameRequired:            resp.RealNameRequired,
			Eligibility: types.DrawEligibilitySummary{
				EligibilityStatus:     resp.EligibilityStatus,
				EligibilityCode:       resp.EligibilityCode,
				EligibilityMessage:    resp.EligibilityMessage,
				NextAction:            resp.NextAction,
				RemainingLotteryTimes: resp.RemainingLotteryTimes,
			},
			Identity: mapIdentitySummary(
				resp.RealNameStatusText,
				resp.RealNameStatus,
				resp.RealNameMasked,
				resp.CredentialRef,
				resp.VerifiedAt,
			),
			CardPreviews: mapCardPreviews(resp.CardPreviews),
			Pools:        mapPoolPreviews(resp.Pools),
			RecentWins:   mapRecentWins(resp.RecentWins),
			MyRecords:    mapDrawRecords(resp.MyRecords),
		},
	}
}

func buildDrawScope(ctx context.Context) *smsclient.GovernanceScope {
	current := frontcommon.ResolveEffectiveGovernanceScope(ctx)
	return frontcommon.SMSGovernanceScope(current)
}

func queryLanding(ctx context.Context, svcCtx *svc.ServiceContext, req *types.DrawActivityLandingReq, memberID int64) (*types.DrawActivityLandingResp, error) {
	resp, err := svcCtx.DrawParticipationService.QueryDrawActivityLanding(ctx, &smsclient.QueryDrawActivityLandingReq{
		ActivityId:  req.ActivityId,
		MemberId:    memberID,
		Channel:     strings.TrimSpace(req.Channel),
		EntrySource: strings.TrimSpace(req.EntrySource),
		Scope:       buildDrawScope(ctx),
	})
	if err != nil {
		return nil, rpcError(ctx, "查询抽卡活动落地页", req, err)
	}
	return mapLandingResponse(resp), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func currentMemberID(ctx context.Context) (int64, error) {
	return frontcommon.GetMemberId(ctx)
}
