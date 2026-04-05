package channelintegrationtemplateservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChannelIntegrationTemplateStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChannelIntegrationTemplateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChannelIntegrationTemplateStatusLogic {
	return &UpdateChannelIntegrationTemplateStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateChannelIntegrationTemplateStatusLogic) UpdateChannelIntegrationTemplateStatus(in *sysclient.UpdateChannelIntegrationTemplateStatusReq) (*sysclient.UpdateChannelIntegrationTemplateStatusResp, error) {
	if len(in.Ids) == 0 {
		return nil, errors.New("模板ID不能为空")
	}
	nextStatus, err := channeltemplate.NormalizeStatus(in.Status)
	if err != nil {
		return nil, err
	}
	ids := normalizeIDList(in.Ids)

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		rows := make([]channelIntegrationTemplateRow, 0, len(ids))
		if err := tx.Where("id IN ?", ids).Find(&rows).Error; err != nil {
			logc.Errorf(l.ctx, "查询待流转模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新模板状态失败")
		}
		if len(rows) != len(ids) {
			return errors.New("存在不存在的模板")
		}
		for _, row := range rows {
			if err := channeltemplate.ValidateStatusTransition(row.Status, nextStatus); err != nil {
				return err
			}
		}
		updateBy := strings.TrimSpace(in.UpdateBy)
		if updateBy == "" {
			updateBy = "system"
		}
		updates := map[string]interface{}{
			"status":      nextStatus,
			"update_by":   updateBy,
			"update_time": time.Now(),
		}
		if err := tx.Model(&channelIntegrationTemplateRow{}).Where("id IN ?", ids).Updates(updates).Error; err != nil {
			logc.Errorf(l.ctx, "批量更新模板状态失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新模板状态失败")
		}
		updateTime := updates["update_time"].(time.Time)
		for _, row := range rows {
			row.Status = nextStatus
			row.UpdateBy = updateBy
			row.UpdateTime = &updateTime
			if err := syncTemplateBindings(l.ctx, tx, row, updateBy); err != nil {
				return err
			}
			action := "status_change"
			switch nextStatus {
			case channeltemplate.StatusEnabled:
				action = "enable"
			case channeltemplate.StatusDisabled:
				action = "disable"
			case channeltemplate.StatusArchived:
				action = "archive"
			}
			if err := recordTemplateOperateLog(l.ctx, tx, row, action, updateBy, fmt.Sprintf("template=%s,status=%s", row.TemplateCode, row.Status)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.UpdateChannelIntegrationTemplateStatusResp{Pong: "ok"}, nil
}
