package types

// ProductFulfillmentConflictItem 商品履约模式配置冲突项
type ProductFulfillmentConflictItem struct {
	ProductId   int64  `json:"productId"`   // 商品ID
	ProductName string `json:"productName"` // 商品名称
	HasConflict bool   `json:"hasConflict"` // 是否有冲突
	ConflictMsg string `json:"conflictMsg"` // 冲突信息
}

// QueryProductFulfillmentConflictReq 查询商品履约模式配置冲突请求
type QueryProductFulfillmentConflictReq struct {
	Ids        []int64 `json:"ids"`                  // 商品ID列表
	ScopeType  string  `json:"scopeType,optional"`   // 治理范围
	PlatformId int64   `json:"platformId,optional"`  // 平台ID
	TenantId   int64   `json:"tenantId,optional"`    // 租户ID
	MerchantId int64   `json:"merchantId,optional"`  // 商户ID
}

// QueryProductFulfillmentConflictResp 查询商品履约模式配置冲突响应
type QueryProductFulfillmentConflictResp struct {
	Code    string                              `json:"code"`
	Message string                              `json:"message"`
	Data    []ProductFulfillmentConflictItem    `json:"data"`
}
