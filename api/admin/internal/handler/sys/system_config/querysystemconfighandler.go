package system_config

import (
	"net/http"

	systemconfiglogic "github.com/feihua/zero-admin/api/admin/internal/logic/sys/system_config"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QuerySystemConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := systemconfiglogic.NewQuerySystemConfigLogic(r.Context(), svcCtx)
		resp, err := l.QuerySystemConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
