// Hand-written gRPC bindings for MemberAuthService.
// 与 sysclient.ChannelIntegrationTemplateService_grpc.pb.go 同样模式，
// 不依赖 protoc 生成的 raw descriptor，靠 ServiceDesc + struct tag 反射工作。

package umsclient

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	MemberAuthService_SendSmsCode_FullMethodName = "/umsclient.MemberAuthService/SendSmsCode"
	MemberAuthService_LoginByCode_FullMethodName = "/umsclient.MemberAuthService/LoginByCode"
)

// MemberAuthServiceClient 客户端接口
type MemberAuthServiceClient interface {
	// SendSmsCode 发送短信验证码（手机号+场景），写入 Redis 等待校验
	SendSmsCode(ctx context.Context, in *SendSmsCodeReq, opts ...grpc.CallOption) (*SendSmsCodeResp, error)
	// LoginByCode 手机号+验证码合并接口（已注册→直接登录；未注册→自动建号并登录）
	LoginByCode(ctx context.Context, in *LoginByCodeReq, opts ...grpc.CallOption) (*LoginByCodeResp, error)
}

type memberAuthServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMemberAuthServiceClient(cc grpc.ClientConnInterface) MemberAuthServiceClient {
	return &memberAuthServiceClient{cc}
}

func (c *memberAuthServiceClient) SendSmsCode(ctx context.Context, in *SendSmsCodeReq, opts ...grpc.CallOption) (*SendSmsCodeResp, error) {
	out := new(SendSmsCodeResp)
	err := c.cc.Invoke(ctx, MemberAuthService_SendSmsCode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *memberAuthServiceClient) LoginByCode(ctx context.Context, in *LoginByCodeReq, opts ...grpc.CallOption) (*LoginByCodeResp, error) {
	out := new(LoginByCodeResp)
	err := c.cc.Invoke(ctx, MemberAuthService_LoginByCode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MemberAuthServiceServer 服务端接口
type MemberAuthServiceServer interface {
	SendSmsCode(context.Context, *SendSmsCodeReq) (*SendSmsCodeResp, error)
	LoginByCode(context.Context, *LoginByCodeReq) (*LoginByCodeResp, error)
	mustEmbedUnimplementedMemberAuthServiceServer()
}

type UnimplementedMemberAuthServiceServer struct{}

func (UnimplementedMemberAuthServiceServer) SendSmsCode(context.Context, *SendSmsCodeReq) (*SendSmsCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSmsCode not implemented")
}
func (UnimplementedMemberAuthServiceServer) LoginByCode(context.Context, *LoginByCodeReq) (*LoginByCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoginByCode not implemented")
}
func (UnimplementedMemberAuthServiceServer) mustEmbedUnimplementedMemberAuthServiceServer() {}

type UnsafeMemberAuthServiceServer interface {
	mustEmbedUnimplementedMemberAuthServiceServer()
}

func RegisterMemberAuthServiceServer(s grpc.ServiceRegistrar, srv MemberAuthServiceServer) {
	s.RegisterService(&MemberAuthService_ServiceDesc, srv)
}

func _MemberAuthService_SendSmsCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendSmsCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemberAuthServiceServer).SendSmsCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: MemberAuthService_SendSmsCode_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemberAuthServiceServer).SendSmsCode(ctx, req.(*SendSmsCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _MemberAuthService_LoginByCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginByCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MemberAuthServiceServer).LoginByCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: MemberAuthService_LoginByCode_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MemberAuthServiceServer).LoginByCode(ctx, req.(*LoginByCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

// MemberAuthService_ServiceDesc grpc 服务注册描述符
var MemberAuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "umsclient.MemberAuthService",
	HandlerType: (*MemberAuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "SendSmsCode", Handler: _MemberAuthService_SendSmsCode_Handler},
		{MethodName: "LoginByCode", Handler: _MemberAuthService_LoginByCode_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/ums/ums.proto",
}
