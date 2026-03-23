package product_brand

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productbrandservice"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"google.golang.org/grpc"
)

type mockProductBrandService struct {
	productbrandservice.ProductBrandService
	addFn func(ctx context.Context, in *pmsclient.AddProductBrandReq, opts ...grpc.CallOption) (*pmsclient.AddProductBrandResp, error)
}

func (m *mockProductBrandService) AddProductBrand(ctx context.Context, in *pmsclient.AddProductBrandReq, opts ...grpc.CallOption) (*pmsclient.AddProductBrandResp, error) {
	return m.addFn(ctx, in, opts...)
}

func newAdminScopeContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("1001"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	if platformID != 1 {
		ctx = context.WithValue(ctx, "platformId", platformID)
	}
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductBrandPassesResolvedScopeToRPC(t *testing.T) {
	ctx := newAdminScopeContext("merchant", 1, 88, 3001)
	var captured *pmsclient.AddProductBrandReq

	logic := NewAddProductBrandLogic(ctx, &svc.ServiceContext{
		ProductBrandService: &mockProductBrandService{
			addFn: func(ctx context.Context, in *pmsclient.AddProductBrandReq, opts ...grpc.CallOption) (*pmsclient.AddProductBrandResp, error) {
				captured = in
				return &pmsclient.AddProductBrandResp{}, nil
			},
		},
	})

	_, err := logic.AddProductBrand(&types.AddProductBrandReq{
		Name:            "品牌A",
		Logo:            "logo.png",
		BigPic:          "big.png",
		FirstLetter:     "A",
		Sort:            1,
		RecommendStatus: 1,
		IsEnabled:       1,
	})
	if err != nil {
		t.Fatalf("AddProductBrand returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.CreateBy != 1001 {
		t.Fatalf("expected CreateBy=1001, got %d", captured.CreateBy)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
}
