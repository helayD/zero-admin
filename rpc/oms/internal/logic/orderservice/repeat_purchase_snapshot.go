package orderservicelogic

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
)

type repeatPurchaseEvent struct {
	Order             repeatPurchaseQualifiedOrderRow
	FirstValidPayTime time.Time
	RepeatDays        float64
}

type repeatPurchaseSnapshot struct {
	Overview          *omsclient.RepeatPurchaseOverview
	Trends            []*omsclient.RepeatPurchaseTrendPoint
	DetailRows        []*omsclient.RepeatPurchaseDetailRow
	TrackingStartedAt string
	PartialMetrics    []string
}

type repeatPurchaseTrendAccumulator struct {
	BucketLabel      string
	BucketStart      string
	BucketEnd        string
	PaidBuyers       map[int64]struct{}
	RepeatBuyers     map[int64]struct{}
	RepeatOrderCount int64
	RepeatGmv        float64
	RepeatDaysSum    float64
	RepeatDaysCount  int64
}

type repeatPurchaseMemberAccumulator struct {
	MemberID            int64
	FirstValidPayTime   time.Time
	LatestRepeatPayTime time.Time
	RepeatOrderCount    int64
	RepeatGmv           float64
	LatestChannel       string
	LatestActivityType  string
	LatestActivityID    int64
	PlatformID          int64
	TenantID            int64
	MerchantID          int64
}

func (b *repeatPurchaseQueryBuilder) BuildRepeatPurchaseSnapshot(filter repeatPurchaseFilter, bucket string) (*repeatPurchaseSnapshot, error) {
	currentOrders, err := b.queryCurrentAnalysisOrders(filter)
	if err != nil {
		return nil, err
	}
	memberIDs := collectRepeatPurchaseMemberIDs(currentOrders)
	historyOrders, err := b.queryScopeEffectiveOrders(filter.Scope, memberIDs, filter.EndTime)
	if err != nil {
		return nil, err
	}

	historyByMember := make(map[int64][]repeatPurchaseQualifiedOrderRow, len(memberIDs))
	firstValidByMember := make(map[int64]time.Time, len(memberIDs))
	for _, row := range historyOrders {
		historyByMember[row.MemberID] = append(historyByMember[row.MemberID], row)
		if existing, ok := firstValidByMember[row.MemberID]; !ok || row.PayTime.Before(existing) {
			firstValidByMember[row.MemberID] = row.PayTime
		}
	}

	events := make([]repeatPurchaseEvent, 0)
	for _, row := range currentOrders {
		firstValid, ok := firstValidByMember[row.MemberID]
		if !ok {
			firstValid = row.PayTime
		}
		if hasPriorRepeatOrder(row, historyByMember[row.MemberID]) {
			events = append(events, repeatPurchaseEvent{
				Order:             row,
				FirstValidPayTime: firstValid,
				RepeatDays:        row.PayTime.Sub(firstValid).Hours() / 24,
			})
		}
	}

	trackingStartedAt, partialMetrics, err := b.queryRepeatPurchaseTrackingState(filter)
	if err != nil {
		return nil, err
	}

	trends := buildRepeatPurchaseTrendPoints(filter.StartTime, filter.EndTime, bucket, currentOrders, events)
	detailRows, overview := buildRepeatPurchaseDetailRowsAndOverview(events, currentOrders)
	return &repeatPurchaseSnapshot{
		Overview:          overview,
		Trends:            trends,
		DetailRows:        detailRows,
		TrackingStartedAt: trackingStartedAt,
		PartialMetrics:    partialMetrics,
	}, nil
}

func (b *repeatPurchaseQueryBuilder) queryCurrentAnalysisOrders(filter repeatPurchaseFilter) ([]repeatPurchaseQualifiedOrderRow, error) {
	rows := make([]repeatPurchaseQualifiedOrderRow, 0)
	query := b.db.WithContext(b.ctx).
		Table("oms_order_main curr").
		Select(repeatPurchaseSelectColumns("curr"))
	query = applyRepeatPurchaseCurrentOrderFilters(query, filter, "curr")
	err := query.Order("curr.user_id ASC").Order("curr.pay_time ASC").Order("curr.id ASC").Scan(&rows).Error
	return rows, err
}

func (b *repeatPurchaseQueryBuilder) queryScopeEffectiveOrders(scope pkgscope.GovernanceScope, memberIDs []int64, endTime time.Time) ([]repeatPurchaseQualifiedOrderRow, error) {
	rows := make([]repeatPurchaseQualifiedOrderRow, 0)
	if len(memberIDs) == 0 {
		return rows, nil
	}
	scopeSQL, scopeArgs := pkgscope.ScopeFilterSQL("hist", scope)
	query := b.db.WithContext(b.ctx).
		Table("oms_order_main hist").
		Select(repeatPurchaseSelectColumns("hist")).
		Where("hist.is_deleted = ?", 0).
		Where("hist.pay_time IS NOT NULL").
		Where("hist.order_status IN ?", []int32{2, 3, 4, 7}).
		Where("hist.user_id IN ?", memberIDs).
		Where("hist.pay_time < ?", endTime).
		Where(scopeSQL, scopeArgs...)
	err := query.Order("hist.user_id ASC").Order("hist.pay_time ASC").Order("hist.id ASC").Scan(&rows).Error
	return rows, err
}

func (b *repeatPurchaseQueryBuilder) queryRepeatPurchaseTrackingState(filter repeatPurchaseFilter) (string, []string, error) {
	activityRequested := false
	if activityType, ok := operatefunnel.OptionalActivityType(filter.ActivityType); ok {
		activityRequested = activityType != ""
	}
	if !activityRequested && filter.ActivityID <= 0 {
		return "", nil, nil
	}

	var value interface{}
	row := b.db.WithContext(b.ctx).
		Table("oms_order_main").
		Select("MIN(pay_time)").
		Where("is_deleted = 0").
		Where("pay_time IS NOT NULL").
		Where("activity_type <> '' AND activity_type <> 'none'").
		Row()
	err := row.Scan(&value)
	if err != nil {
		return "", nil, err
	}
	trackingTime, ok, err := normalizeRepeatPurchaseTrackingTime(value)
	if err != nil {
		return "", nil, err
	}
	if !ok {
		return "", nil, nil
	}
	trackingStartedAt := operatefunnel.FormatDateTime(trackingTime)
	partialMetrics := operatefunnel.RepeatPurchasePartialMetrics(trackingTime, filter.StartTime, filter.ActivityType, filter.ActivityID)
	return trackingStartedAt, partialMetrics, nil
}

func buildRepeatPurchaseTrendPoints(startTime, endTime time.Time, bucket string, currentOrders []repeatPurchaseQualifiedOrderRow, events []repeatPurchaseEvent) []*omsclient.RepeatPurchaseTrendPoint {
	windows := operatefunnel.BuildBucketWindows(startTime, endTime, bucket)
	accumulators := make(map[string]*repeatPurchaseTrendAccumulator, len(windows))
	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		accumulators[key] = &repeatPurchaseTrendAccumulator{
			BucketLabel:  window.Label,
			BucketStart:  key,
			BucketEnd:    operatefunnel.FormatDateTime(window.End),
			PaidBuyers:   make(map[int64]struct{}),
			RepeatBuyers: make(map[int64]struct{}),
		}
	}
	for _, row := range currentOrders {
		key := operatefunnel.FormatDateTime(repeatPurchaseBucketStart(row.PayTime, bucket))
		if acc, ok := accumulators[key]; ok {
			acc.PaidBuyers[row.MemberID] = struct{}{}
		}
	}
	for _, event := range events {
		key := operatefunnel.FormatDateTime(repeatPurchaseBucketStart(event.Order.PayTime, bucket))
		if acc, ok := accumulators[key]; ok {
			acc.RepeatBuyers[event.Order.MemberID] = struct{}{}
			acc.RepeatOrderCount++
			acc.RepeatGmv += event.Order.PayAmount
			acc.RepeatDaysSum += event.RepeatDays
			acc.RepeatDaysCount++
		}
	}
	result := make([]*omsclient.RepeatPurchaseTrendPoint, 0, len(windows))
	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		acc := accumulators[key]
		avgDays := 0.0
		if acc.RepeatDaysCount > 0 {
			avgDays = acc.RepeatDaysSum / float64(acc.RepeatDaysCount)
		}
		paidBuyerCount := int64(len(acc.PaidBuyers))
		repeatBuyerCount := int64(len(acc.RepeatBuyers))
		result = append(result, &omsclient.RepeatPurchaseTrendPoint{
			BucketLabel:      acc.BucketLabel,
			BucketStart:      acc.BucketStart,
			BucketEnd:        acc.BucketEnd,
			PaidBuyerCount:   paidBuyerCount,
			RepeatBuyerCount: repeatBuyerCount,
			RepeatRate:       operatefunnel.SafeRate(repeatBuyerCount, paidBuyerCount),
			RepeatOrderCount: acc.RepeatOrderCount,
			RepeatGmv:        acc.RepeatGmv,
			AvgDaysToRepeat:  avgDays,
		})
	}
	return result
}

func buildRepeatPurchaseDetailRowsAndOverview(events []repeatPurchaseEvent, currentOrders []repeatPurchaseQualifiedOrderRow) ([]*omsclient.RepeatPurchaseDetailRow, *omsclient.RepeatPurchaseOverview) {
	paidBuyerSet := make(map[int64]struct{}, len(currentOrders))
	for _, row := range currentOrders {
		paidBuyerSet[row.MemberID] = struct{}{}
	}
	memberStats := make(map[int64]*repeatPurchaseMemberAccumulator, len(events))
	for _, event := range events {
		row := event.Order
		stat, ok := memberStats[row.MemberID]
		if !ok {
			stat = &repeatPurchaseMemberAccumulator{
				MemberID:          row.MemberID,
				FirstValidPayTime: event.FirstValidPayTime,
			}
			memberStats[row.MemberID] = stat
		}
		stat.RepeatOrderCount++
		stat.RepeatGmv += row.PayAmount
		if stat.LatestRepeatPayTime.IsZero() || row.PayTime.After(stat.LatestRepeatPayTime) || (row.PayTime.Equal(stat.LatestRepeatPayTime) && row.OrderID > 0) {
			stat.LatestRepeatPayTime = row.PayTime
			stat.LatestChannel = operatefunnel.OrderChannelFromSource(row.SourceType)
			stat.LatestActivityType = normalizeRepeatPurchaseActivityType(row.ActivityType)
			stat.LatestActivityID = row.ActivityID
			stat.PlatformID = row.PlatformID
			stat.TenantID = row.TenantID
			stat.MerchantID = row.MerchantID
		}
		if event.FirstValidPayTime.Before(stat.FirstValidPayTime) || stat.FirstValidPayTime.IsZero() {
			stat.FirstValidPayTime = event.FirstValidPayTime
		}
	}
	detailRows := make([]*omsclient.RepeatPurchaseDetailRow, 0, len(memberStats))
	repeatOrderCount := int64(0)
	repeatGmv := 0.0
	totalRepeatDays := 0.0
	for _, stat := range memberStats {
		repeatOrderCount += stat.RepeatOrderCount
		repeatGmv += stat.RepeatGmv
		totalRepeatDays += stat.LatestRepeatPayTime.Sub(stat.FirstValidPayTime).Hours() / 24
		detailRows = append(detailRows, &omsclient.RepeatPurchaseDetailRow{
			MemberId:            stat.MemberID,
			FirstValidPayTime:   operatefunnel.FormatDateTime(stat.FirstValidPayTime),
			LatestRepeatPayTime: operatefunnel.FormatDateTime(stat.LatestRepeatPayTime),
			RepeatOrderCount:    stat.RepeatOrderCount,
			RepeatGmv:           stat.RepeatGmv,
			LatestChannel:       stat.LatestChannel,
			LatestActivityType:  stat.LatestActivityType,
			LatestActivityId:    stat.LatestActivityID,
			PlatformId:          stat.PlatformID,
			TenantId:            stat.TenantID,
			MerchantId:          stat.MerchantID,
		})
	}
	sort.Slice(detailRows, func(i, j int) bool {
		if detailRows[i].LatestRepeatPayTime == detailRows[j].LatestRepeatPayTime {
			return detailRows[i].MemberId < detailRows[j].MemberId
		}
		return detailRows[i].LatestRepeatPayTime > detailRows[j].LatestRepeatPayTime
	})
	repeatBuyerCount := int64(len(memberStats))
	avgDays := 0.0
	if repeatBuyerCount > 0 {
		avgDays = totalRepeatDays / float64(repeatBuyerCount)
	}
	return detailRows, &omsclient.RepeatPurchaseOverview{
		PaidBuyerCount:   int64(len(paidBuyerSet)),
		RepeatBuyerCount: repeatBuyerCount,
		RepeatRate:       operatefunnel.SafeRate(repeatBuyerCount, int64(len(paidBuyerSet))),
		RepeatOrderCount: repeatOrderCount,
		RepeatGmv:        repeatGmv,
		AvgDaysToRepeat:  avgDays,
	}
}

func hasPriorRepeatOrder(current repeatPurchaseQualifiedOrderRow, history []repeatPurchaseQualifiedOrderRow) bool {
	lookbackStart, _, err := operatefunnel.RepeatPurchaseLookbackWindow(current.PayTime)
	if err != nil {
		return false
	}
	for _, item := range history {
		if item.MemberID != current.MemberID {
			continue
		}
		if item.PayTime.Before(lookbackStart) {
			continue
		}
		if item.PayTime.After(current.PayTime) {
			break
		}
		if item.PayTime.Equal(current.PayTime) && item.OrderID >= current.OrderID {
			continue
		}
		if item.PayTime.Before(current.PayTime) || (item.PayTime.Equal(current.PayTime) && item.OrderID < current.OrderID) {
			return true
		}
	}
	return false
}

func repeatPurchaseSelectColumns(alias string) string {
	return alias + ".id AS order_id, " +
		alias + ".user_id AS member_id, " +
		alias + ".pay_time, " +
		alias + ".pay_amount, " +
		alias + ".source_type, " +
		alias + ".activity_type, " +
		alias + ".activity_id, " +
		alias + ".platform_id, " +
		alias + ".tenant_id, " +
		alias + ".merchant_id"
}

func collectRepeatPurchaseMemberIDs(rows []repeatPurchaseQualifiedOrderRow) []int64 {
	result := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.MemberID]; ok {
			continue
		}
		seen[row.MemberID] = struct{}{}
		result = append(result, row.MemberID)
	}
	return result
}

func repeatPurchaseBucketStart(value time.Time, bucket string) time.Time {
	local := value.In(operatefunnel.Location())
	if bucket == operatefunnel.BucketHour {
		return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, operatefunnel.Location())
	}
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, operatefunnel.Location())
}

func normalizeRepeatPurchaseActivityType(value string) string {
	if value == "" {
		return operatefunnel.ActivityNone
	}
	return operatefunnel.NormalizeActivityType(value)
}

func normalizeRepeatPurchaseTrackingTime(value interface{}) (time.Time, bool, error) {
	switch typed := value.(type) {
	case nil:
		return time.Time{}, false, nil
	case time.Time:
		return typed, true, nil
	case *time.Time:
		if typed == nil {
			return time.Time{}, false, nil
		}
		return *typed, true, nil
	case string:
		if typed == "" {
			return time.Time{}, false, nil
		}
		parsed, err := parseRepeatPurchaseTimeText(typed)
		if err != nil {
			return time.Time{}, false, err
		}
		return parsed, true, nil
	case []byte:
		if len(typed) == 0 {
			return time.Time{}, false, nil
		}
		parsed, err := parseRepeatPurchaseTimeText(string(typed))
		if err != nil {
			return time.Time{}, false, err
		}
		return parsed, true, nil
	default:
		return time.Time{}, false, fmt.Errorf("无法解析复购追踪起点时间: %T", value)
	}
}

func parseRepeatPurchaseTimeText(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05-07:00",
		time.RFC3339,
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, operatefunnel.Location())
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func newRepeatPurchaseBuilder(ctx context.Context, svcCtx *svc.ServiceContext) *repeatPurchaseQueryBuilder {
	return newRepeatPurchaseQueryBuilder(ctx, svcCtx)
}
