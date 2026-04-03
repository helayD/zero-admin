package cartitemservicelogic

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

func newOperateCartDryRunLogic(t *testing.T) *QueryOperateCartFunnelLogic {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open sqlite dry run db failed: %v", err)
	}

	return NewQueryOperateCartFunnelLogic(context.Background(), &svc.ServiceContext{DB: db})
}

func TestOperateFunnelCartQueryBuildsScopedSQL(t *testing.T) {
	logic := newOperateCartDryRunLogic(t)
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 88)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	startTime := time.Date(2026, 4, 3, 9, 0, 0, 0, operatefunnel.Location())
	endTime := startTime.Add(2 * time.Hour)
	rows := make([]operateCartBucketRow, 0)
	stmt := logic.buildOperateCartBucketQuery(scope, &omsclient.QueryOperateCartFunnelReq{
		Channel:      operatefunnel.ChannelApp,
		ActivityType: operatefunnel.ActivityHomeAdvertise,
		ActivityId:   501,
	}, startTime, endTime, operatefunnel.BucketHour).
		Group("bucket_start").
		Order("bucket_start ASC").
		Scan(&rows).Statement

	sqlText := stmt.SQL.String()
	if !strings.Contains(sqlText, "JOIN pms_product_spu spu ON spu.id = c.product_id") {
		t.Fatalf("expected product scope join in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "spu.platform_id = ? AND spu.tenant_id = ? AND spu.merchant_id = ?") {
		t.Fatalf("expected explicit merchant scope filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, operateCartChannelCaseSQL()+" = ?") {
		t.Fatalf("expected unified channel filter in SQL, got %s", sqlText)
	}
	if !strings.Contains(sqlText, "c.activity_type = ?") || !strings.Contains(sqlText, "c.activity_id = ?") {
		t.Fatalf("expected activity filter in SQL, got %s", sqlText)
	}
}
