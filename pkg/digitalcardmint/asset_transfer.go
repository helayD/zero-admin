package digitalcardmint

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

const (
	TransferRecipientStatusRegistered   = "registered"
	TransferRecipientStatusNeedRegister = "need_register"
)

type AssetTransferRecipientInput struct {
	AssetInstanceID   int64
	FromMemberID      int64
	RecipientMobile   string
	RegisterH5BaseURL string
}

type AssetTransferRecipientResult struct {
	RecipientStatus         string
	RecipientStatusText     string
	RecipientMemberID       int64
	RecipientNicknameMasked string
	RecipientMobileMasked   string
	RegisterURL             string
	CanTransfer             bool
	ActionHint              string
}

type TransferDigitalCardAssetInput struct {
	AssetInstanceID int64
	FromMemberID    int64
	RecipientMobile string
	RequestID       string
	TraceID         string
}

type TransferDigitalCardAssetResult struct {
	AssetInstanceID         int64
	FromMemberID            int64
	ToMemberID              int64
	RecipientNicknameMasked string
	RecipientMobileMasked   string
	TransferStatus          string
	TransferStatusText      string
}

type RequestDigitalCardAssetWithdrawInput struct {
	AssetInstanceID int64
	MemberID        int64
	Reason          string
	RequestID       string
	TraceID         string
}

type RequestDigitalCardAssetWithdrawResult struct {
	AssetInstanceID int64
	WithdrawStatus  string
	WithdrawText    string
}

type transferRecipientMemberRow struct {
	MemberID  int64  `gorm:"column:member_id"`
	Nickname  string `gorm:"column:nickname"`
	Mobile    string `gorm:"column:mobile"`
	IsEnabled int32  `gorm:"column:is_enabled"`
	IsDeleted int32  `gorm:"column:is_deleted"`
}

func (s *Service) ResolveAssetTransferRecipient(ctx context.Context, currentScope pkgscope.GovernanceScope, input AssetTransferRecipientInput) (*AssetTransferRecipientResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	mobile, err := normalizeTransferMobile(input.RecipientMobile)
	if err != nil {
		return nil, err
	}
	instance, err := s.loadCardInstance(ctx, s.DB, input.AssetInstanceID, false)
	if err != nil {
		return nil, err
	}
	if err = validateOwnedAssetAction(instance, currentScope, input.FromMemberID); err != nil {
		return nil, err
	}
	if err = validateTransferableAsset(ctx, s.DB, instance); err != nil {
		return nil, err
	}

	recipient, err := s.loadTransferRecipientByMobile(ctx, s.DB, mobile)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return &AssetTransferRecipientResult{
			RecipientStatus:       TransferRecipientStatusNeedRegister,
			RecipientStatusText:   "接收人未注册",
			RecipientMobileMasked: maskReceiverPhone(mobile),
			RegisterURL:           buildTransferRegisterURL(input.RegisterH5BaseURL, mobile, input.AssetInstanceID),
			CanTransfer:           false,
			ActionHint:            "请先引导对方完成注册",
		}, nil
	case err != nil:
		return nil, err
	}
	if recipient.MemberID == input.FromMemberID {
		return nil, errors.New("不能转赠给自己")
	}
	return &AssetTransferRecipientResult{
		RecipientStatus:         TransferRecipientStatusRegistered,
		RecipientStatusText:     "接收人已注册",
		RecipientMemberID:       recipient.MemberID,
		RecipientNicknameMasked: maskNickname(recipient.Nickname),
		RecipientMobileMasked:   maskReceiverPhone(recipient.Mobile),
		CanTransfer:             true,
		ActionHint:              "可直接转赠给该接收人",
	}, nil
}

func (s *Service) TransferDigitalCardAsset(ctx context.Context, currentScope pkgscope.GovernanceScope, input TransferDigitalCardAssetInput) (*TransferDigitalCardAssetResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	mobile, err := normalizeTransferMobile(input.RecipientMobile)
	if err != nil {
		return nil, err
	}

	var result *TransferDigitalCardAssetResult
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		instance, txErr := s.loadCardInstance(ctx, tx, input.AssetInstanceID, true)
		if txErr != nil {
			return txErr
		}
		if txErr = validateOwnedAssetAction(instance, currentScope, input.FromMemberID); txErr != nil {
			return txErr
		}
		if txErr = validateTransferableAsset(ctx, tx, instance); txErr != nil {
			return txErr
		}

		recipient, txErr := s.loadTransferRecipientByMobile(ctx, tx, mobile)
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return errors.New("接收人未注册，请先引导对方完成注册")
		}
		if txErr != nil {
			return txErr
		}
		if recipient.MemberID == input.FromMemberID {
			return errors.New("不能转赠给自己")
		}

		now := s.now()
		if txErr = tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND member_id = ? AND is_deleted = 0", instance.ID, input.FromMemberID).
			Updates(map[string]interface{}{
				"member_id":   recipient.MemberID,
				"update_by":   input.FromMemberID,
				"update_time": now,
			}).Error; txErr != nil {
			return txErr
		}
		taskUpdates := map[string]interface{}{
			"member_id":   recipient.MemberID,
			"update_time": now,
		}
		if s.hasColumn(tx, CardMintTaskRow{}.TableName(), "update_by") {
			taskUpdates["update_by"] = input.FromMemberID
		}
		if txErr = tx.WithContext(ctx).
			Table(CardMintTaskRow{}.TableName()).
			Where("asset_instance_id = ? AND is_deleted = 0", instance.ID).
			Updates(taskUpdates).Error; txErr != nil {
			return txErr
		}
		if txErr = s.resetPendingPhysicalFulfillmentForTransfer(ctx, tx, instance.ID, recipient.MemberID, input.FromMemberID, now); txErr != nil {
			return txErr
		}

		instance.MemberID = recipient.MemberID
		if txErr = s.appendAssetLogTx(ctx, tx, instance, instance.AssetStatus, instance.AssetStatus, OperationAssetTransferred, OperatorManual, firstNonEmpty(input.TraceID, instance.TraceID), "", "卡片已转赠给已注册接收人", map[string]interface{}{
			"assetInstanceId": instance.ID,
			"fromMemberId":    input.FromMemberID,
			"toMemberId":      recipient.MemberID,
			"recipientMobile": maskReceiverPhone(recipient.Mobile),
			"requestId":       firstNonEmpty(input.RequestID, instance.RequestID),
			"transferredAt":   now.Format("2006-01-02 15:04:05"),
		}); txErr != nil {
			return txErr
		}

		// Story 10.11 Follow-up: 转赠事务内给双方写站内消息，闭环 C 端体验。
		// 模板名/转赠人昵称缺失时回退到「数字卡片」/ 手机号脱敏，避免空文案。
		template, templateErr := s.loadCardTemplate(ctx, tx, instance.TemplateID)
		templateName := ""
		if templateErr == nil && template != nil {
			templateName = template.TemplateName
		}
		fromMember, senderErr := s.loadSenderMemberByID(ctx, tx, input.FromMemberID)
		if senderErr != nil {
			return fmt.Errorf("loadSenderMemberByID(id=%d): %w", input.FromMemberID, senderErr)
		}
		if txErr = s.appendTransferMessagesTx(ctx, tx, instance, fromMember, recipient, templateName, now); txErr != nil {
			return txErr
		}

		result = &TransferDigitalCardAssetResult{
			AssetInstanceID:         instance.ID,
			FromMemberID:            input.FromMemberID,
			ToMemberID:              recipient.MemberID,
			RecipientNicknameMasked: maskNickname(recipient.Nickname),
			RecipientMobileMasked:   maskReceiverPhone(recipient.Mobile),
			TransferStatus:          "transferred",
			TransferStatusText:      "转赠成功",
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) RequestDigitalCardAssetWithdraw(ctx context.Context, currentScope pkgscope.GovernanceScope, input RequestDigitalCardAssetWithdrawInput) (*RequestDigitalCardAssetWithdrawResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if input.AssetInstanceID <= 0 || input.MemberID <= 0 {
		return nil, errors.New("资产实例和会员不能为空")
	}
	reason := firstNonEmpty(input.Reason, "会员提交提现申请")

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		instance, err := s.loadCardInstance(ctx, tx, input.AssetInstanceID, true)
		if err != nil {
			return err
		}
		if err = validateOwnedAssetAction(instance, currentScope, input.MemberID); err != nil {
			return err
		}
		if err = validateAssetWithdrawable(instance); err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, instance.AssetStatus, instance.AssetStatus, OperationAssetWithdrawRequested, OperatorManual, firstNonEmpty(input.TraceID, instance.TraceID), "", reason, map[string]interface{}{
			"assetInstanceId": input.AssetInstanceID,
			"memberId":        input.MemberID,
			"requestId":       firstNonEmpty(input.RequestID, instance.RequestID),
			"requestedAt":     s.now().Format("2006-01-02 15:04:05"),
		})
	})
	if err != nil {
		return nil, err
	}
	return &RequestDigitalCardAssetWithdrawResult{
		AssetInstanceID: input.AssetInstanceID,
		WithdrawStatus:  "requested",
		WithdrawText:    "提现申请已提交",
	}, nil
}

func (s *Service) loadTransferRecipientByMobile(ctx context.Context, db *gorm.DB, mobile string) (*transferRecipientMemberRow, error) {
	var row transferRecipientMemberRow
	err := db.WithContext(ctx).
		Table("ums_member_info").
		Select("member_id, nickname, mobile, is_enabled, is_deleted").
		Where("mobile = ? AND is_deleted = 0 AND is_enabled = 1", mobile).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func validateOwnedAssetAction(instance *CardInstanceRow, currentScope pkgscope.GovernanceScope, memberID int64) error {
	if instance == nil || instance.ID <= 0 {
		return errors.New("资产实例不存在")
	}
	if memberID <= 0 {
		return errors.New("会员ID不能为空")
	}
	if instance.MemberID != memberID {
		return errors.New("无权操作他人的卡片")
	}
	if err := validateScope(instance.PlatformID, instance.TenantID, instance.MerchantID, currentScope); err != nil {
		return err
	}
	return nil
}

func validateTransferableAsset(ctx context.Context, db *gorm.DB, instance *CardInstanceRow) error {
	if err := validateAssetWithdrawable(instance); err != nil {
		return err
	}
	var fulfillment PhysicalFulfillmentRow
	err := db.WithContext(ctx).
		Table(fulfillment.TableName()).
		Where("asset_instance_id = ? AND is_deleted = 0", instance.ID).
		Take(&fulfillment).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return err
	}
	switch strings.TrimSpace(fulfillment.FulfillmentStatus) {
	case "", PhysicalFulfillmentStatusPendingAddress:
		return nil
	default:
		return errors.New("实体卡已进入履约流程，暂不支持转赠")
	}
}

func validateAssetWithdrawable(instance *CardInstanceRow) error {
	if strings.TrimSpace(instance.MintStatus) != MintStatusSuccess {
		return errors.New("卡片到账完成后才可操作")
	}
	if strings.TrimSpace(instance.DisplayStatus) == DisplayStatusOfflined ||
		strings.TrimSpace(instance.DisplayStatus) == DisplayStatusRecycled ||
		strings.TrimSpace(instance.ComplianceStatus) == ComplianceStatusRestricted ||
		strings.TrimSpace(instance.ComplianceStatus) == ComplianceStatusRecycleRequested ||
		strings.TrimSpace(instance.ComplianceStatus) == ComplianceStatusRecycled {
		return errors.New("卡片当前状态暂不可操作")
	}
	return nil
}

func (s *Service) resetPendingPhysicalFulfillmentForTransfer(ctx context.Context, tx *gorm.DB, assetInstanceID int64, toMemberID int64, operatorID int64, now time.Time) error {
	var row PhysicalFulfillmentRow
	err := tx.WithContext(ctx).
		Table(row.TableName()).
		Where("asset_instance_id = ? AND is_deleted = 0", assetInstanceID).
		Take(&row).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return err
	}
	fulfillmentStatus := strings.TrimSpace(row.FulfillmentStatus)
	if fulfillmentStatus != "" && fulfillmentStatus != PhysicalFulfillmentStatusPendingAddress {
		return errors.New("实体卡已进入履约流程，暂不支持转赠")
	}
	fromStatus := row.FulfillmentStatus
	row.MemberID = toMemberID
	row.AddressID = 0
	row.ReceiverName = ""
	row.ReceiverPhone = ""
	row.ReceiverProvince = ""
	row.ReceiverCity = ""
	row.ReceiverDistrict = ""
	row.ReceiverAddress = ""
	row.ReceiverPostalCode = ""
	row.UpdateBy = operatorID
	row.OperatorID = operatorID
	row.UpdateTime = &now
	if err = tx.WithContext(ctx).
		Table(row.TableName()).
		Where("id = ? AND is_deleted = 0", row.ID).
		Updates(map[string]interface{}{
			"member_id":            row.MemberID,
			"address_id":           row.AddressID,
			"receiver_name":        row.ReceiverName,
			"receiver_phone":       row.ReceiverPhone,
			"receiver_province":    row.ReceiverProvince,
			"receiver_city":        row.ReceiverCity,
			"receiver_district":    row.ReceiverDistrict,
			"receiver_address":     row.ReceiverAddress,
			"receiver_postal_code": row.ReceiverPostalCode,
			"operator_id":          row.OperatorID,
			"update_by":            row.UpdateBy,
			"update_time":          now,
		}).Error; err != nil {
		return err
	}
	return s.appendPhysicalFulfillmentLogTx(ctx, tx, &row, fromStatus, row.FulfillmentStatus, OperationAssetTransferred, OperatorManual, operatorID, "转赠后等待接收人确认地址", map[string]interface{}{
		"assetInstanceId": assetInstanceID,
		"toMemberId":      toMemberID,
	}, row.TraceID, row.RequestID)
}

func normalizeTransferMobile(value string) (string, error) {
	mobile := strings.TrimSpace(value)
	if mobile == "" {
		return "", errors.New("接收人手机号不能为空")
	}
	if len(mobile) != 11 {
		return "", errors.New("接收人手机号格式不正确")
	}
	for _, r := range mobile {
		if r < '0' || r > '9' {
			return "", errors.New("接收人手机号格式不正确")
		}
	}
	return mobile, nil
}

func buildTransferRegisterURL(baseURL string, mobile string, assetInstanceID int64) string {
	value := strings.TrimSpace(baseURL)
	if value == "" {
		value = "/h5/digitalCard/register"
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	query := parsed.Query()
	query.Set("mobile", mobile)
	query.Set("assetInstanceId", strconv.FormatInt(assetInstanceID, 10))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func maskNickname(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return string(runes[0]) + "*"
	}
	return string(runes[0]) + "*" + string(runes[len(runes)-1])
}
