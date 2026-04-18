package digital_card_asset

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyDigitalCardAssetDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyDigitalCardAssetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyDigitalCardAssetDetailLogic {
	return &QueryMyDigitalCardAssetDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyDigitalCardAssetDetailLogic) QueryMyDigitalCardAssetDetail(req *types.QueryMyDigitalCardAssetDetailReq) (*types.QueryMyDigitalCardAssetDetailResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	detail, err := l.svcCtx.CardMintService.QueryMemberDigitalCardAssetDetail(l.ctx, currentGovernanceScope(l.ctx), memberID, req.AssetInstanceId)
	if err != nil {
		return nil, assetServiceError(l.ctx, "查询我的数字卡片资产详情", req, err)
	}

	return &types.QueryMyDigitalCardAssetDetailResp{
		Code:    assetCodeSuccess,
		Message: "查询我的数字卡片资产详情成功",
		Data: types.DigitalCardAssetDetailData{
			Item:                mapAssetItem(detail.Item),
			TokenIdMasked:       detail.TokenIDMasked,
			LatestStatusSummary: detail.LatestStatusSummary,
			RestrictionReason:   detail.RestrictionReason,
			DrawSummary:         mapDrawSummary(detail.DrawSummary),
			Timeline:            mapTimeline(detail.Timeline),
		},
	}, nil
}
