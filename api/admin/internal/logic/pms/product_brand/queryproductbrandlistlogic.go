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

type QueryProductBrandListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductBrandListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductBrandListLogic {
	return &QueryProductBrandListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *QueryProductBrandListLogic) QueryProductBrandList(req *types.QueryProductBrandListReq) (*types.QueryProductBrandListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.ProductBrandService.QueryProductBrandList(l.ctx, &pmsclient.QueryProductBrandListReq{PageNum: req.Current, PageSize: req.PageSize, Name: req.Name, RecommendStatus: req.RecommendStatus, IsEnabled: req.IsEnabled, Scope: admincommon.PMSGovernanceScope(queryScope)})
	if err != nil {
		logc.Errorf(l.ctx, "查询商品品牌列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	list := make([]*types.QueryProductBrandListData, 0, len(result.List))
	for _, detail := range result.List {
		list = append(list, &types.QueryProductBrandListData{Id: detail.Id, Name: detail.Name, Logo: detail.Logo, BigPic: detail.BigPic, Description: detail.Description, FirstLetter: detail.FirstLetter, Sort: detail.Sort, RecommendStatus: detail.RecommendStatus, ProductCount: detail.ProductCount, ProductCommentCount: detail.ProductCommentCount, IsEnabled: detail.IsEnabled, CreateBy: detail.CreateBy, CreateTime: detail.CreateTime, UpdateBy: detail.UpdateBy, UpdateTime: detail.UpdateTime, ScopeType: detail.ScopeType, PlatformId: detail.PlatformId, TenantId: detail.TenantId, MerchantId: detail.MerchantId})
	}
	return &types.QueryProductBrandListResp{Code: "000000", Message: "查询商品品牌列表成功", Current: req.Current, Data: list, PageSize: req.PageSize, Success: true, Total: result.Total}, nil
}
