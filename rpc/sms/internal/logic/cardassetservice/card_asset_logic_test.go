package cardassetservicelogic

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCardAssetTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "card-asset.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_participation_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			request_id TEXT NOT NULL DEFAULT '',
			scope TEXT NOT NULL DEFAULT '',
			result_type TEXT NOT NULL DEFAULT '',
			result_status TEXT NOT NULL DEFAULT '',
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			rarity TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			asset_instance_id INTEGER NOT NULL DEFAULT 0,
			asset_no TEXT NOT NULL DEFAULT '',
			asset_status TEXT NOT NULL DEFAULT '',
			asset_created_at DATETIME NULL,
			update_time DATETIME NULL,
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
		`CREATE UNIQUE INDEX uk_card_instance_participation ON sms_card_instance(participation_record_id, is_deleted)`,
		`CREATE UNIQUE INDEX uk_card_instance_asset_no ON sms_card_instance(asset_no, is_deleted)`,
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
			payload_json TEXT,
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

func seedWinningParticipationRecord(t *testing.T, db *gorm.DB, recordID int64, requestID string) {
	t.Helper()
	if err := db.Exec(`
		INSERT INTO sms_draw_activity (id, platform_id, tenant_id, merchant_id, is_deleted)
		VALUES (1, 1, 10, 88, 0)
	`).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, request_id, scope, result_type, result_status, pool_id, template_id, rarity, trace_id, is_deleted)
		VALUES (?, 1, 3001, ?, 'platform:1,tenant:10,merchant:88', 'won', 'won_pending_asset', 11, 22, 'SSR', ?, 0)
	`, recordID, requestID, requestID).Error; err != nil {
		t.Fatalf("seed participation record failed: %v", err)
	}
}

func newMerchantScope(merchantID int64) *smsclient.GovernanceScope {
	return &smsclient.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformId: 1,
		TenantId:   10,
		MerchantId: merchantID,
	}
}

func TestEnsureCardInstanceByParticipationRecordCreatesLedgerAndSnapshot(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-1")

	asset, err := EnsureCardInstanceByParticipationRecord(context.Background(), db, 1, cardAssetOperatorSystem, "trace-1")
	if err != nil {
		t.Fatalf("EnsureCardInstanceByParticipationRecord returned error: %v", err)
	}
	if asset.AssetNo == "" {
		t.Fatal("expected asset_no to be generated")
	}
	if asset.AssetStatus != cardAssetStatusCreated {
		t.Fatalf("expected asset_status %s, got %s", cardAssetStatusCreated, asset.AssetStatus)
	}

	var record participationRecordSnapshot
	if err := db.Table(record.TableName()).Where("id = 1").Take(&record).Error; err != nil {
		t.Fatalf("load participation snapshot failed: %v", err)
	}
	if record.AssetInstanceID <= 0 || record.AssetNo == "" || record.AssetStatus != cardAssetStatusCreated || record.AssetCreatedAt == nil {
		t.Fatalf("expected participation snapshot to be updated, got %+v", record)
	}

	var logCount int64
	if err := db.Table(cardAssetLogRow{}.TableName()).Count(&logCount).Error; err != nil {
		t.Fatalf("count asset log failed: %v", err)
	}
	if logCount != 1 {
		t.Fatalf("expected 1 asset log, got %d", logCount)
	}

	var instance cardInstanceRow
	if err := db.Table(instance.TableName()).Where("participation_record_id = 1 AND is_deleted = 0").Take(&instance).Error; err != nil {
		t.Fatalf("load created asset failed: %v", err)
	}
	if instance.CreateTime == nil || instance.CreateTime.IsZero() {
		t.Fatalf("expected database default create_time to be populated, got %+v", instance.CreateTime)
	}
}

func TestEnsureCardInstanceByParticipationRecordIsIdempotent(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-2")

	first, err := EnsureCardInstanceByParticipationRecord(context.Background(), db, 1, cardAssetOperatorSystem, "trace-2")
	if err != nil {
		t.Fatalf("first ensure returned error: %v", err)
	}
	second, err := EnsureCardInstanceByParticipationRecord(context.Background(), db, 1, cardAssetOperatorJob, "trace-2-replay")
	if err != nil {
		t.Fatalf("second ensure returned error: %v", err)
	}
	if first.ID != second.ID || first.AssetNo != second.AssetNo {
		t.Fatalf("expected same asset instance on replay, first=%+v second=%+v", first, second)
	}

	var count int64
	if err := db.Table(cardInstanceRow{}.TableName()).Count(&count).Error; err != nil {
		t.Fatalf("count instances failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 asset instance, got %d", count)
	}
}

func TestEnsureCardInstanceByParticipationRecordRetriesAssetNoCollision(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-3")

	originalGenerator := assetNoGenerator
	defer func() {
		assetNoGenerator = originalGenerator
	}()
	calls := 0
	assetNoGenerator = func(now time.Time) string {
		calls++
		if calls == 1 {
			return "CARD-COLLISION"
		}
		return "CARD-UNIQUE"
	}

	if err := db.Exec(`
		INSERT INTO sms_card_instance
			(id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, scope, pool_id, template_id, rarity, asset_no, asset_status, mint_status, issued_at, create_by, is_deleted)
		VALUES (9, 1, 10, 88, 9, 999, 999, 'other', 'other', 'platform:1,tenant:10,merchant:88', 1, 2, 'R', 'CARD-COLLISION', 'asset_created', 'mint_pending', CURRENT_TIMESTAMP, 0, 0)
	`).Error; err != nil {
		t.Fatalf("seed existing asset failed: %v", err)
	}

	asset, err := EnsureCardInstanceByParticipationRecord(context.Background(), db, 1, cardAssetOperatorSystem, "trace-3")
	if err != nil {
		t.Fatalf("EnsureCardInstanceByParticipationRecord returned error: %v", err)
	}
	if asset.AssetNo != "CARD-UNIQUE" {
		t.Fatalf("expected retry asset_no CARD-UNIQUE, got %s", asset.AssetNo)
	}
}

func TestEnsureCardInstanceByParticipationRecordRejectsNonWinningRecord(t *testing.T) {
	db := newCardAssetTestDB(t)
	if err := db.Exec(`
		INSERT INTO sms_draw_activity (id, platform_id, tenant_id, merchant_id, is_deleted)
		VALUES (1, 1, 10, 88, 0)
	`).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, request_id, scope, result_type, result_status, is_deleted)
		VALUES (1, 1, 3001, 'req-not-won', 'platform:1,tenant:10,merchant:88', 'not_won', 'not_won', 0)
	`).Error; err != nil {
		t.Fatalf("seed non-winning record failed: %v", err)
	}

	if _, err := EnsureCardInstanceByParticipationRecord(context.Background(), db, 1, cardAssetOperatorSystem, "trace-4"); err == nil {
		t.Fatal("expected non-winning record to be rejected")
	}
}

func TestEnsureCardInstanceByParticipationRecordRollsBackWithTransaction(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-5")

	err := db.Transaction(func(tx *gorm.DB) error {
		if _, innerErr := EnsureCardInstanceByParticipationRecord(context.Background(), tx, 1, cardAssetOperatorSystem, "trace-5"); innerErr != nil {
			return innerErr
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected transaction rollback error")
	}

	var instanceCount int64
	if err := db.Table(cardInstanceRow{}.TableName()).Count(&instanceCount).Error; err != nil {
		t.Fatalf("count instances failed: %v", err)
	}
	if instanceCount != 0 {
		t.Fatalf("expected no asset instance after rollback, got %d", instanceCount)
	}

	var logCount int64
	if err := db.Table(cardAssetLogRow{}.TableName()).Count(&logCount).Error; err != nil {
		t.Fatalf("count asset logs failed: %v", err)
	}
	if logCount != 0 {
		t.Fatalf("expected no asset logs after rollback, got %d", logCount)
	}
}

func TestBackfillWinningCardInstancesIsReentrant(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-6")

	total, assets, err := BackfillWinningCardInstances(context.Background(), db, pkgscope.DefaultScope(1, 10, 88), 1, 3001, 10, cardAssetOperatorJob, "trace-6")
	if err != nil {
		t.Fatalf("BackfillWinningCardInstances returned error: %v", err)
	}
	if total != 1 || len(assets) != 1 {
		t.Fatalf("expected one backfilled asset, total=%d len=%d", total, len(assets))
	}

	total, assets, err = BackfillWinningCardInstances(context.Background(), db, pkgscope.DefaultScope(1, 10, 88), 1, 3001, 10, cardAssetOperatorJob, "trace-6-repeat")
	if err != nil {
		t.Fatalf("replay backfill returned error: %v", err)
	}
	if total != 1 || len(assets) != 1 {
		t.Fatalf("expected one replayed asset, total=%d len=%d", total, len(assets))
	}

	var count int64
	if err := db.Table(cardInstanceRow{}.TableName()).Count(&count).Error; err != nil {
		t.Fatalf("count instances failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one instance after repeated backfill, got %d", count)
	}
}

func TestBackfillWinningCardInstancesDoesNotProcessNonWinningRecord(t *testing.T) {
	db := newCardAssetTestDB(t)
	if err := db.Exec(`
		INSERT INTO sms_draw_activity (id, platform_id, tenant_id, merchant_id, is_deleted)
		VALUES (1, 1, 10, 88, 0)
	`).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, request_id, scope, result_type, result_status, is_deleted)
		VALUES (1, 1, 3001, 'req-not-win', 'platform:1,tenant:10,merchant:88', 'not_won', 'not_won', 0)
	`).Error; err != nil {
		t.Fatalf("seed participation record failed: %v", err)
	}

	total, assets, err := BackfillWinningCardInstances(context.Background(), db, pkgscope.DefaultScope(1, 10, 88), 1, 3001, 10, cardAssetOperatorJob, "trace-non-win")
	if err != nil {
		t.Fatalf("BackfillWinningCardInstances returned error: %v", err)
	}
	if total != 0 || len(assets) != 0 {
		t.Fatalf("expected non-winning records to be excluded, total=%d len=%d", total, len(assets))
	}
}

func TestEnsureCardInstanceByParticipationRecordLogicRejectsScopeMismatch(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-scope-1")

	logic := NewEnsureCardInstanceByParticipationRecordLogic(context.Background(), &svc.ServiceContext{DB: db})
	_, err := logic.EnsureCardInstanceByParticipationRecord(&smsclient.EnsureCardInstanceByParticipationRecordReq{
		ParticipationRecordId: 1,
		OperatorType:          cardAssetOperatorManual,
		TraceId:               "trace-scope-1",
		Scope:                 newMerchantScope(99),
	})
	if err == nil || !strings.Contains(err.Error(), "当前主体无权操作该资产记录") {
		t.Fatalf("expected scope mismatch error, got %v", err)
	}
}

func TestQueryCardInstanceByParticipationRecordLogicRejectsScopeMismatch(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-scope-2")

	if _, err := EnsureCardInstanceByParticipationRecord(context.Background(), db, 1, cardAssetOperatorSystem, "trace-scope-2"); err != nil {
		t.Fatalf("seed asset instance failed: %v", err)
	}

	logic := NewQueryCardInstanceByParticipationRecordLogic(context.Background(), &svc.ServiceContext{DB: db})
	_, err := logic.QueryCardInstanceByParticipationRecord(&smsclient.QueryCardInstanceByParticipationRecordReq{
		ParticipationRecordId: 1,
		Scope:                 newMerchantScope(99),
	})
	if err == nil || !strings.Contains(err.Error(), "当前主体无权操作该资产记录") {
		t.Fatalf("expected scope mismatch error, got %v", err)
	}
}

func TestBackfillWinningCardInstancesLogicFiltersByScope(t *testing.T) {
	db := newCardAssetTestDB(t)
	seedWinningParticipationRecord(t, db, 1, "req-asset-scope-3")

	logic := NewBackfillWinningCardInstancesLogic(context.Background(), &svc.ServiceContext{DB: db})
	allowed, err := logic.BackfillWinningCardInstances(&smsclient.BackfillWinningCardInstancesReq{
		ActivityId:   1,
		MemberId:     3001,
		Limit:        10,
		OperatorType: cardAssetOperatorJob,
		TraceId:      "trace-scope-3-ok",
		Scope:        newMerchantScope(88),
	})
	if err != nil {
		t.Fatalf("allowed backfill returned error: %v", err)
	}
	if allowed.TotalCandidates != 1 || allowed.ProcessedCount != 1 {
		t.Fatalf("expected allowed scope to process one asset, got %+v", allowed)
	}

	blocked, err := logic.BackfillWinningCardInstances(&smsclient.BackfillWinningCardInstancesReq{
		ActivityId:   1,
		MemberId:     3001,
		Limit:        10,
		OperatorType: cardAssetOperatorJob,
		TraceId:      "trace-scope-3-blocked",
		Scope:        newMerchantScope(99),
	})
	if err != nil {
		t.Fatalf("blocked backfill returned error: %v", err)
	}
	if blocked.TotalCandidates != 0 || blocked.ProcessedCount != 0 || len(blocked.Assets) != 0 {
		t.Fatalf("expected blocked scope to see no candidates, got %+v", blocked)
	}
}
