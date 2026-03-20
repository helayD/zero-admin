package noticeservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newNoticeTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "notice-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sys_tenant (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE sys_notice (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			notice_title TEXT NOT NULL,
			notice_type INTEGER NOT NULL DEFAULT 1,
			notice_content TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			remark TEXT NOT NULL DEFAULT '',
			create_by TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by TEXT NOT NULL DEFAULT '',
			update_time DATETIME NULL,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0
		)`,
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	return &svc.ServiceContext{DB: db}
}

func TestAddNoticePersistsGovernanceScope(t *testing.T) {
	svcCtx := newNoticeTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(`INSERT INTO sys_tenant (id, status) VALUES (10, ?)`, tenantmodel.TenantStatusEnabled).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}

	logic := NewAddNoticeLogic(context.Background(), svcCtx)
	_, err := logic.AddNotice(&sysclient.AddNoticeReq{
		NoticeTitle:   "租户运营通知",
		NoticeType:    1,
		NoticeContent: "请完成租户初始化检查",
		Status:        1,
		Remark:        "story-1.3",
		CreateBy:      "tester",
		Scope: &sysclient.GovernanceScope{
			ScopeType:  scope.SubjectTypeTenant,
			PlatformId: 1,
			TenantId:   10,
		},
	})
	if err != nil {
		t.Fatalf("add notice failed: %v", err)
	}

	var row struct {
		PlatformID int64  `gorm:"column:platform_id"`
		TenantID   int64  `gorm:"column:tenant_id"`
		MerchantID int64  `gorm:"column:merchant_id"`
		CreateBy   string `gorm:"column:create_by"`
	}
	if err := db.Table("sys_notice").
		Select("platform_id, tenant_id, merchant_id, create_by").
		Where("notice_title = ?", "租户运营通知").
		Take(&row).Error; err != nil {
		t.Fatalf("load notice failed: %v", err)
	}

	if row.PlatformID != 1 || row.TenantID != 10 || row.MerchantID != 0 {
		t.Fatalf("unexpected notice scope: %+v", row)
	}
	if row.CreateBy != "tester" {
		t.Fatalf("expected create_by tester, got %q", row.CreateBy)
	}
}

func TestAddNoticeRejectsDisabledTenantScope(t *testing.T) {
	svcCtx := newNoticeTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(`INSERT INTO sys_tenant (id, status) VALUES (11, ?)`, tenantmodel.TenantStatusDisabled).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}

	logic := NewAddNoticeLogic(context.Background(), svcCtx)
	_, err := logic.AddNotice(&sysclient.AddNoticeReq{
		NoticeTitle:   "失效租户通知",
		NoticeType:    2,
		NoticeContent: "不应写入",
		Status:        1,
		CreateBy:      "tester",
		Scope: &sysclient.GovernanceScope{
			ScopeType:  scope.SubjectTypeTenant,
			PlatformId: 1,
			TenantId:   11,
		},
	})
	if err == nil {
		t.Fatal("expected disabled tenant validation to fail")
	}
	if !strings.Contains(err.Error(), "已停用或已归档租户") {
		t.Fatalf("expected disabled tenant error, got %v", err)
	}

	var count int64
	if err := db.Table("sys_notice").Count(&count).Error; err != nil {
		t.Fatalf("count notices failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no notice rows to be written, got %d", count)
	}
}

func TestQueryNoticeListFiltersByScope(t *testing.T) {
	svcCtx := newNoticeTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(
		`INSERT INTO sys_notice (id, notice_title, notice_type, notice_content, status, remark, create_by, update_by, platform_id, tenant_id, merchant_id)
		 VALUES (1, '平台通知', 1, 'platform', 1, '', 'seed', 'seed', 1, 0, 0),
		        (2, '租户通知', 2, 'tenant', 1, '', 'seed', 'seed', 1, 10, 0)`,
	).Error; err != nil {
		t.Fatalf("seed notices failed: %v", err)
	}

	logic := NewQueryNoticeListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryNoticeList(&sysclient.QueryNoticeListReq{
		PageNum:  1,
		PageSize: 10,
		Status:   2,
		Scope: &sysclient.GovernanceScope{
			ScopeType:  scope.SubjectTypeTenant,
			PlatformId: 1,
			TenantId:   10,
		},
	})
	if err != nil {
		t.Fatalf("query notice list failed: %v", err)
	}

	if resp.Total != 1 {
		t.Fatalf("expected 1 scoped notice, got %d", resp.Total)
	}
	if len(resp.List) != 1 {
		t.Fatalf("expected 1 notice row, got %d", len(resp.List))
	}
	if resp.List[0].NoticeTitle != "租户通知" {
		t.Fatalf("expected tenant notice, got %q", resp.List[0].NoticeTitle)
	}
	if resp.List[0].Scope == nil || resp.List[0].Scope.ScopeLabel == "" {
		t.Fatalf("expected scope label in response, got %+v", resp.List[0].Scope)
	}
}
