package svc

import (
	"time"

	"github.com/feihua/zero-admin/job/internal/config"
	"github.com/feihua/zero-admin/pkg/antchain"
	"github.com/feihua/zero-admin/pkg/chainclient"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/fisco"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/client/ordersettingservice"
	"github.com/feihua/zero-admin/rpc/pms/client/productskuservice"
	"github.com/feihua/zero-admin/rpc/sms/client/couponrecordservice"
	"github.com/feihua/zero-admin/rpc/ums/client/memberinfoservice"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config          config.Config
	Redis           *redis.Redis
	DB              *gorm.DB
	ChainClient     chainclient.ChainClient
	CardMintService *digitalcardmint.Service

	UmsRpc              zrpc.RpcClientConf
	PmsRpc              zrpc.RpcClientConf
	OmsRpc              zrpc.RpcClientConf
	SmsRpc              zrpc.RpcClientConf
	MemberService       memberinfoservice.MemberInfoService
	ProductSkuService   productskuservice.ProductSkuService
	OrderService        orderservice.OrderService
	OrderSettingService ordersettingservice.OrderSettingService
	CouponRecordService couponrecordservice.CouponRecordService
}

func NewServiceContext(c config.Config) *ServiceContext {
	umsClient := zrpc.MustNewClient(c.UmsRpc)
	smsClient := zrpc.MustNewClient(c.SmsRpc)
	omsClient := zrpc.MustNewClient(c.OmsRpc)
	pmsClient := zrpc.MustNewClient(c.PmsRpc)

	var db *gorm.DB
	if c.Mysql.Datasource != "" {
		var err error
		db, err = gorm.Open(mysql.Open(c.Mysql.Datasource), &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			Logger:                 logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			panic(err)
		}
	}
	chainClient := buildChainClient(c)
	cardMintService := digitalcardmint.NewService(db, nil, chainClient)
	cardMintService.RunningTimeout = 2 * time.Minute

	redisConf := redis.RedisConf{
		Host: c.Redis.Address,
		Type: "node",
		Pass: c.Redis.Pass,
		Tls:  false,
	}
	r := redis.MustNewRedis(redisConf)

	memberService := memberinfoservice.NewMemberInfoService(umsClient)
	skuService := productskuservice.NewProductSkuService(pmsClient)
	orderService := orderservice.NewOrderService(omsClient)
	couponRecordService := couponrecordservice.NewCouponRecordService(smsClient)
	orderSettingService := ordersettingservice.NewOrderSettingService(omsClient)

	return &ServiceContext{
		Config:              c,
		Redis:               r,
		DB:                  db,
		ChainClient:         chainClient,
		CardMintService:     cardMintService,
		UmsRpc:              c.UmsRpc,
		PmsRpc:              c.PmsRpc,
		OmsRpc:              c.OmsRpc,
		SmsRpc:              c.SmsRpc,
		MemberService:       memberService,
		ProductSkuService:   skuService,
		OrderService:        orderService,
		CouponRecordService: couponRecordService,
		OrderSettingService: orderSettingService,
	}
}

func buildChainClient(c config.Config) chainclient.ChainClient {
	if c.Blockchain.Primary == "antchain" {
		return antchain.NewClient(antchain.Config{
			Endpoint:        c.AntChain.Endpoint,
			ReceiptEndpoint: c.AntChain.ReceiptEndpoint,
			AppID:           c.AntChain.AppId,
			AccessKey:       c.AntChain.AccessKey,
			Secret:          c.AntChain.Secret,
			TimeoutSeconds:  c.AntChain.TimeoutSeconds,
			Enabled:         c.AntChain.Enabled,
		})
	}
	return fisco.NewClient(fisco.Config{
		NodeAddr:       c.Fisco.NodeAddr,
		GroupID:        c.Fisco.GroupID,
		ChainID:        c.Fisco.ChainID,
		ContractAddr:   c.Fisco.ContractAddr,
		PrivateKey:     c.Fisco.PrivateKey,
		TimeoutSeconds: c.Fisco.TimeoutSeconds,
		Enabled:        c.Fisco.Enabled,
	})
}
