package productspuservicelogic

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
)

const (
	productESSyncAction   = "pms.product_spu.es_sync"
	productESDeleteAction = "pms.product_spu.es_delete"
)

func sendProductESSync(ctx context.Context, svcCtx *svc.ServiceContext, productID int64, current pkgscope.GovernanceScope) {
	if svcCtx == nil || svcCtx.RabbitMQ == nil {
		return
	}
	traceID := audit.NewTraceID(productESSyncAction, productID)
	body, _ := sonic.Marshal(pkgscope.NewProductESSyncPayload(productID, current, traceID))
	if err := svcCtx.RabbitMQ.SendMessage("product.event.exchange", "syn.product.to.es.queue", "syn.product.key", body); err != nil {
		logc.Errorf(ctx, "发送商品ES同步消息失败,spuId:%d,traceId:%s,scope:%+v,异常:%s", productID, traceID, current, err.Error())
	}
}

func sendProductESSyncBatch(ctx context.Context, svcCtx *svc.ServiceContext, productIDs []int64, current pkgscope.GovernanceScope) {
	for _, productID := range pkgscope.UniquePositiveIDs(productIDs) {
		sendProductESSync(ctx, svcCtx, productID, current)
	}
}

func sendProductESDelete(ctx context.Context, svcCtx *svc.ServiceContext, productIDs []int64, current pkgscope.GovernanceScope) {
	if svcCtx == nil || svcCtx.RabbitMQ == nil {
		return
	}
	uniqueIDs := pkgscope.UniquePositiveIDs(productIDs)
	if len(uniqueIDs) == 0 {
		return
	}

	traceID := audit.NewTraceID(productESDeleteAction, uniqueIDs[0])
	body, _ := sonic.Marshal(pkgscope.NewProductESDeletePayload(uniqueIDs, current, traceID))
	if err := svcCtx.RabbitMQ.SendMessage("product.event.exchange", "delete.product.from.es.queue", "delete.product.key", body); err != nil {
		logc.Errorf(ctx, "发送商品ES删除消息失败,ids:%+v,traceId:%s,scope:%+v,异常:%s", uniqueIDs, traceID, current, err.Error())
	}
}
