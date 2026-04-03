package chain_monitor

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/oms/chain_monitor"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryChainActionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := bindChainActionsReq(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chain_monitor.NewQueryChainActionsLogic(r.Context(), svcCtx)
		resp, err := l.QueryChainActions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
