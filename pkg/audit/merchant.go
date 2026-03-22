package audit

import "encoding/json"

const (
	ActionMerchantCreated            = "merchant.created"
	ActionMerchantApproved           = "merchant.approved"
	ActionMerchantRejected           = "merchant.rejected"
	ActionMerchantMaterialRequested  = "merchant.material_requested"
	ActionMerchantEnabled            = "merchant.enabled"
	ActionMerchantDisabled           = "merchant.disabled"
	ActionMerchantArchived           = "merchant.archived"
)

type MerchantPayload struct {
	TraceID               string   `json:"traceId,omitempty"`
	TenantID              int64    `json:"tenantId"`
	MerchantID            int64    `json:"merchantId"`
	MerchantCode          string   `json:"merchantCode"`
	MerchantName          string   `json:"merchantName,omitempty"`
	Action                string   `json:"action"`
	PreviousReviewStatus  int32    `json:"previousReviewStatus,omitempty"`
	CurrentReviewStatus   int32    `json:"currentReviewStatus,omitempty"`
	PreviousBusinessStatus int32   `json:"previousBusinessStatus,omitempty"`
	CurrentBusinessStatus int32    `json:"currentBusinessStatus,omitempty"`
	AvailableChannels     []string `json:"availableChannels,omitempty"`
	CapabilityFlags       []string `json:"capabilityFlags,omitempty"`
	PrimaryAdminUserID    int64    `json:"primaryAdminUserId,omitempty"`
	VisibleScopeHint      string   `json:"visibleScopeHint,omitempty"`
	Result                string   `json:"result,omitempty"`
}

func EncodeMerchantPayload(payload MerchantPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
