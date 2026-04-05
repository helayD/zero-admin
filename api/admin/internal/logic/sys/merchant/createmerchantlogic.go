// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package merchant

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

type CreateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMerchantLogic {
	return &CreateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMerchantLogic) CreateMerchant(req *types.CreateMerchantReq) (resp *types.CreateMerchantResp, err error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MerchantService.CreateMerchant(l.ctx, &sysclient.CreateMerchantReq{
		TenantId:           req.TenantId,
		MerchantName:       strings.TrimSpace(req.MerchantName),
		MerchantShortName:  strings.TrimSpace(req.MerchantShortName),
		MerchantCode:       strings.TrimSpace(req.MerchantCode),
		ContactName:        strings.TrimSpace(req.ContactName),
		ContactMobile:      strings.TrimSpace(req.ContactMobile),
		ContactEmail:       strings.TrimSpace(req.ContactEmail),
		AvailableChannels:  normalizeChannelSlice(req.AvailableChannels),
		CapabilityFlags:    trimStringSlice(req.CapabilityFlags),
		VisibleScopeHint:   strings.TrimSpace(req.VisibleScopeHint),
		PrimaryAdminUserId: req.PrimaryAdminUserId,
		Remark:             strings.TrimSpace(req.Remark),
		CreateBy:           userName,
		OperatorId:         userID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "创建商户失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.CreateMerchantResp{
		Code:    "000000",
		Message: "创建商户成功",
		Data: types.CreateMerchantData{
			MerchantId:     result.MerchantId,
			MerchantCode:   result.MerchantCode,
			ReviewStatus:   result.ReviewStatus,
			BusinessStatus: result.BusinessStatus,
		},
	}, nil
}
