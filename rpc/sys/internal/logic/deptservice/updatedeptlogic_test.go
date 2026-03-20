package deptservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/rpc/sys/gen/query"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDeptTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "dept-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sys_tenant (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE sys_dept (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		parent_id INTEGER NOT NULL DEFAULT 0,
		ancestors TEXT NOT NULL DEFAULT '0',
		dept_name TEXT NOT NULL,
		sort INTEGER NOT NULL DEFAULT 1,
		leader TEXT NOT NULL DEFAULT '',
		phone TEXT NOT NULL DEFAULT '',
		email TEXT NOT NULL DEFAULT '',
		status INTEGER NOT NULL DEFAULT 1,
		del_flag INTEGER NOT NULL DEFAULT 1,
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
			t.Fatalf("create test schema failed: %v", err)
		}
	}

	query.SetDefault(db)
	return &svc.ServiceContext{DB: db}
}

func TestUpdateDeptRejectsMovingDeptUnderItsChild(t *testing.T) {
	svcCtx := newDeptTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(
		`INSERT INTO sys_tenant (id, status) VALUES (10, 1)`,
	).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO sys_dept (id, parent_id, ancestors, dept_name, sort, status, del_flag, create_by, update_by, platform_id, tenant_id, merchant_id)
		 VALUES (1, 0, '0', '总部', 1, 1, 1, 'seed', 'seed', 1, 10, 0),
		        (2, 1, '0,1', '运营部', 1, 1, 1, 'seed', 'seed', 1, 10, 0)`,
	).Error; err != nil {
		t.Fatalf("seed departments failed: %v", err)
	}

	logic := NewUpdateDeptLogic(context.Background(), svcCtx)
	_, err := logic.UpdateDept(&sysclient.UpdateDeptReq{
		Id:       1,
		ParentId: 2,
		DeptName: "总部",
		Sort:     1,
		Status:   1,
		UpdateBy: "tester",
		Scope: &sysclient.GovernanceScope{
			ScopeType:  "tenant",
			PlatformId: 1,
			TenantId:   10,
		},
	})
	if err == nil {
		t.Fatal("expected dept cycle validation to fail")
	}
	if !strings.Contains(err.Error(), "子部门") {
		t.Fatalf("expected child-dept validation error, got %v", err)
	}
}
