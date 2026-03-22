package coupon

import (
	"testing"
	"time"

	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

func TestBuildCouponScopeIDsDeduplicatesProductAndCategory(t *testing.T) {
	ids := buildCouponScopeIDs([]types.CarItemtPromotionListData{
		{ProductId: 11, ProductCategoryId: 101},
		{ProductId: 11, ProductCategoryId: 101},
		{ProductId: 12, ProductCategoryId: 102},
	})

	if len(ids) != 4 {
		t.Fatalf("unexpected ids length: %d", len(ids))
	}
}

func TestCouponSubtotalByScopeType(t *testing.T) {
	cartItems := []types.CarItemtPromotionListData{
		{ProductId: 11, ProductCategoryId: 101, Price: 100, ReduceAmount: 10, Quantity: 2},
		{ProductId: 12, ProductCategoryId: 102, Price: 80, ReduceAmount: 0, Quantity: 1},
	}

	fullTotal := couponSubtotal(0, nil, cartItems)
	if fullTotal != 260 {
		t.Fatalf("unexpected full coupon total: %v", fullTotal)
	}

	categoryTotal := couponSubtotal(1, []*smsclient.CouponScopeListData{
		{ScopeId: 101},
	}, cartItems)
	if categoryTotal != 180 {
		t.Fatalf("unexpected category coupon total: %v", categoryTotal)
	}

	productTotal := couponSubtotal(2, []*smsclient.CouponScopeListData{
		{ScopeId: 12},
	}, cartItems)
	if productTotal != 80 {
		t.Fatalf("unexpected product coupon total: %v", productTotal)
	}
}

func TestCouponUsableNow(t *testing.T) {
	start := time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05")
	end := time.Now().Add(time.Hour).Format("2006-01-02 15:04:05")

	if !couponUsableNow(start, end) {
		t.Fatalf("expected coupon to be usable")
	}

	expiredStart := time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05")
	expiredEnd := time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05")
	if couponUsableNow(expiredStart, expiredEnd) {
		t.Fatalf("expected expired coupon to be unusable")
	}
}
