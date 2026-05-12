package sms

import (
	"context"

	"github.com/zeromicro/go-zero/core/logc"
)

// MockProviderCode mock provider 唯一标识
const MockProviderCode = "mock"

// MockFixedCode mock provider 固定下发的验证码（前期调试用）。
// 长度固定 6 位以保持与生产 provider（aliyun/tencent）一致，
// 让前端校验 ^\d{6}$ 在 mock/真实环境共用同一条规则。
// 安全约束: 此 provider 仅供开发/测试环境激活，生产环境运维必须切换到真实 provider。
const MockFixedCode = "123456"

const defaultExpireSeconds = 300

// MockProvider mock 短信网关实现，验证码固定为 123456。
//
// 行为:
//   - 不调用任何外部网关，不引入任何第三方 SDK
//   - 通过 logc.Infof 打印 mobile + scene + 固定验证码（仅开发环境可见）
//   - 永远返回成功
type MockProvider struct{}

// Code 实现 Provider.Code
func (MockProvider) Code() string { return MockProviderCode }

// Send 实现 Provider.Send。验证码固定 123456，仅写日志，不发起任何网络调用。
func (MockProvider) Send(ctx context.Context, mobile string, scene Scene, cfg Config) (*SendResult, error) {
	logc.Infof(ctx, "[MOCK-SMS] mobile=%s, scene=%s, code=%s", mobile, scene, MockFixedCode)

	expireSeconds := cfg.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = defaultExpireSeconds
	}

	return &SendResult{
		Code:          MockFixedCode,
		ExpireSeconds: expireSeconds,
		ProviderCode:  MockProviderCode,
	}, nil
}

func init() {
	RegisterProvider(MockProvider{})
}
