package productspuservicelogic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type reviewMetadataStub struct {
	records map[int64][]*model.ProductVertifyRecord
	errs    map[int64]error
}

func (s *reviewMetadataStub) Insert(ctx context.Context, data *model.ProductVertifyRecord) error {
	return errors.New("not implemented")
}

func (s *reviewMetadataStub) FindOne(ctx context.Context, id string) (*model.ProductVertifyRecord, error) {
	return nil, errors.New("not implemented")
}

func (s *reviewMetadataStub) Update(ctx context.Context, data *model.ProductVertifyRecord) (*mongo.UpdateResult, error) {
	return nil, errors.New("not implemented")
}

func (s *reviewMetadataStub) Delete(ctx context.Context, id string) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *reviewMetadataStub) FindAll(ctx context.Context, productID int64) ([]*model.ProductVertifyRecord, error) {
	if err, ok := s.errs[productID]; ok {
		return nil, err
	}
	return s.records[productID], nil
}

func TestBuildProductScopeType(t *testing.T) {
	tests := []struct {
		name       string
		platformID int64
		tenantID   int64
		merchantID int64
		want       string
	}{
		{name: "merchant takes precedence", platformID: 1, tenantID: 2, merchantID: 3, want: "merchant"},
		{name: "tenant scope", platformID: 1, tenantID: 2, merchantID: 0, want: "tenant"},
		{name: "platform scope", platformID: 1, tenantID: 0, merchantID: 0, want: "platform"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildProductScopeType(tt.platformID, tt.tenantID, tt.merchantID); got != tt.want {
				t.Fatalf("buildProductScopeType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadProductReviewMetadataReturnsLatestRecord(t *testing.T) {
	stub := &reviewMetadataStub{
		records: map[int64][]*model.ProductVertifyRecord{
			101: {
				{
					ProductId: 101,
					ReviewMan: "Alice",
					Detail:    "第一次驳回",
					CreateAt:  time.Date(2026, 3, 20, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
				},
				{
					ProductId: 101,
					ReviewMan: "Bob",
					Detail:    "补图后审核通过",
					UpdateAt:  time.Date(2026, 3, 21, 9, 30, 0, 0, time.FixedZone("CST", 8*3600)),
				},
			},
		},
		errs: map[int64]error{
			102: model.ErrNotFound,
		},
	}

	got, err := loadProductReviewMetadata(context.Background(), stub, []int64{101, 102})
	if err != nil {
		t.Fatalf("loadProductReviewMetadata() error = %v", err)
	}

	meta, ok := got[101]
	if !ok {
		t.Fatalf("expected product 101 metadata to exist")
	}
	if meta.ReviewMan != "Bob" {
		t.Fatalf("expected latest reviewer Bob, got %q", meta.ReviewMan)
	}
	if meta.ReviewDetail != "补图后审核通过" {
		t.Fatalf("unexpected review detail %q", meta.ReviewDetail)
	}
	if meta.ReviewTime != "2026-03-21 09:30:00" {
		t.Fatalf("unexpected review time %q", meta.ReviewTime)
	}
	if _, ok := got[102]; ok {
		t.Fatalf("did not expect metadata for product 102")
	}
}
