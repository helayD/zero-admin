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

type LockSkuStockLockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockSkuStockLockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockSkuStockLockLogic {
	return &LockSkuStockLockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LockSkuStockLock 下单的时候,锁定库存
func (l *LockSkuStockLockLogic) LockSkuStockLock(in *pmsclient.UpdateSkuStockReq) (*pmsclient.UpdateSkuStockLockResp, error) {
	sql := "update pms_product_sku set stock=stock-?,lock_stock=lock_stock+? where `id` = ? and stock>=?"
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range in.Data {
			result := tx.Exec(sql, item.ProductQuantity, item.ProductQuantity, item.Id, item.ProductQuantity)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("商品库存不足")
			}
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "下单的时候,锁定库存失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("下单的时候,锁定库存失败")
	}

	return &pmsclient.UpdateSkuStockLockResp{}, nil
}
