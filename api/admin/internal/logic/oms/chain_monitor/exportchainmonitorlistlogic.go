package chain_monitor

import (
	"bytes"
	"context"
	"fmt"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/xuri/excelize/v2"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportChainMonitorListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportChainMonitorListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportChainMonitorListLogic {
	return &ExportChainMonitorListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportChainMonitorListLogic) ExportChainMonitorList(req *types.ExportChainMonitorListReq) ([]byte, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.OrderService.QueryCompensationChainList(l.ctx, &omsclient.QueryCompensationChainListReq{
		PageNum:            1,
		PageSize:           10000,
		ChainType:         req.ChainType,
		ConsistencyStage:  req.ConsistencyStage,
		ConsistencyResult: req.ConsistencyResult,
		ManualRequired:    req.ManualRequired,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		OrderNo:           req.OrderNo,
		Scope:             admincommon.OMSGovernanceScope(queryScope),
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "链路监控"
	idx, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(idx)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"链路追踪ID", "链路类型", "业务编号", "业务ID",
		"阶段", "结果", "重试次数", "最近错误",
		"最近执行时间", "链路创建时间", "触发操作人ID",
		"平台ID", "租户ID", "商户ID",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	entityTypeMap := map[int32]string{1: "订单", 2: "商品", 3: "优惠券", 4: "积分"}
	for rowIdx, item := range result.List {
		row := rowIdx + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.TraceId)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ChainTypeText)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.EntityNo)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.EntityId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.StageText)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.ResultText)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.RetryCount)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), item.LastError)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), item.LastExecuteAt)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), item.CreatedAt)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), item.ActorId)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), item.PlatformId)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), item.TenantId)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), item.MerchantId)
		_ = entityTypeMap
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
