package order

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/zeromicro/go-zero/core/logc"
)

func OrderCreate(ctx context.Context, body []byte) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderCreate 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderCreate 收到订单创建事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	return nil
}
