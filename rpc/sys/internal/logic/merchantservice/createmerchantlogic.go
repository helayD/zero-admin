package merchantservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/audit"
	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMerchantLogic {
	return &CreateMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateMerchantLogic) CreateMerchant(in *sysclient.CreateMerchantReq) (*sysclient.CreateMerchantResp, error) {
	req, err := normalizeCreateMerchantReq(in)
	if err != nil {
		return nil, err
	}

	var resp *sysclient.CreateMerchantResp
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := ensureTenantExists(l.ctx, tx, req.TenantID); err != nil {
			return err
		}
		if err := ensureMerchantNameAvailable(tx, req.TenantID, req.MerchantName); err != nil {
			return err
		}
		if err := ensureMerchantCodeAvailable(tx, req.TenantID, req.MerchantCode); err != nil {
			return err
		}
		if err := ensurePrimaryAdminUserExists(tx, req.PrimaryAdminUserID); err != nil {
			return err
		}

		now := time.Now()
		merchantCode := strings.TrimSpace(req.MerchantCode)
		if merchantCode == "" {
			merchantCode, err = createUniqueMerchantCode(tx, req.TenantID, now)
			if err != nil {
				return err
			}
		}

		availableChannels, err := encodeStringList(req.AvailableChannels)
		if err != nil {
			return err
		}
		capabilityFlags, err := encodeStringList(req.CapabilityFlags)
		if err != nil {
			return err
		}

		merchant := &merchantmodel.SysMerchant{
			PlatformID:         defaultPlatformID,
			TenantID:           req.TenantID,
			MerchantCode:       merchantCode,
			MerchantName:       req.MerchantName,
			MerchantShortName:  req.MerchantShortName,
			ContactName:        req.ContactName,
			ContactMobile:      req.ContactMobile,
			ContactEmail:       req.ContactEmail,
			AvailableChannels:  availableChannels,
			CapabilityFlags:    capabilityFlags,
			ReviewStatus:       merchantmodel.MerchantReviewPending,
			ReviewReason:       "待平台审核",
			BusinessStatus:     merchantmodel.MerchantBusinessPendingActivation,
			StatusReason:       "审核通过后待启用",
			VisibleScopeHint:   req.VisibleScopeHint,
			PrimaryAdminUserID: req.PrimaryAdminUserID,
			Remark:             req.Remark,
			CreatedBy:          req.CreateBy,
			UpdatedBy:          req.CreateBy,
		}
		if err := tx.Create(merchant).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return errors.New("商户编码或商户名称冲突，请重试")
			}
			return err
		}

		if err := upsertPrimaryAdminBinding(l.ctx, tx, merchant, scope.ActivationStatusPending, req.CreateBy); err != nil {
			return err
		}

		traceID := buildTraceID(merchant.ID, now)
		payload, err := audit.EncodeMerchantPayload(audit.MerchantPayload{
			TraceID:               traceID,
			TenantID:              merchant.TenantID,
			MerchantID:            merchant.ID,
			MerchantCode:          merchant.MerchantCode,
			MerchantName:          merchant.MerchantName,
			Action:                audit.ActionMerchantCreated,
			CurrentReviewStatus:   merchant.ReviewStatus,
			CurrentBusinessStatus: merchant.BusinessStatus,
			AvailableChannels:     req.AvailableChannels,
			CapabilityFlags:       req.CapabilityFlags,
			PrimaryAdminUserID:    merchant.PrimaryAdminUserID,
			VisibleScopeHint:      merchant.VisibleScopeHint,
			Result:                "success",
		})
		if err != nil {
			return err
		}

		afterReviewStatus := merchant.ReviewStatus
		afterBusinessStatus := merchant.BusinessStatus
		if err := recordMerchantAudit(tx, merchant, traceID, audit.ActionMerchantCreated, payload, payload, nil, &afterReviewStatus, nil, &afterBusinessStatus, req.OperatorID, req.CreateBy); err != nil {
			return err
		}

		resp = &sysclient.CreateMerchantResp{
			MerchantId:     merchant.ID,
			MerchantCode:   merchant.MerchantCode,
			ReviewStatus:   merchant.ReviewStatus,
			BusinessStatus: merchant.BusinessStatus,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
