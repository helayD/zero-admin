package physical_fulfillment

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmPhysicalCardReceiptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmPhysicalCardReceiptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmPhysicalCardReceiptLogic {
	return &ConfirmPhysicalCardReceiptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmPhysicalCardReceiptLogic) ConfirmPhysicalCardReceipt(req *types.ConfirmPhysicalCardReceiptReq) (*types.ConfirmPhysicalCardReceiptResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	scope := currentGovernanceScope(l.ctx)
	result, err := l.svcCtx.CardMintService.ConfirmPhysicalCardReceipt(l.ctx, scope, digitalcardmint.ConfirmPhysicalCardReceiptInput{
		FulfillmentID: req.FulfillmentId,
		MemberID:      memberID,
		Reason:        req.Reason,
	})
	if err != nil {
		return nil, physicalServiceError(l.ctx, "确认实体卡签收", req, err)
	}
	detail, err := l.svcCtx.CardMintService.QueryMemberPhysicalFulfillmentDetail(l.ctx, scope, memberID, result.AssetInstanceID)
	if err != nil {
		return nil, physicalServiceError(l.ctx, "查询实体卡履约详情", req, err)
	}
	return &types.ConfirmPhysicalCardReceiptResp{
		Code:    physicalCodeSuccess,
		Message: "确认实体卡签收成功",
		Data:    mapPhysicalDetail(detail, nil, result.AssetInstanceID),
	}, nil
}
