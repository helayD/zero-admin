package menutemplateservicelogic

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMenuTemplateListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMenuTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuTemplateListLogic {
	return &QueryMenuTemplateListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询菜单模板列表
func (l *QueryMenuTemplateListLogic) QueryMenuTemplateList(in *sysclient.QueryMenuTemplateListReq) (*sysclient.QueryMenuTemplateListResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	queryBuilder := l.svcCtx.DB.WithContext(l.ctx).Table("sys_menu_template")

	if name := strings.TrimSpace(in.Name); name != "" {
		queryBuilder = queryBuilder.Where("name LIKE ?", "%"+name+"%")
	}
	if scopeType := strings.TrimSpace(in.ScopeType); scopeType != "" {
		queryBuilder = queryBuilder.Where("scope_type = ?", scopeType)
	}
	if in.Status != 2 {
		queryBuilder = queryBuilder.Where("status = ?", in.Status)
	}

	var total int64
	if err := queryBuilder.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "统计菜单模板列表失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询菜单模板列表失败")
	}

	rows := make([]menuTemplateListRow, 0)
	err := queryBuilder.
		Select("sys_menu_template.id, sys_menu_template.name, sys_menu_template.scope_type, sys_menu_template.platform_id, sys_menu_template.status, sys_menu_template.remark, sys_menu_template.create_by, sys_menu_template.create_time, sys_menu_template.update_by, sys_menu_template.update_time, COUNT(sys_menu_template_item.id) AS menu_count").
		Joins("LEFT JOIN sys_menu_template_item ON sys_menu_template_item.template_id = sys_menu_template.id").
		Group("sys_menu_template.id, sys_menu_template.name, sys_menu_template.scope_type, sys_menu_template.platform_id, sys_menu_template.status, sys_menu_template.remark, sys_menu_template.create_by, sys_menu_template.create_time, sys_menu_template.update_by, sys_menu_template.update_time").
		Order("sys_menu_template.id ASC").
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Scan(&rows).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询菜单模板列表失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询菜单模板列表失败")
	}

	list := make([]*sysclient.MenuTemplateListData, 0, len(rows))
	for _, row := range rows {
		list = append(list, &sysclient.MenuTemplateListData{
			Id:         row.ID,
			Name:       row.Name,
			ScopeType:  row.ScopeType,
			PlatformId: row.PlatformID,
			Status:     row.Status,
			Remark:     row.Remark,
			CreateBy:   row.CreateBy,
			CreateTime: time_util.TimeToStr(row.CreateTime),
			UpdateBy:   row.UpdateBy,
			UpdateTime: formatNullableTime(row.UpdateTime),
			MenuCount:  int32(row.MenuCount),
		})
	}

	return &sysclient.QueryMenuTemplateListResp{
		Total: total,
		List:  list,
	}, nil
}
