package svc

import (
	"github.com/feihua/zero-admin/job/internal/config"
	"github.com/feihua/zero-admin/pkg/antchain"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
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
	"time"
)

type ServiceContext struct {
	Config          config.Config
	Redis           *redis.Redis
	DB              *gorm.DB
	AntChain        antchain.Client
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
	antChainClient := antchain.NewClient(antchain.Config{
		Endpoint:       c.AntChain.Endpoint,
		AppID:          c.AntChain.AppId,
		AccessKey:      c.AntChain.AccessKey,
		Secret:         c.AntChain.Secret,
		TimeoutSeconds: c.AntChain.TimeoutSeconds,
		Enabled:        c.AntChain.Enabled,
	})
	cardMintService := digitalcardmint.NewService(db, nil, antChainClient)
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
		AntChain:            antChainClient,
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
