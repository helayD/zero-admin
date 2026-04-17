package drawactivityservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateDrawActivityStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDrawActivityStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDrawActivityStatusLogic {
	return &UpdateDrawActivityStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDrawActivityStatusLogic) UpdateDrawActivityStatus(in *smsclient.UpdateDrawActivityStatusReq) (*smsclient.UpdateDrawActivityStatusResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureDrawActivityScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "sms.draw_activity.status", in.UpdateBy, in.OperatorName, "update draw activity status"); err != nil {
		return nil, err
	}

	now := time.Now()
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range in.Ids {
			aggregate, err := loadDrawActivityAggregate(l.ctx, tx, currentScope, id)
			if err != nil {
				return err
			}
			switch in.Status {
			case drawStatusPublished:
				if !aggregate.Readiness.ReadyToPublish {
					if err := syncDrawReadiness(l.ctx, tx, id, aggregate.Readiness); err != nil {
						return err
					}
					return errors.New(strings.TrimSpace(aggregate.Readiness.Summary))
				}
			case drawStatusDraft:
				aggregate.Activity.PublishReadiness = drawReadinessOffline
				aggregate.Activity.PublishFailureSummary = ""
			}

			updates := map[string]interface{}{
				"status":      in.Status,
				"update_by":   in.UpdateBy,
				"update_time": now,
			}
			if in.Status == drawStatusPublished {
				updates["publish_readiness"] = drawReadinessReady
				updates["publish_failure_summary"] = ""
				aggregate.Activity.PublishReadiness = drawReadinessReady
				aggregate.Activity.PublishFailureSummary = ""
			}
			if in.Status == drawStatusDraft {
				updates["publish_readiness"] = drawReadinessOffline
				updates["publish_failure_summary"] = ""
			}
			if err := tx.WithContext(l.ctx).Table(drawActivityRow{}.TableName()).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
			aggregate.Activity.Status = in.Status
			if err := appendDrawActivityAudit(l.ctx, tx, &aggregate.Activity, aggregate.Templates, aggregate.Pools, in.UpdateBy, in.OperatorName, statusActionName(in.Status)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新抽卡活动状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}
	return &smsclient.UpdateDrawActivityStatusResp{}, nil
}

func statusActionName(status int32) string {
	switch status {
	case drawStatusPublished:
		return "published"
	case drawStatusArchived:
		return "archived"
	default:
		return "offline"
	}
}
