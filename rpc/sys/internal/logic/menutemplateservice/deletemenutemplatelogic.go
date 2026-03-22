package menutemplateservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMenuTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMenuTemplateLogic {
	return &DeleteMenuTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除菜单模板
func (l *DeleteMenuTemplateLogic) DeleteMenuTemplate(in *sysclient.DeleteMenuTemplateReq) (*sysclient.DeleteMenuTemplateResp, error) {
	templateIDs := normalizeMenuIDs(in.Ids)
	if len(templateIDs) == 0 {
		return nil, errors.New("请选择要删除的菜单模板")
	}

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("sys_menu_template_item").Where("template_id IN ?", templateIDs).Delete(&menuTemplateItemRow{}).Error; err != nil {
			logc.Errorf(l.ctx, "删除菜单模板关联失败, 参数:%+v, 异常:%s", in, err.Error())
			return errors.New("删除菜单模板失败")
		}

		result := tx.Table("sys_menu_template").Where("id IN ?", templateIDs).Delete(&menuTemplateRow{})
		if result.Error != nil {
			logc.Errorf(l.ctx, "删除菜单模板失败, 参数:%+v, 异常:%s", in, result.Error.Error())
			return errors.New("删除菜单模板失败")
		}
		if result.RowsAffected == 0 {
			return errors.New("菜单模板不存在")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &sysclient.DeleteMenuTemplateResp{Pong: "ok"}, nil
}
