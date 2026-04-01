package types

// RetryChainReq 重试链路
type RetryChainReq struct {
	OrderId int64  `json:"orderId"`
	Remark  string `json:"remark"`
}

// RetryChainResp 重试响应
type RetryChainResp struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	NewRetryCount int64  `json:"newRetryCount"`
	TraceId       string `json:"traceId"`
	Success       bool   `json:"success"`
}

// ReplayChainReq 回放链路
type ReplayChainReq struct {
	OrderId      int64  `json:"orderId"`
	ReplayReason string `json:"replayReason"`
}

// ReplayChainResp 回放响应
type ReplayChainResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceId string `json:"traceId"`
	Success bool   `json:"success"`
}

// PauseChainReq 暂停链路
type PauseChainReq struct {
	OrderId     int64  `json:"orderId"`
	PauseReason string `json:"pauseReason"`
}

// PauseChainResp 暂停响应
type PauseChainResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Paused  bool   `json:"paused"`
	Success bool   `json:"success"`
}

// EscalateChainReq 升级链路
type EscalateChainReq struct {
	OrderId         int64  `json:"orderId"`
	EscalateReason string `json:"escalateReason"`
}

// EscalateChainResp 升级响应
type EscalateChainResp struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	ManualRequired bool   `json:"manualRequired"`
	Success        bool   `json:"success"`
}

// ChainActionsReq 查询可用干预动作
type ChainActionsReq struct {
	OrderId int64 `json:"orderId"`
}

// ChainActionsResp 查询可用干预动作响应
type ChainActionsResp struct {
	Code              string   `json:"code"`
	Message           string   `json:"message"`
	AvailableActions  []string `json:"availableActions"`
	Success           bool     `json:"success"`
}
