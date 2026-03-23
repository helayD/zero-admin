package productattributeservicelogic
import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/pkg/pointerprocess"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductAttributeListLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewQueryProductAttributeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductAttributeListLogic { return &QueryProductAttributeListLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *QueryProductAttributeListLogic) QueryProductAttributeList(in *pmsclient.QueryProductAttributeListReq) (*pmsclient.QueryProductAttributeListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope); if err != nil { return nil, err }
	db := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Where("is_deleted = 0")
	db = pkgscope.ApplyGovernanceScope(db, current, "")
	if in.GroupId != 0 { db = db.Where("group_id = ?", in.GroupId) }
	if in.Name != "" { db = db.Where("name LIKE ?", "%"+in.Name+"%") }
	if in.InputType != 0 { db = db.Where("input_type = ?", in.InputType) }
	if in.IsRequired != 2 { db = db.Where("is_required = ?", in.IsRequired) }
	if in.IsSearchable != 2 { db = db.Where("is_searchable = ?", in.IsSearchable) }
	if in.IsShow != 2 { db = db.Where("is_show = ?", in.IsShow) }
	if in.Status != 2 { db = db.Where("status = ?", in.Status) }
	var total int64
	if err := db.Count(&total).Error; err != nil { logc.Errorf(l.ctx, "查询商品属性总数失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("查询商品属性列表失败") }
	var rows []logiccommon.ProductAttributeRow
	if err := db.Order("sort asc, id desc").Offset(int((in.PageNum-1)*in.PageSize)).Limit(int(in.PageSize)).Scan(&rows).Error; err != nil { logc.Errorf(l.ctx, "查询商品属性列表失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("查询商品属性列表失败") }
	list := make([]*pmsclient.ProductAttributeListData,0,len(rows)); for _, item := range rows { list = append(list, &pmsclient.ProductAttributeListData{Id:item.ID,GroupId:item.GroupID,Name:item.Name,InputType:item.InputType,ValueType:item.ValueType,InputList:item.InputList,Unit:item.Unit,IsRequired:item.IsRequired,IsSearchable:item.IsSearchable,IsShow:item.IsShow,Sort:item.Sort,Status:item.Status,CreateBy:item.CreateBy,CreateTime:time_util.TimeToStr(item.CreateTime),UpdateBy:pointerprocess.DefaltData(item.UpdateBy).(int64),UpdateTime:time_util.TimeToString(item.UpdateTime),ScopeType:current.ScopeType,PlatformId:item.PlatformID,TenantId:item.TenantID,MerchantId:item.MerchantID}) }
	return &pmsclient.QueryProductAttributeListResp{Total:total,List:list}, nil
}
