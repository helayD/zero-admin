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
	// 重复评价校验
	_, err := l.svcCtx.ProductCommentModel.FindOneByMemberProductOrder(l.ctx, in.MemberId, in.ProductId, in.OrderId)
	if err == nil {
		return nil, errors.New("您已评价过该商品")
	}
	if !errors.Is(err, mymongo.ErrNotFound) {
		logc.Errorf(l.ctx, "重复评价查询异常,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("系统异常，请稍后重试")
	}

	err = l.svcCtx.ProductCommentModel.Insert(l.ctx, &mymongo.ProductComment{
		ProductId:        in.ProductId,          // 商品id
		MemberNickName:   in.MemberNickName,    // 评价者昵称
		ProductName:      in.ProductName,        // 商品名称
		Star:             in.Star,              // 评价星数：0->5
		MemberIp:         in.MemberIp,           // 评价的ip
		ShowStatus:       in.ShowStatus,        // 是否显示，0-不显示，1-显示
		ProductAttribute: in.ProductAttribute,  // 购买时的商品属性
		CollectCount:     in.CollectCount,      // 点赞数
		ReadCount:        in.ReadCount,          // 阅读数
		Content:          in.Content,            // 内容
		Pics:             in.Pics,               // 上传图片地址，以逗号隔开
		MemberIcon:       in.MemberIcon,         // 评论用户头像
		ReplayCount:      in.ReplayCount,        // 回复数量
		MemberId:         in.MemberId,           // 会员ID
		PlatformId:       in.PlatformId,         // 平台ID
		TenantId:         in.TenantId,           // 租户ID
		MerchantId:       in.MerchantId,         // 商户ID
		OrderId:          in.OrderId,            // 订单ID
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加商品评价失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加商品评价失败")
	}

	// 异步更新商品评价统计（productCommentCount / productStar）
	go l.updateSpuCommentStats(in.ProductId)

	return &pmsclient.AddCommentResp{}, nil
}

// updateSpuCommentStats 异步更新商品评价统计
func (l *AddCommentLogic) updateSpuCommentStats(productId int64) {
	ctx := context.Background()

	// 统计已通过评价数量
	count, err := l.svcCtx.ProductCommentModel.CountByProductId(ctx, productId)
	if err != nil {
		logc.Errorf(ctx, "统计商品[%d]评价数量失败:%s", productId, err.Error())
		return
	}

	// 统计平均评分
	avgStar, err := l.svcCtx.ProductCommentModel.AvgStarByProductId(ctx, productId)
	if err != nil {
		logc.Errorf(ctx, "统计商品[%d]平均评分失败:%s", productId, err.Error())
		return
	}

	// 更新 pms_product_spu 表（使用原生 SQL，因为 GORM 模型未定义 productCommentCount/productStar）
	err = l.svcCtx.DB.WithContext(ctx).Exec(
		"UPDATE pms_product_spu SET product_comment_count = ?, product_star = ? WHERE id = ?",
		count, avgStar, productId,
	).Error
	if err != nil {
		logc.Errorf(ctx, "更新商品[%d]评价统计失败:%s", productId, err.Error())
	}
}
