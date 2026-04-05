package channeltemplate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
)

const (
	TemplateTypeChannel     = "channel"
	TemplateTypeIntegration = "integration"
)

const (
	TargetH5                = operatefunnel.ChannelH5
	TargetMiniProgram       = operatefunnel.ChannelMiniProgram
	TargetLogisticsTracking = "logistics_tracking"
	TargetSMSProvider       = "sms_provider"
	TargetMemberMessage     = "member_message"
)

const (
	StatusDraft    = "draft"
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
	StatusArchived = "archived"
)

var allowedSecretRefPrefixes = []string{
	"credential://",
	"secret://",
	"vault://",
	"kms://",
	"ref:",
}

type IntentContract struct {
	Intent   string            `json:"intent"`
	RouteKey string            `json:"routeKey"`
	Scene    string            `json:"scene,optional"`
	Params   map[string]string `json:"params,optional"`
}

func NormalizeTemplateType(templateType string) (string, error) {
	switch strings.TrimSpace(templateType) {
	case TemplateTypeChannel, TemplateTypeIntegration:
		return strings.TrimSpace(templateType), nil
	default:
		return "", errors.New("模板类型仅支持 channel 或 integration")
	}
}

func NormalizeTargetCode(templateType, targetCode string) (string, error) {
	normalizedTemplateType, err := NormalizeTemplateType(templateType)
	if err != nil {
		return "", err
	}

	trimmedTargetCode := strings.TrimSpace(targetCode)
	switch normalizedTemplateType {
	case TemplateTypeChannel:
		normalizedChannel := operatefunnel.NormalizeChannel(trimmedTargetCode)
		switch normalizedChannel {
		case TargetH5, TargetMiniProgram:
			return normalizedChannel, nil
		default:
			return "", fmt.Errorf("渠道模板目标仅支持 %s 或 %s", TargetH5, TargetMiniProgram)
		}
	case TemplateTypeIntegration:
		switch trimmedTargetCode {
		case TargetLogisticsTracking, TargetSMSProvider, TargetMemberMessage:
			return trimmedTargetCode, nil
		default:
			return "", fmt.Errorf("集成模板目标仅支持 %s、%s 或 %s", TargetLogisticsTracking, TargetSMSProvider, TargetMemberMessage)
		}
	default:
		return "", errors.New("不支持的模板类型")
	}
}

func NormalizeStatus(status string) (string, error) {
	switch strings.TrimSpace(status) {
	case StatusDraft, StatusEnabled, StatusDisabled, StatusArchived:
		return strings.TrimSpace(status), nil
	default:
		return "", errors.New("模板状态仅支持 draft、enabled、disabled、archived")
	}
}

func ValidateStatusTransition(currentStatus, nextStatus string) error {
	normalizedCurrentStatus, err := NormalizeStatus(currentStatus)
	if err != nil {
		return err
	}
	normalizedNextStatus, err := NormalizeStatus(nextStatus)
	if err != nil {
		return err
	}
	if normalizedCurrentStatus == normalizedNextStatus {
		return nil
	}

	switch normalizedCurrentStatus {
	case StatusDraft:
		if normalizedNextStatus == StatusEnabled || normalizedNextStatus == StatusDisabled || normalizedNextStatus == StatusArchived {
			return nil
		}
	case StatusEnabled:
		if normalizedNextStatus == StatusDisabled || normalizedNextStatus == StatusArchived {
			return nil
		}
	case StatusDisabled:
		if normalizedNextStatus == StatusEnabled || normalizedNextStatus == StatusArchived {
			return nil
		}
	case StatusArchived:
		return errors.New("已归档模板不允许重新进入可编辑状态")
	}

	return fmt.Errorf("非法的模板状态流转: %s -> %s", normalizedCurrentStatus, normalizedNextStatus)
}

func RequiresIntentContracts(targetCode string) bool {
	switch strings.TrimSpace(targetCode) {
	case TargetH5, TargetMiniProgram, TargetMemberMessage:
		return true
	default:
		return false
	}
}

func ValidateSecretRefMap(secretRefs map[string]string) error {
	for key, value := range secretRefs {
		normalizedKey := strings.TrimSpace(key)
		normalizedValue := strings.TrimSpace(value)
		if normalizedKey == "" {
			return errors.New("敏感配置引用键不能为空")
		}
		if normalizedValue == "" {
			return fmt.Errorf("敏感配置引用[%s]不能为空", normalizedKey)
		}
		if !hasAllowedSecretRefPrefix(normalizedValue) {
			return fmt.Errorf("敏感配置引用[%s]必须使用 credential ref，禁止保存明文", normalizedKey)
		}
	}

	return nil
}

func ValidateIntentContracts(targetCode string, contracts map[string]IntentContract) error {
	if !RequiresIntentContracts(targetCode) {
		return nil
	}
	if len(contracts) == 0 {
		return fmt.Errorf("目标[%s]必须配置意图契约映射", strings.TrimSpace(targetCode))
	}

	for key, contract := range contracts {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			return errors.New("意图契约键不能为空")
		}
		if strings.TrimSpace(contract.Intent) == "" {
			return fmt.Errorf("意图契约[%s]的 intent 不能为空", normalizedKey)
		}
		routeKey := strings.TrimSpace(contract.RouteKey)
		if routeKey == "" {
			return fmt.Errorf("意图契约[%s]的 routeKey 不能为空", normalizedKey)
		}
		if looksLikeRawRoute(routeKey) {
			return fmt.Errorf("意图契约[%s]必须使用语义化 routeKey，禁止保存原始 URL/hash/query", normalizedKey)
		}
	}

	return nil
}

func hasAllowedSecretRefPrefix(value string) bool {
	for _, prefix := range allowedSecretRefPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}

func looksLikeRawRoute(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	if strings.Contains(trimmed, "http://") || strings.Contains(trimmed, "https://") {
		return true
	}
	if strings.ContainsAny(trimmed, "?#&=") {
		return true
	}
	if strings.Contains(trimmed, "/") {
		return true
	}

	return false
}
