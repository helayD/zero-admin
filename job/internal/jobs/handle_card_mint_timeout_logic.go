package jobs

import (
	"context"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logc"
)

func HandleCardMintTimeout(ctx context.Context, service *digitalcardmint.Service) {
	if service == nil {
		logc.Errorf(ctx, "数字卡片发放服务未初始化")
		return
	}

	stats, err := service.ScanDueTasks(ctx, 50)
	if err != nil {
		logc.Errorf(ctx, "扫描数字卡片发放补偿任务失败, err=%v", err)
		return
	}

	logc.Infof(ctx, "数字卡片发放补偿扫描完成, dispatched=%d, executed=%d, escalated=%d", stats.Dispatched, stats.Executed, stats.Escalated)
}
