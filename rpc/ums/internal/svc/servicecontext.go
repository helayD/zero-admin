package svc

import (
	"fmt"
	"time"

	"github.com/feihua/zero-admin/pkg/mq"
	"github.com/feihua/zero-admin/pkg/sms"
	"github.com/feihua/zero-admin/rpc/sys/client/channelintegrationtemplateservice"
	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config                             config.Config
	DB                                 *gorm.DB
	RabbitMQ                           *mq.RabbitMQ
	MemberBrandAttentionModel          model.MemberBrandAttentionModel
	MemberBrowseRecordModel            model.MemberBrowseRecordModel
	MemberProductCategoryRelationModel model.MemberProductCategoryRelationModel
	MemberProductCollectionModel       model.MemberProductCollectionModel
	Redis                              *redis.Redis
	RedisKey                           string // redis的模块统一前缀

	// Story 3.1.1: 短信验证码登录注册合并
	// SmsSender 全项目共享的 SMS 抽象层，业务模块通过其下发短信
	SmsSender sms.Sender
	// ChannelIntegrationTemplateService 用于 SmsSender 的 ConfigResolver 拉取激活模板
	ChannelIntegrationTemplateService channelintegrationtemplateservice.ChannelIntegrationTemplateService
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisKey := c.RpcServerConf.Redis.Key
	rds := redis.MustNewRedis(c.RpcServerConf.Redis.RedisConf)
	db, err := gorm.Open(mysql.Open(c.Mysql.Datasource), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 settingLogConfig(),
	})
	if err != nil {
		logx.Errorf("mysql连接失败：%+v", err)
		panic(err)
	}

	logx.Info("mysql连接成功")
	query.SetDefault(db)

	mqUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/", c.Rabbitmq.UserName, c.Rabbitmq.Password, c.Rabbitmq.Host, c.Rabbitmq.Port)
	rabbitmq := mq.NewRabbitMQSimple(mqUrl)

	_, err = mon.NewModel(c.Mongo.Datasource, c.Mongo.Db, "test")
	if err != nil {
		logx.Errorf("mongo连接失败：%+v", err)
		panic(err)
	}
	logx.Info("mongo连接成功")
	MemberBrandAttention := model.NewMemberBrandAttentionModel(c.Mongo.Datasource, c.Mongo.Db, "ums_member_brand_attention")
	MemberBrowseRecordModel := model.NewMemberBrowseRecordModel(c.Mongo.Datasource, c.Mongo.Db, "ums_member_browse_record")
	MemberProductCategoryRelationModel := model.NewMemberProductCategoryRelationModel(c.Mongo.Datasource, c.Mongo.Db, "ums_member_product_category_relation")
	MemberProductCollection := model.NewMemberProductCollectionModel(c.Mongo.Datasource, c.Mongo.Db, "ums_member_product_collection")

	// Story 3.1.1: 注入 SMS 抽象层
	//   - sys-rpc client → ChannelIntegrationTemplateService 提供激活模板查询
	//   - SmsConfigResolver 解析模板 default_config_json → sms.Config（30s 正向 / 5s 负向缓存）
	//   - sms.NewSender 路由到对应 Provider（默认注册的 mock 验证码固定 123456）
	//
	// ⚠️ 部署依赖（Story 3.1.1 新增）:
	//   ums-rpc 启动前 sys-rpc 必须已注册到 Etcd/Nacos，否则 zrpc.MustNewClient
	//   会 panic 导致 ums 启动失败。建议启动顺序:
	//     1) sys-rpc 先就绪
	//     2) ums-rpc、pms-rpc、oms-rpc、... 其他 RPC 服务
	//     3) admin-api、front-api 等 HTTP API 层
	//   服务器部署脚本 script/ / Makefile 已按此顺序编排。
	sysRpcClient := zrpc.MustNewClient(c.SysRpc)
	templateService := channelintegrationtemplateservice.NewChannelIntegrationTemplateService(sysRpcClient)
	smsResolver := NewSmsConfigResolver(templateService)
	smsSender := sms.NewSender(smsResolver)

	return &ServiceContext{
		Config:                             c,
		DB:                                 db,
		RabbitMQ:                           rabbitmq,
		MemberBrandAttentionModel:          MemberBrandAttention,
		MemberBrowseRecordModel:            MemberBrowseRecordModel,
		MemberProductCategoryRelationModel: MemberProductCategoryRelationModel,
		MemberProductCollectionModel:       MemberProductCollection,
		Redis:                              rds,
		RedisKey:                           redisKey,

		SmsSender:                         smsSender,
		ChannelIntegrationTemplateService: templateService,
	}
}

type Writer struct {
}

func (w Writer) Printf(format string, args ...interface{}) {
	logx.Infof(format, args...)
}

// init log config
func settingLogConfig() logger.Interface {
	newLogger := logger.New(
		Writer{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Info,            // Log level
			IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                   // Disable color
		},
	)
	return newLogger
}
