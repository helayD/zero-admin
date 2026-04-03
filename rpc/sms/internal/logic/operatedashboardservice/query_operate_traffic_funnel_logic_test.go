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

func newOperateTrafficDryRunLogic(t *testing.T) *QueryOperateTrafficFunnelLogic {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open sqlite dry run db failed: %v", err)
	}

	return NewQueryOperateTrafficFunnelLogic(context.Background(), &svc.ServiceContext{DB: db})
}

func TestOperateFunnelTrafficQueryBuildsScopedSQL(t *testing.T) {
	logic := newOperateTrafficDryRunLogic(t)
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	startTime := time.Date(2026, 4, 3, 10, 0, 0, 0, operatefunnel.Location())
	endTime := startTime.Add(4 * time.Hour)
	rows := make([]trafficBucketRow, 0)
	stmt := logic.buildOperateTrafficBucketQuery(scope, &smsclient.QueryOperateTrafficFunnelReq{
		Channel:      operatefunnel.ChannelMiniProgram,
		ActivityType: operatefunnel.ActivityHomeAdvertise,
		ActivityId:   18,
	}, startTime, endTime, operatefunnel.BucketHour).
		Group("bucket_start").
		Order("bucket_start ASC").
		Scan(&rows).Statement

	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, "sms_operate_funnel_event e") {
		t.Fatalf("expected operate funnel event table in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "e.platform_id = ? AND e.tenant_id = ? AND e.merchant_id = ?") {
		t.Fatalf("expected explicit merchant scope filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "e.channel = ?") {
		t.Fatalf("expected channel filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "e.activity_type = ?") || !strings.Contains(sqlText, "e.activity_id = ?") {
		t.Fatalf("expected activity filter in SQL, got %s", sqlText)
	}
}

func TestOperateFunnelTrafficReturnsPartialMetricsWhenTrackingStartsLater(t *testing.T) {
	trackingStart := time.Date(2026, 4, 3, 12, 0, 0, 0, operatefunnel.Location())
	startTime := trackingStart.Add(-2 * time.Hour)
	partialMetrics := buildTrafficPartialMetrics(sql.NullTime{Time: trackingStart, Valid: true}, startTime)
	if len(partialMetrics) != 2 || partialMetrics[0] != operatefunnel.EventExposure || partialMetrics[1] != operatefunnel.EventClick {
		t.Fatalf("expected exposure/click partial metrics, got %+v", partialMetrics)
	}
}
