package chain_monitor

import (
	"net/http"
	"time"

	"github.com/feihua/zero-admin/api/admin/internal/logic/oms/chain_monitor"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ExportChainMonitorListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExportChainMonitorListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chain_monitor.NewExportChainMonitorListLogic(r.Context(), svcCtx)
		data, err := l.ExportChainMonitorList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		filename := "链路监控_" + time.Now().Format("2006-01-02") + ".xlsx"
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+filename)
		w.Write(data)
	}
}
