package productspuservicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
)

func buildProductEventMeta(action string, actorID int64, actorName string) pkgscope.ProductEventMeta {
	now := time.Now()
	if actorName == "" {
		actorName = fmt.Sprintf("user-%d", actorID) // 无姓名时用 ID 占位，保持日志可读
	}
	return pkgscope.ProductEventMeta{
		Action:     action,
		ActorID:    actorID,
		ActorName:  actorName,
		OccurredAt: now.Format(time.RFC3339Nano),
		Version:    now.UnixMilli(),
	}
}

func sendProductESSync(ctx context.Context, svcCtx *svc.ServiceContext, productID int64, current pkgscope.GovernanceScope, meta pkgscope.ProductEventMeta) {
	if svcCtx == nil || svcCtx.RabbitMQ == nil {
		return
	}
	if productID <= 0 {
		logc.Errorf(ctx, "[PMS→MQ] 商品ID无效，跳过ES同步消息, action=%s, actorId=%d, scope=%+v",
			meta.Action, meta.ActorID, current)
		return
	}
	traceID := audit.NewTraceID(meta.Action, productID)

	// 根据 action 判断路由键：新增 → published，更新 → updated
	// ⚠️ 防御性检查：delete action 不应进入此函数，应调用 sendProductESDelete
	var routingKey string
	switch meta.Action {
	case "pms.product_spu.create":
		routingKey = "pms.product.published.key"
	case "pms.product_spu.update":
		routingKey = "pms.product.updated.key"
	default:
		// 无法识别的 action，记录告警但不阻断主业务（符合异步解耦原则）
		logc.Errorf(ctx, "[PMS→MQ] 无法识别的商品事件 action，无法发送 ES 同步消息, spuId=%d, traceId=%s, action=%s, scope=%+v",
			productID, traceID, meta.Action, current)
		return
	}

	body, _ := sonic.Marshal(pkgscope.NewProductESSyncPayload(productID, current, traceID, meta))
	if err := svcCtx.RabbitMQ.SendMessage("product.event.exchange", "topic", "pms.product.sync.queue", routingKey, body); err != nil {
		logc.Errorf(ctx, "[PMS→MQ] 发送商品ES同步消息失败, spuId=%d, traceId=%s, routingKey=%s, scope=%+v, err=%s",
			productID, traceID, routingKey, current, err.Error())
	} else {
		logc.Infof(ctx, "[PMS→MQ] 发送商品ES同步消息成功, spuId=%d, traceId=%s, routingKey=%s, scope=%+v",
			productID, traceID, routingKey, current)
	}
}

func sendProductESSyncBatch(ctx context.Context, svcCtx *svc.ServiceContext, productIDs []int64, current pkgscope.GovernanceScope, meta pkgscope.ProductEventMeta) {
	for _, productID := range pkgscope.UniquePositiveIDs(productIDs) {
		sendProductESSync(ctx, svcCtx, productID, current, meta)
	}
}

func sendProductESDelete(ctx context.Context, svcCtx *svc.ServiceContext, productIDs []int64, current pkgscope.GovernanceScope, meta pkgscope.ProductEventMeta) {
	if svcCtx == nil || svcCtx.RabbitMQ == nil {
		return
	}
	uniqueIDs := pkgscope.UniquePositiveIDs(productIDs)
	if len(uniqueIDs) == 0 {
		return
	}

	traceID := audit.NewTraceID(meta.Action, uniqueIDs[0])
	body, _ := sonic.Marshal(pkgscope.NewProductESDeletePayload(uniqueIDs, current, traceID, meta))
	if err := svcCtx.RabbitMQ.SendMessage("product.event.exchange", "topic", "pms.product.delete.queue", "pms.product.deleted.key", body); err != nil {
		logc.Errorf(ctx, "[PMS→MQ] 发送商品ES删除消息失败, ids=%+v, traceId=%s, scope=%+v, err=%s",
			uniqueIDs, traceID, current, err.Error())
	} else {
		logc.Infof(ctx, "[PMS→MQ] 发送商品ES删除消息成功, ids=%+v, traceId=%s, scope=%+v",
			uniqueIDs, traceID, current)
	}
}

func syncProductIndexVisibility(ctx context.Context, svcCtx *svc.ServiceContext, current pkgscope.GovernanceScope, productIDs []int64, meta pkgscope.ProductEventMeta) {
	if svcCtx == nil || svcCtx.DB == nil {
		return
	}
	syncIDs, deleteIDs, err := logiccommon.PartitionProductIndexIDs(ctx, svcCtx.DB, current, productIDs)
	if err != nil {
		logc.Errorf(ctx, "划分商品索引同步动作失败,ids:%+v,action:%s,scope:%+v,异常:%s", productIDs, meta.Action, current, err.Error())
		return
	}
	if len(syncIDs) > 0 {
		sendProductESSyncBatch(ctx, svcCtx, syncIDs, current, meta)
	}
	if len(deleteIDs) > 0 {
		sendProductESDelete(ctx, svcCtx, deleteIDs, current, meta)
	}
}
