package digitalcardmint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	PhysicalFulfillmentStatusPendingDigitalConfirmation = "pending_digital_confirmation"
	PhysicalFulfillmentStatusPendingRealName            = "pending_real_name"
	PhysicalFulfillmentStatusPendingAddress             = "pending_address"
	PhysicalFulfillmentStatusAddressConfirmed           = "address_confirmed"
	PhysicalFulfillmentStatusProductionPending          = "production_pending"
	PhysicalFulfillmentStatusProductionInProgress       = "production_in_progress"
	PhysicalFulfillmentStatusQualityChecking            = "quality_checking"
	PhysicalFulfillmentStatusReadyToShip                = "ready_to_ship"
	PhysicalFulfillmentStatusShipped                    = "shipped"
	PhysicalFulfillmentStatusInTransit                  = "in_transit"
	PhysicalFulfillmentStatusSigned                     = "signed"
	PhysicalFulfillmentStatusException                  = "exception"
	PhysicalFulfillmentStatusReissuePending             = "reissue_pending"
	PhysicalFulfillmentStatusCancelled                  = "cancelled"

	PhysicalProductionStatusNotStarted      = "not_started"
	PhysicalProductionStatusQueued          = "queued"
	PhysicalProductionStatusPrinting        = "printing"
	PhysicalProductionStatusQualityChecking = "quality_checking"
	PhysicalProductionStatusCompleted       = "completed"
	PhysicalProductionStatusFailed          = "failed"

	PhysicalShippingStatusPending   = "pending"
	PhysicalShippingStatusShipped   = "shipped"
	PhysicalShippingStatusInTransit = "in_transit"
	PhysicalShippingStatusDelivered = "delivered"
	PhysicalShippingStatusSigned    = "signed"
	PhysicalShippingStatusException = "exception"

	PhysicalQualityStatusPending = "pending"
	PhysicalQualityStatusPassed  = "passed"
	PhysicalQualityStatusFailed  = "failed"

	PhysicalActionCreated           = "physical_fulfillment_created"
	PhysicalActionAddressConfirmed  = "physical_address_confirmed"
	PhysicalActionProductionUpdated = "physical_production_updated"
	PhysicalActionShipped           = "physical_shipped"
	PhysicalActionSigned            = "physical_signed"
	PhysicalActionException         = "physical_exception"
	PhysicalActionReissueRequested  = "physical_reissue_requested"
	PhysicalActionShippingFeePaid   = "physical_shipping_fee_paid"

	PhysicalShippingFeeStatusPending = "pending"
	PhysicalShippingFeeStatusPaid    = "paid"

	PhysicalBlockRealName           = "blocked_real_name"
	PhysicalBlockDigitalPending     = "blocked_digital_pending"
	PhysicalBlockManualReview       = "blocked_manual_review"
	PhysicalBlockCompliance         = "blocked_compliance"
	PhysicalBlockChannelUnavailable = "blocked_channel_unavailable"
)

type PhysicalFulfillmentRow struct {
	ID                    int64      `gorm:"column:id"`
	PlatformID            int64      `gorm:"column:platform_id"`
	TenantID              int64      `gorm:"column:tenant_id"`
	MerchantID            int64      `gorm:"column:merchant_id"`
	AssetInstanceID       int64      `gorm:"column:asset_instance_id"`
	ParticipationRecordID int64      `gorm:"column:participation_record_id"`
	ActivityID            int64      `gorm:"column:activity_id"`
	TemplateID            int64      `gorm:"column:template_id"`
	MemberID              int64      `gorm:"column:member_id"`
	AssetNo               string     `gorm:"column:asset_no"`
	FulfillmentNo         string     `gorm:"column:fulfillment_no"`
	FulfillmentStatus     string     `gorm:"column:fulfillment_status"`
	ProductionStatus      string     `gorm:"column:production_status"`
	QualityStatus         string     `gorm:"column:quality_status"`
	ShippingStatus        string     `gorm:"column:shipping_status"`
	AddressID             int64      `gorm:"column:address_id"`
	ReceiverName          string     `gorm:"column:receiver_name"`
	ReceiverPhone         string     `gorm:"column:receiver_phone"`
	ReceiverProvince      string     `gorm:"column:receiver_province"`
	ReceiverCity          string     `gorm:"column:receiver_city"`
	ReceiverDistrict      string     `gorm:"column:receiver_district"`
	ReceiverAddress       string     `gorm:"column:receiver_address"`
	ReceiverPostalCode    string     `gorm:"column:receiver_postal_code"`
	ProductionBatchNo     string     `gorm:"column:production_batch_no"`
	CarrierCode           string     `gorm:"column:carrier_code"`
	CarrierName           string     `gorm:"column:carrier_name"`
	TrackingNo            string     `gorm:"column:tracking_no"`
	ShippedAt             *time.Time `gorm:"column:shipped_at"`
	SignedAt              *time.Time `gorm:"column:signed_at"`
	AutoSignAt            *time.Time `gorm:"column:auto_sign_at"`
	FailureCode           string     `gorm:"column:failure_code"`
	FailureReason         string     `gorm:"column:failure_reason"`
	OperatorID            int64      `gorm:"column:operator_id"`
	TraceID               string     `gorm:"column:trace_id"`
	RequestID             string     `gorm:"column:request_id"`
	CreateBy              int64      `gorm:"column:create_by"`
	UpdateBy              int64      `gorm:"column:update_by"`
	CreateTime            time.Time  `gorm:"column:create_time"`
	UpdateTime            *time.Time `gorm:"column:update_time"`
	IsDeleted             int32      `gorm:"column:is_deleted"`
}

func (PhysicalFulfillmentRow) TableName() string {
	return "sms_card_physical_fulfillment"
}

type PhysicalFulfillmentLogRow struct {
	ID              int64     `gorm:"column:id"`
	FulfillmentID   int64     `gorm:"column:fulfillment_id"`
	AssetInstanceID int64     `gorm:"column:asset_instance_id"`
	Action          string    `gorm:"column:action"`
	FromStatus      string    `gorm:"column:from_status"`
	ToStatus        string    `gorm:"column:to_status"`
	OperatorType    string    `gorm:"column:operator_type"`
	OperatorID      int64     `gorm:"column:operator_id"`
	Reason          string    `gorm:"column:reason"`
	PayloadJSON     string    `gorm:"column:payload_json"`
	TraceID         string    `gorm:"column:trace_id"`
	RequestID       string    `gorm:"column:request_id"`
	CreateTime      time.Time `gorm:"column:create_time"`
}

func (PhysicalFulfillmentLogRow) TableName() string {
	return "sms_card_physical_fulfillment_log"
}

type memberAddressSnapshotRow struct {
	ID            int64  `gorm:"column:id"`
	MemberID      int64  `gorm:"column:member_id"`
	ReceiverName  string `gorm:"column:receiver_name"`
	ReceiverPhone string `gorm:"column:receiver_phone"`
	Province      string `gorm:"column:province"`
	City          string `gorm:"column:city"`
	District      string `gorm:"column:district"`
	DetailAddress string `gorm:"column:detail_address"`
	PostalCode    string `gorm:"column:postal_code"`
	IsDefault     int32  `gorm:"column:is_default"`
	IsDeleted     int32  `gorm:"column:is_deleted"`
}

func (memberAddressSnapshotRow) TableName() string {
	return "ums_member_address"
}

type PhysicalFulfillmentInput struct {
	AssetInstanceID int64
	AddressID       int64
	OperatorType    string
	OperatorID      int64
	RequestID       string
	TraceID         string
}

type ConfirmPhysicalFulfillmentAddressInput struct {
	FulfillmentID   int64
	AssetInstanceID int64
	AddressID       int64
	MemberID        int64
	RequestID       string
	TraceID         string
}

type UpdatePhysicalCardProductionStatusInput struct {
	FulfillmentID     int64
	OperatorID        int64
	ProductionBatchNo string
	ProductionStatus  string
	Reason            string
	RequestID         string
	TraceID           string
}

type ShipPhysicalCardInput struct {
	FulfillmentID int64
	OperatorID    int64
	CarrierCode   string
	CarrierName   string
	TrackingNo    string
	Reason        string
	RequestID     string
	TraceID       string
}

type ConfirmPhysicalCardReceiptInput struct {
	FulfillmentID int64
	MemberID      int64
	Reason        string
	RequestID     string
	TraceID       string
}

type ConfirmPhysicalFulfillmentShippingFeeInput struct {
	FulfillmentID   int64
	AssetInstanceID int64
	MemberID        int64
	PayAmount       int64
	PayChannel      string
	PaymentNo       string
	RequestID       string
	TraceID         string
}

type PhysicalFulfillmentExceptionInput struct {
	FulfillmentID int64
	OperatorID    int64
	FailureCode   string
	Reason        string
	RequestID     string
	TraceID       string
}

type PhysicalFulfillmentResult struct {
	FulfillmentID         int64
	AssetInstanceID       int64
	FulfillmentNo         string
	FulfillmentStatus     string
	FulfillmentStatusText string
	ProductionStatus      string
	ProductionStatusText  string
	ShippingStatus        string
	ShippingStatusText    string
	ShippingFeeStatus     string
	ShippingFeeStatusText string
	ShippingFeeAmount     int64
	BlockedReason         string
	BlockedReasonText     string
}

type MemberPhysicalFulfillmentDetail struct {
	FulfillmentID         int64
	FulfillmentNo         string
	AssetInstanceID       int64
	AssetNo               string
	TemplateName          string
	ActivityName          string
	ObtainedAt            string
	MintStatusText        string
	FulfillmentStatus     string
	FulfillmentStatusText string
	ProductionStatusText  string
	ShippingStatusText    string
	ShippingFeeStatus     string
	ShippingFeeStatusText string
	ShippingFeeAmount     int64
	ReceiverNameMasked    string
	ReceiverPhoneMasked   string
	AddressSummary        string
	CarrierName           string
	TrackingNo            string
	ComplianceTipSummary  string
	Timeline              []PhysicalFulfillmentTimelineItem
}

type PhysicalFulfillmentTimelineItem struct {
	Action     string
	ActionText string
	StatusText string
	Reason     string
	CreateTime string
}

type PhysicalFulfillmentFilter struct {
	PageNum           int32
	PageSize          int32
	ActivityID        int64
	TemplateID        int64
	MemberID          int64
	AssetNo           string
	FulfillmentNo     string
	ProductionBatchNo string
	FulfillmentStatus string
	ProductionStatus  string
	ShippingStatus    string
	TrackingNo        string
	FailureCode       string
	StartTime         string
	EndTime           string
}

type PhysicalFulfillmentItem struct {
	FulfillmentID         int64  `json:"fulfillmentId"`
	FulfillmentNo         string `json:"fulfillmentNo"`
	AssetInstanceID       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	ActivityID            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	TemplateID            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	MemberID              int64  `json:"memberId"`
	FulfillmentStatus     string `json:"fulfillmentStatus"`
	FulfillmentStatusText string `json:"fulfillmentStatusText"`
	ProductionStatus      string `json:"productionStatus"`
	ProductionStatusText  string `json:"productionStatusText"`
	ShippingStatus        string `json:"shippingStatus"`
	ShippingStatusText    string `json:"shippingStatusText"`
	ProductionBatchNo     string `json:"productionBatchNo"`
	CarrierName           string `json:"carrierName"`
	TrackingNo            string `json:"trackingNo"`
	FailureCode           string `json:"failureCode"`
	FailureReason         string `json:"failureReason"`
	ShippedAt             string `json:"shippedAt"`
	SignedAt              string `json:"signedAt"`
	UpdateTime            string `json:"updateTime"`
}

type PhysicalFulfillmentAdminDetail struct {
	Item           PhysicalFulfillmentItem           `json:"item"`
	ReceiverName   string                            `json:"receiverName"`
	ReceiverPhone  string                            `json:"receiverPhone"`
	AddressSummary string                            `json:"addressSummary"`
	RequestID      string                            `json:"requestId"`
	TraceID        string                            `json:"traceId"`
	Timeline       []PhysicalFulfillmentTimelineItem `json:"timeline"`
}

type physicalFulfillmentDetailRow struct {
	PhysicalFulfillmentRow
	TemplateName string       `gorm:"column:template_name"`
	ActivityName string       `gorm:"column:activity_name"`
	ObtainedAt   nullableTime `gorm:"column:obtained_at"`
	MintStatus   string       `gorm:"column:mint_status"`
}

func (s *Service) EnsurePhysicalFulfillmentByAsset(ctx context.Context, currentScope pkgscope.GovernanceScope, input PhysicalFulfillmentInput) (*PhysicalFulfillmentResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if input.AssetInstanceID <= 0 {
		return nil, errors.New("资产实例ID不能为空")
	}

	var result *PhysicalFulfillmentResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		instance, err := s.loadCardInstance(ctx, tx, input.AssetInstanceID, true)
		if err != nil {
			return err
		}
		if err = validateScope(instance.PlatformID, instance.TenantID, instance.MerchantID, currentScope); err != nil {
			return err
		}

		blockReason, blockText, err := s.physicalFulfillmentBlockReason(ctx, tx, instance)
		if err != nil {
			return err
		}
		if blockReason != "" {
			fulfillmentStatus := physicalFulfillmentStatusForBlockReason(blockReason)
			result = &PhysicalFulfillmentResult{
				AssetInstanceID:       input.AssetInstanceID,
				FulfillmentStatus:     fulfillmentStatus,
				FulfillmentStatusText: physicalFulfillmentStatusText(fulfillmentStatus),
				ShippingFeeStatus:     PhysicalShippingFeeStatusPending,
				ShippingFeeStatusText: "待发放后确认",
				BlockedReason:         blockReason,
				BlockedReasonText:     blockText,
			}
			return nil
		}

		if existing, findErr := s.loadPhysicalFulfillmentByAsset(ctx, tx, input.AssetInstanceID, true); findErr == nil {
			result = buildPhysicalFulfillmentResult(existing)
			return nil
		} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		now := s.now()
		row := &PhysicalFulfillmentRow{
			PlatformID:            instance.PlatformID,
			TenantID:              instance.TenantID,
			MerchantID:            instance.MerchantID,
			AssetInstanceID:       instance.ID,
			ParticipationRecordID: instance.ParticipationRecordID,
			ActivityID:            instance.ActivityID,
			TemplateID:            instance.TemplateID,
			MemberID:              instance.MemberID,
			AssetNo:               instance.AssetNo,
			FulfillmentNo:         buildPhysicalFulfillmentNo(instance.ID, now),
			FulfillmentStatus:     PhysicalFulfillmentStatusPendingAddress,
			ProductionStatus:      PhysicalProductionStatusNotStarted,
			QualityStatus:         PhysicalQualityStatusPending,
			ShippingStatus:        PhysicalShippingStatusPending,
			OperatorID:            input.OperatorID,
			TraceID:               firstNonEmpty(input.TraceID, instance.TraceID),
			RequestID:             firstNonEmpty(input.RequestID, instance.RequestID),
			CreateBy:              input.OperatorID,
			UpdateBy:              input.OperatorID,
			CreateTime:            now,
			UpdateTime:            &now,
		}

		var address *memberAddressSnapshotRow
		if input.AddressID > 0 {
			address, err = s.loadMemberAddress(ctx, tx, instance.MemberID, input.AddressID)
			if err != nil {
				return err
			}
			applyPhysicalAddressSnapshot(row, address)
			row.FulfillmentStatus = PhysicalFulfillmentStatusProductionPending
			row.ProductionStatus = PhysicalProductionStatusQueued
		}

		if err = tx.WithContext(ctx).Table(row.TableName()).Create(row).Error; err != nil {
			return err
		}
		if err = s.appendPhysicalFulfillmentLogTx(ctx, tx, row, "", row.FulfillmentStatus, PhysicalActionCreated, normalizeOperatorType(input.OperatorType), input.OperatorID, "创建实体卡履约单", map[string]interface{}{
			"assetInstanceId": input.AssetInstanceID,
			"addressId":       input.AddressID,
		}, input.TraceID, input.RequestID); err != nil {
			return err
		}
		if address != nil {
			if err = s.appendPhysicalFulfillmentLogTx(ctx, tx, row, PhysicalFulfillmentStatusPendingAddress, row.FulfillmentStatus, PhysicalActionAddressConfirmed, normalizeOperatorType(input.OperatorType), input.OperatorID, "确认收货地址", map[string]interface{}{
				"addressId": address.ID,
			}, input.TraceID, input.RequestID); err != nil {
				return err
			}
		}

		result = buildPhysicalFulfillmentResult(row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = &PhysicalFulfillmentResult{AssetInstanceID: input.AssetInstanceID}
	}
	return result, nil
}

func (s *Service) ConfirmPhysicalFulfillmentAddress(ctx context.Context, currentScope pkgscope.GovernanceScope, input ConfirmPhysicalFulfillmentAddressInput) (*PhysicalFulfillmentResult, error) {
	if input.AddressID <= 0 || input.MemberID <= 0 {
		return nil, errors.New("会员和地址不能为空")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, input.AssetInstanceID, "确认收货地址", func(tx *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		if row.MemberID != input.MemberID {
			return "", "", 0, "", nil, errors.New("无权确认他人的实体卡履约地址")
		}
		address, err := s.loadMemberAddress(ctx, tx, input.MemberID, input.AddressID)
		if err != nil {
			return "", "", 0, "", nil, err
		}
		fromStatus := row.FulfillmentStatus
		applyPhysicalAddressSnapshot(row, address)
		row.FulfillmentStatus = PhysicalFulfillmentStatusProductionPending
		row.ProductionStatus = PhysicalProductionStatusQueued
		row.ShippingStatus = PhysicalShippingStatusPending
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return fromStatus, PhysicalActionAddressConfirmed, input.MemberID, "确认收货地址", map[string]interface{}{"addressId": input.AddressID}, nil
	})
}

func (s *Service) UpdatePhysicalCardProductionStatus(ctx context.Context, currentScope pkgscope.GovernanceScope, input UpdatePhysicalCardProductionStatusInput) (*PhysicalFulfillmentResult, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("制作状态更新原因不能为空")
	}
	nextProduction := strings.TrimSpace(input.ProductionStatus)
	if !isValidPhysicalProductionStatus(nextProduction) {
		return nil, errors.New("制作状态非法")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, 0, input.Reason, func(_ *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		fromStatus := row.FulfillmentStatus
		row.ProductionBatchNo = firstNonEmpty(input.ProductionBatchNo, row.ProductionBatchNo)
		row.ProductionStatus = nextProduction
		switch nextProduction {
		case PhysicalProductionStatusQueued:
			row.FulfillmentStatus = PhysicalFulfillmentStatusProductionPending
		case PhysicalProductionStatusPrinting:
			row.FulfillmentStatus = PhysicalFulfillmentStatusProductionInProgress
		case PhysicalProductionStatusQualityChecking:
			row.FulfillmentStatus = PhysicalFulfillmentStatusQualityChecking
			row.QualityStatus = PhysicalQualityStatusPending
		case PhysicalProductionStatusCompleted:
			row.FulfillmentStatus = PhysicalFulfillmentStatusReadyToShip
			row.QualityStatus = PhysicalQualityStatusPassed
		case PhysicalProductionStatusFailed:
			row.FulfillmentStatus = PhysicalFulfillmentStatusException
			row.QualityStatus = PhysicalQualityStatusFailed
			row.FailureCode = "production_failed"
			row.FailureReason = strings.TrimSpace(input.Reason)
		}
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return fromStatus, PhysicalActionProductionUpdated, input.OperatorID, input.Reason, map[string]interface{}{
			"productionStatus":  nextProduction,
			"productionBatchNo": row.ProductionBatchNo,
		}, nil
	})
}

func (s *Service) ShipPhysicalCard(ctx context.Context, currentScope pkgscope.GovernanceScope, input ShipPhysicalCardInput) (*PhysicalFulfillmentResult, error) {
	if strings.TrimSpace(input.CarrierCode) == "" || strings.TrimSpace(input.TrackingNo) == "" {
		return nil, errors.New("承运商和物流单号不能为空")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("发货原因不能为空")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, 0, input.Reason, func(_ *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		if strings.TrimSpace(row.ReceiverPhone) == "" || strings.TrimSpace(row.ReceiverAddress) == "" {
			return "", "", 0, "", nil, errors.New("发货前必须确认收货地址")
		}
		if row.FulfillmentStatus == PhysicalFulfillmentStatusShipped &&
			row.CarrierCode == strings.TrimSpace(input.CarrierCode) &&
			row.TrackingNo == strings.TrimSpace(input.TrackingNo) {
			return row.FulfillmentStatus, "", input.OperatorID, input.Reason, nil, nil
		}
		if row.ProductionStatus != PhysicalProductionStatusCompleted && row.FulfillmentStatus != PhysicalFulfillmentStatusReadyToShip {
			return "", "", 0, "", nil, errors.New("实体卡未完成制作，不能发货")
		}
		now := s.now()
		fromStatus := row.FulfillmentStatus
		row.CarrierCode = strings.TrimSpace(input.CarrierCode)
		row.CarrierName = strings.TrimSpace(input.CarrierName)
		row.TrackingNo = strings.TrimSpace(input.TrackingNo)
		row.FulfillmentStatus = PhysicalFulfillmentStatusShipped
		row.ShippingStatus = PhysicalShippingStatusShipped
		row.ShippedAt = &now
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return fromStatus, PhysicalActionShipped, input.OperatorID, input.Reason, map[string]interface{}{
			"carrierCode": row.CarrierCode,
			"carrierName": row.CarrierName,
			"trackingNo":  row.TrackingNo,
		}, nil
	})
}

func (s *Service) ConfirmPhysicalCardReceipt(ctx context.Context, currentScope pkgscope.GovernanceScope, input ConfirmPhysicalCardReceiptInput) (*PhysicalFulfillmentResult, error) {
	if input.MemberID <= 0 {
		return nil, errors.New("会员ID不能为空")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, 0, firstNonEmpty(input.Reason, "会员确认收货"), func(_ *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		if row.MemberID != input.MemberID {
			return "", "", 0, "", nil, errors.New("无权确认他人的实体卡签收")
		}
		if row.FulfillmentStatus == PhysicalFulfillmentStatusSigned {
			return row.FulfillmentStatus, "", input.MemberID, input.Reason, nil, nil
		}
		if row.FulfillmentStatus != PhysicalFulfillmentStatusShipped && row.FulfillmentStatus != PhysicalFulfillmentStatusInTransit {
			return "", "", 0, "", nil, errors.New("当前状态不允许确认签收")
		}
		now := s.now()
		fromStatus := row.FulfillmentStatus
		row.FulfillmentStatus = PhysicalFulfillmentStatusSigned
		row.ShippingStatus = PhysicalShippingStatusSigned
		row.SignedAt = &now
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return fromStatus, PhysicalActionSigned, input.MemberID, firstNonEmpty(input.Reason, "会员确认收货"), nil, nil
	})
}

func (s *Service) ConfirmPhysicalFulfillmentShippingFee(ctx context.Context, currentScope pkgscope.GovernanceScope, input ConfirmPhysicalFulfillmentShippingFeeInput) (*PhysicalFulfillmentResult, error) {
	if input.MemberID <= 0 {
		return nil, errors.New("会员ID不能为空")
	}
	if input.PayAmount < 0 {
		return nil, errors.New("邮费金额不能小于0")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, input.AssetInstanceID, "确认邮费支付", func(tx *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		if row.MemberID != input.MemberID {
			return "", "", 0, "", nil, errors.New("无权确认他人的实体卡邮费")
		}
		if row.FulfillmentStatus == PhysicalFulfillmentStatusShipped ||
			row.FulfillmentStatus == PhysicalFulfillmentStatusInTransit ||
			row.FulfillmentStatus == PhysicalFulfillmentStatusSigned {
			return "", "", 0, "", nil, errors.New("实体卡已进入配送或签收，不能重复支付邮费")
		}
		var paidCount int64
		if err := tx.WithContext(ctx).
			Table(PhysicalFulfillmentLogRow{}.TableName()).
			Where("fulfillment_id = ? AND action = ?", row.ID, PhysicalActionShippingFeePaid).
			Count(&paidCount).Error; err != nil {
			return "", "", 0, "", nil, err
		}
		if paidCount > 0 {
			return row.FulfillmentStatus, "", input.MemberID, "邮费已支付", nil, nil
		}
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return row.FulfillmentStatus, PhysicalActionShippingFeePaid, input.MemberID, "确认邮费支付", map[string]interface{}{
			"payAmount":  input.PayAmount,
			"payChannel": strings.TrimSpace(input.PayChannel),
			"paymentNo":  strings.TrimSpace(input.PaymentNo),
		}, nil
	})
}

func (s *Service) MarkPhysicalFulfillmentException(ctx context.Context, currentScope pkgscope.GovernanceScope, input PhysicalFulfillmentExceptionInput) (*PhysicalFulfillmentResult, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("异常原因不能为空")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, 0, input.Reason, func(_ *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		fromStatus := row.FulfillmentStatus
		row.FulfillmentStatus = PhysicalFulfillmentStatusException
		row.ShippingStatus = PhysicalShippingStatusException
		row.FailureCode = firstNonEmpty(input.FailureCode, row.FailureCode)
		row.FailureReason = strings.TrimSpace(input.Reason)
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return fromStatus, PhysicalActionException, input.OperatorID, input.Reason, map[string]interface{}{
			"failureCode": row.FailureCode,
		}, nil
	})
}

func (s *Service) RequestPhysicalCardReissue(ctx context.Context, currentScope pkgscope.GovernanceScope, input PhysicalFulfillmentExceptionInput) (*PhysicalFulfillmentResult, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("补发原因不能为空")
	}
	return s.mutatePhysicalFulfillment(ctx, currentScope, input.FulfillmentID, 0, input.Reason, func(_ *gorm.DB, row *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error) {
		if row.FulfillmentStatus != PhysicalFulfillmentStatusException {
			return "", "", 0, "", nil, errors.New("只有异常状态可以发起补发")
		}
		fromStatus := row.FulfillmentStatus
		row.FulfillmentStatus = PhysicalFulfillmentStatusReissuePending
		row.ProductionStatus = PhysicalProductionStatusQueued
		row.ShippingStatus = PhysicalShippingStatusPending
		row.FailureCode = firstNonEmpty(input.FailureCode, row.FailureCode)
		row.FailureReason = strings.TrimSpace(input.Reason)
		row.RequestID = firstNonEmpty(input.RequestID, row.RequestID)
		row.TraceID = firstNonEmpty(input.TraceID, row.TraceID)
		return fromStatus, PhysicalActionReissueRequested, input.OperatorID, input.Reason, map[string]interface{}{
			"failureCode": row.FailureCode,
		}, nil
	})
}

func (s *Service) QueryMemberPhysicalFulfillmentDetail(ctx context.Context, currentScope pkgscope.GovernanceScope, memberID int64, assetInstanceID int64) (*MemberPhysicalFulfillmentDetail, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if memberID <= 0 || assetInstanceID <= 0 {
		return nil, errors.New("会员和资产实例ID不能为空")
	}

	var row physicalFulfillmentDetailRow
	base := s.DB.WithContext(ctx).
		Table("sms_card_physical_fulfillment AS fulfillment").
		Joins("LEFT JOIN sms_draw_activity AS activity ON activity.id = fulfillment.activity_id AND activity.is_deleted = 0").
		Joins("LEFT JOIN sms_card_template AS template ON template.id = fulfillment.template_id AND template.is_deleted = 0").
		Joins("LEFT JOIN sms_card_instance AS instance ON instance.id = fulfillment.asset_instance_id AND instance.is_deleted = 0").
		Where("fulfillment.asset_instance_id = ? AND fulfillment.member_id = ? AND fulfillment.is_deleted = 0", assetInstanceID, memberID)
	base = pkgscope.ApplyGovernanceScope(base, currentScope, "fulfillment")
	if err := base.Select(`
			fulfillment.*,
			COALESCE(activity.name, '') AS activity_name,
			COALESCE(template.template_name, '') AS template_name,
			COALESCE(instance.issued_at, instance.create_time) AS obtained_at,
			COALESCE(instance.mint_status, '') AS mint_status`).
		Take(&row).Error; err != nil {
		return nil, err
	}

	logs, err := s.queryPhysicalFulfillmentLogs(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return buildMemberPhysicalFulfillmentDetail(row, logs), nil
}

func (s *Service) QueryPhysicalFulfillmentList(ctx context.Context, currentScope pkgscope.GovernanceScope, filter PhysicalFulfillmentFilter) (int64, []PhysicalFulfillmentItem, error) {
	if s.DB == nil {
		return 0, nil, errors.New("数据库未初始化")
	}
	if filter.PageNum <= 0 {
		filter.PageNum = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	base := s.physicalFulfillmentBaseQuery(ctx, currentScope)
	base = applyPhysicalFulfillmentFilter(base, filter)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var rows []physicalFulfillmentDetailRow
	if err := base.Select(physicalFulfillmentSelectColumns()).
		Order("fulfillment.id DESC").
		Offset(int((filter.PageNum - 1) * filter.PageSize)).
		Limit(int(filter.PageSize)).
		Find(&rows).Error; err != nil {
		return 0, nil, err
	}

	items := make([]PhysicalFulfillmentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildPhysicalFulfillmentItem(row))
	}
	return total, items, nil
}

func (s *Service) QueryPhysicalFulfillmentAdminDetail(ctx context.Context, currentScope pkgscope.GovernanceScope, fulfillmentID int64) (*PhysicalFulfillmentAdminDetail, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if fulfillmentID <= 0 {
		return nil, errors.New("履约单ID不能为空")
	}

	var row physicalFulfillmentDetailRow
	base := s.physicalFulfillmentBaseQuery(ctx, currentScope).Where("fulfillment.id = ?", fulfillmentID)
	if err := base.Select(physicalFulfillmentSelectColumns()).Take(&row).Error; err != nil {
		return nil, err
	}
	logs, err := s.queryPhysicalFulfillmentLogs(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	addressSummary := strings.TrimSpace(strings.Join([]string{row.ReceiverProvince, row.ReceiverCity, row.ReceiverDistrict, row.ReceiverAddress}, ""))
	return &PhysicalFulfillmentAdminDetail{
		Item:           buildPhysicalFulfillmentItem(row),
		ReceiverName:   row.ReceiverName,
		ReceiverPhone:  row.ReceiverPhone,
		AddressSummary: addressSummary,
		RequestID:      row.RequestID,
		TraceID:        row.TraceID,
		Timeline:       buildPhysicalTimeline(logs),
	}, nil
}

func (s *Service) physicalFulfillmentBaseQuery(ctx context.Context, currentScope pkgscope.GovernanceScope) *gorm.DB {
	base := s.DB.WithContext(ctx).
		Table("sms_card_physical_fulfillment AS fulfillment").
		Joins("LEFT JOIN sms_draw_activity AS activity ON activity.id = fulfillment.activity_id AND activity.is_deleted = 0").
		Joins("LEFT JOIN sms_card_template AS template ON template.id = fulfillment.template_id AND template.is_deleted = 0").
		Joins("LEFT JOIN sms_card_instance AS instance ON instance.id = fulfillment.asset_instance_id AND instance.is_deleted = 0").
		Where("fulfillment.is_deleted = 0")
	return pkgscope.ApplyGovernanceScope(base, currentScope, "fulfillment")
}

func (s *Service) mutatePhysicalFulfillment(
	ctx context.Context,
	currentScope pkgscope.GovernanceScope,
	fulfillmentID int64,
	assetInstanceID int64,
	defaultReason string,
	mutator func(*gorm.DB, *PhysicalFulfillmentRow) (string, string, int64, string, map[string]interface{}, error),
) (*PhysicalFulfillmentResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	var result *PhysicalFulfillmentResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		row, err := s.loadPhysicalFulfillment(ctx, tx, fulfillmentID, assetInstanceID, true)
		if err != nil {
			return err
		}
		if err = validateScope(row.PlatformID, row.TenantID, row.MerchantID, currentScope); err != nil {
			return err
		}

		fromStatus, action, operatorID, reason, payload, err := mutator(tx, row)
		if err != nil {
			return err
		}
		now := s.now()
		row.UpdateBy = operatorID
		row.OperatorID = operatorID
		row.UpdateTime = &now
		if err = tx.WithContext(ctx).
			Table(row.TableName()).
			Where("id = ? AND is_deleted = 0", row.ID).
			Updates(map[string]interface{}{
				"fulfillment_status":   row.FulfillmentStatus,
				"production_status":    row.ProductionStatus,
				"quality_status":       row.QualityStatus,
				"shipping_status":      row.ShippingStatus,
				"address_id":           row.AddressID,
				"receiver_name":        row.ReceiverName,
				"receiver_phone":       row.ReceiverPhone,
				"receiver_province":    row.ReceiverProvince,
				"receiver_city":        row.ReceiverCity,
				"receiver_district":    row.ReceiverDistrict,
				"receiver_address":     row.ReceiverAddress,
				"receiver_postal_code": row.ReceiverPostalCode,
				"production_batch_no":  row.ProductionBatchNo,
				"carrier_code":         row.CarrierCode,
				"carrier_name":         row.CarrierName,
				"tracking_no":          row.TrackingNo,
				"shipped_at":           row.ShippedAt,
				"signed_at":            row.SignedAt,
				"auto_sign_at":         row.AutoSignAt,
				"failure_code":         row.FailureCode,
				"failure_reason":       row.FailureReason,
				"operator_id":          row.OperatorID,
				"trace_id":             row.TraceID,
				"request_id":           row.RequestID,
				"update_by":            row.UpdateBy,
				"update_time":          now,
			}).Error; err != nil {
			return err
		}
		if action != "" {
			if err = s.appendPhysicalFulfillmentLogTx(ctx, tx, row, fromStatus, row.FulfillmentStatus, action, OperatorManual, operatorID, firstNonEmpty(reason, defaultReason), payload, row.TraceID, row.RequestID); err != nil {
				return err
			}
		}
		result = buildPhysicalFulfillmentResult(row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) physicalFulfillmentBlockReason(ctx context.Context, tx *gorm.DB, instance *CardInstanceRow) (string, string, error) {
	mintStatus := strings.TrimSpace(instance.MintStatus)
	chainStatus := strings.TrimSpace(instance.ChainStatus)
	if mintStatus != MintStatusSuccess || chainStatus != ChainStatusSuccess || strings.TrimSpace(instance.TokenID) == "" {
		switch mintStatus {
		case MintStatusManualReview, MintStatusFrozen:
			return PhysicalBlockManualReview, physicalBlockReasonText(PhysicalBlockManualReview), nil
		case MintStatusFailed:
			return PhysicalBlockChannelUnavailable, physicalBlockReasonText(PhysicalBlockChannelUnavailable), nil
		default:
			return PhysicalBlockDigitalPending, physicalBlockReasonText(PhysicalBlockDigitalPending), nil
		}
	}

	// 购买型资产 activity_id=0，无抽奖活动，跳过实名校验
	var realNameRequired int32
	if instance.ActivityID > 0 {
		activity, err := s.loadDrawActivity(ctx, tx, instance.ActivityID)
		if err != nil {
			return "", "", err
		}
		realNameRequired = activity.RealNameRequired
	}
	if realNameRequired == mintEnabledStatus {
		realNameStatus, realNameErr := s.loadMemberRealNameStatus(ctx, tx, instance.MemberID)
		if realNameErr != nil {
			return "", "", errors.New("实名状态查询失败，暂不可履约")
		}
		if realNameStatus != mintVerifiedRealNameCode {
			return PhysicalBlockRealName, physicalBlockReasonText(PhysicalBlockRealName), nil
		}
	}

	if s.hasColumn(tx, CardInstanceRow{}.TableName(), "compliance_status") {
		var complianceStatus string
		if complianceErr := tx.WithContext(ctx).
			Table(CardInstanceRow{}.TableName()).
			Select("COALESCE(compliance_status, '')").
			Where("id = ? AND is_deleted = 0", instance.ID).
			Scan(&complianceStatus).Error; complianceErr != nil {
			return "", "", complianceErr
		}
		switch strings.TrimSpace(complianceStatus) {
		case ComplianceStatusRestricted, ComplianceStatusRecycleRequested, ComplianceStatusRecycled:
			return PhysicalBlockCompliance, physicalBlockReasonText(PhysicalBlockCompliance), nil
		}
	}
	return "", "", nil
}

func physicalFulfillmentStatusForBlockReason(reason string) string {
	switch strings.TrimSpace(reason) {
	case PhysicalBlockRealName:
		return PhysicalFulfillmentStatusPendingRealName
	default:
		return PhysicalFulfillmentStatusPendingDigitalConfirmation
	}
}

func (s *Service) loadPhysicalFulfillmentByAsset(ctx context.Context, tx *gorm.DB, assetInstanceID int64, forUpdate bool) (*PhysicalFulfillmentRow, error) {
	var row PhysicalFulfillmentRow
	query := tx.WithContext(ctx).Table(row.TableName()).Where("asset_instance_id = ? AND is_deleted = 0", assetInstanceID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadPhysicalFulfillment(ctx context.Context, tx *gorm.DB, fulfillmentID int64, assetInstanceID int64, forUpdate bool) (*PhysicalFulfillmentRow, error) {
	var row PhysicalFulfillmentRow
	query := tx.WithContext(ctx).Table(row.TableName()).Where("is_deleted = 0")
	if fulfillmentID > 0 {
		query = query.Where("id = ?", fulfillmentID)
	} else if assetInstanceID > 0 {
		query = query.Where("asset_instance_id = ?", assetInstanceID)
	} else {
		return nil, errors.New("履约单ID或资产实例ID不能为空")
	}
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadMemberAddress(ctx context.Context, tx *gorm.DB, memberID int64, addressID int64) (*memberAddressSnapshotRow, error) {
	var row memberAddressSnapshotRow
	if err := tx.WithContext(ctx).
		Table(row.TableName()).
		Where("id = ? AND member_id = ? AND is_deleted = 0", addressID, memberID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会员收货地址不存在")
		}
		return nil, err
	}
	return &row, nil
}

func (s *Service) appendPhysicalFulfillmentLogTx(ctx context.Context, tx *gorm.DB, row *PhysicalFulfillmentRow, fromStatus string, toStatus string, action string, operatorType string, operatorID int64, reason string, payload interface{}, traceID string, requestID string) error {
	body := ""
	if payload != nil {
		if raw, err := json.Marshal(payload); err == nil {
			body = string(raw)
		}
	}
	logRow := &PhysicalFulfillmentLogRow{
		FulfillmentID:   row.ID,
		AssetInstanceID: row.AssetInstanceID,
		Action:          strings.TrimSpace(action),
		FromStatus:      strings.TrimSpace(fromStatus),
		ToStatus:        strings.TrimSpace(toStatus),
		OperatorType:    normalizeOperatorType(operatorType),
		OperatorID:      operatorID,
		Reason:          strings.TrimSpace(reason),
		PayloadJSON:     body,
		TraceID:         firstNonEmpty(traceID, row.TraceID),
		RequestID:       firstNonEmpty(requestID, row.RequestID),
		CreateTime:      s.now(),
	}
	return tx.WithContext(ctx).Table(logRow.TableName()).Create(logRow).Error
}

func (s *Service) queryPhysicalFulfillmentLogs(ctx context.Context, fulfillmentID int64) ([]PhysicalFulfillmentLogRow, error) {
	var rows []PhysicalFulfillmentLogRow
	if err := s.DB.WithContext(ctx).
		Table(PhysicalFulfillmentLogRow{}.TableName()).
		Where("fulfillment_id = ?", fulfillmentID).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) hasColumn(db *gorm.DB, tableName string, columnName string) bool {
	if s == nil || db == nil {
		return false
	}
	return db.Migrator().HasColumn(tableName, columnName)
}

func applyPhysicalAddressSnapshot(row *PhysicalFulfillmentRow, address *memberAddressSnapshotRow) {
	row.AddressID = address.ID
	row.ReceiverName = strings.TrimSpace(address.ReceiverName)
	row.ReceiverPhone = strings.TrimSpace(address.ReceiverPhone)
	row.ReceiverProvince = strings.TrimSpace(address.Province)
	row.ReceiverCity = strings.TrimSpace(address.City)
	row.ReceiverDistrict = strings.TrimSpace(address.District)
	row.ReceiverAddress = strings.TrimSpace(address.DetailAddress)
	row.ReceiverPostalCode = strings.TrimSpace(address.PostalCode)
}

func buildPhysicalFulfillmentResult(row *PhysicalFulfillmentRow) *PhysicalFulfillmentResult {
	if row == nil {
		return &PhysicalFulfillmentResult{}
	}
	return &PhysicalFulfillmentResult{
		FulfillmentID:         row.ID,
		AssetInstanceID:       row.AssetInstanceID,
		FulfillmentNo:         row.FulfillmentNo,
		FulfillmentStatus:     row.FulfillmentStatus,
		FulfillmentStatusText: physicalFulfillmentStatusText(row.FulfillmentStatus),
		ProductionStatus:      row.ProductionStatus,
		ProductionStatusText:  physicalProductionStatusText(row.ProductionStatus),
		ShippingStatus:        row.ShippingStatus,
		ShippingStatusText:    physicalShippingStatusText(row.ShippingStatus),
		ShippingFeeStatus:     PhysicalShippingFeeStatusPending,
		ShippingFeeStatusText: physicalShippingFeeStatusText(PhysicalShippingFeeStatusPending),
	}
}

func buildMemberPhysicalFulfillmentDetail(row physicalFulfillmentDetailRow, logs []PhysicalFulfillmentLogRow) *MemberPhysicalFulfillmentDetail {
	addressSummary := strings.TrimSpace(strings.Join([]string{
		row.ReceiverProvince,
		row.ReceiverCity,
		row.ReceiverDistrict,
		row.ReceiverAddress,
	}, ""))
	feeStatus, feeStatusText, feeAmount := resolveShippingFeeSnapshot(logs)
	return &MemberPhysicalFulfillmentDetail{
		FulfillmentID:   row.ID,
		FulfillmentNo:   row.FulfillmentNo,
		AssetInstanceID: row.AssetInstanceID,
		AssetNo:         row.AssetNo,
		TemplateName:    row.TemplateName,
		ActivityName:    row.ActivityName,
		ObtainedAt:      formatNullableTime(row.ObtainedAt),
		// Story 10.11 / Task 8.2 / AC4：C 端实物履约详情同样使用消费者口径文案。
		MintStatusText:        MintStatusConsumerText(row.MintStatus),
		FulfillmentStatus:     row.FulfillmentStatus,
		FulfillmentStatusText: physicalFulfillmentStatusText(row.FulfillmentStatus),
		ProductionStatusText:  physicalProductionStatusText(row.ProductionStatus),
		ShippingStatusText:    physicalShippingStatusText(row.ShippingStatus),
		ShippingFeeStatus:     feeStatus,
		ShippingFeeStatusText: feeStatusText,
		ShippingFeeAmount:     feeAmount,
		ReceiverNameMasked:    maskReceiverName(row.ReceiverName),
		ReceiverPhoneMasked:   maskReceiverPhone(row.ReceiverPhone),
		AddressSummary:        addressSummary,
		CarrierName:           row.CarrierName,
		TrackingNo:            row.TrackingNo,
		ComplianceTipSummary:  "实体卡履约仅展示制作、配送和签收进度。",
		Timeline:              buildPhysicalTimeline(logs),
	}
}

func applyPhysicalFulfillmentFilter(db *gorm.DB, filter PhysicalFulfillmentFilter) *gorm.DB {
	query := db
	if filter.ActivityID > 0 {
		query = query.Where("fulfillment.activity_id = ?", filter.ActivityID)
	}
	if filter.TemplateID > 0 {
		query = query.Where("fulfillment.template_id = ?", filter.TemplateID)
	}
	if filter.MemberID > 0 {
		query = query.Where("fulfillment.member_id = ?", filter.MemberID)
	}
	if strings.TrimSpace(filter.AssetNo) != "" {
		query = query.Where("fulfillment.asset_no LIKE ?", "%"+strings.TrimSpace(filter.AssetNo)+"%")
	}
	if strings.TrimSpace(filter.FulfillmentNo) != "" {
		query = query.Where("fulfillment.fulfillment_no LIKE ?", "%"+strings.TrimSpace(filter.FulfillmentNo)+"%")
	}
	if strings.TrimSpace(filter.ProductionBatchNo) != "" {
		query = query.Where("fulfillment.production_batch_no LIKE ?", "%"+strings.TrimSpace(filter.ProductionBatchNo)+"%")
	}
	if strings.TrimSpace(filter.FulfillmentStatus) != "" {
		query = query.Where("fulfillment.fulfillment_status = ?", strings.TrimSpace(filter.FulfillmentStatus))
	}
	if strings.TrimSpace(filter.ProductionStatus) != "" {
		query = query.Where("fulfillment.production_status = ?", strings.TrimSpace(filter.ProductionStatus))
	}
	if strings.TrimSpace(filter.ShippingStatus) != "" {
		query = query.Where("fulfillment.shipping_status = ?", strings.TrimSpace(filter.ShippingStatus))
	}
	if strings.TrimSpace(filter.TrackingNo) != "" {
		query = query.Where("fulfillment.tracking_no LIKE ?", "%"+strings.TrimSpace(filter.TrackingNo)+"%")
	}
	if strings.TrimSpace(filter.FailureCode) != "" {
		query = query.Where("fulfillment.failure_code = ?", strings.TrimSpace(filter.FailureCode))
	}
	if start, ok := parseStartTime(filter.StartTime); ok {
		query = query.Where("fulfillment.create_time >= ?", start)
	}
	if end, ok := parseEndTime(filter.EndTime); ok {
		query = query.Where("fulfillment.create_time <= ?", end)
	}
	return query
}

func physicalFulfillmentSelectColumns() string {
	return `
		fulfillment.*,
		COALESCE(activity.name, '') AS activity_name,
		COALESCE(template.template_name, '') AS template_name,
		COALESCE(instance.issued_at, instance.create_time) AS obtained_at,
		COALESCE(instance.mint_status, '') AS mint_status`
}

func buildPhysicalFulfillmentItem(row physicalFulfillmentDetailRow) PhysicalFulfillmentItem {
	return PhysicalFulfillmentItem{
		FulfillmentID:         row.ID,
		FulfillmentNo:         row.FulfillmentNo,
		AssetInstanceID:       row.AssetInstanceID,
		AssetNo:               row.AssetNo,
		ActivityID:            row.ActivityID,
		ActivityName:          row.ActivityName,
		TemplateID:            row.TemplateID,
		TemplateName:          row.TemplateName,
		MemberID:              row.MemberID,
		FulfillmentStatus:     row.FulfillmentStatus,
		FulfillmentStatusText: physicalFulfillmentStatusText(row.FulfillmentStatus),
		ProductionStatus:      row.ProductionStatus,
		ProductionStatusText:  physicalProductionStatusText(row.ProductionStatus),
		ShippingStatus:        row.ShippingStatus,
		ShippingStatusText:    physicalShippingStatusText(row.ShippingStatus),
		ProductionBatchNo:     row.ProductionBatchNo,
		CarrierName:           row.CarrierName,
		TrackingNo:            row.TrackingNo,
		FailureCode:           row.FailureCode,
		FailureReason:         row.FailureReason,
		ShippedAt:             formatPhysicalTimePtr(row.ShippedAt),
		SignedAt:              formatPhysicalTimePtr(row.SignedAt),
		UpdateTime:            formatPhysicalTimePtr(row.UpdateTime),
	}
}

func buildPhysicalTimeline(logs []PhysicalFulfillmentLogRow) []PhysicalFulfillmentTimelineItem {
	timeline := make([]PhysicalFulfillmentTimelineItem, 0, len(logs))
	for _, item := range logs {
		timeline = append(timeline, PhysicalFulfillmentTimelineItem{
			Action:     item.Action,
			ActionText: physicalActionText(item.Action),
			StatusText: physicalFulfillmentStatusText(item.ToStatus),
			Reason:     item.Reason,
			CreateTime: item.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return timeline
}

func formatPhysicalTimePtr(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func buildPhysicalFulfillmentNo(assetInstanceID int64, now time.Time) string {
	return fmt.Sprintf("PF%s%08d", now.Format("20060102"), assetInstanceID)
}

func isValidPhysicalProductionStatus(status string) bool {
	switch status {
	case PhysicalProductionStatusNotStarted,
		PhysicalProductionStatusQueued,
		PhysicalProductionStatusPrinting,
		PhysicalProductionStatusQualityChecking,
		PhysicalProductionStatusCompleted,
		PhysicalProductionStatusFailed:
		return true
	default:
		return false
	}
}

func physicalBlockReasonText(reason string) string {
	switch strings.TrimSpace(reason) {
	case PhysicalBlockRealName:
		return "待完成实名认证后发放"
	case PhysicalBlockDigitalPending:
		return "待完成权益确认后制作"
	case PhysicalBlockManualReview:
		return "当前需人工处理，暂不可制作"
	case PhysicalBlockCompliance:
		return "合规状态暂不允许交付"
	case PhysicalBlockChannelUnavailable:
		return "权益确认服务暂不可用"
	default:
		return ""
	}
}

func physicalFulfillmentStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case PhysicalFulfillmentStatusPendingDigitalConfirmation:
		return "待权益确认"
	case PhysicalFulfillmentStatusPendingRealName:
		return "待实名后发放"
	case PhysicalFulfillmentStatusPendingAddress:
		return "待确认地址"
	case PhysicalFulfillmentStatusAddressConfirmed:
		return "地址已确认"
	case PhysicalFulfillmentStatusProductionPending:
		return "待制作"
	case PhysicalFulfillmentStatusProductionInProgress:
		return "制作中"
	case PhysicalFulfillmentStatusQualityChecking:
		return "质检中"
	case PhysicalFulfillmentStatusReadyToShip:
		return "待发货"
	case PhysicalFulfillmentStatusShipped:
		return "已发货"
	case PhysicalFulfillmentStatusInTransit:
		return "运输中"
	case PhysicalFulfillmentStatusSigned:
		return "已签收"
	case PhysicalFulfillmentStatusException:
		return "履约异常"
	case PhysicalFulfillmentStatusReissuePending:
		return "补发中"
	case PhysicalFulfillmentStatusCancelled:
		return "已取消"
	default:
		return "状态待同步"
	}
}

func physicalProductionStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case PhysicalProductionStatusNotStarted:
		return "未开始"
	case PhysicalProductionStatusQueued:
		return "待制作"
	case PhysicalProductionStatusPrinting:
		return "制作中"
	case PhysicalProductionStatusQualityChecking:
		return "质检中"
	case PhysicalProductionStatusCompleted:
		return "制作完成"
	case PhysicalProductionStatusFailed:
		return "制作失败"
	default:
		return "制作状态待同步"
	}
}

func physicalShippingStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case PhysicalShippingStatusPending:
		return "待发货"
	case PhysicalShippingStatusShipped:
		return "已发货"
	case PhysicalShippingStatusInTransit:
		return "运输中"
	case PhysicalShippingStatusDelivered:
		return "已送达"
	case PhysicalShippingStatusSigned:
		return "已签收"
	case PhysicalShippingStatusException:
		return "物流异常"
	default:
		return "物流状态待同步"
	}
}

func physicalShippingFeeStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case PhysicalShippingFeeStatusPaid:
		return "邮费已支付"
	default:
		return "待支付邮费"
	}
}

func physicalActionText(action string) string {
	switch strings.TrimSpace(action) {
	case PhysicalActionCreated:
		return "创建履约单"
	case PhysicalActionAddressConfirmed:
		return "确认收货地址"
	case PhysicalActionProductionUpdated:
		return "更新制作状态"
	case PhysicalActionShipped:
		return "实体卡发货"
	case PhysicalActionSigned:
		return "确认签收"
	case PhysicalActionException:
		return "标记异常"
	case PhysicalActionReissueRequested:
		return "发起补发"
	case PhysicalActionShippingFeePaid:
		return "支付邮费"
	case OperationAssetTransferred:
		return "卡片转赠"
	default:
		return strings.TrimSpace(action)
	}
}

func resolveShippingFeeSnapshot(logs []PhysicalFulfillmentLogRow) (string, string, int64) {
	status := PhysicalShippingFeeStatusPending
	text := physicalShippingFeeStatusText(status)
	var amount int64
	for _, item := range logs {
		if strings.TrimSpace(item.Action) != PhysicalActionShippingFeePaid {
			continue
		}
		status = PhysicalShippingFeeStatusPaid
		text = physicalShippingFeeStatusText(status)
		var payload struct {
			PayAmount int64 `json:"payAmount"`
		}
		if strings.TrimSpace(item.PayloadJSON) != "" && json.Unmarshal([]byte(item.PayloadJSON), &payload) == nil {
			amount = payload.PayAmount
		}
	}
	return status, text, amount
}

func maskReceiverName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return ""
	}
	return string(runes[0]) + "*"
}

func maskReceiverPhone(phone string) string {
	value := strings.TrimSpace(phone)
	if len(value) < 7 {
		return value
	}
	return value[:3] + "****" + value[len(value)-4:]
}
