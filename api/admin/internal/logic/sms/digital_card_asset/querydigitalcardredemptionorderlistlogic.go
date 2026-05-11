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

// QueryDigitalCardRedemptionOrderListLogic Story 10.7 Task 2.8 / S3 — 后台提货单管理
type QueryDigitalCardRedemptionOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardRedemptionOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardRedemptionOrderListLogic {
	return &QueryDigitalCardRedemptionOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardRedemptionOrderListLogic) QueryDigitalCardRedemptionOrderList(req *types.QueryDigitalCardRedemptionOrderListReq) (*types.QueryDigitalCardRedemptionOrderListResp, error) {
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

	total, list, err := l.svcCtx.CardMintAdminService.QueryDigitalCardRedemptionOrderList(l.ctx, queryScope, digitalcardmint.DigitalCardRedemptionOrderFilter{
		PageNum:        pageNum,
		PageSize:       pageSize,
		OrderID:        req.OrderId,
		OrderNo:        req.OrderNo,
		CardInstanceID: req.CardInstanceId,
		AssetNo:        req.AssetNo,
		HolderID:       req.HolderId,
		Status:         req.Status,
		OmsOrderID:     req.OmsOrderId,
		DateFrom:       req.DateFrom,
		DateTo:         req.DateTo,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	items := make([]types.DigitalCardRedemptionOrderListItem, 0, len(list))
	for _, item := range list {
		items = append(items, types.DigitalCardRedemptionOrderListItem{
			Id:              item.ID,
			OrderNo:         item.OrderNo,
			CardInstanceId:  item.CardInstanceID,
			AssetNo:         item.AssetNo,
			TemplateName:    item.TemplateName,
			HolderId:        item.HolderID,
			ReceiverName:    item.ReceiverName,
			ReceiverPhone:   item.ReceiverPhone,
			ReceiverAddress: item.ReceiverAddress,
			Status:          item.Status,
			ShippedAt:       item.ShippedAt,
			DeliveredAt:     item.DeliveredAt,
			CancelReason:    item.CancelReason,
			OmsOrderId:      item.OmsOrderID,
			PlatformId:      item.PlatformID,
			TenantId:        item.TenantID,
			MerchantId:      item.MerchantID,
			CreateTime:      item.CreatedAt,
			UpdateTime:      item.UpdatedAt,
		})
	}

	return &types.QueryDigitalCardRedemptionOrderListResp{
		Code:     "000000",
		Message:  "查询提货单成功",
		Total:    total,
		Current:  req.Current,
		PageSize: req.PageSize,
		Data:     items,
		Success:  true,
	}, nil
}
