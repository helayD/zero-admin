package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		Datasource string
	}

	Rabbitmq struct {
		Host     string
		Port     int64
		UserName string
		Password string
	}

	AntChain struct {
		Endpoint       string
		AppId          string
		AccessKey      string
		Secret         string
		TimeoutSeconds int64
		Enabled        bool
	}
}
