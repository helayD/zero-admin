package message

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"google.golang.org/grpc"
)

type mockPathMemberMessageService struct {
	membermessageservice.MemberMessageService
	markFn   func(context.Context, *membermessageservice.MarkMessageAsReadReq, ...grpc.CallOption) (*membermessageservice.MarkMessageAsReadResp, error)
	deleteFn func(context.Context, *membermessageservice.DeleteMemberMessageReq, ...grpc.CallOption) (*membermessageservice.DeleteMemberMessageResp, error)
}

func (m *mockPathMemberMessageService) MarkMessageAsRead(ctx context.Context, in *membermessageservice.MarkMessageAsReadReq, opts ...grpc.CallOption) (*membermessageservice.MarkMessageAsReadResp, error) {
	return m.markFn(ctx, in, opts...)
}

func (m *mockPathMemberMessageService) DeleteMemberMessage(ctx context.Context, in *membermessageservice.DeleteMemberMessageReq, opts ...grpc.CallOption) (*membermessageservice.DeleteMemberMessageResp, error) {
	return m.deleteFn(ctx, in, opts...)
}

func TestMarkMessageReadByPathAcceptsJsonNullBody(t *testing.T) {
	var gotID int64
	var gotMemberID int64
	svcCtx := &svc.ServiceContext{
		MemberMessageService: &mockPathMemberMessageService{
			markFn: func(_ context.Context, in *membermessageservice.MarkMessageAsReadReq, _ ...grpc.CallOption) (*membermessageservice.MarkMessageAsReadResp, error) {
				gotID = in.Id
				gotMemberID = in.MemberId
				return &membermessageservice.MarkMessageAsReadResp{Code: 0, Msg: "标记成功"}, nil
			},
		},
	}
	req := newPathRequest(http.MethodPost, "/api/member/message/12/read", "12")
	rr := httptest.NewRecorder()

	MarkMessageReadByPathHandler(svcCtx)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if gotID != 12 || gotMemberID != 1001 {
		t.Fatalf("unexpected rpc params: id=%d memberId=%d", gotID, gotMemberID)
	}
}

func TestDeleteMessageByPathAcceptsJsonNullBody(t *testing.T) {
	var gotID int64
	var gotMemberID int64
	svcCtx := &svc.ServiceContext{
		MemberMessageService: &mockPathMemberMessageService{
			deleteFn: func(_ context.Context, in *membermessageservice.DeleteMemberMessageReq, _ ...grpc.CallOption) (*membermessageservice.DeleteMemberMessageResp, error) {
				gotID = in.Id
				gotMemberID = in.MemberId
				return &membermessageservice.DeleteMemberMessageResp{Code: 0, Msg: "删除成功"}, nil
			},
		},
	}
	req := newPathRequest(http.MethodDelete, "/api/member/message/12", "12")
	rr := httptest.NewRecorder()

	DeleteMessageByPathHandler(svcCtx)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if gotID != 12 || gotMemberID != 1001 {
		t.Fatalf("unexpected rpc params: id=%d memberId=%d", gotID, gotMemberID)
	}
}

func newPathRequest(method, target, id string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader("null"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "memberId", json.Number("1001")))
	return pathvar.WithVars(req, map[string]string{"id": id})
}
