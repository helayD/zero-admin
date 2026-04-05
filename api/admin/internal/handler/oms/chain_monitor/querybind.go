package chain_monitor

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/types"
)

func bindQueryChainMonitorListReq(r *http.Request) types.QueryChainMonitorListReq {
	query := r.URL.Query()

	req := types.QueryChainMonitorListReq{
		ScopeType:         strings.TrimSpace(query.Get("scopeType")),
		PlatformId:        parseQueryInt64(query.Get("platformId"), 0),
		TenantId:          parseQueryInt64(query.Get("tenantId"), 0),
		MerchantId:        parseQueryInt64(query.Get("merchantId"), 0),
		PageSize:          parseQueryInt(query.Get("pageSize"), 20),
		Current:           parseQueryInt(query.Get("current"), 1),
		ChainType:         parseQueryInt32(query.Get("chainType"), 0),
		ConsistencyStage:  parseQueryInt32(query.Get("consistencyStage"), 0),
		ConsistencyResult: parseQueryInt32(query.Get("consistencyResult"), 0),
		ManualRequired:    parseQueryInt32(query.Get("manualRequired"), 0),
		StartTime:         strings.TrimSpace(query.Get("startTime")),
		EndTime:           strings.TrimSpace(query.Get("endTime")),
		OrderNo:           strings.TrimSpace(query.Get("orderNo")),
	}

	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Current <= 0 {
		req.Current = 1
	}

	return req
}

func bindChainActionsReq(r *http.Request) (types.ChainActionsReq, error) {
	query := r.URL.Query()
	orderIDText := strings.TrimSpace(query.Get("orderId"))
	if orderIDText == "" {
		return types.ChainActionsReq{}, errors.New("field \"orderId\" is not set")
	}

	orderID, err := strconv.ParseInt(orderIDText, 10, 64)
	if err != nil {
		return types.ChainActionsReq{}, err
	}

	return types.ChainActionsReq{OrderId: orderID}, nil
}

func parseQueryInt(text string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return fallback
	}
	return value
}

func parseQueryInt32(text string, fallback int32) int32 {
	value, err := strconv.ParseInt(strings.TrimSpace(text), 10, 32)
	if err != nil {
		return fallback
	}
	return int32(value)
}

func parseQueryInt64(text string, fallback int64) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
