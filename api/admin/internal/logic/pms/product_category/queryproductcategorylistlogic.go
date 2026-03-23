package product_category

import (
	"context"
	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type QueryProductCategoryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductCategoryListLogic {
	return &QueryProductCategoryListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *QueryProductCategoryListLogic) QueryProductCategoryList(req *types.QueryProductCategoryListReq) (*types.QueryProductCategoryListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.ProductCategoryService.QueryProductCategoryList(l.ctx, &pmsclient.QueryProductCategoryListReq{PageNum: req.Current, PageSize: req.PageSize, ParentId: req.ParentId, Name: req.Name, NavStatus: req.NavStatus, Keywords: req.Keywords, IsEnabled: req.IsEnabled, Scope: admincommon.PMSGovernanceScope(queryScope)})
	if err != nil {
		logc.Errorf(l.ctx, "查询产品分类列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	list := make([]*types.QueryProductCategoryListData, 0, len(result.List))
	for _, detail := range result.List {
		list = append(list, &types.QueryProductCategoryListData{Id: detail.Id, ParentId: detail.ParentId, Name: detail.Name, Level: detail.Level, ProductCount: detail.ProductCount, ProductUnit: detail.ProductUnit, NavStatus: detail.NavStatus, Sort: detail.Sort, Icon: detail.Icon, Keywords: detail.Keywords, Description: detail.Description, IsEnabled: detail.IsEnabled, CreateBy: detail.CreateBy, CreateTime: detail.CreateTime, UpdateBy: detail.UpdateBy, UpdateTime: detail.UpdateTime, ScopeType: detail.ScopeType, PlatformId: detail.PlatformId, TenantId: detail.TenantId, MerchantId: detail.MerchantId})
	}
	return &types.QueryProductCategoryListResp{Code: "000000", Message: "查询产品分类列表成功", Current: req.Current, Data: list, PageSize: req.PageSize, Success: true, Total: result.Total}, nil
}
