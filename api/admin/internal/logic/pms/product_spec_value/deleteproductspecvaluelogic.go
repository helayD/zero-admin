package product_spec_value

import (
	"context"
	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteProductSpecValueLogic 删除商品规格值
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:37
*/
type DeleteProductSpecValueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductSpecValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductSpecValueLogic {
	return &DeleteProductSpecValueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteProductSpecValue 删除商品规格值
func (l *DeleteProductSpecValueLogic) DeleteProductSpecValue(req *types.DeleteProductSpecValueReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	_, err = l.svcCtx.ProductSpecValueService.DeleteProductSpecValue(l.ctx, &pmsclient.DeleteProductSpecValueReq{
		Ids:      req.Ids,
		UpdateBy: userId,
		Scope:    common.PMSGovernanceScope(writeScope),
	})

	if err != nil {
		logc.Errorf(l.ctx, "删除商品规格值失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "删除商品规格值成功",
	}, nil
}
