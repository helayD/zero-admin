package membermessageservicelogic

import (
	"testing"
	"time"
)

func TestResolveLegacyRecallIntentUsesRelatedOrderFallback(t *testing.T) {
	t.Parallel()

	intent := resolveLegacyRecallIntent(
		1,
		"订单提醒",
		"请尽快支付",
		"order",
		"",
		9527,
		101,
		time.Date(2026, 4, 5, 12, 30, 0, 0, time.Local),
	)

	if intent.TargetType != recallTargetOrderDetail {
		t.Fatalf("expected target type %s, got %s", recallTargetOrderDetail, intent.TargetType)
	}
	if intent.TargetID != 9527 {
		t.Fatalf("expected target id 9527, got %d", intent.TargetID)
	}
	if intent.FallbackType != recallTargetOrderList || intent.FallbackTab != 1 {
		t.Fatalf("expected order_list fallback tab=1, got %s/%d", intent.FallbackType, intent.FallbackTab)
	}
	if intent.Blocked {
		t.Fatalf("expected order detail recall to stay available, got blocked intent %+v", intent)
	}
	if intent.IntentID != "member_message:101" {
		t.Fatalf("expected generated intent id member_message:101, got %s", intent.IntentID)
	}
}

func TestResolveLegacyRecallIntentBuildsStructuredFallbackForActivity(t *testing.T) {
	t.Parallel()

	intent := resolveLegacyRecallIntent(
		4,
		"活动提醒",
		"春季活动进行中",
		"activity",
		"88",
		0,
		102,
		time.Date(2026, 4, 5, 12, 45, 0, 0, time.Local),
	)

	if !intent.Blocked {
		t.Fatalf("expected activity intent to be blocked until real detail page exists, got %+v", intent)
	}
	if intent.FailureReason != recallFailureUnsupportedTarget {
		t.Fatalf("expected failure reason %s, got %s", recallFailureUnsupportedTarget, intent.FailureReason)
	}
	if intent.FallbackType != recallTargetHome || intent.FallbackTab != 0 {
		t.Fatalf("expected home fallback, got %s/%d", intent.FallbackType, intent.FallbackTab)
	}
	if intent.TargetType != recallTargetActivity {
		t.Fatalf("expected original target type activity for telemetry, got %s", intent.TargetType)
	}
	if intent.RecoveryHint == "" {
		t.Fatal("expected recovery hint to be populated")
	}
}
