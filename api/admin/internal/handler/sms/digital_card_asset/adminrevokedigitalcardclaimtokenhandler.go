package digital_card_asset

import (
	"net/http"

	"github.com/feihua/zero-admin/api/admin/internal/logic/sms/digital_card_asset"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// AdminRevokeDigitalCardClaimTokenHandler Story 10.7 Task 8.8 / S5 — 后台手动吊销分享凭证
func AdminRevokeDigitalCardClaimTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminRevokeDigitalCardClaimTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := digital_card_asset.NewAdminRevokeDigitalCardClaimTokenLogic(r.Context(), svcCtx)
		resp, err := l.AdminRevokeDigitalCardClaimToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
