package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		Datasource string
	}

	Mongo struct {
		Datasource string
		Db         string
	}
	Rabbitmq struct {
		Host     string
		Port     int64
		UserName string
		Password string
	}
	JWT struct {
		AccessSecret string
		AccessExpire int64
	}

	// SysRpc 用于 sms ConfigResolver 调用 sys-rpc.ChannelIntegrationTemplateService
	// 拉取激活的 SMS provider 模板。Story 3.1.1 引入。
	SysRpc zrpc.RpcClientConf
}
