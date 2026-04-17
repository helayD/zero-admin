package drawactivityservicelogic

import (
	"context"
	"errors"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateDrawActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDrawActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDrawActivityLogic {
	return &UpdateDrawActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDrawActivityLogic) UpdateDrawActivity(in *smsclient.UpdateDrawActivityReq) (*smsclient.UpdateDrawActivityResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureDrawActivityScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "sms.draw_activity.update", in.UpdateBy, in.OperatorName, "update draw activity"); err != nil {
		return nil, err
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var row drawActivityRow
		if err := pkgscope.ApplyGovernanceScope(
			tx.WithContext(l.ctx).Table(drawActivityRow{}.TableName()).Where("is_deleted = 0"),
			currentScope,
			"",
		).Where("id = ?", in.Id).Take(&row).Error; err != nil {
			return err
		}
		if err := buildActivityRowFromUpdate(&row, in); err != nil {
			return err
		}
		return saveDrawAggregate(l.ctx, tx, &row, currentScope, in.Templates, in.Pools, in.UpdateBy, in.OperatorName, "updated", false)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("抽卡活动不存在")
		}
		logc.Errorf(l.ctx, "更新抽卡活动失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	return &smsclient.UpdateDrawActivityResp{}, nil
}
