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
	logc.Infof(ctx, "需要删除es中商品的索引信息: %s", body)
	payload, current, err := pkgscope.DecodeProductESDeletePayload(body)
	if err != nil {
		logc.Errorf(ctx, "解析商品 ES 删除消息失败: %v", err)
		return
	}
	logc.Infof(ctx, "处理商品ES删除消息,ids:%+v,traceId:%s,action:%s,actor:%d/%s,version:%d,scope:%s/%d/%d/%d", payload.IDs, payload.TraceID, payload.Action, payload.ActorID, payload.ActorName, payload.Version, current.ScopeType, current.PlatformID, current.TenantID, current.MerchantID)

	_, err = Search.Delete(ctx, &search_client.DeleteReq{
		Ids: payload.IDs,
	})

	if err != nil {
		logc.Errorf(ctx, "删除es中商品的索引信息失败,请求参数：%s,错误信息：%+v", body, err)
		return
	}

	logc.Errorf(ctx, "删除es中商品的索引信息成功,请求参数：%s", body)
}
