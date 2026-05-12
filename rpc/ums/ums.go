package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/feihua/zero-admin/rpc/ums/internal/config"
	umsserver "github.com/feihua/zero-admin/rpc/ums/internal/server"
	memberaddressserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberaddressservice"
	memberbrandattentionserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberbrandattentionservice"
	memberconsumesettingserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberconsumesettingservice"
	membergrowthlogserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membergrowthlogservice"
	memberidentityserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberidentityservice"
	memberinfoserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberinfoservice"
	memberlevelserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberlevelservice"
	memberloginlogserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberloginlogservice"
	membermessageserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membermessageservice"
	memberpointslogserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberpointslogservice"
	memberproductcategoryrelationserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberproductcategoryrelationservice"
	memberproductcollectionserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberproductcollectionservice"
	memberreadhistoryserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberreadhistoryservice"
	memberrulesettingserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberrulesettingservice"
	membersignlogserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membersignlogservice"
	memberstatisticsinfoserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberstatisticsinfoservice"
	membertagrelationserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membertagrelationservice"
	membertagserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membertagservice"
	membertaskrelationserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membertaskrelationservice"
	membertaskserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/membertaskservice"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "rpc/ums/etc/ums.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.Log)                     // 设置日志配置
	logx.AddWriter(logx.NewWriter(os.Stdout)) // 添加控制台输出
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		umsclient.RegisterMemberInfoServiceServer(grpcServer, memberinfoserviceServer.NewMemberInfoServiceServer(ctx))
		umsclient.RegisterMemberIdentityServiceServer(grpcServer, memberidentityserviceServer.NewMemberIdentityServiceServer(ctx))
		umsclient.RegisterMemberLevelServiceServer(grpcServer, memberlevelserviceServer.NewMemberLevelServiceServer(ctx))
		umsclient.RegisterMemberLoginLogServiceServer(grpcServer, memberloginlogserviceServer.NewMemberLoginLogServiceServer(ctx))
		umsclient.RegisterMemberTagRelationServiceServer(grpcServer, membertagrelationserviceServer.NewMemberTagRelationServiceServer(ctx))
		umsclient.RegisterMemberTaskRelationServiceServer(grpcServer, membertaskrelationserviceServer.NewMemberTaskRelationServiceServer(ctx))
		umsclient.RegisterMemberProductCategoryRelationServiceServer(grpcServer, memberproductcategoryrelationserviceServer.NewMemberProductCategoryRelationServiceServer(ctx))
		umsclient.RegisterMemberAddressServiceServer(grpcServer, memberaddressserviceServer.NewMemberAddressServiceServer(ctx))
		umsclient.RegisterMemberRuleSettingServiceServer(grpcServer, memberrulesettingserviceServer.NewMemberRuleSettingServiceServer(ctx))
		umsclient.RegisterMemberStatisticsInfoServiceServer(grpcServer, memberstatisticsinfoserviceServer.NewMemberStatisticsInfoServiceServer(ctx))
		umsclient.RegisterMemberTaskServiceServer(grpcServer, membertaskserviceServer.NewMemberTaskServiceServer(ctx))
		umsclient.RegisterMemberTagServiceServer(grpcServer, membertagserviceServer.NewMemberTagServiceServer(ctx))
		umsclient.RegisterMemberProductCollectionServiceServer(grpcServer, memberproductcollectionserviceServer.NewMemberProductCollectionServiceServer(ctx))
		umsclient.RegisterMemberReadHistoryServiceServer(grpcServer, memberreadhistoryserviceServer.NewMemberReadHistoryServiceServer(ctx))
		umsclient.RegisterMemberBrandAttentionServiceServer(grpcServer, memberbrandattentionserviceServer.NewMemberBrandAttentionServiceServer(ctx))
		umsclient.RegisterMemberSignLogServiceServer(grpcServer, membersignlogserviceServer.NewMemberSignLogServiceServer(ctx))
		umsclient.RegisterMemberGrowthLogServiceServer(grpcServer, membergrowthlogserviceServer.NewMemberGrowthLogServiceServer(ctx))
		umsclient.RegisterMemberPointsLogServiceServer(grpcServer, memberpointslogserviceServer.NewMemberPointsLogServiceServer(ctx))
		umsclient.RegisterMemberConsumeSettingServiceServer(grpcServer, memberconsumesettingserviceServer.NewMemberConsumeSettingServiceServer(ctx))
		umsclient.RegisterMemberMessageServiceServer(grpcServer, membermessageserviceServer.NewMemberMessageServiceServer(ctx))

		// Story 3.1.1: 注册手写的 MemberAuthService（手机号+验证码合并登录注册）
		umsserver.RegisterExtraServices(grpcServer, ctx)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
