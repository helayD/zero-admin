package digital_card_asset

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/digital_card/digital_card_asset"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// SendClaimVerifyCodeHandler POST /api/digitalCard/sendClaimVerifyCode（匿名）
func SendClaimVerifyCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendClaimVerifyCodeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := digital_card_asset.NewSendClaimVerifyCodeLogic(r.Context(), svcCtx)
		resp, err := l.SendClaimVerifyCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// ClaimByMobileHandler POST /api/digitalCard/claimByMobile（匿名）
func ClaimByMobileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClaimByMobileReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := digital_card_asset.NewClaimByMobileLogic(r.Context(), svcCtx, r)
		resp, err := l.ClaimByMobile(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// GetAppDownloadHandler GET /api/config/appDownload（匿名）
func GetAppDownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := digital_card_asset.NewGetAppDownloadLogic(r.Context(), svcCtx)
		resp, err := l.GetAppDownload()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
