package digitalcardmint

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
	TargetMobile   string     `gorm:"column:target_mobile"`
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
	ID            int64  `gorm:"column:id"`
	PlatformID    int64  `gorm:"column:platform_id"`
	TenantID      int64  `gorm:"column:tenant_id"`
	MerchantID    int64  `gorm:"column:merchant_id"`
	MemberID      int64  `gorm:"column:member_id"`
	AssetStatus   string `gorm:"column:asset_status"`
	MintStatus    string `gorm:"column:mint_status"`
	AssetNo       string `gorm:"column:asset_no"`
	TemplateID    int64  `gorm:"column:template_id"`
	Transferable  int32  `gorm:"column:transferable"`
	TransferLimit int32  `gorm:"column:transfer_limit"`
	IsDeleted     int32  `gorm:"column:is_deleted"`
}

func (claimTokenInstanceRow) TableName() string {
	return "sms_card_instance"
}

type claimTokenAssetLogRow struct {
	AssetInstanceID int64  `gorm:"column:asset_instance_id"`
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
	// TargetMobile 指定接收人手机号（11 位）。
	// 监管约束：分享卡片必须指定接收人，只有该手机号能领取。
	// 空字符串视为未指定（保留向后兼容，但当前业务流程要求非空）。
	TargetMobile string
	TraceID      string
	RequestID    string
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
	TargetMobile   string // 完整手机号，仅 issuer / 后台审计可见
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

	// 幂等：若同卡已有 active 且未过期的凭证，直接复用而不是新建。
	// 产品语义：朋友未接收前，分享人重新打开分享页应得到「同一条」分享链接，不应失败。
	// 顺便把过期但状态仍为 active 的凭证修正为 expired，避免库里堆积。
	now := time.Now()
	var existing claimTokenRow
	err := tx.WithContext(ctx).
		Table(existing.TableName()).
		Where("card_instance_id = ? AND status = ? AND is_deleted = 0",
			input.CardInstanceID, claimTokenStatusActive).
		Order("id DESC").
		Take(&existing).Error
	if err == nil {
		if existing.ExpireAt != nil && existing.ExpireAt.Before(now) {
			// 凭证已过期 → 标记 expired，让本次走新建分支
			if upErr := tx.WithContext(ctx).
				Table(existing.TableName()).
				Where("id = ?", existing.ID).
				Updates(map[string]interface{}{
					"status":     claimTokenStatusExpired,
					"updated_at": now,
				}).Error; upErr != nil {
				return nil, fmt.Errorf("过期凭证状态更新失败: %w", upErr)
			}
		} else if existing.ClaimedCount < existing.MaxClaims {
			// 未过期 + 仍可领取
			// 监管约束：如果分享人重新进入分享页 *改了接收人手机号*，旧链接必须作废，
			// 否则会出现「同一卡两个有效手机号」的歧义。
			if input.TargetMobile != "" && existing.TargetMobile != "" && existing.TargetMobile != input.TargetMobile {
				if upErr := tx.WithContext(ctx).
					Table(existing.TableName()).
					Where("id = ?", existing.ID).
					Updates(map[string]interface{}{
						"status":     claimTokenStatusRevoked,
						"updated_at": now,
					}).Error; upErr != nil {
					return nil, fmt.Errorf("作废旧分享凭证失败: %w", upErr)
				}
			} else {
				// 接收人未变 → 直接复用同一条凭证
				return &existing, nil
			}
		}
		// 否则（已达上限）走新建分支
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
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
	tokenStr, tokenErr := generateRandomClaimToken()
	if tokenErr != nil {
		return nil, fmt.Errorf("生成分享凭证失败: %w", tokenErr)
	}
	row := &claimTokenRow{
		Token:          tokenStr,
		CardInstanceID: input.CardInstanceID,
		IssuerID:       input.IssuerID,
		IssuerType:     "member",
		ExpireAt:       &expireAt,
		MaxClaims:      maxClaims,
		ClaimedCount:   0,
		Status:         claimTokenStatusActive,
		TargetMobile:   input.TargetMobile,
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

// ValidateClaimTokenResult 是凭证校验结果（脱敏：不暴露 issuer 信息给陌生人）
type ValidateClaimTokenResult struct {
	Valid         bool
	FailureReason string
	Token         *ClaimTokenResult
	TemplateName  string
	CardFaceImage string
	AssetNo       string
	SenderName    string
	// TargetMobileMasked 接收人手机号（仅返回掩码，如 138****8888），
	// H5 提示「此卡片只能由 138****8888 领取」。原始手机号永不下传给陌生人。
	TargetMobileMasked string
}

// ValidateClaimToken H5 领取页 / Flutter 朋友端预览凭证。
// 只读接口，不消费 token，可重复调用。
//
// 不存在 / 已吊销 / 已过期 / 已被领取完 / 关联卡片状态异常 → Valid=false + FailureReason
func (s *Service) ValidateClaimToken(ctx context.Context, tokenStr string) (*ValidateClaimTokenResult, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if tokenStr == "" {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证不能为空"}, nil
	}

	var row claimTokenRow
	err := s.DB.WithContext(ctx).
		Table(row.TableName()).
		Where("token = ? AND is_deleted = 0", tokenStr).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证不存在"}, nil
		}
		return nil, err
	}

	if row.Status == claimTokenStatusRevoked {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证已被吊销"}, nil
	}
	if row.Status == claimTokenStatusClaimed {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证已被完全领取"}, nil
	}
	if row.Status == claimTokenStatusExpired || (row.ExpireAt != nil && row.ExpireAt.Before(time.Now())) {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证已过期"}, nil
	}
	if row.Status != claimTokenStatusActive {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证状态无效"}, nil
	}
	if row.ClaimedCount >= row.MaxClaims {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "凭证领取次数已达上限"}, nil
	}

	var instance claimTokenInstanceRow
	if err := s.DB.WithContext(ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", row.CardInstanceID).
		Take(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ValidateClaimTokenResult{Valid: false, FailureReason: "关联的卡片实例不存在"}, nil
		}
		return nil, err
	}
	if instance.AssetStatus != cardAssetStatusClaimed {
		return &ValidateClaimTokenResult{Valid: false, FailureReason: "卡片状态不允许领取"}, nil
	}

	// 卡面 + 模板名（H5 预览要显示）
	type templatePreviewRow struct {
		TemplateName  string `gorm:"column:template_name"`
		CardFaceImage string `gorm:"column:card_face_image"`
	}
	var preview templatePreviewRow
	_ = s.DB.WithContext(ctx).
		Table("sms_card_template").
		Select("template_name, card_face_image").
		Where("id = ? AND is_deleted = 0", instance.TemplateID).
		Take(&preview).Error

	data := buildClaimTokenResult(&row)
	if data != nil {
		// 匿名接口不暴露分享人敏感信息 / 完整接收人手机号
		data.IssuerID = 0
		data.IssuerType = ""
		data.TargetMobile = "" // 不下传完整手机号给陌生人
	}
	return &ValidateClaimTokenResult{
		Valid:              true,
		Token:              data,
		TemplateName:       preview.TemplateName,
		CardFaceImage:      preview.CardFaceImage,
		AssetNo:            instance.AssetNo,
		TargetMobileMasked: maskMobile(row.TargetMobile),
	}, nil
}

// maskMobile 将 11 位手机号脱敏为 138****8888 格式。
func maskMobile(m string) string {
	if len(m) != 11 {
		return ""
	}
	return m[:3] + "****" + m[7:]
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
		TargetMobile:   row.TargetMobile,
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

// ============================================================
// 一步式领取（手机号 + 验证码 mock = 123456）
// ============================================================

// ClaimByMobileInput H5 朋友端领取入参
type ClaimByMobileInput struct {
	Token      string // 分享凭证
	Mobile     string // 朋友输入的手机号（必须 == sms_card_claim_token.target_mobile）
	VerifyCode string // 验证码（mock 阶段：必须 == "123456"）
	RequestID  string // 幂等 ID，建议 H5 端用 timestamp + random
	IPAddress  string // 来源 IP，写入审计日志
}

// ClaimByMobileResult H5 朋友端领取结果
type ClaimByMobileResult struct {
	Success       bool
	FailureReason string             // Success=false 时填写
	FailureCode   string             // mismatch / expired / consumed / invalid_code 等机器可读 code
	Card          *ClaimedCardDetail // 仅 Success=true 时填写，给 H5 显示「你拿到了什么」
}

// ClaimedCardDetail 领取成功后给朋友看的最小信息（不暴露任何区块链/底层字段）
type ClaimedCardDetail struct {
	CardInstanceID int64
	AssetNo        string
	TemplateName   string
	CardFaceImage  string
	NewMemberID    int64 // 卡片现归属的 member_id（朋友自己）
}

// ClaimByMobile 一步式领取（H5 朋友端唯一入口）。
//
// 业务规则：
//   - mock 验证码：仅当 verify_code == "123456" 通过；后期接真短信通道时改这里
//   - mobile 必须 == sms_card_claim_token.target_mobile（命中分享指定接收人）
//   - 不命中：返回 Success=false / FailureCode="mismatch"，让 H5 显示「领取失败 + 下载 App」
//   - 命中：自动按手机号查 ums_member_info；不存在则在事务里创建一条新会员
//     → 转移卡片 sms_card_instance.member_id
//     → 标记 token claimed / claimed_by / claimed_at
//     → 写 sms_card_asset_log holder_transferred
//   - 全部 SQL 在同一个事务里完成，任何一步失败都回滚
func (s *Service) ClaimByMobile(ctx context.Context, input ClaimByMobileInput) (*ClaimByMobileResult, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if input.Token == "" {
		return &ClaimByMobileResult{Success: false, FailureReason: "凭证不能为空", FailureCode: "invalid_token"}, nil
	}
	if !isValidChineseMobile(input.Mobile) {
		return &ClaimByMobileResult{Success: false, FailureReason: "请输入有效的手机号", FailureCode: "invalid_mobile"}, nil
	}
	// mock：验证码固定 "123456"，后期换真短信通道时改这里
	if input.VerifyCode != "123456" {
		return &ClaimByMobileResult{Success: false, FailureReason: "验证码错误", FailureCode: "invalid_code"}, nil
	}

	var (
		result *ClaimByMobileResult
	)
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		result, txErr = s.claimByMobileInTx(ctx, tx, input)
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) claimByMobileInTx(ctx context.Context, tx *gorm.DB, input ClaimByMobileInput) (*ClaimByMobileResult, error) {
	// 1. 加锁读 token 行
	var row claimTokenRow
	err := tx.WithContext(ctx).
		Table(row.TableName()).
		Where("token = ? AND is_deleted = 0", input.Token).
		Set("gorm:query_option", "FOR UPDATE").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ClaimByMobileResult{Success: false, FailureReason: "凭证不存在", FailureCode: "not_found"}, nil
		}
		return nil, err
	}

	now := time.Now()
	if row.Status == claimTokenStatusRevoked {
		return &ClaimByMobileResult{Success: false, FailureReason: "凭证已被吊销", FailureCode: "revoked"}, nil
	}
	if row.Status == claimTokenStatusClaimed || row.ClaimedCount >= row.MaxClaims {
		return &ClaimByMobileResult{Success: false, FailureReason: "凭证已被领取", FailureCode: "consumed"}, nil
	}
	if row.Status == claimTokenStatusExpired || (row.ExpireAt != nil && row.ExpireAt.Before(now)) {
		return &ClaimByMobileResult{Success: false, FailureReason: "凭证已过期", FailureCode: "expired"}, nil
	}
	if row.Status != claimTokenStatusActive {
		return &ClaimByMobileResult{Success: false, FailureReason: "凭证状态无效", FailureCode: "invalid_status"}, nil
	}

	// 2. 命中接收人手机号校验（核心业务规则）
	if row.TargetMobile != "" && row.TargetMobile != input.Mobile {
		// 不命中：直接 return failure，但**不**消费 token，不让外人通过暴力测试占用 token
		return &ClaimByMobileResult{
			Success:       false,
			FailureReason: "此卡片仅限指定接收人领取",
			FailureCode:   "mismatch",
		}, nil
	}

	// 3. 查关联卡片
	var instance claimTokenInstanceRow
	if err := tx.WithContext(ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", row.CardInstanceID).
		Set("gorm:query_option", "FOR UPDATE").
		Take(&instance).Error; err != nil {
		return nil, fmt.Errorf("查询卡片失败: %w", err)
	}
	if instance.AssetStatus != cardAssetStatusClaimed {
		return &ClaimByMobileResult{Success: false, FailureReason: "卡片状态不允许领取", FailureCode: "invalid_card_status"}, nil
	}

	// 4. find-or-create 接收人会员
	memberID, err := findOrCreateMemberByMobile(ctx, tx, input.Mobile)
	if err != nil {
		return nil, fmt.Errorf("接收人会员处理失败: %w", err)
	}

	// 5. 转移卡片归属
	if updRes := tx.WithContext(ctx).
		Table(instance.TableName()).
		Where("id = ? AND member_id = ?", instance.ID, instance.MemberID).
		Updates(map[string]interface{}{
			"member_id":   memberID,
			"update_time": now,
		}); updRes.Error != nil {
		return nil, fmt.Errorf("卡片归属转移失败: %w", updRes.Error)
	} else if updRes.RowsAffected == 0 {
		return nil, errors.New("卡片归属转移失败：CAS 冲突")
	}

	// 6. 标记 token 已领取
	if updRes := tx.WithContext(ctx).
		Table(row.TableName()).
		Where("id = ? AND status = ? AND claimed_count = ?", row.ID, claimTokenStatusActive, row.ClaimedCount).
		Updates(map[string]interface{}{
			"status":        claimTokenStatusClaimed,
			"claimed_count": row.ClaimedCount + 1,
			"claimed_by":    memberID,
			"claimed_at":    now,
			"updated_at":    now,
		}); updRes.Error != nil {
		return nil, fmt.Errorf("凭证状态更新失败: %w", updRes.Error)
	} else if updRes.RowsAffected == 0 {
		return nil, errors.New("凭证状态更新失败：CAS 冲突，请重试")
	}

	// 7. 写资产转赠日志
	// sms_card_asset_log 实际表结构：asset_instance_id / from_status / to_status / operation_type /
	//   operator_type / trace_id / reason_code / reason_text / payload_json / create_time
	// 业务详细信息（operator_id / before_member_id / after_member_id / ip / request_id / scope）
	// 写到 payload_json 里。
	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"operator_id":      memberID,
		"before_member_id": row.IssuerID,
		"after_member_id":  memberID,
		"trigger_source":   "h5_claim_by_mobile",
		"request_id":       input.RequestID,
		"ip_address":       input.IPAddress,
		"platform_id":      row.PlatformID,
		"tenant_id":        row.TenantID,
		"merchant_id":      row.MerchantID,
		"target_mobile":    row.TargetMobile,
	})
	logRow := map[string]interface{}{
		"asset_instance_id":       instance.ID,
		"participation_record_id": 0, // NOT NULL，转赠场景无关联抽卡参与，写 0
		"from_status":             cardAssetStatusClaimed,
		"to_status":               cardAssetStatusClaimed, // holder 切换不改 asset_status
		"operation_type":          cardAssetOperationHolderTransferred,
		"operator_type":           "member",
		"trace_id":                input.RequestID,
		"reason_code":             "h5_claim_by_mobile",
		"reason_text":             "H5 朋友端按手机号一步式领取",
		"payload_json":            string(payloadBytes),
		"create_time":             now,
	}
	if err := tx.WithContext(ctx).
		Table("sms_card_asset_log").
		Create(logRow).Error; err != nil {
		// 日志写入失败不应阻塞领取，但也要 fail-fast 让上层感知
		return nil, fmt.Errorf("写入资产日志失败: %w", err)
	}

	// 8. 查模板名 / 卡面给 H5 显示
	type templatePreviewRow struct {
		TemplateName  string `gorm:"column:template_name"`
		CardFaceImage string `gorm:"column:card_face_image"`
	}
	var preview templatePreviewRow
	_ = tx.WithContext(ctx).
		Table("sms_card_template").
		Select("template_name, card_face_image").
		Where("id = ? AND is_deleted = 0", instance.TemplateID).
		Take(&preview).Error

	return &ClaimByMobileResult{
		Success: true,
		Card: &ClaimedCardDetail{
			CardInstanceID: instance.ID,
			AssetNo:        instance.AssetNo,
			TemplateName:   preview.TemplateName,
			CardFaceImage:  preview.CardFaceImage,
			NewMemberID:    memberID,
		},
	}, nil
}

// isValidChineseMobile 仅做最简单的格式校验：1[3-9] 开头共 11 位数字
func isValidChineseMobile(m string) bool {
	if len(m) != 11 {
		return false
	}
	if m[0] != '1' {
		return false
	}
	if m[1] < '3' || m[1] > '9' {
		return false
	}
	for i := 2; i < 11; i++ {
		if m[i] < '0' || m[i] > '9' {
			return false
		}
	}
	return true
}

// findOrCreateMemberByMobile 按手机号查 ums_member_info；不存在则插入一条新会员，返回 member_id。
//
// mock 阶段策略：
//   - nickname 用 mobile 后 4 位组成"九克城用户8888"
//   - password 写一个安全占位 hash（用户首次正式登录前应当走"忘记密码 / 重设"流程）
//   - source = 1 (APP)，level_id = 1（默认普通会员）
//
// 后期接真实注册流程时，替换这里改调 ums-rpc Register 即可，调用点只 in-process 看到 member_id。
func findOrCreateMemberByMobile(ctx context.Context, tx *gorm.DB, mobile string) (int64, error) {
	type memberRow struct {
		ID       int64  `gorm:"column:id"`
		MemberID int64  `gorm:"column:member_id"`
		Mobile   string `gorm:"column:mobile"`
	}
	var existing memberRow
	err := tx.WithContext(ctx).
		Table("ums_member_info").
		Select("id, member_id, mobile").
		Where("mobile = ? AND is_deleted = 0", mobile).
		Take(&existing).Error
	if err == nil {
		return existing.MemberID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	// 生成新 member_id：用 max(member_id) + 1，简单稳妥
	var maxRow struct {
		MaxID int64 `gorm:"column:max_id"`
	}
	if err := tx.WithContext(ctx).
		Table("ums_member_info").
		Select("COALESCE(MAX(member_id), 1000) AS max_id").
		Take(&maxRow).Error; err != nil {
		return 0, err
	}
	newMemberID := maxRow.MaxID + 1

	nickname := "九克城用户" + mobile[7:]
	now := time.Now()
	insertRow := map[string]interface{}{
		"member_id":          newMemberID,
		"wx_openid":          "",
		"level_id":           1,
		"nickname":           nickname,
		"mobile":             mobile,
		"source":             1,                                                           // APP
		"password":           "$2a$10$placeholder.bcrypt.hash.replace.via.password.reset", // 占位，正式登录前必须重设
		"avatar":             "",
		"signature":          "",
		"gender":             0,
		"growth_point":       0,
		"points":             0,
		"total_points":       0,
		"spend_amount":       0,
		"order_count":        0,
		"coupon_count":       0,
		"comment_count":      0,
		"return_count":       0,
		"lottery_times":      0,
		"first_login_status": 1,
		"is_enabled":         1,
		"create_time":        now,
		"is_deleted":         0,
	}
	if err := tx.WithContext(ctx).
		Table("ums_member_info").
		Create(insertRow).Error; err != nil {
		return 0, fmt.Errorf("创建会员失败: %w", err)
	}
	return newMemberID, nil
}
