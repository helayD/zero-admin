package scope

import (
	"encoding/json"
	"errors"
	"strings"
)

type ProductESSyncPayload struct {
	ID         int64  `json:"id"`
	TraceID    string `json:"traceId,omitempty"`
	Action     string `json:"action,omitempty"`
	ActorID    int64  `json:"actorId,omitempty"`
	ActorName  string `json:"actorName,omitempty"`
	OccurredAt string `json:"occurredAt,omitempty"`
	Version    int64  `json:"version,omitempty"`
	ScopeType  string `json:"scopeType"`
	PlatformID int64  `json:"platformId"`
	TenantID   int64  `json:"tenantId,omitempty"`
	MerchantID int64  `json:"merchantId,omitempty"`
}

type ProductESDeletePayload struct {
	IDs        []int64 `json:"ids"`
	TraceID    string  `json:"traceId,omitempty"`
	Action     string  `json:"action,omitempty"`
	ActorID    int64   `json:"actorId,omitempty"`
	ActorName  string  `json:"actorName,omitempty"`
	OccurredAt string  `json:"occurredAt,omitempty"`
	Version    int64   `json:"version,omitempty"`
	ScopeType  string  `json:"scopeType"`
	PlatformID int64   `json:"platformId"`
	TenantID   int64   `json:"tenantId,omitempty"`
	MerchantID int64   `json:"merchantId,omitempty"`
}

type ProductEventMeta struct {
	Action     string
	ActorID    int64
	ActorName  string
	OccurredAt string
	Version    int64
}

func NewProductESSyncPayload(id int64, current GovernanceScope, traceID string, meta ProductEventMeta) ProductESSyncPayload {
	return ProductESSyncPayload{
		ID:         id,
		TraceID:    strings.TrimSpace(traceID),
		Action:     strings.TrimSpace(meta.Action),
		ActorID:    meta.ActorID,
		ActorName:  strings.TrimSpace(meta.ActorName),
		OccurredAt: strings.TrimSpace(meta.OccurredAt),
		Version:    meta.Version,
		ScopeType:  current.ScopeType,
		PlatformID: current.PlatformID,
		TenantID:   current.TenantID,
		MerchantID: current.MerchantID,
	}
}

func NewProductESDeletePayload(ids []int64, current GovernanceScope, traceID string, meta ProductEventMeta) ProductESDeletePayload {
	return ProductESDeletePayload{
		IDs:        UniquePositiveIDs(ids),
		TraceID:    strings.TrimSpace(traceID),
		Action:     strings.TrimSpace(meta.Action),
		ActorID:    meta.ActorID,
		ActorName:  strings.TrimSpace(meta.ActorName),
		OccurredAt: strings.TrimSpace(meta.OccurredAt),
		Version:    meta.Version,
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

func DecodeProductESDeletePayload(body []byte) (ProductESDeletePayload, GovernanceScope, error) {
	var payload ProductESDeletePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProductESDeletePayload{}, GovernanceScope{}, err
	}
	if strings.TrimSpace(payload.ScopeType) == "" {
		return ProductESDeletePayload{}, GovernanceScope{}, errors.New("商品 ES 删除消息缺少治理范围")
	}

	current, err := NormalizeGovernanceScope(payload.ScopeType, payload.PlatformID, payload.TenantID, payload.MerchantID)
	if err != nil {
		return ProductESDeletePayload{}, GovernanceScope{}, err
	}

	payload.IDs = UniquePositiveIDs(payload.IDs)
	return payload, current, nil
}
