package sms

// Scene 短信发送场景枚举。
// 新增场景时在此追加常量，禁止业务模块绕过本包自定义场景标识。
type Scene string

const (
	// SceneMemberLogin 会员登录/注册合并场景（手机号+验证码）
	SceneMemberLogin Scene = "member_login"

	// 预留扩展位（暂未使用，新增前需走架构评审）
	// SceneOrderShipped Scene = "order_shipped"
	// SceneSecurityAlert Scene = "security_alert"
	// SceneRealName Scene = "real_name"
)

// Code 返回场景的字符串表示。
func (s Scene) Code() string { return string(s) }

// SendResult Sender 返回值。
//
// Code 字段为实际下发的验证码原文，业务侧负责写入 Redis 等待校验；
// 接口响应体绝不能直接回传 Code 给客户端。
type SendResult struct {
	// Code 实际下发的验证码（mock 固定 123456，aliyun/tencent 用 crypto/rand 生成 6 位数字）
	Code string
	// ExpireSeconds 验证码有效期，业务侧写 Redis 时使用此 TTL
	ExpireSeconds int
	// ProviderCode 实际使用的 provider 代码（用于审计/排障）
	ProviderCode string
}

// Config 由 ConfigResolver 提供的运行时配置，与 sys_channel_integration_template
// 表的 default_config_json 字段一一对应。
type Config struct {
	// ProviderCode 当前激活的 provider 代码（mock / aliyun / tencent ...）
	ProviderCode string
	// SignName 短信签名（阿里云等需要）
	SignName string
	// TemplateCode 短信模板代码（阿里云等需要）
	TemplateCode string
	// CredentialRef 凭据引用，格式形如 credential://xxx 或 ref:env:ALIYUN_AK，
	// 解析责任由 Provider 自身承担；mock 不需要凭据。
	CredentialRef string
	// TimeoutSeconds 单次下发超时时间（秒）
	TimeoutSeconds int
	// ExpireSeconds 验证码有效期（秒），默认 300
	ExpireSeconds int
}
