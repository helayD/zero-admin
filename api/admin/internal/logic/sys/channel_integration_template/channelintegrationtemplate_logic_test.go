package channel_integration_template

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	channelintegrationtemplateservice "github.com/feihua/zero-admin/rpc/sys/client/channelintegrationtemplateservice"
	"google.golang.org/grpc"
)

type fakeChannelIntegrationTemplateService struct {
	createFn       func(context.Context, *channelintegrationtemplateservice.CreateChannelIntegrationTemplateReq, ...grpc.CallOption) (*channelintegrationtemplateservice.CreateChannelIntegrationTemplateResp, error)
	updateFn       func(context.Context, *channelintegrationtemplateservice.UpdateChannelIntegrationTemplateReq, ...grpc.CallOption) (*channelintegrationtemplateservice.UpdateChannelIntegrationTemplateResp, error)
	updateStatusFn func(context.Context, *channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusReq, ...grpc.CallOption) (*channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusResp, error)
	queryDetailFn  func(context.Context, *channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailReq, ...grpc.CallOption) (*channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailResp, error)
	queryListFn    func(context.Context, *channelintegrationtemplateservice.QueryChannelIntegrationTemplateListReq, ...grpc.CallOption) (*channelintegrationtemplateservice.QueryChannelIntegrationTemplateListResp, error)
}

func (f *fakeChannelIntegrationTemplateService) CreateChannelIntegrationTemplate(ctx context.Context, in *channelintegrationtemplateservice.CreateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*channelintegrationtemplateservice.CreateChannelIntegrationTemplateResp, error) {
	if f.createFn == nil {
		return &channelintegrationtemplateservice.CreateChannelIntegrationTemplateResp{}, nil
	}
	return f.createFn(ctx, in, opts...)
}

func (f *fakeChannelIntegrationTemplateService) UpdateChannelIntegrationTemplate(ctx context.Context, in *channelintegrationtemplateservice.UpdateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*channelintegrationtemplateservice.UpdateChannelIntegrationTemplateResp, error) {
	if f.updateFn == nil {
		return &channelintegrationtemplateservice.UpdateChannelIntegrationTemplateResp{}, nil
	}
	return f.updateFn(ctx, in, opts...)
}

func (f *fakeChannelIntegrationTemplateService) UpdateChannelIntegrationTemplateStatus(ctx context.Context, in *channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusReq, opts ...grpc.CallOption) (*channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusResp, error) {
	if f.updateStatusFn == nil {
		return &channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusResp{}, nil
	}
	return f.updateStatusFn(ctx, in, opts...)
}

func (f *fakeChannelIntegrationTemplateService) QueryChannelIntegrationTemplateDetail(ctx context.Context, in *channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailReq, opts ...grpc.CallOption) (*channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailResp, error) {
	if f.queryDetailFn == nil {
		return &channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailResp{}, nil
	}
	return f.queryDetailFn(ctx, in, opts...)
}

func (f *fakeChannelIntegrationTemplateService) QueryChannelIntegrationTemplateList(ctx context.Context, in *channelintegrationtemplateservice.QueryChannelIntegrationTemplateListReq, opts ...grpc.CallOption) (*channelintegrationtemplateservice.QueryChannelIntegrationTemplateListResp, error) {
	if f.queryListFn == nil {
		return &channelintegrationtemplateservice.QueryChannelIntegrationTemplateListResp{}, nil
	}
	return f.queryListFn(ctx, in, opts...)
}

func channelIntegrationTemplateTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userName", "platform-admin")
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	ctx = context.WithValue(ctx, "scopeType", "platform")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	return ctx
}

func TestCreateChannelIntegrationTemplateLogicPassesOperatorAndJSONPayload(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ChannelIntegrationTemplateService: &fakeChannelIntegrationTemplateService{
			createFn: func(_ context.Context, in *channelintegrationtemplateservice.CreateChannelIntegrationTemplateReq, _ ...grpc.CallOption) (*channelintegrationtemplateservice.CreateChannelIntegrationTemplateResp, error) {
				if in.CreateBy != "platform-admin" {
					t.Fatalf("unexpected createBy: %+v", in)
				}
				if in.TemplateCode != "mini_default" || in.TargetCode != "mini-program" {
					t.Fatalf("unexpected template payload: %+v", in)
				}
				if in.MetadataConfig != `{"channelCode":"mini_program"}` {
					t.Fatalf("unexpected metadata config: %s", in.MetadataConfig)
				}
				return &channelintegrationtemplateservice.CreateChannelIntegrationTemplateResp{Id: 11}, nil
			},
		},
	}

	logic := NewCreateChannelIntegrationTemplateLogic(channelIntegrationTemplateTestContext(), svcCtx)
	resp, err := logic.CreateChannelIntegrationTemplate(&types.CreateChannelIntegrationTemplateReq{
		TemplateCode:   "mini_default",
		TemplateName:   "小程序模板",
		TemplateType:   "channel",
		TargetCode:     "mini-program",
		MetadataConfig: json.RawMessage("{\n  \"channelCode\": \"mini_program\"\n}"),
	})
	if err != nil {
		t.Fatalf("CreateChannelIntegrationTemplate returned error: %v", err)
	}
	if resp.Code != "000000" || resp.Data.Id != 11 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestQueryChannelIntegrationTemplateListLogicMapsRPCResponse(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ChannelIntegrationTemplateService: &fakeChannelIntegrationTemplateService{
			queryListFn: func(_ context.Context, in *channelintegrationtemplateservice.QueryChannelIntegrationTemplateListReq, _ ...grpc.CallOption) (*channelintegrationtemplateservice.QueryChannelIntegrationTemplateListResp, error) {
				if in.TargetCode != "mini-program" || in.TemplateType != "channel" || in.PageNum != 2 {
					t.Fatalf("unexpected query params: %+v", in)
				}
				return &channelintegrationtemplateservice.QueryChannelIntegrationTemplateListResp{
					Total: 1,
					List: []*channelintegrationtemplateservice.ChannelIntegrationTemplateData{
						{
							Id:                   1,
							TemplateCode:         "mini_default",
							TemplateName:         "小程序模板",
							TemplateType:         "channel",
							TargetCode:           "mini_program",
							Status:               "enabled",
							MetadataConfig:       `{"channelCode":"mini_program"}`,
							SecretRefConfig:      `{"appSecretRef":"credential://sys/channel/mini_program/app-secret"}`,
							IntentContractConfig: `{"home":{"intent":"home","routeKey":"home"}}`,
							ImpactScopeConfig:    `{"subjectTypes":["tenant"]}`,
						},
					},
				}, nil
			},
		},
	}

	logic := NewQueryChannelIntegrationTemplateListLogic(channelIntegrationTemplateTestContext(), svcCtx)
	resp, err := logic.QueryChannelIntegrationTemplateList(&types.QueryChannelIntegrationTemplateListReq{
		Current:      2,
		PageSize:     20,
		TemplateType: "channel",
		TargetCode:   "mini-program",
		Status:       "enabled",
	})
	if err != nil {
		t.Fatalf("QueryChannelIntegrationTemplateList returned error: %v", err)
	}
	if !resp.Success || resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", resp)
	}
	if string(resp.Data[0].MetadataConfig) != `{"channelCode":"mini_program"}` {
		t.Fatalf("unexpected mapped metadata config: %s", string(resp.Data[0].MetadataConfig))
	}
}

func TestUpdateChannelIntegrationTemplateStatusLogicPassesOperator(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		ChannelIntegrationTemplateService: &fakeChannelIntegrationTemplateService{
			queryDetailFn: func(_ context.Context, in *channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailReq, _ ...grpc.CallOption) (*channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailResp, error) {
				return &channelintegrationtemplateservice.QueryChannelIntegrationTemplateDetailResp{Data: &channelintegrationtemplateservice.ChannelIntegrationTemplateData{
					Id:         in.Id,
					ScopeType:  "platform",
					PlatformId: 1,
				}}, nil
			},
			updateStatusFn: func(_ context.Context, in *channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusReq, _ ...grpc.CallOption) (*channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusResp, error) {
				if in.UpdateBy != "platform-admin" {
					t.Fatalf("unexpected operator info: %+v", in)
				}
				if len(in.Ids) != 2 || in.Status != "disabled" {
					t.Fatalf("unexpected status request: %+v", in)
				}
				return &channelintegrationtemplateservice.UpdateChannelIntegrationTemplateStatusResp{Pong: "ok"}, nil
			},
		},
	}

	logic := NewUpdateChannelIntegrationTemplateStatusLogic(channelIntegrationTemplateTestContext(), svcCtx)
	resp, err := logic.UpdateChannelIntegrationTemplateStatus(&types.UpdateChannelIntegrationTemplateStatusReq{
		Ids:    []int64{1, 2},
		Status: "disabled",
	})
	if err != nil {
		t.Fatalf("UpdateChannelIntegrationTemplateStatus returned error: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
