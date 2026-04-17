package drawparticipationservicelogic

import (
	"context"
	"errors"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PreviewDrawEligibilityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPreviewDrawEligibilityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewDrawEligibilityLogic {
	return &PreviewDrawEligibilityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PreviewDrawEligibilityLogic) PreviewDrawEligibility(in *smsclient.PreviewDrawEligibilityReq) (*smsclient.PreviewDrawEligibilityResp, error) {
	if in.ActivityId <= 0 {
		return nil, errors.New("活动ID不能为空")
	}
	if in.MemberId <= 0 {
		return &smsclient.PreviewDrawEligibilityResp{
			EligibilityStatus:  drawEligibilityNeedLogin,
			EligibilityCode:    drawEligibilityNeedLogin,
			EligibilityMessage: "登录后可查看参与资格",
			NextAction:         drawNextActionLogin,
		}, nil
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
		logc.Errorf(l.ctx, "预检抽卡资格失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("预检抽卡资格失败")
	}
	member, err := loadMemberInfoSnapshot(l.ctx, l.svcCtx.DB, in.MemberId, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("查询会员信息失败")
	}
	identity, err := loadMemberIdentitySnapshot(l.ctx, l.svcCtx.DB, in.MemberId)
	if err != nil {
		return nil, errors.New("查询实名状态失败")
	}
	totalCount, dailyCount, err := countConsumedRecords(l.ctx, l.svcCtx.DB, activity.ID, in.MemberId)
	if err != nil {
		return nil, errors.New("统计参与次数失败")
	}
	pools, err := loadPoolSnapshots(l.ctx, l.svcCtx.DB, activity.ID)
	if err != nil {
		return nil, errors.New("查询卡池失败")
	}
	poolIDs := make([]int64, 0, len(pools))
	for _, pool := range pools {
		poolIDs = append(poolIDs, pool.ID)
	}
	poolTemplates, err := loadPoolTemplates(l.ctx, l.svcCtx.DB, activity.ID, poolIDs, false)
	if err != nil {
		return nil, errors.New("查询卡池模板失败")
	}

	eligibility := buildEligibility(activity, member, identity, totalCount, dailyCount, hasAvailableInventory(poolTemplates), true)
	return &smsclient.PreviewDrawEligibilityResp{
		EligibilityStatus:     eligibility.Status,
		EligibilityCode:       eligibility.Code,
		EligibilityMessage:    eligibility.Message,
		NextAction:            eligibility.NextAction,
		RemainingLotteryTimes: eligibility.RemainingLotteryTime,
		RealNameStatus:        eligibility.RealNameStatus,
		RealNameStatusText:    eligibility.RealNameStatusText,
		RealNameMasked:        eligibility.RealNameMasked,
		CredentialRef:         eligibility.CredentialRef,
		VerifiedAt:            eligibility.VerifiedAt,
	}, nil
}
