package operatedashboardservicelogic

import (
	"context"
	"fmt"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryOperateActivityOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryOperateActivityOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOperateActivityOptionsLogic {
	return &QueryOperateActivityOptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryOperateActivityOptionsLogic) QueryOperateActivityOptions(in *smsclient.QueryOperateActivityOptionsReq) (*smsclient.QueryOperateActivityOptionsResp, error) {
	scope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	list := []*smsclient.OperateFunnelActivityOption{
		{
			ActivityType:  operatefunnel.ActivityNone,
			ActivityId:    0,
			ActivityName:  "全部非活动",
			ActivityLabel: "none / 全部非活动",
		},
	}

	rows := make([]activityOptionRow, 0)
	queryText, queryArgs := buildOperateActivityOptionsQuery(scope)
	if err = l.svcCtx.DB.WithContext(l.ctx).Raw(queryText, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		list = append(list, &smsclient.OperateFunnelActivityOption{
			ActivityType:  row.ActivityType,
			ActivityId:    row.ActivityID,
			ActivityName:  row.ActivityName,
			ActivityLabel: row.ActivityLabel,
		})
	}

	return &smsclient.QueryOperateActivityOptionsResp{List: list}, nil
}

func buildOperateActivityOptionsQuery(scope pkgscope.GovernanceScope) (string, []interface{}) {
	eventScopeSQL, eventScopeArgs := scopeClause("e", scope)
	cartScopeSQL, cartScopeArgs := scopeClause("spu", scope)
	orderScopeSQL, orderScopeArgs := scopeClause("o", scope)
	couponOrderScopeSQL, couponOrderScopeArgs := scopeClause("o2", scope)
	couponFallbackScopeSQL, couponFallbackScopeArgs := scopeClause("c", scope)

	queryText := fmt.Sprintf(`
		SELECT
			refs.activity_type,
			refs.activity_id,
			CASE refs.activity_type
				WHEN 'home_advertise' THEN COALESCE(ha.name, CONCAT('首页广告#', refs.activity_id))
				WHEN 'coupon' THEN COALESCE(cp.name, CONCAT('优惠券#', refs.activity_id))
				WHEN 'seckill_activity' THEN COALESCE(sa.name, CONCAT('秒杀活动#', refs.activity_id))
				ELSE CONCAT(refs.activity_type, '#', refs.activity_id)
			END AS activity_name,
			CASE refs.activity_type
				WHEN 'home_advertise' THEN CONCAT('home_advertise / ', COALESCE(ha.name, CONCAT('首页广告#', refs.activity_id)))
				WHEN 'coupon' THEN CONCAT('coupon / ', COALESCE(cp.name, CONCAT('优惠券#', refs.activity_id)))
				WHEN 'seckill_activity' THEN CONCAT('seckill_activity / ', COALESCE(sa.name, CONCAT('秒杀活动#', refs.activity_id)))
				ELSE CONCAT(refs.activity_type, ' / #', refs.activity_id)
			END AS activity_label
		FROM (
			SELECT DISTINCT
				e.activity_type,
				e.activity_id
			FROM sms_operate_funnel_event e
			WHERE COALESCE(NULLIF(e.activity_type, ''), 'none') <> 'none'
			  AND e.activity_id > 0
			  AND %s

			UNION

			SELECT DISTINCT
				c.activity_type,
				c.activity_id
			FROM oms_cart_item c
			JOIN pms_product_spu spu ON spu.id = c.product_id AND spu.is_deleted = 0
			WHERE c.delete_status = 0
			  AND COALESCE(NULLIF(c.activity_type, ''), 'none') <> 'none'
			  AND c.activity_id > 0
			  AND %s

			UNION

			SELECT DISTINCT
				o.activity_type,
				o.activity_id
			FROM oms_order_main o
			WHERE o.is_deleted = 0
			  AND COALESCE(NULLIF(o.activity_type, ''), 'none') <> 'none'
			  AND o.activity_id > 0
			  AND %s

			UNION

			SELECT DISTINCT
				CASE
					WHEN COALESCE(NULLIF(o2.activity_type, ''), 'none') <> 'none' THEN o2.activity_type
					WHEN c.id IS NOT NULL THEN 'coupon'
					ELSE 'none'
				END AS activity_type,
				CASE
					WHEN COALESCE(NULLIF(o2.activity_type, ''), 'none') <> 'none' THEN o2.activity_id
					WHEN c.id IS NOT NULL THEN c.id
					ELSE 0
				END AS activity_id
			FROM sms_coupon_record cr
			LEFT JOIN oms_order_main o2 ON o2.id = cr.order_id AND o2.is_deleted = 0
			LEFT JOIN sms_coupon c ON c.id = cr.coupon_id AND c.is_deleted = 0
			WHERE cr.status = 1
			  AND cr.use_time IS NOT NULL
			  AND (
			  	(o2.id IS NOT NULL AND %s)
			  	OR (o2.id IS NULL AND %s)
			  )
		) refs
		LEFT JOIN sms_home_advertise ha
			ON refs.activity_type = 'home_advertise' AND ha.id = refs.activity_id
		LEFT JOIN sms_coupon cp
			ON refs.activity_type = 'coupon' AND cp.id = refs.activity_id
		LEFT JOIN sms_seckill_activity sa
			ON refs.activity_type = 'seckill_activity' AND sa.id = refs.activity_id
		WHERE refs.activity_type <> 'none'
		  AND refs.activity_id > 0
		ORDER BY refs.activity_type ASC, refs.activity_id DESC
	`, eventScopeSQL, cartScopeSQL, orderScopeSQL, couponOrderScopeSQL, couponFallbackScopeSQL)

	queryArgs := append([]interface{}{}, eventScopeArgs...)
	queryArgs = append(queryArgs, cartScopeArgs...)
	queryArgs = append(queryArgs, orderScopeArgs...)
	queryArgs = append(queryArgs, couponOrderScopeArgs...)
	queryArgs = append(queryArgs, couponFallbackScopeArgs...)
	return queryText, queryArgs
}
