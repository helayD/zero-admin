package cardassetservicelogic

import (
	"context"

	cardminttaskservicelogic "github.com/feihua/zero-admin/rpc/sms/internal/logic/cardminttaskservice"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type BackfillWinningCardInstancesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBackfillWinningCardInstancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackfillWinningCardInstancesLogic {
	return &BackfillWinningCardInstancesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BackfillWinningCardInstancesLogic) BackfillWinningCardInstances(in *smsclient.BackfillWinningCardInstancesReq) (*smsclient.BackfillWinningCardInstancesResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	total, assets, err := BackfillWinningCardInstances(l.ctx, l.svcCtx.DB, currentScope, in.ActivityId, in.MemberId, in.Limit, in.OperatorType, in.TraceId)
	if err != nil {
		return nil, err
	}
	dispatchTaskIDs := make([]int64, 0, len(assets))
	result := make([]*smsclient.CardInstanceData, 0, len(assets))
	for _, asset := range assets {
		var dispatchTaskID int64
		err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var txErr error
			dispatchTaskID, txErr = cardminttaskservicelogic.EnsureCardMintTaskByAssetInstance(l.ctx, l.svcCtx, tx, asset.ID, in.OperatorType)
			return txErr
		})
		if err != nil {
			return nil, err
		}
		if dispatchTaskID > 0 {
			dispatchTaskIDs = append(dispatchTaskIDs, dispatchTaskID)
		}
		result = append(result, buildCardInstanceData(asset))
	}
	for _, taskID := range dispatchTaskIDs {
		cardminttaskservicelogic.DispatchCardMintTask(l.ctx, l.svcCtx, taskID, "历史资产回填后自动派发链上发放任务")
	}
	return &smsclient.BackfillWinningCardInstancesResp{
		TotalCandidates: total,
		ProcessedCount:  int32(len(result)),
		Assets:          result,
	}, nil
}
