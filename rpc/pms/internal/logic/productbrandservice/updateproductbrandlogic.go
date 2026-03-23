package productbrandservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductBrandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductBrandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductBrandLogic {
	return &UpdateProductBrandLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateProductBrandLogic) UpdateProductBrand(in *pmsclient.UpdateProductBrandReq) (*pmsclient.UpdateProductBrandResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureBrandScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "pms.product_brand.update", in.UpdateBy, "", fmt.Sprintf("brandId=%d", in.Id)); err != nil {
		return nil, err
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("is_deleted = 0 AND id <> ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.Id, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "校验商品品牌重复失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("校验商品品牌重复失败")
	}
	if count > 0 {
		return nil, errors.New(fmt.Sprintf("品牌名称：%s,已存在", in.Name))
	}

	updates := map[string]interface{}{"name": strings.TrimSpace(in.Name), "logo": in.Logo, "big_pic": in.BigPic, "description": in.Description, "first_letter": strings.TrimSpace(in.FirstLetter), "sort": in.Sort, "recommend_status": in.RecommendStatus, "is_enabled": in.IsEnabled, "update_by": in.UpdateBy, "update_time": time.Now()}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		logc.Errorf(l.ctx, "更新商品品牌失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新商品品牌失败")
	}
	return &pmsclient.UpdateProductBrandResp{}, nil
}
