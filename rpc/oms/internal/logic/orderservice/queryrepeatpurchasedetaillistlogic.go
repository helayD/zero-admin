package orderservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryRepeatPurchaseDetailListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryRepeatPurchaseDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRepeatPurchaseDetailListLogic {
	return &QueryRepeatPurchaseDetailListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryRepeatPurchaseDetailListLogic) QueryRepeatPurchaseDetailList(in *omsclient.QueryRepeatPurchaseDetailListReq) (*omsclient.QueryRepeatPurchaseDetailListResp, error) {
	scope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	startTime, endTime, err := operatefunnel.ParseTimeRange(in.StartTime, in.EndTime)
	if err != nil {
		return nil, err
	}
	builder := newRepeatPurchaseBuilder(l.ctx, l.svcCtx)
	snapshot, err := builder.BuildRepeatPurchaseSnapshot(repeatPurchaseFilter{
		Scope:        scope,
		StartTime:    startTime,
		EndTime:      endTime,
		Channel:      in.Channel,
		ActivityType: in.ActivityType,
		ActivityID:   in.ActivityId,
	}, operatefunnel.BucketDay)
	if err != nil {
		return nil, err
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	start := int((pageNum - 1) * pageSize)
	if start > len(snapshot.DetailRows) {
		start = len(snapshot.DetailRows)
	}
	end := start + int(pageSize)
	if end > len(snapshot.DetailRows) {
		end = len(snapshot.DetailRows)
	}
	list := make([]*omsclient.RepeatPurchaseDetailRow, 0, end-start)
	list = append(list, snapshot.DetailRows[start:end]...)
	return &omsclient.QueryRepeatPurchaseDetailListResp{
		Total:             int64(len(snapshot.DetailRows)),
		List:              list,
		TrackingStartedAt: snapshot.TrackingStartedAt,
		PartialMetrics:    snapshot.PartialMetrics,
	}, nil
}
