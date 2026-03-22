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

type ApproveMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveMerchantLogic {
	return &ApproveMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveMerchantLogic) ApproveMerchant(req *types.ReviewMerchantReq) (resp *types.BaseResp, err error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.MerchantService.ApproveMerchant(l.ctx, &sysclient.ReviewMerchantReq{
		Ids:          req.Ids,
		ReviewReason: strings.TrimSpace(req.ReviewReason),
		UpdateBy:     userName,
		OperatorId:   userID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "审核通过商户失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.BaseResp{Code: "000000", Message: "审核通过成功"}, nil
}
