package sms

import "context"

// Provider 真实短信供应商实现入口。
//
// 每个真实网关（mock / aliyun / tencent ...）实现一个 Provider，
// 通过 RegisterProvider 注册到全局 registry 后由 DefaultSender 路由。
//
// 实现约束:
//   - 验证码生成在 Provider 内部完成（mock 固定 123456，真实 provider 用
//     crypto/rand 生成 6 位数字）。
//   - Provider 必须自行解析 cfg.CredentialRef 拉取凭据，禁止业务层下传明文 AK/SK。
//   - 生产 provider 实现禁止打印验证码原文到日志。
type Provider interface {
	// Code 返回 provider 唯一标识，用于 ConfigResolver 路由匹配。
	Code() string
	// Send 在给定场景下向手机号发送短信。
	Send(ctx context.Context, mobile string, scene Scene, cfg Config) (*SendResult, error)
}
