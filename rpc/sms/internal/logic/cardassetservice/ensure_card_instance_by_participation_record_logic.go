package cardassetservicelogic

import (
	"context"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type EnsureCardInstanceByParticipationRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnsureCardInstanceByParticipationRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnsureCardInstanceByParticipationRecordLogic {
	return &EnsureCardInstanceByParticipationRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnsureCardInstanceByParticipationRecordLogic) EnsureCardInstanceByParticipationRecord(in *smsclient.EnsureCardInstanceByParticipationRecordReq) (*smsclient.EnsureCardInstanceByParticipationRecordResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	var asset *CardInstanceSnapshot
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if _, txErr := validateParticipationRecordScope(l.ctx, tx, currentScope, in.ParticipationRecordId, true); txErr != nil {
			return txErr
		}

		var txErr error
		asset, txErr = EnsureCardInstanceByParticipationRecord(l.ctx, tx, in.ParticipationRecordId, in.OperatorType, in.TraceId)
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return &smsclient.EnsureCardInstanceByParticipationRecordResp{
		Asset: buildCardInstanceData(asset),
	}, nil
}
