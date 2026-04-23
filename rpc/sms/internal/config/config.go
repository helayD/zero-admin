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
}
