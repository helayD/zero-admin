package cardclaimtokenservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
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
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table(token.TableName()).
		Where("id = ?", token.ID).
		Updates(map[string]interface{}{
			"status":     claimTokenStatusRevoked,
			"updated_at": now,
		}).Error; err != nil {
		return nil, err
	}

	return &smsclient.RevokeClaimTokenResp{
		Success: true,
		Message: "凭证已吊销",
	}, nil
}
