package productspuservicelogic

import (
	"context"
	"errors"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"strconv"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdatePublishStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePublishStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePublishStatusLogic {
	return &UpdatePublishStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdatePublishStatus 上下架商品
func (l *UpdatePublishStatusLogic) UpdatePublishStatus(in *pmsclient.UpdateProductSpuStatusReq) (*pmsclient.UpdateProductSpuStatusResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureProductScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spu.publish_status", in.UpdateBy, in.ReviewMan, "status="+strconv.Itoa(int(in.Status))); err != nil {
		return nil, err
	}
	if in.Status == logiccommon.ProductPublishStatusOnShelf {
		if err := logiccommon.EnsureProductsPublishable(l.ctx, l.svcCtx.DB, currentScope, in.Ids); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	detail := strings.TrimSpace(in.Detail)

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"publish_status": in.Status,
			"update_by":      in.UpdateBy,
			"update_time":    now,
			"publish_man":    in.ReviewMan,
			"publish_time":   now,
			"publish_detail": detail,
		}
		if in.Status == logiccommon.ProductPublishStatusOffShelf {
			updates["recommend_status"] = logiccommon.ProductRecommendStatusOff
			updates["recommend_man"] = in.ReviewMan
			updates["recommend_time"] = now
			updates["recommend_detail"] = buildAutoRecommendDetail(detail)
		}
		if err := tx.Table("pms_product_spu").Where("id IN ?", in.Ids).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Table("pms_product_sku").Where("spu_id IN ?", in.Ids).Updates(map[string]interface{}{
			"publish_status": in.Status,
			"update_by":      in.UpdateBy,
			"update_time":    now,
		}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logc.Errorf(l.ctx, "批量上下架商品失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("批量上下架商品失败")
	}

	syncProductIndexVisibility(l.ctx, l.svcCtx, currentScope, in.Ids, buildProductEventMeta("pms.product_spu.publish_status", in.UpdateBy, in.ReviewMan))

	return &pmsclient.UpdateProductSpuStatusResp{}, nil
}
