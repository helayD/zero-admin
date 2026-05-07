package smsclient

import (
	"context"

	"google.golang.org/grpc"
)

const (
	ProductFulfillmentRuleService_AddProductFulfillmentRule_FullMethodName          = "/smsclient.ProductFulfillmentRuleService/AddProductFulfillmentRule"
	ProductFulfillmentRuleService_UpdateProductFulfillmentRule_FullMethodName       = "/smsclient.ProductFulfillmentRuleService/UpdateProductFulfillmentRule"
	ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_FullMethodName    = "/smsclient.ProductFulfillmentRuleService/QueryProductFulfillmentRuleList"
	ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_FullMethodName  = "/smsclient.ProductFulfillmentRuleService/QueryProductFulfillmentRuleDetail"
	ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_FullMethodName = "/smsclient.ProductFulfillmentRuleService/UpdateProductFulfillmentRuleStatus"
	ProductFulfillmentRuleService_DeleteProductFulfillmentRule_FullMethodName       = "/smsclient.ProductFulfillmentRuleService/DeleteProductFulfillmentRule"
	ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_FullMethodName = "/smsclient.ProductFulfillmentRuleService/CheckProductFulfillmentRuleBinding"
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
	out := new(AddProductFulfillmentRuleResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_AddProductFulfillmentRule_FullMethodName, in, out, opts...)
	return out, err
}

func (c *productFulfillmentRuleServiceImpl) UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleResp, error) {
	out := new(UpdateProductFulfillmentRuleResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_UpdateProductFulfillmentRule_FullMethodName, in, out, opts...)
	return out, err
}

func (c *productFulfillmentRuleServiceImpl) QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleListResp, error) {
	out := new(QueryProductFulfillmentRuleListResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_FullMethodName, in, out, opts...)
	return out, err
}

func (c *productFulfillmentRuleServiceImpl) QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleDetailResp, error) {
	out := new(QueryProductFulfillmentRuleDetailResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_FullMethodName, in, out, opts...)
	return out, err
}

func (c *productFulfillmentRuleServiceImpl) UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleStatusResp, error) {
	out := new(UpdateProductFulfillmentRuleStatusResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_FullMethodName, in, out, opts...)
	return out, err
}

func (c *productFulfillmentRuleServiceImpl) DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq, opts ...grpc.CallOption) (*DeleteProductFulfillmentRuleResp, error) {
	out := new(DeleteProductFulfillmentRuleResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_DeleteProductFulfillmentRule_FullMethodName, in, out, opts...)
	return out, err
}

func (c *productFulfillmentRuleServiceImpl) CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq, opts ...grpc.CallOption) (*CheckProductFulfillmentRuleBindingResp, error) {
	out := new(CheckProductFulfillmentRuleBindingResp)
	err := c.conn.Invoke(ctx, ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_FullMethodName, in, out, opts...)
	return out, err
}

// RegisterProductFulfillmentRuleServiceServer 注册发卡规则服务到 gRPC 服务器
func RegisterProductFulfillmentRuleServiceServer(s *grpc.Server, srv ProductFulfillmentRuleService) {
	s.RegisterService(&ProductFulfillmentRuleService_ServiceDesc, srv)
}

var ProductFulfillmentRuleService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "smsclient.ProductFulfillmentRuleService",
	HandlerType: (*ProductFulfillmentRuleService)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "AddProductFulfillmentRule", Handler: _ProductFulfillmentRuleService_AddProductFulfillmentRule_Handler},
		{MethodName: "UpdateProductFulfillmentRule", Handler: _ProductFulfillmentRuleService_UpdateProductFulfillmentRule_Handler},
		{MethodName: "QueryProductFulfillmentRuleList", Handler: _ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_Handler},
		{MethodName: "QueryProductFulfillmentRuleDetail", Handler: _ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_Handler},
		{MethodName: "UpdateProductFulfillmentRuleStatus", Handler: _ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_Handler},
		{MethodName: "DeleteProductFulfillmentRule", Handler: _ProductFulfillmentRuleService_DeleteProductFulfillmentRule_Handler},
		{MethodName: "CheckProductFulfillmentRuleBinding", Handler: _ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/sms/product_fulfillment_rule.proto",
}

func _ProductFulfillmentRuleService_AddProductFulfillmentRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddProductFulfillmentRuleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).AddProductFulfillmentRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_AddProductFulfillmentRule_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).AddProductFulfillmentRule(ctx, req.(*AddProductFulfillmentRuleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_UpdateProductFulfillmentRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProductFulfillmentRuleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).UpdateProductFulfillmentRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_UpdateProductFulfillmentRule_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).UpdateProductFulfillmentRule(ctx, req.(*UpdateProductFulfillmentRuleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProductFulfillmentRuleListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).QueryProductFulfillmentRuleList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).QueryProductFulfillmentRuleList(ctx, req.(*QueryProductFulfillmentRuleListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProductFulfillmentRuleDetailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).QueryProductFulfillmentRuleDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).QueryProductFulfillmentRuleDetail(ctx, req.(*QueryProductFulfillmentRuleDetailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProductFulfillmentRuleStatusReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).UpdateProductFulfillmentRuleStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).UpdateProductFulfillmentRuleStatus(ctx, req.(*UpdateProductFulfillmentRuleStatusReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_DeleteProductFulfillmentRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteProductFulfillmentRuleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).DeleteProductFulfillmentRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_DeleteProductFulfillmentRule_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).DeleteProductFulfillmentRule(ctx, req.(*DeleteProductFulfillmentRuleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckProductFulfillmentRuleBindingReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleService).CheckProductFulfillmentRuleBinding(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleService).CheckProductFulfillmentRuleBinding(ctx, req.(*CheckProductFulfillmentRuleBindingReq))
	}
	return interceptor(ctx, in, info, handler)
}
