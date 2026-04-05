package orderservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOperateOrderDryRunLogic(t *testing.T) *QueryOperateOrderFunnelLogic {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open sqlite dry run db failed: %v", err)
	}

	return NewQueryOperateOrderFunnelLogic(context.Background(), &svc.ServiceContext{DB: db})
}

func TestOperateFunnelOrderQueryBuildsScopedSQL(t *testing.T) {
	logic := newOperateOrderDryRunLogic(t)
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	startTime := time.Date(2026, 4, 3, 0, 0, 0, 0, operatefunnel.Location())
	endTime := startTime.Add(24 * time.Hour)
	rows := make([]operateOrderBucketRow, 0)
	stmt := logic.buildOperateOrderBucketQuery("pay_time", scope, &omsclient.QueryOperateOrderFunnelReq{
		Channel:      operatefunnel.ChannelPC,
		ActivityType: operatefunnel.ActivityNone,
	}, startTime, endTime, operatefunnel.BucketDay).
		Group("bucket_start").
		Order("bucket_start ASC").
		Scan(&rows).Statement

	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, "o.platform_id = ? AND o.tenant_id = ?") {
		t.Fatalf("expected tenant scope filter in SQL, got %s", sqlText)
	}
	if strings.Contains(sqlText, "o.merchant_id = ?") {
		t.Fatalf("tenant scope should not pin merchant_id, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "o.order_status = 2") {
		t.Fatalf("expected pay success status filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, operateOrderChannelCaseSQL()+" = ?") {
		t.Fatalf("expected unified order channel filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "COALESCE(NULLIF(o.activity_type, ''), 'none') = 'none'") {
		t.Fatalf("expected none-activity filter in SQL, got %s", sqlText)
	}
}

func TestOperateFunnelBucketSwitchesBetweenHourAndDay(t *testing.T) {
	startTime := time.Date(2026, 4, 3, 0, 0, 0, 0, operatefunnel.Location())

	if got := operatefunnel.NormalizeBucket(operatefunnel.BucketHour, startTime, startTime.Add(12*time.Hour)); got != operatefunnel.BucketHour {
		t.Fatalf("expected hour bucket within 48h, got %s", got)
	}
	if got := operatefunnel.NormalizeBucket(operatefunnel.BucketHour, startTime, startTime.Add(49*time.Hour)); got != operatefunnel.BucketDay {
		t.Fatalf("expected day bucket beyond 48h, got %s", got)
	}
}
