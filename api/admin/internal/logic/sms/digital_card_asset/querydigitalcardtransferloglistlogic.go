package digital_card_asset

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryDigitalCardTransferLogListLogic Story 10.7 Task 9.2 — 后台合规审计跨资产检索
type QueryDigitalCardTransferLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardTransferLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardTransferLogListLogic {
	return &QueryDigitalCardTransferLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardTransferLogListLogic) QueryDigitalCardTransferLogList(req *types.QueryDigitalCardTransferLogListReq) (*types.QueryDigitalCardTransferLogListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	pageNum := int64(req.Current)
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := int64(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	total, list, err := l.svcCtx.CardMintAdminService.QueryDigitalCardTransferLogList(l.ctx, queryScope, digitalcardmint.DigitalCardTransferLogFilter{
		PageNum:         pageNum,
		PageSize:        pageSize,
		AssetInstanceID: req.AssetInstanceId,
		AssetNo:         req.AssetNo,
		OperationType:   req.OperationType,
		TraceID:         req.TraceId,
		DateFrom:        req.DateFrom,
		DateTo:          req.DateTo,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	items := make([]types.DigitalCardTransferLogListItem, 0, len(list))
	for _, item := range list {
		items = append(items, types.DigitalCardTransferLogListItem{
			Id:              item.ID,
			AssetInstanceId: item.AssetInstanceID,
			AssetNo:         item.AssetNo,
			TemplateName:    item.TemplateName,
			OperationType:   item.OperationType,
			OperatorType:    item.OperatorType,
			FromStatus:      item.FromStatus,
			ToStatus:        item.ToStatus,
			ReasonCode:      item.ReasonCode,
			ReasonText:      item.ReasonText,
			TraceId:         item.TraceID,
			PayloadJson:     item.PayloadJSON,
			CreateTime:      item.CreateTime,
		})
	}

	return &types.QueryDigitalCardTransferLogListResp{
		Code:     "000000",
		Message:  "查询转赠审计记录成功",
		Total:    total,
		Current:  req.Current,
		PageSize: req.PageSize,
		Data:     items,
		Success:  true,
	}, nil
}
