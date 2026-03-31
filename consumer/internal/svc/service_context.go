package svc

import (
	"context"
	"fmt"
	"github.com/feihua/zero-admin/consumer/internal/config"
	"github.com/feihua/zero-admin/consumer/internal/mq/coupon"
	"github.com/feihua/zero-admin/consumer/internal/mq/order"
	"github.com/feihua/zero-admin/consumer/internal/mq/product"
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
	"github.com/feihua/zero-admin/rpc/ums/client/memberpointslogservice"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config   config.Config
	RabbitMQ *mq.RabbitMQ
	Redis    *redis.Redis

	// 会员相关
	MemberInfoService      memberinfoservice.MemberInfoService
	MemberGrowthLogService membergrowthlogservice.MemberGrowthLogService
	MemberPointsLogService memberpointslogservice.MemberPointsLogService

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

	redisConf := redis.RedisConf{
		Host: c.Redis.Address,
		Type: "node",
		Pass: c.Redis.Pass,
		Tls:  false,
	}
	r := redis.MustNewRedis(redisConf)

	memberInfoService := memberinfoservice.NewMemberInfoService(umsClient)
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
		MemberInfoService:      memberInfoService,
		MemberGrowthLogService: membergrowthlogservice.NewMemberGrowthLogService(umsClient),
		MemberPointsLogService: memberpointslogservice.NewMemberPointsLogService(umsClient),
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
			return order.OrderReturn(context.Background(), body)
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
			return order.OrderDelivery(context.Background(), body)
		})
	}()

	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.confirm.queue", func(body []byte) error {
			return order.OrderConfirm(context.Background(), body)
		})
	}()
	go func() {
		rabbitmq.ConsumeSimpleWithAck("order.create.queue", func(body []byte) error {
			return order.OrderCreate(context.Background(), body)
		})
	}()
	return s
}
