package productfulfillmentruleservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/internal/logic/productfulfillmentruleservice"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

type ProductFulfillmentRuleServiceServer struct {
	svcCtx *svc.ServiceContext
	smsclient.UnimplementedProductFulfillmentRuleServiceServer
}

func NewProductFulfillmentRuleServiceServer(svcCtx *svc.ServiceContext) *ProductFulfillmentRuleServiceServer {
	return &ProductFulfillmentRuleServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *ProductFulfillmentRuleServiceServer) AddProductFulfillmentRule(ctx context.Context, in *smsclient.AddProductFulfillmentRuleReq) (*smsclient.AddProductFulfillmentRuleResp, error) {
	l := productfulfillmentruleservice.NewAddProductFulfillmentRuleLogic(ctx, s.svcCtx)
	return l.AddProductFulfillmentRule(in)
}

func (s *ProductFulfillmentRuleServiceServer) UpdateProductFulfillmentRule(ctx context.Context, in *smsclient.UpdateProductFulfillmentRuleReq) (*smsclient.UpdateProductFulfillmentRuleResp, error) {
	l := productfulfillmentruleservice.NewUpdateProductFulfillmentRuleLogic(ctx, s.svcCtx)
	return l.UpdateProductFulfillmentRule(in)
}

func (s *ProductFulfillmentRuleServiceServer) QueryProductFulfillmentRuleList(ctx context.Context, in *smsclient.QueryProductFulfillmentRuleListReq) (*smsclient.QueryProductFulfillmentRuleListResp, error) {
	l := productfulfillmentruleservice.NewQueryProductFulfillmentRuleListLogic(ctx, s.svcCtx)
	return l.QueryProductFulfillmentRuleList(in)
}

func (s *ProductFulfillmentRuleServiceServer) QueryProductFulfillmentRuleDetail(ctx context.Context, in *smsclient.QueryProductFulfillmentRuleDetailReq) (*smsclient.QueryProductFulfillmentRuleDetailResp, error) {
	l := productfulfillmentruleservice.NewQueryProductFulfillmentRuleDetailLogic(ctx, s.svcCtx)
	return l.QueryProductFulfillmentRuleDetail(in)
}

func (s *ProductFulfillmentRuleServiceServer) UpdateProductFulfillmentRuleStatus(ctx context.Context, in *smsclient.UpdateProductFulfillmentRuleStatusReq) (*smsclient.UpdateProductFulfillmentRuleStatusResp, error) {
	l := productfulfillmentruleservice.NewUpdateProductFulfillmentRuleStatusLogic(ctx, s.svcCtx)
	return l.UpdateProductFulfillmentRuleStatus(in)
}

func (s *ProductFulfillmentRuleServiceServer) DeleteProductFulfillmentRule(ctx context.Context, in *smsclient.DeleteProductFulfillmentRuleReq) (*smsclient.DeleteProductFulfillmentRuleResp, error) {
	l := productfulfillmentruleservice.NewDeleteProductFulfillmentRuleLogic(ctx, s.svcCtx)
	return l.DeleteProductFulfillmentRule(in)
}

func (s *ProductFulfillmentRuleServiceServer) CheckProductFulfillmentRuleBinding(ctx context.Context, in *smsclient.CheckProductFulfillmentRuleBindingReq) (*smsclient.CheckProductFulfillmentRuleBindingResp, error) {
	l := productfulfillmentruleservice.NewCheckProductFulfillmentRuleBindingLogic(ctx, s.svcCtx)
	return l.CheckProductFulfillmentRuleBinding(in)
}
