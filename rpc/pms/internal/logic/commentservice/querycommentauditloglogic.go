package commentservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryCommentAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCommentAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCommentAuditLogLogic {
	return &QueryCommentAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCommentAuditLog 查询评价审核日志
func (l *QueryCommentAuditLogLogic) QueryCommentAuditLog(in *pmsclient.QueryCommentAuditLogReq) (*pmsclient.QueryCommentAuditLogResp, error) {
	if _, err := loadScopedComment(l.ctx, l.svcCtx, in.CommentId, in.PlatformId, in.TenantId, in.MerchantId); err != nil {
		return nil, err
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProductCommentAuditLog{}).Where("comment_id = ?", in.CommentId)
	if in.PlatformId > 0 {
		query = query.Where("platform_id = ?", in.PlatformId)
	}
	if in.TenantId > 0 {
		query = query.Where("tenant_id = ?", in.TenantId)
	}
	if in.MerchantId > 0 {
		query = query.Where("merchant_id = ?", in.MerchantId)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "统计评价审核日志失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询评价审核日志失败")
	}

	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	var logs []model.ProductCommentAuditLog
	if err := query.Order("id desc").Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).Find(&logs).Error; err != nil {
		logc.Errorf(l.ctx, "查询评价审核日志列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询评价审核日志失败")
	}

	list := make([]*pmsclient.CommentAuditLogData, 0, len(logs))
	for _, item := range logs {
		list = append(list, &pmsclient.CommentAuditLogData{
			Id:           item.ID,
			CommentId:    item.CommentID,
			Action:       item.Action,
			FromStatus:   item.FromStatus,
			ToStatus:     item.ToStatus,
			OperatorId:   item.OperatorID,
			OperatorName: item.OperatorName,
			Remark:       item.Remark,
			CreatedAt:    time_util.TimeToStr(item.CreatedAt),
		})
	}

	return &pmsclient.QueryCommentAuditLogResp{Total: total, List: list}, nil
}
