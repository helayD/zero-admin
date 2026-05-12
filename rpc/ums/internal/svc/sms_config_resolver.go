package svc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"

	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/pkg/sms"
	"github.com/feihua/zero-admin/rpc/sys/client/channelintegrationtemplateservice"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

// smsConfigCacheTTL ConfigResolver 正向缓存有效期。
// Story 3.1.1 决策: 30 秒 — 后台切换 provider 模板后最多 30s 生效，
// 同时避免每次发送都打 sys-rpc。
const smsConfigCacheTTL = 30 * time.Second

// smsConfigNegativeCacheTTL ConfigResolver 负向缓存有效期。
// Review M-3: sys-rpc 短暂不可用或模板被误删时，避免在高 QPS 下持续打 sys-rpc。
// 5s 窗口足够让请求"快速失败"且不会让运维切换激活模板的窗口拉得太长。
const smsConfigNegativeCacheTTL = 5 * time.Second

// smsTargetCode 仅消费 sys_channel_integration_template 中
// target_code='sms_provider' 的模板。
const smsTargetCode = channeltemplate.TargetSMSProvider

// SmsConfigResolver 实现 pkg/sms.ConfigResolver。
// 通过 sys-rpc 的 ChannelIntegrationTemplateService 拉取当前激活的
// 短信网关模板（target_code='sms_provider' AND status='enabled'），
// 解析其 default_config_json 字段为 sms.Config。
type SmsConfigResolver struct {
	templateClient channelintegrationtemplateservice.ChannelIntegrationTemplateService

	mu        sync.Mutex
	cached    sms.Config
	cachedAt  time.Time
	cacheHits bool // 是否已经成功缓存过一次（避免初次零值误判）

	// 负向缓存：上一次 Resolve 失败的错误与时间，避免 sys-rpc 抖动时被高 QPS 打穿。
	negErr  error
	negAt   time.Time
	negHits bool
}

// NewSmsConfigResolver 构造一个新的 ConfigResolver。
// templateClient 必须是已连接 sys-rpc 的 client；nil 会导致 Resolve 立即报错。
func NewSmsConfigResolver(templateClient channelintegrationtemplateservice.ChannelIntegrationTemplateService) *SmsConfigResolver {
	return &SmsConfigResolver{templateClient: templateClient}
}

// Resolve 实现 pkg/sms.ConfigResolver.Resolve。
// 当前所有场景都共享同一条 sys_channel_integration_template.target_code='sms_provider'
// 的 enabled 模板；后续若按场景拆分模板可改造此处。
//
// 缓存策略:
//   - 正向缓存 smsConfigCacheTTL（30s）：命中直接返回 cfg
//   - 负向缓存 smsConfigNegativeCacheTTL（5s）：失败也短时缓存以保护 sys-rpc
func (r *SmsConfigResolver) Resolve(ctx context.Context, scene sms.Scene) (sms.Config, error) {
	if r == nil || r.templateClient == nil {
		return sms.Config{}, errors.New("ums-rpc: SmsConfigResolver 未注入 sys-rpc client")
	}

	if cfg, ok := r.readCache(); ok {
		return cfg, nil
	}
	if err, ok := r.readNegativeCache(); ok {
		return sms.Config{}, err
	}

	resp, err := r.templateClient.QueryChannelIntegrationTemplateList(ctx, &sysclient.QueryChannelIntegrationTemplateListReq{
		PageNum:      1,
		PageSize:     1,
		TemplateType: channeltemplate.TemplateTypeIntegration,
		TargetCode:   smsTargetCode,
		Status:       channeltemplate.StatusEnabled,
	})
	if err != nil {
		wrapped := fmt.Errorf("ums-rpc: 拉取 sms provider 模板失败: %w", err)
		r.writeNegativeCache(wrapped)
		return sms.Config{}, wrapped
	}
	if resp == nil || len(resp.List) == 0 {
		wrapped := errors.New("ums-rpc: 未找到 enabled 的 sms_provider 模板，请先在系统管理→短信网关配置中启用一条")
		r.writeNegativeCache(wrapped)
		return sms.Config{}, wrapped
	}

	cfg, err := parseSmsConfig(resp.List[0])
	if err != nil {
		r.writeNegativeCache(err)
		return sms.Config{}, err
	}
	r.writeCache(cfg)
	_ = scene // 当前所有场景共享同一模板，预留参数以便未来按 scene 拆分
	return cfg, nil
}

func (r *SmsConfigResolver) readCache() (sms.Config, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.cacheHits {
		return sms.Config{}, false
	}
	if time.Since(r.cachedAt) > smsConfigCacheTTL {
		return sms.Config{}, false
	}
	return r.cached, true
}

func (r *SmsConfigResolver) writeCache(cfg sms.Config) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cached = cfg
	r.cachedAt = time.Now()
	r.cacheHits = true
	// 写正向缓存时同步清空负向缓存
	r.negErr = nil
	r.negHits = false
}

// readNegativeCache 读取负向缓存。
// 在 smsConfigNegativeCacheTTL 窗口内的失败直接复用上一次错误，避免连续打 sys-rpc。
func (r *SmsConfigResolver) readNegativeCache() (error, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.negHits {
		return nil, false
	}
	if time.Since(r.negAt) > smsConfigNegativeCacheTTL {
		return nil, false
	}
	return r.negErr, true
}

// writeNegativeCache 记录一次失败，给 sys-rpc 喘息窗口。
func (r *SmsConfigResolver) writeNegativeCache(err error) {
	if err == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.negErr = err
	r.negAt = time.Now()
	r.negHits = true
}

// smsConfigPayload sys_channel_integration_template.default_config_json 期望结构
type smsConfigPayload struct {
	ProviderCode   string `json:"providerCode"`
	SignName       string `json:"signName"`
	TemplateCode   string `json:"templateCode"`
	CredentialRef  string `json:"credentialRef"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	ExpireSeconds  int    `json:"expireSeconds"`
}

func parseSmsConfig(template *sysclient.ChannelIntegrationTemplateData) (sms.Config, error) {
	if template == nil {
		return sms.Config{}, errors.New("ums-rpc: sms provider 模板为空")
	}
	raw := strings.TrimSpace(template.MetadataConfig)
	if raw == "" {
		return sms.Config{}, fmt.Errorf("ums-rpc: sms provider 模板[%s] default_config_json 为空", template.TemplateCode)
	}

	var payload smsConfigPayload
	if err := sonic.Unmarshal([]byte(raw), &payload); err != nil {
		return sms.Config{}, fmt.Errorf("ums-rpc: sms provider 模板[%s] default_config_json 解析失败: %w", template.TemplateCode, err)
	}

	providerCode := strings.TrimSpace(payload.ProviderCode)
	if providerCode == "" {
		return sms.Config{}, fmt.Errorf("ums-rpc: sms provider 模板[%s] providerCode 缺失", template.TemplateCode)
	}

	cfg := sms.Config{
		ProviderCode:   providerCode,
		SignName:       strings.TrimSpace(payload.SignName),
		TemplateCode:   strings.TrimSpace(payload.TemplateCode),
		CredentialRef:  strings.TrimSpace(payload.CredentialRef),
		TimeoutSeconds: payload.TimeoutSeconds,
		ExpireSeconds:  payload.ExpireSeconds,
	}
	if cfg.ExpireSeconds <= 0 {
		cfg.ExpireSeconds = 300
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 5
	}
	return cfg, nil
}
