package orderservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOrderScopeTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "order-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE oms_order_main (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			express_order_number TEXT NOT NULL,
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
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0
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
		`CREATE TABLE oms_order_operation_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			order_no TEXT NOT NULL,
			operation_type INTEGER NOT NULL,
			operator_id INTEGER NOT NULL,
			operator_type INTEGER NOT NULL,
			operator_note TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE oms_order_promotion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			order_no TEXT NOT NULL,
			promotion_type INTEGER NOT NULL,
			promotion_id INTEGER NULL,
			promotion_name TEXT NOT NULL,
			discount_amount REAL NOT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_order_delivery (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			order_no TEXT NOT NULL,
			receiver_name TEXT NOT NULL,
			receiver_phone TEXT NOT NULL,
			receiver_province TEXT NOT NULL,
			receiver_city TEXT NOT NULL,
			receiver_district TEXT NOT NULL,
			receiver_address TEXT NOT NULL,
			delivery_company TEXT NOT NULL,
			delivery_no TEXT NOT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_order_payment (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			order_no TEXT NOT NULL,
			pay_type INTEGER NOT NULL,
			transaction_id TEXT NOT NULL,
			total_amount REAL NOT NULL,
			pay_amount REAL NOT NULL,
			pay_status INTEGER NOT NULL,
			pay_time DATETIME NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec order schema failed: %v", err)
		}
	}
	query.SetDefault(db)

	return &svc.ServiceContext{DB: db}
}

func seedOrderMain(t *testing.T, db *gorm.DB, item model.OmsOrderMain, current pkgscope.GovernanceScope) {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed order main failed: %v", err)
	}
	if err := db.Exec(
		`UPDATE oms_order_main SET platform_id = ?, tenant_id = ?, merchant_id = ? WHERE id = ?`,
		current.PlatformID,
		current.TenantID,
		current.MerchantID,
		item.ID,
	).Error; err != nil {
		t.Fatalf("update order scope failed: %v", err)
	}
}

func TestQueryOrderListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newOrderScopeTestSvc(t)
	now := time.Now()
	payType := int32(1)

	tenantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	if err != nil {
		t.Fatalf("normalize tenant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 20, 0)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedOrderMain(t, svcCtx.DB, model.OmsOrderMain{
		ID:                 1,
		OrderNo:            "TENANT-10-ORDER",
		UserID:             100,
		OrderStatus:        2,
		TotalAmount:        200,
		PromotionAmount:    10,
		CouponAmount:       20,
		PointsAmount:       0,
		DiscountAmount:     30,
		FreightAmount:      10,
		PayAmount:          170,
		PayType:            &payType,
		PayTime:            &now,
		SourceType:         1,
		ExpressOrderNumber: "EXP-10",
		UsePoints:          0,
		ReceiveStatus:      0,
		Remark:             "tenant 10",
		CreateTime:         now,
		IsDeleted:          0,
	}, tenantScope)
	if err := svcCtx.DB.Create(&model.OmsOrderItem{
		ID:              1,
		OrderID:         1,
		OrderNo:         "TENANT-10-ORDER",
		OrderItemStatus: 1,
		SkuID:           11,
		SkuName:         "tenant 10 sku",
		SkuPic:          "sku.png",
		SkuPrice:        200,
		SkuQuantity:     1,
		SpecData:        "{}",
		SkuTotalAmount:  200,
		RealAmount:      170,
		CreateTime:      now,
		IsDeleted:       false,
	}).Error; err != nil {
		t.Fatalf("seed order item failed: %v", err)
	}
	if err := svcCtx.DB.Create(&model.OmsOrderOperationLog{
		ID:            1,
		OrderID:       1,
		OrderNo:       "TENANT-10-ORDER",
		OperationType: 1,
		OperatorID:    100,
		OperatorType:  1,
		OperatorNote:  "created",
		CreateTime:    now,
	}).Error; err != nil {
		t.Fatalf("seed order operation log failed: %v", err)
	}

	seedOrderMain(t, svcCtx.DB, model.OmsOrderMain{
		ID:                 2,
		OrderNo:            "TENANT-20-ORDER",
		UserID:             200,
		OrderStatus:        2,
		TotalAmount:        300,
		PromotionAmount:    0,
		CouponAmount:       0,
		PointsAmount:       0,
		DiscountAmount:     0,
		FreightAmount:      10,
		PayAmount:          310,
		PayType:            &payType,
		PayTime:            &now,
		SourceType:         1,
		ExpressOrderNumber: "EXP-20",
		UsePoints:          0,
		ReceiveStatus:      0,
		Remark:             "tenant 20",
		CreateTime:         now,
		IsDeleted:          0,
	}, otherScope)

	logic := NewQueryOrderListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryOrderList(&omsclient.QueryOrderListReq{
		PageNum:  1,
		PageSize: 10,
		Scope: &omsclient.GovernanceScope{
			ScopeType:  tenantScope.ScopeType,
			PlatformId: tenantScope.PlatformID,
			TenantId:   tenantScope.TenantID,
			MerchantId: tenantScope.MerchantID,
		},
	})
	if err != nil {
		t.Fatalf("query order list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("expected one scoped order, got total=%d len=%d", resp.Total, len(resp.List))
	}
	if resp.List[0].OrderNo != "TENANT-10-ORDER" {
		t.Fatalf("expected tenant 10 order, got %+v", resp.List[0])
	}
}

func TestQueryOrderDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newOrderScopeTestSvc(t)
	now := time.Now()
	payType := int32(1)

	merchantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize merchant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedOrderMain(t, svcCtx.DB, model.OmsOrderMain{
		ID:                 11,
		OrderNo:            "M-301-ORDER",
		UserID:             301,
		OrderStatus:        2,
		TotalAmount:        120,
		PromotionAmount:    10,
		CouponAmount:       5,
		PointsAmount:       0,
		DiscountAmount:     15,
		FreightAmount:      10,
		PayAmount:          105,
		PayType:            &payType,
		PayTime:            &now,
		SourceType:         1,
		ExpressOrderNumber: "EXP-301",
		UsePoints:          0,
		ReceiveStatus:      0,
		Remark:             "merchant 301",
		CreateTime:         now,
		IsDeleted:          0,
	}, merchantScope)
	for _, create := range []func() error{
		func() error {
			return svcCtx.DB.Create(&model.OmsOrderItem{
				ID:              11,
				OrderID:         11,
				OrderNo:         "M-301-ORDER",
				OrderItemStatus: 1,
				SkuID:           101,
				SkuName:         "merchant sku",
				SkuPic:          "sku.png",
				SkuPrice:        120,
				SkuQuantity:     1,
				SpecData:        "{}",
				SkuTotalAmount:  120,
				RealAmount:      105,
				CreateTime:      now,
				IsDeleted:       false,
			}).Error
		},
		func() error {
			return svcCtx.DB.Create(&model.OmsOrderOperationLog{
				ID:            11,
				OrderID:       11,
				OrderNo:       "M-301-ORDER",
				OperationType: 1,
				OperatorID:    301,
				OperatorType:  1,
				OperatorNote:  "created",
				CreateTime:    now,
			}).Error
		},
		func() error {
			return svcCtx.DB.Create(&model.OmsOrderPromotion{
				ID:             11,
				OrderID:        11,
				OrderNo:        "M-301-ORDER",
				PromotionType:  1,
				PromotionName:  "coupon",
				DiscountAmount: 15,
				CreateTime:     now,
				IsDeleted:      0,
			}).Error
		},
		func() error {
			return svcCtx.DB.Create(&model.OmsOrderDelivery{
				ID:               11,
				OrderID:          11,
				OrderNo:          "M-301-ORDER",
				ReceiverName:     "tester",
				ReceiverPhone:    "13800000000",
				ReceiverProvince: "广东",
				ReceiverCity:     "深圳",
				ReceiverDistrict: "南山",
				ReceiverAddress:  "科技园",
				DeliveryCompany:  "SF",
				DeliveryNo:       "SF-11",
				CreateTime:       now,
				IsDeleted:        0,
			}).Error
		},
		func() error {
			return svcCtx.DB.Create(&model.OmsOrderPayment{
				ID:            11,
				OrderID:       11,
				OrderNo:       "M-301-ORDER",
				PayType:       1,
				TransactionID: "TX-11",
				TotalAmount:   120,
				PayAmount:     105,
				PayStatus:     1,
				PayTime:       &now,
				CreateTime:    now,
				IsDeleted:     0,
			}).Error
		},
	} {
		if err := create(); err != nil {
			t.Fatalf("seed order detail child data failed: %v", err)
		}
	}

	logic := NewQueryOrderDetailLogic(context.Background(), svcCtx)
	_, err = logic.QueryOrderDetail(&omsclient.QueryOrderDetailReq{
		Id: 11,
		Scope: &omsclient.GovernanceScope{
			ScopeType:  otherScope.ScopeType,
			PlatformId: otherScope.PlatformID,
			TenantId:   otherScope.TenantID,
			MerchantId: otherScope.MerchantID,
		},
	})
	if err == nil {
		t.Fatal("expected cross-scope order detail to fail")
	}
	if !strings.Contains(err.Error(), "订单不存在") {
		t.Fatalf("expected not found error, got %v", err)
	}

	resp, err := logic.QueryOrderDetail(&omsclient.QueryOrderDetailReq{
		Id: 11,
		Scope: &omsclient.GovernanceScope{
			ScopeType:  merchantScope.ScopeType,
			PlatformId: merchantScope.PlatformID,
			TenantId:   merchantScope.TenantID,
			MerchantId: merchantScope.MerchantID,
		},
	})
	if err != nil {
		t.Fatalf("query order detail with correct scope failed: %v", err)
	}
	if resp.Data == nil || resp.Data.OrderNo != "M-301-ORDER" {
		t.Fatalf("expected scoped order detail, got %+v", resp.Data)
	}
	if len(resp.Data.OrderItemData) != 1 || len(resp.Data.OptLogData) != 1 || len(resp.Data.PromotionData) != 1 || len(resp.Data.PaymentData) != 1 {
		t.Fatalf("expected child data to load after scope validation, got %+v", resp.Data)
	}
	if resp.Data.DeliveryData == nil || resp.Data.DeliveryData.OrderId != 11 {
		t.Fatalf("expected delivery data, got %+v", resp.Data.DeliveryData)
	}
}
