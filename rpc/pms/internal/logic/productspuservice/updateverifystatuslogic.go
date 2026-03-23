package productspuservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"strconv"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateVerifyStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateVerifyStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVerifyStatusLogic {
	return &UpdateVerifyStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateVerifyStatus 修改审核状态
func (l *UpdateVerifyStatusLogic) UpdateVerifyStatus(in *pmsclient.UpdateProductSpuStatusReq) (*pmsclient.UpdateProductSpuStatusResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureProductScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spu.verify_status", in.UpdateBy, in.ReviewMan, "status="+strconv.Itoa(int(in.Status))); err != nil {
		return nil, err
	}
	if in.Status == logiccommon.ProductVerifyStatusApproved {
		if err := logiccommon.EnsureProductsReviewReady(l.ctx, l.svcCtx.DB, currentScope, in.Ids); err != nil {
			return nil, err
		}
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"verify_status": in.Status,
			"update_by":     in.UpdateBy,
			"update_time":   time.Now(),
		}
		if in.Status == logiccommon.ProductVerifyStatusRejected {
			updates["publish_status"] = logiccommon.ProductPublishStatusOffShelf
			updates["recommend_status"] = logiccommon.ProductRecommendStatusOff
		}

		if err := tx.Table("pms_product_spu").Where("id IN ?", in.Ids).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Table("pms_product_sku").Where("spu_id IN ?", in.Ids).Updates(map[string]interface{}{
			"verify_status": in.Status,
			"update_by":     in.UpdateBy,
			"update_time":   time.Now(),
		}).Error; err != nil {
			return err
		}
		if in.Status == logiccommon.ProductVerifyStatusRejected {
			if err := tx.Table("pms_product_sku").Where("spu_id IN ?", in.Ids).Updates(map[string]interface{}{
				"publish_status": logiccommon.ProductPublishStatusOffShelf,
				"update_by":      in.UpdateBy,
				"update_time":    time.Now(),
			}).Error; err != nil {
				return err
			}
		}

		for _, id := range in.Ids {
			if err := l.svcCtx.ProductVertifyRecordModel.Insert(l.ctx, &model.ProductVertifyRecord{
				ProductId: id,
				ReviewMan: in.ReviewMan,
				Status:    in.Status,
				Detail:    in.Detail,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logc.Errorf(l.ctx, "批量修改审核状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("批量修改审核状态失败")
	}

	syncProductIndexVisibility(l.ctx, l.svcCtx, currentScope, in.Ids, buildProductEventMeta("pms.product_spu.verify_status", in.UpdateBy, in.ReviewMan))

	return &pmsclient.UpdateProductSpuStatusResp{}, nil
}
