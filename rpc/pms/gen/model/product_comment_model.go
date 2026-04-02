package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	_ ProductCommentModel = (*customProductCommentModel)(nil)
)

type (
	// ProductCommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductCommentModel.
	ProductCommentModel interface {
		productCommentModel
		FindPage(ctx context.Context, productId, platformId, tenantId, merchantId, pageNo, pageSize int64, showStatus int32) ([]*ProductComment, int64, error)
		FindOneByMemberProductOrder(ctx context.Context, memberId, productId, orderId int64) (*ProductComment, error)
		CountByProductId(ctx context.Context, productId int64) (int64, error)
		AvgStarByProductId(ctx context.Context, productId int64) (float64, error)
		// Review Fix H-NEW-1: 批量审核/屏蔽 - 仅更新 showStatus 和 updateBy 字段
		UpdateStatus(ctx context.Context, id string, showStatus int32, updateBy string) error
	}

	customProductCommentModel struct {
		*defaultProductCommentModel
	}
)

// NewProductCommentModel returns a model for the mongo.
func NewProductCommentModel(url, db, collection string) ProductCommentModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customProductCommentModel{
		defaultProductCommentModel: newDefaultProductCommentModel(conn),
	}
}

func (m *customProductCommentModel) FindPage(ctx context.Context, productId, platformId, tenantId, merchantId, pageNo, pageSize int64, showStatus int32) ([]*ProductComment, int64, error) {
	var data []*ProductComment
	filter := bson.M{
		"productId": productId,
	}

	// 作用域过滤
	if platformId > 0 {
		filter["platformId"] = platformId
	}
	if tenantId > 0 {
		filter["tenantId"] = tenantId
	}
	if merchantId > 0 {
		filter["merchantId"] = merchantId
	}
	// 审核状态过滤（showStatus=-1 表示查全部）
	if showStatus >= 0 {
		filter["showStatus"] = showStatus
	}

	// 统计总数
	total, err := m.conn.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSkip((pageNo-1)*pageSize).SetLimit(pageSize).SetSort(bson.D{{Key: "createAt", Value: -1}})
	err = m.conn.Find(ctx, &data, filter, opts)
	switch err {
	case nil:
		return data, total, nil
	case mon.ErrNotFound:
		return nil, 0, ErrNotFound
	default:
		return nil, 0, err
	}
}

func (m *customProductCommentModel) FindOneByMemberProductOrder(ctx context.Context, memberId, productId, orderId int64) (*ProductComment, error) {
	var data ProductComment
	filter := bson.M{
		"memberId":  memberId,
		"productId": productId,
		"orderId":   orderId,
	}
	err := m.conn.FindOne(ctx, &data, filter)
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// CountByProductId 统计商品已通过评价数量
func (m *customProductCommentModel) CountByProductId(ctx context.Context, productId int64) (int64, error) {
	filter := bson.M{
		"productId":  productId,
		"showStatus": 1, // 只统计已通过的评价
	}
	return m.conn.CountDocuments(ctx, filter)
}

// AvgStarByProductId 统计商品已通过评价的平均评分
func (m *customProductCommentModel) AvgStarByProductId(ctx context.Context, productId int64) (float64, error) {
	filter := bson.M{
		"productId":  productId,
		"showStatus": 1,
	}

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":    nil,
			"avgStar": bson.M{"$avg": "$star"},
		}},
	}

	var result struct {
		AvgStar float64 `bson:"avgStar"`
	}
	err := m.conn.Aggregate(ctx, &result, pipeline)
	if err != nil {
		return 0, err
	}
	return result.AvgStar, nil
}

// UpdateStatus 批量审核/屏蔽 - 仅更新 showStatus 和 updateBy 字段，避免零值覆盖
func (m *customProductCommentModel) UpdateStatus(ctx context.Context, id string, showStatus int32, updateBy string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"showStatus": showStatus,
			"updateBy":   updateBy,
			"updateAt":   time.Now(),
		},
	}
	_, err := m.conn.UpdateOne(ctx, filter, update)
	return err
}
