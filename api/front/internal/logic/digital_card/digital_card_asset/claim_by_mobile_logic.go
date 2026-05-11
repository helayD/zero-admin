package digital_card_asset

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// Story 10.7 闭环修复（2026-05-12）
// H5 朋友端一步式领取入口：手机号 + 验证码（mock=123456）。
//
// 三个匿名接口在同一个文件里：
//   - SendClaimVerifyCode: mock 验证码下发
//   - ClaimByMobile:       一步式领取（命中校验+自动注册+转赠）
//   - GetAppDownload:      读 sys_system_config 中 digital_card_app_download 配置
//
// 都不需要 JWT，因为「朋友」此刻还没有 App 端账号。

// ============================================================
// SendClaimVerifyCode
// ============================================================

type SendClaimVerifyCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendClaimVerifyCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendClaimVerifyCodeLogic {
	return &SendClaimVerifyCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendClaimVerifyCodeLogic) SendClaimVerifyCode(req *types.SendClaimVerifyCodeReq) (*types.SendClaimVerifyCodeResp, error) {
	if strings.TrimSpace(req.Token) == "" {
		return nil, errors.New("分享凭证不能为空")
	}
	if !isValidChineseMobile(req.Mobile) {
		return nil, errors.New("请输入有效的手机号")
	}
	// Mock 阶段：不真发短信，前端拿到 MockCode 直接填入。
	// 后期接真短信通道时：此处下发 6 位随机码到 Redis（key=mobile+token，TTL=60s），
	//   并把 MockCode 字段清空。
	return &types.SendClaimVerifyCodeResp{
		Code:    0,
		Message: "验证码已发送",
		Data: types.SendClaimVerifyCodeRespData{
			MockCode:    "123456",
			CountdownMs: 60000,
		},
	}, nil
}

// ============================================================
// ClaimByMobile
// ============================================================

type ClaimByMobileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request // 仅用于读 X-Forwarded-For / RemoteAddr
}

func NewClaimByMobileLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ClaimByMobileLogic {
	return &ClaimByMobileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *ClaimByMobileLogic) ClaimByMobile(req *types.ClaimByMobileReq) (*types.ClaimByMobileResp, error) {
	if l.svcCtx.CardMintService == nil {
		return nil, errors.New("领取服务未初始化")
	}

	clientIP := ""
	if l.r != nil {
		if v := l.r.Header.Get("X-Forwarded-For"); v != "" {
			clientIP = strings.TrimSpace(strings.Split(v, ",")[0])
		}
		if clientIP == "" {
			clientIP = l.r.RemoteAddr
		}
	}

	res, err := l.svcCtx.CardMintService.ClaimByMobile(l.ctx, digitalcardmint.ClaimByMobileInput{
		Token:      strings.TrimSpace(req.Token),
		Mobile:     strings.TrimSpace(req.Mobile),
		VerifyCode: strings.TrimSpace(req.VerifyCode),
		RequestID:  req.RequestId,
		IPAddress:  clientIP,
	})
	if err != nil {
		return nil, err
	}

	// 不论成功失败，都把 App 下载配置一并下传，给 H5 显示「下载 App 注册 / 下载 App 查看」
	appCfg, _ := loadAppDownloadConfig(l.ctx, l.svcCtx)

	resp := &types.ClaimByMobileResp{
		Code:    0,
		Message: "ok",
		Data: types.ClaimByMobileData{
			Success:       res.Success,
			FailureReason: res.FailureReason,
			FailureCode:   res.FailureCode,
			AppDownload:   appCfg,
		},
	}
	if res.Success && res.Card != nil {
		resp.Data.Card = &types.ClaimedCardSummary{
			CardInstanceId: res.Card.CardInstanceID,
			AssetNo:        res.Card.AssetNo,
			TemplateName:   res.Card.TemplateName,
			CardFaceImage:  res.Card.CardFaceImage,
		}
	}
	return resp, nil
}

// ============================================================
// GetAppDownload
// ============================================================

type GetAppDownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppDownloadLogic {
	return &GetAppDownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppDownloadLogic) GetAppDownload() (*types.GetAppDownloadResp, error) {
	cfg, err := loadAppDownloadConfig(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &types.AppDownloadConfig{}
	}
	return &types.GetAppDownloadResp{
		Code:    0,
		Message: "ok",
		Data:    *cfg,
	}, nil
}

// loadAppDownloadConfig 读取 sys_system_config 中 config_group=digital_card_app_download 的配置
//
// 缺失任何字段都允许（前端兜底处理），整体读不到才返回错误（DB 异常）。
func loadAppDownloadConfig(ctx context.Context, svcCtx *svc.ServiceContext) (*types.AppDownloadConfig, error) {
	if svcCtx == nil || svcCtx.DB == nil {
		return nil, nil
	}
	type row struct {
		ConfigKey   string `gorm:"column:config_key"`
		ConfigValue string `gorm:"column:config_value"`
	}
	var rows []row
	if err := svcCtx.DB.WithContext(ctx).
		Table("sys_system_config").
		Select("config_key, config_value").
		Where("config_group = ? AND is_deleted = 0", "digital_card_app_download").
		Find(&rows).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.AppDownloadConfig{}, nil
		}
		return nil, err
	}

	cfg := &types.AppDownloadConfig{}
	for _, r := range rows {
		switch r.ConfigKey {
		case "androidUrl":
			cfg.AndroidUrl = r.ConfigValue
		case "iosUrl":
			cfg.IosUrl = r.ConfigValue
		case "appName":
			cfg.AppName = r.ConfigValue
		case "tagline":
			cfg.Tagline = r.ConfigValue
		}
	}
	return cfg, nil
}
