package jobs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newReconcileTestDB 构造一个最小可运行的 SQLite schema，列名严格对齐生产 SQL。
// 主要意图：把 reconcile job 用到的列（特别是 main.user_id 而非 main.member_id、
// 不依赖 main.pay_status）固化在测试里，回归 Story 10.6 Code Review Fix #1/#2/#6。
func newReconcileTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		// oms_order_main 严格只有 user_id（没有 member_id / pay_status）
		`CREATE TABLE oms_order_main (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no TEXT NOT NULL DEFAULT '',
			user_id INTEGER NOT NULL DEFAULT 0,
			order_status INTEGER NOT NULL DEFAULT 0,
			pay_time DATETIME,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_order_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			sku_id INTEGER NOT NULL,
			sku_name TEXT NOT NULL DEFAULT '',
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE pms_product_spu (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			fulfillment_mode TEXT NOT NULL DEFAULT '',
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE pms_product_sku (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			spu_id INTEGER NOT NULL,
			fulfillment_mode TEXT NOT NULL DEFAULT '',
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_instance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			member_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			asset_no TEXT NOT NULL DEFAULT '',
			asset_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			source_type TEXT NOT NULL DEFAULT '',
			source_id INTEGER NOT NULL DEFAULT 0,
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			transferable INTEGER NOT NULL DEFAULT 0,
			transfer_limit INTEGER NOT NULL DEFAULT 0,
			claim_condition TEXT NOT NULL DEFAULT '',
			redemption_condition TEXT NOT NULL DEFAULT '',
			refund_policy TEXT NOT NULL DEFAULT '',
			compliance_status TEXT NOT NULL DEFAULT 'clear',
			display_status TEXT NOT NULL DEFAULT 'visible',
			scope TEXT NOT NULL DEFAULT '',
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			participation_record_id INTEGER NOT NULL DEFAULT 0,
			issued_at DATETIME,
			create_time DATETIME,
			update_time DATETIME,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_asset_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_instance_id INTEGER NOT NULL DEFAULT 0,
			participation_record_id INTEGER NOT NULL DEFAULT 0,
			from_status TEXT NOT NULL DEFAULT '',
			to_status TEXT NOT NULL DEFAULT '',
			operation_type TEXT NOT NULL DEFAULT '',
			operator_type TEXT NOT NULL DEFAULT 'system',
			trace_id TEXT NOT NULL DEFAULT '',
			reason_code TEXT NOT NULL DEFAULT '',
			reason_text TEXT NOT NULL DEFAULT '',
			payload_json TEXT,
			create_time DATETIME
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table failed: %v\n%s", err, stmt)
		}
	}
	return db
}

// TestReconcileMissingCardAssetsRunsAgainstRealColumns 锁定 reconcile_order_card_fulfillment.go
// 中 SQL 不能再引用 oms_order_main.member_id 或 oms_order_main.pay_status 这两个不存在的列。
// 任何回归一旦把列名写错（例如手抖改回 main.member_id），SQLite 的 sql 解析就会报错使本测试失败。
func TestReconcileMissingCardAssetsRunsAgainstRealColumns(t *testing.T) {
	db := newReconcileTestDB(t)

	// 准备一笔已支付的提货卡订单，明细未发卡（缺少 sms_card_instance 行）
	if err := db.Exec(
		`INSERT INTO oms_order_main (id, order_no, user_id, order_status, pay_time, platform_id, tenant_id, merchant_id, is_deleted)
		 VALUES (1001, 'OR1001', 9001, 2, datetime('now', '-10 minutes'), 1, 10, 88, 0)`,
	).Error; err != nil {
		t.Fatalf("seed order_main failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO oms_order_item (id, order_id, sku_id, sku_name, is_deleted) VALUES (3001, 1001, 2001, '提货卡商品', 0)`,
	).Error; err != nil {
		t.Fatalf("seed order_item failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO pms_product_spu (id, platform_id, tenant_id, merchant_id, fulfillment_mode, fulfillment_rule_id, is_deleted)
		 VALUES (5001, 1, 10, 88, 'digital_asset', 7001, 0)`,
	).Error; err != nil {
		t.Fatalf("seed product_spu failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO pms_product_sku (id, spu_id, fulfillment_mode, fulfillment_rule_id, is_deleted) VALUES (2001, 5001, '', 0, 0)`,
	).Error; err != nil {
		t.Fatalf("seed product_sku failed: %v", err)
	}

	// 不传 cardMintService（reconcileMissingCardAssets 内部会因为 nil service 报错），
	// 但 SQL 必须先能执行。我们传 nil 让函数早返回，并验证 SQL 阶段没报"Unknown column"。
	cnt, err := reconcileMissingCardAssets(context.Background(), db, nil)
	if err != nil {
		t.Fatalf("reconcileMissingCardAssets returned SQL error: %v", err)
	}
	// 由于没 service，cardMintService 调用会失败，行不会被修复但 cnt 仍是 0
	if cnt != 0 {
		t.Fatalf("expected 0 fixed (no service), got %d", cnt)
	}
}

// TestReconcileRefundedOrderCardsRunsAgainstRealColumns 锁定退款分支同样不能引用 main.pay_status，
// 且 order_status=6 才表示已退款。
func TestReconcileRefundedOrderCardsRunsAgainstRealColumns(t *testing.T) {
	db := newReconcileTestDB(t)

	// 已退款订单 + 仍处于 clear 的提货卡资产
	if err := db.Exec(
		`INSERT INTO oms_order_main (id, order_no, user_id, order_status, pay_time, platform_id, tenant_id, merchant_id, is_deleted)
		 VALUES (1002, 'OR1002', 9002, 6, datetime('now', '-1 hour'), 1, 10, 88, 0)`,
	).Error; err != nil {
		t.Fatalf("seed order_main failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO oms_order_item (id, order_id, sku_id, sku_name, is_deleted) VALUES (3002, 1002, 2002, '提货卡商品', 0)`,
	).Error; err != nil {
		t.Fatalf("seed order_item failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO sms_card_instance (
			id, platform_id, tenant_id, merchant_id, member_id, asset_no,
			asset_status, mint_status, chain_status,
			source_type, source_id,
			refund_policy, compliance_status, display_status, is_deleted
		) VALUES (
			970002, 1, 10, 88, 9002, 'CARD-REFUND-1',
			'created', 'mint_pending', 'unknown',
			'purchase', 3002,
			'freeze_card', 'clear', 'visible', 0
		)`,
	).Error; err != nil {
		t.Fatalf("seed card_instance failed: %v", err)
	}

	service := digitalcardmint.NewService(db, nil, nil)
	cnt, err := reconcileRefundedOrderCards(context.Background(), db, service)
	if err != nil {
		t.Fatalf("reconcileRefundedOrderCards returned error: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 refunded card disposed, got %d", cnt)
	}

	// 资产应该被冻结
	type row struct {
		ComplianceStatus string `gorm:"column:compliance_status"`
	}
	var got row
	if err := db.Table("sms_card_instance").Where("id = ?", 970002).Take(&got).Error; err != nil {
		t.Fatalf("verify card_instance failed: %v", err)
	}
	if got.ComplianceStatus != digitalcardmint.ComplianceStatusFrozen {
		t.Fatalf("expected compliance_status=frozen, got %s", got.ComplianceStatus)
	}

	// 触发幂等：第二次扫描不应再处置（compliance_status 已 frozen，会被 NOT IN 过滤）
	cnt2, err := reconcileRefundedOrderCards(context.Background(), db, service)
	if err != nil {
		t.Fatalf("second reconcileRefundedOrderCards returned error: %v", err)
	}
	if cnt2 != 0 {
		t.Fatalf("expected 0 disposes on idempotent rerun, got %d", cnt2)
	}
}

// 留个空引用避免未来误删 time/digitalcardmint 导入
var _ = time.Now
var _ = digitalcardmint.ComplianceStatusFrozen
