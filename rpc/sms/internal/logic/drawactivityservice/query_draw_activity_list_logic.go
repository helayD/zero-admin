package drawactivityservicelogic

import (
	"context"
	"errors"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDrawActivityListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDrawActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDrawActivityListLogic {
	return &QueryDrawActivityListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryDrawActivityListLogic) QueryDrawActivityList(in *smsclient.QueryDrawActivityListReq) (*smsclient.QueryDrawActivityListResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	q := pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Table(drawActivityRow{}.TableName()).Where("is_deleted = 0"),
		currentScope,
		"",
	)
	if name := strings.TrimSpace(in.Name); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if in.Status >= 0 {
		q = q.Where("status = ?", in.Status)
	}
	if in.AuditStatus >= 0 {
		q = q.Where("audit_status = ?", in.AuditStatus)
	}
	if in.PublishReadiness >= 0 {
		q = q.Where("publish_readiness = ?", in.PublishReadiness)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "统计抽卡活动列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询抽卡活动列表失败")
	}

	rows := make([]drawActivityRow, 0)
	if total > 0 {
		if err := q.Order("id desc").Offset(int((in.PageNum - 1) * in.PageSize)).Limit(int(in.PageSize)).Find(&rows).Error; err != nil {
			logc.Errorf(l.ctx, "查询抽卡活动列表失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("查询抽卡活动列表失败")
		}
	}

	list := make([]*smsclient.DrawActivityListData, 0, len(rows))
	for _, row := range rows {
		list = append(list, &smsclient.DrawActivityListData{
			Id:                    row.ID,
			ActivityCode:          row.ActivityCode,
			Name:                  row.Name,
			StartTime:             time_util.TimeToStr(row.StartTime),
			EndTime:               time_util.TimeToStr(row.EndTime),
			Status:                row.Status,
			AuditStatus:           row.AuditStatus,
			IsEnabled:             row.IsEnabled,
			PublishReadiness:      row.PublishReadiness,
			PublishFailureSummary: row.PublishFailureSummary,
			ScopeType:             activityScopeType(row.PlatformID, row.TenantID, row.MerchantID),
			PlatformId:            row.PlatformID,
			TenantId:              row.TenantID,
			MerchantId:            row.MerchantID,
			UpdateTime:            time_util.TimeToString(row.UpdateTime),
			ShowOnHome:            row.ShowOnHome,
			HomeEntryTitle:        row.HomeEntryTitle,
			ReadinessLabel:        drawReadinessLabel(row.PublishReadiness),
		})
	}
	return &smsclient.QueryDrawActivityListResp{Total: total, List: list}, nil
}
