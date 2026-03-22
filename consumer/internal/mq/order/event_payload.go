package order

type eventPayload struct {
	ID         int64  `json:"id"`
	TraceID    string `json:"traceId,omitempty"`
	ScopeType  string `json:"scopeType,omitempty"`
	PlatformID int64  `json:"platformId,omitempty"`
	TenantID   int64  `json:"tenantId,omitempty"`
	MerchantID int64  `json:"merchantId,omitempty"`
}
