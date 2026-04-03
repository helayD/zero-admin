// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package comment

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/pms/comment"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryCommentAuditLogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryCommentAuditLogReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := comment.NewQueryCommentAuditLogLogic(r.Context(), svcCtx)
		resp, err := l.QueryCommentAuditLog(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
