package digital_card_asset

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
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
	// Story 10.7 安全要求 #5/#6：作用域必须从服务端身份上下文取，禁止信任客户端入参
	scope := currentGovernanceScope(l.ctx)

	resp, err := l.svcCtx.CardRedemptionOrderService.CreateRedemptionOrder(l.ctx, &cardredemptionorderservice.CreateRedemptionOrderReq{
		CardInstanceId:  req.CardInstanceId,
		HolderId:        memberID,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		PlatformId:      scope.PlatformID,
		TenantId:        scope.TenantID,
		MerchantId:      scope.MerchantID,
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

	scope := currentGovernanceScope(l.ctx)
	resp, err := l.svcCtx.CardRedemptionOrderService.QueryRedemptionOrder(l.ctx, &cardredemptionorderservice.QueryRedemptionOrderReq{
		OrderId:        req.OrderId,
		CardInstanceId: req.CardInstanceId,
		HolderId:       memberID,
		PlatformId:     scope.PlatformID,
		TenantId:       scope.TenantID,
		MerchantId:     scope.MerchantID,
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

// resolveShareDomain 解析并校验分享链接域名：
// - 若客户端未传，则使用配置白名单首项作为兜底域名，避免调用方未配置时失败
// - 若客户端传入，使用 url.Parse 严格解析，再与白名单精确匹配 host:port
// - 禁止含 userinfo（user@host）、路径、查询、IPv6 字面量等可能引入解析差异的字符
// - 仅允许：与白名单完全相等，或白名单为 a.com 时 sub.a.com 形式（不含端口）
// - Review #5 M3: host 比较使用 strings.EqualFold，遵循 RFC 3986 host 大小写不敏感
func resolveShareDomain(domain string, allowedDomains []string) (string, error) {
	if domain == "" {
		if len(allowedDomains) == 0 {
			return "", errors.New("分享域名未配置")
		}
		return allowedDomains[0], nil
	}
	// 禁止任何路径/查询/认证字符，提前阻断绕过攻击向量
	if strings.ContainsAny(domain, "/?#&=\\@[]") {
		return "", errors.New("域名格式非法")
	}
	// 通过 url.Parse 严格解析 host:port，拒绝歧义 URL（host 大小写不敏感比较）
	parsed, err := url.Parse("https://" + domain)
	if err != nil || parsed.Host == "" || !strings.EqualFold(parsed.Host, domain) {
		return "", errors.New("域名格式非法")
	}
	domainLower := strings.ToLower(domain)
	for _, allowed := range allowedDomains {
		allowedLower := strings.ToLower(allowed)
		if domainLower == allowedLower {
			return domain, nil
		}
		// 子域名匹配仅当白名单不含端口时生效，避免 a.com:9999 与 sub.a.com:9999 歧义
		if !strings.Contains(allowedLower, ":") && !strings.Contains(domainLower, ":") && strings.HasSuffix(domainLower, "."+allowedLower) {
			return domain, nil
		}
	}
	return "", errors.New("域名不在白名单内")
}

func (l *GenerateShareLinkLogic) GenerateShareLink(req *types.GenerateShareLinkReq) (*types.GenerateShareLinkResp, error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}

	domain, err := resolveShareDomain(req.Domain, l.svcCtx.Config.Share.AllowedDomains)
	if err != nil {
		return nil, err
	}

	scope := currentGovernanceScope(l.ctx)
	// Story 10.7 Task 8.x — 因 CardClaimTokenService 的手写 struct 未实现 proto.Message
	// 跨进程 gRPC marshal 必然失败，分享凭证生成改走 pkg/digitalcardmint 的 in-process
	// 服务，直接读写本进程 DB（与 sms-rpc 共享同一个 MySQL 实例）。
	tokenResult, err := l.svcCtx.CardMintService.GenerateClaimToken(l.ctx, scope, digitalcardmint.GenerateClaimTokenInput{
		CardInstanceID: req.CardInstanceId,
		IssuerID:       memberID,
		ExpireHours:    req.ExpireHours,
		MaxClaims:      req.MaxClaims,
		TraceID:        req.TraceId,
		RequestID:      req.RequestId,
	})
	if err != nil {
		return nil, fmt.Errorf("生成分享链接失败: %w", err)
	}

	// 分享链接指向 H5 领取页，朋友扫码或点击后由 H5 引导登录/注册并完成领取
	// 若配置域名已带 scheme 则直接使用；否则按是否为内网 IP/端口动态选择 http/https
	baseURL := domain
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		scheme := "https"
		if strings.Contains(baseURL, ":") || strings.Contains(baseURL, "127.0.0.1") || strings.Contains(baseURL, "localhost") {
			// 含端口或本地/内网地址，默认走 http，便于联调
			scheme = "http"
		}
		baseURL = scheme + "://" + baseURL
	}
	shareLink := fmt.Sprintf("%s/h5/digital-card/claim?token=%s", baseURL, tokenResult.Token)

	return &types.GenerateShareLinkResp{
		Code:    0,
		Message: "生成分享链接成功",
		Data: types.ShareLinkData{
			Token:     tokenResult.Token,
			ShareLink: shareLink,
			ExpireAt:  tokenResult.ExpireAt,
			MaxClaims: tokenResult.MaxClaims,
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
	// C-2: 携带领取人的作用域以便 RPC 端校验跨租户/平台越权
	scope := currentGovernanceScope(l.ctx)
	consumeResp, err := l.svcCtx.CardClaimTokenService.ConsumeClaimToken(l.ctx, &cardclaimtokenservice.ConsumeClaimTokenReq{
		Token:      req.Token,
		ClaimedBy:  memberID,
		PlatformId: scope.PlatformID,
		TenantId:   scope.TenantID,
		MerchantId: scope.MerchantID,
		TraceId:    req.TraceId,
		RequestId:  req.RequestId,
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
	// Story 10.7 Task 8.x — 同 GenerateClaimToken，绕过坏的 CardClaimTokenService gRPC，
	// 直接走 in-process service。匿名接口（H5 朋友打开链接），不需要 scope。
	result, err := l.svcCtx.CardMintService.ValidateClaimToken(l.ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("校验凭证失败: %w", err)
	}

	var tokenData *types.ClaimTokenData
	if result.Token != nil {
		tokenData = &types.ClaimTokenData{
			Id:             result.Token.ID,
			Token:          result.Token.Token,
			CardInstanceId: result.Token.CardInstanceID,
			IssuerId:       result.Token.IssuerID,
			IssuerType:     result.Token.IssuerType,
			ExpireAt:       result.Token.ExpireAt,
			MaxClaims:      result.Token.MaxClaims,
			ClaimedCount:   result.Token.ClaimedCount,
			Status:         result.Token.Status,
			ClaimedBy:      result.Token.ClaimedBy,
			ClaimedAt:      result.Token.ClaimedAt,
			PlatformId:     result.Token.PlatformID,
			TenantId:       result.Token.TenantID,
			MerchantId:     result.Token.MerchantID,
			CreateTime:     result.Token.CreateTime,
			UpdateTime:     result.Token.UpdateTime,
			TemplateName:   result.TemplateName,
			CardFaceImage:  result.CardFaceImage,
		}
	}

	return &types.ValidateClaimTokenResp{
		Code:    "0",
		Message: "校验成功",
		Data: types.ValidateClaimTokenData{
			Valid:         result.Valid,
			FailureReason: result.FailureReason,
			Token:         tokenData,
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
