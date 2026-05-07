package digitalcardmint

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAssetServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			compliance_rule_summary TEXT NOT NULL DEFAULT '',
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_name TEXT NOT NULL DEFAULT '',
			card_face_image TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			display_status INTEGER NOT NULL DEFAULT 1,
			content_audit_status INTEGER NOT NULL DEFAULT 0,
			audit_status INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_participation_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			result_type TEXT NOT NULL DEFAULT '',
			result_status TEXT NOT NULL DEFAULT '',
			failure_reason TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_instance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			scope TEXT NOT NULL DEFAULT '',
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			rarity TEXT NOT NULL DEFAULT '',
			asset_no TEXT NOT NULL DEFAULT '',
			asset_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			source_type TEXT NOT NULL DEFAULT 'draw',
			source_id INTEGER NOT NULL DEFAULT 0,
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			transferable INTEGER NOT NULL DEFAULT 0,
			transfer_limit INTEGER NOT NULL DEFAULT 0,
			claim_condition TEXT NOT NULL DEFAULT '',
			redemption_condition TEXT NOT NULL DEFAULT '',
			refund_policy TEXT NOT NULL DEFAULT '',
			display_status TEXT NOT NULL DEFAULT '',
			compliance_status TEXT NOT NULL DEFAULT '',
			display_reason TEXT NOT NULL DEFAULT '',
			compliance_reason TEXT NOT NULL DEFAULT '',
			rule_snapshot_json TEXT NOT NULL DEFAULT '',
			mint_task_id INTEGER NOT NULL DEFAULT 0,
			last_receipt_at DATETIME NULL,
			issued_at DATETIME NULL,
			disposed_at DATETIME NULL,
			disposed_by INTEGER NOT NULL DEFAULT 0,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by INTEGER NULL,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_mint_task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			asset_instance_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL DEFAULT 0,
			activity_id INTEGER NOT NULL DEFAULT 0,
			member_id INTEGER NOT NULL DEFAULT 0,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			idempotency_key TEXT NOT NULL DEFAULT '',
			task_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_tx_id TEXT NOT NULL DEFAULT '',
			retry_count INTEGER NOT NULL DEFAULT 0,
			max_retry_count INTEGER NOT NULL DEFAULT 3,
			last_error_code TEXT NOT NULL DEFAULT '',
			last_error_reason TEXT NOT NULL DEFAULT '',
			last_receipt_summary TEXT NOT NULL DEFAULT '',
			last_receipt_json TEXT NOT NULL DEFAULT '',
			last_execute_at DATETIME NULL,
			next_retry_at DATETIME NULL,
			manual_required INTEGER NOT NULL DEFAULT 0,
			frozen INTEGER NOT NULL DEFAULT 0,
			freeze_reason TEXT NOT NULL DEFAULT '',
			create_by INTEGER NOT NULL DEFAULT 0,
			update_by INTEGER NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_asset_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_instance_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL DEFAULT 0,
			from_status TEXT NOT NULL DEFAULT '',
			to_status TEXT NOT NULL DEFAULT '',
			operation_type TEXT NOT NULL DEFAULT '',
			operator_type TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			reason_code TEXT NOT NULL DEFAULT '',
			reason_text TEXT NOT NULL DEFAULT '',
			payload_json TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE sms_product_fulfillment_rule (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			rule_name TEXT NOT NULL DEFAULT '',
			rule_status INTEGER NOT NULL DEFAULT 1,
			card_template_id INTEGER NOT NULL DEFAULT 0,
			expire_days INTEGER NOT NULL DEFAULT 0,
			transferable INTEGER NOT NULL DEFAULT 0,
			transfer_limit INTEGER NOT NULL DEFAULT 0,
			claim_condition TEXT NOT NULL DEFAULT '',
			redemption_condition TEXT NOT NULL DEFAULT '',
			refund_policy TEXT NOT NULL DEFAULT '',
			create_by INTEGER NOT NULL DEFAULT 0,
			update_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	seeds := []string{
		`INSERT INTO sms_draw_activity (id, name, compliance_rule_summary, platform_id, tenant_id, merchant_id, is_deleted) VALUES
			(2001, '春季抽卡', '默认禁止集中竞价、连续挂牌和收益承诺', 1, 10, 88, 0)`,
		`INSERT INTO sms_card_template (id, template_name, card_face_image, status, is_deleted) VALUES
			(21, 'SSR 兔兔', 'https://img.example.com/ssr-rabbit.png', 1, 0)`,
		`INSERT INTO sms_draw_participation_record (id, activity_id, member_id, request_id, trace_id, result_type, result_status, failure_reason, create_time, is_deleted) VALUES
			(1, 2001, 3001, 'req-asset-1', 'trace-asset-1', 'won', 'won_pending_asset', '', '2026-04-18 10:00:00', 0),
			(2, 2001, 3001, 'req-asset-2', 'trace-asset-2', 'won', 'won_pending_asset', '', '2026-04-18 10:05:00', 0),
			(3, 2001, 3999, 'req-asset-3', 'trace-asset-3', 'won', 'won_pending_asset', '', '2026-04-18 10:10:00', 0)`,
		`INSERT INTO sms_card_instance (
			id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id,
			template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status,
			source_type, source_id, fulfillment_rule_id, transferable, transfer_limit, claim_condition, redemption_condition, refund_policy,
			display_status, compliance_status, display_reason, compliance_reason, rule_snapshot_json,
			mint_task_id, issued_at, create_time, update_time, is_deleted
		) VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-asset-1', 'trace-asset-1', 21, 'SSR', 'CARD-001', 'asset_created', 'mint_processing', '', 'processing', 'draw', 1, 0, 0, 0, '', '', '', 'display_visible', 'compliance_clear', '', '', '', 11, '2026-04-18 10:00:00', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(2, 1, 10, 88, 2001, 3001, 2, 'req-asset-2', 'trace-asset-2', 21, 'SSR', 'CARD-002', 'asset_created', 'mint_success', 'token-1234567890', 'success', 'draw', 2, 0, 0, 0, '', '', '', 'display_hidden', 'compliance_review', '合规复核中', '合规复核中', '{\"scene\":\"review\"}', 12, '2026-04-18 10:05:00', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0),
			(3, 1, 10, 88, 2001, 3999, 3, 'req-asset-3', 'trace-asset-3', 21, 'SSR', 'CARD-003', 'asset_created', 'mint_success', 'token-abcdef', 'success', 'draw', 3, 0, 0, 0, '', '', '', 'display_visible', 'compliance_clear', '', '', '', 13, '2026-04-18 10:10:00', '2026-04-18 10:10:00', '2026-04-18 10:11:00', 0)`,
		`INSERT INTO sms_card_mint_task (
			id, platform_id, tenant_id, merchant_id, asset_instance_id, participation_record_id, activity_id, member_id,
			request_id, trace_id, task_status, mint_status, chain_status, token_id, chain_tx_id,
			last_receipt_summary, last_receipt_json, create_time, update_time, is_deleted
		) VALUES
			(11, 1, 10, 88, 1, 1, 2001, 3001, 'req-asset-1', 'trace-asset-1', 'running', 'mint_processing', 'processing', '', '', '链上发放处理中', '{\"status\":\"processing\"}', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(12, 1, 10, 88, 2, 2, 2001, 3001, 'req-asset-2', 'trace-asset-2', 'succeeded', 'mint_success', 'success', 'token-1234567890', 'tx-asset-2', '链上铸造成功', '{\"status\":\"success\"}', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0),
			(13, 1, 10, 88, 3, 3, 2001, 3999, 'req-asset-3', 'trace-asset-3', 'succeeded', 'mint_success', 'success', 'token-abcdef', 'tx-asset-3', '链上铸造成功', '{\"status\":\"success\"}', '2026-04-18 10:10:00', '2026-04-18 10:11:00', 0)`,
		`INSERT INTO sms_card_asset_log (
			id, asset_instance_id, participation_record_id, from_status, to_status, operation_type, operator_type, trace_id, reason_code, reason_text, payload_json, create_time
		) VALUES
			(1, 1, 1, '', 'mint_processing', 'mint_dispatching', 'system', 'trace-asset-1', '', '已进入链上处理中', '{}', '2026-04-18 10:01:00'),
			(2, 2, 2, 'mint_processing', 'mint_success', 'mint_succeeded', 'system', 'trace-asset-2', '', '链上铸造成功', '{}', '2026-04-18 10:06:00'),
			(3, 2, 2, 'compliance_clear', 'compliance_review', 'asset_compliance_review', 'manual', 'trace-asset-2', '', '合规复核中', '{}', '2026-04-18 10:07:00')`,
		`INSERT INTO sms_product_fulfillment_rule (
			id, platform_id, tenant_id, merchant_id, rule_name, rule_status, card_template_id, expire_days,
			transferable, transfer_limit, claim_condition, redemption_condition, refund_policy,
			create_by, update_by, create_time, update_time, is_deleted
		) VALUES
			(1, 1, 10, 88, '数字卡包发卡规则', 1, 21, 365, 1, 3, '', '', 'freeze_card', 0, 0, '2026-05-06 00:00:00', NULL, 0)`,
		`INSERT INTO sms_card_instance (
			id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id,
			template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status,
			source_type, source_id, fulfillment_rule_id, transferable, transfer_limit, claim_condition, redemption_condition, refund_policy,
			display_status, compliance_status, display_reason, compliance_reason, rule_snapshot_json,
			mint_task_id, issued_at, create_time, update_time, is_deleted
		) VALUES
			(4, 1, 10, 88, 0, 3001, 0, 'req-purchase-1', 'trace-purchase-1', 21, '', 'CARD-PURCHASE-001', 'asset_created', 'mint_pending', '', '', 'purchase', 1001, 1, 1, 3, '', '', 'freeze_card', 'display_visible', 'compliance_clear', '', '', '', 0, '2026-05-07 10:00:00', '2026-05-07 10:00:00', NULL, 0)`,
	}
	for _, seed := range seeds {
		if err := db.Exec(seed).Error; err != nil {
			t.Fatalf("seed data failed: %v", err)
		}
	}

	return db
}

func merchantScope() pkgscope.GovernanceScope {
	return pkgscope.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformID: 1,
		TenantID:   10,
		MerchantID: 88,
	}
}

func TestQueryMemberDigitalCardAssetListKeepsRestrictedAssetVisible(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	total, list, err := service.QueryMemberDigitalCardAssetList(context.Background(), merchantScope(), 3001, MemberDigitalCardAssetFilter{
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("QueryMemberDigitalCardAssetList returned error: %v", err)
	}
	// 3 个资产：2 个抽卡 + 1 个订单购买
	if total != 3 || len(list) != 3 {
		t.Fatalf("expected three assets for member, got total=%d len=%d", total, len(list))
	}

	var restricted *MemberDigitalCardAssetItem
	for index := range list {
		if list[index].AssetNo == "CARD-002" {
			restricted = &list[index]
			break
		}
	}
	if restricted == nil {
		t.Fatalf("expected restricted asset CARD-002 in %+v", list)
	}
	if restricted.DisplayStatus != DisplayStatusHidden || restricted.ComplianceStatus != ComplianceStatusReview {
		t.Fatalf("expected hidden/review asset, got %+v", restricted)
	}
	if restricted.ComplianceRuleSummary != "合规复核中" {
		t.Fatalf("expected restricted reason to win over activity summary, got %q", restricted.ComplianceRuleSummary)
	}
}

func TestQueryMemberDigitalCardAssetDetailReturnsTimelineWithoutToken(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	detail, err := service.QueryMemberDigitalCardAssetDetail(context.Background(), merchantScope(), 3001, 2)
	if err != nil {
		t.Fatalf("QueryMemberDigitalCardAssetDetail returned error: %v", err)
	}
	if len(detail.Timeline) == 0 {
		t.Fatalf("expected timeline to be populated")
	}
	if detail.RestrictionReason != "合规复核中" {
		t.Fatalf("expected restriction reason to be preserved, got %q", detail.RestrictionReason)
	}
}

func TestRecycleDigitalCardAssetRequiresRestrictedContext(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	_, err := service.RecycleDigitalCardAsset(context.Background(), merchantScope(), 1, 9001, "直接回收")
	if err == nil || !strings.Contains(err.Error(), "回收处置前必须已进入受限或复核上下文") {
		t.Fatalf("expected restricted-context error, got %v", err)
	}
}

func TestReviewThenRecycleDigitalCardAssetUpdatesState(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	reviewResp, err := service.ReviewDigitalCardAssetCompliance(context.Background(), merchantScope(), 1, 9001, "存在投诉需复核")
	if err != nil {
		t.Fatalf("ReviewDigitalCardAssetCompliance returned error: %v", err)
	}
	if reviewResp.DisplayStatus != DisplayStatusHidden || reviewResp.ComplianceStatus != ComplianceStatusReview {
		t.Fatalf("unexpected review result: %+v", reviewResp)
	}

	recycleResp, err := service.RecycleDigitalCardAsset(context.Background(), merchantScope(), 1, 9001, "复核后回收")
	if err != nil {
		t.Fatalf("RecycleDigitalCardAsset returned error: %v", err)
	}
	if recycleResp.DisplayStatus != DisplayStatusRecycled || recycleResp.ComplianceStatus != ComplianceStatusRecycled {
		t.Fatalf("unexpected recycle result: %+v", recycleResp)
	}

	detail, err := service.QueryDigitalCardAssetAuditDetail(context.Background(), merchantScope(), 1)
	if err != nil {
		t.Fatalf("QueryDigitalCardAssetAuditDetail returned error: %v", err)
	}
	if detail.Item.DisplayStatus != DisplayStatusRecycled || detail.Item.ComplianceStatus != ComplianceStatusRecycled {
		t.Fatalf("expected recycled detail, got %+v", detail.Item)
	}
	if len(detail.Logs) < 2 {
		t.Fatalf("expected appended asset logs, got %+v", detail.Logs)
	}
	if detail.Logs[0].OperationType != OperationAssetRecycled {
		t.Fatalf("expected latest log to be recycled, got %+v", detail.Logs[0])
	}
}

func TestEnsureOrderPurchaseAssetCreatesIdempotentAsset(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	input := EnsureOrderPurchaseAssetInput{
		OrderID:           1001,
		OrderItemID:       2001,
		ProductID:         3001,
		SkuID:             4001,
		MemberID:          3001,
		FulfillmentRuleID: 1,
		PlatformID:        1,
		TenantID:          10,
		MerchantID:        88,
		RequestID:         "req-test-purchase-1",
		TraceID:           "trace-test-purchase-1",
		OperatorType:      "system",
	}

	// 首次创建
	result1, err := service.EnsureOrderPurchaseAsset(context.Background(), input)
	if err != nil {
		t.Fatalf("EnsureOrderPurchaseAsset first call returned error: %v", err)
	}
	if result1.AssetInstanceID <= 0 {
		t.Fatalf("expected positive asset instance ID, got %d", result1.AssetInstanceID)
	}
	if result1.AssetNo == "" {
		t.Fatalf("expected non-empty asset no")
	}
	if result1.MintStatus != MintStatusPending {
		t.Fatalf("expected mint status pending, got %s", result1.MintStatus)
	}

	// 幂等性验证：再次调用应返回相同结果
	result2, err := service.EnsureOrderPurchaseAsset(context.Background(), input)
	if err != nil {
		t.Fatalf("EnsureOrderPurchaseAsset second call returned error: %v", err)
	}
	if result2.AssetInstanceID != result1.AssetInstanceID {
		t.Fatalf("expected same asset instance ID for idempotent call, got %d vs %d", result2.AssetInstanceID, result1.AssetInstanceID)
	}
}

func TestHandleRefundCardDisposeFreezeCard(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	// 先创建订单购买型资产
	input := EnsureOrderPurchaseAssetInput{
		OrderID:           1002,
		OrderItemID:       2002,
		ProductID:         3002,
		SkuID:             4002,
		MemberID:          3001,
		FulfillmentRuleID: 1,
		PlatformID:        1,
		TenantID:          10,
		MerchantID:        88,
		RequestID:         "req-test-purchase-2",
		TraceID:           "trace-test-purchase-2",
		OperatorType:      "system",
	}

	result, err := service.EnsureOrderPurchaseAsset(context.Background(), input)
	if err != nil {
		t.Fatalf("EnsureOrderPurchaseAsset returned error: %v", err)
	}

	// 测试冻结处置
	disposeResult, err := service.HandleRefundCardDispose(context.Background(), RefundCardDisposeInput{
		OrderID:      1002,
		OrderItemID:  2002,
		RefundPolicy: "freeze_card",
		OperatorID:   9001,
		Reason:       "用户申请退款",
		TraceID:      "trace-refund-1",
	})
	if err != nil {
		t.Fatalf("HandleRefundCardDispose returned error: %v", err)
	}
	if disposeResult.AssetInstanceID != result.AssetInstanceID {
		t.Fatalf("expected same asset instance ID, got %d vs %d", disposeResult.AssetInstanceID, result.AssetInstanceID)
	}
	if disposeResult.ComplianceStatus != ComplianceStatusFrozen {
		t.Fatalf("expected compliance status frozen, got %s", disposeResult.ComplianceStatus)
	}
	if disposeResult.RefundAction != "freeze_card" {
		t.Fatalf("expected action freeze_card, got %s", disposeResult.RefundAction)
	}

	// 幂等性验证：再次处置应返回 already_disposed
	disposeResult2, err := service.HandleRefundCardDispose(context.Background(), RefundCardDisposeInput{
		OrderID:      1002,
		OrderItemID:  2002,
		RefundPolicy: "freeze_card",
		OperatorID:   9001,
		Reason:       "用户申请退款",
		TraceID:      "trace-refund-2",
	})
	if err != nil {
		t.Fatalf("HandleRefundCardDispose second call returned error: %v", err)
	}
	if disposeResult2.RefundAction != "already_disposed" {
		t.Fatalf("expected action already_disposed, got %s", disposeResult2.RefundAction)
	}
}

func TestQueryTaskListIncludesSourceType(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	// 先创建订单购买型资产
	input := EnsureOrderPurchaseAssetInput{
		OrderID:           1003,
		OrderItemID:       2003,
		ProductID:         3003,
		SkuID:             4003,
		MemberID:          3001,
		FulfillmentRuleID: 1,
		PlatformID:        1,
		TenantID:          10,
		MerchantID:        88,
		RequestID:         "req-test-purchase-3",
		TraceID:           "trace-test-purchase-3",
		OperatorType:      "system",
	}

	_, err := service.EnsureOrderPurchaseAsset(context.Background(), input)
	if err != nil {
		t.Fatalf("EnsureOrderPurchaseAsset returned error: %v", err)
	}

	// 查询任务列表
	_, tasks, err := service.QueryTaskList(context.Background(), merchantScope(), QueryFilter{
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("QueryTaskList returned error: %v", err)
	}

	// 查找订单购买型任务
	found := false
	for _, task := range tasks {
		if task.SourceType == "purchase" {
			found = true
			if task.SourceDisplayName != "订单购买" {
				t.Fatalf("expected source display name '订单购买', got '%s'", task.SourceDisplayName)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected to find purchase type task in list")
	}
}
