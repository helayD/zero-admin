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

func QueryMyDrawRecordListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryMyDrawRecordListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := draw_activity.NewQueryMyDrawRecordListLogic(r.Context(), svcCtx)
		resp, err := l.QueryMyDrawRecordList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
