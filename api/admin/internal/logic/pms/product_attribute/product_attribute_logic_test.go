package product_attribute

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productattributeservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProductAttributeService struct {
	addFn       func(context.Context, *productattributeservice.AddProductAttributeReq, ...grpc.CallOption) (*productattributeservice.AddProductAttributeResp, error)
	deleteFn    func(context.Context, *productattributeservice.DeleteProductAttributeReq, ...grpc.CallOption) (*productattributeservice.DeleteProductAttributeResp, error)
	queryListFn func(context.Context, *productattributeservice.QueryProductAttributeListReq, ...grpc.CallOption) (*productattributeservice.QueryProductAttributeListResp, error)
}

func (f *fakeProductAttributeService) AddProductAttribute(ctx context.Context, in *productattributeservice.AddProductAttributeReq, opts ...grpc.CallOption) (*productattributeservice.AddProductAttributeResp, error) {
	if f.addFn == nil {
		return &productattributeservice.AddProductAttributeResp{}, nil
	}
	return f.addFn(ctx, in, opts...)
}

func (f *fakeProductAttributeService) DeleteProductAttribute(ctx context.Context, in *productattributeservice.DeleteProductAttributeReq, opts ...grpc.CallOption) (*productattributeservice.DeleteProductAttributeResp, error) {
	if f.deleteFn == nil {
		return &productattributeservice.DeleteProductAttributeResp{}, nil
	}
	return f.deleteFn(ctx, in, opts...)
}

func (f *fakeProductAttributeService) UpdateProductAttribute(context.Context, *productattributeservice.UpdateProductAttributeReq, ...grpc.CallOption) (*productattributeservice.UpdateProductAttributeResp, error) {
	return &productattributeservice.UpdateProductAttributeResp{}, nil
}

func (f *fakeProductAttributeService) UpdateProductAttributeStatus(context.Context, *productattributeservice.UpdateProductAttributeStatusReq, ...grpc.CallOption) (*productattributeservice.UpdateProductAttributeStatusResp, error) {
	return &productattributeservice.UpdateProductAttributeStatusResp{}, nil
}

func (f *fakeProductAttributeService) QueryProductAttributeDetail(context.Context, *productattributeservice.QueryProductAttributeDetailReq, ...grpc.CallOption) (*productattributeservice.QueryProductAttributeDetailResp, error) {
	return &productattributeservice.QueryProductAttributeDetailResp{}, nil
}

func (f *fakeProductAttributeService) QueryProductAttributeList(ctx context.Context, in *productattributeservice.QueryProductAttributeListReq, opts ...grpc.CallOption) (*productattributeservice.QueryProductAttributeListResp, error) {
	if f.queryListFn == nil {
		return &productattributeservice.QueryProductAttributeListResp{}, nil
	}
	return f.queryListFn(ctx, in, opts...)
}

func productAttributeTestContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", platformID)
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductAttributePassesRequestedScope(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductAttributeService: &fakeProductAttributeService{
			addFn: func(_ context.Context, in *productattributeservice.AddProductAttributeReq, _ ...grpc.CallOption) (*productattributeservice.AddProductAttributeResp, error) {
				if in.CreateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return &productattributeservice.AddProductAttributeResp{}, nil
			},
		},
	}

	logic := NewAddProductAttributeLogic(productAttributeTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.AddProductAttribute(&types.AddProductAttributeReq{
		GroupId:      2,
		Name:         "颜色",
		InputType:    1,
		ValueType:    1,
		InputList:    "红,蓝",
		Unit:         "",
		IsRequired:   1,
		IsSearchable: 1,
		IsShow:       1,
		Sort:         1,
		Status:       1,
		ScopeType:    "merchant",
		PlatformId:   1,
		TenantId:     10,
		MerchantId:   301,
	})
	if err != nil {
		t.Fatalf("add product attribute failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDeleteProductAttributeMapsRpcErrorMessage(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductAttributeService: &fakeProductAttributeService{
			deleteFn: func(_ context.Context, in *productattributeservice.DeleteProductAttributeReq, _ ...grpc.CallOption) (*productattributeservice.DeleteProductAttributeResp, error) {
				if in.UpdateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return nil, status.Error(codes.InvalidArgument, "商品属性已被分类引用，无法删除")
			},
		},
	}

	logic := NewDeleteProductAttributeLogic(productAttributeTestContext("merchant", 1, 10, 301), svcCtx)
	_, err := logic.DeleteProductAttribute(&types.DeleteProductAttributeReq{Ids: []int64{5}})
	if err == nil {
		t.Fatal("expected delete product attribute error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "商品属性已被分类引用，无法删除" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestQueryProductAttributeListMapsScopeFields(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductAttributeService: &fakeProductAttributeService{
			queryListFn: func(_ context.Context, in *productattributeservice.QueryProductAttributeListReq, _ ...grpc.CallOption) (*productattributeservice.QueryProductAttributeListResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "tenant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 0 {
					t.Fatalf("unexpected query scope: %+v", in.Scope)
				}
				return &productattributeservice.QueryProductAttributeListResp{
					Total: 1,
					List: []*productattributeservice.ProductAttributeListData{
						{
							Id:           5,
							GroupId:      2,
							Name:         "颜色",
							InputType:    1,
							ValueType:    1,
							InputList:    "红,蓝",
							Unit:         "",
							IsRequired:   1,
							IsSearchable: 1,
							IsShow:       1,
							Sort:         1,
							Status:       1,
							CreateBy:     77,
							CreateTime:   "2026-04-03 10:00:00",
							UpdateBy:     77,
							UpdateTime:   "2026-04-03 10:30:00",
							ScopeType:    "tenant",
							PlatformId:   1,
							TenantId:     10,
							MerchantId:   0,
						},
					},
				}, nil
			},
		},
	}

	logic := NewQueryProductAttributeListLogic(productAttributeTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.QueryProductAttributeList(&types.QueryProductAttributeListReq{
		Current:    1,
		PageSize:   20,
		GroupId:    2,
		ScopeType:  "tenant",
		PlatformId: 1,
		TenantId:   10,
	})
	if err != nil {
		t.Fatalf("query product attribute list failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("unexpected response data: %+v", resp.Data)
	}
	item := resp.Data[0]
	if item.ScopeType != "tenant" || item.PlatformId != 1 || item.TenantId != 10 || item.MerchantId != 0 {
		t.Fatalf("unexpected mapped scope fields: %+v", item)
	}
}
