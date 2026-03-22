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

type RejectMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRejectMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectMerchantLogic {
	return &RejectMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RejectMerchantLogic) RejectMerchant(req *types.ReviewMerchantReq) (resp *types.BaseResp, err error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.MerchantService.RejectMerchant(l.ctx, &sysclient.ReviewMerchantReq{
		Ids:          req.Ids,
		ReviewReason: strings.TrimSpace(req.ReviewReason),
		UpdateBy:     userName,
		OperatorId:   userID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "驳回商户失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.BaseResp{Code: "000000", Message: "驳回成功"}, nil
}
