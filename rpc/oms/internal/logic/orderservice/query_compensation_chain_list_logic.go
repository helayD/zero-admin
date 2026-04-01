package orderservicelogic

import (
	"context"
	"fmt"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryCompensationChainListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCompensationChainListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCompensationChainListLogic {
	return &QueryCompensationChainListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCompensationChainList 查询补偿链路列表（支持一致性筛选+时间范围）
func (l *QueryCompensationChainListLogic) QueryCompensationChainList(in *omsclient.QueryCompensationChainListReq) (*omsclient.QueryCompensationChainListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		logc.Errorf(l.ctx, "查询补偿链路列表scope非法,参数:%+v,异常:%s", in, err.Error())
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
	).Where("is_deleted = 0")

	if in.ConsistencyStage > 0 {
		q = q.Where("consistency_stage = ?", in.ConsistencyStage)
	}
	if in.ConsistencyResult > 0 {
		q = q.Where("consistency_result = ?", in.ConsistencyResult)
	}
	if in.ManualRequired == 1 {
		q = q.Where("manual_required = 1")
	} else if in.ManualRequired == 2 {
		q = q.Where("manual_required = 0")
	}
	if len(in.StartTime) > 0 {
		startTime, parseErr := time.ParseInLocation("2006-01-02 15:04:05", in.StartTime, time.Local)
		if parseErr == nil {
			q = q.Where("create_time >= ?", startTime)
		}
	}
	if len(in.EndTime) > 0 {
		endTime, parseErr := time.ParseInLocation("2006-01-02 15:04:05", in.EndTime+" 23:59:59", time.Local)
		if parseErr == nil {
			q = q.Where("create_time <= ?", endTime)
		}
	}
	if len(in.OrderNo) > 0 {
		q = q.Where("order_no LIKE ?", "%"+in.OrderNo+"%")
	}

	var (
		result []model.OmsOrderMain
		count  int64
	)
	err = q.Session(&gorm.Session{}).Count(&count).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询补偿链路列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	err = q.Offset(int((in.PageNum-1)*in.PageSize)).Limit(int(in.PageSize)).
		Order("id DESC").Find(&result).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询补偿链路列表分页失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	list := make([]*omsclient.ChainMonitorItem, 0, len(result))
	for _, item := range result {
		chainItem := buildChainMonitorItemFromOrder(&item, &current)
		list = append(list, chainItem)
	}

	logc.Infof(l.ctx, "查询补偿链路列表完成,total=%d,pageNum=%d,pageSize=%d", count, in.PageNum, in.PageSize)

	return &omsclient.QueryCompensationChainListResp{
		Total: count,
		List:  list,
	}, nil
}

func buildChainMonitorItemFromOrder(item *model.OmsOrderMain, current *pkgscope.GovernanceScope) *omsclient.ChainMonitorItem {
	return &omsclient.ChainMonitorItem{
		TraceId:       fmt.Sprintf("order-comp-%d", item.ID),
		PlatformId:   item.PlatformID,
		TenantId:     item.TenantID,
		MerchantId:   item.MerchantID,
		ChainType:    int32(omsclient.ChainType_CHAIN_TYPE_COMPENSATION),
		ChainTypeText: "订单超时补偿",
		EntityId:     item.ID,
		EntityNo:     item.OrderNo,
		EntityType:   1,
		Stage:        item.ConsistencyStage,
		StageText:    queryConsistencyStageText(item.ConsistencyStage),
		Result:       item.ConsistencyResult,
		ResultText:   queryConsistencyResultText(item.ConsistencyResult),
		RetryCount:   item.RetryCount,
		LastError:    item.LastError,
		LastExecuteAt: time_util.TimeToString(item.LastCompensationAt),
		CreatedAt:    time_util.TimeToStr(item.CreateTime),
		ActorId:      item.UserID,
		Paused:        item.Paused,
		PauseReason:   item.PauseReason,
	}
}

func queryConsistencyStageText(stage int32) string {
	// ConsistencyStage 严格对齐 7.3B 规范
	// 0=Unknown, 1=Pending, 2=Paying, 3=PayConfirming, 4=Cancelling, 5=Cancelled, 6=Completed, 7=AfterSale, 8=Closed, 9=ManualRequired
	switch stage {
	case 0:
		return "正常"
	case 1:
		return "待处理"
	case 2:
		return "支付中"
	case 3:
		return "支付确认中"
	case 4:
		return "取消补偿中"
	case 5:
		return "已取消/已关闭"
	case 6:
		return "已完成"
	case 7:
		return "售后处理中"
	case 8:
		return "已关闭"
	case 9:
		return "需人工介入"
	default:
		return "未知阶段"
	}
}

func queryConsistencyResultText(result int32) string {
	switch result {
	case 0:
		return "无结果"
	case 1:
		return "处理中"
	case 2:
		return "成功"
	case 3:
		return "失败"
	case 4:
		return "人工处理"
	default:
		return "未知"
	}
}
