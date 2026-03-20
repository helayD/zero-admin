package audit

import "encoding/json"

type GovernancePayload struct {
	Action       string `json:"action"`
	ResourceType string `json:"resourceType"`
	ResourceID   int64  `json:"resourceId,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
	ScopeType    string `json:"scopeType"`
	PlatformID   int64  `json:"platformId"`
	TenantID     int64  `json:"tenantId,omitempty"`
	MerchantID   int64  `json:"merchantId,omitempty"`
	ScopeLabel   string `json:"scopeLabel,omitempty"`
}

func EncodeGovernancePayload(payload GovernancePayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
