package product

import (
	"testing"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
)

func TestBuildProductDetailResponseForHiddenProduct(t *testing.T) {
	visibility := frontcommon.BuildFrontProductVisibility(nil)

	resp := buildProductDetailResponse(nil, nil, visibility, map[int64]int64{})
	if resp.Code != productDetailCodeHidden {
		t.Fatalf("expected hidden code %d, got %d", productDetailCodeHidden, resp.Code)
	}
	if resp.Message != visibility.ReasonMessage {
		t.Fatalf("expected message %q, got %q", visibility.ReasonMessage, resp.Message)
	}
	if resp.Data.Visibility.ReasonCode != frontcommon.FrontProductVisibilityReasonNotFound {
		t.Fatalf("expected reason code %q, got %q", frontcommon.FrontProductVisibilityReasonNotFound, resp.Data.Visibility.ReasonCode)
	}
	if len(resp.Data.CouponList) != 0 || len(resp.Data.SkuList) != 0 {
		t.Fatalf("expected hidden product to omit business data")
	}
}

func TestBuildProductDetailResponseForNotSaleableProduct(t *testing.T) {
	visibility := frontcommon.ProductVisibilityResult{
		Visible:        true,
		Purchasable:    false,
		ShowPrice:      true,
		ShowStock:      true,
		Status:         frontcommon.FrontProductVisibilityStatusNotSaleable,
		ReasonCode:     frontcommon.FrontProductVisibilityReasonSoldOut,
		ReasonMessage:  "商品暂时缺货",
		RecoveryHint:   "可先浏览同类商品",
		FallbackAction: frontcommon.FrontProductFallbackBrowseSimilar,
		FallbackTarget: "product_list",
	}

	resp := buildUnavailableProductDetailData(visibility)
	if resp.Visibility.Status != frontcommon.FrontProductVisibilityStatusNotSaleable {
		t.Fatalf("expected status %q, got %q", frontcommon.FrontProductVisibilityStatusNotSaleable, resp.Visibility.Status)
	}
	if resp.Visibility.Purchasable {
		t.Fatalf("expected not purchasable visibility")
	}
	if !resp.Visibility.ShowPrice || !resp.Visibility.ShowStock {
		t.Fatalf("expected price/stock to remain visible for not-saleable product")
	}
}

func TestMapProductDetailQueryError(t *testing.T) {
	visibility, ok := mapProductDetailQueryError("rpc error: code = Unknown desc = 商品SPU不存在")
	if !ok {
		t.Fatalf("expected product not found error to be mapped")
	}
	if visibility.ReasonCode != frontcommon.FrontProductVisibilityReasonNotFound {
		t.Fatalf("expected not found reason, got %q", visibility.ReasonCode)
	}
}

func TestMapProductDetailQueryErrorForDisabledOwner(t *testing.T) {
	visibility, ok := mapProductDetailQueryError("rpc error: code = Unknown desc = 商品[88]商品所属商户已停用")
	if !ok {
		t.Fatalf("expected disabled owner error to be mapped")
	}
	if visibility.ReasonCode != frontcommon.FrontProductVisibilityReasonOwnerDisabled {
		t.Fatalf("expected owner disabled reason, got %q", visibility.ReasonCode)
	}
}
