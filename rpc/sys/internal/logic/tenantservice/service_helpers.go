package tenantservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/audit"
	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/gorm"
)

const (
	defaultPlatformID        int64 = 1
	defaultDeptID            int64 = 1
	defaultDataRetentionDays int32 = 180
	defaultTenantAdminAvatar       = "https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png"
	defaultTenantAdminRemark       = "租户首个管理员（待激活）"
)

const tenantSelectColumns = `
t.id,
t.tenant_code,
t.tenant_name,
t.tenant_short_name,
t.contact_name,
t.contact_mobile,
t.contact_email,
t.available_channels,
t.data_retention_days,
t.feature_flags,
t.status,
t.status_reason,
COALESCE(tu.user_id, 0) AS primary_admin_user_id,
COALESCE(su.user_name, '') AS primary_admin_user_name,
COALESCE(su.mobile, '') AS primary_admin_mobile,
COALESCE(tu.activation_status, t.status) AS admin_activation_status,
t.created_by,
t.created_at,
t.updated_by,
t.updated_at
`

type createTenantInput struct {
	TenantName        string
	TenantShortName   string
	ContactName       string
	ContactMobile     string
	ContactEmail      string
	AvailableChannels []string
	DataRetentionDays int32
	FeatureFlags      []string
	AdminUserName     string
	AdminNickName     string
	AdminMobile       string
	AdminEmail        string
	AdminPassword     string
	CreateBy          string
	OperatorID        int64
}

type tenantQueryRow struct {
	ID                    int64      `gorm:"column:id"`
	TenantCode            string     `gorm:"column:tenant_code"`
	TenantName            string     `gorm:"column:tenant_name"`
	TenantShortName       string     `gorm:"column:tenant_short_name"`
	ContactName           string     `gorm:"column:contact_name"`
	ContactMobile         string     `gorm:"column:contact_mobile"`
	ContactEmail          string     `gorm:"column:contact_email"`
	AvailableChannels     string     `gorm:"column:available_channels"`
	DataRetentionDays     int32      `gorm:"column:data_retention_days"`
	FeatureFlags          string     `gorm:"column:feature_flags"`
	Status                int32      `gorm:"column:status"`
	StatusReason          string     `gorm:"column:status_reason"`
	PrimaryAdminUserID    int64      `gorm:"column:primary_admin_user_id"`
	PrimaryAdminUserName  string     `gorm:"column:primary_admin_user_name"`
	PrimaryAdminMobile    string     `gorm:"column:primary_admin_mobile"`
	AdminActivationStatus int32      `gorm:"column:admin_activation_status"`
	CreatedBy             string     `gorm:"column:created_by"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedBy             string     `gorm:"column:updated_by"`
	UpdatedAt             *time.Time `gorm:"column:updated_at"`
}

func normalizeCreateTenantReq(in *sysclient.CreateTenantReq) (*createTenantInput, error) {
	if in == nil {
		return nil, errors.New("请求不能为空")
	}

	req := &createTenantInput{
		TenantName:        strings.TrimSpace(in.TenantName),
		TenantShortName:   strings.TrimSpace(in.TenantShortName),
		ContactName:       strings.TrimSpace(in.ContactName),
		ContactMobile:     strings.TrimSpace(in.ContactMobile),
		ContactEmail:      strings.TrimSpace(in.ContactEmail),
		AvailableChannels: normalizeStringList(in.AvailableChannels),
		DataRetentionDays: in.DataRetentionDays,
		FeatureFlags:      normalizeStringList(in.FeatureFlags),
		AdminUserName:     strings.TrimSpace(in.AdminUserName),
		AdminNickName:     strings.TrimSpace(in.AdminNickName),
		AdminMobile:       strings.TrimSpace(in.AdminMobile),
		AdminEmail:        strings.TrimSpace(in.AdminEmail),
		AdminPassword:     strings.TrimSpace(in.AdminPassword),
		CreateBy:          strings.TrimSpace(in.CreateBy),
		OperatorID:        in.OperatorId,
	}

	if req.TenantName == "" {
		return nil, errors.New("租户名称不能为空")
	}
	if req.ContactName == "" {
		return nil, errors.New("联系人不能为空")
	}
	if req.ContactMobile == "" {
		return nil, errors.New("联系人手机号不能为空")
	}
	if len(req.AvailableChannels) == 0 {
		return nil, errors.New("至少选择一个可用渠道")
	}
	if req.AdminUserName == "" {
		return nil, errors.New("管理员账号不能为空")
	}
	if req.AdminNickName == "" {
		return nil, errors.New("管理员昵称不能为空")
	}
	if req.AdminMobile == "" {
		return nil, errors.New("管理员手机号不能为空")
	}
	if req.AdminPassword == "" {
		return nil, errors.New("管理员密码不能为空")
	}
	if req.DataRetentionDays <= 0 {
		req.DataRetentionDays = defaultDataRetentionDays
	}
	if req.CreateBy == "" {
		req.CreateBy = "admin"
	}

	return req, nil
}

func encodeStringList(values []string) (string, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func decodeStringList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}

	var result []string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []string{}
	}

	return result
}

func formatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format(time.DateTime)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format(time.DateTime)
}

func mapTenantRowToProto(row tenantQueryRow) *sysclient.TenantData {
	return &sysclient.TenantData{
		Id:                    row.ID,
		TenantCode:            row.TenantCode,
		TenantName:            row.TenantName,
		TenantShortName:       row.TenantShortName,
		ContactName:           row.ContactName,
		ContactMobile:         row.ContactMobile,
		ContactEmail:          row.ContactEmail,
		AvailableChannels:     decodeStringList(row.AvailableChannels),
		DataRetentionDays:     row.DataRetentionDays,
		FeatureFlags:          decodeStringList(row.FeatureFlags),
		Status:                row.Status,
		StatusReason:          row.StatusReason,
		PrimaryAdminUserId:    row.PrimaryAdminUserID,
		PrimaryAdminUserName:  row.PrimaryAdminUserName,
		PrimaryAdminMobile:    row.PrimaryAdminMobile,
		AdminActivationStatus: activationStatusByTenantStatus(row.AdminActivationStatus),
		CreatedBy:             row.CreatedBy,
		CreatedAt:             formatTime(row.CreatedAt),
		UpdatedBy:             row.UpdatedBy,
		UpdatedAt:             formatTimePtr(row.UpdatedAt),
	}
}

func buildTenantBaseQuery(db *gorm.DB) *gorm.DB {
	return db.Table("sys_tenant AS t").
		Joins("LEFT JOIN sys_tenant_user tu ON tu.tenant_id = t.id AND tu.is_primary_admin = ?", 1).
		Joins("LEFT JOIN sys_user su ON su.id = tu.user_id")
}

func applyTenantFilters(db *gorm.DB, in *sysclient.QueryTenantListReq) *gorm.DB {
	if in == nil {
		return db
	}

	if tenantName := strings.TrimSpace(in.TenantName); tenantName != "" {
		db = db.Where("t.tenant_name LIKE ?", "%"+tenantName+"%")
	}
	if tenantCode := strings.TrimSpace(in.TenantCode); tenantCode != "" {
		db = db.Where("t.tenant_code LIKE ?", "%"+tenantCode+"%")
	}
	if in.Status >= 0 {
		db = db.Where("t.status = ?", in.Status)
	}
	if channel := strings.TrimSpace(in.Channel); channel != "" {
		db = db.Where("t.available_channels LIKE ?", "%\""+channel+"\"%")
	}

	return db
}

func normalizeIDList(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}

func createUniqueTenantCode(tx *gorm.DB, now time.Time) (string, error) {
	for i := 0; i < 8; i++ {
		candidate := generateTenantCode(now.Add(time.Duration(i) * time.Millisecond))
		var count int64
		if err := tx.Model(&tenantmodel.SysTenant{}).Where("tenant_code = ?", candidate).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}

	return "", errors.New("生成唯一租户标识失败")
}

func ensureTenantNameAvailable(tx *gorm.DB, tenantName string) error {
	var count int64
	if err := tx.Model(&tenantmodel.SysTenant{}).Where("tenant_name = ?", tenantName).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("租户名称已存在")
	}

	return nil
}

func findReusableAdminUser(tx *gorm.DB, req *createTenantInput, now time.Time) (*model.SysUser, bool, error) {
	query := tx.Model(&model.SysUser{})
	conditions := []string{"user_name = ?", "mobile = ?"}
	args := []interface{}{req.AdminUserName, req.AdminMobile}
	if req.AdminEmail != "" {
		conditions = append(conditions, "email = ?")
		args = append(args, req.AdminEmail)
	}

	var users []model.SysUser
	if err := query.Where(strings.Join(conditions, " OR "), args...).Find(&users).Error; err != nil {
		return nil, false, err
	}
	if len(users) > 1 {
		return nil, false, errors.New("管理员身份信息命中了多个既有账号，请先清理存量账号")
	}
	if len(users) == 0 {
		values := map[string]interface{}{
			"mobile":        req.AdminMobile,
			"user_name":     req.AdminUserName,
			"nick_name":     req.AdminNickName,
			"user_type":     "00",
			"avatar":        defaultTenantAdminAvatar,
			"email":         req.AdminEmail,
			"password":      req.AdminPassword,
			"status":        0,
			"dept_id":       defaultDeptID,
			"login_ip":      "",
			"login_browser": "",
			"login_os":      "",
			"remark":        defaultTenantAdminRemark,
			"del_flag":      1,
			"create_by":     req.CreateBy,
			"update_by":     req.CreateBy,
		}
		if err := tx.Model(&model.SysUser{}).Create(values).Error; err != nil {
			return nil, false, err
		}
		user := &model.SysUser{}
		if err := tx.Where("mobile = ?", req.AdminMobile).First(user).Error; err != nil {
			return nil, false, err
		}
		return user, false, nil
	}

	user := &users[0]
	if user.UserName != req.AdminUserName || user.Mobile != req.AdminMobile {
		return nil, false, errors.New("管理员账号、手机号与既有账号不一致，无法安全复用")
	}
	if req.AdminEmail != "" && user.Email != "" && user.Email != req.AdminEmail {
		return nil, false, errors.New("管理员邮箱与既有账号不一致，无法安全复用")
	}

	var roleCount int64
	if err := tx.Model(&model.SysUserRole{}).Where("user_id = ?", user.ID).Count(&roleCount).Error; err != nil {
		return nil, false, err
	}
	if roleCount > 0 {
		return nil, false, errors.New("该管理员账号已持有平台角色，不能直接复用于租户 bootstrap")
	}

	var bindingCount int64
	if err := tx.Model(&tenantmodel.SysTenantUser{}).Where("user_id = ?", user.ID).Count(&bindingCount).Error; err != nil {
		return nil, false, err
	}
	if bindingCount > 0 {
		return nil, false, errors.New("该管理员账号已绑定租户，不能重复初始化")
	}

	updates := map[string]interface{}{
		"nick_name":   req.AdminNickName,
		"password":    req.AdminPassword,
		"status":      0,
		"dept_id":     defaultDeptID,
		"remark":      defaultTenantAdminRemark,
		"update_by":   req.CreateBy,
		"update_time": now,
	}
	if req.AdminEmail != "" {
		updates["email"] = req.AdminEmail
	}
	if err := tx.Model(&model.SysUser{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return nil, false, err
	}

	user.NickName = req.AdminNickName
	user.Password = req.AdminPassword
	user.Status = 0
	user.DeptID = defaultDeptID
	user.Remark = defaultTenantAdminRemark
	user.UpdateBy = req.CreateBy
	user.UpdateTime = &now
	if req.AdminEmail != "" {
		user.Email = req.AdminEmail
	}

	return user, true, nil
}

func createTenantAuditRows(tx *gorm.DB, tenant *tenantmodel.SysTenant, user *model.SysUser, req *createTenantInput) error {
	createdPayload, err := audit.EncodeTenantPayload(audit.TenantPayload{
		TenantID:              tenant.ID,
		TenantCode:            tenant.TenantCode,
		TenantName:            tenant.TenantName,
		Action:                audit.ActionTenantCreated,
		CurrentStatus:         tenant.Status,
		AvailableChannels:     req.AvailableChannels,
		FeatureFlags:          req.FeatureFlags,
		PrimaryAdminUserID:    user.ID,
		PrimaryAdminUserName:  user.UserName,
		PrimaryAdminMobile:    user.Mobile,
		AdminActivationStatus: scope.ActivationStatusPending,
	})
	if err != nil {
		return err
	}

	adminPayload, err := audit.EncodeTenantPayload(audit.TenantPayload{
		TenantID:              tenant.ID,
		TenantCode:            tenant.TenantCode,
		TenantName:            tenant.TenantName,
		Action:                audit.ActionTenantAdminInitialized,
		CurrentStatus:         tenant.Status,
		PrimaryAdminUserID:    user.ID,
		PrimaryAdminUserName:  user.UserName,
		PrimaryAdminMobile:    user.Mobile,
		AdminActivationStatus: scope.ActivationStatusPending,
	})
	if err != nil {
		return err
	}

	records := []tenantmodel.SysTenantAudit{
		{
			PlatformID:     tenant.PlatformID,
			TenantID:       tenant.ID,
			TenantCode:     tenant.TenantCode,
			Action:         audit.ActionTenantCreated,
			AfterStatus:    tenant.Status,
			OperatorID:     req.OperatorID,
			OperatorName:   req.CreateBy,
			RequestPayload: createdPayload,
			EventPayload:   createdPayload,
			Result:         "success",
			CreatedBy:      req.CreateBy,
			UpdatedBy:      req.CreateBy,
		},
		{
			PlatformID:     tenant.PlatformID,
			TenantID:       tenant.ID,
			TenantCode:     tenant.TenantCode,
			Action:         audit.ActionTenantAdminInitialized,
			AfterStatus:    tenant.Status,
			OperatorID:     req.OperatorID,
			OperatorName:   req.CreateBy,
			RequestPayload: adminPayload,
			EventPayload:   adminPayload,
			Result:         "success",
			CreatedBy:      req.CreateBy,
			UpdatedBy:      req.CreateBy,
		},
	}

	return tx.Create(&records).Error
}

func statusAction(nextStatus int32) string {
	switch nextStatus {
	case tenantmodel.TenantStatusEnabled:
		return audit.ActionTenantEnabled
	case tenantmodel.TenantStatusDisabled:
		return audit.ActionTenantDisabled
	default:
		return audit.ActionTenantArchived
	}
}

func defaultStatusReason(nextStatus int32) string {
	switch nextStatus {
	case tenantmodel.TenantStatusEnabled:
		return "平台管理员启用了租户"
	case tenantmodel.TenantStatusDisabled:
		return "平台管理员停用了租户"
	default:
		return "平台管理员归档了租户"
	}
}

func changeTenantStatus(ctx context.Context, svcCtx *svc.ServiceContext, req *sysclient.ChangeTenantStatusReq, nextStatus int32) (*sysclient.ChangeTenantStatusResp, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}

	ids := normalizeIDList(req.Ids)
	if len(ids) == 0 {
		return nil, errors.New("至少选择一个租户")
	}

	updateBy := strings.TrimSpace(req.UpdateBy)
	if updateBy == "" {
		updateBy = "admin"
	}
	reason := strings.TrimSpace(req.StatusReason)
	if reason == "" {
		reason = defaultStatusReason(nextStatus)
	}

	now := time.Now()
	err := svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tenants []tenantmodel.SysTenant
		if err := tx.Where("id IN ?", ids).Find(&tenants).Error; err != nil {
			return err
		}
		if len(tenants) != len(ids) {
			return errors.New("存在未找到的租户记录")
		}

		tenantMap := make(map[int64]*tenantmodel.SysTenant, len(tenants))
		for i := range tenants {
			tenantMap[tenants[i].ID] = &tenants[i]
		}

		for _, id := range ids {
			tenant := tenantMap[id]
			if err := validateStatusTransition(tenant.Status, nextStatus); err != nil {
				return fmt.Errorf("租户[%d]状态流转失败: %w", id, err)
			}

			updates := map[string]interface{}{
				"status":        nextStatus,
				"status_reason": reason,
				"updated_by":    updateBy,
				"updated_at":    now,
			}
			if nextStatus == tenantmodel.TenantStatusArchived {
				updates["archived_at"] = now
			}

			if err := tx.Model(&tenantmodel.SysTenant{}).Where("id = ?", tenant.ID).Updates(updates).Error; err != nil {
				return err
			}

			var bindings []tenantmodel.SysTenantUser
			if err := tx.Where("tenant_id = ?", tenant.ID).Find(&bindings).Error; err != nil {
				return err
			}
			for _, binding := range bindings {
				scopeMetadata, err := scope.BuildTenantAdminMetadata(binding.PlatformID, tenant.ID, tenant.TenantCode, activationStatusByTenantStatus(nextStatus))
				if err != nil {
					return err
				}
				if err := tx.Model(&tenantmodel.SysTenantUser{}).Where("id = ?", binding.ID).Updates(map[string]interface{}{
					"activation_status": nextStatus,
					"scope_metadata":    scopeMetadata,
					"updated_by":        updateBy,
					"updated_at":        now,
				}).Error; err != nil {
					return err
				}
			}

			payload, err := audit.EncodeTenantPayload(audit.TenantPayload{
				TenantID:       tenant.ID,
				TenantCode:     tenant.TenantCode,
				TenantName:     tenant.TenantName,
				Action:         statusAction(nextStatus),
				PreviousStatus: tenant.Status,
				CurrentStatus:  nextStatus,
			})
			if err != nil {
				return err
			}

			beforeStatus := tenant.Status
			record := &tenantmodel.SysTenantAudit{
				PlatformID:     tenant.PlatformID,
				TenantID:       tenant.ID,
				TenantCode:     tenant.TenantCode,
				Action:         statusAction(nextStatus),
				BeforeStatus:   &beforeStatus,
				AfterStatus:    nextStatus,
				OperatorID:     req.OperatorId,
				OperatorName:   updateBy,
				RequestPayload: payload,
				EventPayload:   payload,
				Result:         "success",
				CreatedBy:      updateBy,
				UpdatedBy:      updateBy,
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.ChangeTenantStatusResp{Pong: "ok"}, nil
}
