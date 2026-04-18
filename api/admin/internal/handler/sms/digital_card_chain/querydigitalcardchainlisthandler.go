package digital_card_chain

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/sms/digital_card_chain"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryDigitalCardChainListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryDigitalCardChainListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := digital_card_chain.NewQueryDigitalCardChainListLogic(r.Context(), svcCtx)
		resp, err := l.QueryDigitalCardChainList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
