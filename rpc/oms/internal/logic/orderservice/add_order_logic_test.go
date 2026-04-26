package orderservicelogic

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAddOrderTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "add-order.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sys_user (
			id INTEGER PRIMARY KEY,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_order_main (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			order_no TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			order_status INTEGER NOT NULL DEFAULT 1,
			total_amount REAL NOT NULL,
			promotion_amount REAL NOT NULL DEFAULT 0,
			coupon_amount REAL NOT NULL DEFAULT 0,
			points_amount REAL NOT NULL DEFAULT 0,
			discount_amount REAL NOT NULL DEFAULT 0,
			freight_amount REAL NOT NULL DEFAULT 0,
			pay_amount REAL NOT NULL,
			pay_type INTEGER NULL,
			pay_time DATETIME NULL,
			delivery_time DATETIME NULL,
			receive_time DATETIME NULL,
			comment_time DATETIME NULL,
			source_type INTEGER NOT NULL DEFAULT 1,
			express_order_number TEXT NOT NULL DEFAULT '',
			use_points INTEGER NOT NULL DEFAULT 0,
			receive_status INTEGER NOT NULL DEFAULT 0,
			remark TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0,
			consistency_stage INTEGER NOT NULL DEFAULT 0,
			consistency_result INTEGER NOT NULL DEFAULT 0,
			last_error TEXT NOT NULL DEFAULT '',
			retry_count INTEGER NOT NULL DEFAULT 0,
			last_compensation_at DATETIME NULL,
			manual_required INTEGER NOT NULL DEFAULT 0,
			paused INTEGER NOT NULL DEFAULT 0,
			paused_at DATETIME NULL,
			pause_reason TEXT NOT NULL DEFAULT '',
			pause_operator_id INTEGER NOT NULL DEFAULT 0,
			activity_type TEXT NOT NULL DEFAULT '',
			activity_id INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_order_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			order_no TEXT NOT NULL,
			order_item_status INTEGER NOT NULL DEFAULT 1,
			sku_id INTEGER NOT NULL,
			sku_name TEXT NOT NULL,
			sku_pic TEXT NOT NULL,
			sku_price REAL NOT NULL,
			sku_quantity INTEGER NOT NULL,
			spec_data TEXT NOT NULL,
			sku_total_amount REAL NOT NULL,
			promotion_amount REAL NOT NULL DEFAULT 0,
			coupon_amount REAL NOT NULL DEFAULT 0,
			points_amount REAL NOT NULL DEFAULT 0,
			discount_amount REAL NOT NULL DEFAULT 0,
			real_amount REAL NOT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_cart_item (
			id INTEGER PRIMARY KEY,
			member_id INTEGER NOT NULL,
			delete_status INTEGER NOT NULL DEFAULT 0,
			update_time DATETIME NULL
		)`,
		`INSERT INTO sys_user (id, platform_id, tenant_id, merchant_id) VALUES (4, 1, 0, 0)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	return &svc.ServiceContext{DB: db}
}

func newAddOrderReq(orderItemID int64) *omsclient.AddOrderReq {
	return &omsclient.AddOrderReq{
		OrderNo:     "OR-TEST-001",
		UserId:      4,
		OrderStatus: 1,
		TotalAmount: 8999,
		PayAmount:   8999,
		SourceType:  1,
		OrderItemData: []*omsclient.OrderItemData{
			{
				Id:             orderItemID,
				SkuId:          36,
				SkuName:        "iPhone",
				SkuPic:         "pic",
				SkuPrice:       8999,
				SkuQuantity:    1,
				SpecData:       "{}",
				SkuTotalAmount: 8999,
				RealAmount:     8999,
			},
		},
	}
}

func TestAddOrderMarksCartItemDeletedByCartItemID(t *testing.T) {
	svcCtx := newAddOrderTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO oms_cart_item (id, member_id, delete_status) VALUES (10, 4, 0), (36, 4, 0)`).Error; err != nil {
		t.Fatalf("seed cart items failed: %v", err)
	}

	resp, err := NewAddOrderLogic(context.Background(), svcCtx).AddOrder(newAddOrderReq(10))
	if err != nil {
		t.Fatalf("AddOrder returned error: %v", err)
	}
	if resp.Id <= 0 {
		t.Fatalf("expected response order id, got %d", resp.Id)
	}

	var deletedCartItem int32
	if err := svcCtx.DB.Table("oms_cart_item").Select("delete_status").Where("id = ?", 10).Scan(&deletedCartItem).Error; err != nil {
		t.Fatalf("query deleted cart item failed: %v", err)
	}
	if deletedCartItem != 1 {
		t.Fatalf("expected cart item 10 to be marked deleted, got %d", deletedCartItem)
	}

	var skuMatchedCartItem int32
	if err := svcCtx.DB.Table("oms_cart_item").Select("delete_status").Where("id = ?", 36).Scan(&skuMatchedCartItem).Error; err != nil {
		t.Fatalf("query sku-matched cart item failed: %v", err)
	}
	if skuMatchedCartItem != 0 {
		t.Fatalf("expected cart item 36 to stay active, got %d", skuMatchedCartItem)
	}

	var orderItem model.OmsOrderItem
	if err := svcCtx.DB.Where("order_id = ?", resp.Id).First(&orderItem).Error; err != nil {
		t.Fatalf("query order item failed: %v", err)
	}
	if orderItem.OrderNo != "OR-TEST-001" {
		t.Fatalf("expected order item order no to be set from order, got %q", orderItem.OrderNo)
	}
}

func TestAddOrderDirectBuyDoesNotDeleteCartItems(t *testing.T) {
	svcCtx := newAddOrderTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO oms_cart_item (id, member_id, delete_status) VALUES (36, 4, 0)`).Error; err != nil {
		t.Fatalf("seed cart item failed: %v", err)
	}

	if _, err := NewAddOrderLogic(context.Background(), svcCtx).AddOrder(newAddOrderReq(0)); err != nil {
		t.Fatalf("AddOrder returned error: %v", err)
	}

	var deleteStatus int32
	if err := svcCtx.DB.Table("oms_cart_item").Select("delete_status").Where("id = ?", 36).Scan(&deleteStatus).Error; err != nil {
		t.Fatalf("query cart item failed: %v", err)
	}
	if deleteStatus != 0 {
		t.Fatalf("expected direct buy to leave cart item active, got %d", deleteStatus)
	}
}

func TestAddOrderFallsBackToPlatformScopeWhenMemberIsNotSysUser(t *testing.T) {
	svcCtx := newAddOrderTestSvc(t)
	if err := svcCtx.DB.Exec(`DELETE FROM sys_user WHERE id = 4`).Error; err != nil {
		t.Fatalf("delete sys user failed: %v", err)
	}

	resp, err := NewAddOrderLogic(context.Background(), svcCtx).AddOrder(newAddOrderReq(0))
	if err != nil {
		t.Fatalf("AddOrder returned error: %v", err)
	}
	if resp.Id <= 0 {
		t.Fatalf("expected response order id, got %d", resp.Id)
	}
}
