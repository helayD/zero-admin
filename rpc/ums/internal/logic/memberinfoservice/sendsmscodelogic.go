package memberinfoservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/pkg/sms"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// SendSmsCodeLogic 发送短信验证码（Story 3.1.1）。
//
// 流程:
//  1. 校验手机号非空（front-api 已做正则校验，本层做兜底）
//  2. 通过 Redis SETNX 检查并加 60s 冷却窗口（ums:sms:cooldown:{mobile}）
//  3. 调用 svcCtx.SmsSender.Send 完成验证码下发
//     - mock provider 固定 123456（前期调试），aliyun/tencent 走 crypto/rand 6 位
//  4. 把验证码 + 错误计数（"{code}:0"）写入 Redis（ums:sms:login:{mobile}）
//  5. 返回 expireSeconds，**响应体绝不包含验证码原文**
type SendSmsCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsCodeLogic {
	return &SendSmsCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendSmsCode 实现 MemberAuthService.SendSmsCode
func (l *SendSmsCodeLogic) SendSmsCode(in *umsclient.SendSmsCodeReq) (*umsclient.SendSmsCodeResp, error) {
	mobile := strings.TrimSpace(in.Mobile)
	if mobile == "" {
		return nil, errors.New("手机号不能为空")
	}

	scene := mapSmsScene(in.Scene)

	// 频控：60 秒冷却窗口
	cooldownKey := smsCooldownKey(mobile)
	if err := l.acquireCooldown(cooldownKey); err != nil {
		return nil, err
	}

	// 委托 pkg/sms.Sender 完成下发（mock 固定 123456，真实 provider 走 crypto/rand）
	result, err := l.svcCtx.SmsSender.Send(l.ctx, mobile, scene)
	if err != nil {
		// Review M-4: 网关下发失败时清掉冷却 key，避免用户因瞬时网关错误被锁 60s。
		// 验证码本身没写入 Redis，重置冷却不会带来安全风险。
		l.releaseCooldown(cooldownKey)
		logc.Errorf(l.ctx, "下发短信验证码失败,手机号:%s,异常:%s", mobile, err.Error())
		return nil, errors.New("验证码下发失败，请稍后重试")
	}
	if result == nil || strings.TrimSpace(result.Code) == "" {
		l.releaseCooldown(cooldownKey)
		logc.Errorf(l.ctx, "下发短信验证码返回空,手机号:%s", mobile)
		return nil, errors.New("验证码下发失败，请稍后重试")
	}

	expireSeconds := result.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = 300
	}

	// 写验证码 + 错误计数到 Redis（"{code}:0"，TTL = expireSeconds）
	codeKey := smsLoginCodeKey(mobile)
	if err := l.svcCtx.Redis.SetexCtx(l.ctx, codeKey, formatMemberSequence(result.Code, 0), expireSeconds); err != nil {
		// 写 Redis 失败：清掉冷却 key，让用户能立即重新触发下发。
		l.releaseCooldown(cooldownKey)
		logc.Errorf(l.ctx, "写验证码到 Redis 失败,手机号:%s,异常:%s", mobile, err.Error())
		return nil, errors.New("验证码下发失败，请稍后重试")
	}

	return &umsclient.SendSmsCodeResp{
		ExpireSeconds: int32(expireSeconds),
	}, nil
}

// acquireCooldown 通过 Redis SETNX 抢占冷却窗口。
// 抢占失败说明 60 秒内已发送过，直接返回频控错误。
func (l *SendSmsCodeLogic) acquireCooldown(key string) error {
	ok, err := l.svcCtx.Redis.SetnxExCtx(l.ctx, key, "1", smsCooldownSeconds)
	if err != nil {
		logc.Errorf(l.ctx, "校验短信冷却窗口失败,key:%s,异常:%s", key, err.Error())
		return errors.New("验证码下发失败，请稍后重试")
	}
	if !ok {
		return errors.New("验证码请求过于频繁，请稍后再试")
	}
	return nil
}

// releaseCooldown 清掉冷却 key，使用户可立即重试。
// 仅在网关下发失败 / 写 Redis 失败时调用，AC-2 正常成功路径下不调用。
func (l *SendSmsCodeLogic) releaseCooldown(key string) {
	if _, err := l.svcCtx.Redis.DelCtx(l.ctx, key); err != nil {
		logc.Errorf(l.ctx, "清理短信冷却 key 失败,key:%s,异常:%s", key, err.Error())
	}
}

// mapSmsScene 把 RPC 入参的 scene 整数映射为 pkg/sms.Scene。
// 当前仅支持 1 (member_login)，未来可拓展。
func mapSmsScene(raw int32) sms.Scene {
	if raw == 0 {
		return sms.SceneMemberLogin
	}
	switch raw {
	case 1:
		return sms.SceneMemberLogin
	default:
		// 当前仅支持 member_login；未识别的场景默认走登录场景以保持兼容
		return sms.SceneMemberLogin
	}
}

// smsCooldownKey 同手机号 60 秒冷却 Redis key
func smsCooldownKey(mobile string) string {
	return fmt.Sprintf("ums:sms:cooldown:%s", mobile)
}

// smsLoginCodeKey 验证码（含错误计数）Redis key
func smsLoginCodeKey(mobile string) string {
	return fmt.Sprintf("ums:sms:login:%s", mobile)
}

const (
	smsCooldownSeconds = 60
	smsMaxErrAttempts  = 5
)
