package scope

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	SubjectTypePlatform = "platform"
	SubjectTypeMerchant = "merchant"

	DefaultPlatformID int64 = 1
)

type GovernanceScope struct {
	ScopeType  string `json:"scopeType"`
	PlatformID int64  `json:"platformId"`
	TenantID   int64  `json:"tenantId,omitempty"`
	MerchantID int64  `json:"merchantId,omitempty"`
}

type GovernanceScopeMetadata struct {
	GovernanceScope
	ActivationStatus string `json:"activationStatus,omitempty"`
	RoleMode         string `json:"roleMode,omitempty"`
	ScopeLabel       string `json:"scopeLabel,omitempty"`
}

func NormalizeGovernanceScope(scopeType string, platformID, tenantID, merchantID int64) (GovernanceScope, error) {
	normalized := GovernanceScope{
		ScopeType:  strings.TrimSpace(scopeType),
		PlatformID: platformID,
		TenantID:   tenantID,
		MerchantID: merchantID,
	}
	if normalized.ScopeType == "" {
		normalized.ScopeType = inferScopeType(tenantID, merchantID)
	}

	if normalized.PlatformID == 0 {
		normalized.PlatformID = DefaultPlatformID
	}

	switch normalized.ScopeType {
	case SubjectTypePlatform:
		normalized.TenantID = 0
		normalized.MerchantID = 0
	case SubjectTypeTenant:
		if normalized.TenantID <= 0 {
			return GovernanceScope{}, errors.New("租户级治理必须提供 tenantId")
		}
		normalized.MerchantID = 0
	case SubjectTypeMerchant:
		if normalized.MerchantID <= 0 {
			return GovernanceScope{}, errors.New("商户级治理必须提供 merchantId")
		}
	default:
		return GovernanceScope{}, fmt.Errorf("不支持的主体范围：%s", normalized.ScopeType)
	}

	return normalized, nil
}

func inferScopeType(tenantID, merchantID int64) string {
	if merchantID > 0 {
		return SubjectTypeMerchant
	}
	if tenantID > 0 {
		return SubjectTypeTenant
	}
	return SubjectTypePlatform
}

func (s GovernanceScope) Label() string {
	switch s.ScopeType {
	case SubjectTypeTenant:
		return fmt.Sprintf("租户 #%d", s.TenantID)
	case SubjectTypeMerchant:
		if s.TenantID > 0 {
			return fmt.Sprintf("租户 #%d / 商户 #%d", s.TenantID, s.MerchantID)
		}
		return fmt.Sprintf("商户 #%d", s.MerchantID)
	default:
		return "平台级"
	}
}

func (s GovernanceScope) Key() string {
	return fmt.Sprintf("%s:%d:%d:%d", s.ScopeType, s.PlatformID, s.TenantID, s.MerchantID)
}

func (s GovernanceScope) SameScope(other GovernanceScope) bool {
	return s.ScopeType == other.ScopeType &&
		s.PlatformID == other.PlatformID &&
		s.TenantID == other.TenantID &&
		s.MerchantID == other.MerchantID
}

func EncodeGovernanceScopeMetadata(scope GovernanceScope, activationStatus, roleMode string) (string, error) {
	payload, err := json.Marshal(GovernanceScopeMetadata{
		GovernanceScope:  scope,
		ActivationStatus: strings.TrimSpace(activationStatus),
		RoleMode:         strings.TrimSpace(roleMode),
		ScopeLabel:       scope.Label(),
	})
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
