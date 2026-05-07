package order

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logc"
)

type EventPayload struct {
	EventID    string                 `json:"eventId"`
	OccurredAt int64                  `json:"occurredAt"`
	TraceID    string                 `json:"traceId"`
	PlatformID int64                  `json:"platformId"`
	TenantID   int64                  `json:"tenantId"`
	MerchantID int64                  `json:"merchantId"`
	ActorID    int64                  `json:"actorId"`
	EntityID   int64                  `json:"entityId"`
	Action     string                 `json:"action"`
	Version    string                 `json:"version"`
	ScopeType  string                 `json:"scopeType"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func (p *EventPayload) ToContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "traceId", p.TraceID)
	ctx = context.WithValue(ctx, "platformId", p.PlatformID)
	ctx = context.WithValue(ctx, "tenantId", p.TenantID)
	ctx = context.WithValue(ctx, "merchantId", p.MerchantID)
	ctx = context.WithValue(ctx, "actorId", p.ActorID)
	ctx = context.WithValue(ctx, "entityId", p.EntityID)
	ctx = context.WithValue(ctx, "eventId", p.EventID)
	ctx = context.WithValue(ctx, "action", p.Action)
	return ctx
}

func (p *EventPayload) EntityIDToString() string {
	return fmt.Sprintf("%d", p.EntityID)
}

func (p *EventPayload) NormalizeLegacyPaymentPayload() {
	if p == nil || p.EntityID > 0 {
		return
	}
	p.EntityID = numberFromData(p.Data, "orderId")
	p.ActorID = numberFromData(p.Data, "memberId")
	p.PlatformID = numberFromData(p.Data, "platformId")
	p.TenantID = numberFromData(p.Data, "tenantId")
	p.MerchantID = numberFromData(p.Data, "merchantId")
	if p.Action == "" {
		p.Action = "paid"
	}
	if p.EntityID > 0 && p.Data == nil {
		p.Data = map[string]interface{}{}
	}
}

func numberFromData(data map[string]interface{}, key string) int64 {
	if data == nil {
		return 0
	}
	switch value := data[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case float64:
		return int64(value)
	case string:
		parsed, _ := strconv.ParseInt(value, 10, 64)
		return parsed
	default:
		return 0
	}
}

func LogWithEventPayload(ctx context.Context, msg string, args ...interface{}) {
	if traceId, ok := ctx.Value("traceId").(string); ok && traceId != "" {
		logc.Infof(ctx, "%s | traceId=%s", msg, traceId, args)
	} else {
		logc.Infof(ctx, msg, args)
	}
}
