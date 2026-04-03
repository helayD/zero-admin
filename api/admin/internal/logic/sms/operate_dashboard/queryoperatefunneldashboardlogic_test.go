package operate_dashboard

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/client/cartitemservice"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/sms/client/operatedashboardservice"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"google.golang.org/grpc"
)

type mockOperateDashboardService struct {
	operatedashboardservice.OperateDashboardService
	queryTrafficFunnelFn   func(context.Context, *smsclient.QueryOperateTrafficFunnelReq, ...grpc.CallOption) (*smsclient.QueryOperateTrafficFunnelResp, error)
	queryCouponRedeemFn    func(context.Context, *smsclient.QueryOperateCouponRedeemReq, ...grpc.CallOption) (*smsclient.QueryOperateCouponRedeemResp, error)
	queryActivityOptionsFn func(context.Context, *smsclient.QueryOperateActivityOptionsReq, ...grpc.CallOption) (*smsclient.QueryOperateActivityOptionsResp, error)
}

func (m *mockOperateDashboardService) QueryOperateTrafficFunnel(ctx context.Context, in *smsclient.QueryOperateTrafficFunnelReq, opts ...grpc.CallOption) (*smsclient.QueryOperateTrafficFunnelResp, error) {
	return m.queryTrafficFunnelFn(ctx, in, opts...)
}

func (m *mockOperateDashboardService) QueryOperateCouponRedeem(ctx context.Context, in *smsclient.QueryOperateCouponRedeemReq, opts ...grpc.CallOption) (*smsclient.QueryOperateCouponRedeemResp, error) {
	return m.queryCouponRedeemFn(ctx, in, opts...)
}

func (m *mockOperateDashboardService) QueryOperateActivityOptions(ctx context.Context, in *smsclient.QueryOperateActivityOptionsReq, opts ...grpc.CallOption) (*smsclient.QueryOperateActivityOptionsResp, error) {
	return m.queryActivityOptionsFn(ctx, in, opts...)
}

type mockCartItemService struct {
	cartitemservice.CartItemService
	queryOperateCartFunnelFn func(context.Context, *omsclient.QueryOperateCartFunnelReq, ...grpc.CallOption) (*omsclient.QueryOperateCartFunnelResp, error)
}

func (m *mockCartItemService) QueryOperateCartFunnel(ctx context.Context, in *omsclient.QueryOperateCartFunnelReq, opts ...grpc.CallOption) (*omsclient.QueryOperateCartFunnelResp, error) {
	return m.queryOperateCartFunnelFn(ctx, in, opts...)
}

type mockOrderService struct {
	orderservice.OrderService
	queryOperateOrderFunnelFn func(context.Context, *omsclient.QueryOperateOrderFunnelReq, ...grpc.CallOption) (*omsclient.QueryOperateOrderFunnelResp, error)
}

func (m *mockOrderService) QueryOperateOrderFunnel(ctx context.Context, in *omsclient.QueryOperateOrderFunnelReq, opts ...grpc.CallOption) (*omsclient.QueryOperateOrderFunnelResp, error) {
	return m.queryOperateOrderFunnelFn(ctx, in, opts...)
}

func TestQueryOperateFunnelDashboardMergesRPCResponses(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "platform")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("0"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("0"))

	logic := NewQueryOperateFunnelDashboardLogic(ctx, &svc.ServiceContext{
		OperateDashboardService: &mockOperateDashboardService{
			queryTrafficFunnelFn: func(_ context.Context, in *smsclient.QueryOperateTrafficFunnelReq, _ ...grpc.CallOption) (*smsclient.QueryOperateTrafficFunnelResp, error) {
				if in.Bucket != "day" {
					t.Fatalf("expected day bucket, got %s", in.Bucket)
				}
				return &smsclient.QueryOperateTrafficFunnelResp{
					Buckets: []*smsclient.OperateTrafficBucketPoint{
						{BucketLabel: "2026-04-03", BucketStart: "2026-04-03 00:00:00", BucketEnd: "2026-04-04 00:00:00", Exposure: 120, Click: 30},
					},
					TotalExposure:     120,
					TotalClick:        30,
					TrackingStartedAt: "2026-04-01 00:00:00",
					PartialMetrics:    []string{"exposure", "click"},
				}, nil
			},
			queryCouponRedeemFn: func(_ context.Context, _ *smsclient.QueryOperateCouponRedeemReq, _ ...grpc.CallOption) (*smsclient.QueryOperateCouponRedeemResp, error) {
				return &smsclient.QueryOperateCouponRedeemResp{
					Buckets: []*smsclient.OperateCouponRedeemBucketPoint{
						{BucketLabel: "2026-04-03", BucketStart: "2026-04-03 00:00:00", BucketEnd: "2026-04-04 00:00:00", CouponRedeem: 2},
					},
					TotalCouponRedeem: 2,
					TrackingStartedAt: "2026-04-04 00:00:00",
					PartialMetrics:    []string{"coupon_redeem"},
				}, nil
			},
			queryActivityOptionsFn: func(_ context.Context, _ *smsclient.QueryOperateActivityOptionsReq, _ ...grpc.CallOption) (*smsclient.QueryOperateActivityOptionsResp, error) {
				return &smsclient.QueryOperateActivityOptionsResp{
					List: []*smsclient.OperateFunnelActivityOption{
						{ActivityType: "coupon", ActivityId: 501, ActivityName: "满100减20", ActivityLabel: "coupon / 满100减20"},
					},
				}, nil
			},
		},
		CartItemService: &mockCartItemService{
			queryOperateCartFunnelFn: func(_ context.Context, _ *omsclient.QueryOperateCartFunnelReq, _ ...grpc.CallOption) (*omsclient.QueryOperateCartFunnelResp, error) {
				return &omsclient.QueryOperateCartFunnelResp{
					Buckets: []*omsclient.OperateCartBucketPoint{
						{BucketLabel: "2026-04-03", BucketStart: "2026-04-03 00:00:00", BucketEnd: "2026-04-04 00:00:00", AddCart: 12},
					},
					TotalAddCart:      12,
					TrackingStartedAt: "2026-04-02 00:00:00",
					PartialMetrics:    []string{"add_cart"},
				}, nil
			},
		},
		OrderService: &mockOrderService{
			queryOperateOrderFunnelFn: func(_ context.Context, _ *omsclient.QueryOperateOrderFunnelReq, _ ...grpc.CallOption) (*omsclient.QueryOperateOrderFunnelResp, error) {
				return &omsclient.QueryOperateOrderFunnelResp{
					Buckets: []*omsclient.OperateOrderBucketPoint{
						{BucketLabel: "2026-04-03", BucketStart: "2026-04-03 00:00:00", BucketEnd: "2026-04-04 00:00:00", OrderCreated: 6, PaySuccess: 3},
					},
					TotalOrderCreated: 6,
					TotalPaySuccess:   3,
					TrackingStartedAt: "2026-04-03 00:00:00",
					PartialMetrics:    []string{"order_created", "pay_success"},
				}, nil
			},
		},
	})

	resp, err := logic.QueryOperateFunnelDashboard(&types.QueryOperateFunnelDashboardReq{
		StartTime: "2026-04-03 00:00:00",
		EndTime:   "2026-04-04 00:00:00",
	})
	if err != nil {
		t.Fatalf("QueryOperateFunnelDashboard returned error: %v", err)
	}

	if resp.Code != "000000" || !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}
	if resp.Data.Overview.Click != 30 || resp.Data.Overview.AddCart != 12 || resp.Data.Overview.PaySuccess != 3 || resp.Data.Overview.CouponRedeem != 2 {
		t.Fatalf("unexpected overview metrics: %+v", resp.Data.Overview)
	}
	if got := resp.Data.Overview.AddCartRate; got != 0.4 {
		t.Fatalf("expected addCartRate 0.4, got %v", got)
	}
	if got := resp.Data.TrackingStartedAt; got != "2026-04-04 00:00:00" {
		t.Fatalf("expected latest trackingStartedAt, got %s", got)
	}
	if got := resp.Data.PartialMetrics; len(got) != 6 || got[0] != "exposure" || got[5] != "coupon_redeem" {
		t.Fatalf("unexpected partial metrics: %+v", got)
	}
	if len(resp.Data.Series) != 1 {
		t.Fatalf("expected 1 series point, got %d", len(resp.Data.Series))
	}
	if resp.Data.Series[0].OrderRate != 0.5 || resp.Data.Series[0].CouponRedeemRate != (2.0/3.0) {
		t.Fatalf("unexpected series rates: %+v", resp.Data.Series[0])
	}
	if len(resp.Data.ActivityOptions) != 1 || resp.Data.ActivityOptions[0].ActivityId != 501 {
		t.Fatalf("unexpected activity options: %+v", resp.Data.ActivityOptions)
	}
}

func TestQueryOperateFunnelDashboardPassesResolvedGovernanceScope(t *testing.T) {
	testCases := []struct {
		name           string
		scopeType      string
		platformID     json.Number
		tenantID       json.Number
		merchantID     json.Number
		expectedScope  string
		expectedTenant int64
		expectedSeller int64
	}{
		{
			name:           "platform",
			scopeType:      "platform",
			platformID:     json.Number("1"),
			tenantID:       json.Number("0"),
			merchantID:     json.Number("0"),
			expectedScope:  "platform",
			expectedTenant: 0,
			expectedSeller: 0,
		},
		{
			name:           "tenant",
			scopeType:      "tenant",
			platformID:     json.Number("1"),
			tenantID:       json.Number("10"),
			merchantID:     json.Number("0"),
			expectedScope:  "tenant",
			expectedTenant: 10,
			expectedSeller: 0,
		},
		{
			name:           "merchant",
			scopeType:      "merchant",
			platformID:     json.Number("1"),
			tenantID:       json.Number("10"),
			merchantID:     json.Number("88"),
			expectedScope:  "merchant",
			expectedTenant: 10,
			expectedSeller: 88,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "scopeType", testCase.scopeType)
			ctx = context.WithValue(ctx, "platformId", testCase.platformID)
			ctx = context.WithValue(ctx, "tenantId", testCase.tenantID)
			ctx = context.WithValue(ctx, "merchantId", testCase.merchantID)

			assertScope := func(scopeType string, tenantID, merchantID int64) {
				t.Helper()
				if scopeType != testCase.expectedScope || tenantID != testCase.expectedTenant || merchantID != testCase.expectedSeller {
					t.Fatalf("expected scope %s/%d/%d, got %s/%d/%d",
						testCase.expectedScope, testCase.expectedTenant, testCase.expectedSeller,
						scopeType, tenantID, merchantID,
					)
				}
			}

			logic := NewQueryOperateFunnelDashboardLogic(ctx, &svc.ServiceContext{
				OperateDashboardService: &mockOperateDashboardService{
					queryTrafficFunnelFn: func(_ context.Context, in *smsclient.QueryOperateTrafficFunnelReq, _ ...grpc.CallOption) (*smsclient.QueryOperateTrafficFunnelResp, error) {
						assertScope(in.Scope.ScopeType, in.Scope.TenantId, in.Scope.MerchantId)
						return &smsclient.QueryOperateTrafficFunnelResp{}, nil
					},
					queryCouponRedeemFn: func(_ context.Context, in *smsclient.QueryOperateCouponRedeemReq, _ ...grpc.CallOption) (*smsclient.QueryOperateCouponRedeemResp, error) {
						assertScope(in.Scope.ScopeType, in.Scope.TenantId, in.Scope.MerchantId)
						return &smsclient.QueryOperateCouponRedeemResp{}, nil
					},
					queryActivityOptionsFn: func(_ context.Context, in *smsclient.QueryOperateActivityOptionsReq, _ ...grpc.CallOption) (*smsclient.QueryOperateActivityOptionsResp, error) {
						assertScope(in.Scope.ScopeType, in.Scope.TenantId, in.Scope.MerchantId)
						return &smsclient.QueryOperateActivityOptionsResp{}, nil
					},
				},
				CartItemService: &mockCartItemService{
					queryOperateCartFunnelFn: func(_ context.Context, in *omsclient.QueryOperateCartFunnelReq, _ ...grpc.CallOption) (*omsclient.QueryOperateCartFunnelResp, error) {
						assertScope(in.Scope.ScopeType, in.Scope.TenantId, in.Scope.MerchantId)
						return &omsclient.QueryOperateCartFunnelResp{}, nil
					},
				},
				OrderService: &mockOrderService{
					queryOperateOrderFunnelFn: func(_ context.Context, in *omsclient.QueryOperateOrderFunnelReq, _ ...grpc.CallOption) (*omsclient.QueryOperateOrderFunnelResp, error) {
						assertScope(in.Scope.ScopeType, in.Scope.TenantId, in.Scope.MerchantId)
						return &omsclient.QueryOperateOrderFunnelResp{}, nil
					},
				},
			})

			if _, err := logic.QueryOperateFunnelDashboard(&types.QueryOperateFunnelDashboardReq{
				StartTime: "2026-04-03 00:00:00",
				EndTime:   "2026-04-04 00:00:00",
			}); err != nil {
				t.Fatalf("QueryOperateFunnelDashboard returned error: %v", err)
			}
		})
	}
}

func TestQueryOperateFunnelDashboardRejectsScopeEscalation(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "tenant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("0"))

	logic := NewQueryOperateFunnelDashboardLogic(ctx, &svc.ServiceContext{})
	_, err := logic.QueryOperateFunnelDashboard(&types.QueryOperateFunnelDashboardReq{
		ScopeType: "platform",
		StartTime: "2026-04-03 00:00:00",
		EndTime:   "2026-04-04 00:00:00",
	})
	if err == nil {
		t.Fatalf("expected scope escalation error, got nil")
	}
	if !strings.Contains(err.Error(), "当前主体不允许切换查询范围") {
		t.Fatalf("expected scope escalation message, got %v", err)
	}
}
