package commentservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCommentListLogic 查询商品评价列表
/*
Author: LiuFeiHua
Date: 2024/6/12 16:36
*/
type QueryCommentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCommentListLogic {
	return &QueryCommentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCommentList 查询商品评价列表
func (l *QueryCommentListLogic) QueryCommentList(in *pmsclient.QueryCommentListReq) (*pmsclient.QueryCommentListResp, error) {
	var hidden *int32
	if in.Hidden >= 0 {
		hidden = &in.Hidden
	}

	result, total, err := l.svcCtx.ProductCommentModel.FindPageWithAuditStatus(l.ctx, model.CommentQueryFilter{
		ProductID:   in.ProductId,
		PlatformID:  in.PlatformId,
		TenantID:    in.TenantId,
		MerchantID:  in.MerchantId,
		PageNo:      in.PageNum,
		PageSize:    in.PageSize,
		ShowStatus:  in.ShowStatus,
		AuditStatus: in.AuditStatus,
		Hidden:      hidden,
		StartTime:   in.StartTime,
		EndTime:     in.EndTime,
		ProductName: in.ProductName,
		MemberName:  in.MemberName,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询商品评价列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品评价列表失败")
	}

	list := make([]*pmsclient.CommentListData, 0, len(result))
	for _, item := range result {
		list = append(list, &pmsclient.CommentListData{
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
			AuditStatus:      item.AuditStatus,
			Hidden:           item.Hidden,
			AuditRemark:      item.AuditRemark,
			AuditorId:        item.AuditorId,
			AuditorName:      item.AuditorName,
			AuditedAt:        formatMongoTime(item.AuditedAt),
			AppealStatus:     item.AppealStatus,
			AppealReason:     item.AppealReason,
			AppealReply:      item.AppealReply,
			AppealedAt:       formatMongoTime(item.AppealedAt),
			AppealHandledAt:  formatMongoTime(item.AppealHandledAt),
		})
	}

	return &pmsclient.QueryCommentListResp{Total: total, List: list}, nil
}
