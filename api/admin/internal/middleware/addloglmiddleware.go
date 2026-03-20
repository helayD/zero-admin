package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/feihua/zero-admin/rpc/sys/client/operatelogservice"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/ua-parser/uap-go/uaparser"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"io/ioutil"
	"net/http"
	"time"

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
		currentScope, scopeExtra := readGovernanceScopeContext(r)
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
			Extra:           scopeExtra,
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

func readGovernanceScopeContext(r *http.Request) (scope.GovernanceScope, string) {
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

	extra, _ := json.Marshal(map[string]interface{}{
		"scopeType":  currentScope.ScopeType,
		"scopeLabel": currentScope.Label(),
		"platformId": currentScope.PlatformID,
		"tenantId":   currentScope.TenantID,
		"merchantId": currentScope.MerchantID,
	})

	return currentScope, string(extra)
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
