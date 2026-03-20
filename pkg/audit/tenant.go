package audit

import "encoding/json"

const (
	ActionTenantCreated          = "tenant.created"
	ActionTenantAdminInitialized = "tenant.admin_initialized"
	ActionTenantEnabled          = "tenant.enabled"
	ActionTenantDisabled         = "tenant.disabled"
	ActionTenantArchived         = "tenant.archived"
)

type TenantPayload struct {
	TenantID              int64    `json:"tenantId"`
	TenantCode            string   `json:"tenantCode"`
	TenantName            string   `json:"tenantName,omitempty"`
	Action                string   `json:"action"`
	PreviousStatus        int32    `json:"previousStatus,omitempty"`
	CurrentStatus         int32    `json:"currentStatus"`
	AvailableChannels     []string `json:"availableChannels,omitempty"`
	FeatureFlags          []string `json:"featureFlags,omitempty"`
	PrimaryAdminUserID    int64    `json:"primaryAdminUserId,omitempty"`
	PrimaryAdminUserName  string   `json:"primaryAdminUserName,omitempty"`
	PrimaryAdminMobile    string   `json:"primaryAdminMobile,omitempty"`
	AdminActivationStatus string   `json:"adminActivationStatus,omitempty"`
}

func EncodeTenantPayload(payload TenantPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
