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
	Redis struct {
		Host string
		Type string
		Pass string
	}
	Cart struct {
		Timeout int
	}
}
