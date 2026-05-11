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
		Enabled         bool
		Host            string
		Port            int
		DisableSsl      bool
		IsSMCrypto      bool
		GroupID         string
		ChainID         string
		ContractAddr    string
		ContractABI     string
		ContractABIPath string
		PrivateKey      string
		CaCertPath      string
		SdkCertPath     string
		SdkKeyPath      string
		TimeoutSeconds  int64
		PollIntervalMs  int64
		PollMaxAttempts int
	}
}
