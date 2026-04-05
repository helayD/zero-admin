package operate_dashboard

import (
	"bytes"
	"context"
	"fmt"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ExportRepeatPurchaseAnalysisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportRepeatPurchaseAnalysisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportRepeatPurchaseAnalysisLogic {
	return &ExportRepeatPurchaseAnalysisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportRepeatPurchaseAnalysisLogic) ExportRepeatPurchaseAnalysis(req *types.ExportRepeatPurchaseAnalysisReq) ([]byte, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	startText, endText, startTime, endTime, err := normalizeRepeatPurchaseTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	_ = operatefunnel.NormalizeBucket(req.Bucket, startTime, endTime)

	f := excelize.NewFile()
	defer f.Close()
	sheetName := "复购分析"
	idx, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(idx)
	f.DeleteSheet("Sheet1")
	stream, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}
	headers := repeatPurchaseExportHeader()
	headerValues := make([]interface{}, 0, len(headers))
	for _, header := range headers {
		headerValues = append(headerValues, header)
	}
	if err := stream.SetRow("A1", headerValues); err != nil {
		return nil, err
	}

	const maxExportRows = 50000
	pageNum := int32(1)
	pageSize := int32(500)
	rowIndex := 2
	for {
		detailResp, err := l.svcCtx.OrderService.QueryRepeatPurchaseDetailList(l.ctx, &omsclient.QueryRepeatPurchaseDetailListReq{
			Scope:        admincommon.OMSGovernanceScope(queryScope),
			StartTime:    startText,
			EndTime:      endText,
			Channel:      req.Channel,
			ActivityType: req.ActivityType,
			ActivityId:   req.ActivityId,
			PageNum:      pageNum,
			PageSize:     pageSize,
		})
		if err != nil {
			logc.Errorf(l.ctx, "导出复购分析查询详情失败, req=%+v, err=%s", req, err.Error())
			return nil, errorx.NewDefaultError(rpcErrorMessage(err))
		}
		if len(detailResp.List) == 0 {
			break
		}
		briefMap, err := queryRepeatPurchaseMemberBriefMap(l.ctx, l.svcCtx.MemberInfoService, detailResp.List)
		if err != nil {
			logc.Errorf(l.ctx, "导出复购分析查询会员简要信息失败, req=%+v, err=%s", req, err.Error())
			return nil, errorx.NewDefaultError(rpcErrorMessage(err))
		}
		items := buildRepeatPurchaseDetailItems(detailResp.List, briefMap)
		for _, item := range items {
			if err := stream.SetRow(fmt.Sprintf("A%d", rowIndex), repeatPurchaseExportRow(item)); err != nil {
				return nil, err
			}
			rowIndex++
		}
		if rowIndex-1 >= maxExportRows {
			logc.Infof(l.ctx, "导出复购分析达到最大行数上限 %d, 截断导出", maxExportRows)
			break
		}
		if int64((pageNum)*pageSize) >= detailResp.Total {
			break
		}
		pageNum++
	}
	if err := stream.Flush(); err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
