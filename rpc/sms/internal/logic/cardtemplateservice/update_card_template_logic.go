package cardtemplateservice

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateCardTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCardTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCardTemplateLogic {
	return &UpdateCardTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCardTemplate 更新卡片模板
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09
*/
func (l *UpdateCardTemplateLogic) UpdateCardTemplate(in *smsclient.UpdateCardTemplateReq) (*smsclient.UpdateCardTemplateResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID无效")
	}

	platformId, tenantId, merchantId := resolveScope(in.Scope)

	// 1. 检查模板是否存在且在 scope 内
	var existing cardTemplateRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Select(cardTemplateSelectColumns).
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Take(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片模板不存在")
		}
		logc.Errorf(l.ctx, "查询卡片模板失败: %v", err)
		return nil, errors.New("查询卡片模板失败")
	}

	// 2. 增量构建 updates。注意 template_code 不允许修改（避免破坏发卡规则引用）。
	now := time.Now()
	updates := map[string]interface{}{
		"update_time": now,
	}
	if name := strings.TrimSpace(in.TemplateName); name != "" {
		updates["template_name"] = name
	}
	if v := strings.TrimSpace(in.CardFaceImage); v != "" {
		updates["card_face_image"] = v
	}
	if v := strings.TrimSpace(in.CopyrightOwner); v != "" {
		updates["copyright_owner"] = v
	}
	if v := strings.TrimSpace(in.CopyrightProofSummary); v != "" {
		updates["copyright_proof_summary"] = v
	}
	if v := strings.TrimSpace(in.Rarity); v != "" {
		updates["rarity"] = v
	}
	// Story 10.10 修复 M1: 允许把发行上限改为 0（不限）。admin-api 层 default=-1 表示"不更新"。
	if in.IssueLimit >= 0 {
		updates["issue_limit"] = in.IssueLimit
	}
	if v := strings.TrimSpace(in.DisplayCopy); v != "" {
		updates["display_copy"] = v
	}
	if v := strings.TrimSpace(in.CirculationLimitSummary); v != "" {
		updates["circulation_limit_summary"] = v
	}
	if in.DisplayStatus == 0 || in.DisplayStatus == 1 {
		updates["display_status"] = in.DisplayStatus
	}
	if in.ContentAuditStatus >= 0 && in.ContentAuditStatus <= 3 {
		updates["content_audit_status"] = in.ContentAuditStatus
	}
	if v := strings.TrimSpace(in.ProviderCode); v != "" {
		updates["provider_code"] = v
	}
	if v := strings.TrimSpace(in.CredentialRef); v != "" {
		updates["credential_ref"] = v
	}

	// 3. 执行 UPDATE
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Updates(updates).Error; err != nil {
		logc.Errorf(l.ctx, "更新卡片模板失败: %v", err)
		return nil, errors.New("更新卡片模板失败")
	}

	// 4. 重新查最新数据返回
	var updated cardTemplateRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Select(cardTemplateSelectColumns).
		Where("id = ?", in.Id).
		Take(&updated).Error; err != nil {
		logc.Errorf(l.ctx, "查询更新后的模板失败: %v", err)
		return nil, errors.New("查询更新后的模板失败")
	}

	return &smsclient.UpdateCardTemplateResp{Template: updated.toProto(0)}, nil
}
