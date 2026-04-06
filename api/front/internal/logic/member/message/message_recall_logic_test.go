package message

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"google.golang.org/grpc"
)

type mockFrontMemberMessageService struct {
	membermessageservice.MemberMessageService
	queryListFn func(context.Context, *membermessageservice.QueryMemberMessageListReq, ...grpc.CallOption) (*membermessageservice.QueryMemberMessageListResp, error)
}

func (m *mockFrontMemberMessageService) QueryMemberMessageList(ctx context.Context, in *membermessageservice.QueryMemberMessageListReq, opts ...grpc.CallOption) (*membermessageservice.QueryMemberMessageListResp, error) {
	return m.queryListFn(ctx, in, opts...)
}

func newFrontMessageTestContext(appVersion string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "memberId", json.Number("1001"))
	ctx = WithRecallRequestMetadata(ctx, RecallRequestMetadata{
		AppVersion: appVersion,
		Platform:   "android",
	})
	return ctx
}

func TestMessageListMapsStructuredRecallIntent(t *testing.T) {
	t.Parallel()

	logic := NewMessageListLogic(newFrontMessageTestContext("1.0.0"), &svc.ServiceContext{
		MemberMessageService: &mockFrontMemberMessageService{
			queryListFn: func(_ context.Context, _ *membermessageservice.QueryMemberMessageListReq, _ ...grpc.CallOption) (*membermessageservice.QueryMemberMessageListResp, error) {
				return &membermessageservice.QueryMemberMessageListResp{
					Total:    1,
					PageNum:  1,
					PageSize: 20,
					List: []*membermessageservice.MemberMessageData{
						{
							Id:             7,
							MessageType:    1,
							Title:          "订单提醒",
							Content:        "点击查看订单详情",
							LinkType:       "order",
							LinkId:         "9001",
							RelatedOrderId: 9001,
							Status:         0,
							CreateTime:     "2026-04-05 12:00:00",
							Intent: &membermessageservice.MemberMessageRecallIntent{
								IntentType:   "order_recall",
								TargetType:   "order_detail",
								TargetId:     9001,
								FallbackType: "order_list",
								FallbackTab:  1,
								RequiresAuth: true,
								IntentId:     "member_message:7",
								IssuedAt:     "2026-04-05 12:00:00",
								Source:       "member_message",
							},
						},
					},
				}, nil
			},
		},
	})

	resp, err := logic.MessageList(&types.QueryMemberMessageReq{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("MessageList returned error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 message item, got %d", len(resp.Data))
	}
	if resp.Data[0].Intent == nil {
		t.Fatal("expected structured intent to be mapped to front response")
	}
	if resp.Data[0].Intent.TargetType != "order_detail" || resp.Data[0].Intent.TargetID == nil || *resp.Data[0].Intent.TargetID != 9001 {
		t.Fatalf("unexpected mapped intent: %+v", resp.Data[0].Intent)
	}
	if resp.Data[0].Intent.IntentID != "member_message:7" {
		t.Fatalf("expected intent id member_message:7, got %s", resp.Data[0].Intent.IntentID)
	}
}

func TestMessageListBlocksRecallWhenAppVersionIsBelowMinimum(t *testing.T) {
	t.Parallel()

	logic := NewMessageListLogic(newFrontMessageTestContext("1.0.0"), &svc.ServiceContext{
		MemberMessageService: &mockFrontMemberMessageService{
			queryListFn: func(_ context.Context, _ *membermessageservice.QueryMemberMessageListReq, _ ...grpc.CallOption) (*membermessageservice.QueryMemberMessageListResp, error) {
				return &membermessageservice.QueryMemberMessageListResp{
					Total:    1,
					PageNum:  1,
					PageSize: 20,
					List: []*membermessageservice.MemberMessageData{
						{
							Id:          8,
							MessageType: 5,
							Title:       "优惠券到账提醒",
							Content:     "登录领取新券",
							Status:      0,
							CreateTime:  "2026-04-05 12:05:00",
							Intent: &membermessageservice.MemberMessageRecallIntent{
								IntentType:    "coupon_center_recall",
								TargetType:    "coupon_center",
								FallbackType:  "coupon_list",
								FallbackTab:   0,
								RequiresAuth:  true,
								MinAppVersion: "1.2.0",
								IntentId:      "member_message:8",
								IssuedAt:      "2026-04-05 12:05:00",
								Source:        "member_message",
							},
						},
					},
				}, nil
			},
		},
	})

	resp, err := logic.MessageList(&types.QueryMemberMessageReq{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("MessageList returned error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Intent == nil {
		t.Fatalf("expected mapped intent response, got %+v", resp.Data)
	}
	if !resp.Data[0].Intent.Blocked {
		t.Fatalf("expected min version mismatch to block recall, got %+v", resp.Data[0].Intent)
	}
	if resp.Data[0].Intent.FailureReason != recallFailureMinVersionUnmet {
		t.Fatalf("expected failure reason %s, got %s", recallFailureMinVersionUnmet, resp.Data[0].Intent.FailureReason)
	}
}

func TestMessageIntentResponseOmitsZeroValueOptionalNumericFields(t *testing.T) {
	t.Parallel()

	intent := messageIntentResponseFromRPC(context.Background(), &membermessageservice.MemberMessageData{
		Intent: &membermessageservice.MemberMessageRecallIntent{
			IntentType:   "coupon_center_recall",
			TargetType:   "coupon_center",
			TargetTab:    0,
			FallbackType: "coupon_list",
			FallbackTab:  0,
			RequiresAuth: true,
			IntentId:     "member_message:9",
			IssuedAt:     "2026-04-05 12:10:00",
			Source:       "member_message",
		},
	})
	if intent == nil {
		t.Fatal("expected intent response to be mapped")
	}
	if intent.TargetID != nil {
		t.Fatalf("expected target id to be omitted, got %v", *intent.TargetID)
	}
	if intent.TargetTab != nil {
		t.Fatalf("expected target tab to be omitted for coupon_center, got %v", *intent.TargetTab)
	}
	if intent.FallbackTargetID != nil {
		t.Fatalf("expected fallback target id to be omitted, got %v", *intent.FallbackTargetID)
	}
	if intent.FallbackTab == nil || *intent.FallbackTab != 0 {
		t.Fatalf("expected fallback tab to stay available, got %+v", intent.FallbackTab)
	}

	encoded, err := json.Marshal(intent)
	if err != nil {
		t.Fatalf("marshal intent response failed: %v", err)
	}
	payload := string(encoded)
	if strings.Contains(payload, "\"targetId\"") {
		t.Fatalf("expected payload to omit targetId, got %s", payload)
	}
	if strings.Contains(payload, "\"targetTab\"") {
		t.Fatalf("expected payload to omit targetTab, got %s", payload)
	}
	if strings.Contains(payload, "\"fallbackTargetId\"") {
		t.Fatalf("expected payload to omit fallbackTargetId, got %s", payload)
	}
	if !strings.Contains(payload, "\"fallbackTab\":0") {
		t.Fatalf("expected payload to preserve fallbackTab=0, got %s", payload)
	}
}
