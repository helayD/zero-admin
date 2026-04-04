package orderservicelogic

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepeatPurchaseTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "repeat-purchase.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	stmt := `CREATE TABLE oms_order_main (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform_id INTEGER NOT NULL DEFAULT 1,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		merchant_id INTEGER NOT NULL DEFAULT 0,
		user_id INTEGER NOT NULL,
		order_status INTEGER NOT NULL DEFAULT 1,
		pay_amount REAL NOT NULL DEFAULT 0,
		pay_time DATETIME NULL,
		source_type INTEGER NOT NULL DEFAULT 1,
		activity_type TEXT NOT NULL DEFAULT 'none',
		activity_id INTEGER NOT NULL DEFAULT 0,
		is_deleted INTEGER NOT NULL DEFAULT 0
	)`
	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("create oms_order_main failed: %v", err)
	}
	return &svc.ServiceContext{DB: db}
}

func seedRepeatPurchaseOrder(t *testing.T, db *gorm.DB, id, userID int64, scope pkgscope.GovernanceScope, status int32, payAmount float64, payTime time.Time, sourceType int32, activityType string, activityID int64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO oms_order_main (id, platform_id, tenant_id, merchant_id, user_id, order_status, pay_amount, pay_time, source_type, activity_type, activity_id, is_deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		id, scope.PlatformID, scope.TenantID, scope.MerchantID, userID, status, payAmount, payTime, sourceType, activityType, activityID,
	).Error; err != nil {
		t.Fatalf("seed repeat purchase order failed: %v", err)
	}
}

func TestQueryRepeatPurchaseAnalysisHonorsEffectiveOrderScopeAndCurrentFilters(t *testing.T) {
	svcCtx := newRepeatPurchaseTestSvc(t)
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	otherMerchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 99)
	otherTenantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 20, 66)
	loc := time.FixedZone("CST", 8*3600)

	seedRepeatPurchaseOrder(t, svcCtx.DB, 1, 1001, merchantScope, 2, 100, time.Date(2026, 1, 1, 10, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 2, 1001, merchantScope, 2, 200, time.Date(2026, 3, 1, 9, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 3, 1002, merchantScope, 2, 180, time.Date(2025, 8, 1, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 4, 1002, merchantScope, 2, 260, time.Date(2026, 3, 10, 9, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 5, 1003, merchantScope, 5, 120, time.Date(2026, 2, 15, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 6, 1003, merchantScope, 2, 220, time.Date(2026, 3, 15, 9, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 7, 1004, merchantScope, 6, 130, time.Date(2026, 2, 18, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 8, 1004, merchantScope, 2, 230, time.Date(2026, 3, 20, 9, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 9, 1005, otherMerchantScope, 2, 110, time.Date(2026, 2, 1, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 10, 1005, otherMerchantScope, 2, 210, time.Date(2026, 3, 5, 9, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 11, 1006, otherTenantScope, 2, 140, time.Date(2026, 2, 5, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 12, 1006, otherTenantScope, 2, 240, time.Date(2026, 3, 6, 9, 0, 0, 0, loc), 2, "coupon", 501)

	logic := NewQueryRepeatPurchaseAnalysisLogic(context.Background(), svcCtx)
	resp, err := logic.QueryRepeatPurchaseAnalysis(&omsclient.QueryRepeatPurchaseAnalysisReq{
		Scope: &omsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID},
		StartTime:    "2026-03-01 00:00:00",
		EndTime:      "2026-04-01 00:00:00",
		Channel:      "pc",
		ActivityType: "coupon",
		ActivityId:   501,
		Bucket:       "day",
	})
	if err != nil {
		t.Fatalf("QueryRepeatPurchaseAnalysis returned error: %v", err)
	}
	if resp.Overview.PaidBuyerCount != 4 {
		t.Fatalf("expected paidBuyerCount 4, got %d", resp.Overview.PaidBuyerCount)
	}
	if resp.Overview.RepeatBuyerCount != 1 {
		t.Fatalf("expected repeatBuyerCount 1, got %d", resp.Overview.RepeatBuyerCount)
	}
	if resp.Overview.RepeatOrderCount != 1 {
		t.Fatalf("expected repeatOrderCount 1, got %d", resp.Overview.RepeatOrderCount)
	}
	if resp.Overview.RepeatGmv != 200 {
		t.Fatalf("expected repeatGmv 200, got %v", resp.Overview.RepeatGmv)
	}
	if resp.Overview.RepeatRate != 0.25 {
		t.Fatalf("expected repeatRate 0.25, got %v", resp.Overview.RepeatRate)
	}
	if len(resp.Trends) == 0 {
		t.Fatalf("expected trend points, got none")
	}
}

func TestQueryRepeatPurchaseDetailListUsesStableSortAndAggregatesRepeatOrders(t *testing.T) {
	svcCtx := newRepeatPurchaseTestSvc(t)
	tenantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	merchant88, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	merchant99, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 99)
	loc := time.FixedZone("CST", 8*3600)

	seedRepeatPurchaseOrder(t, svcCtx.DB, 1, 1001, merchant88, 2, 100, time.Date(2026, 2, 1, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 2, 1001, merchant88, 2, 220, time.Date(2026, 3, 12, 10, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 3, 1007, merchant88, 2, 90, time.Date(2026, 2, 10, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 4, 1007, merchant88, 2, 180, time.Date(2026, 3, 12, 10, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 5, 1007, merchant88, 2, 260, time.Date(2026, 3, 25, 10, 0, 0, 0, loc), 2, "coupon", 501)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 6, 1008, merchant99, 2, 80, time.Date(2026, 2, 12, 9, 0, 0, 0, loc), 1, "none", 0)
	seedRepeatPurchaseOrder(t, svcCtx.DB, 7, 1008, merchant99, 2, 280, time.Date(2026, 3, 18, 10, 0, 0, 0, loc), 2, "coupon", 501)

	logic := NewQueryRepeatPurchaseDetailListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryRepeatPurchaseDetailList(&omsclient.QueryRepeatPurchaseDetailListReq{
		Scope: &omsclient.GovernanceScope{ScopeType: tenantScope.ScopeType, PlatformId: tenantScope.PlatformID, TenantId: tenantScope.TenantID, MerchantId: tenantScope.MerchantID},
		StartTime:    "2026-03-01 00:00:00",
		EndTime:      "2026-04-01 00:00:00",
		Channel:      "pc",
		ActivityType: "coupon",
		ActivityId:   501,
		PageNum:      1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatalf("QueryRepeatPurchaseDetailList returned error: %v", err)
	}
	if resp.Total != 3 {
		t.Fatalf("expected 3 repeat buyers in tenant scope, got %d", resp.Total)
	}
	if len(resp.List) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(resp.List))
	}
	if resp.List[0].MemberId != 1007 || resp.List[0].RepeatOrderCount != 2 || resp.List[0].LatestRepeatPayTime != "2026-03-25 10:00:00" {
		t.Fatalf("unexpected first detail row: %+v", resp.List[0])
	}
	if resp.List[1].LatestRepeatPayTime < resp.List[2].LatestRepeatPayTime {
		t.Fatalf("expected stable descending sort by latestRepeatPayTime, got %+v", resp.List)
	}
}
