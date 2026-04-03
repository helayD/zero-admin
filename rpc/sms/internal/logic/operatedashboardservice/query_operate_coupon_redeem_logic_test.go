package operatedashboardservicelogic

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOperateCouponRedeemDryRunLogic(t *testing.T) *QueryOperateCouponRedeemLogic {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open sqlite dry run db failed: %v", err)
	}

	return NewQueryOperateCouponRedeemLogic(context.Background(), &svc.ServiceContext{DB: db})
}

func TestBaseCouponRedeemQueryUsesCouponFallbackActivityFilter(t *testing.T) {
	logic := newOperateCouponRedeemDryRunLogic(t)
	startTime := time.Date(2026, 4, 3, 11, 0, 0, 0, operatefunnel.Location())
	endTime := startTime.Add(2 * time.Hour)

	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	rows := make([]couponRedeemBucketRow, 0)
	stmt := logic.baseCouponRedeemQuery(scope, &smsclient.QueryOperateCouponRedeemReq{
		ActivityType: operatefunnel.ActivityCoupon,
		ActivityId:   501,
		StartTime:    operatefunnel.FormatDateTime(startTime),
		EndTime:      operatefunnel.FormatDateTime(endTime),
	}, startTime, endTime).
		Select(operateBucketExpr("cr.use_time", operatefunnel.BucketDay) + " AS bucket_start, COUNT(*) AS coupon_redeem").
		Group("bucket_start").
		Order("bucket_start ASC").
		Scan(&rows).Statement

	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, couponRedeemActivityTypeCaseSQL()+" = ?") {
		t.Fatalf("expected coupon fallback activity type filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, couponRedeemActivityIDCaseSQL()+" = ?") {
		t.Fatalf("expected coupon fallback activity id filter in SQL, got %s", sqlText)
	}

	if len(stmt.Vars) < 10 {
		t.Fatalf("expected activity vars in SQL vars, got %+v", stmt.Vars)
	}
	if stmt.Vars[len(stmt.Vars)-2] != operatefunnel.ActivityCoupon {
		t.Fatalf("expected activity type var %s, got %+v", operatefunnel.ActivityCoupon, stmt.Vars)
	}
	if stmt.Vars[len(stmt.Vars)-1] != int64(501) {
		t.Fatalf("expected activity id var 501, got %+v", stmt.Vars)
	}
}

func TestCouponTrackingStateUsesCouponFallbackActivityFilter(t *testing.T) {
	logic := newOperateCouponRedeemDryRunLogic(t)
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	value := sql.NullTime{}
	stmt := applyCouponRedeemFilters(
		logic.couponRedeemScopedQuery(scope).Select("MIN(cr.use_time)"),
		&smsclient.QueryOperateCouponRedeemReq{
			ActivityType: operatefunnel.ActivityCoupon,
			ActivityId:   501,
		},
	).Scan(&value).Statement

	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, couponRedeemActivityTypeCaseSQL()+" = ?") {
		t.Fatalf("expected coupon fallback activity type filter in tracking SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, couponRedeemActivityIDCaseSQL()+" = ?") {
		t.Fatalf("expected coupon fallback activity id filter in tracking SQL, got %s", sqlText)
	}
}
