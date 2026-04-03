package commentservicelogic

import (
	"context"
	"errors"

	mymongo "github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// AddCommentLogic 添加商品评价
/*
Author: LiuFeiHua
Date: 2024/6/12 16:35
*/
type AddCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
	return &AddCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddComment 添加商品评价
func (l *AddCommentLogic) AddComment(in *pmsclient.AddCommentReq) (*pmsclient.AddCommentResp, error) {
	_, err := l.svcCtx.ProductCommentModel.FindOneByMemberProductOrder(l.ctx, in.MemberId, in.ProductId, in.OrderId)
	if err == nil {
		return nil, errors.New("您已评价过该商品")
	}
	if !errors.Is(err, mymongo.ErrNotFound) {
		logc.Errorf(l.ctx, "重复评价查询异常,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("系统异常，请稍后重试")
	}

	err = l.svcCtx.ProductCommentModel.Insert(l.ctx, &mymongo.ProductComment{
		ProductId:        in.ProductId,
		MemberNickName:   in.MemberNickName,
		ProductName:      in.ProductName,
		Star:             in.Star,
		MemberIp:         in.MemberIp,
		ShowStatus:       in.ShowStatus,
		AuditStatus:      0,
		Hidden:           0,
		AppealStatus:     0,
		ProductAttribute: in.ProductAttribute,
		CollectCount:     in.CollectCount,
		ReadCount:        in.ReadCount,
		Content:          in.Content,
		Pics:             in.Pics,
		MemberIcon:       in.MemberIcon,
		ReplayCount:      in.ReplayCount,
		MemberId:         in.MemberId,
		PlatformId:       in.PlatformId,
		TenantId:         in.TenantId,
		MerchantId:       in.MerchantId,
		OrderId:          in.OrderId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加商品评价失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加商品评价失败")
	}

	updateCommentStatsAsync(l.svcCtx, in.ProductId)
	return &pmsclient.AddCommentResp{}, nil
}
