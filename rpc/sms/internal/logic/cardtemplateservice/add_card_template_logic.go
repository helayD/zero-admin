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
)

type AddCardTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCardTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCardTemplateLogic {
	return &AddCardTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddCardTemplate 新增卡片模板
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09
*/
func (l *AddCardTemplateLogic) AddCardTemplate(in *smsclient.AddCardTemplateReq) (*smsclient.AddCardTemplateResp, error) {
	// 1. 校验必填字段
	templateCode := strings.TrimSpace(in.TemplateCode)
	if templateCode == "" {
		return nil, errors.New("模板编码不能为空")
	}
	templateName := strings.TrimSpace(in.TemplateName)
	if templateName == "" {
		return nil, errors.New("模板名称不能为空")
	}

	// 2. 解析治理范围
	platformId, tenantId, merchantId := resolveScope(in.Scope)

	// 3. 模板编码唯一性校验（同 scope 下）
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND template_code = ? AND is_deleted = 0",
			platformId, tenantId, merchantId, templateCode).
		Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "检查卡片模板编码重复失败: %v", err)
		return nil, errors.New("检查模板编码失败")
	}
	if count > 0 {
		return nil, errors.New("模板编码已存在")
	}

	// 4. 默认值兜底
	displayStatus := in.DisplayStatus
	if displayStatus < 0 || displayStatus > 1 {
		displayStatus = 1
	}
	contentAuditStatus := in.ContentAuditStatus
	if contentAuditStatus < 0 {
		contentAuditStatus = 0
	}

	// 5. 写入
	now := time.Now()
	record := map[string]interface{}{
		"platform_id":               platformId,
		"tenant_id":                 tenantId,
		"merchant_id":               merchantId,
		"template_code":             templateCode,
		"template_name":             templateName,
		"card_face_image":           strings.TrimSpace(in.CardFaceImage),
		"copyright_owner":           strings.TrimSpace(in.CopyrightOwner),
		"copyright_proof_summary":   strings.TrimSpace(in.CopyrightProofSummary),
		"rarity":                    strings.TrimSpace(in.Rarity),
		"issue_limit":               in.IssueLimit,
		"display_copy":              strings.TrimSpace(in.DisplayCopy),
		"circulation_limit_summary": strings.TrimSpace(in.CirculationLimitSummary),
		"display_status":            displayStatus,
		"content_audit_status":      contentAuditStatus,
		"provider_code":             strings.TrimSpace(in.ProviderCode),
		"credential_ref":            strings.TrimSpace(in.CredentialRef),
		"status":                    1, // 默认启用
		"audit_status":              0,
		"create_by":                 0, // TODO: 从上下文获取操作人ID
		"create_time":               now,
		"is_deleted":                0,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Create(record).Error; err != nil {
		logc.Errorf(l.ctx, "插入卡片模板失败: %v", err)
		return nil, errors.New("创建卡片模板失败")
	}

	// 6. 取回新插入的 ID
	var inserted struct {
		Id int64 `gorm:"column:id"`
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Select("id").
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND template_code = ? AND is_deleted = 0",
			platformId, tenantId, merchantId, templateCode).
		Order("id DESC").
		Take(&inserted).Error; err != nil {
		logc.Errorf(l.ctx, "获取新创建的卡片模板 ID 失败: %v", err)
		return nil, errors.New("获取创建的模板失败")
	}

	return &smsclient.AddCardTemplateResp{
		Template: &smsclient.CardTemplateData{
			Id:                      inserted.Id,
			TemplateCode:            templateCode,
			TemplateName:            templateName,
			CardFaceImage:           in.CardFaceImage,
			CopyrightOwner:          in.CopyrightOwner,
			CopyrightProofSummary:   in.CopyrightProofSummary,
			Rarity:                  in.Rarity,
			IssueLimit:              in.IssueLimit,
			DisplayCopy:             in.DisplayCopy,
			CirculationLimitSummary: in.CirculationLimitSummary,
			DisplayStatus:           displayStatus,
			ContentAuditStatus:      contentAuditStatus,
			ProviderCode:            in.ProviderCode,
			CredentialRef:           in.CredentialRef,
			Status:                  1,
			PlatformId:              platformId,
			TenantId:                tenantId,
			MerchantId:              merchantId,
			CreateTime:              now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
