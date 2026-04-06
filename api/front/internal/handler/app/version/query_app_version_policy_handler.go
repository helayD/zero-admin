// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package version

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/app/version"
	logiccommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryAppVersionPolicyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AppVersionPolicyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		ctx := logiccommon.WithClientRequestMetadata(
			r.Context(),
			logiccommon.ReadClientRequestMetadata(r),
		)
		l := version.NewQueryAppVersionPolicyLogic(ctx, svcCtx)
		resp, err := l.QueryAppVersionPolicy(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
