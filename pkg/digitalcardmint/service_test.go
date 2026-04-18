package digitalcardmint

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/antchain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDigitalCardMintTestDB(t *testing.T) *gorm.DB {
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
		`CREATE UNIQUE INDEX uk_card_mint_task_token_guard ON sms_card_mint_task(token_id) WHERE is_deleted = 0 AND token_id <> ''`,
		`CREATE UNIQUE INDEX uk_card_instance_token_guard ON sms_card_instance(token_id) WHERE is_deleted = 0 AND token_id <> ''`,
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
		t.Fatalf("seed draw activity failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_template
			(id, template_name, display_status, content_audit_status, credential_ref, status, audit_status, is_deleted)
		VALUES
			(21, 'SSR 兔兔', 1, 2, 'cred-1', 1, 2, 0)
	`).Error; err != nil {
		t.Fatalf("seed card template failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, pool_id, template_id, request_id, trace_id, eligibility_snapshot_json, asset_instance_id, is_deleted)
		VALUES
			(1, 2001, 3001, 11, 21, 'req-1', 'trace-1', '{"realNameStatus":"verified"}', 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed participation record failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_instance
			(id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, scope, pool_id, template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status, mint_task_id, create_by, is_deleted)
		VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-1', 'trace-1', 'platform:1,tenant:10,merchant:88', 11, 21, 'SSR', 'CARD-001', 'asset_created', 'mint_pending', '', '', 0, 0, 0)
	`).Error; err != nil {
		t.Fatalf("seed card instance failed: %v", err)
	}

	return db
}

func createCardAssetLogTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec(`CREATE TABLE sms_card_asset_log (
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
	)`).Error; err != nil {
		t.Fatalf("create asset log table failed: %v", err)
	}
}

type countingPublisher struct {
	calls int
}

func (p *countingPublisher) SendMessage(string, string, string, string, []byte) error {
	p.calls++
	return nil
}

type stubAntChainClient struct {
	response *antchain.MintTokenResponse
	err      error
	calls    int
}

func (c *stubAntChainClient) MintToken(context.Context, *antchain.MintTokenRequest) (*antchain.MintTokenResponse, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	if c.response == nil {
		return nil, errors.New("missing antchain response")
	}
	out := *c.response
	return &out, nil
}

func TestEnsureTaskTxIsIdempotent(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	service := NewService(db, nil, nil)

	var firstTaskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		firstTaskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("first EnsureTaskTx returned error: %v", err)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		if task.ID != firstTaskID {
			t.Fatalf("expected same task id, got %d want %d", task.ID, firstTaskID)
		}
		return nil
	}); err != nil {
		t.Fatalf("second EnsureTaskTx returned error: %v", err)
	}

	var taskCount int64
	if err := db.Table(CardMintTaskRow{}.TableName()).Count(&taskCount).Error; err != nil {
		t.Fatalf("count tasks failed: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("expected 1 task, got %d", taskCount)
	}

	var instance CardInstanceRow
	if err := db.Table(instance.TableName()).Where("id = 1").Take(&instance).Error; err != nil {
		t.Fatalf("load instance failed: %v", err)
	}
	if instance.MintTaskID != firstTaskID {
		t.Fatalf("expected mint_task_id=%d, got %d", firstTaskID, instance.MintTaskID)
	}
}

func TestEnsureTaskTxRejectsForbiddenPrerequisites(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func(*gorm.DB) error
		wantErr string
	}{
		{
			name: "activity offline",
			prepare: func(db *gorm.DB) error {
				return db.Exec(`UPDATE sms_draw_activity SET status = 0 WHERE id = 2001`).Error
			},
			wantErr: "所属活动已下线",
		},
		{
			name: "activity compliance rejected",
			prepare: func(db *gorm.DB) error {
				return db.Exec(`UPDATE sms_draw_activity SET content_audit_status = 3 WHERE id = 2001`).Error
			},
			wantErr: "所属活动合规状态禁止发链",
		},
		{
			name: "template hidden",
			prepare: func(db *gorm.DB) error {
				return db.Exec(`UPDATE sms_card_template SET display_status = 0 WHERE id = 21`).Error
			},
			wantErr: "所属模板已隐藏",
		},
		{
			name: "real name snapshot not verified",
			prepare: func(db *gorm.DB) error {
				if err := db.Exec(`UPDATE sms_draw_activity SET real_name_required = 1 WHERE id = 2001`).Error; err != nil {
					return err
				}
				return db.Exec(`UPDATE sms_draw_participation_record SET eligibility_snapshot_json = '{"realNameStatus":"pending"}' WHERE id = 1`).Error
			},
			wantErr: "实名快照未通过",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := newDigitalCardMintTestDB(t)
			if err := tc.prepare(db); err != nil {
				t.Fatalf("prepare test data failed: %v", err)
			}

			service := NewService(db, nil, nil)
			err := db.Transaction(func(tx *gorm.DB) error {
				_, ensureErr := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
				return ensureErr
			})
			if err == nil {
				t.Fatal("expected EnsureTaskTx to reject forbidden prerequisite")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error to contain %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDispatchTaskSkipsTerminalTask(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	publisher := &countingPublisher{}
	service := NewService(db, publisher, nil)

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	if err := db.Table(CardMintTaskRow{}.TableName()).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"task_status":  TaskStatusSucceeded,
			"mint_status":  MintStatusSuccess,
			"chain_status": ChainStatusSuccess,
			"token_id":     "token-locked",
		}).Error; err != nil {
		t.Fatalf("update task terminal status failed: %v", err)
	}
	if err := db.Table(CardInstanceRow{}.TableName()).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"mint_status":  MintStatusSuccess,
			"chain_status": ChainStatusSuccess,
			"token_id":     "token-locked",
		}).Error; err != nil {
		t.Fatalf("update instance terminal status failed: %v", err)
	}

	if err := service.DispatchTask(context.Background(), taskID, "replay dispatch"); err != nil {
		t.Fatalf("DispatchTask returned error: %v", err)
	}
	if publisher.calls != 0 {
		t.Fatalf("expected dispatch to be skipped, got %d publish calls", publisher.calls)
	}

	var task CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.TaskStatus != TaskStatusSucceeded || task.TokenID != "token-locked" {
		t.Fatalf("expected terminal task to stay unchanged, got %+v", task)
	}
}

func TestScanDueTasksReplaysStaleDispatchedTaskWithoutExecuteLease(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	now := time.Date(2026, 4, 18, 11, 45, 0, 0, time.Local)
	client := &stubAntChainClient{
		response: &antchain.MintTokenResponse{
			TokenID:        "token-stale-dispatch",
			ChainTxID:      "tx-stale-dispatch",
			ChainStatus:    ChainStatusSuccess,
			ReceiptSummary: "recovered stale dispatch",
			ReceiptJSON:    `{"tokenId":"token-stale-dispatch","chainTxId":"tx-stale-dispatch"}`,
			ConfirmedAt:    now,
		},
	}
	service := NewService(db, nil, client)
	service.Now = func() time.Time { return now }

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	staleAt := now.Add(-5 * time.Minute)
	if err := db.Table(CardMintTaskRow{}.TableName()).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"task_status":       TaskStatusDispatched,
			"mint_status":       MintStatusProcessing,
			"chain_status":      ChainStatusProcessing,
			"last_execute_at":   nil,
			"update_time":       staleAt,
			"last_error_code":   "",
			"last_error_reason": "",
		}).Error; err != nil {
		t.Fatalf("prepare dispatched task failed: %v", err)
	}
	if err := db.Table(CardInstanceRow{}.TableName()).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"mint_status":  MintStatusProcessing,
			"chain_status": ChainStatusProcessing,
			"update_time":  staleAt,
		}).Error; err != nil {
		t.Fatalf("prepare instance status failed: %v", err)
	}

	stats, err := service.ScanDueTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("ScanDueTasks returned error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected stale dispatched task to be executed once, got %d calls", client.calls)
	}
	if stats.Executed != 1 {
		t.Fatalf("expected one executed task in stats, got %+v", stats)
	}

	var task CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.TaskStatus != TaskStatusSucceeded || task.TokenID != "token-stale-dispatch" {
		t.Fatalf("expected stale dispatched task to recover successfully, got %+v", task)
	}
}

func TestExecuteTaskSkipsLeasedRunningTask(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	fixedNow := time.Date(2026, 4, 18, 11, 0, 0, 0, time.Local)
	client := &stubAntChainClient{
		response: &antchain.MintTokenResponse{
			TokenID:        "token-unused",
			ChainTxID:      "tx-unused",
			ChainStatus:    ChainStatusSuccess,
			ReceiptSummary: "unused",
			ReceiptJSON:    `{"tokenId":"token-unused"}`,
			ConfirmedAt:    fixedNow,
		},
	}
	service := NewService(db, nil, client)
	service.Now = func() time.Time { return fixedNow }

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	lastExecuteAt := fixedNow.Add(-30 * time.Second)
	if err := db.Table(CardMintTaskRow{}.TableName()).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"task_status":     TaskStatusRunning,
			"mint_status":     MintStatusProcessing,
			"chain_status":    ChainStatusProcessing,
			"last_execute_at": lastExecuteAt,
		}).Error; err != nil {
		t.Fatalf("update task running status failed: %v", err)
	}
	if err := db.Table(CardInstanceRow{}.TableName()).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"mint_status":  MintStatusProcessing,
			"chain_status": ChainStatusProcessing,
		}).Error; err != nil {
		t.Fatalf("update instance processing status failed: %v", err)
	}

	result, err := service.ExecuteTask(context.Background(), taskID, OperatorSystem)
	if err != nil {
		t.Fatalf("ExecuteTask returned error: %v", err)
	}
	if client.calls != 0 {
		t.Fatalf("expected no new mint call, got %d", client.calls)
	}
	if result.TaskStatus != TaskStatusRunning {
		t.Fatalf("expected task to remain running, got %+v", result)
	}
}

func TestExecuteTaskMovesToManualReviewWhenTokenAlreadyBound(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	fixedNow := time.Date(2026, 4, 18, 11, 30, 0, 0, time.Local)
	client := &stubAntChainClient{
		response: &antchain.MintTokenResponse{
			TokenID:        "token-conflict",
			ChainTxID:      "tx-conflict",
			ChainStatus:    ChainStatusSuccess,
			ReceiptSummary: "duplicate token",
			ReceiptJSON:    `{"tokenId":"token-conflict","chainTxId":"tx-conflict"}`,
			ConfirmedAt:    fixedNow,
		},
	}
	service := NewService(db, nil, client)
	service.Now = func() time.Time { return fixedNow }

	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, pool_id, template_id, request_id, trace_id, asset_instance_id, is_deleted)
		VALUES
			(2, 2001, 3002, 12, 22, 'req-2', 'trace-2', 2, 0)
	`).Error; err != nil {
		t.Fatalf("seed second participation record failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_instance
			(id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, scope, pool_id, template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status, mint_task_id, create_by, is_deleted)
		VALUES
			(2, 1, 10, 88, 2001, 3002, 2, 'req-2', 'trace-2', 'platform:1,tenant:10,merchant:88', 12, 22, 'SSR', 'CARD-002', 'asset_created', 'mint_success', 'token-conflict', 'success', 0, 0, 0)
	`).Error; err != nil {
		t.Fatalf("seed conflicting instance failed: %v", err)
	}

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	result, err := service.ExecuteTask(context.Background(), taskID, OperatorSystem)
	if err != nil {
		t.Fatalf("ExecuteTask returned error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected one mint call, got %d", client.calls)
	}
	if result.TaskStatus != TaskStatusManualReview || !result.ManualRequired {
		t.Fatalf("expected manual review result, got %+v", result)
	}

	var task CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.LastErrorCode != ErrorCodeTokenBindingConflict || task.TokenID != "" {
		t.Fatalf("expected token conflict snapshot without local binding, got %+v", task)
	}

	var instance CardInstanceRow
	if err := db.Table(instance.TableName()).Where("id = ?", 1).Take(&instance).Error; err != nil {
		t.Fatalf("load instance failed: %v", err)
	}
	if instance.MintStatus != MintStatusManualReview || instance.TokenID != "" {
		t.Fatalf("expected instance to enter manual review without token binding, got %+v", instance)
	}
}

func TestExecuteTaskSuccessWritesBackTokenAndLogs(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	mockClient := antchain.NewMockClient()
	fixedNow := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	service := NewService(db, nil, mockClient)
	service.Now = func() time.Time { return fixedNow }

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	result, err := service.ExecuteTask(context.Background(), taskID, OperatorSystem)
	if err != nil {
		t.Fatalf("ExecuteTask returned error: %v", err)
	}
	if result.TaskStatus != TaskStatusSucceeded {
		t.Fatalf("expected task status %s, got %s", TaskStatusSucceeded, result.TaskStatus)
	}
	if result.TokenID == "" {
		t.Fatal("expected token id to be written back")
	}

	var task CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.TokenID == "" || task.ChainStatus != ChainStatusSuccess {
		t.Fatalf("expected task to persist success receipt, got %+v", task)
	}

	var instance CardInstanceRow
	if err := db.Table(instance.TableName()).Where("id = 1").Take(&instance).Error; err != nil {
		t.Fatalf("load instance failed: %v", err)
	}
	if instance.MintStatus != MintStatusSuccess || instance.TokenID == "" {
		t.Fatalf("expected instance success snapshot, got %+v", instance)
	}

	var logCount int64
	if err := db.Table(CardAssetLogRow{}.TableName()).Count(&logCount).Error; err != nil {
		t.Fatalf("count asset logs failed: %v", err)
	}
	if logCount != 2 {
		t.Fatalf("expected 2 asset logs, got %d", logCount)
	}
}

func TestExecuteTaskReplaysStoredReceiptAfterWritebackFailure(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	now := time.Date(2026, 4, 18, 12, 0, 0, 0, time.Local)
	client := &stubAntChainClient{
		response: &antchain.MintTokenResponse{
			TokenID:        "token-recover",
			ChainTxID:      "tx-recover",
			ChainStatus:    ChainStatusSuccess,
			ReceiptSummary: "receipt persisted later",
			ReceiptJSON:    `{"tokenId":"token-recover","chainTxId":"tx-recover"}`,
			ConfirmedAt:    now,
		},
	}
	service := NewService(db, nil, client)
	service.Now = func() time.Time { return now }

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	if err := db.Exec(`DROP TABLE sms_card_asset_log`).Error; err != nil {
		t.Fatalf("drop asset log table failed: %v", err)
	}

	firstResult, err := service.ExecuteTask(context.Background(), taskID, OperatorSystem)
	if err != nil {
		t.Fatalf("first ExecuteTask returned error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected one mint call after first execute, got %d", client.calls)
	}
	if firstResult.TaskStatus != TaskStatusFailed || firstResult.ChainStatus != ChainStatusSuccess {
		t.Fatalf("expected receipt recovery pending result, got %+v", firstResult)
	}

	var task CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.LastErrorCode != ErrorCodeReceiptWritebackFailed || task.TokenID != "token-recover" {
		t.Fatalf("expected receipt snapshot to be preserved for recovery, got %+v", task)
	}

	createCardAssetLogTable(t, db)
	now = now.Add(2 * time.Minute)

	secondResult, err := service.ExecuteTask(context.Background(), taskID, OperatorSystem)
	if err != nil {
		t.Fatalf("second ExecuteTask returned error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected receipt replay without second mint call, got %d", client.calls)
	}
	if secondResult.TaskStatus != TaskStatusSucceeded || secondResult.TokenID != "token-recover" {
		t.Fatalf("expected task to recover from stored receipt, got %+v", secondResult)
	}

	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("reload task failed: %v", err)
	}
	if task.TaskStatus != TaskStatusSucceeded || task.RetryCount != 1 {
		t.Fatalf("expected successful replay without extra mint attempt, got %+v", task)
	}
}

func TestExecuteTaskEscalatesToManualReviewAtRetryLimit(t *testing.T) {
	db := newDigitalCardMintTestDB(t)
	mockClient := antchain.NewMockClient()
	service := NewService(db, nil, mockClient)
	service.Now = func() time.Time {
		return time.Date(2026, 4, 18, 10, 30, 0, 0, time.Local)
	}

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := service.EnsureTaskTx(context.Background(), tx, 1, OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	if err := db.Table(CardMintTaskRow{}.TableName()).
		Where("id = ?", taskID).
		Update("max_retry_count", 1).Error; err != nil {
		t.Fatalf("update max_retry_count failed: %v", err)
	}
	mockClient.SetError("card-mint:1", errors.New("chain timeout"))

	result, err := service.ExecuteTask(context.Background(), taskID, OperatorJob)
	if err != nil {
		t.Fatalf("ExecuteTask returned error: %v", err)
	}
	if result.TaskStatus != TaskStatusManualReview {
		t.Fatalf("expected task status %s, got %s", TaskStatusManualReview, result.TaskStatus)
	}
	if !result.ManualRequired {
		t.Fatal("expected manual review to be required")
	}

	var task CardMintTaskRow
	if err := db.Table(task.TableName()).Where("id = ?", taskID).Take(&task).Error; err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if task.ManualRequired != 1 || task.NextRetryAt != nil {
		t.Fatalf("expected manual review without next retry, got %+v", task)
	}

	var instance CardInstanceRow
	if err := db.Table(instance.TableName()).Where("id = 1").Take(&instance).Error; err != nil {
		t.Fatalf("load instance failed: %v", err)
	}
	if instance.MintStatus != MintStatusManualReview || instance.ChainStatus != ChainStatusFailed {
		t.Fatalf("expected instance manual review snapshot, got %+v", instance)
	}
}
