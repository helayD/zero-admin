package physical_fulfillment

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmPhysicalFulfillmentAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmPhysicalFulfillmentAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmPhysicalFulfillmentAddressLogic {
	return &ConfirmPhysicalFulfillmentAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmPhysicalFulfillmentAddressLogic) ConfirmPhysicalFulfillmentAddress(req *types.ConfirmPhysicalFulfillmentAddressReq) (*types.ConfirmPhysicalFulfillmentAddressResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	scope := currentGovernanceScope(l.ctx)
	ensured, err := l.svcCtx.CardMintService.EnsurePhysicalFulfillmentByAsset(l.ctx, scope, digitalcardmint.PhysicalFulfillmentInput{
		AssetInstanceID: req.AssetInstanceId,
		AddressID:       req.AddressId,
		OperatorType:    digitalcardmint.OperatorSystem,
	})
	if err != nil {
		return nil, physicalServiceError(l.ctx, "确认实体卡履约地址", req, err)
	}
	if ensured.BlockedReason != "" {
		return &types.ConfirmPhysicalFulfillmentAddressResp{
			Code:    physicalCodeSuccess,
			Message: "实体卡暂不可确认地址",
			Data:    mapPhysicalDetail(nil, ensured, req.AssetInstanceId),
		}, nil
	}
	_, err = l.svcCtx.CardMintService.ConfirmPhysicalFulfillmentAddress(l.ctx, scope, digitalcardmint.ConfirmPhysicalFulfillmentAddressInput{
		FulfillmentID:   ensured.FulfillmentID,
		AssetInstanceID: req.AssetInstanceId,
		MemberID:        memberID,
		AddressID:       req.AddressId,
	})
	if err != nil {
		return nil, physicalServiceError(l.ctx, "确认实体卡履约地址", req, err)
	}
	detail, err := l.svcCtx.CardMintService.QueryMemberPhysicalFulfillmentDetail(l.ctx, scope, memberID, req.AssetInstanceId)
	if err != nil {
		return nil, physicalServiceError(l.ctx, "查询实体卡履约详情", req, err)
	}
	return &types.ConfirmPhysicalFulfillmentAddressResp{
		Code:    physicalCodeSuccess,
		Message: "确认实体卡履约地址成功",
		Data:    mapPhysicalDetail(detail, nil, req.AssetInstanceId),
	}, nil
}
