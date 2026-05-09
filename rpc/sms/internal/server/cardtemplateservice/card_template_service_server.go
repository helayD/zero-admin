package cardtemplateservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/internal/logic/cardtemplateservice"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

type CardTemplateServiceServer struct {
	svcCtx *svc.ServiceContext
	smsclient.UnimplementedCardTemplateServiceServer
}

func NewCardTemplateServiceServer(svcCtx *svc.ServiceContext) *CardTemplateServiceServer {
	return &CardTemplateServiceServer{svcCtx: svcCtx}
}

func (s *CardTemplateServiceServer) AddCardTemplate(ctx context.Context, in *smsclient.AddCardTemplateReq) (*smsclient.AddCardTemplateResp, error) {
	l := cardtemplateservice.NewAddCardTemplateLogic(ctx, s.svcCtx)
	return l.AddCardTemplate(in)
}

func (s *CardTemplateServiceServer) UpdateCardTemplate(ctx context.Context, in *smsclient.UpdateCardTemplateReq) (*smsclient.UpdateCardTemplateResp, error) {
	l := cardtemplateservice.NewUpdateCardTemplateLogic(ctx, s.svcCtx)
	return l.UpdateCardTemplate(in)
}

func (s *CardTemplateServiceServer) QueryCardTemplateList(ctx context.Context, in *smsclient.QueryCardTemplateListReq) (*smsclient.QueryCardTemplateListResp, error) {
	l := cardtemplateservice.NewQueryCardTemplateListLogic(ctx, s.svcCtx)
	return l.QueryCardTemplateList(in)
}

func (s *CardTemplateServiceServer) QueryCardTemplateDetail(ctx context.Context, in *smsclient.QueryCardTemplateDetailReq) (*smsclient.QueryCardTemplateDetailResp, error) {
	l := cardtemplateservice.NewQueryCardTemplateDetailLogic(ctx, s.svcCtx)
	return l.QueryCardTemplateDetail(in)
}

func (s *CardTemplateServiceServer) UpdateCardTemplateStatus(ctx context.Context, in *smsclient.UpdateCardTemplateStatusReq) (*smsclient.UpdateCardTemplateStatusResp, error) {
	l := cardtemplateservice.NewUpdateCardTemplateStatusLogic(ctx, s.svcCtx)
	return l.UpdateCardTemplateStatus(in)
}

func (s *CardTemplateServiceServer) DeleteCardTemplate(ctx context.Context, in *smsclient.DeleteCardTemplateReq) (*smsclient.DeleteCardTemplateResp, error) {
	l := cardtemplateservice.NewDeleteCardTemplateLogic(ctx, s.svcCtx)
	return l.DeleteCardTemplate(in)
}

func (s *CardTemplateServiceServer) CheckCardTemplateUsage(ctx context.Context, in *smsclient.CheckCardTemplateUsageReq) (*smsclient.CheckCardTemplateUsageResp, error) {
	l := cardtemplateservice.NewCheckCardTemplateUsageLogic(ctx, s.svcCtx)
	return l.CheckCardTemplateUsage(in)
}
