package types

// QueryOperateFunnelDashboardReq 查询经营漏斗看板请求
type QueryOperateFunnelDashboardReq struct {
	ScopeType    string `form:"scopeType,optional" json:"scopeType"`
	PlatformId   int64  `form:"platformId,optional" json:"platformId"`
	TenantId     int64  `form:"tenantId,optional" json:"tenantId"`
	MerchantId   int64  `form:"merchantId,optional" json:"merchantId"`
	StartTime    string `form:"startTime,optional" json:"startTime"`
	EndTime      string `form:"endTime,optional" json:"endTime"`
	Channel      string `form:"channel,optional" json:"channel"`
	ActivityType string `form:"activityType,optional" json:"activityType"`
	ActivityId   int64  `form:"activityId,optional" json:"activityId"`
	Bucket       string `form:"bucket,optional" json:"bucket"`
}

// OperateFunnelMetricCard 看板指标卡
type OperateFunnelMetricCard struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Value     int64   `json:"value"`
	Rate      float64 `json:"rate"`
	RateLabel string  `json:"rateLabel"`
}

// OperateFunnelOverview 经营漏斗总览
type OperateFunnelOverview struct {
	Exposure         int64                     `json:"exposure"`
	Click            int64                     `json:"click"`
	AddCart          int64                     `json:"addCart"`
	OrderCreated     int64                     `json:"orderCreated"`
	PaySuccess       int64                     `json:"paySuccess"`
	CouponRedeem     int64                     `json:"couponRedeem"`
	ClickRate        float64                   `json:"clickRate"`
	AddCartRate      float64                   `json:"addCartRate"`
	OrderRate        float64                   `json:"orderRate"`
	PayRate          float64                   `json:"payRate"`
	CouponRedeemRate float64                   `json:"couponRedeemRate"`
	Cards            []OperateFunnelMetricCard `json:"cards"`
}

// OperateFunnelSeriesPoint 经营漏斗时间序列点
type OperateFunnelSeriesPoint struct {
	BucketLabel      string  `json:"bucketLabel"`
	BucketStart      string  `json:"bucketStart"`
	BucketEnd        string  `json:"bucketEnd"`
	Exposure         int64   `json:"exposure"`
	Click            int64   `json:"click"`
	AddCart          int64   `json:"addCart"`
	OrderCreated     int64   `json:"orderCreated"`
	PaySuccess       int64   `json:"paySuccess"`
	CouponRedeem     int64   `json:"couponRedeem"`
	ClickRate        float64 `json:"clickRate"`
	AddCartRate      float64 `json:"addCartRate"`
	OrderRate        float64 `json:"orderRate"`
	PayRate          float64 `json:"payRate"`
	CouponRedeemRate float64 `json:"couponRedeemRate"`
}

// OperateFunnelActivityOption 经营漏斗活动选项
type OperateFunnelActivityOption struct {
	ActivityType  string `json:"activityType"`
	ActivityId    int64  `json:"activityId"`
	ActivityName  string `json:"activityName"`
	ActivityLabel string `json:"activityLabel"`
}

// QueryOperateFunnelDashboardData 经营漏斗看板数据
type QueryOperateFunnelDashboardData struct {
	Overview          OperateFunnelOverview         `json:"overview"`
	Series            []OperateFunnelSeriesPoint    `json:"series"`
	ActivityOptions   []OperateFunnelActivityOption `json:"activityOptions"`
	TrackingStartedAt string                        `json:"trackingStartedAt"`
	PartialMetrics    []string                      `json:"partialMetrics"`
	Bucket            string                        `json:"bucket"`
}

// QueryOperateFunnelDashboardResp 查询经营漏斗看板响应
type QueryOperateFunnelDashboardResp struct {
	Code    string                          `json:"code"`
	Message string                          `json:"message"`
	Data    QueryOperateFunnelDashboardData `json:"data"`
	Success bool                            `json:"success"`
}

// RepeatPurchaseOverview 复购分析总览
type RepeatPurchaseOverview struct {
	PaidBuyerCount   int64   `json:"paidBuyerCount"`
	RepeatBuyerCount int64   `json:"repeatBuyerCount"`
	RepeatRate       float64 `json:"repeatRate"`
	RepeatOrderCount int64   `json:"repeatOrderCount"`
	RepeatGmv        float64 `json:"repeatGmv"`
	AvgDaysToRepeat  float64 `json:"avgDaysToRepeat"`
}

// RepeatPurchaseTrendPoint 复购分析趋势点
type RepeatPurchaseTrendPoint struct {
	BucketLabel      string  `json:"bucketLabel"`
	BucketStart      string  `json:"bucketStart"`
	BucketEnd        string  `json:"bucketEnd"`
	PaidBuyerCount   int64   `json:"paidBuyerCount"`
	RepeatBuyerCount int64   `json:"repeatBuyerCount"`
	RepeatRate       float64 `json:"repeatRate"`
	RepeatOrderCount int64   `json:"repeatOrderCount"`
	RepeatGmv        float64 `json:"repeatGmv"`
	AvgDaysToRepeat  float64 `json:"avgDaysToRepeat"`
}

// RepeatPurchaseDetailItem 复购分析详情行
type RepeatPurchaseDetailItem struct {
	MemberId            int64   `json:"memberId"`
	NicknameMasked      string  `json:"nicknameMasked"`
	MobileMasked        string  `json:"mobileMasked"`
	FirstValidPayTime   string  `json:"firstValidPayTime"`
	LatestRepeatPayTime string  `json:"latestRepeatPayTime"`
	RepeatOrderCount    int64   `json:"repeatOrderCount"`
	RepeatGmv           float64 `json:"repeatGmv"`
	LatestChannel       string  `json:"latestChannel"`
	LatestActivityType  string  `json:"latestActivityType"`
	LatestActivityId    int64   `json:"latestActivityId"`
	PlatformId          int64   `json:"platformId"`
	TenantId            int64   `json:"tenantId"`
	MerchantId          int64   `json:"merchantId"`
}

// QueryRepeatPurchaseAnalysisReq 查询复购分析请求
type QueryRepeatPurchaseAnalysisReq struct {
	ScopeType    string `form:"scopeType,optional" json:"scopeType"`
	PlatformId   int64  `form:"platformId,optional" json:"platformId"`
	TenantId     int64  `form:"tenantId,optional" json:"tenantId"`
	MerchantId   int64  `form:"merchantId,optional" json:"merchantId"`
	StartTime    string `form:"startTime,optional" json:"startTime"`
	EndTime      string `form:"endTime,optional" json:"endTime"`
	Channel      string `form:"channel,optional" json:"channel"`
	ActivityType string `form:"activityType,optional" json:"activityType"`
	ActivityId   int64  `form:"activityId,optional" json:"activityId"`
	Bucket       string `form:"bucket,optional" json:"bucket"`
	PageNum      int32  `form:"pageNum,default=1" json:"pageNum"`
	PageSize     int32  `form:"pageSize,default=20" json:"pageSize"`
}

// QueryRepeatPurchaseAnalysisData 复购分析数据
type QueryRepeatPurchaseAnalysisData struct {
	Overview          RepeatPurchaseOverview        `json:"overview"`
	Trends            []RepeatPurchaseTrendPoint    `json:"trends"`
	Details           []RepeatPurchaseDetailItem    `json:"details"`
	Total             int64                         `json:"total"`
	PageNum           int32                         `json:"pageNum"`
	PageSize          int32                         `json:"pageSize"`
	ActivityOptions   []OperateFunnelActivityOption `json:"activityOptions"`
	TrackingStartedAt string                        `json:"trackingStartedAt"`
	PartialMetrics    []string                      `json:"partialMetrics"`
	Bucket            string                        `json:"bucket"`
}

// QueryRepeatPurchaseAnalysisResp 查询复购分析响应
type QueryRepeatPurchaseAnalysisResp struct {
	Code    string                          `json:"code"`
	Message string                          `json:"message"`
	Data    QueryRepeatPurchaseAnalysisData `json:"data"`
	Success bool                            `json:"success"`
}

// ExportRepeatPurchaseAnalysisReq 导出复购分析请求
type ExportRepeatPurchaseAnalysisReq = QueryRepeatPurchaseAnalysisReq
