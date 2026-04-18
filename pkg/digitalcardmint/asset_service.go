package digitalcardmint

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MemberDigitalCardAssetFilter struct {
	PageNum  int32
	PageSize int32
}

type MemberDigitalCardAssetItem struct {
	AssetInstanceID       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	TemplateID            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	CardFaceImage         string `json:"cardFaceImage"`
	ActivityID            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	Rarity                string `json:"rarity"`
	ObtainedAt            string `json:"obtainedAt"`
	MintStatus            string `json:"mintStatus"`
	MintStatusText        string `json:"mintStatusText"`
	ChainStatus           string `json:"chainStatus"`
	ChainStatusText       string `json:"chainStatusText"`
	DisplayStatus         string `json:"displayStatus"`
	DisplayStatusText     string `json:"displayStatusText"`
	ComplianceStatus      string `json:"complianceStatus"`
	ComplianceStatusText  string `json:"complianceStatusText"`
	TokenStatusText       string `json:"tokenStatusText"`
	ComplianceRuleSummary string `json:"complianceRuleSummary"`
}

type MemberDigitalCardAssetTimelineItem struct {
	OperationType string `json:"operationType"`
	OperationText string `json:"operationText"`
	StatusText    string `json:"statusText"`
	ReasonText    string `json:"reasonText"`
	CreateTime    string `json:"createTime"`
}

type MemberDigitalCardAssetDrawSummary struct {
	ParticipationRecordID int64  `json:"participationRecordId"`
	ResultType            string `json:"resultType"`
	ResultStatus          string `json:"resultStatus"`
	ResultStatusText      string `json:"resultStatusText"`
	FailureReason         string `json:"failureReason"`
	CreateTime            string `json:"createTime"`
}

type MemberDigitalCardAssetDetail struct {
	Item                MemberDigitalCardAssetItem           `json:"item"`
	TokenIDMasked       string                               `json:"tokenIdMasked"`
	LatestStatusSummary string                               `json:"latestStatusSummary"`
	RestrictionReason   string                               `json:"restrictionReason"`
	DrawSummary         MemberDigitalCardAssetDrawSummary    `json:"drawSummary"`
	Timeline            []MemberDigitalCardAssetTimelineItem `json:"timeline"`
}

type DigitalCardAssetAuditFilter struct {
	PageNum          int32
	PageSize         int32
	ActivityID       int64
	ActivityName     string
	MemberID         int64
	TemplateID       int64
	TemplateName     string
	AssetNo          string
	TokenID          string
	MintStatus       string
	ChainStatus      string
	DisplayStatus    string
	ComplianceStatus string
	StartTime        string
	EndTime          string
}

type DigitalCardAssetAuditItem struct {
	AssetInstanceID       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	MemberID              int64  `json:"memberId"`
	ActivityID            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	TemplateID            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	CardFaceImage         string `json:"cardFaceImage"`
	Rarity                string `json:"rarity"`
	TokenID               string `json:"tokenId"`
	MintTaskID            int64  `json:"mintTaskId"`
	MintStatus            string `json:"mintStatus"`
	MintStatusText        string `json:"mintStatusText"`
	ChainStatus           string `json:"chainStatus"`
	ChainStatusText       string `json:"chainStatusText"`
	DisplayStatus         string `json:"displayStatus"`
	DisplayStatusText     string `json:"displayStatusText"`
	ComplianceStatus      string `json:"complianceStatus"`
	ComplianceStatusText  string `json:"complianceStatusText"`
	TokenStatusText       string `json:"tokenStatusText"`
	ComplianceRuleSummary string `json:"complianceRuleSummary"`
	LastReceiptSummary    string `json:"lastReceiptSummary"`
	ObtainedAt            string `json:"obtainedAt"`
	DisposedAt            string `json:"disposedAt"`
	LatestReasonSummary   string `json:"latestReasonSummary"`
}

type DigitalCardAssetAuditParticipationSummary struct {
	ParticipationRecordID int64  `json:"participationRecordId"`
	RequestID             string `json:"requestId"`
	ResultType            string `json:"resultType"`
	ResultStatus          string `json:"resultStatus"`
	ResultStatusText      string `json:"resultStatusText"`
	FailureReason         string `json:"failureReason"`
	CreateTime            string `json:"createTime"`
}

type DigitalCardAssetAuditMintTaskSummary struct {
	TaskID             int64    `json:"taskId"`
	RequestID          string   `json:"requestId"`
	TraceID            string   `json:"traceId"`
	TaskStatus         string   `json:"taskStatus"`
	TaskStatusText     string   `json:"taskStatusText"`
	MintStatus         string   `json:"mintStatus"`
	MintStatusText     string   `json:"mintStatusText"`
	ChainStatus        string   `json:"chainStatus"`
	ChainStatusText    string   `json:"chainStatusText"`
	ChainTxID          string   `json:"chainTxId"`
	LastReceiptSummary string   `json:"lastReceiptSummary"`
	AvailableActions   []string `json:"availableActions"`
}

type DigitalCardAssetAuditDetail struct {
	Item                  DigitalCardAssetAuditItem                 `json:"item"`
	ParticipationSummary  DigitalCardAssetAuditParticipationSummary `json:"participationSummary"`
	MintTaskSummary       DigitalCardAssetAuditMintTaskSummary      `json:"mintTaskSummary"`
	TraceID               string                                    `json:"traceId"`
	RequestID             string                                    `json:"requestId"`
	RuleSnapshotJSON      string                                    `json:"ruleSnapshotJson"`
	Logs                  []AssetLogItem                            `json:"logs"`
	AvailableAssetActions []string                                  `json:"availableAssetActions"`
}

type DigitalCardAssetActionResult struct {
	AssetInstanceID      int64  `json:"assetInstanceId"`
	DisplayStatus        string `json:"displayStatus"`
	DisplayStatusText    string `json:"displayStatusText"`
	ComplianceStatus     string `json:"complianceStatus"`
	ComplianceStatusText string `json:"complianceStatusText"`
}

type digitalCardAssetBaseRow struct {
	AssetInstanceID           int64        `gorm:"column:asset_instance_id"`
	PlatformID                int64        `gorm:"column:platform_id"`
	TenantID                  int64        `gorm:"column:tenant_id"`
	MerchantID                int64        `gorm:"column:merchant_id"`
	ParticipationRecordID     int64        `gorm:"column:participation_record_id"`
	RequestID                 string       `gorm:"column:request_id"`
	TraceID                   string       `gorm:"column:trace_id"`
	ActivityID                int64        `gorm:"column:activity_id"`
	ActivityName              string       `gorm:"column:activity_name"`
	ActivityComplianceSummary string       `gorm:"column:activity_compliance_summary"`
	MemberID                  int64        `gorm:"column:member_id"`
	TemplateID                int64        `gorm:"column:template_id"`
	TemplateName              string       `gorm:"column:template_name"`
	CardFaceImage             string       `gorm:"column:card_face_image"`
	Rarity                    string       `gorm:"column:rarity"`
	AssetNo                   string       `gorm:"column:asset_no"`
	AssetStatus               string       `gorm:"column:asset_status"`
	MintStatus                string       `gorm:"column:mint_status"`
	ChainStatus               string       `gorm:"column:chain_status"`
	DisplayStatus             string       `gorm:"column:display_status"`
	ComplianceStatus          string       `gorm:"column:compliance_status"`
	DisplayReason             string       `gorm:"column:display_reason"`
	ComplianceReason          string       `gorm:"column:compliance_reason"`
	RuleSnapshotJSON          string       `gorm:"column:rule_snapshot_json"`
	TokenID                   string       `gorm:"column:token_id"`
	MintTaskID                int64        `gorm:"column:mint_task_id"`
	LastReceiptSummary        string       `gorm:"column:last_receipt_summary"`
	LastReceiptJSON           string       `gorm:"column:last_receipt_json"`
	ChainTxID                 string       `gorm:"column:chain_tx_id"`
	TaskStatus                string       `gorm:"column:task_status"`
	DisposedAt                nullableTime `gorm:"column:disposed_at"`
	DisposedBy                int64        `gorm:"column:disposed_by"`
	ObtainedAt                nullableTime `gorm:"column:obtained_at"`
	ResultType                string       `gorm:"column:result_type"`
	ResultStatus              string       `gorm:"column:result_status"`
	FailureReason             string       `gorm:"column:failure_reason"`
	ParticipationCreateTime   nullableTime `gorm:"column:participation_create_time"`
}

func (s *Service) QueryMemberDigitalCardAssetList(ctx context.Context, currentScope pkgscope.GovernanceScope, memberID int64, filter MemberDigitalCardAssetFilter) (int64, []MemberDigitalCardAssetItem, error) {
	if s.DB == nil {
		return 0, nil, errors.New("数据库未初始化")
	}
	if memberID <= 0 {
		return 0, nil, errors.New("会员ID不能为空")
	}
	if filter.PageNum <= 0 {
		filter.PageNum = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	base := s.memberAssetBaseQuery(ctx, currentScope, memberID)
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var rows []digitalCardAssetBaseRow
	if err := base.Select(memberAssetSelectColumns()).
		Order("instance.id DESC").
		Offset(int((filter.PageNum - 1) * filter.PageSize)).
		Limit(int(filter.PageSize)).
		Find(&rows).Error; err != nil {
		return 0, nil, err
	}

	items := make([]MemberDigitalCardAssetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildMemberAssetItem(row))
	}
	return total, items, nil
}

func (s *Service) QueryMemberDigitalCardAssetDetail(ctx context.Context, currentScope pkgscope.GovernanceScope, memberID int64, assetInstanceID int64) (*MemberDigitalCardAssetDetail, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if memberID <= 0 || assetInstanceID <= 0 {
		return nil, errors.New("资产实例ID不能为空")
	}

	var row digitalCardAssetBaseRow
	if err := s.memberAssetBaseQuery(ctx, currentScope, memberID).
		Where("instance.id = ?", assetInstanceID).
		Select(memberAssetSelectColumns()).
		Take(&row).Error; err != nil {
		return nil, err
	}

	logs, err := s.queryAssetLogs(ctx, row.AssetInstanceID)
	if err != nil {
		return nil, err
	}

	return &MemberDigitalCardAssetDetail{
		Item:                buildMemberAssetItem(row),
		TokenIDMasked:       maskTokenID(row.TokenID),
		LatestStatusSummary: ResolveAssetStatusText(row.AssetStatus, row.MintStatus, row.ChainStatus),
		RestrictionReason:   firstNonEmpty(row.ComplianceReason, row.DisplayReason),
		DrawSummary: MemberDigitalCardAssetDrawSummary{
			ParticipationRecordID: row.ParticipationRecordID,
			ResultType:            row.ResultType,
			ResultStatus:          row.ResultStatus,
			ResultStatusText:      row.ResultStatus,
			FailureReason:         row.FailureReason,
			CreateTime:            formatNullableTime(row.ParticipationCreateTime),
		},
		Timeline: buildMemberTimeline(logs),
	}, nil
}

func (s *Service) QueryDigitalCardAssetAuditList(ctx context.Context, currentScope pkgscope.GovernanceScope, filter DigitalCardAssetAuditFilter) (int64, []DigitalCardAssetAuditItem, error) {
	if s.DB == nil {
		return 0, nil, errors.New("数据库未初始化")
	}
	if filter.PageNum <= 0 {
		filter.PageNum = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	base := s.auditAssetBaseQuery(ctx, currentScope, filter)
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var rows []digitalCardAssetBaseRow
	if err := base.Select(auditAssetSelectColumns()).
		Order("instance.id DESC").
		Offset(int((filter.PageNum - 1) * filter.PageSize)).
		Limit(int(filter.PageSize)).
		Find(&rows).Error; err != nil {
		return 0, nil, err
	}

	items := make([]DigitalCardAssetAuditItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildAuditAssetItem(row))
	}
	return total, items, nil
}

func (s *Service) QueryDigitalCardAssetAuditDetail(ctx context.Context, currentScope pkgscope.GovernanceScope, assetInstanceID int64) (*DigitalCardAssetAuditDetail, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if assetInstanceID <= 0 {
		return nil, errors.New("资产实例ID不能为空")
	}

	var row digitalCardAssetBaseRow
	if err := s.auditAssetBaseQuery(ctx, currentScope, DigitalCardAssetAuditFilter{}).
		Where("instance.id = ?", assetInstanceID).
		Select(auditAssetSelectColumns()).
		Take(&row).Error; err != nil {
		return nil, err
	}

	logs, err := s.queryAssetLogs(ctx, row.AssetInstanceID)
	if err != nil {
		return nil, err
	}

	chainActions := []string{}
	if row.MintTaskID > 0 {
		if actions, actionErr := s.QueryAvailableActions(ctx, currentScope, row.MintTaskID); actionErr == nil {
			chainActions = actions
		}
	}

	return &DigitalCardAssetAuditDetail{
		Item: buildAuditAssetItem(row),
		ParticipationSummary: DigitalCardAssetAuditParticipationSummary{
			ParticipationRecordID: row.ParticipationRecordID,
			RequestID:             row.RequestID,
			ResultType:            row.ResultType,
			ResultStatus:          row.ResultStatus,
			ResultStatusText:      row.ResultStatus,
			FailureReason:         row.FailureReason,
			CreateTime:            formatNullableTime(row.ParticipationCreateTime),
		},
		MintTaskSummary: DigitalCardAssetAuditMintTaskSummary{
			TaskID:             row.MintTaskID,
			RequestID:          row.RequestID,
			TraceID:            row.TraceID,
			TaskStatus:         row.TaskStatus,
			TaskStatusText:     taskStatusText(row.TaskStatus),
			MintStatus:         row.MintStatus,
			MintStatusText:     mintStatusText(row.MintStatus),
			ChainStatus:        row.ChainStatus,
			ChainStatusText:    chainStatusText(row.ChainStatus),
			ChainTxID:          row.ChainTxID,
			LastReceiptSummary: row.LastReceiptSummary,
			AvailableActions:   chainActions,
		},
		TraceID:               row.TraceID,
		RequestID:             row.RequestID,
		RuleSnapshotJSON:      row.RuleSnapshotJSON,
		Logs:                  logs,
		AvailableAssetActions: availableAssetActions(row.DisplayStatus, row.ComplianceStatus),
	}, nil
}

func (s *Service) ReviewDigitalCardAssetCompliance(ctx context.Context, currentScope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string) (*DigitalCardAssetActionResult, error) {
	return s.mutateAssetComplianceState(ctx, currentScope, assetInstanceID, operatorID, reason, DisplayStatusHidden, ComplianceStatusReview, OperationAssetComplianceReview, false)
}

func (s *Service) OfflineDigitalCardAssetDisplay(ctx context.Context, currentScope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string) (*DigitalCardAssetActionResult, error) {
	return s.mutateAssetComplianceState(ctx, currentScope, assetInstanceID, operatorID, reason, DisplayStatusOfflined, ComplianceStatusRestricted, OperationAssetOfflined, false)
}

func (s *Service) RecycleDigitalCardAsset(ctx context.Context, currentScope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string) (*DigitalCardAssetActionResult, error) {
	return s.mutateAssetComplianceState(ctx, currentScope, assetInstanceID, operatorID, reason, DisplayStatusRecycled, ComplianceStatusRecycled, OperationAssetRecycled, true)
}

func (s *Service) memberAssetBaseQuery(ctx context.Context, currentScope pkgscope.GovernanceScope, memberID int64) *gorm.DB {
	base := s.DB.WithContext(ctx).
		Table("sms_card_instance AS instance").
		Joins("LEFT JOIN sms_draw_activity AS activity ON activity.id = instance.activity_id AND activity.is_deleted = 0").
		Joins("LEFT JOIN sms_card_template AS template ON template.id = instance.template_id AND template.is_deleted = 0").
		Joins("LEFT JOIN sms_card_mint_task AS task ON task.asset_instance_id = instance.id AND task.is_deleted = 0").
		Joins("LEFT JOIN sms_draw_participation_record AS record ON record.id = instance.participation_record_id AND record.is_deleted = 0").
		Where("instance.is_deleted = 0 AND instance.member_id = ?", memberID)
	return pkgscope.ApplyGovernanceScope(base, currentScope, "instance")
}

func (s *Service) auditAssetBaseQuery(ctx context.Context, currentScope pkgscope.GovernanceScope, filter DigitalCardAssetAuditFilter) *gorm.DB {
	base := s.DB.WithContext(ctx).
		Table("sms_card_instance AS instance").
		Joins("LEFT JOIN sms_draw_activity AS activity ON activity.id = instance.activity_id AND activity.is_deleted = 0").
		Joins("LEFT JOIN sms_card_template AS template ON template.id = instance.template_id AND template.is_deleted = 0").
		Joins("LEFT JOIN sms_card_mint_task AS task ON task.asset_instance_id = instance.id AND task.is_deleted = 0").
		Joins("LEFT JOIN sms_draw_participation_record AS record ON record.id = instance.participation_record_id AND record.is_deleted = 0").
		Where("instance.is_deleted = 0")
	base = pkgscope.ApplyGovernanceScope(base, currentScope, "instance")

	if filter.ActivityID > 0 {
		base = base.Where("instance.activity_id = ?", filter.ActivityID)
	}
	if strings.TrimSpace(filter.ActivityName) != "" {
		base = base.Where("activity.name LIKE ?", "%"+strings.TrimSpace(filter.ActivityName)+"%")
	}
	if filter.MemberID > 0 {
		base = base.Where("instance.member_id = ?", filter.MemberID)
	}
	if filter.TemplateID > 0 {
		base = base.Where("instance.template_id = ?", filter.TemplateID)
	}
	if strings.TrimSpace(filter.TemplateName) != "" {
		base = base.Where("template.template_name LIKE ?", "%"+strings.TrimSpace(filter.TemplateName)+"%")
	}
	if strings.TrimSpace(filter.AssetNo) != "" {
		base = base.Where("instance.asset_no LIKE ?", "%"+strings.TrimSpace(filter.AssetNo)+"%")
	}
	if strings.TrimSpace(filter.TokenID) != "" {
		base = base.Where("instance.token_id LIKE ?", "%"+strings.TrimSpace(filter.TokenID)+"%")
	}
	if strings.TrimSpace(filter.MintStatus) != "" {
		base = base.Where("instance.mint_status = ?", strings.TrimSpace(filter.MintStatus))
	}
	if strings.TrimSpace(filter.ChainStatus) != "" {
		base = base.Where("COALESCE(NULLIF(instance.chain_status, ''), task.chain_status, ?) = ?", ChainStatusUnknown, strings.TrimSpace(filter.ChainStatus))
	}
	if strings.TrimSpace(filter.DisplayStatus) != "" {
		base = base.Where("COALESCE(NULLIF(instance.display_status, ''), ?) = ?", DisplayStatusVisible, strings.TrimSpace(filter.DisplayStatus))
	}
	if strings.TrimSpace(filter.ComplianceStatus) != "" {
		base = base.Where("COALESCE(NULLIF(instance.compliance_status, ''), ?) = ?", ComplianceStatusClear, strings.TrimSpace(filter.ComplianceStatus))
	}
	if start, ok := parseStartTime(filter.StartTime); ok {
		base = base.Where("COALESCE(instance.update_time, instance.create_time) >= ?", start)
	}
	if end, ok := parseEndTime(filter.EndTime); ok {
		base = base.Where("COALESCE(instance.update_time, instance.create_time) <= ?", end)
	}
	return base
}

func memberAssetSelectColumns() string {
	return `
		instance.id AS asset_instance_id,
		instance.platform_id AS platform_id,
		instance.tenant_id AS tenant_id,
		instance.merchant_id AS merchant_id,
		instance.participation_record_id AS participation_record_id,
		COALESCE(instance.request_id, '') AS request_id,
		COALESCE(instance.trace_id, '') AS trace_id,
		instance.activity_id AS activity_id,
		COALESCE(activity.name, '') AS activity_name,
		COALESCE(activity.compliance_rule_summary, '') AS activity_compliance_summary,
		instance.member_id AS member_id,
		instance.template_id AS template_id,
		COALESCE(template.template_name, '') AS template_name,
		COALESCE(template.card_face_image, '') AS card_face_image,
		COALESCE(instance.rarity, '') AS rarity,
		COALESCE(instance.asset_no, '') AS asset_no,
		COALESCE(instance.asset_status, '') AS asset_status,
		COALESCE(instance.mint_status, '') AS mint_status,
		COALESCE(NULLIF(instance.chain_status, ''), task.chain_status, '` + ChainStatusUnknown + `') AS chain_status,
		COALESCE(NULLIF(instance.display_status, ''), '` + DisplayStatusVisible + `') AS display_status,
		COALESCE(NULLIF(instance.compliance_status, ''), '` + ComplianceStatusClear + `') AS compliance_status,
		COALESCE(instance.display_reason, '') AS display_reason,
		COALESCE(instance.compliance_reason, '') AS compliance_reason,
		COALESCE(instance.rule_snapshot_json, '') AS rule_snapshot_json,
		COALESCE(instance.token_id, '') AS token_id,
		instance.mint_task_id AS mint_task_id,
		COALESCE(task.last_receipt_summary, '') AS last_receipt_summary,
		COALESCE(task.last_receipt_json, '') AS last_receipt_json,
		COALESCE(task.chain_tx_id, '') AS chain_tx_id,
		COALESCE(task.task_status, '') AS task_status,
		instance.disposed_at AS disposed_at,
		COALESCE(instance.disposed_by, 0) AS disposed_by,
		COALESCE(instance.issued_at, instance.create_time) AS obtained_at,
		COALESCE(record.result_type, '') AS result_type,
		COALESCE(record.result_status, '') AS result_status,
		COALESCE(record.failure_reason, '') AS failure_reason,
		record.create_time AS participation_create_time`
}

func auditAssetSelectColumns() string {
	return memberAssetSelectColumns()
}

func buildMemberAssetItem(row digitalCardAssetBaseRow) MemberDigitalCardAssetItem {
	return MemberDigitalCardAssetItem{
		AssetInstanceID:       row.AssetInstanceID,
		AssetNo:               row.AssetNo,
		TemplateID:            row.TemplateID,
		TemplateName:          row.TemplateName,
		CardFaceImage:         row.CardFaceImage,
		ActivityID:            row.ActivityID,
		ActivityName:          row.ActivityName,
		Rarity:                row.Rarity,
		ObtainedAt:            formatNullableTime(row.ObtainedAt),
		MintStatus:            row.MintStatus,
		MintStatusText:        mintStatusText(row.MintStatus),
		ChainStatus:           row.ChainStatus,
		ChainStatusText:       chainStatusText(row.ChainStatus),
		DisplayStatus:         row.DisplayStatus,
		DisplayStatusText:     displayStatusText(row.DisplayStatus),
		ComplianceStatus:      row.ComplianceStatus,
		ComplianceStatusText:  complianceStatusText(row.ComplianceStatus),
		TokenStatusText:       tokenStatusText(row.TokenID, row.ChainStatus),
		ComplianceRuleSummary: complianceRuleSummary(row.ActivityComplianceSummary, row.DisplayReason, row.ComplianceReason),
	}
}

func buildAuditAssetItem(row digitalCardAssetBaseRow) DigitalCardAssetAuditItem {
	return DigitalCardAssetAuditItem{
		AssetInstanceID:       row.AssetInstanceID,
		AssetNo:               row.AssetNo,
		MemberID:              row.MemberID,
		ActivityID:            row.ActivityID,
		ActivityName:          row.ActivityName,
		TemplateID:            row.TemplateID,
		TemplateName:          row.TemplateName,
		CardFaceImage:         row.CardFaceImage,
		Rarity:                row.Rarity,
		TokenID:               row.TokenID,
		MintTaskID:            row.MintTaskID,
		MintStatus:            row.MintStatus,
		MintStatusText:        mintStatusText(row.MintStatus),
		ChainStatus:           row.ChainStatus,
		ChainStatusText:       chainStatusText(row.ChainStatus),
		DisplayStatus:         row.DisplayStatus,
		DisplayStatusText:     displayStatusText(row.DisplayStatus),
		ComplianceStatus:      row.ComplianceStatus,
		ComplianceStatusText:  complianceStatusText(row.ComplianceStatus),
		TokenStatusText:       tokenStatusText(row.TokenID, row.ChainStatus),
		ComplianceRuleSummary: complianceRuleSummary(row.ActivityComplianceSummary, row.DisplayReason, row.ComplianceReason),
		LastReceiptSummary:    row.LastReceiptSummary,
		ObtainedAt:            formatNullableTime(row.ObtainedAt),
		DisposedAt:            formatNullableTime(row.DisposedAt),
		LatestReasonSummary:   firstNonEmpty(row.ComplianceReason, row.DisplayReason, row.LastReceiptSummary),
	}
}

func buildMemberTimeline(logs []AssetLogItem) []MemberDigitalCardAssetTimelineItem {
	if len(logs) == 0 {
		return []MemberDigitalCardAssetTimelineItem{}
	}
	result := make([]MemberDigitalCardAssetTimelineItem, 0, len(logs))
	for index := len(logs) - 1; index >= 0; index-- {
		item := logs[index]
		result = append(result, MemberDigitalCardAssetTimelineItem{
			OperationType: item.OperationType,
			OperationText: assetLogOperationText(item.OperationType),
			StatusText:    resolveTimelineStatusText(item.ToStatus),
			ReasonText:    item.ReasonText,
			CreateTime:    item.CreateTime,
		})
	}
	return result
}

func availableAssetActions(displayStatus string, complianceStatus string) []string {
	display := strings.TrimSpace(displayStatus)
	if display == "" {
		display = DisplayStatusVisible
	}
	compliance := strings.TrimSpace(complianceStatus)
	if compliance == "" {
		compliance = ComplianceStatusClear
	}
	if compliance == ComplianceStatusRecycled || display == DisplayStatusRecycled {
		return []string{}
	}

	actions := []string{"review"}
	if display != DisplayStatusOfflined {
		actions = append(actions, "offline")
	}
	if compliance == ComplianceStatusReview || compliance == ComplianceStatusRestricted || compliance == ComplianceStatusRecycleRequested || display == DisplayStatusOfflined || display == DisplayStatusHidden {
		actions = append(actions, "recycle")
	}
	return actions
}

func (s *Service) mutateAssetComplianceState(ctx context.Context, currentScope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, nextDisplayStatus string, nextComplianceStatus string, operationType string, requireRestrictedContext bool) (*DigitalCardAssetActionResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if assetInstanceID <= 0 {
		return nil, errors.New("资产实例ID不能为空")
	}
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("处置原因不能为空")
	}

	var result *DigitalCardAssetActionResult
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var row digitalCardAssetBaseRow
		base := tx.WithContext(ctx).
			Table("sms_card_instance AS instance").
			Joins("LEFT JOIN sms_draw_activity AS activity ON activity.id = instance.activity_id AND activity.is_deleted = 0").
			Joins("LEFT JOIN sms_card_template AS template ON template.id = instance.template_id AND template.is_deleted = 0").
			Joins("LEFT JOIN sms_card_mint_task AS task ON task.asset_instance_id = instance.id AND task.is_deleted = 0").
			Joins("LEFT JOIN sms_draw_participation_record AS record ON record.id = instance.participation_record_id AND record.is_deleted = 0").
			Where("instance.id = ? AND instance.is_deleted = 0", assetInstanceID).
			Select(memberAssetSelectColumns()).
			Clauses(clause.Locking{Strength: "UPDATE"})
		if err := base.Take(&row).Error; err != nil {
			return err
		}
		if err := validateScope(row.PlatformID, row.TenantID, row.MerchantID, currentScope); err != nil {
			return err
		}
		if strings.TrimSpace(row.DisplayStatus) == DisplayStatusRecycled || strings.TrimSpace(row.ComplianceStatus) == ComplianceStatusRecycled {
			return errors.New("已回收资产不允许重复处置")
		}
		if requireRestrictedContext && strings.TrimSpace(row.ComplianceStatus) != ComplianceStatusReview && strings.TrimSpace(row.ComplianceStatus) != ComplianceStatusRestricted && strings.TrimSpace(row.DisplayStatus) != DisplayStatusOfflined && strings.TrimSpace(row.DisplayStatus) != DisplayStatusHidden {
			return errors.New("回收处置前必须已进入受限或复核上下文")
		}

		snapshotJSON, snapshotErr := json.Marshal(map[string]interface{}{
			"platformId":         row.PlatformID,
			"tenantId":           row.TenantID,
			"merchantId":         row.MerchantID,
			"operatorId":         operatorID,
			"requestId":          row.RequestID,
			"traceId":            row.TraceID,
			"mintTaskId":         row.MintTaskID,
			"previousDisplay":    row.DisplayStatus,
			"previousCompliance": row.ComplianceStatus,
			"nextDisplay":        nextDisplayStatus,
			"nextCompliance":     nextComplianceStatus,
			"reason":             strings.TrimSpace(reason),
			"lastReceiptSummary": row.LastReceiptSummary,
			"disposedAt":         now.Format("2006-01-02 15:04:05"),
		})
		if snapshotErr != nil {
			return snapshotErr
		}

		updateMap := map[string]interface{}{
			"display_status":     nextDisplayStatus,
			"compliance_status":  nextComplianceStatus,
			"display_reason":     strings.TrimSpace(reason),
			"compliance_reason":  strings.TrimSpace(reason),
			"rule_snapshot_json": string(snapshotJSON),
			"disposed_at":        now,
			"disposed_by":        operatorID,
			"update_by":          operatorID,
			"update_time":        now,
		}
		if err := tx.WithContext(ctx).
			Table(CardInstanceRow{}.TableName()).
			Where("id = ? AND is_deleted = 0", assetInstanceID).
			Updates(updateMap).Error; err != nil {
			return err
		}

		instance := &CardInstanceRow{
			ID:                    row.AssetInstanceID,
			ParticipationRecordID: row.ParticipationRecordID,
		}
		if err := s.appendAssetLogTx(ctx, tx, instance, row.ComplianceStatus, nextComplianceStatus, operationType, OperatorManual, row.TraceID, "", strings.TrimSpace(reason), map[string]interface{}{
			"assetInstanceId":    row.AssetInstanceID,
			"requestId":          row.RequestID,
			"traceId":            row.TraceID,
			"platformId":         row.PlatformID,
			"tenantId":           row.TenantID,
			"merchantId":         row.MerchantID,
			"operatorId":         operatorID,
			"mintTaskId":         row.MintTaskID,
			"displayStatus":      nextDisplayStatus,
			"complianceStatus":   nextComplianceStatus,
			"lastReceiptSummary": row.LastReceiptSummary,
		}); err != nil {
			return err
		}

		result = &DigitalCardAssetActionResult{
			AssetInstanceID:      row.AssetInstanceID,
			DisplayStatus:        nextDisplayStatus,
			DisplayStatusText:    displayStatusText(nextDisplayStatus),
			ComplianceStatus:     nextComplianceStatus,
			ComplianceStatusText: complianceStatusText(nextComplianceStatus),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func formatTimePtr(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func resolveTimelineStatusText(status string) string {
	value := strings.TrimSpace(status)
	switch value {
	case MintStatusPending, MintStatusProcessing, MintStatusSuccess, MintStatusFailed, MintStatusCompensating, MintStatusManualReview, MintStatusFrozen:
		return mintStatusText(value)
	case DisplayStatusVisible, DisplayStatusHidden, DisplayStatusOfflined, DisplayStatusRecycled:
		return displayStatusText(value)
	case ComplianceStatusClear, ComplianceStatusReview, ComplianceStatusRestricted, ComplianceStatusRecycleRequested, ComplianceStatusRecycled:
		return complianceStatusText(value)
	default:
		return value
	}
}
