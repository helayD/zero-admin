package logic

import (
	"context"

	"github.com/feihua/zero-admin/job/internal/jobs"
	"github.com/feihua/zero-admin/job/internal/svc"
	"github.com/feihua/zero-admin/job/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type JobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JobLogic {
	return &JobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JobLogic) Job(req *types.Request) (resp *types.Response, err error) {
	switch req.Name {
	case "cancel_timeout_order":
		l.CancelTimeOutOrderJob()
	case "handle_card_mint_timeout":
		l.HandleCardMintTimeout()
	case "reconcile_order_card_fulfillment":
		l.ReconcileOrderCardFulfillment()
	default:
		logx.Errorf("unknown job name: %s", req.Name)
	}

	return &types.Response{
		Message: "job executed",
	}, nil
}

func (l *JobLogic) CancelTimeOutOrderJob() {
	jobs.CancelTimeOutOrder(l.ctx, l.svcCtx.Redis, l.svcCtx.ProductSkuService, l.svcCtx.OrderService,
		l.svcCtx.CouponRecordService, l.svcCtx.MemberService, l.svcCtx.OrderSettingService)
}

func (l *JobLogic) HandleCardMintTimeout() {
	jobs.HandleCardMintTimeout(l.ctx, l.svcCtx.CardMintService)
}

func (l *JobLogic) ReconcileOrderCardFulfillment() {
	jobs.ReconcileOrderCardFulfillment(l.ctx, l.svcCtx.DB, l.svcCtx.CardMintService)
}
