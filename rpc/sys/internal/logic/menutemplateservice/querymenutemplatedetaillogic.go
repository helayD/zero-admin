package menutemplateservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMenuTemplateDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMenuTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuTemplateDetailLogic {
	return &QueryMenuTemplateDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询菜单模板详情
func (l *QueryMenuTemplateDetailLogic) QueryMenuTemplateDetail(in *sysclient.QueryMenuTemplateDetailReq) (*sysclient.QueryMenuTemplateDetailResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("菜单模板ID不能为空")
	}

	var template menuTemplateRow
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_menu_template").
		Where("id = ?", in.Id).
		Take(&template).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errors.New("菜单模板不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询菜单模板详情失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询菜单模板详情失败")
	}

	menuIDs, err := loadTemplateMenuIDs(l.ctx, l.svcCtx.DB, template.ID)
	if err != nil {
		logc.Errorf(l.ctx, "查询菜单模板详情菜单关联失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询菜单模板详情失败")
	}

	return &sysclient.QueryMenuTemplateDetailResp{
		Id:         template.ID,
		Name:       template.Name,
		ScopeType:  template.ScopeType,
		PlatformId: template.PlatformID,
		Status:     template.Status,
		Remark:     template.Remark,
		CreateBy:   template.CreateBy,
		CreateTime: time_util.TimeToStr(template.CreateTime),
		UpdateBy:   template.UpdateBy,
		UpdateTime: formatNullableTime(template.UpdateTime),
		MenuIds:    menuIDs,
	}, nil
}
