package productspuservicelogic

import (
	"context"
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
	traceID := audit.NewTraceID(meta.Action, productID)
	body, _ := sonic.Marshal(pkgscope.NewProductESSyncPayload(productID, current, traceID, meta))
	if err := svcCtx.RabbitMQ.SendMessage("product.event.exchange", "syn.product.to.es.queue", "syn.product.key", body); err != nil {
		logc.Errorf(ctx, "发送商品ES同步消息失败,spuId:%d,traceId:%s,scope:%+v,异常:%s", productID, traceID, current, err.Error())
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
	if err := svcCtx.RabbitMQ.SendMessage("product.event.exchange", "delete.product.from.es.queue", "delete.product.key", body); err != nil {
		logc.Errorf(ctx, "发送商品ES删除消息失败,ids:%+v,traceId:%s,scope:%+v,异常:%s", uniqueIDs, traceID, current, err.Error())
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
