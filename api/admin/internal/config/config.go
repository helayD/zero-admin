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

	// 内容
	CmsRpc zrpc.RpcClientConf

	Auth struct {
		AccessSecret string
		AccessExpire int64
		ExcludeUrl   string
	}

	Redis struct {
		Address string
		Pass    string
	}

	Mysql struct {
		Datasource string
	}

	Rabbitmq struct {
		Host     string
		Port     int64
		UserName string
		Password string
	}
	Swagger struct {
		IsTest bool
		Path   string
	}

	SystemConfig SystemConfig `json:",optional"`
}

type SystemConfig struct {
	OSS  OSSConfig
	SMS  SMSConfig
	Push PushConfig
}

type OSSConfig struct {
	Endpoint        string `json:",env=OSS_ENDPOINT,default=oss-cn-shenzhen.aliyuncs.com"`
	AccessKeyID     string `json:",optional,env=OSS_ACCESS_KEY_ID"`
	AccessKeySecret string `json:",optional,env=OSS_ACCESS_KEY_SECRET"`
	BucketName      string `json:",env=OSS_BUCKET_NAME,default=mbjq"`
	URL             string `json:",env=OSS_URL,default=https://speed.maibanjk.com/"`
	MaxSizeMB       int64  `json:",env=OSS_MAX_SIZE_MB,default=20"`
}

type SMSConfig struct {
	Enabled         bool   `json:",env=SMS_ENABLED,default=false"`
	Provider        string `json:",optional,env=SMS_PROVIDER"`
	Endpoint        string `json:",optional,env=SMS_ENDPOINT"`
	AccessKeyID     string `json:",optional,env=SMS_ACCESS_KEY_ID"`
	AccessKeySecret string `json:",optional,env=SMS_ACCESS_KEY_SECRET"`
	SignName        string `json:",optional,env=SMS_SIGN_NAME"`
	TemplateCode    string `json:",optional,env=SMS_TEMPLATE_CODE"`
}

type PushConfig struct {
	Enabled   bool   `json:",env=PUSH_ENABLED,default=false"`
	Provider  string `json:",optional,env=PUSH_PROVIDER"`
	Endpoint  string `json:",optional,env=PUSH_ENDPOINT"`
	AppKey    string `json:",optional,env=PUSH_APP_KEY"`
	AppSecret string `json:",optional,env=PUSH_APP_SECRET"`
}
