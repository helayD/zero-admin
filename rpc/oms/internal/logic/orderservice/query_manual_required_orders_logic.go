package orderservicelogic

import (
	"context"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryManualRequiredOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryManualRequiredOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryManualRequiredOrdersLogic {
	return &QueryManualRequiredOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryManualRequiredOrders 查询需要人工介入的补偿订单
func (l *QueryManualRequiredOrdersLogic) QueryManualRequiredOrders(in *omsclient.QueryManualRequiredOrdersReq) (*omsclient.QueryManualRequiredOrdersResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		logc.Errorf(l.ctx, "查询人工介入订单scope非法,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	if in.PageNum <= 0 {
		in.PageNum = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	q := pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.OmsOrderMain{}),
		current,
		"",
	).Where("is_deleted = 0 AND manual_required = 1")

	var (
		result []model.OmsOrderMain
		count  int64
	)
	err = q.Session(&gorm.Session{}).Count(&count).Error
	if err == nil {
		err = q.Offset(int((in.PageNum - 1) * in.PageSize)).Limit(int(in.PageSize)).Find(&result).Error
	}
	if err != nil {
		logc.Errorf(l.ctx, "查询人工介入订单列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	list := make([]*omsclient.OrderListData, 0, len(result))
	for _, item := range result {
		orderData := &omsclient.OrderListData{
			Id:          item.ID,
			OrderNo:     item.OrderNo,
			UserId:      item.UserID,
			OrderStatus: item.OrderStatus,
		}
		list = append(list, orderData)
	}

	logc.Infof(l.ctx, "查询人工介入订单完成,total=%d,pageNum=%d,pageSize=%d", count, in.PageNum, in.PageSize)

	return &omsclient.QueryManualRequiredOrdersResp{
		Total: count,
		List:  list,
	}, nil
}
