package operate_dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/feihua/zero-admin/rpc/ums/client/memberinfoservice"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc"
)

type mockRepeatPurchaseOrderService struct {
	orderservice.OrderService
	queryRepeatPurchaseAnalysisFn   func(context.Context, *omsclient.QueryRepeatPurchaseAnalysisReq, ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseAnalysisResp, error)
	queryRepeatPurchaseDetailListFn func(context.Context, *omsclient.QueryRepeatPurchaseDetailListReq, ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseDetailListResp, error)
}

func (m *mockRepeatPurchaseOrderService) QueryRepeatPurchaseAnalysis(ctx context.Context, in *omsclient.QueryRepeatPurchaseAnalysisReq, opts ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseAnalysisResp, error) {
	return m.queryRepeatPurchaseAnalysisFn(ctx, in, opts...)
}

func (m *mockRepeatPurchaseOrderService) QueryRepeatPurchaseDetailList(ctx context.Context, in *omsclient.QueryRepeatPurchaseDetailListReq, opts ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseDetailListResp, error) {
	return m.queryRepeatPurchaseDetailListFn(ctx, in, opts...)
}

type mockRepeatPurchaseMemberInfoService struct {
	memberinfoservice.MemberInfoService
	queryMemberBriefByIDsFn func(context.Context, *umsclient.QueryMemberBriefByIdsReq, ...grpc.CallOption) (*umsclient.QueryMemberBriefByIdsResp, error)
}

func (m *mockRepeatPurchaseMemberInfoService) QueryMemberBriefByIds(ctx context.Context, in *umsclient.QueryMemberBriefByIdsReq, opts ...grpc.CallOption) (*umsclient.QueryMemberBriefByIdsResp, error) {
	return m.queryMemberBriefByIDsFn(ctx, in, opts...)
}

func TestQueryRepeatPurchaseAnalysisAggregatesResponses(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("88"))

	logic := NewQueryRepeatPurchaseAnalysisLogic(ctx, &svc.ServiceContext{
		OrderService: &mockRepeatPurchaseOrderService{
			queryRepeatPurchaseAnalysisFn: func(_ context.Context, in *omsclient.QueryRepeatPurchaseAnalysisReq, _ ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseAnalysisResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 88 {
					t.Fatalf("unexpected analysis scope: %+v", in.Scope)
				}
				if in.Channel != "app" || in.ActivityType != "coupon" || in.ActivityId != 9 || in.Bucket != "day" {
					t.Fatalf("unexpected analysis filters: %+v", in)
				}
				return &omsclient.QueryRepeatPurchaseAnalysisResp{
					Overview: &omsclient.RepeatPurchaseOverview{
						PaidBuyerCount:   12,
						RepeatBuyerCount: 3,
						RepeatRate:       0.25,
						RepeatOrderCount: 5,
						RepeatGmv:        188.6,
						AvgDaysToRepeat:  6.5,
					},
					Trends: []*omsclient.RepeatPurchaseTrendPoint{
						{
							BucketLabel:      "2026-04-01",
							BucketStart:      "2026-04-01 00:00:00",
							BucketEnd:        "2026-04-02 00:00:00",
							PaidBuyerCount:   4,
							RepeatBuyerCount: 1,
							RepeatRate:       0.25,
							RepeatOrderCount: 2,
							RepeatGmv:        66.6,
							AvgDaysToRepeat:  3.5,
						},
					},
					TrackingStartedAt: "2026-03-01 00:00:00",
					PartialMetrics:    []string{"paidBuyerCount", "repeatRate"},
				}, nil
			},
			queryRepeatPurchaseDetailListFn: func(_ context.Context, in *omsclient.QueryRepeatPurchaseDetailListReq, _ ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseDetailListResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 88 {
					t.Fatalf("unexpected detail scope: %+v", in.Scope)
				}
				if in.PageNum != 2 || in.PageSize != 5 {
					t.Fatalf("unexpected detail page: %d/%d", in.PageNum, in.PageSize)
				}
				return &omsclient.QueryRepeatPurchaseDetailListResp{
					Total: 1,
					List: []*omsclient.RepeatPurchaseDetailRow{
						{
							MemberId:            3001,
							FirstValidPayTime:   "2026-03-20 10:00:00",
							LatestRepeatPayTime: "2026-04-02 11:00:00",
							RepeatOrderCount:    2,
							RepeatGmv:           88.8,
							LatestChannel:       "app",
							LatestActivityType:  "coupon",
							LatestActivityId:    9,
							PlatformId:          1,
							TenantId:            10,
							MerchantId:          88,
						},
					},
					TrackingStartedAt: "2026-03-05 00:00:00",
					PartialMetrics:    []string{"repeatOrderCount", "customMetric"},
				}, nil
			},
		},
		OperateDashboardService: &mockOperateDashboardService{
			queryActivityOptionsFn: func(_ context.Context, in *smsclient.QueryOperateActivityOptionsReq, _ ...grpc.CallOption) (*smsclient.QueryOperateActivityOptionsResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 88 {
					t.Fatalf("unexpected activity scope: %+v", in.Scope)
				}
				return &smsclient.QueryOperateActivityOptionsResp{
					List: []*smsclient.OperateFunnelActivityOption{{
						ActivityType:  "coupon",
						ActivityId:    9,
						ActivityName:  "满100减20",
						ActivityLabel: "coupon / 满100减20",
					}},
				}, nil
			},
		},
		MemberInfoService: &mockRepeatPurchaseMemberInfoService{
			queryMemberBriefByIDsFn: func(_ context.Context, in *umsclient.QueryMemberBriefByIdsReq, _ ...grpc.CallOption) (*umsclient.QueryMemberBriefByIdsResp, error) {
				if !reflect.DeepEqual(in.MemberIds, []int64{3001}) {
					t.Fatalf("unexpected member ids: %+v", in.MemberIds)
				}
				return &umsclient.QueryMemberBriefByIdsResp{
					List: []*umsclient.MemberBriefData{{
						MemberId:       3001,
						NicknameMasked: "会***员",
						MobileMasked:   "138****0001",
					}},
				}, nil
			},
		},
	})

	resp, err := logic.QueryRepeatPurchaseAnalysis(&types.QueryRepeatPurchaseAnalysisReq{
		StartTime:    "2026-04-01 00:00:00",
		EndTime:      "2026-04-08 00:00:00",
		Channel:      "app",
		ActivityType: "coupon",
		ActivityId:   9,
		Bucket:       "day",
		PageNum:      2,
		PageSize:     5,
	})
	if err != nil {
		t.Fatalf("QueryRepeatPurchaseAnalysis returned error: %v", err)
	}

	if resp.Code != "000000" || !resp.Success {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Data.Overview.RepeatBuyerCount != 3 || resp.Data.Overview.RepeatGmv != 188.6 {
		t.Fatalf("unexpected overview: %+v", resp.Data.Overview)
	}
	if len(resp.Data.Trends) != 1 || resp.Data.Trends[0].BucketLabel != "2026-04-01" {
		t.Fatalf("unexpected trends: %+v", resp.Data.Trends)
	}
	if resp.Data.Total != 1 || resp.Data.PageNum != 2 || resp.Data.PageSize != 5 {
		t.Fatalf("unexpected paging data: %+v", resp.Data)
	}
	if len(resp.Data.Details) != 1 {
		t.Fatalf("unexpected details: %+v", resp.Data.Details)
	}
	if resp.Data.Details[0].NicknameMasked != "会***员" || resp.Data.Details[0].MobileMasked != "138****0001" {
		t.Fatalf("unexpected detail masks: %+v", resp.Data.Details[0])
	}
	if len(resp.Data.ActivityOptions) != 1 || resp.Data.ActivityOptions[0].ActivityId != 9 {
		t.Fatalf("unexpected activity options: %+v", resp.Data.ActivityOptions)
	}
	if resp.Data.TrackingStartedAt != "2026-03-05 00:00:00" {
		t.Fatalf("unexpected trackingStartedAt: %s", resp.Data.TrackingStartedAt)
	}
	wantMetrics := []string{"paidBuyerCount", "repeatRate", "repeatOrderCount", "customMetric"}
	if !reflect.DeepEqual(resp.Data.PartialMetrics, wantMetrics) {
		t.Fatalf("unexpected partial metrics: got %+v want %+v", resp.Data.PartialMetrics, wantMetrics)
	}
	if resp.Data.Bucket != "day" {
		t.Fatalf("unexpected bucket: %s", resp.Data.Bucket)
	}
}

func TestExportRepeatPurchaseAnalysisWritesExcelRows(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("88"))

	pageCalls := make([]int32, 0, 2)
	memberCalls := make([][]int64, 0, 2)
	logic := NewExportRepeatPurchaseAnalysisLogic(ctx, &svc.ServiceContext{
		OrderService: &mockRepeatPurchaseOrderService{
			queryRepeatPurchaseDetailListFn: func(_ context.Context, in *omsclient.QueryRepeatPurchaseDetailListReq, _ ...grpc.CallOption) (*omsclient.QueryRepeatPurchaseDetailListResp, error) {
				if in.Scope == nil || in.Scope.ScopeType != "merchant" || in.Scope.PlatformId != 1 || in.Scope.TenantId != 10 || in.Scope.MerchantId != 88 {
					t.Fatalf("unexpected export scope: %+v", in.Scope)
				}
				pageCalls = append(pageCalls, in.PageNum)
				switch in.PageNum {
				case 1:
					return &omsclient.QueryRepeatPurchaseDetailListResp{
						Total: 501,
						List: []*omsclient.RepeatPurchaseDetailRow{{
							MemberId:            4001,
							FirstValidPayTime:   "2026-04-01 08:00:00",
							LatestRepeatPayTime: "2026-04-03 09:00:00",
							RepeatOrderCount:    2,
							RepeatGmv:           66.6,
							LatestChannel:       "mini_program",
							LatestActivityType:  "coupon",
							LatestActivityId:    11,
							PlatformId:          1,
							TenantId:            10,
							MerchantId:          88,
						}},
					}, nil
				case 2:
					return &omsclient.QueryRepeatPurchaseDetailListResp{
						Total: 501,
						List: []*omsclient.RepeatPurchaseDetailRow{{
							MemberId:            4002,
							FirstValidPayTime:   "2026-04-02 08:00:00",
							LatestRepeatPayTime: "2026-04-04 09:00:00",
							RepeatOrderCount:    3,
							RepeatGmv:           99.9,
							LatestChannel:       "app",
							LatestActivityType:  "coupon",
							LatestActivityId:    12,
							PlatformId:          1,
							TenantId:            10,
							MerchantId:          88,
						}},
					}, nil
				default:
					t.Fatalf("unexpected page number: %d", in.PageNum)
					return nil, nil
				}
			},
		},
		MemberInfoService: &mockRepeatPurchaseMemberInfoService{
			queryMemberBriefByIDsFn: func(_ context.Context, in *umsclient.QueryMemberBriefByIdsReq, _ ...grpc.CallOption) (*umsclient.QueryMemberBriefByIdsResp, error) {
				ids := append([]int64(nil), in.MemberIds...)
				memberCalls = append(memberCalls, ids)
				resp := &umsclient.QueryMemberBriefByIdsResp{List: make([]*umsclient.MemberBriefData, 0, len(in.MemberIds))}
				for _, memberID := range in.MemberIds {
					resp.List = append(resp.List, &umsclient.MemberBriefData{
						MemberId:       memberID,
						NicknameMasked: "昵***称",
						MobileMasked:   "139****0000",
					})
				}
				return resp, nil
			},
		},
	})

	content, err := logic.ExportRepeatPurchaseAnalysis(&types.ExportRepeatPurchaseAnalysisReq{
		StartTime:    "2026-04-01 00:00:00",
		EndTime:      "2026-04-08 00:00:00",
		Channel:      "app",
		ActivityType: "coupon",
		ActivityId:   9,
		Bucket:       "day",
	})
	if err != nil {
		t.Fatalf("ExportRepeatPurchaseAnalysis returned error: %v", err)
	}
	if !reflect.DeepEqual(pageCalls, []int32{1, 2}) {
		t.Fatalf("unexpected page calls: %+v", pageCalls)
	}
	if !reflect.DeepEqual(memberCalls, [][]int64{{4001}, {4002}}) {
		t.Fatalf("unexpected member calls: %+v", memberCalls)
	}

	file, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("OpenReader returned error: %v", err)
	}
	defer file.Close()

	rows, err := file.GetRows("复购分析")
	if err != nil {
		t.Fatalf("GetRows returned error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("unexpected row count: %d", len(rows))
	}
	if rows[0][0] != "会员ID" || rows[0][1] != "昵称" || rows[0][2] != "手机号" {
		t.Fatalf("unexpected header row: %+v", rows[0])
	}
	if rows[1][0] != "4001" || rows[1][1] != "昵***称" || rows[1][2] != "139****0000" {
		t.Fatalf("unexpected first data row: %+v", rows[1])
	}
	if rows[2][0] != "4002" || rows[2][7] != "app" || rows[2][9] != "12" {
		t.Fatalf("unexpected second data row: %+v", rows[2])
	}
}
