package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
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
		FindPageWithAuditStatus(ctx context.Context, req CommentQueryFilter) ([]*ProductComment, int64, error)
		FindOneByMemberProductOrder(ctx context.Context, memberId, productId, orderId int64) (*ProductComment, error)
		CountByProductId(ctx context.Context, productId int64) (int64, error)
		AvgStarByProductId(ctx context.Context, productId int64) (float64, error)
		UpdateStatus(ctx context.Context, id string, showStatus int32, updateBy string) error
		UpdateAuditStatus(ctx context.Context, id string, auditStatus int32, hidden int32, auditRemark string, auditorId int64, auditorName string) error
		RestoreAuditSnapshot(ctx context.Context, id string, snapshot CommentAuditSnapshot) error
		UpdateAppeal(ctx context.Context, id string, update CommentAppealUpdate) error
		RestoreAppealSnapshot(ctx context.Context, id string, snapshot CommentAppealSnapshot) error
		RestoreComment(ctx context.Context, id string, auditorId int64, auditorName string) error
	}

	CommentAuditSnapshot struct {
		ShowStatus  int32
		AuditStatus int32
		Hidden      int32
		AuditRemark string
		AuditorID   int64
		AuditorName string
		AuditedAt   time.Time
		UpdateBy    string
		UpdateAt    time.Time
	}

	CommentAppealUpdate struct {
		AppealStatus    int32
		AppealReason    string
		AppealReply     string
		AppealedAt      *time.Time
		AppealHandledAt *time.Time
		UpdateBy        string
	}

	CommentAppealSnapshot struct {
		AppealStatus    int32
		AppealReason    string
		AppealReply     string
		AppealedAt      time.Time
		AppealHandledAt time.Time
		UpdateBy        string
		UpdateAt        time.Time
	}

	CommentQueryFilter struct {
		ProductID   int64
		PlatformID  int64
		TenantID    int64
		MerchantID  int64
		PageNo      int64
		PageSize    int64
		ShowStatus  int32
		AuditStatus int32
		Hidden      *int32
		StartTime   string
		EndTime     string
		ProductName string
		MemberName  string
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
	return m.FindPageWithAuditStatus(ctx, CommentQueryFilter{
		ProductID:   productId,
		PlatformID:  platformId,
		TenantID:    tenantId,
		MerchantID:  merchantId,
		PageNo:      pageNo,
		PageSize:    pageSize,
		ShowStatus:  showStatus,
		AuditStatus: -1,
	})
}

func (m *customProductCommentModel) FindPageWithAuditStatus(ctx context.Context, req CommentQueryFilter) ([]*ProductComment, int64, error) {
	var data []*ProductComment
	filter, err := buildCommentQueryFilter(req)
	if err != nil {
		return nil, 0, err
	}

	total, err := m.conn.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	opts := options.Find().
		SetSkip((pageNo - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.D{{Key: "createAt", Value: -1}})

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

// CountByProductId 统计商品已通过评价数量。
func (m *customProductCommentModel) CountByProductId(ctx context.Context, productId int64) (int64, error) {
	filter := bson.M{
		"productId":   productId,
		"showStatus":  1,
		"auditStatus": 1,
		"hidden":      bson.M{"$ne": 1},
	}
	return m.conn.CountDocuments(ctx, filter)
}

// AvgStarByProductId 统计商品已通过评价的平均评分。
func (m *customProductCommentModel) AvgStarByProductId(ctx context.Context, productId int64) (float64, error) {
	filter := bson.M{
		"productId":   productId,
		"showStatus":  1,
		"auditStatus": 1,
		"hidden":      bson.M{"$ne": 1},
	}

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":     nil,
			"avgStar": bson.M{"$avg": "$star"},
		}},
	}

	var result struct {
		AvgStar float64 `bson:"avgStar"`
	}
	err := m.conn.Aggregate(ctx, &result, pipeline)
	if err != nil {
		if err == mon.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return result.AvgStar, nil
}

// UpdateStatus 兼容旧的批量通过/屏蔽接口，同步维护审核元数据。
func (m *customProductCommentModel) UpdateStatus(ctx context.Context, id string, showStatus int32, updateBy string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidObjectId
	}

	auditStatus := int32(3)
	hidden := int32(1)
	if showStatus == 1 {
		auditStatus = 1
		hidden = 0
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"showStatus":  showStatus,
			"auditStatus": auditStatus,
			"hidden":      hidden,
			"auditorName": updateBy,
			"auditedAt":   now,
			"updateBy":    updateBy,
			"updateAt":    now,
		},
	}
	_, err = m.conn.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// UpdateAuditStatus 更新审核状态（审核通过/拒绝/屏蔽）。
func (m *customProductCommentModel) UpdateAuditStatus(ctx context.Context, id string, auditStatus int32, hidden int32, auditRemark string, auditorId int64, auditorName string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidObjectId
	}

	showStatus := int32(0)
	if auditStatus == 1 && hidden == 0 {
		showStatus = 1
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"showStatus":  showStatus,
			"auditStatus": auditStatus,
			"hidden":      hidden,
			"auditRemark": auditRemark,
			"auditorId":   auditorId,
			"auditorName": auditorName,
			"auditedAt":   now,
			"updateBy":    auditorName,
			"updateAt":    now,
		},
	}
	_, err = m.conn.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// RestoreAuditSnapshot 回滚审核相关字段到更新前快照。
func (m *customProductCommentModel) RestoreAuditSnapshot(ctx context.Context, id string, snapshot CommentAuditSnapshot) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidObjectId
	}

	update := bson.M{
		"$set": bson.M{
			"showStatus":  snapshot.ShowStatus,
			"auditStatus": snapshot.AuditStatus,
			"hidden":      snapshot.Hidden,
			"auditRemark": snapshot.AuditRemark,
			"auditorId":   snapshot.AuditorID,
			"auditorName": snapshot.AuditorName,
			"auditedAt":   snapshot.AuditedAt,
			"updateBy":    snapshot.UpdateBy,
			"updateAt":    snapshot.UpdateAt,
		},
	}
	_, err = m.conn.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// UpdateAppeal 更新申诉状态与处理信息。
func (m *customProductCommentModel) UpdateAppeal(ctx context.Context, id string, update CommentAppealUpdate) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidObjectId
	}

	now := time.Now()
	setFields := bson.M{
		"appealStatus": update.AppealStatus,
		"appealReason": update.AppealReason,
		"appealReply":  update.AppealReply,
		"updateBy":     update.UpdateBy,
		"updateAt":     now,
	}
	if update.AppealedAt != nil {
		setFields["appealedAt"] = *update.AppealedAt
	}
	if update.AppealHandledAt != nil {
		setFields["appealHandledAt"] = *update.AppealHandledAt
	}

	_, err = m.conn.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": setFields})
	return err
}

// RestoreAppealSnapshot 回滚申诉相关字段到更新前快照。
func (m *customProductCommentModel) RestoreAppealSnapshot(ctx context.Context, id string, snapshot CommentAppealSnapshot) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidObjectId
	}

	update := bson.M{
		"$set": bson.M{
			"appealStatus":    snapshot.AppealStatus,
			"appealReason":    snapshot.AppealReason,
			"appealReply":     snapshot.AppealReply,
			"appealedAt":      snapshot.AppealedAt,
			"appealHandledAt": snapshot.AppealHandledAt,
			"updateBy":        snapshot.UpdateBy,
			"updateAt":        snapshot.UpdateAt,
		},
	}
	_, err = m.conn.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// RestoreComment 恢复已屏蔽的评价。
func (m *customProductCommentModel) RestoreComment(ctx context.Context, id string, auditorId int64, auditorName string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidObjectId
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"showStatus":  1,
			"auditStatus": 1,
			"hidden":      0,
			"auditorId":   auditorId,
			"auditorName": auditorName,
			"auditedAt":   now,
			"updateBy":    auditorName,
			"updateAt":    now,
		},
	}
	_, err = m.conn.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func buildCommentQueryFilter(req CommentQueryFilter) (bson.M, error) {
	filter := bson.M{}
	if req.ProductID > 0 {
		filter["productId"] = req.ProductID
	}
	if req.PlatformID > 0 {
		filter["platformId"] = req.PlatformID
	}
	if req.TenantID > 0 {
		filter["tenantId"] = req.TenantID
	}
	if req.MerchantID > 0 {
		filter["merchantId"] = req.MerchantID
	}
	if req.ShowStatus >= 0 {
		filter["showStatus"] = req.ShowStatus
	}
	if req.AuditStatus >= 0 {
		filter["auditStatus"] = req.AuditStatus
	}
	if req.Hidden != nil && *req.Hidden >= 0 {
		filter["hidden"] = *req.Hidden
	}
	if strings.TrimSpace(req.ProductName) != "" {
		filter["productName"] = bson.M{"$regex": bson.Regex{Pattern: regexpQuoteMeta(strings.TrimSpace(req.ProductName)), Options: "i"}}
	}
	if strings.TrimSpace(req.MemberName) != "" {
		filter["memberNickName"] = bson.M{"$regex": bson.Regex{Pattern: regexpQuoteMeta(strings.TrimSpace(req.MemberName)), Options: "i"}}
	}

	createAtRange, err := buildCommentCreateAtRange(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	if len(createAtRange) > 0 {
		filter["createAt"] = createAtRange
	}

	return filter, nil
}

func buildCommentCreateAtRange(startTime, endTime string) (bson.M, error) {
	result := bson.M{}
	if strings.TrimSpace(startTime) != "" {
		start, err := parseCommentTime(startTime, false)
		if err != nil {
			return nil, err
		}
		result["$gte"] = start
	}
	if strings.TrimSpace(endTime) != "" {
		end, err := parseCommentTime(endTime, true)
		if err != nil {
			return nil, err
		}
		result["$lte"] = end
	}
	return result, nil
}

func parseCommentTime(value string, endOfDay bool) (time.Time, error) {
	layouts := []string{time.DateTime, time.DateOnly, "2006-01-02 15:04"}
	trimmed := strings.TrimSpace(value)
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			if layout == time.DateOnly && endOfDay {
				return parsed.Add(24*time.Hour - time.Second), nil
			}
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("时间格式无效: %s", value)
}

func regexpQuoteMeta(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		`.`, `\\.`,
		`*`, `\\*`,
		`+`, `\\+`,
		`?`, `\\?`,
		`|`, `\\|`,
		`(`, `\\(`,
		`)`, `\\)`,
		`[`, `\\[`,
		`]`, `\\]`,
		`{`, `\\{`,
		`}`, `\\}`,
		`^`, `\\^`,
		`$`, `\\$`,
	)
	return replacer.Replace(value)
}
