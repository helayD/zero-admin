package cardclaimtokenservicelogic

import (
	"context"
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

	if instance.MemberID == in.ClaimedBy {
		return nil, errors.New("不能领取自己的卡片")
	}

	var existingClaimCount int64
	if err := tx.WithContext(l.ctx).
		Table(cardAssetLogRow{}.TableName()).
		Where("asset_instance_id = ? AND operation_type = ? AND payload_json LIKE ?", token.CardInstanceID, cardAssetOperationHolderTransferred, fmt.Sprintf("%%\"toHolderId\":%d%%", in.ClaimedBy)).
		Count(&existingClaimCount).Error; err != nil {
		return nil, err
	}
	if existingClaimCount > 0 {
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
	if err := tx.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ?", instance.ID).
		Updates(map[string]interface{}{
			"member_id":   in.ClaimedBy,
			"update_time": now,
		}).Error; err != nil {
		return nil, fmt.Errorf("更新卡片持有人失败: %w", err)
	}

	payload := fmt.Sprintf(`{"fromHolderId":%d,"toHolderId":%d,"claimTokenId":%d,"transferReason":"%s"}`, fromHolderID, in.ClaimedBy, token.ID, cardAssetReasonClaimTokenConsumed)
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
		PayloadJSON:           fmt.Sprintf(`{"token":"%s","tokenId":%d,"claimedBy":%d,"result":"%s","failureReason":"%s"}`, in.Token, 0, in.ClaimedBy, result, failureReason),
	}
	if token != nil {
		logRow.PayloadJSON = fmt.Sprintf(`{"token":"%s","tokenId":%d,"claimedBy":%d,"result":"%s","failureReason":"%s"}`, in.Token, token.ID, in.ClaimedBy, result, failureReason)
	}
	
	if err := l.svcCtx.DB.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		logc.Errorf(l.ctx, "记录领取尝试日志失败: %v", err)
	}
}
