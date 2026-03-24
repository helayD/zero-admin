package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/sys/client/operatelogservice"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/ua-parser/uap-go/uaparser"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/feihua/zero-admin/pkg/audit"
	"github.com/feihua/zero-admin/pkg/scope"
)

type AddLogMiddleware struct {
	Sys operatelogservice.OperateLogService
}

func NewAddLogMiddleware(Sys operatelogservice.OperateLogService) *AddLogMiddleware {
	return &AddLogMiddleware{Sys: Sys}
}

func (m *AddLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		uri := r.RequestURI

		var userName string
		var deptName string
		// 从上下文中获取 userName，并进行类型断言和 nil 检查
		if userNameRaw, ok := r.Context().Value("userName").(string); !ok || userNameRaw == "" {
			next(w, r)
			return
		} else {
			userName = userNameRaw
		}

		// 从上下文中获取 deptName，并进行类型断言和 nil 检查
		if deptNameRaw, ok := r.Context().Value("deptName").(string); !ok || deptNameRaw == "" {
			next(w, r)
			return
		} else {
			deptName = deptNameRaw
		}

		startTime := time.Now()

		// 读取请求主体
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("Failed to read request body: %v", err)
		}

		// 创建一个新的请求主体用于后续读取
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// 打印请求参数日志
		logx.WithContext(r.Context()).Infof("Request: %s %s %s", r.Method, uri, body)

		// 创建一个自定义的 ResponseWriter，用于记录响应
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           make([]byte, 0),
		}

		// 调用下一个处理器，捕获响应
		next(recorder, r)

		// 打印响应日志
		responseBoy := string(recorder.body)
		// 响应参数较多,可以不打印
		// logx.WithContext(r.Context()).Infof("Response: %s %s %s", r.Method, r.RequestURI, responseBoy)

		userAgent := r.Header.Get("User-Agent")
		parser := uaparser.NewFromSaved()
		ua := parser.Parse(userAgent)

		browser := ua.UserAgent.Family + " " + ua.UserAgent.Major
		os := ua.Os.Family + " " + ua.Os.Major
		currentScope := readGovernanceScopeContext(r)
		extra := buildGovernanceOperateExtra(r, body, recorder.body, currentScope)
		// 打印请求和响应耗时
		duration := time.Since(startTime)
		opLog := &sysclient.AddOperateLogReq{
			Title:           "",
			BusinessType:    0,
			Method:          r.Method,
			RequestMethod:   r.Method,
			OperatorType:    0,
			OperateName:     userName,
			DeptName:        deptName,
			OperateUrl:      uri,
			OperateIp:       httpx.GetRemoteAddr(r),
			OperateLocation: "",
			OperateParam:    string(body),
			JsonResult:      responseBoy,
			Platform:        currentScope.Label(),
			Browser:         browser,
			Version:         "",
			Os:              os,
			Arch:            "",
			Engine:          "",
			EngineDetails:   "",
			Extra:           extra,
			Status:          0,
			ErrorMsg:        "",
			OperateTime:     "",
			CostTime:        duration.Milliseconds(),
		}
		_, _ = m.Sys.AddOperateLog(r.Context(), opLog)
	}
}

// 自定义的 ResponseWriter
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// WriteHeader 重写 WriteHeader 方法，捕获状态码
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// 重写 Write 方法，捕获响应数据
func (r *responseRecorder) Write(body []byte) (int, error) {
	r.body = body
	return r.ResponseWriter.Write(body)
}

func readGovernanceScopeContext(r *http.Request) scope.GovernanceScope {
	scopeType, _ := r.Context().Value("scopeType").(string)
	platformID := readContextInt64(r, "platformId")
	tenantID := readContextInt64(r, "tenantId")
	merchantID := readContextInt64(r, "merchantId")

	currentScope, err := scope.NormalizeGovernanceScope(scopeType, platformID, tenantID, merchantID)
	if err != nil {
		currentScope = scope.GovernanceScope{
			ScopeType:  scope.SubjectTypePlatform,
			PlatformID: scope.DefaultPlatformID,
		}
	}

	return currentScope
}

func readContextInt64(r *http.Request, key string) int64 {
	value := r.Context().Value(key)
	switch typed := value.(type) {
	case json.Number:
		number, _ := typed.Int64()
		return number
	case float64:
		return int64(typed)
	case float32:
		return int64(typed)
	case int64:
		return typed
	case int32:
		return int64(typed)
	case int:
		return int64(typed)
	default:
		return 0
	}
}

func buildGovernanceOperateExtra(
	r *http.Request,
	body []byte,
	responseBody []byte,
	currentScope scope.GovernanceScope,
) string {
	action, resourceType := deriveOperateActionAndResource(r.RequestURI)
	resourceIDs, payload := extractOperateResourceIDs(r, body)
	resourceID := int64(0)
	if len(resourceIDs) > 0 {
		resourceID = resourceIDs[0]
	}

	requestSummary := buildOperateRequestSummary(r.Method, action, resourceIDs, payload)
	result := operateResultFromResponse(responseBody)
	traceID := audit.NewTraceID(action, resourceID)

	extra := map[string]interface{}{
		"traceId":        traceID,
		"action":         action,
		"resourceType":   resourceType,
		"resourceId":     resourceID,
		"resourceIds":    resourceIDs,
		"scopeType":      currentScope.ScopeType,
		"scopeLabel":     currentScope.Label(),
		"platformId":     currentScope.PlatformID,
		"tenantId":       currentScope.TenantID,
		"merchantId":     currentScope.MerchantID,
		"result":         result,
		"requestSummary": requestSummary,
	}

	raw, err := json.Marshal(extra)
	if err != nil {
		fallback, _ := audit.EncodeGovernancePayload(audit.GovernancePayload{
			TraceID:        traceID,
			Action:         action,
			ResourceType:   resourceType,
			ResourceID:     resourceID,
			ScopeType:      currentScope.ScopeType,
			PlatformID:     currentScope.PlatformID,
			TenantID:       currentScope.TenantID,
			MerchantID:     currentScope.MerchantID,
			ScopeLabel:     currentScope.Label(),
			Result:         result,
			RequestSummary: requestSummary,
		})
		return fallback
	}

	if len(raw) <= 1000 {
		return string(raw)
	}

	extra["requestSummary"] = trimString(requestSummary, 180)
	raw, err = json.Marshal(extra)
	if err != nil {
		fallback, _ := audit.EncodeGovernancePayload(audit.GovernancePayload{
			TraceID:        traceID,
			Action:         action,
			ResourceType:   resourceType,
			ResourceID:     resourceID,
			ScopeType:      currentScope.ScopeType,
			PlatformID:     currentScope.PlatformID,
			TenantID:       currentScope.TenantID,
			MerchantID:     currentScope.MerchantID,
			ScopeLabel:     currentScope.Label(),
			Result:         result,
			RequestSummary: trimString(requestSummary, 180),
		})
		return fallback
	}
	if len(raw) <= 1000 {
		return string(raw)
	}

	minimal, _ := audit.EncodeGovernancePayload(audit.GovernancePayload{
		TraceID:        traceID,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		ScopeType:      currentScope.ScopeType,
		PlatformID:     currentScope.PlatformID,
		TenantID:       currentScope.TenantID,
		MerchantID:     currentScope.MerchantID,
		ScopeLabel:     currentScope.Label(),
		Result:         result,
		RequestSummary: trimString(requestSummary, 120),
	})

	return minimal
}

func deriveOperateActionAndResource(uri string) (string, string) {
	cleanURI := strings.TrimSpace(strings.Split(uri, "?")[0])
	action := strings.TrimSpace(path.Base(cleanURI))
	if action == "" || action == "." || action == "/" {
		action = "unknown"
	}

	resourceType := "unknown"
	switch {
	case strings.Contains(cleanURI, "/api/pms/product/"):
		if strings.Contains(strings.ToLower(cleanURI), "sku") {
			resourceType = "product_sku"
		} else {
			resourceType = "product_spu"
		}
	case strings.Contains(cleanURI, "/api/oms/order/"):
		resourceType = "order"
	case strings.Contains(cleanURI, "/api/sms/coupon"):
		resourceType = "coupon"
	case strings.Contains(cleanURI, "/api/cms/subject"):
		resourceType = "subject"
	case strings.Contains(cleanURI, "/api/ums/level"):
		resourceType = "member_level"
	case strings.Contains(cleanURI, "/api/ums/tag"):
		resourceType = "member_tag"
	case strings.Contains(cleanURI, "/api/ums/task"):
		resourceType = "member_task"
	case strings.Contains(cleanURI, "/api/ums/ruleSetting"):
		resourceType = "member_rule_setting"
	case strings.Contains(cleanURI, "/api/ums/consumeSetting"):
		resourceType = "member_consume_setting"
	}

	return action, resourceType
}

func extractOperateResourceIDs(r *http.Request, body []byte) ([]int64, map[string]interface{}) {
	ids := make([]int64, 0)
	if len(body) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err == nil {
			collectOperateResourceIDs(payload, &ids)
			return normalizeOperateIDs(ids), payload
		}
	}

	payload := make(map[string]interface{})
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue
		}
		payload[key] = values[0]
	}
	collectQueryIDs(r.URL.Query(), &ids)
	return normalizeOperateIDs(ids), payload
}

func collectOperateResourceIDs(input interface{}, ids *[]int64) {
	switch typed := input.(type) {
	case map[string]interface{}:
		for key, value := range typed {
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if lowerKey == "id" || lowerKey == "ids" || lowerKey == "orderid" {
				collectIDValue(value, ids)
			}
			collectOperateResourceIDs(value, ids)
		}
	case []interface{}:
		for _, value := range typed {
			collectOperateResourceIDs(value, ids)
		}
	}
}

func collectIDValue(value interface{}, ids *[]int64) {
	switch typed := value.(type) {
	case float64:
		*ids = append(*ids, int64(typed))
	case int64:
		*ids = append(*ids, typed)
	case string:
		for _, item := range strings.Split(typed, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			number, err := strconv.ParseInt(item, 10, 64)
			if err == nil {
				*ids = append(*ids, number)
			}
		}
	case []interface{}:
		for _, item := range typed {
			collectIDValue(item, ids)
		}
	}
}

func collectQueryIDs(values map[string][]string, ids *[]int64) {
	for key, value := range values {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if lowerKey != "id" && lowerKey != "ids" && lowerKey != "orderid" {
			continue
		}
		for _, item := range value {
			collectIDValue(item, ids)
		}
	}
}

func normalizeOperateIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return []int64{}
	}

	seen := make(map[int64]struct{})
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })

	return result
}

func operateResultFromResponse(responseBody []byte) string {
	type response struct {
		Code string `json:"code"`
	}

	var resp response
	if len(responseBody) > 0 && json.Unmarshal(responseBody, &resp) == nil && resp.Code == "000000" {
		return "success"
	}

	return "failed"
}

func buildOperateRequestSummary(
	method string,
	action string,
	resourceIDs []int64,
	payload map[string]interface{},
) string {
	parts := []string{strings.ToUpper(strings.TrimSpace(method)), action}
	if len(resourceIDs) > 0 {
		parts = append(parts, fmt.Sprintf("ids=%v", resourceIDs))
	}

	appendPayloadSummary(&parts, payload, "status")
	appendPayloadSummary(&parts, payload, "showStatus")
	appendPayloadSummary(&parts, payload, "recommendStatus")
	appendPayloadSummary(&parts, payload, "note")
	appendPayloadSummary(&parts, payload, "deliveryCompany")
	appendPayloadSummary(&parts, payload, "deliverySn")

	return trimString(strings.Join(parts, " "), 500)
}

func appendPayloadSummary(parts *[]string, payload map[string]interface{}, key string) {
	if payload == nil {
		return
	}
	value, ok := payload[key]
	if !ok {
		return
	}
	text := trimString(strings.TrimSpace(fmt.Sprint(value)), 80)
	if text == "" {
		return
	}
	*parts = append(*parts, fmt.Sprintf("%s=%s", key, text))
}

func trimString(value string, limit int) string {
	if limit <= 0 || value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
