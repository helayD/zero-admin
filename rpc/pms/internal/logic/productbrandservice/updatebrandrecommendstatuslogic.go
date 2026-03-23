package productbrandservicelogic

import (
	"context"
	"errors"
	"strconv"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBrandRecommendStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBrandRecommendStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBrandRecommendStatusLogic {
	return &UpdateBrandRecommendStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UpdateBrandRecommendStatusLogic) UpdateBrandRecommendStatus(in *pmsclient.UpdateProductBrandStatusReq) (*pmsclient.UpdateProductBrandStatusResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureBrandScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_brand.recommend_status", in.UpdateBy, "", "status="+strconv.Itoa(int(in.Status))); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"recommend_status": in.Status, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "更新商品品牌推荐状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新商品品牌推荐状态失败")
	}
	return &pmsclient.UpdateProductBrandStatusResp{}, nil
}
