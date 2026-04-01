package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	UmsRpc zrpc.RpcClientConf
	PmsRpc zrpc.RpcClientConf
	OmsRpc zrpc.RpcClientConf
	SmsRpc zrpc.RpcClientConf
	Redis  struct {
		Address string
		Pass    string
	}
}
