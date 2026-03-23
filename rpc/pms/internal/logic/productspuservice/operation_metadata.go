package productspuservicelogic

import (
	"context"
	"strings"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	"gorm.io/gorm"
)

type productOperationMetadata struct {
	PublishMan      string
	PublishTime     string
	PublishDetail   string
	RecommendMan    string
	RecommendTime   string
	RecommendDetail string
}

type productOperationMetadataRow struct {
	ID              int64      `gorm:"column:id"`
	PublishMan      string     `gorm:"column:publish_man"`
	PublishTime     *time.Time `gorm:"column:publish_time"`
	PublishDetail   string     `gorm:"column:publish_detail"`
	RecommendMan    string     `gorm:"column:recommend_man"`
	RecommendTime   *time.Time `gorm:"column:recommend_time"`
	RecommendDetail string     `gorm:"column:recommend_detail"`
}

func loadProductOperationMetadata(
	ctx context.Context,
	db *gorm.DB,
	current pkgscope.GovernanceScope,
	productIDs []int64,
) (map[int64]productOperationMetadata, error) {
	metadata := make(map[int64]productOperationMetadata)
	uniqueIDs := pkgscope.UniquePositiveIDs(productIDs)
	if len(uniqueIDs) == 0 {
		return metadata, nil
	}

	var rows []productOperationMetadataRow
	if err := pkgscope.ApplyGovernanceScope(
		db.WithContext(ctx).Table("pms_product_spu"),
		current,
		"",
	).Select(
		"id, publish_man, publish_time, publish_detail, recommend_man, recommend_time, recommend_detail",
	).Where("id IN ? AND is_deleted = 0", uniqueIDs).Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		metadata[row.ID] = productOperationMetadata{
			PublishMan:      strings.TrimSpace(row.PublishMan),
			PublishTime:     time_util.TimeToString(row.PublishTime),
			PublishDetail:   strings.TrimSpace(row.PublishDetail),
			RecommendMan:    strings.TrimSpace(row.RecommendMan),
			RecommendTime:   time_util.TimeToString(row.RecommendTime),
			RecommendDetail: strings.TrimSpace(row.RecommendDetail),
		}
	}

	return metadata, nil
}

func buildAutoRecommendDetail(detail string) string {
	trimmed := strings.TrimSpace(detail)
	if trimmed == "" {
		return "商品下架，已同步取消推荐"
	}
	return "商品下架同步取消推荐：" + trimmed
}
