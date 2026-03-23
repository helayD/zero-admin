package productbrandservicelogic

import (
	"context"
	"errors"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryProductBrandDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductBrandDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductBrandDetailLogic {
	return &QueryProductBrandDetailLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *QueryProductBrandDetailLogic) QueryProductBrandDetail(in *pmsclient.QueryProductBrandDetailReq) (*pmsclient.QueryProductBrandDetailResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	var item logiccommon.CatalogScopeRow
	err = l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("id = ? AND is_deleted = 0", in.Id).Where(scopeWhere, scopeArgs...).Take(&item).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "商品品牌不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("商品品牌不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询商品品牌异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询商品品牌异常")
	}

	return &pmsclient.QueryProductBrandDetailResp{Id: item.ID, Name: item.Name, Logo: item.Logo, BigPic: item.BigPic, Description: item.Description, FirstLetter: item.FirstLetter, Sort: item.Sort, RecommendStatus: item.RecommendStatus, ProductCount: item.ProductCount, ProductCommentCount: item.ProductCommentCount, IsEnabled: item.IsEnabled, CreateBy: item.CreateBy, CreateTime: time_util.TimeToStr(item.CreateTime), UpdateBy: derefInt64(item.UpdateBy), UpdateTime: time_util.TimeToString(item.UpdateTime), ScopeType: item.GovernanceScope().ScopeType, PlatformId: item.PlatformID, TenantId: item.TenantID, MerchantId: item.MerchantID}, nil
}
