package drawactivityservicelogic

import (
	"context"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type PreviewDrawActivityPublishReadinessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPreviewDrawActivityPublishReadinessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewDrawActivityPublishReadinessLogic {
	return &PreviewDrawActivityPublishReadinessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PreviewDrawActivityPublishReadinessLogic) PreviewDrawActivityPublishReadiness(in *smsclient.PreviewDrawActivityPublishReadinessReq) (*smsclient.PreviewDrawActivityPublishReadinessResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	aggregate, err := loadDrawActivityAggregate(l.ctx, l.svcCtx.DB, currentScope, in.Id)
	if err != nil {
		return nil, err
	}
	if err := syncDrawReadiness(l.ctx, l.svcCtx.DB, in.Id, aggregate.Readiness); err != nil {
		return nil, err
	}
	return &smsclient.PreviewDrawActivityPublishReadinessResp{
		ReadyToPublish:   aggregate.Readiness.ReadyToPublish,
		PublishReadiness: aggregate.Readiness.PublishReadiness,
		ReadinessLabel:   aggregate.Readiness.Label,
		Summary:          aggregate.Readiness.Summary,
		Items:            aggregate.Readiness.Items,
	}, nil
}
