package product_fulfillment_rule

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/sms/product_fulfillment_rule"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateProductFulfillmentRuleStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateProductFulfillmentRuleStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := product_fulfillment_rule.NewUpdateProductFulfillmentRuleStatusLogic(r.Context(), svcCtx)
		resp, err := l.UpdateProductFulfillmentRuleStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
