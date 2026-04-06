package channelintegrationtemplateservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/audit"
	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"gorm.io/gorm"
)

type channelIntegrationBindingRow struct {
	ID                  int64      `gorm:"column:id;primaryKey"`
	TemplateID          int64      `gorm:"column:template_id"`
	SubjectType         string     `gorm:"column:subject_type"`
	PlatformID          int64      `gorm:"column:platform_id"`
	TenantID            int64      `gorm:"column:tenant_id"`
	MerchantID          int64      `gorm:"column:merchant_id"`
	TargetCode          string     `gorm:"column:target_code"`
	BindingStatus       string     `gorm:"column:binding_status"`
	BindingSource       string     `gorm:"column:binding_source"`
	EffectScopeSnapshot string     `gorm:"column:effect_scope_snapshot"`
	PublishedBy         string     `gorm:"column:published_by"`
	PublishedAt         *time.Time `gorm:"column:published_at"`
	Remark              string     `gorm:"column:remark"`
	CreateBy            string     `gorm:"column:create_by"`
	CreateTime          time.Time  `gorm:"column:create_time"`
	UpdateBy            string     `gorm:"column:update_by"`
	UpdateTime          *time.Time `gorm:"column:update_time"`
}

func (channelIntegrationBindingRow) TableName() string {
	return "sys_channel_integration_binding"
}

type templateImpactScopeConfig struct {
	SubjectTypes     []string `json:"subjectTypes,omitempty"`
	RequiresBinding  bool     `json:"requiresBinding,omitempty"`
	TenantIDs        []int64  `json:"tenantIds,omitempty"`
	MerchantIDs      []int64  `json:"merchantIds,omitempty"`
	BindingTenantIDs []int64  `json:"bindingTenantIds,omitempty"`
}

type templateImpactSummary struct {
	BindingTenantCount   int64    `json:"bindingTenantCount"`
	BindingMerchantCount int64    `json:"bindingMerchantCount"`
	BindingSampleLabels  []string `json:"bindingSampleLabels,omitempty"`
	LastStatusChangeTime string   `json:"lastStatusChangeTime,omitempty"`
	StatusReason         string   `json:"statusReason,omitempty"`
}

type bindingSubjectRow struct {
	SubjectID         int64  `gorm:"column:subject_id"`
	TenantID          int64  `gorm:"column:tenant_id"`
	SubjectCode       string `gorm:"column:subject_code"`
	SubjectName       string `gorm:"column:subject_name"`
	AvailableChannels string `gorm:"column:available_channels"`
}

type effectScopeSnapshot struct {
	SubjectType  string `json:"subjectType"`
	SubjectID    int64  `json:"subjectId"`
	SubjectCode  string `json:"subjectCode,omitempty"`
	SubjectName  string `json:"subjectName,omitempty"`
	SubjectLabel string `json:"subjectLabel,omitempty"`
	TemplateCode string `json:"templateCode,omitempty"`
	TargetCode   string `json:"targetCode,omitempty"`
	ScopeType    string `json:"scopeType,omitempty"`
	PlatformID   int64  `json:"platformId,omitempty"`
	TenantID     int64  `json:"tenantId,omitempty"`
	MerchantID   int64  `json:"merchantId,omitempty"`
}

const bindingSourceTemplateSync = "template_sync"

func ensureTemplateScopeSubjectExists(ctx context.Context, db *gorm.DB, scopeType string, tenantID, merchantID int64) error {
	switch scopeType {
	case "platform":
		return nil
	case "tenant":
		var count int64
		if err := db.WithContext(ctx).Table("sys_tenant").Where("id = ?", tenantID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("指定的租户不存在")
		}
		return nil
	case "merchant":
		if !db.Migrator().HasTable(&merchantmodel.SysMerchant{}) {
			return errors.New("当前环境未初始化商户主体表")
		}
		var merchant merchantTenantRow
		if err := db.WithContext(ctx).Table("sys_merchant").Select("id, tenant_id").Where("id = ?", merchantID).Take(&merchant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("指定的商户不存在")
			}
			return err
		}
		if merchant.TenantID != tenantID {
			return errors.New("商户不属于指定租户")
		}
		return nil
	default:
		return errors.New("模板作用域仅支持 platform、tenant、merchant")
	}
}

type merchantTenantRow struct {
	ID       int64 `gorm:"column:id"`
	TenantID int64 `gorm:"column:tenant_id"`
}

func requiresCredentialRefs(templateType, targetCode string) bool {
	if templateType == channeltemplate.TemplateTypeChannel {
		return targetCode == channeltemplate.TargetMiniProgram
	}
	switch targetCode {
	case channeltemplate.TargetLogisticsTracking, channeltemplate.TargetSMSProvider:
		return true
	default:
		return false
	}
}

func validateCredentialRefsRequired(templateType, targetCode, secretRefConfig string) error {
	if !requiresCredentialRefs(templateType, targetCode) {
		return nil
	}
	secretRefs := make(map[string]string)
	if err := json.Unmarshal([]byte(secretRefConfig), &secretRefs); err != nil {
		return errors.New("敏感配置引用必须是 string map")
	}
	if len(secretRefs) == 0 {
		return errors.New("当前模板必须配置至少一个 credentialRef")
	}
	return nil
}

func rejectSensitivePlaintext(metadataConfig, impactScopeConfig string) error {
	for _, item := range []struct {
		label string
		raw   string
	}{
		{label: "元数据配置", raw: metadataConfig},
		{label: "影响范围配置", raw: impactScopeConfig},
	} {
		if err := rejectSensitivePlaintextInJSON(item.raw); err != nil {
			return fmt.Errorf("%s存在明文敏感字段: %w", item.label, err)
		}
	}
	return nil
}

func rejectSensitivePlaintextInJSON(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "{}" {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		return nil
	}
	return walkSensitiveJSON("", value)
}

func walkSensitiveJSON(parentKey string, value interface{}) error {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			if err := walkSensitiveJSON(key, nested); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, nested := range typed {
			if err := walkSensitiveJSON(parentKey, nested); err != nil {
				return err
			}
		}
	case string:
		if looksLikeSensitiveKey(parentKey) {
			trimmed := strings.TrimSpace(typed)
			if trimmed != "" && !hasAllowedSecretRefPrefixLocal(trimmed) {
				return fmt.Errorf("字段[%s]禁止保存明文敏感值", parentKey)
			}
		}
	}
	return nil
}

func hasAllowedSecretRefPrefixLocal(value string) bool {
	prefixes := []string{"credential://", "secret://", "vault://", "kms://", "ref:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func looksLikeSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	keywords := []string{"secret", "privatekey", "private_key", "token", "accesskey", "access_key", "appsecret", "signature"}
	for _, keyword := range keywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func syncTemplateBindings(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow, actor string) error {
	impactConfig, err := parseTemplateImpactScopeConfig(template.ImpactScopeConfig)
	if err != nil {
		return err
	}
	bindings, err := resolveDesiredBindings(ctx, tx, template, impactConfig, actor)
	if err != nil {
		return err
	}
	bindingMap := make(map[string]channelIntegrationBindingRow, len(bindings))
	for _, item := range bindings {
		bindingMap[bindingIdentity(item.SubjectType, item.PlatformID, item.TenantID, item.MerchantID, item.TargetCode)] = item
	}

	existing := make([]channelIntegrationBindingRow, 0)
	if err := tx.WithContext(ctx).Where("template_id = ?", template.ID).Find(&existing).Error; err != nil {
		return err
	}

	now := time.Now()
	for _, row := range bindings {
		var current channelIntegrationBindingRow
		err := tx.WithContext(ctx).
			Where("template_id = ? AND subject_type = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND target_code = ?", row.TemplateID, row.SubjectType, row.PlatformID, row.TenantID, row.MerchantID, row.TargetCode).
			Take(&current).Error
		switch {
		case err == nil:
			updates := map[string]interface{}{
				"binding_status":        row.BindingStatus,
				"binding_source":        row.BindingSource,
				"effect_scope_snapshot": row.EffectScopeSnapshot,
				"published_by":          row.PublishedBy,
				"published_at":          row.PublishedAt,
				"remark":                row.Remark,
				"update_by":             actor,
				"update_time":           now,
			}
			if err := tx.WithContext(ctx).Model(&channelIntegrationBindingRow{}).Where("id = ?", current.ID).Updates(updates).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
				return err
			}
		default:
			return err
		}
	}

	for _, row := range existing {
		identity := bindingIdentity(row.SubjectType, row.PlatformID, row.TenantID, row.MerchantID, row.TargetCode)
		if _, ok := bindingMap[identity]; ok {
			continue
		}
		nextStatus := channeltemplate.StatusDisabled
		if template.Status == channeltemplate.StatusArchived {
			nextStatus = channeltemplate.StatusArchived
		}
		if err := tx.WithContext(ctx).Model(&channelIntegrationBindingRow{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
			"binding_status": nextStatus,
			"update_by":      actor,
			"update_time":    now,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func resolveDesiredBindings(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow, impactConfig templateImpactScopeConfig, actor string) ([]channelIntegrationBindingRow, error) {
	subjectTypes := resolveTemplateSubjectTypes(template.ScopeType, impactConfig.SubjectTypes)
	now := time.Now()
	bindings := make([]channelIntegrationBindingRow, 0)
	for _, subjectType := range subjectTypes {
		subjects, err := queryBindingSubjects(ctx, tx, template, impactConfig, subjectType)
		if err != nil {
			return nil, err
		}
		for _, subject := range subjects {
			snapshot, err := encodeEffectScopeSnapshot(template, subjectType, subject)
			if err != nil {
				return nil, err
			}
			binding := channelIntegrationBindingRow{
				TemplateID:          template.ID,
				SubjectType:         subjectType,
				PlatformID:          template.PlatformID,
				TenantID:            subject.TenantID,
				MerchantID:          0,
				TargetCode:          template.TargetCode,
				BindingStatus:       template.Status,
				BindingSource:       bindingSourceTemplateSync,
				EffectScopeSnapshot: snapshot,
				Remark:              template.Remark,
				CreateBy:            actor,
				UpdateBy:            actor,
				CreateTime:          now,
			}
			if template.Status == channeltemplate.StatusEnabled {
				binding.PublishedBy = actor
				binding.PublishedAt = &now
			}
			if subjectType == "merchant" {
				binding.MerchantID = subject.SubjectID
			}
			if subjectType == "tenant" {
				binding.TenantID = subject.SubjectID
			}
			bindings = append(bindings, binding)
		}
	}
	return bindings, nil
}

func resolveTemplateSubjectTypes(scopeType string, subjectTypes []string) []string {
	normalized := normalizeSubjectTypes(subjectTypes)
	if len(normalized) > 0 {
		if scopeType == "merchant" {
			result := make([]string, 0, len(normalized))
			for _, item := range normalized {
				if item == "merchant" {
					result = append(result, item)
				}
			}
			if len(result) > 0 {
				return result
			}
			return []string{"merchant"}
		}
		return normalized
	}
	switch scopeType {
	case "merchant":
		return []string{"merchant"}
	case "tenant":
		return []string{"tenant", "merchant"}
	default:
		return []string{"tenant", "merchant"}
	}
}

func normalizeSubjectTypes(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, item := range values {
		normalized := strings.TrimSpace(item)
		if normalized != "tenant" && normalized != "merchant" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func parseTemplateImpactScopeConfig(raw string) (templateImpactScopeConfig, error) {
	config := templateImpactScopeConfig{}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "{}" {
		return config, nil
	}
	if err := json.Unmarshal([]byte(trimmed), &config); err != nil {
		return templateImpactScopeConfig{}, fmt.Errorf("影响范围配置非法: %w", err)
	}
	config.SubjectTypes = normalizeSubjectTypes(config.SubjectTypes)
	config.TenantIDs = uniquePositiveInt64(config.TenantIDs)
	config.MerchantIDs = uniquePositiveInt64(config.MerchantIDs)
	config.BindingTenantIDs = uniquePositiveInt64(config.BindingTenantIDs)
	return config, nil
}

func uniquePositiveInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func queryBindingSubjects(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow, impactConfig templateImpactScopeConfig, subjectType string) ([]bindingSubjectRow, error) {
	switch subjectType {
	case "tenant":
		return queryTenantBindingSubjects(ctx, tx, template, impactConfig)
	case "merchant":
		return queryMerchantBindingSubjects(ctx, tx, template, impactConfig)
	default:
		return nil, fmt.Errorf("不支持的绑定主体类型: %s", subjectType)
	}
}

func queryTenantBindingSubjects(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow, impactConfig templateImpactScopeConfig) ([]bindingSubjectRow, error) {
	q := tx.WithContext(ctx).Table("sys_tenant").Select("id AS subject_id, id AS tenant_id, tenant_code AS subject_code, tenant_name AS subject_name, available_channels").Where("status <> ?", tenantmodel.TenantStatusArchived)
	if template.ScopeType == "tenant" || template.ScopeType == "merchant" {
		q = q.Where("id = ?", template.TenantID)
	}
	if len(impactConfig.TenantIDs) > 0 {
		q = q.Where("id IN ?", impactConfig.TenantIDs)
	}
	rows := make([]bindingSubjectRow, 0)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return filterSubjectsByTemplateTarget(template, rows), nil
}

func queryMerchantBindingSubjects(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow, impactConfig templateImpactScopeConfig) ([]bindingSubjectRow, error) {
	if !tx.Migrator().HasTable(&merchantmodel.SysMerchant{}) {
		return []bindingSubjectRow{}, nil
	}
	q := tx.WithContext(ctx).Table("sys_merchant").Select("id AS subject_id, tenant_id, merchant_code AS subject_code, merchant_name AS subject_name, available_channels").Where("business_status <> ?", merchantmodel.MerchantBusinessArchived)
	if template.ScopeType == "tenant" || template.ScopeType == "merchant" {
		q = q.Where("tenant_id = ?", template.TenantID)
	}
	if template.ScopeType == "merchant" {
		q = q.Where("id = ?", template.MerchantID)
	}
	if len(impactConfig.TenantIDs) > 0 {
		q = q.Where("tenant_id IN ?", impactConfig.TenantIDs)
	}
	if len(impactConfig.BindingTenantIDs) > 0 {
		q = q.Where("tenant_id IN ?", impactConfig.BindingTenantIDs)
	}
	if len(impactConfig.MerchantIDs) > 0 {
		q = q.Where("id IN ?", impactConfig.MerchantIDs)
	}
	rows := make([]bindingSubjectRow, 0)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return filterSubjectsByTemplateTarget(template, rows), nil
}

func filterSubjectsByTemplateTarget(template channelIntegrationTemplateRow, rows []bindingSubjectRow) []bindingSubjectRow {
	if template.TemplateType != channeltemplate.TemplateTypeChannel {
		return rows
	}
	result := make([]bindingSubjectRow, 0, len(rows))
	for _, row := range rows {
		if subjectContainsChannel(row.AvailableChannels, template.TargetCode) {
			result = append(result, row)
		}
	}
	return result
}

func subjectContainsChannel(raw, targetCode string) bool {
	channels := decodeJSONStrings(raw)
	for _, channel := range channels {
		if operatefunnel.NormalizeChannel(channel) == targetCode {
			return true
		}
	}
	return false
}

func decodeJSONStrings(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}
	result := make([]string, 0)
	if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
		return []string{}
	}
	return result
}

func encodeEffectScopeSnapshot(template channelIntegrationTemplateRow, subjectType string, subject bindingSubjectRow) (string, error) {
	subjectLabel := subject.SubjectName
	if strings.TrimSpace(subject.SubjectCode) != "" {
		subjectLabel = fmt.Sprintf("%s(%s)", subject.SubjectName, subject.SubjectCode)
	}
	snapshot := effectScopeSnapshot{
		SubjectType:  subjectType,
		SubjectID:    subject.SubjectID,
		SubjectCode:  subject.SubjectCode,
		SubjectName:  subject.SubjectName,
		SubjectLabel: subjectLabel,
		TemplateCode: template.TemplateCode,
		TargetCode:   template.TargetCode,
		ScopeType:    template.ScopeType,
		PlatformID:   template.PlatformID,
		TenantID:     subject.TenantID,
		MerchantID:   0,
	}
	if subjectType == "merchant" {
		snapshot.MerchantID = subject.SubjectID
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func bindingIdentity(subjectType string, platformID, tenantID, merchantID int64, targetCode string) string {
	return fmt.Sprintf("%s:%d:%d:%d:%s", subjectType, platformID, tenantID, merchantID, targetCode)
}

func summarizeTemplateBindings(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow) (templateImpactSummary, error) {
	rows := make([]channelIntegrationBindingRow, 0)
	if err := tx.WithContext(ctx).Where("template_id = ? AND binding_status <> ?", template.ID, channeltemplate.StatusArchived).Order("id ASC").Find(&rows).Error; err != nil {
		return templateImpactSummary{}, err
	}
	if len(rows) == 0 {
		if err := syncTemplateBindings(ctx, tx, template, fallbackActor(template.UpdateBy, template.CreateBy)); err != nil {
			return templateImpactSummary{}, err
		}
		if err := tx.WithContext(ctx).Where("template_id = ? AND binding_status <> ?", template.ID, channeltemplate.StatusArchived).Order("id ASC").Find(&rows).Error; err != nil {
			return templateImpactSummary{}, err
		}
	}
	summary := templateImpactSummary{
		StatusReason: templateStatusReason(template.Status),
	}
	latest := template.CreateTime
	if template.UpdateTime != nil && template.UpdateTime.After(latest) {
		latest = *template.UpdateTime
	}
	for _, row := range rows {
		switch row.SubjectType {
		case "tenant":
			summary.BindingTenantCount++
		case "merchant":
			summary.BindingMerchantCount++
		}
		if row.PublishedAt != nil && row.PublishedAt.After(latest) {
			latest = *row.PublishedAt
		}
		if row.UpdateTime != nil && row.UpdateTime.After(latest) {
			latest = *row.UpdateTime
		}
		if len(summary.BindingSampleLabels) < 3 {
			label := effectScopeSnapshotLabel(row.EffectScopeSnapshot)
			if strings.TrimSpace(label) != "" {
				summary.BindingSampleLabels = append(summary.BindingSampleLabels, label)
			}
		}
	}
	summary.LastStatusChangeTime = timeToString(latest)
	return summary, nil
}

func effectScopeSnapshotLabel(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	var snapshot effectScopeSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return ""
	}
	return strings.TrimSpace(snapshot.SubjectLabel)
}

func mergeImpactScopeConfigWithSummary(raw string, summary templateImpactSummary) string {
	payload := make(map[string]interface{})
	trimmed := strings.TrimSpace(raw)
	if trimmed != "" && trimmed != "{}" {
		_ = json.Unmarshal([]byte(trimmed), &payload)
	}
	payload["impactSummary"] = summary
	encoded, err := json.Marshal(payload)
	if err != nil {
		return trimmed
	}
	return string(encoded)
}

func templateStatusReason(status string) string {
	switch status {
	case channeltemplate.StatusDraft:
		return "模板处于草稿态，仅用于治理预览，尚未正式生效"
	case channeltemplate.StatusEnabled:
		return "模板已启用，符合条件的主体会按绑定真相源参与后续治理"
	case channeltemplate.StatusDisabled:
		return "模板已停用，既有绑定保留但不再作为新的生效配置来源"
	case channeltemplate.StatusArchived:
		return "模板已归档，仅保留历史治理轨迹与影响范围快照"
	default:
		return ""
	}
}

func timeToString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.DateTime)
}

func fallbackActor(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return "system"
}

func recordTemplateOperateLog(ctx context.Context, tx *gorm.DB, template channelIntegrationTemplateRow, action, actor, requestSummary string) error {
	operator := fallbackActor(actor)
	extra, err := audit.EncodeGovernancePayload(audit.GovernancePayload{
		Action:         action,
		ResourceType:   "channel_integration_template",
		ResourceID:     template.ID,
		ResourceName:   template.TemplateName,
		ScopeType:      template.ScopeType,
		PlatformID:     template.PlatformID,
		TenantID:       template.TenantID,
		MerchantID:     template.MerchantID,
		OperatorName:   operator,
		Result:         "success",
		RequestSummary: requestSummary,
	})
	if err != nil {
		return err
	}
	logRow := &model.SysOperateLog{
		Title:           "渠道与集成模板",
		BusinessType:    templateOperateBusinessType(action),
		Method:          action,
		RequestMethod:   "RPC",
		OperatorType:    1,
		OperateName:     operator,
		DeptName:        "",
		OperateURL:      "/rpc/sys/channelintegrationtemplate",
		OperateIP:       "",
		OperateLocation: "",
		OperateParam:    truncateForOperateLog(requestSummary, 1800),
		JSONResult:      truncateForOperateLog(fmt.Sprintf("template=%d,status=%s", template.ID, template.Status), 1800),
		Platform:        "go-zero",
		Browser:         "",
		Version:         "",
		Os:              "",
		Arch:            "",
		Engine:          "",
		EngineDetails:   "",
		Extra:           truncateForOperateLog(extra, 900),
		Status:          1,
		ErrorMsg:        "",
		OperateTime:     time.Now(),
		CostTime:        0,
	}
	return tx.WithContext(ctx).Create(logRow).Error
}

func templateOperateBusinessType(action string) int32 {
	switch action {
	case "create":
		return 1
	case "update", "enable", "disable", "archive", "status_change", "sync_binding":
		return 2
	default:
		return 0
	}
}

func truncateForOperateLog(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if limit <= 0 || len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit]
}
