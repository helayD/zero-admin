package digitalcardmint

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OrderCardSummary 订单详情中展示的提货卡摘要（C 端安全字段）。
//
// Story 10.7：C 端订单详情页需要展示订单包含的提货卡，但严禁暴露任何区块链
// 底层信息。此结构体只承载 C 端可见字段；服务端在读取数据库时不返回 chainStatus /
// tokenId / chainTxId / lastReceiptJson 等敏感字段，避免上游误用。
type OrderCardSummary struct {
	AssetInstanceID  int64  `json:"assetInstanceId"`
	AssetNoMasked    string `json:"assetNoMasked"`
	TemplateID       int64  `json:"templateId"`
	TemplateName     string `json:"templateName"`
	TemplateImage    string `json:"templateImage"`
	MintStatus       string `json:"mintStatus"`
	MintStatusText   string `json:"mintStatusText"`
	DisplayStatus    string `json:"displayStatus"`
	ComplianceStatus string `json:"complianceStatus"`
	ComplianceTip    string `json:"complianceTip"`
	OrderItemID      int64  `json:"orderItemId"`
	IssuedAt         string `json:"issuedAt"`
}

type orderCardSummaryRow struct {
	AssetInstanceID  int64      `gorm:"column:asset_instance_id"`
	AssetNo          string     `gorm:"column:asset_no"`
	TemplateID       int64      `gorm:"column:template_id"`
	TemplateName     string     `gorm:"column:template_name"`
	TemplateImage    string     `gorm:"column:template_image"`
	MintStatus       string     `gorm:"column:mint_status"`
	DisplayStatus    string     `gorm:"column:display_status"`
	ComplianceStatus string     `gorm:"column:compliance_status"`
	ComplianceReason string     `gorm:"column:compliance_reason"`
	OrderItemID      int64      `gorm:"column:order_item_id"`
	IssuedAt         *time.Time `gorm:"column:issued_at"`
}

// LoadOrdersHasDigitalCards 批量查询订单是否含提货卡（Story 10.7）。
//
// 返回 map[orderID]bool，仅包含至少有一张卡片的订单。订单列表页可以一次性获取
// 全页订单的「含提货卡」标识，避免 N+1 查询。结果对 C 端安全（只返回布尔值）。
func (s *Service) LoadOrdersHasDigitalCards(ctx context.Context, orderIDs []int64, memberID int64) (map[int64]bool, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("digitalcardmint service 未初始化")
	}
	result := make(map[int64]bool, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}

	type batchRow struct {
		OrderID int64 `gorm:"column:order_id"`
	}
	var rows []batchRow
	query := s.DB.WithContext(ctx).
		Table("sms_card_instance AS instance").
		Select("DISTINCT item.order_id AS order_id").
		Joins("JOIN oms_order_item AS item ON item.id = instance.source_id AND item.is_deleted = 0").
		Where("instance.is_deleted = 0").
		Where("instance.source_type = ?", sourceTypePurchase).
		Where("item.order_id IN ?", orderIDs)
	if memberID > 0 {
		query = query.Where("instance.member_id = ?", memberID)
	}
	if err := query.Scan(&rows).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, nil
		}
		return nil, err
	}
	for _, row := range rows {
		result[row.OrderID] = true
	}
	return result, nil
}

// LoadOrderCardSummaries 查询订单关联的提货卡摘要列表（C 端安全）。
//
// 仅返回 C 端可见字段，绝不返回 chain_status / token_id / chain_tx_id /
// last_receipt_json 等链上字段。
func (s *Service) LoadOrderCardSummaries(ctx context.Context, orderID int64, memberID int64) ([]OrderCardSummary, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("digitalcardmint service 未初始化")
	}
	if orderID <= 0 {
		return nil, errors.New("订单ID不能为空")
	}

	var rows []orderCardSummaryRow
	query := s.DB.WithContext(ctx).
		Table("sms_card_instance AS instance").
		Select(`instance.id AS asset_instance_id,
			instance.asset_no AS asset_no,
			instance.template_id AS template_id,
			COALESCE(template.template_name, '') AS template_name,
			COALESCE(template.card_face_image, '') AS template_image,
			instance.mint_status AS mint_status,
			COALESCE(instance.display_status, '') AS display_status,
			COALESCE(instance.compliance_status, '') AS compliance_status,
			COALESCE(instance.compliance_reason, '') AS compliance_reason,
			instance.source_id AS order_item_id,
			instance.issued_at AS issued_at`).
		Joins("LEFT JOIN sms_card_template AS template ON template.id = instance.template_id AND template.is_deleted = 0").
		Joins("JOIN oms_order_item AS item ON item.id = instance.source_id AND item.is_deleted = 0").
		Where("instance.is_deleted = 0").
		Where("instance.source_type = ?", sourceTypePurchase).
		Where("item.order_id = ?", orderID)

	if memberID > 0 {
		query = query.Where("instance.member_id = ?", memberID)
	}

	if err := query.Order("instance.id ASC").Scan(&rows).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	summaries := make([]OrderCardSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, OrderCardSummary{
			AssetInstanceID:  row.AssetInstanceID,
			AssetNoMasked:    maskAssetNo(row.AssetNo),
			TemplateID:       row.TemplateID,
			TemplateName:     row.TemplateName,
			TemplateImage:    row.TemplateImage,
			MintStatus:       row.MintStatus,
			MintStatusText:   MintStatusConsumerText(row.MintStatus),
			DisplayStatus:    row.DisplayStatus,
			ComplianceStatus: row.ComplianceStatus,
			ComplianceTip:    consumerComplianceTip(row.ComplianceStatus, row.ComplianceReason),
			OrderItemID:      row.OrderItemID,
			IssuedAt:         formatIssuedAt(row.IssuedAt),
		})
	}
	return summaries, nil
}

// MintStatusConsumerText 将 mint_status 转换成 C 端口径文案。
//
// 严守 AGENTS.md 提货卡监管约束：禁止出现「链上 / 区块链 / AntChain / FISCO」
// 等关键词，也不暴露「补偿 / 复核」这类内部链路状态语义，全部归一到
// 「处理中 / 已到账 / 处理异常」三种用户视角。
func MintStatusConsumerText(status string) string {
	switch strings.TrimSpace(status) {
	case MintStatusPending:
		return "处理中"
	case MintStatusProcessing:
		return "处理中"
	case MintStatusCompensating:
		return "处理中"
	case MintStatusManualReview:
		return "审核中"
	case MintStatusSuccess:
		return "已到账"
	case MintStatusFailed:
		return "处理异常"
	case MintStatusFrozen:
		return "已冻结"
	default:
		return "处理中"
	}
}

// maskAssetNo 把卡片编号脱敏，仅展示前 4 位 + **** + 后 4 位。
func maskAssetNo(assetNo string) string {
	value := strings.TrimSpace(assetNo)
	if len(value) <= 8 {
		return value
	}
	return value[:4] + "****" + value[len(value)-4:]
}

// consumerComplianceTip 把合规状态映射为 C 端简短提示，去掉运营/链路术语。
func consumerComplianceTip(status string, reason string) string {
	trimmedReason := strings.TrimSpace(reason)
	switch strings.TrimSpace(status) {
	case ComplianceStatusClear, "":
		return ""
	case ComplianceStatusReview, ComplianceStatusManualReview:
		return "正在合规审核"
	case ComplianceStatusRestricted:
		return "已限制展示"
	case ComplianceStatusRecycleRequested, ComplianceStatusRecycled:
		return "卡片已回收"
	case ComplianceStatusFrozen:
		return "卡片已冻结"
	default:
		if trimmedReason != "" {
			return "合规处理中"
		}
		return ""
	}
}

func formatIssuedAt(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
