package customer_service_workstation

import (
	"context"
	"fmt"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/client/companyaddressservice"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompanyAddressListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompanyAddressListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompanyAddressListLogic {
	return &CompanyAddressListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CompanyAddressList 公司退货地址列表
func (l *CompanyAddressListLogic) CompanyAddressList(req *types.CompanyAddressListReq) (resp *types.CompanyAddressListResp, err error) {
	_, err = admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.CompanyAddressService.QueryCompanyAddressList(l.ctx, &companyaddressservice.QueryCompanyAddressListReq{
		PageNum:       req.Current,
		PageSize:      req.PageSize,
		AddressName:   req.AddressName,
		Name:         req.Name,
		Phone:        req.Phone,
		SendStatus:    2,
		ReceiveStatus: 2,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询公司退货地址列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var list []*types.CompanyAddressItem
	for _, item := range result.List {
		fullAddress := fmt.Sprintf("%s%s%s%s", item.Province, item.City, item.Region, item.DetailAddress)
		defaultStatus := int32(0)
		if item.SendStatus == 1 || item.ReceiveStatus == 1 {
			defaultStatus = 1
		}
		list = append(list, &types.CompanyAddressItem{
			Id:             item.Id,
			AddressName:    item.AddressName,
			ReceiverName:   item.Name,
			Phone:         item.Phone,
			Province:       item.Province,
			City:           item.City,
			Region:         item.Region,
			DetailAddress:  item.DetailAddress,
			FullAddress:    fullAddress,
			DefaultStatus: defaultStatus,
			SendStatus:     item.SendStatus,
			ReceiveStatus:  item.ReceiveStatus,
			CreateTime:     item.CreateTime,
		})
	}

	return &types.CompanyAddressListResp{
		Code:     "000000",
		Message:  "查询成功",
		Data:     list,
		Total:    result.Total,
		Current:  req.Current,
		PageSize: req.PageSize,
		Success:  true,
	}, nil
}
