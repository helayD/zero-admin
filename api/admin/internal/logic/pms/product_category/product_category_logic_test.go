package product_category

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productcategoryservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProductCategoryService struct {
	addFn func(context.Context, *productcategoryservice.AddProductCategoryReq, ...grpc.CallOption) (*productcategoryservice.AddProductCategoryResp, error)
}

func (f *fakeProductCategoryService) AddProductCategory(ctx context.Context, in *productcategoryservice.AddProductCategoryReq, opts ...grpc.CallOption) (*productcategoryservice.AddProductCategoryResp, error) {
	if f.addFn == nil {
		return &productcategoryservice.AddProductCategoryResp{}, nil
	}
	return f.addFn(ctx, in, opts...)
}
func (f *fakeProductCategoryService) DeleteProductCategory(context.Context, *productcategoryservice.DeleteProductCategoryReq, ...grpc.CallOption) (*productcategoryservice.DeleteProductCategoryResp, error) {
	return &productcategoryservice.DeleteProductCategoryResp{}, nil
}
func (f *fakeProductCategoryService) UpdateProductCategory(context.Context, *productcategoryservice.UpdateProductCategoryReq, ...grpc.CallOption) (*productcategoryservice.UpdateProductCategoryResp, error) {
	return &productcategoryservice.UpdateProductCategoryResp{}, nil
}
func (f *fakeProductCategoryService) UpdateCategoryNavStatus(context.Context, *productcategoryservice.UpdateProductCategoryStatusReq, ...grpc.CallOption) (*productcategoryservice.UpdateProductCategoryStatusResp, error) {
	return &productcategoryservice.UpdateProductCategoryStatusResp{}, nil
}
func (f *fakeProductCategoryService) UpdateProductCategoryStatus(context.Context, *productcategoryservice.UpdateProductCategoryStatusReq, ...grpc.CallOption) (*productcategoryservice.UpdateProductCategoryStatusResp, error) {
	return &productcategoryservice.UpdateProductCategoryStatusResp{}, nil
}
func (f *fakeProductCategoryService) QueryProductCategoryDetail(context.Context, *productcategoryservice.QueryProductCategoryDetailReq, ...grpc.CallOption) (*productcategoryservice.QueryProductCategoryDetailResp, error) {
	return &productcategoryservice.QueryProductCategoryDetailResp{}, nil
}
func (f *fakeProductCategoryService) QueryProductCategoryList(context.Context, *productcategoryservice.QueryProductCategoryListReq, ...grpc.CallOption) (*productcategoryservice.QueryProductCategoryListResp, error) {
	return &productcategoryservice.QueryProductCategoryListResp{}, nil
}
func (f *fakeProductCategoryService) QueryProductCategoryTreeList(context.Context, *productcategoryservice.QueryProductCategoryTreeListReq, ...grpc.CallOption) (*productcategoryservice.QueryProductCategoryListTreeResp, error) {
	return &productcategoryservice.QueryProductCategoryListTreeResp{}, nil
}

func productCategoryTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", int64(1))
	ctx = context.WithValue(ctx, "tenantId", int64(10))
	ctx = context.WithValue(ctx, "merchantId", int64(301))
	return ctx
}

func TestAddProductCategoryLogicPassesCurrentScopeByDefault(t *testing.T) {
	svcCtx := &svc.ServiceContext{ProductCategoryService: &fakeProductCategoryService{addFn: func(_ context.Context, in *productcategoryservice.AddProductCategoryReq, _ ...grpc.CallOption) (*productcategoryservice.AddProductCategoryResp, error) {
		if in.CreateBy != 77 {
			t.Fatalf("unexpected operator: %+v", in)
		}
		if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
			t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
		}
		if len(in.ProductAttributeIdList) != 2 || in.ProductAttributeIdList[0] != 8 {
			t.Fatalf("unexpected attribute binding payload: %+v", in.ProductAttributeIdList)
		}
		return &productcategoryservice.AddProductCategoryResp{Pong: "ok"}, nil
	}}}

	logic := NewAddProductCategoryLogic(productCategoryTestContext(), svcCtx)
	resp, err := logic.AddProductCategory(&types.AddProductCategoryReq{ParentId: 0, Name: "手机", Level: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, ProductAttributeIdList: []int64{8, 9}})
	if err != nil {
		t.Fatalf("add product category failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAddProductCategoryLogicMapsRpcErrorMessage(t *testing.T) {
	svcCtx := &svc.ServiceContext{ProductCategoryService: &fakeProductCategoryService{addFn: func(_ context.Context, _ *productcategoryservice.AddProductCategoryReq, _ ...grpc.CallOption) (*productcategoryservice.AddProductCategoryResp, error) {
		return nil, status.Error(codes.InvalidArgument, "只能绑定启用状态的商品属性")
	}}}

	logic := NewAddProductCategoryLogic(productCategoryTestContext(), svcCtx)
	_, err := logic.AddProductCategory(&types.AddProductCategoryReq{Name: "手机", Level: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, IsEnabled: 1})
	if err == nil {
		t.Fatal("expected add product category error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "只能绑定启用状态的商品属性" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}
