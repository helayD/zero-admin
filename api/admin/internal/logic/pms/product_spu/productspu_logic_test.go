package product_spu

import (
	"context"
	"encoding/json"
	"testing"

	adminerrorx "github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/client/productspuservice"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockProductSpuService struct {
	productspuservice.ProductSpuService
	addFn             func(context.Context, *pmsclient.ProductSpuReq, ...grpc.CallOption) (*pmsclient.ProductSpuResp, error)
	queryListFn       func(context.Context, *pmsclient.QueryProductSpuListReq, ...grpc.CallOption) (*pmsclient.QueryProductSpuListResp, error)
	updateVerifyFn    func(context.Context, *pmsclient.UpdateProductSpuStatusReq, ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error)
	updatePublishFn   func(context.Context, *pmsclient.UpdateProductSpuStatusReq, ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error)
	updateRecommendFn func(context.Context, *pmsclient.UpdateProductSpuStatusReq, ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error)
}

func (m *mockProductSpuService) AddProductSpu(ctx context.Context, in *pmsclient.ProductSpuReq, opts ...grpc.CallOption) (*pmsclient.ProductSpuResp, error) {
	return m.addFn(ctx, in, opts...)
}

func (m *mockProductSpuService) QueryProductSpuList(ctx context.Context, in *pmsclient.QueryProductSpuListReq, opts ...grpc.CallOption) (*pmsclient.QueryProductSpuListResp, error) {
	return m.queryListFn(ctx, in, opts...)
}

func (m *mockProductSpuService) UpdateVerifyStatus(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
	return m.updateVerifyFn(ctx, in, opts...)
}

func (m *mockProductSpuService) UpdatePublishStatus(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
	return m.updatePublishFn(ctx, in, opts...)
}

func (m *mockProductSpuService) UpdateRecommendStatus(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
	return m.updateRecommendFn(ctx, in, opts...)
}

func newAdminProductSpuContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", json.Number("1001"))
	ctx = context.WithValue(ctx, "userName", "tester")
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", platformID)
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func TestAddProductSpuPassesScopeAndNestedDetailIDsToRPC(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.ProductSpuReq

	logic := NewAddProductSpuLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			addFn: func(ctx context.Context, in *pmsclient.ProductSpuReq, opts ...grpc.CallOption) (*pmsclient.ProductSpuResp, error) {
				captured = in
				return &pmsclient.ProductSpuResp{SpuId: 2001}, nil
			},
		},
	})

	_, err := logic.AddProductSpu(&types.AddProductSpuReq{
		ProductData: types.ProductSpuData{
			Name:          "测试商品",
			ProductSn:     "SN-2001",
			CategoryId:    301,
			CategoryName:  "手机",
			BrandId:       401,
			BrandName:     "品牌A",
			Unit:          "件",
			Keywords:      "5G",
			AlbumPics:     "a.png,b.png",
			MainPic:       "main.png",
			PublishStatus: 0,
			VerifyStatus:  0,
		},
		MemberPriceList: []types.MemberPriceReq{{
			Id:              11,
			MemberLevelId:   101,
			MemberPrice:     8800,
			MemberLevelName: "黄金会员",
		}},
		AttributeValueList: []types.AddProductAttributeValueReq{{
			Id:          12,
			AttributeId: 202,
			Value:       "黑色",
			Status:      1,
		}},
		FullList: []types.ProductFullReductionReq{{
			Id:          13,
			FullPrice:   10000,
			ReducePrice: 1000,
		}},
		LadderList: []types.ProductLadderReq{{
			Id:       14,
			Count:    3,
			Discount: 80,
			Price:    7600,
		}},
		SkuList: []types.AddProductSkuReq{{
			Id:        15,
			SpuId:     2001,
			Name:      "黑-L",
			SkuCode:   "SKU-001",
			MainPic:   "sku-main.png",
			AlbumPics: "sku-a.png,sku-b.png",
			Price:     99.5,
			Stock:     10,
			LowStock:  2,
			SpecData:  `{"颜色":"黑","尺码":"L"}`,
			Weight:    1.2,
		}},
	})
	if err != nil {
		t.Fatalf("AddProductSpu returned error: %v", err)
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
	if len(captured.MemberPriceList) != 1 || captured.MemberPriceList[0].Id != 11 {
		t.Fatalf("member price detail id not passed: %+v", captured.MemberPriceList)
	}
	if len(captured.ProductAttributeValueList) != 1 || captured.ProductAttributeValueList[0].Id != 12 {
		t.Fatalf("attribute value detail id not passed: %+v", captured.ProductAttributeValueList)
	}
	if len(captured.ProductFullReductionList) != 1 || captured.ProductFullReductionList[0].Id != 13 {
		t.Fatalf("full reduction detail id not passed: %+v", captured.ProductFullReductionList)
	}
	if len(captured.ProductLadderList) != 1 || captured.ProductLadderList[0].Id != 14 {
		t.Fatalf("ladder detail id not passed: %+v", captured.ProductLadderList)
	}
	if len(captured.SkuStockList) != 1 || captured.SkuStockList[0].Id != 15 || captured.SkuStockList[0].SpuId != 2001 {
		t.Fatalf("sku detail mapping not passed: %+v", captured.SkuStockList)
	}
}

func TestAddProductSpuMapsRpcErrorMessage(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
	logic := NewAddProductSpuLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			addFn: func(ctx context.Context, in *pmsclient.ProductSpuReq, opts ...grpc.CallOption) (*pmsclient.ProductSpuResp, error) {
				return nil, status.Error(codes.InvalidArgument, "商品品牌已失效，无法继续建档")
			},
		},
	})

	_, err := logic.AddProductSpu(&types.AddProductSpuReq{
		ProductData: types.ProductSpuData{
			Name:         "测试商品",
			ProductSn:    "SN-2002",
			CategoryId:   301,
			CategoryName: "手机",
			BrandId:      401,
			BrandName:    "品牌A",
			Unit:         "件",
			MainPic:      "main.png",
		},
		SkuList: []types.AddProductSkuReq{{
			Name:     "黑-L",
			Price:    99.5,
			Stock:    10,
			LowStock: 2,
			SpecData: `{"颜色":"黑","尺码":"L"}`,
		}},
	})
	if err == nil {
		t.Fatal("expected AddProductSpu error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "商品品牌已失效，无法继续建档" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestQueryProductSpuListPassesProductSnAndScopeToRPC(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.QueryProductSpuListReq

	logic := NewQueryProductSpuListLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			queryListFn: func(ctx context.Context, in *pmsclient.QueryProductSpuListReq, opts ...grpc.CallOption) (*pmsclient.QueryProductSpuListResp, error) {
				captured = in
				return &pmsclient.QueryProductSpuListResp{
					Total: 1,
					List: []*pmsclient.ProductSpuListData{{
						Id:         2001,
						Name:       "测试商品",
						ProductSn:  "SN-2001",
						CategoryId: 301,
						BrandId:    401,
						MainPic:    "main.png",
					}},
				}, nil
			},
		},
	})

	resp, err := logic.QueryProductSpuList(&types.QueryProductSpuListReq{
		Current:   1,
		PageSize:  20,
		Name:      "测试商品",
		ProductSn: "SN-2001",
	})
	if err != nil {
		t.Fatalf("QueryProductSpuList returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.ProductSn != "SN-2001" {
		t.Fatalf("expected ProductSn to be passed, got %q", captured.ProductSn)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
	if resp.Total != 1 || len(resp.Data) != 1 || resp.Data[0].ProductSn != "SN-2001" {
		t.Fatalf("unexpected response mapping: %+v", resp)
	}
}

func TestUpdateVerifyStatusPassesOperatorScopeAndDetailToRPC(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.UpdateProductSpuStatusReq

	logic := NewUpdateVerifyStatusLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			updateVerifyFn: func(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
				captured = in
				return &pmsclient.UpdateProductSpuStatusResp{}, nil
			},
		},
	})

	_, err := logic.UpdateVerifyStatus(&types.UpdateProductSpuStatusReq{
		Ids:    []int64{2001, 2002},
		Status: 1,
		Detail: "审核通过",
	})
	if err != nil {
		t.Fatalf("UpdateVerifyStatus returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.UpdateBy != 1001 {
		t.Fatalf("expected UpdateBy=1001, got %d", captured.UpdateBy)
	}
	if captured.ReviewMan != "tester" {
		t.Fatalf("expected ReviewMan=tester, got %q", captured.ReviewMan)
	}
	if captured.Detail != "审核通过" {
		t.Fatalf("expected detail to pass through, got %q", captured.Detail)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
}

func TestUpdateVerifyStatusMapsRpcErrorMessage(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)

	logic := NewUpdateVerifyStatusLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			updateVerifyFn: func(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
				return nil, status.Error(codes.PermissionDenied, "当前主体无权送审该商品")
			},
		},
	})

	_, err := logic.UpdateVerifyStatus(&types.UpdateProductSpuStatusReq{
		Ids:    []int64{2001},
		Status: 1,
		Detail: "准备送审",
	})
	if err == nil {
		t.Fatal("expected UpdateVerifyStatus error")
	}
	codeErr, ok := err.(*adminerrorx.CodeError)
	if !ok || codeErr.Code != adminerrorx.DefaultCode || codeErr.Message != "当前主体无权送审该商品" {
		t.Fatalf("unexpected mapped error: %#v", err)
	}
}

func TestUpdatePublishStatusPassesOperatorScopeAndDetailToRPC(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.UpdateProductSpuStatusReq

	logic := NewUpdatePublishStatusLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			updatePublishFn: func(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
				captured = in
				return &pmsclient.UpdateProductSpuStatusResp{}, nil
			},
		},
	})

	_, err := logic.UpdatePublishStatus(&types.UpdateProductSpuStatusReq{
		Ids:    []int64{2001},
		Status: 0,
		Detail: "库存盘点后暂时下架",
	})
	if err != nil {
		t.Fatalf("UpdatePublishStatus returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.UpdateBy != 1001 {
		t.Fatalf("expected UpdateBy=1001, got %d", captured.UpdateBy)
	}
	if captured.ReviewMan != "tester" {
		t.Fatalf("expected ReviewMan=tester, got %q", captured.ReviewMan)
	}
	if captured.Detail != "库存盘点后暂时下架" {
		t.Fatalf("expected detail to pass through, got %q", captured.Detail)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
}

func TestUpdateRecommendStatusPassesOperatorScopeAndDetailToRPC(t *testing.T) {
	ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
	var captured *pmsclient.UpdateProductSpuStatusReq

	logic := NewUpdateRecommendStatusLogic(ctx, &svc.ServiceContext{
		ProductSpuService: &mockProductSpuService{
			updateRecommendFn: func(ctx context.Context, in *pmsclient.UpdateProductSpuStatusReq, opts ...grpc.CallOption) (*pmsclient.UpdateProductSpuStatusResp, error) {
				captured = in
				return &pmsclient.UpdateProductSpuStatusResp{}, nil
			},
		},
	})

	_, err := logic.UpdateRecommendStatus(&types.UpdateProductSpuStatusReq{
		Ids:    []int64{2001},
		Status: 1,
		Detail: "加入本周精选推荐",
	})
	if err != nil {
		t.Fatalf("UpdateRecommendStatus returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected rpc request to be captured")
	}
	if captured.UpdateBy != 1001 {
		t.Fatalf("expected UpdateBy=1001, got %d", captured.UpdateBy)
	}
	if captured.ReviewMan != "tester" {
		t.Fatalf("expected ReviewMan=tester, got %q", captured.ReviewMan)
	}
	if captured.Detail != "加入本周精选推荐" {
		t.Fatalf("expected detail to pass through, got %q", captured.Detail)
	}
	if captured.Scope == nil || captured.Scope.ScopeType != "merchant" || captured.Scope.TenantId != 88 || captured.Scope.MerchantId != 3001 {
		t.Fatalf("unexpected scope passed to rpc: %+v", captured.Scope)
	}
}
