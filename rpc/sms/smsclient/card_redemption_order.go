package smsclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CardRedemptionOrderService_CreateRedemptionOrder_FullMethodName  = "/smsclient.CardRedemptionOrderService/CreateRedemptionOrder"
	CardRedemptionOrderService_QueryRedemptionOrder_FullMethodName   = "/smsclient.CardRedemptionOrderService/QueryRedemptionOrder"
	CardRedemptionOrderService_CancelRedemptionOrder_FullMethodName  = "/smsclient.CardRedemptionOrderService/CancelRedemptionOrder"
	CardRedemptionOrderService_UpdateRedemptionOrderStatus_FullMethodName = "/smsclient.CardRedemptionOrderService/UpdateRedemptionOrderStatus"
)

type RedemptionOrderData struct {
	Id              int64  `json:"id"`
	OrderNo         string `json:"orderNo"`
	CardInstanceId  int64  `json:"cardInstanceId"`
	HolderId        int64  `json:"holderId"`
	ReceiverName    string `json:"receiverName"`
	ReceiverPhone   string `json:"receiverPhone"`
	ReceiverAddress string `json:"receiverAddress"`
	Status          string `json:"status"`
	ShippedAt       string `json:"shippedAt"`
	DeliveredAt     string `json:"deliveredAt"`
	CancelReason    string `json:"cancelReason"`
	OmsOrderId      int64  `json:"omsOrderId"`
	PlatformId      int64  `json:"platformId"`
	TenantId        int64  `json:"tenantId"`
	MerchantId      int64  `json:"merchantId"`
	CreateTime      string `json:"createTime"`
	UpdateTime      string `json:"updateTime"`
}

type CreateRedemptionOrderReq struct {
	CardInstanceId  int64  `json:"cardInstanceId"`
	HolderId        int64  `json:"holderId"`
	ReceiverName    string `json:"receiverName"`
	ReceiverPhone   string `json:"receiverPhone"`
	ReceiverAddress string `json:"receiverAddress"`
	PlatformId      int64  `json:"platformId"`
	TenantId        int64  `json:"tenantId"`
	MerchantId      int64  `json:"merchantId"`
	TraceId         string `json:"traceId"`
	RequestId       string `json:"requestId"`
}

type CreateRedemptionOrderResp struct {
	Order *RedemptionOrderData `json:"order"`
}

type QueryRedemptionOrderReq struct {
	OrderId        int64 `json:"orderId"`
	CardInstanceId int64 `json:"cardInstanceId"`
	HolderId       int64 `json:"holderId"`
}

type QueryRedemptionOrderResp struct {
	Order *RedemptionOrderData `json:"order"`
}

type CancelRedemptionOrderReq struct {
	OrderId      int64  `json:"orderId"`
	HolderId     int64  `json:"holderId"`
	CancelReason string `json:"cancelReason"`
}

type CancelRedemptionOrderResp struct {
	Order *RedemptionOrderData `json:"order"`
}

type UpdateRedemptionOrderStatusReq struct {
	OrderId    int64  `json:"orderId"`
	Status     string `json:"status"`
	OmsOrderId int64  `json:"omsOrderId"`
}

type UpdateRedemptionOrderStatusResp struct {
	Order *RedemptionOrderData `json:"order"`
}

type CardRedemptionOrderService interface {
	CreateRedemptionOrder(ctx context.Context, in *CreateRedemptionOrderReq, opts ...grpc.CallOption) (*CreateRedemptionOrderResp, error)
	QueryRedemptionOrder(ctx context.Context, in *QueryRedemptionOrderReq, opts ...grpc.CallOption) (*QueryRedemptionOrderResp, error)
	CancelRedemptionOrder(ctx context.Context, in *CancelRedemptionOrderReq, opts ...grpc.CallOption) (*CancelRedemptionOrderResp, error)
	UpdateRedemptionOrderStatus(ctx context.Context, in *UpdateRedemptionOrderStatusReq, opts ...grpc.CallOption) (*UpdateRedemptionOrderStatusResp, error)
}

type CardRedemptionOrderServiceServer interface {
	CreateRedemptionOrder(ctx context.Context, in *CreateRedemptionOrderReq) (*CreateRedemptionOrderResp, error)
	QueryRedemptionOrder(ctx context.Context, in *QueryRedemptionOrderReq) (*QueryRedemptionOrderResp, error)
	CancelRedemptionOrder(ctx context.Context, in *CancelRedemptionOrderReq) (*CancelRedemptionOrderResp, error)
	UpdateRedemptionOrderStatus(ctx context.Context, in *UpdateRedemptionOrderStatusReq) (*UpdateRedemptionOrderStatusResp, error)
	mustEmbedUnimplementedCardRedemptionOrderServiceServer()
}

type UnimplementedCardRedemptionOrderServiceServer struct{}

func (UnimplementedCardRedemptionOrderServiceServer) CreateRedemptionOrder(ctx context.Context, in *CreateRedemptionOrderReq) (*CreateRedemptionOrderResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRedemptionOrder not implemented")
}
func (UnimplementedCardRedemptionOrderServiceServer) QueryRedemptionOrder(ctx context.Context, in *QueryRedemptionOrderReq) (*QueryRedemptionOrderResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryRedemptionOrder not implemented")
}
func (UnimplementedCardRedemptionOrderServiceServer) CancelRedemptionOrder(ctx context.Context, in *CancelRedemptionOrderReq) (*CancelRedemptionOrderResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelRedemptionOrder not implemented")
}
func (UnimplementedCardRedemptionOrderServiceServer) UpdateRedemptionOrderStatus(ctx context.Context, in *UpdateRedemptionOrderStatusReq) (*UpdateRedemptionOrderStatusResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRedemptionOrderStatus not implemented")
}
func (UnimplementedCardRedemptionOrderServiceServer) mustEmbedUnimplementedCardRedemptionOrderServiceServer() {}

type cardRedemptionOrderServiceImpl struct {
	conn grpc.ClientConnInterface
}

func NewCardRedemptionOrderServiceClient(conn grpc.ClientConnInterface) CardRedemptionOrderService {
	return &cardRedemptionOrderServiceImpl{conn: conn}
}

func (c *cardRedemptionOrderServiceImpl) CreateRedemptionOrder(ctx context.Context, in *CreateRedemptionOrderReq, opts ...grpc.CallOption) (*CreateRedemptionOrderResp, error) {
	out := new(CreateRedemptionOrderResp)
	err := c.conn.Invoke(ctx, CardRedemptionOrderService_CreateRedemptionOrder_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardRedemptionOrderServiceImpl) QueryRedemptionOrder(ctx context.Context, in *QueryRedemptionOrderReq, opts ...grpc.CallOption) (*QueryRedemptionOrderResp, error) {
	out := new(QueryRedemptionOrderResp)
	err := c.conn.Invoke(ctx, CardRedemptionOrderService_QueryRedemptionOrder_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardRedemptionOrderServiceImpl) CancelRedemptionOrder(ctx context.Context, in *CancelRedemptionOrderReq, opts ...grpc.CallOption) (*CancelRedemptionOrderResp, error) {
	out := new(CancelRedemptionOrderResp)
	err := c.conn.Invoke(ctx, CardRedemptionOrderService_CancelRedemptionOrder_FullMethodName, in, out, opts...)
	return out, err
}

func (c *cardRedemptionOrderServiceImpl) UpdateRedemptionOrderStatus(ctx context.Context, in *UpdateRedemptionOrderStatusReq, opts ...grpc.CallOption) (*UpdateRedemptionOrderStatusResp, error) {
	out := new(UpdateRedemptionOrderStatusResp)
	err := c.conn.Invoke(ctx, CardRedemptionOrderService_UpdateRedemptionOrderStatus_FullMethodName, in, out, opts...)
	return out, err
}

func RegisterCardRedemptionOrderServiceServer(s *grpc.Server, srv CardRedemptionOrderServiceServer) {
	s.RegisterService(&CardRedemptionOrderService_ServiceDesc, srv)
}

var CardRedemptionOrderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "smsclient.CardRedemptionOrderService",
	HandlerType: (*CardRedemptionOrderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateRedemptionOrder", Handler: _CardRedemptionOrderService_CreateRedemptionOrder_Handler},
		{MethodName: "QueryRedemptionOrder", Handler: _CardRedemptionOrderService_QueryRedemptionOrder_Handler},
		{MethodName: "CancelRedemptionOrder", Handler: _CardRedemptionOrderService_CancelRedemptionOrder_Handler},
		{MethodName: "UpdateRedemptionOrderStatus", Handler: _CardRedemptionOrderService_UpdateRedemptionOrderStatus_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/sms/card_redemption_order.proto",
}

func _CardRedemptionOrderService_CreateRedemptionOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRedemptionOrderReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardRedemptionOrderServiceServer).CreateRedemptionOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardRedemptionOrderService_CreateRedemptionOrder_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardRedemptionOrderServiceServer).CreateRedemptionOrder(ctx, req.(*CreateRedemptionOrderReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardRedemptionOrderService_QueryRedemptionOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRedemptionOrderReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardRedemptionOrderServiceServer).QueryRedemptionOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardRedemptionOrderService_QueryRedemptionOrder_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardRedemptionOrderServiceServer).QueryRedemptionOrder(ctx, req.(*QueryRedemptionOrderReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardRedemptionOrderService_CancelRedemptionOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelRedemptionOrderReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardRedemptionOrderServiceServer).CancelRedemptionOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardRedemptionOrderService_CancelRedemptionOrder_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardRedemptionOrderServiceServer).CancelRedemptionOrder(ctx, req.(*CancelRedemptionOrderReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _CardRedemptionOrderService_UpdateRedemptionOrderStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRedemptionOrderStatusReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CardRedemptionOrderServiceServer).UpdateRedemptionOrderStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CardRedemptionOrderService_UpdateRedemptionOrderStatus_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CardRedemptionOrderServiceServer).UpdateRedemptionOrderStatus(ctx, req.(*UpdateRedemptionOrderStatusReq))
	}
	return interceptor(ctx, in, info, handler)
}
