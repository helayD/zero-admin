package operatedashboardservicelogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordOperateFunnelEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordOperateFunnelEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordOperateFunnelEventLogic {
	return &RecordOperateFunnelEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RecordOperateFunnelEventLogic) RecordOperateFunnelEvent(in *smsclient.RecordOperateFunnelEventReq) (*smsclient.RecordOperateFunnelEventResp, error) {
	scope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	statTime := time.Now().In(operatefunnel.Location())
	if strings.TrimSpace(in.StatTime) != "" {
		statTime, err = operatefunnel.ParseTime(in.StatTime)
		if err != nil {
			return nil, err
		}
	}

	traceID := strings.TrimSpace(in.TraceId)
	if traceID == "" {
		traceID = fmt.Sprintf(
			"%s:%d:%d:%d:%d:%d:%d",
			in.EventType,
			in.ActivityId,
			in.MemberId,
			in.ProductId,
			in.OrderId,
			in.CouponId,
			statTime.UnixNano(),
		)
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Exec(`
		INSERT INTO sms_operate_funnel_event (
			event_type, stat_time, bucket_date, bucket_hour,
			platform_id, tenant_id, merchant_id,
			channel, activity_type, activity_id,
			member_id, product_id, order_id, coupon_id,
			trace_id, extra_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			extra_json = VALUES(extra_json)`,
		in.EventType,
		statTime,
		operatefunnel.BucketStart(statTime, operatefunnel.BucketDay),
		operatefunnel.BucketStart(statTime, operatefunnel.BucketHour),
		scope.PlatformID,
		scope.TenantID,
		scope.MerchantID,
		operatefunnel.NormalizeChannel(in.Channel),
		operatefunnel.NormalizeActivityType(in.ActivityType),
		in.ActivityId,
		in.MemberId,
		in.ProductId,
		in.OrderId,
		in.CouponId,
		traceID,
		strings.TrimSpace(in.ExtraJson),
	).Error
	if err != nil {
		return nil, err
	}

	return &smsclient.RecordOperateFunnelEventResp{}, nil
}
