// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package draw_activity

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/digital_card/draw_activity"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryMyDrawActivityLandingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DrawActivityLandingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := draw_activity.NewQueryMyDrawActivityLandingLogic(r.Context(), svcCtx)
		resp, err := l.QueryMyDrawActivityLanding(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
