package physical_fulfillment

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmPhysicalFulfillmentShippingFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmPhysicalFulfillmentShippingFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmPhysicalFulfillmentShippingFeeLogic {
	return &ConfirmPhysicalFulfillmentShippingFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmPhysicalFulfillmentShippingFeeLogic) ConfirmPhysicalFulfillmentShippingFee(req *types.ConfirmPhysicalFulfillmentShippingFeeReq) (*types.ConfirmPhysicalFulfillmentShippingFeeResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	scope := currentGovernanceScope(l.ctx)
	ensured, err := l.svcCtx.CardMintService.EnsurePhysicalFulfillmentByAsset(l.ctx, scope, digitalcardmint.PhysicalFulfillmentInput{
		AssetInstanceID: req.AssetInstanceId,
		OperatorType:    digitalcardmint.OperatorSystem,
	})
	if err != nil {
		return nil, physicalServiceError(l.ctx, "初始化实体卡履约详情", req, err)
	}
	if ensured.BlockedReason != "" {
		return &types.ConfirmPhysicalFulfillmentShippingFeeResp{
			Code:    physicalCodeSuccess,
			Message: "实体卡暂不可支付邮费",
			Data:    mapPhysicalDetail(nil, ensured, req.AssetInstanceId),
		}, nil
	}
	_, err = l.svcCtx.CardMintService.ConfirmPhysicalFulfillmentShippingFee(l.ctx, scope, digitalcardmint.ConfirmPhysicalFulfillmentShippingFeeInput{
		FulfillmentID:   firstPositive(req.FulfillmentId, ensured.FulfillmentID),
		AssetInstanceID: req.AssetInstanceId,
		MemberID:        memberID,
		PayAmount:       req.PayAmount,
		PayChannel:      req.PayChannel,
		PaymentNo:       req.PaymentNo,
	})
	if err != nil {
		return nil, physicalServiceError(l.ctx, "确认实体卡邮费支付", req, err)
	}
	detail, err := l.svcCtx.CardMintService.QueryMemberPhysicalFulfillmentDetail(l.ctx, scope, memberID, req.AssetInstanceId)
	if err != nil {
		return nil, physicalServiceError(l.ctx, "查询实体卡履约详情", req, err)
	}
	return &types.ConfirmPhysicalFulfillmentShippingFeeResp{
		Code:    physicalCodeSuccess,
		Message: "确认邮费支付成功",
		Data:    mapPhysicalDetail(detail, nil, req.AssetInstanceId),
	}, nil
}

func firstPositive(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
