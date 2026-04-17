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

type PreviewDrawEligibilityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreviewDrawEligibilityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewDrawEligibilityLogic {
	return &PreviewDrawEligibilityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreviewDrawEligibilityLogic) PreviewDrawEligibility(req *types.PreviewDrawEligibilityReq) (resp *types.PreviewDrawEligibilityResp, err error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.DrawParticipationService.PreviewDrawEligibility(l.ctx, &smsclient.PreviewDrawEligibilityReq{
		ActivityId:  req.ActivityId,
		MemberId:    memberID,
		Channel:     strings.TrimSpace(req.Channel),
		EntrySource: strings.TrimSpace(req.EntrySource),
		Scope:       buildDrawScope(l.ctx),
	})
	if err != nil {
		return nil, rpcError(l.ctx, "预检抽卡资格", req, err)
	}
	return &types.PreviewDrawEligibilityResp{
		Code:    apiCodeFromEligibility(result.EligibilityCode),
		Message: firstNonEmpty(result.EligibilityMessage, "预检资格成功"),
		Data:    mapEligibilitySummary(result),
	}, nil
}
