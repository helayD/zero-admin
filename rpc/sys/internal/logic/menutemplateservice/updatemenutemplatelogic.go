package menutemplateservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMenuTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuTemplateLogic {
	return &UpdateMenuTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新菜单模板
func (l *UpdateMenuTemplateLogic) UpdateMenuTemplate(in *sysclient.UpdateMenuTemplateReq) (*sysclient.UpdateMenuTemplateResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("菜单模板ID不能为空")
	}
	name, err := normalizeTemplateName(in.Name)
	if err != nil {
		return nil, err
	}
	scopeType, err := normalizeTemplateScope(in.ScopeType)
	if err != nil {
		return nil, err
	}
	menuIDs := normalizeMenuIDs(in.MenuIds)
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var existing menuTemplateRow
		if err := tx.Table("sys_menu_template").Where("id = ?", in.Id).Take(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("菜单模板不存在")
			}
			logc.Errorf(l.ctx, "查询待更新菜单模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新菜单模板失败")
		}

		updates := map[string]interface{}{
			"name":        name,
			"scope_type":  scopeType,
			"platform_id": normalizePlatformID(in.PlatformId),
			"status":      in.Status,
			"remark":      strings.TrimSpace(in.Remark),
			"update_by":   strings.TrimSpace(in.UpdateBy),
			"update_time": time.Now(),
		}
		if err := tx.Table("sys_menu_template").Where("id = ?", in.Id).Updates(updates).Error; err != nil {
			logc.Errorf(l.ctx, "更新菜单模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新菜单模板失败")
		}

		if err := replaceTemplateMenuIDs(l.ctx, tx, in.Id, menuIDs); err != nil {
			logc.Errorf(l.ctx, "更新菜单模板菜单关联失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("更新菜单模板失败")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.UpdateMenuTemplateResp{Pong: "ok"}, nil
}
