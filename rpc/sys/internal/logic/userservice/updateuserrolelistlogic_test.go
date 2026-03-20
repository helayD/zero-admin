package userservicelogic

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

func newUserRoleTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "user-role-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sys_user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			mobile TEXT NOT NULL,
			user_name TEXT NOT NULL,
			nick_name TEXT NOT NULL,
			user_type TEXT NOT NULL DEFAULT '00',
			avatar TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			dept_id INTEGER NOT NULL DEFAULT 0,
			login_ip TEXT NOT NULL DEFAULT '',
			login_date DATETIME NULL,
			login_browser TEXT NOT NULL DEFAULT '',
			login_os TEXT NOT NULL DEFAULT '',
			pwd_update_date DATETIME NULL,
			remark TEXT NOT NULL DEFAULT '',
			del_flag INTEGER NOT NULL DEFAULT 1,
			create_by TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by TEXT NOT NULL DEFAULT '',
			update_time DATETIME NULL,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sys_role (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_name TEXT NOT NULL,
			role_key TEXT NOT NULL,
			scope_type TEXT NOT NULL DEFAULT 'platform',
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			is_admin INTEGER NOT NULL DEFAULT 0,
			data_scope INTEGER NOT NULL DEFAULT 1,
			status INTEGER NOT NULL DEFAULT 1,
			remark TEXT NOT NULL DEFAULT '',
			del_flag INTEGER NOT NULL DEFAULT 1,
			create_by TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by TEXT NOT NULL DEFAULT '',
			update_time DATETIME NULL
		)`,
		`CREATE TABLE sys_user_role (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL
		)`,
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	query.SetDefault(db)
	return &svc.ServiceContext{DB: db}
}

func TestUpdateUserRoleListRejectsCrossScopeRoleAssignment(t *testing.T) {
	svcCtx := newUserRoleTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(
		`INSERT INTO sys_user (id, mobile, user_name, nick_name, email, password, status, dept_id, create_by, update_by, platform_id, tenant_id, merchant_id)
		 VALUES (11, '13800138000', 'tenant_user', '租户用户', 'tenant@example.com', '123456', 1, 1, 'seed', 'seed', 1, 10, 0)`,
	).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO sys_role (id, role_name, role_key, scope_type, platform_id, tenant_id, merchant_id, status, del_flag, create_by, update_by)
		 VALUES (21, '平台角色', 'platform-role', 'platform', 1, 0, 0, 1, 1, 'seed', 'seed')`,
	).Error; err != nil {
		t.Fatalf("seed role failed: %v", err)
	}

	logic := NewUpdateUserRoleListLogic(context.Background(), svcCtx)
	_, err := logic.UpdateUserRoleList(&sysclient.UpdateUserRoleListReq{
		UserId:  11,
		RoleIds: []int64{21},
	})
	if err == nil {
		t.Fatal("expected cross-scope assignment to fail")
	}
	if !strings.Contains(err.Error(), "主体范围") {
		t.Fatalf("expected scope validation error, got %v", err)
	}

	var count int64
	if err := db.Table("sys_user_role").Count(&count).Error; err != nil {
		t.Fatalf("count user roles failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no user-role bindings to be created, got %d", count)
	}
}
