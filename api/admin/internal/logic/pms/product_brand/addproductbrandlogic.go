package product_brand

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

type AddProductBrandLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddProductBrandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductBrandLogic {
	return &AddProductBrandLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *AddProductBrandLogic) AddProductBrand(req *types.AddProductBrandReq) (*types.BaseResp, error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	_, err = l.svcCtx.ProductBrandService.AddProductBrand(l.ctx, &pmsclient.AddProductBrandReq{Name: req.Name, Logo: req.Logo, BigPic: req.BigPic, Description: req.Description, FirstLetter: req.FirstLetter, Sort: req.Sort, RecommendStatus: req.RecommendStatus, IsEnabled: req.IsEnabled, CreateBy: userId, Scope: common.PMSGovernanceScope(writeScope)})
	if err != nil {
		logc.Errorf(l.ctx, "添加商品品牌失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	return res.Success()
}
