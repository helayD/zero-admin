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

type QueryProductCategoryDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductCategoryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductCategoryDetailLogic {
	return &QueryProductCategoryDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *QueryProductCategoryDetailLogic) QueryProductCategoryDetail(req *types.QueryProductCategoryDetailReq) (*types.QueryProductCategoryDetailResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.ProductCategoryService.QueryProductCategoryDetail(l.ctx, &pmsclient.QueryProductCategoryDetailReq{Id: req.Id, Scope: admincommon.PMSGovernanceScope(queryScope)})
	if err != nil {
		logc.Errorf(l.ctx, "查询产品分类详情失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	return &types.QueryProductCategoryDetailResp{Code: "000000", Message: "查询产品分类详情成功", Data: types.QueryProductCategoryDetailData{Id: result.Id, ParentId: result.ParentId, Name: result.Name, Level: result.Level, ProductCount: result.ProductCount, ProductUnit: result.ProductUnit, NavStatus: result.NavStatus, Sort: result.Sort, Icon: result.Icon, Keywords: result.Keywords, Description: result.Description, IsEnabled: result.IsEnabled, CreateBy: result.CreateBy, CreateTime: result.CreateTime, UpdateBy: result.UpdateBy, UpdateTime: result.UpdateTime, ScopeType: result.ScopeType, PlatformId: result.PlatformId, TenantId: result.TenantId, MerchantId: result.MerchantId}}, nil
}
