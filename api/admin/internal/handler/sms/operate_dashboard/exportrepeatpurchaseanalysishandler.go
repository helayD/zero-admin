package operate_dashboard

import (
	"fmt"
	"net/http"
	"time"

	"github.com/feihua/zero-admin/api/admin/internal/logic/sms/operate_dashboard"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ExportRepeatPurchaseAnalysisHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExportRepeatPurchaseAnalysisReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := operate_dashboard.NewExportRepeatPurchaseAnalysisLogic(r.Context(), svcCtx)
		data, err := l.ExportRepeatPurchaseAnalysis(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		filename := fmt.Sprintf("复购分析_%s.xlsx", time.Now().Format("2006-01-02"))
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
