package drawparticipationservicelogic

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDrawParticipationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "draw-participation.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_code TEXT NOT NULL,
			name TEXT NOT NULL,
			rule_summary TEXT NOT NULL DEFAULT '',
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			real_name_required INTEGER NOT NULL DEFAULT 0,
			participant_condition_summary TEXT NOT NULL DEFAULT '',
			consume_rule_summary TEXT NOT NULL DEFAULT '',
			probability_rule TEXT NOT NULL DEFAULT '',
			compliance_rule_summary TEXT NOT NULL DEFAULT '',
			circulation_limit_summary TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			consume_type TEXT NOT NULL DEFAULT 'lottery_times',
			consume_amount INTEGER NOT NULL DEFAULT 1,
			quota_per_member INTEGER NOT NULL DEFAULT 0,
			daily_quota_per_member INTEGER NOT NULL DEFAULT 0,
			eligibility_rule_json TEXT NOT NULL DEFAULT '',
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_pool (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			pool_name TEXT NOT NULL,
			probability_rule TEXT NOT NULL DEFAULT '',
			sort INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 1,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_pool_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			pool_id INTEGER NOT NULL,
			template_id INTEGER NOT NULL,
			rarity TEXT NOT NULL DEFAULT '',
			probability REAL NOT NULL DEFAULT 1,
			sale_limit INTEGER NOT NULL DEFAULT 1,
			remaining_limit INTEGER NOT NULL DEFAULT 1,
			config_limit INTEGER NOT NULL DEFAULT 1,
			status INTEGER NOT NULL DEFAULT 1,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_code TEXT NOT NULL,
			template_name TEXT NOT NULL,
			card_face_image TEXT NOT NULL DEFAULT '',
			rarity TEXT NOT NULL DEFAULT '',
			display_copy TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE ums_member_info (
			member_id INTEGER PRIMARY KEY,
			lottery_times INTEGER NOT NULL DEFAULT 0,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			nickname TEXT NOT NULL DEFAULT '',
			update_time DATETIME NULL
		)`,
		`CREATE TABLE ums_member_identity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			real_name_status TEXT NOT NULL DEFAULT 'need_real_name',
			real_name_masked TEXT NOT NULL DEFAULT '',
			credential_ref TEXT NOT NULL DEFAULT '',
			verified_at DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_participation_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			request_id TEXT NOT NULL,
			scope TEXT NOT NULL DEFAULT '',
			eligibility_snapshot_json TEXT,
			consume_type TEXT NOT NULL DEFAULT 'lottery_times',
			consume_amount INTEGER NOT NULL DEFAULT 0,
			lottery_times_before INTEGER NOT NULL DEFAULT 0,
			lottery_times_after INTEGER NOT NULL DEFAULT 0,
			result_type TEXT NOT NULL DEFAULT '',
			result_status TEXT NOT NULL DEFAULT '',
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			rarity TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			failure_code TEXT NOT NULL DEFAULT '',
			failure_reason TEXT NOT NULL DEFAULT '',
			asset_instance_id INTEGER NOT NULL DEFAULT 0,
			asset_no TEXT NOT NULL DEFAULT '',
			asset_status TEXT NOT NULL DEFAULT '',
			asset_created_at DATETIME NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

func newDrawParticipationSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{DB: newDrawParticipationTestDB(t)}
}

func testMerchantScope() pkgscope.GovernanceScope {
	scope, _ := pkgscope.NormalizeGovernanceScope(
		pkgscope.SubjectTypeMerchant,
		1,
		10,
		88,
	)
	return scope
}

func TestLoadActivitySnapshotHonorsScope(t *testing.T) {
	db := newDrawParticipationTestDB(t)
	now := time.Now()
	if err := db.Exec(`
		INSERT INTO sms_draw_activity
			(id, activity_code, name, start_time, end_time, platform_id, tenant_id, merchant_id)
		VALUES
			(1, 'DRAW-1', '同域活动', ?, ?, 1, 10, 88),
			(2, 'DRAW-2', '越域活动', ?, ?, 1, 10, 99)
	`, now.Add(-time.Hour), now.Add(time.Hour), now.Add(-time.Hour), now.Add(time.Hour)).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}

	activity, err := loadActivitySnapshot(context.Background(), db, testMerchantScope(), 1, false)
	if err != nil {
		t.Fatalf("loadActivitySnapshot returned error: %v", err)
	}
	if activity.ID != 1 {
		t.Fatalf("expected activity 1, got %d", activity.ID)
	}

	if _, err = loadActivitySnapshot(context.Background(), db, testMerchantScope(), 2, false); err == nil {
		t.Fatal("expected cross-scope activity query to fail")
	}
}

func TestBuildEligibilityDoesNotRequireRealNameForDraw(t *testing.T) {
	activity := &drawActivitySnapshot{
		RealNameRequired:    1,
		Status:              1,
		IsEnabled:           1,
		ConsumeAmount:       1,
		StartTime:           time.Now().Add(-time.Hour),
		EndTime:             time.Now().Add(time.Hour),
		EligibilityRuleJSON: `{"minimumLotteryTimes":3,"requiredRealNameStatus":["verified"]}`,
	}
	member := &drawMemberInfoSnapshot{
		MemberID:     2001,
		LotteryTimes: 2,
		IsEnabled:    1,
	}
	identity := &drawMemberIdentitySnapshot{
		MemberID:       2001,
		RealNameStatus: "pending",
	}

	quotaRejected := buildEligibility(activity, member, identity, 0, 0, true, true)
	if quotaRejected.Code != drawEligibilityQuotaExhausted {
		t.Fatalf("expected quota_exhausted, got %s", quotaRejected.Code)
	}

	member.LotteryTimes = 3
	eligible := buildEligibility(activity, member, identity, 0, 0, true, true)
	if eligible.Code != drawEligibilityEligible {
		t.Fatalf("expected eligible without real-name gate, got %s", eligible.Code)
	}
}

func TestParticipateDrawPassesScopeAndReturnsRejectedWhenRuleFails(t *testing.T) {
	svcCtx := newDrawParticipationSvc(t)
	now := time.Now()
	if err := svcCtx.DB.Exec(`
		INSERT INTO sms_draw_activity
			(id, activity_code, name, start_time, end_time, real_name_required, consume_amount, eligibility_rule_json, platform_id, tenant_id, merchant_id)
		VALUES
			(1, 'DRAW-AC', '抽卡活动', ?, ?, 0, 1, '{"minimumLotteryTimes":3}', 1, 10, 88)
	`, now.Add(-time.Hour), now.Add(time.Hour)).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO ums_member_info (member_id, lottery_times, is_enabled, nickname)
		VALUES (3001, 2, 1, '测试用户')
	`).Error; err != nil {
		t.Fatalf("seed member failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO ums_member_identity (member_id, real_name_status)
		VALUES (3001, 'verified')
	`).Error; err != nil {
		t.Fatalf("seed identity failed: %v", err)
	}

	scope := &smsclient.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 88,
	}
	logic := NewParticipateDrawLogic(context.Background(), svcCtx)
	resp, err := logic.ParticipateDraw(&smsclient.ParticipateDrawReq{
		ActivityId: 1,
		MemberId:   3001,
		RequestId:  "req-10-2",
		Scope:      scope,
	})
	if err != nil {
		t.Fatalf("ParticipateDraw returned error: %v", err)
	}
	if resp.EligibilityCode != drawEligibilityQuotaExhausted {
		t.Fatalf("expected quota_exhausted, got %s", resp.EligibilityCode)
	}
	if resp.Record.ConsumeAmount != 0 {
		t.Fatalf("expected rejected record consume amount 0, got %d", resp.Record.ConsumeAmount)
	}
}

func TestCreateParticipationRecordSetsCreateTime(t *testing.T) {
	db := newDrawParticipationTestDB(t)
	record := &drawParticipationRecordRow{
		ActivityID:    1,
		MemberID:      3001,
		RequestID:     "req-create-time",
		ConsumeType:   drawConsumeTypeLottery,
		ResultType:    drawResultTypeRejected,
		ResultStatus:  drawResultStatusQuota,
		FailureCode:   drawEligibilityQuotaExhausted,
		FailureReason: "剩余抽奖次数不足",
	}

	if err := createParticipationRecord(context.Background(), db, record); err != nil {
		t.Fatalf("createParticipationRecord returned error: %v", err)
	}
	if record.CreateTime.IsZero() {
		t.Fatal("expected create time to be set before insert")
	}
}

func TestParticipateDrawCreatesAssetSnapshotForWinningRecordAndKeepsRequestIdempotent(t *testing.T) {
	svcCtx := newDrawParticipationSvc(t)
	now := time.Now()
	if err := svcCtx.DB.Exec(`
		INSERT INTO sms_draw_activity
			(id, activity_code, name, start_time, end_time, real_name_required, consume_amount, platform_id, tenant_id, merchant_id)
		VALUES
			(1, 'DRAW-WIN', '抽卡活动', ?, ?, 1, 1, 1, 10, 88)
	`, now.Add(-time.Hour), now.Add(time.Hour)).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO ums_member_info (member_id, lottery_times, is_enabled, nickname)
		VALUES (3001, 3, 1, '测试用户')
	`).Error; err != nil {
		t.Fatalf("seed member failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO ums_member_identity (member_id, real_name_status)
		VALUES (3001, 'pending')
	`).Error; err != nil {
		t.Fatalf("seed identity failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO sms_draw_pool (id, activity_id, pool_name, probability_rule, sort, status, is_deleted)
		VALUES (11, 1, 'SSR池', '固定中签', 1, 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed pool failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO sms_card_template (id, template_code, template_name, card_face_image, rarity, display_copy, status, is_deleted)
		VALUES (22, 'CARD-SSR', 'SSR卡', '', 'SSR', '恭喜中签', 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed template failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`
		INSERT INTO sms_draw_pool_template
			(id, activity_id, pool_id, template_id, rarity, probability, sale_limit, remaining_limit, config_limit, status, is_deleted)
		VALUES
			(33, 1, 11, 22, 'SSR', 1, 10, 10, 10, 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed pool template failed: %v", err)
	}

	scope := &smsclient.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 88,
	}
	logic := NewParticipateDrawLogic(context.Background(), svcCtx)
	first, err := logic.ParticipateDraw(&smsclient.ParticipateDrawReq{
		ActivityId: 1,
		MemberId:   3001,
		RequestId:  "req-10-3",
		Scope:      scope,
	})
	if err != nil {
		t.Fatalf("ParticipateDraw returned error: %v", err)
	}
	if first.Record.AssetInstanceId <= 0 {
		t.Fatalf("expected asset instance id on winning response, got %+v", first.Record)
	}
	if first.Record.AssetNo == "" {
		t.Fatalf("expected asset no on winning response, got %+v", first.Record)
	}
	if first.Record.AssetStatus != "asset_created" {
		t.Fatalf("expected asset_created, got %s", first.Record.AssetStatus)
	}

	second, err := logic.ParticipateDraw(&smsclient.ParticipateDrawReq{
		ActivityId: 1,
		MemberId:   3001,
		RequestId:  "req-10-3",
		Scope:      scope,
	})
	if err != nil {
		t.Fatalf("second ParticipateDraw returned error: %v", err)
	}
	if first.Record.AssetInstanceId != second.Record.AssetInstanceId || first.Record.AssetNo != second.Record.AssetNo {
		t.Fatalf("expected repeated request to return same asset snapshot, first=%+v second=%+v", first.Record, second.Record)
	}
}
