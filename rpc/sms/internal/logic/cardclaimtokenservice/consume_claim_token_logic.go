package cardclaimtokenservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ConsumeClaimTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeClaimTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeClaimTokenLogic {
	return &ConsumeClaimTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConsumeClaimTokenLogic) ConsumeClaimToken(in *smsclient.ConsumeClaimTokenReq) (*smsclient.ConsumeClaimTokenResp, error) {
	if in.Token == "" {
		l.recordClaimAttempt(in, nil, 0, "failed", "凭证不能为空")
		return &smsclient.ConsumeClaimTokenResp{
			Success:       false,
			FailureReason: "凭证不能为空",
		}, nil
	}
	if in.ClaimedBy <= 0 {
		l.recordClaimAttempt(in, nil, 0, "failed", "领取人ID无效")
		return &smsclient.ConsumeClaimTokenResp{
			Success:       false,
			FailureReason: "领取人ID无效",
		}, nil
	}

	var result *consumeResult
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		result, txErr = l.consumeClaimTokenInTx(tx, in)
		return txErr
	})
	if err != nil {
		logc.Errorf(l.ctx, "消耗分享凭证失败: %v, token=%s", err, in.Token)
		l.recordClaimAttempt(in, nil, 0, "failed", err.Error())
		return &smsclient.ConsumeClaimTokenResp{
			Success:       false,
			FailureReason: err.Error(),
		}, nil
	}

	l.recordClaimAttempt(in, result.token, result.cardInstance.Id, "success", "")

	return &smsclient.ConsumeClaimTokenResp{
		Success:      true,
		Token:        buildClaimTokenData(result.token),
		CardInstance: result.cardInstance,
	}, nil
}

type consumeResult struct {
	token        *claimTokenRow
	cardInstance *smsclient.CardInstanceData
}

func (l *ConsumeClaimTokenLogic) consumeClaimTokenInTx(tx *gorm.DB, in *smsclient.ConsumeClaimTokenReq) (*consumeResult, error) {
	var token claimTokenRow
	if err := tx.WithContext(l.ctx).
		Table(token.TableName()).
		Where("token = ? AND is_deleted = 0", in.Token).
		Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("凭证不存在")
		}
		return nil, err
	}

	if token.Status != claimTokenStatusActive {
		return nil, fmt.Errorf("凭证状态无效:%s", token.Status)
	}
	if token.ExpireAt != nil && token.ExpireAt.Before(time.Now()) {
		return nil, errors.New("凭证已过期")
	}
	if token.ClaimedCount >= token.MaxClaims {
		return nil, errors.New("凭证领取次数已达上限")
	}

	// C-2 / Review #5 H2: 关闭 fail-open——只要 token 自身带有任意作用域，
	// 请求必须精确匹配；若调用方未携带（全 0）将直接被拒绝，禁止匿名/零作用域绕过。
	tokenHasScope := token.PlatformID > 0 || token.TenantID > 0 || token.MerchantID > 0
	if tokenHasScope {
		if token.PlatformID != in.PlatformId || token.TenantID != in.TenantId || token.MerchantID != in.MerchantId {
			return nil, errors.New("凭证作用域与当前用户不匹配")
		}
	}

	var instance cardInstanceRow
	if err := tx.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", token.CardInstanceID).
		Take(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片实例不存在")
		}
		return nil, err
	}

	if instance.AssetStatus != cardAssetStatusClaimed {
		return nil, fmt.Errorf("卡片状态不允许领取:%s", instance.AssetStatus)
	}

	// 卡片实例作用域必须与 token 一致，确保历史数据/外部调用不会跨作用域转移
	if instance.PlatformID != token.PlatformID || instance.TenantID != token.TenantID || instance.MerchantID != token.MerchantID {
		return nil, errors.New("卡片实例作用域与凭证不一致")
	}

	// 当前卡片持有人 = 领取人：不能自己给自己领
	if instance.MemberID == in.ClaimedBy {
		return nil, errors.New("您已持有该卡片，无需重复领取")
	}

	// 检查历史领取记录：精确匹配 payload 中的 toHolderId，避免
	// LIKE '%"toHolderId":101%' 误匹配 toHolderId:1011 等相邻数值的伪阳性。
	// 匹配两种合法边界：后跟 "," 表示还有其他字段，或后跟 "}" 表示末尾字段。
	// 建议为 sms_card_asset_log 增加复合索引 (asset_instance_id, operation_type) 优化此查询
	var existingLog cardAssetLogRow
	claimedKey := fmt.Sprintf("\"toHolderId\":%d", in.ClaimedBy)
	if err := tx.WithContext(l.ctx).
		Table(cardAssetLogRow{}.TableName()).
		Where("asset_instance_id = ? AND operation_type = ? AND (payload_json LIKE ? OR payload_json LIKE ?)",
			token.CardInstanceID,
			cardAssetOperationHolderTransferred,
			"%"+claimedKey+",%",
			"%"+claimedKey+"}%").
		Limit(1).
		Take(&existingLog).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, errors.New("您已领取过该卡片")
	}

	now := time.Now()
	newClaimedCount := token.ClaimedCount + 1
	newStatus := claimTokenStatusActive
	if newClaimedCount >= token.MaxClaims {
		newStatus = claimTokenStatusClaimed
	}

	updateResult := tx.WithContext(l.ctx).
		Table(token.TableName()).
		Where("id = ? AND version = ? AND claimed_count < max_claims", token.ID, token.Version).
		Updates(map[string]interface{}{
			"claimed_count": newClaimedCount,
			"claimed_by":    in.ClaimedBy,
			"claimed_at":    now,
			"status":        newStatus,
			"version":       token.Version + 1,
			"updated_at":    now,
		})
	if updateResult.Error != nil {
		return nil, fmt.Errorf("更新凭证状态失败: %w", updateResult.Error)
	}
	if updateResult.RowsAffected == 0 {
		return nil, errors.New("领取失败，凭证可能已被他人领取或并发冲突")
	}

	fromHolderID := instance.MemberID
	// H-1: 并发保护——UPDATE 必须带 member_id = fromHolderID，避免多个事务同时把
	// instance.MemberID 改给不同领取人；若 RowsAffected==0 说明已被并发抢先，回滚事务。
	instanceUpd := tx.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ? AND member_id = ? AND is_deleted = 0", instance.ID, fromHolderID).
		Updates(map[string]interface{}{
			"member_id":   in.ClaimedBy,
			"update_time": now,
		})
	if instanceUpd.Error != nil {
		return nil, fmt.Errorf("更新卡片持有人失败: %w", instanceUpd.Error)
	}
	if instanceUpd.RowsAffected == 0 {
		return nil, errors.New("卡片持有人已被并发更新，请重试")
	}

	// M-3: payload 用 json.Marshal，避免 fmt.Sprintf 拼接遇到特殊字符时被 JSON 注入
	payloadBytes, mErr := json.Marshal(map[string]interface{}{
		"fromHolderId":   fromHolderID,
		"toHolderId":     in.ClaimedBy,
		"claimTokenId":   token.ID,
		"transferReason": cardAssetReasonClaimTokenConsumed,
	})
	if mErr != nil {
		return nil, fmt.Errorf("序列化审计负载失败: %w", mErr)
	}
	payload := string(payloadBytes)
	logRow := &cardAssetLogRow{
		AssetInstanceID:       instance.ID,
		ParticipationRecordID: 0,
		FromStatus:            instance.AssetStatus,
		ToStatus:              instance.AssetStatus,
		OperationType:         cardAssetOperationHolderTransferred,
		OperatorType:          "member",
		TraceID:               in.TraceId,
		ReasonCode:            cardAssetReasonClaimTokenConsumed,
		ReasonText:            "通过分享凭证领取提货卡",
		PayloadJSON:           payload,
	}
	if err := tx.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		return nil, fmt.Errorf("记录资产日志失败: %w", err)
	}

	token.ClaimedCount = newClaimedCount
	token.ClaimedBy = in.ClaimedBy
	token.Status = newStatus

	return &consumeResult{
		token: &token,
		cardInstance: &smsclient.CardInstanceData{
			Id:          instance.ID,
			AssetNo:     instance.AssetNo,
			AssetStatus: instance.AssetStatus,
			MintStatus:  instance.MintStatus,
			TemplateId:  instance.TemplateID,
			MemberId:    in.ClaimedBy,
		},
	}, nil
}

func (l *ConsumeClaimTokenLogic) recordClaimAttempt(in *smsclient.ConsumeClaimTokenReq, token *claimTokenRow, cardInstanceID int64, result string, failureReason string) {
	cardInstanceIDVal := cardInstanceID
	if token != nil {
		cardInstanceIDVal = token.CardInstanceID
	}

	// 当 token 为 nil 但 in.Token 不为空时，尝试反查数据库获取凭证信息
	// 确保审计日志能关联到具体凭证
	if token == nil && in.Token != "" {
		var found claimTokenRow
		if err := l.svcCtx.DB.WithContext(l.ctx).Table(found.TableName()).
			Where("token = ? AND is_deleted = 0", in.Token).
			Take(&found).Error; err == nil {
			token = &found
			cardInstanceIDVal = found.CardInstanceID
		}
	}

	// L-2 + M-3: payload 用 json.Marshal 避免拼接歧义，token 不存在时 tokenId=0 表示未识别
	tokenIDVal := int64(0)
	if token != nil {
		tokenIDVal = token.ID
	}
	payloadBytes, mErr := json.Marshal(map[string]interface{}{
		"token":         in.Token,
		"tokenId":       tokenIDVal,
		"claimedBy":     in.ClaimedBy,
		"result":        result,
		"failureReason": failureReason,
	})
	payloadJSON := ""
	if mErr != nil {
		logc.Errorf(l.ctx, "序列化领取尝试 payload 失败: %v", mErr)
		payloadJSON = `{"serializeError":true}`
	} else {
		payloadJSON = string(payloadBytes)
	}
	logRow := &cardAssetLogRow{
		AssetInstanceID:       cardInstanceIDVal,
		ParticipationRecordID: 0,
		FromStatus:            "",
		ToStatus:              "",
		OperationType:         "claim_attempt",
		OperatorType:          "member",
		TraceID:               in.TraceId,
		ReasonCode:            result,
		ReasonText:            fmt.Sprintf("领取尝试:%s", result),
		PayloadJSON:           payloadJSON,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		logc.Errorf(l.ctx, "记录领取尝试日志失败: %v", err)
	}
}
