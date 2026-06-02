package product

import (
	"context"
	"errors"

	"github.com/bytedance/sonic"

	"github.com/feihua/zero-admin/consumer/internal/svc"
	"github.com/feihua/zero-admin/consumer/internal/types"
	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductFromEsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductFromEsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductFromEsLogic {
	return &DeleteProductFromEsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteProductFromEs 测试删除商品
func (l *DeleteProductFromEsLogic) DeleteProductFromEs(req *types.ProductEsReq) (resp *types.Response, err error) {
	current, err := pkgscope.NormalizeGovernanceScope(req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, errors.New("删除 ES 商品索引必须携带合法的治理范围")
	}

	traceID := audit.NewTraceID("consumer.product_es.delete", 0)
	if ids := pkgscope.UniquePositiveIDs(req.Ids); len(ids) > 0 {
		traceID = audit.NewTraceID("consumer.product_es.delete", ids[0])
	}

	message := pkgscope.NewProductESDeletePayload(req.Ids, current, traceID, pkgscope.ProductEventMeta{
		Action: "consumer.product_es.delete",
	})
	body, _ := sonic.Marshal(message)
	err = l.svcCtx.RabbitMQ.SendMessage("product.event.exchange", "topic", "pms.product.delete.queue", "pms.product.deleted.key", body)

	return &types.Response{
		Message: "从es删除商品索引成功",
	}, nil
}
