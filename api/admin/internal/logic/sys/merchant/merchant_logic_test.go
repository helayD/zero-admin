package merchant

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	rpcmerchant "github.com/feihua/zero-admin/rpc/sys/client/merchantservice"
	"google.golang.org/grpc"
)

type fakeMerchantService struct {
	createFn  func(context.Context, *rpcmerchant.CreateMerchantReq, ...grpc.CallOption) (*rpcmerchant.CreateMerchantResp, error)
	approveFn func(context.Context, *rpcmerchant.ReviewMerchantReq, ...grpc.CallOption) (*rpcmerchant.ReviewMerchantResp, error)
}

func (f *fakeMerchantService) CreateMerchant(ctx context.Context, in *rpcmerchant.CreateMerchantReq, opts ...grpc.CallOption) (*rpcmerchant.CreateMerchantResp, error) {
	if f.createFn == nil {
		return &rpcmerchant.CreateMerchantResp{}, nil
	}
	return f.createFn(ctx, in, opts...)
}

func (f *fakeMerchantService) QueryMerchantDetail(context.Context, *rpcmerchant.QueryMerchantDetailReq, ...grpc.CallOption) (*rpcmerchant.QueryMerchantDetailResp, error) {
	return &rpcmerchant.QueryMerchantDetailResp{}, nil
}

func (f *fakeMerchantService) QueryMerchantList(context.Context, *rpcmerchant.QueryMerchantListReq, ...grpc.CallOption) (*rpcmerchant.QueryMerchantListResp, error) {
	return &rpcmerchant.QueryMerchantListResp{}, nil
}

func (f *fakeMerchantService) ApproveMerchant(ctx context.Context, in *rpcmerchant.ReviewMerchantReq, opts ...grpc.CallOption) (*rpcmerchant.ReviewMerchantResp, error) {
	if f.approveFn == nil {
		return &rpcmerchant.ReviewMerchantResp{}, nil
	}
	return f.approveFn(ctx, in, opts...)
}

func (f *fakeMerchantService) RejectMerchant(context.Context, *rpcmerchant.ReviewMerchantReq, ...grpc.CallOption) (*rpcmerchant.ReviewMerchantResp, error) {
	return &rpcmerchant.ReviewMerchantResp{}, nil
}

func (f *fakeMerchantService) RequestMerchantMaterial(context.Context, *rpcmerchant.ReviewMerchantReq, ...grpc.CallOption) (*rpcmerchant.ReviewMerchantResp, error) {
	return &rpcmerchant.ReviewMerchantResp{}, nil
}

func (f *fakeMerchantService) EnableMerchant(context.Context, *rpcmerchant.ChangeMerchantStatusReq, ...grpc.CallOption) (*rpcmerchant.ChangeMerchantStatusResp, error) {
	return &rpcmerchant.ChangeMerchantStatusResp{}, nil
}

func (f *fakeMerchantService) DisableMerchant(context.Context, *rpcmerchant.ChangeMerchantStatusReq, ...grpc.CallOption) (*rpcmerchant.ChangeMerchantStatusResp, error) {
	return &rpcmerchant.ChangeMerchantStatusResp{}, nil
}

func (f *fakeMerchantService) ArchiveMerchant(context.Context, *rpcmerchant.ChangeMerchantStatusReq, ...grpc.CallOption) (*rpcmerchant.ChangeMerchantStatusResp, error) {
	return &rpcmerchant.ChangeMerchantStatusResp{}, nil
}

func merchantTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userName", "platform-admin")
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	return ctx
}

func TestCreateMerchantLogicPassesOperatorContext(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		MerchantService: &fakeMerchantService{
			createFn: func(_ context.Context, in *rpcmerchant.CreateMerchantReq, _ ...grpc.CallOption) (*rpcmerchant.CreateMerchantResp, error) {
				if in.CreateBy != "platform-admin" || in.OperatorId != 77 {
					t.Fatalf("unexpected operator info: %+v", in)
				}
				if in.TenantId != 11 || len(in.CapabilityFlags) != 2 {
					t.Fatalf("unexpected merchant payload: %+v", in)
				}
				if len(in.AvailableChannels) != 2 || in.AvailableChannels[0] != "app" || in.AvailableChannels[1] != "mini_program" {
					t.Fatalf("unexpected channels: %+v", in.AvailableChannels)
				}
				return &rpcmerchant.CreateMerchantResp{
					MerchantId:     101,
					MerchantCode:   "MER202603210001",
					ReviewStatus:   0,
					BusinessStatus: 0,
				}, nil
			},
		},
	}

	logic := NewCreateMerchantLogic(merchantTestContext(), svcCtx)
	resp, err := logic.CreateMerchant(&types.CreateMerchantReq{
		TenantId:          11,
		MerchantName:      "  华东旗舰店  ",
		ContactName:       "李四",
		ContactMobile:     "13800138001",
		AvailableChannels: []string{" app ", "mini-program"},
		CapabilityFlags:   []string{"oms", " crm "},
	})
	if err != nil {
		t.Fatalf("create merchant logic failed: %v", err)
	}
	if resp.Code != "000000" || resp.Data.MerchantId != 101 || resp.Data.MerchantCode != "MER202603210001" {
		t.Fatalf("unexpected create response: %+v", resp)
	}
}

func TestApproveMerchantLogicPassesReviewCommand(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		MerchantService: &fakeMerchantService{
			approveFn: func(_ context.Context, in *rpcmerchant.ReviewMerchantReq, _ ...grpc.CallOption) (*rpcmerchant.ReviewMerchantResp, error) {
				if in.UpdateBy != "platform-admin" || in.OperatorId != 77 {
					t.Fatalf("unexpected operator info: %+v", in)
				}
				if len(in.Ids) != 2 || in.ReviewReason != "资料齐全" {
					t.Fatalf("unexpected review request: %+v", in)
				}
				return &rpcmerchant.ReviewMerchantResp{Pong: "ok"}, nil
			},
		},
	}

	logic := NewApproveMerchantLogic(merchantTestContext(), svcCtx)
	resp, err := logic.ApproveMerchant(&types.ReviewMerchantReq{
		Ids:          []int64{1, 2},
		ReviewReason: "资料齐全",
	})
	if err != nil {
		t.Fatalf("approve merchant logic failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected approve response: %+v", resp)
	}
}
