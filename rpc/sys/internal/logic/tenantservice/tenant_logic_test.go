package tenantservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTenantTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "tenant-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(
		&model.SysUser{},
		&model.SysUserRole{},
		&tenantmodel.SysTenant{},
		&tenantmodel.SysTenantUser{},
		&tenantmodel.SysTenantAudit{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return &svc.ServiceContext{DB: db}
}

func TestCreateTenantCreatesTenantAdminBindingAndAudit(t *testing.T) {
	svcCtx := newTenantTestSvc(t)
	logic := NewCreateTenantLogic(context.Background(), svcCtx)

	resp, err := logic.CreateTenant(&sysclient.CreateTenantReq{
		TenantName:        "华东体验店",
		TenantShortName:   "华东",
		ContactName:       "张三",
		ContactMobile:     "13800138000",
		ContactEmail:      "tenant@example.com",
		AvailableChannels: []string{"app", "mini-program"},
		DataRetentionDays: 365,
		FeatureFlags:      []string{"inventory", "crm"},
		AdminUserName:     "tenant_admin_east",
		AdminNickName:     "华东管理员",
		AdminMobile:       "13900139000",
		AdminEmail:        "admin@example.com",
		AdminPassword:     "123456",
		CreateBy:          "platform-admin",
		OperatorId:        99,
	})
	if err != nil {
		t.Fatalf("create tenant failed: %v", err)
	}

	if resp.TenantId == 0 || resp.AdminUserId == 0 {
		t.Fatalf("unexpected create response: %+v", resp)
	}
	if resp.AdminActivationStatus != "pending-activation" {
		t.Fatalf("unexpected admin activation status: %s", resp.AdminActivationStatus)
	}

	var tenant tenantmodel.SysTenant
	if err := svcCtx.DB.First(&tenant, resp.TenantId).Error; err != nil {
		t.Fatalf("load tenant failed: %v", err)
	}
	if tenant.Status != tenantmodel.TenantStatusPendingActivation {
		t.Fatalf("unexpected tenant status: %d", tenant.Status)
	}
	if !strings.Contains(tenant.AvailableChannels, "mini-program") {
		t.Fatalf("channels not persisted: %s", tenant.AvailableChannels)
	}

	var user model.SysUser
	if err := svcCtx.DB.First(&user, resp.AdminUserId).Error; err != nil {
		t.Fatalf("load user failed: %v", err)
	}
	if user.Status != 0 {
		t.Fatalf("unexpected user status: %d", user.Status)
	}

	var binding tenantmodel.SysTenantUser
	if err := svcCtx.DB.Where("tenant_id = ? AND user_id = ?", resp.TenantId, resp.AdminUserId).First(&binding).Error; err != nil {
		t.Fatalf("load tenant binding failed: %v", err)
	}
	if binding.RoleMode != "bootstrap" || binding.ActivationStatus != tenantmodel.TenantStatusPendingActivation {
		t.Fatalf("unexpected binding: %+v", binding)
	}
	if !strings.Contains(binding.ScopeMetadata, resp.TenantCode) {
		t.Fatalf("scope metadata missing tenant code: %s", binding.ScopeMetadata)
	}

	var auditCount int64
	if err := svcCtx.DB.Model(&tenantmodel.SysTenantAudit{}).Where("tenant_id = ?", resp.TenantId).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit failed: %v", err)
	}
	if auditCount != 2 {
		t.Fatalf("unexpected audit count: %d", auditCount)
	}
}

func TestCreateTenantReusesDormantUser(t *testing.T) {
	svcCtx := newTenantTestSvc(t)
	user := &model.SysUser{
		Mobile:   "13900139001",
		UserName: "tenant_admin_reuse",
		NickName: "旧昵称",
		UserType: "00",
		Avatar:   defaultTenantAdminAvatar,
		Email:    "reuse@example.com",
		Password: "old-password",
		Status:   1,
		DeptID:   1,
		Remark:   "old",
		CreateBy: "seed",
		UpdateBy: "seed",
	}
	if err := svcCtx.DB.Create(user).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	logic := NewCreateTenantLogic(context.Background(), svcCtx)
	resp, err := logic.CreateTenant(&sysclient.CreateTenantReq{
		TenantName:        "华南体验店",
		ContactName:       "李四",
		ContactMobile:     "13800138001",
		AvailableChannels: []string{"app"},
		DataRetentionDays: 180,
		AdminUserName:     user.UserName,
		AdminNickName:     "复用管理员",
		AdminMobile:       user.Mobile,
		AdminEmail:        user.Email,
		AdminPassword:     "new-password",
		CreateBy:          "platform-admin",
	})
	if err != nil {
		t.Fatalf("create tenant with reused user failed: %v", err)
	}
	if resp.AdminUserId != user.ID {
		t.Fatalf("expected reused user id %d, got %d", user.ID, resp.AdminUserId)
	}

	var userCount int64
	if err := svcCtx.DB.Model(&model.SysUser{}).Count(&userCount).Error; err != nil {
		t.Fatalf("count user failed: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("expected 1 user after reuse, got %d", userCount)
	}

	var refreshed model.SysUser
	if err := svcCtx.DB.First(&refreshed, user.ID).Error; err != nil {
		t.Fatalf("reload reused user failed: %v", err)
	}
	if refreshed.Password != "new-password" || refreshed.Status != 0 {
		t.Fatalf("reused user was not refreshed: %+v", refreshed)
	}
}

func TestCreateTenantRejectsBoundUser(t *testing.T) {
	svcCtx := newTenantTestSvc(t)
	user := &model.SysUser{
		Mobile:   "13900139002",
		UserName: "tenant_admin_bound",
		NickName: "绑定管理员",
		UserType: "00",
		Avatar:   defaultTenantAdminAvatar,
		Password: "123456",
		Status:   0,
		DeptID:   1,
		Remark:   "seed",
		CreateBy: "seed",
		UpdateBy: "seed",
	}
	if err := svcCtx.DB.Create(user).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	tenant := &tenantmodel.SysTenant{
		PlatformID:        1,
		TenantCode:        "TEN202603200000000001",
		TenantName:        "已有租户",
		ContactName:       "王五",
		ContactMobile:     "13800138088",
		AvailableChannels: `["app"]`,
		DataRetentionDays: 180,
		FeatureFlags:      `[]`,
		Status:            tenantmodel.TenantStatusPendingActivation,
		CreatedBy:         "seed",
		UpdatedBy:         "seed",
	}
	if err := svcCtx.DB.Create(tenant).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}
	if err := svcCtx.DB.Create(&tenantmodel.SysTenantUser{
		PlatformID:       1,
		TenantID:         tenant.ID,
		UserID:           user.ID,
		RoleMode:         "bootstrap",
		ActivationStatus: tenantmodel.TenantStatusPendingActivation,
		IsPrimaryAdmin:   1,
		ScopeMetadata:    "{}",
		CreatedBy:        "seed",
		UpdatedBy:        "seed",
	}).Error; err != nil {
		t.Fatalf("seed tenant binding failed: %v", err)
	}

	logic := NewCreateTenantLogic(context.Background(), svcCtx)
	_, err := logic.CreateTenant(&sysclient.CreateTenantReq{
		TenantName:        "新租户",
		ContactName:       "赵六",
		ContactMobile:     "13800138002",
		AvailableChannels: []string{"app"},
		DataRetentionDays: 180,
		AdminUserName:     user.UserName,
		AdminNickName:     "重复管理员",
		AdminMobile:       user.Mobile,
		AdminPassword:     "654321",
		CreateBy:          "platform-admin",
	})
	if err == nil || !strings.Contains(err.Error(), "已绑定租户") {
		t.Fatalf("expected bound user error, got %v", err)
	}
}

func TestTenantStatusFlowAndQueries(t *testing.T) {
	svcCtx := newTenantTestSvc(t)
	createLogic := NewCreateTenantLogic(context.Background(), svcCtx)

	createResp, err := createLogic.CreateTenant(&sysclient.CreateTenantReq{
		TenantName:        "西南门店",
		TenantShortName:   "西南",
		ContactName:       "陈七",
		ContactMobile:     "13800138003",
		AvailableChannels: []string{"app", "h5"},
		DataRetentionDays: 90,
		FeatureFlags:      []string{"oms"},
		AdminUserName:     "tenant_admin_sw",
		AdminNickName:     "西南管理员",
		AdminMobile:       "13900139003",
		AdminPassword:     "123456",
		CreateBy:          "platform-admin",
		OperatorId:        7,
	})
	if err != nil {
		t.Fatalf("create tenant failed: %v", err)
	}

	enableLogic := NewEnableTenantLogic(context.Background(), svcCtx)
	if _, err := enableLogic.EnableTenant(&sysclient.ChangeTenantStatusReq{
		Ids:          []int64{createResp.TenantId},
		StatusReason: "通过验收",
		UpdateBy:     "platform-admin",
		OperatorId:   7,
	}); err != nil {
		t.Fatalf("enable tenant failed: %v", err)
	}

	queryListLogic := NewQueryTenantListLogic(context.Background(), svcCtx)
	listResp, err := queryListLogic.QueryTenantList(&sysclient.QueryTenantListReq{
		Status:   tenantmodel.TenantStatusEnabled,
		Channel:  "h5",
		PageNum:  1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("query tenant list failed: %v", err)
	}
	if listResp.Total != 1 || len(listResp.List) != 1 {
		t.Fatalf("unexpected tenant list response: %+v", listResp)
	}
	if listResp.List[0].AdminActivationStatus != "active" {
		t.Fatalf("unexpected list activation status: %+v", listResp.List[0])
	}

	queryDetailLogic := NewQueryTenantDetailLogic(context.Background(), svcCtx)
	detailResp, err := queryDetailLogic.QueryTenantDetail(&sysclient.QueryTenantDetailReq{Id: createResp.TenantId})
	if err != nil {
		t.Fatalf("query tenant detail failed: %v", err)
	}
	if detailResp.Data.TenantCode != createResp.TenantCode || len(detailResp.Data.AvailableChannels) != 2 {
		t.Fatalf("unexpected detail response: %+v", detailResp.Data)
	}

	archiveLogic := NewArchiveTenantLogic(context.Background(), svcCtx)
	if _, err := archiveLogic.ArchiveTenant(&sysclient.ChangeTenantStatusReq{
		Ids:        []int64{createResp.TenantId},
		UpdateBy:   "platform-admin",
		OperatorId: 7,
	}); err != nil {
		t.Fatalf("archive tenant failed: %v", err)
	}

	if _, err := enableLogic.EnableTenant(&sysclient.ChangeTenantStatusReq{
		Ids:      []int64{createResp.TenantId},
		UpdateBy: "platform-admin",
	}); err == nil {
		t.Fatal("expected archived tenant enable to fail")
	}

	var binding tenantmodel.SysTenantUser
	if err := svcCtx.DB.Where("tenant_id = ?", createResp.TenantId).First(&binding).Error; err != nil {
		t.Fatalf("reload binding failed: %v", err)
	}
	if binding.ActivationStatus != tenantmodel.TenantStatusArchived {
		t.Fatalf("unexpected archived binding status: %d", binding.ActivationStatus)
	}

	var auditCount int64
	if err := svcCtx.DB.Model(&tenantmodel.SysTenantAudit{}).Where("tenant_id = ?", createResp.TenantId).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit failed: %v", err)
	}
	if auditCount != 4 {
		t.Fatalf("unexpected audit count after status changes: %d", auditCount)
	}
}
