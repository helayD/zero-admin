package digital_card_chain

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

const digitalCardChainBusinessType int32 = 2

func resolveDigitalCardChainWriteScope(ctx context.Context, requested admincommon.RequestedGovernanceScope) (pkgscope.GovernanceScope, error) {
	current, err := admincommon.ResolveWriteGovernanceScope(ctx, requested)
	if err != nil {
		return pkgscope.GovernanceScope{}, errorx.NewDefaultError(err.Error())
	}
	return current, nil
}

func validateActionReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errorx.NewDefaultError("处置原因不能为空")
	}
	return nil
}

func mapTaskItem(item *digitalcardmint.TaskListItem) *types.DigitalCardChainItem {
	if item == nil {
		return &types.DigitalCardChainItem{}
	}
	return &types.DigitalCardChainItem{
		TaskId:                item.TaskID,
		TraceId:               item.TraceID,
		AssetInstanceId:       item.AssetInstanceID,
		AssetNo:               item.AssetNo,
		MemberId:              item.MemberID,
		ActivityId:            item.ActivityID,
		ActivityName:          item.ActivityName,
		TemplateId:            item.TemplateID,
		TemplateName:          item.TemplateName,
		TokenId:               item.TokenID,
		TaskStatus:            item.TaskStatus,
		TaskStatusText:        item.TaskStatusText,
		MintStatus:            item.MintStatus,
		MintStatusText:        item.MintStatusText,
		ChainStatus:           item.ChainStatus,
		ChainStatusText:       item.ChainStatusText,
		RetryCount:            item.RetryCount,
		LastError:             item.LastError,
		LastExecuteAt:         item.LastExecuteAt,
		ManualRequired:        item.ManualRequired,
		Frozen:                item.Frozen,
		FreezeReason:          item.FreezeReason,
		LastReceiptSummary:    item.LastReceiptSummary,
		AssetStatusText:       item.AssetStatusText,
		ParticipationRecordId: item.ParticipationRecordID,
		ChainType:             item.ChainType,
	}
}

func writeDigitalCardChainOperateLog(ctx context.Context, svcCtx *svc.ServiceContext, operatorId int64, action string, taskID int64, reason string, result *digitalcardmint.ActionResult) {
	if operatorId <= 0 || svcCtx == nil || svcCtx.Operatelogservice == nil || result == nil {
		return
	}
	current, _ := admincommon.CurrentGovernanceScope(ctx)
	operateParam := marshalOperateLogJSON(map[string]interface{}{
		"taskId": taskID,
		"reason": reason,
	})
	jsonResult := marshalOperateLogJSON(map[string]interface{}{
		"taskStatus":  result.TaskStatus,
		"mintStatus":  result.MintStatus,
		"chainStatus": result.ChainStatus,
	})
	extra := marshalOperateLogJSON(map[string]interface{}{
		"action": action,
		"taskId": taskID,
	})
	_, _ = svcCtx.Operatelogservice.AddOperateLog(ctx, &sysclient.AddOperateLogReq{
		Title:         "数字卡片链路-" + action,
		BusinessType:  digitalCardChainBusinessType,
		Method:        "/api/sms/digitalCardChain/" + action,
		RequestMethod: "POST",
		OperatorType:  1,
		OperateUrl:    "/api/sms/digitalCardChain/" + action,
		Platform:      "admin",
		Status:        0,
		OperateTime:   time.Now().Format("2006-01-02 15:04:05"),
		OperateName:   strconv.FormatInt(operatorId, 10),
		DeptName:      strconv.FormatInt(current.TenantID, 10),
		OperateParam:  operateParam,
		JsonResult:    jsonResult,
		Extra:         extra,
	})
}

func marshalOperateLogJSON(payload interface{}) string {
	body, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(body)
}
