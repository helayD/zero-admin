package productcategoryservicelogic

import (
	"context"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductCategoryTreeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductCategoryTreeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductCategoryTreeListLogic {
	return &QueryProductCategoryTreeListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *QueryProductCategoryTreeListLogic) QueryProductCategoryTreeList(in *pmsclient.QueryProductCategoryTreeListReq) (*pmsclient.QueryProductCategoryListTreeResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	categoryList, err := queryTreeLevel(l, current, 0)
	if err != nil {
		return nil, err
	}
	list := make([]*pmsclient.QueryProductCategoryListTreeData, 0, len(categoryList))
	for _, item := range categoryList {
		children, err := queryTreeLevel(l, current, item.Id)
		if err != nil {
			return nil, err
		}
		list = append(list, &pmsclient.QueryProductCategoryListTreeData{Id: item.Id, Name: item.Name, ImageUrl: item.ImageUrl, Children: children})
	}
	return &pmsclient.QueryProductCategoryListTreeResp{List: list}, nil
}
func queryTreeLevel(l *QueryProductCategoryTreeListLogic, current pkgscope.GovernanceScope, parentID int64) ([]*pmsclient.QueryProductCategoryListTreeData, error) {
	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	var rows []logiccommon.CatalogScopeRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("is_deleted = 0 AND is_enabled = 1 AND parent_id = ?", parentID).Where(scopeWhere, scopeArgs...).Order("sort asc, id asc").Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]*pmsclient.QueryProductCategoryListTreeData, 0, len(rows))
	for _, category := range rows {
		list = append(list, &pmsclient.QueryProductCategoryListTreeData{Id: category.ID, Name: category.Name, ImageUrl: category.Icon})
	}
	return list, nil
}
