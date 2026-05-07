package productfulfillmentruleservice

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CheckProductFulfillmentRuleBindingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckProductFulfillmentRuleBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckProductFulfillmentRuleBindingLogic {
	return &CheckProductFulfillmentRuleBindingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckProductFulfillmentRuleBindingLogic) CheckProductFulfillmentRuleBinding(in *smsclient.CheckProductFulfillmentRuleBindingReq) (*smsclient.CheckProductFulfillmentRuleBindingResp, error) {
	// 1. 验证参数
	if in.Id <= 0 {
		return nil, errors.New("规则ID无效")
	}

	// 2. 解析治理范围
	platformId := int64(1)
	tenantId := int64(0)
	merchantId := int64(0)
	if in.Scope != nil {
		platformId = in.Scope.PlatformId
		tenantId = in.Scope.TenantId
		merchantId = in.Scope.MerchantId
	}

	// 3. 检查规则是否存在
	var count int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Count(&count).Error
	if err != nil {
		logc.Errorf(l.ctx, "检查发卡规则是否存在失败: %v", err)
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("发卡规则不存在")
	}

	// 4. 查询绑定商品数量
	type bindingRow struct {
		RuleId       int64 `gorm:"column:rule_id"`
		ProductSpuId int64 `gorm:"column:product_spu_id"`
		ProductSkuId int64 `gorm:"column:product_sku_id"`
	}

	var bindings []bindingRow
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_binding").
		Select("rule_id, product_spu_id, product_sku_id").
		Where("rule_id = ? AND is_deleted = 0", in.Id).
		Find(&bindings).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询规则绑定商品失败: %v", err)
		return nil, err
	}

	bindingCount := len(bindings)

	// 5. 查询绑定的商品名称
	var bindingProductNames []string
	if bindingCount > 0 {
		// 收集所有绑定的商品ID
		var spuIds []int64
		for _, b := range bindings {
			if b.ProductSpuId > 0 {
				spuIds = append(spuIds, b.ProductSpuId)
			}
		}

		// 查询商品名称
		if len(spuIds) > 0 {
			type productRow struct {
				Name string `gorm:"column:name"`
			}
			var products []productRow
			err = l.svcCtx.DB.WithContext(l.ctx).
				Table("pms_product_spu").
				Select("name").
				Where("id IN ? AND is_deleted = 0", spuIds).
				Find(&products).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				logc.Errorf(l.ctx, "查询商品名称失败: %v", err)
			}
			for _, p := range products {
				bindingProductNames = append(bindingProductNames, p.Name)
			}
		}
	}

	// 6. 判断是否可以禁用/删除
	canDisable := bindingCount == 0
	canDelete := bindingCount == 0

	message := "该规则未绑定任何商品，可以禁用或删除"
	if bindingCount > 0 {
		message = "该规则已绑定商品，无法禁用或删除"
	}

	return &smsclient.CheckProductFulfillmentRuleBindingResp{
		BindingCount:        int32(bindingCount),
		BindingProductNames: bindingProductNames,
		CanDisable:          canDisable,
		CanDelete:           canDelete,
		Message:             message,
	}, nil
}
