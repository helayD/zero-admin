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

type RevokeClaimTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeClaimTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeClaimTokenLogic {
	return &RevokeClaimTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RevokeClaimTokenLogic) RevokeClaimToken(in *smsclient.RevokeClaimTokenReq) (*smsclient.RevokeClaimTokenResp, error) {
	if in.TokenId <= 0 {
		return &smsclient.RevokeClaimTokenResp{
			Success: false,
			Message: "凭证ID无效",
		}, nil
	}

	var token claimTokenRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table(token.TableName()).
		Where("id = ? AND is_deleted = 0", in.TokenId).
		Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &smsclient.RevokeClaimTokenResp{
				Success: false,
				Message: "凭证不存在",
			}, nil
		}
		return nil, err
	}

	if token.IssuerID != in.IssuerId {
		return &smsclient.RevokeClaimTokenResp{
			Success: false,
			Message: "只有分享人才能吊销凭证",
		}, nil
	}

	if token.Status != claimTokenStatusActive {
		return &smsclient.RevokeClaimTokenResp{
			Success: false,
			Message: "只有活跃状态的凭证才能吊销",
		}, nil
	}

	now := time.Now()
	// M-5: 吊销操作改在事务内完成，确保 token 状态变更与 asset_log 审计原子写入
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// CAS：仅当 status 仍为 active 才能吊销，防止并发竞态
		upd := tx.WithContext(l.ctx).
			Table(token.TableName()).
			Where("id = ? AND status = ? AND is_deleted = 0", token.ID, claimTokenStatusActive).
			Updates(map[string]interface{}{
				"status":     claimTokenStatusRevoked,
				"updated_at": now,
			})
		if upd.Error != nil {
			return upd.Error
		}
		if upd.RowsAffected == 0 {
			return errors.New("凭证状态已变化，吊销失败")
		}

		payloadBytes, mErr := json.Marshal(map[string]interface{}{
			"tokenId":    token.ID,
			"cardId":     token.CardInstanceID,
			"issuerId":   token.IssuerID,
			"reason":     in.Reason,
			"fromStatus": claimTokenStatusActive,
			"toStatus":   claimTokenStatusRevoked,
		})
		if mErr != nil {
			return fmt.Errorf("序列化审计负载失败: %w", mErr)
		}
		logRow := &cardAssetLogRow{
			AssetInstanceID:       token.CardInstanceID,
			ParticipationRecordID: 0,
			FromStatus:            "",
			ToStatus:              "",
			OperationType:         "claim_token_revoked",
			OperatorType:          "member",
			TraceID:               in.TraceId,
			ReasonCode:            "claim_token_revoked",
			ReasonText:            "分享人吊销分享凭证",
			PayloadJSON:           string(payloadBytes),
		}
		if err := tx.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
			return fmt.Errorf("记录吊销审计日志失败: %w", err)
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "吊销分享凭证失败: %v, tokenId=%d", err, token.ID)
		return &smsclient.RevokeClaimTokenResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &smsclient.RevokeClaimTokenResp{
		Success: true,
		Message: "凭证已吊销",
	}, nil
}
