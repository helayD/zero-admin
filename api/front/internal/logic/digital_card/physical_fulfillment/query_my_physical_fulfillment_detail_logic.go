package physical_fulfillment

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryMyPhysicalFulfillmentDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyPhysicalFulfillmentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyPhysicalFulfillmentDetailLogic {
	return &QueryMyPhysicalFulfillmentDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyPhysicalFulfillmentDetailLogic) QueryMyPhysicalFulfillmentDetail(req *types.QueryMyPhysicalFulfillmentDetailReq) (*types.QueryMyPhysicalFulfillmentDetailResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	scope := currentGovernanceScope(l.ctx)

	detail, err := l.svcCtx.CardMintService.QueryMemberPhysicalFulfillmentDetail(l.ctx, scope, memberID, req.AssetInstanceId)
	if err == nil {
		return &types.QueryMyPhysicalFulfillmentDetailResp{
			Code:    physicalCodeSuccess,
			Message: "查询实体卡履约详情成功",
			Data:    mapPhysicalDetail(detail, nil, req.AssetInstanceId),
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, physicalServiceError(l.ctx, "查询实体卡履约详情", req, err)
	}

	ensured, ensureErr := l.svcCtx.CardMintService.EnsurePhysicalFulfillmentByAsset(l.ctx, scope, digitalcardmint.PhysicalFulfillmentInput{
		AssetInstanceID: req.AssetInstanceId,
		OperatorType:    digitalcardmint.OperatorSystem,
	})
	if ensureErr != nil {
		return nil, physicalServiceError(l.ctx, "初始化实体卡履约详情", req, ensureErr)
	}
	if ensured.BlockedReason != "" {
		return &types.QueryMyPhysicalFulfillmentDetailResp{
			Code:    physicalCodeSuccess,
			Message: "查询实体卡履约详情成功",
			Data:    mapPhysicalDetail(nil, ensured, req.AssetInstanceId),
		}, nil
	}

	detail, err = l.svcCtx.CardMintService.QueryMemberPhysicalFulfillmentDetail(l.ctx, scope, memberID, req.AssetInstanceId)
	if err != nil {
		return nil, physicalServiceError(l.ctx, "查询实体卡履约详情", req, err)
	}
	return &types.QueryMyPhysicalFulfillmentDetailResp{
		Code:    physicalCodeSuccess,
		Message: "查询实体卡履约详情成功",
		Data:    mapPhysicalDetail(detail, nil, req.AssetInstanceId),
	}, nil
}
