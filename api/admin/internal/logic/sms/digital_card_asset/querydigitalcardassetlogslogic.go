package digital_card_asset

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardAssetLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardAssetLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardAssetLogsLogic {
	return &QueryDigitalCardAssetLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardAssetLogsLogic) QueryDigitalCardAssetLogs(req *types.QueryDigitalCardAssetLogsReq) (*types.QueryDigitalCardAssetLogsResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	detail, err := l.svcCtx.CardMintAdminService.QueryAssetAuditDetail(l.ctx, current, req.AssetInstanceId)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	return &types.QueryDigitalCardAssetLogsResp{
		Code:    "000000",
		Message: "查询数字卡片资产日志成功",
		Logs:    mapAssetLogs(detail.Logs),
		Success: true,
	}, nil
}
