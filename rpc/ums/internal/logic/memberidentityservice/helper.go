package memberidentityservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"gorm.io/gorm"
)

type memberIdentityRow struct {
	ID               int64       `gorm:"column:id"`
	MemberID         int64       `gorm:"column:member_id"`
	RealNameStatus   string      `gorm:"column:real_name_status"`
	RealNameMasked   string      `gorm:"column:real_name_masked"`
	IdentityNoMasked string      `gorm:"column:identity_no_masked"`
	ProviderCode     string      `gorm:"column:provider_code"`
	CredentialRef    string      `gorm:"column:credential_ref"`
	VerifiedAt       *time.Time  `gorm:"column:verified_at"`
	FailureReason    string      `gorm:"column:failure_reason"`
	AuditStatus      string      `gorm:"column:audit_status"`
}

func (memberIdentityRow) TableName() string {
	return "ums_member_identity"
}

func LoadMemberIdentityProfile(ctx context.Context, db *gorm.DB, memberID int64) (*umsclient.QueryMemberIdentityProfileResp, error) {
	var row memberIdentityRow
	err := db.WithContext(ctx).
		Table(row.TableName()).
		Where("member_id = ? AND is_deleted = 0", memberID).
		Take(&row).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return defaultMemberIdentityProfile(memberID), nil
	case err != nil:
		return nil, err
	default:
		return buildMemberIdentityProfile(&row), nil
	}
}

func buildMemberIdentityProfile(row *memberIdentityRow) *umsclient.QueryMemberIdentityProfileResp {
	if row == nil {
		return defaultMemberIdentityProfile(0)
	}

	verifiedAt := ""
	if row.VerifiedAt != nil {
		verifiedAt = time_util.TimeToStr(*row.VerifiedAt)
	}

	status := strings.TrimSpace(row.RealNameStatus)
	if status == "" {
		status = "need_real_name"
	}

	auditStatus := strings.TrimSpace(row.AuditStatus)
	if auditStatus == "" {
		auditStatus = "pending"
	}

	return &umsclient.QueryMemberIdentityProfileResp{
		Id:                 row.ID,
		MemberId:           row.MemberID,
		RealNameStatus:     status,
		RealNameStatusText: memberIdentityStatusText(status),
		RealNameMasked:     strings.TrimSpace(row.RealNameMasked),
		IdentityNoMasked:   strings.TrimSpace(row.IdentityNoMasked),
		ProviderCode:       strings.TrimSpace(row.ProviderCode),
		CredentialRef:      strings.TrimSpace(row.CredentialRef),
		VerifiedAt:         verifiedAt,
		FailureReason:      strings.TrimSpace(row.FailureReason),
		AuditStatus:        auditStatus,
	}
}

func defaultMemberIdentityProfile(memberID int64) *umsclient.QueryMemberIdentityProfileResp {
	return &umsclient.QueryMemberIdentityProfileResp{
		MemberId:           memberID,
		RealNameStatus:     "need_real_name",
		RealNameStatusText: memberIdentityStatusText("need_real_name"),
		AuditStatus:        "pending",
	}
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
