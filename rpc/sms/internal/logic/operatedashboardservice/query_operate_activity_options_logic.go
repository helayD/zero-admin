package operatedashboardservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
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

	advertiseRows := make([]activityOptionRow, 0)
	if err = l.svcCtx.DB.WithContext(l.ctx).Raw(`
		SELECT DISTINCT
			'home_advertise' AS activity_type,
			e.activity_id AS activity_id,
			COALESCE(a.name, CONCAT('首页广告#', e.activity_id)) AS activity_name,
			CONCAT('home_advertise / ', COALESCE(a.name, CONCAT('首页广告#', e.activity_id))) AS activity_label
		FROM sms_operate_funnel_event e
		LEFT JOIN sms_home_advertise a ON a.id = e.activity_id
		WHERE e.activity_type = 'home_advertise'
		  AND e.platform_id = ? AND e.tenant_id = ? AND e.merchant_id = ?
		ORDER BY e.activity_id DESC
	`, scopeArgs(scope)...).Scan(&advertiseRows).Error; err != nil {
		return nil, err
	}

	couponRows := make([]activityOptionRow, 0)
	if err = l.svcCtx.DB.WithContext(l.ctx).Raw(`
		SELECT
			'coupon' AS activity_type,
			id AS activity_id,
			name AS activity_name,
			CONCAT('coupon / ', name) AS activity_label
		FROM sms_coupon
		WHERE is_deleted = 0
		  AND platform_id = ? AND tenant_id = ? AND merchant_id = ?
		ORDER BY id DESC
	`, scopeArgs(scope)...).Scan(&couponRows).Error; err != nil {
		return nil, err
	}

	seckillRows := make([]activityOptionRow, 0)
	if err = l.svcCtx.DB.WithContext(l.ctx).Raw(`
		SELECT
			'seckill_activity' AS activity_type,
			id AS activity_id,
			name AS activity_name,
			CONCAT('seckill_activity / ', name) AS activity_label
		FROM sms_seckill_activity
		WHERE is_deleted = 0
		  AND platform_id = ? AND tenant_id = ? AND merchant_id = ?
		ORDER BY id DESC
	`, scopeArgs(scope)...).Scan(&seckillRows).Error; err != nil {
		return nil, err
	}

	for _, row := range append(append(advertiseRows, couponRows...), seckillRows...) {
		list = append(list, &smsclient.OperateFunnelActivityOption{
			ActivityType:  row.ActivityType,
			ActivityId:    row.ActivityID,
			ActivityName:  row.ActivityName,
			ActivityLabel: row.ActivityLabel,
		})
	}

	return &smsclient.QueryOperateActivityOptionsResp{List: list}, nil
}
