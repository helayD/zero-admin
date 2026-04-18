package cardmintadminservice

import (
	"context"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/cardmintadminrpc"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type CardMintAdminService interface {
	QueryTaskList(ctx context.Context, scope pkgscope.GovernanceScope, filter digitalcardmint.QueryFilter, opts ...grpc.CallOption) (int64, []*digitalcardmint.TaskListItem, error)
	QueryTaskDetail(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, opts ...grpc.CallOption) (*digitalcardmint.TaskDetail, error)
	QueryAvailableTaskActions(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, opts ...grpc.CallOption) ([]string, error)
	RetryTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error)
	FreezeTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error)
	EscalateTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error)
	QueryAssetAuditList(ctx context.Context, scope pkgscope.GovernanceScope, filter digitalcardmint.DigitalCardAssetAuditFilter, opts ...grpc.CallOption) (int64, []digitalcardmint.DigitalCardAssetAuditItem, error)
	QueryAssetAuditDetail(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetAuditDetail, error)
	ReviewAssetCompliance(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error)
	OfflineAssetDisplay(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error)
	RecycleAsset(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error)
}

type cardMintAdminService struct {
	cli zrpc.Client
}

func NewCardMintAdminService(cli zrpc.Client) CardMintAdminService {
	return &cardMintAdminService{cli: cli}
}

func (m *cardMintAdminService) QueryTaskList(ctx context.Context, scope pkgscope.GovernanceScope, filter digitalcardmint.QueryFilter, opts ...grpc.CallOption) (int64, []*digitalcardmint.TaskListItem, error) {
	var out cardmintadminrpc.QueryTaskListResponse
	if err := m.invoke(ctx, cardmintadminrpc.MethodQueryTaskList, cardmintadminrpc.QueryTaskListRequest{
		Scope:  scope,
		Filter: filter,
	}, &out, opts...); err != nil {
		return 0, nil, err
	}
	return out.Total, out.List, nil
}

func (m *cardMintAdminService) QueryTaskDetail(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, opts ...grpc.CallOption) (*digitalcardmint.TaskDetail, error) {
	var out cardmintadminrpc.QueryTaskDetailResponse
	if err := m.invoke(ctx, cardmintadminrpc.MethodQueryTaskDetail, cardmintadminrpc.QueryTaskDetailRequest{
		Scope:  scope,
		TaskID: taskID,
	}, &out, opts...); err != nil {
		return nil, err
	}
	return out.Detail, nil
}

func (m *cardMintAdminService) QueryAvailableTaskActions(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, opts ...grpc.CallOption) ([]string, error) {
	var out cardmintadminrpc.QueryAvailableTaskActionsResponse
	if err := m.invoke(ctx, cardmintadminrpc.MethodQueryAvailableTaskActions, cardmintadminrpc.QueryAvailableTaskActionsRequest{
		Scope:  scope,
		TaskID: taskID,
	}, &out, opts...); err != nil {
		return nil, err
	}
	return out.Actions, nil
}

func (m *cardMintAdminService) RetryTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return m.invokeTaskAction(ctx, cardmintadminrpc.MethodRetryTask, scope, taskID, operatorID, reason, opts...)
}

func (m *cardMintAdminService) FreezeTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return m.invokeTaskAction(ctx, cardmintadminrpc.MethodFreezeTask, scope, taskID, operatorID, reason, opts...)
}

func (m *cardMintAdminService) EscalateTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return m.invokeTaskAction(ctx, cardmintadminrpc.MethodEscalateTask, scope, taskID, operatorID, reason, opts...)
}

func (m *cardMintAdminService) QueryAssetAuditList(ctx context.Context, scope pkgscope.GovernanceScope, filter digitalcardmint.DigitalCardAssetAuditFilter, opts ...grpc.CallOption) (int64, []digitalcardmint.DigitalCardAssetAuditItem, error) {
	var out cardmintadminrpc.QueryAssetAuditListResponse
	if err := m.invoke(ctx, cardmintadminrpc.MethodQueryAssetAuditList, cardmintadminrpc.QueryAssetAuditListRequest{
		Scope:  scope,
		Filter: filter,
	}, &out, opts...); err != nil {
		return 0, nil, err
	}
	return out.Total, out.List, nil
}

func (m *cardMintAdminService) QueryAssetAuditDetail(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetAuditDetail, error) {
	var out cardmintadminrpc.QueryAssetAuditDetailResponse
	if err := m.invoke(ctx, cardmintadminrpc.MethodQueryAssetAuditDetail, cardmintadminrpc.QueryAssetAuditDetailRequest{
		Scope:           scope,
		AssetInstanceID: assetInstanceID,
	}, &out, opts...); err != nil {
		return nil, err
	}
	return out.Detail, nil
}

func (m *cardMintAdminService) ReviewAssetCompliance(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return m.invokeAssetAction(ctx, cardmintadminrpc.MethodReviewAssetCompliance, scope, assetInstanceID, operatorID, reason, opts...)
}

func (m *cardMintAdminService) OfflineAssetDisplay(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return m.invokeAssetAction(ctx, cardmintadminrpc.MethodOfflineAssetDisplay, scope, assetInstanceID, operatorID, reason, opts...)
}

func (m *cardMintAdminService) RecycleAsset(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return m.invokeAssetAction(ctx, cardmintadminrpc.MethodRecycleAsset, scope, assetInstanceID, operatorID, reason, opts...)
}

func (m *cardMintAdminService) invokeTaskAction(ctx context.Context, method string, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	var out cardmintadminrpc.TaskActionResponse
	if err := m.invoke(ctx, method, cardmintadminrpc.TaskActionRequest{
		Scope:      scope,
		TaskID:     taskID,
		OperatorID: operatorID,
		Reason:     reason,
	}, &out, opts...); err != nil {
		return nil, err
	}
	return out.Result, nil
}

func (m *cardMintAdminService) invokeAssetAction(ctx context.Context, method string, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, opts ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	var out cardmintadminrpc.AssetActionResponse
	if err := m.invoke(ctx, method, cardmintadminrpc.AssetActionRequest{
		Scope:           scope,
		AssetInstanceID: assetInstanceID,
		OperatorID:      operatorID,
		Reason:          reason,
	}, &out, opts...); err != nil {
		return nil, err
	}
	return out.Result, nil
}

func (m *cardMintAdminService) invoke(ctx context.Context, method string, request interface{}, response interface{}, opts ...grpc.CallOption) error {
	in, err := cardmintadminrpc.EncodePayload(request)
	if err != nil {
		return err
	}
	out := new(structpb.Struct)
	if err = m.cli.Conn().Invoke(ctx, method, in, out, opts...); err != nil {
		return err
	}
	return cardmintadminrpc.DecodePayload(out, response)
}
