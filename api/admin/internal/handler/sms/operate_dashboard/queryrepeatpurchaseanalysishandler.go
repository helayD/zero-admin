package operate_dashboard

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/sms/operate_dashboard"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryRepeatPurchaseAnalysisHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryRepeatPurchaseAnalysisReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := operate_dashboard.NewQueryRepeatPurchaseAnalysisLogic(r.Context(), svcCtx)
		resp, err := l.QueryRepeatPurchaseAnalysis(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
