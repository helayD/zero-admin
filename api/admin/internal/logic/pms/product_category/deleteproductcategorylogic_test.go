package product_category

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productcategoryservice"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockProductCategoryService struct {
	productcategoryservice.ProductCategoryService
	deleteFn func(ctx context.Context, in *pmsclient.DeleteProductCategoryReq, opts ...grpc.CallOption) (*pmsclient.DeleteProductCategoryResp, error)
}

func (m *mockProductCategoryService) DeleteProductCategory(ctx context.Context, in *pmsclient.DeleteProductCategoryReq, opts ...grpc.CallOption) (*pmsclient.DeleteProductCategoryResp, error) {
	return m.deleteFn(ctx, in, opts...)
}

func newAdminCategoryScopeContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("2002"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	if platformID != 1 {
		ctx = context.WithValue(ctx, "platformId", platformID)
	}
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestDeleteProductCategoryPassesScopeAndMapsRPCError(t *testing.T) {
	ctx := newAdminCategoryScopeContext("platform", 1, 0, 0)
	var captured *pmsclient.DeleteProductCategoryReq

	logic := NewDeleteProductCategoryLogic(ctx, &svc.ServiceContext{
		ProductCategoryService: &mockProductCategoryService{
			deleteFn: func(ctx context.Context, in *pmsclient.DeleteProductCategoryReq, opts ...grpc.CallOption) (*pmsclient.DeleteProductCategoryResp, error) {
				captured = in
				return nil, status.Error(codes.InvalidArgument, "商品分类已被商品SPU引用，暂不可删除")
			},
		},
	})

	_, err := logic.DeleteProductCategory(&types.DeleteProductCategoryReq{
		Ids:        []int64{9},
		ScopeType:  "merchant",
		PlatformId: 1,
		TenantId:   88,
		MerchantId: 3001,
	})
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.UpdateBy != 2002 {
		t.Fatalf("expected UpdateBy=2002, got %d", captured.UpdateBy)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
	if err == nil {
		t.Fatal("expected mapped error")
	}
	codeErr, ok := err.(*errorx.CodeError)
	if !ok {
		t.Fatalf("expected CodeError, got %T", err)
	}
	if codeErr.Message != "商品分类已被商品SPU引用，暂不可删除" {
		t.Fatalf("unexpected mapped message: %s", codeErr.Message)
	}
}
