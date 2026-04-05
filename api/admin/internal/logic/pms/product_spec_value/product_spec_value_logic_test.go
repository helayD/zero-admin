package product_spec_value

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productspecvalueservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProductSpecValueService struct {
	addFn       func(context.Context, *productspecvalueservice.AddProductSpecValueReq, ...grpc.CallOption) (*productspecvalueservice.AddProductSpecValueResp, error)
	deleteFn    func(context.Context, *productspecvalueservice.DeleteProductSpecValueReq, ...grpc.CallOption) (*productspecvalueservice.DeleteProductSpecValueResp, error)
	queryListFn func(context.Context, *productspecvalueservice.QueryProductSpecValueListReq, ...grpc.CallOption) (*productspecvalueservice.QueryProductSpecValueListResp, error)
}

func (f *fakeProductSpecValueService) AddProductSpecValue(ctx context.Context, in *productspecvalueservice.AddProductSpecValueReq, opts ...grpc.CallOption) (*productspecvalueservice.AddProductSpecValueResp, error) {
	if f.addFn == nil {
		return &productspecvalueservice.AddProductSpecValueResp{}, nil
	}
	return f.addFn(ctx, in, opts...)
}

func (f *fakeProductSpecValueService) DeleteProductSpecValue(ctx context.Context, in *productspecvalueservice.DeleteProductSpecValueReq, opts ...grpc.CallOption) (*productspecvalueservice.DeleteProductSpecValueResp, error) {
	if f.deleteFn == nil {
		return &productspecvalueservice.DeleteProductSpecValueResp{}, nil
	}
	return f.deleteFn(ctx, in, opts...)
}

func (f *fakeProductSpecValueService) UpdateProductSpecValue(context.Context, *productspecvalueservice.UpdateProductSpecValueReq, ...grpc.CallOption) (*productspecvalueservice.UpdateProductSpecValueResp, error) {
	return &productspecvalueservice.UpdateProductSpecValueResp{}, nil
}

func (f *fakeProductSpecValueService) UpdateProductSpecValueStatus(context.Context, *productspecvalueservice.UpdateProductSpecValueStatusReq, ...grpc.CallOption) (*productspecvalueservice.UpdateProductSpecValueStatusResp, error) {
	return &productspecvalueservice.UpdateProductSpecValueStatusResp{}, nil
}

func (f *fakeProductSpecValueService) QueryProductSpecValueDetail(context.Context, *productspecvalueservice.QueryProductSpecValueDetailReq, ...grpc.CallOption) (*productspecvalueservice.QueryProductSpecValueDetailResp, error) {
	return &productspecvalueservice.QueryProductSpecValueDetailResp{}, nil
}

func (f *fakeProductSpecValueService) QueryProductSpecValueList(ctx context.Context, in *productspecvalueservice.QueryProductSpecValueListReq, opts ...grpc.CallOption) (*productspecvalueservice.QueryProductSpecValueListResp, error) {
	if f.queryListFn == nil {
		return &productspecvalueservice.QueryProductSpecValueListResp{}, nil
	}
	return f.queryListFn(ctx, in, opts...)
}

func productSpecValueTestContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", platformID)
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductSpecValuePassesRequestedScope(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductSpecValueService: &fakeProductSpecValueService{
			addFn: func(_ context.Context, in *productspecvalueservice.AddProductSpecValueReq, _ ...grpc.CallOption) (*productspecvalueservice.AddProductSpecValueResp, error) {
				if in.CreateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return &productspecvalueservice.AddProductSpecValueResp{}, nil
			},
		},
	}

	logic := NewAddProductSpecValueLogic(productSpecValueTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.AddProductSpecValue(&types.AddProductSpecValueReq{
		SpecId:     3,
		Value:      "红色",
		Sort:       1,
		Status:     1,
		ScopeType:  "merchant",
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 301,
	})
	if err != nil {
		t.Fatalf("add product spec value failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDeleteProductSpecValueMapsRpcErrorMessage(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductSpecValueService: &fakeProductSpecValueService{
			deleteFn: func(_ context.Context, in *productspecvalueservice.DeleteProductSpecValueReq, _ ...grpc.CallOption) (*productspecvalueservice.DeleteProductSpecValueResp, error) {
				if in.UpdateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return nil, status.Error(codes.InvalidArgument, "商品规格值已被 SKU 引用，无法删除")
			},
		},
	}

	logic := NewDeleteProductSpecValueLogic(productSpecValueTestContext("merchant", 1, 10, 301), svcCtx)
	_, err := logic.DeleteProductSpecValue(&types.DeleteProductSpecValueReq{Ids: []int64{7}})
	if err == nil {
		t.Fatal("expected delete product spec value error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "商品规格值已被 SKU 引用，无法删除" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestQueryProductSpecValueListMapsScopeFields(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductSpecValueService: &fakeProductSpecValueService{
			queryListFn: func(_ context.Context, in *productspecvalueservice.QueryProductSpecValueListReq, _ ...grpc.CallOption) (*productspecvalueservice.QueryProductSpecValueListResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "tenant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 0 {
					t.Fatalf("unexpected query scope: %+v", in.Scope)
				}
				return &productspecvalueservice.QueryProductSpecValueListResp{
					Total: 1,
					List: []*productspecvalueservice.ProductSpecValueListData{
						{
							Id:         7,
							SpecId:     3,
							Value:      "红色",
							Sort:       1,
							Status:     1,
							CreateBy:   77,
							CreateTime: "2026-04-03 10:00:00",
							UpdateBy:   77,
							UpdateTime: "2026-04-03 10:30:00",
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

	logic := NewQueryProductSpecValueListLogic(productSpecValueTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.QueryProductSpecValueList(&types.QueryProductSpecValueListReq{
		Current:    1,
		PageSize:   20,
		SpecId:     3,
		ScopeType:  "tenant",
		PlatformId: 1,
		TenantId:   10,
	})
	if err != nil {
		t.Fatalf("query product spec value list failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("unexpected response data: %+v", resp.Data)
	}
	item := resp.Data[0]
	if item.ScopeType != "tenant" || item.PlatformId != 1 || item.TenantId != 10 || item.MerchantId != 0 {
		t.Fatalf("unexpected mapped scope fields: %+v", item)
	}
}
