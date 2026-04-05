package sysclient

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	ChannelIntegrationTemplateService_CreateChannelIntegrationTemplate_FullMethodName       = "/sysclient.ChannelIntegrationTemplateService/CreateChannelIntegrationTemplate"
	ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplate_FullMethodName       = "/sysclient.ChannelIntegrationTemplateService/UpdateChannelIntegrationTemplate"
	ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplateStatus_FullMethodName = "/sysclient.ChannelIntegrationTemplateService/UpdateChannelIntegrationTemplateStatus"
	ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateDetail_FullMethodName  = "/sysclient.ChannelIntegrationTemplateService/QueryChannelIntegrationTemplateDetail"
	ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateList_FullMethodName    = "/sysclient.ChannelIntegrationTemplateService/QueryChannelIntegrationTemplateList"
)

type ChannelIntegrationTemplateServiceClient interface {
	CreateChannelIntegrationTemplate(ctx context.Context, in *CreateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*CreateChannelIntegrationTemplateResp, error)
	UpdateChannelIntegrationTemplate(ctx context.Context, in *UpdateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateResp, error)
	UpdateChannelIntegrationTemplateStatus(ctx context.Context, in *UpdateChannelIntegrationTemplateStatusReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateStatusResp, error)
	QueryChannelIntegrationTemplateDetail(ctx context.Context, in *QueryChannelIntegrationTemplateDetailReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateDetailResp, error)
	QueryChannelIntegrationTemplateList(ctx context.Context, in *QueryChannelIntegrationTemplateListReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateListResp, error)
}

type channelIntegrationTemplateServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewChannelIntegrationTemplateServiceClient(cc grpc.ClientConnInterface) ChannelIntegrationTemplateServiceClient {
	return &channelIntegrationTemplateServiceClient{cc}
}

func (c *channelIntegrationTemplateServiceClient) CreateChannelIntegrationTemplate(ctx context.Context, in *CreateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*CreateChannelIntegrationTemplateResp, error) {
	out := new(CreateChannelIntegrationTemplateResp)
	err := c.cc.Invoke(ctx, ChannelIntegrationTemplateService_CreateChannelIntegrationTemplate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelIntegrationTemplateServiceClient) UpdateChannelIntegrationTemplate(ctx context.Context, in *UpdateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateResp, error) {
	out := new(UpdateChannelIntegrationTemplateResp)
	err := c.cc.Invoke(ctx, ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelIntegrationTemplateServiceClient) UpdateChannelIntegrationTemplateStatus(ctx context.Context, in *UpdateChannelIntegrationTemplateStatusReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateStatusResp, error) {
	out := new(UpdateChannelIntegrationTemplateStatusResp)
	err := c.cc.Invoke(ctx, ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplateStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelIntegrationTemplateServiceClient) QueryChannelIntegrationTemplateDetail(ctx context.Context, in *QueryChannelIntegrationTemplateDetailReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateDetailResp, error) {
	out := new(QueryChannelIntegrationTemplateDetailResp)
	err := c.cc.Invoke(ctx, ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateDetail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *channelIntegrationTemplateServiceClient) QueryChannelIntegrationTemplateList(ctx context.Context, in *QueryChannelIntegrationTemplateListReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateListResp, error) {
	out := new(QueryChannelIntegrationTemplateListResp)
	err := c.cc.Invoke(ctx, ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateList_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ChannelIntegrationTemplateServiceServer interface {
	CreateChannelIntegrationTemplate(context.Context, *CreateChannelIntegrationTemplateReq) (*CreateChannelIntegrationTemplateResp, error)
	UpdateChannelIntegrationTemplate(context.Context, *UpdateChannelIntegrationTemplateReq) (*UpdateChannelIntegrationTemplateResp, error)
	UpdateChannelIntegrationTemplateStatus(context.Context, *UpdateChannelIntegrationTemplateStatusReq) (*UpdateChannelIntegrationTemplateStatusResp, error)
	QueryChannelIntegrationTemplateDetail(context.Context, *QueryChannelIntegrationTemplateDetailReq) (*QueryChannelIntegrationTemplateDetailResp, error)
	QueryChannelIntegrationTemplateList(context.Context, *QueryChannelIntegrationTemplateListReq) (*QueryChannelIntegrationTemplateListResp, error)
	mustEmbedUnimplementedChannelIntegrationTemplateServiceServer()
}

type UnimplementedChannelIntegrationTemplateServiceServer struct{}

func (UnimplementedChannelIntegrationTemplateServiceServer) CreateChannelIntegrationTemplate(context.Context, *CreateChannelIntegrationTemplateReq) (*CreateChannelIntegrationTemplateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChannelIntegrationTemplate not implemented")
}
func (UnimplementedChannelIntegrationTemplateServiceServer) UpdateChannelIntegrationTemplate(context.Context, *UpdateChannelIntegrationTemplateReq) (*UpdateChannelIntegrationTemplateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateChannelIntegrationTemplate not implemented")
}
func (UnimplementedChannelIntegrationTemplateServiceServer) UpdateChannelIntegrationTemplateStatus(context.Context, *UpdateChannelIntegrationTemplateStatusReq) (*UpdateChannelIntegrationTemplateStatusResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateChannelIntegrationTemplateStatus not implemented")
}
func (UnimplementedChannelIntegrationTemplateServiceServer) QueryChannelIntegrationTemplateDetail(context.Context, *QueryChannelIntegrationTemplateDetailReq) (*QueryChannelIntegrationTemplateDetailResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryChannelIntegrationTemplateDetail not implemented")
}
func (UnimplementedChannelIntegrationTemplateServiceServer) QueryChannelIntegrationTemplateList(context.Context, *QueryChannelIntegrationTemplateListReq) (*QueryChannelIntegrationTemplateListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryChannelIntegrationTemplateList not implemented")
}
func (UnimplementedChannelIntegrationTemplateServiceServer) mustEmbedUnimplementedChannelIntegrationTemplateServiceServer() {
}

type UnsafeChannelIntegrationTemplateServiceServer interface {
	mustEmbedUnimplementedChannelIntegrationTemplateServiceServer()
}

func RegisterChannelIntegrationTemplateServiceServer(s grpc.ServiceRegistrar, srv ChannelIntegrationTemplateServiceServer) {
	s.RegisterService(&ChannelIntegrationTemplateService_ServiceDesc, srv)
}

func _ChannelIntegrationTemplateService_CreateChannelIntegrationTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateChannelIntegrationTemplateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelIntegrationTemplateServiceServer).CreateChannelIntegrationTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ChannelIntegrationTemplateService_CreateChannelIntegrationTemplate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelIntegrationTemplateServiceServer).CreateChannelIntegrationTemplate(ctx, req.(*CreateChannelIntegrationTemplateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateChannelIntegrationTemplateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelIntegrationTemplateServiceServer).UpdateChannelIntegrationTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplate_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelIntegrationTemplateServiceServer).UpdateChannelIntegrationTemplate(ctx, req.(*UpdateChannelIntegrationTemplateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplateStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateChannelIntegrationTemplateStatusReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelIntegrationTemplateServiceServer).UpdateChannelIntegrationTemplateStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplateStatus_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelIntegrationTemplateServiceServer).UpdateChannelIntegrationTemplateStatus(ctx, req.(*UpdateChannelIntegrationTemplateStatusReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryChannelIntegrationTemplateDetailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelIntegrationTemplateServiceServer).QueryChannelIntegrationTemplateDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateDetail_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelIntegrationTemplateServiceServer).QueryChannelIntegrationTemplateDetail(ctx, req.(*QueryChannelIntegrationTemplateDetailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryChannelIntegrationTemplateListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChannelIntegrationTemplateServiceServer).QueryChannelIntegrationTemplateList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateList_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChannelIntegrationTemplateServiceServer).QueryChannelIntegrationTemplateList(ctx, req.(*QueryChannelIntegrationTemplateListReq))
	}
	return interceptor(ctx, in, info, handler)
}

var ChannelIntegrationTemplateService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sysclient.ChannelIntegrationTemplateService",
	HandlerType: (*ChannelIntegrationTemplateServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateChannelIntegrationTemplate", Handler: _ChannelIntegrationTemplateService_CreateChannelIntegrationTemplate_Handler},
		{MethodName: "UpdateChannelIntegrationTemplate", Handler: _ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplate_Handler},
		{MethodName: "UpdateChannelIntegrationTemplateStatus", Handler: _ChannelIntegrationTemplateService_UpdateChannelIntegrationTemplateStatus_Handler},
		{MethodName: "QueryChannelIntegrationTemplateDetail", Handler: _ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateDetail_Handler},
		{MethodName: "QueryChannelIntegrationTemplateList", Handler: _ChannelIntegrationTemplateService_QueryChannelIntegrationTemplateList_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/sys/sys.proto",
}
