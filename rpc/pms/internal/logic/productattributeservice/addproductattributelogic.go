package productattributeservicelogic

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

type AddProductAttributeLogic struct { ctx context.Context; svcCtx *svc.ServiceContext; logx.Logger }
func NewAddProductAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductAttributeLogic { return &AddProductAttributeLogic{ctx:ctx, svcCtx:svcCtx, Logger:logx.WithContext(ctx)} }
func (l *AddProductAttributeLogic) AddProductAttribute(in *pmsclient.AddProductAttributeReq) (*pmsclient.AddProductAttributeResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy); if err != nil { return nil, err }
	if err := logiccommon.EnsureScopedAttributeGroupExists(l.ctx, l.svcCtx.DB, currentScope, in.GroupId, "当前主体无权在该属性分组下创建商品属性"); err != nil { return nil, err }
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Where("is_deleted = 0 AND group_id = ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.GroupId, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil { logc.Errorf(l.ctx, "校验商品属性重复失败,参数:%+v,异常:%s", in, err.Error()); return nil, errors.New("校验商品属性重复失败") }
	if count > 0 { return nil, errors.New(fmt.Sprintf("商品属性名称：%s,已存在", in.Name)) }
	item := map[string]interface{}{"group_id":in.GroupId,"name":strings.TrimSpace(in.Name),"input_type":in.InputType,"value_type":in.ValueType,"input_list":in.InputList,"unit":in.Unit,"is_required":in.IsRequired,"is_searchable":in.IsSearchable,"is_show":in.IsShow,"sort":in.Sort,"status":in.Status,"create_by":in.CreateBy,"platform_id":currentScope.PlatformID,"tenant_id":currentScope.TenantID,"merchant_id":currentScope.MerchantID}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Create(item).Error; err != nil { logc.Errorf(l.ctx, "添加商品属性失败,参数:%+v,异常:%s", item, err.Error()); return nil, errors.New("添加商品属性失败") }
	return &pmsclient.AddProductAttributeResp{Pong:"ok"}, nil
}
