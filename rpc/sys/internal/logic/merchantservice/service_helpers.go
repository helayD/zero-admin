package merchantservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/audit"
	"github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/gorm"
)

const (
	defaultPlatformID          int64 = 1
	defaultMerchantScopeHint         = "启用后会同步影响商户后台访问、能力包、菜单模板继承与作用域生效状态。"
	merchantPermissionCacheKey       = "zero:mall:token"
)

const merchantSelectColumns = `
m.id,
m.tenant_id,
COALESCE(t.tenant_code, '') AS tenant_code,
COALESCE(t.tenant_name, '') AS tenant_name,
COALESCE(t.status, 0) AS tenant_status,
m.merchant_code,
m.merchant_name,
m.merchant_short_name,
m.contact_name,
m.contact_mobile,
m.contact_email,
m.available_channels,
m.capability_flags,
m.review_status,
m.review_reason,
m.reviewed_by,
m.reviewed_by_name,
m.reviewed_at,
m.business_status,
m.status_reason,
m.visible_scope_hint,
m.primary_admin_user_id,
m.remark,
m.created_by,
m.created_at,
m.updated_by,
m.updated_at
`

type createMerchantInput struct {
	TenantID           int64
	MerchantName       string
	MerchantShortName  string
	MerchantCode       string
	ContactName        string
	ContactMobile      string
	ContactEmail       string
	AvailableChannels  []string
	CapabilityFlags    []string
	VisibleScopeHint   string
	PrimaryAdminUserID int64
	Remark             string
	CreateBy           string
	OperatorID         int64
}

type tenantLiteRow struct {
	ID         int64  `gorm:"column:id"`
	TenantCode string `gorm:"column:tenant_code"`
	TenantName string `gorm:"column:tenant_name"`
	Status     int32  `gorm:"column:status"`
}

type merchantQueryRow struct {
	ID                 int64      `gorm:"column:id"`
	TenantID           int64      `gorm:"column:tenant_id"`
	TenantCode         string     `gorm:"column:tenant_code"`
	TenantName         string     `gorm:"column:tenant_name"`
	TenantStatus       int32      `gorm:"column:tenant_status"`
	MerchantCode       string     `gorm:"column:merchant_code"`
	MerchantName       string     `gorm:"column:merchant_name"`
	MerchantShortName  string     `gorm:"column:merchant_short_name"`
	ContactName        string     `gorm:"column:contact_name"`
	ContactMobile      string     `gorm:"column:contact_mobile"`
	ContactEmail       string     `gorm:"column:contact_email"`
	AvailableChannels  string     `gorm:"column:available_channels"`
	CapabilityFlags    string     `gorm:"column:capability_flags"`
	ReviewStatus       int32      `gorm:"column:review_status"`
	ReviewReason       string     `gorm:"column:review_reason"`
	ReviewedBy         int64      `gorm:"column:reviewed_by"`
	ReviewedByName     string     `gorm:"column:reviewed_by_name"`
	ReviewedAt         *time.Time `gorm:"column:reviewed_at"`
	BusinessStatus     int32      `gorm:"column:business_status"`
	StatusReason       string     `gorm:"column:status_reason"`
	VisibleScopeHint   string     `gorm:"column:visible_scope_hint"`
	PrimaryAdminUserID int64      `gorm:"column:primary_admin_user_id"`
	Remark             string     `gorm:"column:remark"`
	CreatedBy          string     `gorm:"column:created_by"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedBy          string     `gorm:"column:updated_by"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

type merchantScopeBindingRow struct {
	ID       int64  `gorm:"column:id"`
	UserID   int64  `gorm:"column:user_id"`
	RoleMode string `gorm:"column:role_mode"`
}

func normalizeCreateMerchantReq(in *sysclient.CreateMerchantReq) (*createMerchantInput, error) {
	if in == nil {
		return nil, errors.New("请求不能为空")
	}

	req := &createMerchantInput{
		TenantID:           in.TenantId,
		MerchantName:       strings.TrimSpace(in.MerchantName),
		MerchantShortName:  strings.TrimSpace(in.MerchantShortName),
		MerchantCode:       strings.TrimSpace(in.MerchantCode),
		ContactName:        strings.TrimSpace(in.ContactName),
		ContactMobile:      strings.TrimSpace(in.ContactMobile),
		ContactEmail:       strings.TrimSpace(in.ContactEmail),
		AvailableChannels:  normalizeChannelList(in.AvailableChannels),
		CapabilityFlags:    normalizeStringList(in.CapabilityFlags),
		VisibleScopeHint:   strings.TrimSpace(in.VisibleScopeHint),
		PrimaryAdminUserID: in.PrimaryAdminUserId,
		Remark:             strings.TrimSpace(in.Remark),
		CreateBy:           strings.TrimSpace(in.CreateBy),
		OperatorID:         in.OperatorId,
	}

	if req.TenantID <= 0 {
		return nil, errors.New("请选择归属租户")
	}
	if req.MerchantName == "" {
		return nil, errors.New("商户名称不能为空")
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
	if len(req.CapabilityFlags) == 0 {
		return nil, errors.New("至少选择一个能力包")
	}
	if req.VisibleScopeHint == "" {
		req.VisibleScopeHint = defaultMerchantScopeHint
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
	if value == nil || value.IsZero() {
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

func nextActionsForMerchant(reviewStatus, businessStatus int32) []string {
	switch {
	case reviewStatus == merchantmodel.MerchantReviewApproved && businessStatus == merchantmodel.MerchantBusinessPendingActivation:
		return []string{"启用商户主体", "分配菜单模板与作用域", "绑定商户管理员并引导首次登录"}
	case reviewStatus == merchantmodel.MerchantReviewApproved && businessStatus == merchantmodel.MerchantBusinessEnabled:
		return []string{"继续分配菜单模板/能力包", "绑定或确认商户管理员", "引导补全店铺资料"}
	case reviewStatus == merchantmodel.MerchantReviewMaterialRequired:
		return []string{"补充材料后重新审核", "核对租户归属与能力包", "确认是否继续保留当前申请"}
	case reviewStatus == merchantmodel.MerchantReviewRejected:
		return []string{"修正主体信息后重新提交", "核对拒绝原因与租户归属", "必要时归档当前申请"}
	default:
		return []string{"执行审核结论", "核对租户归属与能力包", "准备后续启停治理动作"}
	}
}

func mapMerchantRowToProto(row merchantQueryRow) *sysclient.MerchantData {
	return &sysclient.MerchantData{
		Id:                 row.ID,
		TenantId:           row.TenantID,
		TenantCode:         row.TenantCode,
		TenantName:         row.TenantName,
		TenantStatus:       row.TenantStatus,
		MerchantCode:       row.MerchantCode,
		MerchantName:       row.MerchantName,
		MerchantShortName:  row.MerchantShortName,
		ContactName:        row.ContactName,
		ContactMobile:      row.ContactMobile,
		ContactEmail:       row.ContactEmail,
		AvailableChannels:  decodeChannelList(row.AvailableChannels),
		CapabilityFlags:    decodeStringList(row.CapabilityFlags),
		ReviewStatus:       row.ReviewStatus,
		ReviewReason:       row.ReviewReason,
		ReviewedBy:         row.ReviewedBy,
		ReviewedByName:     row.ReviewedByName,
		ReviewedAt:         formatTimePtr(row.ReviewedAt),
		BusinessStatus:     row.BusinessStatus,
		StatusReason:       row.StatusReason,
		VisibleScopeHint:   row.VisibleScopeHint,
		PrimaryAdminUserId: row.PrimaryAdminUserID,
		Remark:             row.Remark,
		NextActions:        nextActionsForMerchant(row.ReviewStatus, row.BusinessStatus),
		CreatedBy:          row.CreatedBy,
		CreatedAt:          formatTime(row.CreatedAt),
		UpdatedBy:          row.UpdatedBy,
		UpdatedAt:          formatTimePtr(row.UpdatedAt),
	}
}

func buildMerchantBaseQuery(db *gorm.DB) *gorm.DB {
	return db.Table("sys_merchant AS m").
		Joins("LEFT JOIN sys_tenant t ON t.id = m.tenant_id")
}

func applyMerchantFilters(db *gorm.DB, in *sysclient.QueryMerchantListReq) *gorm.DB {
	if in == nil {
		return db
	}

	if in.TenantId > 0 {
		db = db.Where("m.tenant_id = ?", in.TenantId)
	}
	if merchantName := strings.TrimSpace(in.MerchantName); merchantName != "" {
		db = db.Where("m.merchant_name LIKE ?", "%"+merchantName+"%")
	}
	if merchantCode := strings.TrimSpace(in.MerchantCode); merchantCode != "" {
		db = db.Where("m.merchant_code LIKE ?", "%"+merchantCode+"%")
	}
	if in.ReviewStatus >= 0 {
		db = db.Where("m.review_status = ?", in.ReviewStatus)
	}
	if in.BusinessStatus >= 0 {
		db = db.Where("m.business_status = ?", in.BusinessStatus)
	}
	if channel := normalizeChannelFilterValue(in.Channel); channel != "" {
		db = db.Where("m.available_channels LIKE ?", "%\""+channel+"\"%")
	}
	if capabilityFlag := strings.TrimSpace(in.CapabilityFlag); capabilityFlag != "" {
		db = db.Where("m.capability_flags LIKE ?", "%\""+capabilityFlag+"\"%")
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

func createUniqueMerchantCode(tx *gorm.DB, tenantID int64, now time.Time) (string, error) {
	for i := 0; i < 8; i++ {
		candidate := generateMerchantCode(now.Add(time.Duration(i) * time.Millisecond))
		var count int64
		if err := tx.Model(&merchantmodel.SysMerchant{}).Where("tenant_id = ? AND merchant_code = ?", tenantID, candidate).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}

	return "", errors.New("生成唯一商户编码失败")
}

func ensureMerchantNameAvailable(tx *gorm.DB, tenantID int64, merchantName string) error {
	var count int64
	if err := tx.Model(&merchantmodel.SysMerchant{}).Where("tenant_id = ? AND merchant_name = ?", tenantID, merchantName).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("同一租户下商户名称已存在")
	}

	return nil
}

func ensureMerchantCodeAvailable(tx *gorm.DB, tenantID int64, merchantCode string) error {
	if strings.TrimSpace(merchantCode) == "" {
		return nil
	}

	var count int64
	if err := tx.Model(&merchantmodel.SysMerchant{}).Where("tenant_id = ? AND merchant_code = ?", tenantID, merchantCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("商户编码已存在")
	}

	return nil
}

func ensureTenantExists(ctx context.Context, db *gorm.DB, tenantID int64) (*tenantLiteRow, error) {
	var tenant tenantLiteRow
	err := db.WithContext(ctx).
		Table("sys_tenant").
		Select("id, tenant_code, tenant_name, status").
		Where("id = ?", tenantID).
		Take(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("归属租户不存在")
		}
		return nil, err
	}

	return &tenant, nil
}

func ensureTenantActiveForMerchant(ctx context.Context, db *gorm.DB, tenantID int64) (*tenantLiteRow, error) {
	tenant, err := ensureTenantExists(ctx, db, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant.Status == tenantmodel.TenantStatusDisabled || tenant.Status == tenantmodel.TenantStatusArchived {
		return nil, errors.New("目标租户已停用或已归档，不能继续审核通过或启用商户")
	}

	return tenant, nil
}

func ensurePrimaryAdminUserExists(tx *gorm.DB, userID int64) error {
	if userID <= 0 {
		return nil
	}

	var count int64
	if err := tx.Table("sys_user").Where("id = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("指定的商户管理员不存在")
	}

	return nil
}

func upsertPrimaryAdminBinding(ctx context.Context, tx *gorm.DB, merchant *merchantmodel.SysMerchant, activationStatus, actor string) error {
	if merchant.PrimaryAdminUserID <= 0 {
		return nil
	}

	currentScope, err := scope.NormalizeGovernanceScope(scope.SubjectTypeMerchant, merchant.PlatformID, merchant.TenantID, merchant.ID)
	if err != nil {
		return err
	}

	return logiccommon.UpsertUserScopeBinding(ctx, tx, merchant.PrimaryAdminUserID, 0, currentScope, activationStatus, scope.RoleModeBootstrap, actor)
}

func reviewAction(nextStatus int32) string {
	switch nextStatus {
	case merchantmodel.MerchantReviewApproved:
		return audit.ActionMerchantApproved
	case merchantmodel.MerchantReviewRejected:
		return audit.ActionMerchantRejected
	default:
		return audit.ActionMerchantMaterialRequested
	}
}

func defaultReviewReason(nextStatus int32) string {
	switch nextStatus {
	case merchantmodel.MerchantReviewApproved:
		return "平台管理员审核通过商户入驻申请"
	case merchantmodel.MerchantReviewRejected:
		return "平台管理员驳回商户入驻申请"
	default:
		return "平台管理员要求商户补充材料"
	}
}

func businessAction(nextStatus int32) string {
	switch nextStatus {
	case merchantmodel.MerchantBusinessEnabled:
		return audit.ActionMerchantEnabled
	case merchantmodel.MerchantBusinessDisabled:
		return audit.ActionMerchantDisabled
	default:
		return audit.ActionMerchantArchived
	}
}

func defaultBusinessReason(nextStatus int32) string {
	switch nextStatus {
	case merchantmodel.MerchantBusinessEnabled:
		return "平台管理员启用了商户主体"
	case merchantmodel.MerchantBusinessDisabled:
		return "平台管理员停用了商户主体"
	default:
		return "平台管理员归档了商户主体"
	}
}

func buildTraceID(merchantID int64, now time.Time) string {
	return fmt.Sprintf("merchant-%d-%d", merchantID, now.UnixNano())
}

func recordMerchantAudit(
	tx *gorm.DB,
	merchant *merchantmodel.SysMerchant,
	traceID, action, requestPayload, eventPayload string,
	beforeReviewStatus, afterReviewStatus, beforeBusinessStatus, afterBusinessStatus *int32,
	operatorID int64,
	operatorName string,
) error {
	record := &merchantmodel.SysMerchantAudit{
		TraceID:              traceID,
		PlatformID:           merchant.PlatformID,
		TenantID:             merchant.TenantID,
		MerchantID:           merchant.ID,
		MerchantCode:         merchant.MerchantCode,
		Action:               action,
		BeforeReviewStatus:   beforeReviewStatus,
		AfterReviewStatus:    afterReviewStatus,
		BeforeBusinessStatus: beforeBusinessStatus,
		AfterBusinessStatus:  afterBusinessStatus,
		OperatorID:           operatorID,
		OperatorName:         operatorName,
		RequestPayload:       requestPayload,
		EventPayload:         eventPayload,
		Result:               "success",
		CreatedBy:            operatorName,
		UpdatedBy:            operatorName,
	}

	return tx.Create(record).Error
}

func updateMerchantScopeBindings(ctx context.Context, tx *gorm.DB, merchant *merchantmodel.SysMerchant, actor string) error {
	var bindings []merchantScopeBindingRow
	if err := tx.WithContext(ctx).
		Table("sys_user_scope").
		Select("id, user_id, role_mode").
		Where("scope_type = ? AND merchant_id = ?", scope.SubjectTypeMerchant, merchant.ID).
		Find(&bindings).Error; err != nil {
		return err
	}

	currentScope, err := scope.NormalizeGovernanceScope(scope.SubjectTypeMerchant, merchant.PlatformID, merchant.TenantID, merchant.ID)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, binding := range bindings {
		metadata, err := scope.EncodeGovernanceScopeMetadata(currentScope, scopeActivationByBusinessStatus(merchant.BusinessStatus), strings.TrimSpace(binding.RoleMode))
		if err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Table("sys_user_scope").
			Where("id = ?", binding.ID).
			Updates(map[string]interface{}{
				"activation_status": logiccommon.ActivationStatusCode(scopeActivationByBusinessStatus(merchant.BusinessStatus)),
				"scope_metadata":    metadata,
				"updated_by":        actor,
				"updated_at":        now,
			}).Error; err != nil {
			return err
		}
	}

	return upsertPrimaryAdminBinding(ctx, tx, merchant, scopeActivationByBusinessStatus(merchant.BusinessStatus), actor)
}

func listMerchantRelatedUserIDs(ctx context.Context, tx *gorm.DB, merchantID int64) ([]int64, error) {
	userIDs := make([]int64, 0)
	if err := tx.WithContext(ctx).
		Table("sys_user_scope").
		Where("merchant_id = ?", merchantID).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}

	var defaultScopeUserIDs []int64
	if err := tx.WithContext(ctx).
		Table("sys_user").
		Where("merchant_id = ?", merchantID).
		Distinct("id").
		Pluck("id", &defaultScopeUserIDs).Error; err != nil {
		return nil, err
	}

	return mergeUserIDs(userIDs, defaultScopeUserIDs), nil
}

func mergeUserIDs(groups ...[]int64) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0)
	for _, group := range groups {
		for _, id := range group {
			if id <= 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}

	return result
}

func invalidatePermissionCache(ctx context.Context, svcCtx *svc.ServiceContext, userIDs []int64) {
	if svcCtx.Redis == nil || len(userIDs) == 0 {
		return
	}

	fields := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		fields = append(fields, strconv.FormatInt(userID, 10))
	}
	_, _ = svcCtx.Redis.HdelCtx(ctx, merchantPermissionCacheKey, fields...)
}

func changeMerchantReviewStatus(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	req *sysclient.ReviewMerchantReq,
	nextReviewStatus int32,
) (*sysclient.ReviewMerchantResp, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}

	ids := normalizeIDList(req.Ids)
	if len(ids) == 0 {
		return nil, errors.New("至少选择一个商户")
	}

	updateBy := strings.TrimSpace(req.UpdateBy)
	if updateBy == "" {
		updateBy = "admin"
	}
	reason := strings.TrimSpace(req.ReviewReason)
	if reason == "" {
		reason = defaultReviewReason(nextReviewStatus)
	}

	err := svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var merchants []merchantmodel.SysMerchant
		if err := tx.Where("id IN ?", ids).Find(&merchants).Error; err != nil {
			return err
		}
		if len(merchants) != len(ids) {
			return errors.New("存在未找到的商户记录")
		}

		merchantMap := make(map[int64]*merchantmodel.SysMerchant, len(merchants))
		for i := range merchants {
			merchantMap[merchants[i].ID] = &merchants[i]
		}

		now := time.Now()
		for _, id := range ids {
			merchant := merchantMap[id]
			if err := validateReviewTransition(merchant.ReviewStatus, nextReviewStatus); err != nil {
				return fmt.Errorf("商户[%d]审核流转失败: %w", id, err)
			}
			if nextReviewStatus == merchantmodel.MerchantReviewApproved {
				if _, err := ensureTenantActiveForMerchant(ctx, tx, merchant.TenantID); err != nil {
					return err
				}
			}

			beforeReviewStatus := merchant.ReviewStatus
			beforeBusinessStatus := merchant.BusinessStatus
			updates := map[string]interface{}{
				"review_status":    nextReviewStatus,
				"review_reason":    reason,
				"reviewed_by":      req.OperatorId,
				"reviewed_by_name": updateBy,
				"reviewed_at":      now,
				"updated_by":       updateBy,
				"updated_at":       now,
			}
			switch nextReviewStatus {
			case merchantmodel.MerchantReviewApproved:
				updates["status_reason"] = "审核通过，待平台启用"
			case merchantmodel.MerchantReviewRejected:
				updates["status_reason"] = "审核驳回，商户主体未启用"
			default:
				updates["status_reason"] = "待补充材料，商户主体未启用"
			}

			if err := tx.Model(&merchantmodel.SysMerchant{}).Where("id = ?", merchant.ID).Updates(updates).Error; err != nil {
				return err
			}

			merchant.ReviewStatus = nextReviewStatus
			merchant.ReviewReason = reason
			merchant.ReviewedBy = req.OperatorId
			merchant.ReviewedByName = updateBy
			merchant.ReviewedAt = &now
			merchant.StatusReason = updates["status_reason"].(string)
			merchant.UpdatedBy = updateBy
			merchant.UpdatedAt = &now

			if nextReviewStatus == merchantmodel.MerchantReviewApproved {
				if err := upsertPrimaryAdminBinding(ctx, tx, merchant, scope.ActivationStatusPending, updateBy); err != nil {
					return err
				}
			}

			traceID := buildTraceID(merchant.ID, now)
			payload, err := audit.EncodeMerchantPayload(audit.MerchantPayload{
				TraceID:                traceID,
				TenantID:               merchant.TenantID,
				MerchantID:             merchant.ID,
				MerchantCode:           merchant.MerchantCode,
				MerchantName:           merchant.MerchantName,
				Action:                 reviewAction(nextReviewStatus),
				PreviousReviewStatus:   beforeReviewStatus,
				CurrentReviewStatus:    nextReviewStatus,
				PreviousBusinessStatus: beforeBusinessStatus,
				CurrentBusinessStatus:  merchant.BusinessStatus,
				AvailableChannels:      decodeStringList(merchant.AvailableChannels),
				CapabilityFlags:        decodeStringList(merchant.CapabilityFlags),
				PrimaryAdminUserID:     merchant.PrimaryAdminUserID,
				VisibleScopeHint:       merchant.VisibleScopeHint,
				Result:                 "success",
			})
			if err != nil {
				return err
			}

			afterReviewStatus := nextReviewStatus
			afterBusinessStatus := merchant.BusinessStatus
			if err := recordMerchantAudit(tx, merchant, traceID, reviewAction(nextReviewStatus), payload, payload, &beforeReviewStatus, &afterReviewStatus, &beforeBusinessStatus, &afterBusinessStatus, req.OperatorId, updateBy); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.ReviewMerchantResp{Pong: "ok"}, nil
}

func changeMerchantBusinessStatus(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	req *sysclient.ChangeMerchantStatusReq,
	nextBusinessStatus int32,
) (*sysclient.ChangeMerchantStatusResp, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}

	ids := normalizeIDList(req.Ids)
	if len(ids) == 0 {
		return nil, errors.New("至少选择一个商户")
	}

	updateBy := strings.TrimSpace(req.UpdateBy)
	if updateBy == "" {
		updateBy = "admin"
	}
	reason := strings.TrimSpace(req.StatusReason)
	if reason == "" {
		reason = defaultBusinessReason(nextBusinessStatus)
	}

	userIDsToInvalidate := make([]int64, 0)
	err := svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var merchants []merchantmodel.SysMerchant
		if err := tx.Where("id IN ?", ids).Find(&merchants).Error; err != nil {
			return err
		}
		if len(merchants) != len(ids) {
			return errors.New("存在未找到的商户记录")
		}

		merchantMap := make(map[int64]*merchantmodel.SysMerchant, len(merchants))
		for i := range merchants {
			merchantMap[merchants[i].ID] = &merchants[i]
		}

		now := time.Now()
		for _, id := range ids {
			merchant := merchantMap[id]
			if err := validateBusinessTransition(merchant.ReviewStatus, merchant.BusinessStatus, nextBusinessStatus); err != nil {
				return fmt.Errorf("商户[%d]状态流转失败: %w", id, err)
			}
			if nextBusinessStatus == merchantmodel.MerchantBusinessEnabled {
				if _, err := ensureTenantActiveForMerchant(ctx, tx, merchant.TenantID); err != nil {
					return err
				}
			}

			beforeReviewStatus := merchant.ReviewStatus
			beforeBusinessStatus := merchant.BusinessStatus
			updates := map[string]interface{}{
				"business_status": nextBusinessStatus,
				"status_reason":   reason,
				"updated_by":      updateBy,
				"updated_at":      now,
			}
			if nextBusinessStatus == merchantmodel.MerchantBusinessArchived {
				updates["archived_at"] = now
			}

			if err := tx.Model(&merchantmodel.SysMerchant{}).Where("id = ?", merchant.ID).Updates(updates).Error; err != nil {
				return err
			}

			merchant.BusinessStatus = nextBusinessStatus
			merchant.StatusReason = reason
			merchant.UpdatedBy = updateBy
			merchant.UpdatedAt = &now
			if nextBusinessStatus == merchantmodel.MerchantBusinessArchived {
				merchant.ArchivedAt = &now
			}

			if err := updateMerchantScopeBindings(ctx, tx, merchant, updateBy); err != nil {
				return err
			}

			relatedUserIDs, err := listMerchantRelatedUserIDs(ctx, tx, merchant.ID)
			if err != nil {
				return err
			}
			userIDsToInvalidate = mergeUserIDs(userIDsToInvalidate, relatedUserIDs)

			traceID := buildTraceID(merchant.ID, now)
			payload, err := audit.EncodeMerchantPayload(audit.MerchantPayload{
				TraceID:                traceID,
				TenantID:               merchant.TenantID,
				MerchantID:             merchant.ID,
				MerchantCode:           merchant.MerchantCode,
				MerchantName:           merchant.MerchantName,
				Action:                 businessAction(nextBusinessStatus),
				PreviousReviewStatus:   beforeReviewStatus,
				CurrentReviewStatus:    merchant.ReviewStatus,
				PreviousBusinessStatus: beforeBusinessStatus,
				CurrentBusinessStatus:  nextBusinessStatus,
				AvailableChannels:      decodeStringList(merchant.AvailableChannels),
				CapabilityFlags:        decodeStringList(merchant.CapabilityFlags),
				PrimaryAdminUserID:     merchant.PrimaryAdminUserID,
				VisibleScopeHint:       merchant.VisibleScopeHint,
				Result:                 "success",
			})
			if err != nil {
				return err
			}

			afterReviewStatus := merchant.ReviewStatus
			afterBusinessStatus := nextBusinessStatus
			if err := recordMerchantAudit(tx, merchant, traceID, businessAction(nextBusinessStatus), payload, payload, &beforeReviewStatus, &afterReviewStatus, &beforeBusinessStatus, &afterBusinessStatus, req.OperatorId, updateBy); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	invalidatePermissionCache(ctx, svcCtx, userIDsToInvalidate)
	return &sysclient.ChangeMerchantStatusResp{Pong: "ok"}, nil
}
