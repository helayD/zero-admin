package productattributegroupservicelogic
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

type UpdateProductAttributeGroupLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewUpdateProductAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductAttributeGroupLogic { return &UpdateProductAttributeGroupLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *UpdateProductAttributeGroupLogic) UpdateProductAttributeGroup(in *pmsclient.UpdateProductAttributeGroupReq) (*pmsclient.UpdateProductAttributeGroupResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy); if err != nil { return nil, err }
	if _, err := logiccommon.EnsureAttributeGroupScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "pms.product_attribute_group.update", in.UpdateBy, "", fmt.Sprintf("attributeGroupId=%d", in.Id)); err != nil { return nil, err }
	if err := logiccommon.EnsureScopedCategoryExists(l.ctx, l.svcCtx.DB, currentScope, in.CategoryId, "当前主体无权将属性分组归属到该商品分类"); err != nil { return nil, err }
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute_group").Where("is_deleted = 0 AND id <> ? AND category_id = ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.Id, in.CategoryId, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil { logc.Errorf(l.ctx, "校验商品属性分组重复失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("校验商品属性分组重复失败") }
	if count > 0 { return nil, errors.New(fmt.Sprintf("商品属性分组名称：%s,已存在", in.Name)) }
	updates := map[string]interface{}{"category_id":in.CategoryId,"name":strings.TrimSpace(in.Name),"sort":in.Sort,"status":in.Status,"update_by":in.UpdateBy,"update_time":time.Now()}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute_group").Where("id = ?", in.Id).Updates(updates).Error; err != nil { logc.Errorf(l.ctx, "更新商品属性分组失败,参数:%+v,异常:%s", updates, err.Error()); return nil, errors.New("更新商品属性分组失败") }
	return &pmsclient.UpdateProductAttributeGroupResp{Pong:"ok"}, nil
}
