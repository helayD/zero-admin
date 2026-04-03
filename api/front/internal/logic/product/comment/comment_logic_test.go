package comment

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	pmscommentreplayservice "github.com/feihua/zero-admin/rpc/pms/client/commentreplayservice"
	pmscommentservice "github.com/feihua/zero-admin/rpc/pms/client/commentservice"
	"google.golang.org/grpc"
)

type mockFrontCommentService struct {
	pmscommentservice.CommentService
	queryListFn   func(context.Context, *pmscommentservice.QueryCommentListReq, ...grpc.CallOption) (*pmscommentservice.QueryCommentListResp, error)
	queryDetailFn func(context.Context, *pmscommentservice.QueryCommentDetailReq, ...grpc.CallOption) (*pmscommentservice.QueryCommentDetailResp, error)
}

func (m *mockFrontCommentService) QueryCommentList(ctx context.Context, in *pmscommentservice.QueryCommentListReq, opts ...grpc.CallOption) (*pmscommentservice.QueryCommentListResp, error) {
	return m.queryListFn(ctx, in, opts...)
}

func (m *mockFrontCommentService) QueryCommentDetail(ctx context.Context, in *pmscommentservice.QueryCommentDetailReq, opts ...grpc.CallOption) (*pmscommentservice.QueryCommentDetailResp, error) {
	return m.queryDetailFn(ctx, in, opts...)
}

type mockFrontCommentReplayService struct {
	pmscommentreplayservice.CommentReplayService
	queryListFn func(context.Context, *pmscommentreplayservice.QueryCommentReplayListReq, ...grpc.CallOption) (*pmscommentreplayservice.QueryCommentReplayListResp, error)
}

func (m *mockFrontCommentReplayService) QueryCommentReplayList(ctx context.Context, in *pmscommentreplayservice.QueryCommentReplayListReq, opts ...grpc.CallOption) (*pmscommentreplayservice.QueryCommentReplayListResp, error) {
	return m.queryListFn(ctx, in, opts...)
}

func newFrontCommentTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("88"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("3001"))
	return ctx
}

func TestQueryCommentListUsesApprovedVisibleFilterAndScope(t *testing.T) {
	t.Parallel()

	var captured *pmscommentservice.QueryCommentListReq
	logic := NewQueryCommentListLogic(newFrontCommentTestContext(), &svc.ServiceContext{
		CommentService: &mockFrontCommentService{
			queryListFn: func(_ context.Context, in *pmscommentservice.QueryCommentListReq, _ ...grpc.CallOption) (*pmscommentservice.QueryCommentListResp, error) {
				captured = in
				return &pmscommentservice.QueryCommentListResp{
					Total: 1,
					List: []*pmscommentservice.CommentListData{
						{
							Id:               "comment-1",
							ProductId:        9527,
							MemberId:         2001,
							MemberNickName:   "张三",
							MemberIcon:       "avatar.png",
							Star:             5,
							Content:          "很好用，值得推荐",
							Pics:             "a.png,b.png",
							ProductAttribute: "红色;64G",
							ShowStatus:       1,
							ReplayCount:      2,
							MemberIp:         "127.0.0.1",
							CreateTime:       "2026-04-02 10:00:00",
						},
					},
				}, nil
			},
		},
	})

	resp, err := logic.QueryCommentList(&types.QueryCommentListReq{
		ProductId: 9527,
		PageNum:   1,
		PageSize:  20,
	})
	if err != nil {
		t.Fatalf("QueryCommentList returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected QueryCommentList RPC request to be captured")
	}
	if captured.ShowStatus != 1 || captured.AuditStatus != 1 || captured.Hidden != 0 {
		t.Fatalf("expected front filter to force show/audit/hidden = 1/1/0, got %d/%d/%d", captured.ShowStatus, captured.AuditStatus, captured.Hidden)
	}
	if captured.PlatformId != 1 || captured.TenantId != 88 || captured.MerchantId != 3001 {
		t.Fatalf("expected governance scope 1/88/3001, got %d/%d/%d", captured.PlatformId, captured.TenantId, captured.MerchantId)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 comment item, got %d", len(resp.Data))
	}
	if resp.Data[0].Star != 5 || resp.Data[0].ShowStatus != 1 || resp.Data[0].ReplayCount != 2 {
		t.Fatalf("expected int32 fields to be mapped unchanged, got star=%d showStatus=%d replayCount=%d", resp.Data[0].Star, resp.Data[0].ShowStatus, resp.Data[0].ReplayCount)
	}
}

func TestQueryCommentDetailRejectsUnapprovedComment(t *testing.T) {
	t.Parallel()

	logic := NewQueryCommentDetailLogic(newFrontCommentTestContext(), &svc.ServiceContext{
		CommentService: &mockFrontCommentService{
			queryDetailFn: func(_ context.Context, in *pmscommentservice.QueryCommentDetailReq, _ ...grpc.CallOption) (*pmscommentservice.QueryCommentDetailResp, error) {
				return &pmscommentservice.QueryCommentDetailResp{
					Id:          in.Id,
					ProductId:   9527,
					AuditStatus: 2,
					Hidden:      0,
				}, nil
			},
		},
	})

	_, err := logic.QueryCommentDetail(&types.QueryCommentDetailReq{
		Id:        "comment-2",
		ProductId: 9527,
	})
	if err == nil {
		t.Fatal("expected hidden/unapproved comment to be rejected")
	}
	if err.Error() != "评价不存在或已下架" {
		t.Fatalf("expected visibility error, got %q", err.Error())
	}
}

func TestQueryCommentDetailMapsVisibleCommentAndReplay(t *testing.T) {
	t.Parallel()

	var replayCalled bool
	logic := NewQueryCommentDetailLogic(newFrontCommentTestContext(), &svc.ServiceContext{
		CommentService: &mockFrontCommentService{
			queryDetailFn: func(_ context.Context, in *pmscommentservice.QueryCommentDetailReq, _ ...grpc.CallOption) (*pmscommentservice.QueryCommentDetailResp, error) {
				return &pmscommentservice.QueryCommentDetailResp{
					Id:               in.Id,
					ProductId:        9527,
					MemberId:         2001,
					MemberNickName:   "张三",
					MemberIcon:       "avatar.png",
					Star:             4,
					Content:          "包装很好，物流很快",
					Pics:             "a.png",
					ProductAttribute: "蓝色;128G",
					ShowStatus:       1,
					ReplayCount:      1,
					MemberIp:         "127.0.0.1",
					CreateTime:       "2026-04-02 11:00:00",
					AuditStatus:      1,
					Hidden:           0,
				}, nil
			},
		},
		CommentReplayService: &mockFrontCommentReplayService{
			queryListFn: func(_ context.Context, in *pmscommentreplayservice.QueryCommentReplayListReq, _ ...grpc.CallOption) (*pmscommentreplayservice.QueryCommentReplayListResp, error) {
				replayCalled = true
				if in.CommentId != "comment-3" {
					t.Fatalf("expected replay query to use comment id comment-3, got %s", in.CommentId)
				}
				return &pmscommentreplayservice.QueryCommentReplayListResp{
					List: []*pmscommentreplayservice.CommentReplayListData{
						{
							Id:             "reply-1",
							CommentId:      "comment-3",
							Type:           1,
							MemberNickName: "运营小助手",
							Content:        "感谢反馈",
							CreateTime:     "2026-04-02 11:05:00",
						},
					},
				}, nil
			},
		},
	})

	resp, err := logic.QueryCommentDetail(&types.QueryCommentDetailReq{
		Id:        "comment-3",
		ProductId: 9527,
	})
	if err != nil {
		t.Fatalf("QueryCommentDetail returned error: %v", err)
	}
	if !replayCalled {
		t.Fatal("expected replay service to be called for visible comment")
	}
	if resp.Data.Star != 4 || resp.Data.ShowStatus != 1 || resp.Data.ReplayCount != 1 {
		t.Fatalf("expected visible comment fields to be mapped unchanged, got star=%d showStatus=%d replayCount=%d", resp.Data.Star, resp.Data.ShowStatus, resp.Data.ReplayCount)
	}
	if len(resp.Replays) != 1 || resp.Replays[0].Type != 1 {
		t.Fatalf("expected replay type to stay int32=1, got %+v", resp.Replays)
	}
}
