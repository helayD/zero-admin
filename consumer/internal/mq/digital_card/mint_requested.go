package digital_card

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
)

func MintRequested(ctx context.Context, body []byte, service *digitalcardmint.Service) error {
	if service == nil {
		return errors.New("数字卡片发放服务未初始化")
	}

	var event digitalcardmint.MintRequestedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	if event.TaskID <= 0 {
		return errors.New("taskId 不能为空")
	}

	_, err := service.ExecuteTask(ctx, event.TaskID, digitalcardmint.OperatorSystem)
	return err
}

