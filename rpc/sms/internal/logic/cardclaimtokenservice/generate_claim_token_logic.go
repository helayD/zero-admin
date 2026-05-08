package cardclaimtokenservicelogic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	claimTokenStatusActive  = "active"
	claimTokenStatusClaimed = "claimed"
	claimTokenStatusExpired = "expired"
	claimTokenStatusRevoked = "revoked"

	cardAssetStatusClaimed           = "claimed"
	cardAssetStatusPendingRedemption = "pending_redemption"

	cardAssetOperationHolderTransferred = "holder_transferred"
	cardAssetReasonClaimTokenConsumed   = "claim_token_consumed"
)

type claimTokenRow struct {
	ID             int64      `gorm:"column:id"`
	Token          string     `gorm:"column:token"`
	CardInstanceID int64      `gorm:"column:card_instance_id"`
	IssuerID       int64      `gorm:"column:issuer_id"`
	IssuerType     string     `gorm:"column:issuer_type"`
	ExpireAt       *time.Time `gorm:"column:expire_at"`
	MaxClaims      int32      `gorm:"column:max_claims"`
	ClaimedCount   int32      `gorm:"column:claimed_count"`
	Version        int32      `gorm:"column:version"`
	Status         string     `gorm:"column:status"`
	ClaimedBy      int64      `gorm:"column:claimed_by"`
	ClaimedAt      *time.Time `gorm:"column:claimed_at"`
	PlatformID     int64      `gorm:"column:platform_id"`
	TenantID       int64      `gorm:"column:tenant_id"`
	MerchantID     int64      `gorm:"column:merchant_id"`
	CreatedAt      *time.Time `gorm:"column:created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
	IsDeleted      int32      `gorm:"column:is_deleted"`
}

func (claimTokenRow) TableName() string {
	return "sms_card_claim_token"
}

type cardInstanceRow struct {
	ID                  int64  `gorm:"column:id"`
	PlatformID          int64  `gorm:"column:platform_id"`
	TenantID            int64  `gorm:"column:tenant_id"`
	MerchantID          int64  `gorm:"column:merchant_id"`
	MemberID            int64  `gorm:"column:member_id"`
	AssetStatus         string `gorm:"column:asset_status"`
	MintStatus          string `gorm:"column:mint_status"`
	AssetNo             string `gorm:"column:asset_no"`
	TemplateID          int64  `gorm:"column:template_id"`
	Transferable        int32  `gorm:"column:transferable"`
	TransferLimit       int32  `gorm:"column:transfer_limit"`
	ClaimCondition      string `gorm:"column:claim_condition"`
	RedemptionCondition string `gorm:"column:redemption_condition"`
	IsDeleted           int32  `gorm:"column:is_deleted"`
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

type redemptionOrderRow struct {
	ID             int64  `gorm:"column:id"`
	CardInstanceID int64  `gorm:"column:card_instance_id"`
	Status         string `gorm:"column:status"`
	IsDeleted      int32  `gorm:"column:is_deleted"`
}

func (redemptionOrderRow) TableName() string {
	return "sms_card_redemption_order"
}

type GenerateClaimTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateClaimTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateClaimTokenLogic {
	return &GenerateClaimTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateClaimTokenLogic) GenerateClaimToken(in *smsclient.GenerateClaimTokenReq) (*smsclient.GenerateClaimTokenResp, error) {
	if in.CardInstanceId <= 0 {
		return nil, errors.New("卡片实例ID无效")
	}
	if in.IssuerId <= 0 {
		return nil, errors.New("分享人ID无效")
	}

	var token *claimTokenRow
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		token, txErr = l.generateClaimTokenInTx(tx, in)
		return txErr
	})
	if err != nil {
		logc.Errorf(l.ctx, "生成分享凭证失败: %v, cardInstanceId=%d", err, in.CardInstanceId)
		return nil, err
	}

	return &smsclient.GenerateClaimTokenResp{
		Token: buildClaimTokenData(token),
	}, nil
}

func (l *GenerateClaimTokenLogic) generateClaimTokenInTx(tx *gorm.DB, in *smsclient.GenerateClaimTokenReq) (*claimTokenRow, error) {
	var instance cardInstanceRow
	if err := tx.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", in.CardInstanceId).
		Take(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片实例不存在")
		}
		return nil, err
	}

	if instance.MemberID != in.IssuerId {
		return nil, errors.New("只有卡片持有人才能分享")
	}

	if instance.Transferable != 1 {
		return nil, errors.New("该卡片不允许转赠")
	}

	if instance.AssetStatus != cardAssetStatusClaimed {
		return nil, fmt.Errorf("卡片状态不允许分享,当前状态:%s", instance.AssetStatus)
	}

	var activeOrderCount int64
	if err := tx.WithContext(l.ctx).
		Table(redemptionOrderRow{}.TableName()).
		Where("card_instance_id = ? AND status IN (?, ?) AND is_deleted = 0", in.CardInstanceId, "pending", "processing").
		Count(&activeOrderCount).Error; err != nil {
		return nil, err
	}
	if activeOrderCount > 0 {
		return nil, errors.New("该卡片存在进行中的提货单,不能分享")
	}

	if instance.TransferLimit > 0 {
		var transferCount int64
		if err := tx.WithContext(l.ctx).
			Table(cardAssetLogRow{}.TableName()).
			Where("asset_instance_id = ? AND operation_type = ?", in.CardInstanceId, cardAssetOperationHolderTransferred).
			Count(&transferCount).Error; err != nil {
			return nil, err
		}
		if int32(transferCount) >= instance.TransferLimit {
			return nil, fmt.Errorf("该卡片转赠次数已达上限:%d", instance.TransferLimit)
		}
	}

	expireHours := in.ExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}
	maxClaims := in.MaxClaims
	if maxClaims <= 0 {
		maxClaims = 1
	}

	expireAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
	token := generateClaimToken()
	now := time.Now()
	row := &claimTokenRow{
		Token:          token,
		CardInstanceID: in.CardInstanceId,
		IssuerID:       in.IssuerId,
		IssuerType:     "member",
		ExpireAt:       &expireAt,
		MaxClaims:      maxClaims,
		ClaimedCount:   0,
		Status:         claimTokenStatusActive,
		PlatformID:     in.PlatformId,
		TenantID:       in.TenantId,
		MerchantID:     in.MerchantId,
		CreatedAt:      &now,
	}
	if err := tx.WithContext(l.ctx).
		Table(row.TableName()).
		Create(row).Error; err != nil {
		return nil, fmt.Errorf("创建分享凭证失败: %w", err)
	}

	return row, nil
}

func generateClaimToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("token_%d", time.Now().UnixNano())))
	}
	return base64.URLEncoding.EncodeToString(buf)
}

func buildClaimTokenData(row *claimTokenRow) *smsclient.ClaimTokenData {
	if row == nil {
		return nil
	}
	data := &smsclient.ClaimTokenData{
		Id:             row.ID,
		Token:          row.Token,
		CardInstanceId: row.CardInstanceID,
		IssuerId:       row.IssuerID,
		IssuerType:     row.IssuerType,
		MaxClaims:      row.MaxClaims,
		ClaimedCount:   row.ClaimedCount,
		Status:         row.Status,
		ClaimedBy:      row.ClaimedBy,
		PlatformId:     row.PlatformID,
		TenantId:       row.TenantID,
		MerchantId:     row.MerchantID,
	}
	if row.ExpireAt != nil {
		data.ExpireAt = row.ExpireAt.Format("2006-01-02 15:04:05")
	}
	if row.ClaimedAt != nil {
		data.ClaimedAt = row.ClaimedAt.Format("2006-01-02 15:04:05")
	}
	if row.CreatedAt != nil {
		data.CreateTime = row.CreatedAt.Format("2006-01-02 15:04:05")
	}
	if row.UpdatedAt != nil {
		data.UpdateTime = row.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return data
}
