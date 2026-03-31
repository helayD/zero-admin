package order

import (
	"context"
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

func LogWithEventPayload(ctx context.Context, msg string, args ...interface{}) {
	if traceId, ok := ctx.Value("traceId").(string); ok && traceId != "" {
		logc.Infof(ctx, "%s | traceId=%s", msg, traceId, args)
	} else {
		logc.Infof(ctx, msg, args)
	}
}
