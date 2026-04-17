// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package draw_activity

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ParticipateDrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewParticipateDrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ParticipateDrawLogic {
	return &ParticipateDrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ParticipateDrawLogic) ParticipateDraw(req *types.ParticipateDrawReq) (resp *types.ParticipateDrawResp, err error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.DrawParticipationService.ParticipateDraw(l.ctx, &smsclient.ParticipateDrawReq{
		ActivityId:  req.ActivityId,
		MemberId:    memberID,
		RequestId:   strings.TrimSpace(req.RequestId),
		Channel:     strings.TrimSpace(req.Channel),
		EntrySource: strings.TrimSpace(req.EntrySource),
		Scope:       buildDrawScope(l.ctx),
	})
	if err != nil {
		return nil, rpcError(l.ctx, "参与抽卡", req, err)
	}
	return &types.ParticipateDrawResp{
		Code:    apiCodeFromEligibility(result.EligibilityCode),
		Message: firstNonEmpty(result.Record.ResultStatusText, result.EligibilityMessage, "参与抽卡成功"),
		Data: types.ParticipateDrawData{
			Record: mapDrawRecord(result.Record),
			Eligibility: types.DrawEligibilitySummary{
				EligibilityStatus:  result.EligibilityCode,
				EligibilityCode:    result.EligibilityCode,
				EligibilityMessage: result.EligibilityMessage,
				NextAction:         result.NextAction,
			},
		},
	}, nil
}
