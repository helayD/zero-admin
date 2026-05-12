package sms

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// defaultSender Sender 接口默认实现，根据 ConfigResolver 路由到对应 Provider。
type defaultSender struct {
	resolver ConfigResolver
}

// NewSender 构造默认 Sender 实现。
// resolver 不能为 nil，否则后续 Send 调用会立即返回错误。
func NewSender(resolver ConfigResolver) Sender {
	return &defaultSender{resolver: resolver}
}

// Send 实现 Sender.Send。
//
// 流程:
//  1. 通过 ConfigResolver 拉取当前激活的 provider 配置
//  2. 按 Config.ProviderCode 在 registry 查找已注册 Provider
//  3. 调用 Provider.Send 完成下发
//
// 任意一步失败都返回包装错误，保证调用方能区分配置错误与下发错误。
func (s *defaultSender) Send(ctx context.Context, mobile string, scene Scene) (*SendResult, error) {
	if s.resolver == nil {
		return nil, errors.New("sms: ConfigResolver 未注入")
	}
	if strings.TrimSpace(mobile) == "" {
		return nil, errors.New("sms: 手机号不能为空")
	}

	cfg, err := s.resolver.Resolve(ctx, scene)
	if err != nil {
		return nil, fmt.Errorf("sms: 拉取场景[%s]配置失败: %w", scene, err)
	}
	providerCode := strings.TrimSpace(cfg.ProviderCode)
	if providerCode == "" {
		return nil, fmt.Errorf("sms: 场景[%s]未配置 ProviderCode", scene)
	}

	provider, ok := ProviderOf(providerCode)
	if !ok {
		return nil, fmt.Errorf("sms: provider[%s] 未注册", providerCode)
	}

	result, err := provider.Send(ctx, mobile, scene, cfg)
	if err != nil {
		return nil, fmt.Errorf("sms: provider[%s] 下发失败: %w", providerCode, err)
	}
	if result == nil {
		return nil, fmt.Errorf("sms: provider[%s] 返回空结果", providerCode)
	}
	if result.ProviderCode == "" {
		result.ProviderCode = providerCode
	}
	return result, nil
}
