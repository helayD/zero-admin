package channelintegrationtemplateservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChannelIntegrationTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChannelIntegrationTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChannelIntegrationTemplateLogic {
	return &CreateChannelIntegrationTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateChannelIntegrationTemplateLogic) CreateChannelIntegrationTemplate(in *sysclient.CreateChannelIntegrationTemplateReq) (*sysclient.CreateChannelIntegrationTemplateResp, error) {
	payload, err := normalizeTemplatePayload(
		in.TemplateCode,
		in.TemplateName,
		in.TemplateType,
		in.TargetCode,
		in.ScopeType,
		in.PlatformId,
		in.TenantId,
		in.MerchantId,
		in.Status,
		in.MetadataConfig,
		in.SecretRefConfig,
		in.IntentContractConfig,
		in.ImpactScopeConfig,
		in.Remark,
	)
	if err != nil {
		return nil, err
	}
	if err := ensureTemplateScopeSubjectExists(l.ctx, l.svcCtx.DB, payload.ScopeType, payload.TenantID, payload.MerchantID); err != nil {
		return nil, err
	}

	row := &channelIntegrationTemplateRow{
		TemplateCode:         payload.TemplateCode,
		TemplateName:         payload.TemplateName,
		TemplateType:         payload.TemplateType,
		TargetCode:           payload.TargetCode,
		ScopeType:            payload.ScopeType,
		PlatformID:           payload.PlatformID,
		TenantID:             payload.TenantID,
		MerchantID:           payload.MerchantID,
		Status:               payload.Status,
		MetadataConfig:       payload.MetadataConfig,
		SecretRefConfig:      payload.SecretRefConfig,
		IntentContractConfig: payload.IntentContractConfig,
		ImpactScopeConfig:    payload.ImpactScopeConfig,
		Remark:               payload.Remark,
		CreateBy:             strings.TrimSpace(in.CreateBy),
		CreateTime:           time.Now(),
		UpdateBy:             strings.TrimSpace(in.CreateBy),
	}
	if row.CreateBy == "" {
		row.CreateBy = "system"
		row.UpdateBy = row.CreateBy
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			logc.Errorf(l.ctx, "创建渠道集成模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("创建渠道集成模板失败")
		}
		if err := syncTemplateBindings(l.ctx, tx, *row, row.UpdateBy); err != nil {
			return err
		}
		if err := recordTemplateOperateLog(l.ctx, tx, *row, "create", row.UpdateBy, fmt.Sprintf("template=%s,status=%s", row.TemplateCode, row.Status)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &sysclient.CreateChannelIntegrationTemplateResp{Id: row.ID}, nil
}
