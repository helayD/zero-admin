package cardmintadminservice

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/rpc/sms/cardmintadminrpc"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"google.golang.org/protobuf/types/known/structpb"
)

type CardMintAdminServiceServer struct {
	svcCtx *svc.ServiceContext
}

func NewCardMintAdminServiceServer(svcCtx *svc.ServiceContext) *CardMintAdminServiceServer {
	return &CardMintAdminServiceServer{svcCtx: svcCtx}
}

func (s *CardMintAdminServiceServer) QueryTaskList(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryTaskListRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	total, list, err := s.svcCtx.CardMintService.QueryTaskList(ctx, req.Scope, req.Filter)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryTaskListResponse{
		Total: total,
		List:  list,
	})
}

func (s *CardMintAdminServiceServer) QueryTaskDetail(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryTaskDetailRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	detail, err := s.svcCtx.CardMintService.QueryTaskDetail(ctx, req.Scope, req.TaskID)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryTaskDetailResponse{Detail: detail})
}

func (s *CardMintAdminServiceServer) QueryAvailableTaskActions(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryAvailableTaskActionsRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	actions, err := s.svcCtx.CardMintService.QueryAvailableActions(ctx, req.Scope, req.TaskID)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryAvailableTaskActionsResponse{Actions: actions})
}

func (s *CardMintAdminServiceServer) RetryTask(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	return s.handleTaskAction(ctx, in, func(req cardmintadminrpc.TaskActionRequest) (interface{}, error) {
		result, err := s.svcCtx.CardMintService.RetryTask(ctx, req.Scope, req.TaskID, req.OperatorID, req.Reason)
		if err != nil {
			return nil, err
		}
		return cardmintadminrpc.TaskActionResponse{Result: result}, nil
	})
}

func (s *CardMintAdminServiceServer) FreezeTask(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	return s.handleTaskAction(ctx, in, func(req cardmintadminrpc.TaskActionRequest) (interface{}, error) {
		result, err := s.svcCtx.CardMintService.FreezeTask(ctx, req.Scope, req.TaskID, req.OperatorID, req.Reason)
		if err != nil {
			return nil, err
		}
		return cardmintadminrpc.TaskActionResponse{Result: result}, nil
	})
}

func (s *CardMintAdminServiceServer) EscalateTask(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	return s.handleTaskAction(ctx, in, func(req cardmintadminrpc.TaskActionRequest) (interface{}, error) {
		result, err := s.svcCtx.CardMintService.EscalateTask(ctx, req.Scope, req.TaskID, req.OperatorID, req.Reason)
		if err != nil {
			return nil, err
		}
		return cardmintadminrpc.TaskActionResponse{Result: result}, nil
	})
}

func (s *CardMintAdminServiceServer) QueryAssetAuditList(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryAssetAuditListRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	total, list, err := s.svcCtx.CardMintService.QueryDigitalCardAssetAuditList(ctx, req.Scope, req.Filter)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryAssetAuditListResponse{
		Total: total,
		List:  list,
	})
}

func (s *CardMintAdminServiceServer) QueryAssetAuditDetail(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryAssetAuditDetailRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	detail, err := s.svcCtx.CardMintService.QueryDigitalCardAssetAuditDetail(ctx, req.Scope, req.AssetInstanceID)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryAssetAuditDetailResponse{Detail: detail})
}

func (s *CardMintAdminServiceServer) ReviewAssetCompliance(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	return s.handleAssetAction(ctx, in, func(req cardmintadminrpc.AssetActionRequest) (interface{}, error) {
		result, err := s.svcCtx.CardMintService.ReviewDigitalCardAssetCompliance(ctx, req.Scope, req.AssetInstanceID, req.OperatorID, req.Reason)
		if err != nil {
			return nil, err
		}
		return cardmintadminrpc.AssetActionResponse{Result: result}, nil
	})
}

func (s *CardMintAdminServiceServer) OfflineAssetDisplay(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	return s.handleAssetAction(ctx, in, func(req cardmintadminrpc.AssetActionRequest) (interface{}, error) {
		result, err := s.svcCtx.CardMintService.OfflineDigitalCardAssetDisplay(ctx, req.Scope, req.AssetInstanceID, req.OperatorID, req.Reason)
		if err != nil {
			return nil, err
		}
		return cardmintadminrpc.AssetActionResponse{Result: result}, nil
	})
}

func (s *CardMintAdminServiceServer) RecycleAsset(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	return s.handleAssetAction(ctx, in, func(req cardmintadminrpc.AssetActionRequest) (interface{}, error) {
		result, err := s.svcCtx.CardMintService.RecycleDigitalCardAsset(ctx, req.Scope, req.AssetInstanceID, req.OperatorID, req.Reason)
		if err != nil {
			return nil, err
		}
		return cardmintadminrpc.AssetActionResponse{Result: result}, nil
	})
}

func (s *CardMintAdminServiceServer) QueryPhysicalFulfillmentList(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryPhysicalFulfillmentListRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if err := s.ensureCardMintService(); err != nil {
		return nil, err
	}
	total, list, err := s.svcCtx.CardMintService.QueryPhysicalFulfillmentList(ctx, req.Scope, req.Filter)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryPhysicalFulfillmentListResponse{
		Total: total,
		List:  list,
	})
}

func (s *CardMintAdminServiceServer) QueryPhysicalFulfillmentDetail(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryPhysicalFulfillmentDetailRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if err := s.ensureCardMintService(); err != nil {
		return nil, err
	}
	detail, err := s.svcCtx.CardMintService.QueryPhysicalFulfillmentAdminDetail(ctx, req.Scope, req.FulfillmentID)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryPhysicalFulfillmentDetailResponse{Detail: detail})
}

func (s *CardMintAdminServiceServer) EnsurePhysicalFulfillment(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.EnsurePhysicalFulfillmentRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	return s.handlePhysicalResult(func(service *digitalcardmint.Service) (*digitalcardmint.PhysicalFulfillmentResult, error) {
		return service.EnsurePhysicalFulfillmentByAsset(ctx, req.Scope, req.Input)
	})
}

func (s *CardMintAdminServiceServer) UpdatePhysicalCardProductionStatus(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.UpdatePhysicalCardProductionStatusRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	return s.handlePhysicalResult(func(service *digitalcardmint.Service) (*digitalcardmint.PhysicalFulfillmentResult, error) {
		return service.UpdatePhysicalCardProductionStatus(ctx, req.Scope, req.Input)
	})
}

func (s *CardMintAdminServiceServer) ShipPhysicalCard(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.ShipPhysicalCardRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	return s.handlePhysicalResult(func(service *digitalcardmint.Service) (*digitalcardmint.PhysicalFulfillmentResult, error) {
		return service.ShipPhysicalCard(ctx, req.Scope, req.Input)
	})
}

func (s *CardMintAdminServiceServer) MarkPhysicalFulfillmentException(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.PhysicalFulfillmentExceptionRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	return s.handlePhysicalResult(func(service *digitalcardmint.Service) (*digitalcardmint.PhysicalFulfillmentResult, error) {
		return service.MarkPhysicalFulfillmentException(ctx, req.Scope, req.Input)
	})
}

func (s *CardMintAdminServiceServer) RequestPhysicalCardReissue(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.PhysicalFulfillmentExceptionRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	return s.handlePhysicalResult(func(service *digitalcardmint.Service) (*digitalcardmint.PhysicalFulfillmentResult, error) {
		return service.RequestPhysicalCardReissue(ctx, req.Scope, req.Input)
	})
}

func (s *CardMintAdminServiceServer) handleTaskAction(ctx context.Context, in *structpb.Struct, handler func(cardmintadminrpc.TaskActionRequest) (interface{}, error)) (*structpb.Struct, error) {
	var req cardmintadminrpc.TaskActionRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	payload, err := handler(req)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(payload)
}

func (s *CardMintAdminServiceServer) handlePhysicalResult(handler func(*digitalcardmint.Service) (*digitalcardmint.PhysicalFulfillmentResult, error)) (*structpb.Struct, error) {
	if err := s.ensureCardMintService(); err != nil {
		return nil, err
	}
	result, err := handler(s.svcCtx.CardMintService)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.PhysicalFulfillmentResultResponse{Result: result})
}

// QueryDigitalCardTransferLogList Story 10.7 Task 9.2 — 后台合规审计跨资产检索
func (s *CardMintAdminServiceServer) QueryDigitalCardTransferLogList(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var req cardmintadminrpc.QueryDigitalCardTransferLogListRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if err := s.ensureCardMintService(); err != nil {
		return nil, err
	}
	total, list, err := s.svcCtx.CardMintService.QueryDigitalCardTransferLogList(ctx, req.Scope, req.Filter)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(cardmintadminrpc.QueryDigitalCardTransferLogListResponse{
		Total: total,
		List:  list,
	})
}

func (s *CardMintAdminServiceServer) ensureCardMintService() error {
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return errors.New("提货卡服务未初始化")
	}
	return nil
}

func (s *CardMintAdminServiceServer) handleAssetAction(ctx context.Context, in *structpb.Struct, handler func(cardmintadminrpc.AssetActionRequest) (interface{}, error)) (*structpb.Struct, error) {
	var req cardmintadminrpc.AssetActionRequest
	if err := cardmintadminrpc.DecodePayload(in, &req); err != nil {
		return nil, err
	}
	if s.svcCtx == nil || s.svcCtx.CardMintService == nil {
		return nil, errors.New("提货卡服务未初始化")
	}
	payload, err := handler(req)
	if err != nil {
		return nil, err
	}
	return cardmintadminrpc.EncodePayload(payload)
}
