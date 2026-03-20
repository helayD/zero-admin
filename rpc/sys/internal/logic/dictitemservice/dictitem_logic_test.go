package dictitemservicelogic

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDictItemTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "dict-item-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sys_tenant (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE sys_dict_type (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dict_name TEXT NOT NULL,
			dict_type TEXT NOT NULL,
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
		`CREATE TABLE sys_dict_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dict_sort INTEGER NOT NULL,
			dict_label TEXT NOT NULL,
			dict_value TEXT NOT NULL,
			dict_type TEXT NOT NULL DEFAULT '',
			css_class TEXT NOT NULL DEFAULT '',
			list_class TEXT NOT NULL DEFAULT '',
			is_default TEXT NOT NULL DEFAULT 'N',
			status INTEGER NOT NULL DEFAULT 1,
			remark TEXT NOT NULL DEFAULT '',
			create_by TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by TEXT NOT NULL DEFAULT '',
			update_time DATETIME NULL,
			dict_type_id INTEGER NOT NULL DEFAULT 0,
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

func TestAddDictItemResolvesScopedDictTypeByCode(t *testing.T) {
	svcCtx := newDictItemTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(
		`INSERT INTO sys_tenant (id, status) VALUES (10, 1)`,
	).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO sys_dict_type (id, dict_name, dict_type, status, create_by, update_by, platform_id, tenant_id, merchant_id)
		 VALUES (1, '平台订单状态', 'shared_status', 1, 'seed', 'seed', 1, 0, 0),
		        (2, '租户订单状态', 'shared_status', 1, 'seed', 'seed', 1, 10, 0)`,
	).Error; err != nil {
		t.Fatalf("seed dict types failed: %v", err)
	}

	logic := NewAddDictItemLogic(context.Background(), svcCtx)
	_, err := logic.AddDictItem(&sysclient.AddDictItemReq{
		DictSort:  1,
		DictLabel: "待处理",
		DictValue: "pending",
		DictType:  "shared_status",
		IsDefault: "N",
		Status:    1,
		CreateBy:  "tester",
		Scope: &sysclient.GovernanceScope{
			ScopeType:  scope.SubjectTypeTenant,
			PlatformId: 1,
			TenantId:   10,
		},
	})
	if err != nil {
		t.Fatalf("add dict item failed: %v", err)
	}

	var row struct {
		DictTypeID int64  `gorm:"column:dict_type_id"`
		DictType   string `gorm:"column:dict_type"`
		PlatformID int64  `gorm:"column:platform_id"`
		TenantID   int64  `gorm:"column:tenant_id"`
	}
	if err := db.Table("sys_dict_item").
		Select("dict_type_id, dict_type, platform_id, tenant_id").
		Where("dict_label = ?", "待处理").
		Take(&row).Error; err != nil {
		t.Fatalf("load dict item failed: %v", err)
	}

	if row.DictTypeID != 2 {
		t.Fatalf("expected tenant dict type id 2, got %d", row.DictTypeID)
	}
	if row.DictType != "shared_status" {
		t.Fatalf("expected dict type shared_status, got %q", row.DictType)
	}
	if row.PlatformID != 1 || row.TenantID != 10 {
		t.Fatalf("unexpected item scope: %+v", row)
	}
}

func TestAddDictItemUsesResolvedDictTypeWhenOnlyIDProvided(t *testing.T) {
	svcCtx := newDictItemTestSvc(t)
	db := svcCtx.DB

	if err := db.Exec(
		`INSERT INTO sys_tenant (id, status) VALUES (10, 1)`,
	).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO sys_dict_type (id, dict_name, dict_type, status, create_by, update_by, platform_id, tenant_id, merchant_id)
		 VALUES (11, '支付状态', 'order_status', 1, 'seed', 'seed', 1, 10, 0)`,
	).Error; err != nil {
		t.Fatalf("seed dict type failed: %v", err)
	}

	logic := NewAddDictItemLogic(context.Background(), svcCtx)
	_, err := logic.AddDictItem(&sysclient.AddDictItemReq{
		DictSort:   2,
		DictLabel:  "已支付",
		DictValue:  "paid",
		DictTypeId: 11,
		IsDefault:  "N",
		Status:     1,
		CreateBy:   "tester",
		Scope: &sysclient.GovernanceScope{
			ScopeType:  scope.SubjectTypeTenant,
			PlatformId: 1,
			TenantId:   10,
		},
	})
	if err != nil {
		t.Fatalf("add dict item by id failed: %v", err)
	}

	var row struct {
		DictType string `gorm:"column:dict_type"`
	}
	if err := db.Table("sys_dict_item").
		Select("dict_type").
		Where("dict_label = ?", "已支付").
		Take(&row).Error; err != nil {
		t.Fatalf("load dict item failed: %v", err)
	}

	if row.DictType != "order_status" {
		t.Fatalf("expected dict type order_status, got %q", row.DictType)
	}
}
