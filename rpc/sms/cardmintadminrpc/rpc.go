package cardmintadminrpc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	ServiceName = "sms.CardMintAdminService"

	MethodQueryTaskList                      = "/" + ServiceName + "/QueryTaskList"
	MethodQueryTaskDetail                    = "/" + ServiceName + "/QueryTaskDetail"
	MethodQueryAvailableTaskActions          = "/" + ServiceName + "/QueryAvailableTaskActions"
	MethodRetryTask                          = "/" + ServiceName + "/RetryTask"
	MethodFreezeTask                         = "/" + ServiceName + "/FreezeTask"
	MethodEscalateTask                       = "/" + ServiceName + "/EscalateTask"
	MethodQueryAssetAuditList                = "/" + ServiceName + "/QueryAssetAuditList"
	MethodQueryAssetAuditDetail              = "/" + ServiceName + "/QueryAssetAuditDetail"
	MethodReviewAssetCompliance              = "/" + ServiceName + "/ReviewAssetCompliance"
	MethodOfflineAssetDisplay                = "/" + ServiceName + "/OfflineAssetDisplay"
	MethodRecycleAsset                       = "/" + ServiceName + "/RecycleAsset"
	MethodQueryPhysicalFulfillmentList       = "/" + ServiceName + "/QueryPhysicalFulfillmentList"
	MethodQueryPhysicalFulfillmentDetail     = "/" + ServiceName + "/QueryPhysicalFulfillmentDetail"
	MethodEnsurePhysicalFulfillment          = "/" + ServiceName + "/EnsurePhysicalFulfillment"
	MethodUpdatePhysicalCardProductionStatus = "/" + ServiceName + "/UpdatePhysicalCardProductionStatus"
	MethodShipPhysicalCard                   = "/" + ServiceName + "/ShipPhysicalCard"
	MethodMarkPhysicalFulfillmentException   = "/" + ServiceName + "/MarkPhysicalFulfillmentException"
	MethodRequestPhysicalCardReissue         = "/" + ServiceName + "/RequestPhysicalCardReissue"
	// Story 10.7 Task 9.2 — 后台合规审计跨资产检索
	MethodQueryDigitalCardTransferLogList = "/" + ServiceName + "/QueryDigitalCardTransferLogList"
	// Story 10.7 Task 2.8 / S3 — 后台提货单管理
	MethodQueryDigitalCardRedemptionOrderList = "/" + ServiceName + "/QueryDigitalCardRedemptionOrderList"
)

type (
	QueryTaskListRequest struct {
		Scope  pkgscope.GovernanceScope    `json:"scope"`
		Filter digitalcardmint.QueryFilter `json:"filter"`
	}

	QueryTaskListResponse struct {
		Total int64                           `json:"total"`
		List  []*digitalcardmint.TaskListItem `json:"list"`
	}

	QueryTaskDetailRequest struct {
		Scope  pkgscope.GovernanceScope `json:"scope"`
		TaskID int64                    `json:"taskId"`
	}

	QueryTaskDetailResponse struct {
		Detail *digitalcardmint.TaskDetail `json:"detail"`
	}

	QueryAvailableTaskActionsRequest struct {
		Scope  pkgscope.GovernanceScope `json:"scope"`
		TaskID int64                    `json:"taskId"`
	}

	QueryAvailableTaskActionsResponse struct {
		Actions []string `json:"actions"`
	}

	TaskActionRequest struct {
		Scope      pkgscope.GovernanceScope `json:"scope"`
		TaskID     int64                    `json:"taskId"`
		OperatorID int64                    `json:"operatorId"`
		Reason     string                   `json:"reason"`
	}

	TaskActionResponse struct {
		Result *digitalcardmint.ActionResult `json:"result"`
	}

	QueryAssetAuditListRequest struct {
		Scope  pkgscope.GovernanceScope                    `json:"scope"`
		Filter digitalcardmint.DigitalCardAssetAuditFilter `json:"filter"`
	}

	QueryAssetAuditListResponse struct {
		Total int64                                       `json:"total"`
		List  []digitalcardmint.DigitalCardAssetAuditItem `json:"list"`
	}

	QueryAssetAuditDetailRequest struct {
		Scope           pkgscope.GovernanceScope `json:"scope"`
		AssetInstanceID int64                    `json:"assetInstanceId"`
	}

	QueryAssetAuditDetailResponse struct {
		Detail *digitalcardmint.DigitalCardAssetAuditDetail `json:"detail"`
	}

	AssetActionRequest struct {
		Scope           pkgscope.GovernanceScope `json:"scope"`
		AssetInstanceID int64                    `json:"assetInstanceId"`
		OperatorID      int64                    `json:"operatorId"`
		Reason          string                   `json:"reason"`
	}

	AssetActionResponse struct {
		Result *digitalcardmint.DigitalCardAssetActionResult `json:"result"`
	}

	QueryPhysicalFulfillmentListRequest struct {
		Scope  pkgscope.GovernanceScope                  `json:"scope"`
		Filter digitalcardmint.PhysicalFulfillmentFilter `json:"filter"`
	}

	QueryPhysicalFulfillmentListResponse struct {
		Total int64                                     `json:"total"`
		List  []digitalcardmint.PhysicalFulfillmentItem `json:"list"`
	}

	QueryPhysicalFulfillmentDetailRequest struct {
		Scope         pkgscope.GovernanceScope `json:"scope"`
		FulfillmentID int64                    `json:"fulfillmentId"`
	}

	QueryPhysicalFulfillmentDetailResponse struct {
		Detail *digitalcardmint.PhysicalFulfillmentAdminDetail `json:"detail"`
	}

	EnsurePhysicalFulfillmentRequest struct {
		Scope pkgscope.GovernanceScope                 `json:"scope"`
		Input digitalcardmint.PhysicalFulfillmentInput `json:"input"`
	}

	PhysicalFulfillmentResultResponse struct {
		Result *digitalcardmint.PhysicalFulfillmentResult `json:"result"`
	}

	UpdatePhysicalCardProductionStatusRequest struct {
		Scope pkgscope.GovernanceScope                                `json:"scope"`
		Input digitalcardmint.UpdatePhysicalCardProductionStatusInput `json:"input"`
	}

	ShipPhysicalCardRequest struct {
		Scope pkgscope.GovernanceScope              `json:"scope"`
		Input digitalcardmint.ShipPhysicalCardInput `json:"input"`
	}

	PhysicalFulfillmentExceptionRequest struct {
		Scope pkgscope.GovernanceScope                          `json:"scope"`
		Input digitalcardmint.PhysicalFulfillmentExceptionInput `json:"input"`
	}

	// Story 10.7 Task 9.2 — 后台合规审计跨资产检索
	QueryDigitalCardTransferLogListRequest struct {
		Scope  pkgscope.GovernanceScope                     `json:"scope"`
		Filter digitalcardmint.DigitalCardTransferLogFilter `json:"filter"`
	}

	QueryDigitalCardTransferLogListResponse struct {
		Total int64                                        `json:"total"`
		List  []digitalcardmint.DigitalCardTransferLogItem `json:"list"`
	}

	// Story 10.7 Task 2.8 / S3 — 后台提货单管理
	QueryDigitalCardRedemptionOrderListRequest struct {
		Scope  pkgscope.GovernanceScope                         `json:"scope"`
		Filter digitalcardmint.DigitalCardRedemptionOrderFilter `json:"filter"`
	}

	QueryDigitalCardRedemptionOrderListResponse struct {
		Total int64                                            `json:"total"`
		List  []digitalcardmint.DigitalCardRedemptionOrderItem `json:"list"`
	}
)

type CardMintAdminServiceServer interface {
	QueryTaskList(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryTaskDetail(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryAvailableTaskActions(context.Context, *structpb.Struct) (*structpb.Struct, error)
	RetryTask(context.Context, *structpb.Struct) (*structpb.Struct, error)
	FreezeTask(context.Context, *structpb.Struct) (*structpb.Struct, error)
	EscalateTask(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryAssetAuditList(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryAssetAuditDetail(context.Context, *structpb.Struct) (*structpb.Struct, error)
	ReviewAssetCompliance(context.Context, *structpb.Struct) (*structpb.Struct, error)
	OfflineAssetDisplay(context.Context, *structpb.Struct) (*structpb.Struct, error)
	RecycleAsset(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryPhysicalFulfillmentList(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryPhysicalFulfillmentDetail(context.Context, *structpb.Struct) (*structpb.Struct, error)
	EnsurePhysicalFulfillment(context.Context, *structpb.Struct) (*structpb.Struct, error)
	UpdatePhysicalCardProductionStatus(context.Context, *structpb.Struct) (*structpb.Struct, error)
	ShipPhysicalCard(context.Context, *structpb.Struct) (*structpb.Struct, error)
	MarkPhysicalFulfillmentException(context.Context, *structpb.Struct) (*structpb.Struct, error)
	RequestPhysicalCardReissue(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryDigitalCardTransferLogList(context.Context, *structpb.Struct) (*structpb.Struct, error)
	QueryDigitalCardRedemptionOrderList(context.Context, *structpb.Struct) (*structpb.Struct, error)
}

func RegisterCardMintAdminServiceServer(registrar grpc.ServiceRegistrar, server CardMintAdminServiceServer) {
	registrar.RegisterService(&CardMintAdminServiceDesc, server)
}

var CardMintAdminServiceDesc = grpc.ServiceDesc{
	ServiceName: ServiceName,
	HandlerType: (*CardMintAdminServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "QueryTaskList", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryTaskList(ctx, in)
		}, MethodQueryTaskList)},
		{MethodName: "QueryTaskDetail", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryTaskDetail(ctx, in)
		}, MethodQueryTaskDetail)},
		{MethodName: "QueryAvailableTaskActions", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryAvailableTaskActions(ctx, in)
		}, MethodQueryAvailableTaskActions)},
		{MethodName: "RetryTask", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.RetryTask(ctx, in)
		}, MethodRetryTask)},
		{MethodName: "FreezeTask", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.FreezeTask(ctx, in)
		}, MethodFreezeTask)},
		{MethodName: "EscalateTask", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.EscalateTask(ctx, in)
		}, MethodEscalateTask)},
		{MethodName: "QueryAssetAuditList", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryAssetAuditList(ctx, in)
		}, MethodQueryAssetAuditList)},
		{MethodName: "QueryAssetAuditDetail", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryAssetAuditDetail(ctx, in)
		}, MethodQueryAssetAuditDetail)},
		{MethodName: "ReviewAssetCompliance", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.ReviewAssetCompliance(ctx, in)
		}, MethodReviewAssetCompliance)},
		{MethodName: "OfflineAssetDisplay", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.OfflineAssetDisplay(ctx, in)
		}, MethodOfflineAssetDisplay)},
		{MethodName: "RecycleAsset", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.RecycleAsset(ctx, in)
		}, MethodRecycleAsset)},
		{MethodName: "QueryPhysicalFulfillmentList", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryPhysicalFulfillmentList(ctx, in)
		}, MethodQueryPhysicalFulfillmentList)},
		{MethodName: "QueryPhysicalFulfillmentDetail", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryPhysicalFulfillmentDetail(ctx, in)
		}, MethodQueryPhysicalFulfillmentDetail)},
		{MethodName: "EnsurePhysicalFulfillment", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.EnsurePhysicalFulfillment(ctx, in)
		}, MethodEnsurePhysicalFulfillment)},
		{MethodName: "UpdatePhysicalCardProductionStatus", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.UpdatePhysicalCardProductionStatus(ctx, in)
		}, MethodUpdatePhysicalCardProductionStatus)},
		{MethodName: "ShipPhysicalCard", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.ShipPhysicalCard(ctx, in)
		}, MethodShipPhysicalCard)},
		{MethodName: "MarkPhysicalFulfillmentException", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.MarkPhysicalFulfillmentException(ctx, in)
		}, MethodMarkPhysicalFulfillmentException)},
		{MethodName: "RequestPhysicalCardReissue", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.RequestPhysicalCardReissue(ctx, in)
		}, MethodRequestPhysicalCardReissue)},
		{MethodName: "QueryDigitalCardTransferLogList", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryDigitalCardTransferLogList(ctx, in)
		}, MethodQueryDigitalCardTransferLogList)},
		{MethodName: "QueryDigitalCardRedemptionOrderList", Handler: buildUnaryHandler(func(s CardMintAdminServiceServer, ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
			return s.QueryDigitalCardRedemptionOrderList(ctx, in)
		}, MethodQueryDigitalCardRedemptionOrderList)},
	},
}

func EncodePayload(payload interface{}) (*structpb.Struct, error) {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return structpb.NewStruct(map[string]interface{}{
		"payload": string(raw),
	})
}

func DecodePayload(message *structpb.Struct, out interface{}) error {
	if message == nil {
		return errors.New("rpc payload 不能为空")
	}
	value, ok := message.Fields["payload"]
	if !ok {
		return errors.New("rpc payload 缺少 payload 字段")
	}
	return json.Unmarshal([]byte(value.GetStringValue()), out)
}

func buildUnaryHandler(
	handler func(CardMintAdminServiceServer, context.Context, *structpb.Struct) (*structpb.Struct, error),
	fullMethod string,
) grpc.MethodHandler {
	return func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
		in := new(structpb.Struct)
		if err := dec(in); err != nil {
			return nil, err
		}
		if interceptor == nil {
			return handler(srv.(CardMintAdminServiceServer), ctx, in)
		}
		info := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: fullMethod,
		}
		return interceptor(ctx, in, info, func(innerCtx context.Context, req interface{}) (interface{}, error) {
			return handler(srv.(CardMintAdminServiceServer), innerCtx, req.(*structpb.Struct))
		})
	}
}
