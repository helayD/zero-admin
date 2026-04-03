package chain_monitor

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/oms/chain_monitor"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryChainMonitorListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := bindQueryChainMonitorListReq(r)

		l := chain_monitor.NewQueryChainMonitorListLogic(r.Context(), svcCtx)
		resp, err := l.QueryChainMonitorList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
