package channelintegrationtemplateservicelogic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

type channelIntegrationTemplateRow struct {
	ID                   int64      `gorm:"column:id;primaryKey"`
	TemplateCode         string     `gorm:"column:template_code"`
	TemplateName         string     `gorm:"column:template_name"`
	TemplateType         string     `gorm:"column:template_type"`
	TargetCode           string     `gorm:"column:target_code"`
	ScopeType            string     `gorm:"column:scope_type"`
	PlatformID           int64      `gorm:"column:platform_id"`
	TenantID             int64      `gorm:"column:tenant_id"`
	MerchantID           int64      `gorm:"column:merchant_id"`
	Status               string     `gorm:"column:status"`
	MetadataConfig       string     `gorm:"column:metadata_config"`
	SecretRefConfig      string     `gorm:"column:secret_ref_config"`
	IntentContractConfig string     `gorm:"column:intent_contract_config"`
	ImpactScopeConfig    string     `gorm:"column:impact_scope_config"`
	Remark               string     `gorm:"column:remark"`
	CreateBy             string     `gorm:"column:create_by"`
	CreateTime           time.Time  `gorm:"column:create_time"`
	UpdateBy             string     `gorm:"column:update_by"`
	UpdateTime           *time.Time `gorm:"column:update_time"`
}

func (channelIntegrationTemplateRow) TableName() string {
	return "sys_channel_integration_template"
}

type normalizedTemplatePayload struct {
	TemplateCode         string
	TemplateName         string
	TemplateType         string
	TargetCode           string
	ScopeType            string
	PlatformID           int64
	TenantID             int64
	MerchantID           int64
	Status               string
	MetadataConfig       string
	SecretRefConfig      string
	IntentContractConfig string
	ImpactScopeConfig    string
	Remark               string
}

func normalizeTemplatePayload(templateCode, templateName, templateType, targetCode, scopeType string, platformID, tenantID, merchantID int64, status, metadataConfig, secretRefConfig, intentContractConfig, impactScopeConfig, remark string) (*normalizedTemplatePayload, error) {
	normalizedTemplateCode := strings.TrimSpace(templateCode)
	if normalizedTemplateCode == "" {
		return nil, errors.New("模板编码不能为空")
	}
	normalizedTemplateName := strings.TrimSpace(templateName)
	if normalizedTemplateName == "" {
		return nil, errors.New("模板名称不能为空")
	}
	normalizedTemplateType, err := channeltemplate.NormalizeTemplateType(templateType)
	if err != nil {
		return nil, err
	}
	normalizedTargetCode, err := channeltemplate.NormalizeTargetCode(normalizedTemplateType, targetCode)
	if err != nil {
		return nil, err
	}
	normalizedScopeType, normalizedPlatformID, normalizedTenantID, normalizedMerchantID, err := normalizeTemplateScope(scopeType, platformID, tenantID, merchantID)
	if err != nil {
		return nil, err
	}
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus == "" {
		normalizedStatus = channeltemplate.StatusDraft
	}
	normalizedStatus, err = channeltemplate.NormalizeStatus(normalizedStatus)
	if err != nil {
		return nil, err
	}
	normalizedMetadataConfig, err := normalizeJSONObjectString(metadataConfig)
	if err != nil {
		return nil, fmt.Errorf("元数据配置非法: %w", err)
	}
	normalizedSecretRefConfig, err := normalizeSecretRefConfig(secretRefConfig)
	if err != nil {
		return nil, err
	}
	if err := validateCredentialRefsRequired(normalizedTemplateType, normalizedTargetCode, normalizedSecretRefConfig); err != nil {
		return nil, err
	}
	normalizedIntentContractConfig, err := normalizeIntentContractConfig(normalizedTargetCode, intentContractConfig)
	if err != nil {
		return nil, err
	}
	normalizedImpactScopeConfig, err := normalizeJSONObjectString(impactScopeConfig)
	if err != nil {
		return nil, fmt.Errorf("影响范围配置非法: %w", err)
	}
	if err := rejectSensitivePlaintext(normalizedMetadataConfig, normalizedImpactScopeConfig); err != nil {
		return nil, err
	}

	return &normalizedTemplatePayload{
		TemplateCode:         normalizedTemplateCode,
		TemplateName:         normalizedTemplateName,
		TemplateType:         normalizedTemplateType,
		TargetCode:           normalizedTargetCode,
		ScopeType:            normalizedScopeType,
		PlatformID:           normalizedPlatformID,
		TenantID:             normalizedTenantID,
		MerchantID:           normalizedMerchantID,
		Status:               normalizedStatus,
		MetadataConfig:       normalizedMetadataConfig,
		SecretRefConfig:      normalizedSecretRefConfig,
		IntentContractConfig: normalizedIntentContractConfig,
		ImpactScopeConfig:    normalizedImpactScopeConfig,
		Remark:               strings.TrimSpace(remark),
	}, nil
}

func normalizeTemplateScope(scopeType string, platformID, tenantID, merchantID int64) (string, int64, int64, int64, error) {
	normalizedScopeType := strings.TrimSpace(scopeType)
	if normalizedScopeType == "" {
		normalizedScopeType = "platform"
	}
	normalizedPlatformID := platformID
	if normalizedPlatformID <= 0 {
		normalizedPlatformID = 1
	}

	switch normalizedScopeType {
	case "platform":
		return normalizedScopeType, normalizedPlatformID, 0, 0, nil
	case "tenant":
		if tenantID <= 0 {
			return "", 0, 0, 0, errors.New("租户级模板必须指定 tenant_id")
		}
		return normalizedScopeType, normalizedPlatformID, tenantID, 0, nil
	case "merchant":
		if tenantID <= 0 {
			return "", 0, 0, 0, errors.New("商户级模板必须指定 tenant_id")
		}
		if merchantID <= 0 {
			return "", 0, 0, 0, errors.New("商户级模板必须指定 merchant_id")
		}
		return normalizedScopeType, normalizedPlatformID, tenantID, merchantID, nil
	default:
		return "", 0, 0, 0, errors.New("模板作用域仅支持 platform、tenant、merchant")
	}
}

func normalizeJSONObjectString(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "{}", nil
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return "", errors.New("必须是合法 JSON 对象")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func normalizeSecretRefConfig(raw string) (string, error) {
	normalized, err := normalizeJSONObjectString(raw)
	if err != nil {
		return "", fmt.Errorf("敏感配置引用非法: %w", err)
	}
	secretRefs := make(map[string]string)
	if err := json.Unmarshal([]byte(normalized), &secretRefs); err != nil {
		return "", errors.New("敏感配置引用必须是 string map")
	}
	if err := channeltemplate.ValidateSecretRefMap(secretRefs); err != nil {
		return "", err
	}
	return normalized, nil
}

func normalizeIntentContractConfig(targetCode, raw string) (string, error) {
	normalized, err := normalizeJSONObjectString(raw)
	if err != nil {
		return "", fmt.Errorf("意图契约配置非法: %w", err)
	}
	contracts := make(map[string]channeltemplate.IntentContract)
	if err := json.Unmarshal([]byte(normalized), &contracts); err != nil {
		return "", errors.New("意图契约配置必须是对象 map")
	}
	if err := channeltemplate.ValidateIntentContracts(targetCode, contracts); err != nil {
		return "", err
	}
	return normalized, nil
}

func normalizePage(pageNum, pageSize int64) (int64, int64) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageNum, pageSize
}

func formatNullableTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return time_util.TimeToStr(*value)
}

func normalizeStatusFilter(status string) (string, error) {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return "", nil
	}
	return channeltemplate.NormalizeStatus(trimmed)
}

func normalizeTargetCodeFilter(targetCode string) string {
	trimmed := strings.TrimSpace(targetCode)
	if trimmed == "" {
		return ""
	}
	if strings.EqualFold(trimmed, "mini-program") {
		return channeltemplate.TargetMiniProgram
	}
	return trimmed
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

func mapTemplateRowToProto(row channelIntegrationTemplateRow) *sysclient.ChannelIntegrationTemplateData {
	return &sysclient.ChannelIntegrationTemplateData{
		Id:                   row.ID,
		TemplateCode:         row.TemplateCode,
		TemplateName:         row.TemplateName,
		TemplateType:         row.TemplateType,
		TargetCode:           row.TargetCode,
		ScopeType:            row.ScopeType,
		PlatformId:           row.PlatformID,
		TenantId:             row.TenantID,
		MerchantId:           row.MerchantID,
		Status:               row.Status,
		MetadataConfig:       compactJSONOrDefault(row.MetadataConfig),
		SecretRefConfig:      compactJSONOrDefault(row.SecretRefConfig),
		IntentContractConfig: compactJSONOrDefault(row.IntentContractConfig),
		ImpactScopeConfig:    compactJSONOrDefault(row.ImpactScopeConfig),
		Remark:               row.Remark,
		CreateBy:             row.CreateBy,
		CreateTime:           time_util.TimeToStr(row.CreateTime),
		UpdateBy:             row.UpdateBy,
		UpdateTime:           formatNullableTime(row.UpdateTime),
	}
}

func compactJSONOrDefault(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "{}"
	}
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, []byte(trimmed)); err != nil {
		return trimmed
	}
	return buffer.String()
}
