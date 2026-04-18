package cardassetservicelogic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	cardAssetOperatorSystem = "system"
	cardAssetOperatorJob    = "job"
	cardAssetOperatorManual = "manual"

	cardAssetStatusCreated = "asset_created"
	cardMintStatusPending  = "mint_pending"

	cardAssetOperationCreate = "asset_created"
	cardAssetReasonWon       = "participation_won"

	drawResultTypeWon   = "won"
	drawResultStatusWon = "won_pending_asset"

	maxAssetNoRetryCount = 8
)

var assetNoGenerator = generateAssetNo

type participationRecordSnapshot struct {
	ID              int64      `gorm:"column:id"`
	ActivityID      int64      `gorm:"column:activity_id"`
	MemberID        int64      `gorm:"column:member_id"`
	RequestID       string     `gorm:"column:request_id"`
	Scope           string     `gorm:"column:scope"`
	ResultType      string     `gorm:"column:result_type"`
	ResultStatus    string     `gorm:"column:result_status"`
	PoolID          int64      `gorm:"column:pool_id"`
	TemplateID      int64      `gorm:"column:template_id"`
	Rarity          string     `gorm:"column:rarity"`
	TraceID         string     `gorm:"column:trace_id"`
	AssetInstanceID int64      `gorm:"column:asset_instance_id"`
	AssetNo         string     `gorm:"column:asset_no"`
	AssetStatus     string     `gorm:"column:asset_status"`
	AssetCreatedAt  *time.Time `gorm:"column:asset_created_at"`
}

func (participationRecordSnapshot) TableName() string {
	return "sms_draw_participation_record"
}

type cardInstanceRow struct {
	ID                    int64      `gorm:"column:id"`
	PlatformID            int64      `gorm:"column:platform_id"`
	TenantID              int64      `gorm:"column:tenant_id"`
	MerchantID            int64      `gorm:"column:merchant_id"`
	ActivityID            int64      `gorm:"column:activity_id"`
	MemberID              int64      `gorm:"column:member_id"`
	ParticipationRecordID int64      `gorm:"column:participation_record_id"`
	RequestID             string     `gorm:"column:request_id"`
	TraceID               string     `gorm:"column:trace_id"`
	Scope                 string     `gorm:"column:scope"`
	PoolID                int64      `gorm:"column:pool_id"`
	TemplateID            int64      `gorm:"column:template_id"`
	Rarity                string     `gorm:"column:rarity"`
	AssetNo               string     `gorm:"column:asset_no"`
	AssetStatus           string     `gorm:"column:asset_status"`
	MintStatus            string     `gorm:"column:mint_status"`
	TokenID               string     `gorm:"column:token_id"`
	ChainStatus           string     `gorm:"column:chain_status"`
	LastReceiptAt         *time.Time `gorm:"column:last_receipt_at"`
	MintTaskID            int64      `gorm:"column:mint_task_id"`
	IssuedAt              *time.Time `gorm:"column:issued_at"`
	CreateBy              int64      `gorm:"column:create_by"`
	CreateTime            *time.Time `gorm:"column:create_time"`
	UpdateBy              *int64     `gorm:"column:update_by"`
	UpdateTime            *time.Time `gorm:"column:update_time"`
}

func (cardInstanceRow) TableName() string {
	return "sms_card_instance"
}

type cardAssetLogRow struct {
	ID                    int64  `gorm:"column:id"`
	AssetInstanceID       int64  `gorm:"column:asset_instance_id"`
	ParticipationRecordID int64  `gorm:"column:participation_record_id"`
	FromStatus            string `gorm:"column:from_status"`
	ToStatus              string `gorm:"column:to_status"`
	OperationType         string `gorm:"column:operation_type"`
	OperatorType          string `gorm:"column:operator_type"`
	TraceID               string `gorm:"column:trace_id"`
	ReasonCode            string `gorm:"column:reason_code"`
	ReasonText            string `gorm:"column:reason_text"`
	PayloadJSON           string `gorm:"column:payload_json"`
}

func (cardAssetLogRow) TableName() string {
	return "sms_card_asset_log"
}

type drawActivityScopeRow struct {
	ID         int64 `gorm:"column:id"`
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func (drawActivityScopeRow) TableName() string {
	return "sms_draw_activity"
}

type CardInstanceSnapshot struct {
	ID                    int64
	ActivityID            int64
	MemberID              int64
	ParticipationRecordID int64
	PoolID                int64
	TemplateID            int64
	RequestID             string
	TraceID               string
	Rarity                string
	AssetNo               string
	AssetStatus           string
	AssetStatusText       string
	MintStatus            string
	IssuedAt              string
}

func EnsureCardInstanceByParticipationRecord(ctx context.Context, tx *gorm.DB, participationRecordID int64, operatorType string, traceID string) (*CardInstanceSnapshot, error) {
	if participationRecordID <= 0 {
		return nil, errors.New("参与记录ID不能为空")
	}
	if tx == nil {
		return nil, errors.New("数据库事务不能为空")
	}
	record, err := loadParticipationRecord(ctx, tx, participationRecordID, true)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.ResultType) != drawResultTypeWon || strings.TrimSpace(record.ResultStatus) != drawResultStatusWon {
		return nil, errors.New("仅中奖且待建账记录允许创建资产实例")
	}

	instance, err := loadCardInstanceByParticipationRecord(ctx, tx, participationRecordID, true)
	switch {
	case err == nil:
		if err = ensureInitialAssetLog(ctx, tx, record, instance, normalizeOperatorType(operatorType), traceID); err != nil {
			return nil, err
		}
		if err = syncParticipationAssetSnapshot(ctx, tx, record.ID, instance); err != nil {
			return nil, err
		}
		return buildCardInstanceSnapshot(instance), nil
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	instance, err = createCardInstanceWithRetry(ctx, tx, record, normalizeOperatorType(operatorType), traceID)
	if err != nil {
		return nil, err
	}
	if err = ensureInitialAssetLog(ctx, tx, record, instance, normalizeOperatorType(operatorType), traceID); err != nil {
		return nil, err
	}
	if err = syncParticipationAssetSnapshot(ctx, tx, record.ID, instance); err != nil {
		return nil, err
	}
	return buildCardInstanceSnapshot(instance), nil
}

func QueryCardInstanceByParticipationRecord(ctx context.Context, db *gorm.DB, participationRecordID int64) (*CardInstanceSnapshot, error) {
	if participationRecordID <= 0 {
		return nil, errors.New("参与记录ID不能为空")
	}
	instance, err := loadCardInstanceByParticipationRecord(ctx, db, participationRecordID, false)
	if err != nil {
		return nil, err
	}
	return buildCardInstanceSnapshot(instance), nil
}

func BackfillWinningCardInstances(ctx context.Context, db *gorm.DB, currentScope pkgscope.GovernanceScope, activityID int64, memberID int64, limit int32, operatorType string, traceID string) (int32, []*CardInstanceSnapshot, error) {
	if db == nil {
		return 0, nil, errors.New("数据库不能为空")
	}
	if limit <= 0 {
		limit = 50
	}

	baseQuery := db.WithContext(ctx).
		Table(participationRecordSnapshot{}.TableName()+" AS record").
		Joins("JOIN "+drawActivityScopeRow{}.TableName()+" AS activity ON activity.id = record.activity_id AND activity.is_deleted = 0").
		Where("record.is_deleted = 0 AND record.result_type = ? AND record.result_status = ?", drawResultTypeWon, drawResultStatusWon)
	baseQuery = pkgscope.ApplyGovernanceScope(baseQuery, currentScope, "activity")
	if activityID > 0 {
		baseQuery = baseQuery.Where("record.activity_id = ?", activityID)
	}
	if memberID > 0 {
		baseQuery = baseQuery.Where("record.member_id = ?", memberID)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var ids []int64
	if err := baseQuery.Select("record.id").Order("record.id asc").Limit(int(limit)).Find(&ids).Error; err != nil {
		return 0, nil, err
	}

	result := make([]*CardInstanceSnapshot, 0, len(ids))
	for _, id := range ids {
		var asset *CardInstanceSnapshot
		if err := db.Transaction(func(tx *gorm.DB) error {
			var txErr error
			asset, txErr = EnsureCardInstanceByParticipationRecord(ctx, tx, id, normalizeOperatorType(operatorType), traceID)
			return txErr
		}); err != nil {
			return int32(total), nil, err
		}
		result = append(result, asset)
	}

	return int32(total), result, nil
}

func buildCardInstanceData(asset *CardInstanceSnapshot) *smsclient.CardInstanceData {
	if asset == nil {
		return &smsclient.CardInstanceData{}
	}
	return &smsclient.CardInstanceData{
		Id:                    asset.ID,
		ActivityId:            asset.ActivityID,
		MemberId:              asset.MemberID,
		ParticipationRecordId: asset.ParticipationRecordID,
		PoolId:                asset.PoolID,
		TemplateId:            asset.TemplateID,
		RequestId:             asset.RequestID,
		TraceId:               asset.TraceID,
		Rarity:                asset.Rarity,
		AssetNo:               asset.AssetNo,
		AssetStatus:           asset.AssetStatus,
		AssetStatusText:       asset.AssetStatusText,
		MintStatus:            asset.MintStatus,
		IssuedAt:              asset.IssuedAt,
	}
}

func buildCardInstanceSnapshot(instance *cardInstanceRow) *CardInstanceSnapshot {
	if instance == nil {
		return nil
	}
	issuedAt := ""
	if instance.IssuedAt != nil {
		issuedAt = time_util.TimeToStr(*instance.IssuedAt)
	}
	return &CardInstanceSnapshot{
		ID:                    instance.ID,
		ActivityID:            instance.ActivityID,
		MemberID:              instance.MemberID,
		ParticipationRecordID: instance.ParticipationRecordID,
		PoolID:                instance.PoolID,
		TemplateID:            instance.TemplateID,
		RequestID:             instance.RequestID,
		TraceID:               instance.TraceID,
		Rarity:                instance.Rarity,
		AssetNo:               instance.AssetNo,
		AssetStatus:           instance.AssetStatus,
		AssetStatusText:       cardAssetStatusText(instance.AssetStatus, instance.MintStatus, instance.ChainStatus),
		MintStatus:            instance.MintStatus,
		IssuedAt:              issuedAt,
	}
}

func loadParticipationRecord(ctx context.Context, db *gorm.DB, participationRecordID int64, forUpdate bool) (*participationRecordSnapshot, error) {
	var row participationRecordSnapshot
	query := db.WithContext(ctx).
		Table(row.TableName()).
		Where("id = ? AND is_deleted = 0", participationRecordID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func loadCardInstanceByParticipationRecord(ctx context.Context, db *gorm.DB, participationRecordID int64, forUpdate bool) (*cardInstanceRow, error) {
	var row cardInstanceRow
	query := db.WithContext(ctx).
		Table(row.TableName()).
		Where("participation_record_id = ? AND is_deleted = 0", participationRecordID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func syncParticipationAssetSnapshot(ctx context.Context, tx *gorm.DB, participationRecordID int64, instance *cardInstanceRow) error {
	if instance == nil {
		return errors.New("资产实例不能为空")
	}
	return tx.WithContext(ctx).
		Table(participationRecordSnapshot{}.TableName()).
		Where("id = ? AND is_deleted = 0", participationRecordID).
		Updates(map[string]interface{}{
			"asset_instance_id": instance.ID,
			"asset_no":          instance.AssetNo,
			"asset_status":      instance.AssetStatus,
			"asset_created_at":  instance.IssuedAt,
			"update_time":       time.Now(),
		}).Error
}

func ensureInitialAssetLog(ctx context.Context, tx *gorm.DB, record *participationRecordSnapshot, instance *cardInstanceRow, operatorType string, traceID string) error {
	if record == nil || instance == nil {
		return errors.New("记录或资产实例不能为空")
	}
	var count int64
	if err := tx.WithContext(ctx).
		Table(cardAssetLogRow{}.TableName()).
		Where("asset_instance_id = ? AND participation_record_id = ? AND to_status = ?", instance.ID, record.ID, instance.AssetStatus).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"assetNo":               instance.AssetNo,
		"activityId":            instance.ActivityID,
		"memberId":              instance.MemberID,
		"participationRecordId": instance.ParticipationRecordID,
		"requestId":             instance.RequestID,
	})
	logRow := &cardAssetLogRow{
		AssetInstanceID:       instance.ID,
		ParticipationRecordID: record.ID,
		FromStatus:            "",
		ToStatus:              instance.AssetStatus,
		OperationType:         cardAssetOperationCreate,
		OperatorType:          operatorType,
		TraceID:               firstNonEmpty(strings.TrimSpace(traceID), strings.TrimSpace(record.TraceID), strings.TrimSpace(record.RequestID)),
		ReasonCode:            cardAssetReasonWon,
		ReasonText:            "中奖后创建本地资产实例",
		PayloadJSON:           string(payload),
	}
	return tx.WithContext(ctx).Table(logRow.TableName()).Create(logRow).Error
}

func createCardInstanceWithRetry(ctx context.Context, tx *gorm.DB, record *participationRecordSnapshot, operatorType string, traceID string) (*cardInstanceRow, error) {
	platformID, tenantID, merchantID, err := resolveScopeFromRecord(ctx, tx, record)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < maxAssetNoRetryCount; attempt++ {
		issuedAt := time.Now()
		row := &cardInstanceRow{
			PlatformID:            platformID,
			TenantID:              tenantID,
			MerchantID:            merchantID,
			ActivityID:            record.ActivityID,
			MemberID:              record.MemberID,
			ParticipationRecordID: record.ID,
			RequestID:             strings.TrimSpace(record.RequestID),
			TraceID:               firstNonEmpty(strings.TrimSpace(traceID), strings.TrimSpace(record.TraceID), strings.TrimSpace(record.RequestID)),
			Scope:                 strings.TrimSpace(record.Scope),
			PoolID:                record.PoolID,
			TemplateID:            record.TemplateID,
			Rarity:                strings.TrimSpace(record.Rarity),
			AssetNo:               assetNoGenerator(issuedAt),
			AssetStatus:           cardAssetStatusCreated,
			MintStatus:            cardMintStatusPending,
			IssuedAt:              &issuedAt,
			CreateBy:              0,
		}
		err = tx.WithContext(ctx).
			Table(row.TableName()).
			Omit("create_time", "update_by", "update_time").
			Create(row).Error
		if err == nil {
			return row, nil
		}
		if !isDuplicateEntryError(err) {
			return nil, err
		}
		existing, existingErr := loadCardInstanceByParticipationRecord(ctx, tx, record.ID, false)
		if existingErr == nil {
			return existing, nil
		}
		if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
			return nil, existingErr
		}
	}
	return nil, errors.New("生成资产编号失败，请稍后重试")
}

func resolveScopeFromRecord(ctx context.Context, db *gorm.DB, record *participationRecordSnapshot) (int64, int64, int64, error) {
	if record == nil {
		return 0, 0, 0, errors.New("参与记录不能为空")
	}
	platformID, tenantID, merchantID, ok := parseScopeSnapshot(record.Scope)
	if ok {
		return platformID, tenantID, merchantID, nil
	}
	var activity drawActivityScopeRow
	if err := db.WithContext(ctx).
		Table(activity.TableName()).
		Select("id, platform_id, tenant_id, merchant_id").
		Where("id = ? AND is_deleted = 0", record.ActivityID).
		Take(&activity).Error; err != nil {
		return 0, 0, 0, err
	}
	return activity.PlatformID, activity.TenantID, activity.MerchantID, nil
}

func validateParticipationRecordScope(ctx context.Context, db *gorm.DB, currentScope pkgscope.GovernanceScope, participationRecordID int64, forUpdate bool) (*participationRecordSnapshot, error) {
	record, err := loadParticipationRecord(ctx, db, participationRecordID, forUpdate)
	if err != nil {
		return nil, err
	}

	platformID, tenantID, merchantID, err := resolveScopeFromRecord(ctx, db, record)
	if err != nil {
		return nil, err
	}
	if err = pkgscope.EnsureScopeMatch(currentScope, platformID, tenantID, merchantID, "当前主体无权操作该资产记录"); err != nil {
		return nil, err
	}

	return record, nil
}

func parseScopeSnapshot(scope string) (int64, int64, int64, bool) {
	var platformID int64
	var tenantID int64
	var merchantID int64
	parts := strings.Split(strings.TrimSpace(scope), ",")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(pair) != 2 {
			continue
		}
		switch strings.TrimSpace(pair[0]) {
		case "platform":
			_, _ = fmt.Sscan(strings.TrimSpace(pair[1]), &platformID)
		case "tenant":
			_, _ = fmt.Sscan(strings.TrimSpace(pair[1]), &tenantID)
		case "merchant":
			_, _ = fmt.Sscan(strings.TrimSpace(pair[1]), &merchantID)
		}
	}
	return platformID, tenantID, merchantID, platformID > 0
}

func generateAssetNo(now time.Time) string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("CARD%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
	}
	return fmt.Sprintf("CARD%s%s", now.Format("20060102150405"), strings.ToUpper(hex.EncodeToString(buf)))
}

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") || strings.Contains(message, "unique constraint failed")
}

func normalizeOperatorType(operatorType string) string {
	switch strings.TrimSpace(operatorType) {
	case cardAssetOperatorJob:
		return cardAssetOperatorJob
	case cardAssetOperatorManual:
		return cardAssetOperatorManual
	default:
		return cardAssetOperatorSystem
	}
}

func cardAssetStatusText(status string, mintStatus string, chainStatus string) string {
	return digitalcardmint.ResolveAssetStatusText(status, mintStatus, chainStatus)
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
