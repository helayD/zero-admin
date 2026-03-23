package product_brand

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductBrandDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductBrandDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductBrandDetailLogic {
	return &QueryProductBrandDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *QueryProductBrandDetailLogic) QueryProductBrandDetail(req *types.QueryProductBrandDetailReq) (*types.QueryProductBrandDetailResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.ProductBrandService.QueryProductBrandDetail(l.ctx, &pmsclient.QueryProductBrandDetailReq{Id: req.Id, Scope: admincommon.PMSGovernanceScope(queryScope)})
	if err != nil {
		logc.Errorf(l.ctx, "查询商品品牌详情失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	return &types.QueryProductBrandDetailResp{Code: "000000", Message: "查询商品品牌详情成功", Data: types.QueryProductBrandDetailData{Id: result.Id, Name: result.Name, Logo: result.Logo, BigPic: result.BigPic, Description: result.Description, FirstLetter: result.FirstLetter, Sort: result.Sort, RecommendStatus: result.RecommendStatus, ProductCount: result.ProductCount, ProductCommentCount: result.ProductCommentCount, IsEnabled: result.IsEnabled, CreateBy: result.CreateBy, CreateTime: result.CreateTime, UpdateBy: result.UpdateBy, UpdateTime: result.UpdateTime, ScopeType: result.ScopeType, PlatformId: result.PlatformId, TenantId: result.TenantId, MerchantId: result.MerchantId}}, nil
}
