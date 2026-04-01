package svc

import (
	"github.com/feihua/zero-admin/job/internal/config"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/client/ordersettingservice"
	"github.com/feihua/zero-admin/rpc/pms/client/productskuservice"
	"github.com/feihua/zero-admin/rpc/sms/client/couponrecordservice"
	"github.com/feihua/zero-admin/rpc/ums/client/memberinfoservice"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis

	UmsRpc              zrpc.RpcClientConf
	PmsRpc              zrpc.RpcClientConf
	OmsRpc              zrpc.RpcClientConf
	SmsRpc              zrpc.RpcClientConf
	MemberService       memberinfoservice.MemberInfoService
	ProductSkuService   productskuservice.ProductSkuService
	OrderService        orderservice.OrderService
	OrderSettingService  ordersettingservice.OrderSettingService
	CouponRecordService couponrecordservice.CouponRecordService
}

func NewServiceContext(c config.Config) *ServiceContext {
	umsClient := zrpc.MustNewClient(c.UmsRpc)
	smsClient := zrpc.MustNewClient(c.SmsRpc)
	omsClient := zrpc.MustNewClient(c.OmsRpc)
	pmsClient := zrpc.MustNewClient(c.PmsRpc)

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
