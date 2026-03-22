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

type AddMenuTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMenuTemplateLogic {
	return &AddMenuTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加菜单模板
func (l *AddMenuTemplateLogic) AddMenuTemplate(in *sysclient.AddMenuTemplateReq) (*sysclient.AddMenuTemplateResp, error) {
	name, err := normalizeTemplateName(in.Name)
	if err != nil {
		return nil, err
	}
	scopeType, err := normalizeTemplateScope(in.ScopeType)
	if err != nil {
		return nil, err
	}
	menuIDs := normalizeMenuIDs(in.MenuIds)
	template := &menuTemplateRow{
		Name:       name,
		ScopeType:  scopeType,
		PlatformID: normalizePlatformID(in.PlatformId),
		Status:     in.Status,
		Remark:     strings.TrimSpace(in.Remark),
		CreateBy:   strings.TrimSpace(in.CreateBy),
		CreateTime: time.Now(),
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("sys_menu_template").Create(template).Error; err != nil {
			logc.Errorf(l.ctx, "创建菜单模板失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("创建菜单模板失败")
		}

		if err := replaceTemplateMenuIDs(l.ctx, tx, template.ID, menuIDs); err != nil {
			logc.Errorf(l.ctx, "保存菜单模板菜单关联失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("保存菜单模板菜单关联失败")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.AddMenuTemplateResp{Pong: "ok"}, nil
}
