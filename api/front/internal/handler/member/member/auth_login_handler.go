// Story 3.1.1: POST /api/member/auth/login
// 手机号+验证码合并登录注册（已注册直接登录，未注册自动建号并登录）。无需 JWT。

package member

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/member/member"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuthLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := member.NewAuthLoginLogic(r.Context(), svcCtx)
		resp, err := l.AuthLogin(&req, httpx.GetRemoteAddr(r))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
