package digital_card_asset

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveDigitalCardTransferRecipientLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveDigitalCardTransferRecipientLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveDigitalCardTransferRecipientLogic {
	return &ResolveDigitalCardTransferRecipientLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResolveDigitalCardTransferRecipientLogic) ResolveDigitalCardTransferRecipient(req *types.ResolveDigitalCardTransferRecipientReq) (*types.ResolveDigitalCardTransferRecipientResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintService.ResolveAssetTransferRecipient(l.ctx, currentGovernanceScope(l.ctx), digitalcardmint.AssetTransferRecipientInput{
		AssetInstanceID:   req.AssetInstanceId,
		FromMemberID:      memberID,
		RecipientMobile:   req.RecipientMobile,
		RegisterH5BaseURL: "/h5/digitalCard/register",
	})
	if err != nil {
		return nil, assetServiceError(l.ctx, "识别数字卡片转赠接收人", req, err)
	}
	return &types.ResolveDigitalCardTransferRecipientResp{
		Code:    assetCodeSuccess,
		Message: "识别数字卡片转赠接收人成功",
		Data: types.DigitalCardTransferRecipientData{
			RecipientStatus:         result.RecipientStatus,
			RecipientStatusText:     result.RecipientStatusText,
			RecipientMemberId:       result.RecipientMemberID,
			RecipientNicknameMasked: result.RecipientNicknameMasked,
			RecipientMobileMasked:   result.RecipientMobileMasked,
			RegisterUrl:             result.RegisterURL,
			CanTransfer:             result.CanTransfer,
			ActionHint:              result.ActionHint,
		},
	}, nil
}

type TransferDigitalCardAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTransferDigitalCardAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferDigitalCardAssetLogic {
	return &TransferDigitalCardAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TransferDigitalCardAssetLogic) TransferDigitalCardAsset(req *types.TransferDigitalCardAssetReq) (*types.TransferDigitalCardAssetResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintService.TransferDigitalCardAsset(l.ctx, currentGovernanceScope(l.ctx), digitalcardmint.TransferDigitalCardAssetInput{
		AssetInstanceID: req.AssetInstanceId,
		FromMemberID:    memberID,
		RecipientMobile: req.RecipientMobile,
		RequestID:       req.RequestId,
	})
	if err != nil {
		return nil, assetServiceError(l.ctx, "转赠数字卡片", req, err)
	}
	return &types.TransferDigitalCardAssetResp{
		Code:    assetCodeSuccess,
		Message: "转赠数字卡片成功",
		Data: types.TransferDigitalCardAssetData{
			AssetInstanceId:         result.AssetInstanceID,
			FromMemberId:            result.FromMemberID,
			ToMemberId:              result.ToMemberID,
			RecipientNicknameMasked: result.RecipientNicknameMasked,
			RecipientMobileMasked:   result.RecipientMobileMasked,
			TransferStatus:          result.TransferStatus,
			TransferStatusText:      result.TransferStatusText,
		},
	}, nil
}

type RequestDigitalCardWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRequestDigitalCardWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestDigitalCardWithdrawLogic {
	return &RequestDigitalCardWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RequestDigitalCardWithdrawLogic) RequestDigitalCardWithdraw(req *types.RequestDigitalCardWithdrawReq) (*types.RequestDigitalCardWithdrawResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.CardMintService.RequestDigitalCardAssetWithdraw(l.ctx, currentGovernanceScope(l.ctx), digitalcardmint.RequestDigitalCardAssetWithdrawInput{
		AssetInstanceID: req.AssetInstanceId,
		MemberID:        memberID,
		Reason:          req.Reason,
		RequestID:       req.RequestId,
	})
	if err != nil {
		return nil, assetServiceError(l.ctx, "提交数字卡片提现申请", req, err)
	}
	return &types.RequestDigitalCardWithdrawResp{
		Code:    assetCodeSuccess,
		Message: "提交数字卡片提现申请成功",
		Data: types.RequestDigitalCardWithdrawData{
			AssetInstanceId: result.AssetInstanceID,
			WithdrawStatus:  result.WithdrawStatus,
			WithdrawText:    result.WithdrawText,
		},
	}, nil
}
