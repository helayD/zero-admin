package digital_card_asset

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardAssetDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardAssetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardAssetDetailLogic {
	return &QueryDigitalCardAssetDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardAssetDetailLogic) QueryDigitalCardAssetDetail(req *types.QueryDigitalCardAssetDetailReq) (*types.QueryDigitalCardAssetDetailResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	detail, err := l.svcCtx.CardMintAdminService.QueryAssetAuditDetail(l.ctx, current, req.AssetInstanceId)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	return &types.QueryDigitalCardAssetDetailResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.DigitalCardAssetDetailData{
			Item: *mapAuditItem(&detail.Item),
			ParticipationSummary: types.DigitalCardAssetParticipationSummary{
				ParticipationRecordId: detail.ParticipationSummary.ParticipationRecordID,
				RequestId:             detail.ParticipationSummary.RequestID,
				ResultType:            detail.ParticipationSummary.ResultType,
				ResultStatus:          detail.ParticipationSummary.ResultStatus,
				ResultStatusText:      detail.ParticipationSummary.ResultStatusText,
				FailureReason:         detail.ParticipationSummary.FailureReason,
				CreateTime:            detail.ParticipationSummary.CreateTime,
			},
			MintTaskSummary: types.DigitalCardAssetMintTaskSummary{
				TaskId:             detail.MintTaskSummary.TaskID,
				RequestId:          detail.MintTaskSummary.RequestID,
				TraceId:            detail.MintTaskSummary.TraceID,
				TaskStatus:         detail.MintTaskSummary.TaskStatus,
				TaskStatusText:     detail.MintTaskSummary.TaskStatusText,
				MintStatus:         detail.MintTaskSummary.MintStatus,
				MintStatusText:     detail.MintTaskSummary.MintStatusText,
				ChainStatus:        detail.MintTaskSummary.ChainStatus,
				ChainStatusText:    detail.MintTaskSummary.ChainStatusText,
				ChainTxId:          detail.MintTaskSummary.ChainTxID,
				LastReceiptSummary: detail.MintTaskSummary.LastReceiptSummary,
				AvailableActions:   detail.MintTaskSummary.AvailableActions,
				ChainType:          detail.MintTaskSummary.ChainType,
			},
			TraceId:               detail.TraceID,
			RequestId:             detail.RequestID,
			RuleSnapshotJson:      detail.RuleSnapshotJSON,
			Logs:                  mapAssetLogs(detail.Logs),
			AvailableAssetActions: detail.AvailableAssetActions,
		},
		Success: true,
	}, nil
}
