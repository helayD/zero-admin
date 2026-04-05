package operatefunnel

import (
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func TestRepeatPurchaseLookbackWindowUses180Days(t *testing.T) {
	payTime := time.Date(2026, 4, 4, 12, 30, 0, 0, Location())

	start, end, err := RepeatPurchaseLookbackWindow(payTime)
	if err != nil {
		t.Fatalf("RepeatPurchaseLookbackWindow returned error: %v", err)
	}
	if !end.Equal(payTime) {
		t.Fatalf("expected lookback end %s, got %s", payTime, end)
	}
	if want := payTime.AddDate(0, 0, -RepeatPurchaseWindowDays); !start.Equal(want) {
		t.Fatalf("expected lookback start %s, got %s", want, start)
	}
}

func TestIsRepeatPurchaseEffectiveOrder(t *testing.T) {
	payTime := time.Date(2026, 4, 4, 12, 0, 0, 0, Location())

	validStatuses := []int32{2, 3, 4, 7}
	for _, status := range validStatuses {
		if !IsRepeatPurchaseEffectiveOrder(0, payTime, status) {
			t.Fatalf("expected status %d to be treated as effective order", status)
		}
	}

	invalidStatuses := []int32{0, 1, 5, 6, 8}
	for _, status := range invalidStatuses {
		if IsRepeatPurchaseEffectiveOrder(0, payTime, status) {
			t.Fatalf("expected status %d to be excluded from effective orders", status)
		}
	}

	if IsRepeatPurchaseEffectiveOrder(1, payTime, 2) {
		t.Fatalf("deleted order should not be treated as effective")
	}
	if IsRepeatPurchaseEffectiveOrder(0, time.Time{}, 2) {
		t.Fatalf("zero pay time should not be treated as effective")
	}
}

func TestRepeatPurchaseScopeOrderColumns(t *testing.T) {
	testCases := []struct {
		name      string
		scope     pkgscope.GovernanceScope
		want      []string
		wantError bool
	}{
		{
			name: "platform",
			scope: pkgscope.GovernanceScope{
				ScopeType:  pkgscope.SubjectTypePlatform,
				PlatformID: 1,
			},
			want: []string{"platform_id"},
		},
		{
			name: "tenant",
			scope: pkgscope.GovernanceScope{
				ScopeType:  pkgscope.SubjectTypeTenant,
				PlatformID: 1,
				TenantID:   10,
			},
			want: []string{"platform_id", "tenant_id"},
		},
		{
			name: "merchant",
			scope: pkgscope.GovernanceScope{
				ScopeType:  pkgscope.SubjectTypeMerchant,
				PlatformID: 1,
				TenantID:   10,
				MerchantID: 88,
			},
			want: []string{"platform_id", "tenant_id", "merchant_id"},
		},
		{
			name: "merchant missing tenant",
			scope: pkgscope.GovernanceScope{
				ScopeType:  pkgscope.SubjectTypeMerchant,
				PlatformID: 1,
				MerchantID: 88,
			},
			wantError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := RepeatPurchaseScopeOrderColumns(testCase.scope)
			if testCase.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("RepeatPurchaseScopeOrderColumns returned error: %v", err)
			}
			if len(got) != len(testCase.want) {
				t.Fatalf("expected %d columns, got %d (%v)", len(testCase.want), len(got), got)
			}
			for index := range got {
				if got[index] != testCase.want[index] {
					t.Fatalf("expected columns %v, got %v", testCase.want, got)
				}
			}
		})
	}
}

func TestRepeatPurchaseOverviewDefinitionsAndDetailColumns(t *testing.T) {
	overview := RepeatPurchaseOverviewDefinitions()
	if len(overview) != 6 {
		t.Fatalf("expected 6 overview definitions, got %d", len(overview))
	}
	if overview[0].Key != RepeatPurchaseMetricPaidBuyerCount {
		t.Fatalf("expected first overview metric to be %s, got %s", RepeatPurchaseMetricPaidBuyerCount, overview[0].Key)
	}
	if overview[5].Key != RepeatPurchaseMetricAvgDaysToRepeat {
		t.Fatalf("expected last overview metric to be %s, got %s", RepeatPurchaseMetricAvgDaysToRepeat, overview[5].Key)
	}

	columns := RepeatPurchaseDetailColumns()
	wantKeys := []string{
		"memberId",
		"nicknameMasked",
		"mobileMasked",
		"firstValidPayTime",
		"latestRepeatPayTime",
		"repeatOrderCount",
		"repeatGmv",
		"latestChannel",
		"latestActivityType",
		"latestActivityId",
		"platformId",
		"tenantId",
		"merchantId",
	}
	if len(columns) != len(wantKeys) {
		t.Fatalf("expected %d detail columns, got %d", len(wantKeys), len(columns))
	}
	for index, wantKey := range wantKeys {
		if columns[index].Key != wantKey {
			t.Fatalf("expected detail column %d to be %s, got %s", index, wantKey, columns[index].Key)
		}
	}
	if got := RepeatPurchaseStableSortClause(); got != "latest_repeat_pay_time DESC, member_id ASC" {
		t.Fatalf("unexpected repeat purchase stable sort clause: %s", got)
	}
}

func TestRepeatPurchasePartialMetricsOnlyWhenActivityTrackingIsIncomplete(t *testing.T) {
	trackingStartedAt := time.Date(2026, 4, 1, 0, 0, 0, 0, Location())
	queryStart := time.Date(2026, 3, 28, 0, 0, 0, 0, Location())

	metrics := RepeatPurchasePartialMetrics(trackingStartedAt, queryStart, ActivityCoupon, 1001)
	if len(metrics) != 6 {
		t.Fatalf("expected all repeat purchase metrics to be partial, got %v", metrics)
	}
	if metrics[0] != RepeatPurchaseMetricPaidBuyerCount || metrics[5] != RepeatPurchaseMetricAvgDaysToRepeat {
		t.Fatalf("unexpected partial metrics: %v", metrics)
	}

	if got := RepeatPurchasePartialMetrics(trackingStartedAt, queryStart, "", 0); len(got) != 0 {
		t.Fatalf("expected no partial metrics without activity filter, got %v", got)
	}
	if got := RepeatPurchasePartialMetrics(trackingStartedAt, trackingStartedAt, ActivityCoupon, 1001); len(got) != 0 {
		t.Fatalf("expected no partial metrics when query starts at tracking start, got %v", got)
	}
}

func TestMaskRepeatPurchaseFields(t *testing.T) {
	if got := MaskRepeatPurchaseMobile("13812345678"); got != "138****5678" {
		t.Fatalf("expected masked mobile, got %s", got)
	}
	if got := MaskRepeatPurchaseNickname("九克城用户"); got != "九**户" {
		t.Fatalf("expected masked nickname, got %s", got)
	}
	if got := MaskRepeatPurchaseNickname("张三"); got != "张*" {
		t.Fatalf("expected two-rune nickname masked, got %s", got)
	}
}
