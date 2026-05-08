package smsclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CardClaimTokenService_GenerateClaimToken_FullMethodName = "/smsclient.CardClaimTokenService/GenerateClaimToken"
	CardClaimTokenService_ValidateClaimToken_FullMethodName = "/smsclient.CardClaimTokenService/ValidateClaimToken"
	CardClaimTokenService_ConsumeClaimToken_FullMethodName  = "/smsclient.CardClaimTokenService/ConsumeClaimToken"
	CardClaimTokenService_RevokeClaimToken_FullMethodName   = "/smsclient.CardClaimTokenService/RevokeClaimToken"
)

type ClaimTokenData struct {
	Id             int64  `json:"id"`
	Token          string `json:"token"`
	CardInstanceId int64  `json:"cardInstanceId"`
	IssuerId       int64  `json:"issuerId"`
	IssuerType     string `json:"issuerType"`
	ExpireAt       string `json:"expireAt"`
	MaxClaims      int32  `json:"maxClaims"`
	ClaimedCount   int32  `json:"claimedCount"`
	Status         string `json:"status"`
	ClaimedBy      int64  `json:"claimedBy"`
	ClaimedAt      string `json:"claimedAt"`
	PlatformId     int64  `json:"platformId"`
	TenantId       int64  `json:"tenantId"`
	MerchantId     int64  `json:"merchantId"`
	CreateTime     string `json:"createTime"`
	UpdateTime     string `json:"updateTime"`
}

type GenerateClaimTokenReq struct {
	CardInstanceId int64  `json:"cardInstanceId"`
	IssuerId       int64  `json:"issuerId"`
	ExpireHours    int32  `json:"expireHours"`
	MaxClaims      int32  `json:"maxClaims"`
	PlatformId     int64  `json:"platformId"`
	TenantId       int64  `json:"tenantId"`
	MerchantId     int64  `json:"merchantId"`
	TraceId        string `json:"traceId"`
	RequestId      string `json:"requestId"`
}

type GenerateClaimTokenResp struct {
	Token *ClaimTokenData `json:"token"`
}

type ValidateClaimTokenReq struct {
	Token string `json:"token"`
}

type ValidateClaimTokenResp struct {
	Valid          bool   `json:"valid"`
	Token          *ClaimTokenData `json:"token"`
	FailureReason  string `json:"failureReason"`
}

type ConsumeClaimTokenReq struct {
	Token      string `json:"token"`
	ClaimedBy  int64  `json:"claimedBy"`
	TraceId    string `json:"traceId"`
	RequestId  string `json:"requestId"`
}

type ConsumeClaimTokenResp struct {
	Success        bool            `json:"success"`
	Token          *ClaimTokenData `json:"token"`
	CardInstance   *CardInstanceData `json:"cardInstance"`
	FailureReason  string          `json:"failureReason"`
}

type RevokeClaimTokenReq struct {
	TokenId    int64  `json:"tokenId"`
	IssuerId   int64  `json:"issuerId"`
	Reason     string `json:"reason"`
}

type RevokeClaimTokenResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CardClaimTokenService interface {
	GenerateClaimToken(ctx context.Context, in *GenerateClaimTokenReq, opts ...grpc.CallOption) (*GenerateClaimTokenResp, error)
	ValidateClaimToken(ctx context.Context, in *ValidateClaimTokenReq, opts ...grpc.CallOption) (*ValidateClaimTokenResp, error)
	ConsumeClaimToken(ctx context.Context, in *ConsumeClaimTokenReq, opts ...grpc.CallOption) (*ConsumeClaimTokenResp, error)
	RevokeClaimToken(ctx context.Context, in *RevokeClaimTokenReq, opts ...grpc.CallOption) (*RevokeClaimTokenResp, error)
}

type CardClaimTokenServiceServer interface {
	GenerateClaimToken(ctx context.Context, in *GenerateClaimTokenReq) (*GenerateClaimTokenResp, error)
	ValidateClaimToken(ctx context.Context, in *ValidateClaimTokenReq) (*ValidateClaimTokenResp, error)
	ConsumeClaimToken(ctx context.Context, in *ConsumeClaimTokenReq) (*ConsumeClaimTokenResp, error)
	RevokeClaimToken(ctx context.Context, in *RevokeClaimTokenReq) (*RevokeClaimTokenResp, error)
	mustEmbedUnimplementedCardClaimTokenServiceServer()
}

type UnimplementedCardClaimTokenServiceServer struct{}

func (UnimplementedCardClaimTokenServiceServer) GenerateClaimToken(ctx context.Context, in *GenerateClaimTokenReq) (*GenerateClaimTokenResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateClaimToken not implemented")
}
func (UnimplementedCardClaimTokenServiceServer) ValidateClaimToken(ctx context.Context, in *ValidateClaimTokenReq) (*ValidateClaimTokenResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateClaimToken not implemented")
}
func (UnimplementedCardClaimTokenServiceServer) ConsumeClaimToken(ctx context.Context, in *ConsumeClaimTokenReq) (*ConsumeClaimTokenResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConsumeClaimToken not implemented")
}
func (UnimplementedCardClaimTokenServiceServer) RevokeClaimToken(ctx context.Context, in *RevokeClaimTokenReq) (*RevokeClaimTokenResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevokeClaimToken not implemented")
}
func (UnimplementedCardClaimTokenServiceServer) mustEmbedUnimplementedCardClaimTokenServiceServer() {}

type cardClaimTokenServiceImpl struct {
	conn grpc.ClientConnInterface
}

func NewCardClaimTokenServiceClient(conn grpc.ClientConnInterface) CardClaimTokenService {
	return &cardClaimTokenServiceImpl{conn: conn}
}

func (c *cardClaimTokenServiceImpl) GenerateClaimToken(ctx context.Context, in *GenerateClaimTokenReq, opts ...grpc.CallOption) (*GenerateClaimTokenResp, error) {
	out := new(GenerateClaimTokenResp)
	err := c.conn.Invoke(ctx, CardClaimTokenService_GenerateClaimToken_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardClaimTokenServiceImpl) ValidateClaimToken(ctx context.Context, in *ValidateClaimTokenReq, opts ...grpc.CallOption) (*ValidateClaimTokenResp, error) {
	out := new(ValidateClaimTokenResp)
	err := c.conn.Invoke(ctx, CardClaimTokenService_ValidateClaimToken_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardClaimTokenServiceImpl) ConsumeClaimToken(ctx context.Context, in *ConsumeClaimTokenReq, opts ...grpc.CallOption) (*ConsumeClaimTokenResp, error) {
	out := new(ConsumeClaimTokenResp)
	err := c.conn.Invoke(ctx, CardClaimTokenService_ConsumeClaimToken_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardClaimTokenServiceImpl) RevokeClaimToken(ctx context.Context, in *RevokeClaimTokenReq, opts ...grpc.CallOption) (*RevokeClaimTokenResp, error) {
	out := new(RevokeClaimTokenResp)
	err := c.conn.Invoke(ctx, CardClaimTokenService_RevokeClaimToken_FullMethodName, in, out, opts...)
	return out, err
}

func RegisterCardClaimTokenServiceServer(s *grpc.Server, srv CardClaimTokenServiceServer) {
	s.RegisterService(&CardClaimTokenService_ServiceDesc, srv)
}

var CardClaimTokenService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "smsclient.CardClaimTokenService",
	HandlerType: (*CardClaimTokenServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GenerateClaimToken", Handler: _CardClaimTokenService_GenerateClaimToken_Handler},
		{MethodName: "ValidateClaimToken", Handler: _CardClaimTokenService_ValidateClaimToken_Handler},
		{MethodName: "ConsumeClaimToken", Handler: _CardClaimTokenService_ConsumeClaimToken_Handler},
		{MethodName: "RevokeClaimToken", Handler: _CardClaimTokenService_RevokeClaimToken_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/sms/card_claim_token.proto",
}

func _CardClaimTokenService_GenerateClaimToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateClaimTokenReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardClaimTokenServiceServer).GenerateClaimToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardClaimTokenService_GenerateClaimToken_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardClaimTokenServiceServer).GenerateClaimToken(ctx, req.(*GenerateClaimTokenReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardClaimTokenService_ValidateClaimToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateClaimTokenReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardClaimTokenServiceServer).ValidateClaimToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardClaimTokenService_ValidateClaimToken_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardClaimTokenServiceServer).ValidateClaimToken(ctx, req.(*ValidateClaimTokenReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardClaimTokenService_ConsumeClaimToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConsumeClaimTokenReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardClaimTokenServiceServer).ConsumeClaimToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardClaimTokenService_ConsumeClaimToken_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardClaimTokenServiceServer).ConsumeClaimToken(ctx, req.(*ConsumeClaimTokenReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardClaimTokenService_RevokeClaimToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RevokeClaimTokenReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardClaimTokenServiceServer).RevokeClaimToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardClaimTokenService_RevokeClaimToken_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardClaimTokenServiceServer).RevokeClaimToken(ctx, req.(*RevokeClaimTokenReq))
	}
	return interceptor(ctx, in, info, handler)
}
