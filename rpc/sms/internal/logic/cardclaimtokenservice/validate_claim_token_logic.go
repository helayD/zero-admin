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

type ValidateClaimTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateClaimTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateClaimTokenLogic {
	return &ValidateClaimTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateClaimTokenLogic) ValidateClaimToken(in *smsclient.ValidateClaimTokenReq) (*smsclient.ValidateClaimTokenResp, error) {
	if in.Token == "" {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "凭证不能为空",
		}, nil
	}

	var row claimTokenRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table(row.TableName()).
		Where("token = ? AND is_deleted = 0", in.Token).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &smsclient.ValidateClaimTokenResp{
				Valid:         false,
				FailureReason: "凭证不存在",
			}, nil
		}
		return nil, err
	}

	if row.Status == claimTokenStatusRevoked {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "凭证已被吊销",
		}, nil
	}

	if row.Status == claimTokenStatusClaimed {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "凭证已被完全领取",
		}, nil
	}

	if row.Status == claimTokenStatusExpired || (row.ExpireAt != nil && row.ExpireAt.Before(time.Now())) {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "凭证已过期",
		}, nil
	}

	if row.Status != claimTokenStatusActive {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "凭证状态无效",
		}, nil
	}

	if row.ClaimedCount >= row.MaxClaims {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "凭证领取次数已达上限",
		}, nil
	}

	var instance cardInstanceRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", row.CardInstanceID).
		Take(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &smsclient.ValidateClaimTokenResp{
				Valid:         false,
				FailureReason: "关联的卡片实例不存在",
			}, nil
		}
		return nil, err
	}

	if instance.AssetStatus != cardAssetStatusClaimed {
		return &smsclient.ValidateClaimTokenResp{
			Valid:         false,
			FailureReason: "卡片状态不允许领取",
		}, nil
	}

	data := buildClaimTokenData(&row)
	// 匿名接口不暴露分享人敏感信息
	if data != nil {
		data.IssuerId = 0
		data.IssuerType = ""
	}
	return &smsclient.ValidateClaimTokenResp{
		Valid: true,
		Token: data,
	}, nil
}
