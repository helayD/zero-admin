package audit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	SecurityEventTypeGovernanceDenied = "governance.denied"
	SecurityEventResultDenied         = "denied"
	SecurityEventResultBlocked        = "blocked"
)

type SecurityEvent struct {
	TraceID        string
	EventType      string
	Action         string
	ResourceType   string
	ResourceID     int64
	ScopeType      string
	PlatformID     int64
	TenantID       int64
	MerchantID     int64
	OperatorID     int64
	OperatorName   string
	RequestSummary string
	Result         string
	Payload        string
}

func NewTraceID(action string, resourceID int64) string {
	sanitized := strings.NewReplacer(" ", "_", ".", "_", "/", "_").Replace(strings.TrimSpace(action))
	if sanitized == "" {
		sanitized = "governance"
	}

	return fmt.Sprintf("%s-%d-%d", sanitized, resourceID, time.Now().UnixNano())
}

func RecordSecurityEvent(ctx context.Context, db *gorm.DB, event SecurityEvent) error {
	if db == nil {
		return nil
	}

	if strings.TrimSpace(event.EventType) == "" {
		event.EventType = SecurityEventTypeGovernanceDenied
	}
	if strings.TrimSpace(event.Result) == "" {
		event.Result = SecurityEventResultDenied
	}
	if strings.TrimSpace(event.TraceID) == "" {
		event.TraceID = NewTraceID(event.Action, event.ResourceID)
	}
	if strings.TrimSpace(event.Payload) == "" {
		payload, err := EncodeGovernancePayload(GovernancePayload{
			TraceID:        event.TraceID,
			Action:         event.Action,
			ResourceType:   event.ResourceType,
			ResourceID:     event.ResourceID,
			ScopeType:      event.ScopeType,
			PlatformID:     event.PlatformID,
			TenantID:       event.TenantID,
			MerchantID:     event.MerchantID,
			ScopeLabel:     buildScopeLabel(event.ScopeType, event.PlatformID, event.TenantID, event.MerchantID),
			OperatorID:     event.OperatorID,
			OperatorName:   event.OperatorName,
			Result:         event.Result,
			RequestSummary: event.RequestSummary,
		})
		if err != nil {
			return err
		}
		event.Payload = payload
	}

	return db.WithContext(ctx).
		Table("sys_security_event").
		Create(map[string]interface{}{
			"trace_id":        event.TraceID,
			"event_type":      event.EventType,
			"action":          event.Action,
			"resource_type":   event.ResourceType,
			"resource_id":     event.ResourceID,
			"scope_type":      event.ScopeType,
			"platform_id":     event.PlatformID,
			"tenant_id":       event.TenantID,
			"merchant_id":     event.MerchantID,
			"operator_id":     event.OperatorID,
			"operator_name":   event.OperatorName,
			"request_summary": event.RequestSummary,
			"result":          event.Result,
			"payload":         event.Payload,
			"created_at":      time.Now(),
		}).Error
}

func buildScopeLabel(scopeType string, platformID, tenantID, merchantID int64) string {
	switch strings.TrimSpace(scopeType) {
	case "tenant":
		if tenantID > 0 {
			return fmt.Sprintf("租户 #%d", tenantID)
		}
		return "租户级"
	case "merchant":
		if tenantID > 0 && merchantID > 0 {
			return fmt.Sprintf("租户 #%d / 商户 #%d", tenantID, merchantID)
		}
		if merchantID > 0 {
			return fmt.Sprintf("商户 #%d", merchantID)
		}
		return "商户级"
	default:
		return fmt.Sprintf("平台 #%d", platformID)
	}
}
