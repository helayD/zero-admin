// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package merchant

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMerchantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMerchantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMerchantListLogic {
	return &QueryMerchantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMerchantListLogic) QueryMerchantList(req *types.QueryMerchantListReq) (resp *types.QueryMerchantListResp, err error) {
	result, err := l.svcCtx.MerchantService.QueryMerchantList(l.ctx, &sysclient.QueryMerchantListReq{
		TenantId:       req.TenantId,
		MerchantName:   strings.TrimSpace(req.MerchantName),
		MerchantCode:   strings.TrimSpace(req.MerchantCode),
		ReviewStatus:   req.ReviewStatus,
		BusinessStatus: req.BusinessStatus,
		Channel:        normalizeChannelFilterValue(req.Channel),
		CapabilityFlag: strings.TrimSpace(req.CapabilityFlag),
		PageNum:        req.Current,
		PageSize:       req.PageSize,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询商户列表失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	list := make([]*types.MerchantData, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, mapMerchantData(item))
	}

	return &types.QueryMerchantListResp{
		Code:     "000000",
		Message:  "查询商户列表成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
