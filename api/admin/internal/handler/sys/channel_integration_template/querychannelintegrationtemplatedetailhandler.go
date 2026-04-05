package channel_integration_template

import (
	"net/http"

	channelintegrationtemplatelogic "github.com/feihua/zero-admin/api/admin/internal/logic/sys/channel_integration_template"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryChannelIntegrationTemplateDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryChannelIntegrationTemplateDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := channelintegrationtemplatelogic.NewQueryChannelIntegrationTemplateDetailLogic(r.Context(), svcCtx)
		resp, err := l.QueryChannelIntegrationTemplateDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
