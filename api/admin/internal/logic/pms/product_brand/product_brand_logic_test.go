package product_brand

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productbrandservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProductBrandService struct {
	addFn    func(context.Context, *productbrandservice.AddProductBrandReq, ...grpc.CallOption) (*productbrandservice.AddProductBrandResp, error)
	deleteFn func(context.Context, *productbrandservice.DeleteProductBrandReq, ...grpc.CallOption) (*productbrandservice.DeleteProductBrandResp, error)
}

func (f *fakeProductBrandService) AddProductBrand(ctx context.Context, in *productbrandservice.AddProductBrandReq, opts ...grpc.CallOption) (*productbrandservice.AddProductBrandResp, error) {
	if f.addFn == nil {
		return &productbrandservice.AddProductBrandResp{}, nil
	}
	return f.addFn(ctx, in, opts...)
}

func (f *fakeProductBrandService) DeleteProductBrand(ctx context.Context, in *productbrandservice.DeleteProductBrandReq, opts ...grpc.CallOption) (*productbrandservice.DeleteProductBrandResp, error) {
	if f.deleteFn == nil {
		return &productbrandservice.DeleteProductBrandResp{}, nil
	}
	return f.deleteFn(ctx, in, opts...)
}

func (f *fakeProductBrandService) UpdateProductBrand(context.Context, *productbrandservice.UpdateProductBrandReq, ...grpc.CallOption) (*productbrandservice.UpdateProductBrandResp, error) {
	return &productbrandservice.UpdateProductBrandResp{}, nil
}
func (f *fakeProductBrandService) UpdateProductBrandStatus(context.Context, *productbrandservice.UpdateProductBrandStatusReq, ...grpc.CallOption) (*productbrandservice.UpdateProductBrandStatusResp, error) {
	return &productbrandservice.UpdateProductBrandStatusResp{}, nil
}
func (f *fakeProductBrandService) QueryProductBrandDetail(context.Context, *productbrandservice.QueryProductBrandDetailReq, ...grpc.CallOption) (*productbrandservice.QueryProductBrandDetailResp, error) {
	return &productbrandservice.QueryProductBrandDetailResp{}, nil
}
func (f *fakeProductBrandService) QueryProductBrandList(context.Context, *productbrandservice.QueryProductBrandListReq, ...grpc.CallOption) (*productbrandservice.QueryProductBrandListResp, error) {
	return &productbrandservice.QueryProductBrandListResp{}, nil
}
func (f *fakeProductBrandService) QueryBrandListByIds(context.Context, *productbrandservice.QueryBrandListByIdsReq, ...grpc.CallOption) (*productbrandservice.QueryProductBrandListResp, error) {
	return &productbrandservice.QueryProductBrandListResp{}, nil
}
func (f *fakeProductBrandService) UpdateBrandRecommendStatus(context.Context, *productbrandservice.UpdateProductBrandStatusReq, ...grpc.CallOption) (*productbrandservice.UpdateProductBrandStatusResp, error) {
	return &productbrandservice.UpdateProductBrandStatusResp{}, nil
}
func (f *fakeProductBrandService) UpdateBrandSort(context.Context, *productbrandservice.UpdateProductBrandSortReq, ...grpc.CallOption) (*productbrandservice.UpdateProductBrandStatusResp, error) {
	return &productbrandservice.UpdateProductBrandStatusResp{}, nil
}

func productBrandTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", "platform")
	ctx = context.WithValue(ctx, "platformId", int64(1))
	ctx = context.WithValue(ctx, "tenantId", int64(0))
	ctx = context.WithValue(ctx, "merchantId", int64(0))
	return ctx
}

func TestAddProductBrandLogicPassesRequestedScope(t *testing.T) {
	svcCtx := &svc.ServiceContext{ProductBrandService: &fakeProductBrandService{addFn: func(_ context.Context, in *productbrandservice.AddProductBrandReq, _ ...grpc.CallOption) (*productbrandservice.AddProductBrandResp, error) {
		if in.CreateBy != 77 {
			t.Fatalf("unexpected operator: %+v", in)
		}
		if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
			t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
		}
		return &productbrandservice.AddProductBrandResp{}, nil
	}}}

	logic := NewAddProductBrandLogic(productBrandTestContext(), svcCtx)
	resp, err := logic.AddProductBrand(&types.AddProductBrandReq{Name: "品牌A", Logo: "a", BigPic: "b", Description: "desc", FirstLetter: "A", Sort: 1, RecommendStatus: 1, IsEnabled: 1, ScopeType: "merchant", PlatformId: 1, TenantId: 10, MerchantId: 301})
	if err != nil {
		t.Fatalf("add product brand failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDeleteProductBrandLogicMapsRpcErrorMessage(t *testing.T) {
	svcCtx := &svc.ServiceContext{ProductBrandService: &fakeProductBrandService{deleteFn: func(_ context.Context, in *productbrandservice.DeleteProductBrandReq, _ ...grpc.CallOption) (*productbrandservice.DeleteProductBrandResp, error) {
		if in.Scope == nil || in.Scope.ScopeType != "platform" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 0 || in.Scope.MerchantId != 0 {
			t.Fatalf("unexpected scope: %+v", in.Scope)
		}
		return nil, status.Error(codes.InvalidArgument, "商品品牌已被商品建档引用，无法删除")
	}}}

	logic := NewDeleteProductBrandLogic(productBrandTestContext(), svcCtx)
	_, err := logic.DeleteProductBrand(&types.DeleteProductBrandReq{Ids: []int64{1}})
	if err == nil {
		t.Fatal("expected delete product brand error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "商品品牌已被商品建档引用，无法删除" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}
