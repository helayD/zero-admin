package home

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type RecordHomeAdvertiseClickLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordHomeAdvertiseClickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordHomeAdvertiseClickLogic {
	return &RecordHomeAdvertiseClickLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordHomeAdvertiseClickLogic) RecordHomeAdvertiseClick(req *types.RecordHomeAdvertiseClickReq) (resp *types.RecordHomeAdvertiseClickResp, err error) {
	if req.AdvertiseId <= 0 {
		return nil, errorx.NewDefaultError("首页广告ID不能为空")
	}

	detail, err := l.svcCtx.HomeAdvertiseService.QueryHomeAdvertiseDetail(l.ctx, &smsclient.QueryHomeAdvertiseDetailReq{
		Id: req.AdvertiseId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询首页广告详情失败, advertiseId=%d, err=%s", req.AdvertiseId, err.Error())
		s, _ := status.FromError(err)
		message := strings.TrimSpace(s.Message())
		if message == "" {
			message = err.Error()
		}
		return nil, errorx.NewDefaultError(message)
	}

	scope := frontcommon.ResolveEffectiveGovernanceScope(l.ctx)
	channel := operatefunnel.ChannelApp
	if detail.Type == 0 {
		channel = operatefunnel.ChannelPC
	}

	_, err = l.svcCtx.OperateDashboardService.RecordOperateFunnelEvent(l.ctx, &smsclient.RecordOperateFunnelEventReq{
		Scope:        frontcommon.SMSGovernanceScope(scope),
		EventType:    operatefunnel.EventClick,
		Channel:      channel,
		ActivityType: operatefunnel.ActivityHomeAdvertise,
		ActivityId:   detail.Id,
		MemberId:     tryGetOptionalMemberID(l.ctx),
		TraceId:      strings.TrimSpace(req.TraceId),
		ExtraJson:    fmt.Sprintf(`{"advertiseId":%d,"advertiseName":%q}`, detail.Id, detail.Name),
	})
	if err != nil {
		logc.Errorf(l.ctx, "记录首页广告点击失败, advertiseId=%d, err=%s", req.AdvertiseId, err.Error())
		s, _ := status.FromError(err)
		message := strings.TrimSpace(s.Message())
		if message == "" {
			message = err.Error()
		}
		return nil, errorx.NewDefaultError(message)
	}

	return &types.RecordHomeAdvertiseClickResp{
		Code:    0,
		Message: "记录首页广告点击成功",
	}, nil
}

func tryGetOptionalMemberID(ctx context.Context) int64 {
	rawMemberID := ctx.Value("memberId")
	jsonNumber, ok := rawMemberID.(json.Number)
	if !ok {
		return 0
	}

	memberID, err := jsonNumber.Int64()
	if err != nil {
		return 0
	}

	return memberID
}
