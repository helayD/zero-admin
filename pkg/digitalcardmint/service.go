package digitalcardmint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/chainclient"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MQPublisher interface {
	SendMessage(exchange, exchangeType, queueName, key string, message []byte) error
}

type Service struct {
	DB                 *gorm.DB
	MQ                 MQPublisher
	Chain              chainclient.ChainClient
	Now                func() time.Time
	MaxRetryCount      int32
	RunningTimeout     time.Duration
	DispatchExchange   string
	DispatchType       string
	DispatchQueue      string
	DispatchRoutingKey string
}

const (
	mintPublishedStatus      int32 = 1
	mintEnabledStatus        int32 = 1
	mintDisplayVisible       int32 = 1
	mintReadinessOffline     int32 = 3
	mintApprovalRejected     int32 = 3
	mintComplianceRejected   int32 = 3
	mintVerifiedRealNameCode       = "verified"
)

type mintEligibilitySnapshot struct {
	Status         string `json:"status"`
	Code           string `json:"code"`
	RealNameStatus string `json:"realNameStatus"`
}

type executeSnapshot struct {
	TaskStatus    string
	LastErrorCode string
	LastExecuteAt *time.Time
	RetryCount    int32
}

func NewService(db *gorm.DB, mq MQPublisher, client chainclient.ChainClient) *Service {
	return &Service{
		DB:                 db,
		MQ:                 mq,
		Chain:              client,
		Now:                time.Now,
		MaxRetryCount:      DefaultMaxRetryCount,
		RunningTimeout:     2 * time.Minute,
		DispatchExchange:   EventExchange,
		DispatchType:       EventExchangeType,
		DispatchQueue:      EventQueue,
		DispatchRoutingKey: EventRoutingKey,
	}
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) EnsureTaskTx(ctx context.Context, tx *gorm.DB, assetInstanceID int64, operatorType string) (*CardMintTaskRow, error) {
	if tx == nil {
		return nil, errors.New("数据库事务不能为空")
	}
	if assetInstanceID <= 0 {
		return nil, errors.New("资产实例ID不能为空")
	}

	instance, err := s.loadCardInstance(ctx, tx, assetInstanceID, true)
	if err != nil {
		return nil, err
	}
	if instance.IsDeleted != 0 {
		return nil, errors.New("资产实例不存在")
	}
	if strings.TrimSpace(instance.AssetStatus) != AssetStatusCreated {
		return nil, errors.New("当前资产状态不允许发链")
	}

	if existing, findErr := s.loadTaskByAssetInstance(ctx, tx, assetInstanceID, true); findErr == nil {
		if instance.MintTaskID == 0 || instance.MintTaskID != existing.ID {
			now := s.now()
			if err = tx.WithContext(ctx).
				Table(instance.TableName()).
				Where("id = ? AND is_deleted = 0", instance.ID).
				Updates(map[string]interface{}{
					"mint_task_id": existing.ID,
					"update_time":  now,
				}).Error; err != nil {
				return nil, err
			}
		}
		return existing, nil
	} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil, findErr
	}

	record, err := s.loadParticipationRecord(ctx, tx, instance.ParticipationRecordID)
	if err != nil && !isOrderPurchaseAsset(instance) {
		return nil, err
	}
	if err = s.validateMintPrerequisites(ctx, tx, instance, record); err != nil {
		return nil, err
	}

	currentMintStatus := normalizeMintStatus(instance.MintStatus)
	switch currentMintStatus {
	case MintStatusPending, MintStatusFailed, MintStatusCompensating, "":
	default:
		return nil, fmt.Errorf("当前资产发放状态[%s]不允许创建任务", currentMintStatus)
	}
	if currentMintStatus == "" {
		currentMintStatus = MintStatusPending
	}

	now := s.now()
	maxRetry := s.MaxRetryCount
	if maxRetry <= 0 {
		maxRetry = DefaultMaxRetryCount
	}
	task := &CardMintTaskRow{
		PlatformID:            instance.PlatformID,
		TenantID:              instance.TenantID,
		MerchantID:            instance.MerchantID,
		AssetInstanceID:       instance.ID,
		ParticipationRecordID: instance.ParticipationRecordID,
		ActivityID:            instance.ActivityID,
		MemberID:              instance.MemberID,
		RequestID:             firstNonEmpty(strings.TrimSpace(instance.RequestID), participationRequestID(record)),
		TraceID:               firstNonEmpty(strings.TrimSpace(instance.TraceID), participationTraceID(record)),
		IdempotencyKey:        buildIdempotencyKey(instance.ID),
		TaskStatus:            TaskStatusPendingDispatch,
		MintStatus:            currentMintStatus,
		ChainStatus:           normalizeChainStatus(instance.ChainStatus),
		MaxRetryCount:         maxRetry,
		CreateTime:            now,
		UpdateTime:            &now,
	}
	if task.ChainStatus == "" {
		task.ChainStatus = ChainStatusUnknown
	}

	if err = tx.WithContext(ctx).Table(task.TableName()).Create(task).Error; err != nil {
		return nil, err
	}

	if err = tx.WithContext(ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", instance.ID).
		Updates(map[string]interface{}{
			"mint_task_id": task.ID,
			"mint_status":  currentMintStatus,
			"chain_status": task.ChainStatus,
			"update_time":  now,
		}).Error; err != nil {
		return nil, err
	}

	if err = s.appendAssetLogTx(ctx, tx, instance, "", currentMintStatus, OperationMintRequested, normalizeOperatorType(operatorType), task.TraceID, "", "已创建链上发放任务", map[string]interface{}{
		"taskId":          task.ID,
		"idempotencyKey":  task.IdempotencyKey,
		"taskStatus":      task.TaskStatus,
		"participationId": task.ParticipationRecordID,
	}); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) DispatchTask(ctx context.Context, taskID int64, reason string) error {
	if s.DB == nil {
		return errors.New("数据库未初始化")
	}
	if taskID <= 0 {
		return errors.New("任务ID不能为空")
	}

	now := s.now()
	var (
		task         *CardMintTaskRow
		instance     *CardInstanceRow
		skipDispatch bool
	)
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		task, txErr = s.loadTaskByID(ctx, tx, taskID, true)
		if txErr != nil {
			return txErr
		}
		instance, txErr = s.loadCardInstance(ctx, tx, task.AssetInstanceID, true)
		if txErr != nil {
			return txErr
		}
		if task.IsDeleted != 0 || instance.IsDeleted != 0 {
			return errors.New("任务或资产不存在")
		}
		if shouldSkipDispatchTask(task, now, s.runningTimeout()) {
			skipDispatch = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if skipDispatch {
		return nil
	}

	event := MintRequestedEvent{
		EventName:       EventNameMintRequested,
		TaskID:          task.ID,
		AssetInstanceID: task.AssetInstanceID,
		RequestID:       task.RequestID,
		TraceID:         task.TraceID,
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if s.MQ == nil {
		sendErr := errors.New("RabbitMQ 未配置")
		if err = s.markDispatchFailure(ctx, task, instance, reason, sendErr); err != nil {
			return err
		}
		return sendErr
	}

	if err = s.MQ.SendMessage(s.DispatchExchange, s.DispatchType, s.DispatchQueue, s.DispatchRoutingKey, body); err != nil {
		sendErr := err
		if err = s.markDispatchFailure(ctx, task, instance, reason, sendErr); err != nil {
			return err
		}
		return sendErr
	}

	nextMintStatus := MintStatusProcessing
	nextChainStatus := ChainStatusProcessing
	dispatchReason := firstNonEmpty(reason, "已派发异步发链任务")
	taskUpdates := map[string]interface{}{
		"task_status":  TaskStatusDispatched,
		"mint_status":  nextMintStatus,
		"chain_status": nextChainStatus,
		"update_time":  now,
	}
	if hasReceiptWritebackPending(task) {
		nextMintStatus = MintStatusCompensating
		nextChainStatus = normalizeChainStatus(firstNonEmpty(task.ChainStatus, ChainStatusSuccess))
		dispatchReason = firstNonEmpty(reason, "已派发回执补写任务")
		taskUpdates["mint_status"] = nextMintStatus
		taskUpdates["chain_status"] = nextChainStatus
		taskUpdates["next_retry_at"] = nil
	} else {
		taskUpdates["last_error_code"] = ""
		taskUpdates["last_error_reason"] = ""
		taskUpdates["next_retry_at"] = nil
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(taskUpdates).Error; err != nil {
			return err
		}

		if err = tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status":  nextMintStatus,
				"chain_status": nextChainStatus,
				"update_time":  now,
			}).Error; err != nil {
			return err
		}

		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, nextMintStatus, OperationMintDispatching, OperatorSystem, task.TraceID, "", dispatchReason, map[string]interface{}{
			"taskId":      task.ID,
			"taskStatus":  TaskStatusDispatched,
			"dispatchKey": s.DispatchRoutingKey,
		})
	})
}

func (s *Service) ExecuteTask(ctx context.Context, taskID int64, operatorType string) (*ExecuteResult, error) {
	if s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if taskID <= 0 {
		return nil, errors.New("任务ID不能为空")
	}

	now := s.now()
	timeout := s.runningTimeout()
	var (
		task          *CardMintTaskRow
		instance      *CardInstanceRow
		attempt       int32
		replayReceipt bool
		skipExecute   bool
		snapshot      executeSnapshot
		err           error
	)
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		task, txErr = s.loadTaskByID(ctx, tx, taskID, true)
		if txErr != nil {
			return txErr
		}
		instance, txErr = s.loadCardInstance(ctx, tx, task.AssetInstanceID, true)
		if txErr != nil {
			return txErr
		}

		if task.IsDeleted != 0 || instance.IsDeleted != 0 {
			return errors.New("任务或资产不存在")
		}
		if isTaskTerminal(task) || isTaskExecutionLeased(task, now, timeout) {
			skipExecute = true
			return nil
		}

		snapshot = executeSnapshot{
			TaskStatus:    task.TaskStatus,
			LastErrorCode: task.LastErrorCode,
			LastExecuteAt: cloneTimePointer(task.LastExecuteAt),
			RetryCount:    task.RetryCount,
		}
		replayReceipt = hasReceiptWritebackPending(task)
		if replayReceipt {
			attempt = task.RetryCount
		} else {
			attempt = task.RetryCount + 1
		}
		taskUpdates := map[string]interface{}{
			"task_status":     TaskStatusRunning,
			"last_execute_at": now,
			"update_time":     now,
			"next_retry_at":   nil,
		}
		instanceUpdates := map[string]interface{}{
			"update_time": now,
		}
		if replayReceipt {
			taskUpdates["mint_status"] = MintStatusCompensating
			taskUpdates["chain_status"] = normalizeChainStatus(firstNonEmpty(task.ChainStatus, ChainStatusSuccess))
			instanceUpdates["mint_status"] = MintStatusCompensating
			instanceUpdates["chain_status"] = normalizeChainStatus(firstNonEmpty(task.ChainStatus, ChainStatusSuccess))
		} else {
			taskUpdates["mint_status"] = MintStatusProcessing
			taskUpdates["chain_status"] = ChainStatusProcessing
			taskUpdates["retry_count"] = attempt
			instanceUpdates["mint_status"] = MintStatusProcessing
			instanceUpdates["chain_status"] = ChainStatusProcessing
		}
		if err = tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(taskUpdates).Error; err != nil {
			return err
		}
		if err = tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(instanceUpdates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if skipExecute {
		task, err = s.loadTaskByID(ctx, s.DB, taskID, false)
		if err != nil {
			return nil, err
		}
		return buildExecuteResult(task), nil
	}

	task, err = s.loadTaskByID(ctx, s.DB, taskID, false)
	if err != nil {
		return nil, err
	}
	instance, err = s.loadCardInstance(ctx, s.DB, task.AssetInstanceID, false)
	if err != nil {
		return nil, err
	}

	if isTaskTerminal(task) {
		return buildExecuteResult(task), nil
	}

	var receipt *chainclient.MintTokenResponse
	if replayReceipt {
		receipt = s.buildReceiptFromTask(task)
		if receipt == nil {
			if err = s.transitionReceiptReconcileRequired(ctx, task, instance, normalizeOperatorType(operatorType), "回执补写快照缺失，无法确认历史执行结果", snapshot); err != nil {
				return nil, err
			}
			task, _ = s.loadTaskByID(ctx, s.DB, task.ID, false)
			return buildExecuteResult(task), nil
		}
	} else {
		receipt, skipExecute, err = s.reconcilePreviousReceipt(ctx, task, instance, snapshot, normalizeOperatorType(operatorType))
		if err != nil {
			return nil, err
		}
		if skipExecute {
			task, _ = s.loadTaskByID(ctx, s.DB, task.ID, false)
			return buildExecuteResult(task), nil
		}
		if receipt == nil {
			if err = s.validateExecutionPrerequisites(ctx, task, instance); err != nil {
				if transitionErr := s.transitionPrerequisiteRejected(ctx, task, instance, normalizeOperatorType(operatorType), err); transitionErr != nil {
					return nil, transitionErr
				}
				task, _ = s.loadTaskByID(ctx, s.DB, task.ID, false)
				return buildExecuteResult(task), nil
			}
		}
		if receipt != nil {
			goto finalize
		}
		if s.Chain == nil {
			return nil, errors.New("链客户端未初始化")
		}
		var mintErr error
		receipt, mintErr = s.Chain.MintToken(ctx, &chainclient.MintTokenRequest{
			IdempotencyKey:  task.IdempotencyKey,
			TaskID:          task.ID,
			AssetInstanceID: task.AssetInstanceID,
			ActivityID:      task.ActivityID,
			TemplateID:      instance.TemplateID,
			MemberID:        task.MemberID,
			RequestID:       task.RequestID,
			TraceID:         task.TraceID,
			AssetNo:         instance.AssetNo,
			ScopeType:       resolveScopeType(task.PlatformID, task.TenantID, task.MerchantID),
			PlatformID:      task.PlatformID,
			TenantID:        task.TenantID,
			MerchantID:      task.MerchantID,
		})
		if mintErr != nil {
			if err = s.transitionFailure(ctx, task, instance, attempt, normalizeOperatorType(operatorType), mintErr); err != nil {
				return nil, err
			}
			task, _ = s.loadTaskByID(ctx, s.DB, task.ID, false)
			return buildExecuteResult(task), nil
		}
	}

finalize:
	if err = s.transitionSuccess(ctx, task, instance, receipt, normalizeOperatorType(operatorType)); err != nil {
		if markErr := s.handleSuccessPersistenceFailure(ctx, task, instance, receipt, err); markErr != nil {
			return nil, markErr
		}
		task, _ = s.loadTaskByID(ctx, s.DB, task.ID, false)
		return buildExecuteResult(task), nil
	}

	task, err = s.loadTaskByID(ctx, s.DB, task.ID, false)
	if err != nil {
		return nil, err
	}
	return buildExecuteResult(task), nil
}

func (s *Service) QueryTaskList(ctx context.Context, currentScope pkgscope.GovernanceScope, filter QueryFilter) (int64, []*TaskListItem, error) {
	if s.DB == nil {
		return 0, nil, errors.New("数据库未初始化")
	}
	if filter.PageNum <= 0 {
		filter.PageNum = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	base := s.DB.WithContext(ctx).
		Table("sms_card_mint_task AS task").
		Joins("JOIN sms_card_instance AS instance ON instance.id = task.asset_instance_id AND instance.is_deleted = 0").
		Joins("LEFT JOIN sms_draw_activity AS activity ON activity.id = task.activity_id AND activity.is_deleted = 0").
		Joins("LEFT JOIN sms_card_template AS template ON template.id = instance.template_id AND template.is_deleted = 0").
		Where("task.is_deleted = 0")
	base = pkgscope.ApplyGovernanceScope(base, currentScope, "task")
	base = applyQueryFilter(base, filter)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	type taskListRow struct {
		TaskID                int64      `gorm:"column:task_id"`
		TraceID               string     `gorm:"column:trace_id"`
		AssetInstanceID       int64      `gorm:"column:asset_instance_id"`
		AssetNo               string     `gorm:"column:asset_no"`
		MemberID              int64      `gorm:"column:member_id"`
		ActivityID            int64      `gorm:"column:activity_id"`
		ActivityName          string     `gorm:"column:activity_name"`
		TemplateID            int64      `gorm:"column:template_id"`
		TemplateName          string     `gorm:"column:template_name"`
		TokenID               string     `gorm:"column:token_id"`
		TaskStatus            string     `gorm:"column:task_status"`
		MintStatus            string     `gorm:"column:mint_status"`
		ChainStatus           string     `gorm:"column:chain_status"`
		RetryCount            int32      `gorm:"column:retry_count"`
		LastError             string     `gorm:"column:last_error"`
		ManualRequired        int32      `gorm:"column:manual_required"`
		Frozen                int32      `gorm:"column:frozen"`
		FreezeReason          string     `gorm:"column:freeze_reason"`
		LastReceiptSummary    string     `gorm:"column:last_receipt_summary"`
		ParticipationRecordID int64      `gorm:"column:participation_record_id"`
		AssetStatus           string     `gorm:"column:asset_status"`
		LastExecuteAt         *time.Time `gorm:"column:last_execute_at"`
		SourceType            string     `gorm:"column:source_type"`
		SourceID              int64      `gorm:"column:source_id"`
	}
	var rows []taskListRow
	err := base.Select(`
			task.id AS task_id,
			task.trace_id AS trace_id,
			task.asset_instance_id AS asset_instance_id,
			instance.asset_no AS asset_no,
			task.member_id AS member_id,
			task.activity_id AS activity_id,
			COALESCE(activity.name, '') AS activity_name,
			instance.template_id AS template_id,
			COALESCE(template.template_name, '') AS template_name,
			task.token_id AS token_id,
			task.task_status AS task_status,
			task.mint_status AS mint_status,
			task.chain_status AS chain_status,
			task.retry_count AS retry_count,
			task.last_error_reason AS last_error,
			task.manual_required AS manual_required,
			task.frozen AS frozen,
			task.freeze_reason AS freeze_reason,
			task.last_receipt_summary AS last_receipt_summary,
			task.participation_record_id AS participation_record_id,
			instance.asset_status AS asset_status,
			task.last_execute_at AS last_execute_at,
			instance.source_type AS source_type,
			instance.source_id AS source_id`).
		Order("task.id DESC").
		Offset(int((filter.PageNum - 1) * filter.PageSize)).
		Limit(int(filter.PageSize)).
		Find(&rows).Error
	if err != nil {
		return 0, nil, err
	}

	result := make([]*TaskListItem, 0, len(rows))
	for _, row := range rows {
		item := TaskListItem{
			TaskID:                row.TaskID,
			TraceID:               row.TraceID,
			AssetInstanceID:       row.AssetInstanceID,
			AssetNo:               row.AssetNo,
			MemberID:              row.MemberID,
			ActivityID:            row.ActivityID,
			ActivityName:          row.ActivityName,
			TemplateID:            row.TemplateID,
			TemplateName:          row.TemplateName,
			TokenID:               row.TokenID,
			TaskStatus:            row.TaskStatus,
			TaskStatusText:        taskStatusText(row.TaskStatus),
			MintStatus:            row.MintStatus,
			MintStatusText:        mintStatusText(row.MintStatus),
			ChainStatus:           row.ChainStatus,
			ChainStatusText:       chainStatusText(row.ChainStatus),
			RetryCount:            row.RetryCount,
			LastError:             row.LastError,
			ManualRequired:        row.ManualRequired == 1,
			Frozen:                row.Frozen == 1,
			FreezeReason:          row.FreezeReason,
			LastReceiptSummary:    row.LastReceiptSummary,
			AssetStatusText:       ResolveAssetStatusText(row.AssetStatus, row.MintStatus, row.ChainStatus),
			ParticipationRecordID: row.ParticipationRecordID,
			SourceType:            row.SourceType,
			SourceDisplayName:     resolveSourceDisplayName(row.SourceType, row.ActivityName),
		}
		if row.LastExecuteAt != nil {
			item.LastExecuteAt = row.LastExecuteAt.Format("2006-01-02 15:04:05")
		}
		result = append(result, &item)
	}

	chainTypeVal := s.chainType()
	for _, item := range result {
		item.ChainType = chainTypeVal
	}

	return total, result, nil
}

func (s *Service) QueryTaskDetail(ctx context.Context, currentScope pkgscope.GovernanceScope, taskID int64) (*TaskDetail, error) {
	if taskID <= 0 {
		return nil, errors.New("任务ID不能为空")
	}
	task, err := s.loadTaskByID(ctx, s.DB, taskID, false)
	if err != nil {
		return nil, err
	}
	if err = validateScope(task.PlatformID, task.TenantID, task.MerchantID, currentScope); err != nil {
		return nil, err
	}
	instance, loadErr := s.loadCardInstance(ctx, s.DB, task.AssetInstanceID, false)
	if loadErr != nil {
		return nil, loadErr
	}
	item := TaskListItem{
		TaskID:                task.ID,
		TraceID:               task.TraceID,
		AssetInstanceID:       task.AssetInstanceID,
		AssetNo:               instance.AssetNo,
		MemberID:              task.MemberID,
		ActivityID:            task.ActivityID,
		TemplateID:            instance.TemplateID,
		TokenID:               task.TokenID,
		TaskStatus:            task.TaskStatus,
		TaskStatusText:        taskStatusText(task.TaskStatus),
		MintStatus:            task.MintStatus,
		MintStatusText:        mintStatusText(task.MintStatus),
		ChainStatus:           task.ChainStatus,
		ChainStatusText:       chainStatusText(task.ChainStatus),
		RetryCount:            task.RetryCount,
		LastError:             task.LastErrorReason,
		ManualRequired:        task.ManualRequired == 1,
		Frozen:                task.Frozen == 1,
		FreezeReason:          task.FreezeReason,
		LastReceiptSummary:    task.LastReceiptSummary,
		AssetStatusText:       ResolveAssetStatusText(instance.AssetStatus, task.MintStatus, task.ChainStatus),
		ParticipationRecordID: task.ParticipationRecordID,
	}
	if task.LastExecuteAt != nil {
		item.LastExecuteAt = task.LastExecuteAt.Format("2006-01-02 15:04:05")
	}

	logs, err := s.queryAssetLogs(ctx, task.AssetInstanceID)
	if err != nil {
		return nil, err
	}
	chainTypeVal := s.chainType()
	item.ChainType = chainTypeVal
	return &TaskDetail{
		Item:            item,
		RequestID:       task.RequestID,
		ChainTxID:       task.ChainTxID,
		LastReceiptJSON: task.LastReceiptJSON,
		Logs:            logs,
		ChainType:       chainTypeVal,
	}, nil
}

func (s *Service) QueryAvailableActions(ctx context.Context, currentScope pkgscope.GovernanceScope, taskID int64) ([]string, error) {
	task, err := s.loadTaskByID(ctx, s.DB, taskID, false)
	if err != nil {
		return nil, err
	}
	if err = validateScope(task.PlatformID, task.TenantID, task.MerchantID, currentScope); err != nil {
		return nil, err
	}
	return availableActions(task), nil
}

func (s *Service) RetryTask(ctx context.Context, currentScope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string) (*ActionResult, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("处置原因不能为空")
	}
	var (
		task     *CardMintTaskRow
		instance *CardInstanceRow
		err      error
	)
	now := s.now()
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		task, txErr = s.loadTaskByID(ctx, tx, taskID, true)
		if txErr != nil {
			return txErr
		}
		if txErr = validateScope(task.PlatformID, task.TenantID, task.MerchantID, currentScope); txErr != nil {
			return txErr
		}
		instance, txErr = s.loadCardInstance(ctx, tx, task.AssetInstanceID, true)
		if txErr != nil {
			return txErr
		}
		if task.Frozen == 1 {
			return errors.New("已冻结链路不允许重试")
		}
		if task.TaskStatus == TaskStatusSucceeded {
			return errors.New("已成功链路无需重试")
		}
		if task.TaskStatus == TaskStatusRunning {
			return errors.New("任务正在执行中")
		}

		if err = tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(map[string]interface{}{
				"task_status":     TaskStatusPendingDispatch,
				"mint_status":     MintStatusCompensating,
				"manual_required": 0,
				"frozen":          0,
				"freeze_reason":   "",
				"next_retry_at":   now,
				"last_error_code": task.LastErrorCode,
				"update_by":       operatorID,
				"update_time":     now,
			}).Error; err != nil {
			return err
		}
		if err = tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status": MintStatusCompensating,
				"update_by":   operatorID,
				"update_time": now,
			}).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, MintStatusCompensating, OperationMintRetryRequested, OperatorManual, task.TraceID, "", reason, map[string]interface{}{
			"taskId": task.ID,
		})
	})
	if err != nil {
		return nil, err
	}

	if err = s.DispatchTask(ctx, taskID, "人工重试触发派发"); err != nil {
		return nil, err
	}
	task, err = s.loadTaskByID(ctx, s.DB, taskID, false)
	if err != nil {
		return nil, err
	}
	return buildActionResult(task, s.chainType()), nil
}

func (s *Service) FreezeTask(ctx context.Context, currentScope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string) (*ActionResult, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("处置原因不能为空")
	}
	var (
		task     *CardMintTaskRow
		instance *CardInstanceRow
		err      error
	)
	now := s.now()
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		task, txErr = s.loadTaskByID(ctx, tx, taskID, true)
		if txErr != nil {
			return txErr
		}
		if txErr = validateScope(task.PlatformID, task.TenantID, task.MerchantID, currentScope); txErr != nil {
			return txErr
		}
		instance, txErr = s.loadCardInstance(ctx, tx, task.AssetInstanceID, true)
		if txErr != nil {
			return txErr
		}
		if err = tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(map[string]interface{}{
				"task_status":     TaskStatusFrozen,
				"mint_status":     MintStatusFrozen,
				"chain_status":    ChainStatusFrozen,
				"frozen":          1,
				"freeze_reason":   reason,
				"manual_required": 0,
				"update_by":       operatorID,
				"update_time":     now,
			}).Error; err != nil {
			return err
		}
		if err = tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status":  MintStatusFrozen,
				"chain_status": ChainStatusFrozen,
				"update_time":  now,
			}).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, MintStatusFrozen, OperationMintFrozen, OperatorManual, task.TraceID, "", reason, map[string]interface{}{
			"taskId": task.ID,
		})
	})
	if err != nil {
		return nil, err
	}
	task, err = s.loadTaskByID(ctx, s.DB, taskID, false)
	if err != nil {
		return nil, err
	}
	return buildActionResult(task, s.chainType()), nil
}

func (s *Service) EscalateTask(ctx context.Context, currentScope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string) (*ActionResult, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("处置原因不能为空")
	}
	var (
		task     *CardMintTaskRow
		instance *CardInstanceRow
		err      error
	)
	now := s.now()
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		task, txErr = s.loadTaskByID(ctx, tx, taskID, true)
		if txErr != nil {
			return txErr
		}
		if txErr = validateScope(task.PlatformID, task.TenantID, task.MerchantID, currentScope); txErr != nil {
			return txErr
		}
		instance, txErr = s.loadCardInstance(ctx, tx, task.AssetInstanceID, true)
		if txErr != nil {
			return txErr
		}
		if err = tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(map[string]interface{}{
				"task_status":       TaskStatusManualReview,
				"mint_status":       MintStatusManualReview,
				"manual_required":   1,
				"last_error_reason": reason,
				"update_by":         operatorID,
				"update_time":       now,
			}).Error; err != nil {
			return err
		}
		if err = tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status": MintStatusManualReview,
				"update_time": now,
			}).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, MintStatusManualReview, OperationMintManualReview, OperatorManual, task.TraceID, "", reason, map[string]interface{}{
			"taskId": task.ID,
		})
	})
	if err != nil {
		return nil, err
	}
	task, err = s.loadTaskByID(ctx, s.DB, taskID, false)
	if err != nil {
		return nil, err
	}
	return buildActionResult(task, s.chainType()), nil
}

func (s *Service) ScanDueTasks(ctx context.Context, batchSize int) (*RecoveryStats, error) {
	if batchSize <= 0 {
		batchSize = 50
	}
	now := s.now()
	cutoff := now.Add(-s.runningTimeout())
	var dueTasks []CardMintTaskRow
	err := s.DB.WithContext(ctx).
		Table(CardMintTaskRow{}.TableName()).
		Where("is_deleted = 0 AND frozen = 0 AND manual_required = 0").
		Where(`
			task_status = ? OR
			(task_status = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ?) OR
			(task_status = ? AND (
				(last_execute_at IS NOT NULL AND last_execute_at <= ?) OR
				(last_execute_at IS NULL AND COALESCE(update_time, create_time) <= ?)
			)) OR
			(task_status = ? AND last_execute_at IS NOT NULL AND last_execute_at <= ?)
		`, TaskStatusPendingDispatch, TaskStatusFailed, now, TaskStatusDispatched, cutoff, cutoff, TaskStatusRunning, cutoff).
		Order("id asc").
		Limit(batchSize).
		Find(&dueTasks).Error
	if err != nil {
		return nil, err
	}

	stats := &RecoveryStats{}
	for _, task := range dueTasks {
		switch task.TaskStatus {
		case TaskStatusPendingDispatch:
			if s.MQ != nil {
				if err = s.DispatchTask(ctx, task.ID, "job 扫描补发"); err == nil {
					stats.Dispatched++
				}
				continue
			}
			fallthrough
		case TaskStatusFailed, TaskStatusDispatched, TaskStatusRunning:
			if _, err = s.ExecuteTask(ctx, task.ID, OperatorJob); err == nil {
				stats.Executed++
				updated, loadErr := s.loadTaskByID(ctx, s.DB, task.ID, false)
				if loadErr == nil && updated.TaskStatus == TaskStatusManualReview {
					stats.Escalated++
				}
			}
		}
	}
	return stats, nil
}

func (s *Service) markDispatchFailure(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, reason string, sendErr error) error {
	now := s.now()
	lastErrorCode := ErrorCodeMQDispatchFailed
	chainStatus := normalizeChainStatus(task.ChainStatus)
	instanceUpdates := map[string]interface{}{
		"mint_status": MintStatusCompensating,
		"update_time": now,
	}
	taskUpdates := map[string]interface{}{
		"task_status":       TaskStatusPendingDispatch,
		"mint_status":       MintStatusCompensating,
		"last_error_reason": sendErr.Error(),
		"next_retry_at":     now.Add(1 * time.Minute),
		"update_time":       now,
	}
	if hasReceiptWritebackPending(task) {
		lastErrorCode = ErrorCodeReceiptWritebackFailed
		chainStatus = normalizeChainStatus(firstNonEmpty(task.ChainStatus, ChainStatusSuccess))
		taskUpdates["chain_status"] = chainStatus
		instanceUpdates["chain_status"] = chainStatus
	} else {
		taskUpdates["last_error_code"] = lastErrorCode
	}
	taskUpdates["last_error_code"] = lastErrorCode
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(taskUpdates).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(instanceUpdates).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, MintStatusCompensating, OperationMintRetryRequested, OperatorSystem, task.TraceID, ErrorCodeMQDispatchFailed, firstNonEmpty(reason, sendErr.Error()), map[string]interface{}{
			"taskId": task.ID,
		})
	})
}

func (s *Service) transitionSuccess(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, receipt *chainclient.MintTokenResponse, operatorType string) error {
	if task == nil || instance == nil {
		return errors.New("任务或资产不能为空")
	}
	if receipt == nil {
		return errors.New("链上回执不能为空")
	}
	if strings.TrimSpace(receipt.TokenID) == "" {
		return errors.New("链上回执缺少 token_id")
	}

	chainStatus := normalizeChainStatus(firstNonEmpty(receipt.ChainStatus, ChainStatusSuccess))
	now := s.now()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.ensureTokenOwnershipTx(ctx, tx, task.ID, instance.ID, receipt.TokenID); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(map[string]interface{}{
				"task_status":          TaskStatusSucceeded,
				"mint_status":          MintStatusSuccess,
				"chain_status":         chainStatus,
				"token_id":             receipt.TokenID,
				"chain_tx_id":          receipt.ChainTxID,
				"last_receipt_summary": receipt.ReceiptSummary,
				"last_receipt_json":    receipt.ReceiptJSON,
				"last_error_code":      "",
				"last_error_reason":    "",
				"manual_required":      0,
				"frozen":               0,
				"next_retry_at":        nil,
				"last_execute_at":      now,
				"update_time":          now,
			}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status":     MintStatusSuccess,
				"chain_status":    chainStatus,
				"token_id":        receipt.TokenID,
				"last_receipt_at": receipt.ConfirmedAt,
				"mint_task_id":    task.ID,
				"update_time":     now,
			}).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, MintStatusSuccess, OperationMintSucceeded, operatorType, task.TraceID, "", firstNonEmpty(receipt.ReceiptSummary, "链上发放成功"), map[string]interface{}{
			"taskId":      task.ID,
			"tokenId":     receipt.TokenID,
			"chainTxId":   receipt.ChainTxID,
			"chainStatus": chainStatus,
		})
	})
}

func (s *Service) transitionFailure(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, attempt int32, operatorType string, execErr error) error {
	now := s.now()
	nextRetryAt := now.Add(time.Duration(attempt*5) * time.Minute)
	nextTaskStatus := TaskStatusFailed
	nextMintStatus := MintStatusCompensating
	manualRequired := int32(0)
	if attempt >= maxRetry(task.MaxRetryCount, s.MaxRetryCount) {
		nextTaskStatus = TaskStatusManualReview
		nextMintStatus = MintStatusManualReview
		manualRequired = 1
		nextRetryAt = time.Time{}
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"task_status":       nextTaskStatus,
			"mint_status":       nextMintStatus,
			"chain_status":      ChainStatusFailed,
			"last_error_code":   ErrorCodeMintExecuteFailed,
			"last_error_reason": execErr.Error(),
			"manual_required":   manualRequired,
			"last_execute_at":   now,
			"update_time":       now,
		}
		if !nextRetryAt.IsZero() {
			updates["next_retry_at"] = nextRetryAt
		} else {
			updates["next_retry_at"] = nil
		}
		if err := tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status":  nextMintStatus,
				"chain_status": ChainStatusFailed,
				"update_time":  now,
			}).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, nextMintStatus, OperationMintFailed, operatorType, task.TraceID, ErrorCodeMintExecuteFailed, execErr.Error(), map[string]interface{}{
			"taskId":       task.ID,
			"retryCount":   attempt,
			"manualReview": manualRequired == 1,
		})
	})
}

func (s *Service) reconcilePreviousReceipt(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, snapshot executeSnapshot, operatorType string) (*chainclient.MintTokenResponse, bool, error) {
	if !needsReceiptReconcile(snapshot) {
		return nil, false, nil
	}

	receipt, err := s.queryExistingReceipt(ctx, task)
	if err == nil {
		return receipt, false, nil
	}
	if errors.Is(err, chainclient.ErrReceiptNotFound) && !requiresStrictReceiptReconcile(snapshot) {
		return nil, false, nil
	}

	reason := "历史执行结果待人工对账确认"
	switch {
	case errors.Is(err, chainclient.ErrReceiptNotFound):
		reason = "链上未查询到历史回执，当前任务需人工对账确认"
	case err != nil:
		reason = fmt.Sprintf("链上回执对账失败：%s", err.Error())
	}
	if err = s.transitionReceiptReconcileRequired(ctx, task, instance, operatorType, reason, snapshot); err != nil {
		return nil, true, err
	}
	return nil, true, nil
}

func (s *Service) queryExistingReceipt(ctx context.Context, task *CardMintTaskRow) (*chainclient.MintTokenResponse, error) {
	if s.Chain == nil {
		return nil, errors.New("链客户端未初始化")
	}
	if task == nil {
		return nil, errors.New("任务不能为空")
	}
	return s.Chain.QueryMintToken(ctx, &chainclient.QueryMintTokenRequest{
		IdempotencyKey:  task.IdempotencyKey,
		TaskID:          task.ID,
		AssetInstanceID: task.AssetInstanceID,
		RequestID:       task.RequestID,
		TraceID:         task.TraceID,
	})
}

func (s *Service) validateExecutionPrerequisites(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow) error {
	if task == nil || instance == nil {
		return errors.New("任务或资产不能为空")
	}
	recordID := instance.ParticipationRecordID
	if recordID <= 0 {
		recordID = task.ParticipationRecordID
	}
	record, err := s.loadParticipationRecord(ctx, s.DB, recordID)
	if err != nil {
		return err
	}
	return s.validateMintPrerequisites(ctx, s.DB, instance, record)
}

func (s *Service) transitionReceiptReconcileRequired(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, operatorType string, reason string, snapshot executeSnapshot) error {
	return s.transitionManualReview(ctx, task, instance, operatorType, ErrorCodeReceiptReconcileRequired, reason, normalizeChainStatus(firstNonEmpty(task.ChainStatus, instance.ChainStatus, ChainStatusUnknown)), map[string]interface{}{
		"taskId":         task.ID,
		"previousStatus": snapshot.TaskStatus,
		"lastErrorCode":  snapshot.LastErrorCode,
		"lastExecuteAt":  formatTimePtr(snapshot.LastExecuteAt),
		"retryCount":     snapshot.RetryCount,
	})
}

func (s *Service) transitionPrerequisiteRejected(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, operatorType string, cause error) error {
	return s.transitionManualReview(ctx, task, instance, operatorType, ErrorCodeMintPrerequisiteRejected, cause.Error(), normalizeChainStatus(firstNonEmpty(task.ChainStatus, instance.ChainStatus, ChainStatusUnknown)), map[string]interface{}{
		"taskId":                task.ID,
		"activityId":            task.ActivityID,
		"participationRecordId": task.ParticipationRecordID,
		"assetInstanceId":       task.AssetInstanceID,
	})
}

func (s *Service) transitionManualReview(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, operatorType string, errorCode string, reason string, chainStatus string, payload map[string]interface{}) error {
	now := s.now()
	nextChainStatus := normalizeChainStatus(firstNonEmpty(chainStatus, task.ChainStatus, instance.ChainStatus, ChainStatusUnknown))
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).
			Table(task.TableName()).
			Where("id = ? AND is_deleted = 0", task.ID).
			Updates(map[string]interface{}{
				"task_status":       TaskStatusManualReview,
				"mint_status":       MintStatusManualReview,
				"chain_status":      nextChainStatus,
				"last_error_code":   errorCode,
				"last_error_reason": reason,
				"manual_required":   1,
				"next_retry_at":     nil,
				"last_execute_at":   now,
				"update_time":       now,
			}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Table(instance.TableName()).
			Where("id = ? AND is_deleted = 0", instance.ID).
			Updates(map[string]interface{}{
				"mint_status":  MintStatusManualReview,
				"chain_status": nextChainStatus,
				"update_time":  now,
			}).Error; err != nil {
			return err
		}
		return s.appendAssetLogTx(ctx, tx, instance, task.MintStatus, MintStatusManualReview, OperationMintManualReview, operatorType, task.TraceID, errorCode, reason, payload)
	})
}

func (s *Service) bestEffortMarkWritebackFailure(ctx context.Context, taskID int64, assetInstanceID int64, receipt *chainclient.MintTokenResponse, cause error) error {
	if s.DB == nil {
		return cause
	}
	if receipt == nil {
		return cause
	}
	now := s.now()
	chainStatus := normalizeChainStatus(firstNonEmpty(receipt.ChainStatus, ChainStatusSuccess))
	taskErr := s.DB.WithContext(ctx).
		Table(CardMintTaskRow{}.TableName()).
		Where("id = ? AND is_deleted = 0", taskID).
		Updates(map[string]interface{}{
			"task_status":          TaskStatusFailed,
			"mint_status":          MintStatusCompensating,
			"chain_status":         chainStatus,
			"token_id":             receipt.TokenID,
			"chain_tx_id":          receipt.ChainTxID,
			"last_receipt_summary": receipt.ReceiptSummary,
			"last_receipt_json":    receipt.ReceiptJSON,
			"last_error_code":      ErrorCodeReceiptWritebackFailed,
			"last_error_reason":    cause.Error(),
			"next_retry_at":        now.Add(1 * time.Minute),
			"last_execute_at":      now,
			"update_time":          now,
		}).Error
	_ = s.DB.WithContext(ctx).
		Table(CardInstanceRow{}.TableName()).
		Where("id = ? AND is_deleted = 0", assetInstanceID).
		Updates(map[string]interface{}{
			"mint_status":     MintStatusCompensating,
			"chain_status":    chainStatus,
			"token_id":        receipt.TokenID,
			"last_receipt_at": receipt.ConfirmedAt,
			"update_time":     now,
		}).Error
	if taskErr != nil {
		return cause
	}
	return nil
}

func (s *Service) handleSuccessPersistenceFailure(ctx context.Context, task *CardMintTaskRow, instance *CardInstanceRow, receipt *chainclient.MintTokenResponse, cause error) error {
	if isTokenBindingConflictError(cause) {
		return s.markTokenBindingConflict(ctx, task.ID, instance.ID, receipt, cause)
	}
	return s.bestEffortMarkWritebackFailure(ctx, task.ID, instance.ID, receipt, cause)
}

func (s *Service) markTokenBindingConflict(ctx context.Context, taskID int64, assetInstanceID int64, receipt *chainclient.MintTokenResponse, cause error) error {
	if s.DB == nil || receipt == nil {
		return cause
	}

	now := s.now()
	chainStatus := normalizeChainStatus(firstNonEmpty(receipt.ChainStatus, ChainStatusSuccess))
	taskErr := s.DB.WithContext(ctx).
		Table(CardMintTaskRow{}.TableName()).
		Where("id = ? AND is_deleted = 0", taskID).
		Updates(map[string]interface{}{
			"task_status":          TaskStatusManualReview,
			"mint_status":          MintStatusManualReview,
			"chain_status":         chainStatus,
			"chain_tx_id":          receipt.ChainTxID,
			"last_receipt_summary": receipt.ReceiptSummary,
			"last_receipt_json":    receipt.ReceiptJSON,
			"last_error_code":      ErrorCodeTokenBindingConflict,
			"last_error_reason":    cause.Error(),
			"manual_required":      1,
			"next_retry_at":        nil,
			"last_execute_at":      now,
			"update_time":          now,
		}).Error
	_ = s.DB.WithContext(ctx).
		Table(CardInstanceRow{}.TableName()).
		Where("id = ? AND is_deleted = 0", assetInstanceID).
		Updates(map[string]interface{}{
			"mint_status": MintStatusManualReview,
			"update_time": now,
		}).Error
	if taskErr != nil {
		return cause
	}
	return nil
}

func (s *Service) buildReceiptFromTask(task *CardMintTaskRow) *chainclient.MintTokenResponse {
	if task == nil || strings.TrimSpace(task.TokenID) == "" {
		return nil
	}

	confirmedAt := s.now()
	if task.LastExecuteAt != nil && !task.LastExecuteAt.IsZero() {
		confirmedAt = *task.LastExecuteAt
	} else if task.UpdateTime != nil && !task.UpdateTime.IsZero() {
		confirmedAt = *task.UpdateTime
	}

	return &chainclient.MintTokenResponse{
		TokenID:        strings.TrimSpace(task.TokenID),
		ChainTxID:      strings.TrimSpace(task.ChainTxID),
		ChainStatus:    normalizeChainStatus(firstNonEmpty(task.ChainStatus, ChainStatusSuccess)),
		ReceiptSummary: strings.TrimSpace(task.LastReceiptSummary),
		ReceiptJSON:    strings.TrimSpace(task.LastReceiptJSON),
		ConfirmedAt:    confirmedAt,
	}
}

func (s *Service) ensureTokenOwnershipTx(ctx context.Context, tx *gorm.DB, taskID int64, assetInstanceID int64, tokenID string) error {
	value := strings.TrimSpace(tokenID)
	if value == "" {
		return errors.New("链上回执缺少 token_id")
	}

	var taskRow CardMintTaskRow
	if err := tx.WithContext(ctx).
		Table(taskRow.TableName()).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_id = ? AND is_deleted = 0 AND id <> ?", value, taskID).
		Take(&taskRow).Error; err == nil {
		return &tokenBindingConflictError{TokenID: value, OwnerTaskID: taskRow.ID}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var instanceRow CardInstanceRow
	if err := tx.WithContext(ctx).
		Table(instanceRow.TableName()).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_id = ? AND is_deleted = 0 AND id <> ?", value, assetInstanceID).
		Take(&instanceRow).Error; err == nil {
		return &tokenBindingConflictError{TokenID: value, OwnerAssetID: instanceRow.ID}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (s *Service) runningTimeout() time.Duration {
	if s != nil && s.RunningTimeout > 0 {
		return s.RunningTimeout
	}
	return 2 * time.Minute
}

func shouldSkipDispatchTask(task *CardMintTaskRow, now time.Time, timeout time.Duration) bool {
	if isTaskTerminal(task) {
		return true
	}
	switch task.TaskStatus {
	case TaskStatusDispatched, TaskStatusRunning:
		return true
	case TaskStatusPendingDispatch, TaskStatusFailed:
		return false
	default:
		return isTaskExecutionLeased(task, now, timeout)
	}
}

func isTaskTerminal(task *CardMintTaskRow) bool {
	if task == nil {
		return false
	}
	if task.Frozen == 1 || task.TaskStatus == TaskStatusFrozen || task.MintStatus == MintStatusFrozen {
		return true
	}
	if task.ManualRequired == 1 || task.TaskStatus == TaskStatusManualReview || task.MintStatus == MintStatusManualReview {
		return true
	}
	return task.TaskStatus == TaskStatusSucceeded || (task.MintStatus == MintStatusSuccess && strings.TrimSpace(task.TokenID) != "")
}

func isTaskExecutionLeased(task *CardMintTaskRow, now time.Time, timeout time.Duration) bool {
	if task == nil || task.LastExecuteAt == nil {
		return false
	}
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	return (task.TaskStatus == TaskStatusRunning || task.TaskStatus == TaskStatusDispatched) && now.Sub(*task.LastExecuteAt) < timeout
}

func hasReceiptWritebackPending(task *CardMintTaskRow) bool {
	return task != nil &&
		strings.TrimSpace(task.LastErrorCode) == ErrorCodeReceiptWritebackFailed &&
		strings.TrimSpace(task.TokenID) != ""
}

func needsReceiptReconcile(snapshot executeSnapshot) bool {
	if snapshot.LastExecuteAt != nil && !snapshot.LastExecuteAt.IsZero() {
		return true
	}
	return snapshot.RetryCount > 0 || strings.TrimSpace(snapshot.TaskStatus) == TaskStatusRunning
}

func requiresStrictReceiptReconcile(snapshot executeSnapshot) bool {
	status := strings.TrimSpace(snapshot.TaskStatus)
	if status == TaskStatusRunning || status == TaskStatusDispatched {
		return true
	}
	if snapshot.LastExecuteAt == nil || snapshot.LastExecuteAt.IsZero() {
		return false
	}
	errorCode := strings.TrimSpace(snapshot.LastErrorCode)
	if errorCode == "" {
		return true
	}
	return errorCode != ErrorCodeMintExecuteFailed && errorCode != ErrorCodeReceiptWritebackFailed
}

func isTokenBindingConflictError(err error) bool {
	if err == nil {
		return false
	}
	var conflict *tokenBindingConflictError
	if errors.As(err, &conflict) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "uk_card_mint_task_token_guard") ||
		strings.Contains(lower, "uk_card_instance_token_guard") ||
		(strings.Contains(lower, "duplicate") && strings.Contains(lower, "token"))
}

type tokenBindingConflictError struct {
	TokenID      string
	OwnerTaskID  int64
	OwnerAssetID int64
}

func (e *tokenBindingConflictError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.OwnerAssetID > 0:
		return fmt.Sprintf("token[%s] 已绑定资产实例[%d]", e.TokenID, e.OwnerAssetID)
	case e.OwnerTaskID > 0:
		return fmt.Sprintf("token[%s] 已绑定发放任务[%d]", e.TokenID, e.OwnerTaskID)
	default:
		return fmt.Sprintf("token[%s] 已被其他链路占用", e.TokenID)
	}
}

func (s *Service) queryAssetLogs(ctx context.Context, assetInstanceID int64) ([]AssetLogItem, error) {
	var rows []CardAssetLogRow
	if err := s.DB.WithContext(ctx).
		Table(CardAssetLogRow{}.TableName()).
		Where("asset_instance_id = ?", assetInstanceID).
		Order("id desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]AssetLogItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, AssetLogItem{
			OperationType: row.OperationType,
			OperatorType:  row.OperatorType,
			FromStatus:    row.FromStatus,
			ToStatus:      row.ToStatus,
			ReasonText:    row.ReasonText,
			TraceID:       row.TraceID,
			PayloadJSON:   row.PayloadJSON,
			CreateTime:    row.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *Service) appendAssetLogTx(ctx context.Context, tx *gorm.DB, instance *CardInstanceRow, fromStatus, toStatus, operationType, operatorType, traceID, reasonCode, reasonText string, payload interface{}) error {
	body := ""
	if payload != nil {
		if raw, err := json.Marshal(payload); err == nil {
			body = string(raw)
		}
	}
	// Story 10.7 Review Fix: 显式注入 CreateTime，避免 time.Time 零值
	// 在 MySQL NO_ZERO_DATE 严格模式下被写成 '0000-00-00 00:00:00' 而被拒绝
	// （历史 bug：表 DDL 有 DEFAULT CURRENT_TIMESTAMP，但 GORM 仍会显式发送 '0001-01-01' 触发严格模式失败）
	logRow := &CardAssetLogRow{
		AssetInstanceID:       instance.ID,
		ParticipationRecordID: instance.ParticipationRecordID,
		FromStatus:            strings.TrimSpace(fromStatus),
		ToStatus:              strings.TrimSpace(toStatus),
		OperationType:         strings.TrimSpace(operationType),
		OperatorType:          normalizeOperatorType(operatorType),
		TraceID:               strings.TrimSpace(traceID),
		ReasonCode:            strings.TrimSpace(reasonCode),
		ReasonText:            strings.TrimSpace(reasonText),
		PayloadJSON:           body,
		CreateTime:            s.now(),
	}
	return tx.WithContext(ctx).Table(logRow.TableName()).Create(logRow).Error
}

func (s *Service) loadTaskByAssetInstance(ctx context.Context, db *gorm.DB, assetInstanceID int64, forUpdate bool) (*CardMintTaskRow, error) {
	var row CardMintTaskRow
	query := db.WithContext(ctx).Table(row.TableName()).Where("asset_instance_id = ? AND is_deleted = 0", assetInstanceID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadTaskByID(ctx context.Context, db *gorm.DB, taskID int64, forUpdate bool) (*CardMintTaskRow, error) {
	var row CardMintTaskRow
	query := db.WithContext(ctx).Table(row.TableName()).Where("id = ? AND is_deleted = 0", taskID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadCardInstance(ctx context.Context, db *gorm.DB, assetInstanceID int64, forUpdate bool) (*CardInstanceRow, error) {
	var row CardInstanceRow
	query := db.WithContext(ctx).Table(row.TableName()).Where("id = ? AND is_deleted = 0", assetInstanceID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadParticipationRecord(ctx context.Context, db *gorm.DB, recordID int64) (*ParticipationRecordRow, error) {
	var row ParticipationRecordRow
	if err := db.WithContext(ctx).Table(row.TableName()).Where("id = ? AND is_deleted = 0", recordID).Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadMemberRealNameStatus(ctx context.Context, db *gorm.DB, memberID int64) (string, error) {
	var row MemberIdentityRow
	err := db.WithContext(ctx).
		Table(row.TableName()).
		Select("real_name_status").
		Where("member_id = ? AND is_deleted = 0", memberID).
		Take(&row).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "", nil
	case err != nil:
		return "", err
	default:
		return strings.TrimSpace(row.RealNameStatus), nil
	}
}

func (s *Service) loadDrawActivity(ctx context.Context, db *gorm.DB, activityID int64) (*DrawActivityRow, error) {
	var row DrawActivityRow
	if err := db.WithContext(ctx).Table(row.TableName()).Where("id = ? AND is_deleted = 0", activityID).Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) loadCardTemplate(ctx context.Context, db *gorm.DB, templateID int64) (*CardTemplateRow, error) {
	var row CardTemplateRow
	if err := db.WithContext(ctx).Table(row.TableName()).Where("id = ? AND is_deleted = 0", templateID).Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) validateMintPrerequisites(ctx context.Context, tx *gorm.DB, instance *CardInstanceRow, record *ParticipationRecordRow) error {
	if isOrderPurchaseAsset(instance) {
		template, err := s.loadCardTemplate(ctx, tx, instance.TemplateID)
		if err != nil {
			return err
		}
		if template.Status != mintPublishedStatus {
			return errors.New("所属模板已下线，当前资产不允许发放")
		}
		if template.DisplayStatus != mintDisplayVisible {
			return errors.New("所属模板已隐藏，当前资产不允许发放")
		}
		if template.AuditStatus == mintApprovalRejected || template.ContentAuditStatus == mintComplianceRejected {
			return errors.New("所属模板合规状态禁止发放")
		}
		return nil
	}
	if record == nil {
		return errors.New("参与记录不能为空")
	}
	activity, err := s.loadDrawActivity(ctx, tx, instance.ActivityID)
	if err != nil {
		return err
	}
	if activity.Status != mintPublishedStatus || activity.PublishReadiness == mintReadinessOffline {
		return errors.New("所属活动已下线，当前资产不允许发链")
	}
	if activity.IsEnabled != mintEnabledStatus {
		return errors.New("所属活动已停用，当前资产不允许发链")
	}
	if activity.AuditStatus == mintApprovalRejected ||
		activity.CopyrightStatus == mintComplianceRejected ||
		activity.ContentAuditStatus == mintComplianceRejected {
		return errors.New("所属活动合规状态禁止发链")
	}

	template, err := s.loadCardTemplate(ctx, tx, instance.TemplateID)
	if err != nil {
		return err
	}
	if template.Status != mintPublishedStatus {
		return errors.New("所属模板已下线，当前资产不允许发链")
	}
	if template.DisplayStatus != mintDisplayVisible {
		return errors.New("所属模板已隐藏，当前资产不允许发链")
	}
	if template.AuditStatus == mintApprovalRejected || template.ContentAuditStatus == mintComplianceRejected {
		return errors.New("所属模板合规状态禁止发链")
	}

	if activity.RealNameRequired != mintEnabledStatus {
		return nil
	}
	memberID := record.MemberID
	if memberID <= 0 {
		memberID = instance.MemberID
	}
	realNameStatus, err := s.loadMemberRealNameStatus(ctx, tx, memberID)
	if err != nil {
		return errors.New("实名状态查询失败，当前资产不允许发放")
	}
	if realNameStatus != mintVerifiedRealNameCode {
		return errors.New("实名未通过，当前资产不允许发放")
	}
	return nil
}

func parseMintEligibilitySnapshot(raw string) (*mintEligibilitySnapshot, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("eligibility snapshot is empty")
	}
	var snapshot mintEligibilitySnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func buildIdempotencyKey(assetInstanceID int64) string {
	return fmt.Sprintf("card-mint:%d", assetInstanceID)
}

func isOrderPurchaseAsset(instance *CardInstanceRow) bool {
	return instance != nil && strings.TrimSpace(instance.SourceType) == sourceTypePurchase
}

func participationRequestID(record *ParticipationRecordRow) string {
	if record == nil {
		return ""
	}
	return strings.TrimSpace(record.RequestID)
}

func participationTraceID(record *ParticipationRecordRow) string {
	if record == nil {
		return ""
	}
	return strings.TrimSpace(record.TraceID)
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func normalizeOperatorType(operatorType string) string {
	switch strings.TrimSpace(operatorType) {
	case OperatorJob:
		return OperatorJob
	case OperatorManual:
		return OperatorManual
	default:
		return OperatorSystem
	}
}

func normalizeMintStatus(status string) string {
	switch strings.TrimSpace(status) {
	case MintStatusPending, MintStatusProcessing, MintStatusSuccess, MintStatusFailed, MintStatusCompensating, MintStatusManualReview, MintStatusFrozen:
		return strings.TrimSpace(status)
	default:
		return strings.TrimSpace(status)
	}
}

func normalizeChainStatus(status string) string {
	switch strings.TrimSpace(status) {
	case ChainStatusProcessing, ChainStatusSuccess, ChainStatusFailed, ChainStatusFrozen, ChainStatusUnknown:
		return strings.TrimSpace(status)
	case "":
		return ChainStatusUnknown
	default:
		return strings.TrimSpace(status)
	}
}

func taskStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case TaskStatusPendingDispatch:
		return "待派发"
	case TaskStatusDispatched:
		return "已派发"
	case TaskStatusRunning:
		return "执行中"
	case TaskStatusSucceeded:
		return "已成功"
	case TaskStatusFailed:
		return "待补偿"
	case TaskStatusManualReview:
		return "人工复核"
	case TaskStatusFrozen:
		return "已冻结"
	default:
		return "未知"
	}
}

func mintStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case MintStatusPending:
		return "待发放"
	case MintStatusProcessing:
		return "链上处理中"
	case MintStatusSuccess:
		return "已到账"
	case MintStatusFailed:
		return "发放失败"
	case MintStatusCompensating:
		return "补偿中"
	case MintStatusManualReview:
		return "人工复核中"
	case MintStatusFrozen:
		return "已冻结"
	default:
		return "未知"
	}
}

func chainStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case ChainStatusUnknown:
		return "待确认"
	case ChainStatusProcessing:
		return "处理中"
	case ChainStatusSuccess:
		return "成功"
	case ChainStatusFailed:
		return "失败"
	case ChainStatusFrozen:
		return "已冻结"
	default:
		return "未知"
	}
}

func availableActions(task *CardMintTaskRow) []string {
	if task == nil {
		return []string{}
	}
	if task.Frozen == 1 || task.TaskStatus == TaskStatusFrozen {
		return []string{}
	}
	actions := make([]string, 0, 3)
	if task.TaskStatus != TaskStatusSucceeded && task.TaskStatus != TaskStatusRunning {
		actions = append(actions, "retry")
	}
	if task.TaskStatus != TaskStatusSucceeded && task.ManualRequired == 0 {
		actions = append(actions, "escalate")
	}
	if task.TaskStatus != TaskStatusSucceeded {
		actions = append(actions, "freeze")
	}
	return actions
}

func validateScope(platformID, tenantID, merchantID int64, current pkgscope.GovernanceScope) error {
	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		if current.PlatformID != platformID {
			return errors.New("当前主体无权访问该链路")
		}
	case pkgscope.SubjectTypeTenant:
		if current.PlatformID != platformID || current.TenantID != tenantID {
			return errors.New("当前主体无权访问该链路")
		}
	case pkgscope.SubjectTypeMerchant:
		if current.PlatformID != platformID || current.TenantID != tenantID || current.MerchantID != merchantID {
			return errors.New("当前主体无权访问该链路")
		}
	default:
		return errors.New("当前主体范围无效")
	}
	return nil
}

func applyQueryFilter(db *gorm.DB, filter QueryFilter) *gorm.DB {
	query := db
	if filter.ActivityID > 0 {
		query = query.Where("task.activity_id = ?", filter.ActivityID)
	}
	if strings.TrimSpace(filter.ActivityName) != "" {
		query = query.Where("activity.name LIKE ?", "%"+strings.TrimSpace(filter.ActivityName)+"%")
	}
	if filter.MemberID > 0 {
		query = query.Where("task.member_id = ?", filter.MemberID)
	}
	if filter.TemplateID > 0 {
		query = query.Where("instance.template_id = ?", filter.TemplateID)
	}
	if strings.TrimSpace(filter.AssetNo) != "" {
		query = query.Where("instance.asset_no LIKE ?", "%"+strings.TrimSpace(filter.AssetNo)+"%")
	}
	if strings.TrimSpace(filter.TokenID) != "" {
		query = query.Where("task.token_id LIKE ?", "%"+strings.TrimSpace(filter.TokenID)+"%")
	}
	if strings.TrimSpace(filter.TaskStatus) != "" {
		query = query.Where("task.task_status = ?", strings.TrimSpace(filter.TaskStatus))
	}
	if strings.TrimSpace(filter.MintStatus) != "" {
		query = query.Where("task.mint_status = ?", strings.TrimSpace(filter.MintStatus))
	}
	if strings.TrimSpace(filter.ChainStatus) != "" {
		query = query.Where("task.chain_status = ?", strings.TrimSpace(filter.ChainStatus))
	}
	if filter.ManualRequired == 1 {
		query = query.Where("task.manual_required = 1")
	} else if filter.ManualRequired == 2 {
		query = query.Where("task.manual_required = 0")
	}
	if start, ok := parseStartTime(filter.StartTime); ok {
		query = query.Where("task.create_time >= ?", start)
	}
	if end, ok := parseEndTime(filter.EndTime); ok {
		query = query.Where("task.create_time <= ?", end)
	}
	return query
}

func parseStartTime(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseEndTime(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			if layout == "2006-01-02" {
				return t.Add(23*time.Hour + 59*time.Minute + 59*time.Second), true
			}
			return t, true
		}
	}
	return time.Time{}, false
}

func resolveScopeType(platformID, tenantID, merchantID int64) string {
	if merchantID > 0 {
		return pkgscope.SubjectTypeMerchant
	}
	if tenantID > 0 {
		return pkgscope.SubjectTypeTenant
	}
	if platformID > 0 {
		return pkgscope.SubjectTypePlatform
	}
	return ""
}

func resolveSourceDisplayName(sourceType string, activityName string) string {
	switch strings.TrimSpace(sourceType) {
	case sourceTypePurchase:
		return "订单购买"
	case "draw":
		if activityName != "" {
			return activityName
		}
		return "抽卡获取"
	default:
		return "未知来源"
	}
}

func buildExecuteResult(task *CardMintTaskRow) *ExecuteResult {
	if task == nil {
		return &ExecuteResult{}
	}
	return &ExecuteResult{
		TaskID:         task.ID,
		TokenID:        task.TokenID,
		TaskStatus:     task.TaskStatus,
		MintStatus:     task.MintStatus,
		ChainStatus:    task.ChainStatus,
		RetryCount:     task.RetryCount,
		ManualRequired: task.ManualRequired == 1,
	}
}

func buildActionResult(task *CardMintTaskRow, chainType string) *ActionResult {
	if task == nil {
		return &ActionResult{}
	}
	return &ActionResult{
		TaskID:         task.ID,
		TraceID:        task.TraceID,
		TaskStatus:     task.TaskStatus,
		MintStatus:     task.MintStatus,
		ChainStatus:    task.ChainStatus,
		RetryCount:     task.RetryCount,
		ManualRequired: task.ManualRequired == 1,
		Frozen:         task.Frozen == 1,
		ChainType:      chainType,
	}
}

func maxRetry(taskValue int32, fallback int32) int32 {
	if taskValue > 0 {
		return taskValue
	}
	if fallback > 0 {
		return fallback
	}
	return DefaultMaxRetryCount
}

func (s *Service) chainType() string {
	if s.Chain != nil {
		return s.Chain.ChainType()
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
