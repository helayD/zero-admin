package channelintegrationtemplateservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChannelIntegrationTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChannelIntegrationTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChannelIntegrationTemplateLogic {
	return &UpdateChannelIntegrationTemplateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateChannelIntegrationTemplateLogic) UpdateChannelIntegrationTemplate(in *sysclient.UpdateChannelIntegrationTemplateReq) (*sysclient.UpdateChannelIntegrationTemplateResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID不能为空")
	}

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

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var existing channelIntegrationTemplateRow
		if err := tx.Where("id = ?", in.Id).Take(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("模板不存在")
			}
			logc.Errorf(l.ctx, "查询待更新模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新渠道集成模板失败")
		}
		if err := channeltemplate.ValidateStatusTransition(existing.Status, payload.Status); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"template_code":          payload.TemplateCode,
			"template_name":          payload.TemplateName,
			"template_type":          payload.TemplateType,
			"target_code":            payload.TargetCode,
			"scope_type":             payload.ScopeType,
			"platform_id":            payload.PlatformID,
			"tenant_id":              payload.TenantID,
			"merchant_id":            payload.MerchantID,
			"status":                 payload.Status,
			"metadata_config":        payload.MetadataConfig,
			"secret_ref_config":      payload.SecretRefConfig,
			"intent_contract_config": payload.IntentContractConfig,
			"impact_scope_config":    payload.ImpactScopeConfig,
			"remark":                 payload.Remark,
			"update_by":              strings.TrimSpace(in.UpdateBy),
			"update_time":            time.Now(),
		}
		if updates["update_by"] == "" {
			updates["update_by"] = "system"
		}
		if err := tx.Model(&channelIntegrationTemplateRow{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
			logc.Errorf(l.ctx, "更新渠道集成模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新渠道集成模板失败")
		}
		updatedRow := existing
		updatedRow.TemplateCode = payload.TemplateCode
		updatedRow.TemplateName = payload.TemplateName
		updatedRow.TemplateType = payload.TemplateType
		updatedRow.TargetCode = payload.TargetCode
		updatedRow.ScopeType = payload.ScopeType
		updatedRow.PlatformID = payload.PlatformID
		updatedRow.TenantID = payload.TenantID
		updatedRow.MerchantID = payload.MerchantID
		updatedRow.Status = payload.Status
		updatedRow.MetadataConfig = payload.MetadataConfig
		updatedRow.SecretRefConfig = payload.SecretRefConfig
		updatedRow.IntentContractConfig = payload.IntentContractConfig
		updatedRow.ImpactScopeConfig = payload.ImpactScopeConfig
		updatedRow.Remark = payload.Remark
		updatedRow.UpdateBy = fmt.Sprint(updates["update_by"])
		updateTime := updates["update_time"].(time.Time)
		updatedRow.UpdateTime = &updateTime
		if err := syncTemplateBindings(l.ctx, tx, updatedRow, updatedRow.UpdateBy); err != nil {
			return err
		}
		if err := recordTemplateOperateLog(l.ctx, tx, updatedRow, "update", updatedRow.UpdateBy, fmt.Sprintf("template=%s,status=%s", updatedRow.TemplateCode, updatedRow.Status)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.UpdateChannelIntegrationTemplateResp{Pong: "ok"}, nil
}
