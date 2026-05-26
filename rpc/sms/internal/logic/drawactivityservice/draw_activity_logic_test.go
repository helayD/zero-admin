package drawactivityservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDrawActivityTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "draw-activity.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL,
			activity_code TEXT NOT NULL,
			name TEXT NOT NULL,
			rule_summary TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			real_name_required INTEGER NOT NULL DEFAULT 0,
			participant_condition_summary TEXT NOT NULL DEFAULT '',
			consume_rule_summary TEXT NOT NULL DEFAULT '',
			probability_rule TEXT NOT NULL DEFAULT '',
			compliance_rule_summary TEXT NOT NULL DEFAULT '',
			circulation_limit_summary TEXT NOT NULL DEFAULT '',
			approval_record_ref TEXT NOT NULL DEFAULT '',
			publish_failure_summary TEXT NOT NULL DEFAULT '',
			publish_readiness INTEGER NOT NULL DEFAULT 0,
			copyright_status INTEGER NOT NULL DEFAULT 0,
			content_audit_status INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			audit_status INTEGER NOT NULL DEFAULT 0,
			consume_type TEXT NOT NULL DEFAULT 'points',
			consume_amount INTEGER NOT NULL DEFAULT 1,
			quota_per_member INTEGER NOT NULL DEFAULT 0,
			daily_quota_per_member INTEGER NOT NULL DEFAULT 0,
			eligibility_rule_json TEXT NOT NULL DEFAULT '',
			is_enabled INTEGER NOT NULL DEFAULT 1,
			show_on_home INTEGER NOT NULL DEFAULT 0,
			home_entry_title TEXT NOT NULL DEFAULT '',
			home_entry_subtitle TEXT NOT NULL DEFAULT '',
			home_entry_image TEXT NOT NULL DEFAULT '',
			home_entry_sort INTEGER NOT NULL DEFAULT 0,
			home_entry_enabled INTEGER NOT NULL DEFAULT 1,
			landing_target_type TEXT NOT NULL DEFAULT '',
			landing_target_value TEXT NOT NULL DEFAULT '',
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL,
			update_by INTEGER,
			update_time DATETIME,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_pool (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL,
			pool_code TEXT NOT NULL,
			pool_name TEXT NOT NULL,
			probability_rule TEXT NOT NULL DEFAULT '',
			wheel_slot_count INTEGER NOT NULL DEFAULT 5,
			sort INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			audit_status INTEGER NOT NULL DEFAULT 0,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL,
			update_by INTEGER,
			update_time DATETIME,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL,
			template_code TEXT NOT NULL,
			template_name TEXT NOT NULL,
			card_face_image TEXT NOT NULL DEFAULT '',
			copyright_owner TEXT NOT NULL DEFAULT '',
			copyright_proof_summary TEXT NOT NULL DEFAULT '',
			rarity TEXT NOT NULL DEFAULT '',
			issue_limit INTEGER NOT NULL DEFAULT 0,
			display_copy TEXT NOT NULL DEFAULT '',
			circulation_limit_summary TEXT NOT NULL DEFAULT '',
			display_status INTEGER NOT NULL DEFAULT 1,
			content_audit_status INTEGER NOT NULL DEFAULT 0,
			provider_code TEXT NOT NULL DEFAULT '',
			credential_ref TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 0,
			audit_status INTEGER NOT NULL DEFAULT 0,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL,
			update_by INTEGER,
			update_time DATETIME,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_pool_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			pool_id INTEGER NOT NULL,
			template_id INTEGER NOT NULL,
			slot_index INTEGER NOT NULL DEFAULT 0,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL,
			rarity TEXT NOT NULL DEFAULT '',
			probability REAL NOT NULL DEFAULT 0,
			sale_limit INTEGER NOT NULL DEFAULT 0,
			remaining_limit INTEGER NOT NULL DEFAULT 0,
			config_limit INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			audit_status INTEGER NOT NULL DEFAULT 0,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL,
			update_by INTEGER,
			update_time DATETIME,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_activity_audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL,
			operation_type TEXT NOT NULL,
			operator_id INTEGER NOT NULL DEFAULT 0,
			operator_name TEXT NOT NULL DEFAULT '',
			approval_result TEXT NOT NULL DEFAULT '',
			rule_snapshot TEXT,
			circulation_limit_snapshot TEXT,
			failure_summary TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			payload_json TEXT,
			status INTEGER NOT NULL DEFAULT 0,
			audit_status INTEGER NOT NULL DEFAULT 0,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL
		)`,
		`CREATE TABLE sys_security_event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trace_id TEXT,
			event_type TEXT,
			action TEXT,
			resource_type TEXT,
			resource_id INTEGER,
			scope_type TEXT,
			platform_id INTEGER,
			tenant_id INTEGER,
			merchant_id INTEGER,
			operator_id INTEGER,
			operator_name TEXT,
			request_summary TEXT,
			result TEXT,
			payload TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE sys_user (
			id INTEGER PRIMARY KEY,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}
	if err := db.Exec(`
		INSERT INTO sys_user (id, platform_id, tenant_id, merchant_id) VALUES
		(1001, 1, 10, 88),
		(1002, 1, 10, 99)
	`).Error; err != nil {
		t.Fatalf("seed sys_user failed: %v", err)
	}

	return db
}

func newDrawActivityLogicSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{DB: newDrawActivityTestDB(t)}
}

func merchantScope() *smsclient.GovernanceScope {
	return &smsclient.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 88,
	}
}

func validDrawAddReq() *smsclient.AddDrawActivityReq {
	return &smsclient.AddDrawActivityReq{
		Scope:                       merchantScope(),
		ActivityCode:                "DRAW-20260416",
		Name:                        "春季提货卡抽赏",
		RuleSummary:                 "完成实名后可参与，每次消耗 10 积分",
		StartTime:                   "2026-04-16 10:00:00",
		EndTime:                     "2026-04-30 23:59:59",
		RealNameRequired:            1,
		ParticipantConditionSummary: "完成实名校验且账号正常",
		ConsumeRuleSummary:          "每次抽卡消耗 10 积分",
		ProbabilityRule:             "所有概率使用 0-1 小数表达，总和必须为 1",
		ComplianceRuleSummary:       "活动文案与资格规则已通过法务审阅",
		CirculationLimitSummary:     "禁止集中竞价；禁止连续挂牌；禁止收益承诺",
		ApprovalRecordRef:           "APPROVAL-1001",
		CopyrightStatus:             2,
		ContentAuditStatus:          2,
		AuditStatus:                 2,
		Status:                      0,
		IsEnabled:                   1,
		OperatorName:                "tester",
		CreateBy:                    1001,
		HomeEntry: &smsclient.DrawHomeEntryConfig{
			ShowOnHome:         1,
			HomeEntryTitle:     "春季抽卡",
			HomeEntrySubtitle:  "首页显著入口",
			HomeEntryImage:     "https://cdn.example.com/draw.png",
			HomeEntrySort:      10,
			LandingTargetType:  "activity",
			LandingTargetValue: "DRAW-20260416",
			IsEnabled:          1,
		},
		Templates: []*smsclient.DrawCardTemplateData{
			{TemplateName: "SSR 兔兔", TemplateCode: "TPL-SSR-01", CopyrightOwner: "Zero Admin", CopyrightProofSummary: "版权登记号 2026-01", Rarity: "SSR", IssueLimit: 100, DisplayCopy: "限定 SSR", CirculationLimitSummary: "禁止集中竞价；禁止连续挂牌；禁止收益承诺", ContentAuditStatus: 2, DisplayStatus: 1, Status: 0, AuditStatus: 2},
			{TemplateName: "SR 火雉", TemplateCode: "TPL-SR-01", CopyrightOwner: "Zero Admin", CopyrightProofSummary: "版权登记号 2026-02", Rarity: "SR", IssueLimit: 200, DisplayCopy: "SR", CirculationLimitSummary: "禁止集中竞价；禁止连续挂牌；禁止收益承诺", ContentAuditStatus: 2, DisplayStatus: 1, Status: 0, AuditStatus: 2},
			{TemplateName: "R 小熊", TemplateCode: "TPL-R-01", CopyrightOwner: "Zero Admin", CopyrightProofSummary: "版权登记号 2026-03", Rarity: "R", IssueLimit: 300, DisplayCopy: "R", CirculationLimitSummary: "禁止集中竞价；禁止连续挂牌；禁止收益承诺", ContentAuditStatus: 2, DisplayStatus: 1, Status: 0, AuditStatus: 2},
			{TemplateName: "N 普通1", TemplateCode: "TPL-N-01", CopyrightOwner: "Zero Admin", CopyrightProofSummary: "版权登记号 2026-04", Rarity: "N", IssueLimit: 400, DisplayCopy: "N", CirculationLimitSummary: "禁止集中竞价；禁止连续挂牌；禁止收益承诺", ContentAuditStatus: 2, DisplayStatus: 1, Status: 0, AuditStatus: 2},
			{TemplateName: "N 普通2", TemplateCode: "TPL-N-02", CopyrightOwner: "Zero Admin", CopyrightProofSummary: "版权登记号 2026-05", Rarity: "N", IssueLimit: 400, DisplayCopy: "N", CirculationLimitSummary: "禁止集中竞价；禁止连续挂牌；禁止收益承诺", ContentAuditStatus: 2, DisplayStatus: 1, Status: 0, AuditStatus: 2},
		},
		Pools: []*smsclient.DrawPoolData{
			{
				PoolName:        "限定池",
				PoolCode:        "POOL-SSR",
				ProbabilityRule: "概率总和必须为 1",
				Templates: []*smsclient.DrawPoolTemplateData{
					{TemplateCode: "TPL-SSR-01", TemplateName: "SSR 兔兔", Rarity: "SSR", Probability: 0.1, SaleLimit: 100, RemainingLimit: 100, ConfigLimit: 100, SlotIndex: 1},
					{TemplateCode: "TPL-SR-01", TemplateName: "SR 火雉", Rarity: "SR", Probability: 0.2, SaleLimit: 200, RemainingLimit: 200, ConfigLimit: 200, SlotIndex: 2},
					{TemplateCode: "TPL-R-01", TemplateName: "R 小熊", Rarity: "R", Probability: 0.3, SaleLimit: 300, RemainingLimit: 300, ConfigLimit: 300, SlotIndex: 3},
					{TemplateCode: "TPL-N-01", TemplateName: "N 普通1", Rarity: "N", Probability: 0.2, SaleLimit: 400, RemainingLimit: 400, ConfigLimit: 400, SlotIndex: 4},
					{TemplateCode: "TPL-N-02", TemplateName: "N 普通2", Rarity: "N", Probability: 0.2, SaleLimit: 400, RemainingLimit: 400, ConfigLimit: 400, SlotIndex: 5},
				},
			},
		},
	}
}

func TestAddDrawActivityRejectsInvalidSlotIndex(t *testing.T) {
	logic := NewAddDrawActivityLogic(context.Background(), newDrawActivityLogicSvc(t))
	req := validDrawAddReq()
	req.Pools[0].Templates[0].SlotIndex = 6

	_, err := logic.AddDrawActivity(req)
	if err == nil || !strings.Contains(err.Error(), "格位序号必须在 1-5") {
		t.Fatalf("expected slot_index out of range error, got %v", err)
	}
}

func TestAddDrawActivityRejectsDuplicateSlotIndex(t *testing.T) {
	logic := NewAddDrawActivityLogic(context.Background(), newDrawActivityLogicSvc(t))
	req := validDrawAddReq()
	req.Pools[0].Templates[1].SlotIndex = 1

	_, err := logic.AddDrawActivity(req)
	if err == nil || !strings.Contains(err.Error(), "格位序号 1 重复") {
		t.Fatalf("expected duplicate slot_index error, got %v", err)
	}
}

func TestAddDrawActivityRejectsInvalidProbabilitySum(t *testing.T) {
	logic := NewAddDrawActivityLogic(context.Background(), newDrawActivityLogicSvc(t))
	req := validDrawAddReq()
	req.Pools[0].Templates[0].Probability = 0.9

	_, err := logic.AddDrawActivity(req)
	if err == nil || !strings.Contains(err.Error(), "概率总和必须为1") {
		t.Fatalf("expected invalid probability error, got %v", err)
	}
}

func TestAddDrawActivityRejectsCrossScopeWrite(t *testing.T) {
	logic := NewAddDrawActivityLogic(context.Background(), newDrawActivityLogicSvc(t))
	req := validDrawAddReq()
	req.CreateBy = 1002

	_, err := logic.AddDrawActivity(req)
	if err == nil || !strings.Contains(err.Error(), "当前主体不允许切换写入范围") {
		t.Fatalf("expected cross-scope rejection, got %v", err)
	}
}

func TestPreviewDrawActivityPersistsStructuredFailures(t *testing.T) {
	svcCtx := newDrawActivityLogicSvc(t)
	addLogic := NewAddDrawActivityLogic(context.Background(), svcCtx)
	req := validDrawAddReq()
	req.ApprovalRecordRef = ""
	req.AuditStatus = 0
	req.CopyrightStatus = 0

	resp, err := addLogic.AddDrawActivity(req)
	if err != nil {
		t.Fatalf("AddDrawActivity returned error: %v", err)
	}

	previewLogic := NewPreviewDrawActivityPublishReadinessLogic(context.Background(), svcCtx)
	previewResp, err := previewLogic.PreviewDrawActivityPublishReadiness(&smsclient.PreviewDrawActivityPublishReadinessReq{
		Id:    resp.Id,
		Scope: merchantScope(),
	})
	if err != nil {
		t.Fatalf("PreviewDrawActivityPublishReadiness returned error: %v", err)
	}
	if previewResp.ReadyToPublish {
		t.Fatal("expected readiness preview to block publish")
	}
	if len(previewResp.Items) == 0 {
		t.Fatal("expected structured readiness items")
	}

	var activity struct {
		PublishReadiness      int32  `gorm:"column:publish_readiness"`
		PublishFailureSummary string `gorm:"column:publish_failure_summary"`
	}
	if err := svcCtx.DB.Table("sms_draw_activity").Where("id = ?", resp.Id).Take(&activity).Error; err != nil {
		t.Fatalf("query activity failed: %v", err)
	}
	if activity.PublishReadiness != 0 {
		t.Fatalf("expected persisted readiness=0, got %d", activity.PublishReadiness)
	}
	if !strings.Contains(activity.PublishFailureSummary, "审批记录") {
		t.Fatalf("expected failure summary to mention approval record, got %s", activity.PublishFailureSummary)
	}
}

func TestUpdateDrawActivityStatusBlocksPublishWhenPreviewFails(t *testing.T) {
	svcCtx := newDrawActivityLogicSvc(t)
	addLogic := NewAddDrawActivityLogic(context.Background(), svcCtx)
	req := validDrawAddReq()
	req.ContentAuditStatus = 1

	resp, err := addLogic.AddDrawActivity(req)
	if err != nil {
		t.Fatalf("AddDrawActivity returned error: %v", err)
	}

	statusLogic := NewUpdateDrawActivityStatusLogic(context.Background(), svcCtx)
	_, err = statusLogic.UpdateDrawActivityStatus(&smsclient.UpdateDrawActivityStatusReq{
		Ids:          []int64{resp.Id},
		Status:       1,
		Scope:        merchantScope(),
		UpdateBy:     1001,
		OperatorName: "tester",
	})
	if err == nil || !strings.Contains(err.Error(), "内容审核") {
		t.Fatalf("expected content audit blocking error, got %v", err)
	}
}

func TestAddAndUpdateAppendAuditRecords(t *testing.T) {
	svcCtx := newDrawActivityLogicSvc(t)
	addLogic := NewAddDrawActivityLogic(context.Background(), svcCtx)
	addResp, err := addLogic.AddDrawActivity(validDrawAddReq())
	if err != nil {
		t.Fatalf("AddDrawActivity returned error: %v", err)
	}

	updateLogic := NewUpdateDrawActivityLogic(context.Background(), svcCtx)
	_, err = updateLogic.UpdateDrawActivity(&smsclient.UpdateDrawActivityReq{
		Id:                          addResp.Id,
		Scope:                       merchantScope(),
		ActivityCode:                "DRAW-20260416",
		Name:                        "春季提货卡抽赏-更新",
		RuleSummary:                 "规则更新",
		StartTime:                   "2026-04-16 10:00:00",
		EndTime:                     "2026-04-30 23:59:59",
		RealNameRequired:            1,
		ParticipantConditionSummary: "完成实名校验且账号正常",
		ConsumeRuleSummary:          "每次抽卡消耗 10 积分",
		ProbabilityRule:             "所有概率使用 0-1 小数表达，总和必须为 1",
		ComplianceRuleSummary:       "活动文案与资格规则已通过法务审阅",
		CirculationLimitSummary:     "禁止集中竞价；禁止连续挂牌；禁止收益承诺",
		ApprovalRecordRef:           "APPROVAL-1001",
		CopyrightStatus:             2,
		ContentAuditStatus:          2,
		AuditStatus:                 2,
		Status:                      0,
		IsEnabled:                   1,
		OperatorName:                "tester",
		UpdateBy:                    1001,
		HomeEntry:                   validDrawAddReq().HomeEntry,
		Templates:                   validDrawAddReq().Templates,
		Pools:                       validDrawAddReq().Pools,
	})
	if err != nil {
		t.Fatalf("UpdateDrawActivity returned error: %v", err)
	}

	var count int64
	if err := svcCtx.DB.Table("sms_draw_activity_audit").Where("activity_id = ?", addResp.Id).Count(&count).Error; err != nil {
		t.Fatalf("count audit records failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 audit records, got %d", count)
	}
}
