package commentservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockConsistencyCommentModel struct {
	comment              *model.ProductComment
	restoreAuditInvoked  bool
	restoreAppealInvoked bool
}

func (m *mockConsistencyCommentModel) Insert(context.Context, *model.ProductComment) error {
	return nil
}

func (m *mockConsistencyCommentModel) FindOne(_ context.Context, id string) (*model.ProductComment, error) {
	if m.comment == nil || m.comment.ID.Hex() != id {
		return nil, model.ErrNotFound
	}
	cloned := *m.comment
	return &cloned, nil
}

func (m *mockConsistencyCommentModel) Update(context.Context, *model.ProductComment) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{}, nil
}

func (m *mockConsistencyCommentModel) Delete(context.Context, string) (int64, error) {
	return 0, nil
}

func (m *mockConsistencyCommentModel) FindPage(context.Context, int64, int64, int64, int64, int64, int64, int32) ([]*model.ProductComment, int64, error) {
	return nil, 0, nil
}

func (m *mockConsistencyCommentModel) FindPageWithAuditStatus(context.Context, model.CommentQueryFilter) ([]*model.ProductComment, int64, error) {
	return nil, 0, nil
}

func (m *mockConsistencyCommentModel) FindOneByMemberProductOrder(context.Context, int64, int64, int64) (*model.ProductComment, error) {
	return nil, model.ErrNotFound
}

func (m *mockConsistencyCommentModel) CountByProductId(context.Context, int64) (int64, error) {
	return 0, nil
}

func (m *mockConsistencyCommentModel) AvgStarByProductId(context.Context, int64) (float64, error) {
	return 0, nil
}

func (m *mockConsistencyCommentModel) UpdateStatus(_ context.Context, _ string, showStatus int32, updateBy string) error {
	if m.comment == nil {
		return errors.New("comment missing")
	}
	m.comment.ShowStatus = showStatus
	m.comment.UpdateBy = updateBy
	if showStatus == 1 {
		m.comment.AuditStatus = 1
		m.comment.Hidden = 0
	} else {
		m.comment.AuditStatus = 3
		m.comment.Hidden = 1
	}
	return nil
}

func (m *mockConsistencyCommentModel) UpdateAuditStatus(_ context.Context, _ string, auditStatus int32, hidden int32, auditRemark string, auditorId int64, auditorName string) error {
	if m.comment == nil {
		return errors.New("comment missing")
	}
	m.comment.AuditStatus = auditStatus
	m.comment.Hidden = hidden
	m.comment.AuditRemark = auditRemark
	m.comment.AuditorId = auditorId
	m.comment.AuditorName = auditorName
	if auditStatus == 1 && hidden == 0 {
		m.comment.ShowStatus = 1
	} else {
		m.comment.ShowStatus = 0
	}
	return nil
}

func (m *mockConsistencyCommentModel) RestoreAuditSnapshot(_ context.Context, _ string, snapshot model.CommentAuditSnapshot) error {
	if m.comment == nil {
		return errors.New("comment missing")
	}
	m.restoreAuditInvoked = true
	m.comment.ShowStatus = snapshot.ShowStatus
	m.comment.AuditStatus = snapshot.AuditStatus
	m.comment.Hidden = snapshot.Hidden
	m.comment.AuditRemark = snapshot.AuditRemark
	m.comment.AuditorId = snapshot.AuditorID
	m.comment.AuditorName = snapshot.AuditorName
	m.comment.AuditedAt = snapshot.AuditedAt
	m.comment.UpdateBy = snapshot.UpdateBy
	m.comment.UpdateAt = snapshot.UpdateAt
	return nil
}

func (m *mockConsistencyCommentModel) UpdateAppeal(_ context.Context, _ string, update model.CommentAppealUpdate) error {
	if m.comment == nil {
		return errors.New("comment missing")
	}
	m.comment.AppealStatus = update.AppealStatus
	m.comment.AppealReason = update.AppealReason
	m.comment.AppealReply = update.AppealReply
	if update.AppealedAt != nil {
		m.comment.AppealedAt = *update.AppealedAt
	}
	if update.AppealHandledAt != nil {
		m.comment.AppealHandledAt = *update.AppealHandledAt
	}
	m.comment.UpdateBy = update.UpdateBy
	return nil
}

func (m *mockConsistencyCommentModel) RestoreAppealSnapshot(_ context.Context, _ string, snapshot model.CommentAppealSnapshot) error {
	if m.comment == nil {
		return errors.New("comment missing")
	}
	m.restoreAppealInvoked = true
	m.comment.AppealStatus = snapshot.AppealStatus
	m.comment.AppealReason = snapshot.AppealReason
	m.comment.AppealReply = snapshot.AppealReply
	m.comment.AppealedAt = snapshot.AppealedAt
	m.comment.AppealHandledAt = snapshot.AppealHandledAt
	m.comment.UpdateBy = snapshot.UpdateBy
	m.comment.UpdateAt = snapshot.UpdateAt
	return nil
}

func (m *mockConsistencyCommentModel) RestoreComment(_ context.Context, _ string, auditorId int64, auditorName string) error {
	if m.comment == nil {
		return errors.New("comment missing")
	}
	m.comment.ShowStatus = 1
	m.comment.AuditStatus = 1
	m.comment.Hidden = 0
	m.comment.AuditorId = auditorId
	m.comment.AuditorName = auditorName
	return nil
}

func newAuditFailureDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	return db
}

func newAuditSuccessDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := newAuditFailureDB(t)
	if err := db.Exec(`
		CREATE TABLE pms_comment_audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comment_id TEXT NOT NULL,
			platform_id INTEGER NOT NULL DEFAULT 0,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			action TEXT NOT NULL DEFAULT '',
			from_status INTEGER NOT NULL DEFAULT 0,
			to_status INTEGER NOT NULL DEFAULT 0,
			operator_id INTEGER NOT NULL DEFAULT 0,
			operator_name TEXT NOT NULL DEFAULT '',
			remark TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		t.Fatalf("create audit log table failed: %v", err)
	}
	return db
}

func TestUpdateCommentRollsBackWhenAuditLogInsertFails(t *testing.T) {
	t.Parallel()

	comment := &model.ProductComment{
		ID:          bson.NewObjectID(),
		ShowStatus:  0,
		AuditStatus: 0,
		Hidden:      0,
		PlatformId:  1,
		TenantId:    88,
		MerchantId:  3001,
	}
	commentModel := &mockConsistencyCommentModel{comment: comment}
	logic := NewUpdateCommentLogic(context.Background(), &svc.ServiceContext{
		DB:                  newAuditFailureDB(t),
		ProductCommentModel: commentModel,
	})

	_, err := logic.UpdateComment(&pmsclient.UpdateCommentReq{
		Id:          comment.ID.Hex(),
		AuditStatus: 1,
		AuditRemark: "通过",
		AuditorId:   1001,
		UpdateBy:    "审核员",
		PlatformId:  1,
		TenantId:    88,
		MerchantId:  3001,
	})
	if err == nil {
		t.Fatal("expected audit log failure to be returned")
	}
	if !commentModel.restoreAuditInvoked {
		t.Fatal("expected audit snapshot rollback to run")
	}
	if comment.AuditStatus != 0 || comment.Hidden != 0 || comment.ShowStatus != 0 {
		t.Fatalf("expected comment state rolled back, got audit=%d hidden=%d show=%d", comment.AuditStatus, comment.Hidden, comment.ShowStatus)
	}
}

func TestRestoreCommentRollsBackWhenAuditLogInsertFails(t *testing.T) {
	t.Parallel()

	comment := &model.ProductComment{
		ID:          bson.NewObjectID(),
		ShowStatus:  0,
		AuditStatus: 3,
		Hidden:      1,
		PlatformId:  1,
		TenantId:    88,
		MerchantId:  3001,
	}
	commentModel := &mockConsistencyCommentModel{comment: comment}
	logic := NewRestoreCommentLogic(context.Background(), &svc.ServiceContext{
		DB:                  newAuditFailureDB(t),
		ProductCommentModel: commentModel,
	})

	_, err := logic.RestoreComment(&pmsclient.RestoreCommentReq{
		Id:           comment.ID.Hex(),
		AuditorId:    1001,
		OperatorName: "审核员",
		PlatformId:   1,
		TenantId:     88,
		MerchantId:   3001,
	})
	if err == nil {
		t.Fatal("expected restore log failure to be returned")
	}
	if !commentModel.restoreAuditInvoked {
		t.Fatal("expected restore rollback to run")
	}
	if comment.AuditStatus != 3 || comment.Hidden != 1 || comment.ShowStatus != 0 {
		t.Fatalf("expected comment state rolled back, got audit=%d hidden=%d show=%d", comment.AuditStatus, comment.Hidden, comment.ShowStatus)
	}
}

func TestSubmitCommentAppealClearsHandledAtAfterResubmission(t *testing.T) {
	t.Parallel()

	comment := &model.ProductComment{
		ID:              bson.NewObjectID(),
		ProductId:       9527,
		MemberId:        2001,
		MemberNickName:  "张三",
		AuditStatus:     3,
		Hidden:          1,
		AppealStatus:    3,
		AppealReason:    "旧申诉原因",
		AppealReply:     "已驳回",
		AppealedAt:      time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC),
		AppealHandledAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		PlatformId:      1,
		TenantId:        88,
		MerchantId:      3001,
	}
	commentModel := &mockConsistencyCommentModel{comment: comment}
	logic := NewSubmitCommentAppealLogic(context.Background(), &svc.ServiceContext{
		DB:                  newAuditSuccessDB(t),
		ProductCommentModel: commentModel,
	})

	_, err := logic.SubmitCommentAppeal(&pmsclient.SubmitCommentAppealReq{
		Id:           comment.ID.Hex(),
		MemberId:     2001,
		OperatorName: "张三",
		AppealReason: "请重新复核",
		PlatformId:   1,
		TenantId:     88,
		MerchantId:   3001,
	})
	if err != nil {
		t.Fatalf("SubmitCommentAppeal returned error: %v", err)
	}
	if comment.AppealStatus != 1 {
		t.Fatalf("expected appeal status to become pending, got %d", comment.AppealStatus)
	}
	if !comment.AppealHandledAt.IsZero() {
		t.Fatalf("expected handledAt to be cleared on resubmission, got %v", comment.AppealHandledAt)
	}
	if comment.AppealReply != "" {
		t.Fatalf("expected old appeal reply to be cleared, got %q", comment.AppealReply)
	}
}

func TestHandleCommentAppealRollsBackWhenAuditLogInsertFails(t *testing.T) {
	t.Parallel()

	comment := &model.ProductComment{
		ID:           bson.NewObjectID(),
		ProductId:    9527,
		ShowStatus:   0,
		AuditStatus:  3,
		Hidden:       1,
		AuditRemark:  "内容违规",
		MemberId:     2001,
		AppealStatus: 1,
		AppealReason: "请人工复核",
		AppealedAt:   time.Date(2026, 4, 2, 8, 0, 0, 0, time.UTC),
		PlatformId:   1,
		TenantId:     88,
		MerchantId:   3001,
	}
	commentModel := &mockConsistencyCommentModel{comment: comment}
	logic := NewHandleCommentAppealLogic(context.Background(), &svc.ServiceContext{
		DB:                  newAuditFailureDB(t),
		ProductCommentModel: commentModel,
	})

	_, err := logic.HandleCommentAppeal(&pmsclient.HandleCommentAppealReq{
		Id:           comment.ID.Hex(),
		AppealStatus: 2,
		AppealReply:  "申诉通过，已恢复展示",
		OperatorId:   1001,
		OperatorName: "审核员",
		PlatformId:   1,
		TenantId:     88,
		MerchantId:   3001,
	})
	if err == nil {
		t.Fatal("expected appeal log failure to be returned")
	}
	if !commentModel.restoreAppealInvoked {
		t.Fatal("expected appeal snapshot rollback to run")
	}
	if !commentModel.restoreAuditInvoked {
		t.Fatal("expected audit snapshot rollback to run")
	}
	if comment.AppealStatus != 1 || comment.AuditStatus != 3 || comment.Hidden != 1 || comment.ShowStatus != 0 {
		t.Fatalf("expected comment state rolled back, got appeal=%d audit=%d hidden=%d show=%d", comment.AppealStatus, comment.AuditStatus, comment.Hidden, comment.ShowStatus)
	}
}
