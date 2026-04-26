package digital_card_physical_fulfillment

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardPhysicalFulfillmentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardPhysicalFulfillmentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardPhysicalFulfillmentListLogic {
	return &QueryDigitalCardPhysicalFulfillmentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardPhysicalFulfillmentListLogic) QueryDigitalCardPhysicalFulfillmentList(req *types.QueryDigitalCardPhysicalFulfillmentListReq) (*types.QueryDigitalCardPhysicalFulfillmentListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, requestedPhysicalScope(req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId))
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	total, list, err := l.svcCtx.CardMintAdminService.QueryPhysicalFulfillmentList(l.ctx, queryScope, digitalcardmint.PhysicalFulfillmentFilter{
		PageNum:           int32(req.Current),
		PageSize:          int32(req.PageSize),
		ActivityID:        req.ActivityId,
		TemplateID:        req.TemplateId,
		MemberID:          req.MemberId,
		AssetNo:           req.AssetNo,
		FulfillmentNo:     req.FulfillmentNo,
		ProductionBatchNo: req.ProductionBatchNo,
		FulfillmentStatus: req.FulfillmentStatus,
		ProductionStatus:  req.ProductionStatus,
		ShippingStatus:    req.ShippingStatus,
		TrackingNo:        req.TrackingNo,
		FailureCode:       req.FailureCode,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	items := make([]*types.DigitalCardPhysicalFulfillmentItem, 0, len(list))
	for index := range list {
		items = append(items, mapPhysicalFulfillmentItem(&list[index]))
	}
	return &types.QueryDigitalCardPhysicalFulfillmentListResp{
		Code:    "000000",
		Message: "查询实体卡履约单成功",
		Data: types.QueryDigitalCardPhysicalFulfillmentListData{
			List:  items,
			Total: total,
		},
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    total,
		Success:  true,
	}, nil
}

type QueryDigitalCardPhysicalFulfillmentDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardPhysicalFulfillmentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardPhysicalFulfillmentDetailLogic {
	return &QueryDigitalCardPhysicalFulfillmentDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardPhysicalFulfillmentDetailLogic) QueryDigitalCardPhysicalFulfillmentDetail(req *types.QueryDigitalCardPhysicalFulfillmentDetailReq) (*types.QueryDigitalCardPhysicalFulfillmentDetailResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	detail, err := l.svcCtx.CardMintAdminService.QueryPhysicalFulfillmentDetail(l.ctx, current, req.FulfillmentId)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	return &types.QueryDigitalCardPhysicalFulfillmentDetailResp{
		Code:    "000000",
		Message: "查询实体卡履约详情成功",
		Data: types.DigitalCardPhysicalFulfillmentDetailData{
			Item:           *mapPhysicalFulfillmentItem(&detail.Item),
			ReceiverName:   detail.ReceiverName,
			ReceiverPhone:  detail.ReceiverPhone,
			AddressSummary: detail.AddressSummary,
			RequestId:      detail.RequestID,
			TraceId:        detail.TraceID,
			Timeline:       mapPhysicalTimeline(detail.Timeline),
		},
		Success: true,
	}, nil
}

type EnsureDigitalCardPhysicalFulfillmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnsureDigitalCardPhysicalFulfillmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnsureDigitalCardPhysicalFulfillmentLogic {
	return &EnsureDigitalCardPhysicalFulfillmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnsureDigitalCardPhysicalFulfillmentLogic) EnsureDigitalCardPhysicalFulfillment(req *types.EnsureDigitalCardPhysicalFulfillmentReq) (*types.DigitalCardPhysicalFulfillmentActionResp, error) {
	writeScope, operatorID, err := resolvePhysicalActionContext(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintAdminService.EnsurePhysicalFulfillment(l.ctx, writeScope, digitalcardmint.PhysicalFulfillmentInput{
		AssetInstanceID: req.AssetInstanceId,
		OperatorType:    digitalcardmint.OperatorManual,
		OperatorID:      operatorID,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	fulfillmentID := int64(0)
	if result != nil {
		fulfillmentID = result.FulfillmentID
	}
	writePhysicalFulfillmentOperateLog(l.ctx, l.svcCtx, operatorID, "ensure", fulfillmentID, "", result)
	return mapPhysicalActionResp("实体卡履约单已初始化", result), nil
}

type UpdateDigitalCardPhysicalProductionStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDigitalCardPhysicalProductionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDigitalCardPhysicalProductionStatusLogic {
	return &UpdateDigitalCardPhysicalProductionStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDigitalCardPhysicalProductionStatusLogic) UpdateDigitalCardPhysicalProductionStatus(req *types.UpdateDigitalCardPhysicalProductionStatusReq) (*types.DigitalCardPhysicalFulfillmentActionResp, error) {
	writeScope, operatorID, err := resolvePhysicalActionContext(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintAdminService.UpdatePhysicalCardProductionStatus(l.ctx, writeScope, digitalcardmint.UpdatePhysicalCardProductionStatusInput{
		FulfillmentID:     req.FulfillmentId,
		OperatorID:        operatorID,
		ProductionBatchNo: req.ProductionBatchNo,
		ProductionStatus:  req.ProductionStatus,
		Reason:            req.Reason,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	writePhysicalFulfillmentOperateLog(l.ctx, l.svcCtx, operatorID, "production", req.FulfillmentId, req.Reason, result)
	return mapPhysicalActionResp("实体卡制作状态已更新", result), nil
}

type ShipDigitalCardPhysicalFulfillmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShipDigitalCardPhysicalFulfillmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipDigitalCardPhysicalFulfillmentLogic {
	return &ShipDigitalCardPhysicalFulfillmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShipDigitalCardPhysicalFulfillmentLogic) ShipDigitalCardPhysicalFulfillment(req *types.ShipDigitalCardPhysicalFulfillmentReq) (*types.DigitalCardPhysicalFulfillmentActionResp, error) {
	writeScope, operatorID, err := resolvePhysicalActionContext(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintAdminService.ShipPhysicalCard(l.ctx, writeScope, digitalcardmint.ShipPhysicalCardInput{
		FulfillmentID: req.FulfillmentId,
		OperatorID:    operatorID,
		CarrierCode:   req.CarrierCode,
		CarrierName:   req.CarrierName,
		TrackingNo:    req.TrackingNo,
		Reason:        req.Reason,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	writePhysicalFulfillmentOperateLog(l.ctx, l.svcCtx, operatorID, "ship", req.FulfillmentId, req.Reason, result)
	return mapPhysicalActionResp("实体卡发货成功", result), nil
}

type MarkDigitalCardPhysicalFulfillmentExceptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkDigitalCardPhysicalFulfillmentExceptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkDigitalCardPhysicalFulfillmentExceptionLogic {
	return &MarkDigitalCardPhysicalFulfillmentExceptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkDigitalCardPhysicalFulfillmentExceptionLogic) MarkDigitalCardPhysicalFulfillmentException(req *types.DigitalCardPhysicalFulfillmentExceptionReq) (*types.DigitalCardPhysicalFulfillmentActionResp, error) {
	writeScope, operatorID, err := resolvePhysicalActionContext(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintAdminService.MarkPhysicalFulfillmentException(l.ctx, writeScope, digitalcardmint.PhysicalFulfillmentExceptionInput{
		FulfillmentID: req.FulfillmentId,
		OperatorID:    operatorID,
		FailureCode:   req.FailureCode,
		Reason:        req.Reason,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	writePhysicalFulfillmentOperateLog(l.ctx, l.svcCtx, operatorID, "exception", req.FulfillmentId, req.Reason, result)
	return mapPhysicalActionResp("实体卡履约异常已标记", result), nil
}

type RequestDigitalCardPhysicalReissueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRequestDigitalCardPhysicalReissueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestDigitalCardPhysicalReissueLogic {
	return &RequestDigitalCardPhysicalReissueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RequestDigitalCardPhysicalReissueLogic) RequestDigitalCardPhysicalReissue(req *types.DigitalCardPhysicalFulfillmentExceptionReq) (*types.DigitalCardPhysicalFulfillmentActionResp, error) {
	writeScope, operatorID, err := resolvePhysicalActionContext(l.ctx, req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintAdminService.RequestPhysicalCardReissue(l.ctx, writeScope, digitalcardmint.PhysicalFulfillmentExceptionInput{
		FulfillmentID: req.FulfillmentId,
		OperatorID:    operatorID,
		FailureCode:   req.FailureCode,
		Reason:        req.Reason,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	writePhysicalFulfillmentOperateLog(l.ctx, l.svcCtx, operatorID, "reissue", req.FulfillmentId, req.Reason, result)
	return mapPhysicalActionResp("实体卡补发已发起", result), nil
}
