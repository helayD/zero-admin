package order

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/zeromicro/go-zero/core/logc"
)

// OrderDelivery 订单发货通知
func OrderDelivery(ctx context.Context, body []byte) {
	logc.Infof(ctx, "订单发货通知mq消息: %s", body)
	var orderInfo eventPayload
	err := sonic.Unmarshal(body, &orderInfo)
	if err != nil {
		logc.Errorf(ctx, "序列化 JSON 失败: %v", err)
		return
	}
	logc.Infof(ctx, "订单发货事件,orderId:%d,traceId:%s,scope:%s/%d/%d/%d", orderInfo.ID, orderInfo.TraceID, orderInfo.ScopeType, orderInfo.PlatformID, orderInfo.TenantID, orderInfo.MerchantID)
}
