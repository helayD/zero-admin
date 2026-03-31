package order

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/zeromicro/go-zero/core/logc"
)

func OrderConfirm(ctx context.Context, body []byte) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderConfirm 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderConfirm 收到订单确认事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	return nil
}
