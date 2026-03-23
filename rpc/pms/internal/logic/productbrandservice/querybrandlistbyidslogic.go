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
)

type QueryBrandListByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryBrandListByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryBrandListByIdsLogic {
	return &QueryBrandListByIdsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *QueryBrandListByIdsLogic) QueryBrandListByIds(in *pmsclient.QueryBrandListByIdsReq) (*pmsclient.QueryProductBrandListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	var rows []logiccommon.CatalogScopeRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("id IN ? AND is_deleted = 0", in.Ids).Where(scopeWhere, scopeArgs...).Order("sort asc, id desc").Scan(&rows).Error; err != nil {
		logc.Errorf(l.ctx, "根据ids查询品牌信息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("根据ids查询品牌信息失败")
	}
	list := make([]*pmsclient.ProductBrandListData, 0, len(rows))
	for _, item := range rows {
		list = append(list, &pmsclient.ProductBrandListData{Id: item.ID, Name: item.Name, Logo: item.Logo, BigPic: item.BigPic, Description: item.Description, FirstLetter: item.FirstLetter, Sort: item.Sort, RecommendStatus: item.RecommendStatus, ProductCount: item.ProductCount, ProductCommentCount: item.ProductCommentCount, IsEnabled: item.IsEnabled, CreateBy: item.CreateBy, CreateTime: time_util.TimeToStr(item.CreateTime), UpdateBy: derefInt64(item.UpdateBy), UpdateTime: time_util.TimeToString(item.UpdateTime), ScopeType: item.GovernanceScope().ScopeType, PlatformId: item.PlatformID, TenantId: item.TenantID, MerchantId: item.MerchantID})
	}
	return &pmsclient.QueryProductBrandListResp{Total: int64(len(list)), List: list}, nil
}
