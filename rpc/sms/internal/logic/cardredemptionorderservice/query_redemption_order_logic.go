package cardredemptionorderservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryRedemptionOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryRedemptionOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRedemptionOrderLogic {
	return &QueryRedemptionOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryRedemptionOrderLogic) QueryRedemptionOrder(in *smsclient.QueryRedemptionOrderReq) (*smsclient.QueryRedemptionOrderResp, error) {
	if in.OrderId <= 0 && in.CardInstanceId <= 0 {
		return nil, errors.New("订单ID或卡片实例ID至少提供一个")
	}

	var row redemptionOrderRow
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table(row.TableName()).
		Where("is_deleted = 0")

	if in.OrderId > 0 {
		query = query.Where("id = ?", in.OrderId)
	}
	if in.CardInstanceId > 0 {
		query = query.Where("card_instance_id = ?", in.CardInstanceId)
	}
	if in.HolderId > 0 {
		query = query.Where("holder_id = ?", in.HolderId)
	}

	if err := query.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("提货单不存在")
		}
		return nil, err
	}

	return &smsclient.QueryRedemptionOrderResp{
		Order: buildRedemptionOrderData(&row),
	}, nil
}
