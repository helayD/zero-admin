package digitalcardmint

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

// Story 10.7 Task 8.x — claim token in-process service.
//
// 历史背景：rpc/sms/smsclient/CardClaimTokenService 由于 GenerateClaimTokenReq
// 等 struct 未实现 proto.Message 接口，跨进程 gRPC marshal 会直接报错
// "GenerateClaimTokenReq, want proto.Message"。这套从 Story 10.7 起就从未在
// 真实环境跑通。本文件把核心逻辑提到 pkg 层，让 front-api 直接 in-process
// 调用，绕过坏的 RPC 边界，恢复「分享给朋友」功能。
//
// 后续如需恢复 RPC，需仿 product_fulfillment_rule.proto 写一份 .proto 并生成
// .pb.go，本文件可作为参考实现保留。

const (
	claimTokenStatusActive  = "active"
	claimTokenStatusClaimed = "claimed"
	claimTokenStatusExpired = "expired"
	claimTokenStatusRevoked = "revoked"

	cardAssetStatusClaimed = "asset_created"

	cardAssetOperationHolderTransferred = "holder_transferred"
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

type claimTokenInstanceRow struct {
	ID             int64  `gorm:"column:id"`
	PlatformID     int64  `gorm:"column:platform_id"`
	TenantID       int64  `gorm:"column:tenant_id"`
	MerchantID     int64  `gorm:"column:merchant_id"`
	MemberID       int64  `gorm:"column:member_id"`
	AssetStatus    string `gorm:"column:asset_status"`
	MintStatus     string `gorm:"column:mint_status"`
	AssetNo        string `gorm:"column:asset_no"`
	TemplateID     int64  `gorm:"column:template_id"`
	Transferable   int32  `gorm:"column:transferable"`
	TransferLimit  int32  `gorm:"column:transfer_limit"`
	IsDeleted      int32  `gorm:"column:is_deleted"`
}

func (claimTokenInstanceRow) TableName() string {
	return "sms_card_instance"
}

type claimTokenAssetLogRow struct {
	AssetInstanceID int64 `gorm:"column:asset_instance_id"`
	OperationType   string `gorm:"column:operation_type"`
}

func (claimTokenAssetLogRow) TableName() string {
	return "sms_card_asset_log"
}

type claimTokenRedemptionOrderRow struct {
	ID             int64  `gorm:"column:id"`
	CardInstanceID int64  `gorm:"column:card_instance_id"`
	Status         string `gorm:"column:status"`
	IsDeleted      int32  `gorm:"column:is_deleted"`
}

func (claimTokenRedemptionOrderRow) TableName() string {
	return "sms_card_redemption_order"
}

// GenerateClaimTokenInput 生成分享凭证入参（in-process 版本）
type GenerateClaimTokenInput struct {
	CardInstanceID int64
	IssuerID       int64
	ExpireHours    int32
	MaxClaims      int32
	TraceID        string
	RequestID      string
}

// ClaimTokenResult 生成分享凭证返回结果
type ClaimTokenResult struct {
	ID             int64
	Token          string
	CardInstanceID int64
	IssuerID       int64
	IssuerType     string
	ExpireAt       string
	MaxClaims      int32
	ClaimedCount   int32
	Status         string
	ClaimedBy      int64
	ClaimedAt      string
	PlatformID     int64
	TenantID       int64
	MerchantID     int64
	CreateTime     string
	UpdateTime     string
}

// GenerateClaimToken 生成分享凭证 (in-process)。
//
// 监管约束：
//   - 卡片必须属于 issuer 本人 (instance.member_id == issuer_id)
//   - 卡片 transferable=1 才允许分享
//   - 卡片 asset_status = asset_created（已到账，未被锁占用）
//   - 不能存在进行中的提货单（pending/processing）
//   - 不能超过 transfer_limit 上限（基于 sms_card_asset_log 历史转赠次数）
//   - token 必须使用 crypto/rand，crypto/rand 不可用时直接 fail-closed
func (s *Service) GenerateClaimToken(ctx context.Context, currentScope pkgscope.GovernanceScope, input GenerateClaimTokenInput) (*ClaimTokenResult, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if input.CardInstanceID <= 0 {
		return nil, errors.New("卡片实例ID无效")
	}
	if input.IssuerID <= 0 {
		return nil, errors.New("分享人ID无效")
	}

	var token *claimTokenRow
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		token, txErr = s.generateClaimTokenInTx(ctx, tx, currentScope, input)
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return buildClaimTokenResult(token), nil
}

func (s *Service) generateClaimTokenInTx(ctx context.Context, tx *gorm.DB, currentScope pkgscope.GovernanceScope, input GenerateClaimTokenInput) (*claimTokenRow, error) {
	var instance claimTokenInstanceRow
	if err := tx.WithContext(ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", input.CardInstanceID).
		Take(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片实例不存在")
		}
		return nil, err
	}

	if instance.MemberID != input.IssuerID {
		return nil, errors.New("只有卡片持有人才能分享")
	}
	if instance.Transferable != 1 {
		return nil, errors.New("该卡片不允许转赠")
	}
	if instance.AssetStatus != cardAssetStatusClaimed {
		return nil, fmt.Errorf("卡片状态不允许分享,当前状态:%s", instance.AssetStatus)
	}

	var activeOrderCount int64
	if err := tx.WithContext(ctx).
		Table(claimTokenRedemptionOrderRow{}.TableName()).
		Where("card_instance_id = ? AND status IN (?, ?, ?) AND is_deleted = 0",
			input.CardInstanceID, "pending", "processing", "shipped").
		Count(&activeOrderCount).Error; err != nil {
		return nil, err
	}
	if activeOrderCount > 0 {
		return nil, errors.New("该卡片存在进行中的提货单,不能分享")
	}

	// 已存在 active 凭证则禁止重复签发，避免单卡多个有效 token 并发
	var activeTokenCount int64
	if err := tx.WithContext(ctx).
		Table(claimTokenRow{}.TableName()).
		Where("card_instance_id = ? AND status = ? AND is_deleted = 0",
			input.CardInstanceID, claimTokenStatusActive).
		Count(&activeTokenCount).Error; err != nil {
		return nil, err
	}
	if activeTokenCount > 0 {
		return nil, errors.New("该卡片已存在活跃的分享凭证,请先吊销后再分享")
	}

	if instance.TransferLimit > 0 {
		var transferCount int64
		if err := tx.WithContext(ctx).
			Table(claimTokenAssetLogRow{}.TableName()).
			Where("asset_instance_id = ? AND operation_type = ?",
				input.CardInstanceID, cardAssetOperationHolderTransferred).
			Count(&transferCount).Error; err != nil {
			return nil, err
		}
		if int32(transferCount) >= instance.TransferLimit {
			return nil, fmt.Errorf("该卡片转赠次数已达上限:%d", instance.TransferLimit)
		}
	}

	expireHours := input.ExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}
	maxClaims := input.MaxClaims
	if maxClaims <= 0 {
		maxClaims = 1
	}

	expireAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
	tokenStr, err := generateRandomClaimToken()
	if err != nil {
		return nil, fmt.Errorf("生成分享凭证失败: %w", err)
	}
	now := time.Now()
	row := &claimTokenRow{
		Token:          tokenStr,
		CardInstanceID: input.CardInstanceID,
		IssuerID:       input.IssuerID,
		IssuerType:     "member",
		ExpireAt:       &expireAt,
		MaxClaims:      maxClaims,
		ClaimedCount:   0,
		Status:         claimTokenStatusActive,
		PlatformID:     currentScope.PlatformID,
		TenantID:       currentScope.TenantID,
		MerchantID:     currentScope.MerchantID,
		CreatedAt:      &now,
	}
	if err := tx.WithContext(ctx).
		Table(row.TableName()).
		Create(row).Error; err != nil {
		return nil, fmt.Errorf("创建分享凭证失败: %w", err)
	}
	return row, nil
}

// generateRandomClaimToken 生成 32 字节加密安全随机 token，使用 RawURLEncoding（无 padding）。
// crypto/rand 失败时返回 error，由调用方触发事务回滚——禁止任何可枚举/可预测的兜底，
// 避免违反 Story 10.7 架构约束 #1（凭证必须不可枚举）。
func generateRandomClaimToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("crypto/rand 不可用: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func buildClaimTokenResult(row *claimTokenRow) *ClaimTokenResult {
	if row == nil {
		return nil
	}
	r := &ClaimTokenResult{
		ID:             row.ID,
		Token:          row.Token,
		CardInstanceID: row.CardInstanceID,
		IssuerID:       row.IssuerID,
		IssuerType:     row.IssuerType,
		MaxClaims:      row.MaxClaims,
		ClaimedCount:   row.ClaimedCount,
		Status:         row.Status,
		ClaimedBy:      row.ClaimedBy,
		PlatformID:     row.PlatformID,
		TenantID:       row.TenantID,
		MerchantID:     row.MerchantID,
	}
	if row.ExpireAt != nil {
		r.ExpireAt = row.ExpireAt.Format("2006-01-02 15:04:05")
	}
	if row.ClaimedAt != nil {
		r.ClaimedAt = row.ClaimedAt.Format("2006-01-02 15:04:05")
	}
	if row.CreatedAt != nil {
		r.CreateTime = row.CreatedAt.Format("2006-01-02 15:04:05")
	}
	if row.UpdatedAt != nil {
		r.UpdateTime = row.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return r
}
