// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantLogic) CreateTenant(req *types.CreateTenantReq) (resp *types.CreateTenantResp, err error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.TenantService.CreateTenant(l.ctx, &sysclient.CreateTenantReq{
		TenantName:        strings.TrimSpace(req.TenantName),
		TenantShortName:   strings.TrimSpace(req.TenantShortName),
		ContactName:       strings.TrimSpace(req.ContactName),
		ContactMobile:     strings.TrimSpace(req.ContactMobile),
		ContactEmail:      strings.TrimSpace(req.ContactEmail),
		AvailableChannels: normalizeChannelSlice(req.AvailableChannels),
		DataRetentionDays: req.DataRetentionDays,
		FeatureFlags:      trimStringSlice(req.FeatureFlags),
		AdminUserName:     strings.TrimSpace(req.AdminUserName),
		AdminNickName:     strings.TrimSpace(req.AdminNickName),
		AdminMobile:       strings.TrimSpace(req.AdminMobile),
		AdminEmail:        strings.TrimSpace(req.AdminEmail),
		AdminPassword:     strings.TrimSpace(req.AdminPassword),
		CreateBy:          userName,
		OperatorId:        userID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "创建租户失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.CreateTenantResp{
		Code:    "000000",
		Message: "创建租户成功",
		Data: types.CreateTenantData{
			TenantId:              result.TenantId,
			TenantCode:            result.TenantCode,
			AdminUserId:           result.AdminUserId,
			AdminActivationStatus: result.AdminActivationStatus,
		},
	}, nil
}
