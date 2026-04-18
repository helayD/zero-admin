package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Mysql struct {
		Datasource string
	}

	AntChain struct {
		Endpoint       string
		AppId          string
		AccessKey      string
		Secret         string
		TimeoutSeconds int64
		Enabled        bool
	}

	UmsRpc zrpc.RpcClientConf
	PmsRpc zrpc.RpcClientConf
	OmsRpc zrpc.RpcClientConf
	SmsRpc zrpc.RpcClientConf
	Redis  struct {
		Address string
		Pass    string
	}
}
