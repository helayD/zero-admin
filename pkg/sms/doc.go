// Package sms 提供全项目共享的短信发送抽象层。
//
// 设计目标:
//   - 业务模块（会员登录、订单通知、密码找回等）统一通过 Sender 接口下发短信，
//     禁止直接 import 第三方厂商 SDK。
//   - 真实供应商通过实现 Provider 接口并调用 RegisterProvider 自动接入，
//     业务侧无需感知具体厂商差异。
//   - 当前内置 MockProvider，验证码固定为 "123456"，仅用于开发/测试环境调试。
//   - 生产环境的 Provider 与配置由 sys_channel_integration_template 表
//     （target_code='sms_provider'）控制，通过 ConfigResolver 在运行时拉取激活模板。
//
// 典型用法:
//
//	resolver := myConfigResolver{}
//	sender := sms.NewSender(resolver)
//	result, err := sender.Send(ctx, "13800138001", sms.SceneMemberLogin)
//	if err != nil { ... }
//	// result.Code 是实际下发的验证码，业务侧负责写入 Redis 等待校验。
package sms
