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

// QueryProductBrandListLogic 查询商品品牌列表
type QueryProductBrandListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductBrandListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductBrandListLogic {
	return &QueryProductBrandListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *QueryProductBrandListLogic) QueryProductBrandList(in *pmsclient.QueryProductBrandListReq) (*pmsclient.QueryProductBrandListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	db := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("is_deleted = 0")
	if in.Name != "" {
		db = db.Where("name LIKE ?", "%"+in.Name+"%")
	}
	if in.RecommendStatus != 2 {
		db = db.Where("recommend_status = ?", in.RecommendStatus)
	}
	if in.IsEnabled != 2 {
		db = db.Where("is_enabled = ?", in.IsEnabled)
	}
	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	db = db.Where(scopeWhere, scopeArgs...)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "查询商品品牌总数失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品品牌列表失败")
	}

	var rows []logiccommon.CatalogScopeRow
	if err := db.Order("sort asc, id desc").Offset(int((in.PageNum - 1) * in.PageSize)).Limit(int(in.PageSize)).Scan(&rows).Error; err != nil {
		logc.Errorf(l.ctx, "查询商品品牌列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品品牌列表失败")
	}

	list := make([]*pmsclient.ProductBrandListData, 0, len(rows))
	for _, item := range rows {
		list = append(list, &pmsclient.ProductBrandListData{
			Id:                  item.ID,
			Name:                item.Name,
			Logo:                item.Logo,
			BigPic:              item.BigPic,
			Description:         item.Description,
			FirstLetter:         item.FirstLetter,
			Sort:                item.Sort,
			RecommendStatus:     item.RecommendStatus,
			ProductCount:        item.ProductCount,
			ProductCommentCount: item.ProductCommentCount,
			IsEnabled:           item.IsEnabled,
			CreateBy:            item.CreateBy,
			CreateTime:          time_util.TimeToStr(item.CreateTime),
			UpdateBy:            derefInt64(item.UpdateBy),
			UpdateTime:          time_util.TimeToString(item.UpdateTime),
			ScopeType:           item.GovernanceScope().ScopeType,
			PlatformId:          item.PlatformID,
			TenantId:            item.TenantID,
			MerchantId:          item.MerchantID,
		})
	}

	return &pmsclient.QueryProductBrandListResp{Total: total, List: list}, nil
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
