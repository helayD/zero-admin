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

type QueryProductAttributeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductAttributeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductAttributeListLogic {
	return &QueryProductAttributeListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *QueryProductAttributeListLogic) QueryProductAttributeList(req *types.QueryProductAttributeListReq) (resp *types.QueryProductAttributeListResp, err error) {
	queryScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.ProductAttributeService.QueryProductAttributeList(l.ctx, &pmsclient.QueryProductAttributeListReq{PageNum: req.Current, PageSize: req.PageSize, GroupId: req.GroupId, Name: req.Name, InputType: req.InputType, IsRequired: req.IsRequired, IsSearchable: req.IsSearchable, IsShow: req.IsShow, Status: req.Status, Scope: common.PMSGovernanceScope(queryScope)})
	if err != nil {
		logc.Errorf(l.ctx, "查询商品属性列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	list := make([]*types.QueryProductAttributeListData, 0, len(result.List))
	for _, d := range result.List {
		list = append(list, &types.QueryProductAttributeListData{Id: d.Id, GroupId: d.GroupId, Name: d.Name, InputType: d.InputType, ValueType: d.ValueType, InputList: d.InputList, Unit: d.Unit, IsRequired: d.IsRequired, IsSearchable: d.IsSearchable, IsShow: d.IsShow, Sort: d.Sort, Status: d.Status, CreateBy: d.CreateBy, CreateTime: d.CreateTime, UpdateBy: d.UpdateBy, UpdateTime: d.UpdateTime, ScopeType: d.ScopeType, PlatformId: d.PlatformId, TenantId: d.TenantId, MerchantId: d.MerchantId})
	}
	return &types.QueryProductAttributeListResp{Code: "000000", Message: "查询商品属性列表成功", Data: list, Current: req.Current, PageSize: req.PageSize, Total: result.Total, Success: true}, nil
}
