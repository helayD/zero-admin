package operatedashboardservicelogic

import (
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func TestOperateActivityOptionsQueryUsesFactTablesOnly(t *testing.T) {
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	queryText, queryArgs := buildOperateActivityOptionsQuery(scope)
	if len(queryArgs) != 10 {
		t.Fatalf("expected tenant scope args for all unions, got %#v", queryArgs)
	}
	if !strings.Contains(queryText, "FROM sms_operate_funnel_event e") {
		t.Fatalf("expected funnel event source in query, got %s", queryText)
	}
	if !strings.Contains(queryText, "FROM oms_cart_item c") {
		t.Fatalf("expected cart fact source in query, got %s", queryText)
	}
	if !strings.Contains(queryText, "FROM oms_order_main o") {
		t.Fatalf("expected order fact source in query, got %s", queryText)
	}
	if !strings.Contains(queryText, "FROM sms_coupon_record cr") {
		t.Fatalf("expected coupon redeem fact source in query, got %s", queryText)
	}
	if !strings.Contains(queryText, "spu.platform_id = ? AND spu.tenant_id = ?") {
		t.Fatalf("expected tenant scope on cart join, got %s", queryText)
	}
	if strings.Contains(queryText, "spu.merchant_id = ?") {
		t.Fatalf("tenant scope should not pin merchant_id in cart query, got %s", queryText)
	}
	if strings.Contains(queryText, "FROM sms_coupon\n") {
		t.Fatalf("query should not enumerate all coupons directly, got %s", queryText)
	}
	if strings.Contains(queryText, "FROM sms_seckill_activity\n") {
		t.Fatalf("query should not enumerate all seckill activities directly, got %s", queryText)
	}
}
