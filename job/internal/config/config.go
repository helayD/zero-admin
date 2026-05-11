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
