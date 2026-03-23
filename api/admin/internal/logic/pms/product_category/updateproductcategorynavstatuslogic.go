package product_category

import (
	"context"
	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/common/res"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type UpdateProductCategoryNavStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductCategoryNavStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductCategoryNavStatusLogic {
	return &UpdateProductCategoryNavStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *UpdateProductCategoryNavStatusLogic) UpdateProductCategoryNavStatus(req *types.UpdateProductCategoryStatusReq) (*types.BaseResp, error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	_, err = l.svcCtx.ProductCategoryService.UpdateCategoryNavStatus(l.ctx, &pmsclient.UpdateProductCategoryStatusReq{Ids: req.Ids, Status: req.Status, UpdateBy: userId, Scope: common.PMSGovernanceScope(writeScope)})
	if err != nil {
		logc.Errorf(l.ctx, "更新产品分类导航状态失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	return res.Success()
}
