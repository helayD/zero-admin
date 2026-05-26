package drawactivityservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/gorm"
)

const (
	drawStatusDraft     int32 = 0
	drawStatusPublished int32 = 1
	drawStatusArchived  int32 = 2

	drawReadinessMissing int32 = 0
	drawReadinessReady   int32 = 1
	drawReadinessPending int32 = 2
	drawReadinessOffline int32 = 3

	drawApprovalMissing int32 = 0
	drawApprovalPending int32 = 1
	drawApprovalPassed  int32 = 2
	drawApprovalReject  int32 = 3

	drawAssetMissing int32 = 0
	drawAssetPending int32 = 1
	drawAssetPassed  int32 = 2
	drawAssetReject  int32 = 3
)

type drawActivityRow struct {
	ID                          int64      `gorm:"column:id"`
	PlatformID                  int64      `gorm:"column:platform_id"`
	TenantID                    int64      `gorm:"column:tenant_id"`
	MerchantID                  int64      `gorm:"column:merchant_id"`
	ActivityCode                string     `gorm:"column:activity_code"`
	Name                        string     `gorm:"column:name"`
	RuleSummary                 string     `gorm:"column:rule_summary"`
	StartTime                   time.Time  `gorm:"column:start_time"`
	EndTime                     time.Time  `gorm:"column:end_time"`
	RealNameRequired            int32      `gorm:"column:real_name_required"`
	ParticipantConditionSummary string     `gorm:"column:participant_condition_summary"`
	ConsumeRuleSummary          string     `gorm:"column:consume_rule_summary"`
	ProbabilityRule             string     `gorm:"column:probability_rule"`
	ComplianceRuleSummary       string     `gorm:"column:compliance_rule_summary"`
	CirculationLimitSummary     string     `gorm:"column:circulation_limit_summary"`
	ApprovalRecordRef           string     `gorm:"column:approval_record_ref"`
	PublishFailureSummary       string     `gorm:"column:publish_failure_summary"`
	PublishReadiness            int32      `gorm:"column:publish_readiness"`
	CopyrightStatus             int32      `gorm:"column:copyright_status"`
	ContentAuditStatus          int32      `gorm:"column:content_audit_status"`
	Status                      int32      `gorm:"column:status"`
	AuditStatus                 int32      `gorm:"column:audit_status"`
	ConsumeType                 string     `gorm:"column:consume_type"`
	ConsumeAmount               int32      `gorm:"column:consume_amount"`
	QuotaPerMember              int32      `gorm:"column:quota_per_member"`
	DailyQuotaPerMember         int32      `gorm:"column:daily_quota_per_member"`
	IsEnabled                   int32      `gorm:"column:is_enabled"`
	ShowOnHome                  int32      `gorm:"column:show_on_home"`
	HomeEntryTitle              string     `gorm:"column:home_entry_title"`
	HomeEntrySubtitle           string     `gorm:"column:home_entry_subtitle"`
	HomeEntryImage              string     `gorm:"column:home_entry_image"`
	HomeEntrySort               int32      `gorm:"column:home_entry_sort"`
	HomeEntryEnabled            int32      `gorm:"column:home_entry_enabled"`
	LandingTargetType           string     `gorm:"column:landing_target_type"`
	LandingTargetValue          string     `gorm:"column:landing_target_value"`
	CreateBy                    int64      `gorm:"column:create_by"`
	CreateTime                  time.Time  `gorm:"column:create_time"`
	UpdateBy                    *int64     `gorm:"column:update_by"`
	UpdateTime                  *time.Time `gorm:"column:update_time"`
	IsDeleted                   int32      `gorm:"column:is_deleted"`
}

func (drawActivityRow) TableName() string {
	return "sms_draw_activity"
}

type drawPoolRow struct {
	ID              int64      `gorm:"column:id"`
	ActivityID      int64      `gorm:"column:activity_id"`
	PlatformID      int64      `gorm:"column:platform_id"`
	TenantID        int64      `gorm:"column:tenant_id"`
	MerchantID      int64      `gorm:"column:merchant_id"`
	PoolCode        string     `gorm:"column:pool_code"`
	PoolName        string     `gorm:"column:pool_name"`
	ProbabilityRule string     `gorm:"column:probability_rule"`
	WheelSlotCount  int32      `gorm:"column:wheel_slot_count"`
	Sort            int32      `gorm:"column:sort"`
	Status          int32      `gorm:"column:status"`
	AuditStatus     int32      `gorm:"column:audit_status"`
	CreateBy        int64      `gorm:"column:create_by"`
	CreateTime      time.Time  `gorm:"column:create_time"`
	UpdateBy        *int64     `gorm:"column:update_by"`
	UpdateTime      *time.Time `gorm:"column:update_time"`
	IsDeleted       int32      `gorm:"column:is_deleted"`
}

func (drawPoolRow) TableName() string {
	return "sms_draw_pool"
}

type drawCardTemplateRow struct {
	ID                      int64      `gorm:"column:id"`
	PlatformID              int64      `gorm:"column:platform_id"`
	TenantID                int64      `gorm:"column:tenant_id"`
	MerchantID              int64      `gorm:"column:merchant_id"`
	TemplateCode            string     `gorm:"column:template_code"`
	TemplateName            string     `gorm:"column:template_name"`
	CardFaceImage           string     `gorm:"column:card_face_image"`
	CopyrightOwner          string     `gorm:"column:copyright_owner"`
	CopyrightProofSummary   string     `gorm:"column:copyright_proof_summary"`
	Rarity                  string     `gorm:"column:rarity"`
	IssueLimit              int64      `gorm:"column:issue_limit"`
	DisplayCopy             string     `gorm:"column:display_copy"`
	CirculationLimitSummary string     `gorm:"column:circulation_limit_summary"`
	DisplayStatus           int32      `gorm:"column:display_status"`
	ContentAuditStatus      int32      `gorm:"column:content_audit_status"`
	ProviderCode            string     `gorm:"column:provider_code"`
	CredentialRef           string     `gorm:"column:credential_ref"`
	Status                  int32      `gorm:"column:status"`
	AuditStatus             int32      `gorm:"column:audit_status"`
	CreateBy                int64      `gorm:"column:create_by"`
	CreateTime              time.Time  `gorm:"column:create_time"`
	UpdateBy                *int64     `gorm:"column:update_by"`
	UpdateTime              *time.Time `gorm:"column:update_time"`
	IsDeleted               int32      `gorm:"column:is_deleted"`
}

func (drawCardTemplateRow) TableName() string {
	return "sms_card_template"
}

type drawPoolTemplateRow struct {
	ID             int64      `gorm:"column:id"`
	ActivityID     int64      `gorm:"column:activity_id"`
	PoolID         int64      `gorm:"column:pool_id"`
	TemplateID     int64      `gorm:"column:template_id"`
	SlotIndex      int32      `gorm:"column:slot_index"`
	PlatformID     int64      `gorm:"column:platform_id"`
	TenantID       int64      `gorm:"column:tenant_id"`
	MerchantID     int64      `gorm:"column:merchant_id"`
	Rarity         string     `gorm:"column:rarity"`
	Probability    float64    `gorm:"column:probability"`
	SaleLimit      int64      `gorm:"column:sale_limit"`
	RemainingLimit int64      `gorm:"column:remaining_limit"`
	ConfigLimit    int64      `gorm:"column:config_limit"`
	Status         int32      `gorm:"column:status"`
	AuditStatus    int32      `gorm:"column:audit_status"`
	CreateBy       int64      `gorm:"column:create_by"`
	CreateTime     time.Time  `gorm:"column:create_time"`
	UpdateBy       *int64     `gorm:"column:update_by"`
	UpdateTime     *time.Time `gorm:"column:update_time"`
	IsDeleted      int32      `gorm:"column:is_deleted"`
}

func (drawPoolTemplateRow) TableName() string {
	return "sms_draw_pool_template"
}

type drawActivityAuditRow struct {
	ID                       int64     `gorm:"column:id"`
	ActivityID               int64     `gorm:"column:activity_id"`
	PlatformID               int64     `gorm:"column:platform_id"`
	TenantID                 int64     `gorm:"column:tenant_id"`
	MerchantID               int64     `gorm:"column:merchant_id"`
	OperationType            string    `gorm:"column:operation_type"`
	OperatorID               int64     `gorm:"column:operator_id"`
	OperatorName             string    `gorm:"column:operator_name"`
	ApprovalResult           string    `gorm:"column:approval_result"`
	RuleSnapshot             string    `gorm:"column:rule_snapshot"`
	CirculationLimitSnapshot string    `gorm:"column:circulation_limit_snapshot"`
	FailureSummary           string    `gorm:"column:failure_summary"`
	TraceID                  string    `gorm:"column:trace_id"`
	PayloadJSON              string    `gorm:"column:payload_json"`
	Status                   int32     `gorm:"column:status"`
	AuditStatus              int32     `gorm:"column:audit_status"`
	CreateBy                 int64     `gorm:"column:create_by"`
	CreateTime               time.Time `gorm:"column:create_time"`
}

func (drawActivityAuditRow) TableName() string {
	return "sms_draw_activity_audit"
}

type drawPoolTemplateJoinRow struct {
	PoolID                  int64   `gorm:"column:pool_id"`
	PoolTemplateID          int64   `gorm:"column:pool_template_id"`
	TemplateID              int64   `gorm:"column:template_id"`
	SlotIndex               int32   `gorm:"column:slot_index"`
	TemplateCode            string  `gorm:"column:template_code"`
	TemplateName            string  `gorm:"column:template_name"`
	Rarity                  string  `gorm:"column:rarity"`
	Probability             float64 `gorm:"column:probability"`
	SaleLimit               int64   `gorm:"column:sale_limit"`
	RemainingLimit          int64   `gorm:"column:remaining_limit"`
	ConfigLimit             int64   `gorm:"column:config_limit"`
	CardFaceImage           string  `gorm:"column:card_face_image"`
	CopyrightOwner          string  `gorm:"column:copyright_owner"`
	CopyrightProofSummary   string  `gorm:"column:copyright_proof_summary"`
	DisplayCopy             string  `gorm:"column:display_copy"`
	CirculationLimitSummary string  `gorm:"column:circulation_limit_summary"`
	DisplayStatus           int32   `gorm:"column:display_status"`
	ContentAuditStatus      int32   `gorm:"column:content_audit_status"`
	ProviderCode            string  `gorm:"column:provider_code"`
	CredentialRef           string  `gorm:"column:credential_ref"`
	IssueLimit              int64   `gorm:"column:issue_limit"`
	TemplateStatus          int32   `gorm:"column:template_status"`
	TemplateAuditStatus     int32   `gorm:"column:template_audit_status"`
}

type drawActivityAggregate struct {
	Activity  drawActivityRow
	Templates []*smsclient.DrawCardTemplateData
	Pools     []*smsclient.DrawPoolData
	Audits    []*smsclient.DrawActivityAuditRecord
	Readiness *drawReadinessResult
}

type drawReadinessResult struct {
	ReadyToPublish   bool
	PublishReadiness int32
	Label            string
	Summary          string
	Items            []*smsclient.DrawReadinessItem
}

type drawAuditPayload struct {
	Activity  *drawActivityRow                  `json:"activity"`
	Templates []*smsclient.DrawCardTemplateData `json:"templates"`
	Pools     []*smsclient.DrawPoolData         `json:"pools"`
}

func parseDrawDateTime(raw, field string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, fmt.Errorf("%s格式无效，请使用 yyyy-MM-dd HH:mm:ss 格式", field)
	}
	return parsed, nil
}

func validateDrawAggregate(activity *drawActivityRow, templates []*smsclient.DrawCardTemplateData, pools []*smsclient.DrawPoolData) error {
	if strings.TrimSpace(activity.Name) == "" {
		return errors.New("活动名称不能为空")
	}
	if strings.TrimSpace(activity.ActivityCode) == "" {
		return errors.New("活动编码不能为空")
	}
	if !activity.StartTime.Before(activity.EndTime) {
		return errors.New("开始时间必须早于结束时间")
	}
	if len(templates) == 0 {
		return errors.New("至少需要配置一个卡片模板")
	}
	if len(pools) == 0 {
		return errors.New("至少需要配置一个卡池")
	}

	templateCodeSeen := make(map[string]struct{}, len(templates))
	for _, item := range templates {
		if strings.TrimSpace(item.TemplateCode) == "" {
			return errors.New("模板编码不能为空")
		}
		if strings.TrimSpace(item.TemplateName) == "" {
			return errors.New("模板名称不能为空")
		}
		if item.IssueLimit <= 0 {
			return fmt.Errorf("模板[%s]发行上限必须大于0", item.TemplateName)
		}
		code := strings.TrimSpace(item.TemplateCode)
		if _, ok := templateCodeSeen[code]; ok {
			return fmt.Errorf("模板编码[%s]重复", code)
		}
		templateCodeSeen[code] = struct{}{}
	}

	poolCodeSeen := make(map[string]struct{}, len(pools))
	for _, pool := range pools {
		if strings.TrimSpace(pool.PoolName) == "" {
			return errors.New("卡池名称不能为空")
		}
		if strings.TrimSpace(pool.PoolCode) == "" {
			return fmt.Errorf("卡池[%s]编码不能为空", pool.PoolName)
		}
		if _, ok := poolCodeSeen[strings.TrimSpace(pool.PoolCode)]; ok {
			return fmt.Errorf("卡池编码[%s]重复", pool.PoolCode)
		}
		poolCodeSeen[strings.TrimSpace(pool.PoolCode)] = struct{}{}
		if len(pool.Templates) == 0 {
			return fmt.Errorf("卡池[%s]至少需要配置一个模板", pool.PoolName)
		}
		var probabilitySum float64
		slotIndexSeen := make(map[int32]struct{}, len(pool.Templates))
		for _, mapping := range pool.Templates {
			if strings.TrimSpace(mapping.TemplateCode) == "" && mapping.TemplateId <= 0 {
				return fmt.Errorf("卡池[%s]存在未绑定模板的概率配置", pool.PoolName)
			}
			if mapping.Probability <= 0 || mapping.Probability > 1 {
				return fmt.Errorf("卡池[%s]的模板概率必须在(0,1]之间", pool.PoolName)
			}
			if mapping.SaleLimit <= 0 || mapping.ConfigLimit < 0 {
				return fmt.Errorf("卡池[%s]的模板数量配置非法", pool.PoolName)
			}
			if mapping.SlotIndex < 1 || mapping.SlotIndex > 5 {
				return fmt.Errorf("卡池[%s]格位序号必须在 1-5 之间，当前値: %d", pool.PoolName, mapping.SlotIndex)
			}
			if _, ok := slotIndexSeen[mapping.SlotIndex]; ok {
				return fmt.Errorf("卡池[%s]格位序号 %d 重复", pool.PoolName, mapping.SlotIndex)
			}
			slotIndexSeen[mapping.SlotIndex] = struct{}{}
			probabilitySum += mapping.Probability
		}
		if len(slotIndexSeen) != 5 {
			return fmt.Errorf("卡池[%s]必须配置5个格位（slot_index 1-5全部存在），当前配置了 %d 个", pool.PoolName, len(slotIndexSeen))
		}
		for idx := int32(1); idx <= 5; idx++ {
			if _, ok := slotIndexSeen[idx]; !ok {
				return fmt.Errorf("卡池[%s]格位序号 %d 缺失，slot_index 1-5 必须连续且全部存在", pool.PoolName, idx)
			}
		}
		if math.Abs(probabilitySum-1) > 0.0001 {
			return fmt.Errorf("卡池[%s]概率总和必须为1", pool.PoolName)
		}
	}

	return nil
}

func buildDrawReadiness(activity *drawActivityRow, templates []*smsclient.DrawCardTemplateData, pools []*smsclient.DrawPoolData, auditCount int64) *drawReadinessResult {
	result := &drawReadinessResult{
		ReadyToPublish:   true,
		PublishReadiness: drawReadinessReady,
		Label:            drawReadinessLabel(drawReadinessReady),
	}

	addItem := func(code, field, message string, blocking bool) {
		result.Items = append(result.Items, &smsclient.DrawReadinessItem{
			Code:     code,
			Field:    field,
			Message:  message,
			Blocking: blocking,
		})
	}

	if activity.IsEnabled != 1 {
		addItem("activityDisabled", "isEnabled", "活动处于停用状态，请先启用后再发布", true)
	}
	if strings.TrimSpace(activity.RuleSummary) == "" {
		addItem("missingRuleSummary", "ruleSummary", "活动规则摘要缺失", true)
	}
	if strings.TrimSpace(activity.ParticipantConditionSummary) == "" {
		addItem("missingParticipantRule", "participantConditionSummary", "参与条件摘要缺失", true)
	}
	if strings.TrimSpace(activity.ConsumeRuleSummary) == "" {
		addItem("missingConsumeRule", "consumeRuleSummary", "消耗规则摘要缺失", true)
	}
	if strings.TrimSpace(activity.ProbabilityRule) == "" {
		addItem("invalidProbabilityRule", "probabilityRule", "概率披露方式缺失", true)
	}
	if strings.TrimSpace(activity.ComplianceRuleSummary) == "" {
		addItem("missingComplianceRule", "complianceRuleSummary", "合规规则摘要缺失", true)
	}
	if !containsRequiredCirculationRule(activity.CirculationLimitSummary) {
		addItem("missingCirculationLimit", "circulationLimitSummary", "流转限制必须明确禁止集中竞价、连续挂牌和收益承诺", true)
	}
	if activity.CopyrightStatus == drawAssetMissing || activity.CopyrightStatus == drawAssetReject {
		addItem("missingCopyright", "copyrightStatus", "版权校验未通过，请补齐版权信息", true)
	} else if activity.CopyrightStatus == drawAssetPending {
		addItem("pendingCopyright", "copyrightStatus", "版权校验仍在处理中", false)
	}
	if activity.ContentAuditStatus == drawAssetMissing || activity.ContentAuditStatus == drawAssetReject {
		addItem("missingContentAudit", "contentAuditStatus", "内容审核未通过，请补齐审核结果", true)
	} else if activity.ContentAuditStatus == drawAssetPending {
		addItem("pendingContentAudit", "contentAuditStatus", "内容审核仍在处理中", false)
	}
	if activity.AuditStatus == drawApprovalMissing || strings.TrimSpace(activity.ApprovalRecordRef) == "" || auditCount == 0 {
		addItem("missingAuditRecord", "approvalRecordRef", "审批记录或审计快照不完整", true)
	} else if activity.AuditStatus == drawApprovalPending {
		addItem("pendingApproval", "auditStatus", "活动审批仍在处理中", false)
	} else if activity.AuditStatus == drawApprovalReject {
		addItem("approvalRejected", "auditStatus", "活动审批未通过", true)
	}
	if activity.ShowOnHome == 1 {
		if strings.TrimSpace(activity.HomeEntryTitle) == "" || strings.TrimSpace(activity.LandingTargetType) == "" {
			addItem("missingHomeEntry", "homeEntry", "首页显著入口配置不完整（需填写入口标题和落地页类型）", true)
		}
	}

	if len(templates) == 0 {
		addItem("missingTemplate", "templates", "至少需要配置一个卡片模板", true)
	}
	for _, item := range templates {
		if strings.TrimSpace(item.CopyrightOwner) == "" || strings.TrimSpace(item.CopyrightProofSummary) == "" {
			addItem("missingTemplateCopyright", "templates", fmt.Sprintf("模板[%s]缺少版权归属或凭证摘要", item.TemplateName), true)
		}
		if item.ContentAuditStatus == drawAssetMissing || item.ContentAuditStatus == drawAssetReject {
			addItem("templateAuditRejected", "templates", fmt.Sprintf("模板[%s]内容审核未通过", item.TemplateName), true)
		} else if item.ContentAuditStatus == drawAssetPending {
			addItem("templateAuditPending", "templates", fmt.Sprintf("模板[%s]内容审核仍在处理中", item.TemplateName), false)
		}
		if !containsRequiredCirculationRule(item.CirculationLimitSummary) {
			addItem("templateCirculationLimit", "templates", fmt.Sprintf("模板[%s]默认流转限制不完整", item.TemplateName), true)
		}
	}

	if len(pools) == 0 {
		addItem("missingPool", "pools", "至少需要配置一个卡池", true)
	}
	for _, pool := range pools {
		var sum float64
		if len(pool.Templates) == 0 {
			addItem("missingPoolTemplate", "pools", fmt.Sprintf("卡池[%s]缺少模板映射", pool.PoolName), true)
			continue
		}
		for _, mapping := range pool.Templates {
			sum += mapping.Probability
			if mapping.SaleLimit <= 0 || mapping.ConfigLimit < 0 {
				addItem("invalidSaleLimit", "pools", fmt.Sprintf("卡池[%s]存在非法数量配置", pool.PoolName), true)
			}
		}
		if math.Abs(sum-1) > 0.0001 {
			addItem("invalidProbabilityRule", "pools", fmt.Sprintf("卡池[%s]概率总和必须为1", pool.PoolName), true)
		}
		if len(pool.Templates) != 5 {
			addItem("invalidSlotCount", "pools", fmt.Sprintf("卡池[%s]必须配置恰好5个格位，当前: %d", pool.PoolName, len(pool.Templates)), true)
		}
	}

	if len(result.Items) == 0 {
		result.Summary = "发布预检通过"
		return result
	}

	result.ReadyToPublish = false
	blocking := false
	for _, item := range result.Items {
		if item.Blocking {
			blocking = true
			break
		}
	}
	if blocking {
		result.PublishReadiness = drawReadinessMissing
	} else {
		result.PublishReadiness = drawReadinessPending
	}
	result.Label = drawReadinessLabel(result.PublishReadiness)

	summaryParts := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		summaryParts = append(summaryParts, item.Message)
		if len(summaryParts) == 3 {
			break
		}
	}
	result.Summary = strings.Join(summaryParts, "；")
	return result
}

func containsRequiredCirculationRule(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "禁止集中竞价") &&
		strings.Contains(trimmed, "禁止连续挂牌") &&
		strings.Contains(trimmed, "禁止收益承诺")
}

func drawReadinessLabel(status int32) string {
	switch status {
	case drawReadinessReady:
		return "可发布"
	case drawReadinessPending:
		return "待审批"
	case drawReadinessOffline:
		return "已下线"
	default:
		return "缺少信息"
	}
}

func activityScopeType(platformID, tenantID, merchantID int64) string {
	scope, err := pkgscope.NormalizeGovernanceScope("", platformID, tenantID, merchantID)
	if err != nil {
		return pkgscope.SubjectTypePlatform
	}
	return scope.ScopeType
}

func approvalResultLabel(status int32) string {
	switch status {
	case drawApprovalPassed:
		return "approved"
	case drawApprovalPending:
		return "pending"
	case drawApprovalReject:
		return "rejected"
	default:
		return "missing"
	}
}

func buildDrawHomeEntry(activity drawActivityRow) *smsclient.DrawHomeEntryConfig {
	return &smsclient.DrawHomeEntryConfig{
		ShowOnHome:         activity.ShowOnHome,
		HomeEntryTitle:     activity.HomeEntryTitle,
		HomeEntrySubtitle:  activity.HomeEntrySubtitle,
		HomeEntryImage:     activity.HomeEntryImage,
		HomeEntrySort:      activity.HomeEntrySort,
		IsEnabled:          activity.HomeEntryEnabled,
		LandingTargetType:  activity.LandingTargetType,
		LandingTargetValue: activity.LandingTargetValue,
	}
}

func applyHomeEntry(activity *drawActivityRow, home *smsclient.DrawHomeEntryConfig) {
	if home == nil {
		return
	}
	activity.ShowOnHome = home.ShowOnHome
	activity.HomeEntryTitle = strings.TrimSpace(home.HomeEntryTitle)
	activity.HomeEntrySubtitle = strings.TrimSpace(home.HomeEntrySubtitle)
	activity.HomeEntryImage = strings.TrimSpace(home.HomeEntryImage)
	activity.HomeEntrySort = home.HomeEntrySort
	activity.HomeEntryEnabled = home.IsEnabled
	activity.LandingTargetType = strings.TrimSpace(home.LandingTargetType)
	activity.LandingTargetValue = strings.TrimSpace(home.LandingTargetValue)
}

func loadDrawActivityAggregate(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, id int64) (*drawActivityAggregate, error) {
	var activity drawActivityRow
	if err := pkgscope.ApplyGovernanceScope(
		db.WithContext(ctx).Table(activity.TableName()).Where("is_deleted = 0"),
		current,
		"",
	).Where("id = ?", id).Take(&activity).Error; err != nil {
		return nil, err
	}

	var poolRows []drawPoolRow
	if err := db.WithContext(ctx).Table(drawPoolRow{}.TableName()).
		Where("activity_id = ? AND is_deleted = 0", id).
		Order("sort asc, id asc").
		Find(&poolRows).Error; err != nil {
		return nil, err
	}

	poolMap := make(map[int64]*smsclient.DrawPoolData, len(poolRows))
	pools := make([]*smsclient.DrawPoolData, 0, len(poolRows))
	for _, row := range poolRows {
		item := &smsclient.DrawPoolData{
			Id:              row.ID,
			PoolName:        row.PoolName,
			PoolCode:        row.PoolCode,
			ProbabilityRule: row.ProbabilityRule,
			Sort:            row.Sort,
			Status:          row.Status,
			Templates:       []*smsclient.DrawPoolTemplateData{},
		}
		poolMap[row.ID] = item
		pools = append(pools, item)
	}

	var joinRows []drawPoolTemplateJoinRow
	if len(poolRows) > 0 {
		poolIDs := make([]int64, 0, len(poolRows))
		for _, row := range poolRows {
			poolIDs = append(poolIDs, row.ID)
		}
		err := db.WithContext(ctx).Table("sms_draw_pool_template dpt").
			Select(`
				dpt.pool_id,
				dpt.id AS pool_template_id,
				dpt.template_id,
				dpt.slot_index,
				t.template_code,
				t.template_name,
				COALESCE(dpt.rarity, t.rarity) AS rarity,
				dpt.probability,
				dpt.sale_limit,
				dpt.remaining_limit,
				dpt.config_limit,
				t.card_face_image,
				t.copyright_owner,
				t.copyright_proof_summary,
				t.display_copy,
				t.circulation_limit_summary,
				t.display_status,
				t.content_audit_status,
				t.provider_code,
				t.credential_ref,
				t.issue_limit,
				t.status AS template_status,
				t.audit_status AS template_audit_status`).
			Joins("JOIN sms_card_template t ON t.id = dpt.template_id AND t.is_deleted = 0").
			Where("dpt.pool_id IN ? AND dpt.is_deleted = 0", poolIDs).
			Order("dpt.slot_index asc").
			Find(&joinRows).Error
		if err != nil {
			return nil, err
		}
	}

	templateSeen := make(map[int64]struct{})
	templates := make([]*smsclient.DrawCardTemplateData, 0)
	for _, row := range joinRows {
		if pool, ok := poolMap[row.PoolID]; ok {
			pool.Templates = append(pool.Templates, &smsclient.DrawPoolTemplateData{
				Id:             row.PoolTemplateID,
				TemplateId:     row.TemplateID,
				SlotIndex:      row.SlotIndex,
				TemplateCode:   row.TemplateCode,
				TemplateName:   row.TemplateName,
				Rarity:         row.Rarity,
				Probability:    row.Probability,
				SaleLimit:      row.SaleLimit,
				RemainingLimit: row.RemainingLimit,
				ConfigLimit:    row.ConfigLimit,
			})
		}
		if _, ok := templateSeen[row.TemplateID]; ok {
			continue
		}
		templateSeen[row.TemplateID] = struct{}{}
		templates = append(templates, &smsclient.DrawCardTemplateData{
			Id:                      row.TemplateID,
			TemplateCode:            row.TemplateCode,
			TemplateName:            row.TemplateName,
			CardFaceImage:           row.CardFaceImage,
			CopyrightOwner:          row.CopyrightOwner,
			CopyrightProofSummary:   row.CopyrightProofSummary,
			Rarity:                  row.Rarity,
			IssueLimit:              row.IssueLimit,
			DisplayCopy:             row.DisplayCopy,
			CirculationLimitSummary: row.CirculationLimitSummary,
			DisplayStatus:           row.DisplayStatus,
			ContentAuditStatus:      row.ContentAuditStatus,
			ProviderCode:            row.ProviderCode,
			CredentialRef:           row.CredentialRef,
			Status:                  row.TemplateStatus,
			AuditStatus:             row.TemplateAuditStatus,
		})
	}

	var auditRows []drawActivityAuditRow
	if err := db.WithContext(ctx).Table(drawActivityAuditRow{}.TableName()).
		Where("activity_id = ?", id).
		Order("id desc").
		Limit(20).
		Find(&auditRows).Error; err != nil {
		return nil, err
	}

	audits := make([]*smsclient.DrawActivityAuditRecord, 0, len(auditRows))
	for _, row := range auditRows {
		audits = append(audits, &smsclient.DrawActivityAuditRecord{
			Id:             row.ID,
			OperationType:  row.OperationType,
			OperatorId:     row.OperatorID,
			OperatorName:   row.OperatorName,
			ApprovalResult: row.ApprovalResult,
			FailureSummary: row.FailureSummary,
			TraceId:        row.TraceID,
			CreateTime:     time_util.TimeToStr(row.CreateTime),
		})
	}

	readiness := buildDrawReadiness(&activity, templates, pools, int64(len(auditRows)))
	return &drawActivityAggregate{
		Activity:  activity,
		Templates: templates,
		Pools:     pools,
		Audits:    audits,
		Readiness: readiness,
	}, nil
}

func syncDrawReadiness(ctx context.Context, tx *gorm.DB, activityID int64, readiness *drawReadinessResult) error {
	if readiness == nil {
		return nil
	}
	return tx.WithContext(ctx).
		Table(drawActivityRow{}.TableName()).
		Where("id = ?", activityID).
		Updates(map[string]interface{}{
			"publish_readiness":       readiness.PublishReadiness,
			"publish_failure_summary": readiness.Summary,
		}).Error
}

func upsertDrawTemplates(ctx context.Context, tx *gorm.DB, scope pkgscope.GovernanceScope, templates []*smsclient.DrawCardTemplateData, operatorID int64) (map[string]int64, error) {
	templateIDs := make(map[string]int64, len(templates))
	now := time.Now()
	for _, item := range templates {
		code := strings.TrimSpace(item.TemplateCode)
		var existing drawCardTemplateRow
		err := tx.WithContext(ctx).Table(drawCardTemplateRow{}.TableName()).
			Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND template_code = ? AND is_deleted = 0",
				scope.PlatformID, scope.TenantID, scope.MerchantID, code).
			Take(&existing).Error

		updates := map[string]interface{}{
			"template_name":             strings.TrimSpace(item.TemplateName),
			"card_face_image":           strings.TrimSpace(item.CardFaceImage),
			"copyright_owner":           strings.TrimSpace(item.CopyrightOwner),
			"copyright_proof_summary":   strings.TrimSpace(item.CopyrightProofSummary),
			"rarity":                    strings.TrimSpace(item.Rarity),
			"issue_limit":               item.IssueLimit,
			"display_copy":              strings.TrimSpace(item.DisplayCopy),
			"circulation_limit_summary": strings.TrimSpace(item.CirculationLimitSummary),
			"display_status":            item.DisplayStatus,
			"content_audit_status":      item.ContentAuditStatus,
			"provider_code":             strings.TrimSpace(item.ProviderCode),
			"credential_ref":            strings.TrimSpace(item.CredentialRef),
			"status":                    item.Status,
			"audit_status":              item.AuditStatus,
			"update_by":                 operatorID,
			"update_time":               now,
		}

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			row := drawCardTemplateRow{
				PlatformID:              scope.PlatformID,
				TenantID:                scope.TenantID,
				MerchantID:              scope.MerchantID,
				TemplateCode:            code,
				TemplateName:            strings.TrimSpace(item.TemplateName),
				CardFaceImage:           strings.TrimSpace(item.CardFaceImage),
				CopyrightOwner:          strings.TrimSpace(item.CopyrightOwner),
				CopyrightProofSummary:   strings.TrimSpace(item.CopyrightProofSummary),
				Rarity:                  strings.TrimSpace(item.Rarity),
				IssueLimit:              item.IssueLimit,
				DisplayCopy:             strings.TrimSpace(item.DisplayCopy),
				CirculationLimitSummary: strings.TrimSpace(item.CirculationLimitSummary),
				DisplayStatus:           item.DisplayStatus,
				ContentAuditStatus:      item.ContentAuditStatus,
				ProviderCode:            strings.TrimSpace(item.ProviderCode),
				CredentialRef:           strings.TrimSpace(item.CredentialRef),
				Status:                  item.Status,
				AuditStatus:             item.AuditStatus,
				CreateBy:                operatorID,
				CreateTime:              now,
			}
			if err := tx.WithContext(ctx).Table(drawCardTemplateRow{}.TableName()).Create(&row).Error; err != nil {
				return nil, err
			}
			templateIDs[code] = row.ID
			item.Id = row.ID
		case err != nil:
			return nil, err
		default:
			if err := tx.WithContext(ctx).Table(drawCardTemplateRow{}.TableName()).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return nil, err
			}
			templateIDs[code] = existing.ID
			item.Id = existing.ID
		}
	}
	return templateIDs, nil
}

func replaceDrawPools(ctx context.Context, tx *gorm.DB, activityID int64, scope pkgscope.GovernanceScope, pools []*smsclient.DrawPoolData, templates []*smsclient.DrawCardTemplateData, templateIDs map[string]int64, operatorID int64) error {
	now := time.Now()
	templateByCode := make(map[string]*smsclient.DrawCardTemplateData, len(templates))
	templateByID := make(map[int64]*smsclient.DrawCardTemplateData, len(templates))
	for _, item := range templates {
		if item == nil {
			continue
		}
		code := strings.TrimSpace(item.TemplateCode)
		if code != "" {
			templateByCode[code] = item
		}
		if item.Id > 0 {
			templateByID[item.Id] = item
		}
	}
	if err := tx.WithContext(ctx).Table(drawPoolTemplateRow{}.TableName()).
		Where("activity_id = ? AND is_deleted = 0", activityID).
		Updates(map[string]interface{}{"is_deleted": 1, "update_by": operatorID, "update_time": now}).Error; err != nil {
		return err
	}
	if err := tx.WithContext(ctx).Table(drawPoolRow{}.TableName()).
		Where("activity_id = ? AND is_deleted = 0", activityID).
		Updates(map[string]interface{}{
			"is_deleted":  1,
			"pool_name":   gorm.Expr("CONCAT(pool_name, '_del_', id)"),
			"pool_code":   gorm.Expr("CONCAT(pool_code, '_del_', id)"),
			"update_by":   operatorID,
			"update_time": now,
		}).Error; err != nil {
		return err
	}

	for _, item := range pools {
		row := drawPoolRow{
			ActivityID:      activityID,
			PlatformID:      scope.PlatformID,
			TenantID:        scope.TenantID,
			MerchantID:      scope.MerchantID,
			PoolCode:        strings.TrimSpace(item.PoolCode),
			PoolName:        strings.TrimSpace(item.PoolName),
			ProbabilityRule: strings.TrimSpace(item.ProbabilityRule),
			WheelSlotCount:  5,
			Sort:            item.Sort,
			Status:          item.Status,
			AuditStatus:     drawApprovalPassed,
			CreateBy:        operatorID,
			CreateTime:      now,
		}
		if err := tx.WithContext(ctx).Table(drawPoolRow{}.TableName()).Create(&row).Error; err != nil {
			return err
		}
		item.Id = row.ID

		for _, mapping := range item.Templates {
			templateID := mapping.TemplateId
			if templateID <= 0 {
				templateID = templateIDs[strings.TrimSpace(mapping.TemplateCode)]
			}
			if templateID <= 0 {
				return fmt.Errorf("卡池[%s]关联模板[%s]不存在", item.PoolName, mapping.TemplateCode)
			}
			templateMeta := templateByID[templateID]
			if templateMeta == nil {
				templateMeta = templateByCode[strings.TrimSpace(mapping.TemplateCode)]
			}
			remainingLimit := mapping.RemainingLimit
			if remainingLimit <= 0 {
				remainingLimit = mapping.SaleLimit
			}
			rowMapping := drawPoolTemplateRow{
				ActivityID:     activityID,
				PoolID:         row.ID,
				TemplateID:     templateID,
				SlotIndex:      mapping.SlotIndex,
				PlatformID:     scope.PlatformID,
				TenantID:       scope.TenantID,
				MerchantID:     scope.MerchantID,
				Rarity:         firstNonEmpty(strings.TrimSpace(mapping.Rarity), templateRarity(templateMeta)),
				Probability:    mapping.Probability,
				SaleLimit:      mapping.SaleLimit,
				RemainingLimit: remainingLimit,
				ConfigLimit:    mapping.ConfigLimit,
				Status:         item.Status,
				AuditStatus:    drawApprovalPassed,
				CreateBy:       operatorID,
				CreateTime:     now,
			}
			if err := tx.WithContext(ctx).Table(drawPoolTemplateRow{}.TableName()).Create(&rowMapping).Error; err != nil {
				return err
			}
			mapping.Id = rowMapping.ID
			mapping.TemplateId = templateID
		}
	}
	return nil
}

func templateRarity(item *smsclient.DrawCardTemplateData) string {
	if item == nil {
		return ""
	}
	return strings.TrimSpace(item.Rarity)
}

func appendDrawActivityAudit(ctx context.Context, tx *gorm.DB, activity *drawActivityRow, templates []*smsclient.DrawCardTemplateData, pools []*smsclient.DrawPoolData, operatorID int64, operatorName, action string) error {
	traceID := audit.NewTraceID(action, activity.ID)
	payload, err := json.Marshal(drawAuditPayload{
		Activity:  activity,
		Templates: templates,
		Pools:     pools,
	})
	if err != nil {
		return err
	}

	row := drawActivityAuditRow{
		ActivityID:               activity.ID,
		PlatformID:               activity.PlatformID,
		TenantID:                 activity.TenantID,
		MerchantID:               activity.MerchantID,
		OperationType:            action,
		OperatorID:               operatorID,
		OperatorName:             strings.TrimSpace(operatorName),
		ApprovalResult:           approvalResultLabel(activity.AuditStatus),
		RuleSnapshot:             activity.RuleSummary,
		CirculationLimitSnapshot: activity.CirculationLimitSummary,
		FailureSummary:           activity.PublishFailureSummary,
		TraceID:                  traceID,
		PayloadJSON:              string(payload),
		Status:                   activity.Status,
		AuditStatus:              activity.AuditStatus,
		CreateBy:                 operatorID,
		CreateTime:               time.Now(),
	}
	return tx.WithContext(ctx).Table(drawActivityAuditRow{}.TableName()).Create(&row).Error
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func buildActivityRowFromAdd(in *smsclient.AddDrawActivityReq) (*drawActivityRow, error) {
	startTime, err := parseDrawDateTime(in.StartTime, "开始时间")
	if err != nil {
		return nil, err
	}
	endTime, err := parseDrawDateTime(in.EndTime, "结束时间")
	if err != nil {
		return nil, err
	}
	row := &drawActivityRow{
		ActivityCode:                strings.TrimSpace(in.ActivityCode),
		Name:                        strings.TrimSpace(in.Name),
		RuleSummary:                 strings.TrimSpace(in.RuleSummary),
		StartTime:                   startTime,
		EndTime:                     endTime,
		RealNameRequired:            in.RealNameRequired,
		ParticipantConditionSummary: strings.TrimSpace(in.ParticipantConditionSummary),
		ConsumeRuleSummary:          strings.TrimSpace(in.ConsumeRuleSummary),
		ConsumeType:                 strings.TrimSpace(in.ConsumeType),
		ConsumeAmount:               in.ConsumeAmount,
		QuotaPerMember:              in.QuotaPerMember,
		DailyQuotaPerMember:         in.DailyQuotaPerMember,
		ProbabilityRule:             strings.TrimSpace(in.ProbabilityRule),
		ComplianceRuleSummary:       strings.TrimSpace(in.ComplianceRuleSummary),
		CirculationLimitSummary:     strings.TrimSpace(in.CirculationLimitSummary),
		ApprovalRecordRef:           strings.TrimSpace(in.ApprovalRecordRef),
		CopyrightStatus:             in.CopyrightStatus,
		ContentAuditStatus:          in.ContentAuditStatus,
		Status:                      in.Status,
		AuditStatus:                 in.AuditStatus,
		IsEnabled:                   in.IsEnabled,
		HomeEntryEnabled:            1,
		CreateBy:                    in.CreateBy,
		CreateTime:                  time.Now(),
	}
	applyHomeEntry(row, in.HomeEntry)
	return row, nil
}

func buildActivityRowFromUpdate(current *drawActivityRow, in *smsclient.UpdateDrawActivityReq) error {
	startTime, err := parseDrawDateTime(in.StartTime, "开始时间")
	if err != nil {
		return err
	}
	endTime, err := parseDrawDateTime(in.EndTime, "结束时间")
	if err != nil {
		return err
	}
	current.ActivityCode = strings.TrimSpace(in.ActivityCode)
	current.Name = strings.TrimSpace(in.Name)
	current.RuleSummary = strings.TrimSpace(in.RuleSummary)
	current.StartTime = startTime
	current.EndTime = endTime
	current.RealNameRequired = in.RealNameRequired
	current.ParticipantConditionSummary = strings.TrimSpace(in.ParticipantConditionSummary)
	current.ConsumeRuleSummary = strings.TrimSpace(in.ConsumeRuleSummary)
	current.ConsumeType = strings.TrimSpace(in.ConsumeType)
	current.ConsumeAmount = in.ConsumeAmount
	current.QuotaPerMember = in.QuotaPerMember
	current.DailyQuotaPerMember = in.DailyQuotaPerMember
	current.ProbabilityRule = strings.TrimSpace(in.ProbabilityRule)
	current.ComplianceRuleSummary = strings.TrimSpace(in.ComplianceRuleSummary)
	current.CirculationLimitSummary = strings.TrimSpace(in.CirculationLimitSummary)
	current.ApprovalRecordRef = strings.TrimSpace(in.ApprovalRecordRef)
	current.CopyrightStatus = in.CopyrightStatus
	current.ContentAuditStatus = in.ContentAuditStatus
	current.Status = in.Status
	current.AuditStatus = in.AuditStatus
	current.IsEnabled = in.IsEnabled
	updateBy := in.UpdateBy
	current.UpdateBy = &updateBy
	now := time.Now()
	current.UpdateTime = &now
	applyHomeEntry(current, in.HomeEntry)
	return nil
}

func ensureUniqueActivityCode(ctx context.Context, tx *gorm.DB, scope pkgscope.GovernanceScope, code string, excludeID int64) error {
	query := tx.WithContext(ctx).Table(drawActivityRow{}.TableName()).
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND activity_code = ? AND is_deleted = 0",
			scope.PlatformID, scope.TenantID, scope.MerchantID, code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("活动编码[%s]已存在", code)
	}
	return nil
}

func saveDrawAggregate(ctx context.Context, tx *gorm.DB, activity *drawActivityRow, scope pkgscope.GovernanceScope, templates []*smsclient.DrawCardTemplateData, pools []*smsclient.DrawPoolData, operatorID int64, operatorName string, action string, isCreate bool) error {
	if err := validateDrawAggregate(activity, templates, pools); err != nil {
		return err
	}
	if err := ensureUniqueActivityCode(ctx, tx, scope, activity.ActivityCode, activity.ID); err != nil {
		return err
	}

	activity.PlatformID = scope.PlatformID
	activity.TenantID = scope.TenantID
	activity.MerchantID = scope.MerchantID

	var existingAuditCount int64
	if !isCreate && activity.ID > 0 {
		_ = tx.WithContext(ctx).Table(drawActivityAuditRow{}.TableName()).
			Where("activity_id = ? AND is_deleted = 0", activity.ID).
			Count(&existingAuditCount).Error
	}
	readiness := buildDrawReadiness(activity, templates, pools, existingAuditCount)
	activity.PublishReadiness = readiness.PublishReadiness
	activity.PublishFailureSummary = readiness.Summary

	if isCreate {
		if err := tx.WithContext(ctx).Table(drawActivityRow{}.TableName()).Create(activity).Error; err != nil {
			return err
		}
		if err := logiccommon.ApplyDrawActivityScope(ctx, tx, activity.ID, scope); err != nil {
			return err
		}
	} else {
		updates := map[string]interface{}{
			"activity_code":                 activity.ActivityCode,
			"name":                          activity.Name,
			"rule_summary":                  activity.RuleSummary,
			"start_time":                    activity.StartTime,
			"end_time":                      activity.EndTime,
			"real_name_required":            activity.RealNameRequired,
			"participant_condition_summary": activity.ParticipantConditionSummary,
			"consume_rule_summary":          activity.ConsumeRuleSummary,
			"consume_type":                  activity.ConsumeType,
			"consume_amount":                activity.ConsumeAmount,
			"quota_per_member":              activity.QuotaPerMember,
			"daily_quota_per_member":        activity.DailyQuotaPerMember,
			"probability_rule":              activity.ProbabilityRule,
			"compliance_rule_summary":       activity.ComplianceRuleSummary,
			"circulation_limit_summary":     activity.CirculationLimitSummary,
			"approval_record_ref":           activity.ApprovalRecordRef,
			"publish_failure_summary":       activity.PublishFailureSummary,
			"publish_readiness":             activity.PublishReadiness,
			"copyright_status":              activity.CopyrightStatus,
			"content_audit_status":          activity.ContentAuditStatus,
			"status":                        activity.Status,
			"audit_status":                  activity.AuditStatus,
			"is_enabled":                    activity.IsEnabled,
			"show_on_home":                  activity.ShowOnHome,
			"home_entry_title":              activity.HomeEntryTitle,
			"home_entry_subtitle":           activity.HomeEntrySubtitle,
			"home_entry_image":              activity.HomeEntryImage,
			"home_entry_sort":               activity.HomeEntrySort,
			"home_entry_enabled":            activity.HomeEntryEnabled,
			"landing_target_type":           activity.LandingTargetType,
			"landing_target_value":          activity.LandingTargetValue,
			"update_by":                     safeInt64(activity.UpdateBy),
			"update_time":                   activity.UpdateTime,
		}
		if err := tx.WithContext(ctx).Table(drawActivityRow{}.TableName()).Where("id = ?", activity.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	templateIDs, err := upsertDrawTemplates(ctx, tx, scope, templates, operatorID)
	if err != nil {
		return err
	}
	if err := replaceDrawPools(ctx, tx, activity.ID, scope, pools, templates, templateIDs, operatorID); err != nil {
		return err
	}
	return appendDrawActivityAudit(ctx, tx, activity, templates, pools, operatorID, operatorName, action)
}

func safeInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
