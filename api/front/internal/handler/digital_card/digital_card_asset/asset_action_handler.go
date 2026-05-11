package digital_card_asset

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/digital_card/digital_card_asset"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Story 10.7 Review #5 H1: 越权直转 API 已弃用
// 原 `TransferDigitalCardAssetHandler`/`ResolveDigitalCardTransferRecipientHandler` 是
// 「转赠人输入接收人手机号 → 直接更新 member_id」的越权直转实现，违反 10.7 的 claim_token 流程。
// 现在统一返回 410 Gone，强制使用 GenerateClaimToken + ConsumeClaimToken（分享凭证流程）。
// 原 logic 函数体保留只为防止 dev 误删 import；下一迭代会随产品规格 A-M 一起整体下线。
const deprecatedTransferMessage = "转赠 API 已下线，请使用『分享给朋友』功能（分享凭证流程）"

func ResolveDigitalCardTransferRecipientHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logc.Errorf(r.Context(), "弃用接口被调用: resolveDigitalCardTransferRecipient, ua=%s", r.UserAgent())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"code":410,"message":"` + deprecatedTransferMessage + `","data":null}`))
	}
}

func TransferDigitalCardAssetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logc.Errorf(r.Context(), "弃用接口被调用: transferDigitalCardAsset, ua=%s", r.UserAgent())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"code":410,"message":"` + deprecatedTransferMessage + `","data":null}`))
	}
}

func RequestDigitalCardWithdrawHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RequestDigitalCardWithdrawReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := digital_card_asset.NewRequestDigitalCardWithdrawLogic(r.Context(), svcCtx)
		resp, err := l.RequestDigitalCardWithdraw(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
