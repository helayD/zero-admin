package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Rabbitmq struct {
		Host     string
		Port     int64
		UserName string
		Password string
	}

	Mysql struct {
		Datasource string
	}

	AntChain struct {
		Endpoint        string
		ReceiptEndpoint string
		AppId           string
		AccessKey       string
		Secret          string
		TimeoutSeconds  int64
		Enabled         bool
	}

	Blockchain struct {
		Primary string // "fisco" | "antchain"
	}

	Fisco struct {
		NodeAddr       string
		GroupID        int
		ChainID        int64
		ContractAddr   string
		PrivateKey     string
		TimeoutSeconds int64
		Enabled        bool
	}

	// 会员
	UmsRpc zrpc.RpcClientConf
	// 商品
	PmsRpc zrpc.RpcClientConf
	// 订单
	OmsRpc zrpc.RpcClientConf
	// 营销
	SmsRpc zrpc.RpcClientConf

	// 搜索
	SearchRpc zrpc.RpcClientConf

	Redis struct {
		Address string
		Pass    string
	}

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
}
