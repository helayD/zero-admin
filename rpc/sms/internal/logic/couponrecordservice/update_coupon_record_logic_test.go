package couponrecordservicelogic

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newUpdateCouponRecordTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "coupon-record.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_coupon_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			coupon_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			get_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			get_type INTEGER NOT NULL DEFAULT 1,
			use_time DATETIME NULL,
			order_id INTEGER NOT NULL DEFAULT 0,
			order_amount REAL NOT NULL DEFAULT 0,
			discount_amount REAL NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 0,
			invalid_time DATETIME NULL,
			invalid_reason TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_coupon (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type_id INTEGER NOT NULL DEFAULT 1,
			name TEXT NOT NULL DEFAULT '',
			code TEXT NOT NULL DEFAULT '',
			amount REAL NOT NULL DEFAULT 0,
			min_amount REAL NOT NULL DEFAULT 0,
			start_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			end_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			total_count INTEGER NOT NULL DEFAULT 100,
			received_count INTEGER NOT NULL DEFAULT 1,
			used_count INTEGER NOT NULL DEFAULT 0,
			per_limit INTEGER NOT NULL DEFAULT 1,
			status INTEGER NOT NULL DEFAULT 1,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			description TEXT NOT NULL DEFAULT '',
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by INTEGER NULL,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`INSERT INTO sms_coupon (id, used_count) VALUES (1, 0)`,
		`INSERT INTO sms_coupon_record (id, coupon_id, member_id, status) VALUES (5, 1, 4, 0)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}
	query.SetDefault(db)

	return &svc.ServiceContext{DB: db}
}

func TestUpdateCouponRecordUsesRecordIDCondition(t *testing.T) {
	svcCtx := newUpdateCouponRecordTestSvc(t)

	_, err := NewUpdateCouponRecordLogic(context.Background(), svcCtx).UpdateCouponRecord(&smsclient.UpdateCouponRecordReq{
		CouponIds: []int64{1},
		MemberId:  4,
		OrderId:   10,
		Status:    1,
	})
	if err != nil {
		t.Fatalf("UpdateCouponRecord returned error: %v", err)
	}

	var record struct {
		Status  int32 `gorm:"column:status"`
		OrderID int64 `gorm:"column:order_id"`
	}
	if err := svcCtx.DB.Table("sms_coupon_record").Select("status, order_id").Where("id = ?", 5).Scan(&record).Error; err != nil {
		t.Fatalf("query coupon record failed: %v", err)
	}
	if record.Status != 1 || record.OrderID != 10 {
		t.Fatalf("expected coupon record used by order 10, got status=%d orderID=%d", record.Status, record.OrderID)
	}

	var usedCount int32
	if err := svcCtx.DB.Table("sms_coupon").Select("used_count").Where("id = ?", 1).Scan(&usedCount).Error; err != nil {
		t.Fatalf("query coupon used count failed: %v", err)
	}
	if usedCount != 1 {
		t.Fatalf("expected used_count 1, got %d", usedCount)
	}
}

func TestUpdateCouponRecordRollbackWritesZeroValues(t *testing.T) {
	svcCtx := newUpdateCouponRecordTestSvc(t)
	if _, err := NewUpdateCouponRecordLogic(context.Background(), svcCtx).UpdateCouponRecord(&smsclient.UpdateCouponRecordReq{
		CouponIds: []int64{1},
		MemberId:  4,
		OrderId:   10,
		Status:    1,
	}); err != nil {
		t.Fatalf("mark used returned error: %v", err)
	}

	_, err := NewUpdateCouponRecordLogic(context.Background(), svcCtx).UpdateCouponRecord(&smsclient.UpdateCouponRecordReq{
		CouponIds: []int64{1},
		MemberId:  4,
		Status:    0,
	})
	if err != nil {
		t.Fatalf("rollback coupon returned error: %v", err)
	}

	var record struct {
		Status  int32 `gorm:"column:status"`
		OrderID int64 `gorm:"column:order_id"`
	}
	if err := svcCtx.DB.Table("sms_coupon_record").Select("status, order_id").Where("id = ?", 5).Scan(&record).Error; err != nil {
		t.Fatalf("query coupon record failed: %v", err)
	}
	if record.Status != 0 || record.OrderID != 0 {
		t.Fatalf("expected coupon record rollback to zero values, got status=%d orderID=%d", record.Status, record.OrderID)
	}

	var usedCount int32
	if err := svcCtx.DB.Table("sms_coupon").Select("used_count").Where("id = ?", 1).Scan(&usedCount).Error; err != nil {
		t.Fatalf("query coupon used count failed: %v", err)
	}
	if usedCount != 0 {
		t.Fatalf("expected used_count 0 after rollback, got %d", usedCount)
	}
}
