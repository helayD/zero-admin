package couponservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCouponScopeTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "coupon-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.SmsCoupon{}); err != nil {
		t.Fatalf("auto migrate coupon failed: %v", err)
	}
	for _, stmt := range []string{
		`ALTER TABLE sms_coupon ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE sms_coupon ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE sms_coupon ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter coupon scope columns failed: %v", err)
		}
	}

	return &svc.ServiceContext{DB: db}
}

func seedCoupon(t *testing.T, db *gorm.DB, item model.SmsCoupon, current pkgscope.GovernanceScope) {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed coupon failed: %v", err)
	}
	if err := db.Exec(
		`UPDATE sms_coupon SET platform_id = ?, tenant_id = ?, merchant_id = ? WHERE id = ?`,
		current.PlatformID,
		current.TenantID,
		current.MerchantID,
		item.ID,
	).Error; err != nil {
		t.Fatalf("update coupon scope failed: %v", err)
	}
}

func TestQueryCouponListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newCouponScopeTestSvc(t)
	now := time.Now()

	tenantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	if err != nil {
		t.Fatalf("normalize tenant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 20, 0)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedCoupon(t, svcCtx.DB, model.SmsCoupon{
		ID:            1,
		TypeID:        1,
		Name:          "租户优惠券",
		Code:          "TENANT-10",
		Amount:        30,
		MinAmount:     100,
		StartTime:     now,
		EndTime:       now.Add(24 * time.Hour),
		TotalCount:    100,
		ReceivedCount: 10,
		UsedCount:     2,
		PerLimit:      1,
		Status:        1,
		IsEnabled:     1,
		Description:   "only tenant 10",
		CreateBy:      1,
		CreateTime:    now,
		IsDeleted:     0,
	}, tenantScope)
	seedCoupon(t, svcCtx.DB, model.SmsCoupon{
		ID:            2,
		TypeID:        1,
		Name:          "其它租户优惠券",
		Code:          "TENANT-20",
		Amount:        40,
		MinAmount:     100,
		StartTime:     now,
		EndTime:       now.Add(24 * time.Hour),
		TotalCount:    100,
		ReceivedCount: 5,
		UsedCount:     1,
		PerLimit:      1,
		Status:        1,
		IsEnabled:     1,
		Description:   "only tenant 20",
		CreateBy:      1,
		CreateTime:    now,
		IsDeleted:     0,
	}, otherScope)

	logic := NewQueryCouponListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryCouponList(&smsclient.QueryCouponListReq{
		PageNum:   1,
		PageSize:  10,
		Status:    4,
		IsEnabled: 2,
		Scope: &smsclient.GovernanceScope{
			ScopeType:  tenantScope.ScopeType,
			PlatformId: tenantScope.PlatformID,
			TenantId:   tenantScope.TenantID,
			MerchantId: tenantScope.MerchantID,
		},
	})
	if err != nil {
		t.Fatalf("query coupon list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("expected one scoped coupon, got total=%d len=%d", resp.Total, len(resp.List))
	}
	if resp.List[0].Id != 1 {
		t.Fatalf("expected coupon 1, got %+v", resp.List[0])
	}
}

func TestQueryCouponDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newCouponScopeTestSvc(t)
	now := time.Now()

	merchantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize merchant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedCoupon(t, svcCtx.DB, model.SmsCoupon{
		ID:            11,
		TypeID:        1,
		Name:          "商户券",
		Code:          "M-301",
		Amount:        20,
		MinAmount:     100,
		StartTime:     now,
		EndTime:       now.Add(24 * time.Hour),
		TotalCount:    100,
		ReceivedCount: 0,
		UsedCount:     0,
		PerLimit:      1,
		Status:        1,
		IsEnabled:     1,
		Description:   "only merchant 301",
		CreateBy:      1,
		CreateTime:    now,
		IsDeleted:     0,
	}, merchantScope)

	logic := NewQueryCouponDetailLogic(context.Background(), svcCtx)
	_, err = logic.QueryCouponDetail(&smsclient.QueryCouponDetailReq{
		Id: 11,
		Scope: &smsclient.GovernanceScope{
			ScopeType:  otherScope.ScopeType,
			PlatformId: otherScope.PlatformID,
			TenantId:   otherScope.TenantID,
			MerchantId: otherScope.MerchantID,
		},
	})
	if err == nil {
		t.Fatal("expected cross-scope coupon detail to fail")
	}
	if !strings.Contains(err.Error(), "优惠券不存在") {
		t.Fatalf("expected not found error, got %v", err)
	}
}
