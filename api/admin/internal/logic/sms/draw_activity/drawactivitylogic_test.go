package draw_activity

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/client/drawactivityservice"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"google.golang.org/grpc"
)

type mockDrawActivityService struct {
	drawactivityservice.DrawActivityService
	addFn     func(context.Context, *smsclient.AddDrawActivityReq, ...grpc.CallOption) (*smsclient.AddDrawActivityResp, error)
	listFn    func(context.Context, *smsclient.QueryDrawActivityListReq, ...grpc.CallOption) (*smsclient.QueryDrawActivityListResp, error)
	previewFn func(context.Context, *smsclient.PreviewDrawActivityPublishReadinessReq, ...grpc.CallOption) (*smsclient.PreviewDrawActivityPublishReadinessResp, error)
}

func (m *mockDrawActivityService) AddDrawActivity(ctx context.Context, in *smsclient.AddDrawActivityReq, opts ...grpc.CallOption) (*smsclient.AddDrawActivityResp, error) {
	return m.addFn(ctx, in, opts...)
}

func (m *mockDrawActivityService) QueryDrawActivityList(ctx context.Context, in *smsclient.QueryDrawActivityListReq, opts ...grpc.CallOption) (*smsclient.QueryDrawActivityListResp, error) {
	return m.listFn(ctx, in, opts...)
}

func (m *mockDrawActivityService) PreviewDrawActivityPublishReadiness(ctx context.Context, in *smsclient.PreviewDrawActivityPublishReadinessReq, opts ...grpc.CallOption) (*smsclient.PreviewDrawActivityPublishReadinessResp, error) {
	return m.previewFn(ctx, in, opts...)
}

func TestAddDrawActivityPassesResolvedWriteScope(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("88"))
	ctx = context.WithValue(ctx, "userId", json.Number("1001"))
	ctx = context.WithValue(ctx, "userName", "tester")

	logic := NewAddDrawActivityLogic(ctx, &svc.ServiceContext{
		DrawActivityService: &mockDrawActivityService{
			addFn: func(_ context.Context, in *smsclient.AddDrawActivityReq, _ ...grpc.CallOption) (*smsclient.AddDrawActivityResp, error) {
				if in.Scope.ScopeType != "merchant" || in.Scope.TenantId != 10 || in.Scope.MerchantId != 88 {
					t.Fatalf("unexpected scope: %+v", in.Scope)
				}
				if in.CreateBy != 1001 || in.OperatorName != "tester" {
					t.Fatalf("unexpected operator payload: %+v", in)
				}
				return &smsclient.AddDrawActivityResp{Id: 1}, nil
			},
		},
	})

	_, err := logic.AddDrawActivity(&types.AddDrawActivityReq{
		ActivityCode:                "DRAW-TEST",
		Name:                        "抽卡活动",
		StartTime:                   "2026-04-16 10:00:00",
		EndTime:                     "2026-04-30 23:59:59",
		ParticipantConditionSummary: "条件",
		ConsumeRuleSummary:          "规则",
		ProbabilityRule:             "总和为1",
		ComplianceRuleSummary:       "合规",
		CirculationLimitSummary:     "禁止集中竞价；禁止连续挂牌；禁止收益承诺",
		ApprovalRecordRef:           "APPROVAL-1",
		CopyrightStatus:             2,
		ContentAuditStatus:          2,
		AuditStatus:                 2,
		IsEnabled:                   1,
		Templates: []types.DrawCardTemplateData{
			{TemplateName: "模板", TemplateCode: "TPL-1", IssueLimit: 100},
		},
		Pools: []types.DrawPoolData{
			{
				PoolName: "卡池",
				PoolCode: "POOL-1",
				Templates: []types.DrawPoolTemplateData{
					{TemplateCode: "TPL-1", Probability: 1, SaleLimit: 100, RemainingLimit: 100, ConfigLimit: 100},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("AddDrawActivity returned error: %v", err)
	}
}

func TestQueryDrawActivityListPassesResolvedQueryScope(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "tenant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("0"))

	logic := NewQueryDrawActivityListLogic(ctx, &svc.ServiceContext{
		DrawActivityService: &mockDrawActivityService{
			listFn: func(_ context.Context, in *smsclient.QueryDrawActivityListReq, _ ...grpc.CallOption) (*smsclient.QueryDrawActivityListResp, error) {
				if in.Scope.ScopeType != "tenant" || in.Scope.TenantId != 10 || in.Scope.MerchantId != 0 {
					t.Fatalf("unexpected query scope: %+v", in.Scope)
				}
				return &smsclient.QueryDrawActivityListResp{
					Total: 1,
					List: []*smsclient.DrawActivityListData{
						{Id: 1, Name: "抽卡活动", ActivityCode: "DRAW-1", ScopeType: "tenant", TenantId: 10},
					},
				}, nil
			},
		},
	})

	resp, err := logic.QueryDrawActivityList(&types.QueryDrawActivityListReq{Current: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("QueryDrawActivityList returned error: %v", err)
	}
	if resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", resp)
	}
}

func TestPreviewDrawActivityPublishReadinessMapsStructuredItems(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "platform")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("0"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("0"))

	logic := NewPreviewDrawActivityPublishReadinessLogic(ctx, &svc.ServiceContext{
		DrawActivityService: &mockDrawActivityService{
			previewFn: func(_ context.Context, in *smsclient.PreviewDrawActivityPublishReadinessReq, _ ...grpc.CallOption) (*smsclient.PreviewDrawActivityPublishReadinessResp, error) {
				if in.Scope.ScopeType != "platform" {
					t.Fatalf("unexpected preview scope: %+v", in.Scope)
				}
				return &smsclient.PreviewDrawActivityPublishReadinessResp{
					ReadyToPublish:   false,
					PublishReadiness: 0,
					ReadinessLabel:   "缺少信息",
					Summary:          "审批记录缺失",
					Items: []*smsclient.DrawReadinessItem{
						{Code: "missingAuditRecord", Field: "approvalRecordRef", Message: "审批记录缺失", Blocking: true},
					},
				}, nil
			},
		},
	})

	resp, err := logic.PreviewDrawActivityPublishReadiness(&types.PreviewDrawActivityPublishReadinessReq{Id: 1})
	if err != nil {
		t.Fatalf("PreviewDrawActivityPublishReadiness returned error: %v", err)
	}
	if resp.Data.ReadyToPublish || len(resp.Data.Items) != 1 || resp.Data.Items[0].Code != "missingAuditRecord" {
		t.Fatalf("unexpected preview response: %+v", resp)
	}
}
