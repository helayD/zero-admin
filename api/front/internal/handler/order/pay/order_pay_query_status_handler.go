package pay

import (
	"net/http"
	"strconv"

	"github.com/feihua/zero-admin/api/front/internal/logic/order/pay"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// OrderPayQueryStatusHandler 支付状态查询（Flutter query 参数版本）
// Flutter 使用: GET /api/order/orderPayQueryStatus?orderId=xxx
func OrderPayQueryStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderIdStr := r.URL.Query().Get("orderId")
		orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
		if err != nil || orderId <= 0 {
			httpx.OkJsonCtx(r.Context(), w, &types.OrderPayQueryResp{
				Code:    1,
				Message: "无效的订单ID",
			})
			return
		}

		req := &types.OrderPayQueryReq{OrderId: orderId}
		l := pay.NewOrderPayQueryLogic(r.Context(), svcCtx)
		resp, err := l.OrderPayQuery(req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
