package product_fulfillment_conflict

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/pms/product_fulfillment_conflict"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// QueryProductFulfillmentConflictHandler 查询商品履约模式配置冲突
func QueryProductFulfillmentConflictHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryProductFulfillmentConflictReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := product_fulfillment_conflict.NewQueryProductFulfillmentConflictLogic(r.Context(), svcCtx)
		resp, err := l.QueryProductFulfillmentConflict(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
