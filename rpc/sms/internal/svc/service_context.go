package svc

import (
	"fmt"
	"time"

	"github.com/feihua/zero-admin/pkg/antchain"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/mq"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config          config.Config
	DB              *gorm.DB
	RabbitMQ        *mq.RabbitMQ
	AntChain        antchain.Client
	CardMintService *digitalcardmint.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	DB, err := gorm.Open(mysql.Open(c.Mysql.Datasource), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 settingLogConfig(),
	})
	if err != nil {
		panic(err)
	}

	logx.Debug("mysql已连接")
	query.SetDefault(DB)

	var rabbitmq *mq.RabbitMQ
	if c.Rabbitmq.Host != "" {
		mqUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/", c.Rabbitmq.UserName, c.Rabbitmq.Password, c.Rabbitmq.Host, c.Rabbitmq.Port)
		rabbitmq = mq.NewRabbitMQSimple(mqUrl)
	}
	antChainClient := antchain.NewClient(antchain.Config{
		Endpoint:        c.AntChain.Endpoint,
		ReceiptEndpoint: c.AntChain.ReceiptEndpoint,
		AppID:           c.AntChain.AppId,
		AccessKey:       c.AntChain.AccessKey,
		Secret:          c.AntChain.Secret,
		TimeoutSeconds:  c.AntChain.TimeoutSeconds,
		Enabled:         c.AntChain.Enabled,
	})
	cardMintService := digitalcardmint.NewService(DB, rabbitmq, antChainClient)

	return &ServiceContext{
		Config:          c,
		DB:              DB,
		RabbitMQ:        rabbitmq,
		AntChain:        antChainClient,
		CardMintService: cardMintService,
	}
}

type Writer struct{}

func (w Writer) Printf(format string, args ...interface{}) {
	logx.Infof(format, args...)
}

func settingLogConfig() logger.Interface {
	newLogger := logger.New(
		Writer{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	return newLogger
}
