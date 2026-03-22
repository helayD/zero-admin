package orderservicelogic

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
)

func sendOrderEvent(ctx context.Context, svcCtx *svc.ServiceContext, queue, routingKey, action string, orderID int64, current pkgscope.GovernanceScope) {
	traceID := audit.NewTraceID(action, orderID)
	body, _ := sonic.Marshal(map[string]any{
		"id":         orderID,
		"traceId":    traceID,
		"scopeType":  current.ScopeType,
		"platformId": current.PlatformID,
		"tenantId":   current.TenantID,
		"merchantId": current.MerchantID,
	})
	if err := svcCtx.RabbitMQ.SendMessage("order.event.exchange", queue, routingKey, body); err != nil {
		logc.Errorf(ctx, "发送订单异步消息失败,action:%s,orderId:%d,scope:%+v,异常:%s", action, orderID, current, err.Error())
	}
}
