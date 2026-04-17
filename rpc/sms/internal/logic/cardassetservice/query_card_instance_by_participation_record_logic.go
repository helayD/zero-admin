package cardassetservicelogic

import (
	"context"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryCardInstanceByParticipationRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCardInstanceByParticipationRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCardInstanceByParticipationRecordLogic {
	return &QueryCardInstanceByParticipationRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryCardInstanceByParticipationRecordLogic) QueryCardInstanceByParticipationRecord(in *smsclient.QueryCardInstanceByParticipationRecordReq) (*smsclient.QueryCardInstanceByParticipationRecordResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if _, err = validateParticipationRecordScope(l.ctx, l.svcCtx.DB, currentScope, in.ParticipationRecordId, false); err != nil {
		return nil, err
	}

	asset, err := QueryCardInstanceByParticipationRecord(l.ctx, l.svcCtx.DB, in.ParticipationRecordId)
	if err != nil {
		return nil, err
	}
	return &smsclient.QueryCardInstanceByParticipationRecordResp{
		Asset: buildCardInstanceData(asset),
	}, nil
}
