package product_attribute

import (
	"context"
	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type UpdateProductAttributeStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductAttributeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductAttributeStatusLogic {
	return &UpdateProductAttributeStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *UpdateProductAttributeStatusLogic) UpdateProductAttributeStatus(req *types.UpdateProductAttributeStatusReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	_, err = l.svcCtx.ProductAttributeService.UpdateProductAttributeStatus(l.ctx, &pmsclient.UpdateProductAttributeStatusReq{Ids: req.Ids, Status: req.Status, UpdateBy: userId, Scope: common.PMSGovernanceScope(writeScope)})
	if err != nil {
		logc.Errorf(l.ctx, "更新商品属性状态失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	return &types.BaseResp{Code: "000000", Message: "更新商品属性状态成功"}, nil
}
