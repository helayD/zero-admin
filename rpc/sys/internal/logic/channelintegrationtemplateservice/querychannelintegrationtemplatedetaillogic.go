package channelintegrationtemplateservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChannelIntegrationTemplateDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryChannelIntegrationTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChannelIntegrationTemplateDetailLogic {
	return &QueryChannelIntegrationTemplateDetailLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *QueryChannelIntegrationTemplateDetailLogic) QueryChannelIntegrationTemplateDetail(in *sysclient.QueryChannelIntegrationTemplateDetailReq) (*sysclient.QueryChannelIntegrationTemplateDetailResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID不能为空")
	}

	var row channelIntegrationTemplateRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.Id).Take(&row).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.New("模板不存在")
		default:
			logc.Errorf(l.ctx, "查询模板详情失败, 参数:%+v, 异常:%s", in, err.Error())
			return nil, errors.New("查询模板详情失败")
		}
	}
	summary, err := summarizeTemplateBindings(l.ctx, l.svcCtx.DB, row)
	if err != nil {
		logc.Errorf(l.ctx, "汇总模板详情影响范围失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询模板详情失败")
	}
	row.ImpactScopeConfig = mergeImpactScopeConfigWithSummary(row.ImpactScopeConfig, summary)

	return &sysclient.QueryChannelIntegrationTemplateDetailResp{Data: mapTemplateRowToProto(row)}, nil
}
