package types

// 链路监控查询请求
type QueryChainMonitorListReq struct {
	ScopeType          string `json:"scopeType"`
	PlatformId         int64  `json:"platformId"`
	TenantId           int64  `json:"tenantId"`
	MerchantId         int64  `json:"merchantId"`
	PageSize           int    `json:"pageSize"`
	Current            int    `json:"current"`
	ChainType          int32  `json:"chainType"`           // 链路类型
	ConsistencyStage   int32  `json:"consistencyStage"`    // 一致性阶段
	ConsistencyResult  int32  `json:"consistencyResult"`   // 处理结果
	ManualRequired     int32  `json:"manualRequired"`      // 人工介入
	StartTime          string `json:"startTime"`
	EndTime            string `json:"endTime"`
	OrderNo            string `json:"orderNo"`
}

// 链路监控项
type ChainMonitorItem struct {
	TraceId        string `json:"traceId"`        // 链路追踪ID
	PlatformId     int64  `json:"platformId"`     // 平台ID
	TenantId       int64  `json:"tenantId"`       // 租户ID
	MerchantId     int64  `json:"merchantId"`     // 商户ID
	ChainType      int32  `json:"chainType"`      // 链路类型
	ChainTypeText  string `json:"chainTypeText"`  // 链路类型文本
	EntityId       int64  `json:"entityId"`       // 关联业务对象ID
	EntityNo       string `json:"entityNo"`       // 关联业务对象编号
	EntityType     int32  `json:"entityType"`     // 业务对象类型
	Stage          int32  `json:"stage"`          // 一致性阶段
	StageText      string `json:"stageText"`      // 阶段文本
	Result         int32  `json:"result"`         // 处理结果
	ResultText     string `json:"resultText"`     // 结果文本
	RetryCount     int32  `json:"retryCount"`     // 重试次数
	LastError      string `json:"lastError"`     // 最近错误摘要
	LastExecuteAt  string `json:"lastExecuteAt"` // 最近执行时间
	CreatedAt      string `json:"createdAt"`      // 链路创建时间
	ActorId        int64  `json:"actorId"`        // 触发操作人ID
}

type QueryChainMonitorListResp struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Data    []*ChainMonitorItem `json:"data"`
	Current int                `json:"current"`
	PageSize int               `json:"pageSize"`
	Total   int64              `json:"total"`
	Success bool               `json:"success"`
}

// ExportChainMonitorListReq 导出链路监控列表（与查询共用同一结构，只是 pageSize 固定为 10000）
type ExportChainMonitorListReq = QueryChainMonitorListReq
