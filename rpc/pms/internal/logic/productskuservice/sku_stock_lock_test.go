package productskuservicelogic

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSkuStockLockTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "sku-stock-lock.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE pms_product_sku (
		id INTEGER PRIMARY KEY,
		stock INTEGER NOT NULL,
		lock_stock INTEGER NOT NULL
	)`).Error; err != nil {
		t.Fatalf("create sku table failed: %v", err)
	}
	return &svc.ServiceContext{DB: db}
}

func TestLockSkuStockLockMovesStockToLockedStock(t *testing.T) {
	svcCtx := newSkuStockLockTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_sku (id, stock, lock_stock) VALUES (36, 5, 1)`).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}

	logic := NewLockSkuStockLockLogic(context.Background(), svcCtx)
	_, err := logic.LockSkuStockLock(&pmsclient.UpdateSkuStockReq{
		Data: []*pmsclient.UpdateSkuStockData{{Id: 36, ProductQuantity: 2}},
	})
	if err != nil {
		t.Fatalf("lock stock failed: %v", err)
	}

	var row struct {
		Stock     int32
		LockStock int32 `gorm:"column:lock_stock"`
	}
	if err := svcCtx.DB.Table("pms_product_sku").Select("stock, lock_stock").Where("id = ?", 36).Take(&row).Error; err != nil {
		t.Fatalf("load sku failed: %v", err)
	}
	if row.Stock != 3 || row.LockStock != 3 {
		t.Fatalf("unexpected stock after lock: stock=%d lock_stock=%d", row.Stock, row.LockStock)
	}
}

func TestLockSkuStockLockRollsBackWhenAnySkuInsufficient(t *testing.T) {
	svcCtx := newSkuStockLockTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_sku (id, stock, lock_stock) VALUES (36, 5, 0), (37, 1, 0)`).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}

	logic := NewLockSkuStockLockLogic(context.Background(), svcCtx)
	_, err := logic.LockSkuStockLock(&pmsclient.UpdateSkuStockReq{
		Data: []*pmsclient.UpdateSkuStockData{
			{Id: 36, ProductQuantity: 2},
			{Id: 37, ProductQuantity: 2},
		},
	})
	if err == nil {
		t.Fatal("expected insufficient stock error")
	}

	var stock int32
	if err := svcCtx.DB.Table("pms_product_sku").Select("stock").Where("id = ?", 36).Take(&stock).Error; err != nil {
		t.Fatalf("load sku failed: %v", err)
	}
	if stock != 5 {
		t.Fatalf("expected first sku lock to roll back, got stock=%d", stock)
	}
}

func TestReleaseSkuStockLockMovesLockedStockBackToStock(t *testing.T) {
	svcCtx := newSkuStockLockTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_sku (id, stock, lock_stock) VALUES (36, 3, 2)`).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}

	logic := NewReleaseSkuStockLockLogic(context.Background(), svcCtx)
	_, err := logic.ReleaseSkuStockLock(&pmsclient.UpdateSkuStockReq{
		Data: []*pmsclient.UpdateSkuStockData{{Id: 36, ProductQuantity: 2}},
	})
	if err != nil {
		t.Fatalf("release stock failed: %v", err)
	}

	var row struct {
		Stock     int32
		LockStock int32 `gorm:"column:lock_stock"`
	}
	if err := svcCtx.DB.Table("pms_product_sku").Select("stock, lock_stock").Where("id = ?", 36).Take(&row).Error; err != nil {
		t.Fatalf("load sku failed: %v", err)
	}
	if row.Stock != 5 || row.LockStock != 0 {
		t.Fatalf("unexpected stock after release: stock=%d lock_stock=%d", row.Stock, row.LockStock)
	}
}
