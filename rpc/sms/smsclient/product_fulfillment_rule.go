package smsclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// ProductFulfillmentRuleService 发卡规则服务接口（客户端）
type ProductFulfillmentRuleService interface {
	AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq, opts ...grpc.CallOption) (*AddProductFulfillmentRuleResp, error)
	UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleResp, error)
	QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleListResp, error)
	QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleDetailResp, error)
	UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleStatusResp, error)
	DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq, opts ...grpc.CallOption) (*DeleteProductFulfillmentRuleResp, error)
	CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq, opts ...grpc.CallOption) (*CheckProductFulfillmentRuleBindingResp, error)
}

// ProductFulfillmentRuleServiceServer 发卡规则服务接口（服务端）
type ProductFulfillmentRuleServiceServer interface {
	AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq) (*AddProductFulfillmentRuleResp, error)
	UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq) (*UpdateProductFulfillmentRuleResp, error)
	QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq) (*QueryProductFulfillmentRuleListResp, error)
	QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq) (*QueryProductFulfillmentRuleDetailResp, error)
	UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq) (*UpdateProductFulfillmentRuleStatusResp, error)
	DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq) (*DeleteProductFulfillmentRuleResp, error)
	CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq) (*CheckProductFulfillmentRuleBindingResp, error)
	mustEmbedUnimplementedProductFulfillmentRuleServiceServer()
}

// UnimplementedProductFulfillmentRuleServiceServer 未实现的发卡规则服务（用于前向兼容）
type UnimplementedProductFulfillmentRuleServiceServer struct{}

func (UnimplementedProductFulfillmentRuleServiceServer) AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq) (*AddProductFulfillmentRuleResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddProductFulfillmentRule not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq) (*UpdateProductFulfillmentRuleResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProductFulfillmentRule not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq) (*QueryProductFulfillmentRuleListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryProductFulfillmentRuleList not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq) (*QueryProductFulfillmentRuleDetailResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryProductFulfillmentRuleDetail not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq) (*UpdateProductFulfillmentRuleStatusResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProductFulfillmentRuleStatus not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq) (*DeleteProductFulfillmentRuleResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteProductFulfillmentRule not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq) (*CheckProductFulfillmentRuleBindingResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckProductFulfillmentRuleBinding not implemented")
}
func (UnimplementedProductFulfillmentRuleServiceServer) mustEmbedUnimplementedProductFulfillmentRuleServiceServer() {}

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
func RegisterProductFulfillmentRuleServiceServer(s *grpc.Server, srv ProductFulfillmentRuleServiceServer) {
	s.RegisterService(&ProductFulfillmentRuleService_ServiceDesc, srv)
}

var ProductFulfillmentRuleService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "smsclient.ProductFulfillmentRuleService",
	HandlerType: (*ProductFulfillmentRuleServiceServer)(nil),
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
		return srv.(ProductFulfillmentRuleServiceServer).AddProductFulfillmentRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_AddProductFulfillmentRule_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).AddProductFulfillmentRule(ctx, req.(*AddProductFulfillmentRuleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_UpdateProductFulfillmentRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProductFulfillmentRuleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleServiceServer).UpdateProductFulfillmentRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_UpdateProductFulfillmentRule_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).UpdateProductFulfillmentRule(ctx, req.(*UpdateProductFulfillmentRuleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProductFulfillmentRuleListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleServiceServer).QueryProductFulfillmentRuleList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_QueryProductFulfillmentRuleList_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).QueryProductFulfillmentRuleList(ctx, req.(*QueryProductFulfillmentRuleListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProductFulfillmentRuleDetailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleServiceServer).QueryProductFulfillmentRuleDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_QueryProductFulfillmentRuleDetail_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).QueryProductFulfillmentRuleDetail(ctx, req.(*QueryProductFulfillmentRuleDetailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProductFulfillmentRuleStatusReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleServiceServer).UpdateProductFulfillmentRuleStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_UpdateProductFulfillmentRuleStatus_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).UpdateProductFulfillmentRuleStatus(ctx, req.(*UpdateProductFulfillmentRuleStatusReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_DeleteProductFulfillmentRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteProductFulfillmentRuleReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleServiceServer).DeleteProductFulfillmentRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_DeleteProductFulfillmentRule_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).DeleteProductFulfillmentRule(ctx, req.(*DeleteProductFulfillmentRuleReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckProductFulfillmentRuleBindingReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductFulfillmentRuleServiceServer).CheckProductFulfillmentRuleBinding(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ProductFulfillmentRuleService_CheckProductFulfillmentRuleBinding_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductFulfillmentRuleServiceServer).CheckProductFulfillmentRuleBinding(ctx, req.(*CheckProductFulfillmentRuleBindingReq))
	}
	return interceptor(ctx, in, info, handler)
}
