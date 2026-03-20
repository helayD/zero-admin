package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	UserActivationDisabled int32 = 0
	UserActivationActive   int32 = 1
	UserActivationPending  int32 = 2
	UserActivationArchived int32 = 3
)

type UserScopeBinding struct {
	UserID           int64  `gorm:"column:user_id"`
	ScopeType        string `gorm:"column:scope_type"`
	PlatformID       int64  `gorm:"column:platform_id"`
	TenantID         int64  `gorm:"column:tenant_id"`
	MerchantID       int64  `gorm:"column:merchant_id"`
	DeptID           int64  `gorm:"column:dept_id"`
	ActivationStatus int32  `gorm:"column:activation_status"`
	RoleMode         string `gorm:"column:role_mode"`
}

func (*UserScopeBinding) TableName() string {
	return "sys_user_scope"
}

type UserScopeRow struct {
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

type ScopedDept struct {
	ID         int64      `gorm:"column:id"`
	ParentID   int64      `gorm:"column:parent_id"`
	Ancestors  string     `gorm:"column:ancestors"`
	DeptName   string     `gorm:"column:dept_name"`
	Sort       int32      `gorm:"column:sort"`
	Leader     string     `gorm:"column:leader"`
	Phone      string     `gorm:"column:phone"`
	Email      string     `gorm:"column:email"`
	Status     int32      `gorm:"column:status"`
	DelFlag    int32      `gorm:"column:del_flag"`
	Remark     string     `gorm:"column:remark"`
	CreateBy   string     `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   string     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}

type ScopedPost struct {
	ID         int64      `gorm:"column:id"`
	PostCode   string     `gorm:"column:post_code"`
	PostName   string     `gorm:"column:post_name"`
	Sort       int32      `gorm:"column:sort"`
	Status     int32      `gorm:"column:status"`
	Remark     string     `gorm:"column:remark"`
	CreateBy   string     `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   string     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}

type ScopedDictType struct {
	ID         int64      `gorm:"column:id"`
	DictName   string     `gorm:"column:dict_name"`
	DictType   string     `gorm:"column:dict_type"`
	Status     int32      `gorm:"column:status"`
	Remark     string     `gorm:"column:remark"`
	CreateBy   string     `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   string     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}

type ScopedDictItem struct {
	ID         int64      `gorm:"column:id"`
	DictSort   int32      `gorm:"column:dict_sort"`
	DictLabel  string     `gorm:"column:dict_label"`
	DictValue  string     `gorm:"column:dict_value"`
	DictType   string     `gorm:"column:dict_type"`
	CSSClass   string     `gorm:"column:css_class"`
	ListClass  string     `gorm:"column:list_class"`
	IsDefault  string     `gorm:"column:is_default"`
	Status     int32      `gorm:"column:status"`
	Remark     string     `gorm:"column:remark"`
	CreateBy   string     `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   string     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	DictTypeID int64      `gorm:"column:dict_type_id"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}

type ScopedRole struct {
	ID         int64  `gorm:"column:id"`
	RoleName   string `gorm:"column:role_name"`
	ScopeType  string `gorm:"column:scope_type"`
	PlatformID int64  `gorm:"column:platform_id"`
	TenantID   int64  `gorm:"column:tenant_id"`
	MerchantID int64  `gorm:"column:merchant_id"`
	Status     int32  `gorm:"column:status"`
	DelFlag    int32  `gorm:"column:del_flag"`
}

type ScopedNotice struct {
	ID            int64      `gorm:"column:id"`
	NoticeTitle   string     `gorm:"column:notice_title"`
	NoticeType    int32      `gorm:"column:notice_type"`
	NoticeContent string     `gorm:"column:notice_content"`
	Status        int32      `gorm:"column:status"`
	Remark        string     `gorm:"column:remark"`
	CreateBy      string     `gorm:"column:create_by"`
	CreateTime    time.Time  `gorm:"column:create_time"`
	UpdateBy      string     `gorm:"column:update_by"`
	UpdateTime    *time.Time `gorm:"column:update_time"`
	PlatformID    int64      `gorm:"column:platform_id"`
	TenantID      int64      `gorm:"column:tenant_id"`
	MerchantID    int64      `gorm:"column:merchant_id"`
}

type TenantStatusRow struct {
	ID     int64 `gorm:"column:id"`
	Status int32 `gorm:"column:status"`
}

func NormalizeProtoScope(input *sysclient.GovernanceScope) (scope.GovernanceScope, error) {
	if input == nil {
		return scope.NormalizeGovernanceScope(scope.SubjectTypePlatform, scope.DefaultPlatformID, 0, 0)
	}

	return scope.NormalizeGovernanceScope(input.ScopeType, input.PlatformId, input.TenantId, input.MerchantId)
}

func ProtoScope(input scope.GovernanceScope) *sysclient.GovernanceScope {
	return &sysclient.GovernanceScope{
		ScopeType:  input.ScopeType,
		PlatformId: input.PlatformID,
		TenantId:   input.TenantID,
		MerchantId: input.MerchantID,
		ScopeLabel: input.Label(),
	}
}

func DefaultScope(platformID, tenantID, merchantID int64) scope.GovernanceScope {
	normalized, _ := scope.NormalizeGovernanceScope("", platformID, tenantID, merchantID)
	return normalized
}

func ActivationStatusName(code int32) string {
	switch code {
	case UserActivationDisabled:
		return scope.ActivationStatusDisabled
	case UserActivationPending:
		return scope.ActivationStatusPending
	case UserActivationArchived:
		return scope.ActivationStatusArchived
	default:
		return scope.ActivationStatusActive
	}
}

func ActivationStatusCode(name string) int32 {
	switch strings.TrimSpace(name) {
	case scope.ActivationStatusDisabled:
		return UserActivationDisabled
	case scope.ActivationStatusPending:
		return UserActivationPending
	case scope.ActivationStatusArchived:
		return UserActivationArchived
	default:
		return UserActivationActive
	}
}

func ValidateTenantWritable(ctx context.Context, db *gorm.DB, current scope.GovernanceScope) error {
	if current.TenantID == 0 {
		return nil
	}

	var tenant TenantStatusRow
	err := db.WithContext(ctx).
		Table("sys_tenant").
		Select("id, status").
		Where("id = ?", current.TenantID).
		Take(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("指定的租户不存在")
		}
		return err
	}

	if tenant.Status == tenantmodel.TenantStatusDisabled || tenant.Status == tenantmodel.TenantStatusArchived {
		return errors.New("当前主体绑定到已停用或已归档租户，无法继续写入治理元数据")
	}

	return nil
}

func ValidateDeptInScope(ctx context.Context, db *gorm.DB, current scope.GovernanceScope, deptID int64) (*ScopedDept, error) {
	if deptID <= 0 {
		return nil, errors.New("请选择部门")
	}

	var dept ScopedDept
	err := db.WithContext(ctx).
		Table("sys_dept").
		Select("id, parent_id, ancestors, dept_name, status, platform_id, tenant_id, merchant_id").
		Where("id = ?", deptID).
		Take(&dept).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("部门不存在")
		}
		return nil, err
	}

	if !DefaultScope(dept.PlatformID, dept.TenantID, dept.MerchantID).SameScope(current) {
		return nil, errors.New("部门主体范围与当前操作上下文不一致")
	}
	if dept.Status != 1 {
		return nil, errors.New("部门已停用，无法继续关联")
	}

	return &dept, nil
}

func ValidatePostIDsInScope(ctx context.Context, db *gorm.DB, current scope.GovernanceScope, postIDs []int64) error {
	if len(postIDs) == 0 {
		return nil
	}

	var rows []ScopedPost
	err := db.WithContext(ctx).
		Table("sys_post").
		Select("id, post_code, status, platform_id, tenant_id, merchant_id").
		Where("id IN ?", postIDs).
		Find(&rows).Error
	if err != nil {
		return err
	}
	if len(rows) != len(postIDs) {
		return errors.New("存在岗位不存在或已被删除")
	}

	for _, row := range rows {
		if !DefaultScope(row.PlatformID, row.TenantID, row.MerchantID).SameScope(current) {
			return errors.New("岗位主体范围与当前操作上下文不一致")
		}
	}

	return nil
}

func ValidateRoleIDsInScope(ctx context.Context, db *gorm.DB, current scope.GovernanceScope, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}

	uniqueRoleIDs := make([]int64, 0, len(roleIDs))
	roleIDSet := make(map[int64]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID <= 0 {
			return errors.New("存在非法角色ID")
		}
		if _, ok := roleIDSet[roleID]; ok {
			continue
		}
		roleIDSet[roleID] = struct{}{}
		uniqueRoleIDs = append(uniqueRoleIDs, roleID)
	}

	var rows []ScopedRole
	err := db.WithContext(ctx).
		Table("sys_role").
		Select("id, role_name, scope_type, platform_id, tenant_id, merchant_id, status, del_flag").
		Where("id IN ?", uniqueRoleIDs).
		Find(&rows).Error
	if err != nil {
		return err
	}
	if len(rows) != len(uniqueRoleIDs) {
		return errors.New("存在角色不存在或已被删除")
	}

	for _, row := range rows {
		if row.DelFlag == 0 {
			return fmt.Errorf("角色[%s]已被删除，无法分配给用户", row.RoleName)
		}
		if row.Status != 1 {
			return fmt.Errorf("角色[%s]已停用，无法分配给用户", row.RoleName)
		}

		roleScope, scopeErr := scope.NormalizeGovernanceScope(row.ScopeType, row.PlatformID, row.TenantID, row.MerchantID)
		if scopeErr != nil {
			return fmt.Errorf("角色[%s]主体范围配置非法: %w", row.RoleName, scopeErr)
		}
		if scopeErr = scope.ValidateUserRoleAssignment(current, roleScope); scopeErr != nil {
			return fmt.Errorf("角色[%s]主体范围与当前用户不一致: %w", row.RoleName, scopeErr)
		}
	}

	return nil
}

func ResolveDictType(ctx context.Context, db *gorm.DB, current scope.GovernanceScope, dictTypeID int64, dictType string) (*ScopedDictType, error) {
	query := db.WithContext(ctx).
		Table("sys_dict_type").
		Select("id, dict_name, dict_type, status, platform_id, tenant_id, merchant_id")

	if dictTypeID > 0 {
		query = query.Where("id = ?", dictTypeID)
	} else if strings.TrimSpace(dictType) != "" {
		scopeWhere, scopeArgs := ScopeFilterSQL("", current)
		query = query.Where(scopeWhere, scopeArgs...).Where("dict_type = ?", strings.TrimSpace(dictType))
	} else {
		return nil, errors.New("字典项必须绑定字典类型")
	}

	var record ScopedDictType
	err := query.Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("字典类型不存在")
		}
		return nil, err
	}

	if !DefaultScope(record.PlatformID, record.TenantID, record.MerchantID).SameScope(current) {
		return nil, errors.New("字典项引用了其他主体的字典类型")
	}
	if record.Status != 1 {
		return nil, errors.New("字典类型已停用，无法继续关联字典项")
	}

	return &record, nil
}

func QueryPrimaryUserScope(ctx context.Context, db *gorm.DB, userID int64, fallback scope.GovernanceScope) (*UserScopeBinding, error) {
	var binding UserScopeBinding
	err := db.WithContext(ctx).
		Table("sys_user_scope").
		Where("user_id = ?", userID).
		Order("is_primary desc, id asc").
		Take(&binding).Error
	if err == nil {
		return &binding, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &UserScopeBinding{
		UserID:           userID,
		ScopeType:        fallback.ScopeType,
		PlatformID:       fallback.PlatformID,
		TenantID:         fallback.TenantID,
		MerchantID:       fallback.MerchantID,
		DeptID:           0,
		ActivationStatus: UserActivationActive,
		RoleMode:         "",
	}, nil
}

func QueryUserDefaultScope(ctx context.Context, db *gorm.DB, userID int64) (scope.GovernanceScope, error) {
	var row UserScopeRow
	err := db.WithContext(ctx).
		Table("sys_user").
		Select("platform_id, tenant_id, merchant_id").
		Where("id = ?", userID).
		Take(&row).Error
	if err != nil {
		return scope.GovernanceScope{}, err
	}

	return DefaultScope(row.PlatformID, row.TenantID, row.MerchantID), nil
}

func UpsertUserScopeBinding(ctx context.Context, tx *gorm.DB, userID, deptID int64, current scope.GovernanceScope, activationStatus, roleMode, actor string) error {
	metadata, err := scope.EncodeGovernanceScopeMetadata(current, ActivationStatusName(ActivationStatusCode(activationStatus)), strings.TrimSpace(roleMode))
	if err != nil {
		return err
	}

	row := map[string]interface{}{
		"user_id":           userID,
		"scope_type":        current.ScopeType,
		"platform_id":       current.PlatformID,
		"tenant_id":         current.TenantID,
		"merchant_id":       current.MerchantID,
		"dept_id":           deptID,
		"role_mode":         strings.TrimSpace(roleMode),
		"activation_status": ActivationStatusCode(activationStatus),
		"is_primary":        1,
		"scope_metadata":    metadata,
		"created_by":        actor,
		"updated_by":        actor,
	}

	return tx.WithContext(ctx).
		Table("sys_user_scope").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "scope_type"},
				{Name: "platform_id"},
				{Name: "tenant_id"},
				{Name: "merchant_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"dept_id":           deptID,
				"role_mode":         strings.TrimSpace(roleMode),
				"activation_status": ActivationStatusCode(activationStatus),
				"is_primary":        1,
				"scope_metadata":    metadata,
				"updated_by":        actor,
			}),
		}).
		Create(row).Error
}

func ScopeFilterSQL(alias string, current scope.GovernanceScope) (string, []interface{}) {
	prefix := alias
	if prefix != "" {
		prefix += "."
	}

	return fmt.Sprintf("%splatform_id = ? AND %stenant_id = ? AND %smerchant_id = ?", prefix, prefix, prefix), []interface{}{
		current.PlatformID,
		current.TenantID,
		current.MerchantID,
	}
}

func EnsureScopeMatch(current scope.GovernanceScope, platformID, tenantID, merchantID int64, message string) error {
	if !DefaultScope(platformID, tenantID, merchantID).SameScope(current) {
		return errors.New(message)
	}

	return nil
}
