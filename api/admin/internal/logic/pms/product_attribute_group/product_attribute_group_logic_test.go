package product_attribute_group

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productattributegroupservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProductAttributeGroupService struct {
	addFn       func(context.Context, *productattributegroupservice.AddProductAttributeGroupReq, ...grpc.CallOption) (*productattributegroupservice.AddProductAttributeGroupResp, error)
	deleteFn    func(context.Context, *productattributegroupservice.DeleteProductAttributeGroupReq, ...grpc.CallOption) (*productattributegroupservice.DeleteProductAttributeGroupResp, error)
	queryListFn func(context.Context, *productattributegroupservice.QueryProductAttributeGroupListReq, ...grpc.CallOption) (*productattributegroupservice.QueryProductAttributeGroupListResp, error)
}

func (f *fakeProductAttributeGroupService) AddProductAttributeGroup(ctx context.Context, in *productattributegroupservice.AddProductAttributeGroupReq, opts ...grpc.CallOption) (*productattributegroupservice.AddProductAttributeGroupResp, error) {
	if f.addFn == nil {
		return &productattributegroupservice.AddProductAttributeGroupResp{}, nil
	}
	return f.addFn(ctx, in, opts...)
}

func (f *fakeProductAttributeGroupService) DeleteProductAttributeGroup(ctx context.Context, in *productattributegroupservice.DeleteProductAttributeGroupReq, opts ...grpc.CallOption) (*productattributegroupservice.DeleteProductAttributeGroupResp, error) {
	if f.deleteFn == nil {
		return &productattributegroupservice.DeleteProductAttributeGroupResp{}, nil
	}
	return f.deleteFn(ctx, in, opts...)
}

func (f *fakeProductAttributeGroupService) UpdateProductAttributeGroup(context.Context, *productattributegroupservice.UpdateProductAttributeGroupReq, ...grpc.CallOption) (*productattributegroupservice.UpdateProductAttributeGroupResp, error) {
	return &productattributegroupservice.UpdateProductAttributeGroupResp{}, nil
}

func (f *fakeProductAttributeGroupService) UpdateProductAttributeGroupStatus(context.Context, *productattributegroupservice.UpdateProductAttributeGroupStatusReq, ...grpc.CallOption) (*productattributegroupservice.UpdateProductAttributeGroupStatusResp, error) {
	return &productattributegroupservice.UpdateProductAttributeGroupStatusResp{}, nil
}

func (f *fakeProductAttributeGroupService) QueryProductAttributeGroupDetail(context.Context, *productattributegroupservice.QueryProductAttributeGroupDetailReq, ...grpc.CallOption) (*productattributegroupservice.QueryProductAttributeGroupDetailResp, error) {
	return &productattributegroupservice.QueryProductAttributeGroupDetailResp{}, nil
}

func (f *fakeProductAttributeGroupService) QueryProductAttributeGroupList(ctx context.Context, in *productattributegroupservice.QueryProductAttributeGroupListReq, opts ...grpc.CallOption) (*productattributegroupservice.QueryProductAttributeGroupListResp, error) {
	if f.queryListFn == nil {
		return &productattributegroupservice.QueryProductAttributeGroupListResp{}, nil
	}
	return f.queryListFn(ctx, in, opts...)
}

func productAttributeGroupTestContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", platformID)
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductAttributeGroupPassesRequestedScope(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductAttributeGroupService: &fakeProductAttributeGroupService{
			addFn: func(_ context.Context, in *productattributegroupservice.AddProductAttributeGroupReq, _ ...grpc.CallOption) (*productattributegroupservice.AddProductAttributeGroupResp, error) {
				if in.CreateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return &productattributegroupservice.AddProductAttributeGroupResp{}, nil
			},
		},
	}

	logic := NewAddProductAttributeGroupLogic(productAttributeGroupTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.AddProductAttributeGroup(&types.AddProductAttributeGroupReq{
		CategoryId: 1,
		Name:       "基础参数",
		Sort:       10,
		Status:     1,
		ScopeType:  "merchant",
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 301,
	})
	if err != nil {
		t.Fatalf("add product attribute group failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDeleteProductAttributeGroupMapsRpcErrorMessage(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductAttributeGroupService: &fakeProductAttributeGroupService{
			deleteFn: func(_ context.Context, in *productattributegroupservice.DeleteProductAttributeGroupReq, _ ...grpc.CallOption) (*productattributegroupservice.DeleteProductAttributeGroupResp, error) {
				if in.UpdateBy != 77 {
					t.Fatalf("unexpected operator: %+v", in)
				}
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 301 {
					t.Fatalf("unexpected scope passthrough: %+v", in.Scope)
				}
				return nil, status.Error(codes.InvalidArgument, "商品属性分组已被商品属性引用，无法删除")
			},
		},
	}

	logic := NewDeleteProductAttributeGroupLogic(productAttributeGroupTestContext("merchant", 1, 10, 301), svcCtx)
	_, err := logic.DeleteProductAttributeGroup(&types.DeleteProductAttributeGroupReq{Ids: []int64{9}})
	if err == nil {
		t.Fatal("expected delete product attribute group error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "商品属性分组已被商品属性引用，无法删除" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestQueryProductAttributeGroupListMapsScopeFields(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ProductAttributeGroupService: &fakeProductAttributeGroupService{
			queryListFn: func(_ context.Context, in *productattributegroupservice.QueryProductAttributeGroupListReq, _ ...grpc.CallOption) (*productattributegroupservice.QueryProductAttributeGroupListResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "tenant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 0 {
					t.Fatalf("unexpected query scope: %+v", in.Scope)
				}
				return &productattributegroupservice.QueryProductAttributeGroupListResp{
					Total: 1,
					List: []*productattributegroupservice.ProductAttributeGroupListData{
						{
							Id:         1,
							CategoryId: 2,
							Name:       "销售属性",
							Sort:       10,
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

	logic := NewQueryProductAttributeGroupListLogic(productAttributeGroupTestContext("platform", 1, 0, 0), svcCtx)
	resp, err := logic.QueryProductAttributeGroupList(&types.QueryProductAttributeGroupListReq{
		Current:    1,
		PageSize:   20,
		CategoryId: 2,
		ScopeType:  "tenant",
		PlatformId: 1,
		TenantId:   10,
	})
	if err != nil {
		t.Fatalf("query product attribute group list failed: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("unexpected response data: %+v", resp.Data)
	}
	item := resp.Data[0]
	if item.ScopeType != "tenant" || item.PlatformId != 1 || item.TenantId != 10 || item.MerchantId != 0 || item.IsDeleted != 0 {
		t.Fatalf("unexpected mapped scope fields: %+v", item)
	}
}
