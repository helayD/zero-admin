package menu_template

import (
	"sort"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"google.golang.org/grpc/status"
)

func grpcError(err error) error {
	s, _ := status.FromError(err)
	return errorx.NewDefaultError(s.Message())
}

func normalizeMenuIDs(menuIDs []int64) []int64 {
	result := make([]int64, 0, len(menuIDs))
	seen := make(map[int64]struct{}, len(menuIDs))

	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		result = append(result, menuID)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func mapMenuTemplateDetail(data *sysclient.QueryMenuTemplateDetailResp) types.QueryMenuTemplateDetailData {
	if data == nil {
		return types.QueryMenuTemplateDetailData{}
	}

	return types.QueryMenuTemplateDetailData{
		Id:         data.Id,
		Name:       data.Name,
		ScopeType:  data.ScopeType,
		PlatformId: data.PlatformId,
		Status:     data.Status,
		Remark:     data.Remark,
		CreateBy:   data.CreateBy,
		CreateTime: data.CreateTime,
		UpdateBy:   data.UpdateBy,
		UpdateTime: data.UpdateTime,
		MenuIds:    data.MenuIds,
	}
}

func mapMenuTemplateListItem(item *sysclient.MenuTemplateListData) *types.QueryMenuTemplateListData {
	if item == nil {
		return &types.QueryMenuTemplateListData{}
	}

	return &types.QueryMenuTemplateListData{
		Id:         item.Id,
		Name:       item.Name,
		ScopeType:  item.ScopeType,
		PlatformId: item.PlatformId,
		Status:     item.Status,
		Remark:     item.Remark,
		CreateBy:   item.CreateBy,
		CreateTime: item.CreateTime,
		UpdateBy:   item.UpdateBy,
		UpdateTime: item.UpdateTime,
		MenuCount:  item.MenuCount,
	}
}

func trimOptionalText(value string) string {
	return strings.TrimSpace(value)
}
