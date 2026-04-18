package digital_card_chain

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardChainListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardChainListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardChainListLogic {
	return &QueryDigitalCardChainListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardChainListLogic) QueryDigitalCardChainList(req *types.QueryDigitalCardChainListReq) (*types.QueryDigitalCardChainListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	total, list, err := l.svcCtx.CardMintAdminService.QueryTaskList(l.ctx, queryScope, digitalcardmint.QueryFilter{
		PageNum:        int32(req.Current),
		PageSize:       int32(req.PageSize),
		ActivityID:     req.ActivityId,
		ActivityName:   req.ActivityName,
		MemberID:       req.MemberId,
		TemplateID:     req.TemplateId,
		AssetNo:        req.AssetNo,
		TokenID:        req.TokenId,
		TaskStatus:     req.TaskStatus,
		MintStatus:     req.MintStatus,
		ChainStatus:    req.ChainStatus,
		ManualRequired: req.ManualRequired,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	items := make([]*types.DigitalCardChainItem, 0, len(list))
	for _, item := range list {
		items = append(items, mapTaskItem(item))
	}

	return &types.QueryDigitalCardChainListResp{
		Code:    "000000",
		Message: "查询数字卡片链路成功",
		Data: types.QueryDigitalCardChainListData{
			List:  items,
			Total: total,
		},
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    total,
		Success:  true,
	}, nil
}
