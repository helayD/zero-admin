package digital_card_physical_fulfillment

import (
	"net/http"

	physicalfulfillmentlogic "github.com/feihua/zero-admin/api/admin/internal/logic/sms/digital_card_physical_fulfillment"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryDigitalCardPhysicalFulfillmentListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryDigitalCardPhysicalFulfillmentListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewQueryDigitalCardPhysicalFulfillmentListLogic(r.Context(), svcCtx)
		resp, err := l.QueryDigitalCardPhysicalFulfillmentList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func QueryDigitalCardPhysicalFulfillmentDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryDigitalCardPhysicalFulfillmentDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewQueryDigitalCardPhysicalFulfillmentDetailLogic(r.Context(), svcCtx)
		resp, err := l.QueryDigitalCardPhysicalFulfillmentDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func EnsureDigitalCardPhysicalFulfillmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.EnsureDigitalCardPhysicalFulfillmentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewEnsureDigitalCardPhysicalFulfillmentLogic(r.Context(), svcCtx)
		resp, err := l.EnsureDigitalCardPhysicalFulfillment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func UpdateDigitalCardPhysicalProductionStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateDigitalCardPhysicalProductionStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewUpdateDigitalCardPhysicalProductionStatusLogic(r.Context(), svcCtx)
		resp, err := l.UpdateDigitalCardPhysicalProductionStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func ShipDigitalCardPhysicalFulfillmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ShipDigitalCardPhysicalFulfillmentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewShipDigitalCardPhysicalFulfillmentLogic(r.Context(), svcCtx)
		resp, err := l.ShipDigitalCardPhysicalFulfillment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func MarkDigitalCardPhysicalFulfillmentExceptionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DigitalCardPhysicalFulfillmentExceptionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewMarkDigitalCardPhysicalFulfillmentExceptionLogic(r.Context(), svcCtx)
		resp, err := l.MarkDigitalCardPhysicalFulfillmentException(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func RequestDigitalCardPhysicalReissueHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DigitalCardPhysicalFulfillmentExceptionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physicalfulfillmentlogic.NewRequestDigitalCardPhysicalReissueLogic(r.Context(), svcCtx)
		resp, err := l.RequestDigitalCardPhysicalReissue(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
