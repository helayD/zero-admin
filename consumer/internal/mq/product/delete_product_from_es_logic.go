package product

import (
	"context"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/client/productspuservice"
	"github.com/feihua/zero-admin/rpc/search/search_client"
	"github.com/zeromicro/go-zero/core/logc"
)

// DeleteProductFromEs 删除es中商品的索引
func DeleteProductFromEs(ctx context.Context, body []byte, Search search_client.Search, productSpuService productspuservice.ProductSpuService) {
	logc.Infof(ctx, "[Consumer→ES] 收到商品ES删除消息, body=%s", body)
	payload, current, err := pkgscope.DecodeProductESDeletePayload(body)
	if err != nil {
		logc.Errorf(ctx, "[Consumer→ES] 解析商品ES删除消息失败: %v", err)
		return
	}
	logc.Infof(ctx, "[Consumer→ES] 处理商品ES删除消息, ids=%+v, traceId=%s, action=%s, actor=%d/%s, version=%d, scope=%s/%d/%d/%d",
		payload.IDs, payload.TraceID, payload.Action, payload.ActorID, payload.ActorName, payload.Version,
		current.ScopeType, current.PlatformID, current.TenantID, current.MerchantID)

	_, err = Search.Delete(ctx, &search_client.DeleteReq{
		Ids: payload.IDs,
	})

	if err != nil {
		logc.Errorf(ctx, "[Consumer→ES] 删除商品ES索引失败, ids=%+v, traceId=%s, err=%v", payload.IDs, payload.TraceID, err)
		return
	}

	logc.Infof(ctx, "[Consumer→ES] 删除商品ES索引成功, ids=%+v, traceId=%s, scope=%s/%d/%d/%d",
		payload.IDs, payload.TraceID, current.ScopeType, current.PlatformID, current.TenantID, current.MerchantID)
}
