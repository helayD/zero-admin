package operate_dashboard

import (
	"context"
	"sort"
	"time"

	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"google.golang.org/grpc"
)

func normalizeRepeatPurchaseTimeRange(startText, endText string) (string, string, time.Time, time.Time, error) {
	startText, endText = normalizeDashboardTimeRange(startText, endText)
	startTime, endTime, err := operatefunnel.ParseTimeRange(startText, endText)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	return startText, endText, startTime, endTime, nil
}

func normalizeRepeatPurchasePage(pageNum, pageSize int32) (int32, int32) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return pageNum, pageSize
}

func buildRepeatPurchaseOverview(source *omsclient.RepeatPurchaseOverview) types.RepeatPurchaseOverview {
	if source == nil {
		return types.RepeatPurchaseOverview{}
	}
	return types.RepeatPurchaseOverview{
		PaidBuyerCount:   source.PaidBuyerCount,
		RepeatBuyerCount: source.RepeatBuyerCount,
		RepeatRate:       source.RepeatRate,
		RepeatOrderCount: source.RepeatOrderCount,
		RepeatGmv:        source.RepeatGmv,
		AvgDaysToRepeat:  source.AvgDaysToRepeat,
	}
}

func buildRepeatPurchaseTrendPoints(source []*omsclient.RepeatPurchaseTrendPoint) []types.RepeatPurchaseTrendPoint {
	result := make([]types.RepeatPurchaseTrendPoint, 0, len(source))
	for _, item := range source {
		if item == nil {
			continue
		}
		result = append(result, types.RepeatPurchaseTrendPoint{
			BucketLabel:      item.BucketLabel,
			BucketStart:      item.BucketStart,
			BucketEnd:        item.BucketEnd,
			PaidBuyerCount:   item.PaidBuyerCount,
			RepeatBuyerCount: item.RepeatBuyerCount,
			RepeatRate:       item.RepeatRate,
			RepeatOrderCount: item.RepeatOrderCount,
			RepeatGmv:        item.RepeatGmv,
			AvgDaysToRepeat:  item.AvgDaysToRepeat,
		})
	}
	return result
}

type repeatPurchaseActivityOptionService interface {
	QueryOperateActivityOptions(context.Context, *smsclient.QueryOperateActivityOptionsReq, ...grpc.CallOption) (*smsclient.QueryOperateActivityOptionsResp, error)
}

type repeatPurchaseMemberInfoService interface {
	QueryMemberBriefByIds(context.Context, *umsclient.QueryMemberBriefByIdsReq, ...grpc.CallOption) (*umsclient.QueryMemberBriefByIdsResp, error)
}

func queryRepeatPurchaseActivityOptions(ctx context.Context, service repeatPurchaseActivityOptionService, scope *smsclient.GovernanceScope) ([]types.OperateFunnelActivityOption, error) {
	resp, err := service.QueryOperateActivityOptions(ctx, &smsclient.QueryOperateActivityOptionsReq{
		Scope: scope,
	})
	if err != nil {
		return nil, err
	}
	return buildActivityOptions(resp.List), nil
}

func queryRepeatPurchaseMemberBriefMap(ctx context.Context, service repeatPurchaseMemberInfoService, detailRows []*omsclient.RepeatPurchaseDetailRow) (map[int64]*umsclient.MemberBriefData, error) {
	memberIDs := make([]int64, 0, len(detailRows))
	seen := make(map[int64]struct{}, len(detailRows))
	for _, row := range detailRows {
		if row == nil || row.MemberId <= 0 {
			continue
		}
		if _, ok := seen[row.MemberId]; ok {
			continue
		}
		seen[row.MemberId] = struct{}{}
		memberIDs = append(memberIDs, row.MemberId)
	}
	result := make(map[int64]*umsclient.MemberBriefData, len(memberIDs))
	if len(memberIDs) == 0 {
		return result, nil
	}
	resp, err := service.QueryMemberBriefByIds(ctx, &umsclient.QueryMemberBriefByIdsReq{MemberIds: memberIDs})
	if err != nil {
		return nil, err
	}
	for _, item := range resp.List {
		if item == nil {
			continue
		}
		result[item.MemberId] = item
	}
	return result, nil
}

func buildRepeatPurchaseDetailItems(detailRows []*omsclient.RepeatPurchaseDetailRow, briefMap map[int64]*umsclient.MemberBriefData) []types.RepeatPurchaseDetailItem {
	items := make([]types.RepeatPurchaseDetailItem, 0, len(detailRows))
	for _, row := range detailRows {
		if row == nil {
			continue
		}
		brief := briefMap[row.MemberId]
		item := types.RepeatPurchaseDetailItem{
			MemberId:            row.MemberId,
			FirstValidPayTime:   row.FirstValidPayTime,
			LatestRepeatPayTime: row.LatestRepeatPayTime,
			RepeatOrderCount:    row.RepeatOrderCount,
			RepeatGmv:           row.RepeatGmv,
			LatestChannel:       row.LatestChannel,
			LatestActivityType:  row.LatestActivityType,
			LatestActivityId:    row.LatestActivityId,
			PlatformId:          row.PlatformId,
			TenantId:            row.TenantId,
			MerchantId:          row.MerchantId,
		}
		if brief != nil {
			item.NicknameMasked = brief.NicknameMasked
			item.MobileMasked = brief.MobileMasked
		}
		items = append(items, item)
	}
	return items
}

func mergeRepeatPurchasePartialMetrics(metricGroups ...[]string) []string {
	seen := make(map[string]struct{})
	for _, metrics := range metricGroups {
		for _, metric := range metrics {
			seen[metric] = struct{}{}
		}
	}
	merged := make([]string, 0, len(seen))
	known := make(map[string]struct{})
	for _, definition := range operatefunnel.RepeatPurchaseOverviewDefinitions() {
		known[definition.Key] = struct{}{}
		if _, ok := seen[definition.Key]; ok {
			merged = append(merged, definition.Key)
		}
	}
	remaining := make([]string, 0)
	for metric := range seen {
		if _, ok := known[metric]; ok {
			continue
		}
		remaining = append(remaining, metric)
	}
	sort.Strings(remaining)
	merged = append(merged, remaining...)
	return merged
}

func repeatPurchaseExportHeader() []string {
	columns := operatefunnel.RepeatPurchaseDetailColumns()
	headers := make([]string, 0, len(columns))
	for _, column := range columns {
		headers = append(headers, column.Label)
	}
	return headers
}

func repeatPurchaseExportRow(item types.RepeatPurchaseDetailItem) []interface{} {
	values := make([]interface{}, 0, len(operatefunnel.RepeatPurchaseDetailColumns()))
	for _, column := range operatefunnel.RepeatPurchaseDetailColumns() {
		switch column.Key {
		case "memberId":
			values = append(values, item.MemberId)
		case "nicknameMasked":
			values = append(values, item.NicknameMasked)
		case "mobileMasked":
			values = append(values, item.MobileMasked)
		case "firstValidPayTime":
			values = append(values, item.FirstValidPayTime)
		case "latestRepeatPayTime":
			values = append(values, item.LatestRepeatPayTime)
		case "repeatOrderCount":
			values = append(values, item.RepeatOrderCount)
		case "repeatGmv":
			values = append(values, item.RepeatGmv)
		case "latestChannel":
			values = append(values, item.LatestChannel)
		case "latestActivityType":
			values = append(values, item.LatestActivityType)
		case "latestActivityId":
			values = append(values, item.LatestActivityId)
		case "platformId":
			values = append(values, item.PlatformId)
		case "tenantId":
			values = append(values, item.TenantId)
		case "merchantId":
			values = append(values, item.MerchantId)
		default:
			values = append(values, "")
		}
	}
	return values
}
