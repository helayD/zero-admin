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

type AddProductAttributeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddProductAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductAttributeLogic {
	return &AddProductAttributeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *AddProductAttributeLogic) AddProductAttribute(req *types.AddProductAttributeReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	_, err = l.svcCtx.ProductAttributeService.AddProductAttribute(l.ctx, &pmsclient.AddProductAttributeReq{GroupId: req.GroupId, Name: req.Name, InputType: req.InputType, ValueType: req.ValueType, InputList: req.InputList, Unit: req.Unit, IsRequired: req.IsRequired, IsSearchable: req.IsSearchable, IsShow: req.IsShow, Sort: req.Sort, Status: req.Status, CreateBy: userId, Scope: common.PMSGovernanceScope(writeScope)})
	if err != nil {
		logc.Errorf(l.ctx, "添加商品属性失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	return &types.BaseResp{Code: "000000", Message: "添加商品属性成功"}, nil
}
