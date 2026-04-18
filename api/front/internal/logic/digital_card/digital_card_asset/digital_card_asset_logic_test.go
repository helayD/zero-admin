package digital_card_asset

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newFrontAssetTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_activity (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL DEFAULT '', compliance_rule_summary TEXT NOT NULL DEFAULT '', platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_template (id INTEGER PRIMARY KEY AUTOINCREMENT, template_name TEXT NOT NULL DEFAULT '', card_face_image TEXT NOT NULL DEFAULT '', is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_draw_participation_record (id INTEGER PRIMARY KEY AUTOINCREMENT, activity_id INTEGER NOT NULL, member_id INTEGER NOT NULL, request_id TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', result_type TEXT NOT NULL DEFAULT '', result_status TEXT NOT NULL DEFAULT '', failure_reason TEXT NOT NULL DEFAULT '', create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_instance (id INTEGER PRIMARY KEY AUTOINCREMENT, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0, activity_id INTEGER NOT NULL, member_id INTEGER NOT NULL, participation_record_id INTEGER NOT NULL, request_id TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', template_id INTEGER NOT NULL DEFAULT 0, rarity TEXT NOT NULL DEFAULT '', asset_no TEXT NOT NULL DEFAULT '', asset_status TEXT NOT NULL DEFAULT '', mint_status TEXT NOT NULL DEFAULT '', token_id TEXT NOT NULL DEFAULT '', chain_status TEXT NOT NULL DEFAULT '', display_status TEXT NOT NULL DEFAULT '', compliance_status TEXT NOT NULL DEFAULT '', display_reason TEXT NOT NULL DEFAULT '', compliance_reason TEXT NOT NULL DEFAULT '', rule_snapshot_json TEXT NOT NULL DEFAULT '', mint_task_id INTEGER NOT NULL DEFAULT 0, issued_at DATETIME NULL, disposed_at DATETIME NULL, disposed_by INTEGER NOT NULL DEFAULT 0, create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, update_time DATETIME NULL, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_mint_task (id INTEGER PRIMARY KEY AUTOINCREMENT, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0, asset_instance_id INTEGER NOT NULL, participation_record_id INTEGER NOT NULL DEFAULT 0, activity_id INTEGER NOT NULL DEFAULT 0, member_id INTEGER NOT NULL DEFAULT 0, request_id TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', task_status TEXT NOT NULL DEFAULT '', mint_status TEXT NOT NULL DEFAULT '', chain_status TEXT NOT NULL DEFAULT '', token_id TEXT NOT NULL DEFAULT '', chain_tx_id TEXT NOT NULL DEFAULT '', last_receipt_summary TEXT NOT NULL DEFAULT '', last_receipt_json TEXT NOT NULL DEFAULT '', create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, update_time DATETIME NULL, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_asset_log (id INTEGER PRIMARY KEY AUTOINCREMENT, asset_instance_id INTEGER NOT NULL, participation_record_id INTEGER NOT NULL DEFAULT 0, from_status TEXT NOT NULL DEFAULT '', to_status TEXT NOT NULL DEFAULT '', operation_type TEXT NOT NULL DEFAULT '', operator_type TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', reason_code TEXT NOT NULL DEFAULT '', reason_text TEXT NOT NULL DEFAULT '', payload_json TEXT NOT NULL DEFAULT '', create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	seeds := []string{
		`INSERT INTO sms_draw_activity (id, name, compliance_rule_summary, platform_id, tenant_id, merchant_id, is_deleted) VALUES (2001, '春季抽卡', '默认禁止收益承诺', 1, 10, 88, 0)`,
		`INSERT INTO sms_card_template (id, template_name, card_face_image, is_deleted) VALUES (21, 'SSR 兔兔', 'https://img.example.com/ssr-rabbit.png', 0)`,
		`INSERT INTO sms_draw_participation_record (id, activity_id, member_id, request_id, trace_id, result_type, result_status, failure_reason, create_time, is_deleted) VALUES
			(1, 2001, 3001, 'req-front-1', 'trace-front-1', 'won', 'won_pending_asset', '', '2026-04-18 10:00:00', 0),
			(2, 2001, 3001, 'req-front-2', 'trace-front-2', 'won', 'won_pending_asset', '', '2026-04-18 10:05:00', 0),
			(3, 2001, 4001, 'req-front-3', 'trace-front-3', 'won', 'won_pending_asset', '', '2026-04-18 10:10:00', 0)`,
		`INSERT INTO sms_card_instance (id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status, display_status, compliance_status, display_reason, compliance_reason, rule_snapshot_json, mint_task_id, issued_at, create_time, update_time, is_deleted) VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-front-1', 'trace-front-1', 21, 'SSR', 'CARD-FRONT-001', 'asset_created', 'mint_processing', '', 'processing', 'display_visible', 'compliance_clear', '', '', '', 11, '2026-04-18 10:00:00', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(2, 1, 10, 88, 2001, 3001, 2, 'req-front-2', 'trace-front-2', 21, 'SSR', 'CARD-FRONT-002', 'asset_created', 'mint_success', 'token-front-123456', 'success', 'display_hidden', 'compliance_review', '复核中', '复核中', '{\"scene\":\"review\"}', 12, '2026-04-18 10:05:00', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0),
			(3, 1, 10, 88, 2001, 4001, 3, 'req-front-3', 'trace-front-3', 21, 'SSR', 'CARD-FRONT-003', 'asset_created', 'mint_success', 'token-front-abcdef', 'success', 'display_visible', 'compliance_clear', '', '', '', 13, '2026-04-18 10:10:00', '2026-04-18 10:10:00', '2026-04-18 10:11:00', 0)`,
		`INSERT INTO sms_card_mint_task (id, platform_id, tenant_id, merchant_id, asset_instance_id, participation_record_id, activity_id, member_id, request_id, trace_id, task_status, mint_status, chain_status, token_id, chain_tx_id, last_receipt_summary, last_receipt_json, create_time, update_time, is_deleted) VALUES
			(11, 1, 10, 88, 1, 1, 2001, 3001, 'req-front-1', 'trace-front-1', 'running', 'mint_processing', 'processing', '', '', '链上处理中', '{\"status\":\"processing\"}', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(12, 1, 10, 88, 2, 2, 2001, 3001, 'req-front-2', 'trace-front-2', 'succeeded', 'mint_success', 'success', 'token-front-123456', 'tx-front-2', '链上成功', '{\"status\":\"success\"}', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0),
			(13, 1, 10, 88, 3, 3, 2001, 4001, 'req-front-3', 'trace-front-3', 'succeeded', 'mint_success', 'success', 'token-front-abcdef', 'tx-front-3', '链上成功', '{\"status\":\"success\"}', '2026-04-18 10:10:00', '2026-04-18 10:11:00', 0)`,
		`INSERT INTO sms_card_asset_log (id, asset_instance_id, participation_record_id, from_status, to_status, operation_type, operator_type, trace_id, reason_code, reason_text, payload_json, create_time) VALUES
			(1, 1, 1, '', 'mint_processing', 'mint_dispatching', 'system', 'trace-front-1', '', '链上处理中', '{}', '2026-04-18 10:01:00'),
			(2, 2, 2, 'compliance_clear', 'compliance_review', 'asset_compliance_review', 'manual', 'trace-front-2', '', '复核中', '{}', '2026-04-18 10:06:00')`,
	}
	for _, seed := range seeds {
		if err := db.Exec(seed).Error; err != nil {
			t.Fatalf("seed data failed: %v", err)
		}
	}

	return db
}

func newFrontAssetCtx(memberID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("88"))
	ctx = context.WithValue(ctx, "memberId", json.Number(fmt.Sprintf("%d", memberID)))
	return ctx
}

func newFrontAssetServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		CardMintService: digitalcardmint.NewService(newFrontAssetTestDB(t), nil, nil),
	}
}

func TestQueryMyDigitalCardAssetListOnlyReturnsCurrentMemberAssets(t *testing.T) {
	logic := NewQueryMyDigitalCardAssetListLogic(newFrontAssetCtx(3001), newFrontAssetServiceContext(t))

	resp, err := logic.QueryMyDigitalCardAssetList(&types.QueryMyDigitalCardAssetListReq{
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("QueryMyDigitalCardAssetList returned error: %v", err)
	}
	if resp.Data.Total != 2 || len(resp.Data.List) != 2 {
		t.Fatalf("expected two assets for member 3001, got total=%d len=%d", resp.Data.Total, len(resp.Data.List))
	}
	for _, item := range resp.Data.List {
		if item.AssetNo == "CARD-FRONT-003" {
			t.Fatalf("unexpected other member asset in response: %+v", item)
		}
	}
	if resp.Data.List[0].AssetNo != "CARD-FRONT-002" {
		t.Fatalf("expected latest asset first, got %+v", resp.Data.List[0])
	}
}

func TestQueryMyDigitalCardAssetDetailRejectsOtherMemberAsset(t *testing.T) {
	logic := NewQueryMyDigitalCardAssetDetailLogic(newFrontAssetCtx(3001), newFrontAssetServiceContext(t))

	_, err := logic.QueryMyDigitalCardAssetDetail(&types.QueryMyDigitalCardAssetDetailReq{AssetInstanceId: 3})
	if err == nil {
		t.Fatalf("expected cross-member access to be rejected")
	}
}
