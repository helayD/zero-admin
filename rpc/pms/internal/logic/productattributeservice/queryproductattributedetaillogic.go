package productattributeservicelogic
import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/pkg/pointerprocess"
	"github.com/feihua/zero-admin/pkg/time_util"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryProductAttributeDetailLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewQueryProductAttributeDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductAttributeDetailLogic { return &QueryProductAttributeDetailLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *QueryProductAttributeDetailLogic) QueryProductAttributeDetail(in *pmsclient.QueryProductAttributeDetailReq) (*pmsclient.QueryProductAttributeDetailResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope); if err != nil { return nil, err }
	var item logiccommon.ProductAttributeRow
	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	err = l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Where("id = ? AND is_deleted = 0", in.Id).Where(scopeWhere, scopeArgs...).Take(&item).Error
	switch { case errors.Is(err, gorm.ErrRecordNotFound): logc.Errorf(l.ctx, "商品属性不存在, 请求参数：%+v, 异常信息: %s", in, err.Error()); return nil, errors.New("商品属性不存在"); case err != nil: logc.Errorf(l.ctx, "查询商品属性异常, 请求参数：%+v, 异常信息: %s", in, err.Error()); return nil, errors.New("查询商品属性异常") }
	return &pmsclient.QueryProductAttributeDetailResp{Id:item.ID,GroupId:item.GroupID,Name:item.Name,InputType:item.InputType,ValueType:item.ValueType,InputList:item.InputList,Unit:item.Unit,IsRequired:item.IsRequired,IsSearchable:item.IsSearchable,IsShow:item.IsShow,Sort:item.Sort,Status:item.Status,CreateBy:item.CreateBy,CreateTime:time_util.TimeToStr(item.CreateTime),UpdateBy:pointerprocess.DefaltData(item.UpdateBy).(int64),UpdateTime:time_util.TimeToString(item.UpdateTime),ScopeType:current.ScopeType,PlatformId:item.PlatformID,TenantId:item.TenantID,MerchantId:item.MerchantID}, nil
}
