// Story 3.1.1: POST /api/member/auth/sms/send
// 发送手机号短信验证码（mock 模式固定 123456）。无需 JWT。

package member

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/member/member"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuthSendSmsCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSmsCodeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := member.NewAuthSendSmsCodeLogic(r.Context(), svcCtx)
		resp, err := l.AuthSendSmsCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
