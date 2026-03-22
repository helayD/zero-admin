package scope

import (
	"encoding/json"
	"errors"
	"strings"
)

type ProductESSyncPayload struct {
	ID         int64  `json:"id"`
	ScopeType  string `json:"scopeType"`
	PlatformID int64  `json:"platformId"`
	TenantID   int64  `json:"tenantId,omitempty"`
	MerchantID int64  `json:"merchantId,omitempty"`
}

func NewProductESSyncPayload(id int64, current GovernanceScope) ProductESSyncPayload {
	return ProductESSyncPayload{
		ID:         id,
		ScopeType:  current.ScopeType,
		PlatformID: current.PlatformID,
		TenantID:   current.TenantID,
		MerchantID: current.MerchantID,
	}
}

func DecodeProductESSyncPayload(body []byte) (ProductESSyncPayload, GovernanceScope, error) {
	var payload ProductESSyncPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProductESSyncPayload{}, GovernanceScope{}, err
	}
	if strings.TrimSpace(payload.ScopeType) == "" {
		return ProductESSyncPayload{}, GovernanceScope{}, errors.New("商品 ES 同步消息缺少治理范围")
	}

	current, err := NormalizeGovernanceScope(payload.ScopeType, payload.PlatformID, payload.TenantID, payload.MerchantID)
	if err != nil {
		return ProductESSyncPayload{}, GovernanceScope{}, err
	}

	return payload, current, nil
}
