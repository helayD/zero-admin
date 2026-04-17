package drawactivityservicelogic

import (
	"context"
	"errors"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteDrawActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDrawActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDrawActivityLogic {
	return &DeleteDrawActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDrawActivityLogic) DeleteDrawActivity(in *smsclient.DeleteDrawActivityReq) (*smsclient.DeleteDrawActivityResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureDrawActivityScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "sms.draw_activity.delete", in.UpdateBy, in.OperatorName, "delete draw activity"); err != nil {
		return nil, err
	}

	now := time.Now()
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(l.ctx).Table(drawActivityRow{}.TableName()).
			Where("id IN ? AND is_deleted = 0", in.Ids).
			Updates(map[string]interface{}{
				"is_deleted":  1,
				"status":      drawStatusArchived,
				"update_by":   in.UpdateBy,
				"update_time": now,
			}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(l.ctx).Table(drawPoolRow{}.TableName()).
			Where("activity_id IN ? AND is_deleted = 0", in.Ids).
			Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": now}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(l.ctx).Table(drawPoolTemplateRow{}.TableName()).
			Where("activity_id IN ? AND is_deleted = 0", in.Ids).
			Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": now}).Error; err != nil {
			return err
		}
		for _, id := range in.Ids {
			var row drawActivityRow
			if err := tx.WithContext(l.ctx).Table(drawActivityRow{}.TableName()).Where("id = ?", id).Take(&row).Error; err != nil {
				return err
			}
			row.Status = drawStatusArchived
			row.PublishReadiness = drawReadinessOffline
			row.PublishFailureSummary = "活动已删除"
			if err := appendDrawActivityAudit(l.ctx, tx, &row, nil, nil, in.UpdateBy, in.OperatorName, "deleted"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("抽卡活动不存在")
		}
		logc.Errorf(l.ctx, "删除抽卡活动失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}
	return &smsclient.DeleteDrawActivityResp{}, nil
}
