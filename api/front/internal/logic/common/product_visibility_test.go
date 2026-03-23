package common

import (
	"strings"
	"testing"

	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

func TestBuildFrontProductVisibility(t *testing.T) {
	tests := []struct {
		name            string
		product         *pmsclient.ProductSpuListData
		wantVisible     bool
		wantPurchasable bool
		wantStatus      string
		wantReasonCode  string
		wantMessage     string
	}{
		{
			name: "visible product",
			product: &pmsclient.ProductSpuListData{
				Id:            1,
				VerifyStatus:  1,
				PublishStatus: 1,
				PreviewStatus: 0,
				Stock:         20,
			},
			wantVisible:     true,
			wantPurchasable: true,
			wantStatus:      FrontProductVisibilityStatusVisible,
		},
		{
			name: "pending review",
			product: &pmsclient.ProductSpuListData{
				Id:            2,
				VerifyStatus:  0,
				PublishStatus: 1,
				PreviewStatus: 0,
			},
			wantVisible:    false,
			wantStatus:     FrontProductVisibilityStatusHidden,
			wantReasonCode: FrontProductVisibilityReasonPendingReview,
			wantMessage:    "暂未开放浏览",
		},
		{
			name: "rejected",
			product: &pmsclient.ProductSpuListData{
				Id:            3,
				VerifyStatus:  2,
				PublishStatus: 1,
				PreviewStatus: 0,
			},
			wantVisible:    false,
			wantStatus:     FrontProductVisibilityStatusHidden,
			wantReasonCode: FrontProductVisibilityReasonRejected,
			wantMessage:    "审核未通过",
		},
		{
			name: "off shelf",
			product: &pmsclient.ProductSpuListData{
				Id:            4,
				VerifyStatus:  1,
				PublishStatus: 0,
				PreviewStatus: 0,
			},
			wantVisible:    false,
			wantStatus:     FrontProductVisibilityStatusHidden,
			wantReasonCode: FrontProductVisibilityReasonOffShelf,
			wantMessage:    "已下架",
		},
		{
			name: "preview product",
			product: &pmsclient.ProductSpuListData{
				Id:            5,
				VerifyStatus:  1,
				PublishStatus: 1,
				PreviewStatus: 1,
			},
			wantVisible:    false,
			wantStatus:     FrontProductVisibilityStatusHidden,
			wantReasonCode: FrontProductVisibilityReasonPreviewOnly,
			wantMessage:    "暂未开放浏览",
		},
		{
			name: "sold out",
			product: &pmsclient.ProductSpuListData{
				Id:            6,
				VerifyStatus:  1,
				PublishStatus: 1,
				PreviewStatus: 0,
				Stock:         0,
			},
			wantVisible:     true,
			wantPurchasable: false,
			wantStatus:      FrontProductVisibilityStatusNotSaleable,
			wantReasonCode:  FrontProductVisibilityReasonSoldOut,
			wantMessage:     "暂时缺货",
		},
		{
			name:           "nil product",
			product:        nil,
			wantVisible:    false,
			wantStatus:     FrontProductVisibilityStatusHidden,
			wantReasonCode: FrontProductVisibilityReasonNotFound,
			wantMessage:    "商品不存在或当前作用域不可见",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildFrontProductVisibility(tc.product)
			if result.Visible != tc.wantVisible {
				t.Fatalf("expected visible=%v, got %v", tc.wantVisible, result.Visible)
			}
			if result.Purchasable != tc.wantPurchasable {
				t.Fatalf("expected purchasable=%v, got %v", tc.wantPurchasable, result.Purchasable)
			}
			if result.Status != tc.wantStatus {
				t.Fatalf("expected status=%q, got %q", tc.wantStatus, result.Status)
			}
			if result.ReasonCode != tc.wantReasonCode {
				t.Fatalf("expected reasonCode=%q, got %q", tc.wantReasonCode, result.ReasonCode)
			}
			if tc.wantMessage != "" && !strings.Contains(result.ReasonMessage, tc.wantMessage) {
				t.Fatalf("expected message containing %q, got %q", tc.wantMessage, result.ReasonMessage)
			}
		})
	}
}

func TestEnsureFrontProductVisible(t *testing.T) {
	visibleProduct := &pmsclient.ProductSpuListData{
		Id:            1,
		VerifyStatus:  1,
		PublishStatus: 1,
		PreviewStatus: 0,
		Stock:         0,
	}

	if err := EnsureFrontProductVisible(visibleProduct); err != nil {
		t.Fatalf("sold out product should stay visible, got %v", err)
	}

	invisibleProduct := &pmsclient.ProductSpuListData{
		Id:            2,
		VerifyStatus:  2,
		PublishStatus: 1,
		PreviewStatus: 0,
	}

	err := EnsureFrontProductVisible(invisibleProduct)
	if err == nil || !strings.Contains(err.Error(), "审核未通过") {
		t.Fatalf("expected hidden product error, got %v", err)
	}
}
