package svc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/antchain"
	"github.com/feihua/zero-admin/pkg/chainclient"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/fisco"
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
	ChainClient     chainclient.ChainClient
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
	chainClient := buildChainClient(c)
	cardMintService := digitalcardmint.NewService(DB, rabbitmq, chainClient)

	// Story 10.7 运维修复：sms-rpc 启动后立即做一次 ScanDueTasks，并每 30 秒自动扫描一次。
	// 避免外部 job/cron 缺失时积压任务永远卡在 pending_dispatch/compensating。
	go startMintTaskSelfScan(cardMintService)

	return &ServiceContext{
		Config:          c,
		DB:              DB,
		RabbitMQ:        rabbitmq,
		ChainClient:     chainClient,
		CardMintService: cardMintService,
	}
}

// startMintTaskSelfScan 周期性扫描到期的提货卡发放任务。
// 30 秒跑一次，batchSize=50，异常只记日志不 panic。
func startMintTaskSelfScan(service *digitalcardmint.Service) {
	if service == nil {
		return
	}
	// 启动后等 5 秒让 gRPC server / MQ channel ready，再做第一次扫描
	time.Sleep(5 * time.Second)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		stats, err := service.ScanDueTasks(context.Background(), 50)
		if err != nil {
			logx.Errorf("sms-rpc 自循环扫描提货卡发放任务失败: %+v", err)
		} else if stats != nil && (stats.Dispatched+stats.Executed+stats.Escalated) > 0 {
			logx.Infof("sms-rpc 自循环扫描提货卡任务: dispatched=%d executed=%d escalated=%d", stats.Dispatched, stats.Executed, stats.Escalated)
		}
		<-ticker.C
	}
}

func buildChainClient(c config.Config) chainclient.ChainClient {
	client := buildChainClientImpl(c)
	// Story 10.11 / Task 6.2 / AC6 + M4：启动日志打印 chainType + mode + 节点 + 合约地址，
	// mode 区分 disabled / invalid_config / real，避免「chainType=fisco_bcos_3x」误导运维
	// 以为已接真链。
	mode := "unknown"
	if m, ok := client.(interface{ Mode() fisco.Mode }); ok {
		mode = string(m.Mode())
	}
	logx.Infof("buildChainClient[sms-rpc]: chainType=%q mode=%s nodeAddr=%s:%d contract=%s",
		client.ChainType(), mode, c.Fisco.Host, c.Fisco.Port, c.Fisco.ContractAddr)

	// Story 10.11 / M5：启动后异步做一次 HealthCheck，仅预警，不阻塞服务启动。
	if h, ok := client.(interface{ HealthCheck(context.Context) error }); ok && mode == string(fisco.ModeReal) {
		go func() {
			time.Sleep(5 * time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := h.HealthCheck(ctx); err != nil {
				logx.Errorf("sms-rpc FISCO 启动探活失败: %v", err)
			} else {
				logx.Infof("sms-rpc FISCO 启动探活 OK")
			}
		}()
	}
	return client
}

func buildChainClientImpl(c config.Config) chainclient.ChainClient {
	logx.Infof("buildChainClient: Blockchain.Primary=%q Fisco.Enabled=%v AntChain.Enabled=%v Fisco.Host=%q:%d",
		c.Blockchain.Primary, c.Fisco.Enabled, c.AntChain.Enabled, c.Fisco.Host, c.Fisco.Port)
	switch strings.ToLower(strings.TrimSpace(c.Blockchain.Primary)) {
	case "antchain":
		return antchain.NewClient(antchain.Config{
			Endpoint:        c.AntChain.Endpoint,
			ReceiptEndpoint: c.AntChain.ReceiptEndpoint,
			AppID:           c.AntChain.AppId,
			AccessKey:       c.AntChain.AccessKey,
			Secret:          c.AntChain.Secret,
			TimeoutSeconds:  c.AntChain.TimeoutSeconds,
			Enabled:         c.AntChain.Enabled,
		})
	case "", "fisco", "fisco_bcos_3x":
		return fisco.NewClient(fisco.Config{
			Enabled:         c.Fisco.Enabled,
			Host:            c.Fisco.Host,
			Port:            c.Fisco.Port,
			DisableSsl:      c.Fisco.DisableSsl,
			IsSMCrypto:      c.Fisco.IsSMCrypto,
			GroupID:         c.Fisco.GroupID,
			ChainID:         c.Fisco.ChainID,
			ContractAddr:    c.Fisco.ContractAddr,
			ContractABI:     c.Fisco.ContractABI,
			ContractABIPath: c.Fisco.ContractABIPath,
			PrivateKey:      c.Fisco.PrivateKey,
			CaCertPath:      c.Fisco.CaCertPath,
			SdkCertPath:     c.Fisco.SdkCertPath,
			SdkKeyPath:      c.Fisco.SdkKeyPath,
			TimeoutSeconds:  c.Fisco.TimeoutSeconds,
		})
	default:
		return invalidChainClient{primary: c.Blockchain.Primary}
	}
}

type invalidChainClient struct {
	primary string
}

func (c invalidChainClient) MintToken(context.Context, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, c.err()
}

func (c invalidChainClient) QueryMintToken(context.Context, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, c.err()
}

func (c invalidChainClient) ChainType() string {
	return "invalid"
}

func (c invalidChainClient) err() error {
	return errors.New("提货卡通道配置非法: " + strings.TrimSpace(c.primary))
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
