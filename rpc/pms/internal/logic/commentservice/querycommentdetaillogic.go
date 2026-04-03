package commentservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCommentDetailLogic 查询商品评价详情
/*
Author: LiuFeiHua
Date: 2025/01/24 09:08:05
*/
type QueryCommentDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCommentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCommentDetailLogic {
	return &QueryCommentDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCommentDetail 查询商品评价详情
func (l *QueryCommentDetailLogic) QueryCommentDetail(in *pmsclient.QueryCommentDetailReq) (*pmsclient.QueryCommentDetailResp, error) {
	item, err := loadScopedComment(l.ctx, l.svcCtx, in.Id, in.PlatformId, in.TenantId, in.MerchantId)
	if err != nil {
		logc.Errorf(l.ctx, "查询商品评价详情失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品评价详情失败")
	}

	return &pmsclient.QueryCommentDetailResp{
		Id:               item.ID.Hex(),
		ProductId:        item.ProductId,
		MemberNickName:   item.MemberNickName,
		ProductName:      item.ProductName,
		Star:             item.Star,
		MemberIp:         item.MemberIp,
		CreateTime:       formatMongoTime(item.CreateAt),
		ShowStatus:       item.ShowStatus,
		ProductAttribute: item.ProductAttribute,
		CollectCount:     item.CollectCount,
		ReadCount:        item.ReadCount,
		Content:          item.Content,
		Pics:             item.Pics,
		MemberIcon:       item.MemberIcon,
		ReplayCount:      item.ReplayCount,
		MemberId:         item.MemberId,
		PlatformId:       item.PlatformId,
		TenantId:         item.TenantId,
		MerchantId:       item.MerchantId,
		AuditStatus:      item.AuditStatus,
		AuditRemark:      item.AuditRemark,
		AuditorId:        item.AuditorId,
		AuditorName:      item.AuditorName,
		AuditedAt:        formatMongoTime(item.AuditedAt),
		Hidden:           item.Hidden,
		AppealStatus:     item.AppealStatus,
		AppealReason:     item.AppealReason,
		AppealReply:      item.AppealReply,
		AppealedAt:       formatMongoTime(item.AppealedAt),
		AppealHandledAt:  formatMongoTime(item.AppealHandledAt),
	}, nil
}
