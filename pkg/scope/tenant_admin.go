package scope

import "encoding/json"

const (
	SubjectTypeTenant = "tenant"

	RoleModeBootstrap = "bootstrap"

	ActivationStatusPending  = "pending-activation"
	ActivationStatusActive   = "active"
	ActivationStatusDisabled = "disabled"
	ActivationStatusArchived = "archived"
)

type TenantAdminMetadata struct {
	SubjectType      string `json:"subjectType"`
	PlatformID       int64  `json:"platformId"`
	TenantID         int64  `json:"tenantId"`
	TenantCode       string `json:"tenantCode"`
	ActivationStatus string `json:"activationStatus"`
	RoleMode         string `json:"roleMode"`
}

func BuildTenantAdminMetadata(platformID, tenantID int64, tenantCode, activationStatus string) (string, error) {
	payload, err := json.Marshal(TenantAdminMetadata{
		SubjectType:      SubjectTypeTenant,
		PlatformID:       platformID,
		TenantID:         tenantID,
		TenantCode:       tenantCode,
		ActivationStatus: activationStatus,
		RoleMode:         RoleModeBootstrap,
	})
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
