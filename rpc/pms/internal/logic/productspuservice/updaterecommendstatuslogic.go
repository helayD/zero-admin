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
)

type UpdateRecommendStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRecommendStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRecommendStatusLogic {
	return &UpdateRecommendStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateRecommendStatus 推荐商品
func (l *UpdateRecommendStatusLogic) UpdateRecommendStatus(in *pmsclient.UpdateProductSpuStatusReq) (*pmsclient.UpdateProductSpuStatusResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureProductScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spu.recommend_status", in.UpdateBy, in.ReviewMan, "status="+strconv.Itoa(int(in.Status))); err != nil {
		return nil, err
	}
	if in.Status == logiccommon.ProductRecommendStatusOn {
		if err := logiccommon.EnsureProductsRecommendable(l.ctx, l.svcCtx.DB, currentScope, in.Ids); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	detail := strings.TrimSpace(in.Detail)

	err = l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_spu").Where("id IN ?", in.Ids).Updates(map[string]interface{}{
		"recommend_status": in.Status,
		"recommend_man":    in.ReviewMan,
		"recommend_time":   now,
		"recommend_detail": detail,
		"update_by":        in.UpdateBy,
		"update_time":      now,
	}).Error

	if err != nil {
		logc.Errorf(l.ctx, "批量推荐商品失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("批量推荐商品失败")
	}

	syncProductIndexVisibility(l.ctx, l.svcCtx, currentScope, in.Ids, buildProductEventMeta("pms.product_spu.recommend_status", in.UpdateBy, in.ReviewMan))

	return &pmsclient.UpdateProductSpuStatusResp{}, nil
}
