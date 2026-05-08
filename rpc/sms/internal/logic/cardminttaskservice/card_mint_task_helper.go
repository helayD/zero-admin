package cardminttaskservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

func EnsureCardMintTaskByAssetInstance(ctx context.Context, svcCtx *svc.ServiceContext, tx *gorm.DB, assetInstanceID int64, operatorType string) (int64, error) {
	if svcCtx == nil || svcCtx.CardMintService == nil {
		return 0, nil
	}
	task, err := svcCtx.CardMintService.EnsureTaskTx(ctx, tx, assetInstanceID, operatorType)
	if err != nil {
		return 0, err
	}
	if task == nil {
		return 0, nil
	}
	return task.ID, nil
}

func DispatchCardMintTask(ctx context.Context, svcCtx *svc.ServiceContext, taskID int64, reason string) {
	if taskID <= 0 || svcCtx == nil || svcCtx.CardMintService == nil {
		return
	}
	if err := svcCtx.CardMintService.DispatchTask(ctx, taskID, reason); err != nil {
		logc.Errorf(ctx, "派发提货卡发放任务失败, taskId=%d, reason=%s, err=%s", taskID, reason, err.Error())
	}
}
