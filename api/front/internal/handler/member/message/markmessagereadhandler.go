package message

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/member/message"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func MarkMessageReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MarkMessageReadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := message.NewMarkMessageReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkMessageRead(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func MarkMessageReadByPathHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.MemberMessagePathReq
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := message.NewMarkMessageReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkMessageRead(&types.MarkMessageReadReq{ID: pathReq.ID})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
