package productattributegroupservicelogic
import (
	"context"
	"errors"
	"fmt"
	"strings"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddProductAttributeGroupLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewAddProductAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductAttributeGroupLogic { return &AddProductAttributeGroupLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *AddProductAttributeGroupLogic) AddProductAttributeGroup(in *pmsclient.AddProductAttributeGroupReq) (*pmsclient.AddProductAttributeGroupResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy); if err != nil { return nil, err }
	if err := logiccommon.EnsureScopedCategoryExists(l.ctx, l.svcCtx.DB, currentScope, in.CategoryId, "当前主体无权在该商品分类下创建属性分组"); err != nil { return nil, err }
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute_group").Where("is_deleted = 0 AND category_id = ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.CategoryId, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil { logc.Errorf(l.ctx, "校验商品属性分组重复失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("校验商品属性分组重复失败") }
	if count > 0 { return nil, errors.New(fmt.Sprintf("商品属性分组名称：%s,已存在", in.Name)) }
	item := map[string]interface{}{"category_id":in.CategoryId,"name":strings.TrimSpace(in.Name),"sort":in.Sort,"status":in.Status,"create_by":in.CreateBy,"platform_id":currentScope.PlatformID,"tenant_id":currentScope.TenantID,"merchant_id":currentScope.MerchantID}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute_group").Create(item).Error; err != nil { logc.Errorf(l.ctx, "添加商品属性分组失败,参数:%+v,异常:%s", item, err.Error()); return nil, errors.New("添加商品属性分组失败") }
	return &pmsclient.AddProductAttributeGroupResp{Pong:"ok"}, nil
}
