package orderservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryRepeatPurchaseAnalysisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryRepeatPurchaseAnalysisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRepeatPurchaseAnalysisLogic {
	return &QueryRepeatPurchaseAnalysisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryRepeatPurchaseAnalysisLogic) QueryRepeatPurchaseAnalysis(in *omsclient.QueryRepeatPurchaseAnalysisReq) (*omsclient.QueryRepeatPurchaseAnalysisResp, error) {
	scope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	startTime, endTime, err := operatefunnel.ParseTimeRange(in.StartTime, in.EndTime)
	if err != nil {
		return nil, err
	}
	bucket := operatefunnel.NormalizeBucket(in.Bucket, startTime, endTime)
	builder := newRepeatPurchaseBuilder(l.ctx, l.svcCtx)
	snapshot, err := builder.BuildRepeatPurchaseSnapshot(repeatPurchaseFilter{
		Scope:        scope,
		StartTime:    startTime,
		EndTime:      endTime,
		Channel:      in.Channel,
		ActivityType: in.ActivityType,
		ActivityID:   in.ActivityId,
	}, bucket)
	if err != nil {
		return nil, err
	}
	return &omsclient.QueryRepeatPurchaseAnalysisResp{
		Overview:          snapshot.Overview,
		Trends:            snapshot.Trends,
		TrackingStartedAt: snapshot.TrackingStartedAt,
		PartialMetrics:    snapshot.PartialMetrics,
	}, nil
}
