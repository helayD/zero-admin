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
			template_id INTEGER NOT NULL DEFAULT 0,
			rarity TEXT NOT NULL DEFAULT '',
			asset_no TEXT NOT NULL DEFAULT '',
			asset_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			display_status TEXT NOT NULL DEFAULT '',
			compliance_status TEXT NOT NULL DEFAULT '',
			display_reason TEXT NOT NULL DEFAULT '',
			compliance_reason TEXT NOT NULL DEFAULT '',
			rule_snapshot_json TEXT NOT NULL DEFAULT '',
			mint_task_id INTEGER NOT NULL DEFAULT 0,
			issued_at DATETIME NULL,
			disposed_at DATETIME NULL,
			disposed_by INTEGER NOT NULL DEFAULT 0,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by INTEGER NOT NULL DEFAULT 0,
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
			task_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_tx_id TEXT NOT NULL DEFAULT '',
			last_receipt_summary TEXT NOT NULL DEFAULT '',
			last_receipt_json TEXT NOT NULL DEFAULT '',
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
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	seeds := []string{
		`INSERT INTO sms_draw_activity (id, name, compliance_rule_summary, platform_id, tenant_id, merchant_id, is_deleted) VALUES
			(2001, '春季抽卡', '默认禁止集中竞价、连续挂牌和收益承诺', 1, 10, 88, 0)`,
		`INSERT INTO sms_card_template (id, template_name, card_face_image, is_deleted) VALUES
			(21, 'SSR 兔兔', 'https://img.example.com/ssr-rabbit.png', 0)`,
		`INSERT INTO sms_draw_participation_record (id, activity_id, member_id, request_id, trace_id, result_type, result_status, failure_reason, create_time, is_deleted) VALUES
			(1, 2001, 3001, 'req-asset-1', 'trace-asset-1', 'won', 'won_pending_asset', '', '2026-04-18 10:00:00', 0),
			(2, 2001, 3001, 'req-asset-2', 'trace-asset-2', 'won', 'won_pending_asset', '', '2026-04-18 10:05:00', 0),
			(3, 2001, 3999, 'req-asset-3', 'trace-asset-3', 'won', 'won_pending_asset', '', '2026-04-18 10:10:00', 0)`,
		`INSERT INTO sms_card_instance (
			id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id,
			template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status,
			display_status, compliance_status, display_reason, compliance_reason, rule_snapshot_json,
			mint_task_id, issued_at, create_time, update_time, is_deleted
		) VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-asset-1', 'trace-asset-1', 21, 'SSR', 'CARD-001', 'asset_created', 'mint_processing', '', 'processing', 'display_visible', 'compliance_clear', '', '', '', 11, '2026-04-18 10:00:00', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(2, 1, 10, 88, 2001, 3001, 2, 'req-asset-2', 'trace-asset-2', 21, 'SSR', 'CARD-002', 'asset_created', 'mint_success', 'token-1234567890', 'success', 'display_hidden', 'compliance_review', '合规复核中', '合规复核中', '{\"scene\":\"review\"}', 12, '2026-04-18 10:05:00', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0),
			(3, 1, 10, 88, 2001, 3999, 3, 'req-asset-3', 'trace-asset-3', 21, 'SSR', 'CARD-003', 'asset_created', 'mint_success', 'token-abcdef', 'success', 'display_visible', 'compliance_clear', '', '', '', 13, '2026-04-18 10:10:00', '2026-04-18 10:10:00', '2026-04-18 10:11:00', 0)`,
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
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected two assets for member, got total=%d len=%d", total, len(list))
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

func TestQueryMemberDigitalCardAssetDetailMasksTokenAndReturnsTimeline(t *testing.T) {
	service := NewService(newAssetServiceTestDB(t), nil, nil)

	detail, err := service.QueryMemberDigitalCardAssetDetail(context.Background(), merchantScope(), 3001, 2)
	if err != nil {
		t.Fatalf("QueryMemberDigitalCardAssetDetail returned error: %v", err)
	}
	if detail.TokenIDMasked == "" || strings.Contains(detail.TokenIDMasked, "1234567890") {
		t.Fatalf("expected token id to be masked, got %q", detail.TokenIDMasked)
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
