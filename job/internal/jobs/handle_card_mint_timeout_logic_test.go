package jobs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/antchain"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newJobMintTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_participation_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			eligibility_snapshot_json TEXT NOT NULL DEFAULT '',
			asset_instance_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			real_name_required INTEGER NOT NULL DEFAULT 0,
			publish_readiness INTEGER NOT NULL DEFAULT 1,
			copyright_status INTEGER NOT NULL DEFAULT 2,
			content_audit_status INTEGER NOT NULL DEFAULT 2,
			status INTEGER NOT NULL DEFAULT 1,
			audit_status INTEGER NOT NULL DEFAULT 2,
			is_enabled INTEGER NOT NULL DEFAULT 1,
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
			asset_no TEXT NOT NULL,
			asset_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			last_receipt_at DATETIME NULL,
			mint_task_id INTEGER NOT NULL DEFAULT 0,
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
		`CREATE UNIQUE INDEX uk_card_mint_task_asset ON sms_card_mint_task(asset_instance_id, is_deleted)`,
		`CREATE UNIQUE INDEX uk_card_mint_task_idempotency ON sms_card_mint_task(idempotency_key, is_deleted)`,
		`CREATE TABLE sms_card_asset_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_instance_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL,
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

	if err := db.Exec(`
		INSERT INTO sms_draw_activity
			(id, name, platform_id, tenant_id, merchant_id, real_name_required, publish_readiness, copyright_status, content_audit_status, status, audit_status, is_enabled, is_deleted)
		VALUES
			(2001, '春季抽卡', 1, 10, 88, 0, 1, 2, 2, 1, 2, 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_template
			(id, template_name, display_status, content_audit_status, credential_ref, status, audit_status, is_deleted)
		VALUES
			(21, 'SSR 兔兔', 1, 2, 'cred-1', 1, 2, 0)
	`).Error; err != nil {
		t.Fatalf("seed template failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, pool_id, template_id, request_id, trace_id, eligibility_snapshot_json, asset_instance_id, is_deleted)
		VALUES
			(1, 2001, 3001, 11, 21, 'req-job', 'trace-job', '{"realNameStatus":"verified"}', 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed participation record failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_instance
			(id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, scope, pool_id, template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status, mint_task_id, create_by, is_deleted)
		VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-job', 'trace-job', 'platform:1,tenant:10,merchant:88', 11, 21, 'SSR', 'CARD-JOB', 'asset_created', 'mint_pending', '', '', 0, 0, 0)
	`).Error; err != nil {
		t.Fatalf("seed card instance failed: %v", err)
	}

	return db
}

func newJobMintService(t *testing.T) (*digitalcardmint.Service, int64, *gorm.DB, *antchain.MockClient) {
	t.Helper()

	db := newJobMintTestDB(t)
	mockClient := antchain.NewMockClient()
	service := digitalcardmint.NewService(db, nil, mockClient)
	service.Now = func() time.Time {
		return time.Date(2026, 4, 18, 12, 0, 0, 0, time.Local)
	}

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, digitalcardmint.OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	return service, taskID, db, mockClient
}

func TestHandleCardMintTimeoutExecutesPendingDispatchWithoutMQ(t *testing.T) {
	service, taskID, db, _ := newJobMintService(t)

	HandleCardMintTimeout(context.Background(), service)

	var task digitalcardmint.CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.TaskStatus != digitalcardmint.TaskStatusSucceeded || task.TokenID == "" {
		t.Fatalf("expected pending task to be executed by job, got %+v", task)
	}
}

func TestHandleCardMintTimeoutEscalatesFailedTaskAtRetryLimit(t *testing.T) {
	service, taskID, db, mockClient := newJobMintService(t)

	if err := db.Table(digitalcardmint.CardMintTaskRow{}.TableName()).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"task_status":       digitalcardmint.TaskStatusFailed,
			"mint_status":       digitalcardmint.MintStatusFailed,
			"chain_status":      digitalcardmint.ChainStatusFailed,
			"retry_count":       0,
			"max_retry_count":   1,
			"next_retry_at":     time.Date(2026, 4, 18, 11, 55, 0, 0, time.Local),
			"last_error_code":   "mint_execute_failed",
			"last_error_reason": "previous failed",
		}).Error; err != nil {
		t.Fatalf("prepare failed task failed: %v", err)
	}
	mockClient.SetError("card-mint:1", errors.New("chain timeout"))

	HandleCardMintTimeout(context.Background(), service)

	var task digitalcardmint.CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.TaskStatus != digitalcardmint.TaskStatusManualReview || task.ManualRequired != 1 {
		t.Fatalf("expected failed task to escalate to manual review, got %+v", task)
	}
}
