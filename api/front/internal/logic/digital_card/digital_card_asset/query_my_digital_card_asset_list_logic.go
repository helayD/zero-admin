package digital_card_asset

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyDigitalCardAssetListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyDigitalCardAssetListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyDigitalCardAssetListLogic {
	return &QueryMyDigitalCardAssetListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyDigitalCardAssetListLogic) QueryMyDigitalCardAssetList(req *types.QueryMyDigitalCardAssetListReq) (*types.QueryMyDigitalCardAssetListResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	total, items, err := l.svcCtx.CardMintService.QueryMemberDigitalCardAssetList(l.ctx, currentGovernanceScope(l.ctx), memberID, digitalcardmint.MemberDigitalCardAssetFilter{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, assetServiceError(l.ctx, "查询我的数字卡片资产列表", req, err)
	}

	list := make([]types.DigitalCardAssetItem, 0, len(items))
	for _, item := range items {
		list = append(list, mapAssetItem(item))
	}

	return &types.QueryMyDigitalCardAssetListResp{
		Code:    assetCodeSuccess,
		Message: "查询我的数字卡片资产成功",
		Data: types.QueryMyDigitalCardAssetListData{
			Total: total,
			List:  list,
		},
	}, nil
}
