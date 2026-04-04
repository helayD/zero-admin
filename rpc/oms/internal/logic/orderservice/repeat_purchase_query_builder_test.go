package orderservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepeatPurchaseDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open sqlite dry run db failed: %v", err)
	}
	return db
}

func TestBuildRepeatPurchaseQualifiedOrderQueryAppliesCurrentOrderFilters(t *testing.T) {
	db := newRepeatPurchaseDryRunDB(t)
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}
	filter := repeatPurchaseFilter{
		Scope:        scope,
		StartTime:    time.Date(2026, 4, 1, 0, 0, 0, 0, operatefunnel.Location()),
		EndTime:      time.Date(2026, 4, 8, 0, 0, 0, 0, operatefunnel.Location()),
		Channel:      operatefunnel.ChannelPC,
		ActivityType: operatefunnel.ActivityCoupon,
		ActivityID:   1001,
	}

	query, err := buildRepeatPurchaseQualifiedOrderQuery(context.Background(), db, filter)
	if err != nil {
		t.Fatalf("buildRepeatPurchaseQualifiedOrderQuery returned error: %v", err)
	}

	stmt := query.Find(&[]repeatPurchaseQualifiedOrderRow{}).Statement
	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, "curr.is_deleted = ?") {
		t.Fatalf("expected current order delete filter, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "curr.pay_time IS NOT NULL") {
		t.Fatalf("expected current order pay_time filter, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "curr.order_status IN (?,?,?,?)") {
		t.Fatalf("expected effective order status filter, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "curr.pay_time >= ? AND curr.pay_time < ?") {
		t.Fatalf("expected current order pay range filter, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "curr.platform_id = ? AND curr.tenant_id = ?") {
		t.Fatalf("expected tenant scope filter, got %s", sqlText)
	}
	if strings.Contains(sqlText, "curr.merchant_id = ?") {
		t.Fatalf("tenant scope should not pin merchant scope, got %s", sqlText)
	}
	if !strings.Contains(sqlText, repeatPurchaseOrderChannelCaseSQL("curr")+" = ?") {
		t.Fatalf("expected channel filter, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "curr.activity_type = ?") || !strings.Contains(sqlText, "curr.activity_id = ?") {
		t.Fatalf("expected activity filters, got %s", sqlText)
	}
}

func TestBuildRepeatPurchasePriorOrderExistsSQLHonorsLookbackAndScope(t *testing.T) {
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	sqlText, err := buildRepeatPurchasePriorOrderExistsSQL(scope, "curr", "prev")
	if err != nil {
		t.Fatalf("buildRepeatPurchasePriorOrderExistsSQL returned error: %v", err)
	}
	checks := []string{
		"prev.user_id = curr.user_id",
		"prev.is_deleted = 0",
		"prev.pay_time IS NOT NULL",
		"prev.order_status IN (2,3,4,7)",
		"prev.pay_time >= DATE_SUB(curr.pay_time, INTERVAL 180 DAY)",
		"prev.pay_time < curr.pay_time",
		"prev.platform_id = curr.platform_id",
		"prev.tenant_id = curr.tenant_id",
		"prev.merchant_id = curr.merchant_id",
	}
	for _, item := range checks {
		if !strings.Contains(sqlText, item) {
			t.Fatalf("expected SQL to contain %s, got %s", item, sqlText)
		}
	}
}

func TestBuildRepeatPurchaseQualifiedOrderQuerySupportsNoneActivityAndPlatformScope(t *testing.T) {
	db := newRepeatPurchaseDryRunDB(t)
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, 1, 0, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}
	filter := repeatPurchaseFilter{
		Scope:        scope,
		StartTime:    time.Date(2026, 4, 1, 0, 0, 0, 0, operatefunnel.Location()),
		EndTime:      time.Date(2026, 4, 8, 0, 0, 0, 0, operatefunnel.Location()),
		ActivityType: operatefunnel.ActivityNone,
	}

	query, err := buildRepeatPurchaseQualifiedOrderQuery(context.Background(), db, filter)
	if err != nil {
		t.Fatalf("buildRepeatPurchaseQualifiedOrderQuery returned error: %v", err)
	}
	stmt := query.Find(&[]repeatPurchaseQualifiedOrderRow{}).Statement
	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, "curr.platform_id = ?") {
		t.Fatalf("expected platform scope filter, got %s", sqlText)
	}
	if strings.Contains(sqlText, "curr.tenant_id = ?") || strings.Contains(sqlText, "curr.merchant_id = ?") {
		t.Fatalf("platform scope should not pin tenant/merchant, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "COALESCE(NULLIF(curr.activity_type, ''), 'none') = 'none'") {
		t.Fatalf("expected none activity filter, got %s", sqlText)
	}
}

func TestNewRepeatPurchaseQueryBuilderUsesSvcDB(t *testing.T) {
	db := newRepeatPurchaseDryRunDB(t)
	builder := newRepeatPurchaseQueryBuilder(context.Background(), &svc.ServiceContext{DB: db})
	if builder == nil || builder.db != db {
		t.Fatalf("expected query builder to retain svc db")
	}
}
