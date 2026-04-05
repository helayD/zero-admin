package cartitemservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCartItemListLogic 查询购物车列表
/*
Author: LiuFeiHua
Date: 2024/6/12 9:56
*/
type QueryCartItemListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCartItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCartItemListLogic {
	return &QueryCartItemListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCartItemList 查询购物车列表
func (l *QueryCartItemListLogic) QueryCartItemList(in *omsclient.QueryCartItemListReq) (*omsclient.QueryCartItemListResp, error) {
	type cartItemRow struct {
		ID                int64        `gorm:"column:id"`
		MemberID          int64        `gorm:"column:member_id"`
		ProductID         int64        `gorm:"column:product_id"`
		ProductSkuID      int64        `gorm:"column:product_sku_id"`
		Quantity          int32        `gorm:"column:quantity"`
		Price             float64      `gorm:"column:price"`
		Selected          int32        `gorm:"column:selected"`
		ProductName       string       `gorm:"column:product_name"`
		ProductSubTitle   string       `gorm:"column:product_sub_title"`
		ProductPic        string       `gorm:"column:product_pic"`
		ProductSkuCode    string       `gorm:"column:product_sku_code"`
		ProductSn         string       `gorm:"column:product_sn"`
		ProductBrand      string       `gorm:"column:product_brand"`
		ProductCategoryID int64        `gorm:"column:product_category_id"`
		ProductAttr       string       `gorm:"column:product_attr"`
		MemberNickname    string       `gorm:"column:member_nickname"`
		Source            int32        `gorm:"column:source"`
		DeleteStatus      int32        `gorm:"column:delete_status"`
		ExpireTime        sql.NullTime `gorm:"column:expire_time"`
		CreateTime        sql.NullTime `gorm:"column:create_time"`
		UpdateTime        sql.NullTime `gorm:"column:update_time"`
		ActivityType      string       `gorm:"column:activity_type"`
		ActivityID        int64        `gorm:"column:activity_id"`
	}

	var result []cartItemRow
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("oms_cart_item").
		Where("member_id = ? AND delete_status = 0", in.MemberId).
		Order("id desc").
		Scan(&result).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询购物车列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询购物车列表失败")
	}
	var list []*omsclient.CartItemData
	for _, item := range result {
		data := &omsclient.CartItemData{
			Id:                item.ID,                                 // 主键ID
			MemberId:          item.MemberID,                           // 会员ID
			ProductId:         item.ProductID,                          // 商品ID
			ProductSkuId:      item.ProductSkuID,                       // 商品SKU ID
			Quantity:          item.Quantity,                           // 购买数量
			Price:             float32(item.Price),                     // 添加到购物车时的价格
			Selected:          item.Selected,                           // 是否选中 0-未选中 1-选中
			ProductName:       item.ProductName,                        // 商品名称
			ProductSubTitle:   item.ProductSubTitle,                    // 商品副标题
			ProductPic:        item.ProductPic,                         // 商品主图URL
			ProductSkuCode:    item.ProductSkuCode,                     // 商品SKU编码
			ProductSn:         item.ProductSn,                          // 商品货号
			ProductBrand:      item.ProductBrand,                       // 商品品牌
			ProductCategoryId: item.ProductCategoryID,                  // 商品分类ID
			ProductAttr:       item.ProductAttr,                        // 商品销售属性JSON
			MemberNickname:    item.MemberNickname,                     // 会员昵称
			Source:            item.Source,                             // 来源 1-PC 2-H5 3-小程序 4-APP
			DeleteStatus:      item.DeleteStatus,                       // 删除状态 0-正常 1-删除
			ExpireTime:        time_util.TimeToString(nullTimePtr(item.ExpireTime)), // 过期时间
			CreateTime:        time_util.TimeToString(nullTimePtr(item.CreateTime)), // 创建时间
			UpdateTime:        time_util.TimeToString(nullTimePtr(item.UpdateTime)), // 更新时间
			ActivityType:      item.ActivityType,                       // 活动类型
			ActivityId:        item.ActivityID,                         // 活动ID
		}
		list = append(list, data)
	}

	return &omsclient.QueryCartItemListResp{
		List: list,
	}, nil

}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
