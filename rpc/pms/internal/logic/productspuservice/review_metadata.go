package productspuservicelogic

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
)

type productReviewMetadata struct {
	ReviewMan    string
	ReviewTime   string
	ReviewDetail string
}

func buildProductScopeType(platformID, tenantID, merchantID int64) string {
	switch {
	case merchantID > 0:
		return "merchant"
	case tenantID > 0:
		return "tenant"
	default:
		return "platform"
	}
}

func loadProductReviewMetadata(
	ctx context.Context,
	reviewModel model.ProductVertifyRecordModel,
	productIDs []int64,
) (map[int64]productReviewMetadata, error) {
	metadata := make(map[int64]productReviewMetadata, len(productIDs))
	for _, productID := range productIDs {
		if productID <= 0 {
			continue
		}

		records, err := reviewModel.FindAll(ctx, productID)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			return nil, err
		}
		if len(records) == 0 {
			continue
		}

		latest := latestReviewRecord(records)
		metadata[productID] = productReviewMetadata{
			ReviewMan:    strings.TrimSpace(latest.ReviewMan),
			ReviewTime:   formatReviewTime(latest),
			ReviewDetail: strings.TrimSpace(latest.Detail),
		}
	}

	return metadata, nil
}

func latestReviewRecord(records []*model.ProductVertifyRecord) *model.ProductVertifyRecord {
	sort.SliceStable(records, func(i, j int) bool {
		return reviewRecordTime(records[i]).After(reviewRecordTime(records[j]))
	})
	return records[0]
}

func formatReviewTime(record *model.ProductVertifyRecord) string {
	ts := reviewRecordTime(record)
	if ts.IsZero() {
		return ""
	}
	return ts.Format("2006-01-02 15:04:05")
}

func reviewRecordTime(record *model.ProductVertifyRecord) time.Time {
	if !record.UpdateAt.IsZero() {
		return record.UpdateAt
	}
	return record.CreateAt
}
