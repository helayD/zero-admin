package smsclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CardTemplateService_AddCardTemplate_FullMethodName          = "/smsclient.CardTemplateService/AddCardTemplate"
	CardTemplateService_UpdateCardTemplate_FullMethodName       = "/smsclient.CardTemplateService/UpdateCardTemplate"
	CardTemplateService_QueryCardTemplateList_FullMethodName    = "/smsclient.CardTemplateService/QueryCardTemplateList"
	CardTemplateService_QueryCardTemplateDetail_FullMethodName  = "/smsclient.CardTemplateService/QueryCardTemplateDetail"
	CardTemplateService_UpdateCardTemplateStatus_FullMethodName = "/smsclient.CardTemplateService/UpdateCardTemplateStatus"
	CardTemplateService_DeleteCardTemplate_FullMethodName       = "/smsclient.CardTemplateService/DeleteCardTemplate"
	CardTemplateService_CheckCardTemplateUsage_FullMethodName   = "/smsclient.CardTemplateService/CheckCardTemplateUsage"
)

// CardTemplateService 卡片模板服务接口（客户端）
type CardTemplateService interface {
	AddCardTemplate(ctx context.Context, in *AddCardTemplateReq, opts ...grpc.CallOption) (*AddCardTemplateResp, error)
	UpdateCardTemplate(ctx context.Context, in *UpdateCardTemplateReq, opts ...grpc.CallOption) (*UpdateCardTemplateResp, error)
	QueryCardTemplateList(ctx context.Context, in *QueryCardTemplateListReq, opts ...grpc.CallOption) (*QueryCardTemplateListResp, error)
	QueryCardTemplateDetail(ctx context.Context, in *QueryCardTemplateDetailReq, opts ...grpc.CallOption) (*QueryCardTemplateDetailResp, error)
	UpdateCardTemplateStatus(ctx context.Context, in *UpdateCardTemplateStatusReq, opts ...grpc.CallOption) (*UpdateCardTemplateStatusResp, error)
	DeleteCardTemplate(ctx context.Context, in *DeleteCardTemplateReq, opts ...grpc.CallOption) (*DeleteCardTemplateResp, error)
	CheckCardTemplateUsage(ctx context.Context, in *CheckCardTemplateUsageReq, opts ...grpc.CallOption) (*CheckCardTemplateUsageResp, error)
}

// CardTemplateServiceServer 卡片模板服务接口（服务端）
type CardTemplateServiceServer interface {
	AddCardTemplate(ctx context.Context, in *AddCardTemplateReq) (*AddCardTemplateResp, error)
	UpdateCardTemplate(ctx context.Context, in *UpdateCardTemplateReq) (*UpdateCardTemplateResp, error)
	QueryCardTemplateList(ctx context.Context, in *QueryCardTemplateListReq) (*QueryCardTemplateListResp, error)
	QueryCardTemplateDetail(ctx context.Context, in *QueryCardTemplateDetailReq) (*QueryCardTemplateDetailResp, error)
	UpdateCardTemplateStatus(ctx context.Context, in *UpdateCardTemplateStatusReq) (*UpdateCardTemplateStatusResp, error)
	DeleteCardTemplate(ctx context.Context, in *DeleteCardTemplateReq) (*DeleteCardTemplateResp, error)
	CheckCardTemplateUsage(ctx context.Context, in *CheckCardTemplateUsageReq) (*CheckCardTemplateUsageResp, error)
	mustEmbedUnimplementedCardTemplateServiceServer()
}

// UnimplementedCardTemplateServiceServer 未实现的卡片模板服务（用于前向兼容）
type UnimplementedCardTemplateServiceServer struct{}

func (UnimplementedCardTemplateServiceServer) AddCardTemplate(ctx context.Context, in *AddCardTemplateReq) (*AddCardTemplateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddCardTemplate not implemented")
}
func (UnimplementedCardTemplateServiceServer) UpdateCardTemplate(ctx context.Context, in *UpdateCardTemplateReq) (*UpdateCardTemplateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCardTemplate not implemented")
}
func (UnimplementedCardTemplateServiceServer) QueryCardTemplateList(ctx context.Context, in *QueryCardTemplateListReq) (*QueryCardTemplateListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryCardTemplateList not implemented")
}
func (UnimplementedCardTemplateServiceServer) QueryCardTemplateDetail(ctx context.Context, in *QueryCardTemplateDetailReq) (*QueryCardTemplateDetailResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryCardTemplateDetail not implemented")
}
func (UnimplementedCardTemplateServiceServer) UpdateCardTemplateStatus(ctx context.Context, in *UpdateCardTemplateStatusReq) (*UpdateCardTemplateStatusResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCardTemplateStatus not implemented")
}
func (UnimplementedCardTemplateServiceServer) DeleteCardTemplate(ctx context.Context, in *DeleteCardTemplateReq) (*DeleteCardTemplateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCardTemplate not implemented")
}
func (UnimplementedCardTemplateServiceServer) CheckCardTemplateUsage(ctx context.Context, in *CheckCardTemplateUsageReq) (*CheckCardTemplateUsageResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckCardTemplateUsage not implemented")
}
func (UnimplementedCardTemplateServiceServer) mustEmbedUnimplementedCardTemplateServiceServer() {}

// cardTemplateServiceImpl 卡片模板服务实现（用于客户端调用）
type cardTemplateServiceImpl struct {
	conn grpc.ClientConnInterface
}

// NewCardTemplateServiceClient 创建卡片模板服务客户端
func NewCardTemplateServiceClient(conn grpc.ClientConnInterface) CardTemplateService {
	return &cardTemplateServiceImpl{conn: conn}
}

func (c *cardTemplateServiceImpl) AddCardTemplate(ctx context.Context, in *AddCardTemplateReq, opts ...grpc.CallOption) (*AddCardTemplateResp, error) {
	out := new(AddCardTemplateResp)
	err := c.conn.Invoke(ctx, CardTemplateService_AddCardTemplate_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardTemplateServiceImpl) UpdateCardTemplate(ctx context.Context, in *UpdateCardTemplateReq, opts ...grpc.CallOption) (*UpdateCardTemplateResp, error) {
	out := new(UpdateCardTemplateResp)
	err := c.conn.Invoke(ctx, CardTemplateService_UpdateCardTemplate_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardTemplateServiceImpl) QueryCardTemplateList(ctx context.Context, in *QueryCardTemplateListReq, opts ...grpc.CallOption) (*QueryCardTemplateListResp, error) {
	out := new(QueryCardTemplateListResp)
	err := c.conn.Invoke(ctx, CardTemplateService_QueryCardTemplateList_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardTemplateServiceImpl) QueryCardTemplateDetail(ctx context.Context, in *QueryCardTemplateDetailReq, opts ...grpc.CallOption) (*QueryCardTemplateDetailResp, error) {
	out := new(QueryCardTemplateDetailResp)
	err := c.conn.Invoke(ctx, CardTemplateService_QueryCardTemplateDetail_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardTemplateServiceImpl) UpdateCardTemplateStatus(ctx context.Context, in *UpdateCardTemplateStatusReq, opts ...grpc.CallOption) (*UpdateCardTemplateStatusResp, error) {
	out := new(UpdateCardTemplateStatusResp)
	err := c.conn.Invoke(ctx, CardTemplateService_UpdateCardTemplateStatus_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardTemplateServiceImpl) DeleteCardTemplate(ctx context.Context, in *DeleteCardTemplateReq, opts ...grpc.CallOption) (*DeleteCardTemplateResp, error) {
	out := new(DeleteCardTemplateResp)
	err := c.conn.Invoke(ctx, CardTemplateService_DeleteCardTemplate_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardTemplateServiceImpl) CheckCardTemplateUsage(ctx context.Context, in *CheckCardTemplateUsageReq, opts ...grpc.CallOption) (*CheckCardTemplateUsageResp, error) {
	out := new(CheckCardTemplateUsageResp)
	err := c.conn.Invoke(ctx, CardTemplateService_CheckCardTemplateUsage_FullMethodName, in, out, opts...)
	return out, err
}

// RegisterCardTemplateServiceServer 注册卡片模板服务到 gRPC 服务器
func RegisterCardTemplateServiceServer(s *grpc.Server, srv CardTemplateServiceServer) {
	s.RegisterService(&CardTemplateService_ServiceDesc, srv)
}

var CardTemplateService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "smsclient.CardTemplateService",
	HandlerType: (*CardTemplateServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "AddCardTemplate", Handler: _CardTemplateService_AddCardTemplate_Handler},
		{MethodName: "UpdateCardTemplate", Handler: _CardTemplateService_UpdateCardTemplate_Handler},
		{MethodName: "QueryCardTemplateList", Handler: _CardTemplateService_QueryCardTemplateList_Handler},
		{MethodName: "QueryCardTemplateDetail", Handler: _CardTemplateService_QueryCardTemplateDetail_Handler},
		{MethodName: "UpdateCardTemplateStatus", Handler: _CardTemplateService_UpdateCardTemplateStatus_Handler},
		{MethodName: "DeleteCardTemplate", Handler: _CardTemplateService_DeleteCardTemplate_Handler},
		{MethodName: "CheckCardTemplateUsage", Handler: _CardTemplateService_CheckCardTemplateUsage_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/sms/card_template.proto",
}

func _CardTemplateService_AddCardTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddCardTemplateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).AddCardTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_AddCardTemplate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).AddCardTemplate(ctx, req.(*AddCardTemplateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardTemplateService_UpdateCardTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCardTemplateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).UpdateCardTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_UpdateCardTemplate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).UpdateCardTemplate(ctx, req.(*UpdateCardTemplateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardTemplateService_QueryCardTemplateList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCardTemplateListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).QueryCardTemplateList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_QueryCardTemplateList_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).QueryCardTemplateList(ctx, req.(*QueryCardTemplateListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardTemplateService_QueryCardTemplateDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCardTemplateDetailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).QueryCardTemplateDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_QueryCardTemplateDetail_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).QueryCardTemplateDetail(ctx, req.(*QueryCardTemplateDetailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardTemplateService_UpdateCardTemplateStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCardTemplateStatusReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).UpdateCardTemplateStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_UpdateCardTemplateStatus_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).UpdateCardTemplateStatus(ctx, req.(*UpdateCardTemplateStatusReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardTemplateService_DeleteCardTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCardTemplateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).DeleteCardTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_DeleteCardTemplate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).DeleteCardTemplate(ctx, req.(*DeleteCardTemplateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardTemplateService_CheckCardTemplateUsage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckCardTemplateUsageReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardTemplateServiceServer).CheckCardTemplateUsage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardTemplateService_CheckCardTemplateUsage_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardTemplateServiceServer).CheckCardTemplateUsage(ctx, req.(*CheckCardTemplateUsageReq))
	}
	return interceptor(ctx, in, info, handler)
}
