package product_spec

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productspecservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProductSpecService struct {
	addFn       func(context.Context, *productspecservice.AddProductSpecReq, ...grpc.CallOption) (*productspecservice.AddProductSpecResp, error)
	deleteFn    func(context.Context, *productspecservice.DeleteProductSpecReq, ...grpc.CallOption) (*productspecservice.DeleteProductSpecResp, error)
	queryListFn func(context.Context, *productspecservice.QueryProductSpecListReq, ...grpc.CallOption) (*productspecservice.QueryProductSpecListResp, error)
}

func (f *fakeProductSpecService) AddProductSpec(ctx context.Context, in *productspecservice.AddProductSpecReq, opts ...grpc.CallOption) (*productspecservice.AddProductSpecResp, error) {
	if f.addFn == nil {
		return &productspecservice.AddProductSpecResp{}, nil
	}
	return f.addFn(ctx, in, opts...)
}

func (f *fakeProductSpecService) DeleteProductSpec(ctx context.Context, in *productspecservice.DeleteProductSpecReq, opts ...grpc.CallOption) (*productspecservice.DeleteProductSpecResp, error) {
	if f.deleteFn == nil {
		return &productspecservice.DeleteProductSpecResp{}, nil
	}
	return f.deleteFn(ctx, in, opts...)
}

func (f *fakeProductSpecService) UpdateProductSpec(context.Context, *productspecservice.UpdateProductSpecReq, ...grpc.CallOption) (*productspecservice.UpdateProductSpecResp, error) {
	return &productspecservice.UpdateProductSpecResp{}, nil
}

func (f *fakeProductSpecService) UpdateProductSpecStatus(context.Context, *productspecservice.UpdateProductSpecStatusReq, ...grpc.CallOption) (*productspecservice.UpdateProductSpecStatusResp, error) {
	return &productspecservice.UpdateProductSpecStatusResp{}, nil
}

func (f *fakeProductSpecService) QueryProductSpecDetail(context.Context, *productspecservice.QueryProductSpecDetailReq, ...grpc.CallOption) (*productspecservice.QueryProductSpecDetailResp, error) {
	return &productspecservice.QueryProductSpecDetailResp{}, nil
}

func (f *fakeProductSpecService) QueryProductSpecList(ctx context.Context, in *productspecservice.QueryProductSpecListReq, opts ...grpc.CallOption) (*productspecservice.QueryProductSpecListResp, error) {
	if f.queryListFn == nil {
		return &productspecservice.QueryProductSpecListResp{}, nil
	}
	return f.queryListFn(ctx, in, opts...)
}

func productSpecTestContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", platformID)
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductSpecPassesRequestedScope(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductSpecService: &fakeProductSpecService{
			addFn: func(_ context.Context, in *productspecservice.AddProductSpecReq, _ ...grpc.CallOption) (*productspecservice.AddProductSpecResp, error) {
				if in.CreateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return &productspecservice.AddProductSpecResp{}, nil
			},
		},
	}

	logic := NewAddProductSpecLogic(productSpecTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.AddProductSpec(&types.AddProductSpecReq{
		CategoryId: 2,
		Name:       "颜色",
		Sort:       1,
		Status:     1,
		ScopeType:  "merchant",
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 301,
	})
	if err != nil {
		t.Fatalf("add product spec failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDeleteProductSpecMapsRpcErrorMessage(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductSpecService: &fakeProductSpecService{
			deleteFn: func(_ context.Context, in *productspecservice.DeleteProductSpecReq, _ ...grpc.CallOption) (*productspecservice.DeleteProductSpecResp, error) {
				if in.UpdateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return nil, status.Error(codes.InvalidArgument, "商品规格已被商品建档引用，无法删除")
			},
		},
	}

	logic := NewDeleteProductSpecLogic(productSpecTestContext("merchant", 1, 10, 301), svcCtx)
	_, err := logic.DeleteProductSpec(&types.DeleteProductSpecReq{Ids: []int64{3}})
	if err == nil {
		t.Fatal("expected delete product spec error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "商品规格已被商品建档引用，无法删除" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestQueryProductSpecListMapsScopeFields(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductSpecService: &fakeProductSpecService{
			queryListFn: func(_ context.Context, in *productspecservice.QueryProductSpecListReq, _ ...grpc.CallOption) (*productspecservice.QueryProductSpecListResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "tenant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 0 {
					t.Fatalf("unexpected query scope: %+v", in.Scope)
				}
				return &productspecservice.QueryProductSpecListResp{
					Total: 1,
					List: []*productspecservice.ProductSpecListData{
						{
							Id:         8,
							CategoryId: 2,
							Name:       "颜色",
							Sort:       1,
							Status:     1,
							CreateBy:   77,
							CreateTime: "2026-04-03 10:00:00",
							UpdateBy:   77,
							UpdateTime: "2026-04-03 10:30:00",
							IsDeleted:  0,
							ScopeType:  "tenant",
							PlatformId: 1,
							TenantId:   10,
							MerchantId: 0,
						},
					},
				}, nil
			},
		},
	}

	logic := NewQueryProductSpecListLogic(productSpecTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.QueryProductSpecList(&types.QueryProductSpecListReq{
		Current:    1,
		PageSize:   20,
		CategoryId: 2,
		ScopeType:  "tenant",
		PlatformId: 1,
		TenantId:   10,
	})
	if err != nil {
		t.Fatalf("query product spec list failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("unexpected response data: %+v", resp.Data)
	}
	item := resp.Data[0]
	if item.ScopeType != "tenant" || item.PlatformId != 1 || item.TenantId != 10 || item.MerchantId != 0 || item.IsDeleted != 0 {
		t.Fatalf("unexpected mapped scope fields: %+v", item)
	}
}
