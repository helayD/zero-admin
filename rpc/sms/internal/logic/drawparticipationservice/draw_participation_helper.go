package drawparticipationservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	drawEligibilityEligible           = "eligible"
	drawEligibilityNeedLogin          = "need_login"
	drawEligibilityNeedRealName       = "need_real_name"
	drawEligibilityQuotaExhausted     = "quota_exhausted"
	drawEligibilityActivityOffline    = "activity_offline"
	drawEligibilityMemberDisabled     = "member_disabled"
	drawEligibilityInventoryExhausted = "inventory_exhausted"

	drawNextActionNone        = "none"
	drawNextActionLogin       = "login"
	drawNextActionRealName    = "real_name"
	drawNextActionRetryLater  = "retry_later"
	drawConsumeTypeLottery    = "lottery_times"
	drawResultTypeNotWon      = "not_won"
	drawResultTypeWon         = "won"
	drawResultTypeRejected    = "rejected"
	drawResultStatusNotWon    = "not_won"
	drawResultStatusWon       = "won_pending_asset"
	drawResultStatusNeedReal  = "rejected_need_real_name"
	drawResultStatusNeedLogin = "rejected_need_login"
	drawResultStatusQuota     = "rejected_quota_exhausted"
	drawResultStatusOffline   = "rejected_activity_offline"
	drawResultStatusDisabled  = "rejected_member_disabled"
	drawResultStatusInventory = "rejected_inventory_exhausted"
)

type drawActivitySnapshot struct {
	ID                          int64     `gorm:"column:id"`
	ActivityCode                string    `gorm:"column:activity_code"`
	Name                        string    `gorm:"column:name"`
	RuleSummary                 string    `gorm:"column:rule_summary"`
	StartTime                   time.Time `gorm:"column:start_time"`
	EndTime                     time.Time `gorm:"column:end_time"`
	RealNameRequired            int32     `gorm:"column:real_name_required"`
	ParticipantConditionSummary string    `gorm:"column:participant_condition_summary"`
	ConsumeRuleSummary          string    `gorm:"column:consume_rule_summary"`
	ProbabilityRule             string    `gorm:"column:probability_rule"`
	ComplianceRuleSummary       string    `gorm:"column:compliance_rule_summary"`
	CirculationLimitSummary     string    `gorm:"column:circulation_limit_summary"`
	Status                      int32     `gorm:"column:status"`
	IsEnabled                   int32     `gorm:"column:is_enabled"`
	ConsumeType                 string    `gorm:"column:consume_type"`
	ConsumeAmount               int32     `gorm:"column:consume_amount"`
	QuotaPerMember              int32     `gorm:"column:quota_per_member"`
	DailyQuotaPerMember         int32     `gorm:"column:daily_quota_per_member"`
	EligibilityRuleJSON         string    `gorm:"column:eligibility_rule_json"`
	PlatformID                  int64     `gorm:"column:platform_id"`
	TenantID                    int64     `gorm:"column:tenant_id"`
	MerchantID                  int64     `gorm:"column:merchant_id"`
}

func (drawActivitySnapshot) TableName() string {
	return "sms_draw_activity"
}

type drawPoolSnapshot struct {
	ID              int64  `gorm:"column:id"`
	ActivityID      int64  `gorm:"column:activity_id"`
	PoolName        string `gorm:"column:pool_name"`
	ProbabilityRule string `gorm:"column:probability_rule"`
	Sort            int32  `gorm:"column:sort"`
	Status          int32  `gorm:"column:status"`
}

func (drawPoolSnapshot) TableName() string {
	return "sms_draw_pool"
}

type drawPoolTemplateSnapshot struct {
	ID             int64   `gorm:"column:id"`
	ActivityID     int64   `gorm:"column:activity_id"`
	PoolID         int64   `gorm:"column:pool_id"`
	TemplateID     int64   `gorm:"column:template_id"`
	SlotIndex      int32   `gorm:"column:slot_index"`
	Rarity         string  `gorm:"column:rarity"`
	Probability    float64 `gorm:"column:probability"`
	SaleLimit      int64   `gorm:"column:sale_limit"`
	RemainingLimit int64   `gorm:"column:remaining_limit"`
	ConfigLimit    int64   `gorm:"column:config_limit"`
	Status         int32   `gorm:"column:status"`
}

func (drawPoolTemplateSnapshot) TableName() string {
	return "sms_draw_pool_template"
}

type drawCardTemplateSnapshot struct {
	ID            int64  `gorm:"column:id"`
	TemplateCode  string `gorm:"column:template_code"`
	TemplateName  string `gorm:"column:template_name"`
	CardFaceImage string `gorm:"column:card_face_image"`
	Rarity        string `gorm:"column:rarity"`
	DisplayCopy   string `gorm:"column:display_copy"`
	Status        int32  `gorm:"column:status"`
}

func (drawCardTemplateSnapshot) TableName() string {
	return "sms_card_template"
}

type drawMemberInfoSnapshot struct {
	MemberID     int64 `gorm:"column:member_id"`
	LotteryTimes int32 `gorm:"column:lottery_times"`
	IsEnabled    int32 `gorm:"column:is_enabled"`
}

func (drawMemberInfoSnapshot) TableName() string {
	return "ums_member_info"
}

type drawMemberIdentitySnapshot struct {
	MemberID       int64      `gorm:"column:member_id"`
	RealNameStatus string     `gorm:"column:real_name_status"`
	RealNameMasked string     `gorm:"column:real_name_masked"`
	CredentialRef  string     `gorm:"column:credential_ref"`
	VerifiedAt     *time.Time `gorm:"column:verified_at"`
}

func (drawMemberIdentitySnapshot) TableName() string {
	return "ums_member_identity"
}

type drawParticipationRecordRow struct {
	ID                  int64      `gorm:"column:id"`
	ActivityID          int64      `gorm:"column:activity_id"`
	MemberID            int64      `gorm:"column:member_id"`
	RequestID           string     `gorm:"column:request_id"`
	Scope               string     `gorm:"column:scope"`
	EligibilitySnapshot string     `gorm:"column:eligibility_snapshot_json"`
	ConsumeType         string     `gorm:"column:consume_type"`
	ConsumeAmount       int32      `gorm:"column:consume_amount"`
	LotteryTimesBefore  int32      `gorm:"column:lottery_times_before"`
	LotteryTimesAfter   int32      `gorm:"column:lottery_times_after"`
	ResultType          string     `gorm:"column:result_type"`
	ResultStatus        string     `gorm:"column:result_status"`
	PoolID              int64      `gorm:"column:pool_id"`
	TemplateID          int64      `gorm:"column:template_id"`
	Rarity              string     `gorm:"column:rarity"`
	TraceID             string     `gorm:"column:trace_id"`
	FailureCode         string     `gorm:"column:failure_code"`
	FailureReason       string     `gorm:"column:failure_reason"`
	AssetInstanceID     int64      `gorm:"column:asset_instance_id"`
	AssetNo             string     `gorm:"column:asset_no"`
	AssetStatus         string     `gorm:"column:asset_status"`
	AssetCreatedAt      *time.Time `gorm:"column:asset_created_at"`
	CreateTime          time.Time  `gorm:"column:create_time"`
	UpdateTime          *time.Time `gorm:"column:update_time"`
}

func (drawParticipationRecordRow) TableName() string {
	return "sms_draw_participation_record"
}

type drawRecordDetailRow struct {
	ID                 int64      `gorm:"column:id"`
	ActivityID         int64      `gorm:"column:activity_id"`
	RequestID          string     `gorm:"column:request_id"`
	ResultType         string     `gorm:"column:result_type"`
	ResultStatus       string     `gorm:"column:result_status"`
	FailureCode        string     `gorm:"column:failure_code"`
	FailureReason      string     `gorm:"column:failure_reason"`
	PoolID             int64      `gorm:"column:pool_id"`
	TemplateID         int64      `gorm:"column:template_id"`
	TemplateName       string     `gorm:"column:template_name"`
	Rarity             string     `gorm:"column:rarity"`
	ConsumeAmount      int32      `gorm:"column:consume_amount"`
	LotteryTimesBefore int32      `gorm:"column:lottery_times_before"`
	LotteryTimesAfter  int32      `gorm:"column:lottery_times_after"`
	AssetInstanceID    int64      `gorm:"column:asset_instance_id"`
	AssetNo            string     `gorm:"column:asset_no"`
	AssetStatus        string     `gorm:"column:asset_status"`
	MintStatus         string     `gorm:"column:mint_status"`
	ChainStatus        string     `gorm:"column:chain_status"`
	AssetCreatedAt     *time.Time `gorm:"column:asset_created_at"`
	CreateTime         time.Time  `gorm:"column:create_time"`
}

type drawRecentWinRow struct {
	RecordID     int64     `gorm:"column:record_id"`
	MemberID     int64     `gorm:"column:member_id"`
	Nickname     string    `gorm:"column:nickname"`
	ResultStatus string    `gorm:"column:result_status"`
	TemplateName string    `gorm:"column:template_name"`
	Rarity       string    `gorm:"column:rarity"`
	CreateTime   time.Time `gorm:"column:create_time"`
}

type drawEligibilitySummary struct {
	Status               string `json:"status"`
	Code                 string `json:"code"`
	Message              string `json:"message"`
	NextAction           string `json:"nextAction"`
	RemainingLotteryTime int32  `json:"remainingLotteryTimes"`
	RealNameStatus       string `json:"realNameStatus"`
	RealNameStatusText   string `json:"realNameStatusText"`
	RealNameMasked       string `json:"realNameMasked"`
	CredentialRef        string `json:"credentialRef"`
	VerifiedAt           string `json:"verifiedAt"`
}

type eligibilityRuleConfig struct {
	MinimumLotteryTimes    int32    `json:"minimumLotteryTimes"`
	RequireMemberEnabled   bool     `json:"requireMemberEnabled"`
	RequiredRealNameStatus []string `json:"requiredRealNameStatus"`
	AllowedScopeTypes      []string `json:"allowedScopeTypes"`
}

func loadActivitySnapshot(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, activityID int64, forUpdate bool) (*drawActivitySnapshot, error) {
	var activity drawActivitySnapshot
	query := pkgscope.ApplyGovernanceScope(
		db.WithContext(ctx).Table(activity.TableName()).Where("is_deleted = 0"),
		current,
		"",
	).Where("id = ?", activityID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&activity).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(activity.ConsumeType) == "" {
		activity.ConsumeType = drawConsumeTypeLottery
	}
	if activity.ConsumeAmount <= 0 {
		activity.ConsumeAmount = 1
	}
	return &activity, nil
}

func loadMemberInfoSnapshot(ctx context.Context, db *gorm.DB, memberID int64, forUpdate bool) (*drawMemberInfoSnapshot, error) {
	var member drawMemberInfoSnapshot
	query := db.WithContext(ctx).Table(member.TableName()).Where("member_id = ?", memberID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func loadMemberIdentitySnapshot(ctx context.Context, db *gorm.DB, memberID int64) (*drawMemberIdentitySnapshot, error) {
	var identity drawMemberIdentitySnapshot
	err := db.WithContext(ctx).
		Table(identity.TableName()).
		Where("member_id = ? AND is_deleted = 0", memberID).
		Take(&identity).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return defaultMemberIdentitySnapshot(memberID), nil
	case err != nil:
		return nil, err
	default:
		return &identity, nil
	}
}

func defaultMemberIdentitySnapshot(memberID int64) *drawMemberIdentitySnapshot {
	return &drawMemberIdentitySnapshot{
		MemberID:       memberID,
		RealNameStatus: drawEligibilityNeedRealName,
	}
}

func loadPoolSnapshots(ctx context.Context, db *gorm.DB, activityID int64) ([]drawPoolSnapshot, error) {
	var pools []drawPoolSnapshot
	err := db.WithContext(ctx).
		Table(drawPoolSnapshot{}.TableName()).
		Where("activity_id = ? AND is_deleted = 0", activityID).
		Order("sort asc, id asc").
		Find(&pools).Error
	return pools, err
}

func loadPoolTemplates(ctx context.Context, db *gorm.DB, activityID int64, poolIDs []int64, forUpdate bool) ([]drawPoolTemplateSnapshot, error) {
	if len(poolIDs) == 0 {
		return []drawPoolTemplateSnapshot{}, nil
	}
	var rows []drawPoolTemplateSnapshot
	query := db.WithContext(ctx).
		Table(drawPoolTemplateSnapshot{}.TableName()).
		Where("activity_id = ? AND pool_id IN ? AND is_deleted = 0", activityID, poolIDs).
		Order("id asc")
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Find(&rows).Error
	return rows, err
}

func loadCardTemplates(ctx context.Context, db *gorm.DB, templateIDs []int64) (map[int64]drawCardTemplateSnapshot, error) {
	if len(templateIDs) == 0 {
		return map[int64]drawCardTemplateSnapshot{}, nil
	}
	var rows []drawCardTemplateSnapshot
	err := db.WithContext(ctx).
		Table(drawCardTemplateSnapshot{}.TableName()).
		Where("id IN ? AND is_deleted = 0", templateIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]drawCardTemplateSnapshot, len(rows))
	for _, row := range rows {
		result[row.ID] = row
	}
	return result, nil
}

func collectIDs[T ~int64](items []T) []int64 {
	result := make([]int64, 0, len(items))
	for _, item := range items {
		result = append(result, int64(item))
	}
	return result
}

func buildLandingPools(pools []drawPoolSnapshot, poolTemplates []drawPoolTemplateSnapshot, templateMap map[int64]drawCardTemplateSnapshot) []*smsclient.DrawLandingPoolPreview {
	templatesByPool := make(map[int64][]*smsclient.DrawLandingCardPreview)
	for _, item := range poolTemplates {
		template, ok := templateMap[item.TemplateID]
		if !ok {
			continue
		}
		templatesByPool[item.PoolID] = append(templatesByPool[item.PoolID], &smsclient.DrawLandingCardPreview{
			TemplateId:    template.ID,
			TemplateCode:  template.TemplateCode,
			TemplateName:  template.TemplateName,
			CardFaceImage: template.CardFaceImage,
			Rarity:        firstNonEmpty(strings.TrimSpace(item.Rarity), template.Rarity),
			DisplayCopy:   template.DisplayCopy,
			SlotIndex:     item.SlotIndex,
			Probability:   item.Probability,
		})
	}

	result := make([]*smsclient.DrawLandingPoolPreview, 0, len(pools))
	for _, pool := range pools {
		result = append(result, &smsclient.DrawLandingPoolPreview{
			PoolId:          pool.ID,
			PoolName:        pool.PoolName,
			ProbabilityRule: pool.ProbabilityRule,
			Cards:           templatesByPool[pool.ID],
		})
	}
	return result
}

func buildCardPreviews(templateMap map[int64]drawCardTemplateSnapshot) []*smsclient.DrawLandingCardPreview {
	previews := make([]*smsclient.DrawLandingCardPreview, 0, len(templateMap))
	for _, template := range templateMap {
		previews = append(previews, &smsclient.DrawLandingCardPreview{
			TemplateId:    template.ID,
			TemplateCode:  template.TemplateCode,
			TemplateName:  template.TemplateName,
			CardFaceImage: template.CardFaceImage,
			Rarity:        template.Rarity,
			DisplayCopy:   template.DisplayCopy,
		})
	}
	return previews
}

func buildEligibility(activity *drawActivitySnapshot, member *drawMemberInfoSnapshot, identity *drawMemberIdentitySnapshot, successCount int64, dailySuccessCount int64, hasAvailableInventory bool, loggedIn bool) drawEligibilitySummary {
	summary := drawEligibilitySummary{
		Status:               drawEligibilityEligible,
		Code:                 drawEligibilityEligible,
		Message:              "当前可参与抽卡",
		NextAction:           drawNextActionNone,
		RemainingLotteryTime: 0,
		RealNameStatus:       drawEligibilityNeedRealName,
		RealNameStatusText:   memberIdentityStatusText(drawEligibilityNeedRealName),
	}

	if identity != nil {
		summary.RealNameStatus = firstNonEmpty(strings.TrimSpace(identity.RealNameStatus), drawEligibilityNeedRealName)
		summary.RealNameStatusText = memberIdentityStatusText(summary.RealNameStatus)
		summary.RealNameMasked = strings.TrimSpace(identity.RealNameMasked)
		summary.CredentialRef = strings.TrimSpace(identity.CredentialRef)
		if identity.VerifiedAt != nil {
			summary.VerifiedAt = time_util.TimeToStr(*identity.VerifiedAt)
		}
	}

	if !loggedIn {
		summary.Status = drawEligibilityNeedLogin
		summary.Code = drawEligibilityNeedLogin
		summary.Message = "登录后可查看个人资格和参与记录"
		summary.NextAction = drawNextActionLogin
		return summary
	}

	if activity == nil {
		summary.Status = drawEligibilityActivityOffline
		summary.Code = drawEligibilityActivityOffline
		summary.Message = "活动不存在或已下线"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	now := time.Now()
	if activity.Status != 1 || activity.IsEnabled != 1 || now.Before(activity.StartTime) || now.After(activity.EndTime) {
		summary.Status = drawEligibilityActivityOffline
		summary.Code = drawEligibilityActivityOffline
		summary.Message = "活动暂不可参与，请稍后再试"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	if member == nil || member.IsEnabled != 1 {
		summary.Status = drawEligibilityMemberDisabled
		summary.Code = drawEligibilityMemberDisabled
		summary.Message = "当前会员状态不可参与活动"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	summary.RemainingLotteryTime = member.LotteryTimes
	rules := parseEligibilityRuleConfig(activity)
	if rules.RequireMemberEnabled && member.IsEnabled != 1 {
		summary.Status = drawEligibilityMemberDisabled
		summary.Code = drawEligibilityMemberDisabled
		summary.Message = "当前会员状态不可参与活动"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}
	if rules.MinimumLotteryTimes > 0 && member.LotteryTimes < rules.MinimumLotteryTimes {
		summary.Status = drawEligibilityQuotaExhausted
		summary.Code = drawEligibilityQuotaExhausted
		summary.Message = "当前剩余抽奖次数未达到参与门槛"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	if activity.QuotaPerMember > 0 && successCount >= int64(activity.QuotaPerMember) {
		summary.Status = drawEligibilityQuotaExhausted
		summary.Code = drawEligibilityQuotaExhausted
		summary.Message = "当前活动参与次数已用完"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	if activity.DailyQuotaPerMember > 0 && dailySuccessCount >= int64(activity.DailyQuotaPerMember) {
		summary.Status = drawEligibilityQuotaExhausted
		summary.Code = drawEligibilityQuotaExhausted
		summary.Message = "今日参与次数已用完"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	if member.LotteryTimes < activity.ConsumeAmount {
		summary.Status = drawEligibilityQuotaExhausted
		summary.Code = drawEligibilityQuotaExhausted
		summary.Message = "剩余抽奖次数不足"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	if !hasAvailableInventory {
		summary.Status = drawEligibilityInventoryExhausted
		summary.Code = drawEligibilityInventoryExhausted
		summary.Message = "当前卡池库存不足，请稍后再试"
		summary.NextAction = drawNextActionRetryLater
		return summary
	}

	return summary
}

func parseEligibilityRuleConfig(activity *drawActivitySnapshot) eligibilityRuleConfig {
	rules := eligibilityRuleConfig{RequireMemberEnabled: true}
	if activity == nil || strings.TrimSpace(activity.EligibilityRuleJSON) == "" {
		return rules
	}
	_ = json.Unmarshal([]byte(activity.EligibilityRuleJSON), &rules)
	return rules
}

func containsText(values []string, target string) bool {
	normalizedTarget := strings.TrimSpace(target)
	for _, value := range values {
		if strings.TrimSpace(value) == normalizedTarget {
			return true
		}
	}
	return false
}

func countConsumedRecords(ctx context.Context, db *gorm.DB, activityID int64, memberID int64) (int64, int64, error) {
	query := db.WithContext(ctx).Table(drawParticipationRecordRow{}.TableName()).
		Where("activity_id = ? AND member_id = ? AND is_deleted = 0 AND consume_amount > 0", activityID, memberID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, 0, err
	}

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	var daily int64
	err := db.WithContext(ctx).Table(drawParticipationRecordRow{}.TableName()).
		Where("activity_id = ? AND member_id = ? AND is_deleted = 0 AND consume_amount > 0 AND create_time >= ? AND create_time < ?", activityID, memberID, dayStart, dayEnd).
		Count(&daily).Error
	return total, daily, err
}

func hasAvailableInventory(poolTemplates []drawPoolTemplateSnapshot) bool {
	for _, item := range poolTemplates {
		if item.RemainingLimit > 0 {
			return true
		}
	}
	return false
}

func loadRecentWins(ctx context.Context, db *gorm.DB, activityID int64, limit int) ([]*smsclient.DrawLandingRecentWin, error) {
	var rows []drawRecentWinRow
	err := db.WithContext(ctx).
		Table("sms_draw_participation_record r").
		Select(`
			r.id AS record_id,
			r.member_id,
			COALESCE(mi.nickname, '') AS nickname,
			r.result_status,
			COALESCE(ct.template_name, '') AS template_name,
			r.rarity,
			r.create_time`).
		Joins("LEFT JOIN ums_member_info mi ON mi.member_id = r.member_id").
		Joins("LEFT JOIN sms_card_template ct ON ct.id = r.template_id AND ct.is_deleted = 0").
		Where("r.activity_id = ? AND r.is_deleted = 0 AND r.result_status = ?", activityID, drawResultStatusWon).
		Order("r.id desc").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]*smsclient.DrawLandingRecentWin, 0, len(rows))
	for _, row := range rows {
		result = append(result, &smsclient.DrawLandingRecentWin{
			RecordId:         row.RecordID,
			MemberId:         row.MemberID,
			MemberNameMasked: maskNickname(row.Nickname),
			ResultStatus:     row.ResultStatus,
			ResultStatusText: drawResultStatusText(row.ResultStatus),
			TemplateName:     row.TemplateName,
			Rarity:           row.Rarity,
			CreateTime:       time_util.TimeToStr(row.CreateTime),
		})
	}
	return result, nil
}

func queryMemberRecordList(ctx context.Context, db *gorm.DB, activityID int64, memberID int64, pageNum int32, pageSize int32) (int64, []*smsclient.DrawMemberRecordData, error) {
	if memberID <= 0 {
		return 0, []*smsclient.DrawMemberRecordData{}, nil
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	baseQuery := db.WithContext(ctx).
		Table("sms_draw_participation_record r").
		Where("r.activity_id = ? AND r.member_id = ? AND r.is_deleted = 0", activityID, memberID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var rows []drawRecordDetailRow
	err := db.WithContext(ctx).
		Table("sms_draw_participation_record r").
		Select(`
			r.id,
			r.activity_id,
			r.request_id,
			r.result_type,
			r.result_status,
			r.failure_code,
			r.failure_reason,
			r.pool_id,
			r.template_id,
			COALESCE(ct.template_name, '') AS template_name,
			r.rarity,
			r.consume_amount,
			r.lottery_times_before,
			r.lottery_times_after,
			r.asset_instance_id,
			r.asset_no,
			r.asset_status,
			COALESCE(ci.mint_status, '') AS mint_status,
			COALESCE(ci.chain_status, '') AS chain_status,
			r.asset_created_at,
			r.create_time`).
		Joins("LEFT JOIN sms_card_template ct ON ct.id = r.template_id AND ct.is_deleted = 0").
		Joins("LEFT JOIN sms_card_instance ci ON ci.id = r.asset_instance_id AND ci.is_deleted = 0").
		Where("r.activity_id = ? AND r.member_id = ? AND r.is_deleted = 0", activityID, memberID).
		Order("r.id desc").
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&rows).Error
	if err != nil {
		return 0, nil, err
	}

	result := make([]*smsclient.DrawMemberRecordData, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapDrawRecordDetail(&row))
	}
	return total, result, nil
}

func mapDrawRecordDetail(row *drawRecordDetailRow) *smsclient.DrawMemberRecordData {
	if row == nil {
		return &smsclient.DrawMemberRecordData{}
	}
	return &smsclient.DrawMemberRecordData{
		Id:                 row.ID,
		ActivityId:         row.ActivityID,
		RequestId:          row.RequestID,
		ResultType:         row.ResultType,
		ResultStatus:       row.ResultStatus,
		ResultStatusText:   drawResultStatusText(row.ResultStatus),
		FailureCode:        row.FailureCode,
		FailureReason:      row.FailureReason,
		PoolId:             row.PoolID,
		TemplateId:         row.TemplateID,
		TemplateName:       row.TemplateName,
		Rarity:             row.Rarity,
		ConsumeAmount:      row.ConsumeAmount,
		LotteryTimesBefore: row.LotteryTimesBefore,
		LotteryTimesAfter:  row.LotteryTimesAfter,
		AssetInstanceId:    row.AssetInstanceID,
		AssetNo:            row.AssetNo,
		AssetStatus:        row.AssetStatus,
		AssetStatusText:    cardAssetStatusText(row.AssetStatus, row.MintStatus, row.ChainStatus),
		AssetCreatedAt:     nullableTimeToStr(row.AssetCreatedAt),
		CreateTime:         time_util.TimeToStr(row.CreateTime),
	}
}

func loadRecordByRequest(ctx context.Context, db *gorm.DB, activityID int64, memberID int64, requestID string) (*drawParticipationRecordRow, error) {
	var row drawParticipationRecordRow
	err := db.WithContext(ctx).
		Table(row.TableName()).
		Where("activity_id = ? AND member_id = ? AND request_id = ? AND is_deleted = 0", activityID, memberID, strings.TrimSpace(requestID)).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func loadRecordDetailByID(ctx context.Context, db *gorm.DB, id int64) (*smsclient.DrawMemberRecordData, error) {
	var row drawRecordDetailRow
	err := db.WithContext(ctx).
		Table("sms_draw_participation_record r").
		Select(`
			r.id,
			r.activity_id,
			r.request_id,
			r.result_type,
			r.result_status,
			r.failure_code,
			r.failure_reason,
			r.pool_id,
			r.template_id,
			COALESCE(ct.template_name, '') AS template_name,
			r.rarity,
			r.consume_amount,
			r.lottery_times_before,
			r.lottery_times_after,
			r.asset_instance_id,
			r.asset_no,
			r.asset_status,
			COALESCE(ci.mint_status, '') AS mint_status,
			COALESCE(ci.chain_status, '') AS chain_status,
			r.asset_created_at,
			r.create_time`).
		Joins("LEFT JOIN sms_card_template ct ON ct.id = r.template_id AND ct.is_deleted = 0").
		Joins("LEFT JOIN sms_card_instance ci ON ci.id = r.asset_instance_id AND ci.is_deleted = 0").
		Where("r.id = ?", id).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return mapDrawRecordDetail(&row), nil
}

func chooseWinner(poolTemplates []drawPoolTemplateSnapshot) *drawPoolTemplateSnapshot {
	candidates := make([]drawPoolTemplateSnapshot, 0, len(poolTemplates))
	totalWeight := 0.0
	for _, item := range poolTemplates {
		if item.RemainingLimit <= 0 {
			continue
		}
		candidates = append(candidates, item)
		if item.Probability > 0 {
			totalWeight += item.Probability
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	if totalWeight <= 0 {
		return nil
	}

	loseWeight := 0.0
	if totalWeight < 1 {
		loseWeight = 1 - totalWeight
	}

	randomValue := rand.New(rand.NewSource(time.Now().UnixNano())).Float64() * (totalWeight + loseWeight)
	if randomValue >= totalWeight {
		return nil
	}

	cursor := 0.0
	for _, item := range candidates {
		if item.Probability <= 0 {
			continue
		}
		cursor += item.Probability
		if randomValue <= cursor {
			chosen := item
			return &chosen
		}
	}

	last := candidates[len(candidates)-1]
	return &last
}

func createParticipationRecord(ctx context.Context, tx *gorm.DB, record *drawParticipationRecordRow) error {
	if record == nil {
		return errors.New("参与记录不能为空")
	}
	if record.CreateTime.IsZero() {
		record.CreateTime = time.Now()
	}
	return tx.WithContext(ctx).Table(record.TableName()).Create(record).Error
}

func buildRejectedRecord(activity *drawActivitySnapshot, memberID int64, requestID string, eligibility drawEligibilitySummary) *drawParticipationRecordRow {
	snapshot, _ := json.Marshal(eligibility)
	return &drawParticipationRecordRow{
		ActivityID:          activity.ID,
		MemberID:            memberID,
		RequestID:           strings.TrimSpace(requestID),
		Scope:               buildScopeSnapshot(activity),
		EligibilitySnapshot: string(snapshot),
		ConsumeType:         firstNonEmpty(strings.TrimSpace(activity.ConsumeType), drawConsumeTypeLottery),
		ResultType:          drawResultTypeRejected,
		ResultStatus:        rejectedResultStatus(eligibility.Code),
		TraceID:             strings.TrimSpace(requestID),
		FailureCode:         eligibility.Code,
		FailureReason:       eligibility.Message,
	}
}

func buildWinningOrNotRecord(activity *drawActivitySnapshot, memberID int64, requestID string, eligibility drawEligibilitySummary, before int32, after int32, winner *drawPoolTemplateSnapshot, templateMap map[int64]drawCardTemplateSnapshot) *drawParticipationRecordRow {
	snapshot, _ := json.Marshal(eligibility)
	record := &drawParticipationRecordRow{
		ActivityID:          activity.ID,
		MemberID:            memberID,
		RequestID:           strings.TrimSpace(requestID),
		Scope:               buildScopeSnapshot(activity),
		EligibilitySnapshot: string(snapshot),
		ConsumeType:         firstNonEmpty(strings.TrimSpace(activity.ConsumeType), drawConsumeTypeLottery),
		ConsumeAmount:       activity.ConsumeAmount,
		LotteryTimesBefore:  before,
		LotteryTimesAfter:   after,
		TraceID:             strings.TrimSpace(requestID),
	}

	if winner == nil {
		record.ResultType = drawResultTypeNotWon
		record.ResultStatus = drawResultStatusNotWon
		return record
	}

	record.ResultType = drawResultTypeWon
	record.ResultStatus = drawResultStatusWon
	record.PoolID = winner.PoolID
	record.TemplateID = winner.TemplateID
	record.Rarity = firstNonEmpty(strings.TrimSpace(winner.Rarity), templateMap[winner.TemplateID].Rarity)
	return record
}

func rejectedResultStatus(code string) string {
	switch strings.TrimSpace(code) {
	case drawEligibilityNeedLogin:
		return drawResultStatusNeedLogin
	case drawEligibilityNeedRealName:
		return drawResultStatusNeedReal
	case drawEligibilityQuotaExhausted:
		return drawResultStatusQuota
	case drawEligibilityInventoryExhausted:
		return drawResultStatusInventory
	case drawEligibilityMemberDisabled:
		return drawResultStatusDisabled
	default:
		return drawResultStatusOffline
	}
}

func eligibilityNextAction(code string) string {
	switch strings.TrimSpace(code) {
	case drawEligibilityNeedLogin:
		return drawNextActionLogin
	case drawEligibilityNeedRealName:
		return drawNextActionRealName
	case drawEligibilityQuotaExhausted, drawEligibilityActivityOffline, drawEligibilityInventoryExhausted, drawEligibilityMemberDisabled:
		return drawNextActionRetryLater
	default:
		return drawNextActionNone
	}
}

func drawResultStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case drawResultStatusWon:
		return "已中奖待到账"
	case drawResultStatusNeedReal:
		return "待实名"
	case drawResultStatusNeedLogin:
		return "请先登录"
	case drawResultStatusQuota:
		return "资格不足"
	case drawResultStatusInventory:
		return "库存不足"
	case drawResultStatusDisabled, drawResultStatusOffline:
		return "暂不可参与"
	default:
		return "未中奖"
	}
}

func cardAssetStatusText(status string, mintStatus string, chainStatus string) string {
	return digitalcardmint.ResolveAssetStatusText(status, mintStatus, chainStatus)
}

func nullableTimeToStr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return time_util.TimeToStr(*value)
}

func memberIdentityStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case "verified":
		return "已实名"
	case "pending":
		return "审核中"
	case "rejected":
		return "实名失败"
	default:
		return "待实名"
	}
}

func buildScopeSnapshot(activity *drawActivitySnapshot) string {
	if activity == nil {
		return ""
	}
	return fmt.Sprintf("platform:%d,tenant:%d,merchant:%d", activity.PlatformID, activity.TenantID, activity.MerchantID)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func maskNickname(nickname string) string {
	trimmed := strings.TrimSpace(nickname)
	runes := []rune(trimmed)
	switch len(runes) {
	case 0:
		return "匿名用户"
	case 1:
		return string(runes[0]) + "*"
	case 2:
		return string(runes[0]) + "*"
	default:
		return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
	}
}

func logParticipationFailure(ctx context.Context, action string, payload interface{}, err error) error {
	if err != nil {
		logc.Errorf(ctx, "%s失败,参数:%+v,异常:%s", action, payload, err.Error())
	}
	return err
}
