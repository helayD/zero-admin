package tenant

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	rpctenant "github.com/feihua/zero-admin/rpc/sys/client/tenantservice"
	"google.golang.org/grpc"
)

type fakeTenantService struct {
	createFn      func(context.Context, *rpctenant.CreateTenantReq, ...grpc.CallOption) (*rpctenant.CreateTenantResp, error)
	queryListFn   func(context.Context, *rpctenant.QueryTenantListReq, ...grpc.CallOption) (*rpctenant.QueryTenantListResp, error)
	queryDetailFn func(context.Context, *rpctenant.QueryTenantDetailReq, ...grpc.CallOption) (*rpctenant.QueryTenantDetailResp, error)
	enableFn      func(context.Context, *rpctenant.ChangeTenantStatusReq, ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error)
	disableFn     func(context.Context, *rpctenant.ChangeTenantStatusReq, ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error)
	archiveFn     func(context.Context, *rpctenant.ChangeTenantStatusReq, ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error)
}

func (f *fakeTenantService) CreateTenant(ctx context.Context, in *rpctenant.CreateTenantReq, opts ...grpc.CallOption) (*rpctenant.CreateTenantResp, error) {
	if f.createFn == nil {
		return &rpctenant.CreateTenantResp{}, nil
	}
	return f.createFn(ctx, in, opts...)
}

func (f *fakeTenantService) QueryTenantDetail(ctx context.Context, in *rpctenant.QueryTenantDetailReq, opts ...grpc.CallOption) (*rpctenant.QueryTenantDetailResp, error) {
	if f.queryDetailFn == nil {
		return &rpctenant.QueryTenantDetailResp{}, nil
	}
	return f.queryDetailFn(ctx, in, opts...)
}

func (f *fakeTenantService) QueryTenantList(ctx context.Context, in *rpctenant.QueryTenantListReq, opts ...grpc.CallOption) (*rpctenant.QueryTenantListResp, error) {
	if f.queryListFn == nil {
		return &rpctenant.QueryTenantListResp{}, nil
	}
	return f.queryListFn(ctx, in, opts...)
}

func (f *fakeTenantService) EnableTenant(ctx context.Context, in *rpctenant.ChangeTenantStatusReq, opts ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error) {
	if f.enableFn == nil {
		return &rpctenant.ChangeTenantStatusResp{}, nil
	}
	return f.enableFn(ctx, in, opts...)
}

func (f *fakeTenantService) DisableTenant(ctx context.Context, in *rpctenant.ChangeTenantStatusReq, opts ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error) {
	if f.disableFn == nil {
		return &rpctenant.ChangeTenantStatusResp{}, nil
	}
	return f.disableFn(ctx, in, opts...)
}

func (f *fakeTenantService) ArchiveTenant(ctx context.Context, in *rpctenant.ChangeTenantStatusReq, opts ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error) {
	if f.archiveFn == nil {
		return &rpctenant.ChangeTenantStatusResp{}, nil
	}
	return f.archiveFn(ctx, in, opts...)
}

func tenantTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userName", "platform-admin")
	ctx = context.WithValue(ctx, "userId", json.Number("77"))
	return ctx
}

func TestCreateTenantLogicPassesOperatorContext(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		TenantService: &fakeTenantService{
			createFn: func(_ context.Context, in *rpctenant.CreateTenantReq, _ ...grpc.CallOption) (*rpctenant.CreateTenantResp, error) {
				if in.CreateBy != "platform-admin" || in.OperatorId != 77 {
					t.Fatalf("unexpected operator info: %+v", in)
				}
				if len(in.AvailableChannels) != 2 || in.AvailableChannels[0] != "app" || in.AvailableChannels[1] != "mini_program" {
					t.Fatalf("unexpected channels: %+v", in.AvailableChannels)
				}
				return &rpctenant.CreateTenantResp{
					TenantId:              8,
					TenantCode:            "TEN202603201122330001",
					AdminUserId:           18,
					AdminActivationStatus: "pending-activation",
				}, nil
			},
		},
	}

	logic := NewCreateTenantLogic(tenantTestContext(), svcCtx)
	resp, err := logic.CreateTenant(&types.CreateTenantReq{
		TenantName:        "租户A",
		ContactName:       "张三",
		ContactMobile:     "13800138000",
		AvailableChannels: []string{" app ", "mini-program"},
		DataRetentionDays: 180,
		FeatureFlags:      []string{"oms"},
		AdminUserName:     "tenant_admin",
		AdminNickName:     "租户管理员",
		AdminMobile:       "13900139000",
		AdminPassword:     "123456",
	})
	if err != nil {
		t.Fatalf("create tenant logic failed: %v", err)
	}
	if resp.Code != "000000" || resp.Data.TenantId != 8 || resp.Data.AdminUserId != 18 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestQueryTenantListLogicMapsRpcResponse(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		TenantService: &fakeTenantService{
			queryListFn: func(_ context.Context, in *rpctenant.QueryTenantListReq, _ ...grpc.CallOption) (*rpctenant.QueryTenantListResp, error) {
				if in.Status != 1 || in.Channel != "mini_program" || in.PageNum != 2 {
					t.Fatalf("unexpected query params: %+v", in)
				}
				return &rpctenant.QueryTenantListResp{
					Total: 1,
					List: []*rpctenant.TenantData{
						{
							Id:                    1,
							TenantCode:            "TEN001",
							TenantName:            "租户一",
							AvailableChannels:     []string{"app"},
							FeatureFlags:          []string{"oms"},
							Status:                1,
							PrimaryAdminUserId:    11,
							PrimaryAdminUserName:  "tenant_admin",
							PrimaryAdminMobile:    "13900139000",
							AdminActivationStatus: "active",
						},
					},
				}, nil
			},
		},
	}

	logic := NewQueryTenantListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryTenantList(&types.QueryTenantListReq{
		Current:  2,
		PageSize: 20,
		Status:   1,
		Channel:  "mini-program",
	})
	if err != nil {
		t.Fatalf("query tenant list failed: %v", err)
	}
	if !resp.Success || resp.Total != 1 || len(resp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", resp)
	}
	if resp.Data[0].TenantCode != "TEN001" || resp.Data[0].AdminActivationStatus != "active" {
		t.Fatalf("unexpected list item: %+v", resp.Data[0])
	}
}

func TestEnableTenantLogicPassesStatusCommand(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		TenantService: &fakeTenantService{
			enableFn: func(_ context.Context, in *rpctenant.ChangeTenantStatusReq, _ ...grpc.CallOption) (*rpctenant.ChangeTenantStatusResp, error) {
				if in.UpdateBy != "platform-admin" || in.OperatorId != 77 {
					t.Fatalf("unexpected operator info: %+v", in)
				}
				if len(in.Ids) != 2 || in.StatusReason != "通过验收" {
					t.Fatalf("unexpected status request: %+v", in)
				}
				return &rpctenant.ChangeTenantStatusResp{Pong: "ok"}, nil
			},
		},
	}

	logic := NewEnableTenantLogic(tenantTestContext(), svcCtx)
	resp, err := logic.EnableTenant(&types.ChangeTenantStatusReq{
		Ids:          []int64{1, 2},
		StatusReason: "通过验收",
	})
	if err != nil {
		t.Fatalf("enable tenant failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("unexpected enable response: %+v", resp)
	}
}
