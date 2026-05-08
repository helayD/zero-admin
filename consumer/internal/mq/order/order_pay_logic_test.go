package order

import (
	"context"
	"fmt"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOrderPayTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE oms_order_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			sku_id INTEGER NOT NULL,
			sku_name TEXT NOT NULL DEFAULT '',
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE pms_product_sku (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			spu_id INTEGER NOT NULL,
			fulfillment_mode TEXT NOT NULL DEFAULT 'physical_delivery',
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE pms_product_spu (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			fulfillment_mode TEXT NOT NULL DEFAULT 'physical_delivery',
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
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
			refund_policy TEXT NOT NULL DEFAULT 'freeze_card',
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_name TEXT NOT NULL DEFAULT '',
			display_status INTEGER NOT NULL DEFAULT 1,
			content_audit_status INTEGER NOT NULL DEFAULT 2,
			credential_ref TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			audit_status INTEGER NOT NULL DEFAULT 2,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_participation_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL DEFAULT 0,
			member_id INTEGER NOT NULL DEFAULT 0,
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			eligibility_snapshot_json TEXT NOT NULL DEFAULT '',
			asset_instance_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_instance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			activity_id INTEGER NOT NULL DEFAULT 0,
			member_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL DEFAULT 0,
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
			last_receipt_at DATETIME NULL,
			mint_task_id INTEGER NOT NULL DEFAULT 0,
			disposed_at DATETIME NULL,
			disposed_by INTEGER NOT NULL DEFAULT 0,
			issued_at DATETIME NULL,
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
			update_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX uk_card_mint_task_asset_test ON sms_card_mint_task(asset_instance_id, is_deleted)`,
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

	return db
}

func seedOrderTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 创建商品 SPU（提货卡模式）
	if err := db.Exec(`INSERT INTO pms_product_spu (id, platform_id, tenant_id, merchant_id, fulfillment_mode, fulfillment_rule_id, is_deleted) VALUES (1001, 1, 10, 88, 'digital_asset', 1, 0)`).Error; err != nil {
		t.Fatalf("seed product spu failed: %v", err)
	}

	// 创建商品 SKU（继承 SPU 的履约模式）
	if err := db.Exec(`INSERT INTO pms_product_sku (id, spu_id, fulfillment_mode, fulfillment_rule_id, is_deleted) VALUES (2001, 1001, '', 1, 0)`).Error; err != nil {
		t.Fatalf("seed product sku failed: %v", err)
	}

	// 创建发卡规则
	if err := db.Exec(`INSERT INTO sms_product_fulfillment_rule (id, platform_id, tenant_id, merchant_id, rule_name, rule_status, card_template_id, expire_days, transferable, transfer_limit, claim_condition, redemption_condition, refund_policy, is_deleted) VALUES (1, 1, 10, 88, '测试规则', 1, 21, 365, 1, 3, '', '', 'freeze_card', 0)`).Error; err != nil {
		t.Fatalf("seed fulfillment rule failed: %v", err)
	}

	// 创建卡片模板
	if err := db.Exec(`INSERT INTO sms_card_template (id, template_name, display_status, content_audit_status, credential_ref, status, audit_status, is_deleted) VALUES (21, 'SSR 兔兔', 1, 2, 'cred-1', 1, 2, 0)`).Error; err != nil {
		t.Fatalf("seed card template failed: %v", err)
	}

	// 创建订单明细
	if err := db.Exec(`INSERT INTO oms_order_item (id, order_id, sku_id, sku_name, is_deleted) VALUES (3001, 4001, 2001, '提货卡商品', 0)`).Error; err != nil {
		t.Fatalf("seed order item failed: %v", err)
	}
}

func buildOrderPayEventPayload(orderID int64) []byte {
	payload := EventPayload{
		EventID:    fmt.Sprintf("order-paid-%d", orderID),
		TraceID:    fmt.Sprintf("trace-order-%d", orderID),
		PlatformID: 1,
		TenantID:   10,
		MerchantID: 88,
		ActorID:    5001,
		EntityID:   orderID,
		Action:     "paid",
		Version:    "v1",
		Data: map[string]interface{}{
			"orderNo": fmt.Sprintf("OR%d", orderID),
		},
	}
	body, _ := sonic.Marshal(payload)
	return body
}

func TestProcessPaidOrderDigitalAssetsCreatesAssetForDigitalProduct(t *testing.T) {
	db := newOrderPayTestDB(t)
	seedOrderTestData(t, db)

	service := digitalcardmint.NewService(db, nil, nil)

	payload := &EventPayload{
		EventID:    "order-paid-4001",
		TraceID:    "trace-order-4001",
		PlatformID: 1,
		TenantID:   10,
		MerchantID: 88,
		ActorID:    5001,
		EntityID:   4001,
		Action:     "paid",
	}

	ctx := context.Background()
	err := processPaidOrderDigitalAssets(ctx, payload, service, db)
	if err != nil {
		t.Fatalf("processPaidOrderDigitalAssets returned error: %v", err)
	}

	// 验证卡片资产已创建
	var count int64
	db.Table("sms_card_instance").
		Where("source_type = ? AND source_id = ? AND member_id = ? AND is_deleted = 0", "purchase", 3001, 5001).
		Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 card instance, got %d", count)
	}
}

func TestProcessPaidOrderDigitalAssetsSkipsPhysicalProduct(t *testing.T) {
	db := newOrderPayTestDB(t)

	// 创建实物商品
	if err := db.Exec(`INSERT INTO pms_product_spu (id, platform_id, tenant_id, merchant_id, fulfillment_mode, fulfillment_rule_id, is_deleted) VALUES (1002, 1, 10, 88, 'physical_delivery', 0, 0)`).Error; err != nil {
		t.Fatalf("seed product spu failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_sku (id, spu_id, fulfillment_mode, fulfillment_rule_id, is_deleted) VALUES (2002, 1002, '', 0, 0)`).Error; err != nil {
		t.Fatalf("seed product sku failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO oms_order_item (id, order_id, sku_id, sku_name, is_deleted) VALUES (3002, 4002, 2002, '实物商品', 0)`).Error; err != nil {
		t.Fatalf("seed order item failed: %v", err)
	}

	service := digitalcardmint.NewService(db, nil, nil)

	payload := &EventPayload{
		EventID:    "order-paid-4002",
		TraceID:    "trace-order-4002",
		PlatformID: 1,
		TenantID:   10,
		MerchantID: 88,
		ActorID:    5001,
		EntityID:   4002,
		Action:     "paid",
	}

	ctx := context.Background()
	err := processPaidOrderDigitalAssets(ctx, payload, service, db)
	if err != nil {
		t.Fatalf("processPaidOrderDigitalAssets returned error: %v", err)
	}

	// 验证没有创建卡片资产
	var count int64
	db.Table("sms_card_instance").
		Where("source_type = ? AND source_id = ? AND is_deleted = 0", "purchase", 3002).
		Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 card instance for physical product, got %d", count)
	}
}

func TestProcessPaidOrderDigitalAssetsIdempotent(t *testing.T) {
	db := newOrderPayTestDB(t)
	seedOrderTestData(t, db)

	service := digitalcardmint.NewService(db, nil, nil)

	payload := &EventPayload{
		EventID:    "order-paid-4001",
		TraceID:    "trace-order-4001",
		PlatformID: 1,
		TenantID:   10,
		MerchantID: 88,
		ActorID:    5001,
		EntityID:   4001,
		Action:     "paid",
	}

	ctx := context.Background()

	// 第一次调用
	err := processPaidOrderDigitalAssets(ctx, payload, service, db)
	if err != nil {
		t.Fatalf("first call returned error: %v", err)
	}

	// 第二次调用（幂等）
	err = processPaidOrderDigitalAssets(ctx, payload, service, db)
	if err != nil {
		t.Fatalf("second call returned error: %v", err)
	}

	// 验证只有一个卡片资产
	var count int64
	db.Table("sms_card_instance").
		Where("source_type = ? AND source_id = ? AND member_id = ? AND is_deleted = 0", "purchase", 3001, 5001).
		Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 card instance after idempotent call, got %d", count)
	}
}

func TestProcessPaidOrderDigitalAssetsHandlesMissingPayload(t *testing.T) {
	db := newOrderPayTestDB(t)
	service := digitalcardmint.NewService(db, nil, nil)

	ctx := context.Background()

	// nil payload
	err := processPaidOrderDigitalAssets(ctx, nil, service, db)
	if err != nil {
		t.Fatalf("nil payload should return nil error, got: %v", err)
	}

	// invalid entity ID
	payload := &EventPayload{
		EntityID: 0,
	}
	err = processPaidOrderDigitalAssets(ctx, payload, service, db)
	if err != nil {
		t.Fatalf("invalid entity ID should return nil error, got: %v", err)
	}
}
