package commentservicelogic

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateCommentLogic 更新商品评价
/*
Author: LiuFeiHua
Date: 2024/6/12 16:38
Updated: 2026/04/02 - Review Fix H-NEW-1 (批量操作) & H-NEW-2 (scope 防越权)
*/
type UpdateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentLogic {
	return &UpdateCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateComment 更新商品评价（支持单个更新 + 批量审核/屏蔽）
func (l *UpdateCommentLogic) UpdateComment(in *pmsclient.UpdateCommentReq) (*pmsclient.UpdateCommentResp, error) {
	var affectedCount int64

	// Review Fix H-NEW-1: 批量操作支持
	if in.Ids != "" {
		ids := strings.Split(in.Ids, ",")
		affectedCount = int64(len(ids))

		for _, idStr := range ids {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			err := l.svcCtx.ProductCommentModel.UpdateStatus(l.ctx, idStr, in.ShowStatus, in.UpdateBy)
			if err != nil {
				logc.Errorf(l.ctx, "批量更新评价失败,ID:%s,异常:%s", idStr, err.Error())
			}
		}

		// 异步更新被操作商品的统计
		go l.updateSpuCommentStatsByIds(l.ctx, in)
	} else {
		// 单个更新：带 scope 校验
		item, err := l.svcCtx.ProductCommentModel.FindOne(l.ctx, in.Id)
		if err != nil {
			if errors.Is(err, mon.ErrNotFound) {
				return nil, errors.New("评价不存在")
			}
			logc.Errorf(l.ctx, "查询评价失败,ID:%s,异常:%s", in.Id, err.Error())
			return nil, errors.New("查询评价失败")
		}

		// Review Fix H-NEW-2: scope 防越权校验
		if in.PlatformId > 0 && item.PlatformId != in.PlatformId {
			return nil, errors.New("无权操作该评价")
		}
		if in.TenantId > 0 && item.TenantId != in.TenantId {
			return nil, errors.New("无权操作该评价")
		}
		if in.MerchantId > 0 && item.MerchantId != in.MerchantId {
			return nil, errors.New("无权操作该评价")
		}

		objectID, _ := bson.ObjectIDFromHex(in.Id)
		_, err = l.svcCtx.ProductCommentModel.Update(l.ctx, &model.ProductComment{
			ID:               objectID,
			ProductId:        in.ProductId,
			MemberNickName:   in.MemberNickName,
			ProductName:      in.ProductName,
			Star:             in.Star,
			MemberIp:         in.MemberIp,
			ShowStatus:       in.ShowStatus,
			ProductAttribute: in.ProductAttribute,
			CollectCount:     in.CollectCount,
			ReadCount:        in.ReadCount,
			Content:          in.Content,
			Pics:             in.Pics,
			MemberIcon:       in.MemberIcon,
			ReplayCount:      in.ReplayCount,
			UpdateBy:         in.UpdateBy,
		})
		if err != nil {
			logc.Errorf(l.ctx, "更新商品评价失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("更新商品评价失败")
		}

		affectedCount = 1

		// 异步更新商品评价统计（审核状态变更后重算）
		go l.updateSpuCommentStats(in.ProductId)
	}

	return &pmsclient.UpdateCommentResp{
		Pong:          "ok",
		AffectedCount: affectedCount,
	}, nil
}

// updateSpuCommentStatsByIds 异步更新多个商品的统计（批量操作）
func (l *UpdateCommentLogic) updateSpuCommentStatsByIds(ctx context.Context, in *pmsclient.UpdateCommentReq) {
	if in.Ids == "" {
		return
	}
	ids := strings.Split(in.Ids, ",")
	productIds := make(map[int64]bool)
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		item, err := l.svcCtx.ProductCommentModel.FindOne(ctx, idStr)
		if err != nil {
			continue
		}
		productIds[item.ProductId] = true
	}
	for productId := range productIds {
		l.updateSpuCommentStatsById(ctx, productId)
	}
}

// updateSpuCommentStatsById 异步更新单个商品评价统计
func (l *UpdateCommentLogic) updateSpuCommentStatsById(ctx context.Context, productId int64) {
	count, err := l.svcCtx.ProductCommentModel.CountByProductId(ctx, productId)
	if err != nil {
		logc.Errorf(ctx, "统计商品[%d]评价数量失败:%s", productId, err.Error())
		return
	}
	avgStar, err := l.svcCtx.ProductCommentModel.AvgStarByProductId(ctx, productId)
	if err != nil {
		logc.Errorf(ctx, "统计商品[%d]平均评分失败:%s", productId, err.Error())
		return
	}
	err = l.svcCtx.DB.WithContext(ctx).Exec(
		"UPDATE pms_product_spu SET product_comment_count = ?, product_star = ? WHERE id = ?",
		count, avgStar, productId,
	).Error
	if err != nil {
		logc.Errorf(ctx, "更新商品[%d]评价统计失败:%s", productId, err.Error())
	}
}

// updateSpuCommentStats 异步更新商品评价统计
func (l *UpdateCommentLogic) updateSpuCommentStats(productId int64) {
	ctx := context.Background()
	l.updateSpuCommentStatsById(ctx, productId)
}
