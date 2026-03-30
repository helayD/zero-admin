package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryOrderStatusSnapshotLogic 查询订单状态快照
// 聚合 order_main + order_payment + order_operation_log
//
// Author: Claude AI
// Date: 2026-03-30
// Story: 6-5 Task 8
type QueryOrderStatusSnapshotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOrderStatusSnapshotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderStatusSnapshotLogic {
	return &QueryOrderStatusSnapshotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryOrderStatusSnapshot 查询订单状态快照入口
func (l *QueryOrderStatusSnapshotLogic) QueryOrderStatusSnapshot(req *types.QueryOrderStatusSnapshotReq) (resp *types.QueryOrderStatusSnapshotResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	orderSvc := NewOrderStatusService(l.ctx, l.svcCtx)

	snapshot, err := orderSvc.GetOrderStatusSnapshot(req.OrderId, memberId)
	if err != nil {
		return nil, err
	}

	// 转换操作日志列表
	var optLogs []types.OrderOperationLogItem
	for _, log := range snapshot.OptLogs {
		optLogs = append(optLogs, types.OrderOperationLogItem{
			Id:            log.Id,
			OperationType: log.OperationType,
			OperatorType:  log.OperatorType,
			OperatorNote:  log.OperatorNote,
			CreateTime:    log.CreateTime,
		})
	}

	return &types.QueryOrderStatusSnapshotResp{
		Code:             0,
		Message:          "查询成功",
		OrderId:          snapshot.OrderId,
		OrderNo:          snapshot.OrderNo,
		OrderStatus:      snapshot.OrderStatus,
		PayStatus:        snapshot.PayStatus,
		OrderStatusText:  snapshot.OrderStatusText,
		PayStatusText:    snapshot.PayStatusText,
		OptLogs:          optLogs,
	}, nil
}
