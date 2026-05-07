package smsclient

import (
	"context"

	"google.golang.org/grpc"
)

// ProductFulfillmentRuleData 发卡规则数据
type ProductFulfillmentRuleData struct {
	Id                  int64  `json:"id"`
	RuleName            string `json:"ruleName"`
	RuleStatus          int32  `json:"ruleStatus"`
	CardTemplateId      int64  `json:"cardTemplateId"`
	CardTemplateName    string `json:"cardTemplateName"`
	ExpireDays          int32  `json:"expireDays"`
	Transferable        int32  `json:"transferable"`
	TransferLimit       int32  `json:"transferLimit"`
	ClaimCondition      string `json:"claimCondition"`
	RedemptionCondition string `json:"redemptionCondition"`
	RefundPolicy        string `json:"refundPolicy"`
	RefundPolicyText    string `json:"refundPolicyText"`
	PlatformId          int64  `json:"platformId"`
	TenantId            int64  `json:"tenantId"`
	MerchantId          int64  `json:"merchantId"`
	CreateBy            string `json:"createBy"`
	UpdateBy            string `json:"updateBy"`
	CreateTime          string `json:"createTime"`
	UpdateTime          string `json:"updateTime"`
	BindingCount        int32  `json:"bindingCount"`
}

// AddProductFulfillmentRuleReq 添加发卡规则请求
type AddProductFulfillmentRuleReq struct {
	RuleName            string           `json:"ruleName"`
	CardTemplateId      int64            `json:"cardTemplateId"`
	ExpireDays          int32            `json:"expireDays"`
	Transferable        int32            `json:"transferable"`
	TransferLimit       int32            `json:"transferLimit"`
	ClaimCondition      string           `json:"claimCondition"`
	RedemptionCondition string           `json:"redemptionCondition"`
	RefundPolicy        string           `json:"refundPolicy"`
	Scope               *GovernanceScope `json:"scope"`
	OperatorType        string           `json:"operatorType"`
}

// AddProductFulfillmentRuleResp 添加发卡规则响应
type AddProductFulfillmentRuleResp struct {
	Rule *ProductFulfillmentRuleData `json:"rule"`
}

// UpdateProductFulfillmentRuleReq 更新发卡规则请求
type UpdateProductFulfillmentRuleReq struct {
	Id                  int64            `json:"id"`
	RuleName            string           `json:"ruleName"`
	CardTemplateId      int64            `json:"cardTemplateId"`
	ExpireDays          int32            `json:"expireDays"`
	Transferable        int32            `json:"transferable"`
	TransferLimit       int32            `json:"transferLimit"`
	ClaimCondition      string           `json:"claimCondition"`
	RedemptionCondition string           `json:"redemptionCondition"`
	RefundPolicy        string           `json:"refundPolicy"`
	Scope               *GovernanceScope `json:"scope"`
	OperatorType        string           `json:"operatorType"`
}

// UpdateProductFulfillmentRuleResp 更新发卡规则响应
type UpdateProductFulfillmentRuleResp struct {
	Rule *ProductFulfillmentRuleData `json:"rule"`
}

// QueryProductFulfillmentRuleListReq 查询发卡规则列表请求
type QueryProductFulfillmentRuleListReq struct {
	RuleName       string           `json:"ruleName"`
	CardTemplateId int64            `json:"cardTemplateId"`
	RuleStatus     int32            `json:"ruleStatus"`
	Page           int32            `json:"page"`
	PageSize       int32            `json:"pageSize"`
	Scope          *GovernanceScope `json:"scope"`
}

// QueryProductFulfillmentRuleListResp 查询发卡规则列表响应
type QueryProductFulfillmentRuleListResp struct {
	List  []*ProductFulfillmentRuleData `json:"list"`
	Total int64                         `json:"total"`
}

// QueryProductFulfillmentRuleDetailReq 查询发卡规则详情请求
type QueryProductFulfillmentRuleDetailReq struct {
	Id    int64            `json:"id"`
	Scope *GovernanceScope `json:"scope"`
}

// QueryProductFulfillmentRuleDetailResp 查询发卡规则详情响应
type QueryProductFulfillmentRuleDetailResp struct {
	Rule *ProductFulfillmentRuleData `json:"rule"`
}

// UpdateProductFulfillmentRuleStatusReq 更新发卡规则状态请求
type UpdateProductFulfillmentRuleStatusReq struct {
	Id           int64            `json:"id"`
	RuleStatus   int32            `json:"ruleStatus"`
	Scope        *GovernanceScope `json:"scope"`
	OperatorType string           `json:"operatorType"`
}

// UpdateProductFulfillmentRuleStatusResp 更新发卡规则状态响应
type UpdateProductFulfillmentRuleStatusResp struct {
	Rule *ProductFulfillmentRuleData `json:"rule"`
}

// DeleteProductFulfillmentRuleReq 删除发卡规则请求
type DeleteProductFulfillmentRuleReq struct {
	Id           int64            `json:"id"`
	Scope        *GovernanceScope `json:"scope"`
	OperatorType string           `json:"operatorType"`
}

// DeleteProductFulfillmentRuleResp 删除发卡规则响应
type DeleteProductFulfillmentRuleResp struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// CheckProductFulfillmentRuleBindingReq 检查发卡规则绑定请求
type CheckProductFulfillmentRuleBindingReq struct {
	Id    int64            `json:"id"`
	Scope *GovernanceScope `json:"scope"`
}

// CheckProductFulfillmentRuleBindingResp 检查发卡规则绑定响应
type CheckProductFulfillmentRuleBindingResp struct {
	BindingCount        int32    `json:"bindingCount"`
	BindingProductNames []string `json:"bindingProductNames"`
	CanDisable          bool     `json:"canDisable"`
	CanDelete           bool     `json:"canDelete"`
	Message             string   `json:"message"`
}

// EnsureOrderPurchaseCardInstanceReq 订单购买型卡片资产创建请求
type EnsureOrderPurchaseCardInstanceReq struct {
	OrderId           int64  `json:"orderId"`
	OrderItemId       int64  `json:"orderItemId"`
	ProductId         int64  `json:"productId"`
	SkuId             int64  `json:"skuId"`
	MemberId          int64  `json:"memberId"`
	FulfillmentRuleId int64  `json:"fulfillmentRuleId"`
	PlatformId        int64  `json:"platformId"`
	TenantId          int64  `json:"tenantId"`
	MerchantId        int64  `json:"merchantId"`
	RequestId         string `json:"requestId"`
	TraceId           string `json:"traceId"`
	OperatorType      string `json:"operatorType"`
}

// EnsureOrderPurchaseCardInstanceResp 订单购买型卡片资产创建响应
type EnsureOrderPurchaseCardInstanceResp struct {
	Asset *CardInstanceData `json:"asset"`
}

// ProductFulfillmentRuleService 发卡规则服务接口
type ProductFulfillmentRuleService interface {
	AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq, opts ...grpc.CallOption) (*AddProductFulfillmentRuleResp, error)
	UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleResp, error)
	QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleListResp, error)
	QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleDetailResp, error)
	UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleStatusResp, error)
	DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq, opts ...grpc.CallOption) (*DeleteProductFulfillmentRuleResp, error)
	CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq, opts ...grpc.CallOption) (*CheckProductFulfillmentRuleBindingResp, error)
}

// productFulfillmentRuleServiceImpl 发卡规则服务实现（用于客户端调用）
type productFulfillmentRuleServiceImpl struct {
	conn grpc.ClientConnInterface
}

// NewProductFulfillmentRuleServiceClient 创建发卡规则服务客户端
func NewProductFulfillmentRuleServiceClient(conn grpc.ClientConnInterface) ProductFulfillmentRuleService {
	return &productFulfillmentRuleServiceImpl{conn: conn}
}

func (c *productFulfillmentRuleServiceImpl) AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq, opts ...grpc.CallOption) (*AddProductFulfillmentRuleResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

func (c *productFulfillmentRuleServiceImpl) UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

func (c *productFulfillmentRuleServiceImpl) QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleListResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

func (c *productFulfillmentRuleServiceImpl) QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleDetailResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

func (c *productFulfillmentRuleServiceImpl) UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleStatusResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

func (c *productFulfillmentRuleServiceImpl) DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq, opts ...grpc.CallOption) (*DeleteProductFulfillmentRuleResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

func (c *productFulfillmentRuleServiceImpl) CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq, opts ...grpc.CallOption) (*CheckProductFulfillmentRuleBindingResp, error) {
	// TODO: 实现 gRPC 调用
	return nil, nil
}

// RegisterProductFulfillmentRuleServiceServer 注册发卡规则服务到 gRPC 服务器
func RegisterProductFulfillmentRuleServiceServer(s *grpc.Server, srv ProductFulfillmentRuleService) {
	// TODO: 实现服务注册
}
