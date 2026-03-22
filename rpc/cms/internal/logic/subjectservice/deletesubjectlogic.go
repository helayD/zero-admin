package subjectservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/cms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteSubjectLogic 删除专题
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type DeleteSubjectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSubjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSubjectLogic {
	return &DeleteSubjectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteSubject 删除专题
func (l *DeleteSubjectLogic) DeleteSubject(in *cmsclient.DeleteSubjectReq) (*cmsclient.DeleteSubjectResp, error) {
	q := query.CmsSubject
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, "")
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureSubjectScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "cms.subject.delete", "", "delete subject"); err != nil {
		return nil, err
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除专题失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除专题失败")
	}

	return &cmsclient.DeleteSubjectResp{}, nil
}
