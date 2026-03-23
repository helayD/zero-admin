package productattributeservicelogic

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

type UpdateProductAttributeLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewUpdateProductAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductAttributeLogic { return &UpdateProductAttributeLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *UpdateProductAttributeLogic) UpdateProductAttribute(in *pmsclient.UpdateProductAttributeReq) (*pmsclient.UpdateProductAttributeResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy); if err != nil { return nil, err }
	if _, err := logiccommon.EnsureAttributeScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "pms.product_attribute.update", in.UpdateBy, "", fmt.Sprintf("attributeId=%d", in.Id)); err != nil { return nil, err }
	if err := logiccommon.EnsureScopedAttributeGroupExists(l.ctx, l.svcCtx.DB, currentScope, in.GroupId, "当前主体无权将商品属性归属到该属性分组"); err != nil { return nil, err }
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Where("is_deleted = 0 AND id <> ? AND group_id = ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.Id, in.GroupId, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil { logc.Errorf(l.ctx, "校验商品属性重复失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("校验商品属性重复失败") }
	if count > 0 { return nil, errors.New(fmt.Sprintf("商品属性名称：%s,已存在", in.Name)) }
	updates := map[string]interface{}{"group_id":in.GroupId,"name":strings.TrimSpace(in.Name),"input_type":in.InputType,"value_type":in.ValueType,"input_list":in.InputList,"unit":in.Unit,"is_required":in.IsRequired,"is_searchable":in.IsSearchable,"is_show":in.IsShow,"sort":in.Sort,"status":in.Status,"update_by":in.UpdateBy,"update_time":time.Now()}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Where("id = ?", in.Id).Updates(updates).Error; err != nil { logc.Errorf(l.ctx, "更新商品属性失败,参数:%+v,异常:%s", updates, err.Error()); return nil, errors.New("更新商品属性失败") }
	return &pmsclient.UpdateProductAttributeResp{Pong:"ok"}, nil
}
