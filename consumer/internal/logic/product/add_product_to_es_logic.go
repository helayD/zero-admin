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

type AddProductToEsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddProductToEsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductToEsLogic {
	return &AddProductToEsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AddProductToEs 同步商品到es
func (l *AddProductToEsLogic) AddProductToEs(req *types.ProductEsReq) (resp *types.Response, err error) {
	current, err := pkgscope.NormalizeGovernanceScope(req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, errors.New("同步 ES 必须携带合法的治理范围")
	}

	for _, id := range req.Ids {
		traceID := audit.NewTraceID("consumer.product_es.sync", id)
		message := pkgscope.NewProductESSyncPayload(id, current, traceID, pkgscope.ProductEventMeta{
			Action: "consumer.product_es.sync",
		})
		body, _ := sonic.Marshal(message)
		err = l.svcCtx.RabbitMQ.SendMessage("product.event.exchange", "syn.product.to.es.queue", "syn.product.key", body)
		if err != nil {
			return nil, err
		}
	}

	return &types.Response{
		Message: "同步商品到es成功",
	}, nil
}
