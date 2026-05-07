package productfulfillmentruleservice

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductFulfillmentRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductFulfillmentRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductFulfillmentRuleLogic {
	return &DeleteProductFulfillmentRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteProductFulfillmentRuleLogic) DeleteProductFulfillmentRule(in *smsclient.DeleteProductFulfillmentRuleReq) (*smsclient.DeleteProductFulfillmentRuleResp, error) {
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

	// 4. 检查是否绑定商品
	var bindingCount int64
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_binding").
		Where("rule_id = ? AND is_deleted = 0", in.Id).
		Count(&bindingCount).Error
	if err != nil {
		logc.Errorf(l.ctx, "检查规则绑定商品失败: %v", err)
		return nil, err
	}
	if bindingCount > 0 {
		return nil, errors.New("该规则已绑定商品，无法删除")
	}

	// 5. 执行软删除
	now := time.Now()
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Updates(map[string]interface{}{
			"is_deleted":  1,
			"update_time": now,
		}).Error
	if err != nil {
		logc.Errorf(l.ctx, "删除发卡规则失败: %v", err)
		return nil, errors.New("删除发卡规则失败")
	}

	return &smsclient.DeleteProductFulfillmentRuleResp{
		Code:    0,
		Message: "删除成功",
	}, nil
}
