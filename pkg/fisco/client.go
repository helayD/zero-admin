package fisco

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

var _ chainclient.ChainClient = (*fiscoClient)(nil)
var _ chainclient.ChainClient = (disabledClient{})

type fiscoClient struct {
	cfg     Config
	timeout time.Duration
}

type disabledClient struct{}

func (disabledClient) MintToken(context.Context, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("FISCO BCOS 能力未启用")
}

func (disabledClient) QueryMintToken(context.Context, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("FISCO BCOS 能力未启用")
}

func (disabledClient) ChainType() string { return "fisco" }

func NewClient(cfg Config) chainclient.ChainClient {
	if !cfg.Enabled {
		return disabledClient{}
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &fiscoClient{
		cfg:     cfg,
		timeout: timeout,
	}
}

func (c *fiscoClient) ChainType() string { return "fisco" }

// TODO: 实现 FISCO BCOS Go SDK 调用合约 mint 方法
func (c *fiscoClient) MintToken(_ context.Context, req *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("mint request 不能为空")
	}
	return nil, errors.New("FISCO BCOS MintToken 尚未实现")
}

// TODO: 实现 FISCO BCOS Go SDK 查询链上交易回执
func (c *fiscoClient) QueryMintToken(_ context.Context, req *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("query request 不能为空")
	}
	return nil, errors.New("FISCO BCOS QueryMintToken 尚未实现")
}
