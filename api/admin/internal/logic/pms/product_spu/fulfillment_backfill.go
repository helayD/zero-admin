package product_spu

import (
	"context"

	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

// fulfillmentRow 是 backfillFulfillmentFields 内部用的最小投影。
// 这里特意不引用 rpc/pms/gen/model 包，避免 admin-api 反向依赖 RPC 内部 model。
type fulfillmentRow struct {
	ID                int64  `gorm:"column:id"`
	FulfillmentMode   string `gorm:"column:fulfillment_mode"`
	FulfillmentRuleID int64  `gorm:"column:fulfillment_rule_id"`
}

// backfillFulfillmentFields 旁路修复 ProductSpuListData.FulfillmentMode / FulfillmentRuleId。
//
// 背景：rpc/pms/pmsclient/pms.pb.go 的 file descriptor rawDesc 缺失 ProductSpuListData
// 的 fulfillment_mode (tag=50) 与 fulfillment_rule_id (tag=51) 字段元数据，
// 导致 protobuf-go 序列化时不写入这两个字段的 wire bytes，下游反序列化得到的恒为零值。
// 项目规则禁止跑 protoc 重生成 .pb.go，故在 admin-api 这一层用直连 DB 旁路兜底。
//
// 安全性：本函数仅按已经经过 RPC scope 过滤后的 product id 集合查询，不会扩大可见范围。
// 兼容：若 svcCtx.DB 未配置或查询失败，函数静默退出（保持原有行为，不影响主流程）。
func backfillFulfillmentFields(ctx context.Context, db *gorm.DB, list []*pmsclient.ProductSpuListData) {
	if db == nil || len(list) == 0 {
		return
	}

	ids := make([]int64, 0, len(list))
	for _, item := range list {
		if item == nil || item.Id == 0 {
			continue
		}
		ids = append(ids, item.Id)
	}
	if len(ids) == 0 {
		return
	}

	var rows []fulfillmentRow
	if err := db.WithContext(ctx).
		Table("pms_product_spu").
		Select("id, fulfillment_mode, fulfillment_rule_id").
		Where("id IN ?", ids).
		Find(&rows).Error; err != nil {
		logc.Errorf(ctx, "[fulfillment-backfill] 查询失败 ids=%v err=%s", ids, err.Error())
		return
	}

	byID := make(map[int64]fulfillmentRow, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}

	for _, item := range list {
		if item == nil {
			continue
		}
		if row, ok := byID[item.Id]; ok {
			item.FulfillmentMode = row.FulfillmentMode
			item.FulfillmentRuleId = row.FulfillmentRuleID
		}
	}
}
