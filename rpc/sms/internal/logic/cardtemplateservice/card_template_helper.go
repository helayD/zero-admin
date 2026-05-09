package cardtemplateservice

import (
	"fmt"

	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

// resolveScope 解析治理范围，返回 (platformId, tenantId, merchantId)。
// 默认平台范围：platform_id=1, tenant_id=0, merchant_id=0。
func resolveScope(scope *smsclient.GovernanceScope) (int64, int64, int64) {
	platformId := int64(1)
	tenantId := int64(0)
	merchantId := int64(0)
	if scope != nil {
		platformId = scope.PlatformId
		tenantId = scope.TenantId
		merchantId = scope.MerchantId
	}
	return platformId, tenantId, merchantId
}

// cardTemplateRow 是查询 sms_card_template 表全字段的中间结构。
type cardTemplateRow struct {
	Id                      int64  `gorm:"column:id"`
	PlatformId              int64  `gorm:"column:platform_id"`
	TenantId                int64  `gorm:"column:tenant_id"`
	MerchantId              int64  `gorm:"column:merchant_id"`
	TemplateCode            string `gorm:"column:template_code"`
	TemplateName            string `gorm:"column:template_name"`
	CardFaceImage           string `gorm:"column:card_face_image"`
	CopyrightOwner          string `gorm:"column:copyright_owner"`
	CopyrightProofSummary   string `gorm:"column:copyright_proof_summary"`
	Rarity                  string `gorm:"column:rarity"`
	IssueLimit              int64  `gorm:"column:issue_limit"`
	DisplayCopy             string `gorm:"column:display_copy"`
	CirculationLimitSummary string `gorm:"column:circulation_limit_summary"`
	DisplayStatus           int32  `gorm:"column:display_status"`
	ContentAuditStatus      int32  `gorm:"column:content_audit_status"`
	ProviderCode            string `gorm:"column:provider_code"`
	CredentialRef           string `gorm:"column:credential_ref"`
	Status                  int32  `gorm:"column:status"`
	AuditStatus             int32  `gorm:"column:audit_status"`
	CreateBy                int64  `gorm:"column:create_by"`
	CreateTime              string `gorm:"column:create_time"`
	UpdateBy                int64  `gorm:"column:update_by"`
	UpdateTime              string `gorm:"column:update_time"`
}

// cardTemplateSelectColumns 是查询 sms_card_template 主字段的 SELECT 列表。
const cardTemplateSelectColumns = "id, platform_id, tenant_id, merchant_id, template_code, template_name, " +
	"card_face_image, copyright_owner, copyright_proof_summary, rarity, issue_limit, display_copy, " +
	"circulation_limit_summary, display_status, content_audit_status, provider_code, credential_ref, " +
	"status, audit_status, create_by, create_time, update_by, update_time"

// toProto 把 row 转成 smsclient.CardTemplateData。
func (r *cardTemplateRow) toProto(refRuleCount int32) *smsclient.CardTemplateData {
	return &smsclient.CardTemplateData{
		Id:                      r.Id,
		TemplateCode:            r.TemplateCode,
		TemplateName:            r.TemplateName,
		CardFaceImage:           r.CardFaceImage,
		CopyrightOwner:          r.CopyrightOwner,
		CopyrightProofSummary:   r.CopyrightProofSummary,
		Rarity:                  r.Rarity,
		IssueLimit:              r.IssueLimit,
		DisplayCopy:             r.DisplayCopy,
		CirculationLimitSummary: r.CirculationLimitSummary,
		DisplayStatus:           r.DisplayStatus,
		ContentAuditStatus:      r.ContentAuditStatus,
		ProviderCode:            r.ProviderCode,
		CredentialRef:           r.CredentialRef,
		Status:                  r.Status,
		AuditStatus:             r.AuditStatus,
		PlatformId:              r.PlatformId,
		TenantId:                r.TenantId,
		MerchantId:              r.MerchantId,
		CreateBy:                int64ToString(r.CreateBy),
		UpdateBy:                int64ToString(r.UpdateBy),
		CreateTime:              r.CreateTime,
		UpdateTime:              r.UpdateTime,
		RefRuleCount:            refRuleCount,
	}
}

func int64ToString(n int64) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}
