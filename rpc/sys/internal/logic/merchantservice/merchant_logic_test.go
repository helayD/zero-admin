package merchantservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newMerchantTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "merchant-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&model.SysUser{}, &tenantmodel.SysTenant{}, &merchantmodel.SysMerchant{}, &merchantmodel.SysMerchantAudit{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	for _, statement := range []string{
		`ALTER TABLE sys_user ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE sys_user ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE sys_user ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("extend sys_user table failed: %v", err)
		}
	}

	if err := db.Exec(`CREATE TABLE sys_user_scope (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		scope_type TEXT NOT NULL DEFAULT 'platform',
		platform_id INTEGER NOT NULL DEFAULT 1,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		merchant_id INTEGER NOT NULL DEFAULT 0,
		dept_id INTEGER NOT NULL DEFAULT 0,
		role_mode TEXT NOT NULL DEFAULT '',
		activation_status INTEGER NOT NULL DEFAULT 1,
		is_primary INTEGER NOT NULL DEFAULT 1,
		scope_metadata TEXT NOT NULL DEFAULT '{}',
		created_by TEXT NOT NULL DEFAULT 'admin',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NULL
	)`).Error; err != nil {
		t.Fatalf("create user scope table failed: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX uk_sys_user_scope_binding ON sys_user_scope(user_id, scope_type, platform_id, tenant_id, merchant_id)`).Error; err != nil {
		t.Fatalf("create user scope unique index failed: %v", err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	t.Cleanup(mr.Close)

	rds := redis.New(mr.Addr())
	return &svc.ServiceContext{DB: db, Redis: rds}
}

func seedMerchantTenant(t *testing.T, svcCtx *svc.ServiceContext, tenantID int64, status int32) {
	t.Helper()

	tenant := &tenantmodel.SysTenant{
		ID:                tenantID,
		PlatformID:        1,
		TenantCode:        "TEN-MER-001",
		TenantName:        "华东租户",
		ContactName:       "张三",
		ContactMobile:     "13800138000",
		AvailableChannels: `["app"]`,
		DataRetentionDays: 180,
		FeatureFlags:      `["oms"]`,
		Status:            status,
		CreatedBy:         "seed",
		UpdatedBy:         "seed",
	}
	if err := svcCtx.DB.Create(tenant).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}
}

func seedMerchantUser(t *testing.T, svcCtx *svc.ServiceContext, userID, merchantID int64) {
	t.Helper()

	if err := svcCtx.DB.Exec(
		`INSERT INTO sys_user (id, mobile, user_name, nick_name, user_type, avatar, email, password, status, dept_id, login_ip, login_browser, login_os, remark, del_flag, create_by, update_by, platform_id, tenant_id, merchant_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		"13900139000",
		"merchant_admin",
		"商户管理员",
		"00",
		"",
		"merchant@example.com",
		"123456",
		1,
		1,
		"",
		"",
		"",
		"seed",
		1,
		"seed",
		"seed",
		1,
		11,
		merchantID,
	).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
}

func TestCreateApproveAndEnableMerchantFlow(t *testing.T) {
	svcCtx := newMerchantTestSvc(t)
	seedMerchantTenant(t, svcCtx, 11, tenantmodel.TenantStatusEnabled)
	seedMerchantUser(t, svcCtx, 41, 0)

	createLogic := NewCreateMerchantLogic(context.Background(), svcCtx)
	createResp, err := createLogic.CreateMerchant(&sysclient.CreateMerchantReq{
		TenantId:           11,
		MerchantName:       "华东旗舰店",
		ContactName:        "李四",
		ContactMobile:      "13800138001",
		AvailableChannels:  []string{"app", "mini-program"},
		CapabilityFlags:    []string{"oms", "crm"},
		PrimaryAdminUserId: 41,
		CreateBy:           "platform-admin",
		OperatorId:         9,
	})
	if err != nil {
		t.Fatalf("create merchant failed: %v", err)
	}
	if createResp.MerchantId == 0 || createResp.ReviewStatus != merchantmodel.MerchantReviewPending {
		t.Fatalf("unexpected create response: %+v", createResp)
	}

	approveLogic := NewApproveMerchantLogic(context.Background(), svcCtx)
	if _, err := approveLogic.ApproveMerchant(&sysclient.ReviewMerchantReq{
		Ids:          []int64{createResp.MerchantId},
		ReviewReason: "资料齐全",
		UpdateBy:     "platform-admin",
		OperatorId:   9,
	}); err != nil {
		t.Fatalf("approve merchant failed: %v", err)
	}

	enableLogic := NewEnableMerchantLogic(context.Background(), svcCtx)
	if _, err := enableLogic.EnableMerchant(&sysclient.ChangeMerchantStatusReq{
		Ids:          []int64{createResp.MerchantId},
		StatusReason: "正式开业",
		UpdateBy:     "platform-admin",
		OperatorId:   9,
	}); err != nil {
		t.Fatalf("enable merchant failed: %v", err)
	}

	var merchant merchantmodel.SysMerchant
	if err := svcCtx.DB.First(&merchant, createResp.MerchantId).Error; err != nil {
		t.Fatalf("load merchant failed: %v", err)
	}
	if merchant.ReviewStatus != merchantmodel.MerchantReviewApproved || merchant.BusinessStatus != merchantmodel.MerchantBusinessEnabled {
		t.Fatalf("unexpected merchant status: %+v", merchant)
	}

	var binding logiccommon.UserScopeBinding
	if err := svcCtx.DB.Table("sys_user_scope").Where("user_id = ? AND merchant_id = ?", 41, createResp.MerchantId).Take(&binding).Error; err != nil {
		t.Fatalf("load merchant scope binding failed: %v", err)
	}
	if binding.ActivationStatus != logiccommon.UserActivationActive {
		t.Fatalf("expected active binding, got %+v", binding)
	}

	var auditCount int64
	if err := svcCtx.DB.Model(&merchantmodel.SysMerchantAudit{}).Where("merchant_id = ?", createResp.MerchantId).Count(&auditCount).Error; err != nil {
		t.Fatalf("count merchant audit failed: %v", err)
	}
	if auditCount != 3 {
		t.Fatalf("expected 3 audit rows, got %d", auditCount)
	}
}

func TestApproveMerchantRejectsDisabledTenant(t *testing.T) {
	svcCtx := newMerchantTestSvc(t)
	seedMerchantTenant(t, svcCtx, 11, tenantmodel.TenantStatusDisabled)

	merchant := &merchantmodel.SysMerchant{
		PlatformID:        1,
		TenantID:          11,
		MerchantCode:      "MER-001",
		MerchantName:      "待审核商户",
		ContactName:       "王五",
		ContactMobile:     "13800138002",
		AvailableChannels: `["app"]`,
		CapabilityFlags:   `["oms"]`,
		ReviewStatus:      merchantmodel.MerchantReviewPending,
		BusinessStatus:    merchantmodel.MerchantBusinessPendingActivation,
		CreatedBy:         "seed",
		UpdatedBy:         "seed",
	}
	if err := svcCtx.DB.Create(merchant).Error; err != nil {
		t.Fatalf("seed merchant failed: %v", err)
	}

	logic := NewApproveMerchantLogic(context.Background(), svcCtx)
	_, err := logic.ApproveMerchant(&sysclient.ReviewMerchantReq{
		Ids:          []int64{merchant.ID},
		ReviewReason: "尝试通过",
		UpdateBy:     "platform-admin",
		OperatorId:   9,
	})
	if err == nil || !strings.Contains(err.Error(), "目标租户已停用或已归档") {
		t.Fatalf("expected disabled tenant error, got %v", err)
	}
}

func TestDisableMerchantUpdatesScopeAndEvictsPermissionCache(t *testing.T) {
	svcCtx := newMerchantTestSvc(t)
	seedMerchantTenant(t, svcCtx, 11, tenantmodel.TenantStatusEnabled)
	seedMerchantUser(t, svcCtx, 41, 101)

	merchant := &merchantmodel.SysMerchant{
		ID:                 101,
		PlatformID:         1,
		TenantID:           11,
		MerchantCode:       "MER-101",
		MerchantName:       "已启用商户",
		ContactName:        "赵六",
		ContactMobile:      "13800138003",
		AvailableChannels:  `["app"]`,
		CapabilityFlags:    `["oms"]`,
		ReviewStatus:       merchantmodel.MerchantReviewApproved,
		BusinessStatus:     merchantmodel.MerchantBusinessEnabled,
		PrimaryAdminUserID: 41,
		CreatedBy:          "seed",
		UpdatedBy:          "seed",
	}
	if err := svcCtx.DB.Create(merchant).Error; err != nil {
		t.Fatalf("seed merchant failed: %v", err)
	}
	if err := svcCtx.DB.Exec(
		`INSERT INTO sys_user_scope (user_id, scope_type, platform_id, tenant_id, merchant_id, role_mode, activation_status, scope_metadata, created_by, updated_by)
		 VALUES (41, 'merchant', 1, 11, 101, 'bootstrap', 1, '{}', 'seed', 'seed')`,
	).Error; err != nil {
		t.Fatalf("seed user scope failed: %v", err)
	}
	if err := svcCtx.Redis.HsetCtx(context.Background(), merchantPermissionCacheKey, "41", "/api/sys/merchant/queryMerchantList"); err != nil {
		t.Fatalf("seed redis permission failed: %v", err)
	}

	logic := NewDisableMerchantLogic(context.Background(), svcCtx)
	if _, err := logic.DisableMerchant(&sysclient.ChangeMerchantStatusReq{
		Ids:          []int64{101},
		StatusReason: "风控冻结",
		UpdateBy:     "platform-admin",
		OperatorId:   9,
	}); err != nil {
		t.Fatalf("disable merchant failed: %v", err)
	}

	var binding logiccommon.UserScopeBinding
	if err := svcCtx.DB.Table("sys_user_scope").Where("user_id = ? AND merchant_id = ?", 41, 101).Take(&binding).Error; err != nil {
		t.Fatalf("reload scope binding failed: %v", err)
	}
	if binding.ActivationStatus != logiccommon.UserActivationDisabled {
		t.Fatalf("expected disabled binding, got %+v", binding)
	}

	got, err := svcCtx.Redis.HgetCtx(context.Background(), merchantPermissionCacheKey, "41")
	if err != nil && err.Error() != "redis: nil" {
		t.Fatalf("read redis permission failed: %v", err)
	}
	if got != "" {
		t.Fatalf("expected permission cache to be evicted, got %q", got)
	}
}
