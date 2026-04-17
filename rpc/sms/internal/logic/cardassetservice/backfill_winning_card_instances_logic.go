package cardassetservicelogic

import (
	"context"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
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
	result := make([]*smsclient.CardInstanceData, 0, len(assets))
	for _, asset := range assets {
		result = append(result, buildCardInstanceData(asset))
	}
	return &smsclient.BackfillWinningCardInstancesResp{
		TotalCandidates: total,
		ProcessedCount:  int32(len(result)),
		Assets:          result,
	}, nil
}
