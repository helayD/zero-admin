package digitalcardmint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Story 10.11 Follow-up: 转赠成功后给转赠人 + 接收人写入站内消息，
// 走与订单/活动消息相同的 ums_member_message 通道。
//
// 设计原则：
//   - intent_contract 指向首页（home），避免依赖未实装的 digital_card_detail 路由；
//     等前端独立 Story 扩展 deep link 时再升级 target_type。
//   - 落表与转赠 UPDATE/asset_log 在同一事务，失败时整体回滚，避免消息和持卡人状态错位。
//   - 文案不暴露任何链上信息（chainTxId/tokenId 等），符合 C 端监管约束。

const (
	transferMessageType    int32  = 1
	transferLinkTypeHome   string = "home"
	transferIntentContract string = `{"intentType":"message_recall","targetType":"home","fallbackType":"home","requiresAuth":false,"source":"member_message"}`
)

type memberMessageRow struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement:true"`
	MemberID       int64     `gorm:"column:member_id"`
	MessageType    int32     `gorm:"column:message_type"`
	Title          string    `gorm:"column:title"`
	Content        string    `gorm:"column:content"`
	ImageURL       string    `gorm:"column:image_url"`
	LinkType       string    `gorm:"column:link_type"`
	LinkID         string    `gorm:"column:link_id"`
	RelatedOrderID int64     `gorm:"column:related_order_id"`
	IntentContract string    `gorm:"column:intent_contract"`
	Status         int32     `gorm:"column:status"`
	CreateTime     time.Time `gorm:"column:create_time"`
	PlatformID     int64     `gorm:"column:platform_id"`
	TenantID       int64     `gorm:"column:tenant_id"`
	MerchantID     int64     `gorm:"column:merchant_id"`
}

func (memberMessageRow) TableName() string {
	return "ums_member_message"
}

type senderMemberRow struct {
	MemberID int64  `gorm:"column:member_id"`
	Nickname string `gorm:"column:nickname"`
	Mobile   string `gorm:"column:mobile"`
}

func (s *Service) loadSenderMemberByID(ctx context.Context, db *gorm.DB, memberID int64) (*senderMemberRow, error) {
	var row senderMemberRow
	err := db.WithContext(ctx).
		Table("ums_member_info").
		Select("member_id, nickname, mobile").
		Where("member_id = ? AND is_deleted = 0", memberID).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// appendTransferMessagesTx 在转赠事务内给转赠人 + 接收人各写一条站内消息。
// templateName 为空时回退为「提货卡」（监管命名约束：C 端只能用「提货卡」，禁用「数字卡片」/「NFT」）。
func (s *Service) appendTransferMessagesTx(
	ctx context.Context,
	tx *gorm.DB,
	instance *CardInstanceRow,
	fromMember *senderMemberRow,
	recipient *transferRecipientMemberRow,
	templateName string,
	now time.Time,
) error {
	if instance == nil || fromMember == nil || recipient == nil {
		return errors.New("appendTransferMessagesTx: 参数不能为空")
	}
	displayName := templateName
	if displayName == "" {
		displayName = "提货卡"
	}

	platformID := instance.PlatformID
	if platformID <= 0 {
		platformID = 1
	}
	tenantID := instance.TenantID
	if tenantID <= 0 {
		tenantID = 1
	}

	fromNicknameMasked := maskNickname(fromMember.Nickname)
	if fromNicknameMasked == "" {
		fromNicknameMasked = maskReceiverPhone(fromMember.Mobile)
	}
	recipientNicknameMasked := maskNickname(recipient.Nickname)
	if recipientNicknameMasked == "" {
		recipientNicknameMasked = maskReceiverPhone(recipient.Mobile)
	}

	// 1) 接收人消息：「您收到一张新卡片」
	recipientMsg := &memberMessageRow{
		MemberID:       recipient.MemberID,
		MessageType:    transferMessageType,
		Title:          "您收到一张提货卡",
		Content:        fmt.Sprintf("%s 向您转赠了一张「%s」提货卡，可在「我的资产」查看。", fromNicknameMasked, displayName),
		LinkType:       transferLinkTypeHome,
		IntentContract: transferIntentContract,
		Status:         0,
		CreateTime:     now,
		PlatformID:     platformID,
		TenantID:       tenantID,
		MerchantID:     instance.MerchantID,
	}
	if err := tx.WithContext(ctx).Table(recipientMsg.TableName()).Create(recipientMsg).Error; err != nil {
		return fmt.Errorf("写入接收人转赠消息失败: %w", err)
	}

	// 2) 转赠人消息：「转赠成功」
	senderMsg := &memberMessageRow{
		MemberID:       fromMember.MemberID,
		MessageType:    transferMessageType,
		Title:          "转赠成功",
		Content:        fmt.Sprintf("您已将一张「%s」提货卡转赠给 %s（%s）。", displayName, recipientNicknameMasked, maskReceiverPhone(recipient.Mobile)),
		LinkType:       transferLinkTypeHome,
		IntentContract: transferIntentContract,
		Status:         0,
		CreateTime:     now,
		PlatformID:     platformID,
		TenantID:       tenantID,
		MerchantID:     instance.MerchantID,
	}
	if err := tx.WithContext(ctx).Table(senderMsg.TableName()).Create(senderMsg).Error; err != nil {
		return fmt.Errorf("写入转赠人转赠消息失败: %w", err)
	}

	return nil
}
