package productskuservicelogic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReleaseSkuStockLockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseSkuStockLockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseSkuStockLockLogic {
	return &ReleaseSkuStockLockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReleaseSkuStockLock 取消订单的时候,释放库存
func (l *ReleaseSkuStockLockLogic) ReleaseSkuStockLock(in *pmsclient.UpdateSkuStockReq) (*pmsclient.UpdateSkuStockLockResp, error) {
	sql := "update pms_product_sku set stock=stock+?,lock_stock=lock_stock-? where `id` = ? and lock_stock>=?"
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range in.Data {
			result := tx.Exec(sql, item.ProductQuantity, item.ProductQuantity, item.Id, item.ProductQuantity)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("锁定库存不足")
			}
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新商品库存失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新商品库存失败")
	}

	return &pmsclient.UpdateSkuStockLockResp{}, nil
}
