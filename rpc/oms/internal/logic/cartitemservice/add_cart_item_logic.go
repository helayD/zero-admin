package cartitemservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddCartItemLogic 添加购物车
/*
Author: LiuFeiHua
Date: 2024/6/12 9:40
*/
type AddCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCartItemLogic {
	return &AddCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddCartItem 添加购物车（幂等化改造）
// 使用 MySQL 原生 INSERT ... ON DUPLICATE KEY UPDATE 替代"先查后改"并发不安全模式
// 唯一索引: uk_member_sku_status(member_id, product_sku_id, delete_status)
// - 若记录不存在：INSERT 新记录（quantity = req.Quantity）
// - 若记录存在（同一 member_id + product_sku_id + delete_status=0）：UPDATE 累加数量（quantity = quantity + req.Quantity）
func (l *AddCartItemLogic) AddCartItem(in *omsclient.AddCartItemReq) (*omsclient.CartItemResp, error) {
	q := query.OmsCartItem

	now := time.Now()
	nowPtr := now
	timeoutDays := l.svcCtx.C.Cart.Timeout
	if timeoutDays <= 0 {
		timeoutDays = 30 // 默认30天，防止配置缺失时购物车立即过期
	}
	expireTime := now.AddDate(0, 0, timeoutDays)

	newItem := &model.OmsCartItem{
		MemberID:          in.MemberId,
		ProductID:         in.ProductId,
		ProductSkuID:      in.ProductSkuId,
		Quantity:          in.Quantity,
		Price:             float64(in.Price),
		Selected:          in.Selected,
		ProductName:       in.ProductName,
		ProductSubTitle:   in.ProductSubTitle,
		ProductPic:        in.ProductPic,
		ProductSkuCode:    in.ProductSkuCode,
		ProductSn:         in.ProductSn,
		ProductBrand:      in.ProductBrand,
		ProductCategoryID: in.ProductCategoryId,
		ProductAttr:       in.ProductAttr,
		MemberNickname:    in.MemberNickname,
		Source:            in.Source,
		DeleteStatus:      0,
		ExpireTime:        expireTime,
		CreateTime:        now,
		UpdateTime:        &nowPtr,
	}

	// 幂等化 upsert：INSERT ... ON DUPLICATE KEY UPDATE
	// 无竞态条件，并发重复加购同一商品不会产生 duplicate key 错误
	err := q.WithContext(l.ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "member_id"},
			{Name: "product_sku_id"},
			{Name: "delete_status"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity":     gorm.Expr("quantity + ?", in.Quantity),
			"selected":    in.Selected,
			"update_time": now,
		}),
	}).Create(newItem)
	if err != nil {
		logc.Errorf(l.ctx, "添加购物车失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加购物车失败")
	}

	return &omsclient.CartItemResp{}, nil
}
