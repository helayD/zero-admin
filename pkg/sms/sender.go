package sms

import "context"

// Sender 业务层使用的短信下发统一入口。
// 业务模块只依赖此 interface，无需感知具体 provider 与凭据细节。
type Sender interface {
	// Send 在指定场景下向手机号发送验证码短信。
	// 返回的 SendResult.Code 为实际下发的验证码原文，业务侧负责写入 Redis；
	// 业务接口响应体禁止包含验证码原文（即使 mock 模式）。
	Send(ctx context.Context, mobile string, scene Scene) (*SendResult, error)
}

// ConfigResolver 由调用方注入，负责从配置真相源（sys_channel_integration_template）
// 拉取当前激活的 provider 配置。
//
// 调用方应该在实现内部加适当的内存缓存（建议 30 秒），避免每次发送都打 sys-rpc。
type ConfigResolver interface {
	Resolve(ctx context.Context, scene Scene) (Config, error)
}
