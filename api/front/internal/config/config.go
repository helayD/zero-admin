package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// 系统
	SysRpc zrpc.RpcClientConf
	// 会员
	UmsRpc zrpc.RpcClientConf
	// 商品
	PmsRpc zrpc.RpcClientConf
	// 订单
	OmsRpc zrpc.RpcClientConf
	// 营销
	SmsRpc zrpc.RpcClientConf
	// 支付
	PayRpc zrpc.RpcClientConf
	// 内容相关
	CmsRpc zrpc.RpcClientConf
	// 搜索
	SearchRpc zrpc.RpcClientConf

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	// 支付宝支付配置
	Alipay struct {
		AppId        string
		PrivateKey   string
		ServerDomain string
		NotifyURL    string
		IsProduction bool
	}

	Rabbitmq struct {
		Host     string
		Port     int64
		UserName string
		Password string
	}
	Redis struct {
		Address string
		Pass    string
	}

	UpgradePolicy struct {
		CurrentVersion       string
		MinSupportedVersion  string
		RecommendedVersion   string
		EffectiveAt          string
		DeadlineAt           string
		ReleaseNotesSummary  string
		UpgradeUrl           string
		StoreTarget          string
		RecoveryHint         string
		AffectedCapabilities []string
	}

	Swagger struct {
		IsTest bool
		Path   string
	}
}
