package orderservicelogic

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logc"
)

// OrderEventPayload 统一订单域事件 payload 结构
// 遵循 Architecture 规范: {domain}.{entity}.{action}.v{n}
type OrderEventPayload struct {
	EventID    string                 `json:"eventId"`        // 事件唯一标识
	OccurredAt int64                  `json:"occurredAt"`     // 事件发生时间戳
	TraceID    string                 `json:"traceId"`        // 链路追踪 ID
	PlatformID int64                  `json:"platformId"`     // 平台 ID
	TenantID   int64                  `json:"tenantId"`       // 租户 ID
	MerchantID int64                  `json:"merchantId"`     // 商户 ID
	ActorID    int64                  `json:"actorId"`        // 操作者 ID
	EntityID   int64                  `json:"entityId"`       // 实体 ID (订单 ID)
	Action     string                 `json:"action"`         // 业务动作 (order.created/order.paid/order.cancelled)
	Version    string                 `json:"version"`        // 事件版本 (v1)
	ScopeType  string                 `json:"scopeType"`      // 作用域类型
	Data       map[string]interface{} `json:"data,omitempty"` // 业务数据上下文
}

// sendOrderEvent 统一订单域事件发送入口
// 替换原有的直接拼 JSON payload 模式
func sendOrderEvent(ctx context.Context, svcCtx *svc.ServiceContext, queue, routingKey, action string, orderID int64, current pkgscope.GovernanceScope, actorID int64, data map[string]interface{}) {
	eventID := uuid.New().String()
	traceID := audit.NewTraceID(action, orderID)
	occurredAt := time.Now().UnixMilli()

	payload := OrderEventPayload{
		EventID:    eventID,
		OccurredAt: occurredAt,
		TraceID:    traceID,
		PlatformID: current.PlatformID,
		TenantID:   current.TenantID,
		MerchantID: current.MerchantID,
		ActorID:    actorID,
		EntityID:   orderID,
		Action:     action,
		Version:    "v1",
		ScopeType:  current.ScopeType,
		Data:       data,
	}

	body, err := sonic.Marshal(payload)
	if err != nil {
		logc.Errorf(ctx, "序列化订单事件失败,action:%s,orderId:%d,异常:%s", action, orderID, err.Error())
		return
	}

	if err := svcCtx.RabbitMQ.SendMessage("order.event.exchange", "direct", queue, routingKey, body); err != nil {
		logc.Errorf(ctx, "发送订单异步消息失败,action:%s,orderId:%d,scope:%+v,异常:%s", action, orderID, current, err.Error())
	}
	logc.Infof(ctx, "发送订单事件成功,eventId:%s,action:%s,orderId:%d,traceId:%s", eventID, action, orderID, traceID)
}
