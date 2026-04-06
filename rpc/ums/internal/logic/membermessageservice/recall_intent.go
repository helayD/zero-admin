package membermessageservicelogic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/ums/umsclient"
)

const (
	recallSourceMemberMessage = "member_message"

	recallTargetHome            = "home"
	recallTargetCart            = "cart"
	recallTargetOrderList       = "order_list"
	recallTargetOrderDetail     = "order_detail"
	recallTargetProductDetail   = "product_detail"
	recallTargetCouponList      = "coupon_list"
	recallTargetCouponCenter    = "coupon_center"
	recallTargetAfterSalesApply = "after_sales_apply"
	recallTargetActivity        = "activity"
	recallTargetSubject         = "subject"
	recallTargetPreferredArea   = "preferred_area"

	recallFailureInvalidPayload      = "invalid_payload"
	recallFailureUnsupportedTarget   = "unsupported_target"
	recallFailureResourceUnavailable = "resource_unavailable"
	recallFailureMinVersionUnmet     = "min_version_unmet"
	recallFailureLoginRequired       = "login_required"
)

type memberMessageRecord struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	MemberID       int64      `gorm:"column:member_id"`
	MessageType    int32      `gorm:"column:message_type"`
	Title          string     `gorm:"column:title"`
	Content        string     `gorm:"column:content"`
	ImageURL       string     `gorm:"column:image_url"`
	LinkType       string     `gorm:"column:link_type"`
	LinkID         string     `gorm:"column:link_id"`
	RelatedOrderID int64      `gorm:"column:related_order_id"`
	Status         int32      `gorm:"column:status"`
	ReadTime       *time.Time `gorm:"column:read_time"`
	CreateTime     time.Time  `gorm:"column:create_time"`
	PlatformID     int64      `gorm:"column:platform_id"`
	TenantID       int64      `gorm:"column:tenant_id"`
	MerchantID     int64      `gorm:"column:merchant_id"`
	IntentContract string     `gorm:"column:intent_contract"`
}

func (memberMessageRecord) TableName() string {
	return "ums_member_message"
}

type persistedRecallIntent struct {
	IntentType       string `json:"intentType"`
	TargetType       string `json:"targetType"`
	TargetID         int64  `json:"targetId,omitempty"`
	TargetTab        int32  `json:"targetTab,omitempty"`
	FallbackType     string `json:"fallbackType"`
	FallbackTargetID int64  `json:"fallbackTargetId,omitempty"`
	FallbackTab      int32  `json:"fallbackTab,omitempty"`
	RequiresAuth     bool   `json:"requiresAuth"`
	MinAppVersion    string `json:"minAppVersion,omitempty"`
	IntentID         string `json:"intentId,omitempty"`
	IssuedAt         string `json:"issuedAt,omitempty"`
	FailureReason    string `json:"failureReason,omitempty"`
	RecoveryHint     string `json:"recoveryHint,omitempty"`
	Blocked          bool   `json:"blocked,omitempty"`
	Source           string `json:"source,omitempty"`
}

func newPersistedRecallIntent(targetType string) persistedRecallIntent {
	return persistedRecallIntent{
		IntentType:   defaultIntentType(targetType),
		TargetType:   targetType,
		FallbackType: recallTargetHome,
		Source:       recallSourceMemberMessage,
	}
}

func normalizeMessageCreatePayload(in *umsclient.AddMemberMessageReq) (string, string, int64, string, error) {
	intent := resolveCreateIntent(in)
	intentJSON, err := encodePersistedRecallIntent(intent)
	if err != nil {
		return "", "", 0, "", err
	}

	linkType, linkID, relatedOrderID := deriveLegacyRoute(intent, in.RelatedOrderId, in.LinkType, in.LinkId)
	return linkType, linkID, relatedOrderID, intentJSON, nil
}

func resolveCreateIntent(in *umsclient.AddMemberMessageReq) persistedRecallIntent {
	if in.Intent != nil {
		return normalizeExplicitIntent(in.Intent, in.RelatedOrderId, in.LinkType, in.LinkId)
	}

	return resolveLegacyRecallIntent(
		in.MessageType,
		in.Title,
		in.Content,
		in.LinkType,
		in.LinkId,
		in.RelatedOrderId,
		0,
		time.Time{},
	)
}

func normalizeExplicitIntent(in *umsclient.MemberMessageRecallIntent, relatedOrderID int64, rawLinkType, rawLinkID string) persistedRecallIntent {
	targetType := normalizeTargetAlias(in.TargetType)
	if targetType == "" {
		targetType = normalizeLinkType(rawLinkType)
	}
	intent := newPersistedRecallIntent(targetType)
	intent.IntentType = defaultString(strings.TrimSpace(in.IntentType), defaultIntentType(targetType))
	intent.TargetID = in.TargetId
	intent.TargetTab = in.TargetTab
	intent.FallbackType = normalizeTargetAlias(defaultString(strings.TrimSpace(in.FallbackType), defaultFallbackType(targetType)))
	intent.FallbackTargetID = in.FallbackTargetId
	intent.FallbackTab = in.FallbackTab
	intent.RequiresAuth = in.RequiresAuth || defaultRequiresAuth(targetType)
	intent.MinAppVersion = strings.TrimSpace(in.MinAppVersion)
	intent.IntentID = strings.TrimSpace(in.IntentId)
	intent.IssuedAt = strings.TrimSpace(in.IssuedAt)
	intent.FailureReason = strings.TrimSpace(in.FailureReason)
	intent.RecoveryHint = strings.TrimSpace(in.RecoveryHint)
	intent.Blocked = in.Blocked
	intent.Source = defaultString(strings.TrimSpace(in.Source), recallSourceMemberMessage)

	if intent.TargetID <= 0 {
		switch targetType {
		case recallTargetOrderDetail, recallTargetAfterSalesApply:
			if relatedOrderID > 0 {
				intent.TargetID = relatedOrderID
			}
		}
	}
	if intent.FallbackTargetID <= 0 {
		switch intent.FallbackType {
		case recallTargetOrderDetail:
			intent.FallbackTargetID = intent.TargetID
		}
	}

	return finalizePersistedRecallIntent(intent)
}

func resolvePersistedIntent(record *memberMessageRecord) *umsclient.MemberMessageRecallIntent {
	intent := resolveIntentContract(record)
	intent.IntentID = defaultString(intent.IntentID, fmt.Sprintf("member_message:%d", record.ID))
	intent.IssuedAt = defaultString(intent.IssuedAt, record.CreateTime.Format("2006-01-02 15:04:05"))
	intent.Source = defaultString(intent.Source, recallSourceMemberMessage)

	return &umsclient.MemberMessageRecallIntent{
		IntentType:       intent.IntentType,
		TargetType:       intent.TargetType,
		TargetId:         intent.TargetID,
		TargetTab:        intent.TargetTab,
		FallbackType:     intent.FallbackType,
		FallbackTargetId: intent.FallbackTargetID,
		FallbackTab:      intent.FallbackTab,
		RequiresAuth:     intent.RequiresAuth,
		MinAppVersion:    intent.MinAppVersion,
		IntentId:         intent.IntentID,
		IssuedAt:         intent.IssuedAt,
		FailureReason:    intent.FailureReason,
		RecoveryHint:     intent.RecoveryHint,
		Blocked:          intent.Blocked,
		Source:           intent.Source,
	}
}

func resolveIntentContract(record *memberMessageRecord) persistedRecallIntent {
	trimmedContract := strings.TrimSpace(record.IntentContract)
	if trimmedContract != "" && trimmedContract != "{}" {
		var intent persistedRecallIntent
		if err := json.Unmarshal([]byte(trimmedContract), &intent); err == nil {
			intent = finalizePersistedRecallIntent(intent)
			if intent.TargetType != "" {
				return intent
			}
		}
	}

	return resolveLegacyRecallIntent(
		record.MessageType,
		record.Title,
		record.Content,
		record.LinkType,
		record.LinkID,
		record.RelatedOrderID,
		record.ID,
		record.CreateTime,
	)
}

func resolveLegacyRecallIntent(messageType int32, title, content, rawLinkType, rawLinkID string, relatedOrderID, messageID int64, issuedAt time.Time) persistedRecallIntent {
	linkType := normalizeLinkType(rawLinkType)
	intent := newPersistedRecallIntent(linkType)
	intent.Source = recallSourceMemberMessage
	if messageID > 0 {
		intent.IntentID = fmt.Sprintf("member_message:%d", messageID)
	}
	if !issuedAt.IsZero() {
		intent.IssuedAt = issuedAt.Format("2006-01-02 15:04:05")
	}

	switch linkType {
	case recallTargetOrderDetail:
		orderID := parsePositiveInt64(rawLinkID)
		if orderID <= 0 {
			orderID = relatedOrderID
		}
		if orderID <= 0 {
			intent.TargetType = recallTargetOrderDetail
			intent.FallbackType = recallTargetOrderList
			intent.FallbackTab = 1
			intent.RequiresAuth = true
			intent.FailureReason = recallFailureInvalidPayload
			intent.RecoveryHint = "订单入口缺少必要信息，已返回订单列表"
			intent.Blocked = true
			return finalizePersistedRecallIntent(intent)
		}
		intent.TargetID = orderID
		intent.FallbackType = recallTargetOrderList
		intent.FallbackTab = 1
		intent.RequiresAuth = true
	case recallTargetOrderList:
		intent.TargetTab = 0
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
		intent.RequiresAuth = true
	case recallTargetProductDetail:
		productID := parsePositiveInt64(rawLinkID)
		if productID <= 0 {
			intent.TargetType = recallTargetProductDetail
			intent.FallbackType = recallTargetHome
			intent.FallbackTab = 0
			intent.FailureReason = recallFailureInvalidPayload
			intent.RecoveryHint = "商品入口缺少必要信息，已返回首页继续浏览"
			intent.Blocked = true
			return finalizePersistedRecallIntent(intent)
		}
		intent.TargetID = productID
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
	case recallTargetCouponCenter:
		intent.FallbackType = recallTargetCouponList
		intent.FallbackTab = 0
		intent.RequiresAuth = true
	case recallTargetCouponList:
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
		intent.RequiresAuth = true
	case recallTargetAfterSalesApply:
		orderID := parsePositiveInt64(rawLinkID)
		if orderID <= 0 {
			orderID = relatedOrderID
		}
		if orderID <= 0 {
			intent.TargetType = recallTargetAfterSalesApply
			intent.FallbackType = recallTargetOrderList
			intent.FallbackTab = 1
			intent.RequiresAuth = true
			intent.FailureReason = recallFailureInvalidPayload
			intent.RecoveryHint = "售后入口缺少必要信息，已返回订单列表"
			intent.Blocked = true
			return finalizePersistedRecallIntent(intent)
		}
		intent.TargetID = orderID
		intent.FallbackType = recallTargetOrderDetail
		intent.FallbackTargetID = orderID
		intent.RequiresAuth = true
	case recallTargetCart:
		intent.TargetTab = 2
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
		intent.RequiresAuth = true
	case recallTargetHome:
		intent.TargetTab = 0
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
	case recallTargetActivity, recallTargetSubject, recallTargetPreferredArea:
		intent.TargetID = parsePositiveInt64(rawLinkID)
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
		intent.FailureReason = recallFailureUnsupportedTarget
		intent.RecoveryHint = "当前活动入口暂不可直达，已返回首页继续浏览"
		intent.Blocked = true
	default:
		intent = fallbackIntentForUnknownTarget(messageType, title, content, rawLinkID, relatedOrderID)
	}

	return finalizePersistedRecallIntent(intent)
}

func fallbackIntentForUnknownTarget(messageType int32, title, content, rawLinkID string, relatedOrderID int64) persistedRecallIntent {
	if strings.TrimSpace(rawLinkID) == "" && relatedOrderID > 0 {
		intent := newPersistedRecallIntent(recallTargetOrderDetail)
		intent.TargetID = relatedOrderID
		intent.FallbackType = recallTargetOrderList
		intent.FallbackTab = 1
		intent.RequiresAuth = true
		return finalizePersistedRecallIntent(intent)
	}

	if messageType == 4 {
		intent := newPersistedRecallIntent(recallTargetActivity)
		intent.TargetID = parsePositiveInt64(rawLinkID)
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
		intent.FailureReason = recallFailureUnsupportedTarget
		intent.RecoveryHint = "当前活动入口暂不可直达，已返回首页继续浏览"
		intent.Blocked = true
		return finalizePersistedRecallIntent(intent)
	}

	if isCouponCenterMessage(title, content) {
		intent := newPersistedRecallIntent(recallTargetCouponCenter)
		intent.FallbackType = recallTargetCouponList
		intent.FallbackTab = 0
		intent.RequiresAuth = true
		return finalizePersistedRecallIntent(intent)
	}

	intent := newPersistedRecallIntent(recallTargetHome)
	intent.TargetTab = 0
	intent.FallbackType = recallTargetHome
	intent.FallbackTab = 0
	intent.FailureReason = recallFailureInvalidPayload
	intent.RecoveryHint = "消息目标无效，已返回首页继续浏览"
	intent.Blocked = true
	return finalizePersistedRecallIntent(intent)
}

func finalizePersistedRecallIntent(intent persistedRecallIntent) persistedRecallIntent {
	intent.TargetType = normalizeTargetAlias(intent.TargetType)
	intent.IntentType = defaultString(strings.TrimSpace(intent.IntentType), defaultIntentType(intent.TargetType))
	intent.FallbackType = normalizeTargetAlias(defaultString(intent.FallbackType, defaultFallbackType(intent.TargetType)))
	intent.Source = defaultString(intent.Source, recallSourceMemberMessage)

	switch intent.TargetType {
	case recallTargetCouponCenter:
		intent.RequiresAuth = true
	case recallTargetCouponList, recallTargetOrderList, recallTargetOrderDetail, recallTargetAfterSalesApply, recallTargetCart:
		intent.RequiresAuth = true
	}

	if intent.FallbackType == recallTargetOrderDetail && intent.FallbackTargetID <= 0 {
		intent.FallbackTargetID = intent.TargetID
	}

	switch intent.TargetType {
	case recallTargetActivity, recallTargetSubject, recallTargetPreferredArea:
		intent.Blocked = true
		intent.FailureReason = defaultString(intent.FailureReason, recallFailureUnsupportedTarget)
		intent.RecoveryHint = defaultString(intent.RecoveryHint, "当前活动入口暂不可直达，已返回首页继续浏览")
		intent.FallbackType = recallTargetHome
		intent.FallbackTab = 0
	}

	if requiresPositiveTargetID(intent.TargetType) && intent.TargetID <= 0 {
		intent.Blocked = true
		intent.FailureReason = defaultString(intent.FailureReason, recallFailureInvalidPayload)
		intent.RecoveryHint = defaultString(intent.RecoveryHint, "消息目标缺少必要信息，已返回可用页面")
		intent.FallbackType = normalizeTargetAlias(defaultString(intent.FallbackType, defaultFallbackType(intent.TargetType)))
	}

	return intent
}

func deriveLegacyRoute(intent persistedRecallIntent, defaultRelatedOrderID int64, rawLinkType, rawLinkID string) (string, string, int64) {
	linkType := normalizeLinkType(rawLinkType)
	linkID := strings.TrimSpace(rawLinkID)
	relatedOrderID := defaultRelatedOrderID

	switch intent.TargetType {
	case recallTargetOrderDetail:
		linkType = "order"
		if intent.TargetID > 0 {
			linkID = strconv.FormatInt(intent.TargetID, 10)
			relatedOrderID = intent.TargetID
		}
	case recallTargetOrderList:
		linkType = "order_list"
		linkID = strconv.Itoa(int(intent.TargetTab))
	case recallTargetProductDetail:
		linkType = "product"
		if intent.TargetID > 0 {
			linkID = strconv.FormatInt(intent.TargetID, 10)
		}
	case recallTargetCouponList:
		linkType = "coupon"
		linkID = ""
	case recallTargetCouponCenter:
		linkType = "coupon_center"
		linkID = ""
	case recallTargetAfterSalesApply:
		linkType = "after_sales"
		if intent.TargetID > 0 {
			linkID = strconv.FormatInt(intent.TargetID, 10)
			relatedOrderID = intent.TargetID
		}
	case recallTargetCart:
		linkType = "cart"
		linkID = ""
	case recallTargetActivity:
		linkType = "activity"
		if intent.TargetID > 0 {
			linkID = strconv.FormatInt(intent.TargetID, 10)
		}
	case recallTargetSubject:
		linkType = "subject"
		if intent.TargetID > 0 {
			linkID = strconv.FormatInt(intent.TargetID, 10)
		}
	case recallTargetPreferredArea:
		linkType = "preferred_area"
		if intent.TargetID > 0 {
			linkID = strconv.FormatInt(intent.TargetID, 10)
		}
	case recallTargetHome:
		linkType = "home"
		linkID = ""
	}

	return linkType, linkID, relatedOrderID
}

func encodePersistedRecallIntent(intent persistedRecallIntent) (string, error) {
	encoded, err := json.Marshal(intent)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func normalizeTargetAlias(value string) string {
	switch normalizeLinkType(value) {
	case "order":
		return recallTargetOrderDetail
	case "product":
		return recallTargetProductDetail
	case "coupon":
		return recallTargetCouponList
	case "available_coupon", "coupon_receive_center":
		return recallTargetCouponCenter
	case "after_sale":
		return recallTargetAfterSalesApply
	default:
		return normalizeLinkType(value)
	}
}

func normalizeLinkType(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "home":
		return recallTargetHome
	case "cart":
		return recallTargetCart
	case "order", "order_detail":
		return recallTargetOrderDetail
	case "order_list":
		return recallTargetOrderList
	case "product", "product_detail":
		return recallTargetProductDetail
	case "coupon_list":
		return recallTargetCouponList
	case "coupon", "coupon_center", "coupon_receive_center", "available_coupon":
		if normalized == "coupon" {
			return "coupon"
		}
		return recallTargetCouponCenter
	case "after_sales", "after_sale", "after_sales_apply":
		return recallTargetAfterSalesApply
	case "activity":
		return recallTargetActivity
	case "subject":
		return recallTargetSubject
	case "preferred_area":
		return recallTargetPreferredArea
	default:
		return normalized
	}
}

func defaultFallbackType(targetType string) string {
	switch targetType {
	case recallTargetOrderDetail:
		return recallTargetOrderList
	case recallTargetAfterSalesApply:
		return recallTargetOrderDetail
	case recallTargetCouponCenter:
		return recallTargetCouponList
	case recallTargetCouponList, recallTargetProductDetail, recallTargetCart, recallTargetHome:
		return recallTargetHome
	case recallTargetOrderList:
		return recallTargetHome
	default:
		return recallTargetHome
	}
}

func defaultIntentType(targetType string) string {
	switch targetType {
	case recallTargetOrderDetail:
		return "order_recall"
	case recallTargetOrderList:
		return "order_list_recall"
	case recallTargetProductDetail:
		return "product_recall"
	case recallTargetCouponList:
		return "coupon_list_recall"
	case recallTargetCouponCenter:
		return "coupon_center_recall"
	case recallTargetAfterSalesApply:
		return "after_sales_recall"
	case recallTargetActivity:
		return "activity_recall"
	case recallTargetSubject:
		return "subject_recall"
	case recallTargetPreferredArea:
		return "preferred_area_recall"
	case recallTargetCart:
		return "cart_recall"
	default:
		return "message_recall"
	}
}

func defaultRequiresAuth(targetType string) bool {
	switch targetType {
	case recallTargetOrderDetail, recallTargetOrderList, recallTargetCouponList, recallTargetCouponCenter, recallTargetAfterSalesApply, recallTargetCart:
		return true
	default:
		return false
	}
}

func requiresPositiveTargetID(targetType string) bool {
	switch targetType {
	case recallTargetOrderDetail, recallTargetProductDetail, recallTargetAfterSalesApply:
		return true
	default:
		return false
	}
}

func isCouponCenterMessage(title, content string) bool {
	joined := strings.ToLower(strings.TrimSpace(title + " " + content))
	keywords := []string{"领券", "可领", "领取", "发券", "优惠券到账"}
	for _, keyword := range keywords {
		if strings.Contains(joined, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func parsePositiveInt64(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
