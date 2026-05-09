package digital_card_asset

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/client/cardclaimtokenservice"
	"github.com/feihua/zero-admin/rpc/sms/client/cardredemptionorderservice"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)


type CreateRedemptionOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRedemptionOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRedemptionOrderLogic {
	return &CreateRedemptionOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRedemptionOrderLogic) CreateRedemptionOrder(req *types.CreateRedemptionOrderReq) (*types.CreateRedemptionOrderResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.CardRedemptionOrderService.CreateRedemptionOrder(l.ctx, &cardredemptionorderservice.CreateRedemptionOrderReq{
		CardInstanceId:  req.CardInstanceId,
		HolderId:        memberID,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		PlatformId:      req.PlatformId,
		TenantId:        req.TenantId,
		MerchantId:      req.MerchantId,
		TraceId:         req.TraceId,
		RequestId:       req.RequestId,
	})
	if err != nil {
		return nil, fmt.Errorf("创建提货单失败: %w", err)
	}

	return &types.CreateRedemptionOrderResp{
		Code:    0,
		Message: "创建提货单成功",
		Data:    convertRedemptionOrderData(resp.Order),
	}, nil
}

type QueryRedemptionOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryRedemptionOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRedemptionOrderLogic {
	return &QueryRedemptionOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryRedemptionOrderLogic) QueryRedemptionOrder(req *types.QueryRedemptionOrderReq) (*types.QueryRedemptionOrderResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.CardRedemptionOrderService.QueryRedemptionOrder(l.ctx, &cardredemptionorderservice.QueryRedemptionOrderReq{
		OrderId:        req.OrderId,
		CardInstanceId: req.CardInstanceId,
		HolderId:       memberID,
	})
	if err != nil {
		return nil, fmt.Errorf("查询提货单失败: %w", err)
	}

	return &types.QueryRedemptionOrderResp{
		Code:    0,
		Message: "查询提货单成功",
		Data:    convertRedemptionOrderData(resp.Order),
	}, nil
}

type GenerateShareLinkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateShareLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateShareLinkLogic {
	return &GenerateShareLinkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// validateShareDomain 校验分享链接域名是否在白名单内，防止 SSRF 和钓鱼链接
func validateShareDomain(domain string, allowedDomains []string) error {
	if domain == "" {
		return errors.New("域名不能为空")
	}
	// 防止路径注入、查询参数注入等攻击
	if strings.ContainsAny(domain, "/?#&=\\") {
		return errors.New("域名格式非法")
	}
	for _, allowed := range allowedDomains {
		if domain == allowed || strings.HasSuffix(domain, "."+allowed) {
			return nil
		}
	}
	return errors.New("域名不在白名单内")
}

func (l *GenerateShareLinkLogic) GenerateShareLink(req *types.GenerateShareLinkReq) (*types.GenerateShareLinkResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	if err := validateShareDomain(req.Domain, l.svcCtx.Config.Share.AllowedDomains); err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.CardClaimTokenService.GenerateClaimToken(l.ctx, &cardclaimtokenservice.GenerateClaimTokenReq{
		CardInstanceId: req.CardInstanceId,
		IssuerId:       memberID,
		ExpireHours:    req.ExpireHours,
		MaxClaims:      req.MaxClaims,
		PlatformId:     req.PlatformId,
		TenantId:       req.TenantId,
		MerchantId:     req.MerchantId,
		TraceId:        req.TraceId,
		RequestId:      req.RequestId,
	})
	if err != nil {
		return nil, fmt.Errorf("生成分享链接失败: %w", err)
	}

	shareLink := fmt.Sprintf("https://%s/api/digitalCard/claim?token=%s", req.Domain, resp.Token.Token)

	return &types.GenerateShareLinkResp{
		Code:    0,
		Message: "生成分享链接成功",
		Data: types.ShareLinkData{
			Token:     resp.Token.Token,
			ShareLink: shareLink,
			ExpireAt:  resp.Token.ExpireAt,
			MaxClaims: resp.Token.MaxClaims,
		},
	}, nil
}

func convertRedemptionOrderData(data *smsclient.RedemptionOrderData) *types.RedemptionOrderData {
	if data == nil {
		return nil
	}
	return &types.RedemptionOrderData{
		Id:              data.Id,
		OrderNo:         data.OrderNo,
		CardInstanceId:  data.CardInstanceId,
		HolderId:        data.HolderId,
		ReceiverName:    data.ReceiverName,
		ReceiverPhone:   data.ReceiverPhone,
		ReceiverAddress: data.ReceiverAddress,
		Status:          data.Status,
		ShippedAt:       data.ShippedAt,
		DeliveredAt:     data.DeliveredAt,
		CancelReason:    data.CancelReason,
		OmsOrderId:      data.OmsOrderId,
		PlatformId:      data.PlatformId,
		TenantId:        data.TenantId,
		MerchantId:      data.MerchantId,
		CreateTime:      data.CreateTime,
		UpdateTime:      data.UpdateTime,
	}
}

type ClaimDigitalCardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClaimDigitalCardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimDigitalCardLogic {
	return &ClaimDigitalCardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClaimDigitalCardLogic) ClaimDigitalCard(req *types.ClaimDigitalCardReq) (*types.ClaimDigitalCardResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	// ConsumeClaimToken 内部已包含完整的校验逻辑（token 有效性、过期、claimed_count、卡片状态等）
	// 直接调用即可消除 Validate+Consume 之间的竞态窗口
	consumeResp, err := l.svcCtx.CardClaimTokenService.ConsumeClaimToken(l.ctx, &cardclaimtokenservice.ConsumeClaimTokenReq{
		Token:     req.Token,
		ClaimedBy: memberID,
		TraceId:   req.TraceId,
		RequestId: req.RequestId,
	})
	if err != nil {
		return nil, fmt.Errorf("领取提货卡失败: %w", err)
	}
	if !consumeResp.Success {
		return nil, errors.New(consumeResp.FailureReason)
	}

	return &types.ClaimDigitalCardResp{
		Code:    0,
		Message: "领取提货卡成功",
		Data: types.ClaimDigitalCardData{
			CardInstanceId: consumeResp.CardInstance.Id,
			AssetNo:        consumeResp.CardInstance.AssetNo,
			Status:         consumeResp.CardInstance.AssetStatus,
		},
	}, nil
}

type ValidateClaimTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidateClaimTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateClaimTokenLogic {
	return &ValidateClaimTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ValidateClaimTokenLogic) ValidateClaimToken(req *types.ValidateClaimTokenReq) (*types.ValidateClaimTokenResp, error) {
	resp, err := l.svcCtx.CardClaimTokenService.ValidateClaimToken(l.ctx, &cardclaimtokenservice.ValidateClaimTokenReq{
		Token: req.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("校验凭证失败: %w", err)
	}

	return &types.ValidateClaimTokenResp{
		Code:    "0",
		Message: "校验成功",
		Data: types.ValidateClaimTokenData{
			Valid:         resp.Valid,
			FailureReason: resp.FailureReason,
			Token:         convertClaimTokenData(resp.Token),
		},
	}, nil
}

func convertClaimTokenData(data *smsclient.ClaimTokenData) *types.ClaimTokenData {
	if data == nil {
		return nil
	}
	return &types.ClaimTokenData{
		Id:             data.Id,
		Token:          data.Token,
		CardInstanceId: data.CardInstanceId,
		IssuerId:       data.IssuerId,
		IssuerType:     data.IssuerType,
		ExpireAt:       data.ExpireAt,
		MaxClaims:      data.MaxClaims,
		ClaimedCount:   data.ClaimedCount,
		Status:         data.Status,
		ClaimedBy:      data.ClaimedBy,
		ClaimedAt:      data.ClaimedAt,
		PlatformId:     data.PlatformId,
		TenantId:       data.TenantId,
		MerchantId:     data.MerchantId,
		CreateTime:     data.CreateTime,
		UpdateTime:     data.UpdateTime,
	}
}
