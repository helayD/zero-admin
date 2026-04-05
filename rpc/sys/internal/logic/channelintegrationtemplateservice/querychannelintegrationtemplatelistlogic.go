package channelintegrationtemplateservicelogic

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChannelIntegrationTemplateListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryChannelIntegrationTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChannelIntegrationTemplateListLogic {
	return &QueryChannelIntegrationTemplateListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *QueryChannelIntegrationTemplateListLogic) QueryChannelIntegrationTemplateList(in *sysclient.QueryChannelIntegrationTemplateListReq) (*sysclient.QueryChannelIntegrationTemplateListResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	statusFilter, err := normalizeStatusFilter(in.Status)
	if err != nil {
		return nil, err
	}

	queryBuilder := l.svcCtx.DB.WithContext(l.ctx).Model(&channelIntegrationTemplateRow{})
	if templateName := strings.TrimSpace(in.TemplateName); templateName != "" {
		queryBuilder = queryBuilder.Where("template_name LIKE ?", "%"+templateName+"%")
	}
	if templateType := strings.TrimSpace(in.TemplateType); templateType != "" {
		normalizedTemplateType, err := channeltemplate.NormalizeTemplateType(templateType)
		if err != nil {
			return nil, err
		}
		queryBuilder = queryBuilder.Where("template_type = ?", normalizedTemplateType)
	}
	if targetCode := normalizeTargetCodeFilter(in.TargetCode); targetCode != "" {
		queryBuilder = queryBuilder.Where("target_code = ?", targetCode)
	}
	if statusFilter != "" {
		queryBuilder = queryBuilder.Where("status = ?", statusFilter)
	}

	var total int64
	if err := queryBuilder.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "统计模板列表失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询模板列表失败")
	}

	rows := make([]channelIntegrationTemplateRow, 0)
	if err := queryBuilder.Order("id ASC").Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).Find(&rows).Error; err != nil {
		logc.Errorf(l.ctx, "查询模板列表失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询模板列表失败")
	}

	list := make([]*sysclient.ChannelIntegrationTemplateData, 0, len(rows))
	for _, row := range rows {
		item := mapTemplateRowToProto(row)
		summary, err := summarizeTemplateBindings(l.ctx, l.svcCtx.DB, row)
		if err != nil {
			logc.Errorf(l.ctx, "汇总模板影响范围失败, 模板ID:%d, 异常:%s", row.ID, err.Error())
			return nil, errors.New("查询模板列表失败")
		}
		item.ImpactScopeConfig = mergeImpactScopeConfigWithSummary(item.ImpactScopeConfig, summary)
		list = append(list, item)
	}
	return &sysclient.QueryChannelIntegrationTemplateListResp{Total: total, List: list}, nil
}
