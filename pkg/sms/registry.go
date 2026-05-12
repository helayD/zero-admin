package sms

import "sync"

// registry 进程级 Provider 注册中心。
// 由 Provider 实现的 init() 自动调用 RegisterProvider 注册自身。
var (
	registryMu sync.RWMutex
	registry   = map[string]Provider{}
)

// RegisterProvider 注册一个 Provider 到全局 registry。
// 重复注册同名 provider 将覆盖旧值（用于测试时替换 mock 实现）。
func RegisterProvider(p Provider) {
	if p == nil {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.Code()] = p
}

// ProviderOf 按 code 查找已注册 Provider。
func ProviderOf(code string) (Provider, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[code]
	return p, ok
}

// resetRegistryForTest 仅用于测试场景清空 registry，业务代码不应调用。
func resetRegistryForTest() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = map[string]Provider{}
}
