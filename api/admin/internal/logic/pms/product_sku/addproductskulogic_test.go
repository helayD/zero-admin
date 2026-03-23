package product_sku

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productskuservice"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockProductSkuService struct {
	productskuservice.ProductSkuService
	addFn         func(context.Context, *pmsclient.AddProductSkuReq, ...grpc.CallOption) (*pmsclient.AddProductSkuResp, error)
	updateFn      func(context.Context, *pmsclient.UpdateProductSkuReq, ...grpc.CallOption) (*pmsclient.UpdateProductSkuResp, error)
	deleteFn      func(context.Context, *pmsclient.DeleteProductSkuReq, ...grpc.CallOption) (*pmsclient.DeleteProductSkuResp, error)
	queryListFn   func(context.Context, *pmsclient.QueryProductSkuListReq, ...grpc.CallOption) (*pmsclient.QueryProductSkuListResp, error)
	queryDetailFn func(context.Context, *pmsclient.QueryProductSkuDetailReq, ...grpc.CallOption) (*pmsclient.QueryProductSkuDetailResp, error)
}

func (m *mockProductSkuService) AddProductSku(ctx context.Context, in *pmsclient.AddProductSkuReq, opts ...grpc.CallOption) (*pmsclient.AddProductSkuResp, error) {
	return m.addFn(ctx, in, opts...)
}

func (m *mockProductSkuService) UpdateProductSku(ctx context.Context, in *pmsclient.UpdateProductSkuReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSkuResp, error) {
	return m.updateFn(ctx, in, opts...)
}

func (m *mockProductSkuService) DeleteProductSku(ctx context.Context, in *pmsclient.DeleteProductSkuReq, opts ...grpc.CallOption) (*pmsclient.DeleteProductSkuResp, error) {
	return m.deleteFn(ctx, in, opts...)
}

func (m *mockProductSkuService) QueryProductSkuList(ctx context.Context, in *pmsclient.QueryProductSkuListReq, opts ...grpc.CallOption) (*pmsclient.QueryProductSkuListResp, error) {
	return m.queryListFn(ctx, in, opts...)
}

func (m *mockProductSkuService) QueryProductSkuDetail(ctx context.Context, in *pmsclient.QueryProductSkuDetailReq, opts ...grpc.CallOption) (*pmsclient.QueryProductSkuDetailResp, error) {
	return m.queryDetailFn(ctx, in, opts...)
}

func newAdminProductSkuContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("1001"))
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", platformID)
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductSkuPassesSpuIDAndScopeToRPC(t *testing.T) {
	ctx := newAdminProductSkuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.AddProductSkuReq

	logic := NewAddProductSkuLogic(ctx, &svc.ServiceContext{
		ProductSkuService: &mockProductSkuService{
			addFn: func(ctx context.Context, in *pmsclient.AddProductSkuReq, opts ...grpc.CallOption) (*pmsclient.AddProductSkuResp, error) {
				captured = in
				return &pmsclient.AddProductSkuResp{}, nil
			},
		},
	})

	_, err := logic.AddProductSku(&types.AddProductSkuReq{
		SpuId:         2001,
		Name:          "黑-L",
		MainPic:       "main.png",
		AlbumPics:     "a.png,b.png",
		Price:         99.5,
		Stock:         10,
		LowStock:      2,
		SpecData:      `{"颜色":"黑","尺码":"L"}`,
		Weight:        1.2,
		PublishStatus: 0,
		VerifyStatus:  0,
		Sort:          1,
	})
	if err != nil {
		t.Fatalf("AddProductSku returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.SpuId != 2001 {
		t.Fatalf("expected SpuId=2001, got %d", captured.SpuId)
	}
	if captured.CreateBy != 1001 {
		t.Fatalf("expected CreateBy=1001, got %d", captured.CreateBy)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
}

func TestAddProductSkuMapsRpcErrorMessage(t *testing.T) {
	ctx := newAdminProductSkuContext("merchant", 1, 88, 3001)
	logic := NewAddProductSkuLogic(ctx, &svc.ServiceContext{
		ProductSkuService: &mockProductSkuService{
			addFn: func(ctx context.Context, in *pmsclient.AddProductSkuReq, opts ...grpc.CallOption) (*pmsclient.AddProductSkuResp, error) {
				return nil, status.Error(codes.InvalidArgument, "SKU编码已被当前商户占用")
			},
		},
	})

	_, err := logic.AddProductSku(&types.AddProductSkuReq{
		SpuId:         2001,
		Name:          "黑-L",
		MainPic:       "main.png",
		AlbumPics:     "a.png,b.png",
		Price:         99.5,
		Stock:         10,
		LowStock:      2,
		SpecData:      `{"颜色":"黑","尺码":"L"}`,
		Weight:        1.2,
		PublishStatus: 0,
		VerifyStatus:  0,
		Sort:          1,
	})
	if err == nil {
		t.Fatal("expected AddProductSku error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "SKU编码已被当前商户占用" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestQueryProductSkuListPassesScopeAndFiltersToRPC(t *testing.T) {
	ctx := newAdminProductSkuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.QueryProductSkuListReq

	logic := NewQueryProductSkuListLogic(ctx, &svc.ServiceContext{
		ProductSkuService: &mockProductSkuService{
			queryListFn: func(ctx context.Context, in *pmsclient.QueryProductSkuListReq, opts ...grpc.CallOption) (*pmsclient.QueryProductSkuListResp, error) {
				captured = in
				return &pmsclient.QueryProductSkuListResp{
					Total: 1,
					List: []*pmsclient.ProductSkuListData{{
						Id:        31,
						SpuId:     2001,
						Name:      "黑-L",
						SkuCode:   "SKU-001",
						Stock:     10,
						LowStock:  2,
						Price:     99.5,
						MainPic:   "main.png",
						AlbumPics: "a.png,b.png",
					}},
				}, nil
			},
		},
	})

	resp, err := logic.QueryProductSkuList(&types.QueryProductSkuListReq{
		Current:       1,
		PageSize:      20,
		SpuId:         2001,
		Name:          "黑-L",
		SkuCode:       "SKU-001",
		PublishStatus: 1,
		VerifyStatus:  1,
	})
	if err != nil {
		t.Fatalf("QueryProductSkuList returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.SpuId != 2001 || captured.Name != "黑-L" || captured.SkuCode != "SKU-001" {
		t.Fatalf("unexpected filters passed to rpc: %+v", captured)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
	if resp.Total != 1 || len(resp.Data) != 1 || resp.Data[0].SpuId != 2001 || resp.Data[0].SkuCode != "SKU-001" {
		t.Fatalf("unexpected response mapping: %+v", resp)
	}
}

func TestUpdateProductSkuPassesScopeAndUpdateByToRPC(t *testing.T) {
	ctx := newAdminProductSkuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.UpdateProductSkuReq

	logic := NewUpdateProductSkuLogic(ctx, &svc.ServiceContext{
		ProductSkuService: &mockProductSkuService{
			updateFn: func(ctx context.Context, in *pmsclient.UpdateProductSkuReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSkuResp, error) {
				captured = in
				return &pmsclient.UpdateProductSkuResp{}, nil
			},
		},
	})

	_, err := logic.UpdateProductSku(&types.UpdateProductSkuReq{
		Data: []types.UpdateProductSkuData{{
			Id:            31,
			SpuId:         2001,
			Name:          "黑-L",
			SkuCode:       "SKU-001",
			MainPic:       "main.png",
			AlbumPics:     "a.png,b.png",
			Price:         99.5,
			Stock:         10,
			LowStock:      2,
			SpecData:      `{"颜色":"黑","尺码":"L"}`,
			Weight:        1.2,
			PublishStatus: 0,
			VerifyStatus:  0,
			Sort:          1,
		}},
	})
	if err != nil {
		t.Fatalf("UpdateProductSku returned error: %v", err)
	}
	if captured == nil || len(captured.Data) != 1 {
		t.Fatalf("expected rpc update request to be captured, got %+v", captured)
	}
	if captured.Data[0].UpdateBy != 1001 || captured.Data[0].SpuId != 2001 || captured.Data[0].Id != 31 {
		t.Fatalf("unexpected update payload passed to rpc: %+v", captured.Data[0])
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
}

func TestDeleteProductSkuMapsRpcErrorMessageAndScope(t *testing.T) {
	ctx := newAdminProductSkuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.DeleteProductSkuReq

	logic := NewDeleteProductSkuLogic(ctx, &svc.ServiceContext{
		ProductSkuService: &mockProductSkuService{
			deleteFn: func(ctx context.Context, in *pmsclient.DeleteProductSkuReq, opts ...grpc.CallOption) (*pmsclient.DeleteProductSkuResp, error) {
				captured = in
				return nil, status.Error(codes.PermissionDenied, "当前主体无权删除该SKU")
			},
		},
	})

	_, err := logic.DeleteProductSku(&types.DeleteProductSkuReq{
		Ids: []int64{31},
	})
	if captured == nil {
		t.Fatal("expected rpc delete request to be captured")
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
	if err == nil {
		t.Fatal("expected DeleteProductSku error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "当前主体无权删除该SKU" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}
