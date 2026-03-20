package tenantservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateTenantLogic) CreateTenant(in *sysclient.CreateTenantReq) (*sysclient.CreateTenantResp, error) {
	req, err := normalizeCreateTenantReq(in)
	if err != nil {
		return nil, err
	}

	var resp *sysclient.CreateTenantResp
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureTenantNameAvailable(tx, req.TenantName); err != nil {
			return err
		}

		now := time.Now()
		tenantCode, err := createUniqueTenantCode(tx, now)
		if err != nil {
			return err
		}

		user, _, err := findReusableAdminUser(tx, req, now)
		if err != nil {
			return err
		}

		availableChannels, err := encodeStringList(req.AvailableChannels)
		if err != nil {
			return err
		}
		featureFlags, err := encodeStringList(req.FeatureFlags)
		if err != nil {
			return err
		}

		tenant := &tenantmodel.SysTenant{
			PlatformID:        defaultPlatformID,
			TenantCode:        tenantCode,
			TenantName:        req.TenantName,
			TenantShortName:   req.TenantShortName,
			ContactName:       req.ContactName,
			ContactMobile:     req.ContactMobile,
			ContactEmail:      req.ContactEmail,
			AvailableChannels: availableChannels,
			DataRetentionDays: req.DataRetentionDays,
			FeatureFlags:      featureFlags,
			Status:            tenantmodel.TenantStatusPendingActivation,
			StatusReason:      "租户已创建，等待平台启用",
			CreatedBy:         req.CreateBy,
			UpdatedBy:         req.CreateBy,
		}
		if err := tx.Create(tenant).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return errors.New("租户标识冲突，请重试")
			}
			return err
		}

		scopeMetadata, err := scope.BuildTenantAdminMetadata(defaultPlatformID, tenant.ID, tenant.TenantCode, scope.ActivationStatusPending)
		if err != nil {
			return err
		}

		binding := &tenantmodel.SysTenantUser{
			PlatformID:       defaultPlatformID,
			TenantID:         tenant.ID,
			UserID:           user.ID,
			RoleMode:         scope.RoleModeBootstrap,
			ActivationStatus: tenantmodel.TenantStatusPendingActivation,
			IsPrimaryAdmin:   1,
			ScopeMetadata:    scopeMetadata,
			CreatedBy:        req.CreateBy,
			UpdatedBy:        req.CreateBy,
		}
		if err := tx.Create(binding).Error; err != nil {
			return err
		}

		if err := createTenantAuditRows(tx, tenant, user, req); err != nil {
			return err
		}

		resp = &sysclient.CreateTenantResp{
			TenantId:              tenant.ID,
			TenantCode:            tenant.TenantCode,
			AdminUserId:           user.ID,
			AdminActivationStatus: scope.ActivationStatusPending,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
