package svc

import (
	"context"
	"fmt"

	"time"

	"github.com/feihua/zero-admin/consumer/internal/config"
	"github.com/feihua/zero-admin/consumer/internal/mq/coupon"
	digitalcardconsumer "github.com/feihua/zero-admin/consumer/internal/mq/digital_card"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/consumer/internal/mq/order"
	"github.com/feihua/zero-admin/consumer/internal/mq/product"
	"github.com/feihua/zero-admin/pkg/antchain"
	"github.com/feihua/zero-admin/pkg/chainclient"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/fisco"
	"github.com/feihua/zero-admin/pkg/mq"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/pms/client/productskuservice"
	"github.com/feihua/zero-admin/rpc/pms/client/productspuservice"
	"github.com/feihua/zero-admin/rpc/search/search_client"
	"github.com/feihua/zero-admin/rpc/sms/client/couponrecordservice"
	"github.com/feihua/zero-admin/rpc/sms/client/couponservice"
	"github.com/feihua/zero-admin/rpc/sms/client/coupontypeservice"
	"github.com/feihua/zero-admin/rpc/ums/client/membergrowthlogservice"
	"github.com/feihua/zero-admin/rpc/ums/client/memberinfoservice"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/feihua/zero-admin/rpc/ums/client/memberpointslogservice"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config          config.Config
	RabbitMQ        *mq.RabbitMQ
	Redis           *redis.Redis
	DB              *gorm.DB
	ChainClient     chainclient.ChainClient
	CardMintService *digitalcardmint.Service

	// 会员相关
	MemberInfoService      memberinfoservice.MemberInfoService
	MemberGrowthLogService membergrowthlogservice.MemberGrowthLogService
	MemberPointsLogService memberpointslogservice.MemberPointsLogService
	MemberMessageService   membermessageservice.MemberMessageService

	// 营销相关
	CouponRecordService couponrecordservice.CouponRecordService
	CouponService       couponservice.CouponService
	CouponTypeService   coupontypeservice.CouponTypeService

	// 商品相关
	ProductSkuService productskuservice.ProductSkuService
	ProductSpuService productspuservice.ProductSpuService
	// 订单相关
	OrderService orderservice.OrderService

	// 搜索相关
	Search search_client.Search
}

func NewServiceContext(c config.Config) *ServiceContext {
	umsClient := zrpc.MustNewClient(c.UmsRpc)
	smsClient := zrpc.MustNewClient(c.SmsRpc)
	omsClient := zrpc.MustNewClient(c.OmsRpc)
	pmsClient := zrpc.MustNewClient(c.PmsRpc)
	searchClient := zrpc.MustNewClient(c.SearchRpc)

	mqUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/", c.Rabbitmq.UserName, c.Rabbitmq.Password, c.Rabbitmq.Host, c.Rabbitmq.Port)
	rabbitmq := mq.NewRabbitMQSimple(mqUrl)

	var db *gorm.DB
	if c.Mysql.Datasource != "" {
		var err error
		db, err = gorm.Open(mysql.Open(c.Mysql.Datasource), &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			Logger:                 consumerLogConfig(),
		})
		if err != nil {
			panic(err)
		}
	}
	chainClient := buildChainClient(c)
	cardMintService := digitalcardmint.NewService(db, rabbitmq, chainClient)

	redisConf := redis.RedisConf{
		Host: c.Redis.Address,
		Type: "node",
		Pass: c.Redis.Pass,
		Tls:  false,
	}
	r := redis.MustNewRedis(redisConf)

	memberInfoService := memberinfoservice.NewMemberInfoService(umsClient)
	memberMessageService := membermessageservice.NewMemberMessageService(umsClient)
	couponService := couponservice.NewCouponService(smsClient)
	couponRecordService := couponrecordservice.NewCouponRecordService(smsClient)
	skuService := productskuservice.NewProductSkuService(pmsClient)
	spuService := productspuservice.NewProductSpuService(pmsClient)
	orderService := orderservice.NewOrderService(omsClient)
	search := search_client.NewSearch(searchClient)
	s := &ServiceContext{
		Config:                 c,
		RabbitMQ:               rabbitmq,
		Redis:                  r,
		DB:                     db,
		ChainClient:            chainClient,
		CardMintService:        cardMintService,
		MemberInfoService:      memberInfoService,
		MemberGrowthLogService: membergrowthlogservice.NewMemberGrowthLogService(umsClient),
		MemberPointsLogService: memberpointslogservice.NewMemberPointsLogService(umsClient),
		MemberMessageService:   memberMessageService,
		CouponRecordService:    couponRecordService,
		CouponService:          couponService,
		CouponTypeService:      coupontypeservice.NewCouponTypeService(smsClient),
		ProductSkuService:      skuService,
		ProductSpuService:      spuService,
		OrderService:           orderService,
		Search:                 search,
	}

	go func() {
		rabbitmq.ConsumeSimple("test", func(body []byte) {
			logc.Infof(context.Background(), "收到消息: %s", body)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimple("first.login.queue", func(body []byte) {
			coupon.FirstLogin(context.Background(), body, memberInfoService, couponService, couponRecordService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("coupon.issued.queue", func(body []byte) error {
			return coupon.CouponIssued(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.delay.cancel.queue", func(body []byte) error {
			return order.OrderDelayCancel(context.Background(), body, r, skuService, orderService, couponRecordService, memberInfoService)
		})
	}()

	go func() {
		rabbitmq.ConsumeTopicQueue("pms.product.sync.queue", "product.event.exchange", "pms.product.*.key", func(body []byte) {
			product.SynProductToEs(context.Background(), body, search, spuService)
		})
	}()
	go func() {
		rabbitmq.ConsumeTopicQueue("pms.product.delete.queue", "product.event.exchange", "pms.product.*.key", func(body []byte) {
			product.DeleteProductFromEs(context.Background(), body, search, spuService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.return.queue", func(body []byte) error {
			return order.OrderReturn(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.cancel.queue", func(body []byte) error {
			return order.OrderCancel(context.Background(), body)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.close.queue", func(body []byte) error {
			return order.OrderClose(context.Background(), body)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.delivery.queue", func(body []byte) error {
			return order.OrderDelivery(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.confirm.queue", func(body []byte) error {
			return order.OrderConfirm(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.pay.queue", func(body []byte) error {
			return order.OrderPay(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.create.queue", func(body []byte) error {
			return order.OrderCreate(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("member.message.queue", func(body []byte) error {
			return member.CreateMemberMessage(context.Background(), body, memberMessageService)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck(digitalcardmint.EventQueue, func(body []byte) error {
			return digitalcardconsumer.MintRequested(context.Background(), body, cardMintService)
		})
	}()

	return s
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

type consumerWriter struct{}

func (consumerWriter) Printf(format string, args ...interface{}) {
	logc.Infof(context.Background(), format, args...)
}

func consumerLogConfig() logger.Interface {
	return logger.New(
		consumerWriter{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}
